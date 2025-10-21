package importexport

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/casapps/casnotes/internal/notes"
)

// ImportService per CLAUDE.md Import/Export
type ImportService struct {
	db            *sql.DB
	notesService  *notes.NotesService
	maxImportSize int64 // 100MB per CLAUDE.md
}

// NewImportService creates import service
func NewImportService(db *sql.DB) *ImportService {
	return &ImportService{
		db:            db,
		notesService:  notes.NewNotesService(db),
		maxImportSize: 100 * 1024 * 1024, // 100MB
	}
}

// GoogleKeepNote represents Google Keep takeout format
type GoogleKeepNote struct {
	Title         string   `json:"title"`
	TextContent   string   `json:"textContent"`
	Labels        []string `json:"labels"`
	IsPinned      bool     `json:"isPinned"`
	IsArchived    bool     `json:"isArchived"`
	Color         string   `json:"color"`
	CreatedTime   int64    `json:"createdTimestampUsec"`
	ModifiedTime  int64    `json:"userEditedTimestampUsec"`
}

// ImportGoogleKeep imports from Google Keep takeout per CLAUDE.md
func (s *ImportService) ImportGoogleKeep(userID int, zipPath string) (int, error) {
	// Check file size
	stat, err := os.Stat(zipPath)
	if err != nil {
		return 0, err
	}
	if stat.Size() > s.maxImportSize {
		return 0, fmt.Errorf("import file too large: max 100MB")
	}

	// Open ZIP
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	imported := 0

	// Process each file
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, ".json") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}

		var keepNote GoogleKeepNote
		if err := json.NewDecoder(rc).Decode(&keepNote); err != nil {
			rc.Close()
			continue
		}
		rc.Close()

		// Convert to casnotes format
		content := keepNote.TextContent
		if content == "" {
			content = keepNote.Title
		}

		// Determine note type
		noteType := "note"
		if strings.Contains(content, "- [ ]") || strings.Contains(content, "- [x]") {
			noteType = "checklist"
		}

		// Map colors
		color := s.mapGoogleKeepColor(keepNote.Color)

		// Create note
		_, err = s.notesService.CreateNote(
			userID,
			keepNote.Title,
			content,
			noteType,
			"private",
			color,
			keepNote.IsPinned,
		)

		if err == nil {
			imported++
		}
	}

	return imported, nil
}

// JoplinNote represents Joplin JEX format
type JoplinNote struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt int64  `json:"created_time"`
	UpdatedAt int64  `json:"updated_time"`
	IsConflict int   `json:"is_conflict"`
	IsTodo    int    `json:"is_todo"`
}

// ImportJoplin imports from Joplin JEX per CLAUDE.md
func (s *ImportService) ImportJoplin(userID int, jexPath string) (int, error) {
	// JEX is a tar.gz containing JSON files
	// For now, simplified implementation
	stat, err := os.Stat(jexPath)
	if err != nil {
		return 0, err
	}
	if stat.Size() > s.maxImportSize {
		return 0, fmt.Errorf("import file too large: max 100MB")
	}

	// Implementation placeholder - would parse JEX format
	// JEX contains multiple .md files and metadata .json
	return 0, fmt.Errorf("Joplin JEX import not yet implemented")
}

// EvernoteNote represents Evernote ENEX format (XML)
type EvernoteNote struct {
	XMLName   xml.Name `xml:"note"`
	Title     string   `xml:"title"`
	Content   string   `xml:"content"`
	Created   string   `xml:"created"`
	Updated   string   `xml:"updated"`
	Tags      []string `xml:"tag"`
}

type EvernoteExport struct {
	XMLName xml.Name       `xml:"en-export"`
	Notes   []EvernoteNote `xml:"note"`
}

// ImportEvernote imports from Evernote ENEX per CLAUDE.md
func (s *ImportService) ImportEvernote(userID int, enexPath string) (int, error) {
	stat, err := os.Stat(enexPath)
	if err != nil {
		return 0, err
	}
	if stat.Size() > s.maxImportSize {
		return 0, fmt.Errorf("import file too large: max 100MB")
	}

	// Read ENEX file
	data, err := os.ReadFile(enexPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read ENEX: %w", err)
	}

	// Parse XML
	var export EvernoteExport
	if err := xml.Unmarshal(data, &export); err != nil {
		return 0, fmt.Errorf("failed to parse ENEX: %w", err)
	}

	imported := 0

	// Import each note
	for _, enNote := range export.Notes {
		// Convert ENML to markdown (simplified)
		content := s.convertENMLToMarkdown(enNote.Content)

		_, err := s.notesService.CreateNote(
			userID,
			enNote.Title,
			content,
			"note",
			"private",
			"",
			false,
		)

		if err == nil {
			imported++
		}
	}

	return imported, nil
}

// StandardNotesItem represents Standard Notes export format
type StandardNotesItem struct {
	UUID      string                 `json:"uuid"`
	Content   map[string]interface{} `json:"content"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type StandardNotesExport struct {
	Items []StandardNotesItem `json:"items"`
}

// ImportStandardNotes imports from Standard Notes per CLAUDE.md
func (s *ImportService) ImportStandardNotes(userID int, jsonPath string) (int, error) {
	stat, err := os.Stat(jsonPath)
	if err != nil {
		return 0, err
	}
	if stat.Size() > s.maxImportSize {
		return 0, fmt.Errorf("import file too large: max 100MB")
	}

	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var export StandardNotesExport
	if err := json.Unmarshal(data, &export); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	imported := 0

	// Import each note
	for _, item := range export.Items {
		// Extract title and text from content
		title, _ := item.Content["title"].(string)
		text, _ := item.Content["text"].(string)

		if title == "" && text == "" {
			continue
		}

		_, err := s.notesService.CreateNote(
			userID,
			title,
			text,
			"note",
			"private",
			"",
			false,
		)

		if err == nil {
			imported++
		}
	}

	return imported, nil
}

// ImportMarkdown imports plain markdown files per CLAUDE.md
func (s *ImportService) ImportMarkdown(userID int, dirPath string) (int, error) {
	imported := 0

	// Walk directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		// Check size
		if info.Size() > s.maxImportSize {
			return nil // Skip large files
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip on error
		}

		// Extract title from filename
		title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		// Create note
		_, err = s.notesService.CreateNote(
			userID,
			title,
			string(content),
			"note",
			"private",
			"",
			false,
		)

		if err == nil {
			imported++
		}

		return nil
	})

	if err != nil {
		return imported, err
	}

	return imported, nil
}

// ImportOpenGist imports from OpenGist repository per CLAUDE.md
func (s *ImportService) ImportOpenGist(userID int, repoPath string) (int, error) {
	// OpenGist stores gists as files in a git repository
	// Similar to markdown import but with metadata
	return s.ImportMarkdown(userID, repoPath)
}

// Helper functions

func (s *ImportService) mapGoogleKeepColor(keepColor string) string {
	// Map Google Keep colors to casnotes colors
	colorMap := map[string]string{
		"DEFAULT": "",
		"RED":     "red",
		"ORANGE":  "orange",
		"YELLOW":  "yellow",
		"GREEN":   "green",
		"TEAL":    "cyan",
		"BLUE":    "blue",
		"PURPLE":  "purple",
		"PINK":    "pink",
		"BROWN":   "orange",
		"GRAY":    "",
	}

	if color, ok := colorMap[keepColor]; ok {
		return color
	}
	return ""
}

func (s *ImportService) convertENMLToMarkdown(enml string) string {
	// Simplified ENML to Markdown conversion
	// ENML is Evernote's XML-based format
	content := enml

	// Remove XML tags (very simplified)
	content = strings.ReplaceAll(content, "<div>", "\n")
	content = strings.ReplaceAll(content, "</div>", "")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")

	// Convert basic formatting
	content = strings.ReplaceAll(content, "<b>", "**")
	content = strings.ReplaceAll(content, "</b>", "**")
	content = strings.ReplaceAll(content, "<i>", "*")
	content = strings.ReplaceAll(content, "</i>", "*")

	// Strip remaining tags
	// Production would use proper HTML/XML parser
	for strings.Contains(content, "<") && strings.Contains(content, ">") {
		start := strings.Index(content, "<")
		end := strings.Index(content, ">")
		if start >= 0 && end > start {
			content = content[:start] + content[end+1:]
		} else {
			break
		}
	}

	return strings.TrimSpace(content)
}

// ImportResult represents import operation result
type ImportResult struct {
	Format       string    `json:"format"`
	TotalNotes   int       `json:"total_notes"`
	ImportedNotes int      `json:"imported_notes"`
	FailedNotes  int       `json:"failed_notes"`
	Duration     time.Duration `json:"duration"`
	Errors       []string  `json:"errors,omitempty"`
}

// AutoDetectAndImport detects format and imports
func (s *ImportService) AutoDetectAndImport(userID int, filePath string) (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{}

	// Detect format by extension
	ext := strings.ToLower(filepath.Ext(filePath))

	var imported int
	var err error

	switch ext {
	case ".zip":
		// Could be Google Keep
		result.Format = "Google Keep"
		imported, err = s.ImportGoogleKeep(userID, filePath)

	case ".jex":
		result.Format = "Joplin"
		imported, err = s.ImportJoplin(userID, filePath)

	case ".enex":
		result.Format = "Evernote"
		imported, err = s.ImportEvernote(userID, filePath)

	case ".json":
		result.Format = "Standard Notes"
		imported, err = s.ImportStandardNotes(userID, filePath)

	case ".md", ".markdown":
		result.Format = "Markdown"
		imported, err = s.ImportMarkdown(userID, filepath.Dir(filePath))

	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	result.ImportedNotes = imported
	result.Duration = time.Since(start)

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result, nil
}

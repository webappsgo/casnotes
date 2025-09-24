package importexport

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/casapps/casnotes/internal/notes"
)

// ExportService per CLAUDE.md Import/Export
type ExportService struct {
	notesService *notes.NotesService
}

func NewExportService(notesService *notes.NotesService) *ExportService {
	return &ExportService{notesService: notesService}
}

// ExportFormat per CLAUDE.md Export Formats
type ExportFormat string

const (
	ExportCSV      ExportFormat = "csv"      // CSV data export
	ExportMarkdown ExportFormat = "markdown" // ZIP with markdown+JSON
	ExportPDF      ExportFormat = "pdf"      // PDF (single/bulk)
	ExportHTML     ExportFormat = "html"     // HTML static site
)

// ExportNotes per CLAUDE.md Export Formats
func (s *ExportService) ExportNotes(userID int, format ExportFormat) ([]byte, string, error) {
	// Get all user notes (including archived)
	notes, _, err := s.notesService.ListNotes(userID, 10000, 0, false)
	if err != nil {
		return nil, "", err
	}
	
	archivedNotes, _, err := s.notesService.ListNotes(userID, 10000, 0, true)
	if err != nil {
		return nil, "", err
	}
	
	allNotes := append(notes, archivedNotes...)

	switch format {
	case ExportCSV:
		return s.exportCSV(userID, allNotes)
	case ExportMarkdown:
		return s.exportMarkdown(userID, allNotes)
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportCSV per CLAUDE.md CSV export
func (s *ExportService) exportCSV(userID int, notes []notes.Note) ([]byte, string, error) {
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)
	
	// CSV header per CLAUDE.md
	header := []string{"ID", "Title", "Type", "Visibility", "Created", "Updated", "Content"}
	writer.Write(header)
	
	// Write notes
	for _, note := range notes {
		record := []string{
			note.ID,
			note.Title,
			note.NoteType,
			note.Visibility,
			note.CreatedAt.Format("2006-01-02 15:04:05"),
			note.UpdatedAt.Format("2006-01-02 15:04:05"),
			note.Content,
		}
		writer.Write(record)
	}
	
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}
	
	filename := fmt.Sprintf("casnotes-export-%d-%s.csv", userID, time.Now().Format("20060102-150405"))
	return []byte(csvData.String()), filename, nil
}

// exportMarkdown per CLAUDE.md ZIP with markdown+JSON
func (s *ExportService) exportMarkdown(userID int, notes []notes.Note) ([]byte, string, error) {
	// For now, create a simple text representation
	// Full ZIP implementation would go here
	var content strings.Builder
	
	content.WriteString("# casnotes Export\n\n")
	content.WriteString(fmt.Sprintf("Exported: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Notes: %d\n\n", len(notes)))
	
	for _, note := range notes {
		content.WriteString("---\n\n")
		content.WriteString(fmt.Sprintf("# %s\n\n", note.Title))
		content.WriteString(fmt.Sprintf("**Type:** %s | **Visibility:** %s | **Created:** %s\n\n", 
			note.NoteType, note.Visibility, note.CreatedAt.Format("2006-01-02")))
		content.WriteString(note.Content)
		content.WriteString("\n\n")
	}
	
	filename := fmt.Sprintf("casnotes-export-%d-%s.md", userID, time.Now().Format("20060102-150405"))
	return []byte(content.String()), filename, nil
}

// ImportService per CLAUDE.md Import Support
type ImportService struct {
	notesService *notes.NotesService
	maxSize      int64 // 100MB per CLAUDE.md
}

func NewImportService(notesService *notes.NotesService) *ImportService {
	return &ImportService{
		notesService: notesService,
		maxSize:      100 * 1024 * 1024, // 100MB per CLAUDE.md
	}
}

// ImportFormat per CLAUDE.md Import Support
type ImportFormat string

const (
	ImportGoogleKeep   ImportFormat = "google_keep"    // Google Keep takeout
	ImportJoplin       ImportFormat = "joplin"         // Joplin JEX
	ImportEvernote     ImportFormat = "evernote"       // Evernote ENEX
	ImportStandardNotes ImportFormat = "standard_notes" // Standard Notes
	ImportMarkdown     ImportFormat = "markdown"       // Plain markdown
	ImportOpenGist     ImportFormat = "opengist"       // OpenGist repos
)

// ImportResult per CLAUDE.md
type ImportResult struct {
	Success       bool     `json:"success"`
	NotesImported int      `json:"notes_imported"`
	Errors        []string `json:"errors,omitempty"`
	Duration      string   `json:"duration"`
}

// ImportMarkdown per CLAUDE.md (basic implementation)
func (s *ImportService) ImportMarkdown(userID int, content []byte, filename string) (*ImportResult, error) {
	start := time.Now()
	
	if int64(len(content)) > s.maxSize {
		return nil, fmt.Errorf("file too large (max 100MB)")
	}
	
	// Create note from markdown content
	title := strings.TrimSuffix(filename, ".md")
	if title == "" {
		title = "Imported Note"
	}
	
	note, err := s.notesService.CreateNote(userID, title, string(content), "note", "private", "", false)
	if err != nil {
		return &ImportResult{
			Success: false,
			Errors:  []string{fmt.Sprintf("Failed to create note: %v", err)},
			Duration: time.Since(start).String(),
		}, nil
	}
	
	return &ImportResult{
		Success:       true,
		NotesImported: 1,
		Duration:      time.Since(start).String(),
	}, nil
}
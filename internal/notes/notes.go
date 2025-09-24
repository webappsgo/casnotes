package notes

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Note per CLAUDE.md Note System
type Note struct {
	ID         string    `json:"id"`
	UserID     int       `json:"user_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	NoteType   string    `json:"note_type"`   // note, code, checklist, canvas, encrypted
	Visibility string    `json:"visibility"`  // private, unlisted, public
	Color      string    `json:"color"`
	Pinned     bool      `json:"pinned"`
	Archived   bool      `json:"archived"`
	Encrypted  bool      `json:"encrypted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NotesService per CLAUDE.md
type NotesService struct {
	db      *sql.DB
	maxSize int64 // 10MB per CLAUDE.md Resource Limits
}

func NewNotesService(db *sql.DB) *NotesService {
	return &NotesService{
		db:      db,
		maxSize: 10 * 1024 * 1024, // 10MB per CLAUDE.md
	}
}

// CreateNote per CLAUDE.md Note System
func (s *NotesService) CreateNote(userID int, title, content, noteType, visibility, color string, pinned bool) (*Note, error) {
	// Validate note type per CLAUDE.md
	validTypes := []string{"note", "code", "checklist", "canvas", "encrypted"}
	if noteType == "" {
		noteType = "note"
	}
	if !contains(validTypes, noteType) {
		return nil, fmt.Errorf("invalid note type")
	}

	// Validate visibility per CLAUDE.md
	validVisibility := []string{"private", "unlisted", "public"}
	if visibility == "" {
		visibility = "private" // Default per CLAUDE.md
	}
	if !contains(validVisibility, visibility) {
		visibility = "private"
	}

	// Check size per CLAUDE.md Resource Limits
	if int64(len(content)) > s.maxSize {
		return nil, fmt.Errorf("note content too large (max %d MB)", s.maxSize/(1024*1024))
	}

	// Generate UUID per CLAUDE.md
	noteID := uuid.New().String()

	// Insert note
	_, err := s.db.Exec(`
		INSERT INTO notes (id, user_id, title, content, note_type, visibility, color, pinned, archived, encrypted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		noteID, userID, title, content, noteType, visibility, color, pinned, false, noteType == "encrypted")
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %v", err)
	}

	return s.GetNote(userID, noteID)
}

// GetNote retrieves note by ID
func (s *NotesService) GetNote(userID int, noteID string) (*Note, error) {
	note := &Note{}
	err := s.db.QueryRow(`
		SELECT id, user_id, title, content, note_type, visibility, color, pinned, archived, encrypted, created_at, updated_at
		FROM notes WHERE id = ? AND user_id = ?`, noteID, userID).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.NoteType,
		&note.Visibility, &note.Color, &note.Pinned, &note.Archived, &note.Encrypted,
		&note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, err
	}
	return note, nil
}

// ListNotes per CLAUDE.md with pagination (50 items default)
func (s *NotesService) ListNotes(userID int, limit, offset int, archived bool) ([]Note, int, error) {
	if limit == 0 {
		limit = 50 // Default per CLAUDE.md User Preferences
	}

	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM notes WHERE user_id = ? AND archived = ?", 
		userID, archived).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get notes
	rows, err := s.db.Query(`
		SELECT id, user_id, title, content, note_type, visibility, color, pinned, archived, encrypted, created_at, updated_at
		FROM notes WHERE user_id = ? AND archived = ?
		ORDER BY pinned DESC, updated_at DESC
		LIMIT ? OFFSET ?`, userID, archived, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(
			&note.ID, &note.UserID, &note.Title, &note.Content, &note.NoteType,
			&note.Visibility, &note.Color, &note.Pinned, &note.Archived, &note.Encrypted,
			&note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		notes = append(notes, note)
	}

	return notes, total, nil
}

// UpdateNote updates existing note
func (s *NotesService) UpdateNote(userID int, noteID, title, content string) (*Note, error) {
	// Check size per CLAUDE.md
	if int64(len(content)) > s.maxSize {
		return nil, fmt.Errorf("note content too large")
	}

	_, err := s.db.Exec(`
		UPDATE notes SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND user_id = ?`, title, content, noteID, userID)
	if err != nil {
		return nil, err
	}

	return s.GetNote(userID, noteID)
}

// DeleteNote removes note
func (s *NotesService) DeleteNote(userID int, noteID string) error {
	result, err := s.db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?", noteID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("note not found")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
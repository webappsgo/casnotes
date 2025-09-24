package notes

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Tag per CLAUDE.md Organization (color-coded, max 20 per note)
type Tag struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// Notebook per CLAUDE.md Organization (unlimited nesting)
type Notebook struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ParentID  *int      `json:"parent_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagsService per CLAUDE.md
type TagsService struct {
	db *sql.DB
}

func NewTagsService(db *sql.DB) *TagsService {
	return &TagsService{db: db}
}

// CreateTag per CLAUDE.md (color-coded)
func (s *TagsService) CreateTag(userID int, name, color string) (*Tag, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	// Check if tag exists for user
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tags WHERE user_id = ? AND name = ?", userID, name).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("tag already exists")
	}

	// Insert tag
	result, err := s.db.Exec("INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, name, color)
	if err != nil {
		return nil, err
	}

	tagID, _ := result.LastInsertId()
	return s.GetTag(userID, int(tagID))
}

// GetTag retrieves tag by ID
func (s *TagsService) GetTag(userID, tagID int) (*Tag, error) {
	tag := &Tag{}
	err := s.db.QueryRow(`
		SELECT id, user_id, name, color, created_at
		FROM tags WHERE id = ? AND user_id = ?`, tagID, userID).Scan(
		&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, err
	}
	return tag, nil
}

// ListTags per CLAUDE.md
func (s *TagsService) ListTags(userID int) ([]Tag, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, color, created_at
		FROM tags WHERE user_id = ?
		ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// NotebooksService per CLAUDE.md unlimited nesting
type NotebooksService struct {
	db *sql.DB
}

func NewNotebooksService(db *sql.DB) *NotebooksService {
	return &NotebooksService{db: db}
}

// CreateNotebook per CLAUDE.md (unlimited nesting)
func (s *NotebooksService) CreateNotebook(userID int, name, color string, parentID *int) (*Notebook, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("notebook name is required")
	}

	// Validate parent exists if specified
	if parentID != nil {
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM notebooks WHERE id = ? AND user_id = ?", *parentID, userID).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, fmt.Errorf("parent notebook not found")
		}
	}

	// Insert notebook
	result, err := s.db.Exec("INSERT INTO notebooks (user_id, parent_id, name, color) VALUES (?, ?, ?, ?)", 
		userID, parentID, name, color)
	if err != nil {
		return nil, err
	}

	notebookID, _ := result.LastInsertId()
	return s.GetNotebook(userID, int(notebookID))
}

// GetNotebook retrieves notebook by ID
func (s *NotebooksService) GetNotebook(userID, notebookID int) (*Notebook, error) {
	notebook := &Notebook{}
	err := s.db.QueryRow(`
		SELECT id, user_id, parent_id, name, color, created_at, updated_at
		FROM notebooks WHERE id = ? AND user_id = ?`, notebookID, userID).Scan(
		&notebook.ID, &notebook.UserID, &notebook.ParentID, &notebook.Name, 
		&notebook.Color, &notebook.CreatedAt, &notebook.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notebook not found")
		}
		return nil, err
	}
	return notebook, nil
}

// ListNotebooks per CLAUDE.md with hierarchical structure
func (s *NotebooksService) ListNotebooks(userID int) ([]Notebook, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, parent_id, name, color, created_at, updated_at
		FROM notebooks WHERE user_id = ?
		ORDER BY parent_id, name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notebooks []Notebook
	for rows.Next() {
		var notebook Notebook
		err := rows.Scan(&notebook.ID, &notebook.UserID, &notebook.ParentID, 
			&notebook.Name, &notebook.Color, &notebook.CreatedAt, &notebook.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notebooks = append(notebooks, notebook)
	}
	return notebooks, nil
}
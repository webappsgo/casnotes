package notes

import (
	"database/sql"
	"strings"
)

// SearchService per CLAUDE.md Search (SQLite FTS5 with fallback)
type SearchService struct {
	db     *sql.DB
	hasFTS bool
}

func NewSearchService(db *sql.DB) *SearchService {
	service := &SearchService{db: db}
	
	// Test if FTS5 is available per CLAUDE.md
	_, err := db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS test_fts USING fts5(content)")
	if err == nil {
		service.hasFTS = true
		db.Exec("DROP TABLE test_fts") // Clean up test table
	}
	
	return service
}

// SearchResult per CLAUDE.md with 25 results default
type SearchResult struct {
	Notes   []Note `json:"notes"`
	Total   int    `json:"total"`
	HasMore bool   `json:"has_more"`
}

// SearchNotes per CLAUDE.md Search functionality
func (s *SearchService) SearchNotes(userID int, query string, limit, offset int) (*SearchResult, error) {
	if limit == 0 {
		limit = 25 // Default search results per CLAUDE.md
	}

	if s.hasFTS {
		return s.searchWithFTS5(userID, query, limit, offset)
	}
	
	// Fallback to LIKE search per CLAUDE.md
	return s.searchWithLike(userID, query, limit, offset)
}

// searchWithFTS5 uses SQLite FTS5 per CLAUDE.md
func (s *SearchService) searchWithFTS5(userID int, query string, limit, offset int) (*SearchResult, error) {
	// Create FTS table if it doesn't exist
	createFTS := `
	CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
		note_id,
		title,
		content,
		content=notes,
		content_rowid=rowid
	)`
	s.db.Exec(createFTS)

	// Sync FTS table
	s.db.Exec("INSERT OR REPLACE INTO notes_fts(note_id, title, content) SELECT id, title, content FROM notes WHERE user_id = ?", userID)

	// Search with FTS5
	searchQuery := `
	SELECT n.id, n.user_id, n.title, n.content, n.note_type, n.visibility, n.color, 
	       n.pinned, n.archived, n.encrypted, n.created_at, n.updated_at
	FROM notes n
	JOIN notes_fts fts ON n.id = fts.note_id
	WHERE n.user_id = ? AND notes_fts MATCH ? AND n.archived = false
	ORDER BY rank
	LIMIT ? OFFSET ?`

	return s.executeSearch(searchQuery, userID, query, limit, offset)
}

// searchWithLike uses LIKE fallback per CLAUDE.md
func (s *SearchService) searchWithLike(userID int, query string, limit, offset int) (*SearchResult, error) {
	searchTerm := "%" + strings.ToLower(query) + "%"
	
	searchQuery := `
	SELECT id, user_id, title, content, note_type, visibility, color, pinned, archived, encrypted, created_at, updated_at
	FROM notes 
	WHERE user_id = ? AND archived = false AND (
		LOWER(title) LIKE ? OR LOWER(content) LIKE ?
	)
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?`

	return s.executeSearchLike(searchQuery, userID, searchTerm, limit, offset)
}

func (s *SearchService) executeSearch(query string, userID int, searchQuery string, limit, offset int) (*SearchResult, error) {
	// Get total count
	countQuery := strings.Replace(query, "SELECT n.id, n.user_id, n.title, n.content, n.note_type, n.visibility, n.color, n.pinned, n.archived, n.encrypted, n.created_at, n.updated_at", "SELECT COUNT(*)", 1)
	countQuery = strings.Replace(countQuery, "ORDER BY rank LIMIT ? OFFSET ?", "", 1)
	
	var total int
	err := s.db.QueryRow(countQuery, userID, searchQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Execute search
	rows, err := s.db.Query(query, userID, searchQuery, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		notes = append(notes, note)
	}

	return &SearchResult{
		Notes:   notes,
		Total:   total,
		HasMore: offset+len(notes) < total,
	}, nil
}

func (s *SearchService) executeSearchLike(query string, userID int, searchTerm string, limit, offset int) (*SearchResult, error) {
	// Get total count
	countQuery := `
	SELECT COUNT(*)
	FROM notes 
	WHERE user_id = ? AND archived = false AND (
		LOWER(title) LIKE ? OR LOWER(content) LIKE ?
	)`
	
	var total int
	err := s.db.QueryRow(countQuery, userID, searchTerm, searchTerm).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Execute search
	rows, err := s.db.Query(query, userID, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		notes = append(notes, note)
	}

	return &SearchResult{
		Notes:   notes,
		Total:   total,
		HasMore: offset+len(notes) < total,
	}, nil
}
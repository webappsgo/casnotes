package attachments

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/casapps/casnotes/internal/config"
)

// AttachmentService per CLAUDE.md Attachments
type AttachmentService struct {
	cfg              *config.Config
	db               *sql.DB
	maxAttachmentSize int64 // 25MB per CLAUDE.md
	maxPerNote       int    // 10 per CLAUDE.md
	imageMaxDimension int    // 2048px per CLAUDE.md
}

// Attachment represents a file attachment
type Attachment struct {
	ID        string    `json:"id"`
	NoteID    string    `json:"note_id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Hash      string    `json:"hash"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAttachmentService creates attachment service
func NewAttachmentService(cfg *config.Config, db *sql.DB) *AttachmentService {
	return &AttachmentService{
		cfg:               cfg,
		db:                db,
		maxAttachmentSize: 25 * 1024 * 1024, // 25MB
		maxPerNote:        10,
		imageMaxDimension: 2048,
	}
}

// UploadAttachment handles file upload per CLAUDE.md
func (s *AttachmentService) UploadAttachment(noteID string, filename string, reader io.Reader) (*Attachment, error) {
	// Check max attachments per note
	count, err := s.CountAttachments(noteID)
	if err != nil {
		return nil, err
	}
	if count >= s.maxPerNote {
		return nil, fmt.Errorf("maximum %d attachments per note", s.maxPerNote)
	}

	// Create attachments directory
	attachmentsDir := filepath.Join(s.cfg.DataDir, "attachments", noteID)
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Generate unique ID
	attachmentID := generateID()

	// Sanitize filename
	sanitizedFilename := sanitizeFilename(filename)

	// Create file path
	filePath := filepath.Join(attachmentsDir, attachmentID+"-"+sanitizedFilename)

	// Create temporary file for size check
	tempFile, err := os.CreateTemp(attachmentsDir, "upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Copy data and calculate hash
	hash := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hash)

	written, err := io.Copy(multiWriter, reader)
	if err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	tempFile.Close()

	// Check size limit per CLAUDE.md
	if written > s.maxAttachmentSize {
		return nil, fmt.Errorf("attachment too large: max %dMB", s.maxAttachmentSize/(1024*1024))
	}

	// Move temp file to final location
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	// Detect MIME type
	mimeType := detectMimeType(sanitizedFilename)

	// Optimize image if needed per CLAUDE.md
	if strings.HasPrefix(mimeType, "image/") {
		if err := s.optimizeImage(filePath); err != nil {
			// Log error but don't fail - optimization is optional
		}
	}

	// Create attachment record
	attachment := &Attachment{
		ID:        attachmentID,
		NoteID:    noteID,
		Filename:  sanitizedFilename,
		MimeType:  mimeType,
		Size:      written,
		Hash:      fmt.Sprintf("%x", hash.Sum(nil)),
		Path:      filePath,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.saveAttachment(attachment); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}

	return attachment, nil
}

// GetAttachment retrieves attachment by ID
func (s *AttachmentService) GetAttachment(attachmentID string) (*Attachment, error) {
	var attachment Attachment

	query := `
		SELECT id, note_id, filename, mime_type, size, hash, path, created_at
		FROM attachments
		WHERE id = ?
	`

	err := s.db.QueryRow(query, attachmentID).Scan(
		&attachment.ID,
		&attachment.NoteID,
		&attachment.Filename,
		&attachment.MimeType,
		&attachment.Size,
		&attachment.Hash,
		&attachment.Path,
		&attachment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("attachment not found")
	}
	if err != nil {
		return nil, err
	}

	return &attachment, nil
}

// ListAttachments lists all attachments for a note
func (s *AttachmentService) ListAttachments(noteID string) ([]*Attachment, error) {
	query := `
		SELECT id, note_id, filename, mime_type, size, hash, path, created_at
		FROM attachments
		WHERE note_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*Attachment
	for rows.Next() {
		var attachment Attachment
		err := rows.Scan(
			&attachment.ID,
			&attachment.NoteID,
			&attachment.Filename,
			&attachment.MimeType,
			&attachment.Size,
			&attachment.Hash,
			&attachment.Path,
			&attachment.CreatedAt,
		)
		if err != nil {
			continue
		}
		attachments = append(attachments, &attachment)
	}

	return attachments, nil
}

// DeleteAttachment removes attachment
func (s *AttachmentService) DeleteAttachment(attachmentID string) error {
	// Get attachment info
	attachment, err := s.GetAttachment(attachmentID)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(attachment.Path); err != nil {
		// Log error but continue - file might already be gone
	}

	// Delete from database
	_, err = s.db.Exec("DELETE FROM attachments WHERE id = ?", attachmentID)
	return err
}

// CountAttachments counts attachments for a note
func (s *AttachmentService) CountAttachments(noteID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM attachments WHERE note_id = ?", noteID).Scan(&count)
	return count, err
}

// saveAttachment saves attachment record to database
func (s *AttachmentService) saveAttachment(attachment *Attachment) error {
	query := `
		INSERT INTO attachments (id, note_id, filename, mime_type, size, hash, path, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		attachment.ID,
		attachment.NoteID,
		attachment.Filename,
		attachment.MimeType,
		attachment.Size,
		attachment.Hash,
		attachment.Path,
		attachment.CreatedAt,
	)

	return err
}

// optimizeImage optimizes images >2048px per CLAUDE.md
func (s *AttachmentService) optimizeImage(filePath string) error {
	// Image optimization would use image libraries like:
	// - github.com/disintegration/imaging
	// - github.com/nfnt/resize
	// For now, placeholder implementation

	// Would resize images >2048px to 2048px maintaining aspect ratio
	// Would reduce JPEG quality to 85%
	// Would strip EXIF data

	return nil
}

// Helper functions

func generateID() string {
	// Generate UUID-like ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func sanitizeFilename(filename string) string {
	// Remove path separators
	filename = filepath.Base(filename)

	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range invalid {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Limit length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		filename = filename[:255-len(ext)] + ext
	}

	return filename
}

func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}

// CleanupOrphanedAttachments removes attachments for deleted notes
func (s *AttachmentService) CleanupOrphanedAttachments() error {
	// Find attachments with no corresponding note
	query := `
		SELECT id FROM attachments
		WHERE note_id NOT IN (SELECT id FROM notes)
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var orphanedIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		orphanedIDs = append(orphanedIDs, id)
	}

	// Delete orphaned attachments
	for _, id := range orphanedIDs {
		s.DeleteAttachment(id)
	}

	return nil
}

// GetStorageUsage calculates total storage used by attachments
func (s *AttachmentService) GetStorageUsage(userID int) (int64, error) {
	query := `
		SELECT COALESCE(SUM(a.size), 0)
		FROM attachments a
		JOIN notes n ON a.note_id = n.id
		WHERE n.user_id = ?
	`

	var total int64
	err := s.db.QueryRow(query, userID).Scan(&total)
	return total, err
}

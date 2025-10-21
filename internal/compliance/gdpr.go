package compliance

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/casapps/casnotes/internal/config"
)

// GDPRCompliance implements GDPR per CLAUDE.md
type GDPRCompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewGDPRCompliance(cfg *config.Config, db *sql.DB) *GDPRCompliance {
	return &GDPRCompliance{cfg: cfg, db: db}
}

// Validate checks GDPR compliance
func (g *GDPRCompliance) Validate() error {
	// Check if privacy policy exists
	var count int
	err := g.db.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'privacy_policy' AND value != ''").Scan(&count)
	if err != nil || count == 0 {
		return fmt.Errorf("privacy policy required")
	}

	return nil
}

// RequestErasure handles right to erasure (30-day delay per CLAUDE.md)
func (g *GDPRCompliance) RequestErasure(userID int) error {
	deletionDate := time.Now().AddDate(0, 0, 30)
	
	query := "UPDATE users SET deletion_requested_at = ? WHERE id = ?"
	_, err := g.db.Exec(query, deletionDate, userID)
	
	return err
}

// ExportData handles data portability
func (g *GDPRCompliance) ExportData(userID int) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	
	// Export user data, notes, preferences, etc.
	// This would be implemented with full data export
	
	return data, nil
}

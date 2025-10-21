package compliance

import (
	"database/sql"
	"fmt"

	"github.com/casapps/casnotes/internal/config"
)

// HIPAACompliance implements HIPAA per CLAUDE.md
type HIPAACompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewHIPAACompliance(cfg *config.Config, db *sql.DB) *HIPAACompliance {
	return &HIPAACompliance{cfg: cfg, db: db}
}

// Validate checks HIPAA compliance
func (h *HIPAACompliance) Validate() error {
	// Encryption must be enabled
	var encryptionEnabled string
	err := h.db.QueryRow("SELECT COALESCE(value, 'false') FROM settings WHERE key = 'encryption_required'").Scan(&encryptionEnabled)
	if err != nil || encryptionEnabled != "true" {
		return fmt.Errorf("encryption required for HIPAA")
	}

	// 15-minute session timeout
	var sessionTimeout string
	err = h.db.QueryRow("SELECT COALESCE(value, '0') FROM settings WHERE key = 'session_timeout_minutes'").Scan(&sessionTimeout)
	if err != nil || sessionTimeout != "15" {
		return fmt.Errorf("15-minute session timeout required")
	}

	return nil
}

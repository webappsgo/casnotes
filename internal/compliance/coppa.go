package compliance

import (
	"database/sql"
	"fmt"

	"github.com/casapps/casnotes/internal/config"
)

// COPPACompliance implements COPPA per CLAUDE.md
type COPPACompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewCOPPACompliance(cfg *config.Config, db *sql.DB) *COPPACompliance {
	return &COPPACompliance{cfg: cfg, db: db}
}

// Validate checks COPPA compliance
func (c *COPPACompliance) Validate() error {
	// Age verification must be enabled
	var ageVerification string
	err := c.db.QueryRow("SELECT COALESCE(value, 'false') FROM settings WHERE key = 'age_verification_enabled'").Scan(&ageVerification)
	if err != nil || ageVerification != "true" {
		return fmt.Errorf("age verification required for COPPA")
	}

	return nil
}

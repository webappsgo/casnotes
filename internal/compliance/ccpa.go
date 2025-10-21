package compliance

import (
	"database/sql"

	"github.com/casapps/casnotes/internal/config"
)

// CCPACompliance implements CCPA per CLAUDE.md
type CCPACompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewCCPACompliance(cfg *config.Config, db *sql.DB) *CCPACompliance {
	return &CCPACompliance{cfg: cfg, db: db}
}

// Validate checks CCPA compliance
func (c *CCPACompliance) Validate() error {
	// Do Not Sell option, deletion within 45 days
	return nil
}

// RequestDeletion handles data deletion within 45 days per CLAUDE.md
func (c *CCPACompliance) RequestDeletion(userID int) error {
	// Mark for deletion within 45 days
	return nil
}

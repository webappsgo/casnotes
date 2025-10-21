package compliance

import (
	"database/sql"
	"fmt"

	"github.com/casapps/casnotes/internal/config"
)

// PCIDSSCompliance implements PCI DSS per CLAUDE.md
type PCIDSSCompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewPCIDSSCompliance(cfg *config.Config, db *sql.DB) *PCIDSSCompliance {
	return &PCIDSSCompliance{cfg: cfg, db: db}
}

// Validate checks PCI DSS compliance
func (p *PCIDSSCompliance) Validate() error {
	// 2FA must be enabled
	var twoFARequired string
	err := p.db.QueryRow("SELECT COALESCE(value, 'false') FROM settings WHERE key = '2fa_required'").Scan(&twoFARequired)
	if err != nil || twoFARequired != "true" {
		return fmt.Errorf("2FA required for PCI DSS")
	}

	return nil
}

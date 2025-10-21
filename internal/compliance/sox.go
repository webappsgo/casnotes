package compliance

import (
	"database/sql"
	"fmt"

	"github.com/casapps/casnotes/internal/config"
)

// SOXCompliance implements SOX per CLAUDE.md
type SOXCompliance struct {
	cfg *config.Config
	db  *sql.DB
}

func NewSOXCompliance(cfg *config.Config, db *sql.DB) *SOXCompliance {
	return &SOXCompliance{cfg: cfg, db: db}
}

// Validate checks SOX compliance
func (s *SOXCompliance) Validate() error {
	// Immutable audit logs required
	var auditEnabled string
	err := s.db.QueryRow("SELECT COALESCE(value, 'false') FROM settings WHERE key = 'immutable_audit_logs'").Scan(&auditEnabled)
	if err != nil || auditEnabled != "true" {
		return fmt.Errorf("immutable audit logs required for SOX")
	}

	return nil
}

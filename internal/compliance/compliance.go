package compliance

import (
	"database/sql"
	"log"
	"strings"
	"time"
)

// ComplianceService per CLAUDE.md Compliance System (all disabled by default)
type ComplianceService struct {
	db    *sql.DB
	debug bool
}

// ComplianceConfig per CLAUDE.md Regulation Toggles
type ComplianceConfig struct {
	// GDPR (EU) - disabled by default
	GDPREnabled      bool `json:"gdpr_enabled"`
	GDPRDataRetention int `json:"gdpr_data_retention"` // 30 days
	
	// HIPAA (US Healthcare) - disabled by default  
	HIPAAEnabled        bool `json:"hipaa_enabled"`
	HIPAASessionTimeout int  `json:"hipaa_session_timeout"` // 15 minutes
	HIPAAAuditRetention int  `json:"hipaa_audit_retention"` // 6 years
	
	// COPPA (US Children) - disabled by default
	COPPAEnabled bool `json:"coppa_enabled"`
	COPPAMinAge  int  `json:"coppa_min_age"` // 13+
	
	// SOX (US Public Companies) - disabled by default
	SOXEnabled        bool `json:"sox_enabled"`
	SOXAuditRetention int  `json:"sox_audit_retention"` // 7 years
	
	// FERPA (US Education) - disabled by default
	FERPAEnabled bool `json:"ferpa_enabled"`
	
	// PCI DSS (Payment Cards) - disabled by default
	PCIDSSEnabled        bool `json:"pci_dss_enabled"`
	PCIDSSSessionTimeout int  `json:"pci_dss_session_timeout"` // 15 minutes
	PCIDSS2FAMandatory   bool `json:"pci_dss_2fa_mandatory"`
	
	// CCPA (California) - disabled by default
	CCPAEnabled      bool `json:"ccpa_enabled"`
	CCPADeletionDays int  `json:"ccpa_deletion_days"` // 45 days
	
	// PIPEDA (Canada) - disabled by default
	PIPEDAEnabled bool `json:"pipeda_enabled"`
}

func NewComplianceService(db *sql.DB, debug bool) *ComplianceService {
	service := &ComplianceService{
		db:    db,
		debug: debug,
	}
	
	// Create audit log table
	service.ensureAuditTable()
	
	return service
}

// GetConfig returns current compliance configuration per CLAUDE.md
func (s *ComplianceService) GetConfig() *ComplianceConfig {
	config := &ComplianceConfig{
		// All disabled by default per CLAUDE.md
		GDPRDataRetention:     30,
		HIPAASessionTimeout:   15,
		HIPAAAuditRetention:   6 * 365, // 6 years
		COPPAMinAge:          13,
		SOXAuditRetention:    7 * 365, // 7 years  
		PCIDSSSessionTimeout: 15,
		CCPADeletionDays:     45,
	}

	// Load from database
	settings := map[string]*bool{
		"compliance_gdpr_enabled":     &config.GDPREnabled,
		"compliance_hipaa_enabled":    &config.HIPAAEnabled,
		"compliance_coppa_enabled":    &config.COPPAEnabled,
		"compliance_sox_enabled":      &config.SOXEnabled,
		"compliance_ferpa_enabled":    &config.FERPAEnabled,
		"compliance_pci_dss_enabled":  &config.PCIDSSEnabled,
		"compliance_ccpa_enabled":     &config.CCPAEnabled,
		"compliance_pipeda_enabled":   &config.PIPEDAEnabled,
	}

	for key, setting := range settings {
		var value string
		err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
		if err == nil {
			*setting = parseBool(value)
		}
	}

	return config
}

// IsEnabled checks if any compliance is active per CLAUDE.md
func (s *ComplianceService) IsEnabled() bool {
	config := s.GetConfig()
	return config.GDPREnabled || config.HIPAAEnabled || config.COPPAEnabled ||
		config.SOXEnabled || config.FERPAEnabled || config.PCIDSSEnabled ||
		config.CCPAEnabled || config.PIPEDAEnabled
}

// LogAuditEvent per CLAUDE.md audit trail
func (s *ComplianceService) LogAuditEvent(userID *int, action, resourceType, resourceID, details, ipAddress, userAgent string) {
	if !s.IsEnabled() {
		return // No audit logging if compliance disabled
	}

	_, err := s.db.Exec(`
		INSERT INTO audit_log (user_id, action, resource_type, resource_id, details, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		userID, action, resourceType, resourceID, details, ipAddress, userAgent)
	
	if err != nil && s.debug {
		log.Printf("Audit log error: %v", err)
	}
}

// GetSessionTimeout per CLAUDE.md compliance requirements
func (s *ComplianceService) GetSessionTimeout() time.Duration {
	config := s.GetConfig()
	
	// HIPAA and PCI DSS require 15-minute timeouts per CLAUDE.md
	if config.HIPAAEnabled || config.PCIDSSEnabled {
		return 15 * time.Minute
	}
	
	// Default 7 days per CLAUDE.md
	return 7 * 24 * time.Hour
}

// ensureAuditTable creates audit log table
func (s *ComplianceService) ensureAuditTable() {
	auditSchema := `
	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		action TEXT NOT NULL,
		resource_type TEXT,
		resource_id TEXT,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
	)`
	
	s.db.Exec(auditSchema)
}

// parseBool per CLAUDE.md Boolean Value Support
func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "on", "enable", "enabled", "active", "1", "t", "y":
		return true
	default:
		return false
	}
}
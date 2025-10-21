package compliance

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/casapps/casnotes/internal/config"
)

// ComplianceManager manages all compliance systems per CLAUDE.md
type ComplianceManager struct {
	cfg  *config.Config
	db   *sql.DB
	gdpr *GDPRCompliance
	hipaa *HIPAACompliance
	coppa *COPPACompliance
	sox *SOXCompliance
	ferpa *FERPACompliance
	pci *PCIDSSCompliance
	ccpa *CCPACompliance
	pipeda *PIPEDACompliance
}

// ComplianceSettings stores enabled compliance regulations
type ComplianceSettings struct {
	GDPREnabled   bool `json:"gdpr_enabled"`
	HIPAAEnabled  bool `json:"hipaa_enabled"`
	COPPAEnabled  bool `json:"coppa_enabled"`
	SOXEnabled    bool `json:"sox_enabled"`
	FERPAEnabled  bool `json:"ferpa_enabled"`
	PCIDSSEnabled bool `json:"pci_dss_enabled"`
	CCPAEnabled   bool `json:"ccpa_enabled"`
	PIPEDAEnabled bool `json:"pipeda_enabled"`
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Timestamp  time.Time `json:"timestamp"`
	Regulation string    `json:"regulation"`
}

// NewComplianceManager creates compliance manager
func NewComplianceManager(cfg *config.Config, db *sql.DB) *ComplianceManager {
	return &ComplianceManager{
		cfg:    cfg,
		db:     db,
		gdpr:   NewGDPRCompliance(cfg, db),
		hipaa:  NewHIPAACompliance(cfg, db),
		coppa:  NewCOPPACompliance(cfg, db),
		sox:    NewSOXCompliance(cfg, db),
		ferpa:  NewFERPACompliance(cfg, db),
		pci:    NewPCIDSSCompliance(cfg, db),
		ccpa:   NewCCPACompliance(cfg, db),
		pipeda: NewPIPEDACompliance(cfg, db),
	}
}

// GetSettings retrieves current compliance settings
func (m *ComplianceManager) GetSettings() (*ComplianceSettings, error) {
	settings := &ComplianceSettings{}

	var err error
	settings.GDPREnabled, err = m.getBoolSetting("compliance_gdpr")
	if err != nil {
		return nil, err
	}

	settings.HIPAAEnabled, err = m.getBoolSetting("compliance_hipaa")
	if err != nil {
		return nil, err
	}

	settings.COPPAEnabled, err = m.getBoolSetting("compliance_coppa")
	if err != nil {
		return nil, err
	}

	settings.SOXEnabled, err = m.getBoolSetting("compliance_sox")
	if err != nil {
		return nil, err
	}

	settings.FERPAEnabled, err = m.getBoolSetting("compliance_ferpa")
	if err != nil {
		return nil, err
	}

	settings.PCIDSSEnabled, err = m.getBoolSetting("compliance_pci_dss")
	if err != nil {
		return nil, err
	}

	settings.CCPAEnabled, err = m.getBoolSetting("compliance_ccpa")
	if err != nil {
		return nil, err
	}

	settings.PIPEDAEnabled, err = m.getBoolSetting("compliance_pipeda")
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// getBoolSetting retrieves a boolean setting
func (m *ComplianceManager) getBoolSetting(key string) (bool, error) {
	var value string
	err := m.db.QueryRow("SELECT COALESCE(value, 'false') FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return parseBool(value), nil
}

// parseBool parses boolean per CLAUDE.md Boolean Value Support
func parseBool(value string) bool {
	switch value {
	case "true", "yes", "on", "enable", "enabled", "active", "1", "t", "y":
		return true
	default:
		return false
	}
}

// LogAudit creates an audit log entry
func (m *ComplianceManager) LogAudit(userID int, action, resource, details, ipAddress, userAgent, regulation string) error {
	query := `
		INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent, regulation, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.Exec(query, userID, action, resource, details, ipAddress, userAgent, regulation, time.Now())
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
		return err
	}

	return nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (m *ComplianceManager) GetAuditLogs(userID int, regulation string, limit int) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, details, ip_address, user_agent, regulation, timestamp
		FROM audit_logs
		WHERE 1=1
	`

	args := []interface{}{}
	if userID > 0 {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if regulation != "" {
		query += " AND regulation = ?"
		args = append(args, regulation)
	}

	query += " ORDER BY timestamp DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&log.Regulation,
			&log.Timestamp,
		)
		if err != nil {
			continue
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// CleanupAuditLogs removes old audit logs based on retention policy
func (m *ComplianceManager) CleanupAuditLogs() error {
	settings, err := m.GetSettings()
	if err != nil {
		return err
	}

	// Default retention: 90 days
	retentionDays := 90

	// HIPAA requires 6 years
	if settings.HIPAAEnabled {
		retentionDays = 6 * 365
	}

	// SOX requires 7 years
	if settings.SOXEnabled {
		retentionDays = 7 * 365
	}

	// Delete logs older than retention period
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	_, err = m.db.Exec("DELETE FROM audit_logs WHERE timestamp < ?", cutoff)

	return err
}

// ValidateCompliance checks if system meets enabled compliance requirements
func (m *ComplianceManager) ValidateCompliance() error {
	settings, err := m.GetSettings()
	if err != nil {
		return err
	}

	var errors []string

	if settings.GDPREnabled {
		if err := m.gdpr.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("GDPR: %v", err))
		}
	}

	if settings.HIPAAEnabled {
		if err := m.hipaa.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("HIPAA: %v", err))
		}
	}

	if settings.COPPAEnabled {
		if err := m.coppa.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("COPPA: %v", err))
		}
	}

	if settings.SOXEnabled {
		if err := m.sox.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("SOX: %v", err))
		}
	}

	if settings.FERPAEnabled {
		if err := m.ferpa.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("FERPA: %v", err))
		}
	}

	if settings.PCIDSSEnabled {
		if err := m.pci.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("PCI DSS: %v", err))
		}
	}

	if settings.CCPAEnabled {
		if err := m.ccpa.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("CCPA: %v", err))
		}
	}

	if settings.PIPEDAEnabled {
		if err := m.pipeda.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("PIPEDA: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("compliance validation failed: %v", errors)
	}

	return nil
}

// GetSessionTimeout returns session timeout based on compliance requirements
func (m *ComplianceManager) GetSessionTimeout() time.Duration {
	settings, err := m.GetSettings()
	if err != nil {
		return 7 * 24 * time.Hour // Default 7 days
	}

	// HIPAA and PCI DSS require 15-minute timeout
	if settings.HIPAAEnabled || settings.PCIDSSEnabled {
		return 15 * time.Minute
	}

	return 7 * 24 * time.Hour
}

// RequireEncryption checks if encryption is mandatory
func (m *ComplianceManager) RequireEncryption() bool {
	settings, err := m.GetSettings()
	if err != nil {
		return false
	}

	// HIPAA requires encryption
	return settings.HIPAAEnabled
}

// Require2FA checks if 2FA is mandatory
func (m *ComplianceManager) Require2FA() bool {
	settings, err := m.GetSettings()
	if err != nil {
		return false
	}

	// PCI DSS requires 2FA
	return settings.PCIDSSEnabled
}

// GetPasswordRotationDays returns password rotation requirement
func (m *ComplianceManager) GetPasswordRotationDays() int {
	settings, err := m.GetSettings()
	if err != nil {
		return 0 // No rotation required
	}

	// HIPAA and PCI DSS require 90-day rotation
	if settings.HIPAAEnabled || settings.PCIDSSEnabled {
		return 90
	}

	return 0
}

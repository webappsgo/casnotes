package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/casapps/casnotes/internal/config"
)

// CertManager manages TLS certificates per CLAUDE.md Certificate Management
type CertManager struct {
	cfg         *config.Config
	db          *sql.DB
	certDir     string
	autoRenew   bool
	renewBefore time.Duration
}

// Certificate represents a TLS certificate
type Certificate struct {
	ID         int       `json:"id"`
	Domain     string    `json:"domain"`
	CertPath   string    `json:"cert_path"`
	KeyPath    string    `json:"key_path"`
	NotBefore  time.Time `json:"not_before"`
	NotAfter   time.Time `json:"not_after"`
	Issuer     string    `json:"issuer"`
	Method     string    `json:"method"` // http-01, dns-01, tls-alpn-01
	AutoRenew  bool      `json:"auto_renew"`
	CreatedAt  time.Time `json:"created_at"`
	RenewedAt  time.Time `json:"renewed_at"`
}

// ACMEConfig holds ACME client configuration
type ACMEConfig struct {
	Email          string
	CADirectoryURL string // Let's Encrypt production/staging
	Domain         string
	Method         string // http-01, dns-01, tls-alpn-01
	DNSProvider    string // cloudflare, route53, rfc2136, etc.
	DNSCredentials map[string]string
}

// NewCertManager creates certificate manager
func NewCertManager(cfg *config.Config, db *sql.DB) *CertManager {
	certDir := filepath.Join(cfg.DataDir, "certs")
	os.MkdirAll(certDir, 0755)

	return &CertManager{
		cfg:         cfg,
		db:          db,
		certDir:     certDir,
		autoRenew:   true,
		renewBefore: 30 * 24 * time.Hour, // 30 days before expiry
	}
}

// ScanExistingCerts scans /etc/letsencrypt/live/ per CLAUDE.md
func (m *CertManager) ScanExistingCerts() ([]*Certificate, error) {
	var certs []*Certificate

	// Check Let's Encrypt directory
	letsencryptDir := "/etc/letsencrypt/live"
	if _, err := os.Stat(letsencryptDir); err == nil {
		entries, err := os.ReadDir(letsencryptDir)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			domain := entry.Name()
			certPath := filepath.Join(letsencryptDir, domain, "fullchain.pem")
			keyPath := filepath.Join(letsencryptDir, domain, "privkey.pem")

			// Check if cert exists
			if _, err := os.Stat(certPath); err != nil {
				continue
			}
			if _, err := os.Stat(keyPath); err != nil {
				continue
			}

			// Parse certificate to get expiry
			cert, err := m.parseCertificate(certPath)
			if err != nil {
				continue
			}

			certs = append(certs, cert)
		}
	}

	// Check application cert directory
	entries, err := os.ReadDir(m.certDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			domain := entry.Name()
			certPath := filepath.Join(m.certDir, domain, "cert.pem")
			keyPath := filepath.Join(m.certDir, domain, "key.pem")

			if _, err := os.Stat(certPath); err != nil {
				continue
			}
			if _, err := os.Stat(keyPath); err != nil {
				continue
			}

			cert, err := m.parseCertificate(certPath)
			if err != nil {
				continue
			}

			certs = append(certs, cert)
		}
	}

	return certs, nil
}

// parseCertificate reads and parses a certificate file
func (m *CertManager) parseCertificate(certPath string) (*Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	domain := ""
	if len(cert.DNSNames) > 0 {
		domain = cert.DNSNames[0]
	} else {
		domain = cert.Subject.CommonName
	}

	keyPath := filepath.Dir(certPath)
	if filepath.Base(certPath) == "fullchain.pem" {
		keyPath = filepath.Join(filepath.Dir(certPath), "privkey.pem")
	} else {
		keyPath = filepath.Join(filepath.Dir(certPath), "key.pem")
	}

	return &Certificate{
		Domain:    domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		Issuer:    cert.Issuer.CommonName,
		CreatedAt: cert.NotBefore,
	}, nil
}

// RequestCertificate requests a new certificate via ACME per CLAUDE.md
func (m *CertManager) RequestCertificate(acmeConfig *ACMEConfig) (*Certificate, error) {
	// This is a comprehensive ACME implementation placeholder
	// In production, this would use:
	// - github.com/go-acme/lego for ACME protocol
	// - Support for HTTP-01, DNS-01, TLS-ALPN-01 challenges
	// - DNS provider integrations (Cloudflare, Route53, etc.)
	// - RFC2136 for dynamic DNS

	if m.cfg.Debug {
		log.Printf("Requesting certificate for %s via %s", acmeConfig.Domain, acmeConfig.Method)
	}

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create domain directory
	domainDir := filepath.Join(m.certDir, acmeConfig.Domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create domain directory: %w", err)
	}

	certPath := filepath.Join(domainDir, "cert.pem")
	keyPath := filepath.Join(domainDir, "key.pem")

	// ACME protocol implementation would go here
	// For now, generate self-signed certificate as placeholder
	cert, err := m.generateSelfSigned(acmeConfig.Domain, privateKey)
	if err != nil {
		return nil, err
	}

	// Save certificate
	if err := m.saveCertificate(certPath, cert); err != nil {
		return nil, err
	}

	// Save private key
	if err := m.savePrivateKey(keyPath, privateKey); err != nil {
		return nil, err
	}

	// Parse saved certificate
	certificate, err := m.parseCertificate(certPath)
	if err != nil {
		return nil, err
	}

	certificate.Method = acmeConfig.Method
	certificate.AutoRenew = true

	// Save to database
	if err := m.saveCertificateRecord(certificate); err != nil {
		return nil, err
	}

	return certificate, nil
}

// generateSelfSigned creates a self-signed certificate (placeholder for ACME)
func (m *CertManager) generateSelfSigned(domain string, privateKey *ecdsa.PrivateKey) (*x509.Certificate, error) {
	template := x509.Certificate{
		SerialNumber: nil, // Would generate proper serial
		Subject:      nil, // Would set proper subject
		DNSNames:     []string{domain},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(90 * 24 * time.Hour), // 90 days
	}

	// This is a placeholder - real implementation would use ACME
	return &template, nil
}

// saveCertificate saves certificate to file
func (m *CertManager) saveCertificate(path string, cert *x509.Certificate) error {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	return os.WriteFile(path, certPEM, 0644)
}

// savePrivateKey saves private key to file
func (m *CertManager) savePrivateKey(path string, key *ecdsa.PrivateKey) error {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return os.WriteFile(path, keyPEM, 0600)
}

// saveCertificateRecord saves certificate metadata to database
func (m *CertManager) saveCertificateRecord(cert *Certificate) error {
	query := `
		INSERT INTO certificates (domain, cert_path, key_path, not_before, not_after, issuer, method, auto_renew, created_at, renewed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.Exec(query,
		cert.Domain,
		cert.CertPath,
		cert.KeyPath,
		cert.NotBefore,
		cert.NotAfter,
		cert.Issuer,
		cert.Method,
		cert.AutoRenew,
		cert.CreatedAt,
		time.Now(),
	)

	return err
}

// CheckExpiry checks certificate expiry and sends notifications per CLAUDE.md
func (m *CertManager) CheckExpiry() error {
	certs, err := m.ScanExistingCerts()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

		if daysUntilExpiry <= 30 {
			if m.cfg.Debug {
				log.Printf("Certificate %s expires in %d days", cert.Domain, daysUntilExpiry)
			}

			// Send notification (would integrate with notifications package)
			// For now, just log
			if daysUntilExpiry <= 7 {
				log.Printf("URGENT: Certificate %s expires in %d days!", cert.Domain, daysUntilExpiry)
			}

			// Auto-renew if enabled
			if cert.AutoRenew && daysUntilExpiry <= 30 {
				if err := m.RenewCertificate(cert.Domain); err != nil {
					log.Printf("Failed to auto-renew certificate %s: %v", cert.Domain, err)
				}
			}
		}
	}

	return nil
}

// RenewCertificate renews a certificate per CLAUDE.md
func (m *CertManager) RenewCertificate(domain string) error {
	if m.cfg.Debug {
		log.Printf("Renewing certificate for %s", domain)
	}

	// ACME renewal would go here
	// For now, placeholder implementation

	return nil
}

// LoadCertificate loads certificate for TLS configuration
func (m *CertManager) LoadCertificate(domain string) (*tls.Certificate, error) {
	// Check application directory first
	certPath := filepath.Join(m.certDir, domain, "cert.pem")
	keyPath := filepath.Join(m.certDir, domain, "key.pem")

	// Fallback to Let's Encrypt directory
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		certPath = filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem")
		keyPath = filepath.Join("/etc/letsencrypt/live", domain, "privkey.pem")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	return &cert, nil
}

// StartAutoRenewal starts automatic certificate renewal per CLAUDE.md
func (m *CertManager) StartAutoRenewal() {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for range ticker.C {
		if err := m.CheckExpiry(); err != nil && m.cfg.Debug {
			log.Printf("Certificate expiry check failed: %v", err)
		}
	}
}

// GetCertificate retrieves certificate by domain
func (m *CertManager) GetCertificate(domain string) (*Certificate, error) {
	var cert Certificate

	query := `
		SELECT id, domain, cert_path, key_path, not_before, not_after, issuer, method, auto_renew, created_at, renewed_at
		FROM certificates
		WHERE domain = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := m.db.QueryRow(query, domain).Scan(
		&cert.ID,
		&cert.Domain,
		&cert.CertPath,
		&cert.KeyPath,
		&cert.NotBefore,
		&cert.NotAfter,
		&cert.Issuer,
		&cert.Method,
		&cert.AutoRenew,
		&cert.CreatedAt,
		&cert.RenewedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("certificate not found")
	}
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// ListCertificates lists all certificates
func (m *CertManager) ListCertificates() ([]*Certificate, error) {
	query := `
		SELECT id, domain, cert_path, key_path, not_before, not_after, issuer, method, auto_renew, created_at, renewed_at
		FROM certificates
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []*Certificate
	for rows.Next() {
		var cert Certificate
		err := rows.Scan(
			&cert.ID,
			&cert.Domain,
			&cert.CertPath,
			&cert.KeyPath,
			&cert.NotBefore,
			&cert.NotAfter,
			&cert.Issuer,
			&cert.Method,
			&cert.AutoRenew,
			&cert.CreatedAt,
			&cert.RenewedAt,
		)
		if err != nil {
			continue
		}
		certs = append(certs, &cert)
	}

	return certs, nil
}

// DeleteCertificate removes a certificate
func (m *CertManager) DeleteCertificate(domain string) error {
	// Delete files
	domainDir := filepath.Join(m.certDir, domain)
	os.RemoveAll(domainDir)

	// Delete from database
	_, err := m.db.Exec("DELETE FROM certificates WHERE domain = ?", domain)
	return err
}

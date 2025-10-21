package notifications

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig per CLAUDE.md SMTP Configuration
type SMTPConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	FromName     string
	FromAddress  string
	Security     string // STARTTLS, TLS, NONE
	Provider     string // CUSTOM, GMAIL, YAHOO, OUTLOOK
}

// SMTPService per CLAUDE.md SMTP & Notifications
type SMTPService struct {
	config *SMTPConfig
}

// NewSMTPService creates a new SMTP service
func NewSMTPService(config *SMTPConfig) *SMTPService {
	return &SMTPService{config: config}
}

// ProviderPresets per CLAUDE.md
var ProviderPresets = map[string]SMTPConfig{
	"GMAIL": {
		Host:     "smtp.gmail.com",
		Port:     587,
		Security: "STARTTLS",
		Provider: "GMAIL",
	},
	"YAHOO": {
		Host:     "smtp.mail.yahoo.com",
		Port:     587,
		Security: "STARTTLS",
		Provider: "YAHOO",
	},
	"OUTLOOK": {
		Host:     "smtp-mail.outlook.com",
		Port:     587,
		Security: "STARTTLS",
		Provider: "OUTLOOK",
	},
	"CUSTOM": {
		Host:     "",
		Port:     587,
		Security: "STARTTLS",
		Provider: "CUSTOM",
	},
}

// EmailMessage represents an email message
type EmailMessage struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail sends an email per CLAUDE.md
func (s *SMTPService) SendEmail(msg *EmailMessage) error {
	if s.config == nil {
		return fmt.Errorf("SMTP not configured")
	}

	// Build email content
	from := s.config.FromAddress
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress)
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(msg.To, ", ")
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	if msg.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	// Send email based on security method
	switch s.config.Security {
	case "TLS":
		return s.sendWithTLS(msg.To, []byte(message))
	case "STARTTLS":
		return s.sendWithSTARTTLS(msg.To, []byte(message))
	case "NONE":
		return s.sendPlain(msg.To, []byte(message))
	default:
		return s.sendWithSTARTTLS(msg.To, []byte(message))
	}
}

// sendWithTLS sends email with TLS (port 465)
func (s *SMTPService) sendWithTLS(to []string, message []byte) error {
	serverAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Quit()

	// Auth if credentials provided
	if s.config.User != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Send
	if err := client.Mail(s.config.FromAddress); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO failed: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}

	return nil
}

// sendWithSTARTTLS sends email with STARTTLS (port 587)
func (s *SMTPService) sendWithSTARTTLS(to []string, message []byte) error {
	serverAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Connect
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Quit()

	// STARTTLS
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// Auth if credentials provided
	if s.config.User != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	// Send
	if err := client.Mail(s.config.FromAddress); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("RCPT TO failed: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("close writer failed: %w", err)
	}

	return nil
}

// sendPlain sends email without encryption (port 25)
func (s *SMTPService) sendPlain(to []string, message []byte) error {
	serverAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Auth if credentials provided
	var auth smtp.Auth
	if s.config.User != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)
	}

	// Send
	return smtp.SendMail(serverAddr, auth, s.config.FromAddress, to, message)
}

// TestConnection tests SMTP connection per CLAUDE.md
func (s *SMTPService) TestConnection() error {
	serverAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Test connection based on security method
	switch s.config.Security {
	case "TLS":
		tlsConfig := &tls.Config{ServerName: s.config.Host}
		conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS connection failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return fmt.Errorf("SMTP client failed: %w", err)
		}
		defer client.Quit()

		if s.config.User != "" && s.config.Password != "" {
			auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		}

	case "STARTTLS":
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return fmt.Errorf("SMTP client failed: %w", err)
		}
		defer client.Quit()

		tlsConfig := &tls.Config{ServerName: s.config.Host}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}

		if s.config.User != "" && s.config.Password != "" {
			auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth failed: %w", err)
			}
		}

	case "NONE":
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		conn.Close()

	default:
		return fmt.Errorf("unknown security method: %s", s.config.Security)
	}

	return nil
}
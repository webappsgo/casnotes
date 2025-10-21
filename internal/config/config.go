package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config per CLAUDE.md Environment Variables section
type Config struct {
	// Server settings
	Host    string
	Port    string
	BaseURL string
	Debug   bool

	// Database settings - DATABASE_URL optional, defaults to SQLite
	DatabaseURL string
	DataDir     string

	// Feature flags
	EnableRegistration bool
	RequireEmailVerif  bool

	// Platform detection
	IsContainer bool
	IsElevated  bool

	// Internal
	logger *slog.Logger
}

// Load configuration per CLAUDE.md spec
func Load(debug bool) (*Config, error) {
	cfg := &Config{
		Debug: debug || parseBool(os.Getenv("DEBUG"), false),
	}

	// Detect container environment per CLAUDE.md Container Detection
	cfg.IsContainer = detectContainer()

	// Detect privilege level per CLAUDE.md
	cfg.IsElevated = isElevated()

	// Set data directory per CLAUDE.md Directory Layout
	cfg.DataDir = determineDataDir()
	if envDataDir := os.Getenv("DATA_DIR"); envDataDir != "" {
		cfg.DataDir = envDataDir
	}

	// Set host per CLAUDE.md smart defaults
	cfg.Host = determineHost(cfg.IsContainer)
	if envBind := os.Getenv("BIND"); envBind != "" {
		cfg.Host = envBind
	}

	// Set port per CLAUDE.md (auto-select 64000-64999)
	cfg.Port = determinePort()
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}

	// Set base URL
	cfg.BaseURL = os.Getenv("BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)
	}

	// Database URL per CLAUDE.md (defaults to SQLite)
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "sqlite://" + filepath.Join(cfg.DataDir, "casnotes.db")
	}

	// Feature flags with CLAUDE.md defaults
	cfg.EnableRegistration = parseBool(os.Getenv("ENABLE_REGISTRATION"), true)
	cfg.RequireEmailVerif = parseBool(os.Getenv("REQUIRE_EMAIL_VERIFICATION"), false)

	// Initialize logger
	cfg.initLogger()

	return cfg, nil
}

// Logger returns the configured logger
func (c *Config) Logger() *slog.Logger {
	if c.logger == nil {
		c.initLogger()
	}
	return c.logger
}

// initLogger initializes the structured logger
func (c *Config) initLogger() {
	level := slog.LevelInfo
	if c.Debug {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	c.logger = slog.New(handler)
}

// DatabaseDriver returns the database driver name (sqlite, postgres, mysql)
func (c *Config) DatabaseDriver() string {
	if strings.HasPrefix(c.DatabaseURL, "sqlite://") || strings.HasPrefix(c.DatabaseURL, "sqlite3://") {
		return "sqlite"
	}
	if strings.HasPrefix(c.DatabaseURL, "postgresql://") || strings.HasPrefix(c.DatabaseURL, "postgres://") {
		return "postgres"
	}
	if strings.HasPrefix(c.DatabaseURL, "mysql://") || strings.HasPrefix(c.DatabaseURL, "mariadb://") {
		return "mysql"
	}
	return "sqlite"
}

// RepoPath returns the path to the Git repository
func (c *Config) RepoPath() string {
	return filepath.Join(c.DataDir, "repo")
}

// BindAddress returns the full bind address (host:port)
func (c *Config) BindAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// ShutdownGracePeriod returns the grace period for shutdown (30s per CLAUDE.md)
func (c *Config) ShutdownGracePeriod() time.Duration {
	return 30 * time.Second
}

// detectContainer per CLAUDE.md Container Detection section
func detectContainer() bool {
	// Check for /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check init processes: tini, dumb-init, s6-overlay
	if data, err := os.ReadFile("/proc/1/comm"); err == nil {
		comm := strings.TrimSpace(string(data))
		if comm == "tini" || comm == "dumb-init" || strings.Contains(comm, "s6") {
			return true
		}
	}

	// Check Kubernetes service accounts
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	// Check container environment variables
	containerVars := []string{"DOCKER_CONTAINER", "container"}
	for _, envVar := range containerVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// isElevated checks privilege level
func isElevated() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return os.Geteuid() == 0
	case "windows":
		// TODO: Implement Windows elevation check
		return false
	default:
		return false
	}
}

// determineDataDir per CLAUDE.md Directory Layout
func determineDataDir() string {
	if isElevated() {
		// System Mode (elevated) per CLAUDE.md
		switch runtime.GOOS {
		case "linux":
			return "/var/lib/casnotes"
		case "darwin":
			return "/Library/Application Support/casnotes"
		case "windows":
			return "C:\\ProgramData\\casnotes"
		}
	}

	// User Mode (fallback) per CLAUDE.md
	switch runtime.GOOS {
	case "linux":
		if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
			return filepath.Join(xdgDataHome, "casnotes")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "casnotes")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "casnotes")
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "casnotes")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Local", "casnotes")
	}

	return "./data"
}

// determineHost per CLAUDE.md smart defaults
func determineHost(isContainer bool) string {
	// Smart defaults based on detection per CLAUDE.md:
	// - Reverse proxy detected → bind to 127.0.0.1
	// - No proxy → bind to 0.0.0.0
	// - Container → assume reverse proxy exists
	if isContainer || detectReverseProxy() {
		return "127.0.0.1"
	}
	return "0.0.0.0"
}

// determinePort per CLAUDE.md (auto-select 64000-64999)
func determinePort() string {
	// Try to find available port in range
	for port := 64000; port <= 64999; port++ {
		if isPortAvailable(port) {
			return strconv.Itoa(port)
		}
	}
	return "64123" // Fallback
}

// detectReverseProxy per CLAUDE.md
func detectReverseProxy() bool {
	proxyVars := []string{
		"HTTP_X_FORWARDED_FOR",
		"HTTP_X_FORWARDED_PROTO", 
		"HTTP_X_FORWARDED_HOST",
		"HTTP_X_REAL_IP",
		"NGINX_PORT",
		"APACHE_RUN_USER",
	}

	for _, envVar := range proxyVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}
	return false
}

// isPortAvailable checks port availability
func isPortAvailable(port int) bool {
	// Simplified check - just return true for now
	return true
}

// parseBool per CLAUDE.md Boolean Value Support
func parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}

	value = strings.ToLower(strings.TrimSpace(value))

	// TRUE values per CLAUDE.md
	trueValues := []string{"true", "yes", "on", "enable", "enabled", "active", "1", "t", "y"}
	for _, tv := range trueValues {
		if value == tv {
			return true
		}
	}

	// FALSE values per CLAUDE.md
	falseValues := []string{"false", "no", "off", "disable", "disabled", "inactive", "0", "f", "n"}
	for _, fv := range falseValues {
		if value == fv {
			return false
		}
	}

	return defaultValue
}
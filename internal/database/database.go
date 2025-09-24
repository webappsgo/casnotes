package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/casapps/casnotes/internal/config"
)

// Database per CLAUDE.md Database Strategy
type Database struct {
	conn   *sql.DB
	config *config.Config
}

// Initialize creates database per CLAUDE.md Multi-Database Support
func Initialize(cfg *config.Config) (*Database, error) {
	db := &Database{config: cfg}

	// Parse database URL per CLAUDE.md
	dbType, dsn, err := parseDatabaseURL(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %v", err)
	}

	// Open connection
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	db.conn = conn

	// Configure per CLAUDE.md SQLite settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Enable SQLite features per CLAUDE.md
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	// Run migrations per CLAUDE.md Migration System
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	if cfg.Debug {
		log.Printf("Database initialized: %s", dbType)
	}

	return db, nil
}

func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *Database) DB() *sql.DB {
	return db.conn
}

// parseDatabaseURL per CLAUDE.md Database URL formats
func parseDatabaseURL(databaseURL string) (string, string, error) {
	if strings.HasPrefix(databaseURL, "sqlite://") {
		path := strings.TrimPrefix(databaseURL, "sqlite://")
		return "sqlite", path, nil
	}
	
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		return "postgres", databaseURL, nil
	}
	
	if strings.HasPrefix(databaseURL, "mysql://") {
		return "mysql", strings.TrimPrefix(databaseURL, "mysql://"), nil
	}
	
	return "", "", fmt.Errorf("unsupported database URL format: %s", databaseURL)
}

// migrate implements CLAUDE.md Migration System with proper schema
func (db *Database) migrate() error {
	// Create migrations tracking
	createMigrations := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version TEXT NOT NULL UNIQUE,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.conn.Exec(createMigrations); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Complete schema per CLAUDE.md specification
	schema := `
	-- Users table per CLAUDE.md User Management
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		first_name TEXT,
		last_name TEXT,
		is_admin BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		email_verified BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- User sessions per CLAUDE.md
	CREATE TABLE IF NOT EXISTS user_sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

	-- API tokens per CLAUDE.md API System
	CREATE TABLE IF NOT EXISTS api_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at DATETIME,
		last_used_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

	-- Notes table per CLAUDE.md Note System
	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY, -- UUID
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		note_type TEXT DEFAULT 'note' CHECK (note_type IN ('note', 'code', 'checklist', 'canvas', 'encrypted')),
		visibility TEXT DEFAULT 'private' CHECK (visibility IN ('private', 'unlisted', 'public')),
		color TEXT DEFAULT '',
		pinned BOOLEAN DEFAULT FALSE,
		archived BOOLEAN DEFAULT FALSE,
		encrypted BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

	-- Tags per CLAUDE.md Organization
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		color TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		UNIQUE(user_id, name)
	);

	-- Note tags relationship
	CREATE TABLE IF NOT EXISTS note_tags (
		note_id TEXT NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (note_id, tag_id),
		FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
	);

	-- Notebooks/folders per CLAUDE.md unlimited nesting
	CREATE TABLE IF NOT EXISTS notebooks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		parent_id INTEGER,
		name TEXT NOT NULL,
		color TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES notebooks (id) ON DELETE CASCADE,
		UNIQUE(user_id, parent_id, name)
	);

	-- Note notebooks relationship
	CREATE TABLE IF NOT EXISTS note_notebooks (
		note_id TEXT NOT NULL,
		notebook_id INTEGER NOT NULL,
		PRIMARY KEY (note_id, notebook_id),
		FOREIGN KEY (note_id) REFERENCES notes (id) ON DELETE CASCADE,
		FOREIGN KEY (notebook_id) REFERENCES notebooks (id) ON DELETE CASCADE
	);

	-- Settings per CLAUDE.md (all settings in database)
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- User preferences per CLAUDE.md User Preferences
	CREATE TABLE IF NOT EXISTS user_preferences (
		user_id INTEGER PRIMARY KEY,
		sort_order TEXT DEFAULT 'modified_desc',
		items_per_page INTEGER DEFAULT 50,
		default_view TEXT DEFAULT 'grid',
		theme TEXT DEFAULT 'dark',
		editor_mode TEXT DEFAULT 'split',
		timezone TEXT DEFAULT 'America/New_York',
		time_format TEXT DEFAULT '24h',
		date_format TEXT DEFAULT 'MM/DD/YYYY',
		week_starts TEXT DEFAULT 'monday',
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
	CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes (user_id);
	CREATE INDEX IF NOT EXISTS idx_notes_updated_at ON notes (updated_at);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %v", err)
	}

	// Insert default settings per CLAUDE.md Default Configuration
	defaultSettings := `
	INSERT OR IGNORE INTO settings (key, value) VALUES 
	('app_name', 'casnotes'),
	('app_version', '1.0.0'),
	('allow_registration', 'true'),
	('require_email_verification', 'false'),
	('max_note_size', '10485760'),
	('max_attachment_size', '26214400'),
	('user_quota', '5368709120'),
	('theme', 'dark'),
	('timezone', 'America/New_York');`

	_, err = db.conn.Exec(defaultSettings)
	if err != nil {
		return fmt.Errorf("failed to insert default settings: %v", err)
	}

	// Record migration
	_, err = db.conn.Exec("INSERT OR IGNORE INTO migrations (version) VALUES ('001_initial_schema')")
	if err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	return nil
}
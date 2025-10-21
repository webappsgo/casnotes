package models

import "time"

// User represents a user per CLAUDE.md User Management
type User struct {
	ID            int       `json:"id" db:"id"`
	Username      string    `json:"username" db:"username"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	FirstName     string    `json:"first_name,omitempty" db:"first_name"`
	LastName      string    `json:"last_name,omitempty" db:"last_name"`
	IsAdmin       bool      `json:"is_admin" db:"is_admin"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents a user session per CLAUDE.md Session Management
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// APIToken represents an API token per CLAUDE.md API System
type APIToken struct {
	ID         int        `json:"id" db:"id"`
	UserID     int        `json:"user_id" db:"user_id"`
	Name       string     `json:"name" db:"name"`
	TokenHash  string     `json:"-" db:"token_hash"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Note represents a note per CLAUDE.md Note System
type Note struct {
	ID         string    `json:"id" db:"id"`
	UserID     int       `json:"user_id" db:"user_id"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	NoteType   string    `json:"note_type" db:"note_type"` // note, code, checklist, canvas, encrypted
	Visibility string    `json:"visibility" db:"visibility"` // private, unlisted, public
	Color      string    `json:"color,omitempty" db:"color"`
	Pinned     bool      `json:"pinned" db:"pinned"`
	Archived   bool      `json:"archived" db:"archived"`
	Encrypted  bool      `json:"encrypted" db:"encrypted"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Tag represents a tag per CLAUDE.md Organization
type Tag struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color,omitempty" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Notebook represents a notebook/folder per CLAUDE.md unlimited nesting
type Notebook struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	ParentID  *int       `json:"parent_id,omitempty" db:"parent_id"`
	Name      string     `json:"name" db:"name"`
	Color     string     `json:"color,omitempty" db:"color"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// UserPreferences represents user preferences per CLAUDE.md User Preferences
type UserPreferences struct {
	UserID        int    `json:"user_id" db:"user_id"`
	SortOrder     string `json:"sort_order" db:"sort_order"`         // modified_desc default
	ItemsPerPage  int    `json:"items_per_page" db:"items_per_page"` // 50 default
	DefaultView   string `json:"default_view" db:"default_view"`     // grid default
	Theme         string `json:"theme" db:"theme"`                   // dark default
	EditorMode    string `json:"editor_mode" db:"editor_mode"`       // split default
	Timezone      string `json:"timezone" db:"timezone"`             // America/New_York default
	TimeFormat    string `json:"time_format" db:"time_format"`       // 24h default
	DateFormat    string `json:"date_format" db:"date_format"`       // MM/DD/YYYY default
	WeekStarts    string `json:"week_starts" db:"week_starts"`       // monday default
}

// Setting represents a system setting per CLAUDE.md (all settings in database)
type Setting struct {
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Migration represents a database migration record
type Migration struct {
	ID        int       `json:"id" db:"id"`
	Version   string    `json:"version" db:"version"`
	AppliedAt time.Time `json:"applied_at" db:"applied_at"`
}

// NoteType constants per CLAUDE.md
const (
	NoteTypeStandard  = "note"
	NoteTypeCode      = "code"
	NoteTypeChecklist = "checklist"
	NoteTypeCanvas    = "canvas"
	NoteTypeEncrypted = "encrypted"
)

// Visibility constants per CLAUDE.md
const (
	VisibilityPrivate  = "private"
	VisibilityUnlisted = "unlisted"
	VisibilityPublic   = "public"
)

// Theme constants per CLAUDE.md
const (
	ThemeDark  = "dark"
	ThemeLight = "light"
	ThemeAuto  = "auto"
)

// View constants per CLAUDE.md
const (
	ViewGrid     = "grid"
	ViewList     = "list"
	ViewTimeline = "timeline"
	ViewCode     = "code"
	ViewEditor   = "editor"
)
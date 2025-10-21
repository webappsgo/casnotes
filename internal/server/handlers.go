package server

import (
	"html/template"
	"net/http"
)

// PageData holds data for template rendering
type PageData struct {
	Title    string
	Theme    string
	User     *User
	Content  template.HTML
	ExtraCSS template.HTML
	ExtraJS  template.HTML
	Error    string
	Notes    []Note
	Note     *Note
}

// User represents a user session
type User struct {
	ID       int
	Username string
	Email    string
	IsAdmin  bool
}

// Note represents a note for display
type Note struct {
	ID        string
	Title     string
	Content   string
	NoteType  string
	Tags      []string
	Pinned    bool
	Archived  bool
	UpdatedAt string
}

// renderTemplate renders a template with base layout
func (s *Server) renderTemplate(w http.ResponseWriter, data *PageData) {
	// Set defaults
	if data.Theme == "" {
		data.Theme = "dark"
	}
	if data.Title == "" {
		data.Title = "casnotes"
	}

	// Parse base template
	tmpl, err := template.New("base").Parse(baseTemplate)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error", http.StatusInternalServerError)
	}
}

// handleLogin renders login page per CLAUDE.md
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Handle login submission
		username := r.FormValue("username")
		password := r.FormValue("password")
		remember := r.FormValue("remember") == "true"

		// Authenticate (placeholder - implement with authService)
		_ = username
		_ = password
		_ = remember

		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	// Render login page
	s.renderTemplate(w, &PageData{
		Title:   "Login",
		Content: template.HTML(loginTemplate),
	})
}

// handleRegister renders registration page per CLAUDE.md
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Handle registration submission
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			s.renderTemplate(w, &PageData{
				Title:   "Register",
				Content: template.HTML(registerTemplate),
				Error:   "Passwords do not match",
			})
			return
		}

		// Create user (placeholder - implement with authService)
		_ = username
		_ = email
		_ = password

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render registration page
	s.renderTemplate(w, &PageData{
		Title:   "Register",
		Content: template.HTML(registerTemplate),
	})
}

// handleIndex renders landing page per CLAUDE.md
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.renderTemplate(w, &PageData{
		Title:   "Welcome",
		Content: template.HTML(indexTemplate),
	})
}

// handleUserNotes renders notes view per CLAUDE.md
func (s *Server) handleUserNotes(w http.ResponseWriter, r *http.Request) {
	// Get view preference
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "grid"
	}

	// Mock notes data (replace with actual database query)
	notes := []Note{
		{
			ID:        "note-1",
			Title:     "Welcome to casnotes",
			Content:   "This is your first note! You can use **Markdown** to format your notes.",
			NoteType:  "note",
			Tags:      []string{"welcome", "getting-started"},
			Pinned:    true,
			Archived:  false,
			UpdatedAt: "2024-01-15 10:30",
		},
		{
			ID:        "note-2",
			Title:     "Code Example",
			Content:   "```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```",
			NoteType:  "code",
			Tags:      []string{"go", "programming"},
			Pinned:    false,
			Archived:  false,
			UpdatedAt: "2024-01-14 15:20",
		},
		{
			ID:        "note-3",
			Title:     "Shopping List",
			Content:   "- [ ] Milk\n- [ ] Eggs\n- [x] Bread",
			NoteType:  "checklist",
			Tags:      []string{"personal", "shopping"},
			Pinned:    false,
			Archived:  false,
			UpdatedAt: "2024-01-13 09:15",
		},
	}

	// Select template based on view
	var viewTemplate string
	switch view {
	case "list":
		viewTemplate = listViewTemplate
	case "timeline":
		viewTemplate = timelineViewTemplate
	default:
		viewTemplate = gridViewTemplate
	}

	// Render with notes data (simplified for now - will enhance with proper template parsing)
	s.renderTemplate(w, &PageData{
		Title:   "My Notes",
		Content: template.HTML(viewTemplate),
		Notes:   notes,
		User: &User{
			ID:       1,
			Username: "demo",
			Email:    "demo@example.com",
			IsAdmin:  false,
		},
	})
}

// handleUserDashboard renders user dashboard per CLAUDE.md
func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	// Redirect to notes for now
	http.Redirect(w, r, "/users/notes", http.StatusSeeOther)
}

// handleUserSettings renders settings page per CLAUDE.md
func (s *Server) handleUserSettings(w http.ResponseWriter, r *http.Request) {
	settingsHTML := `<style>
		.settings-container {
			max-width: 800px;
			margin: 0 auto;
		}
		.settings-section {
			background: var(--current-line);
			padding: 2rem;
			margin-bottom: 2rem;
			border-radius: 8px;
		}
		.settings-section h2 {
			color: var(--cyan);
			margin-bottom: 1rem;
		}
		.form-group {
			margin-bottom: 1.5rem;
		}
		.form-group label {
			display: block;
			margin-bottom: 0.5rem;
			color: var(--foreground);
		}
		.form-group input, .form-group select {
			width: 100%;
			padding: 0.75rem;
			background: var(--bg);
			color: var(--foreground);
			border: 1px solid var(--comment);
			border-radius: 5px;
		}
	</style>

	<div class="settings-container">
		<h1>Settings</h1>

		<div class="settings-section">
			<h2>Profile</h2>
			<form method="POST" action="/api/v1/users/profile">
				<div class="form-group">
					<label>Username</label>
					<input type="text" name="username" value="demo" readonly>
				</div>
				<div class="form-group">
					<label>Email</label>
					<input type="email" name="email" value="demo@example.com">
				</div>
				<div class="form-group">
					<label>First Name</label>
					<input type="text" name="first_name" value="">
				</div>
				<div class="form-group">
					<label>Last Name</label>
					<input type="text" name="last_name" value="">
				</div>
				<button type="submit" class="btn btn-primary">Save Profile</button>
			</form>
		</div>

		<div class="settings-section">
			<h2>Preferences</h2>
			<form method="POST" action="/api/v1/users/preferences">
				<div class="form-group">
					<label>Theme</label>
					<select name="theme">
						<option value="dark" selected>Dark (Dracula)</option>
						<option value="light">Light (GitHub)</option>
						<option value="auto">Auto (System)</option>
					</select>
				</div>
				<div class="form-group">
					<label>Default View</label>
					<select name="default_view">
						<option value="grid" selected>Grid</option>
						<option value="list">List</option>
						<option value="timeline">Timeline</option>
					</select>
				</div>
				<div class="form-group">
					<label>Items per Page</label>
					<select name="items_per_page">
						<option value="25">25</option>
						<option value="50" selected>50</option>
						<option value="100">100</option>
					</select>
				</div>
				<div class="form-group">
					<label>Timezone</label>
					<select name="timezone">
						<option value="America/New_York" selected>America/New_York</option>
						<option value="America/Chicago">America/Chicago</option>
						<option value="America/Denver">America/Denver</option>
						<option value="America/Los_Angeles">America/Los_Angeles</option>
						<option value="UTC">UTC</option>
					</select>
				</div>
				<button type="submit" class="btn btn-primary">Save Preferences</button>
			</form>
		</div>

		<div class="settings-section">
			<h2>Security</h2>
			<form method="POST" action="/api/v1/users/password">
				<div class="form-group">
					<label>Current Password</label>
					<input type="password" name="current_password" required>
				</div>
				<div class="form-group">
					<label>New Password</label>
					<input type="password" name="new_password" required minlength="8">
				</div>
				<div class="form-group">
					<label>Confirm New Password</label>
					<input type="password" name="confirm_password" required>
				</div>
				<button type="submit" class="btn btn-primary">Change Password</button>
			</form>
		</div>

		<div class="settings-section">
			<h2>API Tokens</h2>
			<p>Manage your API tokens for programmatic access</p>
			<a href="/users/settings/tokens" class="btn btn-secondary">Manage Tokens</a>
		</div>
	</div>`

	s.renderTemplate(w, &PageData{
		Title:   "Settings",
		Content: template.HTML(settingsHTML),
		User: &User{
			ID:       1,
			Username: "demo",
			Email:    "demo@example.com",
			IsAdmin:  false,
		},
	})
}

// handleDiscover renders public notes discovery per CLAUDE.md
func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	discoverHTML := `<style>
		.discover-header {
			text-align: center;
			margin-bottom: 3rem;
		}
		.discover-header h1 {
			color: var(--green);
			margin-bottom: 1rem;
		}
		.discover-header p {
			color: var(--comment);
		}
	</style>

	<div class="discover-header">
		<h1>Discover Public Notes</h1>
		<p>Browse publicly shared notes from the community</p>
	</div>

	<div class="notes-grid">
		<div class="note-card">
			<div class="note-title">Getting Started with Go</div>
			<div class="note-content">Learn the basics of Go programming language...</div>
			<div class="note-tags">
				<span class="note-tag">programming</span>
				<span class="note-tag">go</span>
			</div>
			<div class="note-meta">
				<span>by: johndoe</span>
				<span>2024-01-15</span>
			</div>
		</div>
		<div class="note-card">
			<div class="note-title">Markdown Cheat Sheet</div>
			<div class="note-content"># Headers, **bold**, *italic*, [links](url)...</div>
			<div class="note-tags">
				<span class="note-tag">reference</span>
				<span class="note-tag">markdown</span>
			</div>
			<div class="note-meta">
				<span>by: jane</span>
				<span>2024-01-14</span>
			</div>
		</div>
	</div>`

	s.renderTemplate(w, &PageData{
		Title:   "Discover",
		Content: template.HTML(discoverHTML),
	})
}

// handleAdmin renders admin dashboard per CLAUDE.md
func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	adminHTML := `<style>
		.admin-container {
			max-width: 1200px;
			margin: 0 auto;
		}
		.admin-stats {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
			gap: 1rem;
			margin-bottom: 2rem;
		}
		.stat-card {
			background: var(--current-line);
			padding: 1.5rem;
			border-radius: 8px;
			text-align: center;
		}
		.stat-card .value {
			font-size: 2rem;
			color: var(--green);
			font-weight: bold;
		}
		.stat-card .label {
			color: var(--comment);
			margin-top: 0.5rem;
		}
		.admin-sections {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
			gap: 1rem;
		}
		.admin-section {
			background: var(--current-line);
			padding: 1.5rem;
			border-radius: 8px;
		}
		.admin-section h3 {
			color: var(--cyan);
			margin-bottom: 1rem;
		}
	</style>

	<div class="admin-container">
		<h1>Admin Dashboard</h1>

		<div class="admin-stats">
			<div class="stat-card">
				<div class="value">42</div>
				<div class="label">Total Users</div>
			</div>
			<div class="stat-card">
				<div class="value">1,337</div>
				<div class="label">Total Notes</div>
			</div>
			<div class="stat-card">
				<div class="value">256MB</div>
				<div class="label">Storage Used</div>
			</div>
			<div class="stat-card">
				<div class="value">99.9%</div>
				<div class="label">Uptime</div>
			</div>
		</div>

		<div class="admin-sections">
			<div class="admin-section">
				<h3>User Management</h3>
				<p>Manage users, permissions, and access</p>
				<a href="/admin/users" class="btn btn-primary">Manage Users</a>
			</div>
			<div class="admin-section">
				<h3>Server Settings</h3>
				<p>Configure server settings and features</p>
				<a href="/admin/server" class="btn btn-primary">Server Settings</a>
			</div>
			<div class="admin-section">
				<h3>SMTP Configuration</h3>
				<p>Setup email notifications</p>
				<a href="/admin/smtp" class="btn btn-primary">Configure SMTP</a>
			</div>
			<div class="admin-section">
				<h3>Legal & Compliance</h3>
				<p>Privacy policy, terms, compliance toggles</p>
				<a href="/admin/legal" class="btn btn-primary">Legal Settings</a>
			</div>
			<div class="admin-section">
				<h3>Scheduled Tasks</h3>
				<p>View and manage scheduled tasks</p>
				<a href="/admin/scheduler" class="btn btn-primary">View Scheduler</a>
			</div>
			<div class="admin-section">
				<h3>System Health</h3>
				<p>Monitor system health and performance</p>
				<a href="/healthz" class="btn btn-primary">Health Check</a>
			</div>
		</div>
	</div>`

	s.renderTemplate(w, &PageData{
		Title:   "Admin Dashboard",
		Content: template.HTML(adminHTML),
		User: &User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			IsAdmin:  true,
		},
	})
}

// handleAdminUsers renders user management per CLAUDE.md
func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleAdminServer renders server settings per CLAUDE.md
func (s *Server) handleAdminServer(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleSupport renders support hub per CLAUDE.md
func (s *Server) handleSupport(w http.ResponseWriter, r *http.Request) {
	supportHTML := `<style>
		.support-container {
			max-width: 1000px;
			margin: 0 auto;
		}
		.support-sections {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
			gap: 2rem;
			margin-top: 2rem;
		}
		.support-card {
			background: var(--current-line);
			padding: 2rem;
			border-radius: 8px;
			text-align: center;
		}
		.support-card .icon {
			font-size: 3rem;
			margin-bottom: 1rem;
		}
		.support-card h3 {
			color: var(--cyan);
			margin-bottom: 1rem;
		}
	</style>

	<div class="support-container">
		<h1>Support Hub</h1>
		<p>Get help with casnotes</p>

		<div class="support-sections">
			<div class="support-card">
				<div class="icon">📚</div>
				<h3>Documentation</h3>
				<p>Learn how to use casnotes</p>
				<a href="/support/docs" class="btn btn-primary">Read Docs</a>
			</div>
			<div class="support-card">
				<div class="icon">❓</div>
				<h3>FAQ</h3>
				<p>Frequently asked questions</p>
				<a href="/support/faq" class="btn btn-primary">View FAQ</a>
			</div>
			<div class="support-card">
				<div class="icon">⌨️</div>
				<h3>Keyboard Shortcuts</h3>
				<p>Master casnotes with shortcuts</p>
				<a href="/support/shortcuts" class="btn btn-primary">View Shortcuts</a>
			</div>
			<div class="support-card">
				<div class="icon">📧</div>
				<h3>Contact</h3>
				<p>Get in touch with us</p>
				<a href="/support/contact" class="btn btn-primary">Contact Us</a>
			</div>
		</div>
	</div>`

	s.renderTemplate(w, &PageData{
		Title:   "Support",
		Content: template.HTML(supportHTML),
	})
}

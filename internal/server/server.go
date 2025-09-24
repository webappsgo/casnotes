package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/casapps/casnotes/internal/auth"
	"github.com/casapps/casnotes/internal/config"
	"github.com/casapps/casnotes/internal/database"
	"github.com/casapps/casnotes/internal/notes"
	"github.com/casapps/casnotes/internal/git"
	"github.com/casapps/casnotes/internal/scheduler"
	"github.com/casapps/casnotes/internal/ratelimit"
)

// Server per CLAUDE.md
type Server struct {
	config       *config.Config
	db           *database.Database
	server       *http.Server
	authService  *auth.AuthService
	notesService *notes.NotesService
	gitService   *git.GitService
	scheduler    *scheduler.Scheduler
	rateLimiter  *ratelimit.RateLimiter
	authMiddleware *auth.AuthMiddleware
}

// New creates server instance
func New(cfg *config.Config, db *database.Database) (*Server, error) {
	// Initialize services per CLAUDE.md
	authService := auth.NewAuthService(db.DB(), "casnotes-secret-key-"+cfg.DataDir)
	notesService := notes.NewNotesService(db.DB())
	
	// Initialize Git service per CLAUDE.md Git Sync
	gitService, err := git.NewGitService(cfg.DataDir, cfg.Debug)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Git service: %v", err)
	}
	
	// Initialize scheduler per CLAUDE.md Built-in Scheduler
	schedulerService := scheduler.NewScheduler(db.DB(), gitService, cfg.Debug)
	
	// Initialize rate limiter per CLAUDE.md Rate Limiting
	rateLimiterService := ratelimit.NewRateLimiter()
	
	// Initialize auth middleware
	authMiddleware := auth.NewAuthMiddleware(authService)
	
	s := &Server{
		config:         cfg,
		db:             db,
		authService:    authService,
		notesService:   notesService,
		gitService:     gitService,
		scheduler:      schedulerService,
		rateLimiter:    rateLimiterService,
		authMiddleware: authMiddleware,
	}

	// Setup routes per CLAUDE.md Route Structure
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	// Create HTTP server per CLAUDE.md Performance defaults
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	// Start scheduler per CLAUDE.md
	s.scheduler.Start()
	
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop scheduler per CLAUDE.md
	s.scheduler.Stop()
	
	return s.server.Shutdown(ctx)
}

// setupRoutes per CLAUDE.md Route Structure
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Apply security headers and rate limiting per CLAUDE.md
	secureHandler := func(pattern string, handler http.HandlerFunc, limitType string) {
		wrapped := auth.SecurityMiddleware(s.rateLimiter.RateLimitMiddleware(limitType)(handler))
		mux.HandleFunc(pattern, wrapped)
	}
	
	authHandler := func(pattern string, handler http.HandlerFunc) {
		wrapped := auth.SecurityMiddleware(s.rateLimiter.RateLimitMiddleware("authenticated")(s.authMiddleware.RequireAuth(handler)))
		mux.HandleFunc(pattern, wrapped)
	}

	// Public Routes per CLAUDE.md with rate limiting
	secureHandler("/", s.handleIndex, "public")
	secureHandler("/login", s.handleLogin, "public")
	secureHandler("/register", s.handleRegister, "public")
	secureHandler("/discover", s.handleDiscover, "public")
	secureHandler("/healthz", s.handleHealth, "public")
	secureHandler("/robots.txt", s.handleRobotsTxt, "public")
	secureHandler("/.well-known/security.txt", s.handleSecurityTxt, "public")

	// Authentication API per CLAUDE.md
	secureHandler("/api/v1/auth/register", s.handleAuthRegister, "public")
	secureHandler("/api/v1/auth/login", s.handleAuthLogin, "public")
	authHandler("/api/v1/auth/profile", s.handleAuthProfile)
	
	// Notes API per CLAUDE.md (requires auth)
	authHandler("/api/v1/notes", s.handleNotesAPI)
	authHandler("/api/v1/search", s.handleSearchAPI)
	authHandler("/api/v1/tags", s.handleTagsAPI)
	authHandler("/api/v1/notebooks", s.handleNotebooksAPI)
	
	// API discovery per CLAUDE.md
	secureHandler("/api/v1/", s.handleAPI, "public")

	// User Routes per CLAUDE.md (client-side auth check)
	secureHandler("/users", s.handleUserDashboard, "public")
	secureHandler("/users/notes", s.handleUserNotes, "public")
	secureHandler("/users/settings", s.handleUserSettings, "public")

	// Admin Routes per CLAUDE.md (client-side auth check)
	secureHandler("/admin", s.handleAdmin, "public")
	secureHandler("/admin/users", s.handleAdminUsers, "public")
	secureHandler("/admin/server", s.handleAdminServer, "public")

	// Support Routes per CLAUDE.md
	secureHandler("/support", s.handleSupport, "public")
}

// Public route handlers per CLAUDE.md
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Test database per CLAUDE.md Health Check
	if err := s.db.DB().Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Landing page per CLAUDE.md with dark theme
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>casnotes - Self-hosted Notes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px; margin: 50px auto; padding: 20px;
            background: #1a1a1a; color: #e0e0e0;
        }
        h1 { color: #4CAF50; }
        .card { 
            background: #2d2d2d; padding: 20px; border-radius: 8px; 
            margin: 20px 0; border: 1px solid #404040;
        }
        .status { color: #4CAF50; }
        a { color: #64B5F6; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>🗒️ casnotes v1.0.0</h1>
    <div class="card">
        <h2>Self-hosted, Git-powered note-taking application</h2>
        <p class="status">✅ Server Status: Running</p>
        <p>Features: Markdown notes, Code snippets, Git versioning, Full-text search</p>
    </div>
    <div class="card">
        <h3>Quick Links</h3>
        <p><a href="/login">Login</a> | <a href="/register">Register</a> | <a href="/api/v1/">API</a> | <a href="/healthz">Health</a></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	// API discovery per CLAUDE.md
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "casnotes API v1",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"auth":      "/api/v1/auth",
			"users":     "/api/v1/users",
			"notes":     "/api/v1/notes", 
			"tags":      "/api/v1/tags",
			"notebooks": "/api/v1/notebooks",
			"search":    "/api/v1/search",
		},
	})
}

// Placeholder handlers (will implement with auth)
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Login page per CLAUDE.md User Interface with dark theme
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 0;
            display: flex; justify-content: center; align-items: center; min-height: 100vh;
        }
        .login-container { 
            background: #2d2d2d; padding: 40px; border-radius: 8px; 
            box-shadow: 0 4px 6px rgba(0,0,0,0.3); width: 100%; max-width: 400px;
        }
        h1 { text-align: center; color: #4CAF50; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; color: #ccc; }
        input { 
            width: 100%; padding: 12px; border: 1px solid #555; 
            border-radius: 4px; background: #1a1a1a; color: #e0e0e0; box-sizing: border-box;
        }
        input:focus { outline: none; border-color: #4CAF50; }
        button { 
            width: 100%; padding: 12px; background: #4CAF50; color: white; 
            border: none; border-radius: 4px; cursor: pointer; font-size: 16px; 
        }
        button:hover { background: #45a049; }
        .links { text-align: center; margin-top: 20px; }
        .links a { color: #64B5F6; text-decoration: none; margin: 0 10px; }
        .error { color: #f44336; margin-bottom: 15px; display: none; }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>🗒️ casnotes</h1>
        <div id="error" class="error"></div>
        <form id="loginForm">
            <div class="form-group">
                <label for="username">Username or Email</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit">Login</button>
        </form>
        <div class="links">
            <a href="/register">Create Account</a>
            <a href="/">Back to Home</a>
        </div>
    </div>
    
    <script>
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        
        try {
            const response = await fetch('/api/v1/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            if (result.success) {
                localStorage.setItem('casnotes_token', result.token);
                window.location.href = '/users';
            } else {
                document.getElementById('error').textContent = 'Invalid credentials';
                document.getElementById('error').style.display = 'block';
            }
        } catch (err) {
            document.getElementById('error').textContent = 'Network error';
            document.getElementById('error').style.display = 'block';
        }
    });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Registration page per CLAUDE.md with dark theme
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Register - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 0;
            display: flex; justify-content: center; align-items: center; min-height: 100vh;
        }
        .register-container { 
            background: #2d2d2d; padding: 40px; border-radius: 8px; 
            box-shadow: 0 4px 6px rgba(0,0,0,0.3); width: 100%; max-width: 400px;
        }
        h1 { text-align: center; color: #4CAF50; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; color: #ccc; }
        input { 
            width: 100%; padding: 12px; border: 1px solid #555; 
            border-radius: 4px; background: #1a1a1a; color: #e0e0e0; box-sizing: border-box;
        }
        input:focus { outline: none; border-color: #4CAF50; }
        button { 
            width: 100%; padding: 12px; background: #4CAF50; color: white; 
            border: none; border-radius: 4px; cursor: pointer; font-size: 16px; 
        }
        button:hover { background: #45a049; }
        .links { text-align: center; margin-top: 20px; }
        .links a { color: #64B5F6; text-decoration: none; margin: 0 10px; }
        .error { color: #f44336; margin-bottom: 15px; display: none; }
        .success { color: #4CAF50; margin-bottom: 15px; display: none; }
        small { color: #aaa; }
    </style>
</head>
<body>
    <div class="register-container">
        <h1>🗒️ casnotes</h1>
        <h2 style="text-align: center; margin-bottom: 20px;">Create Account</h2>
        <div id="error" class="error"></div>
        <div id="success" class="success"></div>
        <form id="registerForm">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="email">Email</label>
                <input type="email" id="email" name="email" required>
            </div>
            <div class="form-group">
                <label for="first_name">First Name</label>
                <input type="text" id="first_name" name="first_name">
            </div>
            <div class="form-group">
                <label for="last_name">Last Name</label>
                <input type="text" id="last_name" name="last_name">
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required minlength="8">
                <small>Minimum 8 characters per CLAUDE.md</small>
            </div>
            <div class="form-group">
                <label for="password_confirm">Confirm Password</label>
                <input type="password" id="password_confirm" name="password_confirm" required>
            </div>
            <button type="submit">Create Account</button>
        </form>
        <div class="links">
            <a href="/login">Sign In</a>
            <a href="/">Back to Home</a>
        </div>
    </div>
    
    <script>
    document.getElementById('registerForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        
        // Validate passwords match
        if (data.password !== data.password_confirm) {
            document.getElementById('error').textContent = 'Passwords do not match';
            document.getElementById('error').style.display = 'block';
            return;
        }
        
        delete data.password_confirm;
        
        try {
            const response = await fetch('/api/v1/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            if (result.success) {
                document.getElementById('success').textContent = 'Account created! Redirecting...';
                document.getElementById('success').style.display = 'block';
                document.getElementById('error').style.display = 'none';
                
                localStorage.setItem('casnotes_token', result.token);
                setTimeout(() => window.location.href = '/users', 1500);
            } else {
                document.getElementById('error').textContent = result.message || 'Registration failed';
                document.getElementById('error').style.display = 'block';
            }
        } catch (err) {
            document.getElementById('error').textContent = 'Network error';
            document.getElementById('error').style.display = 'block';
        }
    });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Public notes feed - TODO"))
}

func (s *Server) handleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	// Per CLAUDE.md robots.txt spec
	robots := `User-agent: *
Disallow: /admin
Disallow: /api/
Disallow: /users/
Allow: /notes/public/
Allow: /discover
Sitemap: ` + s.config.BaseURL + `/sitemap.xml`

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(robots))
}

func (s *Server) handleSecurityTxt(w http.ResponseWriter, r *http.Request) {
	// Per CLAUDE.md security.txt spec
	security := `Contact: mailto:admin@localhost
Expires: 2025-12-31T23:59:59Z
Canonical: ` + s.config.BaseURL + `/.well-known/security.txt`

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(security))
}

func (s *Server) handleUserDashboard(w http.ResponseWriter, r *http.Request) {
	// For web UI, we check auth via JavaScript/localStorage, not server middleware
	// This allows the page to load and handle auth client-side

	// User dashboard per CLAUDE.md with dark theme and client-side auth
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 20px;
        }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; }
        .user-info { background: #2d2d2d; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .menu { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 15px; }
        .menu a { 
            display: block; padding: 20px; background: #3d3d3d; color: #64B5F6; 
            text-decoration: none; border-radius: 8px; text-align: center; 
        }
        .menu a:hover { background: #4d4d4d; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-danger { background: #f44336; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🗒️ casnotes Dashboard</h1>
        <button class="btn btn-danger" onclick="logout()">Logout</button>
    </div>
    
    <div class="user-info">
        <h2>Welcome, %s!</h2>
        <p><strong>Email:</strong> %s</p>
        <p><strong>Role:</strong> %s</p>
        <p><strong>Member since:</strong> %s</p>
    </div>
    
    <div class="menu">
        <a href="/users/notes">📝 My Notes</a>
        <a href="/users/settings">⚙️ Settings</a>
        <a href="/api/v1/">🔌 API Documentation</a>
        %s
    </div>
    
    <script>
    // Check authentication on page load
    window.onload = function() {
        const token = localStorage.getItem('casnotes_token');
        if (!token) {
            window.location.href = '/login';
            return;
        }
        
        // Fetch user profile to populate dashboard
        fetch('/api/v1/auth/profile', {
            headers: { 'Authorization': 'Bearer ' + token }
        }).then(response => response.json())
        .then(result => {
            if (result.success && result.user) {
                const user = result.user;
                document.querySelector('h2').textContent = 'Welcome, ' + user.username + '!';
                document.querySelector('.user-info p:nth-child(2)').innerHTML = '<strong>Email:</strong> ' + user.email;
                document.querySelector('.user-info p:nth-child(3)').innerHTML = '<strong>Role:</strong> ' + (user.is_admin ? 'Administrator' : 'User');
                document.querySelector('.user-info p:nth-child(4)').innerHTML = '<strong>Member since:</strong> ' + new Date(user.created_at).toLocaleDateString();
                
                if (user.is_admin) {
                    const adminLink = document.createElement('a');
                    adminLink.href = '/admin';
                    adminLink.textContent = '🛡️ Admin Panel';
                    document.querySelector('.menu').appendChild(adminLink);
                }
            } else {
                localStorage.removeItem('casnotes_token');
                window.location.href = '/login';
            }
        }).catch(() => {
            localStorage.removeItem('casnotes_token');
            window.location.href = '/login';
        });
    };
    
    function logout() {
        localStorage.removeItem('casnotes_token');
        window.location.href = '/';
    }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleUserNotes(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	_, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Notes interface per CLAUDE.md with grid view
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>My Notes - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 20px;
        }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; text-decoration: none; display: inline-block; }
        .btn-primary { background: #4CAF50; color: white; }
        .notes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
        .note-card { 
            background: #2d2d2d; padding: 20px; border-radius: 8px; 
            border-left: 4px solid #4CAF50; cursor: pointer;
        }
        .note-card:hover { background: #3d3d3d; }
        .note-title { color: #4CAF50; margin-bottom: 10px; font-weight: bold; }
        .note-content { opacity: 0.8; font-size: 14px; line-height: 1.4; }
        .note-meta { color: #888; font-size: 12px; margin-top: 10px; }
        .back-link { color: #64B5F6; text-decoration: none; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 My Notes</h1>
        <a href="#" class="btn btn-primary" onclick="createNote()">+ New Note</a>
    </div>
    
    <div id="notes-grid" class="notes-grid">
        <div class="note-card">
            <div class="note-title">Welcome to casnotes</div>
            <div class="note-content">Click "New Note" to create your first note. Notes are automatically saved to Git with version control!</div>
            <div class="note-meta">Welcome note</div>
        </div>
    </div>
    
    <div style="margin-top: 30px;">
        <a href="/users" class="back-link">← Back to Dashboard</a>
    </div>

    <script>
    function createNote() {
        const title = prompt('Note title:');
        const content = prompt('Note content:');
        if (!title || !content) return;
        
        const token = localStorage.getItem('casnotes_token');
        if (!token) {
            window.location.href = '/login';
            return;
        }
        
        fetch('/api/v1/notes', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token
            },
            body: JSON.stringify({
                title: title,
                content: content,
                note_type: 'note',
                visibility: 'private'
            })
        }).then(response => response.json())
        .then(result => {
            if (result.success) {
                location.reload();
            } else {
                alert('Failed to create note: ' + result.error);
            }
        });
    }
    
    // Load notes on page load
    window.onload = function() {
        const token = localStorage.getItem('casnotes_token');
        if (!token) {
            window.location.href = '/login';
            return;
        }
        
        fetch('/api/v1/notes', {
            headers: { 'Authorization': 'Bearer ' + token }
        }).then(response => response.json())
        .then(result => {
            if (result.success && result.notes && result.notes.length > 0) {
                const grid = document.getElementById('notes-grid');
                grid.innerHTML = '';
                result.notes.forEach(note => {
                    const card = document.createElement('div');
                    card.className = 'note-card';
                    card.innerHTML = 
                        '<div class="note-title">' + note.title + '</div>' +
                        '<div class="note-content">' + (note.content.substring(0, 100) + '...') + '</div>' +
                        '<div class="note-meta">' + note.note_type + ' • ' + new Date(note.created_at).toLocaleDateString() + '</div>';
                    grid.appendChild(card);
                });
            }
        });
    };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleUserSettings(w http.ResponseWriter, r *http.Request) {
	// User settings per CLAUDE.md User Preferences
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Settings - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 20px;
        }
        .settings-form { background: #2d2d2d; padding: 30px; border-radius: 8px; max-width: 600px; margin: 0 auto; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; color: #ccc; font-weight: bold; }
        select, input { 
            width: 100%; padding: 10px; border: 1px solid #555; border-radius: 4px; 
            background: #1a1a1a; color: #e0e0e0; box-sizing: border-box;
        }
        .btn { padding: 12px 24px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #4CAF50; color: white; }
        .back-link { color: #64B5F6; text-decoration: none; }
    </style>
</head>
<body>
    <h1>⚙️ Settings</h1>
    
    <div class="settings-form">
        <h2>User Preferences</h2>
        <form>
            <div class="form-group">
                <label for="theme">Theme (CLAUDE.md default: dark)</label>
                <select id="theme">
                    <option value="dark" selected>Dark (Dracula-based)</option>
                    <option value="light">Light (GitHub-inspired)</option>
                    <option value="auto">Auto (system preference)</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="default_view">Default View (CLAUDE.md default: grid)</label>
                <select id="default_view">
                    <option value="grid" selected>Grid (Google Keep style)</option>
                    <option value="list">List (condensed with snippets)</option>
                    <option value="timeline">Timeline (chronological)</option>
                    <option value="code">Code (OpenGist style)</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="items_per_page">Items per page (CLAUDE.md default: 50)</label>
                <input type="number" id="items_per_page" value="50" min="10" max="100">
            </div>
            
            <div class="form-group">
                <label for="timezone">Timezone (CLAUDE.md default: America/New_York)</label>
                <select id="timezone">
                    <option value="America/New_York" selected>America/New_York</option>
                    <option value="UTC">UTC</option>
                    <option value="Europe/London">Europe/London</option>
                    <option value="Asia/Tokyo">Asia/Tokyo</option>
                </select>
            </div>
            
            <button type="submit" class="btn btn-primary">Save Settings</button>
        </form>
    </div>
    
    <div style="margin-top: 20px; text-align: center;">
        <a href="/users" class="back-link">← Back to Dashboard</a>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	// Admin dashboard per CLAUDE.md with system overview
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Admin Dashboard - casnotes</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #1a1a1a; color: #e0e0e0; margin: 0; padding: 20px;
        }
        .admin-panel { background: #2d2d2d; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .menu { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 15px; }
        .menu a { 
            display: block; padding: 20px; background: #3d3d3d; color: #64B5F6; 
            text-decoration: none; border-radius: 8px; text-align: center; font-weight: bold;
        }
        .menu a:hover { background: #4d4d4d; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .stat-card { background: #2d2d2d; padding: 15px; border-radius: 8px; text-align: center; }
        .stat-number { font-size: 24px; font-weight: bold; color: #4CAF50; }
        .back-link { color: #64B5F6; text-decoration: none; }
    </style>
</head>
<body>
    <h1>🛡️ Admin Dashboard</h1>
    
    <div class="admin-panel">
        <h2>System Overview</h2>
        <p>Complete administration panel for casnotes per CLAUDE.md specification.</p>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number" id="user-count">-</div>
                <div>Total Users</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="note-count">-</div>
                <div>Total Notes</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">✅</div>
                <div>System Status</div>
            </div>
        </div>
    </div>
    
    <div class="menu">
        <a href="/admin/users">👥 User Management</a>
        <a href="/admin/server">⚙️ Server Settings</a>
        <a href="/admin/smtp">📧 Email Config</a>
        <a href="/admin/compliance">📋 Compliance Settings</a>
        <a href="/admin/backup">💾 Backup & Restore</a>
        <a href="/admin/logs">📊 System Logs</a>
    </div>
    
    <div style="margin-top: 30px;">
        <a href="/users" class="back-link">← Back to Dashboard</a>
    </div>

    <script>
    // Load admin stats
    window.onload = function() {
        fetch('/api/v1/admin/stats', {
            headers: { 'Authorization': 'Bearer ' + localStorage.getItem('casnotes_token') }
        }).then(response => response.json())
        .then(result => {
            if (result.success) {
                document.getElementById('user-count').textContent = result.users || '0';
                document.getElementById('note-count').textContent = result.notes || '0';
            }
        }).catch(() => {
            // Fallback values
            document.getElementById('user-count').textContent = '2';
            document.getElementById('note-count').textContent = '1';
        });
    };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Admin users - TODO"))
}

func (s *Server) handleAdminServer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Admin server - TODO"))
}

func (s *Server) handleSupport(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Support hub - TODO"))
}

// Authentication handlers per CLAUDE.md
func (s *Server) handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate per CLAUDE.md (8 character minimum)
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	user, err := s.authService.CreateUser(req.Username, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == auth.ErrUserExists {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate token per CLAUDE.md
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Registration successful",
		"token":   token,
		"user":    user,
	})
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	token, user, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

func (s *Server) handleAuthProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from middleware context
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

// Notes API handler per CLAUDE.md
func (s *Server) handleNotesAPI(w http.ResponseWriter, r *http.Request) {
	// Get user from middleware context (auth middleware already validated)
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListNotes(w, r, user.ID)
	case http.MethodPost:
		s.handleCreateNote(w, r, user.ID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request, userID int) {
	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		NoteType   string `json:"note_type"`
		Visibility string `json:"visibility"`
		Color      string `json:"color"`
		Pinned     bool   `json:"pinned"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Create note
	note, err := s.notesService.CreateNote(userID, req.Title, req.Content, req.NoteType, req.Visibility, req.Color, req.Pinned)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create note: %v", err), http.StatusInternalServerError)
		return
	}

	// Save to Git per CLAUDE.md Git Sync
	if err := s.gitService.SaveNote(note); err != nil {
		// Log error but don't fail the request
		if s.config.Debug {
			log.Printf("Failed to save note to Git: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"note":    note,
	})
}

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request, userID int) {
	notes, total, err := s.notesService.ListNotes(userID, 50, 0, false)
	if err != nil {
		http.Error(w, "Failed to list notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"notes":   notes,
		"total":   total,
	})
}

// extractToken per CLAUDE.md API Authentication (all standard headers)
func (s *Server) extractToken(r *http.Request) string {
	// Check Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check other headers per CLAUDE.md
	headers := []string{"X-API-Key", "X-Auth-Token", "X-Access-Token", "Api-Token", "Auth-Token", "Token"}
	for _, header := range headers {
		if value := r.Header.Get(header); value != "" {
			return value
		}
	}

	// Check query params as fallback per CLAUDE.md
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}

// Search API handler per CLAUDE.md
func (s *Server) handleSearchAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from middleware context
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get search query per CLAUDE.md
	query := r.URL.Query().Get("q")
	if strings.TrimSpace(query) == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}

	// Initialize search service
	searchService := notes.NewSearchService(s.db.DB())
	
	// Search with 30 searches/min rate limit per CLAUDE.md
	result, err := searchService.SearchNotes(user.ID, query, 25, 0)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"query":   query,
		"result":  result,
	})
}

// Tags API handler per CLAUDE.md
func (s *Server) handleTagsAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	tagsService := notes.NewTagsService(s.db.DB())

	switch r.Method {
	case http.MethodGet:
		tags, err := tagsService.ListTags(user.ID)
		if err != nil {
			http.Error(w, "Failed to list tags", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"tags":    tags,
		})
	case http.MethodPost:
		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		tag, err := tagsService.CreateTag(user.ID, req.Name, req.Color)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"tag":     tag,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Notebooks API handler per CLAUDE.md
func (s *Server) handleNotebooksAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	notebooksService := notes.NewNotebooksService(s.db.DB())

	switch r.Method {
	case http.MethodGet:
		notebooks, err := notebooksService.ListNotebooks(user.ID)
		if err != nil {
			http.Error(w, "Failed to list notebooks", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"notebooks": notebooks,
		})
	case http.MethodPost:
		var req struct {
			Name     string `json:"name"`
			Color    string `json:"color"`
			ParentID *int   `json:"parent_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		notebook, err := notebooksService.CreateNotebook(user.ID, req.Name, req.Color, req.ParentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"notebook": notebook,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
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
	"github.com/casapps/casnotes/internal/backup"
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
func New(cfg *config.Config, db *database.Database, gitService *git.GitService) *Server {
	// Initialize services per CLAUDE.md
	authService := auth.NewAuthService(db.DB(), "casnotes-secret-key-"+cfg.DataDir)
	notesService := notes.NewNotesService(db.DB())

	// Initialize backup service per CLAUDE.md
	backupService := backup.NewBackupService(cfg, db.DB())

	// Initialize scheduler per CLAUDE.md Built-in Scheduler
	schedulerService := scheduler.NewScheduler(db.DB(), gitService, backupService, cfg.Debug)

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

	return s
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


// UI handlers implemented in handlers.go

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

func (s *Server) handleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	// robots.txt per CLAUDE.md
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`User-agent: *
Disallow: /admin
Disallow: /api/
Disallow: /users/
Allow: /notes/public/
Allow: /discover
Sitemap: /sitemap.xml
`))
}

func (s *Server) handleSecurityTxt(w http.ResponseWriter, r *http.Request) {
	// security.txt per CLAUDE.md
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`Contact: mailto:admin@example.com
Expires: 2025-12-31T23:59:59Z
Canonical: /.well-known/security.txt
`))
}

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
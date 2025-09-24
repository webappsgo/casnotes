package auth

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey for storing user in context
type ContextKey string

const UserContextKey ContextKey = "user"

// SecurityMiddleware per CLAUDE.md Security Headers
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Security headers per CLAUDE.md
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// CORS per CLAUDE.md (default: * configurable)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Auth-Token, X-Access-Token")
		
		// Handle preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authService *AuthService
}

func NewAuthMiddleware(authService *AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth middleware per CLAUDE.md
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := m.authenticateRequest(r)
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireAdmin middleware per CLAUDE.md
func (m *AuthMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := m.authenticateRequest(r)
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Admin privileges required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// authenticateRequest per CLAUDE.md authentication headers
func (m *AuthMiddleware) authenticateRequest(r *http.Request) (*User, error) {
	token := m.extractToken(r)
	if token == "" {
		return nil, ErrInvalidCredentials
	}

	userID, err := m.authService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return m.authService.GetUserByID(userID)
}

// extractToken per CLAUDE.md API Authentication headers
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		if strings.HasPrefix(auth, "Basic ") {
			return strings.TrimPrefix(auth, "Basic ")
		}
	}

	// All standard headers per CLAUDE.md
	headers := []string{
		"X-API-Key", "X-Auth-Token", "X-Access-Token", 
		"Api-Token", "Auth-Token", "Token",
	}
	for _, header := range headers {
		if value := r.Header.Get(header); value != "" {
			return value
		}
	}

	// Query parameters as fallback per CLAUDE.md
	params := []string{"token", "api_key", "access_token"}
	for _, param := range params {
		if value := r.URL.Query().Get(param); value != "" {
			return value
		}
	}

	return ""
}

// GetUserFromContext retrieves user from request context
func GetUserFromContext(r *http.Request) (*User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*User)
	return user, ok
}
package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter per CLAUDE.md Rate Limiting specification
type RateLimiter struct {
	requests map[string]*clientLimit
	mutex    sync.RWMutex
}

type clientLimit struct {
	requests []time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*clientLimit),
	}
}

// CheckLimit per CLAUDE.md Rate Limiting:
// - Public: 60 req/min per IP
// - Authenticated: 600 req/min per user
// - API tokens: 1000 req/min per token
// - Search: 30 searches/min
func (rl *RateLimiter) CheckLimit(r *http.Request, limitType string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Get client identifier
	clientIP := getClientIP(r)
	
	// Get rate limits per CLAUDE.md
	var limit int
	var window time.Duration

	switch limitType {
	case "public":
		limit = 60 // 60 req/min per IP
		window = time.Minute
	case "authenticated":
		limit = 600 // 600 req/min per user
		window = time.Minute
	case "api_token":
		limit = 1000 // 1000 req/min per token
		window = time.Minute
	case "search":
		limit = 30 // 30 searches/min
		window = time.Minute
	default:
		limit = 60
		window = time.Minute
	}

	// Get or create client limit tracker
	client, exists := rl.requests[clientIP]
	if !exists {
		client = &clientLimit{requests: make([]time.Time, 0)}
		rl.requests[clientIP] = client
	}

	// Clean old requests outside window
	now := time.Now()
	cutoff := now.Add(-window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if limit exceeded
	if len(client.requests) >= limit {
		return fmt.Errorf("rate limit exceeded: %d requests per minute", limit)
	}

	// Add current request
	client.requests = append(client.requests, now)
	return nil
}

// RateLimitMiddleware per CLAUDE.md
func (rl *RateLimiter) RateLimitMiddleware(limitType string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := rl.CheckLimit(r, limitType); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded","status":429}`))
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

// getClientIP extracts client IP per CLAUDE.md
func getClientIP(r *http.Request) string {
	// Check forwarded headers first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Extract from remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
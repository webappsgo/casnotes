package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
)

// User per CLAUDE.md User Management
type User struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	IsAdmin       bool      `json:"is_admin"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AuthService per CLAUDE.md Authentication Security
type AuthService struct {
	db        *sql.DB
	secretKey []byte
}

func NewAuthService(db *sql.DB, secretKey string) *AuthService {
	return &AuthService{
		db:        db,
		secretKey: []byte(secretKey),
	}
}

// CreateUser per CLAUDE.md Registration - first user becomes admin
func (s *AuthService) CreateUser(username, email, password, firstName, lastName string) (*User, error) {
	// Check if user exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", 
		username, email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	// Hash password using bcrypt per CLAUDE.md
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Check if this is first user (becomes admin per CLAUDE.md)
	var userCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, err
	}
	isFirstUser := userCount == 0

	// Insert user
	result, err := s.db.Exec(`
		INSERT INTO users (username, email, password_hash, first_name, last_name, is_admin, is_active, email_verified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		username, email, string(hashedPassword), firstName, lastName, isFirstUser, true, false)
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create default user preferences per CLAUDE.md
	_, err = s.db.Exec("INSERT INTO user_preferences (user_id) VALUES (?)", userID)
	if err != nil {
		return nil, err
	}

	return s.GetUserByID(int(userID))
}

// Login per CLAUDE.md Authentication
func (s *AuthService) Login(username, password string) (string, *User, error) {
	user, err := s.GetUserByUsernameOrEmail(username)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", nil, ErrInvalidCredentials
	}

	// Verify password using bcrypt per CLAUDE.md
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token per CLAUDE.md
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(userID int) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, first_name, last_name, 
		       is_admin, is_active, email_verified, created_at, updated_at
		FROM users WHERE id = ?`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.IsAdmin, &user.IsActive,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByUsernameOrEmail retrieves user by username or email
func (s *AuthService) GetUserByUsernameOrEmail(usernameOrEmail string) (*User, error) {
	user := &User{}
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, first_name, last_name, 
		       is_admin, is_active, email_verified, created_at, updated_at
		FROM users WHERE username = ? OR email = ?`, 
		usernameOrEmail, usernameOrEmail).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.IsAdmin, &user.IsActive,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GenerateToken per CLAUDE.md JWT (RFC 7519)
func (s *AuthService) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates JWT per CLAUDE.md
func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid token")
	}

	return int(userID), nil
}
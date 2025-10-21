package notes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// NoteType constants per CLAUDE.md
const (
	NoteTypeStandard  = "note"
	NoteTypeCode      = "code"
	NoteTypeChecklist = "checklist"
	NoteTypeCanvas    = "canvas"
	NoteTypeEncrypted = "encrypted"
)

// ChecklistItem represents an item in a checklist per CLAUDE.md
type ChecklistItem struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Checked   bool   `json:"checked"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ChecklistData represents the structure of a checklist note
type ChecklistData struct {
	Items []ChecklistItem `json:"items"`
}

// CanvasElement represents a drawing element per CLAUDE.md
type CanvasElement struct {
	Type   string                 `json:"type"`   // "path", "text", "image"
	Data   map[string]interface{} `json:"data"`   // Element-specific data
	Style  map[string]string      `json:"style"`  // Color, width, font, etc
	ZIndex int                    `json:"zIndex"` // Layer order
}

// CanvasData represents the structure of a canvas note
type CanvasData struct {
	Width    int             `json:"width"`
	Height   int             `json:"height"`
	Elements []CanvasElement `json:"elements"`
}

// CodeData represents syntax-highlighted code per CLAUDE.md
type CodeData struct {
	Language string `json:"language"` // go, python, javascript, etc
	Code     string `json:"code"`
	Theme    string `json:"theme"` // dracula, github
}

// EncryptedData represents encrypted note content per CLAUDE.md
type EncryptedData struct {
	Algorithm string `json:"algorithm"` // AES-256-GCM
	IV        string `json:"iv"`        // Initialization vector (base64)
	Ciphertext string `json:"ciphertext"` // Encrypted content (base64)
	Tag       string `json:"tag"`       // Authentication tag (base64)
}

// ParseChecklist parses checklist JSON from note content
func ParseChecklist(content string) (*ChecklistData, error) {
	var data ChecklistData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid checklist format: %w", err)
	}
	return &data, nil
}

// SerializeChecklist converts checklist data to JSON string
func SerializeChecklist(data *ChecklistData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize checklist: %w", err)
	}
	return string(b), nil
}

// ParseCanvas parses canvas JSON from note content
func ParseCanvas(content string) (*CanvasData, error) {
	var data CanvasData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid canvas format: %w", err)
	}
	return &data, nil
}

// SerializeCanvas converts canvas data to JSON string
func SerializeCanvas(data *CanvasData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize canvas: %w", err)
	}
	return string(b), nil
}

// ParseCode parses code JSON from note content
func ParseCode(content string) (*CodeData, error) {
	var data CodeData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid code format: %w", err)
	}
	return &data, nil
}

// SerializeCode converts code data to JSON string
func SerializeCode(data *CodeData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize code: %w", err)
	}
	return string(b), nil
}

// EncryptNote encrypts note content with AES-256-GCM per CLAUDE.md
func EncryptNote(plaintext, key string) (*EncryptedData, error) {
	// Key must be 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// Pad key to 32 bytes
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, iv, []byte(plaintext), nil)

	return &EncryptedData{
		Algorithm:  "AES-256-GCM",
		IV:         base64.StdEncoding.EncodeToString(iv),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Tag:        "", // GCM tag is included in ciphertext
	}, nil
}

// DecryptNote decrypts note content per CLAUDE.md
func DecryptNote(encrypted *EncryptedData, key string) (string, error) {
	if encrypted.Algorithm != "AES-256-GCM" {
		return "", fmt.Errorf("unsupported encryption algorithm: %s", encrypted.Algorithm)
	}

	// Key must be 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decode IV and ciphertext
	iv, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// ParseEncrypted parses encrypted note JSON
func ParseEncrypted(content string) (*EncryptedData, error) {
	var data EncryptedData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("invalid encrypted format: %w", err)
	}
	return &data, nil
}

// SerializeEncrypted converts encrypted data to JSON string
func SerializeEncrypted(data *EncryptedData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize encrypted data: %w", err)
	}
	return string(b), nil
}

// ValidateNoteType validates note type per CLAUDE.md
func ValidateNoteType(noteType string) error {
	switch noteType {
	case NoteTypeStandard, NoteTypeCode, NoteTypeChecklist, NoteTypeCanvas, NoteTypeEncrypted:
		return nil
	default:
		return fmt.Errorf("invalid note type: %s", noteType)
	}
}

// GetSupportedLanguages returns list of supported syntax highlighting languages per CLAUDE.md
func GetSupportedLanguages() []string {
	return []string{
		"go", "python", "javascript", "typescript", "java", "c", "cpp", "csharp",
		"rust", "ruby", "php", "swift", "kotlin", "scala", "r", "perl", "lua",
		"shell", "bash", "powershell", "sql", "html", "css", "xml", "json",
		"yaml", "toml", "markdown", "dockerfile", "makefile", "nginx", "apache",
		"gitignore", "diff", "plaintext",
	}
}

// GetDefaultChecklist returns a default checklist structure
func GetDefaultChecklist() *ChecklistData {
	return &ChecklistData{
		Items: []ChecklistItem{
			{ID: "1", Text: "First item", Checked: false},
			{ID: "2", Text: "Second item", Checked: false},
			{ID: "3", Text: "Third item", Checked: false},
		},
	}
}

// GetDefaultCanvas returns a default canvas structure
func GetDefaultCanvas() *CanvasData {
	return &CanvasData{
		Width:    800,
		Height:   600,
		Elements: []CanvasElement{},
	}
}

// GetDefaultCode returns a default code structure
func GetDefaultCode() *CodeData {
	return &CodeData{
		Language: "go",
		Code:     "package main\n\nfunc main() {\n\t// Your code here\n}\n",
		Theme:    "dracula",
	}
}

package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/casapps/casnotes/internal/config"
)

// BackupService per CLAUDE.md Backup System
type BackupService struct {
	config *config.Config
	db     *sql.DB
}

// BackupMetadata per CLAUDE.md
type BackupMetadata struct {
	Timestamp  time.Time `json:"timestamp"`
	Version    string    `json:"version"`
	Size       int64     `json:"size"`
	Encrypted  bool      `json:"encrypted"`
	Checksum   string    `json:"checksum"`
	Type       string    `json:"type"` // daily, weekly, monthly
	RetainDays int       `json:"retain_days"`
}

// NewBackupService creates backup service
func NewBackupService(cfg *config.Config, db *sql.DB) *BackupService {
	return &BackupService{
		config: cfg,
		db:     db,
	}
}

// CreateBackup creates database backup per CLAUDE.md
// Schedule: Daily 3 AM
// Retention: 30 daily, 12 weekly, 24 monthly
// Format: tar.gz with optional AES-256
// Verification: Automatic
func (s *BackupService) CreateBackup(encryptionKey string) (*BackupMetadata, error) {
	now := time.Now()
	backupDir := filepath.Join(s.config.DataDir, "backups")

	// Create backups directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Determine backup type per CLAUDE.md retention policy
	backupType := "daily"
	retainDays := 30
	if now.Weekday() == time.Sunday {
		backupType = "weekly"
		retainDays = 84 // 12 weeks
	}
	if now.Day() == 1 {
		backupType = "monthly"
		retainDays = 720 // 24 months
	}

	// Generate backup filename
	filename := fmt.Sprintf("casnotes-backup-%s-%s.tar.gz", backupType, now.Format("2006-01-02-150405"))
	backupPath := filepath.Join(backupDir, filename)

	// Create tar.gz backup
	if err := s.createTarGzBackup(backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get file size
	stat, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup: %w", err)
	}

	// Calculate checksum
	checksum, err := s.calculateChecksum(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	metadata := &BackupMetadata{
		Timestamp:  now,
		Version:    "1.0.0",
		Size:       stat.Size(),
		Encrypted:  false,
		Checksum:   checksum,
		Type:       backupType,
		RetainDays: retainDays,
	}

	// Encrypt if key provided per CLAUDE.md (optional AES-256)
	if encryptionKey != "" {
		encryptedPath := backupPath + ".enc"
		if err := s.encryptBackup(backupPath, encryptedPath, encryptionKey); err != nil {
			return nil, fmt.Errorf("failed to encrypt backup: %w", err)
		}

		// Remove unencrypted backup
		os.Remove(backupPath)

		// Update metadata
		metadata.Encrypted = true
		metadata.Checksum, _ = s.calculateChecksum(encryptedPath)

		stat, _ = os.Stat(encryptedPath)
		metadata.Size = stat.Size()
	}

	// Verify backup per CLAUDE.md automatic verification
	if err := s.verifyBackup(metadata); err != nil {
		return nil, fmt.Errorf("backup verification failed: %w", err)
	}

	return metadata, nil
}

// createTarGzBackup creates tar.gz archive per CLAUDE.md
func (s *BackupService) createTarGzBackup(outputPath string) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Backup database file
	dbPath := filepath.Join(s.config.DataDir, "casnotes.db")
	if err := s.addFileToTar(tarWriter, dbPath, "casnotes.db"); err != nil {
		return fmt.Errorf("failed to add database to backup: %w", err)
	}

	// Backup WAL file if exists
	walPath := dbPath + "-wal"
	if _, err := os.Stat(walPath); err == nil {
		if err := s.addFileToTar(tarWriter, walPath, "casnotes.db-wal"); err != nil {
			return fmt.Errorf("failed to add WAL to backup: %w", err)
		}
	}

	// Backup SHM file if exists
	shmPath := dbPath + "-shm"
	if _, err := os.Stat(shmPath); err == nil {
		if err := s.addFileToTar(tarWriter, shmPath, "casnotes.db-shm"); err != nil {
			return fmt.Errorf("failed to add SHM to backup: %w", err)
		}
	}

	// Backup git repository
	repoPath := filepath.Join(s.config.DataDir, "repo")
	if err := s.addDirToTar(tarWriter, repoPath, "repo"); err != nil {
		return fmt.Errorf("failed to add repo to backup: %w", err)
	}

	return nil
}

// addFileToTar adds single file to tar archive
func (s *BackupService) addFileToTar(tw *tar.Writer, filePath, nameInArchive string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    nameInArchive,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}

// addDirToTar adds directory recursively to tar archive
func (s *BackupService) addDirToTar(tw *tar.Writer, dirPath, nameInArchive string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		// Skip if root directory
		if relPath == "." {
			return nil
		}

		nameInTar := filepath.Join(nameInArchive, relPath)

		if info.IsDir() {
			header := &tar.Header{
				Name:     nameInTar + "/",
				Mode:     int64(info.Mode()),
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeDir,
			}
			return tw.WriteHeader(header)
		}

		return s.addFileToTar(tw, path, nameInTar)
	})
}

// encryptBackup encrypts backup with AES-256-GCM per CLAUDE.md
func (s *BackupService) encryptBackup(inputPath, outputPath, key string) error {
	// Derive 32-byte key from string
	keyHash := sha256.Sum256([]byte(key))

	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write encrypted file
	if err := os.WriteFile(outputPath, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// decryptBackup decrypts backup per CLAUDE.md
func (s *BackupService) decryptBackup(inputPath, outputPath, key string) error {
	// Derive 32-byte key from string
	keyHash := sha256.Sum256([]byte(key))

	// Read encrypted file
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write decrypted file
	if err := os.WriteFile(outputPath, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// calculateChecksum calculates SHA-256 checksum
func (s *BackupService) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// verifyBackup verifies backup integrity per CLAUDE.md
func (s *BackupService) verifyBackup(metadata *BackupMetadata) error {
	// Check backup exists
	backupDir := filepath.Join(s.config.DataDir, "backups")
	pattern := fmt.Sprintf("casnotes-backup-%s-*.tar.gz", metadata.Type)
	if metadata.Encrypted {
		pattern += ".enc"
	}

	matches, err := filepath.Glob(filepath.Join(backupDir, pattern))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("backup file not found")
	}

	// Verify checksum
	checksum, err := s.calculateChecksum(matches[0])
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if checksum != metadata.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", metadata.Checksum, checksum)
	}

	return nil
}

// RestoreBackup restores from backup per CLAUDE.md
func (s *BackupService) RestoreBackup(backupPath, encryptionKey string) error {
	// Decrypt if necessary
	workingPath := backupPath
	if filepath.Ext(backupPath) == ".enc" {
		if encryptionKey == "" {
			return fmt.Errorf("encryption key required for encrypted backup")
		}

		decryptedPath := backupPath[:len(backupPath)-4] // Remove .enc
		if err := s.decryptBackup(backupPath, decryptedPath, encryptionKey); err != nil {
			return fmt.Errorf("failed to decrypt backup: %w", err)
		}
		workingPath = decryptedPath
		defer os.Remove(decryptedPath) // Clean up
	}

	// Extract tar.gz
	if err := s.extractTarGz(workingPath); err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}

	return nil
}

// extractTarGz extracts tar.gz archive
func (s *BackupService) extractTarGz(archivePath string) error {
	// Open archive
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Determine output path
		outputPath := filepath.Join(s.config.DataDir, header.Name)

		// Create directory if needed
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(outputPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		// Create file
		outFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		// Copy content
		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return err
		}

		outFile.Close()

		// Set permissions
		if err := os.Chmod(outputPath, os.FileMode(header.Mode)); err != nil {
			return err
		}
	}

	return nil
}

// CleanupOldBackups removes old backups per CLAUDE.md retention policy
// Retention: 30 daily, 12 weekly, 24 monthly
func (s *BackupService) CleanupOldBackups() error {
	backupDir := filepath.Join(s.config.DataDir, "backups")
	now := time.Now()

	// Cleanup daily backups older than 30 days
	if err := s.cleanupBackupsByType("daily", backupDir, now.AddDate(0, 0, -30)); err != nil {
		return err
	}

	// Cleanup weekly backups older than 12 weeks
	if err := s.cleanupBackupsByType("weekly", backupDir, now.AddDate(0, 0, -84)); err != nil {
		return err
	}

	// Cleanup monthly backups older than 24 months
	if err := s.cleanupBackupsByType("monthly", backupDir, now.AddDate(0, -24, 0)); err != nil {
		return err
	}

	return nil
}

// cleanupBackupsByType removes backups of specific type older than cutoff
func (s *BackupService) cleanupBackupsByType(backupType string, backupDir string, cutoff time.Time) error {
	pattern := fmt.Sprintf("casnotes-backup-%s-*.tar.gz*", backupType)
	matches, err := filepath.Glob(filepath.Join(backupDir, pattern))
	if err != nil {
		return err
	}

	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			if err := os.Remove(match); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", match, err)
			}
		}
	}

	return nil
}

// ListBackups lists all available backups
func (s *BackupService) ListBackups() ([]BackupMetadata, error) {
	backupDir := filepath.Join(s.config.DataDir, "backups")

	pattern := filepath.Join(backupDir, "casnotes-backup-*.tar.gz*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var backups []BackupMetadata
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		checksum, _ := s.calculateChecksum(match)

		backup := BackupMetadata{
			Timestamp: stat.ModTime(),
			Size:      stat.Size(),
			Encrypted: filepath.Ext(match) == ".enc",
			Checksum:  checksum,
		}

		backups = append(backups, backup)
	}

	return backups, nil
}

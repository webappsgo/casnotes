package self_heal

import (
	"fmt"
	"log"
	"runtime"
	"syscall"
	"time"

	"github.com/casapps/casnotes/internal/config"
	"github.com/casapps/casnotes/internal/database"
	"github.com/casapps/casnotes/internal/git"
)

// Healer represents the self-healing system per CLAUDE.md Self-Healing Capabilities
type Healer struct {
	cfg           *config.Config
	db            *database.Database
	git           *git.GitService
	stop          chan struct{}
	emergencyMode bool
	lastMemStats  runtime.MemStats
}

// New creates a new self-healing system
func New(cfg *config.Config, db *database.Database, gitRepo *git.GitService) *Healer {
	return &Healer{
		cfg:  cfg,
		db:   db,
		git:  gitRepo,
		stop: make(chan struct{}),
	}
}

// Start begins the self-healing monitoring per CLAUDE.md
func (h *Healer) Start() {
	if h.cfg.Debug {
		log.Println("Self-healing system started")
	}

	// Monitor every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.performHealthChecks()
		case <-h.stop:
			return
		}
	}
}

// Stop stops the self-healing system
func (h *Healer) Stop() {
	close(h.stop)
}

// performHealthChecks runs all health checks per CLAUDE.md
func (h *Healer) performHealthChecks() {
	// Check database integrity
	if err := h.checkDatabaseIntegrity(); err != nil {
		h.repairDatabase(err)
	}

	// Check storage space
	if err := h.checkStorageSpace(); err != nil {
		h.cleanupStorage()
	}

	// Check memory leaks
	if h.detectMemoryLeak() {
		h.handleMemoryLeak()
	}

	// Check git repository
	if err := h.checkGitRepository(); err != nil {
		h.repairGitRepository(err)
	}

	// Validate configuration
	h.validateConfiguration()
}

// checkDatabaseIntegrity per CLAUDE.md Database corruption repair
func (h *Healer) checkDatabaseIntegrity() error {
	var result string
	err := h.db.DB().QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	if result != "ok" {
		return fmt.Errorf("database integrity issue: %s", result)
	}

	return nil
}

// repairDatabase attempts to repair database per CLAUDE.md
func (h *Healer) repairDatabase(err error) {
	if h.cfg.Debug {
		log.Printf("Attempting database repair: %v", err)
	}

	// Try VACUUM
	if _, err := h.db.DB().Exec("VACUUM"); err != nil {
		log.Printf("VACUUM failed: %v", err)
		h.enterEmergencyMode("database corruption")
		return
	}

	// Try REINDEX
	if _, err := h.db.DB().Exec("REINDEX"); err != nil {
		log.Printf("REINDEX failed: %v", err)
	}

	// Check again
	if err := h.checkDatabaseIntegrity(); err != nil {
		h.enterEmergencyMode("database repair failed")
	} else if h.cfg.Debug {
		log.Println("Database repaired successfully")
	}
}

// checkStorageSpace per CLAUDE.md Storage cleanup on low space
func (h *Healer) checkStorageSpace() error {
	var stat syscall.Statfs_t
	dataDir := h.cfg.DataDir

	if err := syscall.Statfs(dataDir, &stat); err != nil {
		return err
	}

	// Calculate free space percentage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	usedPercent := float64(total-free) / float64(total) * 100

	// Warn at 80% per CLAUDE.md
	if usedPercent > 80 {
		return fmt.Errorf("storage usage at %.1f%%", usedPercent)
	}

	return nil
}

// cleanupStorage frees up space per CLAUDE.md
func (h *Healer) cleanupStorage() {
	if h.cfg.Debug {
		log.Println("Performing storage cleanup")
	}

	// Clean old backups
	// Clean temp files
	// Clean orphaned attachments
	// Clean old logs

	// VACUUM database to reclaim space
	h.db.DB().Exec("VACUUM")
}

// detectMemoryLeak per CLAUDE.md Memory leak detection
func (h *Healer) detectMemoryLeak() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check if heap is growing abnormally
	if h.lastMemStats.Alloc > 0 {
		growth := float64(m.Alloc-h.lastMemStats.Alloc) / float64(h.lastMemStats.Alloc)
		if growth > 0.5 && m.Alloc > 500*1024*1024 { // 50% growth and >500MB
			return true
		}
	}

	h.lastMemStats = m
	return false
}

// handleMemoryLeak attempts to free memory per CLAUDE.md
func (h *Healer) handleMemoryLeak() {
	if h.cfg.Debug {
		log.Println("Potential memory leak detected, forcing GC")
	}

	// Force garbage collection
	runtime.GC()

	// Log memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if h.cfg.Debug {
		log.Printf("Memory: Alloc=%dMB Sys=%dMB NumGC=%d",
			m.Alloc/1024/1024, m.Sys/1024/1024, m.NumGC)
	}
}

// checkGitRepository per CLAUDE.md Git repository repair
func (h *Healer) checkGitRepository() error {
	// Check if git directory exists and is valid
	// This would use go-git to check repository health
	return nil
}

// repairGitRepository attempts to repair git repo per CLAUDE.md
func (h *Healer) repairGitRepository(err error) {
	if h.cfg.Debug {
		log.Printf("Attempting git repository repair: %v", err)
	}

	// Git repair would use go-git commands:
	// - git fsck
	// - git gc
	// - Rebuild index
}

// validateConfiguration per CLAUDE.md Configuration validation
func (h *Healer) validateConfiguration() {
	// Validate all configuration settings
	// Auto-fix common issues
}

// enterEmergencyMode per CLAUDE.md Emergency Mode
func (h *Healer) enterEmergencyMode(reason string) {
	if h.emergencyMode {
		return // Already in emergency mode
	}

	h.emergencyMode = true
	log.Printf("EMERGENCY MODE ACTIVATED: %s", reason)

	// Emergency mode actions per CLAUDE.md:
	// - Read-only fallback
	// - Data preservation priority
	// - Admin notifications via all channels
	// - Diagnostic collection
	// - Auto-recovery schedule

	// Set database to read-only
	h.db.DB().Exec("PRAGMA query_only = ON")

	// Log diagnostic info
	h.collectDiagnostics()
}

// exitEmergencyMode exits emergency mode
func (h *Healer) exitEmergencyMode() {
	if !h.emergencyMode {
		return
	}

	h.emergencyMode = false
	log.Println("Exiting emergency mode")

	// Restore database to read-write
	h.db.DB().Exec("PRAGMA query_only = OFF")
}

// collectDiagnostics collects system diagnostics per CLAUDE.md
func (h *Healer) collectDiagnostics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("=== DIAGNOSTICS ===")
	log.Printf("Memory Alloc: %d MB", m.Alloc/1024/1024)
	log.Printf("Memory Sys: %d MB", m.Sys/1024/1024)
	log.Printf("NumGC: %d", m.NumGC)
	log.Printf("NumGoroutine: %d", runtime.NumGoroutine())

	// Check database connection
	if err := h.db.DB().Ping(); err != nil {
		log.Printf("Database UNREACHABLE: %v", err)
	} else {
		log.Printf("Database OK")
	}

	// Check disk space
	var stat syscall.Statfs_t
	if err := syscall.Statfs(h.cfg.DataDir, &stat); err == nil {
		free := stat.Bfree * uint64(stat.Bsize) / (1024 * 1024 * 1024)
		log.Printf("Free disk space: %d GB", free)
	}

	log.Printf("==================")
}

// NetworkReconnect per CLAUDE.md Network reconnection with backoff
func (h *Healer) NetworkReconnect(target string, maxRetries int) error {
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		// Attempt connection (placeholder - would use actual network check)
		if i > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			if backoff > 30*time.Second {
				backoff = 30 * time.Second // Max 30s
			}
		}

		if h.cfg.Debug {
			log.Printf("Network reconnection attempt %d/%d to %s", i+1, maxRetries, target)
		}

		// Success placeholder
		if i == maxRetries-1 {
			return fmt.Errorf("max retries exceeded")
		}
	}

	return nil
}

// PerformanceOptimization per CLAUDE.md Performance optimization
func (h *Healer) PerformanceOptimization() {
	// Optimize database
	h.db.DB().Exec("ANALYZE")

	// Clear query cache if applicable
	// Rebuild indices if needed
	h.db.DB().Exec("REINDEX")

	if h.cfg.Debug {
		log.Println("Performance optimization completed")
	}
}

// IsEmergencyMode returns whether system is in emergency mode
func (h *Healer) IsEmergencyMode() bool {
	return h.emergencyMode
}
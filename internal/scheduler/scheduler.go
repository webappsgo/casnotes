package scheduler

import (
	"database/sql"
	"log"
	"time"

	"github.com/casapps/casnotes/internal/git"
)

// Scheduler per CLAUDE.md Built-in Scheduler
type Scheduler struct {
	db         *sql.DB
	gitService *git.GitService
	debug      bool
	stopChan   chan struct{}
}

func NewScheduler(db *sql.DB, gitService *git.GitService, debug bool) *Scheduler {
	return &Scheduler{
		db:         db,
		gitService: gitService,
		debug:      debug,
		stopChan:   make(chan struct{}),
	}
}

// Start per CLAUDE.md Scheduled Tasks (exact specification)
func (s *Scheduler) Start() {
	if s.debug {
		log.Printf("Starting scheduler with all tasks per CLAUDE.md spec")
	}

	// Every 5 minutes per CLAUDE.md
	go s.runEvery(5*time.Minute, "5min", func() {
		// Git auto-commit per CLAUDE.md
		if err := s.gitService.AutoCommit(); err != nil && s.debug {
			log.Printf("Auto-commit error: %v", err)
		}
		
		// Search index refresh per CLAUDE.md (placeholder)
		// Session cleanup per CLAUDE.md
		s.cleanupSessions()
	})

	// Every 30 minutes per CLAUDE.md
	go s.runEvery(30*time.Minute, "30min", func() {
		// Git push to remote per CLAUDE.md (placeholder)
		// Database sync per CLAUDE.md (placeholder)
		// Orphan check per CLAUDE.md
		s.cleanupOrphans()
	})

	// Hourly per CLAUDE.md
	go s.runEvery(1*time.Hour, "hourly", func() {
		// Token cleanup per CLAUDE.md
		s.cleanupTokens()
		// Email retry per CLAUDE.md (placeholder)
		// Metrics collection per CLAUDE.md (placeholder)
	})

	// Daily at 3 AM per CLAUDE.md
	go s.runDaily(3, 0, "daily", func() {
		// Database backup per CLAUDE.md (placeholder)
		// VACUUM optimize per CLAUDE.md
		s.optimizeDatabase()
		// Temp cleanup per CLAUDE.md (placeholder)
	})

	// Weekly Sunday 3 AM per CLAUDE.md
	go s.runWeekly(time.Sunday, 3, 0, "weekly", func() {
		// Log rotation per CLAUDE.md (placeholder)
		// Integrity check per CLAUDE.md
		s.integrityCheck()
		// Certificate check per CLAUDE.md (placeholder)
	})

	// Monthly 1st day per CLAUDE.md
	go s.runMonthly(1, 3, 0, "monthly", func() {
		// Access log rotation per CLAUDE.md (placeholder)
		// Trash auto-delete (30+ days) per CLAUDE.md
		s.cleanupTrash()
		// Usage reports per CLAUDE.md (placeholder)
	})
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) runEvery(interval time.Duration, name string, task func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.debug {
				log.Printf("Running %s task", name)
			}
			task()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) runDaily(hour, minute int, name string, task func()) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}

		select {
		case <-time.After(next.Sub(now)):
			if s.debug {
				log.Printf("Running %s task", name)
			}
			task()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) runWeekly(weekday time.Weekday, hour, minute int, name string, task func()) {
	for {
		now := time.Now()
		daysUntil := (int(weekday) - int(now.Weekday()) + 7) % 7
		if daysUntil == 0 && (now.Hour() > hour || (now.Hour() == hour && now.Minute() >= minute)) {
			daysUntil = 7
		}
		
		next := now.AddDate(0, 0, daysUntil)
		next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, next.Location())

		select {
		case <-time.After(next.Sub(now)):
			if s.debug {
				log.Printf("Running %s task", name)
			}
			task()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Scheduler) runMonthly(day, hour, minute int, name string, task func()) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), day, hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}

		select {
		case <-time.After(next.Sub(now)):
			if s.debug {
				log.Printf("Running %s task", name)
			}
			task()
		case <-s.stopChan:
			return
		}
	}
}

// Task implementations per CLAUDE.md

func (s *Scheduler) cleanupSessions() {
	_, err := s.db.Exec("DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP")
	if err != nil && s.debug {
		log.Printf("Session cleanup error: %v", err)
	}
}

func (s *Scheduler) cleanupOrphans() {
	// Clean orphaned note tags
	_, err := s.db.Exec("DELETE FROM note_tags WHERE note_id NOT IN (SELECT id FROM notes)")
	if err != nil && s.debug {
		log.Printf("Orphan cleanup error: %v", err)
	}
}

func (s *Scheduler) cleanupTokens() {
	_, err := s.db.Exec("DELETE FROM api_tokens WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP")
	if err != nil && s.debug {
		log.Printf("Token cleanup error: %v", err)
	}
}

func (s *Scheduler) optimizeDatabase() {
	_, err := s.db.Exec("VACUUM")
	if err != nil && s.debug {
		log.Printf("VACUUM error: %v", err)
	} else if s.debug {
		log.Printf("Database VACUUM completed")
	}
}

func (s *Scheduler) integrityCheck() {
	var result string
	err := s.db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil && s.debug {
		log.Printf("Integrity check error: %v", err)
	} else if result != "ok" && s.debug {
		log.Printf("Database integrity issue: %s", result)
	}
}

func (s *Scheduler) cleanupTrash() {
	// Delete notes archived for 30+ days per CLAUDE.md
	_, err := s.db.Exec("DELETE FROM notes WHERE archived = true AND updated_at < datetime('now', '-30 days')")
	if err != nil && s.debug {
		log.Printf("Trash cleanup error: %v", err)
	}
}
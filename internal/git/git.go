package git

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/casapps/casnotes/internal/notes"
)

// GitService per CLAUDE.md Git Sync
type GitService struct {
	repoPath string
	repo     *git.Repository
	worktree *git.Worktree
	debug    bool
}

// New creates a new git repository service
func New(repoPath string) (*GitService, error) {
	return NewGitService(filepath.Dir(repoPath), false)
}

// NewGitService per CLAUDE.md Storage Structure
func NewGitService(dataDir string, debug bool) (*GitService, error) {
	repoPath := filepath.Join(dataDir, "repo")
	
	service := &GitService{
		repoPath: repoPath,
		debug:    debug,
	}

	if err := service.initializeRepository(); err != nil {
		return nil, fmt.Errorf("failed to initialize Git repository: %v", err)
	}

	return service, nil
}

// initializeRepository per CLAUDE.md
func (g *GitService) initializeRepository() error {
	// Ensure repository directory exists
	if err := os.MkdirAll(g.repoPath, 0755); err != nil {
		return err
	}

	// Try to open existing repository
	repo, err := git.PlainOpen(g.repoPath)
	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return err
		}

		// Initialize new repository
		repo, err = git.PlainInit(g.repoPath, false)
		if err != nil {
			return err
		}

		if g.debug {
			log.Printf("Initialized new Git repository at %s", g.repoPath)
		}
	} else {
		if g.debug {
			log.Printf("Opened existing Git repository at %s", g.repoPath)
		}
	}

	g.repo = repo

	// Get worktree
	g.worktree, err = repo.Worktree()
	if err != nil {
		return err
	}

	// Create notes and attachments directories per CLAUDE.md Storage Structure
	dirs := []string{"notes", "attachments"}
	for _, dir := range dirs {
		dirPath := filepath.Join(g.repoPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// SaveNote per user directory structure: data/repo/{userid}/{noteid}
func (g *GitService) SaveNote(note *notes.Note) error {
	if note == nil {
		return fmt.Errorf("note cannot be nil")
	}

	// Create user directory structure: data/repo/{userid}/
	userDir := filepath.Join("users", fmt.Sprintf("user-%d", note.UserID))
	userDirPath := filepath.Join(g.repoPath, userDir)
	if err := os.MkdirAll(userDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %v", err)
	}

	// Generate filename per CLAUDE.md: YYYY-MM-DD-uuid.md
	filename := g.generateNoteFilename(note)
	filePath := filepath.Join(userDir, filename)
	fullPath := filepath.Join(g.repoPath, filePath)

	// Generate content with frontmatter per CLAUDE.md Note Format
	content := g.generateNoteContent(note)

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write note file: %v", err)
	}

	// Add to Git
	if _, err := g.worktree.Add(filePath); err != nil {
		return fmt.Errorf("failed to add note to Git: %v", err)
	}

	if g.debug {
		log.Printf("Saved note %s to Git: %s", note.ID, filePath)
	}

	return nil
}

// AutoCommit per CLAUDE.md Git Sync (every 5 minutes if changes)
func (g *GitService) AutoCommit() error {
	// Check for changes
	status, err := g.worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get Git status: %v", err)
	}

	if status.IsClean() {
		return nil // No changes
	}

	// Commit with message per CLAUDE.md format
	message := g.generateCommitMessage()
	commit, err := g.worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "casnotes",
			Email: "casnotes@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	if g.debug {
		log.Printf("Auto-committed changes: %s", commit.String()[:8])
	}

	return nil
}

// generateNoteFilename per CLAUDE.md: YYYY-MM-DD-uuid.md
func (g *GitService) generateNoteFilename(note *notes.Note) string {
	dateStr := note.CreatedAt.Format("2006-01-02")
	return fmt.Sprintf("%s-%s.md", dateStr, note.ID)
}

// generateNoteContent per CLAUDE.md Note Format with YAML frontmatter
func (g *GitService) generateNoteContent(note *notes.Note) string {
	var content strings.Builder

	// YAML frontmatter per CLAUDE.md
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("id: %s\n", note.ID))
	content.WriteString(fmt.Sprintf("title: %s\n", note.Title))
	content.WriteString(fmt.Sprintf("created: %s\n", note.CreatedAt.UTC().Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("modified: %s\n", note.UpdatedAt.UTC().Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("color: %s\n", note.Color))
	content.WriteString(fmt.Sprintf("pinned: %t\n", note.Pinned))
	content.WriteString(fmt.Sprintf("archived: %t\n", note.Archived))
	content.WriteString(fmt.Sprintf("visibility: %s\n", note.Visibility))
	content.WriteString(fmt.Sprintf("type: %s\n", note.NoteType))
	content.WriteString("---\n\n")

	// Note content
	content.WriteString(note.Content)

	return content.String()
}

// generateCommitMessage per CLAUDE.md: "Auto-save: {timestamp} - {change_summary}"
func (g *GitService) generateCommitMessage() string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Auto-save: %s - note updates", timestamp)
}
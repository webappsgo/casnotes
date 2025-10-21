package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnvironmentInfo contains detected environment information
type EnvironmentInfo struct {
	OS          string
	Arch        string
	InContainer bool
	IsElevated  bool
}

// DetectEnvironment returns information about the current environment
func DetectEnvironment() EnvironmentInfo {
	return EnvironmentInfo{
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		InContainer: detectContainer(),
		IsElevated:  isElevated(),
	}
}

// detectContainer checks if running in a container
func detectContainer() bool {
	// Check for /.dockerenv
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for Kubernetes
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	// Check container env vars
	if os.Getenv("DOCKER_CONTAINER") != "" || os.Getenv("container") != "" {
		return true
	}

	return false
}

// isElevated checks if running with elevated privileges
func isElevated() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return os.Geteuid() == 0
	case "windows":
		// TODO: Windows elevation check
		return false
	}
	return false
}

// TryElevate attempts to elevate privileges per CLAUDE.md Privilege Escalation Strategy
// Returns (true, nil) if re-executed with elevated privileges
// Returns (false, nil) if cannot elevate (will run in user mode)
// Returns (false, err) if error occurred
func TryElevate(osType string) (bool, error) {
	switch osType {
	case "linux":
		return tryElevateLinux()
	case "darwin":
		return tryElevateMacOS()
	case "windows":
		return tryElevateWindows()
	default:
		return false, nil
	}
}

// tryElevateLinux per CLAUDE.md Linux (assume headless/CI/CD)
func tryElevateLinux() (bool, error) {
	// Silent check: sudo -n true
	sudoCheck := exec.Command("sudo", "-n", "true")
	if err := sudoCheck.Run(); err != nil {
		// sudo not available without password, continue in user mode
		return false, nil
	}

	// Re-execute elevated
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	args := append([]string{executable}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to re-exec with sudo: %w", err)
	}

	// Exit current process (elevated process is now running)
	os.Exit(0)
	return true, nil
}

// tryElevateMacOS per CLAUDE.md macOS (GUI environment)
func tryElevateMacOS() (bool, error) {
	// TODO: Implement native macOS admin request dialog
	// For now, same as Linux
	return tryElevateLinux()
}

// tryElevateWindows per CLAUDE.md Windows (GUI environment)
func tryElevateWindows() (bool, error) {
	// TODO: Implement Windows UAC prompt
	return false, nil
}

// CreateDirectories creates necessary directories per CLAUDE.md Directory Layout
func CreateDirectories(dataDir string, isElevated bool) error {
	// Create main data directory
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create subdirectories
	subdirs := []string{
		"repo",
		"backups",
		"logs",
	}

	for _, subdir := range subdirs {
		path := filepath.Join(dataDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", subdir, err)
		}
	}

	return nil
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
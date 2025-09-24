package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// SudoReexec re-executes the current program with sudo per CLAUDE.md Linux strategy
func SudoReexec() error {
	// Check if sudo is available without password prompt (headless/CI/CD)
	sudoCheck := exec.Command("sudo", "-n", "true")
	if err := sudoCheck.Run(); err != nil {
		return fmt.Errorf("sudo not available without password")
	}

	// Re-execute with sudo
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	args := append([]string{executable}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to re-exec with sudo: %v", err)
	}

	// Exit the current process
	os.Exit(0)
	return nil
}
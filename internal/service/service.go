package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/casapps/casnotes/internal/config"
)

// ServiceManager handles service installation per CLAUDE.md
type ServiceManager struct {
	cfg *config.Config
}

// NewServiceManager creates service manager
func NewServiceManager(cfg *config.Config) *ServiceManager {
	return &ServiceManager{cfg: cfg}
}

// Install installs the service per CLAUDE.md Service Installation
func (s *ServiceManager) Install() error {
	switch runtime.GOOS {
	case "linux":
		return s.installSystemd()
	case "darwin":
		return s.installLaunchd()
	case "windows":
		return s.installWindowsService()
	default:
		return fmt.Errorf("service installation not supported on %s", runtime.GOOS)
	}
}

// Uninstall removes the service
func (s *ServiceManager) Uninstall() error {
	switch runtime.GOOS {
	case "linux":
		return s.uninstallSystemd()
	case "darwin":
		return s.uninstallLaunchd()
	case "windows":
		return s.uninstallWindowsService()
	default:
		return fmt.Errorf("service uninstallation not supported on %s", runtime.GOOS)
	}
}

// Start starts the service
func (s *ServiceManager) Start() error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("systemctl", "start", "casnotes").Run()
	case "darwin":
		return exec.Command("launchctl", "load", "/Library/LaunchDaemons/com.casapps.casnotes.plist").Run()
	case "windows":
		return exec.Command("sc", "start", "casnotes").Run()
	default:
		return fmt.Errorf("service start not supported on %s", runtime.GOOS)
	}
}

// Stop stops the service
func (s *ServiceManager) Stop() error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("systemctl", "stop", "casnotes").Run()
	case "darwin":
		return exec.Command("launchctl", "unload", "/Library/LaunchDaemons/com.casapps.casnotes.plist").Run()
	case "windows":
		return exec.Command("sc", "stop", "casnotes").Run()
	default:
		return fmt.Errorf("service stop not supported on %s", runtime.GOOS)
	}
}

// Status checks service status
func (s *ServiceManager) Status() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("systemctl", "status", "casnotes")
	case "darwin":
		cmd = exec.Command("launchctl", "list", "com.casapps.casnotes")
	case "windows":
		cmd = exec.Command("sc", "query", "casnotes")
	default:
		return "", fmt.Errorf("service status not supported on %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// installSystemd installs systemd service per CLAUDE.md
func (s *ServiceManager) installSystemd() error {
	binaryPath := "/usr/local/bin/casnotes"

	// Create systemd unit file
	serviceContent := fmt.Sprintf(`[Unit]
Description=casnotes - Git-powered note-taking application
After=network.target

[Service]
Type=simple
User=casnotes
Group=casnotes
ExecStart=%s
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="DATA_DIR=/var/lib/casnotes"

[Install]
WantedBy=multi-user.target
`, binaryPath)

	// Write service file
	if err := os.WriteFile("/etc/systemd/system/casnotes.service", []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Create service user (UID/GID < 999)
	exec.Command("useradd", "-r", "-s", "/bin/false", "-d", "/var/lib/casnotes", "-u", "998", "casnotes").Run()

	// Create directories
	os.MkdirAll("/var/lib/casnotes", 0755)
	os.MkdirAll("/var/log/casnotes", 0755)
	exec.Command("chown", "-R", "casnotes:casnotes", "/var/lib/casnotes").Run()
	exec.Command("chown", "-R", "casnotes:casnotes", "/var/log/casnotes").Run()

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", "casnotes").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	return nil
}

// uninstallSystemd removes systemd service
func (s *ServiceManager) uninstallSystemd() error {
	exec.Command("systemctl", "stop", "casnotes").Run()
	exec.Command("systemctl", "disable", "casnotes").Run()
	os.Remove("/etc/systemd/system/casnotes.service")
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}

// installLaunchd installs macOS launchd service per CLAUDE.md
func (s *ServiceManager) installLaunchd() error {
	binaryPath := "/usr/local/bin/casnotes"

	// Create launchd plist
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.casapps.casnotes</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/Library/Logs/casnotes/stdout.log</string>
	<key>StandardErrorPath</key>
	<string>/Library/Logs/casnotes/stderr.log</string>
	<key>EnvironmentVariables</key>
	<dict>
		<key>DATA_DIR</key>
		<string>/Library/Application Support/casnotes</string>
	</dict>
</dict>
</plist>
`, binaryPath)

	// Write plist file
	plistPath := "/Library/LaunchDaemons/com.casapps.casnotes.plist"
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Create directories
	os.MkdirAll("/Library/Application Support/casnotes", 0755)
	os.MkdirAll("/Library/Logs/casnotes", 0755)

	// Load service
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load service: %w", err)
	}

	return nil
}

// uninstallLaunchd removes macOS launchd service
func (s *ServiceManager) uninstallLaunchd() error {
	plistPath := "/Library/LaunchDaemons/com.casapps.casnotes.plist"
	exec.Command("launchctl", "unload", plistPath).Run()
	os.Remove(plistPath)
	return nil
}

// installWindowsService installs Windows service per CLAUDE.md
func (s *ServiceManager) installWindowsService() error {
	binaryPath := `C:\Program Files\casnotes\casnotes.exe`

	// Create Windows service using sc command
	cmd := exec.Command("sc", "create", "casnotes",
		"binPath=", binaryPath,
		"start=", "auto",
		"DisplayName=", "casnotes",
		"obj=", "LocalSystem",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Windows service: %w", err)
	}

	// Set description
	exec.Command("sc", "description", "casnotes", "Git-powered note-taking application").Run()

	// Create directories
	os.MkdirAll(`C:\ProgramData\casnotes`, 0755)

	return nil
}

// uninstallWindowsService removes Windows service
func (s *ServiceManager) uninstallWindowsService() error {
	exec.Command("sc", "stop", "casnotes").Run()
	return exec.Command("sc", "delete", "casnotes").Run()
}

// IsInstalled checks if service is installed
func (s *ServiceManager) IsInstalled() bool {
	switch runtime.GOOS {
	case "linux":
		_, err := os.Stat("/etc/systemd/system/casnotes.service")
		return err == nil
	case "darwin":
		_, err := os.Stat("/Library/LaunchDaemons/com.casapps.casnotes.plist")
		return err == nil
	case "windows":
		cmd := exec.Command("sc", "query", "casnotes")
		return cmd.Run() == nil
	default:
		return false
	}
}

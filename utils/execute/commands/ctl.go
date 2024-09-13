package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// SystemctlService represents a systemd service file.
type SystemctlService struct {
	Name        string
	Description string
	ExecStart   string
	Restart     string
}

// NewSystemctlService creates a new SystemctlService instance.
func NewSystemctlService(name, description, execStart, restart string) *SystemctlService {
	return &SystemctlService{
		Name:        name,
		Description: description,
		ExecStart:   execStart,
		Restart:     restart,
	}
}

// GenerateServiceFile generates a systemd service file based on the SystemctlService configuration.
func (s *SystemctlService) GenerateServiceFile(tmpPath string) (string, error) {
	content := fmt.Sprintf(`[Unit]
Description=%s

[Service]
ExecStart=%s
Restart=%s

[Install]
WantedBy=multi-user.target`, s.Description, s.ExecStart, s.Restart)

	serviceFilePath := filepath.Join(tmpPath, fmt.Sprintf("%s.service", s.Name))
	err := ioutil.WriteFile(serviceFilePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write service file: %w", err)
	}

	return serviceFilePath, nil
}

// MoveServiceFile moves the generated systemd service file to a specified location.
func (s *SystemctlService) MoveServiceFile(serviceFilePath, destinationPath string) error {
	destination := filepath.Join(destinationPath, filepath.Base(serviceFilePath))
	err := os.Rename(serviceFilePath, destination)
	if err != nil {
		return fmt.Errorf("failed to move service file: %w", err)
	}
	return nil
}

func trimNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
}

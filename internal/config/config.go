package config

import (
	"os"
	"path/filepath"
)

// AppConfig holds application configuration
type AppConfig struct {
	WorkingDir string
	Editor     string
}

// NewAppConfig creates a new app configuration
func NewAppConfig() *AppConfig {
	wd, _ := os.Getwd()
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	return &AppConfig{
		WorkingDir: wd,
		Editor:     editor,
	}
}

// GetCurlmanDir returns the .curlman directory path
func (c *AppConfig) GetCurlmanDir() string {
	return filepath.Join(c.WorkingDir, ".curlman")
}

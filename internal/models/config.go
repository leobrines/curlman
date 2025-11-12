package models

import "time"

// Config represents the application configuration
type Config struct {
	SelectedEnvironmentID string    `json:"selected_environment_id"`
	SelectedCollectionID  string    `json:"selected_collection_id"`
	Editor                string    `json:"editor"` // Default: vim
	LastUsed              time.Time `json:"last_used"`
}

// NewConfig creates a new config with defaults
func NewConfig() *Config {
	return &Config{
		Editor:   "vim",
		LastUsed: time.Now(),
	}
}

// SetSelectedEnvironment sets the currently selected environment
func (c *Config) SetSelectedEnvironment(envID string) {
	c.SelectedEnvironmentID = envID
	c.LastUsed = time.Now()
}

// SetSelectedCollection sets the currently selected collection
func (c *Config) SetSelectedCollection(collID string) {
	c.SelectedCollectionID = collID
	c.LastUsed = time.Now()
}

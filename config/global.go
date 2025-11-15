package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfig represents global configuration settings
type GlobalConfig struct {
	Variables map[string]string `json:"variables"` // Global variables usable across all collections
}

// NewGlobalConfig creates a new global configuration with default values
func NewGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Variables: make(map[string]string),
	}
}

// GetGlobalConfigPath returns the path to the global config file
func GetGlobalConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	curlmanDir := filepath.Join(homeDir, ".curlman")

	// Ensure directory exists
	if err := os.MkdirAll(curlmanDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(curlmanDir, "global.json"), nil
}

// Save persists the global configuration to disk
func (gc *GlobalConfig) Save() error {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := json.MarshalIndent(gc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads the global configuration from disk
func Load() (*GlobalConfig, error) {
	path, err := GetGlobalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If file doesn't exist, return a new empty config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewGlobalConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var gc GlobalConfig
	if err := json.Unmarshal(data, &gc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure map is initialized
	if gc.Variables == nil {
		gc.Variables = make(map[string]string)
	}

	return &gc, nil
}

// SetVariable sets a global variable
func (gc *GlobalConfig) SetVariable(key, value string) {
	gc.Variables[key] = value
}

// DeleteVariable removes a global variable
func (gc *GlobalConfig) DeleteVariable(key string) {
	delete(gc.Variables, key)
}

// GetVariable retrieves a global variable value
func (gc *GlobalConfig) GetVariable(key string) (string, bool) {
	value, exists := gc.Variables[key]
	return value, exists
}

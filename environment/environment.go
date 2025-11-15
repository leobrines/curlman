package environment

import (
	"curlman/storage"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Environment represents a named set of variables
type Environment struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
}

// NewEnvironment creates a new environment with the given name
func NewEnvironment(name string) *Environment {
	return &Environment{
		Name:      name,
		Variables: make(map[string]string),
	}
}

// GetEnvironmentsDir returns the path to the environments directory
func GetEnvironmentsDir() (string, error) {
	storageDir, err := storage.GetStorageDir()
	if err != nil {
		return "", err
	}

	envDir := filepath.Join(storageDir, "environments")

	// Create environments directory if it doesn't exist
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create environments directory: %w", err)
	}

	return envDir, nil
}

// GetEnvironmentPath returns the full path for an environment file
func GetEnvironmentPath(name string) (string, error) {
	envDir, err := GetEnvironmentsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(envDir, name+".json"), nil
}

// Save saves the environment to a file
func (e *Environment) Save() error {
	envPath, err := GetEnvironmentPath(e.Name)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment: %w", err)
	}

	if err := os.WriteFile(envPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	return nil
}

// Load loads an environment from a file
func Load(name string) (*Environment, error) {
	envPath, err := GetEnvironmentPath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file: %w", err)
	}

	var env Environment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment: %w", err)
	}

	return &env, nil
}

// List returns all available environment names
func List() ([]string, error) {
	envDir, err := GetEnvironmentsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			name := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			environments = append(environments, name)
		}
	}

	return environments, nil
}

// Delete removes an environment file
func Delete(name string) error {
	envPath, err := GetEnvironmentPath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(envPath); err != nil {
		return fmt.Errorf("failed to delete environment file: %w", err)
	}

	return nil
}

// Exists checks if an environment exists
func Exists(name string) bool {
	envPath, err := GetEnvironmentPath(name)
	if err != nil {
		return false
	}

	_, err = os.Stat(envPath)
	return err == nil
}

// Clone creates a copy of an environment with a new name
func (e *Environment) Clone(newName string) *Environment {
	clone := NewEnvironment(newName)
	for k, v := range e.Variables {
		clone.Variables[k] = v
	}
	return clone
}

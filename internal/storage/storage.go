package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leit0/curlman/internal/models"
)

const (
	CurlmanDir       = ".curlman"
	EnvironmentsDir  = "environments"
	CollectionsDir   = "collections"
	RequestsDir      = "requests"
	ConfigFile       = "config.json"
	DefaultEnvName   = "None"
)

// Storage handles all file system operations for curlman
type Storage struct {
	rootDir string
}

// NewStorage creates a new storage instance
func NewStorage(rootDir string) *Storage {
	if rootDir == "" {
		rootDir = "."
	}
	return &Storage{
		rootDir: rootDir,
	}
}

// GetCurlmanPath returns the path to the .curlman directory
func (s *Storage) GetCurlmanPath() string {
	return filepath.Join(s.rootDir, CurlmanDir)
}

// IsInitialized checks if .curlman directory exists
func (s *Storage) IsInitialized() bool {
	info, err := os.Stat(s.GetCurlmanPath())
	return err == nil && info.IsDir()
}

// Initialize creates the .curlman directory structure
func (s *Storage) Initialize() error {
	basePath := s.GetCurlmanPath()

	// Create main directory
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create .curlman directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{EnvironmentsDir, CollectionsDir, RequestsDir}
	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	// Create default "None" environment
	defaultEnv := models.NewEnvironment(DefaultEnvName)
	if err := s.SaveEnvironment(defaultEnv); err != nil {
		return fmt.Errorf("failed to create default environment: %w", err)
	}

	// Create default config
	config := models.NewConfig()
	config.SelectedEnvironmentID = defaultEnv.ID
	if err := s.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	return nil
}

// SaveConfig saves the application configuration
func (s *Storage) SaveConfig(config *models.Config) error {
	path := filepath.Join(s.GetCurlmanPath(), ConfigFile)
	return saveJSON(path, config)
}

// LoadConfig loads the application configuration
func (s *Storage) LoadConfig() (*models.Config, error) {
	path := filepath.Join(s.GetCurlmanPath(), ConfigFile)
	var config models.Config
	if err := loadJSON(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// saveJSON saves data as JSON to a file
func saveJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// loadJSON loads JSON data from a file
func loadJSON(path string, target interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

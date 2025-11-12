package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leit0/curlman/internal/models"
)

// SaveEnvironment saves an environment to disk
func (s *Storage) SaveEnvironment(env *models.Environment) error {
	path := filepath.Join(s.GetCurlmanPath(), EnvironmentsDir, fmt.Sprintf("%s.json", env.ID))
	return saveJSON(path, env)
}

// LoadEnvironment loads an environment by ID
func (s *Storage) LoadEnvironment(id string) (*models.Environment, error) {
	path := filepath.Join(s.GetCurlmanPath(), EnvironmentsDir, fmt.Sprintf("%s.json", id))
	var env models.Environment
	if err := loadJSON(path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// ListEnvironments returns all environments
func (s *Storage) ListEnvironments() ([]*models.Environment, error) {
	envDir := filepath.Join(s.GetCurlmanPath(), EnvironmentsDir)
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []*models.Environment
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		env, err := s.LoadEnvironment(id)
		if err != nil {
			// Skip invalid environment files
			continue
		}
		environments = append(environments, env)
	}

	return environments, nil
}

// DeleteEnvironment deletes an environment by ID
func (s *Storage) DeleteEnvironment(id string) error {
	path := filepath.Join(s.GetCurlmanPath(), EnvironmentsDir, fmt.Sprintf("%s.json", id))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	return nil
}

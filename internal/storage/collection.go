package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leit0/curlman/internal/models"
)

// SaveCollection saves a collection to disk
func (s *Storage) SaveCollection(coll *models.Collection) error {
	path := filepath.Join(s.GetCurlmanPath(), CollectionsDir, fmt.Sprintf("%s.json", coll.ID))
	return saveJSON(path, coll)
}

// LoadCollection loads a collection by ID
func (s *Storage) LoadCollection(id string) (*models.Collection, error) {
	path := filepath.Join(s.GetCurlmanPath(), CollectionsDir, fmt.Sprintf("%s.json", id))
	var coll models.Collection
	if err := loadJSON(path, &coll); err != nil {
		return nil, err
	}
	return &coll, nil
}

// ListCollections returns all collections
func (s *Storage) ListCollections() ([]*models.Collection, error) {
	collDir := filepath.Join(s.GetCurlmanPath(), CollectionsDir)
	entries, err := os.ReadDir(collDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read collections directory: %w", err)
	}

	var collections []*models.Collection
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		coll, err := s.LoadCollection(id)
		if err != nil {
			// Skip invalid collection files
			continue
		}
		collections = append(collections, coll)
	}

	return collections, nil
}

// DeleteCollection deletes a collection by ID
func (s *Storage) DeleteCollection(id string) error {
	path := filepath.Join(s.GetCurlmanPath(), CollectionsDir, fmt.Sprintf("%s.json", id))
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

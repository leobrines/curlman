package services

import (
	"github.com/leobrines/curlman/models"
	"github.com/leobrines/curlman/openapi"
	"github.com/leobrines/curlman/storage"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CollectionService handles all collection-related business logic
type CollectionService struct{}

// NewCollectionService creates a new collection service
func NewCollectionService() *CollectionService {
	return &CollectionService{}
}

// CreateEmptyCollection creates a new empty collection with default values
func (s *CollectionService) CreateEmptyCollection() *models.Collection {
	return &models.Collection{
		Name:              "New Collection",
		Requests:          []*models.Request{},
		Variables:         make(map[string]string),
		Environments:      []models.CollectionEnvironment{},
		ActiveCollectionEnv: "",
		ActiveEnvironment: "",
		EnvironmentVars:   make(map[string]string),
		CollectionEnvVars: make(map[string]string),
	}
}

// ImportFromOpenAPI imports a collection from an OpenAPI file and auto-saves it
func (s *CollectionService) ImportFromOpenAPI(filePath string) (*models.Collection, string, error) {
	if filePath == "" {
		return nil, "", fmt.Errorf("file path cannot be empty")
	}

	collection, err := openapi.ImportFromFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to import OpenAPI file: %w", err)
	}

	if len(collection.Requests) == 0 {
		return nil, "", fmt.Errorf("imported collection contains no requests")
	}

	// Auto-save the collection with the OpenAPI title as filename
	fileName := sanitizeFileName(collection.Name)
	if fileName == "" {
		fileName = "imported-collection"
	}

	savedPath, err := s.SaveCollection(collection, fileName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to auto-save collection: %w", err)
	}

	return collection, savedPath, nil
}

// sanitizeFileName converts a string into a valid filename
func sanitizeFileName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove or replace invalid filename characters
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = reg.ReplaceAllString(name, "")

	// Convert to lowercase for consistency
	name = strings.ToLower(name)

	// Remove leading/trailing hyphens and dots
	name = strings.Trim(name, "-.")

	// Collapse multiple hyphens into one
	reg = regexp.MustCompile(`-+`)
	name = reg.ReplaceAllString(name, "-")

	return name
}

// SaveCollection saves a collection to the storage directory
func (s *CollectionService) SaveCollection(collection *models.Collection, fileName string) (string, error) {
	if collection == nil {
		return "", fmt.Errorf("collection cannot be nil")
	}

	if fileName == "" {
		return "", fmt.Errorf("file name cannot be empty")
	}

	// Ensure .json extension
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	// Get storage directory
	storageDir, err := storage.GetStorageDir()
	if err != nil {
		return "", fmt.Errorf("failed to get storage directory: %w", err)
	}

	// Create full path
	fullPath := filepath.Join(storageDir, fileName)

	// Save collection
	err = openapi.SaveCollection(collection, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to save collection: %w", err)
	}

	return fullPath, nil
}

// LoadCollection loads a collection from the storage directory
func (s *CollectionService) LoadCollection(fileName string) (*models.Collection, error) {
	if fileName == "" {
		return nil, fmt.Errorf("file name cannot be empty")
	}

	// Ensure .json extension
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	// Get storage directory
	storageDir, err := storage.GetStorageDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage directory: %w", err)
	}

	// Create full path
	fullPath := filepath.Join(storageDir, fileName)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("collection file does not exist: %s", fileName)
	}

	// Load collection
	collection, err := openapi.LoadCollection(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	return collection, nil
}

// ListCollections returns a list of all saved collections in the storage directory
func (s *CollectionService) ListCollections() ([]string, error) {
	storageDir, err := storage.GetStorageDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage directory: %w", err)
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var collections []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			collections = append(collections, entry.Name())
		}
	}

	return collections, nil
}

// ValidateCollection validates a collection's data
func (s *CollectionService) ValidateCollection(collection *models.Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	if collection.Name == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	// Validate all requests
	for i, req := range collection.Requests {
		if req.Name == "" {
			return fmt.Errorf("request %d has empty name", i)
		}
		if req.Method == "" {
			return fmt.Errorf("request '%s' has empty method", req.Name)
		}
		if req.URL == "" {
			return fmt.Errorf("request '%s' has empty URL", req.Name)
		}
	}

	return nil
}

// GetCollectionStats returns statistics about the collection
func (s *CollectionService) GetCollectionStats(collection *models.Collection) map[string]interface{} {
	if collection == nil {
		return map[string]interface{}{
			"requests":                0,
			"variables":               0,
			"collection_environments": 0,
		}
	}

	return map[string]interface{}{
		"requests":                len(collection.Requests),
		"variables":               len(collection.Variables),
		"collection_environments": len(collection.Environments),
		"active_environment":      collection.ActiveEnvironment,
		"active_collection_env":   collection.ActiveCollectionEnv,
	}
}

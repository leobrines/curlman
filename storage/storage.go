package storage

import (
	"os"
	"path/filepath"
)

// GetStorageDir returns the curlman storage directory path
// It will create the directory if it doesn't exist
func GetStorageDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	storageDir := filepath.Join(homeDir, ".curlman")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", err
	}

	return storageDir, nil
}

// GetFilePath returns the full path for a file in the storage directory
func GetFilePath(filename string) (string, error) {
	storageDir, err := GetStorageDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(storageDir, filename), nil
}

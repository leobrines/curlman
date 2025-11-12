package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leit0/curlman/internal/models"
)

// SaveRequest saves a request's curl command to a file
func (s *Storage) SaveRequest(req *models.Request) error {
	// Save request metadata within its collection
	coll, err := s.LoadCollection(req.CollectionID)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Update or add request to collection
	found := false
	for i := range coll.Requests {
		if coll.Requests[i].ID == req.ID {
			coll.Requests[i] = *req
			found = true
			break
		}
	}
	if !found {
		coll.Requests = append(coll.Requests, *req)
	}

	// Save the curl command to a file
	reqDir := filepath.Join(s.GetCurlmanPath(), RequestsDir, req.CollectionID)
	if err := os.MkdirAll(reqDir, 0755); err != nil {
		return fmt.Errorf("failed to create request directory: %w", err)
	}

	curlFilePath := filepath.Join(reqDir, fmt.Sprintf("%s.curl", req.ID))
	if err := os.WriteFile(curlFilePath, []byte(req.CurlCommand), 0644); err != nil {
		return fmt.Errorf("failed to write curl file: %w", err)
	}
	req.FilePath = curlFilePath

	// Save updated collection
	return s.SaveCollection(coll)
}

// LoadRequest loads a request by ID from a collection
func (s *Storage) LoadRequest(collectionID, requestID string) (*models.Request, error) {
	coll, err := s.LoadCollection(collectionID)
	if err != nil {
		return nil, err
	}

	for i := range coll.Requests {
		if coll.Requests[i].ID == requestID {
			// Load curl command from file
			req := &coll.Requests[i]
			curlFilePath := filepath.Join(s.GetCurlmanPath(), RequestsDir, collectionID, fmt.Sprintf("%s.curl", requestID))
			curlBytes, err := os.ReadFile(curlFilePath)
			if err == nil {
				req.CurlCommand = string(curlBytes)
				req.FilePath = curlFilePath
			}
			return req, nil
		}
	}

	return nil, fmt.Errorf("request not found")
}

// DeleteRequest deletes a request by ID
func (s *Storage) DeleteRequest(collectionID, requestID string) error {
	coll, err := s.LoadCollection(collectionID)
	if err != nil {
		return err
	}

	// Remove from collection
	found := false
	for i := range coll.Requests {
		if coll.Requests[i].ID == requestID {
			coll.Requests = append(coll.Requests[:i], coll.Requests[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("request not found")
	}

	// Delete curl file
	curlFilePath := filepath.Join(s.GetCurlmanPath(), RequestsDir, collectionID, fmt.Sprintf("%s.curl", requestID))
	if err := os.Remove(curlFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete curl file: %w", err)
	}

	// Save updated collection
	return s.SaveCollection(coll)
}

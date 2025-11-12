package models

import "time"

// Request represents a single API request (stored as a curl command)
type Request struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	CurlCommand  string            `json:"curl_command"`
	FilePath     string            `json:"file_path"` // Path to the curl file
	CollectionID string            `json:"collection_id"`
	Method       string            `json:"method"`       // GET, POST, etc.
	URL          string            `json:"url"`          // Extracted from curl
	Headers      map[string]string `json:"headers"`      // Extracted headers
	Body         string            `json:"body"`         // Request body
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`

	// OpenAPI integration fields
	IsManaged         bool   `json:"is_managed"`          // false = spec request (ephemeral), true = managed request (persistent)
	OpenAPIOperation  string `json:"openapi_operation"`   // Operation ID (e.g., "GET /users/{id}")
	OperationExists   bool   `json:"operation_exists"`    // Whether the linked OpenAPI operation still exists
}

// NewRequest creates a new managed request with a curl command
func NewRequest(name, curlCommand, collectionID string) *Request {
	now := time.Now()
	return &Request{
		ID:              generateID(),
		Name:            name,
		CurlCommand:     curlCommand,
		CollectionID:    collectionID,
		Headers:         make(map[string]string),
		CreatedAt:       now,
		UpdatedAt:       now,
		IsManaged:       true, // New requests are managed by default
		OperationExists: true, // Default to true for non-OpenAPI requests
	}
}

// UpdateCurlCommand updates the curl command and timestamp
func (r *Request) UpdateCurlCommand(curlCommand string) {
	r.CurlCommand = curlCommand
	r.UpdatedAt = time.Now()
}

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

	// OpenAPI spec details for display
	Tags              []string                   `json:"tags"`                // Operation tags
	Parameters        []ParameterDetail          `json:"parameters"`          // Path, query, header parameters
	RequestBodySchema *RequestBodyDetail         `json:"request_body_schema"` // Request body schema info
	Responses         map[string]ResponseDetail  `json:"responses"`           // Response schemas by status code
	Security          []SecurityRequirement      `json:"security"`            // Security requirements
	Deprecated        bool                       `json:"deprecated"`          // Whether the operation is deprecated
}

// ParameterDetail represents details about an operation parameter
type ParameterDetail struct {
	Name        string `json:"name"`
	In          string `json:"in"`          // "path", "query", "header", "cookie"
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`        // e.g., "string", "integer", "boolean"
	Example     string `json:"example"`     // Example value
	Default     string `json:"default"`     // Default value
}

// RequestBodyDetail represents request body schema information
type RequestBodyDetail struct {
	Description string `json:"description"`
	ContentType string `json:"content_type"` // e.g., "application/json"
	Required    bool   `json:"required"`
	Schema      string `json:"schema"`       // Brief schema description
}

// ResponseDetail represents response schema information
type ResponseDetail struct {
	Description string `json:"description"`
	ContentType string `json:"content_type"`
	Schema      string `json:"schema"` // Brief schema description
}

// SecurityRequirement represents a security requirement
type SecurityRequirement struct {
	Type        string   `json:"type"`        // e.g., "apiKey", "http", "oauth2"
	Name        string   `json:"name"`        // Security scheme name
	In          string   `json:"in"`          // For apiKey: "header", "query", "cookie"
	Scheme      string   `json:"scheme"`      // For http: "basic", "bearer"
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`      // For oauth2
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

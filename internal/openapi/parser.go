package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/leit0/curlman/internal/models"
	"gopkg.in/yaml.v3"
)

// OpenAPIParser parses OpenAPI specification files
type OpenAPIParser struct{}

// NewOpenAPIParser creates a new OpenAPI parser
func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{}
}

// OpenAPISpec represents a simplified OpenAPI specification
type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi" yaml:"openapi"`
	Info    map[string]interface{} `json:"info" yaml:"info"`
	Servers []Server               `json:"servers" yaml:"servers"`
	Paths   map[string]PathItem    `json:"paths" yaml:"paths"`
}

// Server represents an OpenAPI server
type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
}

// PathItem represents operations for a path
type PathItem struct {
	Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *Operation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Summary     string                 `json:"summary" yaml:"summary"`
	Description string                 `json:"description" yaml:"description"`
	OperationID string                 `json:"operationId" yaml:"operationId"`
	Parameters  []Parameter            `json:"parameters" yaml:"parameters"`
	RequestBody map[string]interface{} `json:"requestBody" yaml:"requestBody"`
}

// Parameter represents an operation parameter
type Parameter struct {
	Name        string `json:"name" yaml:"name"`
	In          string `json:"in" yaml:"in"` // query, header, path, cookie
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description" yaml:"description"`
}

// ParseFile parses an OpenAPI file and returns a collection (metadata only, no requests)
func (p *OpenAPIParser) ParseFile(filePath string) (*models.Collection, error) {
	spec, err := p.parseSpec(filePath)
	if err != nil {
		return nil, err
	}

	// Create collection metadata only (no requests)
	collection := p.createCollectionMetadata(spec, filePath)
	return collection, nil
}

// ParseFileToSpecRequests parses an OpenAPI file and returns spec requests (ephemeral, in-memory only)
func (p *OpenAPIParser) ParseFileToSpecRequests(filePath string) ([]models.Request, error) {
	spec, err := p.parseSpec(filePath)
	if err != nil {
		return nil, err
	}

	return p.generateSpecRequests(spec), nil
}

// parseSpec parses the OpenAPI specification file
func (p *OpenAPIParser) parseSpec(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var spec OpenAPISpec

	// Try JSON first, then YAML
	if err := json.Unmarshal(data, &spec); err != nil {
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI file: %w", err)
		}
	}

	return &spec, nil
}

// createCollectionMetadata creates collection metadata without requests
func (p *OpenAPIParser) createCollectionMetadata(spec *OpenAPISpec, filePath string) *models.Collection {
	name := "OpenAPI Collection"
	description := ""

	if info, ok := spec.Info["title"].(string); ok {
		name = info
	}
	if desc, ok := spec.Info["description"].(string); ok {
		description = desc
	}

	coll := models.NewCollection(name, description)
	coll.OpenAPIPath = filePath

	return coll
}

// generateSpecRequests generates spec requests from OpenAPI spec
func (p *OpenAPIParser) generateSpecRequests(spec *OpenAPISpec) []models.Request {
	var requests []models.Request

	// Get base URL from servers
	baseURL := ""
	if len(spec.Servers) > 0 {
		baseURL = spec.Servers[0].URL
	}

	// Convert paths to spec requests
	for path, pathItem := range spec.Paths {
		reqs := p.createSpecRequestsFromPath(baseURL, path, &pathItem)
		requests = append(requests, reqs...)
	}

	return requests
}

// createSpecRequestsFromPath creates spec requests from a path item
func (p *OpenAPIParser) createSpecRequestsFromPath(baseURL, path string, pathItem *PathItem) []models.Request {
	operations := map[string]*Operation{
		"GET":    pathItem.Get,
		"POST":   pathItem.Post,
		"PUT":    pathItem.Put,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
	}

	var requests []models.Request

	for method, op := range operations {
		if op == nil {
			continue
		}

		// Build curl command
		curlCmd := p.buildCurlCommand(method, baseURL+path, op)

		// Create spec request (not managed)
		requestName := op.Summary
		if requestName == "" {
			requestName = fmt.Sprintf("%s %s", method, path)
		}

		req := models.NewRequest(requestName, curlCmd, "")
		req.Description = op.Description
		req.Method = method
		req.URL = baseURL + path

		// Mark as spec request (ephemeral)
		req.IsManaged = false
		req.OpenAPIOperation = fmt.Sprintf("%s %s", method, path)
		req.OperationExists = true

		requests = append(requests, *req)
	}

	return requests
}

// ValidateOperation checks if an operation exists in the OpenAPI spec
func (p *OpenAPIParser) ValidateOperation(filePath, operation string) (bool, error) {
	spec, err := p.parseSpec(filePath)
	if err != nil {
		return false, err
	}

	// Parse operation string (e.g., "GET /users/{id}")
	parts := strings.SplitN(operation, " ", 2)
	if len(parts) != 2 {
		return false, nil
	}

	method := parts[0]
	path := parts[1]

	// Check if path exists
	pathItem, exists := spec.Paths[path]
	if !exists {
		return false, nil
	}

	// Check if method exists for this path
	switch strings.ToUpper(method) {
	case "GET":
		return pathItem.Get != nil, nil
	case "POST":
		return pathItem.Post != nil, nil
	case "PUT":
		return pathItem.Put != nil, nil
	case "DELETE":
		return pathItem.Delete != nil, nil
	case "PATCH":
		return pathItem.Patch != nil, nil
	default:
		return false, nil
	}
}

// buildCurlCommand builds a curl command from operation details
func (p *OpenAPIParser) buildCurlCommand(method, url string, op *Operation) string {
	var parts []string
	parts = append(parts, "curl")

	// Add method if not GET
	if method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", method))
	}

	// Add URL
	parts = append(parts, fmt.Sprintf(`"%s"`, url))

	// Add headers for parameters
	for _, param := range op.Parameters {
		if param.In == "header" {
			parts = append(parts, fmt.Sprintf(`-H "%s: {{.%s}}"`, param.Name, param.Name))
		}
	}

	// Add content-type header if request body exists
	if op.RequestBody != nil {
		parts = append(parts, `-H "Content-Type: application/json"`)
		parts = append(parts, `-d '{{.RequestBody}}'`)
	}

	return strings.Join(parts, " ")
}

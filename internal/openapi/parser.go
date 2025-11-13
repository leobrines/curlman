package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/leit0/curlman/internal/models"
)

// OpenAPIParser parses OpenAPI specification files
type OpenAPIParser struct{}

// NewOpenAPIParser creates a new OpenAPI parser
func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{}
}

// ParseFile parses an OpenAPI file and returns a collection (metadata only, no requests)
func (p *OpenAPIParser) ParseFile(filePath string) (*models.Collection, error) {
	doc, err := p.loadSpec(filePath)
	if err != nil {
		return nil, err
	}

	// Create collection metadata only (no requests)
	collection := p.createCollectionMetadata(doc, filePath)
	return collection, nil
}

// ParseFileToSpecRequests parses an OpenAPI file and returns spec requests (ephemeral, in-memory only)
func (p *OpenAPIParser) ParseFileToSpecRequests(filePath string) ([]models.Request, error) {
	doc, err := p.loadSpec(filePath)
	if err != nil {
		return nil, err
	}

	return p.generateSpecRequests(doc), nil
}

// loadSpec loads and parses an OpenAPI specification file
// Handles OpenAPI 2.0 (Swagger) to 3.0 conversion automatically
func (p *OpenAPIParser) loadSpec(filePath string) (*openapi3.T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check version to determine if it's Swagger 2.0 or OpenAPI 3.x
	var versionCheck struct {
		Swagger string `json:"swagger" yaml:"swagger"`
		OpenAPI string `json:"openapi" yaml:"openapi"`
	}

	// Try JSON first
	if err := json.Unmarshal(data, &versionCheck); err != nil {
		// If JSON fails, the loader will handle YAML
		versionCheck = struct {
			Swagger string `json:"swagger" yaml:"swagger"`
			OpenAPI string `json:"openapi" yaml:"openapi"`
		}{}
	}

	// Check if this is Swagger 2.0
	if versionCheck.Swagger != "" {
		return p.convertSwaggerToOpenAPI3(filePath, data)
	}

	// Load as OpenAPI 3.x
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI file: %w", err)
	}

	// Validate OpenAPI version
	if doc.OpenAPI == "" {
		return nil, fmt.Errorf("invalid OpenAPI specification: missing version")
	}

	// Check version >= 3.0
	if !strings.HasPrefix(doc.OpenAPI, "3.") {
		return nil, fmt.Errorf("unsupported OpenAPI version %s: only version 3.0 and above are supported", doc.OpenAPI)
	}

	// Validate the document
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI specification: %w", err)
	}

	return doc, nil
}

// convertSwaggerToOpenAPI3 converts an OpenAPI 2.0 (Swagger) file to OpenAPI 3.0
// and saves it to a temporary file, then loads it
func (p *OpenAPIParser) convertSwaggerToOpenAPI3(originalPath string, data []byte) (*openapi3.T, error) {
	// Parse as Swagger 2.0 using ReadFromURIFunc
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Parse the Swagger 2.0 document
	var swagger2 openapi2.T
	if err := json.Unmarshal(data, &swagger2); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger 2.0 file as JSON: %w", err)
	}

	// Convert to OpenAPI 3.0
	doc, err := openapi2conv.ToV3(&swagger2)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Swagger 2.0 to OpenAPI 3.0: %w", err)
	}

	// Create a temporary file for the converted spec
	tmpDir := os.TempDir()
	baseName := filepath.Base(originalPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	tmpFile, err := os.CreateTemp(tmpDir, fmt.Sprintf("curlman-openapi3-%s-*.json", nameWithoutExt))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	// Marshal the converted spec to JSON
	convertedData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal converted OpenAPI 3.0 spec: %w", err)
	}

	// Write to temporary file
	if _, err := tmpFile.Write(convertedData); err != nil {
		return nil, fmt.Errorf("failed to write converted spec to temporary file: %w", err)
	}

	// Validate the converted document
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("converted OpenAPI 3.0 specification is invalid: %w", err)
	}

	return doc, nil
}

// createCollectionMetadata creates collection metadata without requests
func (p *OpenAPIParser) createCollectionMetadata(doc *openapi3.T, filePath string) *models.Collection {
	name := "OpenAPI Collection"
	description := ""

	if doc.Info != nil {
		if doc.Info.Title != "" {
			name = doc.Info.Title
		}
		if doc.Info.Description != "" {
			description = doc.Info.Description
		}
	}

	coll := models.NewCollection(name, description)
	coll.OpenAPIPath = filePath

	return coll
}

// generateSpecRequests generates spec requests from OpenAPI spec
func (p *OpenAPIParser) generateSpecRequests(doc *openapi3.T) []models.Request {
	var requests []models.Request

	// Get base URL from servers
	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = doc.Servers[0].URL
	}

	// Convert paths to spec requests
	for path, pathItem := range doc.Paths.Map() {
		reqs := p.createSpecRequestsFromPath(baseURL, path, pathItem)
		requests = append(requests, reqs...)
	}

	return requests
}

// createSpecRequestsFromPath creates spec requests from a path item
func (p *OpenAPIParser) createSpecRequestsFromPath(baseURL, path string, pathItem *openapi3.PathItem) []models.Request {
	operations := map[string]*openapi3.Operation{
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
		curlCmd := p.buildCurlCommand(method, baseURL+path, op, pathItem)

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

		// Extract additional spec details
		req.Tags = op.Tags
		req.Deprecated = op.Deprecated
		req.Parameters = p.extractParameters(op, pathItem)
		req.RequestBodySchema = p.extractRequestBodySchema(op)
		req.Responses = p.extractResponses(op)
		req.Security = p.extractSecurity(op)

		requests = append(requests, *req)
	}

	return requests
}

// buildCurlCommand builds a curl command from operation details
// Uses examples and defaults from the OpenAPI specification
func (p *OpenAPIParser) buildCurlCommand(method, urlStr string, op *openapi3.Operation, pathItem *openapi3.PathItem) string {
	var parts []string
	parts = append(parts, "curl")

	// Add method if not GET
	if method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", method))
	}

	// Combine parameters from both path and operation level
	allParams := make([]*openapi3.ParameterRef, 0)
	if pathItem.Parameters != nil {
		allParams = append(allParams, pathItem.Parameters...)
	}
	if op.Parameters != nil {
		allParams = append(allParams, op.Parameters...)
	}

	// Build URL with path and query parameters
	finalURL := urlStr
	queryParams := url.Values{}

	for _, paramRef := range allParams {
		if paramRef.Value == nil {
			continue
		}
		param := paramRef.Value
		value := p.extractParameterValue(param)

		switch param.In {
		case "path":
			// Replace path parameter placeholder
			placeholder := fmt.Sprintf("{%s}", param.Name)
			finalURL = strings.ReplaceAll(finalURL, placeholder, value)
		case "query":
			// Add query parameter
			queryParams.Add(param.Name, value)
		}
	}

	// Add query parameters to URL
	if len(queryParams) > 0 {
		if strings.Contains(finalURL, "?") {
			finalURL += "&" + queryParams.Encode()
		} else {
			finalURL += "?" + queryParams.Encode()
		}
	}

	parts = append(parts, fmt.Sprintf(`"%s"`, finalURL))

	// Add headers
	for _, paramRef := range allParams {
		if paramRef.Value == nil {
			continue
		}
		param := paramRef.Value
		if param.In == "header" {
			value := p.extractParameterValue(param)
			parts = append(parts, fmt.Sprintf(`-H "%s: %s"`, param.Name, value))
		}
	}

	// Add request body if present
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		bodyContent := p.extractRequestBody(op.RequestBody.Value)
		if bodyContent != "" {
			parts = append(parts, `-H "Content-Type: application/json"`)
			parts = append(parts, fmt.Sprintf(`-d '%s'`, bodyContent))
		}
	}

	return strings.Join(parts, " ")
}

// extractParameterValue extracts a value for a parameter
// Priority: Example > Schema.Default > Type-based placeholder
func (p *OpenAPIParser) extractParameterValue(param *openapi3.Parameter) string {
	// Check for example value first
	if param.Example != nil {
		return fmt.Sprintf("%v", param.Example)
	}

	// Check schema for default or example
	if param.Schema != nil && param.Schema.Value != nil {
		schema := param.Schema.Value

		// Check schema example
		if schema.Example != nil {
			return fmt.Sprintf("%v", schema.Example)
		}

		// Check schema default
		if schema.Default != nil {
			return fmt.Sprintf("%v", schema.Default)
		}

		// Generate placeholder based on type
		switch schema.Type.Slice()[0] {
		case "string":
			if len(schema.Enum) > 0 {
				return fmt.Sprintf("%v", schema.Enum[0])
			}
			return fmt.Sprintf("{%s}", param.Name)
		case "integer", "number":
			return "0"
		case "boolean":
			return "false"
		}
	}

	// Fallback to placeholder
	return fmt.Sprintf("{%s}", param.Name)
}

// extractRequestBody extracts a request body from examples
// Returns JSON string or empty string if no examples found
func (p *OpenAPIParser) extractRequestBody(requestBody *openapi3.RequestBody) string {
	// Check for application/json content
	if content, ok := requestBody.Content["application/json"]; ok && content != nil {
		// Check for example
		if content.Example != nil {
			data, err := json.Marshal(content.Example)
			if err == nil {
				return string(data)
			}
		}

		// Check for examples (map)
		if len(content.Examples) > 0 {
			// Get the first example
			for _, exampleRef := range content.Examples {
				if exampleRef.Value != nil && exampleRef.Value.Value != nil {
					data, err := json.Marshal(exampleRef.Value.Value)
					if err == nil {
						return string(data)
					}
				}
			}
		}

		// Check schema for example
		if content.Schema != nil && content.Schema.Value != nil {
			schema := content.Schema.Value
			if schema.Example != nil {
				data, err := json.Marshal(schema.Example)
				if err == nil {
					return string(data)
				}
			}
		}
	}

	return ""
}

// extractParameters extracts parameter details from operation and path item
func (p *OpenAPIParser) extractParameters(op *openapi3.Operation, pathItem *openapi3.PathItem) []models.ParameterDetail {
	var params []models.ParameterDetail

	// Combine parameters from both path and operation level
	allParams := make([]*openapi3.ParameterRef, 0)
	if pathItem.Parameters != nil {
		allParams = append(allParams, pathItem.Parameters...)
	}
	if op.Parameters != nil {
		allParams = append(allParams, op.Parameters...)
	}

	for _, paramRef := range allParams {
		if paramRef.Value == nil {
			continue
		}
		param := paramRef.Value

		paramDetail := models.ParameterDetail{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
			Required:    param.Required,
		}

		// Extract type and example from schema
		if param.Schema != nil && param.Schema.Value != nil {
			schema := param.Schema.Value
			if len(schema.Type.Slice()) > 0 {
				paramDetail.Type = schema.Type.Slice()[0]
			}
			if schema.Example != nil {
				paramDetail.Example = fmt.Sprintf("%v", schema.Example)
			}
			if schema.Default != nil {
				paramDetail.Default = fmt.Sprintf("%v", schema.Default)
			}
		}

		// Check for parameter-level example
		if param.Example != nil {
			paramDetail.Example = fmt.Sprintf("%v", param.Example)
		}

		params = append(params, paramDetail)
	}

	return params
}

// extractRequestBodySchema extracts request body schema details
func (p *OpenAPIParser) extractRequestBodySchema(op *openapi3.Operation) *models.RequestBodyDetail {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}

	rb := op.RequestBody.Value
	detail := &models.RequestBodyDetail{
		Description: rb.Description,
		Required:    rb.Required,
	}

	// Check for application/json content
	if content, ok := rb.Content["application/json"]; ok && content != nil {
		detail.ContentType = "application/json"

		// Extract schema description
		if content.Schema != nil && content.Schema.Value != nil {
			schema := content.Schema.Value

			// Build a brief schema description
			if schema.Type.Is("object") {
				detail.Schema = fmt.Sprintf("Object with %d properties", len(schema.Properties))
			} else if schema.Type.Is("array") {
				detail.Schema = "Array"
				if schema.Items != nil && schema.Items.Value != nil {
					itemType := "unknown"
					if len(schema.Items.Value.Type.Slice()) > 0 {
						itemType = schema.Items.Value.Type.Slice()[0]
					}
					detail.Schema = fmt.Sprintf("Array of %s", itemType)
				}
			} else if len(schema.Type.Slice()) > 0 {
				detail.Schema = schema.Type.Slice()[0]
			}
		}
	}

	return detail
}

// extractResponses extracts response details
func (p *OpenAPIParser) extractResponses(op *openapi3.Operation) map[string]models.ResponseDetail {
	responses := make(map[string]models.ResponseDetail)

	if op.Responses == nil {
		return responses
	}

	for statusCode, respRef := range op.Responses.Map() {
		if respRef.Value == nil {
			continue
		}
		resp := respRef.Value

		detail := models.ResponseDetail{}
		if resp.Description != nil {
			detail.Description = *resp.Description
		}

		// Check for application/json content
		if content, ok := resp.Content["application/json"]; ok && content != nil {
			detail.ContentType = "application/json"

			// Extract schema description
			if content.Schema != nil && content.Schema.Value != nil {
				schema := content.Schema.Value

				// Build a brief schema description
				if schema.Type.Is("object") {
					detail.Schema = fmt.Sprintf("Object with %d properties", len(schema.Properties))
				} else if schema.Type.Is("array") {
					detail.Schema = "Array"
					if schema.Items != nil && schema.Items.Value != nil {
						itemType := "unknown"
						if len(schema.Items.Value.Type.Slice()) > 0 {
							itemType = schema.Items.Value.Type.Slice()[0]
						}
						detail.Schema = fmt.Sprintf("Array of %s", itemType)
					}
				} else if len(schema.Type.Slice()) > 0 {
					detail.Schema = schema.Type.Slice()[0]
				}
			}
		}

		responses[statusCode] = detail
	}

	return responses
}

// extractSecurity extracts security requirements
func (p *OpenAPIParser) extractSecurity(op *openapi3.Operation) []models.SecurityRequirement {
	var secReqs []models.SecurityRequirement

	if op.Security == nil || len(*op.Security) == 0 {
		return secReqs
	}

	// For now, we'll return a simplified representation
	// In a full implementation, we'd need access to the SecuritySchemes
	for _, secReq := range *op.Security {
		for schemeName, scopes := range secReq {
			secReqs = append(secReqs, models.SecurityRequirement{
				Name:   schemeName,
				Scopes: scopes,
			})
		}
	}

	return secReqs
}

// ValidateOperation checks if an operation exists in the OpenAPI spec
func (p *OpenAPIParser) ValidateOperation(filePath, operation string) (bool, error) {
	doc, err := p.loadSpec(filePath)
	if err != nil {
		return false, err
	}

	// Parse operation string (e.g., "GET /users/{id}")
	parts := strings.SplitN(operation, " ", 2)
	if len(parts) != 2 {
		return false, nil
	}

	method := strings.ToLower(parts[0])
	path := parts[1]

	// Check if path exists
	pathItem := doc.Paths.Find(path)
	if pathItem == nil {
		return false, nil
	}

	// Check if method exists for this path
	switch method {
	case "get":
		return pathItem.Get != nil, nil
	case "post":
		return pathItem.Post != nil, nil
	case "put":
		return pathItem.Put != nil, nil
	case "delete":
		return pathItem.Delete != nil, nil
	case "patch":
		return pathItem.Patch != nil, nil
	default:
		return false, nil
	}
}

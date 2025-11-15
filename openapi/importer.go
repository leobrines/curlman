package openapi

import (
	"context"
	"curlman/models"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

// ImportFromFile imports an OpenAPI YAML file and creates a collection
func ImportFromFile(filepath string) (*models.Collection, error) {
	// Load the OpenAPI document
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI file: %w", err)
	}

	// Validate the document
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI document: %w", err)
	}

	return convertToCollection(doc), nil
}

// ImportFromYAML imports an OpenAPI YAML string and creates a collection
func ImportFromYAML(yamlContent []byte) (*models.Collection, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI YAML: %w", err)
	}

	// Validate the document
	if err := doc.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI document: %w", err)
	}

	return convertToCollection(doc), nil
}

// convertToCollection converts an OpenAPI document to a Collection
func convertToCollection(doc *openapi3.T) *models.Collection {
	collection := &models.Collection{
		Name:      doc.Info.Title,
		Requests:  []*models.Request{},
		Variables: make(map[string]string),
	}

	// Extract base URL from servers
	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = doc.Servers[0].URL
		// Add server variables
		for name, serverVar := range doc.Servers[0].Variables {
			if serverVar.Default != "" {
				collection.Variables[name] = serverVar.Default
			}
		}
	}

	// Iterate through all paths and operations
	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		// Process each HTTP method
		operations := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			request := convertOperation(baseURL, path, method, operation)
			collection.Requests = append(collection.Requests, request)
		}
	}

	return collection
}

// convertOperation converts an OpenAPI operation to a Request
func convertOperation(baseURL, path, method string, operation *openapi3.Operation) *models.Request {
	request := &models.Request{
		ID:          uuid.New().String(),
		Name:        operation.Summary,
		Method:      method,
		URL:         baseURL,
		Path:        path,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Description: operation.Description,
	}

	// If no summary, use operationId or generate one
	if request.Name == "" {
		if operation.OperationID != "" {
			request.Name = operation.OperationID
		} else {
			request.Name = fmt.Sprintf("%s %s", method, path)
		}
	}

	// Extract parameters
	for _, paramRef := range operation.Parameters {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}

		param := paramRef.Value
		defaultValue := ""

		// Get default value if it exists
		if param.Schema != nil && param.Schema.Value != nil && param.Schema.Value.Default != nil {
			defaultValue = fmt.Sprintf("%v", param.Schema.Value.Default)
		}

		// Example value
		if defaultValue == "" && param.Example != nil {
			defaultValue = fmt.Sprintf("%v", param.Example)
		}

		switch param.In {
		case "query":
			if defaultValue != "" {
				request.QueryParams[param.Name] = defaultValue
			} else {
				request.QueryParams[param.Name] = "{{" + param.Name + "}}"
			}
		case "header":
			if defaultValue != "" {
				request.Headers[param.Name] = defaultValue
			} else {
				request.Headers[param.Name] = "{{" + param.Name + "}}"
			}
		case "path":
			// Path parameters are already in the path string
			// We'll mark them as variables
		}
	}

	// Add common headers
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for contentType := range operation.RequestBody.Value.Content {
			request.Headers["Content-Type"] = contentType
			break // Use the first content type
		}
	}

	return request
}

// SaveCollection saves a collection to a file
func SaveCollection(collection *models.Collection, filepath string) error {
	data, err := collection.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize collection: %w", err)
	}

	err = os.WriteFile(filepath, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// LoadCollection loads a collection from a file
func LoadCollection(filepath string) (*models.Collection, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read collection file: %w", err)
	}

	collection, err := models.FromJSON(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection: %w", err)
	}

	return collection, nil
}

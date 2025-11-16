package services

import (
	"curlman/executor"
	"curlman/exporter"
	"curlman/models"
	"fmt"
	"strings"
)

// RequestService handles all request-related business logic
type RequestService struct{}

// NewRequestService creates a new request service
func NewRequestService() *RequestService {
	return &RequestService{}
}

// CreateRequest creates a new request with default values
func (s *RequestService) CreateRequest() *models.Request {
	return &models.Request{
		Name:        "New Request",
		Method:      "GET",
		URL:         "https://api.example.com",
		Path:        "",
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        "",
	}
}

// AddRequest adds a request to a collection
func (s *RequestService) AddRequest(collection *models.Collection, request *models.Request) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate request
	if err := s.ValidateRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	collection.Requests = append(collection.Requests, request)
	return nil
}

// UpdateRequest updates an existing request in a collection
func (s *RequestService) UpdateRequest(collection *models.Collection, index int, request *models.Request) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if index < 0 || index >= len(collection.Requests) {
		return fmt.Errorf("invalid request index: %d", index)
	}

	// Validate request
	if err := s.ValidateRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	collection.Requests[index] = request
	return nil
}

// DeleteRequest deletes a request from a collection
func (s *RequestService) DeleteRequest(collection *models.Collection, index int) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if index < 0 || index >= len(collection.Requests) {
		return fmt.Errorf("invalid request index: %d", index)
	}

	collection.Requests = append(collection.Requests[:index], collection.Requests[index+1:]...)
	return nil
}

// CloneRequest creates a copy of an existing request
func (s *RequestService) CloneRequest(collection *models.Collection, index int) (*models.Request, error) {
	if collection == nil {
		return nil, fmt.Errorf("collection cannot be nil")
	}
	if index < 0 || index >= len(collection.Requests) {
		return nil, fmt.Errorf("invalid request index: %d", index)
	}

	original := collection.Requests[index]
	cloned := original.Clone()
	cloned.Name = original.Name + " (Copy)"

	return cloned, nil
}

// GetRequest retrieves a request by index
func (s *RequestService) GetRequest(collection *models.Collection, index int) (*models.Request, error) {
	if collection == nil {
		return nil, fmt.Errorf("collection cannot be nil")
	}
	if index < 0 || index >= len(collection.Requests) {
		return nil, fmt.Errorf("invalid request index: %d", index)
	}

	return collection.Requests[index], nil
}

// ExecuteRequest executes a request with the given variables
func (s *RequestService) ExecuteRequest(request *models.Request, variables map[string]string) (*executor.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request before execution
	if err := s.ValidateRequest(request); err != nil {
		return nil, fmt.Errorf("cannot execute invalid request: %w", err)
	}

	// Execute the request
	response := executor.Execute(request, variables)
	return response, nil
}

// ExportToCurl generates a curl command for the request
func (s *RequestService) ExportToCurl(request *models.Request, variables map[string]string) (string, error) {
	if request == nil {
		return "", fmt.Errorf("request cannot be nil")
	}

	// Validate request before export
	if err := s.ValidateRequest(request); err != nil {
		return "", fmt.Errorf("cannot export invalid request: %w", err)
	}

	// Inject variables and export
	injected := request.InjectVariables(variables)
	curlCmd := exporter.ToCurl(injected)
	return curlCmd, nil
}

// ValidateRequest validates a request's data
func (s *RequestService) ValidateRequest(request *models.Request) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.Name == "" {
		return fmt.Errorf("request name cannot be empty")
	}

	if request.Method == "" {
		return fmt.Errorf("request method cannot be empty")
	}

	// Validate HTTP method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[strings.ToUpper(request.Method)] {
		return fmt.Errorf("invalid HTTP method: %s", request.Method)
	}

	if request.URL == "" {
		return fmt.Errorf("request URL cannot be empty")
	}

	// Basic URL validation
	if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// UpdateRequestField updates a specific field of a request
func (s *RequestService) UpdateRequestField(request *models.Request, field string, value string) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	switch field {
	case "name":
		if value == "" {
			return fmt.Errorf("request name cannot be empty")
		}
		request.Name = value
	case "method":
		if value == "" {
			return fmt.Errorf("request method cannot be empty")
		}
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"PATCH": true, "HEAD": true, "OPTIONS": true,
		}
		upperValue := strings.ToUpper(value)
		if !validMethods[upperValue] {
			return fmt.Errorf("invalid HTTP method: %s", value)
		}
		request.Method = upperValue
	case "url":
		if value == "" {
			return fmt.Errorf("request URL cannot be empty")
		}
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("URL must start with http:// or https://")
		}
		request.URL = value
	case "path":
		request.Path = value
	case "body":
		request.Body = value
	default:
		return fmt.Errorf("unknown field: %s", field)
	}

	return nil
}

// SetHeader sets a header on a request
func (s *RequestService) SetHeader(request *models.Request, key, value string) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("header key cannot be empty")
	}

	if request.Headers == nil {
		request.Headers = make(map[string]string)
	}
	request.Headers[key] = value
	return nil
}

// DeleteHeader deletes a header from a request
func (s *RequestService) DeleteHeader(request *models.Request, key string) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("header key cannot be empty")
	}

	delete(request.Headers, key)
	return nil
}

// SetQueryParam sets a query parameter on a request
func (s *RequestService) SetQueryParam(request *models.Request, key, value string) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("query parameter key cannot be empty")
	}

	if request.QueryParams == nil {
		request.QueryParams = make(map[string]string)
	}
	request.QueryParams[key] = value
	return nil
}

// DeleteQueryParam deletes a query parameter from a request
func (s *RequestService) DeleteQueryParam(request *models.Request, key string) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("query parameter key cannot be empty")
	}

	delete(request.QueryParams, key)
	return nil
}

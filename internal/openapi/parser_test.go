package openapi_test

import (
	"path/filepath"
	"testing"

	"github.com/leit0/curlman/internal/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFile_OpenAPI30 tests parsing an OpenAPI 3.0 YAML file
func TestParseFile_OpenAPI30(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-openapi-v3.0.yaml")

	collection, err := parser.ParseFile(filePath)

	require.NoError(t, err, "ParseFile should not return an error for valid OpenAPI 3.0 spec")
	require.NotNil(t, collection, "Collection should not be nil")

	assert.Equal(t, "Swagger Petstore", collection.Name)
	assert.Contains(t, collection.Description, "sample API")
	assert.Equal(t, filePath, collection.OpenAPIPath)
}

// TestParseFile_OpenAPI31 tests parsing an OpenAPI 3.1 JSON file
func TestParseFile_OpenAPI31(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-openapi-v3.1.json")

	collection, err := parser.ParseFile(filePath)

	require.NoError(t, err, "ParseFile should not return an error for valid OpenAPI 3.1 spec")
	require.NotNil(t, collection, "Collection should not be nil")

	assert.Equal(t, "Product Store API", collection.Name)
	assert.Contains(t, collection.Description, "e-commerce store")
	assert.Equal(t, filePath, collection.OpenAPIPath)
}

// TestParseFile_Swagger20 tests parsing a Swagger 2.0 file and auto-conversion to OpenAPI 3.0
func TestParseFile_Swagger20(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-swagger-v2.0.json")

	collection, err := parser.ParseFile(filePath)

	require.NoError(t, err, "ParseFile should not return an error for Swagger 2.0 spec (should auto-convert)")
	require.NotNil(t, collection, "Collection should not be nil")

	assert.Equal(t, "Library API", collection.Name)
	assert.Contains(t, collection.Description, "library management")
	assert.Equal(t, filePath, collection.OpenAPIPath)
}

// TestParseFileToSpecRequests_OpenAPI30 tests request generation from OpenAPI 3.0
func TestParseFileToSpecRequests_OpenAPI30(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-openapi-v3.0.yaml")

	requests, err := parser.ParseFileToSpecRequests(filePath)

	require.NoError(t, err, "ParseFileToSpecRequests should not return an error")
	require.NotEmpty(t, requests, "Requests should not be empty")

	// Verify we got the expected number of operations
	// /pets: GET, POST
	// /pets/{petId}: GET, DELETE
	assert.Len(t, requests, 4, "Should have 4 requests from the spec")

	// Verify requests have the expected structure
	for _, req := range requests {
		assert.NotEmpty(t, req.Name, "Request should have a name")
		assert.NotEmpty(t, req.Method, "Request should have a method")
		assert.NotEmpty(t, req.URL, "Request should have a URL")
		assert.NotEmpty(t, req.CurlCommand, "Request should have a curl command")
		assert.Contains(t, req.URL, "https://petstore.swagger.io/v2", "URL should contain server base URL")
		assert.False(t, req.IsManaged, "Spec requests should not be managed")
		assert.True(t, req.OperationExists, "Operation should be marked as existing")
	}

	// Find and verify the POST request with request body
	var postRequest *struct {
		Name        string
		Method      string
		CurlCommand string
	}
	for _, req := range requests {
		if req.Method == "POST" && req.URL == "https://petstore.swagger.io/v2/pets" {
			postRequest = &struct {
				Name        string
				Method      string
				CurlCommand string
			}{
				Name:        req.Name,
				Method:      req.Method,
				CurlCommand: req.CurlCommand,
			}
			break
		}
	}

	require.NotNil(t, postRequest, "Should have a POST /pets request")
	assert.Equal(t, "Create a pet", postRequest.Name)
	assert.Contains(t, postRequest.CurlCommand, `-d '`, "POST request should have request body")
	assert.Contains(t, postRequest.CurlCommand, `"name":"Fluffy"`, "Request body should contain example data")

	// Find and verify a GET request with query parameters
	var getRequest *struct {
		CurlCommand string
	}
	for _, req := range requests {
		if req.Method == "GET" && req.URL == "https://petstore.swagger.io/v2/pets" {
			getRequest = &struct {
				CurlCommand string
			}{
				CurlCommand: req.CurlCommand,
			}
			break
		}
	}

	require.NotNil(t, getRequest, "Should have a GET /pets request")
	assert.Contains(t, getRequest.CurlCommand, "limit=20", "GET request should have default query parameter")
}

// TestParseFileToSpecRequests_OpenAPI31 tests request generation from OpenAPI 3.1
func TestParseFileToSpecRequests_OpenAPI31(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-openapi-v3.1.json")

	requests, err := parser.ParseFileToSpecRequests(filePath)

	require.NoError(t, err, "ParseFileToSpecRequests should not return an error")
	require.NotEmpty(t, requests, "Requests should not be empty")

	// /products: GET, POST
	// /products/{productId}: GET, PUT
	assert.Len(t, requests, 4, "Should have 4 requests from the spec")

	// Verify all requests are properly formed
	for _, req := range requests {
		assert.NotEmpty(t, req.Name, "Request should have a name")
		assert.NotEmpty(t, req.Method, "Request should have a method")
		assert.Contains(t, req.URL, "https://api.productstore.io/v1", "URL should contain server base URL")
	}

	// Verify POST request with examples
	var postRequest *struct {
		CurlCommand string
	}
	for _, req := range requests {
		if req.Method == "POST" {
			postRequest = &struct {
				CurlCommand string
			}{
				CurlCommand: req.CurlCommand,
			}
			break
		}
	}

	require.NotNil(t, postRequest, "Should have a POST request")
	assert.Contains(t, postRequest.CurlCommand, `"name":"Gaming Laptop"`, "Should use example from spec")
	assert.Contains(t, postRequest.CurlCommand, `"price":1299.99`, "Should use example from spec")
}

// TestParseFileToSpecRequests_Swagger20 tests request generation from converted Swagger 2.0
func TestParseFileToSpecRequests_Swagger20(t *testing.T) {
	parser := openapi.NewOpenAPIParser()
	filePath := filepath.Join("testdata", "petstore-swagger-v2.0.json")

	requests, err := parser.ParseFileToSpecRequests(filePath)

	require.NoError(t, err, "ParseFileToSpecRequests should not return an error for Swagger 2.0")
	require.NotEmpty(t, requests, "Requests should not be empty")

	// /books: GET, POST
	// /books/{bookId}: GET, DELETE
	assert.Len(t, requests, 4, "Should have 4 requests from the converted spec")

	// Verify server URL was properly converted
	for _, req := range requests {
		assert.Contains(t, req.URL, "https://api.library.io/v1", "URL should contain converted server URL")
	}

	// Verify GET request with path parameter
	var getByIdRequest *struct {
		CurlCommand string
		URL         string
	}
	for _, req := range requests {
		if req.Method == "GET" && req.URL == "https://api.library.io/v1/books/{bookId}" {
			getByIdRequest = &struct {
				CurlCommand string
				URL         string
			}{
				CurlCommand: req.CurlCommand,
				URL:         req.URL,
			}
			break
		}
	}

	require.NotNil(t, getByIdRequest, "Should have a GET /books/{bookId} request")
	assert.Contains(t, getByIdRequest.CurlCommand, "book-001", "Should use default value for path parameter")
}

// TestValidateOperation tests operation validation
func TestValidateOperation(t *testing.T) {
	parser := openapi.NewOpenAPIParser()

	tests := []struct {
		name      string
		filePath  string
		operation string
		wantValid bool
	}{
		{
			name:      "Valid operation - OpenAPI 3.0",
			filePath:  filepath.Join("testdata", "petstore-openapi-v3.0.yaml"),
			operation: "GET /pets",
			wantValid: true,
		},
		{
			name:      "Valid operation with path param - OpenAPI 3.0",
			filePath:  filepath.Join("testdata", "petstore-openapi-v3.0.yaml"),
			operation: "DELETE /pets/{petId}",
			wantValid: true,
		},
		{
			name:      "Invalid operation - wrong method",
			filePath:  filepath.Join("testdata", "petstore-openapi-v3.0.yaml"),
			operation: "PUT /pets",
			wantValid: false,
		},
		{
			name:      "Invalid operation - wrong path",
			filePath:  filepath.Join("testdata", "petstore-openapi-v3.0.yaml"),
			operation: "GET /nonexistent",
			wantValid: false,
		},
		{
			name:      "Valid operation - OpenAPI 3.1",
			filePath:  filepath.Join("testdata", "petstore-openapi-v3.1.json"),
			operation: "POST /products",
			wantValid: true,
		},
		{
			name:      "Valid operation - Swagger 2.0 (converted)",
			filePath:  filepath.Join("testdata", "petstore-swagger-v2.0.json"),
			operation: "GET /books",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := parser.ValidateOperation(tt.filePath, tt.operation)

			require.NoError(t, err, "ValidateOperation should not return an error")
			assert.Equal(t, tt.wantValid, valid, "Operation validation result mismatch")
		})
	}
}

// TestParseFile_InvalidFile tests error handling for invalid files
func TestParseFile_InvalidFile(t *testing.T) {
	parser := openapi.NewOpenAPIParser()

	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "Non-existent file",
			filePath: filepath.Join("testdata", "nonexistent.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection, err := parser.ParseFile(tt.filePath)

			assert.Error(t, err, "ParseFile should return an error for invalid file")
			assert.Nil(t, collection, "Collection should be nil on error")
		})
	}
}

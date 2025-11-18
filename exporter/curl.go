package exporter

import (
	"github.com/leobrines/curlman/models"
	"fmt"
	"strings"
)

// ToCurl converts a request to a curl command
func ToCurl(request *models.Request) string {
	var parts []string

	// Start with curl command
	parts = append(parts, "curl")

	// Add method
	if request.Method != "" && request.Method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", request.Method))
	}

	// Add headers
	for key, value := range request.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", key, value))
	}

	// Add body if present
	if request.Body != "" {
		escapedBody := strings.ReplaceAll(request.Body, "'", "'\\''")
		parts = append(parts, fmt.Sprintf("-d '%s'", escapedBody))
	}

	// Add URL (including query params)
	url := request.FullURL()
	parts = append(parts, fmt.Sprintf("'%s'", url))

	return strings.Join(parts, " ")
}

// ToCurlWithVariables converts a request to a curl command with variables injected
func ToCurlWithVariables(request *models.Request, variables map[string]string) string {
	injected := request.InjectVariables(variables)
	return ToCurl(injected)
}

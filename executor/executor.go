package executor

import (
	"bytes"
	"curlman/models"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response represents the HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       string
	Duration   time.Duration
	Error      error
}

// Execute executes an HTTP request and returns the response
func Execute(request *models.Request, variables map[string]string) *Response {
	start := time.Now()
	response := &Response{}

	// Inject variables
	injected := request.InjectVariables(variables)

	// Create HTTP request
	url := injected.FullURL()
	var bodyReader io.Reader
	if injected.Body != "" {
		bodyReader = bytes.NewBufferString(injected.Body)
	}

	req, err := http.NewRequest(injected.Method, url, bodyReader)
	if err != nil {
		response.Error = fmt.Errorf("failed to create request: %w", err)
		return response
	}

	// Add headers
	for key, value := range injected.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		response.Error = fmt.Errorf("request failed: %w", err)
		response.Duration = time.Since(start)
		return response
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Errorf("failed to read response body: %w", err)
		response.Duration = time.Since(start)
		return response
	}

	response.StatusCode = resp.StatusCode
	response.Status = resp.Status
	response.Headers = resp.Header
	response.Body = string(bodyBytes)
	response.Duration = time.Since(start)

	return response
}

// FormatResponse formats the response for display
func FormatResponse(resp *Response) string {
	if resp.Error != nil {
		return fmt.Sprintf("Error: %s\nDuration: %s", resp.Error, resp.Duration)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Status: %s\n", resp.Status))
	result.WriteString(fmt.Sprintf("Duration: %s\n\n", resp.Duration))

	result.WriteString("Headers:\n")
	for key, values := range resp.Headers {
		for _, value := range values {
			result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	result.WriteString("\nBody:\n")
	result.WriteString(resp.Body)

	return result.String()
}

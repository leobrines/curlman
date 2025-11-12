package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/leit0/curlman/internal/models"
)

// CurlParser parses curl commands
type CurlParser struct{}

// NewCurlParser creates a new curl parser
func NewCurlParser() *CurlParser {
	return &CurlParser{}
}

// Parse parses a curl command and extracts structured information
func (p *CurlParser) Parse(curlCommand string) (*models.Request, error) {
	curlCommand = strings.TrimSpace(curlCommand)

	// Validate it's a curl command
	if !strings.HasPrefix(curlCommand, "curl") {
		return nil, fmt.Errorf("invalid curl command: must start with 'curl'")
	}

	req := &models.Request{
		CurlCommand: curlCommand,
		Headers:     make(map[string]string),
		Method:      "GET", // Default method
	}

	// Extract URL (look for URL pattern)
	url := p.extractURL(curlCommand)
	if url == "" {
		return nil, fmt.Errorf("no URL found in curl command")
	}
	req.URL = url

	// Extract method (-X or --request)
	method := p.extractMethod(curlCommand)
	if method != "" {
		req.Method = method
	}

	// Extract headers (-H or --header)
	headers := p.extractHeaders(curlCommand)
	req.Headers = headers

	// Extract body (-d, --data, --data-raw, --data-binary)
	body := p.extractBody(curlCommand)
	req.Body = body

	return req, nil
}

// extractURL extracts the URL from a curl command
func (p *CurlParser) extractURL(cmd string) string {
	// Match URLs (http/https)
	urlRegex := regexp.MustCompile(`https?://[^\s'"]+`)
	matches := urlRegex.FindString(cmd)
	return strings.Trim(matches, `"'`)
}

// extractMethod extracts the HTTP method
func (p *CurlParser) extractMethod(cmd string) string {
	// Match -X METHOD or --request METHOD
	methodRegex := regexp.MustCompile(`(?:-X|--request)\s+([A-Z]+)`)
	matches := methodRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractHeaders extracts headers from -H or --header flags
func (p *CurlParser) extractHeaders(cmd string) map[string]string {
	headers := make(map[string]string)

	// Match -H "key: value" or --header "key: value"
	headerRegex := regexp.MustCompile(`(?:-H|--header)\s+['"]([^:]+):\s*([^'"]+)['"]`)
	matches := headerRegex.FindAllStringSubmatch(cmd, -1)

	for _, match := range matches {
		if len(match) > 2 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			headers[key] = value
		}
	}

	return headers
}

// extractBody extracts the request body
func (p *CurlParser) extractBody(cmd string) string {
	// Match -d, --data, --data-raw, or --data-binary
	bodyRegex := regexp.MustCompile(`(?:-d|--data|--data-raw|--data-binary)\s+['"]([^'"]+)['"]`)
	matches := bodyRegex.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Validate validates a curl command
func (p *CurlParser) Validate(curlCommand string) error {
	curlCommand = strings.TrimSpace(curlCommand)

	if !strings.HasPrefix(curlCommand, "curl") {
		return fmt.Errorf("command must start with 'curl'")
	}

	if p.extractURL(curlCommand) == "" {
		return fmt.Errorf("no valid URL found")
	}

	return nil
}

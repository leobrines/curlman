package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/leit0/curlman/internal/models"
)

// Engine handles template processing with environment variables
type Engine struct{}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	return &Engine{}
}

// Process processes a template string with environment variables
func (e *Engine) Process(templateStr string, env *models.Environment) (string, error) {
	if env == nil {
		// No environment, return as-is
		return templateStr, nil
	}

	// Create template
	tmpl, err := template.New("request").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with environment variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, env.Variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ProcessRequest processes a request's curl command with environment variables
func (e *Engine) ProcessRequest(req *models.Request, env *models.Environment) (string, error) {
	return e.Process(req.CurlCommand, env)
}

// ProcessFile processes a file's content with environment variables
func (e *Engine) ProcessFile(content string, env *models.Environment) (string, error) {
	return e.Process(content, env)
}

// HasTemplates checks if a string contains template variables
func (e *Engine) HasTemplates(str string) bool {
	return len(str) > 0 && (containsPattern(str, "{{") || containsPattern(str, "}}"))
}

// containsPattern checks if a string contains a pattern
func containsPattern(str, pattern string) bool {
	for i := 0; i < len(str)-len(pattern)+1; i++ {
		if str[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}

package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/leit0/curlman/internal/models"
	"github.com/leit0/curlman/internal/template"
)

// Executor handles request execution
type Executor struct {
	templateEngine *template.Engine
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	return &Executor{
		templateEngine: template.NewEngine(),
	}
}

// ExecuteRequest executes a request with the given environment
func (e *Executor) ExecuteRequest(ctx context.Context, req *models.Request, env *models.Environment) (string, error) {
	// Process template with environment variables
	curlCmd := req.CurlCommand
	if env != nil {
		processedCmd, err := e.templateEngine.ProcessRequest(req, env)
		if err != nil {
			return "", fmt.Errorf("template processing failed: %w", err)
		}
		curlCmd = processedCmd
	}

	// Parse curl command into shell command
	cmdParts := parseCurlCommand(curlCmd)
	if len(cmdParts) == 0 {
		return "", fmt.Errorf("invalid curl command")
	}

	// Create command with context for cancellation
	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	// Build output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Command: %s\n", curlCmd))
	output.WriteString(fmt.Sprintf("Duration: %v\n\n", duration))

	if stdout.Len() > 0 {
		output.WriteString("STDOUT:\n")
		output.WriteString(stdout.String())
		output.WriteString("\n")
	}

	if stderr.Len() > 0 {
		output.WriteString("STDERR:\n")
		output.WriteString(stderr.String())
		output.WriteString("\n")
	}

	if err != nil {
		if ctx.Err() == context.Canceled {
			output.WriteString("\n[Request cancelled by user]")
			return output.String(), nil
		}
		output.WriteString(fmt.Sprintf("\nError: %v", err))
		return output.String(), err
	}

	output.WriteString("\n[Request completed successfully]")
	return output.String(), nil
}

// parseCurlCommand parses a curl command string into command parts
func parseCurlCommand(curlCmd string) []string {
	// Simple parsing - split by spaces but respect quotes
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range curlCmd {
		switch {
		case (ch == '"' || ch == '\'') && !inQuote:
			inQuote = true
			quoteChar = ch
		case ch == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// ExecuteWithTimeout executes a request with a timeout
func (e *Executor) ExecuteWithTimeout(req *models.Request, env *models.Environment, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return e.ExecuteRequest(ctx, req, env)
}

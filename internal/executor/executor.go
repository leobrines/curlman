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

// PrepareCommand prepares a command for execution with the given environment
// Returns the processed curl command string and the exec.Cmd ready for execution
func (e *Executor) PrepareCommand(req *models.Request, env *models.Environment) (string, *exec.Cmd, error) {
	// Process template with environment variables
	curlCmd := req.CurlCommand
	if env != nil {
		processedCmd, err := e.templateEngine.ProcessRequest(req, env)
		if err != nil {
			return "", nil, fmt.Errorf("template processing failed: %w", err)
		}
		curlCmd = processedCmd
	}

	// Parse curl command into shell command
	cmdParts := parseCurlCommand(curlCmd)
	if len(cmdParts) == 0 {
		return "", nil, fmt.Errorf("invalid curl command")
	}

	// Create command without context (will be managed by tea.ExecProcess)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...) //nolint:gosec

	return curlCmd, cmd, nil
}

// ExecuteRequest executes a request with the given environment
// DEPRECATED: This method is kept for backward compatibility but should not be used
// for interactive execution. Use PrepareCommand with tea.ExecProcess instead.
func (e *Executor) ExecuteRequest(ctx context.Context, req *models.Request, env *models.Environment) (string, error) {
	curlCmd, cmd, err := e.PrepareCommand(req, env)
	if err != nil {
		return "", err
	}

	// Add context for cancellation
	if ctx != nil {
		cmd.Cancel = func() error {
			return cmd.Process.Kill()
		}
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	startTime := time.Now()
	err = cmd.Run()
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
		if ctx != nil && ctx.Err() == context.Canceled {
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

package models

import (
	"encoding/json"
	"strings"
)

// Collection represents a collection of HTTP requests
type Collection struct {
	Name               string            `json:"name"`
	Requests           []*Request        `json:"requests"`
	Variables          map[string]string `json:"variables"`
	ActiveEnvironment  string            `json:"active_environment,omitempty"`
	EnvironmentVars    map[string]string `json:"-"` // Runtime environment variables, not persisted
}

// Request represents an HTTP request
type Request struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Body        string            `json:"body,omitempty"`
	Description string            `json:"description,omitempty"`
}

// Clone creates a deep copy of the request
func (r *Request) Clone() *Request {
	clone := &Request{
		ID:          r.ID + "_clone",
		Name:        r.Name + " (Clone)",
		Method:      r.Method,
		URL:         r.URL,
		Path:        r.Path,
		Body:        r.Body,
		Description: r.Description,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}

	for k, v := range r.Headers {
		clone.Headers[k] = v
	}

	for k, v := range r.QueryParams {
		clone.QueryParams[k] = v
	}

	return clone
}

// InjectVariables replaces variables in the request with their values
func (r *Request) InjectVariables(variables map[string]string) *Request {
	injected := r.Clone()
	injected.ID = r.ID
	injected.Name = r.Name

	// Inject into URL
	injected.URL = replaceVariables(r.URL, variables)
	injected.Path = replaceVariables(r.Path, variables)

	// Inject into headers
	for k, v := range injected.Headers {
		injected.Headers[k] = replaceVariables(v, variables)
	}

	// Inject into query params
	for k, v := range injected.QueryParams {
		injected.QueryParams[k] = replaceVariables(v, variables)
	}

	// Inject into body
	injected.Body = replaceVariables(r.Body, variables)

	return injected
}

// replaceVariables replaces {{var}} placeholders with their values
func replaceVariables(text string, variables map[string]string) string {
	result := text
	for k, v := range variables {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// FullURL returns the complete URL including path and query parameters
func (r *Request) FullURL() string {
	url := r.URL
	if r.Path != "" {
		url = strings.TrimSuffix(url, "/") + "/" + strings.TrimPrefix(r.Path, "/")
	}

	if len(r.QueryParams) > 0 {
		params := []string{}
		for k, v := range r.QueryParams {
			params = append(params, k+"="+v)
		}
		url += "?" + strings.Join(params, "&")
	}

	return url
}

// ToJSON exports the collection to JSON
func (c *Collection) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON imports a collection from JSON
func FromJSON(data string) (*Collection, error) {
	var collection Collection
	err := json.Unmarshal([]byte(data), &collection)
	if err != nil {
		return nil, err
	}
	// Initialize EnvironmentVars map if nil
	if collection.EnvironmentVars == nil {
		collection.EnvironmentVars = make(map[string]string)
	}
	return &collection, nil
}

// GetAllVariables merges environment variables and collection variables
// Environment variables take precedence over collection variables
func (c *Collection) GetAllVariables() map[string]string {
	merged := make(map[string]string)

	// First add collection variables
	for k, v := range c.Variables {
		merged[k] = v
	}

	// Then add/override with environment variables
	for k, v := range c.EnvironmentVars {
		merged[k] = v
	}

	return merged
}

// SetEnvironmentVariables updates the runtime environment variables
func (c *Collection) SetEnvironmentVariables(envVars map[string]string) {
	if c.EnvironmentVars == nil {
		c.EnvironmentVars = make(map[string]string)
	}
	c.EnvironmentVars = envVars
}

// ClearEnvironmentVariables clears the runtime environment variables
func (c *Collection) ClearEnvironmentVariables() {
	c.EnvironmentVars = make(map[string]string)
	c.ActiveEnvironment = ""
}

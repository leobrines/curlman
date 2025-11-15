package models

import (
	"encoding/json"
	"strings"
)

// CollectionEnvironment represents an environment specific to a collection
type CollectionEnvironment struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
}

// Collection represents a collection of HTTP requests
type Collection struct {
	Name                  string                   `json:"name"`
	Requests              []*Request               `json:"requests"`
	Variables             map[string]string        `json:"variables"`
	ActiveEnvironment     string                   `json:"active_environment,omitempty"`
	EnvironmentVars       map[string]string        `json:"-"` // Runtime environment variables, not persisted
	Environments          []CollectionEnvironment  `json:"environments,omitempty"`
	ActiveCollectionEnv   string                   `json:"active_collection_environment,omitempty"`
	CollectionEnvVars     map[string]string        `json:"-"` // Runtime collection environment variables, not persisted
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
	// Initialize CollectionEnvVars map if nil
	if collection.CollectionEnvVars == nil {
		collection.CollectionEnvVars = make(map[string]string)
	}
	// Initialize Environments slice if nil
	if collection.Environments == nil {
		collection.Environments = []CollectionEnvironment{}
	}
	return &collection, nil
}

// GetAllVariables merges global, collection, and environment variables
// Precedence (lowest to highest): Global < Collection < Global Environment < Collection Environment
func (c *Collection) GetAllVariables(globalVars map[string]string) map[string]string {
	merged := make(map[string]string)

	// First add global variables (lowest precedence)
	for k, v := range globalVars {
		merged[k] = v
	}

	// Then add collection variables (overrides global)
	for k, v := range c.Variables {
		merged[k] = v
	}

	// Then add global environment variables (overrides collection)
	for k, v := range c.EnvironmentVars {
		merged[k] = v
	}

	// Finally add collection environment variables (highest precedence, overrides all)
	for k, v := range c.CollectionEnvVars {
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

// SetCollectionEnvironmentVariables updates the runtime collection environment variables
func (c *Collection) SetCollectionEnvironmentVariables(envVars map[string]string) {
	if c.CollectionEnvVars == nil {
		c.CollectionEnvVars = make(map[string]string)
	}
	c.CollectionEnvVars = envVars
}

// ClearCollectionEnvironmentVariables clears the runtime collection environment variables
func (c *Collection) ClearCollectionEnvironmentVariables() {
	c.CollectionEnvVars = make(map[string]string)
	c.ActiveCollectionEnv = ""
}

// GetCollectionEnvironment returns a collection environment by name
func (c *Collection) GetCollectionEnvironment(name string) *CollectionEnvironment {
	for i := range c.Environments {
		if c.Environments[i].Name == name {
			return &c.Environments[i]
		}
	}
	return nil
}

// AddCollectionEnvironment adds a new collection environment
func (c *Collection) AddCollectionEnvironment(name string) *CollectionEnvironment {
	if c.Environments == nil {
		c.Environments = []CollectionEnvironment{}
	}

	env := CollectionEnvironment{
		Name:      name,
		Variables: make(map[string]string),
	}
	c.Environments = append(c.Environments, env)
	return &c.Environments[len(c.Environments)-1]
}

// DeleteCollectionEnvironment removes a collection environment by name
func (c *Collection) DeleteCollectionEnvironment(name string) bool {
	for i, env := range c.Environments {
		if env.Name == name {
			c.Environments = append(c.Environments[:i], c.Environments[i+1:]...)
			if c.ActiveCollectionEnv == name {
				c.ClearCollectionEnvironmentVariables()
			}
			return true
		}
	}
	return false
}

// RenameCollectionEnvironment renames a collection environment
func (c *Collection) RenameCollectionEnvironment(oldName, newName string) bool {
	env := c.GetCollectionEnvironment(oldName)
	if env != nil {
		env.Name = newName
		if c.ActiveCollectionEnv == oldName {
			c.ActiveCollectionEnv = newName
		}
		return true
	}
	return false
}

// ListCollectionEnvironments returns a list of collection environment names
func (c *Collection) ListCollectionEnvironments() []string {
	names := make([]string, len(c.Environments))
	for i, env := range c.Environments {
		names[i] = env.Name
	}
	return names
}

// ActivateCollectionEnvironment activates a collection environment
func (c *Collection) ActivateCollectionEnvironment(name string) bool {
	env := c.GetCollectionEnvironment(name)
	if env != nil {
		c.ActiveCollectionEnv = name
		c.SetCollectionEnvironmentVariables(env.Variables)
		return true
	}
	return false
}

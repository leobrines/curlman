package models

import "time"

// Environment represents a set of variables for request templating
type Environment struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewEnvironment creates a new environment with the given name
func NewEnvironment(name string) *Environment {
	now := time.Now()
	return &Environment{
		ID:        generateID(),
		Name:      name,
		Variables: make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetVariable retrieves a variable value by key
func (e *Environment) GetVariable(key string) (string, bool) {
	val, ok := e.Variables[key]
	return val, ok
}

// SetVariable sets a variable value
func (e *Environment) SetVariable(key, value string) {
	e.Variables[key] = value
	e.UpdatedAt = time.Now()
}

// DeleteVariable removes a variable by key
func (e *Environment) DeleteVariable(key string) {
	delete(e.Variables, key)
	e.UpdatedAt = time.Now()
}

package models

import "time"

// Collection represents a group of API requests
type Collection struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	OpenAPIPath  string    `json:"openapi_path,omitempty"` // Absolute path to external OpenAPI file
	IsTemplated  bool      `json:"is_templated"`
	Requests     []Request `json:"requests"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

}

// NewCollection creates a new collection
func NewCollection(name, description string) *Collection {
	now := time.Now()
	return &Collection{
		ID:          generateID(),
		Name:        name,
		Description: description,
		Requests:    []Request{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddRequest adds a request to the collection
func (c *Collection) AddRequest(req Request) {
	c.Requests = append(c.Requests, req)
	c.UpdatedAt = time.Now()
}

// RemoveRequest removes a request by ID
func (c *Collection) RemoveRequest(requestID string) bool {
	for i, req := range c.Requests {
		if req.ID == requestID {
			c.Requests = append(c.Requests[:i], c.Requests[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetRequest retrieves a request by ID
func (c *Collection) GetRequest(requestID string) (*Request, bool) {
	for i := range c.Requests {
		if c.Requests[i].ID == requestID {
			return &c.Requests[i], true
		}
	}
	return nil, false
}

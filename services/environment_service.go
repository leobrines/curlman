package services

import (
	"curlman/environment"
	"curlman/models"
	"fmt"
)

// EnvironmentService handles all environment-related business logic
type EnvironmentService struct{}

// NewEnvironmentService creates a new environment service
func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{}
}

// ========== Global Environments ==========

// ListGlobalEnvironments returns a list of all global environment names
func (s *EnvironmentService) ListGlobalEnvironments() ([]string, error) {
	environments, err := environment.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list global environments: %w", err)
	}
	return environments, nil
}

// GetGlobalEnvironment retrieves a global environment by name
func (s *EnvironmentService) GetGlobalEnvironment(name string) (*environment.Environment, error) {
	if name == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	env, err := environment.Load(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment '%s': %w", name, err)
	}
	return env, nil
}

// CreateGlobalEnvironment creates a new global environment
func (s *EnvironmentService) CreateGlobalEnvironment(name string) error {
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	// Check if environment already exists
	if environment.Exists(name) {
		return fmt.Errorf("environment '%s' already exists", name)
	}

	// Create new environment
	env := environment.NewEnvironment(name)

	if err := env.Save(); err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	return nil
}

// DeleteGlobalEnvironment deletes a global environment
func (s *EnvironmentService) DeleteGlobalEnvironment(name string) error {
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	if err := environment.Delete(name); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	return nil
}

// RenameGlobalEnvironment renames a global environment
func (s *EnvironmentService) RenameGlobalEnvironment(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("environment names cannot be empty")
	}

	if oldName == newName {
		return fmt.Errorf("new name must be different from old name")
	}

	// Check if new name already exists
	if environment.Exists(newName) {
		return fmt.Errorf("environment '%s' already exists", newName)
	}

	// Get existing environment
	env, err := environment.Load(oldName)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Update name and save as new
	env.Name = newName
	if err := env.Save(); err != nil {
		return fmt.Errorf("failed to save renamed environment: %w", err)
	}

	// Delete old environment
	if err := environment.Delete(oldName); err != nil {
		return fmt.Errorf("failed to delete old environment: %w", err)
	}

	return nil
}

// SetGlobalEnvironmentVariable sets a variable in a global environment
func (s *EnvironmentService) SetGlobalEnvironmentVariable(envName, key, value string) error {
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	env, err := environment.Load(envName)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}
	env.Variables[key] = value

	if err := env.Save(); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	return nil
}

// DeleteGlobalEnvironmentVariable deletes a variable from a global environment
func (s *EnvironmentService) DeleteGlobalEnvironmentVariable(envName, key string) error {
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	env, err := environment.Load(envName)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	delete(env.Variables, key)

	if err := env.Save(); err != nil {
		return fmt.Errorf("failed to save environment: %w", err)
	}

	return nil
}

// ActivateGlobalEnvironment activates a global environment in the collection
func (s *EnvironmentService) ActivateGlobalEnvironment(collection *models.Collection, envName string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	// Verify environment exists
	env, err := environment.Load(envName)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	// Set active environment and its variables
	collection.ActiveEnvironment = envName
	collection.SetEnvironmentVariables(env.Variables)

	return nil
}

// DeactivateGlobalEnvironment deactivates the global environment in the collection
func (s *EnvironmentService) DeactivateGlobalEnvironment(collection *models.Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	collection.ClearEnvironmentVariables()

	return nil
}

// ========== Collection Environments ==========

// ListCollectionEnvironments returns a list of all collection environment names
func (s *EnvironmentService) ListCollectionEnvironments(collection *models.Collection) []string {
	if collection == nil {
		return []string{}
	}
	return collection.ListCollectionEnvironments()
}

// GetCollectionEnvironment retrieves a collection environment by name
func (s *EnvironmentService) GetCollectionEnvironment(collection *models.Collection, name string) (*models.CollectionEnvironment, error) {
	if collection == nil {
		return nil, fmt.Errorf("collection cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("environment name cannot be empty")
	}

	env := collection.GetCollectionEnvironment(name)
	if env == nil {
		return nil, fmt.Errorf("collection environment '%s' not found", name)
	}

	return env, nil
}

// CreateCollectionEnvironment creates a new collection environment
func (s *EnvironmentService) CreateCollectionEnvironment(collection *models.Collection, name string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	// Check if environment already exists
	if collection.GetCollectionEnvironment(name) != nil {
		return fmt.Errorf("collection environment '%s' already exists", name)
	}

	// Create new environment using model method
	collection.AddCollectionEnvironment(name)
	return nil
}

// DeleteCollectionEnvironment deletes a collection environment
func (s *EnvironmentService) DeleteCollectionEnvironment(collection *models.Collection, name string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	if !collection.DeleteCollectionEnvironment(name) {
		return fmt.Errorf("collection environment '%s' not found", name)
	}

	return nil
}

// RenameCollectionEnvironment renames a collection environment
func (s *EnvironmentService) RenameCollectionEnvironment(collection *models.Collection, oldName, newName string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if oldName == "" || newName == "" {
		return fmt.Errorf("environment names cannot be empty")
	}
	if oldName == newName {
		return fmt.Errorf("new name must be different from old name")
	}

	// Check if new name already exists
	if collection.GetCollectionEnvironment(newName) != nil {
		return fmt.Errorf("collection environment '%s' already exists", newName)
	}

	// Rename using model method
	if !collection.RenameCollectionEnvironment(oldName, newName) {
		return fmt.Errorf("collection environment '%s' not found", oldName)
	}

	return nil
}

// SetCollectionEnvironmentVariable sets a variable in a collection environment
func (s *EnvironmentService) SetCollectionEnvironmentVariable(collection *models.Collection, envName, key, value string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	env := collection.GetCollectionEnvironment(envName)
	if env == nil {
		return fmt.Errorf("collection environment '%s' not found", envName)
	}

	if env.Variables == nil {
		env.Variables = make(map[string]string)
	}
	env.Variables[key] = value

	return nil
}

// DeleteCollectionEnvironmentVariable deletes a variable from a collection environment
func (s *EnvironmentService) DeleteCollectionEnvironmentVariable(collection *models.Collection, envName, key string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	env := collection.GetCollectionEnvironment(envName)
	if env == nil {
		return fmt.Errorf("collection environment '%s' not found", envName)
	}

	delete(env.Variables, key)
	return nil
}

// ActivateCollectionEnvironment activates a collection environment
func (s *EnvironmentService) ActivateCollectionEnvironment(collection *models.Collection, envName string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if envName == "" {
		return fmt.Errorf("environment name cannot be empty")
	}

	if !collection.ActivateCollectionEnvironment(envName) {
		return fmt.Errorf("collection environment '%s' not found", envName)
	}

	return nil
}

// DeactivateCollectionEnvironment deactivates the collection environment
func (s *EnvironmentService) DeactivateCollectionEnvironment(collection *models.Collection) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	collection.ClearCollectionEnvironmentVariables()
	return nil
}

package services

import (
	"github.com/leobrines/curlman/config"
	"github.com/leobrines/curlman/models"
	"fmt"
	"strings"
)

// VariableService handles all variable-related business logic
type VariableService struct {
	globalConfig *config.GlobalConfig
}

// NewVariableService creates a new variable service
func NewVariableService(globalConfig *config.GlobalConfig) *VariableService {
	return &VariableService{
		globalConfig: globalConfig,
	}
}

// GetAllVariables returns all variables merged with proper precedence
// Precedence: Global < Collection < Global Environment < Collection Environment
func (s *VariableService) GetAllVariables(collection *models.Collection) map[string]string {
	if collection == nil {
		// If no collection, return only global variables
		if s.globalConfig != nil {
			return s.copyMap(s.globalConfig.Variables)
		}
		return make(map[string]string)
	}

	// Use the collection's method which already implements precedence
	globalVars := make(map[string]string)
	if s.globalConfig != nil {
		globalVars = s.globalConfig.Variables
	}

	return collection.GetAllVariables(globalVars)
}

// SetCollectionVariable sets a variable in the collection
func (s *VariableService) SetCollectionVariable(collection *models.Collection, key, value string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	if collection.Variables == nil {
		collection.Variables = make(map[string]string)
	}
	collection.Variables[key] = value
	return nil
}

// DeleteCollectionVariable deletes a variable from the collection
func (s *VariableService) DeleteCollectionVariable(collection *models.Collection, key string) error {
	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	delete(collection.Variables, key)
	return nil
}

// GetCollectionVariable gets a variable value from the collection
func (s *VariableService) GetCollectionVariable(collection *models.Collection, key string) (string, bool) {
	if collection == nil || collection.Variables == nil {
		return "", false
	}

	value, exists := collection.Variables[key]
	return value, exists
}

// ListCollectionVariables returns all collection variables
func (s *VariableService) ListCollectionVariables(collection *models.Collection) map[string]string {
	if collection == nil || collection.Variables == nil {
		return make(map[string]string)
	}
	return s.copyMap(collection.Variables)
}

// SetGlobalVariable sets a global variable
func (s *VariableService) SetGlobalVariable(key, value string) error {
	if s.globalConfig == nil {
		return fmt.Errorf("global config not initialized")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	s.globalConfig.SetVariable(key, value)
	if err := s.globalConfig.Save(); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	return nil
}

// DeleteGlobalVariable deletes a global variable
func (s *VariableService) DeleteGlobalVariable(key string) error {
	if s.globalConfig == nil {
		return fmt.Errorf("global config not initialized")
	}
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	s.globalConfig.DeleteVariable(key)
	if err := s.globalConfig.Save(); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	return nil
}

// GetGlobalVariable gets a global variable value
func (s *VariableService) GetGlobalVariable(key string) (string, bool) {
	if s.globalConfig == nil || s.globalConfig.Variables == nil {
		return "", false
	}

	value, exists := s.globalConfig.Variables[key]
	return value, exists
}

// ListGlobalVariables returns all global variables
func (s *VariableService) ListGlobalVariables() map[string]string {
	if s.globalConfig == nil || s.globalConfig.Variables == nil {
		return make(map[string]string)
	}
	return s.copyMap(s.globalConfig.Variables)
}

// ValidateVariableName validates a variable name
func (s *VariableService) ValidateVariableName(name string) error {
	if name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	// Check for invalid characters (basic validation)
	if strings.Contains(name, " ") {
		return fmt.Errorf("variable name cannot contain spaces")
	}

	if strings.HasPrefix(name, "{{") || strings.HasSuffix(name, "}}") {
		return fmt.Errorf("variable name should not include {{ }} markers")
	}

	return nil
}

// InjectVariables injects variables into a request (returns new request)
func (s *VariableService) InjectVariables(request *models.Request, variables map[string]string) *models.Request {
	if request == nil {
		return nil
	}
	return request.InjectVariables(variables)
}

// FindUnresolvedVariables finds all variable references in a request that are not in the provided variables
func (s *VariableService) FindUnresolvedVariables(request *models.Request, variables map[string]string) []string {
	if request == nil {
		return []string{}
	}

	unresolved := make(map[string]bool)

	// Helper to find variables in a string
	findVars := func(text string) {
		start := 0
		for {
			startIdx := strings.Index(text[start:], "{{")
			if startIdx == -1 {
				break
			}
			startIdx += start

			endIdx := strings.Index(text[startIdx:], "}}")
			if endIdx == -1 {
				break
			}
			endIdx += startIdx

			varName := strings.TrimSpace(text[startIdx+2 : endIdx])
			if _, exists := variables[varName]; !exists {
				unresolved[varName] = true
			}

			start = endIdx + 2
		}
	}

	// Check all fields that support variable injection
	findVars(request.URL)
	findVars(request.Path)
	findVars(request.Body)

	for _, v := range request.Headers {
		findVars(v)
	}

	for _, v := range request.QueryParams {
		findVars(v)
	}

	// Convert map to slice
	result := make([]string, 0, len(unresolved))
	for varName := range unresolved {
		result = append(result, varName)
	}

	return result
}

// Helper function to copy a map
func (s *VariableService) copyMap(original map[string]string) map[string]string {
	copy := make(map[string]string, len(original))
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

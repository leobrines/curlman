package ui

import (
	"curlman/config"
	"curlman/environment"
	"curlman/executor"
	"curlman/models"
	"curlman/services"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view int

const (
	viewMain view = iota
	viewRequestList
	viewRequestDetail
	viewRequestEdit
	viewResponse
	viewVariables
	viewHelp
	viewHeaders
	viewQueryParams
	viewEnvironments
	viewEnvironmentDetail
	viewEnvironmentVariables
	viewGlobalVariables
)

type editField int

const (
	editName editField = iota
	editMethod
	editURL
	editPath
	editHeader
	editQuery
	editBody
)

type Model struct {
	// Data
	collection            *models.Collection
	response              *executor.Response
	environments          []string
	currentEnv            *environment.Environment
	currentCollectionEnv  *models.CollectionEnvironment
	globalConfig          *config.GlobalConfig

	// Services (Business Logic Layer)
	collectionService  *services.CollectionService
	requestService     *services.RequestService
	variableService    *services.VariableService
	environmentService *services.EnvironmentService

	// UI State
	currentView          view
	selectedRequest      int
	selectedField        int
	cursor               int
	textInput            textinput.Model
	editing              bool
	editingField         editField
	editingKey           string
	message              string
	width                int
	height               int
	selectedEnvIdx       int
	viewingCollectionEnv bool // true = collection envs, false = global envs
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 500

	// Load global config
	globalConfig, err := config.Load()
	if err != nil {
		// If loading fails, create a new empty config
		globalConfig = config.NewGlobalConfig()
	}

	// Initialize services
	collectionService := services.NewCollectionService()
	requestService := services.NewRequestService()
	variableService := services.NewVariableService(globalConfig)
	environmentService := services.NewEnvironmentService()

	// Create initial collection using service
	collection := collectionService.CreateEmptyCollection()

	return Model{
		// Data
		collection:   collection,
		globalConfig: globalConfig,
		environments: []string{},

		// Services
		collectionService:  collectionService,
		requestService:     requestService,
		variableService:    variableService,
		environmentService: environmentService,

		// UI State
		currentView: viewMain,
		textInput:   ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			return m.handleEditingInput(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == viewMain {
				return m, tea.Quit
			}
			m.currentView = viewMain
			m.message = ""
			return m, nil

		case "?":
			m.currentView = viewHelp
			return m, nil

		case "i":
			if m.currentView == viewMain {
				m.message = "Enter OpenAPI file path:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				m.editingField = editName
				return m, nil
			}

		case "r":
			if m.currentView == viewMain {
				m.currentView = viewRequestList
				return m, nil
			}

		case "v":
			if m.currentView == viewMain {
				m.currentView = viewVariables
				m.cursor = 0
				return m, nil
			} else if m.currentView == viewEnvironmentDetail && m.currentEnv != nil {
				m.currentView = viewEnvironmentVariables
				m.cursor = 0
				return m, nil
			}

		case "t":
			if m.currentView == viewEnvironments {
				// Toggle between global and collection environments
				m.viewingCollectionEnv = !m.viewingCollectionEnv
				m.cursor = 0
				m.currentEnv = nil
				m.currentCollectionEnv = nil

				if m.viewingCollectionEnv {
					// Switch to collection environments
					m.environments = m.environmentService.ListCollectionEnvironments(m.collection)
					m.message = "Viewing collection environments"
				} else {
					// Switch to global environments
					envs, err := m.environmentService.ListGlobalEnvironments()
					if err != nil {
						m.message = fmt.Sprintf("Error loading environments: %s", err)
						return m, nil
					}
					m.environments = envs
					m.message = "Viewing global environments"
				}
				return m, nil
			}

		case "e":
			if m.currentView == viewMain {
				// Load environments list (global by default)
				m.viewingCollectionEnv = false
				envs, err := m.environmentService.ListGlobalEnvironments()
				if err != nil {
					m.message = fmt.Sprintf("Error loading environments: %s", err)
					return m, nil
				}
				m.environments = envs
				m.currentView = viewEnvironments
				m.cursor = 0
				return m, nil
			} else if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				m.currentView = viewRequestEdit
				m.selectedField = 0
				return m, nil
			} else if m.currentView == viewEnvironmentDetail {
				// Edit environment name
				if m.viewingCollectionEnv && m.currentCollectionEnv != nil {
					m.message = "Enter new environment name:"
					m.textInput.SetValue(m.currentCollectionEnv.Name)
					m.textInput.Focus()
					m.editing = true
					m.editingField = editName
					return m, nil
				} else if !m.viewingCollectionEnv && m.currentEnv != nil {
					m.message = "Enter new environment name:"
					m.textInput.SetValue(m.currentEnv.Name)
					m.textInput.Focus()
					m.editing = true
					m.editingField = editName
					return m, nil
				}
			}

		case "g":
			if m.currentView == viewMain {
				m.currentView = viewGlobalVariables
				m.cursor = 0
				return m, nil
			}

		case "s":
			if m.currentView == viewMain {
				m.message = "Enter filename to save:"
				m.textInput.SetValue("collection.json")
				m.textInput.Focus()
				m.editing = true
				m.editingField = editPath
				return m, nil
			} else if m.currentView == viewResponse && m.response != nil {
				m.message = "Enter filename to save response:"
				m.textInput.SetValue("response.txt")
				m.textInput.Focus()
				m.editing = true
				m.editingField = editBody
				return m, nil
			} else if m.currentView == viewEnvironmentDetail {
				// Save environment
				if m.viewingCollectionEnv {
					// Collection environments are saved with the collection
					m.message = "Collection environment will be saved with the collection"
				} else if m.currentEnv != nil {
					err := m.currentEnv.Save()
					if err != nil {
						m.message = fmt.Sprintf("Error saving environment: %s", err)
					} else {
						m.message = fmt.Sprintf("Environment '%s' saved", m.currentEnv.Name)
						// Reload environments list
						envs, _ := m.environmentService.ListGlobalEnvironments()
						m.environments = envs
					}
				}
				return m, nil
			}

		case "l":
			if m.currentView == viewMain {
				m.message = "Enter filename to load:"
				m.textInput.SetValue("collection.json")
				m.textInput.Focus()
				m.editing = true
				m.editingField = editURL
				return m, nil
			}

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			m.cursor++
			return m, nil

		case "h":
			if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				m.currentView = viewHeaders
				m.cursor = 0
				return m, nil
			}

		case "p":
			if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				m.currentView = viewQueryParams
				m.cursor = 0
				return m, nil
			}

		case "c":
			if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				cloned, err := m.requestService.CloneRequest(m.collection, m.selectedRequest)
				if err != nil {
					m.message = fmt.Sprintf("Error cloning request: %s", err)
				} else {
					m.collection.Requests = append(m.collection.Requests, cloned)
					m.message = "Request cloned successfully!"
				}
				return m, nil
			}

		case "x":
			if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				req := m.collection.Requests[m.selectedRequest]
				// Use merged variables via service
				allVars := m.variableService.GetAllVariables(m.collection)
				curlCmd, err := m.requestService.ExportToCurl(req, allVars)
				if err != nil {
					m.message = fmt.Sprintf("Error exporting: %s", err)
				} else {
					m.message = "Curl command:\n" + curlCmd
				}
				return m, nil
			}

		case "a":
			if m.currentView == viewEnvironmentDetail {
				if m.viewingCollectionEnv && m.currentCollectionEnv != nil {
					// Activate collection environment
					err := m.environmentService.ActivateCollectionEnvironment(m.collection, m.currentCollectionEnv.Name)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Collection environment '%s' activated", m.currentCollectionEnv.Name)
					}
				} else if !m.viewingCollectionEnv && m.currentEnv != nil {
					// Activate global environment
					err := m.environmentService.ActivateGlobalEnvironment(m.collection, m.currentEnv.Name)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Global environment '%s' activated", m.currentEnv.Name)
					}
				}
				return m, nil
			}

		case "d":
			if m.currentView == viewRequestList && m.selectedRequest >= 0 && len(m.collection.Requests) > 0 {
				err := m.requestService.DeleteRequest(m.collection, m.selectedRequest)
				if err != nil {
					m.message = fmt.Sprintf("Error deleting request: %s", err)
				} else {
					if m.selectedRequest >= len(m.collection.Requests) && m.selectedRequest > 0 {
						m.selectedRequest--
					}
					m.message = "Request deleted"
				}
				return m, nil
			} else if m.currentView == viewEnvironments && m.cursor < len(m.environments) {
				// Delete environment
				envName := m.environments[m.cursor]

				if m.viewingCollectionEnv {
					// Delete collection environment
					err := m.environmentService.DeleteCollectionEnvironment(m.collection, envName)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.environments = m.environmentService.ListCollectionEnvironments(m.collection)
						if m.cursor >= len(m.environments) && m.cursor > 0 {
							m.cursor--
						}
						m.message = fmt.Sprintf("Collection environment '%s' deleted", envName)
					}
				} else {
					// Delete global environment
					err := m.environmentService.DeleteGlobalEnvironment(envName)
					if err != nil {
						m.message = fmt.Sprintf("Error deleting environment: %s", err)
					} else {
						// If deleted environment was active, deactivate it
						if m.collection.ActiveEnvironment == envName {
							m.environmentService.DeactivateGlobalEnvironment(m.collection)
						}
						// Reload environments list
						envs, _ := m.environmentService.ListGlobalEnvironments()
						m.environments = envs
						if m.cursor >= len(m.environments) && m.cursor > 0 {
							m.cursor--
						}
						m.message = fmt.Sprintf("Environment '%s' deleted", envName)
					}
				}
				return m, nil
			} else if m.currentView == viewEnvironmentDetail {
				// Delete current environment
				if m.viewingCollectionEnv && m.currentCollectionEnv != nil {
					envName := m.currentCollectionEnv.Name
					err := m.environmentService.DeleteCollectionEnvironment(m.collection, envName)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.environments = m.environmentService.ListCollectionEnvironments(m.collection)
						m.currentCollectionEnv = nil
						m.currentView = viewEnvironments
						m.message = fmt.Sprintf("Collection environment '%s' deleted", envName)
					}
				} else if !m.viewingCollectionEnv && m.currentEnv != nil {
					envName := m.currentEnv.Name
					err := m.environmentService.DeleteGlobalEnvironment(envName)
					if err != nil {
						m.message = fmt.Sprintf("Error deleting environment: %s", err)
					} else {
						// If deleted environment was active, deactivate it
						if m.collection.ActiveEnvironment == envName {
							m.environmentService.DeactivateGlobalEnvironment(m.collection)
						}
						// Reload environments list and go back
						envs, _ := m.environmentService.ListGlobalEnvironments()
						m.environments = envs
						m.currentEnv = nil
						m.currentView = viewEnvironments
						m.message = fmt.Sprintf("Environment '%s' deleted", envName)
					}
				}
				return m, nil
			}
			if m.currentView == viewVariables && len(m.collection.Variables) > 0 {
				varKeys := getSortedVariableKeys(m.collection.Variables)
				if m.cursor >= 0 && m.cursor < len(varKeys) {
					keyToDelete := varKeys[m.cursor]
					err := m.variableService.DeleteCollectionVariable(m.collection, keyToDelete)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Variable '%s' deleted", keyToDelete)
						if m.cursor >= len(m.collection.Variables) && m.cursor > 0 {
							m.cursor--
						}
					}
				}
				return m, nil
			}
			if m.currentView == viewGlobalVariables && len(m.globalConfig.Variables) > 0 {
				varKeys := getSortedVariableKeys(m.globalConfig.Variables)
				if m.cursor >= 0 && m.cursor < len(varKeys) {
					keyToDelete := varKeys[m.cursor]
					err := m.variableService.DeleteGlobalVariable(keyToDelete)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Global variable '%s' deleted", keyToDelete)
						if m.cursor >= len(m.globalConfig.Variables) && m.cursor > 0 {
							m.cursor--
						}
					}
				}
				return m, nil
			}

		case "n":
			if m.currentView == viewRequestList {
				newReq := m.requestService.CreateRequest()
				err := m.requestService.AddRequest(m.collection, newReq)
				if err != nil {
					m.message = fmt.Sprintf("Error creating request: %s", err)
				} else {
					m.selectedRequest = len(m.collection.Requests) - 1
					m.currentView = viewRequestEdit
					m.message = "New request created"
				}
				return m, nil
			}
			if m.currentView == viewVariables {
				m.startEditingNewVariable()
				return m, nil
			}
			if m.currentView == viewGlobalVariables {
				m.startEditingNewGlobalVariable()
				return m, nil
			}

		case "esc", "backspace":
			if m.currentView == viewRequestDetail {
				m.currentView = viewRequestList
				return m, nil
			}
			if m.currentView == viewRequestList || m.currentView == viewVariables || m.currentView == viewEnvironments || m.currentView == viewGlobalVariables {
				m.currentView = viewMain
				return m, nil
			}
			if m.currentView == viewRequestEdit {
				m.currentView = viewRequestDetail
				return m, nil
			}
			if m.currentView == viewResponse {
				m.currentView = viewRequestDetail
				return m, nil
			}
			if m.currentView == viewHeaders || m.currentView == viewQueryParams {
				m.currentView = viewRequestDetail
				return m, nil
			}
			if m.currentView == viewEnvironmentDetail || m.currentView == viewEnvironmentVariables {
				m.currentView = viewEnvironments
				return m, nil
			}
		}
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case viewRequestList:
		if m.cursor < len(m.collection.Requests) {
			m.selectedRequest = m.cursor
			m.currentView = viewRequestDetail
			m.cursor = 0
		}
	case viewRequestDetail:
		if m.selectedRequest >= 0 {
			req := m.collection.Requests[m.selectedRequest]
			// Use merged variables (global + collection + environment) via service
			allVars := m.variableService.GetAllVariables(m.collection)
			response, err := m.requestService.ExecuteRequest(req, allVars)
			if err != nil {
				m.message = fmt.Sprintf("Error executing request: %s", err)
			} else {
				m.response = response
				m.currentView = viewResponse
			}
		}
	case viewRequestEdit:
		m.startEditing()
	case viewVariables:
		m.startEditingVariable()
	case viewHeaders:
		m.startEditingHeader()
	case viewQueryParams:
		m.startEditingQueryParam()
	case viewEnvironments:
		m.handleEnvironmentSelect()
	case viewEnvironmentVariables:
		m.startEditingEnvironmentVariable()
	case viewGlobalVariables:
		m.startEditingGlobalVariable()
	}
	return m, nil
}

func (m *Model) startEditing() {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return
	}

	req := m.collection.Requests[m.selectedRequest]
	m.editing = true
	m.textInput.Focus()

	switch m.selectedField {
	case 0: // Name
		m.editingField = editName
		m.textInput.SetValue(req.Name)
	case 1: // Method
		m.editingField = editMethod
		m.textInput.SetValue(req.Method)
	case 2: // URL
		m.editingField = editURL
		m.textInput.SetValue(req.URL)
	case 3: // Path
		m.editingField = editPath
		m.textInput.SetValue(req.Path)
	case 4: // Body
		m.editingField = editBody
		m.textInput.SetValue(req.Body)
	}
}

func (m *Model) startEditingVariable() {
	varKeys := getSortedVariableKeys(m.collection.Variables)

	// If cursor is on an existing variable, edit it
	if m.cursor >= 0 && m.cursor < len(varKeys) {
		key := varKeys[m.cursor]
		value := m.collection.Variables[key]
		m.editingKey = key
		m.editing = true
		m.textInput.Focus()
		m.editingField = editHeader
		m.textInput.SetValue(value)
		m.message = fmt.Sprintf("Editing variable '%s' (press enter to save):", key)
	} else {
		// Otherwise, create a new variable
		m.startEditingNewVariable()
	}
}

func (m *Model) startEditingNewVariable() {
	m.editing = true
	m.textInput.Focus()
	m.editingField = editHeader
	m.textInput.SetValue("")
	m.editingKey = ""
	m.message = "Enter variable name:"
}

func (m *Model) startEditingHeader() {
	m.editing = true
	m.textInput.Focus()
	m.editingField = editHeader
	m.textInput.SetValue("")
	m.message = "Enter header name:"
}

func (m *Model) startEditingQueryParam() {
	m.editing = true
	m.textInput.Focus()
	m.editingField = editQuery
	m.textInput.SetValue("")
	m.message = "Enter query parameter name:"
}

func (m Model) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		value := m.textInput.Value()
		m.textInput.Blur()
		m.editing = false

		// Handle different editing contexts
		if m.currentView == viewMain {
			if m.editingField == editName { // Import OpenAPI
				collection, err := m.collectionService.ImportFromOpenAPI(value)
				if err != nil {
					m.message = fmt.Sprintf("Error importing: %s", err)
				} else {
					m.collection = collection
					m.message = fmt.Sprintf("Imported %d requests from %s", len(collection.Requests), collection.Name)
				}
			} else if m.editingField == editPath { // Save collection
				fullPath, err := m.collectionService.SaveCollection(m.collection, value)
				if err != nil {
					m.message = fmt.Sprintf("Error saving: %s", err)
				} else {
					m.message = fmt.Sprintf("Collection saved to %s", fullPath)
				}
			} else if m.editingField == editURL { // Load collection
				collection, err := m.collectionService.LoadCollection(value)
				if err != nil {
					m.message = fmt.Sprintf("Error loading: %s", err)
				} else {
					m.collection = collection
					// Initialize EnvironmentVars if nil
					if m.collection.EnvironmentVars == nil {
						m.collection.EnvironmentVars = make(map[string]string)
					}
					// Load active global environment if set
					if m.collection.ActiveEnvironment != "" {
						err := m.environmentService.ActivateGlobalEnvironment(m.collection, m.collection.ActiveEnvironment)
						if err != nil {
							// Silently ignore if environment doesn't exist
							m.collection.ActiveEnvironment = ""
						}
					}
					// Load active collection environment if set
					if m.collection.ActiveCollectionEnv != "" {
						err := m.environmentService.ActivateCollectionEnvironment(m.collection, m.collection.ActiveCollectionEnv)
						if err != nil {
							// Silently ignore if environment doesn't exist
							m.collection.ActiveCollectionEnv = ""
						}
					}
					m.message = fmt.Sprintf("Loaded collection: %s", collection.Name)
				}
			}
		} else if m.currentView == viewRequestEdit && m.selectedRequest >= 0 {
			req := m.collection.Requests[m.selectedRequest]
			var err error
			switch m.editingField {
			case editName:
				err = m.requestService.UpdateRequestField(req, "name", value)
			case editMethod:
				err = m.requestService.UpdateRequestField(req, "method", value)
			case editURL:
				err = m.requestService.UpdateRequestField(req, "url", value)
			case editPath:
				err = m.requestService.UpdateRequestField(req, "path", value)
			case editBody:
				err = m.requestService.UpdateRequestField(req, "body", value)
			}
			if err != nil {
				m.message = fmt.Sprintf("Error: %s", err)
			} else {
				m.message = "Updated successfully"
			}
		} else if m.currentView == viewVariables {
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter variable value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				err := m.variableService.SetCollectionVariable(m.collection, m.editingKey, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
				} else {
					m.message = fmt.Sprintf("Variable '%s' set", m.editingKey)
				}
				m.editingKey = ""
			}
		} else if m.currentView == viewHeaders && m.selectedRequest >= 0 {
			req := m.collection.Requests[m.selectedRequest]
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter header value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				err := m.requestService.SetHeader(req, m.editingKey, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
				} else {
					m.message = fmt.Sprintf("Header '%s' set", m.editingKey)
				}
				m.editingKey = ""
			}
		} else if m.currentView == viewQueryParams && m.selectedRequest >= 0 {
			req := m.collection.Requests[m.selectedRequest]
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter query parameter value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				err := m.requestService.SetQueryParam(req, m.editingKey, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
				} else {
					m.message = fmt.Sprintf("Query parameter '%s' set", m.editingKey)
				}
				m.editingKey = ""
			}
		} else if m.currentView == viewResponse && m.response != nil {
			// Save response body to file
			err := executor.SaveResponseBody(m.response, value)
			if err != nil {
				m.message = fmt.Sprintf("Error saving response: %s", err)
			} else {
				m.message = fmt.Sprintf("Response body saved to %s", value)
			}
		} else if m.currentView == viewEnvironments && m.editingField == editName {
			// Create new environment
			if m.viewingCollectionEnv {
				// Create collection environment
				err := m.environmentService.CreateCollectionEnvironment(m.collection, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
					return m, nil
				}
				newEnv, _ := m.environmentService.GetCollectionEnvironment(m.collection, value)
				m.environments = m.environmentService.ListCollectionEnvironments(m.collection)
				m.currentCollectionEnv = newEnv
				m.currentEnv = nil
				m.currentView = viewEnvironmentDetail
				m.message = fmt.Sprintf("Collection environment '%s' created", value)
			} else {
				// Create global environment
				err := m.environmentService.CreateGlobalEnvironment(value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
					return m, nil
				}
				// Get newly created environment
				newEnv, _ := m.environmentService.GetGlobalEnvironment(value)
				// Reload environments list
				envs, _ := m.environmentService.ListGlobalEnvironments()
				m.environments = envs
				m.currentEnv = newEnv
				m.currentCollectionEnv = nil
				m.currentView = viewEnvironmentDetail
				m.message = fmt.Sprintf("Environment '%s' created", value)
			}
		} else if m.currentView == viewEnvironmentVariables {
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter variable value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				if m.viewingCollectionEnv && m.currentCollectionEnv != nil {
					err := m.environmentService.SetCollectionEnvironmentVariable(m.collection, m.currentCollectionEnv.Name, m.editingKey, value)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Variable '%s' set in collection environment", m.editingKey)
					}
				} else if !m.viewingCollectionEnv && m.currentEnv != nil {
					err := m.environmentService.SetGlobalEnvironmentVariable(m.currentEnv.Name, m.editingKey, value)
					if err != nil {
						m.message = fmt.Sprintf("Error: %s", err)
					} else {
						m.message = fmt.Sprintf("Variable '%s' set", m.editingKey)
					}
				}
				m.editingKey = ""
			}
		} else if m.currentView == viewGlobalVariables {
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter global variable value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				err := m.variableService.SetGlobalVariable(m.editingKey, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
				} else {
					m.message = fmt.Sprintf("Global variable '%s' set", m.editingKey)
				}
				m.editingKey = ""
			}
		} else if m.currentView == viewEnvironmentDetail && m.editingField == editName {
			// Rename environment
			if m.viewingCollectionEnv && m.currentCollectionEnv != nil {
				oldName := m.currentCollectionEnv.Name
				err := m.environmentService.RenameCollectionEnvironment(m.collection, oldName, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
					return m, nil
				}
				m.environments = m.environmentService.ListCollectionEnvironments(m.collection)
				m.message = fmt.Sprintf("Collection environment renamed to '%s'", value)
			} else if !m.viewingCollectionEnv && m.currentEnv != nil {
				oldName := m.currentEnv.Name
				err := m.environmentService.RenameGlobalEnvironment(oldName, value)
				if err != nil {
					m.message = fmt.Sprintf("Error: %s", err)
					return m, nil
				}
				// Update current env reference
				m.currentEnv.Name = value
				// Update active environment name if this was active
				if m.collection.ActiveEnvironment == oldName {
					m.collection.ActiveEnvironment = value
				}
				// Reload environments list
				envs, _ := m.environmentService.ListGlobalEnvironments()
				m.environments = envs
				m.message = fmt.Sprintf("Environment renamed to '%s'", value)
			}
		}

		return m, nil

	case "esc":
		m.textInput.Blur()
		m.editing = false
		m.editingKey = ""
		m.message = ""
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case viewMain:
		return m.viewMain()
	case viewRequestList:
		return m.viewRequestList()
	case viewRequestDetail:
		return m.viewRequestDetail()
	case viewRequestEdit:
		return m.viewRequestEdit()
	case viewResponse:
		return m.viewResponse()
	case viewVariables:
		return m.viewVariables()
	case viewHelp:
		return m.viewHelp()
	case viewHeaders:
		return m.viewHeaders()
	case viewQueryParams:
		return m.viewQueryParams()
	case viewEnvironments:
		return m.viewEnvironments()
	case viewEnvironmentDetail:
		return m.viewEnvironmentDetail()
	case viewEnvironmentVariables:
		return m.viewEnvironmentVariables()
	case viewGlobalVariables:
		return m.viewGlobalVariables()
	}

	return ""
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))
)

// Helper function to get sorted variable keys
func getSortedVariableKeys(variables map[string]string) []string {
	keys := make([]string, 0, len(variables))
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m *Model) handleEnvironmentSelect() {
	if m.cursor < len(m.environments) {
		// Select an existing environment
		envName := m.environments[m.cursor]

		if m.viewingCollectionEnv {
			// Load collection environment
			collEnv := m.collection.GetCollectionEnvironment(envName)
			if collEnv == nil {
				m.message = fmt.Sprintf("Error loading collection environment: %s", envName)
				return
			}
			m.currentCollectionEnv = collEnv
			m.currentEnv = nil
		} else {
			// Load global environment
			env, err := m.environmentService.GetGlobalEnvironment(envName)
			if err != nil {
				m.message = fmt.Sprintf("Error loading environment: %s", err)
				return
			}
			m.currentEnv = env
			m.currentCollectionEnv = nil
		}

		m.selectedEnvIdx = m.cursor
		m.currentView = viewEnvironmentDetail
		m.cursor = 0
	} else if m.cursor == len(m.environments) {
		// Create new environment
		m.message = "Enter new environment name:"
		m.textInput.SetValue("")
		m.textInput.Focus()
		m.editing = true
		m.editingField = editName
	}
}

func (m *Model) startEditingEnvironmentVariable() {
	if m.currentEnv == nil && m.currentCollectionEnv == nil {
		return
	}
	m.editing = true
	m.textInput.Focus()
	m.editingField = editHeader
	m.textInput.SetValue("")
	m.message = "Enter variable name:"
}

func (m *Model) startEditingGlobalVariable() {
	varKeys := getSortedVariableKeys(m.globalConfig.Variables)

	// If cursor is on an existing variable, edit it
	if m.cursor >= 0 && m.cursor < len(varKeys) {
		key := varKeys[m.cursor]
		value := m.globalConfig.Variables[key]
		m.editingKey = key
		m.editing = true
		m.textInput.Focus()
		m.editingField = editHeader
		m.textInput.SetValue(value)
		m.message = fmt.Sprintf("Editing global variable '%s' (press enter to save):", key)
	} else {
		// Otherwise, create a new variable
		m.startEditingNewGlobalVariable()
	}
}

func (m *Model) startEditingNewGlobalVariable() {
	m.editing = true
	m.textInput.Focus()
	m.editingField = editHeader
	m.textInput.SetValue("")
	m.editingKey = ""
	m.message = "Enter global variable name:"
}

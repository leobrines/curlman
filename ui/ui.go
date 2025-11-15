package ui

import (
	"curlman/config"
	"curlman/environment"
	"curlman/executor"
	"curlman/exporter"
	"curlman/models"
	"curlman/openapi"
	"curlman/storage"
	"fmt"
	"sort"
	"strings"

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
	collection      *models.Collection
	currentView     view
	selectedRequest int
	selectedField   int
	cursor          int
	textInput       textinput.Model
	editing         bool
	editingField    editField
	editingKey      string
	response        *executor.Response
	message         string
	width           int
	height          int
	environments    []string
	currentEnv      *environment.Environment
	selectedEnvIdx  int
	globalConfig    *config.GlobalConfig
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

	return Model{
		collection: &models.Collection{
			Name:            "New Collection",
			Requests:        []*models.Request{},
			Variables:       make(map[string]string),
			EnvironmentVars: make(map[string]string),
		},
		currentView:  viewMain,
		textInput:    ti,
		environments: []string{},
		globalConfig: globalConfig,
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

		case "e":
			if m.currentView == viewMain {
				// Load environments list
				envs, err := environment.List()
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
			} else if m.currentView == viewEnvironmentDetail && m.currentEnv != nil {
				// Edit environment name
				m.message = "Enter new environment name:"
				m.textInput.SetValue(m.currentEnv.Name)
				m.textInput.Focus()
				m.editing = true
				m.editingField = editName
				return m, nil
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
			} else if m.currentView == viewEnvironmentDetail && m.currentEnv != nil {
				// Save environment
				err := m.currentEnv.Save()
				if err != nil {
					m.message = fmt.Sprintf("Error saving environment: %s", err)
				} else {
					m.message = fmt.Sprintf("Environment '%s' saved", m.currentEnv.Name)
					// Reload environments list
					envs, _ := environment.List()
					m.environments = envs
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
				cloned := m.collection.Requests[m.selectedRequest].Clone()
				m.collection.Requests = append(m.collection.Requests, cloned)
				m.message = "Request cloned successfully!"
				return m, nil
			}

		case "x":
			if m.currentView == viewRequestDetail && m.selectedRequest >= 0 {
				req := m.collection.Requests[m.selectedRequest]
				// Use merged variables (global + collection + environment)
				curl := exporter.ToCurlWithVariables(req, m.collection.GetAllVariables(m.globalConfig.Variables))
				m.message = "Curl command:\n" + curl
				return m, nil
			}

		case "a":
			if m.currentView == viewEnvironmentDetail && m.currentEnv != nil {
				// Activate environment
				m.collection.ActiveEnvironment = m.currentEnv.Name
				m.collection.SetEnvironmentVariables(m.currentEnv.Variables)
				m.message = fmt.Sprintf("Environment '%s' activated", m.currentEnv.Name)
				return m, nil
			}

		case "d":
			if m.currentView == viewRequestList && m.selectedRequest >= 0 && len(m.collection.Requests) > 0 {
				m.collection.Requests = append(
					m.collection.Requests[:m.selectedRequest],
					m.collection.Requests[m.selectedRequest+1:]...,
				)
				if m.selectedRequest >= len(m.collection.Requests) && m.selectedRequest > 0 {
					m.selectedRequest--
				}
				m.message = "Request deleted"
				return m, nil
			} else if m.currentView == viewEnvironments && m.cursor < len(m.environments) {
				// Delete environment
				envName := m.environments[m.cursor]
				err := environment.Delete(envName)
				if err != nil {
					m.message = fmt.Sprintf("Error deleting environment: %s", err)
				} else {
					// If deleted environment was active, clear it
					if m.collection.ActiveEnvironment == envName {
						m.collection.ClearEnvironmentVariables()
					}
					// Reload environments list
					envs, _ := environment.List()
					m.environments = envs
					if m.cursor >= len(m.environments) && m.cursor > 0 {
						m.cursor--
					}
					m.message = fmt.Sprintf("Environment '%s' deleted", envName)
				}
				return m, nil
			} else if m.currentView == viewEnvironmentDetail && m.currentEnv != nil {
				// Delete current environment
				envName := m.currentEnv.Name
				err := environment.Delete(envName)
				if err != nil {
					m.message = fmt.Sprintf("Error deleting environment: %s", err)
				} else {
					// If deleted environment was active, clear it
					if m.collection.ActiveEnvironment == envName {
						m.collection.ClearEnvironmentVariables()
					}
					// Reload environments list and go back
					envs, _ := environment.List()
					m.environments = envs
					m.currentEnv = nil
					m.currentView = viewEnvironments
					m.message = fmt.Sprintf("Environment '%s' deleted", envName)
				}
				return m, nil
			}
			if m.currentView == viewVariables && len(m.collection.Variables) > 0 {
				varKeys := getSortedVariableKeys(m.collection.Variables)
				if m.cursor >= 0 && m.cursor < len(varKeys) {
					keyToDelete := varKeys[m.cursor]
					delete(m.collection.Variables, keyToDelete)
					m.message = fmt.Sprintf("Variable '%s' deleted", keyToDelete)
					if m.cursor >= len(m.collection.Variables) && m.cursor > 0 {
						m.cursor--
					}
				}
				return m, nil
			}
			if m.currentView == viewGlobalVariables && len(m.globalConfig.Variables) > 0 {
				varKeys := getSortedVariableKeys(m.globalConfig.Variables)
				if m.cursor >= 0 && m.cursor < len(varKeys) {
					keyToDelete := varKeys[m.cursor]
					m.globalConfig.DeleteVariable(keyToDelete)
					m.globalConfig.Save()
					m.message = fmt.Sprintf("Global variable '%s' deleted", keyToDelete)
					if m.cursor >= len(m.globalConfig.Variables) && m.cursor > 0 {
						m.cursor--
					}
				}
				return m, nil
			}

		case "n":
			if m.currentView == viewRequestList {
				newReq := &models.Request{
					Name:        "New Request",
					Method:      "GET",
					URL:         "https://api.example.com",
					Headers:     make(map[string]string),
					QueryParams: make(map[string]string),
				}
				m.collection.Requests = append(m.collection.Requests, newReq)
				m.selectedRequest = len(m.collection.Requests) - 1
				m.currentView = viewRequestEdit
				m.message = "New request created"
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
			// Use merged variables (global + collection + environment)
			m.response = executor.Execute(req, m.collection.GetAllVariables(m.globalConfig.Variables))
			m.currentView = viewResponse
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
				collection, err := openapi.ImportFromFile(value)
				if err != nil {
					m.message = fmt.Sprintf("Error importing: %s", err)
				} else {
					m.collection = collection
					m.message = fmt.Sprintf("Imported %d requests from %s", len(collection.Requests), collection.Name)
				}
			} else if m.editingField == editPath { // Save collection
				err := openapi.SaveCollection(m.collection, value)
				if err != nil {
					m.message = fmt.Sprintf("Error saving: %s", err)
				} else {
					// Show the actual path where the file was saved
					storageDir, _ := storage.GetStorageDir()
					m.message = fmt.Sprintf("Collection saved to ~/.curlman/%s\n(Full path: %s/%s)", value, storageDir, value)
				}
			} else if m.editingField == editURL { // Load collection
				collection, err := openapi.LoadCollection(value)
				if err != nil {
					m.message = fmt.Sprintf("Error loading: %s", err)
				} else {
					m.collection = collection
					// Initialize EnvironmentVars if nil
					if m.collection.EnvironmentVars == nil {
						m.collection.EnvironmentVars = make(map[string]string)
					}
					// Load active environment if set
					if m.collection.ActiveEnvironment != "" {
						env, err := environment.Load(m.collection.ActiveEnvironment)
						if err == nil {
							m.collection.SetEnvironmentVariables(env.Variables)
						}
					}
					storageDir, _ := storage.GetStorageDir()
					m.message = fmt.Sprintf("Loaded collection: %s\n(From: %s/%s)", collection.Name, storageDir, value)
				}
			}
		} else if m.currentView == viewRequestEdit && m.selectedRequest >= 0 {
			req := m.collection.Requests[m.selectedRequest]
			switch m.editingField {
			case editName:
				req.Name = value
			case editMethod:
				req.Method = strings.ToUpper(value)
			case editURL:
				req.URL = value
			case editPath:
				req.Path = value
			case editBody:
				req.Body = value
			}
			m.message = "Updated successfully"
		} else if m.currentView == viewVariables {
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter variable value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				m.collection.Variables[m.editingKey] = value
				m.message = fmt.Sprintf("Variable '%s' set", m.editingKey)
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
				req.Headers[m.editingKey] = value
				m.message = fmt.Sprintf("Header '%s' set", m.editingKey)
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
				req.QueryParams[m.editingKey] = value
				m.message = fmt.Sprintf("Query parameter '%s' set", m.editingKey)
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
			if value == "" {
				m.message = "Environment name cannot be empty"
				return m, nil
			}
			if environment.Exists(value) {
				m.message = fmt.Sprintf("Environment '%s' already exists", value)
				return m, nil
			}
			newEnv := environment.NewEnvironment(value)
			err := newEnv.Save()
			if err != nil {
				m.message = fmt.Sprintf("Error creating environment: %s", err)
			} else {
				// Reload environments list
				envs, _ := environment.List()
				m.environments = envs
				m.currentEnv = newEnv
				m.currentView = viewEnvironmentDetail
				m.message = fmt.Sprintf("Environment '%s' created", value)
			}
		} else if m.currentView == viewEnvironmentVariables && m.currentEnv != nil {
			if m.editingKey == "" {
				m.editingKey = value
				m.message = "Enter variable value:"
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.editing = true
				return m, nil
			} else {
				m.currentEnv.Variables[m.editingKey] = value
				m.message = fmt.Sprintf("Variable '%s' set", m.editingKey)
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
				m.globalConfig.SetVariable(m.editingKey, value)
				m.globalConfig.Save()
				m.message = fmt.Sprintf("Global variable '%s' set", m.editingKey)
				m.editingKey = ""
			}
		} else if m.currentView == viewEnvironmentDetail && m.editingField == editName && m.currentEnv != nil {
			// Rename environment
			oldName := m.currentEnv.Name
			if value == "" {
				m.message = "Environment name cannot be empty"
				return m, nil
			}
			if value != oldName && environment.Exists(value) {
				m.message = fmt.Sprintf("Environment '%s' already exists", value)
				return m, nil
			}
			// Delete old environment file
			environment.Delete(oldName)
			// Update name and save
			m.currentEnv.Name = value
			err := m.currentEnv.Save()
			if err != nil {
				m.message = fmt.Sprintf("Error renaming environment: %s", err)
			} else {
				// Update active environment name if this was active
				if m.collection.ActiveEnvironment == oldName {
					m.collection.ActiveEnvironment = value
				}
				// Reload environments list
				envs, _ := environment.List()
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

func (m Model) viewMain() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("CurlMan - Postman CLI Alternative"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Collection: %s\n", m.collection.Name))
	s.WriteString(fmt.Sprintf("Requests: %d\n", len(m.collection.Requests)))
	s.WriteString(fmt.Sprintf("Variables: %d\n", len(m.collection.Variables)))

	// Display active environment
	if m.collection.ActiveEnvironment != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active Environment: %s (%d vars)\n",
			m.collection.ActiveEnvironment, len(m.collection.EnvironmentVars))))
	} else {
		s.WriteString(dimStyle.Render("Active Environment: None\n"))
	}

	// Display storage directory
	storageDir, err := storage.GetStorageDir()
	if err == nil {
		s.WriteString(dimStyle.Render(fmt.Sprintf("Storage: %s\n", storageDir)))
	}
	s.WriteString("\n")

	s.WriteString("Commands:\n")
	s.WriteString("  i - Import OpenAPI YAML\n")
	s.WriteString("  r - View Requests\n")
	s.WriteString("  v - Manage Variables\n")
	s.WriteString("  g - Manage Global Variables\n")
	s.WriteString("  e - Manage Environments\n")
	s.WriteString("  s - Save Collection (to ~/.curlman/)\n")
	s.WriteString("  l - Load Collection (from ~/.curlman/)\n")
	s.WriteString("  ? - Help\n")
	s.WriteString("  q - Quit\n\n")

	if m.editing {
		s.WriteString(m.message + "\n")
		s.WriteString(m.textInput.View() + "\n")
	} else if m.message != "" {
		s.WriteString(successStyle.Render(m.message) + "\n")
	}

	return s.String()
}

func (m Model) viewRequestList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Requests"))
	s.WriteString("\n\n")

	if len(m.collection.Requests) == 0 {
		s.WriteString(dimStyle.Render("No requests yet. Press 'n' to create one."))
	} else {
		for i, req := range m.collection.Requests {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s [%s] %s\n", cursor, req.Method, req.Name)))
			} else {
				s.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, req.Method, req.Name))
			}
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("n: new | enter: select | d: delete | esc: back"))
	s.WriteString("\n")

	if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewRequestDetail() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render(req.Name))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Method: %s\n", req.Method))
	s.WriteString(fmt.Sprintf("URL: %s\n", req.URL))
	if req.Path != "" {
		s.WriteString(fmt.Sprintf("Path: %s\n", req.Path))
	}
	s.WriteString(fmt.Sprintf("Full URL: %s\n\n", req.FullURL()))

	if len(req.Headers) > 0 {
		s.WriteString("Headers:\n")
		for k, v := range req.Headers {
			s.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		s.WriteString("\n")
	}

	if len(req.QueryParams) > 0 {
		s.WriteString("Query Parameters:\n")
		for k, v := range req.QueryParams {
			s.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		s.WriteString("\n")
	}

	if req.Body != "" {
		s.WriteString("Body:\n")
		s.WriteString(req.Body + "\n\n")
	}

	s.WriteString(dimStyle.Render("enter: execute | e: edit | h: headers | p: query params | c: clone | x: export curl | esc: back"))
	s.WriteString("\n")

	if m.message != "" {
		s.WriteString("\n" + m.message)
	}

	return s.String()
}

func (m Model) viewRequestEdit() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render("Edit Request"))
	s.WriteString("\n\n")

	fields := []string{
		fmt.Sprintf("Name: %s", req.Name),
		fmt.Sprintf("Method: %s", req.Method),
		fmt.Sprintf("URL: %s", req.URL),
		fmt.Sprintf("Path: %s", req.Path),
		fmt.Sprintf("Body: %s", req.Body),
	}

	for i, field := range fields {
		cursor := " "
		if i == m.selectedField {
			cursor = ">"
			s.WriteString(selectedStyle.Render(cursor + " " + field + "\n"))
		} else {
			s.WriteString(cursor + " " + field + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: edit | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.textInput.View())
	}

	if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewResponse() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Response"))
	s.WriteString("\n\n")

	if m.response != nil {
		s.WriteString(executor.FormatResponse(m.response))
	} else {
		s.WriteString("No response yet")
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("s: save body | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

// Helper function to get sorted variable keys
func getSortedVariableKeys(variables map[string]string) []string {
	keys := make([]string, 0, len(variables))
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (m Model) viewVariables() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Variables"))
	s.WriteString("\n\n")

	if len(m.collection.Variables) == 0 {
		s.WriteString(dimStyle.Render("No variables set. Press 'n' or 'enter' to add one."))
	} else {
		varKeys := getSortedVariableKeys(m.collection.Variables)
		for i, key := range varKeys {
			value := m.collection.Variables[key]
			line := fmt.Sprintf("%s = %s", key, value)

			if i == m.cursor {
				s.WriteString(selectedStyle.Render("> " + line))
			} else {
				s.WriteString("  " + line)
			}
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: edit | n: new | d: delete | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewHelp() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("CurlMan - Help"))
	s.WriteString("\n\n")

	// Display storage directory info
	storageDir, err := storage.GetStorageDir()
	if err == nil {
		s.WriteString(dimStyle.Render(fmt.Sprintf("Storage Directory: %s\n", storageDir)))
		s.WriteString(dimStyle.Render("All collections are saved/loaded from this directory by default.\n\n"))
	}

	s.WriteString("Main View:\n")
	s.WriteString("  i - Import OpenAPI YAML file\n")
	s.WriteString("  r - View and manage requests\n")
	s.WriteString("  v - Manage collection variables\n")
	s.WriteString("  e - Manage environments\n")
	s.WriteString("  s - Save collection to JSON (in ~/.curlman/)\n")
	s.WriteString("  l - Load collection from JSON (from ~/.curlman/)\n")
	s.WriteString("  q - Quit application\n\n")

	s.WriteString("Request List View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate requests\n")
	s.WriteString("  enter - View request details\n")
	s.WriteString("  n - Create new request\n")
	s.WriteString("  d - Delete selected request\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Request Detail View:\n")
	s.WriteString("  enter - Execute request\n")
	s.WriteString("  e - Edit request\n")
	s.WriteString("  c - Clone request\n")
	s.WriteString("  x - Export as curl command\n")
	s.WriteString("  esc - Back to request list\n\n")

	s.WriteString("Request Edit View:\n")
	s.WriteString("  ↑/↓ - Navigate fields\n")
	s.WriteString("  enter - Edit selected field\n")
	s.WriteString("  esc - Back to request detail\n\n")

	s.WriteString("Variables View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate variables\n")
	s.WriteString("  enter - Edit selected variable\n")
	s.WriteString("  n - Create new variable\n")
	s.WriteString("  d - Delete selected variable\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Environment Management:\n")
	s.WriteString("  Environments List:\n")
	s.WriteString("    enter - Select/create environment\n")
	s.WriteString("    d - Delete environment\n")
	s.WriteString("  Environment Detail:\n")
	s.WriteString("    a - Activate environment\n")
	s.WriteString("    v - Manage variables\n")
	s.WriteString("    e - Edit name\n")
	s.WriteString("    s - Save environment\n")
	s.WriteString("    d - Delete environment\n\n")

	s.WriteString("Variables Usage:\n")
	s.WriteString("  Use {{variable_name}} in requests\n")
	s.WriteString("  Environment variables override collection variables\n")
	s.WriteString("  Variables are injected before execution\n\n")

	s.WriteString(dimStyle.Render("Press 'esc' or 'q' to go back"))

	return s.String()
}

func (m Model) viewHeaders() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render("Headers"))
	s.WriteString("\n\n")

	if len(req.Headers) == 0 {
		s.WriteString(dimStyle.Render("No headers set. Press 'enter' to add one."))
	} else {
		for k, v := range req.Headers {
			s.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: add header | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewQueryParams() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render("Query Parameters"))
	s.WriteString("\n\n")

	if len(req.QueryParams) == 0 {
		s.WriteString(dimStyle.Render("No query parameters set. Press 'enter' to add one."))
	} else {
		for k, v := range req.QueryParams {
			s.WriteString(fmt.Sprintf("%s = %s\n", k, v))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: add query param | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

// Environment management helper functions
func (m *Model) handleEnvironmentSelect() {
	if m.cursor < len(m.environments) {
		// Select an existing environment
		envName := m.environments[m.cursor]
		env, err := environment.Load(envName)
		if err != nil {
			m.message = fmt.Sprintf("Error loading environment: %s", err)
			return
		}
		m.currentEnv = env
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
	if m.currentEnv == nil {
		return
	}
	m.editing = true
	m.textInput.Focus()
	m.editingField = editHeader
	m.textInput.SetValue("")
	m.message = "Enter variable name:"
}

func (m Model) viewEnvironments() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Environments"))
	s.WriteString("\n\n")

	if m.collection.ActiveEnvironment != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active: %s\n\n", m.collection.ActiveEnvironment)))
	}

	if len(m.environments) == 0 {
		s.WriteString(dimStyle.Render("No environments yet. Press 'enter' to create one."))
	} else {
		for i, envName := range m.environments {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			if envName == m.collection.ActiveEnvironment {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s (active)\n", cursor, envName)))
			} else if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s\n", cursor, envName)))
			} else {
				s.WriteString(fmt.Sprintf("%s %s\n", cursor, envName))
			}
		}
	}

	// Option to create new environment
	cursor := " "
	if m.cursor == len(m.environments) {
		cursor = ">"
		s.WriteString("\n" + selectedStyle.Render(cursor+" [Create New Environment]"))
	} else {
		s.WriteString("\n" + cursor + " [Create New Environment]")
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: select/create | d: delete | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	} else if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewEnvironmentDetail() string {
	if m.currentEnv == nil {
		return "No environment selected"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("Environment: %s", m.currentEnv.Name)))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Variables: %d\n\n", len(m.currentEnv.Variables)))

	if len(m.currentEnv.Variables) > 0 {
		for k, v := range m.currentEnv.Variables {
			s.WriteString(fmt.Sprintf("  %s = %s\n", k, v))
		}
		s.WriteString("\n")
	}

	s.WriteString(dimStyle.Render("a: activate | v: manage variables | e: edit name | s: save | d: delete | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	} else if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewEnvironmentVariables() string {
	if m.currentEnv == nil {
		return "No environment selected"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("Environment Variables: %s", m.currentEnv.Name)))
	s.WriteString("\n\n")

	if len(m.currentEnv.Variables) == 0 {
		s.WriteString(dimStyle.Render("No variables set. Press 'enter' to add one."))
	} else {
		for k, v := range m.currentEnv.Variables {
			s.WriteString(fmt.Sprintf("%s = %s\n", k, v))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: add variable | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	} else if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewGlobalVariables() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Global Variables"))
	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("Global variables are available across all collections"))
	s.WriteString("\n\n")

	if len(m.globalConfig.Variables) == 0 {
		s.WriteString(dimStyle.Render("No global variables set. Press 'n' or 'enter' to add one."))
	} else {
		varKeys := getSortedVariableKeys(m.globalConfig.Variables)
		for i, key := range varKeys {
			value := m.globalConfig.Variables[key]
			line := fmt.Sprintf("%s = %s", key, value)

			if i == m.cursor {
				s.WriteString(selectedStyle.Render("> " + line))
			} else {
				s.WriteString("  " + line)
			}
			s.WriteString("\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: edit | n: new | d: delete | esc: back"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
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

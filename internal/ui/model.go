package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leit0/curlman/internal/config"
	"github.com/leit0/curlman/internal/models"
	"github.com/leit0/curlman/internal/storage"
)

// Panel represents which panel is active
type Panel int

const (
	PanelEnvironments Panel = iota
	PanelCollections
	PanelRequests
)

// DetailPanel represents which panel is focused in the detail view
type DetailPanel int

const (
	DetailPanelSpec DetailPanel = iota // Left panel - spec request details
	DetailPanelManaged                 // Right panel - managed requests list
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeList ViewMode = iota
	ViewModeExecuting
	ViewModeEditing
	ViewModeInput
	ViewModeConfirm
	ViewModeRequestDetail  // Request detail view with split layout (spec + managed list)
	ViewModeCurlActions    // Curl edit/execute/copy/save view
	ViewModeRequestExpanded // Full view for truncated content
)

// Model represents the main application model
type Model struct {
	storage      *storage.Storage
	config       *config.AppConfig
	appConfig    *models.Config
	currentPanel Panel
	viewMode     ViewMode

	// Data
	environments   []*models.Environment
	collections    []*models.Collection
	managedRequests []models.Request // Saved requests (persistent)
	specRequests    []models.Request // Generated from OpenAPI (ephemeral)

	// Selected items
	selectedEnv  *models.Environment
	selectedColl *models.Collection
	selectedReq  *models.Request

	// Lists for rendering
	envList  list.Model
	collList list.Model
	reqList  list.Model

	// Execution state
	executionOutput string
	executing       bool
	spinner         spinner.Model

	// Input state
	textInput      textinput.Model
	inputPrompt    string
	confirmPrompt  string
	confirmAction  func(*Model) tea.Cmd

	// Request detail view state
	detailActionIndex  int              // Currently selected action in menu (deprecated - only Generate now)
	detailRelatedIndex int              // Selected related request in right panel
	detailFocusPanel   DetailPanel      // Which panel is focused (spec or managed list)
	expandedSection    string           // Which section is expanded ("body", "headers")
	relatedRequests    []models.Request // Filtered related requests (managed requests for this spec)

	// CurlActions view state
	curlActionIndex   int             // Selected action (0=Edit, 1=Execute, 2=Copy, 3=Save)
	curlActionSource  *models.Request // Source request (spec or managed)
	curlActionCommand string          // Current curl command being viewed/edited

	// UI state
	width      int
	height     int
	ctrlCCount int // Track Ctrl+C presses for double-tap exit
	err        error
}

// NewModel creates a new application model
func NewModel(store *storage.Storage, appConfig *config.AppConfig) *Model {
	// Create list models with default delegate
	envList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	envList.Title = "Environments"
	envList.SetShowStatusBar(false)
	envList.SetFilteringEnabled(false)

	collList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	collList.Title = "Collections"
	collList.SetShowStatusBar(false)
	collList.SetFilteringEnabled(false)

	reqList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	reqList.Title = "Requests"
	reqList.SetShowStatusBar(false)
	reqList.SetFilteringEnabled(false)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 500
	ti.Width = 80

	m := &Model{
		storage:      store,
		config:       appConfig,
		currentPanel: PanelEnvironments,
		viewMode:     ViewModeList,
		ctrlCCount:   0,
		envList:      envList,
		collList:     collList,
		reqList:      reqList,
		spinner:      s,
		textInput:    ti,
	}

	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	// Load initial data
	return tea.Batch(
		m.loadEnvironments(),
		m.loadCollections(),
		m.spinner.Tick,
	)
}

func (m *Model) loadEnvironments() tea.Cmd {
	return func() tea.Msg {
		envs, err := m.storage.ListEnvironments()
		if err != nil {
			return errMsg{err}
		}
		return environmentsLoadedMsg{envs}
	}
}

func (m *Model) loadCollections() tea.Cmd {
	return func() tea.Msg {
		colls, err := m.storage.ListCollections()
		if err != nil {
			return errMsg{err}
		}
		return collectionsLoadedMsg{colls}
	}
}

// List items for bubble tea lists
type envItem struct {
	env *models.Environment
}

func (i envItem) FilterValue() string { return i.env.Name }
func (i envItem) Title() string       { return i.env.Name }
func (i envItem) Description() string {
	return fmt.Sprintf("%d variables", len(i.env.Variables))
}

type collItem struct {
	coll *models.Collection
}

func (i collItem) FilterValue() string { return i.coll.Name }
func (i collItem) Title() string       { return i.coll.Name }
func (i collItem) Description() string {
	if i.coll.OpenAPIPath != "" {
		return i.coll.Description + " (OpenAPI)"
	}
	return i.coll.Description
}

type reqItem struct {
	req models.Request
}

func (i reqItem) FilterValue() string { return i.req.Name }
func (i reqItem) Title() string {
	// Add visual indicator for spec requests
	if !i.req.IsManaged {
		return "[Spec] " + i.req.Name
	}
	// Add warning for managed requests with missing operations
	if i.req.OpenAPIOperation != "" && !i.req.OperationExists {
		return "⚠ " + i.req.Name
	}
	return i.req.Name
}
func (i reqItem) Description() string { return i.req.Method + " " + i.req.URL }

// Messages
type errMsg struct{ err error }
type environmentsLoadedMsg struct{ envs []*models.Environment }
type collectionsLoadedMsg struct{ colls []*models.Collection }
type requestsLoadedMsg struct{ reqs []models.Request }
type loadOpenAPISpecMsg struct{ openAPIPath string }
type specRequestsLoadedMsg struct{ specReqs []models.Request }
type executionCompleteMsg struct{ output string }
type executionStartedMsg struct{}
type editorFinishedMsg struct {
	err      error
	filePath string
	editType string // "curl" or "request"
}

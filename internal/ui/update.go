package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leit0/curlman/internal/executor"
	"github.com/leit0/curlman/internal/models"
	"github.com/leit0/curlman/internal/openapi"
	"github.com/leit0/curlman/internal/parser"
)

// openEditor opens an external editor for the given file path
func openEditor(filePath string, editType string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, filePath) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err, filePath: filePath, editType: editType}
	})
}

// executeCurl executes a curl command using tea.ExecProcess for proper terminal handling
func executeCurl(req *models.Request, env *models.Environment) tea.Cmd {
	exec := executor.NewExecutor()

	// Prepare the command
	curlCmd, cmd, err := exec.PrepareCommand(req, env)
	if err != nil {
		return func() tea.Msg {
			return curlFinishedMsg{
				err:         err,
				curlCmd:     req.CurlCommand,
				output:      "",
				requestName: req.Name,
			}
		}
	}

	// Execute using tea.ExecProcess for proper terminal handling
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return curlFinishedMsg{
			err:         err,
			curlCmd:     curlCmd,
			output:      "",
			requestName: req.Name,
		}
	})
}

// Update handles all state updates
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Always update spinner
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		model, cmd := m.handleKeyPress(msg)
		cmds = append(cmds, cmd)
		return model, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update list sizes - account for borders (2 chars) and padding (4 chars) = 6 total width reduction
		// and borders (2 lines) and padding (2 lines) = 4 total height reduction
		panelWidth := (msg.Width - 6) / 3
		if panelWidth < 20 {
			panelWidth = 20
		}
		listWidth := panelWidth - 6   // Account for border (2) + padding (4)
		listHeight := msg.Height - 14 // Account for help text (~4 lines) + border (2) + padding (2) + margin (6)
		if listWidth < 10 {
			listWidth = 10
		}
		if listHeight < 5 {
			listHeight = 5
		}
		m.envList.SetSize(listWidth, listHeight)
		m.collList.SetSize(listWidth, listHeight)
		m.reqList.SetSize(listWidth, listHeight)
		return m, tea.Batch(cmds...)

	case environmentsLoadedMsg:
		m.environments = msg.envs
		m.updateEnvList()
		// Load app config to select environment
		if cfg, err := m.storage.LoadConfig(); err == nil {
			m.appConfig = cfg
			m.selectEnvironmentByID(cfg.SelectedEnvironmentID)
		}
		return m, tea.Batch(cmds...)

	case collectionsLoadedMsg:
		m.collections = msg.colls
		m.updateCollList()
		// Select collection if one is configured
		if m.appConfig != nil && m.appConfig.SelectedCollectionID != "" {
			m.selectCollectionByID(m.appConfig.SelectedCollectionID)
		}
		return m, tea.Batch(cmds...)

	case requestsLoadedMsg:
		m.updateReqList()
		return m, tea.Batch(cmds...)

	case loadOpenAPISpecMsg:
		return m, m.loadOpenAPISpec(msg.openAPIPath)

	case specRequestsLoadedMsg:
		m.specRequests = msg.specReqs
		m.updateReqList()
		return m, tea.Batch(cmds...)

	case executionStartedMsg:
		m.executing = true
		m.viewMode = ViewModeExecuting
		m.executionOutput = "Executing request...\n"
		cmds = append(cmds, m.doExecuteRequest())
		return m, tea.Batch(cmds...)

	case executionCompleteMsg:
		m.executionOutput = msg.output
		m.executing = false
		return m, tea.Batch(cmds...)

	case curlFinishedMsg:
		m.executing = false
		if msg.err != nil {
			// Format error output
			m.executionOutput = fmt.Sprintf("Request: %s %s\n\n", m.selectedReq.Method, msg.requestName)
			m.executionOutput += "--- Response ---\n"
			m.executionOutput += fmt.Sprintf("Command: %s\n\n", msg.curlCmd)
			m.executionOutput += fmt.Sprintf("Error: %v\n", msg.err)
		} else {
			// Format success output
			m.executionOutput = fmt.Sprintf("Request: %s %s\n\n", m.selectedReq.Method, msg.requestName)
			m.executionOutput += "--- Response ---\n"
			m.executionOutput += fmt.Sprintf("Command: %s\n\n", msg.curlCmd)
			m.executionOutput += "[Request completed successfully]\n"
			m.executionOutput += "\nNote: Output is displayed in the terminal above."
		}
		return m, tea.Batch(cmds...)

	case editorFinishedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("editor error: %w", msg.err)
		} else {
			// Handle based on edit type
			switch msg.editType {
			case "curl":
				// Read back the edited content from temp file
				content, err := os.ReadFile(msg.filePath)
				if err != nil {
					m.err = fmt.Errorf("failed to read edited file: %w", err)
				} else {
					m.curlActionCommand = string(content)
				}
				// Clean up temp file
				os.Remove(msg.filePath)

			case "request":
				// Reload the request from file
				if m.selectedReq != nil {
					reloadedReq, err := m.storage.LoadRequest(m.selectedColl.ID, m.selectedReq.ID)
					if err == nil {
						// Update the request in managedRequests
						for i := range m.managedRequests {
							if m.managedRequests[i].ID == m.selectedReq.ID {
								m.managedRequests[i] = *reloadedReq
								break
							}
						}
						m.selectedReq = reloadedReq
						m.updateReqList()
					} else {
						m.err = fmt.Errorf("failed to reload request: %w", err)
					}
				}
			}
		}
		return m, tea.Batch(cmds...)

	case errMsg:
		m.err = msg.err
		m.executing = false
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Ctrl+C for exit (double-tap)
	if msg.Type == tea.KeyCtrlC {
		m.ctrlCCount++
		if m.ctrlCCount >= 2 {
			return m, tea.Quit
		}
		return m, nil
	} else {
		m.ctrlCCount = 0 // Reset counter on other key press
	}

	// Handle ESC key
	if msg.Type == tea.KeyEsc {
		if m.viewMode == ViewModeExecuting {
			m.viewMode = ViewModeList
			m.executing = false
			return m, nil
		}
	}

	// Handle keys based on view mode
	switch m.viewMode {
	case ViewModeList:
		return m.handleListKeys(msg)
	case ViewModeExecuting:
		return m.handleExecutionKeys(msg)
	case ViewModeInput:
		return m.handleInputKeys(msg)
	case ViewModeConfirm:
		return m.handleConfirmKeys(msg)
	case ViewModeRequestDetail:
		return m.handleRequestDetailKeys(msg)
	case ViewModeCurlActions:
		return m.handleCurlActionsKeys(msg)
	case ViewModeRequestExpanded:
		return m.handleRequestExpandedKeys(msg)
	}

	return m, nil
}

func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "tab":
		// Switch panels
		m.currentPanel = (m.currentPanel + 1) % 3
		return m, nil

	case "shift+tab":
		// Switch panels backwards
		if m.currentPanel == 0 {
			m.currentPanel = 2
		} else {
			m.currentPanel--
		}
		return m, nil

	case "enter":
		return m.handleEnter()

	case "e":
		return m.handleEdit()

	case "a":
		return m.handleAdd()

	case "d", "delete":
		return m.handleDelete()

	case "s":
		return m.handleSave()

	case "c":
		return m.handleCopy()

	case "r":
		return m.handleRefresh()

	case "up", "k":
		return m.handleNavigateUp()

	case "down", "j":
		return m.handleNavigateDown()
	}

	return m, nil
}

func (m *Model) handleExecutionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only ESC is handled, which is already handled above
	return m, nil
}

func (m *Model) handleInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		// Process input
		value := m.textInput.Value()
		if value == "" {
			m.viewMode = ViewModeList
			return m, nil
		}

		// Determine what to add based on the prompt
		if m.inputPrompt == "Enter collection name:" {
			return m.addCollection(value)
		} else if m.inputPrompt == "Enter name for request:" {
			// Save curl as managed request with the given name
			return m.saveCurlAsManaged(value)
		} else {
			// Add request to collection
			return m.addRequestFromCurl(value)
		}

	case tea.KeyEsc:
		m.viewMode = ViewModeList
		m.textInput.Blur()
		return m, nil

	default:
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

func (m *Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.viewMode = ViewModeList
		if m.confirmAction != nil {
			return m, m.confirmAction(m)
		}
		return m, nil

	case "n", "N", "esc":
		m.viewMode = ViewModeList
		m.confirmAction = nil
		return m, nil
	}

	return m, nil
}

func (m *Model) handleRequestDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewModeList
		return m, nil

	case "tab":
		// Switch focus to right panel (managed list)
		if m.detailFocusPanel == DetailPanelSpec {
			m.detailFocusPanel = DetailPanelManaged
			// Initialize selection if there are managed requests
			if len(m.relatedRequests) > 0 && m.detailRelatedIndex == -1 {
				m.detailRelatedIndex = 0
			}
		} else {
			m.detailFocusPanel = DetailPanelSpec
		}
		return m, nil

	case "shift+tab":
		// Switch focus to left panel (spec detail)
		if m.detailFocusPanel == DetailPanelManaged {
			m.detailFocusPanel = DetailPanelSpec
		} else {
			m.detailFocusPanel = DetailPanelManaged
			// Initialize selection if there are managed requests
			if len(m.relatedRequests) > 0 && m.detailRelatedIndex == -1 {
				m.detailRelatedIndex = 0
			}
		}
		return m, nil

	case "left", "h":
		// Switch to left panel
		m.detailFocusPanel = DetailPanelSpec
		return m, nil

	case "right", "l":
		// Switch to right panel
		m.detailFocusPanel = DetailPanelManaged
		// Initialize selection if there are managed requests
		if len(m.relatedRequests) > 0 && m.detailRelatedIndex == -1 {
			m.detailRelatedIndex = 0
		}
		return m, nil

	case "up", "k":
		// Navigate managed requests up (only when right panel is focused)
		if m.detailFocusPanel == DetailPanelManaged && len(m.relatedRequests) > 0 {
			if m.detailRelatedIndex > 0 {
				m.detailRelatedIndex--
			}
		}
		return m, nil

	case "down", "j":
		// Navigate managed requests down (only when right panel is focused)
		if m.detailFocusPanel == DetailPanelManaged && len(m.relatedRequests) > 0 {
			if m.detailRelatedIndex == -1 {
				m.detailRelatedIndex = 0
			} else if m.detailRelatedIndex < len(m.relatedRequests)-1 {
				m.detailRelatedIndex++
			}
		}
		return m, nil

	case "enter":
		if m.detailFocusPanel == DetailPanelSpec {
			// Generate button pressed - open CurlActions view with generated curl
			return m.handleGenerateCurl()
		} else if m.detailFocusPanel == DetailPanelManaged {
			// Managed request selected - open CurlActions view with that request
			if m.detailRelatedIndex >= 0 && m.detailRelatedIndex < len(m.relatedRequests) {
				return m.handleOpenManagedRequest()
			}
		}
		return m, nil

	case "v":
		// Expand/view full content (only when left panel is focused)
		if m.detailFocusPanel == DetailPanelSpec {
			return m.handleExpandContent()
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleCurlActionsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to detail view
		m.viewMode = ViewModeRequestDetail
		return m, nil

	case "left", "h":
		// Navigate actions left
		if m.curlActionIndex > 0 {
			m.curlActionIndex--
		}
		return m, nil

	case "right", "l":
		// Navigate actions right (4 actions: Edit, Execute, Copy, Save)
		if m.curlActionIndex < 3 {
			m.curlActionIndex++
		}
		return m, nil

	case "enter":
		// Execute selected action
		return m.executeCurlAction()
	}

	return m, nil
}

func (m *Model) handleRequestExpandedKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Return to detail view
		m.viewMode = ViewModeRequestDetail
		m.expandedSection = ""
		return m, nil
	}

	return m, nil
}

func (m *Model) executeDetailAction() (tea.Model, tea.Cmd) {
	// This function is deprecated but kept for compatibility
	// The new flow uses executeCurlAction in ViewModeCurlActions
	if m.selectedReq == nil {
		return m, nil
	}

	switch m.detailActionIndex {
	case 0: // Execute (deprecated)
		return m, m.executeRequest()

	case 2: // Save (deprecated)
		if m.selectedReq.IsManaged {
			m.err = fmt.Errorf("request is already saved")
			return m, nil
		}
		return m, m.saveSpecRequest(m.selectedReq)

	case 3: // Copy (deprecated)
		return m, m.copyToClipboard(m.selectedReq.CurlCommand)
	}

	return m, nil
}

func (m *Model) handleGenerateCurl() (tea.Model, tea.Cmd) {
	// Generate curl command from spec request and open CurlActions view
	if m.selectedReq == nil {
		return m, nil
	}

	// Set up CurlActions view state
	m.curlActionSource = m.selectedReq
	m.curlActionCommand = m.selectedReq.CurlCommand
	m.curlActionIndex = 0 // Default to first action (Edit)
	m.viewMode = ViewModeCurlActions

	return m, nil
}

func (m *Model) handleOpenManagedRequest() (tea.Model, tea.Cmd) {
	// Open CurlActions view with selected managed request
	if m.detailRelatedIndex < 0 || m.detailRelatedIndex >= len(m.relatedRequests) {
		return m, nil
	}

	managedReq := m.relatedRequests[m.detailRelatedIndex]

	// Set up CurlActions view state
	m.curlActionSource = &managedReq
	m.curlActionCommand = managedReq.CurlCommand
	m.curlActionIndex = 0 // Default to first action (Edit)
	m.viewMode = ViewModeCurlActions

	return m, nil
}

func (m *Model) executeCurlAction() (tea.Model, tea.Cmd) {
	// Execute the selected action in CurlActions view
	switch m.curlActionIndex {
	case 0: // Edit (vim)
		return m.handleEditCurlWithVim()

	case 1: // Execute
		return m.handleExecuteCurl()

	case 2: // Copy
		return m, m.copyToClipboard(m.curlActionCommand)

	case 3: // Save
		return m.handleSaveCurl()
	}

	return m, nil
}

func (m *Model) handleEditCurlWithVim() (tea.Model, tea.Cmd) {
	// Create temp file and write curl command
	tmpFile, err := os.CreateTemp("", "curlman-*.curl")
	if err != nil {
		m.err = fmt.Errorf("failed to create temp file: %w", err)
		return m, nil
	}

	// Write curl command to temp file
	if _, err := tmpFile.WriteString(m.curlActionCommand); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		m.err = fmt.Errorf("failed to write to temp file: %w", err)
		return m, nil
	}
	tmpFile.Close()

	// Open editor (cleanup and reading will happen in message handler)
	return m, openEditor(tmpFile.Name(), "curl")
}

func (m *Model) handleExecuteCurl() (tea.Model, tea.Cmd) {
	// Execute the curl command
	if m.curlActionSource == nil {
		return m, nil
	}

	// Create a temporary request with the current curl command
	tempReq := *m.curlActionSource
	tempReq.CurlCommand = m.curlActionCommand
	m.selectedReq = &tempReq

	// Execute using existing executeRequest flow
	return m, m.executeRequest()
}

func (m *Model) handleSaveCurl() (tea.Model, tea.Cmd) {
	// Prompt for name and save as managed request
	if m.curlActionSource == nil {
		return m, nil
	}

	// If it's already a managed request, just update it
	if m.curlActionSource.IsManaged {
		return m, m.updateManagedRequest(m.curlActionSource.ID, m.curlActionCommand)
	}

	// Otherwise, prompt for name to save as new managed request
	m.viewMode = ViewModeInput
	m.inputPrompt = "Enter name for request:"
	m.textInput.SetValue("")
	m.textInput.Focus()

	return m, nil
}

func (m *Model) handleExpandContent() (tea.Model, tea.Cmd) {
	// Cycle through expandable sections: curl -> headers -> body -> (back to none)
	if m.expandedSection == "" {
		m.expandedSection = "curl"
	} else if m.expandedSection == "curl" {
		m.expandedSection = "headers"
	} else if m.expandedSection == "headers" {
		m.expandedSection = "body"
	} else {
		m.expandedSection = ""
	}

	if m.expandedSection != "" {
		m.viewMode = ViewModeRequestExpanded
	}

	return m, nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentPanel {
	case PanelEnvironments:
		// Select environment
		if len(m.environments) > 0 {
			idx := m.getCurrentIndex()
			if idx >= 0 && idx < len(m.environments) {
				m.selectedEnv = m.environments[idx]
				if m.appConfig != nil {
					m.appConfig.SetSelectedEnvironment(m.selectedEnv.ID)
					m.storage.SaveConfig(m.appConfig)
				}
			}
		}

	case PanelCollections:
		// Select collection and load requests
		if len(m.collections) > 0 {
			idx := m.getCurrentIndex()
			if idx >= 0 && idx < len(m.collections) {
				m.selectedColl = m.collections[idx]
				if m.appConfig != nil {
					m.appConfig.SetSelectedCollection(m.selectedColl.ID)
					m.storage.SaveConfig(m.appConfig)
				}
				// Load both managed and spec requests
				return m, m.loadRequestsForCollection()
			}
		}

	case PanelRequests:
		// Show request detail view
		allRequests := m.getAllRequests()
		if len(allRequests) > 0 {
			idx := m.getCurrentIndex()
			if idx >= 0 && idx < len(allRequests) {
				req := allRequests[idx]
				m.selectedReq = &req
				// Find related managed requests (same OpenAPI operation)
				m.relatedRequests = m.findRelatedRequests(&req)
				// Reset detail view state
				m.detailActionIndex = 0
				m.detailRelatedIndex = -1 // -1 means no related request selected
				m.detailFocusPanel = DetailPanelSpec // Default to left panel
				m.expandedSection = ""
				m.viewMode = ViewModeRequestDetail
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *Model) handleEdit() (tea.Model, tea.Cmd) {
	// Only handle editing requests for now
	if m.currentPanel != PanelRequests {
		return m, nil
	}

	allRequests := m.getAllRequests()
	if len(allRequests) == 0 {
		return m, nil
	}

	idx := m.getCurrentIndex()
	if idx < 0 || idx >= len(allRequests) {
		return m, nil
	}

	req := allRequests[idx]

	// Cannot edit spec requests - they must be saved first
	if !req.IsManaged {
		m.err = fmt.Errorf("cannot edit spec request - save it first with 's' key")
		return m, nil
	}

	// Open external editor
	return m, m.handleEditRequest(req)
}

func (m *Model) handleEditRequest(req models.Request) tea.Cmd {
	// Build the curl file path
	curlPath := filepath.Join(m.storage.GetCurlmanPath(), "requests", m.selectedColl.ID, fmt.Sprintf("%s.curl", req.ID))

	// Open editor
	return openEditor(curlPath, "request")
}

func (m *Model) handleAdd() (tea.Model, tea.Cmd) {
	switch m.currentPanel {
	case PanelCollections:
		// Add new collection
		m.viewMode = ViewModeInput
		m.inputPrompt = "Enter collection name:"
		m.textInput.SetValue("")
		m.textInput.Placeholder = "My API Collection"
		m.textInput.Focus()
		return m, nil

	case PanelRequests:
		// Add new request
		if m.selectedColl == nil {
			m.err = fmt.Errorf("select a collection first")
			return m, nil
		}

		m.viewMode = ViewModeInput
		m.inputPrompt = "Enter curl command:"
		m.textInput.SetValue("")
		m.textInput.Placeholder = "curl -X GET https://api.example.com/endpoint"
		m.textInput.Focus()
		return m, nil
	}

	return m, nil
}

func (m *Model) handleDelete() (tea.Model, tea.Cmd) {
	switch m.currentPanel {
	case PanelCollections:
		if len(m.collections) == 0 {
			return m, nil
		}

		idx := m.getCurrentIndex()
		if idx < 0 || idx >= len(m.collections) {
			return m, nil
		}

		coll := m.collections[idx]

		// Show confirmation dialog
		m.viewMode = ViewModeConfirm
		m.confirmPrompt = fmt.Sprintf("Delete collection '%s' and all its requests?", coll.Name)
		m.confirmAction = func(m *Model) tea.Cmd {
			return m.deleteCollection(coll.ID)
		}

		return m, nil

	case PanelRequests:
		allRequests := m.getAllRequests()
		if len(allRequests) == 0 {
			return m, nil
		}

		idx := m.getCurrentIndex()
		if idx < 0 || idx >= len(allRequests) {
			return m, nil
		}

		req := allRequests[idx]

		// Cannot delete spec requests
		if !req.IsManaged {
			m.err = fmt.Errorf("cannot delete spec request - it's generated from OpenAPI")
			return m, nil
		}

		// Show confirmation dialog
		m.viewMode = ViewModeConfirm
		m.confirmPrompt = fmt.Sprintf("Delete request '%s'?", req.Name)
		m.confirmAction = func(m *Model) tea.Cmd {
			return m.deleteRequest(req.ID)
		}

		return m, nil
	}

	return m, nil
}

func (m *Model) handleSave() (tea.Model, tea.Cmd) {
	// Only applicable to requests panel
	if m.currentPanel != PanelRequests {
		return m, nil
	}

	if m.selectedColl == nil {
		m.err = fmt.Errorf("select a collection first")
		return m, nil
	}

	allRequests := m.getAllRequests()
	if len(allRequests) == 0 {
		return m, nil
	}

	idx := m.getCurrentIndex()
	if idx < 0 || idx >= len(allRequests) {
		return m, nil
	}

	req := allRequests[idx]

	// Only save spec requests (managed requests are already saved)
	if req.IsManaged {
		m.err = fmt.Errorf("request is already saved")
		return m, nil
	}

	// Convert spec request to managed request
	return m, m.saveSpecRequest(&req)
}

func (m *Model) handleCopy() (tea.Model, tea.Cmd) {
	// Only applicable to requests panel
	if m.currentPanel != PanelRequests {
		return m, nil
	}

	allRequests := m.getAllRequests()
	if len(allRequests) == 0 {
		return m, nil
	}

	idx := m.getCurrentIndex()
	if idx < 0 || idx >= len(allRequests) {
		return m, nil
	}

	req := allRequests[idx]

	// Copy curl command to clipboard (using xclip, pbcopy, or similar)
	return m, m.copyToClipboard(req.CurlCommand)
}

func (m *Model) handleRefresh() (tea.Model, tea.Cmd) {
	// Only applicable when a collection with OpenAPI is selected
	if m.selectedColl == nil || m.selectedColl.OpenAPIPath == "" {
		return m, nil
	}

	// Reload OpenAPI spec
	return m, m.loadOpenAPISpec(m.selectedColl.OpenAPIPath)
}

func (m *Model) saveSpecRequest(req *models.Request) tea.Cmd {
	return func() tea.Msg {
		// Create a new managed request from the spec request
		managedReq := models.NewRequest(req.Name, req.CurlCommand, m.selectedColl.ID)
		managedReq.Description = req.Description
		managedReq.Method = req.Method
		managedReq.URL = req.URL
		managedReq.Headers = req.Headers
		managedReq.Body = req.Body
		managedReq.OpenAPIOperation = req.OpenAPIOperation
		managedReq.OperationExists = true

		// Save to storage
		if err := m.storage.SaveRequest(managedReq); err != nil {
			return errMsg{err: fmt.Errorf("failed to save request: %w", err)}
		}

		// Update collection
		m.selectedColl.AddRequest(*managedReq)
		m.managedRequests = m.selectedColl.Requests

		return requestsLoadedMsg{}
	}
}

func (m *Model) saveCurlAsManaged(name string) (tea.Model, tea.Cmd) {
	// Save the current curl command as a new managed request
	if m.curlActionSource == nil {
		m.viewMode = ViewModeCurlActions
		return m, nil
	}

	// Create a new managed request from the source request
	managedReq := models.NewRequest(name, m.curlActionCommand, m.selectedColl.ID)
	managedReq.Description = m.curlActionSource.Description
	managedReq.Method = m.curlActionSource.Method
	managedReq.URL = m.curlActionSource.URL
	managedReq.Headers = m.curlActionSource.Headers
	managedReq.Body = m.curlActionSource.Body
	managedReq.OpenAPIOperation = m.curlActionSource.OpenAPIOperation
	managedReq.OperationExists = true

	// Return to detail view after save
	m.viewMode = ViewModeRequestDetail
	m.textInput.Blur()

	// Refresh related requests
	if m.selectedReq != nil {
		m.relatedRequests = m.findRelatedRequests(m.selectedReq)
	}

	return m, func() tea.Msg {
		// Save to storage
		if err := m.storage.SaveRequest(managedReq); err != nil {
			return errMsg{err: fmt.Errorf("failed to save request: %w", err)}
		}

		// Update collection
		m.selectedColl.AddRequest(*managedReq)
		m.managedRequests = m.selectedColl.Requests

		// After saving, update the view
		// Return to detail view
		return requestsLoadedMsg{}
	}
}

func (m *Model) updateManagedRequest(requestID string, newCurl string) tea.Cmd {
	return func() tea.Msg {
		// Load the request
		req, err := m.storage.LoadRequest(m.selectedColl.ID, requestID)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load request: %w", err)}
		}

		// Update the curl command
		req.CurlCommand = newCurl

		// Save the updated request
		if err := m.storage.SaveRequest(req); err != nil {
			return errMsg{err: fmt.Errorf("failed to update request: %w", err)}
		}

		// Update in collection
		for i := range m.selectedColl.Requests {
			if m.selectedColl.Requests[i].ID == requestID {
				m.selectedColl.Requests[i] = *req
				break
			}
		}
		m.managedRequests = m.selectedColl.Requests

		return requestsLoadedMsg{}
	}
}

func (m *Model) copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		// Try different clipboard utilities
		var cmd *exec.Cmd

		// Linux: xclip or xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("pbcopy"); err == nil {
			// macOS: pbcopy
			cmd = exec.Command("pbcopy")
		} else {
			return errMsg{err: fmt.Errorf("no clipboard utility found (install xclip, xsel, or use macOS)")}
		}

		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return errMsg{err: fmt.Errorf("failed to copy to clipboard: %w", err)}
		}

		// Success - maybe show a brief message
		return nil
	}
}

func (m *Model) deleteRequest(requestID string) tea.Cmd {
	return func() tea.Msg {
		if err := m.storage.DeleteRequest(m.selectedColl.ID, requestID); err != nil {
			return errMsg{err: fmt.Errorf("failed to delete request: %w", err)}
		}

		// Reload collection to update managed requests
		coll, err := m.storage.LoadCollection(m.selectedColl.ID)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to reload collection: %w", err)}
		}

		m.selectedColl = coll
		m.managedRequests = coll.Requests

		return requestsLoadedMsg{}
	}
}

func (m *Model) handleNavigateUp() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentPanel {
	case PanelEnvironments:
		m.envList, cmd = m.envList.Update(tea.KeyMsg{Type: tea.KeyUp})
	case PanelCollections:
		m.collList, cmd = m.collList.Update(tea.KeyMsg{Type: tea.KeyUp})
	case PanelRequests:
		m.reqList, cmd = m.reqList.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	return m, cmd
}

func (m *Model) handleNavigateDown() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentPanel {
	case PanelEnvironments:
		m.envList, cmd = m.envList.Update(tea.KeyMsg{Type: tea.KeyDown})
	case PanelCollections:
		m.collList, cmd = m.collList.Update(tea.KeyMsg{Type: tea.KeyDown})
	case PanelRequests:
		m.reqList, cmd = m.reqList.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	return m, cmd
}

func (m *Model) getCurrentIndex() int {
	switch m.currentPanel {
	case PanelEnvironments:
		return m.envList.Index()
	case PanelCollections:
		return m.collList.Index()
	case PanelRequests:
		return m.reqList.Index()
	}
	return 0
}

func (m *Model) executeRequest() tea.Cmd {
	return func() tea.Msg {
		// First send a message to indicate execution started
		return executionStartedMsg{}
	}
}

func (m *Model) doExecuteRequest() tea.Cmd {
	if m.selectedReq == nil {
		return func() tea.Msg {
			return errMsg{fmt.Errorf("no request selected")}
		}
	}

	req := *m.selectedReq
	env := m.selectedEnv

	// Use tea.ExecProcess for proper terminal handling
	return executeCurl(&req, env)
}

func (m *Model) selectEnvironmentByID(id string) {
	for _, env := range m.environments {
		if env.ID == id {
			m.selectedEnv = env
			break
		}
	}
}

func (m *Model) selectCollectionByID(id string) {
	for _, coll := range m.collections {
		if coll.ID == id {
			m.selectedColl = coll
			// Load both managed and spec requests
			m.loadRequestsForCollection()()
			break
		}
	}
}

// getAllRequests returns merged spec and managed requests
func (m *Model) getAllRequests() []models.Request {
	allRequests := append([]models.Request{}, m.specRequests...)
	allRequests = append(allRequests, m.managedRequests...)
	return allRequests
}

// findRelatedRequests finds managed requests with the same OpenAPI operation
func (m *Model) findRelatedRequests(req *models.Request) []models.Request {
	if req.OpenAPIOperation == "" {
		return nil
	}

	var related []models.Request
	for _, managedReq := range m.managedRequests {
		// Skip if it's the same request
		if managedReq.ID == req.ID {
			continue
		}
		// Match by OpenAPI operation
		if managedReq.OpenAPIOperation == req.OpenAPIOperation {
			related = append(related, managedReq)
		}
	}
	return related
}

func (m *Model) updateEnvList() {
	items := make([]list.Item, len(m.environments))
	for i, env := range m.environments {
		items[i] = envItem{env: env}
	}
	m.envList.SetItems(items)
}

func (m *Model) updateCollList() {
	items := make([]list.Item, len(m.collections))
	for i, coll := range m.collections {
		items[i] = collItem{coll: coll}
	}
	m.collList.SetItems(items)
}

func (m *Model) updateReqList() {
	// Merge spec requests and managed requests
	allRequests := append([]models.Request{}, m.specRequests...)
	allRequests = append(allRequests, m.managedRequests...)

	items := make([]list.Item, len(allRequests))
	for i, req := range allRequests {
		items[i] = reqItem{req: req}
	}
	m.reqList.SetItems(items)
}

func (m *Model) loadRequestsForCollection() tea.Cmd {
	return func() tea.Msg {
		// Load managed requests from collection
		m.managedRequests = m.selectedColl.Requests
		m.specRequests = nil // Clear spec requests

		// If collection has OpenAPI path, load spec requests
		if m.selectedColl.OpenAPIPath != "" {
			return loadOpenAPISpecMsg{openAPIPath: m.selectedColl.OpenAPIPath}
		}

		return requestsLoadedMsg{reqs: m.managedRequests}
	}
}

func (m *Model) loadOpenAPISpec(openAPIPath string) tea.Cmd {
	return func() tea.Msg {
		// Parse OpenAPI file to generate spec requests
		parser := openapi.NewOpenAPIParser()
		specReqs, err := parser.ParseFileToSpecRequests(openAPIPath)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load OpenAPI spec: %w", err)}
		}

		// Validate managed requests against current spec
		for i := range m.managedRequests {
			if m.managedRequests[i].OpenAPIOperation != "" {
				exists, _ := parser.ValidateOperation(openAPIPath, m.managedRequests[i].OpenAPIOperation)
				m.managedRequests[i].OperationExists = exists
			}
		}

		return specRequestsLoadedMsg{specReqs: specReqs}
	}
}

func (m *Model) addCollection(name string) (tea.Model, tea.Cmd) {
	// Create new collection
	coll := models.NewCollection(name, "")

	// Save to storage
	if err := m.storage.SaveCollection(coll); err != nil {
		m.err = fmt.Errorf("failed to save collection: %w", err)
		m.viewMode = ViewModeList
		return m, nil
	}

	// Reload collections
	m.viewMode = ViewModeList
	m.textInput.Blur()

	return m, m.loadCollections()
}

func (m *Model) deleteCollection(collectionID string) tea.Cmd {
	return func() tea.Msg {
		if err := m.storage.DeleteCollection(collectionID); err != nil {
			return errMsg{err: fmt.Errorf("failed to delete collection: %w", err)}
		}

		// Reload collections
		colls, err := m.storage.ListCollections()
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to reload collections: %w", err)}
		}

		// Clear selected collection if it was deleted
		if m.selectedColl != nil && m.selectedColl.ID == collectionID {
			m.selectedColl = nil
			m.managedRequests = nil
			m.specRequests = nil
		}

		return collectionsLoadedMsg{colls: colls}
	}
}

func (m *Model) addRequestFromCurl(curlCmd string) (tea.Model, tea.Cmd) {
	// Parse curl command
	curlParser := parser.NewCurlParser()
	req, err := curlParser.Parse(curlCmd)
	if err != nil {
		m.err = fmt.Errorf("failed to parse curl command: %w", err)
		m.viewMode = ViewModeList
		return m, nil
	}

	// Set collection ID
	req.CollectionID = m.selectedColl.ID

	// Save request to storage
	if err := m.storage.SaveRequest(req); err != nil {
		m.err = fmt.Errorf("failed to save request: %w", err)
		m.viewMode = ViewModeList
		return m, nil
	}

	// Update collection and managed requests
	m.selectedColl.AddRequest(*req)
	m.managedRequests = m.selectedColl.Requests
	m.updateReqList()

	m.viewMode = ViewModeList
	m.textInput.Blur()

	return m, nil
}

// Error returns the error if any
func (m *Model) Error() error {
	return m.err
}

func (m Model) String() string {
	return fmt.Sprintf("Model{panel: %d, mode: %d}", m.currentPanel, m.viewMode)
}

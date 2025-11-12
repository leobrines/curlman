package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leit0/curlman/internal/executor"
	"github.com/leit0/curlman/internal/models"
	"github.com/leit0/curlman/internal/openapi"
	"github.com/leit0/curlman/internal/parser"
)

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
		// Update list sizes
		listWidth := msg.Width / 3
		listHeight := msg.Height - 10
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

	case editorClosedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("editor error: %w", msg.err)
		} else {
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
		// Execute request
		allRequests := m.getAllRequests()
		if len(allRequests) > 0 {
			idx := m.getCurrentIndex()
			if idx >= 0 && idx < len(allRequests) {
				req := allRequests[idx]
				m.selectedReq = &req
				return m, m.executeRequest()
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
	return m, m.openEditor(req)
}

func (m *Model) openEditor(req models.Request) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Build the curl file path
	curlPath := filepath.Join(m.storage.GetCurlmanPath(), "requests", m.selectedColl.ID, fmt.Sprintf("%s.curl", req.ID))

	return func() tea.Msg {
		// Suspend bubble tea to run the editor
		cmd := exec.Command(editor, curlPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return editorClosedMsg{err: err}
		}

		return editorClosedMsg{err: nil}
	}
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

	return func() tea.Msg {
		// Create executor and execute the request
		exec := executor.NewExecutor()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := exec.ExecuteRequest(ctx, &req, env)
		if err != nil {
			return errMsg{fmt.Errorf("execution failed: %w", err)}
		}

		// Format output
		output := fmt.Sprintf("Request: %s %s\n\n", req.Method, req.Name)
		output += "--- Response ---\n"
		output += result

		return executionCompleteMsg{output: output}
	}
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

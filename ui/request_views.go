package ui

import (
	"curlman/storage"
	"fmt"
	"strings"
)

func (m Model) viewMain() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("CurlMan - Postman CLI Alternative"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Collection: %s\n", m.collection.Name))
	s.WriteString(fmt.Sprintf("Requests: %d\n", len(m.collection.Requests)))
	s.WriteString(fmt.Sprintf("Variables: %d\n", len(m.collection.Variables)))

	// Display active global environment
	if m.collection.ActiveEnvironment != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active Global Environment: %s (%d vars)\n",
			m.collection.ActiveEnvironment, len(m.collection.EnvironmentVars))))
	} else {
		s.WriteString(dimStyle.Render("Active Global Environment: None\n"))
	}

	// Display active collection environment
	if m.collection.ActiveCollectionEnv != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active Collection Environment: %s (%d vars)\n",
			m.collection.ActiveCollectionEnv, len(m.collection.CollectionEnvVars))))
	} else {
		s.WriteString(dimStyle.Render("Active Collection Environment: None\n"))
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

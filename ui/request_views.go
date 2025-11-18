package ui

import (
	"github.com/leobrines/curlman/storage"
	"fmt"
	"strings"
)

func (m Model) viewMain() string {
	var s strings.Builder

	// Handle collection selection mode
	if m.collectionSelectMode {
		s.WriteString(titleStyle.Render("Select Collection"))
		s.WriteString("\n\n")

		if len(m.availableCollections) == 0 {
			s.WriteString(dimStyle.Render("No collections available."))
			s.WriteString("\n")
		} else {
			for i, name := range m.availableCollections {
				cursor := "  "
				// Remove .json extension for display
				displayName := name
				if len(name) > 5 && name[len(name)-5:] == ".json" {
					displayName = name[:len(name)-5]
				}
				if i == m.collectionCursor {
					cursor = "> "
					s.WriteString(selectedStyle.Render(cursor + displayName) + "\n")
				} else {
					s.WriteString(cursor + displayName + "\n")
				}
			}
		}

		s.WriteString("\n")
		s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | esc: back"))
		s.WriteString("\n")

		if m.message != "" {
			s.WriteString("\n" + errorStyle.Render(m.message) + "\n")
		}

		return s.String()
	}

	// Handle collection detail view (when collection is active)
	if m.hasActiveCollection {
		s.WriteString(titleStyle.Render(m.collection.Name))
		s.WriteString("\n\n")

		s.WriteString(fmt.Sprintf("Requests: %d\n", len(m.collection.Requests)))
		s.WriteString(fmt.Sprintf("Variables: %d\n", len(m.collection.Variables)))
		s.WriteString(fmt.Sprintf("Environments: %d\n", len(m.collection.Environments)))

		// Display active global environment
		if m.collection.ActiveEnvironment != "" {
			s.WriteString(successStyle.Render(fmt.Sprintf("Active Global Env: %s (%d vars)\n",
				m.collection.ActiveEnvironment, len(m.collection.EnvironmentVars))))
		}

		// Display active collection environment
		if m.collection.ActiveCollectionEnv != "" {
			s.WriteString(successStyle.Render(fmt.Sprintf("Active Collection Env: %s (%d vars)\n",
				m.collection.ActiveCollectionEnv, len(m.collection.CollectionEnvVars))))
		}

		s.WriteString("\n")

		// Collection detail menu items
		menuItems := []string{
			"View Requests",
			"Manage Environments",
			"Manage Variables",
			"Edit Collection Name",
			"Save Collection",
			"Delete Collection",
		}

		for i, item := range menuItems {
			cursor := "  "
			if i == m.mainMenuCursor {
				cursor = "> "
				s.WriteString(selectedStyle.Render(cursor + item) + "\n")
			} else {
				s.WriteString(cursor + item + "\n")
			}
		}

		s.WriteString("\n")
		s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | esc: back to main menu"))
		s.WriteString("\n")

		if m.confirmingDelete {
			s.WriteString("\n" + errorStyle.Render(m.message) + "\n")
		} else if m.editing {
			s.WriteString("\n" + m.message + "\n")
			s.WriteString(m.textInput.View() + "\n")
		} else if m.message != "" {
			s.WriteString("\n" + successStyle.Render(m.message) + "\n")
		}

		return s.String()
	}

	// Main menu (no active collection)
	s.WriteString(titleStyle.Render("CurlMan - Main Menu"))
	s.WriteString("\n\n")

	// Display storage directory
	storageDir, err := storage.GetStorageDir()
	if err == nil {
		s.WriteString(dimStyle.Render(fmt.Sprintf("Storage: %s", storageDir)) + "\n")
	}

	// Display available collections count
	s.WriteString(fmt.Sprintf("Available Collections: %d\n", len(m.availableCollections)))
	s.WriteString("\n")

	// Main menu items
	menuItems := []string{
		"Import OpenAPI YAML",
		fmt.Sprintf("Select Collection (%d)", len(m.availableCollections)),
		"Manage Global Variables",
		"Manage Environments",
		"Help",
		"Quit",
	}

	for i, item := range menuItems {
		cursor := "  "
		if i == m.mainMenuCursor {
			cursor = "> "
			s.WriteString(selectedStyle.Render(cursor + item) + "\n")
		} else {
			s.WriteString(cursor + item + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | q: quit"))
	s.WriteString("\n")

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View() + "\n")
	} else if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message) + "\n")
	}

	return s.String()
}

func (m Model) viewRequestList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Requests"))
	s.WriteString("\n\n")

	if len(m.collection.Requests) == 0 {
		s.WriteString(dimStyle.Render("No requests yet."))
		s.WriteString("\n\n")
		// Show "Create New" option
		cursor := "> "
		s.WriteString(selectedStyle.Render(cursor + "[Create New Request]") + "\n")
	} else {
		for i, req := range m.collection.Requests {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s[%s] %s\n", cursor, req.Method, req.Name)))
			} else {
				s.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, req.Method, req.Name))
			}
		}
		// Add "Create New" option at the end
		cursor := "  "
		if m.cursor == len(m.collection.Requests) {
			cursor = "> "
			s.WriteString(selectedStyle.Render(cursor + "[Create New Request]") + "\n")
		} else {
			s.WriteString(cursor + "[Create New Request]\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | d: delete | esc: back"))
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

	// Action menu as a selectable list
	s.WriteString("Actions:\n")
	actions := []string{
		"Execute Request",
		"Edit Request",
		"Manage Headers",
		"Manage Query Params",
		"Clone Request",
		"Export to cURL",
	}

	for i, action := range actions {
		cursor := "  "
		if i == m.detailActionCursor {
			cursor = "> "
			s.WriteString(selectedStyle.Render(cursor + action) + "\n")
		} else {
			s.WriteString(cursor + action + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | esc: back"))
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

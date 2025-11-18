package ui

import (
	"fmt"
	"strings"
)

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

func (m Model) viewGlobalVariables() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Global Variables"))
	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("Global variables are available across all collections"))
	s.WriteString("\n\n")

	if len(m.globalConfig.Variables) == 0 {
		// Show only "Create New Variable" option when empty
		if m.cursor == 0 {
			s.WriteString(selectedStyle.Render("> Create New Variable"))
		} else {
			s.WriteString("  Create New Variable")
		}
		s.WriteString("\n")
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
		// Add "Create New Variable" option at the end
		if m.cursor == len(varKeys) {
			s.WriteString(selectedStyle.Render("> Create New Variable"))
		} else {
			s.WriteString("  Create New Variable")
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | enter: select | esc: back"))
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

func (m Model) viewGlobalVariableDetail() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Global Variable Detail"))
	s.WriteString("\n\n")

	// Show the selected variable
	if m.editingKey != "" {
		value := m.globalConfig.Variables[m.editingKey]
		s.WriteString(fmt.Sprintf("Name:  %s\n", m.editingKey))
		s.WriteString(fmt.Sprintf("Value: %s\n", value))
	}

	s.WriteString("\n")

	// Action menu as a selectable list
	s.WriteString("Actions:\n")
	actions := []string{
		"Edit Value",
		"Rename Variable",
		"Delete Variable",
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

	if m.editing {
		s.WriteString("\n" + m.message + "\n")
		s.WriteString(m.textInput.View())
	}

	if m.message != "" && !m.editing {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

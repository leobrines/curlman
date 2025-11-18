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
		s.WriteString(dimStyle.Render("No variables set."))
		s.WriteString("\n")
	} else {
		varKeys := getSortedVariableKeys(m.collection.Variables)
		for i, key := range varKeys {
			value := m.collection.Variables[key]
			line := fmt.Sprintf("%s = %s", key, value)

			if i == m.cursor && !m.variableActionFocus {
				s.WriteString(selectedStyle.Render("> " + line))
			} else if i == m.cursor {
				s.WriteString("  " + line + " ←")
			} else {
				s.WriteString("  " + line)
			}
			s.WriteString("\n")
		}
	}

	// Actions menu
	s.WriteString("\nActions:\n")
	actions := []string{
		"Add New Variable",
		"Edit Selected",
		"Delete Selected",
	}

	for i, action := range actions {
		if i == m.variableActionCursor && m.variableActionFocus {
			s.WriteString(selectedStyle.Render("> " + action) + "\n")
		} else {
			s.WriteString("  " + action + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | tab: switch section | enter: select | esc: back"))
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
		s.WriteString(dimStyle.Render("No global variables set."))
		s.WriteString("\n")
	} else {
		varKeys := getSortedVariableKeys(m.globalConfig.Variables)
		for i, key := range varKeys {
			value := m.globalConfig.Variables[key]
			line := fmt.Sprintf("%s = %s", key, value)

			if i == m.cursor && !m.variableActionFocus {
				s.WriteString(selectedStyle.Render("> " + line))
			} else if i == m.cursor {
				s.WriteString("  " + line + " ←")
			} else {
				s.WriteString("  " + line)
			}
			s.WriteString("\n")
		}
	}

	// Actions menu
	s.WriteString("\nActions:\n")
	actions := []string{
		"Add New Variable",
		"Edit Selected",
		"Delete Selected",
	}

	for i, action := range actions {
		if i == m.variableActionCursor && m.variableActionFocus {
			s.WriteString(selectedStyle.Render("> " + action) + "\n")
		} else {
			s.WriteString("  " + action + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate | tab: switch section | enter: select | esc: back"))
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

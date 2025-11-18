package ui

import (
	"fmt"
	"strings"
)

func (m Model) viewEnvironments() string {
	var s strings.Builder

	if m.viewingCollectionEnv {
		s.WriteString(titleStyle.Render("Collection Environments"))
	} else {
		s.WriteString(titleStyle.Render("Global Environments"))
	}
	s.WriteString("\n\n")

	if m.viewingCollectionEnv && m.collection.ActiveCollectionEnv != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active: %s\n\n", m.collection.ActiveCollectionEnv)))
	} else if !m.viewingCollectionEnv && m.collection.ActiveEnvironment != "" {
		s.WriteString(successStyle.Render(fmt.Sprintf("Active: %s\n\n", m.collection.ActiveEnvironment)))
	}

	if len(m.environments) == 0 {
		s.WriteString(dimStyle.Render("No environments yet."))
	} else {
		activeEnv := ""
		if m.viewingCollectionEnv {
			activeEnv = m.collection.ActiveCollectionEnv
		} else {
			activeEnv = m.collection.ActiveEnvironment
		}

		for i, envName := range m.environments {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			if envName == activeEnv {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s (active)\n", cursor, envName)))
			} else if i == m.cursor {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s\n", cursor, envName)))
			} else {
				s.WriteString(fmt.Sprintf("%s %s\n", cursor, envName))
			}
		}
	}

	s.WriteString("\n")

	// Action menu as a selectable list
	s.WriteString("Actions:\n")
	actions := []string{
		"View Details",
		"Activate Environment",
		"Create New Environment",
		"Delete Environment",
		"Toggle Global/Collection",
	}

	for i, action := range actions {
		cursor := "  "
		if i == m.envListActionCursor {
			cursor = "> "
			s.WriteString(selectedStyle.Render(cursor + action) + "\n")
		} else {
			s.WriteString(cursor + action + "\n")
		}
	}

	s.WriteString("\n")
	s.WriteString(dimStyle.Render("↑/↓: navigate environments | ←/→: navigate actions | enter: select action | esc: back"))
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
	var envName string
	var variables map[string]string

	if m.viewingCollectionEnv {
		if m.currentCollectionEnv == nil {
			return "No collection environment selected"
		}
		envName = m.currentCollectionEnv.Name
		variables = m.currentCollectionEnv.Variables
	} else {
		if m.currentEnv == nil {
			return "No environment selected"
		}
		envName = m.currentEnv.Name
		variables = m.currentEnv.Variables
	}

	var s strings.Builder

	if m.viewingCollectionEnv {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Collection Environment: %s", envName)))
	} else {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Global Environment: %s", envName)))
	}
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Variables: %d\n\n", len(variables)))

	if len(variables) > 0 {
		for k, v := range variables {
			s.WriteString(fmt.Sprintf("  %s = %s\n", k, v))
		}
		s.WriteString("\n")
	}

	// Action menu as a selectable list
	s.WriteString("Actions:\n")
	var actions []string
	if m.viewingCollectionEnv {
		actions = []string{
			"Activate Environment",
			"Manage Variables",
			"Edit Name",
			"Delete Environment",
		}
	} else {
		actions = []string{
			"Activate Environment",
			"Manage Variables",
			"Edit Name",
			"Save Environment",
			"Delete Environment",
		}
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
	} else if m.message != "" {
		s.WriteString("\n" + successStyle.Render(m.message))
	}

	return s.String()
}

func (m Model) viewEnvironmentVariables() string {
	var envName string
	var variables map[string]string

	if m.viewingCollectionEnv {
		if m.currentCollectionEnv == nil {
			return "No collection environment selected"
		}
		envName = m.currentCollectionEnv.Name
		variables = m.currentCollectionEnv.Variables
	} else {
		if m.currentEnv == nil {
			return "No environment selected"
		}
		envName = m.currentEnv.Name
		variables = m.currentEnv.Variables
	}

	var s strings.Builder

	if m.viewingCollectionEnv {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Collection Environment Variables: %s", envName)))
	} else {
		s.WriteString(titleStyle.Render(fmt.Sprintf("Environment Variables: %s", envName)))
	}
	s.WriteString("\n\n")

	if len(variables) == 0 {
		s.WriteString(dimStyle.Render("No variables set. Press 'enter' to add one."))
	} else {
		for k, v := range variables {
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

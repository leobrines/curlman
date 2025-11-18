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
		s.WriteString(dimStyle.Render("No environments yet. Press 'enter' to create one."))
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

	// Option to create new environment
	cursor := " "
	if m.cursor == len(m.environments) {
		cursor = ">"
		s.WriteString("\n" + selectedStyle.Render(cursor+" [Create New Environment]"))
	} else {
		s.WriteString("\n" + cursor + " [Create New Environment]")
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: select/create | d: delete | t: toggle global/collection | esc: back"))
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

	if m.viewingCollectionEnv {
		s.WriteString(dimStyle.Render("a: activate | v: manage variables | e: edit name | d: delete | esc: back"))
	} else {
		s.WriteString(dimStyle.Render("a: activate | v: manage variables | e: edit name | s: save | d: delete | esc: back"))
	}
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

package ui

import (
	"curlman/executor"
	"strings"
)

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

package ui

import (
	"github.com/leobrines/curlman/executor"
	"strings"
)

func (m Model) viewResponse() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Response"))
	s.WriteString("\n\n")

	// Print mode selector
	s.WriteString(dimStyle.Render("View Mode: "))
	modes := []string{"All", "Body Only", "Headers Only", "Status Only"}
	for i, mode := range modes {
		if printMode(i) == m.currentPrintMode {
			s.WriteString(selectedStyle.Render("[ " + mode + " ]"))
		} else {
			s.WriteString(dimStyle.Render("  " + mode + "  "))
		}
		if i < len(modes)-1 {
			s.WriteString(" ")
		}
	}
	s.WriteString("\n\n")

	if m.response != nil {
		// Format response based on selected mode
		var formatted string
		switch m.currentPrintMode {
		case printAll:
			formatted = executor.FormatResponse(m.response)
		case printBodyOnly:
			formatted = executor.FormatResponseBodyOnly(m.response)
		case printHeadersOnly:
			formatted = executor.FormatResponseHeadersOnly(m.response)
		case printStatusOnly:
			formatted = executor.FormatResponseStatusOnly(m.response)
		}
		s.WriteString(formatted)
	} else {
		s.WriteString("No response yet")
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("← →: change view mode | s: save body | esc: back"))
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

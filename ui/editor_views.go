package ui

import (
	"fmt"
	"strings"
)

func (m Model) viewHeaders() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render("Headers"))
	s.WriteString("\n\n")

	if len(req.Headers) == 0 {
		s.WriteString(dimStyle.Render("No headers set. Press 'enter' to add one."))
	} else {
		for k, v := range req.Headers {
			s.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: add header | esc: back"))
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

func (m Model) viewQueryParams() string {
	if m.selectedRequest < 0 || m.selectedRequest >= len(m.collection.Requests) {
		return "No request selected"
	}

	req := m.collection.Requests[m.selectedRequest]
	var s strings.Builder

	s.WriteString(titleStyle.Render("Query Parameters"))
	s.WriteString("\n\n")

	if len(req.QueryParams) == 0 {
		s.WriteString(dimStyle.Render("No query parameters set. Press 'enter' to add one."))
	} else {
		for k, v := range req.QueryParams {
			s.WriteString(fmt.Sprintf("%s = %s\n", k, v))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimStyle.Render("enter: add query param | esc: back"))
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

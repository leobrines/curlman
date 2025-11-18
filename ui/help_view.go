package ui

import (
	"github.com/leobrines/curlman/storage"
	"fmt"
	"strings"
)

func (m Model) viewHelp() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("CurlMan - Help"))
	s.WriteString("\n\n")

	// Display storage directory info
	storageDir, err := storage.GetStorageDir()
	if err == nil {
		s.WriteString(dimStyle.Render(fmt.Sprintf("Storage Directory: %s\n", storageDir)))
		s.WriteString(dimStyle.Render("All collections are saved/loaded from this directory by default.\n\n"))
	}

	s.WriteString("Main View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate menu\n")
	s.WriteString("  enter - Select menu item\n")
	s.WriteString("  q - Quit application\n\n")

	s.WriteString("Request List View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate requests\n")
	s.WriteString("  enter - View request details or create new\n")
	s.WriteString("  d - Delete selected request\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Request Detail View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate actions\n")
	s.WriteString("  enter - Execute selected action\n")
	s.WriteString("  Actions: Execute, Edit, Headers, Query Params, Clone, Export\n")
	s.WriteString("  esc - Back to request list\n\n")

	s.WriteString("Request Edit View:\n")
	s.WriteString("  ↑/↓ - Navigate fields\n")
	s.WriteString("  enter - Edit selected field\n")
	s.WriteString("  esc - Back to request detail\n\n")

	s.WriteString("Response View:\n")
	s.WriteString("  ← → - Change view mode (All, Body, Headers, Status)\n")
	s.WriteString("  s - Save response body to file\n")
	s.WriteString("  esc - Back to request detail\n\n")

	s.WriteString("Variables View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate variables\n")
	s.WriteString("  enter - Edit selected variable\n")
	s.WriteString("  n - Create new variable\n")
	s.WriteString("  d - Delete selected variable\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Environment Management:\n")
	s.WriteString("  Environments List:\n")
	s.WriteString("    ↑/↓ - Navigate environments\n")
	s.WriteString("    enter - Select/create environment\n")
	s.WriteString("    d - Delete environment\n")
	s.WriteString("    t - Toggle global/collection environments\n")
	s.WriteString("  Environment Detail:\n")
	s.WriteString("    ↑/↓ - Navigate actions\n")
	s.WriteString("    enter - Execute selected action\n")
	s.WriteString("    Actions: Activate, Variables, Edit Name, Save, Delete\n\n")

	s.WriteString("Variables Usage:\n")
	s.WriteString("  Use {{variable_name}} in requests\n")
	s.WriteString("  Environment variables override collection variables\n")
	s.WriteString("  Variables are injected before execution\n\n")

	s.WriteString(dimStyle.Render("Press 'esc' or 'q' to go back"))

	return s.String()
}

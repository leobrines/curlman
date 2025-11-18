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
	s.WriteString("  i - Import OpenAPI YAML file\n")
	s.WriteString("  r - View and manage requests\n")
	s.WriteString("  v - Manage collection variables\n")
	s.WriteString("  e - Manage environments\n")
	s.WriteString("  s - Save collection to JSON (in ~/.curlman/)\n")
	s.WriteString("  l - Load collection from JSON (from ~/.curlman/)\n")
	s.WriteString("  q - Quit application\n\n")

	s.WriteString("Request List View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate requests\n")
	s.WriteString("  enter - View request details\n")
	s.WriteString("  n - Create new request\n")
	s.WriteString("  d - Delete selected request\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Request Detail View:\n")
	s.WriteString("  enter - Execute request\n")
	s.WriteString("  e - Edit request\n")
	s.WriteString("  c - Clone request\n")
	s.WriteString("  x - Export as curl command\n")
	s.WriteString("  esc - Back to request list\n\n")

	s.WriteString("Request Edit View:\n")
	s.WriteString("  ↑/↓ - Navigate fields\n")
	s.WriteString("  enter - Edit selected field\n")
	s.WriteString("  esc - Back to request detail\n\n")

	s.WriteString("Variables View:\n")
	s.WriteString("  ↑/↓ or j/k - Navigate variables\n")
	s.WriteString("  enter - Edit selected variable\n")
	s.WriteString("  n - Create new variable\n")
	s.WriteString("  d - Delete selected variable\n")
	s.WriteString("  esc - Back to main\n\n")

	s.WriteString("Environment Management:\n")
	s.WriteString("  Environments List:\n")
	s.WriteString("    enter - Select/create environment\n")
	s.WriteString("    d - Delete environment\n")
	s.WriteString("  Environment Detail:\n")
	s.WriteString("    a - Activate environment\n")
	s.WriteString("    v - Manage variables\n")
	s.WriteString("    e - Edit name\n")
	s.WriteString("    s - Save environment\n")
	s.WriteString("    d - Delete environment\n\n")

	s.WriteString("Variables Usage:\n")
	s.WriteString("  Use {{variable_name}} in requests\n")
	s.WriteString("  Environment variables override collection variables\n")
	s.WriteString("  Variables are injected before execution\n\n")

	s.WriteString(dimStyle.Render("Press 'esc' or 'q' to go back"))

	return s.String()
}

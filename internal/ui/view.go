package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/leit0/curlman/internal/models"
)

var (
	// Styles
	activeStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	inactiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// View renders the UI
func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	switch m.viewMode {
	case ViewModeExecuting:
		return m.renderExecutionView()
	case ViewModeInput:
		return m.renderInputView()
	case ViewModeConfirm:
		return m.renderConfirmView()
	case ViewModeRequestDetail:
		return m.renderRequestDetailView()
	case ViewModeCurlActions:
		return m.renderCurlActionsView()
	case ViewModeRequestExpanded:
		return m.renderRequestExpandedView()
	default:
		return m.renderMainView()
	}
}

func (m *Model) renderMainView() string {
	// Calculate panel widths
	panelWidth := (m.width - 6) / 3
	if panelWidth < 20 {
		panelWidth = 20
	}

	// Render panels
	envPanel := m.renderEnvironmentsPanel(panelWidth)
	collPanel := m.renderCollectionsPanel(panelWidth)
	reqPanel := m.renderRequestsPanel(panelWidth)

	// Apply active/inactive styles
	if m.currentPanel == PanelEnvironments {
		envPanel = activeStyle.Width(panelWidth).Render(envPanel)
	} else {
		envPanel = inactiveStyle.Width(panelWidth).Render(envPanel)
	}

	if m.currentPanel == PanelCollections {
		collPanel = activeStyle.Width(panelWidth).Render(collPanel)
	} else {
		collPanel = inactiveStyle.Width(panelWidth).Render(collPanel)
	}

	if m.currentPanel == PanelRequests {
		reqPanel = activeStyle.Width(panelWidth).Render(reqPanel)
	} else {
		reqPanel = inactiveStyle.Width(panelWidth).Render(reqPanel)
	}

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, envPanel, collPanel, reqPanel)

	// Add help text
	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, panels, help)
}

func (m *Model) renderEnvironmentsPanel(width int) string {
	if len(m.environments) == 0 {
		var b strings.Builder
		b.WriteString(titleStyle.Render("Environments"))
		b.WriteString("\n\n")
		b.WriteString("No environments\n")
		b.WriteString(helpStyle.Render("Press 'a' to add"))
		return b.String()
	}

	// Use Bubble Tea list's built-in View() method
	return m.envList.View()
}

func (m *Model) renderCollectionsPanel(width int) string {
	if len(m.collections) == 0 {
		var b strings.Builder
		b.WriteString(titleStyle.Render("Collections"))
		b.WriteString("\n\n")
		b.WriteString("No collections\n")
		b.WriteString(helpStyle.Render("Press 'a' to add"))
		return b.String()
	}

	// Use Bubble Tea list's built-in View() method
	return m.collList.View()
}

func (m *Model) renderRequestsPanel(width int) string {
	if m.selectedColl == nil {
		var b strings.Builder
		b.WriteString(titleStyle.Render("Requests"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Select a collection"))
		return b.String()
	}

	if len(m.getAllRequests()) == 0 {
		var b strings.Builder
		b.WriteString(titleStyle.Render("Requests"))
		b.WriteString("\n\n")
		b.WriteString("No requests\n")
		b.WriteString(helpStyle.Render("Press 'a' to add"))
		return b.String()
	}

	// Use Bubble Tea list's built-in View() method
	return m.reqList.View()
}

func (m *Model) renderExecutionView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Executing Request"))
	b.WriteString("\n\n")

	if m.selectedReq != nil {
		b.WriteString(fmt.Sprintf("Request: %s\n", m.selectedReq.Name))
		b.WriteString(fmt.Sprintf("Method: %s\n", m.selectedReq.Method))
		b.WriteString(fmt.Sprintf("URL: %s\n\n", m.selectedReq.URL))
	}

	b.WriteString("Output:\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n")

	if m.executing {
		b.WriteString(m.spinner.View() + " " + m.executionOutput)
	} else {
		b.WriteString(m.executionOutput)
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", 80))
	b.WriteString("\n\n")

	if m.executing {
		b.WriteString(helpStyle.Render("Executing... Press ESC to cancel"))
	} else {
		b.WriteString(helpStyle.Render("Press ESC to return"))
	}

	return b.String()
}

func (m *Model) renderInputView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.inputPrompt))
	b.WriteString("\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter: Submit • ESC: Cancel"))

	return b.String()
}

func (m *Model) renderConfirmView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm"))
	b.WriteString("\n\n")
	b.WriteString(m.confirmPrompt)
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y: Yes • n: No • ESC: Cancel"))

	return b.String()
}

func (m *Model) renderHelp() string {
	var helps []string

	switch m.viewMode {
	case ViewModeList:
		helps = []string{
			"Tab: Switch panels",
			"Enter: View details",
			"e: Edit",
			"a: Add",
			"d: Delete",
			"Ctrl+C twice: Exit",
		}
	case ViewModeExecuting:
		helps = []string{
			"ESC: Cancel/Return",
		}
	case ViewModeInput:
		helps = []string{
			"Enter: Submit",
			"ESC: Cancel",
		}
	case ViewModeConfirm:
		helps = []string{
			"y: Yes",
			"n: No",
			"ESC: Cancel",
		}
	case ViewModeRequestDetail:
		helps = []string{
			"Enter: Generate/Open",
			"Tab/←/→: Switch panels",
			"↑/↓: Navigate managed",
			"v: Expand content",
			"ESC: Back",
		}
	case ViewModeCurlActions:
		if m.curlActionSource != nil && !m.curlActionSource.IsManaged {
			helps = []string{
				"Enter: Execute action",
				"←/→: Navigate actions",
				"Space/Tab: Toggle sections",
				"1-4: Toggle specific section",
				"ESC: Back",
			}
		} else {
			helps = []string{
				"Enter: Execute action",
				"←/→: Navigate actions",
				"ESC: Back",
			}
		}
	case ViewModeRequestExpanded:
		helps = []string{
			"ESC/q: Back",
		}
	}

	return "\n" + helpStyle.Render(strings.Join(helps, " • "))
}

func (m *Model) renderRequestDetailView() string {
	if m.selectedReq == nil {
		return "No request selected"
	}

	req := m.selectedReq

	// Calculate panel widths (60/40 split)
	leftWidth := int(float64(m.width) * 0.6)
	rightWidth := m.width - leftWidth - 2 // Account for borders

	// Render left panel (Spec Request Details)
	leftPanel := m.renderSpecDetailPanel(req, leftWidth)

	// Render right panel (Managed Requests List)
	rightPanel := m.renderManagedListPanel(rightWidth)

	// Apply focus styles
	leftStyle := inactiveStyle.Width(leftWidth).Height(m.height - 4)
	rightStyle := inactiveStyle.Width(rightWidth).Height(m.height - 4)

	if m.detailFocusPanel == DetailPanelSpec {
		leftStyle = activeStyle.Width(leftWidth).Height(m.height - 4)
	} else {
		rightStyle = activeStyle.Width(rightWidth).Height(m.height - 4)
	}

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftPanel),
		rightStyle.Render(rightPanel),
	)

	// Add help text at bottom
	help := m.renderHelp()

	return panels + "\n" + help
}

func (m *Model) renderSpecDetailPanel(req *models.Request, width int) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Spec Request Details"))
	b.WriteString("\n\n")

	// Request metadata
	b.WriteString(fmt.Sprintf("Name: %s\n", req.Name))
	b.WriteString(fmt.Sprintf("Method: %s\n", req.Method))
	b.WriteString(fmt.Sprintf("URL: %s\n", req.URL))

	if req.OpenAPIOperation != "" {
		b.WriteString(fmt.Sprintf("Operation: %s\n", req.OpenAPIOperation))
	}

	b.WriteString("\n")

	// Headers (truncated)
	if len(req.Headers) > 0 {
		b.WriteString(selectedStyle.Render("Headers:"))
		b.WriteString("\n")
		count := 0
		maxHeaders := 3
		for key, value := range req.Headers {
			if count >= maxHeaders {
				b.WriteString(fmt.Sprintf("  ... and %d more (press 'v' to expand)\n", len(req.Headers)-maxHeaders))
				break
			}
			// Truncate long values
			if len(value) > 40 {
				value = value[:37] + "..."
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			count++
		}
		b.WriteString("\n")
	}

	// Body (truncated)
	if req.Body != "" {
		b.WriteString(selectedStyle.Render("Body:"))
		b.WriteString("\n")
		bodyLines := strings.Split(req.Body, "\n")
		maxBodyLines := 3
		if len(bodyLines) > maxBodyLines {
			b.WriteString(strings.Join(bodyLines[:maxBodyLines], "\n"))
			b.WriteString("\n... (press 'v' to expand)\n")
		} else {
			b.WriteString(req.Body)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Generate button (only action)
	b.WriteString(selectedStyle.Render("Actions:"))
	b.WriteString("\n")
	if m.detailFocusPanel == DetailPanelSpec {
		b.WriteString(selectedStyle.Render("▶ Generate Curl"))
	} else {
		b.WriteString("  Generate Curl")
	}
	b.WriteString("\n")

	return b.String()
}

func (m *Model) renderManagedListPanel(width int) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Managed Requests"))
	b.WriteString("\n\n")

	if len(m.relatedRequests) == 0 {
		b.WriteString(helpStyle.Render("No managed requests yet.\n"))
		b.WriteString(helpStyle.Render("Press 'Generate' to create one."))
	} else {
		b.WriteString(fmt.Sprintf("Count: %d\n\n", len(m.relatedRequests)))

		// List managed requests
		for i, relReq := range m.relatedRequests {
			prefix := "  "
			if m.detailFocusPanel == DetailPanelManaged && i == m.detailRelatedIndex {
				prefix = selectedStyle.Render("▶ ")
			} else if m.detailFocusPanel == DetailPanelManaged && i == m.detailRelatedIndex {
				prefix = "▶ "
			}

			// Truncate name if too long
			name := relReq.Name
			if len(name) > width-10 {
				name = name[:width-13] + "..."
			}

			timeSince := ""
			if !relReq.UpdatedAt.IsZero() {
				timeSince = fmt.Sprintf("\n     %s", helpStyle.Render(formatTimeSince(relReq.UpdatedAt)))
			}

			b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, name, timeSince))
		}
	}

	return b.String()
}

// renderSpecDetailsSection renders collapsible spec details sections
func (m *Model) renderSpecDetailsSection() string {
	if m.curlActionSource == nil {
		return ""
	}

	req := m.curlActionSource
	var b strings.Builder

	b.WriteString(selectedStyle.Render("API Specification Details"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Space/Tab to cycle sections, or 1-4 to toggle specific sections"))
	b.WriteString("\n\n")

	// Section 1: Description & Tags
	hasDescription := req.Description != "" || len(req.Tags) > 0 || req.Deprecated
	if hasDescription {
		expanded := m.specDetailExpanded["description"]
		chevron := "▶"
		if expanded {
			chevron = "▼"
		}
		b.WriteString(fmt.Sprintf("%s [1] Description & Tags\n", chevron))

		if expanded {
			if req.Deprecated {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Render("  ⚠ DEPRECATED"))
				b.WriteString("\n")
			}
			if req.Description != "" {
				b.WriteString("  ")
				b.WriteString(req.Description)
				b.WriteString("\n")
			}
			if len(req.Tags) > 0 {
				b.WriteString("  Tags: ")
				b.WriteString(strings.Join(req.Tags, ", "))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	// Section 2: Parameters
	if len(req.Parameters) > 0 {
		expanded := m.specDetailExpanded["parameters"]
		chevron := "▶"
		if expanded {
			chevron = "▼"
		}
		b.WriteString(fmt.Sprintf("%s [2] Parameters (%d)\n", chevron, len(req.Parameters)))

		if expanded {
			// Group by type
			pathParams := []models.ParameterDetail{}
			queryParams := []models.ParameterDetail{}
			headerParams := []models.ParameterDetail{}

			for _, p := range req.Parameters {
				switch p.In {
				case "path":
					pathParams = append(pathParams, p)
				case "query":
					queryParams = append(queryParams, p)
				case "header":
					headerParams = append(headerParams, p)
				}
			}

			if len(pathParams) > 0 {
				b.WriteString("  Path:\n")
				for _, p := range pathParams {
					b.WriteString(fmt.Sprintf("    • %s", p.Name))
					if p.Required {
						b.WriteString(" (required)")
					}
					if p.Type != "" {
						b.WriteString(fmt.Sprintf(" [%s]", p.Type))
					}
					if p.Description != "" {
						b.WriteString(fmt.Sprintf(" - %s", p.Description))
					}
					if p.Example != "" {
						b.WriteString(fmt.Sprintf(" (ex: %s)", p.Example))
					}
					b.WriteString("\n")
				}
			}

			if len(queryParams) > 0 {
				b.WriteString("  Query:\n")
				for _, p := range queryParams {
					b.WriteString(fmt.Sprintf("    • %s", p.Name))
					if p.Required {
						b.WriteString(" (required)")
					}
					if p.Type != "" {
						b.WriteString(fmt.Sprintf(" [%s]", p.Type))
					}
					if p.Description != "" {
						b.WriteString(fmt.Sprintf(" - %s", p.Description))
					}
					if p.Default != "" {
						b.WriteString(fmt.Sprintf(" (default: %s)", p.Default))
					}
					b.WriteString("\n")
				}
			}

			if len(headerParams) > 0 {
				b.WriteString("  Headers:\n")
				for _, p := range headerParams {
					b.WriteString(fmt.Sprintf("    • %s", p.Name))
					if p.Required {
						b.WriteString(" (required)")
					}
					if p.Description != "" {
						b.WriteString(fmt.Sprintf(" - %s", p.Description))
					}
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
		}
	}

	// Section 3: Request & Response Schemas
	hasSchemas := req.RequestBodySchema != nil || len(req.Responses) > 0
	if hasSchemas {
		expanded := m.specDetailExpanded["schemas"]
		chevron := "▶"
		if expanded {
			chevron = "▼"
		}
		b.WriteString(fmt.Sprintf("%s [3] Request & Response Schemas\n", chevron))

		if expanded {
			if req.RequestBodySchema != nil {
				b.WriteString("  Request Body:\n")
				if req.RequestBodySchema.Required {
					b.WriteString("    (required) ")
				}
				if req.RequestBodySchema.ContentType != "" {
					b.WriteString(fmt.Sprintf("[%s] ", req.RequestBodySchema.ContentType))
				}
				if req.RequestBodySchema.Schema != "" {
					b.WriteString(req.RequestBodySchema.Schema)
				}
				if req.RequestBodySchema.Description != "" {
					b.WriteString(fmt.Sprintf(" - %s", req.RequestBodySchema.Description))
				}
				b.WriteString("\n")
			}

			if len(req.Responses) > 0 {
				b.WriteString("  Responses:\n")
				// Common status codes order
				statusOrder := []string{"200", "201", "204", "400", "401", "403", "404", "500"}
				for _, status := range statusOrder {
					if resp, ok := req.Responses[status]; ok {
						b.WriteString(fmt.Sprintf("    • %s: ", status))
						if resp.Description != "" {
							b.WriteString(resp.Description)
						}
						if resp.ContentType != "" && resp.Schema != "" {
							b.WriteString(fmt.Sprintf(" [%s: %s]", resp.ContentType, resp.Schema))
						}
						b.WriteString("\n")
					}
				}
				// Show any remaining status codes not in our order
				for status, resp := range req.Responses {
					found := false
					for _, s := range statusOrder {
						if s == status {
							found = true
							break
						}
					}
					if !found {
						b.WriteString(fmt.Sprintf("    • %s: ", status))
						if resp.Description != "" {
							b.WriteString(resp.Description)
						}
						if resp.ContentType != "" && resp.Schema != "" {
							b.WriteString(fmt.Sprintf(" [%s: %s]", resp.ContentType, resp.Schema))
						}
						b.WriteString("\n")
					}
				}
			}
			b.WriteString("\n")
		}
	}

	// Section 4: Security
	if len(req.Security) > 0 {
		expanded := m.specDetailExpanded["security"]
		chevron := "▶"
		if expanded {
			chevron = "▼"
		}
		b.WriteString(fmt.Sprintf("%s [4] Security Requirements\n", chevron))

		if expanded {
			for _, sec := range req.Security {
				b.WriteString(fmt.Sprintf("  • %s", sec.Name))
				if sec.Type != "" {
					b.WriteString(fmt.Sprintf(" (%s)", sec.Type))
				}
				if sec.Scheme != "" {
					b.WriteString(fmt.Sprintf(" - %s", sec.Scheme))
				}
				if sec.In != "" {
					b.WriteString(fmt.Sprintf(" in %s", sec.In))
				}
				if len(sec.Scopes) > 0 {
					b.WriteString(fmt.Sprintf(" [scopes: %s]", strings.Join(sec.Scopes, ", ")))
				}
				if sec.Description != "" {
					b.WriteString(fmt.Sprintf("\n    %s", sec.Description))
				}
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(strings.Repeat("─", min(m.width-4, 100)))
	b.WriteString("\n")

	return b.String()
}

func (m *Model) renderCurlActionsView() string {
	var b strings.Builder

	// Title with source info
	title := "Curl Command"
	if m.curlActionSource != nil {
		if m.curlActionSource.IsManaged {
			title += " - Managed Request: " + m.curlActionSource.Name
		} else {
			title += " - Generated from Spec"
		}
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Display request info before spec details
	if m.curlActionSource != nil {
		b.WriteString(fmt.Sprintf("Name: %s\n", m.curlActionSource.Name))
		b.WriteString(fmt.Sprintf("Method: %s\n", m.curlActionSource.Method))
		b.WriteString(fmt.Sprintf("URL: %s\n", m.curlActionSource.URL))
		if m.curlActionSource.OpenAPIOperation != "" {
			b.WriteString(fmt.Sprintf("Operation: %s\n", m.curlActionSource.OpenAPIOperation))
		}
		b.WriteString("\n")
	}

	// Render spec details section if this is a spec request
	if m.curlActionSource != nil && !m.curlActionSource.IsManaged {
		specDetails := m.renderSpecDetailsSection()
		if specDetails != "" {
			b.WriteString(specDetails)
			b.WriteString("\n")
		}
	}

	// Display curl command
	b.WriteString(selectedStyle.Render("Curl Command:"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(m.width-4, 100)))
	b.WriteString("\n")

	// Show curl command (scrollable if needed)
	curlLines := strings.Split(m.curlActionCommand, "\n")
	maxLines := m.height - 15 // Leave room for actions and help
	if len(curlLines) > maxLines && maxLines > 0 {
		b.WriteString(strings.Join(curlLines[:maxLines], "\n"))
		b.WriteString(fmt.Sprintf("\n... (%d more lines)", len(curlLines)-maxLines))
	} else {
		b.WriteString(m.curlActionCommand)
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", min(m.width-4, 100)))
	b.WriteString("\n\n")

	// Actions menu: Edit, Execute, Copy, Save
	b.WriteString(selectedStyle.Render("Actions:"))
	b.WriteString("\n")

	actions := []string{"Edit (vim)", "Execute", "Copy", "Save"}
	var actionParts []string
	for i, action := range actions {
		if i == m.curlActionIndex {
			actionParts = append(actionParts, selectedStyle.Render("▶ "+action))
		} else {
			actionParts = append(actionParts, "  "+action)
		}
	}
	b.WriteString(strings.Join(actionParts, "  "))
	b.WriteString("\n\n")

	// Help text
	b.WriteString(m.renderHelp())

	return b.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *Model) renderRequestExpandedView() string {
	if m.selectedReq == nil {
		return "No request selected"
	}

	var b strings.Builder
	req := m.selectedReq

	b.WriteString(titleStyle.Render("Expanded View"))
	b.WriteString("\n\n")

	switch m.expandedSection {
	case "curl":
		b.WriteString(selectedStyle.Render("Full Curl Command:"))
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", 80))
		b.WriteString("\n")
		b.WriteString(req.CurlCommand)
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", 80))
		b.WriteString("\n")

	case "headers":
		b.WriteString(selectedStyle.Render("All Headers:"))
		b.WriteString("\n")
		if len(req.Headers) == 0 {
			b.WriteString("No headers\n")
		} else {
			for key, value := range req.Headers {
				b.WriteString(fmt.Sprintf("  • %s: %s\n", key, value))
			}
		}

	case "body":
		b.WriteString(selectedStyle.Render("Full Body:"))
		b.WriteString("\n")
		if req.Body == "" {
			b.WriteString("No body\n")
		} else {
			b.WriteString(req.Body)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

// Helper function to format time since
func formatTimeSince(t time.Time) string {
	duration := time.Since(t)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes())

	if days > 365 {
		years := days / 365
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
	if days > 30 {
		months := days / 30
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	if days > 0 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if hours > 0 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if minutes > 0 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	return "just now"
}

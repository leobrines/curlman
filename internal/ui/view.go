package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
			"Enter: Select/Execute",
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
	}

	return "\n" + helpStyle.Render(strings.Join(helps, " • "))
}

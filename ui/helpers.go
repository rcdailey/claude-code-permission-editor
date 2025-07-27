package ui

import (
	"strings"

	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleKeyPress handles keyboard input using pure state management
func handleKeyPress(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	key := msg.String()

	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit
	}

	if key == "tab" {
		return handleTabKey(m), nil
	}

	return handleNavigationKeys(m, key), nil
}

// handleTabKey switches between screens
func handleTabKey(m *types.Model) *types.Model {
	if m.CurrentScreen == types.ScreenDuplicates {
		m.CurrentScreen = types.ScreenOrganization
	} else {
		m.CurrentScreen = types.ScreenDuplicates
	}
	return m
}

// handleNavigationKeys handles left/right navigation
func handleNavigationKeys(m *types.Model, key string) *types.Model {
	if m.CurrentScreen != types.ScreenOrganization {
		return m
	}

	switch key {
	case "left", "h":
		if m.FocusedColumn > 0 {
			m.FocusedColumn--
		}
	case "right", "l":
		if m.FocusedColumn < 2 {
			m.FocusedColumn++
		}
	}
	return m
}

// renderModal renders a modal dialog using lipgloss.Place()
func renderModal(m *types.Model) string {
	// Simple modal placeholder - will be implemented properly later
	return "Modal dialog will be implemented here"
}

// renderConfirmation renders the confirmation screen
func renderConfirmation(m *types.Model) string {
	// Create title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).
		Width(m.Width).
		Padding(1)
	title := titleStyle.Render("Confirm Actions")

	// Create action summary
	if len(m.Actions) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(m.Width).
			Height(m.Height-6).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8"))
		content := emptyStyle.Render("No actions queued")

		instructions := "Press ESC to return to main screen"
		instrStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Align(lipgloss.Center).
			Width(m.Width)
		footer := instrStyle.Render(instructions)

		return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
	}

	// Build action list
	actionLines := make([]string, 0, len(m.Actions))
	for i, action := range m.Actions {
		actionText := formatAction(action)
		prefix := "  "
		if i == 0 { // Highlight first action for simplicity
			prefix = "> "
		}
		actionLines = append(actionLines, prefix+actionText)
	}

	contentStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1)
	content := contentStyle.Render(strings.Join(actionLines, "\n"))

	// Instructions
	instructions := "ENTER: execute all actions | ESC: cancel and return"
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Align(lipgloss.Center).
		Width(m.Width)
	footer := instrStyle.Render(instructions)

	return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
}

// formatAction formats an action for display
func formatAction(action types.Action) string {
	switch action.Type {
	case "move":
		return action.Permission + ": " + action.FromLevel + " → " + action.ToLevel
	case "edit":
		return action.Permission + " → " + action.NewName
	case "duplicate":
		return "Remove " + action.Permission + " from " + action.FromLevel
	default:
		return action.Permission + " (" + action.Type + ")"
	}
}

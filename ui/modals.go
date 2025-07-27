package ui

import (
	"strings"

	"claude-permissions/types"

	"github.com/charmbracelet/lipgloss/v2"
)

// SmallModal implements types.Modal for small centered dialog boxes
type SmallModal struct {
	Title  string
	Body   string
	Action string // "continue", "exit", etc.
}

// NewSmallModal creates a new small modal dialog
func NewSmallModal(title, body, action string) *SmallModal {
	return &SmallModal{
		Title:  title,
		Body:   body,
		Action: action,
	}
}

// RenderModal renders the small modal content (extracted from renderModal function)
func (sm *SmallModal) RenderModal(width, height int) string {
	// Calculate modal dimensions
	contentWidth := 60

	// Create modal content with high contrast styling
	modalStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("11")).
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("15")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11")).
		Align(lipgloss.Center).
		Width(contentWidth - 4) // Account for padding

	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Width(contentWidth-4). // Account for padding
		Padding(1, 0)

	// Style instructions consistently with footer hints using AccentStyle
	instructionsStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(contentWidth-4). // Account for padding
		Padding(1, 0, 0, 0)

	title := titleStyle.Render(sm.Title)
	body := bodyStyle.Render(sm.Body)

	// Use AccentStyle from this package
	instructions := instructionsStyle.Render(
		AccentStyle.Render("Y/Enter") + " · Yes  |  " + AccentStyle.Render("N/ESC") + " · No",
	)

	modalContent := modalStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, body, instructions),
	)

	return modalContent
}

// HandleInput processes keyboard input for the small modal
func (sm *SmallModal) HandleInput(key string) (handled bool, result interface{}) {
	switch key {
	case "y", "Y", keyEnter:
		return true, "yes"
	case "n", "N", keyEscapeLong, keyEscape:
		return true, "no"
	default:
		return false, nil
	}
}

// ConfirmChangesModal implements types.Modal for full-screen confirm changes dialog
type ConfirmChangesModal struct {
	model *types.Model
}

// NewConfirmChangesModal creates a new confirm changes modal
func NewConfirmChangesModal(model *types.Model) *ConfirmChangesModal {
	return &ConfirmChangesModal{
		model: model,
	}
}

// RenderModal renders the confirm changes content (extracted from renderConfirmation function)
func (ccm *ConfirmChangesModal) RenderModal(width, height int) string {
	// Create title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Align(lipgloss.Center).
		Width(width).
		Padding(1)
	title := titleStyle.Render("Confirm Changes")

	// Build list of pending changes
	changeLines := buildPendingChangesList(ccm.model)

	if len(changeLines) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(width).
			Height(height-6).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8"))
		content := emptyStyle.Render("No pending changes")

		instructions := AccentStyle.Render("ESC") + " · Return to main screen"
		instrStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width)
		footer := instrStyle.Render(instructions)

		return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
	}

	contentStyle := lipgloss.NewStyle().
		Width(width).
		Height(height - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1)
	content := contentStyle.Render(strings.Join(changeLines, "\n"))

	// Instructions using consistent footer styling
	row1Keys := []string{
		AccentStyle.Render("ENTER") + " · Execute all actions",
		AccentStyle.Render("ESC") + " · Cancel and return",
	}
	row2Keys := []string{
		AccentStyle.Render("Q") + " · Quit without saving",
	}
	instructions := strings.Join(row1Keys, "  |  ") + "\n" + strings.Join(row2Keys, "  |  ")
	instrStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width)
	footer := instrStyle.Render(instructions)

	return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
}

// HandleInput processes keyboard input for the confirm changes modal
func (ccm *ConfirmChangesModal) HandleInput(key string) (handled bool, result interface{}) {
	switch key {
	case keyEnter:
		return true, "execute"
	case keyEscapeLong, keyEscape:
		return true, "cancel"
	case "q", "Q":
		return true, "quit"
	default:
		return false, nil
	}
}

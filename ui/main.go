package ui

import (
	"fmt"
	"os"
	"strings"

	"claude-permissions/debug"
	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// Init initializes the model
func Init(_ *types.Model) tea.Cmd {
	// WindowSizeMsg will be sent automatically in v2
	return nil
}

// Update handles all Bubble Tea messages using pure state management
func Update(m *types.Model, msg tea.Msg) (*types.Model, tea.Cmd) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update terminal dimensions - no layout engine needed
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return handleKeyPress(m, msg)

	case debug.LaunchConfirmChangesMsg:
		return handleLaunchConfirmChanges(m, msg), nil

	default:
		return m, nil
	}
}

// View renders the entire UI using pure lipgloss composition
func View(m *types.Model) string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	// Handle case when terminal dimensions haven't been set yet
	if m.Width == 0 || m.Height == 0 {
		return "Initializing layout... (waiting for terminal size)"
	}

	// Always render main layout as base content - modals will overlay on top
	baseContent := renderMainLayout(m)

	// Overlay modal if shown
	if m.ActiveModal != nil {
		return renderModal(m, baseContent)
	}

	return baseContent
}

// renderMainLayout renders the main UI using pure lipgloss composition
func renderMainLayout(m *types.Model) string {
	// Create header component
	header := NewHeaderComponent(m.Width)
	header.SetContent(renderHeaderContent(m))

	// Create footer component
	footer := NewFooterComponent(m.Width)
	footer.SetContent(renderFooterContent(m))

	// Use lipgloss dynamic height calculation (following best practices)
	headerContent := header.View()
	footerContent := footer.View()
	headerHeight := lipgloss.Height(headerContent)
	footerHeight := lipgloss.Height(footerContent)

	// Create and render status bar
	statusContent := renderStatusBarContent(m)
	statusHeight := lipgloss.Height(statusContent)

	// Calculate content height: total minus header, footer, and status
	contentHeight := m.Height - headerHeight - footerHeight - statusHeight

	// Create content component
	content := NewContentComponent(m.Width, contentHeight, m)

	// Join all components vertically using pure lipgloss
	return lipgloss.JoinVertical(lipgloss.Top,
		headerContent,
		content.View(),
		statusContent,
		footerContent,
	)
}

// renderHeaderContent generates the header content string with file status and current directory
func renderHeaderContent(m *types.Model) string {
	// File status indicators using centralized theme
	userStatus := "X"
	userStatusStyle := ErrorStyle
	if m.UserLevel.Exists {
		userStatus = "OK"
		userStatusStyle = SuccessStyle
	}

	repoStatus := "X"
	repoStatusStyle := ErrorStyle
	if m.RepoLevel.Exists {
		repoStatus = "OK"
		repoStatusStyle = SuccessStyle
	}

	localStatus := "X"
	localStatusStyle := ErrorStyle
	if m.LocalLevel.Exists {
		localStatus = "OK"
		localStatusStyle = SuccessStyle
	}

	// Build file info with themed styling
	fileInfo := fmt.Sprintf(
		"Files: Local:%s%s Repo:%s%s User:%s%s",
		localStatusStyle.Render(localStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.LocalLevel.Permissions))),
		repoStatusStyle.Render(repoStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.RepoLevel.Permissions))),
		userStatusStyle.Render(userStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.UserLevel.Permissions))),
	)

	// Current working directory with accent color
	cwd, _ := os.Getwd()
	currentDir := fmt.Sprintf("%s %s", AccentStyle.Render("Current:"), cwd)

	// Build header text with themed styling
	title := TitleStyle.Render("Claude Code Permission Editor")

	return fmt.Sprintf("%s\n%s | %s", title, fileInfo, currentDir)
}

// renderFooterContent generates the footer content string with context-sensitive hotkeys
func renderFooterContent(m *types.Model) string {
	var row1Actions, row2Actions []string

	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		row1Actions = []string{
			formatFooterAction("TAB", "Switch panel"),
			formatFooterAction("↑↓", "Navigate"),
		}
		row2Actions = []string{
			formatFooterAction("ENTER", "Save"),
			formatFooterAction("ESC", "Reset changes"),
			formatFooterAction("1/2/3", "Keep in LOCAL/REPO/USER"),
		}
	case types.ScreenOrganization:
		row1Actions = []string{
			formatFooterAction("TAB", "Switch panel"),
			formatFooterAction("↑↓", "Navigate within column"),
			formatFooterAction("←→", "Switch between columns"),
		}
		row2Actions = []string{
			formatFooterAction("ENTER", "Save"),
			formatFooterAction("ESC", "Reset changes"),
			formatFooterAction("1/2/3", "Move to LOCAL/REPO/USER"),
		}
	default:
		// Generic footer
		row1Actions = []string{
			formatFooterAction("TAB", "Switch panel"),
			formatFooterAction("↑↓", "Navigate"),
		}
		row2Actions = []string{
			formatFooterAction("SPACE", "Select"),
			formatFooterAction("ENTER", "Confirm"),
			formatFooterAction("Q", "Quit"),
		}
	}

	return buildTwoRowFooter(row1Actions, row2Actions)
}

// renderStatusBarContent generates the status bar with contextual information
func renderStatusBarContent(m *types.Model) string {
	var statusText string

	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		statusText = renderDuplicatesStatusText(m)
	case types.ScreenOrganization:
		statusText = renderOrganizationStatusText(m)
	default:
		statusText = "Claude Code Permission Editor"
	}

	// Style the status bar using centralized theme
	statusBarStyle := StatusBarStyle.Width(m.Width)
	return statusBarStyle.Render(statusText)
}

// renderDuplicatesStatusText generates status text for duplicates screen
func renderDuplicatesStatusText(m *types.Model) string {
	if len(m.Duplicates) > 0 {
		cursor := m.DuplicatesTable.Cursor()
		if cursor < len(m.Duplicates) {
			dup := m.Duplicates[cursor]
			levelsStr := strings.Join(dup.Levels, " vs ")
			return fmt.Sprintf(
				"%s conflict: %s (choose 1/2/3)     [%d conflicts remaining]",
				dup.Name,
				levelsStr,
				len(m.Duplicates),
			)
		}
	}
	return "Resolve duplicate permissions"
}

// renderOrganizationStatusText generates status text for organization screen
func renderOrganizationStatusText(m *types.Model) string {
	// Check if duplicates are blocking permissions organization
	if hasUnresolvedDuplicates(m) {
		return "Duplicates must be resolved before organizing permissions"
	}

	columnPerms := getColumnPermissions(m)
	if len(columnPerms) > 0 && m.ColumnSelections[m.FocusedColumn] < len(columnPerms) {
		selectedPerm := columnPerms[m.ColumnSelections[m.FocusedColumn]]
		return fmt.Sprintf(
			"%s (originally %s → in %s)",
			selectedPerm.Name,
			selectedPerm.OriginalLevel,
			selectedPerm.CurrentLevel,
		)
	}
	return "Ready to organize permissions"
}

// getColumnPermissions returns permissions for the currently focused column
func getColumnPermissions(m *types.Model) []types.Permission {
	var columnPerms []types.Permission
	var targetLevel string

	switch m.FocusedColumn {
	case 0:
		targetLevel = types.LevelLocal
	case 1:
		targetLevel = types.LevelRepo
	case 2:
		targetLevel = types.LevelUser
	}

	for _, perm := range m.Permissions {
		if perm.CurrentLevel == targetLevel {
			columnPerms = append(columnPerms, perm)
		}
	}
	return columnPerms
}

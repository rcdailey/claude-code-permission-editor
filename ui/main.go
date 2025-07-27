package ui

import (
	"fmt"
	"os"
	"strings"

	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the model
func Init(_ *types.Model) tea.Cmd {
	// Request initial terminal size from Bubble Tea
	return tea.WindowSize()
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

	default:
		return m, nil
	}
}

// View renders the entire UI using pure lipgloss composition
func View(m *types.Model) string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	if m.ShowModal {
		return renderModal(m)
	}

	if m.ConfirmMode {
		return renderConfirmation(m)
	}

	// Handle case when terminal dimensions haven't been set yet
	if m.Width == 0 || m.Height == 0 {
		return "Initializing layout... (waiting for terminal size)"
	}

	// Pure lipgloss composition - no layout engine needed
	return renderMainLayout(m)
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

	// Account for status bar, blank line, and column borders in content height calculation
	// -3 for column border overhead + blank line
	contentHeight := m.Height - headerHeight - footerHeight - statusHeight - 3

	// Create content component
	content := NewContentComponent(m.Width, contentHeight, m)

	// Join all components vertically using pure lipgloss
	return lipgloss.JoinVertical(lipgloss.Top,
		headerContent,
		content.View(),
		statusContent,
		"", // Blank line for visual separation between status bar and footer
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
		"Files: User:%s%s Repo:%s%s Local:%s%s",
		userStatusStyle.Render(userStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.UserLevel.Permissions))),
		repoStatusStyle.Render(repoStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.RepoLevel.Permissions))),
		localStatusStyle.Render(localStatus),
		CountStyle.Render(fmt.Sprintf("(%d)", len(m.LocalLevel.Permissions))),
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
	// Use centralized theme for all keys

	var row1Keys, row2Keys []string

	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		row1Keys = []string{
			AccentStyle.Render("TAB") + " · Switch panel",
			AccentStyle.Render("↑↓") + " · Navigate",
			AccentStyle.Render("1/2/3") + " · Keep in LOCAL/REPO/USER",
		}
		row2Keys = []string{
			AccentStyle.Render("ENTER") + " · Save & continue",
			AccentStyle.Render("ESC") + " · Cancel/exit",
		}
	case types.ScreenOrganization:
		row1Keys = []string{
			AccentStyle.Render("TAB") + " · Switch panel",
			AccentStyle.Render("↑↓") + " · Navigate within column",
			AccentStyle.Render("←→") + " · Switch between columns",
		}
		row2Keys = []string{
			AccentStyle.Render("1/2/3") + " · Move to LOCAL/REPO/USER",
			AccentStyle.Render("ENTER") + " · Save & exit",
			AccentStyle.Render("ESC") + " · Cancel/exit",
		}
	default:
		// Generic footer
		row1Keys = []string{
			AccentStyle.Render("TAB") + " · Switch panel",
			AccentStyle.Render("↑↓") + " · Navigate",
		}
		row2Keys = []string{
			AccentStyle.Render("SPACE") + " · Select",
			AccentStyle.Render("ENTER") + " · Confirm",
			AccentStyle.Render("Q") + " · Quit",
		}
	}

	return strings.Join(row1Keys, "  |  ") + "\n" + strings.Join(row2Keys, "  |  ")
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
	return "Step 1: Resolve Duplicates"
}

// renderOrganizationStatusText generates status text for organization screen
func renderOrganizationStatusText(m *types.Model) string {
	columnPerms := getColumnPermissions(m)
	if len(columnPerms) > 0 && m.ColumnSelections[m.FocusedColumn] < len(columnPerms) {
		selectedPerm := columnPerms[m.ColumnSelections[m.FocusedColumn]]
		originalLevel := selectedPerm.CurrentLevel
		currentLocation := originalLevel
		if selectedPerm.PendingMove != "" {
			currentLocation = selectedPerm.PendingMove
		}
		return fmt.Sprintf(
			"%s (originally %s → in %s)",
			selectedPerm.Name,
			originalLevel,
			currentLocation,
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

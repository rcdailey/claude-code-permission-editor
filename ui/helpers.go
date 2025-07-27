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

	// Handle number keys for moving permissions
	if key == "1" || key == "2" || key == "3" {
		return handleNumberKeys(m, key), nil
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

// handleNavigationKeys handles navigation and passes through to components
func handleNavigationKeys(m *types.Model, key string) *types.Model {
	switch key {
	case keyUp, "k":
		return handleUpDownNavigation(m, key)
	case keyDown, "j":
		return handleUpDownNavigation(m, key)
	case "left", "h":
		if m.CurrentScreen == types.ScreenOrganization && m.FocusedColumn > 0 {
			m.FocusedColumn--
		}
	case "right", "l":
		if m.CurrentScreen == types.ScreenOrganization && m.FocusedColumn < 2 {
			m.FocusedColumn++
		}
	}
	return m
}

// handleNumberKeys handles 1/2/3 keys for moving permissions or resolving duplicates
func handleNumberKeys(m *types.Model, key string) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		return handleDuplicateResolution(m, key)
	case types.ScreenOrganization:
		return handlePermissionMove(m, key)
	}
	return m
}

// handleDuplicateResolution handles number keys on duplicates screen
func handleDuplicateResolution(m *types.Model, key string) *types.Model {
	if len(m.Duplicates) == 0 {
		return m
	}

	cursor := m.DuplicatesTable.Cursor()
	if cursor >= len(m.Duplicates) {
		return m
	}

	duplicate := m.Duplicates[cursor]
	var keepLevel string

	switch key {
	case "1":
		keepLevel = types.LevelLocal
	case "2":
		keepLevel = types.LevelRepo
	case "3":
		keepLevel = types.LevelUser
	}

	// Update the duplicate's keep level
	m.Duplicates[cursor].KeepLevel = keepLevel

	// Create actions to remove duplicates from other levels
	for _, level := range duplicate.Levels {
		if level != keepLevel {
			action := types.Action{
				Type:       types.ActionDuplicate,
				Permission: duplicate.Name,
				FromLevel:  level,
				ToLevel:    keepLevel,
			}
			m.Actions = append(m.Actions, action)
		}
	}

	return m
}

// handlePermissionMove handles number keys on organization screen
func handlePermissionMove(m *types.Model, key string) *types.Model {
	currentLevelPerms, fromLevel := getCurrentColumnInfo(m)
	if len(currentLevelPerms) == 0 {
		return m
	}

	currentSelection := m.ColumnSelections[m.FocusedColumn]
	if currentSelection >= len(currentLevelPerms) {
		return m
	}

	permissionToMove := currentLevelPerms[currentSelection]
	toLevel := getTargetLevel(key)

	// Don't move if already in target level
	if fromLevel == toLevel {
		return m
	}

	// Perform the immediate move
	movePermissionBetweenLevels(m, permissionToMove, fromLevel, toLevel)
	updateSelectionAfterMove(m, currentSelection)

	return m
}

// getCurrentColumnInfo returns the permissions and level for the focused column
func getCurrentColumnInfo(m *types.Model) ([]string, string) {
	switch m.FocusedColumn {
	case 0:
		return m.LocalLevel.Permissions, types.LevelLocal
	case 1:
		return m.RepoLevel.Permissions, types.LevelRepo
	case 2:
		return m.UserLevel.Permissions, types.LevelUser
	}
	return []string{}, ""
}

// getTargetLevel converts number key to level constant
func getTargetLevel(key string) string {
	switch key {
	case "1":
		return types.LevelLocal
	case "2":
		return types.LevelRepo
	case "3":
		return types.LevelUser
	}
	return ""
}

// updateSelectionAfterMove updates selection after moving an item
func updateSelectionAfterMove(m *types.Model, oldSelection int) {
	newSourceLength := getSourceColumnLength(m, m.FocusedColumn)
	if oldSelection >= newSourceLength && newSourceLength > 0 {
		m.ColumnSelections[m.FocusedColumn] = newSourceLength - 1
	}
}

// movePermissionBetweenLevels immediately moves a permission between levels
func movePermissionBetweenLevels(m *types.Model, permission, fromLevel, toLevel string) {
	// Remove from source level
	switch fromLevel {
	case types.LevelLocal:
		m.LocalLevel.Permissions = removePermission(m.LocalLevel.Permissions, permission)
	case types.LevelRepo:
		m.RepoLevel.Permissions = removePermission(m.RepoLevel.Permissions, permission)
	case types.LevelUser:
		m.UserLevel.Permissions = removePermission(m.UserLevel.Permissions, permission)
	}

	// Add to target level (alphabetically sorted)
	switch toLevel {
	case types.LevelLocal:
		m.LocalLevel.Permissions = addPermissionSorted(m.LocalLevel.Permissions, permission)
	case types.LevelRepo:
		m.RepoLevel.Permissions = addPermissionSorted(m.RepoLevel.Permissions, permission)
	case types.LevelUser:
		m.UserLevel.Permissions = addPermissionSorted(m.UserLevel.Permissions, permission)
	}

	// Update the Permission struct in the model's consolidated view
	for i := range m.Permissions {
		if m.Permissions[i].Name == permission && m.Permissions[i].CurrentLevel == fromLevel {
			m.Permissions[i].CurrentLevel = toLevel
			break
		}
	}
}

// removePermission removes a permission from a slice
func removePermission(perms []string, permission string) []string {
	for i, perm := range perms {
		if perm == permission {
			return append(perms[:i], perms[i+1:]...)
		}
	}
	return perms
}

// addPermissionSorted adds a permission to a slice in alphabetical order (case-insensitive)
func addPermissionSorted(perms []string, permission string) []string {
	// Find insertion point
	insertIndex := 0
	for i, perm := range perms {
		if strings.ToLower(permission) < strings.ToLower(perm) {
			insertIndex = i
			break
		}
		insertIndex = i + 1
	}

	// Insert at the correct position
	perms = append(perms, "")
	copy(perms[insertIndex+1:], perms[insertIndex:])
	perms[insertIndex] = permission

	return perms
}

// getSourceColumnLength returns the length of permissions in the specified column
func getSourceColumnLength(m *types.Model, columnIndex int) int {
	switch columnIndex {
	case 0:
		return len(m.LocalLevel.Permissions)
	case 1:
		return len(m.RepoLevel.Permissions)
	case 2:
		return len(m.UserLevel.Permissions)
	}
	return 0
}

const (
	keyUp   = "up"
	keyDown = "down"
)

// handleUpDownNavigation passes up/down keys to the appropriate component
func handleUpDownNavigation(m *types.Model, key string) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		return handleDuplicatesNavigation(m, key)
	case types.ScreenOrganization:
		return handleOrganizationNavigation(m, key)
	}
	return m
}

// handleDuplicatesNavigation handles up/down navigation for duplicates screen
func handleDuplicatesNavigation(m *types.Model, key string) *types.Model {
	var keyMsg tea.KeyMsg
	switch key {
	case keyUp, "k":
		keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	case keyDown, "j":
		keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	}
	m.DuplicatesTable, _ = m.DuplicatesTable.Update(keyMsg)
	return m
}

// handleOrganizationNavigation handles up/down navigation for organization screen
func handleOrganizationNavigation(m *types.Model, key string) *types.Model {
	var levelPerms []string
	switch m.FocusedColumn {
	case 0:
		levelPerms = m.LocalLevel.Permissions
	case 1:
		levelPerms = m.RepoLevel.Permissions
	case 2:
		levelPerms = m.UserLevel.Permissions
	}

	if len(levelPerms) == 0 {
		return m
	}

	currentSelection := m.ColumnSelections[m.FocusedColumn]
	maxIndex := len(levelPerms) - 1

	switch key {
	case keyUp, "k":
		if currentSelection > 0 {
			m.ColumnSelections[m.FocusedColumn] = currentSelection - 1
		}
	case keyDown, "j":
		if currentSelection < maxIndex {
			m.ColumnSelections[m.FocusedColumn] = currentSelection + 1
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

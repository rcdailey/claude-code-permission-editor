package ui

import (
	"fmt"
	"strings"

	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// handleKeyPress handles keyboard input using pure state management
func handleKeyPress(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	key := msg.String()

	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle modal input first if modal is shown
	if m.ShowModal {
		return handleModalInput(m, key), nil
	}

	return handleNonModalKeys(m, msg, key)
}

// handleNonModalKeys handles key input when no modal is shown
func handleNonModalKeys(m *types.Model, msg tea.KeyMsg, key string) (*types.Model, tea.Cmd) {
	if key == "tab" {
		return handleTabKey(m), nil
	}

	// Handle ESC key for reset functionality on permissions screen
	if key == "escape" || key == "esc" || msg.Key().Code == tea.KeyEscape {
		return handleEscapeKey(m), nil
	}

	// Handle ENTER key for confirmation screen transition
	if key == "enter" || msg.Key().Code == tea.KeyEnter {
		return handleEnterKey(m), nil
	}

	// Handle number keys for moving permissions
	if key == "1" || key == "2" || key == "3" {
		return handleNumberKeys(m, key), nil
	}

	return handleNavigationKeys(m, key), nil
}

// handleEnterKey handles ENTER key based on current screen
func handleEnterKey(m *types.Model) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenConfirmation:
		// Return to organization screen
		m.CurrentScreen = types.ScreenOrganization
	case types.ScreenDuplicates, types.ScreenOrganization:
		// Transition to confirmation screen if there are pending changes
		if hasPendingChanges(m) {
			m.CurrentScreen = types.ScreenConfirmation
		}
	}
	return m
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
		keyMsg = tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	case keyDown, "j":
		keyMsg = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
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

// renderModal renders a modal dialog using Lipgloss v2 Canvas and Layer compositing
func renderModal(m *types.Model, baseContent string) string {
	if !m.ShowModal {
		return baseContent
	}

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

	title := titleStyle.Render(m.ModalTitle)
	body := bodyStyle.Render(m.ModalBody)
	// Style like footer: AccentStyle for keys, normal text for descriptions
	instructions := instructionsStyle.Render(
		AccentStyle.Render("Y/Enter") + " · Yes  |  " + AccentStyle.Render("N/ESC") + " · No",
	)

	modalContent := modalStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, body, instructions),
	)

	// Calculate modal dimensions after rendering
	modalHeight := lipgloss.Height(modalContent)
	modalWidth := lipgloss.Width(modalContent)

	// Use Lipgloss v2 Canvas and Layer compositing for proper background visibility
	baseLayer := lipgloss.NewLayer(baseContent)
	modalLayer := lipgloss.NewLayer(modalContent).
		X((m.Width - modalWidth) / 2).   // Center horizontally
		Y((m.Height - modalHeight) / 2). // Center vertically
		Z(1)                             // On top

	canvas := lipgloss.NewCanvas(baseLayer, modalLayer)
	return canvas.Render()
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
	title := titleStyle.Render("Confirm Changes")

	// Build list of pending changes
	changeLines := buildPendingChangesList(m)

	if len(changeLines) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(m.Width).
			Height(m.Height-6).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8"))
		content := emptyStyle.Render("No pending changes")

		instructions := "Press ESC to return to main screen"
		instrStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Align(lipgloss.Center).
			Width(m.Width)
		footer := instrStyle.Render(instructions)

		return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
	}

	contentStyle := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height - 6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1)
	content := contentStyle.Render(strings.Join(changeLines, "\n"))

	// Instructions
	instructions := "ENTER: save changes | ESC: cancel and return"
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Align(lipgloss.Center).
		Width(m.Width)
	footer := instrStyle.Render(instructions)

	return lipgloss.JoinVertical(lipgloss.Top, title, content, footer)
}

// buildPendingChangesList builds a list of pending changes for display
func buildPendingChangesList(m *types.Model) []string {
	var changeLines []string

	// Check for moved permissions
	for _, perm := range m.Permissions {
		if perm.CurrentLevel != perm.OriginalLevel {
			changeLines = append(changeLines,
				fmt.Sprintf("• %s: %s → %s", perm.Name, perm.OriginalLevel, perm.CurrentLevel))
		}
	}

	// Check for resolved duplicates
	for _, dup := range m.Duplicates {
		if dup.KeepLevel != "" {
			otherLevels := []string{}
			for _, level := range dup.Levels {
				if level != dup.KeepLevel {
					otherLevels = append(otherLevels, level)
				}
			}
			if len(otherLevels) > 0 {
				changeLines = append(changeLines,
					fmt.Sprintf("• %s: Remove from %s (keep in %s)",
						dup.Name, strings.Join(otherLevels, ", "), dup.KeepLevel))
			}
		}
	}

	return changeLines
}

// handleEscapeKey handles ESC key with screen-specific behavior
func handleEscapeKey(m *types.Model) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenConfirmation:
		// On confirmation screen: ESC returns to the previous screen (organization)
		m.CurrentScreen = types.ScreenOrganization
	case types.ScreenDuplicates:
		// On duplicates screen: ESC should cancel/exit (only if no pending changes)
		if hasPendingChanges(m) {
			m.ShowModal = true
			m.ModalTitle = "Exit with Pending Changes"
			m.ModalBody = "You have pending permission moves or duplicate resolutions.\n\n" +
				"Do you want to discard these changes and exit?"
			m.ModalAction = "exit"
		}
		// If no pending changes, ESC does nothing (user should use Q to quit)
	case types.ScreenOrganization:
		// On organization screen: ESC should reset changes
		if hasPendingChanges(m) {
			m.ShowModal = true
			m.ModalTitle = "Reset All Changes"
			m.ModalBody = "Are you sure you want to reset all permission moves and duplicate resolutions?\n\n" +
				"This will undo all pending changes and return permissions to their original state."
			m.ModalAction = "reset"
		}
		// If no pending changes, ESC does nothing
	}
	return m
}

// handleModalInput handles keyboard input when modal is shown
func handleModalInput(m *types.Model, key string) *types.Model {
	switch key {
	case "y", "Y", "enter":
		switch m.ModalAction {
		case "reset":
			m = resetAllChanges(m)
		case "exit":
			// For exit action, reset changes and quit the application
			m = resetAllChanges(m)
			// Note: We can't directly quit from here, but we clear changes
			// The user will need to press Q to actually quit
		}
		m.ShowModal = false
		m.ModalTitle = ""
		m.ModalBody = ""
		m.ModalAction = ""
	case "n", "N", "escape", "esc":
		m.ShowModal = false
		m.ModalTitle = ""
		m.ModalBody = ""
		m.ModalAction = ""
	}
	return m
}

// hasPendingChanges checks if there are any pending permission moves or duplicate resolutions
func hasPendingChanges(m *types.Model) bool {
	// Check if any permissions have been moved from their original level
	for _, perm := range m.Permissions {
		if perm.CurrentLevel != perm.OriginalLevel {
			return true
		}
	}

	// Check if any duplicates have been resolved
	for _, dup := range m.Duplicates {
		if dup.KeepLevel != "" {
			return true
		}
	}

	return false
}

// resetAllChanges resets all pending permission moves and duplicate resolutions
func resetAllChanges(m *types.Model) *types.Model {
	// Reset permissions to their original levels
	for i := range m.Permissions {
		originalLevel := m.Permissions[i].OriginalLevel
		currentLevel := m.Permissions[i].CurrentLevel

		if originalLevel != currentLevel {
			// Move permission back to original level
			movePermissionBetweenLevels(m, m.Permissions[i].Name, currentLevel, originalLevel)
			m.Permissions[i].CurrentLevel = originalLevel
		}
	}

	// Reset duplicate resolutions
	for i := range m.Duplicates {
		m.Duplicates[i].KeepLevel = ""
	}

	// Reset column selections to 0
	m.ColumnSelections = [3]int{0, 0, 0}

	return m
}

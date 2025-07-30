package ui

import (
	"fmt"
	"strings"

	"claude-permissions/debug"
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
	if m.ActiveModal != nil {
		return handleActiveModalInput(m, key), nil
	}

	return handleNonModalKeys(m, msg, key)
}

// handleNonModalKeys handles key input when no modal is shown
func handleNonModalKeys(m *types.Model, msg tea.KeyMsg, key string) (*types.Model, tea.Cmd) {
	if key == "tab" {
		return handleTabKey(m), nil
	}

	// Handle ESC key for reset functionality on permissions screen
	if key == keyEscapeLong || key == keyEscape || msg.Key().Code == tea.KeyEscape {
		return handleEscapeKey(m), nil
	}

	// Handle ENTER key for confirmation screen transition
	if key == keyEnter || msg.Key().Code == tea.KeyEnter {
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
	case types.ScreenDuplicates, types.ScreenOrganization:
		// Launch confirm changes modal if there are pending changes
		if hasPendingChanges(m) {
			m.ActiveModal = NewConfirmChangesModal(m)
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
		return handleLeftNavigation(m)
	case "right", "l":
		return handleRightNavigation(m)
	}
	return m
}

// handleLeftNavigation handles left arrow navigation
func handleLeftNavigation(m *types.Model) *types.Model {
	if m.CurrentScreen == types.ScreenOrganization && m.FocusedColumn > 0 {
		// Block navigation if there are unresolved duplicates
		if hasUnresolvedDuplicates(m) {
			return m
		}
		m.FocusedColumn--
	}
	return m
}

// handleRightNavigation handles right arrow navigation
func handleRightNavigation(m *types.Model) *types.Model {
	if m.CurrentScreen == types.ScreenOrganization && m.FocusedColumn < 2 {
		// Block navigation if there are unresolved duplicates
		if hasUnresolvedDuplicates(m) {
			return m
		}
		m.FocusedColumn++
	}
	return m
}

// handleNumberKeys handles 1/2/3 keys for moving permissions or resolving duplicates
func handleNumberKeys(m *types.Model, key string) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		return handleDuplicateResolution(m, key)
	case types.ScreenOrganization:
		// Block permission moves if there are unresolved duplicates
		if hasUnresolvedDuplicates(m) {
			return m
		}
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
	keyUp         = "up"
	keyDown       = "down"
	keyEnter      = "enter"
	keyEscape     = "esc"
	keyEscapeLong = "escape"
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
	default:
		return m
	}
	m.DuplicatesTable, _ = m.DuplicatesTable.Update(keyMsg)
	return m
}

// handleOrganizationNavigation handles up/down navigation for organization screen
func handleOrganizationNavigation(m *types.Model, key string) *types.Model {
	// Block navigation if there are unresolved duplicates
	if hasUnresolvedDuplicates(m) {
		return m
	}

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
	if m.ActiveModal == nil {
		return baseContent
	}

	// Ask the modal what to render
	modalContent := m.ActiveModal.RenderModal(m.Width, m.Height)

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

// buildPendingChangesList builds a list of pending changes for display, grouped by destination level
func buildPendingChangesList(m *types.Model) []string {
	var changeLines []string

	// Add permission moves grouped by destination level
	permissionChanges := buildPermissionMovesList(m)
	changeLines = append(changeLines, permissionChanges...)

	// Add duplicate resolutions section
	duplicateChanges := buildDuplicateResolutionsList(m)
	changeLines = append(changeLines, duplicateChanges...)

	return changeLines
}

// buildPermissionMovesList builds the permission moves section
func buildPermissionMovesList(m *types.Model) []string {
	var changeLines []string

	// Group permission moves by destination level
	movesByLevel := map[string][]types.Permission{
		types.LevelLocal: {},
		types.LevelRepo:  {},
		types.LevelUser:  {},
	}

	// Collect moved permissions by destination level
	for _, perm := range m.Permissions {
		if perm.CurrentLevel != perm.OriginalLevel {
			movesByLevel[perm.CurrentLevel] = append(movesByLevel[perm.CurrentLevel], perm)
		}
	}

	// Add permission moves grouped by destination level
	levelOrder := []string{types.LevelLocal, types.LevelRepo, types.LevelUser}
	for _, level := range levelOrder {
		moves := movesByLevel[level]
		if len(moves) > 0 {
			changeLines = append(changeLines, buildLevelSection(level, moves)...)
		}
	}

	return changeLines
}

// buildLevelSection builds a section for a specific level
func buildLevelSection(level string, moves []types.Permission) []string {
	section := make([]string, 0, len(moves)+2) // header + moves + empty line

	// Add section header
	levelStyled := getLevelStyledText(level)
	section = append(section, fmt.Sprintf("Moving to %s Level:", levelStyled))

	// Sort permissions alphabetically within level
	sortPermissionsByName(moves)

	// Add each permission move
	for _, perm := range moves {
		originalLevelStyled := getLevelStyledText(perm.OriginalLevel)
		currentLevelStyled := getLevelStyledText(perm.CurrentLevel)
		section = append(
			section,
			fmt.Sprintf(
				"• %s: %s → %s",
				perm.Name,
				originalLevelStyled,
				currentLevelStyled,
			),
		)
	}
	section = append(section, "") // Empty line after each section

	return section
}

// buildDuplicateResolutionsList builds the duplicate resolutions section
func buildDuplicateResolutionsList(m *types.Model) []string {
	var duplicateResolutions []string

	for _, dup := range m.Duplicates {
		if dup.KeepLevel != "" {
			otherLevels := []string{}
			for _, level := range dup.Levels {
				if level != dup.KeepLevel {
					// Apply level colors to level names
					otherLevels = append(otherLevels, getLevelStyledText(level))
				}
			}
			if len(otherLevels) > 0 {
				// Apply level color to keep level
				keepLevelStyled := getLevelStyledText(dup.KeepLevel)
				duplicateResolutions = append(duplicateResolutions,
					fmt.Sprintf("• %s: Remove from %s (keep in %s)",
						dup.Name, strings.Join(otherLevels, ", "), keepLevelStyled))
			}
		}
	}

	var result []string
	if len(duplicateResolutions) > 0 {
		result = append(result, "Duplicate Resolutions:")
		result = append(result, duplicateResolutions...)
	}

	return result
}

// sortPermissionsByName sorts permissions alphabetically by name
func sortPermissionsByName(perms []types.Permission) {
	for i := 0; i < len(perms)-1; i++ {
		for j := i + 1; j < len(perms); j++ {
			if strings.ToLower(perms[i].Name) > strings.ToLower(perms[j].Name) {
				perms[i], perms[j] = perms[j], perms[i]
			}
		}
	}
}

// handleEscapeKey handles ESC key with screen-specific behavior
func handleEscapeKey(m *types.Model) *types.Model {
	switch m.CurrentScreen {
	case types.ScreenDuplicates:
		// On duplicates screen: ESC should cancel/exit (only if no pending changes)
		if hasPendingChanges(m) {
			m.ActiveModal = NewSmallModal(
				"Exit with Pending Changes",
				"You have pending permission moves or duplicate resolutions.\n\n"+
					"Do you want to discard these changes and exit?",
				"exit",
			)
		}
		// If no pending changes, ESC does nothing (user should use Q to quit)
	case types.ScreenOrganization:
		// On organization screen: ESC should reset changes
		if hasPendingChanges(m) {
			m.ActiveModal = NewSmallModal(
				"Reset All Changes",
				"Are you sure you want to reset all permission moves and duplicate resolutions?\n\n"+
					"This will undo all pending changes and return permissions to their original state.",
				"reset",
			)
		}
		// If no pending changes, ESC does nothing
	}
	return m
}

// handleActiveModalInput handles keyboard input for new modal interface
func handleActiveModalInput(m *types.Model, key string) *types.Model {
	handled, result := m.ActiveModal.HandleInput(key)
	if !handled {
		return m
	}

	// Process the result based on modal type and action
	switch resultStr := result.(string); resultStr {
	case "yes":
		// For small modals, determine action based on the modal's Action field
		if smallModal, ok := m.ActiveModal.(*SmallModal); ok {
			switch smallModal.Action {
			case "reset":
				m = resetAllChanges(m)
			case "exit":
				// For exit action, reset changes and clear modal
				m = resetAllChanges(m)
			}
		}
		m.ActiveModal = nil
	case "no":
		// Just close the modal without action
		m.ActiveModal = nil
	case "execute":
		// For confirm changes modal - execute all changes and close modal
		// TODO: Here we would actually save the changes to files
		// For now, just close the modal (changes remain in memory)
		m.ActiveModal = nil
	case "cancel":
		// For confirm changes modal - just close modal and return to main screen
		m.ActiveModal = nil
	case "quit":
		// For confirm changes modal - quit application
		// The main program loop should handle this by checking for quit signals
		m.ActiveModal = nil
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

// getLevelStyledText returns a styled level name using the appropriate theme color
func getLevelStyledText(level string) string {
	switch level {
	case types.LevelLocal:
		return LocalLevelStyle.Render(level)
	case types.LevelRepo:
		return RepoLevelStyle.Render(level)
	case types.LevelUser:
		return UserLevelStyle.Render(level)
	default:
		return level
	}
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

// handleLaunchConfirmChanges handles the debug message to launch confirm changes screen
func handleLaunchConfirmChanges(
	m *types.Model,
	msg debug.LaunchConfirmChangesMsg,
) *types.Model {
	// Apply mock changes to model
	applyMockChangesToModel(m, msg.Request)

	// Launch confirm changes modal
	m.ActiveModal = NewConfirmChangesModal(m)

	return m
}

// applyMockChangesToModel applies mock permission moves and duplicate resolutions to the model
func applyMockChangesToModel(m *types.Model, request *debug.LaunchConfirmChangesRequest) {
	// Apply permission moves
	for _, move := range request.MockChanges.PermissionMoves {
		// Find the permission in the model
		for i := range m.Permissions {
			if m.Permissions[i].Name == move.Name {
				// Set original level if not already set
				if m.Permissions[i].OriginalLevel == "" {
					m.Permissions[i].OriginalLevel = move.From
				}
				// Update the permission's current level
				m.Permissions[i].CurrentLevel = move.To
				break
			}
		}

		// Also update the level permissions arrays
		updateModelLevelPermissions(m, move.Name, move.From, move.To)
	}

	// Apply duplicate resolutions
	for _, resolution := range request.MockChanges.DuplicateResolutions {
		// Find the duplicate in the model
		for i := range m.Duplicates {
			if m.Duplicates[i].Name == resolution.Name {
				m.Duplicates[i].KeepLevel = resolution.KeepLevel
				break
			}
		}
	}
}

// updateModelLevelPermissions updates the permission arrays in the appropriate levels
func updateModelLevelPermissions(m *types.Model, permName, fromLevel, toLevel string) {
	// Remove from source level
	switch fromLevel {
	case types.LevelLocal:
		m.LocalLevel.Permissions = removePermissionFromArray(m.LocalLevel.Permissions, permName)
	case types.LevelRepo:
		m.RepoLevel.Permissions = removePermissionFromArray(m.RepoLevel.Permissions, permName)
	case types.LevelUser:
		m.UserLevel.Permissions = removePermissionFromArray(m.UserLevel.Permissions, permName)
	}

	// Add to target level (alphabetically sorted)
	switch toLevel {
	case types.LevelLocal:
		m.LocalLevel.Permissions = addPermissionToArraySorted(m.LocalLevel.Permissions, permName)
	case types.LevelRepo:
		m.RepoLevel.Permissions = addPermissionToArraySorted(m.RepoLevel.Permissions, permName)
	case types.LevelUser:
		m.UserLevel.Permissions = addPermissionToArraySorted(m.UserLevel.Permissions, permName)
	}
}

// removePermissionFromArray removes a permission from a slice
func removePermissionFromArray(perms []string, permission string) []string {
	for i, perm := range perms {
		if perm == permission {
			return append(perms[:i], perms[i+1:]...)
		}
	}
	return perms
}

// addPermissionToArraySorted adds a permission to a slice in alphabetical order
func addPermissionToArraySorted(perms []string, permission string) []string {
	// Find insertion point
	insertIndex := 0
	for i, perm := range perms {
		if permission < perm {
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

// hasUnresolvedDuplicates checks if there are duplicates that need to be committed.
//
// Duplicates are auto-assigned KeepLevel values during initialization based on priority
// (User > Repo > Local). However, they are considered "unresolved" until the user
// commits them via ENTER → confirmation modal → save to files.
//
// The presence of ANY duplicates in m.Duplicates means they need resolution/commitment,
// regardless of their KeepLevel assignment. Only after successful commit are duplicates
// removed from m.Duplicates, making the organization screen accessible.
//
// Workflow:
// 1. Duplicates created with auto-selected KeepLevel (highest priority)
// 2. User optionally changes selections with 1/2/3 keys
// 3. User hits ENTER → confirmation modal
// 4. User confirms → duplicates committed to files, m.Duplicates cleared
// 5. Organization screen becomes accessible, app switches to it
func hasUnresolvedDuplicates(m *types.Model) bool {
	return len(m.Duplicates) > 0
}

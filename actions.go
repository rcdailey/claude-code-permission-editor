// Package main provides an interactive TUI for managing Claude Code tool permissions.
package main

import (
	"fmt"
	"sort"
	"strings"

	"claude-permissions/types"
)

// moveSelectedPermissions moves selected permissions to the specified level
func moveSelectedPermissions(m *types.Model, toLevel string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	hasSelected := false
	for i := range m.Permissions {
		if m.Permissions[i].Selected {
			hasSelected = true
			m.Permissions[i].PendingMove = toLevel
			// Update the list item to reflect the change
			m.PermissionsList.SetItem(i, m.Permissions[i])
		}
	}

	// If no selections, move current cursor item
	if !hasSelected {
		currentIdx := m.PermissionsList.Index()
		if currentIdx < len(m.Permissions) {
			m.Permissions[currentIdx].PendingMove = toLevel
			// Update the list item to reflect the change
			m.PermissionsList.SetItem(currentIdx, m.Permissions[currentIdx])
		}
	}

	updateActionQueue(m)
}

// setDuplicateKeepLevel sets which level to keep for a duplicate
func setDuplicateKeepLevel(m *types.Model, level string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	cursor := m.DuplicatesTable.Cursor()
	if cursor < len(m.Duplicates) {
		m.Duplicates[cursor].KeepLevel = level
		updateDuplicatesTableRows(m)
		updateActionQueue(m)
	}
}

// updateActionQueue rebuilds the action queue based on current state
func updateActionQueue(m *types.Model) {
	m.Actions = []types.Action{}

	// Add duplicate resolutions
	for _, dup := range m.Duplicates {
		for _, level := range dup.Levels {
			if level != dup.KeepLevel {
				m.Actions = append(m.Actions, types.Action{
					Type:       types.ActionDuplicate,
					Permission: dup.Name,
					FromLevel:  level,
					ToLevel:    "",
				})
			}
		}
	}

	// Add moves
	for _, perm := range m.Permissions {
		if perm.PendingMove != "" && perm.PendingMove != perm.CurrentLevel {
			m.Actions = append(m.Actions, types.Action{
				Type:       types.ActionMove,
				Permission: perm.Name,
				FromLevel:  perm.CurrentLevel,
				ToLevel:    perm.PendingMove,
			})
		}
	}
}

// clearPendingMoves clears all pending moves and selections
func clearPendingMoves(m *types.Model) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	for i := range m.Permissions {
		m.Permissions[i].PendingMove = ""
		m.Permissions[i].Selected = false
	}

	for i := range m.Duplicates {
		m.Duplicates[i].Selected = false
	}
}

// generateConfirmationText creates the confirmation dialog text
func generateConfirmationText(m *types.Model) string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	var lines []string
	lines = append(lines, "Confirm the following changes:")
	lines = append(lines, "")

	// Group actions by type
	duplicateActions := []types.Action{}
	moveActions := []types.Action{}

	for _, action := range m.Actions {
		switch action.Type {
		case types.ActionDuplicate:
			duplicateActions = append(duplicateActions, action)
		case types.ActionMove:
			moveActions = append(moveActions, action)
		}
	}

	// Show duplicate resolutions
	if len(duplicateActions) > 0 {
		lines = append(lines, "Duplicate Resolutions:")
		for _, action := range duplicateActions {
			lines = append(lines, fmt.Sprintf("  • Remove %s from %s", action.Permission, action.FromLevel))
		}
		lines = append(lines, "")
	}

	// Show moves
	if len(moveActions) > 0 {
		lines = append(lines, "Permission Moves:")
		for _, action := range moveActions {
			lines = append(lines, fmt.Sprintf("  • Move %s: %s → %s", action.Permission, action.FromLevel, action.ToLevel))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "Press ENTER to confirm, ESC to cancel")

	return strings.Join(lines, "\n")
}

// executeActions applies all queued actions to the settings files
func executeActions(m *types.Model) error {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	// Load current settings
	userSettings, repoSettings, localSettings, err := loadAllSettings(m)
	if err != nil {
		return err
	}

	// Apply actions
	applyActionsToSettings(m, &userSettings, &repoSettings, &localSettings)

	// Sort all permission lists
	sort.Strings(userSettings.Allow)
	sort.Strings(repoSettings.Allow)
	sort.Strings(localSettings.Allow)

	// Save settings
	return saveAllSettings(m, userSettings, repoSettings, localSettings)
}

// loadAllSettings loads settings from all three levels
func loadAllSettings(m *types.Model) (types.Settings, types.Settings, types.Settings, error) {
	userSettings, err := loadSettingsFromFile(m.UserLevel.Path)
	if err != nil && m.UserLevel.Exists {
		return types.Settings{}, types.Settings{}, types.Settings{}, fmt.Errorf("failed to load user settings: %w", err)
	}

	repoSettings, err := loadSettingsFromFile(m.RepoLevel.Path)
	if err != nil && m.RepoLevel.Exists {
		return types.Settings{}, types.Settings{}, types.Settings{}, fmt.Errorf("failed to load repo settings: %w", err)
	}

	localSettings, err := loadSettingsFromFile(m.LocalLevel.Path)
	if err != nil && m.LocalLevel.Exists {
		return types.Settings{}, types.Settings{}, types.Settings{}, fmt.Errorf("failed to load local settings: %w", err)
	}

	return userSettings, repoSettings, localSettings, nil
}

// applyActionsToSettings applies all actions to the provided settings
func applyActionsToSettings(m *types.Model, userSettings, repoSettings, localSettings *types.Settings) {
	for _, action := range m.Actions {
		switch action.Type {
		case types.ActionDuplicate:
			removePermissionFromLevel(action.FromLevel, action.Permission, userSettings, repoSettings, localSettings)
		case types.ActionMove:
			removePermissionFromLevel(action.FromLevel, action.Permission, userSettings, repoSettings, localSettings)
			addPermissionToLevel(action.ToLevel, action.Permission, userSettings, repoSettings, localSettings)
		}
	}
}

// removePermissionFromLevel removes a permission from the specified level
func removePermissionFromLevel(
	level, permission string,
	userSettings, repoSettings, localSettings *types.Settings,
) {
	switch level {
	case types.LevelUser:
		userSettings.Allow = removePermission(userSettings.Allow, permission)
	case types.LevelRepo:
		repoSettings.Allow = removePermission(repoSettings.Allow, permission)
	case types.LevelLocal:
		localSettings.Allow = removePermission(localSettings.Allow, permission)
	}
}

// addPermissionToLevel adds a permission to the specified level
func addPermissionToLevel(
	level, permission string,
	userSettings, repoSettings, localSettings *types.Settings,
) {
	switch level {
	case types.LevelUser:
		userSettings.Allow = addPermission(userSettings.Allow, permission)
	case types.LevelRepo:
		repoSettings.Allow = addPermission(repoSettings.Allow, permission)
	case types.LevelLocal:
		localSettings.Allow = addPermission(localSettings.Allow, permission)
	}
}

// saveAllSettings saves all settings to their respective files
func saveAllSettings(m *types.Model, userSettings, repoSettings, localSettings types.Settings) error {
	if err := saveSettingsToFile(m.UserLevel.Path, userSettings); err != nil {
		return fmt.Errorf("failed to save user settings: %w", err)
	}

	if err := saveSettingsToFile(m.RepoLevel.Path, repoSettings); err != nil {
		return fmt.Errorf("failed to save repo settings: %w", err)
	}

	if err := saveSettingsToFile(m.LocalLevel.Path, localSettings); err != nil {
		return fmt.Errorf("failed to save local settings: %w", err)
	}

	return nil
}

// removePermission removes a permission from a slice
func removePermission(perms []string, perm string) []string {
	var result []string
	for _, p := range perms {
		if p != perm {
			result = append(result, p)
		}
	}
	return result
}

// addPermission adds a permission to a slice if it doesn't already exist
func addPermission(perms []string, perm string) []string {
	// Check if already exists
	for _, p := range perms {
		if p == perm {
			return perms
		}
	}
	return append(perms, perm)
}

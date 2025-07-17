// Package main provides an interactive TUI for managing Claude Code tool permissions.
package main

import (
	"fmt"
	"sort"
	"strings"
)

// moveSelectedPermissions moves selected permissions to the specified level
func (m *Model) moveSelectedPermissions(toLevel string) {
	hasSelected := false
	for i := range m.permissions {
		if m.permissions[i].Selected {
			hasSelected = true
			m.permissions[i].PendingMove = toLevel
			// Update the list item to reflect the change
			m.permissionsList.SetItem(i, m.permissions[i])
		}
	}

	// If no selections, move current cursor item
	if !hasSelected {
		currentIdx := m.permissionsList.Index()
		if currentIdx < len(m.permissions) {
			m.permissions[currentIdx].PendingMove = toLevel
			// Update the list item to reflect the change
			m.permissionsList.SetItem(currentIdx, m.permissions[currentIdx])
		}
	}

	m.updateActionQueue()
}

// setDuplicateKeepLevel sets which level to keep for a duplicate
func (m *Model) setDuplicateKeepLevel(level string) {
	cursor := m.duplicatesTable.Cursor()
	if cursor < len(m.duplicates) {
		m.duplicates[cursor].KeepLevel = level
		m.updateDuplicatesTableRows()
		m.updateActionQueue()
	}
}

// updateActionQueue rebuilds the action queue based on current state
func (m *Model) updateActionQueue() {
	m.actions = []Action{}

	// Add duplicate resolutions
	for _, dup := range m.duplicates {
		for _, level := range dup.Levels {
			if level != dup.KeepLevel {
				m.actions = append(m.actions, Action{
					Type:       ActionDuplicate,
					Permission: dup.Name,
					FromLevel:  level,
					ToLevel:    "",
				})
			}
		}
	}

	// Add moves
	for _, perm := range m.permissions {
		if perm.PendingMove != "" && perm.PendingMove != perm.CurrentLevel {
			m.actions = append(m.actions, Action{
				Type:       ActionMove,
				Permission: perm.Name,
				FromLevel:  perm.CurrentLevel,
				ToLevel:    perm.PendingMove,
			})
		}
	}
}

// clearPendingMoves clears all pending moves and selections
func (m *Model) clearPendingMoves() {
	for i := range m.permissions {
		m.permissions[i].PendingMove = ""
		m.permissions[i].Selected = false
	}

	for i := range m.duplicates {
		m.duplicates[i].Selected = false
	}
}

// generateConfirmationText creates the confirmation dialog text
func (m Model) generateConfirmationText() string {
	var lines []string
	lines = append(lines, "Confirm the following changes:")
	lines = append(lines, "")

	// Group actions by type
	duplicateActions := []Action{}
	moveActions := []Action{}

	for _, action := range m.actions {
		switch action.Type {
		case ActionDuplicate:
			duplicateActions = append(duplicateActions, action)
		case ActionMove:
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
func (m Model) executeActions() error {
	// Load current settings
	userSettings, repoSettings, localSettings, err := m.loadAllSettings()
	if err != nil {
		return err
	}

	// Apply actions
	m.applyActionsToSettings(&userSettings, &repoSettings, &localSettings)

	// Sort all permission lists
	sort.Strings(userSettings.Allow)
	sort.Strings(repoSettings.Allow)
	sort.Strings(localSettings.Allow)

	// Save settings
	return m.saveAllSettings(userSettings, repoSettings, localSettings)
}

// loadAllSettings loads settings from all three levels
func (m Model) loadAllSettings() (Settings, Settings, Settings, error) {
	userSettings, err := loadSettingsFromFile(m.userLevel.Path)
	if err != nil && m.userLevel.Exists {
		return Settings{}, Settings{}, Settings{}, fmt.Errorf("failed to load user settings: %w", err)
	}

	repoSettings, err := loadSettingsFromFile(m.repoLevel.Path)
	if err != nil && m.repoLevel.Exists {
		return Settings{}, Settings{}, Settings{}, fmt.Errorf("failed to load repo settings: %w", err)
	}

	localSettings, err := loadSettingsFromFile(m.localLevel.Path)
	if err != nil && m.localLevel.Exists {
		return Settings{}, Settings{}, Settings{}, fmt.Errorf("failed to load local settings: %w", err)
	}

	return userSettings, repoSettings, localSettings, nil
}

// applyActionsToSettings applies all actions to the provided settings
func (m Model) applyActionsToSettings(userSettings, repoSettings, localSettings *Settings) {
	for _, action := range m.actions {
		switch action.Type {
		case ActionDuplicate:
			m.removePermissionFromLevel(action.FromLevel, action.Permission, userSettings, repoSettings, localSettings)
		case ActionMove:
			m.removePermissionFromLevel(action.FromLevel, action.Permission, userSettings, repoSettings, localSettings)
			m.addPermissionToLevel(action.ToLevel, action.Permission, userSettings, repoSettings, localSettings)
		}
	}
}

// removePermissionFromLevel removes a permission from the specified level
func (m Model) removePermissionFromLevel(
	level, permission string,
	userSettings, repoSettings, localSettings *Settings,
) {
	switch level {
	case LevelUser:
		userSettings.Allow = removePermission(userSettings.Allow, permission)
	case LevelRepo:
		repoSettings.Allow = removePermission(repoSettings.Allow, permission)
	case LevelLocal:
		localSettings.Allow = removePermission(localSettings.Allow, permission)
	}
}

// addPermissionToLevel adds a permission to the specified level
func (m Model) addPermissionToLevel(
	level, permission string,
	userSettings, repoSettings, localSettings *Settings,
) {
	switch level {
	case LevelUser:
		userSettings.Allow = addPermission(userSettings.Allow, permission)
	case LevelRepo:
		repoSettings.Allow = addPermission(repoSettings.Allow, permission)
	case LevelLocal:
		localSettings.Allow = addPermission(localSettings.Allow, permission)
	}
}

// saveAllSettings saves all settings to their respective files
func (m Model) saveAllSettings(userSettings, repoSettings, localSettings Settings) error {
	if err := saveSettingsToFile(m.userLevel.Path, userSettings); err != nil {
		return fmt.Errorf("failed to save user settings: %w", err)
	}

	if err := saveSettingsToFile(m.repoLevel.Path, repoSettings); err != nil {
		return fmt.Errorf("failed to save repo settings: %w", err)
	}

	if err := saveSettingsToFile(m.localLevel.Path, localSettings); err != nil {
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

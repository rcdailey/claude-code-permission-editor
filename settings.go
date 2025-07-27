package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"claude-permissions/types"
)

// loadUserLevel loads user-level settings with chezmoi integration
func loadUserLevel() (types.SettingsLevel, error) {
	// Use command line override if provided
	if *userFile != "" {
		return loadSettingsLevel("User", *userFile)
	}

	// Check for chezmoi integration
	if path := getChezmoidUserPath(); path != "" {
		return loadSettingsLevel("User", path)
	}

	// Fallback to standard path
	home, err := os.UserHomeDir()
	if err != nil {
		return types.SettingsLevel{}, err
	}

	path := filepath.Join(home, ".claude", "settings.json")
	return loadSettingsLevel("User", path)
}

// getChezmoidUserPath returns the chezmoi source path for user settings
func getChezmoidUserPath() string {
	// Check if chezmoi is available
	if _, err := exec.LookPath("chezmoi"); err != nil {
		return ""
	}

	// Try to get source path
	cmd := exec.Command("chezmoi", "source-path", "~/.claude/settings.json")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	path := strings.TrimSpace(string(output))
	if path == "" {
		return ""
	}

	// Verify the file exists
	if _, err := os.Stat(path); err != nil {
		return ""
	}

	return path
}

// loadRepoLevel loads repository-level settings
func loadRepoLevel() (types.SettingsLevel, error) {
	// Use command line override if provided
	if *repoFile != "" {
		return loadSettingsLevel("Repo", *repoFile)
	}

	repoRoot, err := findGitRoot()
	if err != nil {
		return types.SettingsLevel{
			Name:        types.LevelRepo,
			Path:        "",
			Permissions: []string{},
			Exists:      false,
		}, nil
	}

	path := filepath.Join(repoRoot, ".claude", "settings.json")
	return loadSettingsLevel("Repo", path)
}

// loadLocalLevel loads local-level settings
func loadLocalLevel() (types.SettingsLevel, error) {
	// Use command line override if provided
	if *localFile != "" {
		return loadSettingsLevel("Local", *localFile)
	}

	repoRoot, err := findGitRoot()
	if err != nil {
		return types.SettingsLevel{
			Name:        types.LevelLocal,
			Path:        "",
			Permissions: []string{},
			Exists:      false,
		}, nil
	}

	path := filepath.Join(repoRoot, ".claude", "settings.local.json")
	return loadSettingsLevel("Local", path)
}

// findGitRoot finds the root of the git repository
func findGitRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git", "config")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a git repository")
}

// loadSettingsLevel loads settings from a specific file
func loadSettingsLevel(name, path string) (types.SettingsLevel, error) {
	level := types.SettingsLevel{
		Name:        name,
		Path:        path,
		Permissions: []string{},
		Exists:      false,
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return level, nil // Not an error, just doesn't exist
	}

	// Read file
	data, err := os.ReadFile(
		path,
	) // #nosec G304 - path is validated and user-controlled config file
	if err != nil {
		return level, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse JSON
	var settings types.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return level, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	level.Exists = true
	level.Permissions = settings.Allow
	if level.Permissions == nil {
		level.Permissions = []string{}
	}

	// Sort permissions alphabetically
	sort.Strings(level.Permissions)

	return level, nil
}

// Removed unused functions loadSettingsFromFile and saveSettingsToFile
// These will be implemented when the action system is activated

// consolidatePermissions creates a unified view of all permissions
func consolidatePermissions(user, repo, local types.SettingsLevel) []types.Permission {
	permMap := make(map[string]types.Permission)

	// Add all permissions from all levels
	for _, perm := range user.Permissions {
		permMap[perm] = types.Permission{
			Name:         perm,
			CurrentLevel: types.LevelUser,
			PendingMove:  "",
			Selected:     false,
		}
	}

	for _, perm := range repo.Permissions {
		if _, exists := permMap[perm]; !exists {
			permMap[perm] = types.Permission{
				Name:         perm,
				CurrentLevel: types.LevelRepo,
				PendingMove:  "",
				Selected:     false,
			}
		}
	}

	for _, perm := range local.Permissions {
		if _, exists := permMap[perm]; !exists {
			permMap[perm] = types.Permission{
				Name:         perm,
				CurrentLevel: types.LevelLocal,
				PendingMove:  "",
				Selected:     false,
			}
		}
	}

	// Convert to slice and sort
	permissions := make([]types.Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	sort.Slice(permissions, func(i, j int) bool {
		return strings.ToLower(permissions[i].Name) < strings.ToLower(permissions[j].Name)
	})

	return permissions
}

// autoResolveSameLevelDuplicates removes duplicate permissions within the same level
func autoResolveSameLevelDuplicates(level *types.SettingsLevel) int {
	seen := make(map[string]bool)
	cleaned := []string{}
	duplicatesRemoved := 0

	for _, perm := range level.Permissions {
		if !seen[perm] {
			seen[perm] = true
			cleaned = append(cleaned, perm)
		} else {
			duplicatesRemoved++
		}
	}

	level.Permissions = cleaned
	return duplicatesRemoved
}

// detectDuplicates finds permissions that exist in multiple levels
func detectDuplicates(user, repo, local types.SettingsLevel) []types.Duplicate {
	permCount := make(map[string][]string)

	// Count occurrences across levels
	for _, perm := range user.Permissions {
		permCount[perm] = append(permCount[perm], types.LevelUser)
	}
	for _, perm := range repo.Permissions {
		permCount[perm] = append(permCount[perm], types.LevelRepo)
	}
	for _, perm := range local.Permissions {
		permCount[perm] = append(permCount[perm], types.LevelLocal)
	}

	// Find duplicates
	var duplicates []types.Duplicate
	for perm, levels := range permCount {
		if len(levels) > 1 {
			// Default to keeping highest priority level (User > Repo > Local)
			keepLevel := types.LevelLocal
			for _, level := range levels {
				if level == types.LevelUser {
					keepLevel = types.LevelUser
					break
				} else if level == types.LevelRepo && keepLevel != types.LevelUser {
					keepLevel = types.LevelRepo
				}
			}

			duplicates = append(duplicates, types.Duplicate{
				Name:      perm,
				Levels:    levels,
				KeepLevel: keepLevel,
				Selected:  false,
			})
		}
	}

	// Sort duplicates alphabetically
	sort.Slice(duplicates, func(i, j int) bool {
		return strings.ToLower(duplicates[i].Name) < strings.ToLower(duplicates[j].Name)
	})

	return duplicates
}

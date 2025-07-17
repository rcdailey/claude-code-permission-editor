package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// loadUserLevel loads user-level settings with chezmoi integration
func loadUserLevel() (SettingsLevel, error) {
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
		return SettingsLevel{}, err
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
func loadRepoLevel() (SettingsLevel, error) {
	// Use command line override if provided
	if *repoFile != "" {
		return loadSettingsLevel("Repo", *repoFile)
	}

	repoRoot, err := findGitRoot()
	if err != nil {
		return SettingsLevel{Name: LevelRepo, Path: "", Permissions: []string{}, Exists: false}, nil
	}

	path := filepath.Join(repoRoot, ".claude", "settings.json")
	return loadSettingsLevel("Repo", path)
}

// loadLocalLevel loads local-level settings
func loadLocalLevel() (SettingsLevel, error) {
	// Use command line override if provided
	if *localFile != "" {
		return loadSettingsLevel("Local", *localFile)
	}

	repoRoot, err := findGitRoot()
	if err != nil {
		return SettingsLevel{Name: LevelLocal, Path: "", Permissions: []string{}, Exists: false}, nil
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
func loadSettingsLevel(name, path string) (SettingsLevel, error) {
	level := SettingsLevel{
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
	data, err := os.ReadFile(path) // #nosec G304 - path is validated and user-controlled config file
	if err != nil {
		return level, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Parse JSON
	var settings Settings
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

// loadSettingsFromFile loads settings from a file path
func loadSettingsFromFile(path string) (Settings, error) {
	settings := Settings{Allow: []string{}}

	if path == "" {
		return settings, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return settings, nil
	}

	data, err := os.ReadFile(path) // #nosec G304 - path is validated and user-controlled config file
	if err != nil {
		return settings, err
	}

	// Try to parse as JSON
	var fullSettings map[string]interface{}
	if err := json.Unmarshal(data, &fullSettings); err != nil {
		return settings, err
	}

	// Extract allow permissions if they exist
	if allow, ok := fullSettings["allow"].([]interface{}); ok {
		for _, perm := range allow {
			if str, ok := perm.(string); ok {
				settings.Allow = append(settings.Allow, str)
			}
		}
	}

	return settings, nil
}

// saveSettingsToFile saves settings to a file path
func saveSettingsToFile(path string, settings Settings) error {
	if path == "" {
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	// Load existing settings to preserve other fields
	var fullSettings map[string]interface{}
	// #nosec G304 - path is validated and user-controlled config file
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &fullSettings) // Ignore error - we'll create new if unmarshal fails
	}

	if fullSettings == nil {
		fullSettings = make(map[string]interface{})
	}

	// Update only the allow field
	fullSettings["allow"] = settings.Allow

	// Write back to file
	data, err := json.MarshalIndent(fullSettings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// consolidatePermissions creates a unified view of all permissions
func consolidatePermissions(user, repo, local SettingsLevel) []Permission {
	permMap := make(map[string]Permission)

	// Add all permissions from all levels
	for _, perm := range user.Permissions {
		permMap[perm] = Permission{
			Name:         perm,
			CurrentLevel: LevelUser,
			PendingMove:  "",
			Selected:     false,
		}
	}

	for _, perm := range repo.Permissions {
		if _, exists := permMap[perm]; !exists {
			permMap[perm] = Permission{
				Name:         perm,
				CurrentLevel: LevelRepo,
				PendingMove:  "",
				Selected:     false,
			}
		}
	}

	for _, perm := range local.Permissions {
		if _, exists := permMap[perm]; !exists {
			permMap[perm] = Permission{
				Name:         perm,
				CurrentLevel: LevelLocal,
				PendingMove:  "",
				Selected:     false,
			}
		}
	}

	// Convert to slice and sort
	permissions := make([]Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	sort.Slice(permissions, func(i, j int) bool {
		return strings.ToLower(permissions[i].Name) < strings.ToLower(permissions[j].Name)
	})

	return permissions
}

// detectDuplicates finds permissions that exist in multiple levels
func detectDuplicates(user, repo, local SettingsLevel) []Duplicate {
	permCount := make(map[string][]string)

	// Count occurrences across levels
	for _, perm := range user.Permissions {
		permCount[perm] = append(permCount[perm], LevelUser)
	}
	for _, perm := range repo.Permissions {
		permCount[perm] = append(permCount[perm], LevelRepo)
	}
	for _, perm := range local.Permissions {
		permCount[perm] = append(permCount[perm], LevelLocal)
	}

	// Find duplicates
	var duplicates []Duplicate
	for perm, levels := range permCount {
		if len(levels) > 1 {
			// Default to keeping highest priority level (User > Repo > Local)
			keepLevel := LevelLocal
			for _, level := range levels {
				if level == LevelUser {
					keepLevel = LevelUser
					break
				} else if level == LevelRepo && keepLevel != LevelUser {
					keepLevel = LevelRepo
				}
			}

			duplicates = append(duplicates, Duplicate{
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

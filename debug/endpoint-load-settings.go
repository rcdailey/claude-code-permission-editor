package debug

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"

	"claude-permissions/types"

	"github.com/charmbracelet/bubbles/v2/table"
)

func init() {
	RegisterEndpoint("/load-settings", handleLoadSettings)
}

// LoadSettingsRequest defines the request payload for loading settings from files
type LoadSettingsRequest struct {
	UserFile  string `json:"user_file,omitempty"`
	RepoFile  string `json:"repo_file,omitempty"`
	LocalFile string `json:"local_file,omitempty"`
}

// LoadSettingsResponse defines the response for load operation
type LoadSettingsResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	FilesLoaded int    `json:"files_loaded"`
	Duplicates  int    `json:"duplicates"`
	Permissions int    `json:"permissions"`
	Timestamp   string `json:"timestamp"`
}

// handleLoadSettings handles the POST /load-settings endpoint
func handleLoadSettings(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	// Parse request body
	var req LoadSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON in request body", http.StatusBadRequest, ds.logger)
		return
	}

	// Get current model
	model := ds.GetModel()
	if model == nil {
		writeErrorResponse(w, "Model not available", http.StatusInternalServerError, ds.logger)
		return
	}

	// Load settings from specified file paths
	userLevel, repoLevel, localLevel, filesLoaded, err := loadAllLevels(req)
	if err != nil {
		writeErrorResponse(w, "Failed to load settings: "+err.Error(),
			http.StatusInternalServerError, ds.logger)
		return
	}

	// Update model with new data
	model.Mutex.Lock()
	model.UserLevel = userLevel
	model.RepoLevel = repoLevel
	model.LocalLevel = localLevel

	// Rebuild permissions and duplicates
	model.Permissions = consolidatePermissions(userLevel, repoLevel, localLevel)
	model.Duplicates = findDuplicates(model.Permissions)

	// Recreate duplicates table with new data
	model.DuplicatesTable = createDuplicatesTable(model.Duplicates)
	model.Mutex.Unlock()

	response := LoadSettingsResponse{
		Success:     true,
		Message:     "Settings loaded successfully",
		FilesLoaded: filesLoaded,
		Duplicates:  len(model.Duplicates),
		Permissions: len(model.Permissions),
		Timestamp:   getCurrentTimestamp(),
	}

	ds.logger.LogEvent("settings_loaded", map[string]interface{}{
		"files_loaded": filesLoaded,
		"duplicates":   response.Duplicates,
		"permissions":  response.Permissions,
	})

	writeJSONResponse(w, response, ds.logger)
}

// loadAllLevels loads settings from specified files or defaults
func loadAllLevels(
	req LoadSettingsRequest,
) (types.SettingsLevel, types.SettingsLevel, types.SettingsLevel, int, error) {
	var filesLoaded int

	// Load user level
	userLevel, err := loadSettingsLevelFromPath("User", req.UserFile)
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, err
	}
	if userLevel.Exists {
		filesLoaded++
	}

	// Load repo level
	repoLevel, err := loadSettingsLevelFromPath("Repo", req.RepoFile)
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, err
	}
	if repoLevel.Exists {
		filesLoaded++
	}

	// Load local level
	localLevel, err := loadSettingsLevelFromPath("Local", req.LocalFile)
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, err
	}
	if localLevel.Exists {
		filesLoaded++
	}

	return userLevel, repoLevel, localLevel, filesLoaded, nil
}

// loadSettingsLevelFromPath loads settings from a specific path
func loadSettingsLevelFromPath(name, path string) (types.SettingsLevel, error) {
	if path == "" {
		// Use default empty level if no path provided
		return types.SettingsLevel{
			Name:        name,
			Path:        "",
			Permissions: []string{},
			Exists:      false,
		}, nil
	}

	// Validate file path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return types.SettingsLevel{
			Name:        name,
			Path:        path,
			Permissions: []string{},
			Exists:      false,
		}, nil
	}

	// Load the settings using the existing logic pattern
	level := types.SettingsLevel{
		Name:        name,
		Path:        path,
		Permissions: []string{},
		Exists:      false,
	}

	// Read and parse file
	data, err := os.ReadFile(path) // #nosec G304 - path is validated by caller
	if err != nil {
		return level, err
	}

	var settings types.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return level, err
	}

	level.Exists = true
	level.Permissions = settings.Allow
	if level.Permissions == nil {
		level.Permissions = []string{}
	}

	return level, nil
}

// consolidatePermissions combines permissions from all levels
func consolidatePermissions(
	userLevel, repoLevel, localLevel types.SettingsLevel,
) []types.Permission {
	// Pre-allocate slice with estimated capacity
	totalPerms := len(
		userLevel.Permissions,
	) + len(
		repoLevel.Permissions,
	) + len(
		localLevel.Permissions,
	)
	permissions := make([]types.Permission, 0, totalPerms)

	// Add user level permissions
	for _, perm := range userLevel.Permissions {
		permissions = append(permissions, types.Permission{
			Name:          perm,
			CurrentLevel:  types.LevelUser,
			OriginalLevel: types.LevelUser,
		})
	}

	// Add repo level permissions
	for _, perm := range repoLevel.Permissions {
		permissions = append(permissions, types.Permission{
			Name:          perm,
			CurrentLevel:  types.LevelRepo,
			OriginalLevel: types.LevelRepo,
		})
	}

	// Add local level permissions
	for _, perm := range localLevel.Permissions {
		permissions = append(permissions, types.Permission{
			Name:          perm,
			CurrentLevel:  types.LevelLocal,
			OriginalLevel: types.LevelLocal,
		})
	}

	// Sort by name for consistent ordering
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Name < permissions[j].Name
	})

	return permissions
}

// findDuplicates identifies duplicate permissions across levels
func findDuplicates(permissions []types.Permission) []types.Duplicate {
	permissionMap := make(map[string][]types.Permission)

	// Group permissions by name
	for _, perm := range permissions {
		permissionMap[perm.Name] = append(permissionMap[perm.Name], perm)
	}

	var duplicates []types.Duplicate
	for name, perms := range permissionMap {
		if len(perms) > 1 {
			levels := make([]string, len(perms))
			for i, perm := range perms {
				levels[i] = perm.CurrentLevel
			}

			// Auto-select keep level using priority (User > Repo > Local)
			keepLevel := determineKeepLevel(levels)

			duplicates = append(duplicates, types.Duplicate{
				Name:      name,
				Levels:    levels,
				KeepLevel: keepLevel,
			})
		}
	}

	// Sort duplicates by name for consistency
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].Name < duplicates[j].Name
	})

	return duplicates
}

// determineKeepLevel selects the highest priority level (User > Repo > Local)
func determineKeepLevel(levels []string) string {
	for _, level := range []string{types.LevelUser, types.LevelRepo, types.LevelLocal} {
		for _, l := range levels {
			if l == level {
				return level
			}
		}
	}
	return levels[0] // fallback
}

// createDuplicatesTable creates a table model for displaying duplicates
func createDuplicatesTable(duplicates []types.Duplicate) table.Model {
	columns := []table.Column{
		{Title: "Permission", Width: 30},
		{Title: "Found In", Width: 25},
		{Title: "Keep Level", Width: 15},
	}

	rows := []table.Row{}
	for _, dup := range duplicates {
		levelsStr := strings.Join(dup.Levels, ", ")
		keepLevel := dup.KeepLevel
		if keepLevel == "" {
			keepLevel = "None"
		}
		rows = append(rows, table.Row{dup.Name, levelsStr, keepLevel})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	return t
}

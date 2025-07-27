package debug

import (
	"net/http"

	"claude-permissions/types"
)

func init() {
	RegisterEndpoint("/state", handleState)
}

// StateResponse represents the complete application state
type StateResponse struct {
	UI        UIState    `json:"ui"`
	Data      DataState  `json:"data"`
	Files     FilesState `json:"files"`
	Errors    []string   `json:"errors"`
	Timestamp string     `json:"timestamp"`
}

// UIState represents the user interface state
type UIState struct {
	ActivePanel   string   `json:"active_panel"`
	SelectedItems []string `json:"selected_items"`
	ListPosition  int      `json:"list_position"`
	FilterText    string   `json:"filter_text"`
	ConfirmMode   bool     `json:"confirm_mode"`
	ConfirmText   string   `json:"confirm_text"`
	StatusMessage string   `json:"status_message"`
}

// DataState represents the application data state
type DataState struct {
	PermissionsCount int      `json:"permissions_count"`
	DuplicatesCount  int      `json:"duplicates_count"`
	PendingEdits     []string `json:"pending_edits"`
}

// FilesState represents the settings files state
type FilesState struct {
	UserExists  bool   `json:"user_exists"`
	RepoExists  bool   `json:"repo_exists"`
	LocalExists bool   `json:"local_exists"`
	UserPath    string `json:"user_path"`
	RepoPath    string `json:"repo_path"`
	LocalPath   string `json:"local_path"`
}

// handleState handles the GET /state endpoint
func handleState(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	model := ds.GetModel()
	if model == nil {
		writeErrorResponse(w, "Model not available", http.StatusInternalServerError, ds.logger)
		return
	}

	response := extractApplicationState(model)
	response.Timestamp = getCurrentTimestamp()

	ds.logger.LogEvent("state_extracted", map[string]interface{}{
		"active_panel":      response.UI.ActivePanel,
		"permissions_count": response.Data.PermissionsCount,
		"duplicates_count":  response.Data.DuplicatesCount,
	})

	writeJSONResponse(w, response, ds.logger)
}

// extractApplicationState extracts state information from the model using direct field access
func extractApplicationState(model *types.Model) StateResponse {
	model.Mutex.RLock()
	defer model.Mutex.RUnlock()

	return StateResponse{
		UI:     extractUIState(model),
		Data:   extractDataState(model),
		Files:  extractFilesState(model),
		Errors: []string{}, // No more reflection errors
	}
}

// extractUIState extracts UI-related state from the model
func extractUIState(model *types.Model) UIState {
	return UIState{
		ActivePanel: panelNumberToName(model.ActivePanel), // Direct field access
		SelectedItems: extractSelectedItems(
			model,
		), // Extract from current column selection
		ListPosition:  model.ColumnSelections[model.FocusedColumn], // Current selection in focused column
		FilterText:    "",                                          // No filter in current UI implementation
		ConfirmMode:   model.ConfirmMode,                           // Direct field access
		ConfirmText:   model.ConfirmText,                           // Direct field access
		StatusMessage: model.StatusMessage,                         // Direct field access
	}
}

// extractDataState extracts data-related state from the model
func extractDataState(model *types.Model) DataState {
	return DataState{
		PermissionsCount: len(model.Permissions), // Direct field access
		DuplicatesCount:  len(model.Duplicates),  // Direct field access
		PendingEdits:     extractPendingEdits(model),
	}
}

// extractFilesState extracts settings files state from the model
func extractFilesState(model *types.Model) FilesState {
	return FilesState{
		UserExists:  model.UserLevel.Exists,  // Direct field access
		RepoExists:  model.RepoLevel.Exists,  // Direct field access
		LocalExists: model.LocalLevel.Exists, // Direct field access
		UserPath:    model.UserLevel.Path,    // Direct field access
		RepoPath:    model.RepoLevel.Path,    // Direct field access
		LocalPath:   model.LocalLevel.Path,   // Direct field access
	}
}

// extractSelectedItems extracts currently selected items based on UI state
func extractSelectedItems(model *types.Model) []string {
	var selectedItems []string

	// Get permissions for the currently focused column
	var targetLevel string
	switch model.FocusedColumn {
	case 0:
		targetLevel = types.LevelLocal
	case 1:
		targetLevel = types.LevelRepo
	case 2:
		targetLevel = types.LevelUser
	}

	// Find permissions in the focused column
	var columnPerms []types.Permission
	for _, perm := range model.Permissions {
		if perm.CurrentLevel == targetLevel {
			columnPerms = append(columnPerms, perm)
		}
	}

	// Add the currently selected permission if it exists
	selectionIndex := model.ColumnSelections[model.FocusedColumn]
	if selectionIndex < len(columnPerms) {
		selectedItems = append(selectedItems, columnPerms[selectionIndex].Name)
	}

	return selectedItems
}

// extractPendingEdits extracts pending edits from model
func extractPendingEdits(model *types.Model) []string {
	var edits []string
	// Since we removed the action queue system, this now returns empty
	// In the future, this could check for actual permission edit state
	return edits
}

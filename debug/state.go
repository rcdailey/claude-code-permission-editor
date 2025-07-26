package debug

import (
	"net/http"

	"claude-permissions/types"
)

// StateResponse represents the complete application state
type StateResponse struct {
	UI        UIState        `json:"ui"`
	Data      DataState      `json:"data"`
	Files     FilesState     `json:"files"`
	Errors    []string       `json:"errors"`
	Timestamp string         `json:"timestamp"`
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
	ActionsQueued    int      `json:"actions_queued"`
	PendingMoves     []string `json:"pending_moves"`
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
func (ds *DebugServer) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	model := ds.GetModel()
	if model == nil {
		ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
		return
	}

	response := extractApplicationState(model)
	response.Timestamp = getCurrentTimestamp()

	ds.logger.LogEvent("state_extracted", map[string]interface{}{
		"active_panel":      response.UI.ActivePanel,
		"permissions_count": response.Data.PermissionsCount,
		"duplicates_count":  response.Data.DuplicatesCount,
		"actions_queued":    response.Data.ActionsQueued,
	})

	ds.writeJSONResponse(w, response)
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
		ActivePanel:   panelNumberToName(model.ActivePanel),  // Direct field access
		SelectedItems: []string{}, // TODO: Extract from PermissionsList if needed
		ListPosition:  0,          // TODO: Extract from PermissionsList if needed
		FilterText:    "",         // TODO: Extract from PermissionsList if needed
		ConfirmMode:   model.ConfirmMode,    // Direct field access
		ConfirmText:   model.ConfirmText,    // Direct field access
		StatusMessage: model.StatusMessage,  // Direct field access
	}
}

// extractDataState extracts data-related state from the model
func extractDataState(model *types.Model) DataState {
	return DataState{
		PermissionsCount: len(model.Permissions),  // Direct field access
		DuplicatesCount:  len(model.Duplicates),   // Direct field access
		ActionsQueued:    len(model.Actions),      // Direct field access
		PendingMoves:     extractPendingMoves(model.Actions),
		PendingEdits:     extractPendingEdits(model.Actions),
	}
}

// extractFilesState extracts settings files state from the model
func extractFilesState(model *types.Model) FilesState {
	return FilesState{
		UserExists:  model.UserLevel.Exists,   // Direct field access
		RepoExists:  model.RepoLevel.Exists,   // Direct field access
		LocalExists: model.LocalLevel.Exists,  // Direct field access
		UserPath:    model.UserLevel.Path,     // Direct field access
		RepoPath:    model.RepoLevel.Path,     // Direct field access
		LocalPath:   model.LocalLevel.Path,    // Direct field access
	}
}

// extractPendingMoves extracts pending moves from actions
func extractPendingMoves(actions []types.Action) []string {
	var moves []string
	for _, action := range actions {
		if action.Type == types.ActionMove {  // Direct field access
			moves = append(moves, action.Permission+"→"+action.ToLevel)
		}
	}
	return moves
}

// extractPendingEdits extracts pending edits from actions
func extractPendingEdits(actions []types.Action) []string {
	var edits []string
	for _, action := range actions {
		if action.Type == "edit" {  // Direct field access
			edits = append(edits, action.Permission+"→"+action.NewName)
		}
	}
	return edits
}

// panelNumberToName converts panel number to name
func panelNumberToName(panel int) string {
	switch panel {
	case 0:
		return "permissions"
	case 1:
		return "duplicates"
	case 2:
		return "actions"
	default:
		return "unknown"
	}
}

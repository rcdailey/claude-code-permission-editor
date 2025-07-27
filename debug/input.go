package debug

import (
	"encoding/json"
	"net/http"
	"time"

	"claude-permissions/types"
)

// InputRequest represents the input injection request
type InputRequest struct {
	Key string `json:"key"`
}

// InputResponse represents the response to input injection
type InputResponse struct {
	PreviousPanel string   `json:"previous_panel"`
	NewPanel      string   `json:"new_panel"`
	StateChanges  []string `json:"state_changes"`
	Success       bool     `json:"success"`
	Error         string   `json:"error,omitempty"`
	Timestamp     string   `json:"timestamp"`
}

// ModelStateCapture represents a snapshot of model state before/after input
type ModelStateCapture struct {
	ActivePanel   int      `json:"active_panel"`
	SelectedItems []string `json:"selected_items"`
	FilterText    string   `json:"filter_text"`
	ConfirmMode   bool     `json:"confirm_mode"`
	StatusMessage string   `json:"status_message"`
}

// handleInput handles the POST /input endpoint
func (ds *DebugServer) handleInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the input request
	var request InputRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ds.writeErrorResponse(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate the key
	if request.Key == "" {
		ds.writeErrorResponse(w, "Key field is required", http.StatusBadRequest)
		return
	}

	// Capture state before input
	beforeState := ds.captureModelState()

	// Send the input to the application
	err := ds.SendInput(request.Key)

	// Give the application a moment to process the input
	time.Sleep(50 * time.Millisecond)

	// Capture state after input
	afterState := ds.captureModelState()

	// Build response
	response := InputResponse{
		Success:   err == nil,
		Timestamp: getCurrentTimestamp(),
	}

	if err != nil {
		response.Error = err.Error()
	} else {
		// Analyze state changes
		response.PreviousPanel = panelNumberToName(beforeState.ActivePanel)
		response.NewPanel = panelNumberToName(afterState.ActivePanel)
		response.StateChanges = analyzeStateChanges(beforeState, afterState)
	}

	ds.logger.LogEvent("input_processed", map[string]interface{}{
		"key":           request.Key,
		"success":       response.Success,
		"state_changes": len(response.StateChanges),
		"panel_change":  response.PreviousPanel != response.NewPanel,
	})

	ds.writeJSONResponse(w, response)
}

// handleReset handles the POST /reset endpoint
func (ds *DebugServer) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, this is a placeholder implementation
	// In a full implementation, this would reset the application state
	response := map[string]interface{}{
		"message":   "Reset functionality not yet implemented",
		"timestamp": getCurrentTimestamp(),
	}

	ds.logger.LogEvent("reset_requested", nil)
	ds.writeJSONResponse(w, response)
}

// captureModelState captures a snapshot of the current model state using direct field access
// extractSelectedItemsForCapture extracts currently selected items for input capture
func extractSelectedItemsForCapture(model *types.Model) []string {
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

func (ds *DebugServer) captureModelState() ModelStateCapture {
	model := ds.GetModel()
	if model == nil {
		return ModelStateCapture{}
	}

	model.Mutex.RLock() // Add explicit locking
	defer model.Mutex.RUnlock()

	return ModelStateCapture{
		ActivePanel: model.ActivePanel, // Direct field access
		SelectedItems: extractSelectedItemsForCapture(
			model,
		), // Extract from current column selection
		FilterText:    "",                  // No filter in current UI implementation
		ConfirmMode:   false,               // Removed confirm mode boolean
		StatusMessage: model.StatusMessage, // Direct field access
	}
}

// analyzeStateChanges compares before and after state to identify changes
func analyzeStateChanges(before, after ModelStateCapture) []string {
	changes := []string{}

	changes = append(changes, checkPanelChange(before, after)...)
	changes = append(changes, checkConfirmModeChange(before, after)...)
	changes = append(changes, checkActionsCountChange(before, after)...)
	changes = append(changes, checkFilterTextChange(before, after)...)
	changes = append(changes, checkStatusMessageChange(before, after)...)
	changes = append(changes, checkSelectedItemsChange(before, after)...)

	// If no specific changes detected but states are different, mark as general change
	if len(changes) == 0 && !statesEqual(before, after) {
		changes = append(changes, "state_modified")
	}

	return changes
}

func checkPanelChange(before, after ModelStateCapture) []string {
	if before.ActivePanel != after.ActivePanel {
		return []string{"panel_switched"}
	}
	return nil
}

func checkConfirmModeChange(before, after ModelStateCapture) []string {
	if before.ConfirmMode != after.ConfirmMode {
		if after.ConfirmMode {
			return []string{"confirm_mode_entered"}
		}
		return []string{"confirm_mode_exited"}
	}
	return nil
}

func checkActionsCountChange(before, after ModelStateCapture) []string {
	// Actions removed from system - no longer tracking
	return nil
}

func checkFilterTextChange(before, after ModelStateCapture) []string {
	if before.FilterText != after.FilterText {
		switch {
		case after.FilterText == "" && before.FilterText != "":
			return []string{"filter_cleared"}
		case before.FilterText == "" && after.FilterText != "":
			return []string{"filter_activated"}
		default:
			return []string{"filter_modified"}
		}
	}
	return nil
}

func checkStatusMessageChange(before, after ModelStateCapture) []string {
	if before.StatusMessage != after.StatusMessage {
		return []string{"status_message_changed"}
	}
	return nil
}

func checkSelectedItemsChange(before, after ModelStateCapture) []string {
	if !stringSlicesEqual(before.SelectedItems, after.SelectedItems) {
		return []string{"selection_changed"}
	}
	return nil
}

// statesEqual compares two model state captures for equality
func statesEqual(a, b ModelStateCapture) bool {
	return a.ActivePanel == b.ActivePanel &&
		a.ConfirmMode == b.ConfirmMode &&
		a.FilterText == b.FilterText &&
		a.StatusMessage == b.StatusMessage &&
		stringSlicesEqual(a.SelectedItems, b.SelectedItems)
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

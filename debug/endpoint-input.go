package debug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func init() {
	RegisterEndpoint("/input", handleInput)
}

// InputRequest represents the input injection request
type InputRequest struct {
	Key string `json:"key"`
}

// InputResponse represents the response to input injection
type InputResponse struct {
	PreviousPanel string        `json:"previous_panel"`
	NewPanel      string        `json:"new_panel"`
	StateChanges  []string      `json:"state_changes"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
	Snapshot      *SnapshotData `json:"snapshot,omitempty"`
	Timestamp     string        `json:"timestamp"`
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
func handleInput(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	// Parse the input request
	var request InputRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeErrorResponse(w, "Invalid JSON in request body", http.StatusBadRequest, ds.logger)
		return
	}

	// Validate the key
	if request.Key == "" {
		writeErrorResponse(w, "Key field is required", http.StatusBadRequest, ds.logger)
		return
	}

	// Capture state before input
	beforeState := captureModelState(ds)

	// Send the input to the application
	err := sendInput(ds, request.Key)

	// Give the application a moment to process the input
	time.Sleep(50 * time.Millisecond)

	// Capture state after input
	afterState := captureModelState(ds)

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

		// Capture snapshot after input processing
		if snapshot, snapshotErr := captureSnapshot(ds, true); snapshotErr == nil {
			response.Snapshot = snapshot
		}
	}

	ds.logger.LogEvent("input_processed", map[string]interface{}{
		"key":           request.Key,
		"success":       response.Success,
		"state_changes": len(response.StateChanges),
		"panel_change":  response.PreviousPanel != response.NewPanel,
	})

	writeJSONResponse(w, response, ds.logger)
}

// sendInput sends a key input to the TUI program
func sendInput(ds *DebugServer, key string) error {
	if ds.program == nil {
		return fmt.Errorf("no program instance available")
	}

	msg, err := convertKeyToMessage(key)
	if err != nil {
		return err
	}

	ds.program.Send(msg)
	ds.logger.LogEvent("input_sent", map[string]interface{}{
		"key": key,
	})

	return nil
}

// captureModelState captures a snapshot of the current model state using direct field access
func captureModelState(ds *DebugServer) ModelStateCapture {
	model := ds.GetModel()
	if model == nil {
		return ModelStateCapture{}
	}

	model.Mutex.RLock()
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

// convertKeyToMessage converts a string key to a tea.Msg
func convertKeyToMessage(key string) (tea.Msg, error) {
	switch key {
	case "up", "arrow-up":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}), nil
	case "down", "arrow-down":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}), nil
	case "left", "arrow-left":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}), nil
	case "right", "arrow-right":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}), nil
	case "tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}), nil
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), nil
	case "escape", "esc":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}), nil
	case "space":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeySpace, Text: " "}), nil
	default:
		return convertRuneKeyToMessage(key)
	}
}

// keyMappings maps key strings to their corresponding rune
var keyMappings = map[string]rune{
	"a": 'a', "A": 'a',
	"u": 'u', "U": 'u',
	"r": 'r', "R": 'r',
	"l": 'l', "L": 'l',
	"e": 'e', "E": 'e',
	"c": 'c', "C": 'c',
	"q": 'q', "Q": 'q',
	"y": 'y', "Y": 'y',
	"n": 'n', "N": 'n',
	"/": '/',
	"1": '1',
	"2": '2',
	"3": '3',
}

// convertRuneKeyToMessage converts single character keys to messages
func convertRuneKeyToMessage(key string) (tea.Msg, error) {
	if r, ok := keyMappings[key]; ok {
		return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}), nil
	}
	return nil, fmt.Errorf("unsupported key: %s", key)
}

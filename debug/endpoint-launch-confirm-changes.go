package debug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func init() {
	RegisterEndpoint("/launch-confirm-changes", handleLaunchConfirmChanges)
}

// LaunchConfirmChangesRequest represents the request to launch confirm changes screen
type LaunchConfirmChangesRequest struct {
	MockChanges struct {
		PermissionMoves []struct {
			Name string `json:"name"`
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"permission_moves"`
		DuplicateResolutions []struct {
			Name       string   `json:"name"`
			KeepLevel  string   `json:"keep_level"`
			RemoveFrom []string `json:"remove_from"`
		} `json:"duplicate_resolutions"`
	} `json:"mock_changes"`
}

// LaunchConfirmChangesResponse represents the response from launching confirm changes screen
type LaunchConfirmChangesResponse struct {
	Success        bool          `json:"success"`
	PreviousScreen string        `json:"previous_screen"`
	NewScreen      string        `json:"new_screen"`
	ChangesApplied int           `json:"changes_applied"`
	Error          string        `json:"error,omitempty"`
	Snapshot       *SnapshotData `json:"snapshot,omitempty"`
	Timestamp      string        `json:"timestamp"`
}

// LaunchConfirmChangesMsg is a custom message for launching confirm changes screen
type LaunchConfirmChangesMsg struct {
	Request *LaunchConfirmChangesRequest
}

// handleLaunchConfirmChanges handles the POST /launch-confirm-changes endpoint
func handleLaunchConfirmChanges(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	request, err := parseLaunchRequest(r)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest, ds.logger)
		return
	}

	if ds.program == nil {
		writeErrorResponse(
			w,
			"No program instance available",
			http.StatusInternalServerError,
			ds.logger,
		)
		return
	}

	response, err := processLaunchRequest(ds, request)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError, ds.logger)
		return
	}

	ds.logger.LogEvent("launch_confirm_changes", map[string]interface{}{
		"changes_applied": response.ChangesApplied,
		"previous_screen": response.PreviousScreen,
		"new_screen":      response.NewScreen,
	})

	writeJSONResponse(w, response, ds.logger)
}

func parseLaunchRequest(r *http.Request) (*LaunchConfirmChangesRequest, error) {
	var request LaunchConfirmChangesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, fmt.Errorf("invalid JSON in request body")
	}
	return &request, nil
}

func processLaunchRequest(
	ds *DebugServer,
	request *LaunchConfirmChangesRequest,
) (*LaunchConfirmChangesResponse, error) {
	model := ds.GetModel()
	if model == nil {
		return nil, fmt.Errorf("no model available")
	}

	// Capture previous screen
	model.Mutex.RLock()
	previousScreen := screenNumberToName(getCurrentScreen(model))
	model.Mutex.RUnlock()

	// Send message to launch confirm changes screen
	msg := LaunchConfirmChangesMsg{Request: request}
	ds.program.Send(msg)

	// Give the application a moment to process the message
	time.Sleep(100 * time.Millisecond)

	// Capture new screen state
	model.Mutex.RLock()
	newScreen := screenNumberToName(getCurrentScreen(model))
	model.Mutex.RUnlock()

	response := &LaunchConfirmChangesResponse{
		Success:        true,
		PreviousScreen: previousScreen,
		NewScreen:      newScreen,
		ChangesApplied: len(
			request.MockChanges.PermissionMoves,
		) + len(
			request.MockChanges.DuplicateResolutions,
		),
		Timestamp: getCurrentTimestamp(),
	}

	// Capture snapshot after launching screen
	if snapshot, err := captureSnapshot(ds, true); err == nil {
		response.Snapshot = snapshot
	}

	return response, nil
}

// getCurrentScreen extracts the current screen value without importing types
func getCurrentScreen(model interface{}) int {
	// Use reflection or type assertion - for now just return 0 as placeholder
	// In a real implementation, this would need access to the CurrentScreen field
	return 0
}

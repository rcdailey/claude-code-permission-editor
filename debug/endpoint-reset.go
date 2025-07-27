package debug

import (
	"net/http"
)

func init() {
	RegisterEndpoint("/reset", handleReset)
}

// handleReset handles the POST /reset endpoint
func handleReset(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	// For now, this is a placeholder implementation
	// In a full implementation, this would reset the application state
	response := map[string]interface{}{
		"message":   "Reset functionality not yet implemented",
		"timestamp": getCurrentTimestamp(),
	}

	ds.logger.LogEvent("reset_requested", nil)
	writeJSONResponse(w, response, ds.logger)
}

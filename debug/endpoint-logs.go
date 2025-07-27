package debug

import (
	"net/http"
)

func init() {
	RegisterEndpoint("/logs", handleLogs)
}

// LogResponse represents the logs endpoint response
type LogResponse struct {
	Entries []LogEntry `json:"entries"`
}

// handleLogs handles the GET /logs endpoint
func handleLogs(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	// Get all entries and clear the buffer
	entries := ds.logger.GetAndClearEntries()

	response := LogResponse{
		Entries: entries,
	}

	writeJSONResponse(w, response, ds.logger)
}

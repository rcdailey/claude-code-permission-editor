package debug

import (
	"net/http"
)

func init() {
	RegisterEndpoint("/snapshot", handleSnapshot)
}

// handleSnapshot handles the GET /snapshot endpoint
func handleSnapshot(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
		return
	}

	// Get query parameters - color is opt-in, raw is default
	color := getQueryParamBool(r, "color", false)
	raw := !color

	// Capture snapshot using shared function
	snapshot, err := captureSnapshot(ds, raw)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusInternalServerError, ds.logger)
		return
	}

	ds.logger.LogEvent("snapshot_captured", map[string]interface{}{
		"width":  snapshot.Width,
		"height": snapshot.Height,
		"raw":    raw,
		"color":  color,
	})

	writeJSONResponse(w, snapshot, ds.logger)
}

package debug

import (
	"encoding/json"
	"net/http"
	"time"
)

func init() {
	RegisterEndpoint("/health", handleHealth)
}

// handleHealth provides a health check endpoint
func handleHealth(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		ds.logger.LogError("health_endpoint_error", err, nil)
	}
}

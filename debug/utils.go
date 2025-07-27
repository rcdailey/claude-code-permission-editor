package debug

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"claude-permissions/types"
)

// writeJSONResponse writes a JSON response with proper headers
func writeJSONResponse(w http.ResponseWriter, data interface{}, logger *Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		if logger != nil {
			logger.LogEvent("json_encode_error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
}

// writeErrorResponse writes a structured error response
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int, logger *Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		if logger != nil {
			logger.LogError("error_response_encode_failed", err, nil)
		}
	}
	if logger != nil {
		logger.LogEvent("error_response", map[string]interface{}{
			"message":     message,
			"status_code": statusCode,
		})
	}
}

// getQueryParamBool safely gets a boolean query parameter with a default value
func getQueryParamBool(r *http.Request, key string, defaultValue bool) bool {
	if value := r.URL.Query().Get(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getCurrentTimestamp returns the current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// getTimestamp returns the current timestamp in RFC3339 format
func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
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

// screenNumberToName converts screen number to name
func screenNumberToName(screen int) string {
	switch screen {
	case types.ScreenDuplicates:
		return "ScreenDuplicates"
	case types.ScreenOrganization:
		return "ScreenOrganization"
	default:
		return "Unknown"
	}
}

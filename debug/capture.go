package debug

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/term"
)

// SnapshotResponse represents the screen snapshot data
type SnapshotResponse struct {
	Content        string    `json:"content"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
	CursorPosition [2]int    `json:"cursor_position"`
	Raw            bool      `json:"raw"`
	Timestamp      string    `json:"timestamp"`
}

// No interfaces needed anymore - we use concrete types.Model

// handleSnapshot handles the GET /snapshot endpoint
func (ds *DebugServer) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	raw := getQueryParamBool(r, "raw", false)

	model := ds.GetModel()
	if model == nil {
		ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
		return
	}

	// Get the terminal dimensions
	width, height := getTerminalDimensions()

	// Get the actual rendered view content using the ViewProvider
	var content string

	if ds.viewProvider != nil {
		content = ds.viewProvider.GetView()
	} else {
		// Fallback: try to get some meaningful content from the model state
		model.Mutex.RLock()
		content = fmt.Sprintf("Permissions: %d, Duplicates: %d, Actions: %d",
			len(model.Permissions), len(model.Duplicates), len(model.Actions))
		model.Mutex.RUnlock()
	}

	// Strip ANSI codes if raw output is requested
	if raw {
		content = stripANSICodes(content)
	}

	// For now, we'll assume cursor is at the end of visible content
	// In a real implementation, this would need to track actual cursor position
	cursorPos := estimateCursorPosition(content)

	response := SnapshotResponse{
		Content:        content,
		Width:          width,
		Height:         height,
		CursorPosition: cursorPos,
		Raw:            raw,
		Timestamp:      getCurrentTimestamp(),
	}

	ds.logger.LogEvent("snapshot_captured", map[string]interface{}{
		"width":  width,
		"height": height,
		"raw":    raw,
	})

	ds.writeJSONResponse(w, response)
}

// getTerminalDimensions returns the current terminal dimensions
func getTerminalDimensions() (width, height int) {
	// Try to get actual terminal size from stdout
	if w, h, err := term.GetSize(0); err == nil {
		return w, h
	}

	// Try stderr if stdout fails
	if w, h, err := term.GetSize(2); err == nil {
		return w, h
	}

	// Fallback to reasonable defaults
	return 80, 24
}

// stripANSICodes removes ANSI escape sequences from text
func stripANSICodes(text string) string {
	// ANSI escape sequence regex pattern
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiEscape.ReplaceAllString(text, "")
}

// estimateCursorPosition attempts to estimate cursor position based on content
func estimateCursorPosition(content string) [2]int {
	lines := strings.Split(content, "\n")

	// Find the last non-empty line for Y position
	y := len(lines) - 1
	for y >= 0 && strings.TrimSpace(lines[y]) == "" {
		y--
	}

	// Use the length of the last non-empty line for X position
	x := 0
	if y >= 0 && y < len(lines) {
		// Strip ANSI codes to get actual text length
		cleanLine := stripANSICodes(lines[y])
		x = len(cleanLine)
	}

	return [2]int{x, y}
}

// getCurrentTimestamp returns the current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return getTimestamp()
}

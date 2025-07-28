package debug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"claude-permissions/types"

	"golang.org/x/term"
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

// ComponentPosition represents the calculated position and dimensions of a UI component
type ComponentPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// LayoutCalculations represents UI layout metrics and component sizing data
type LayoutCalculations struct {
	AvailableHeight int                    `json:"available_height"`
	FixedHeight     int                    `json:"fixed_height"`
	FrameOverhead   map[string]int         `json:"frame_overhead"`
	ComponentSizes  map[string]interface{} `json:"component_sizes"`
}

// SnapshotData represents the combined screen snapshot and layout data
type SnapshotData struct {
	// Rendered content
	Content        string `json:"content"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	CursorPosition [2]int `json:"cursor_position"`
	Raw            bool   `json:"raw"`

	// Layout diagnostics
	Terminal           [2]int                       `json:"terminal"`
	Components         map[string]ComponentPosition `json:"components"`
	LayoutWarnings     []string                     `json:"layout_warnings"`
	LayoutCalculations LayoutCalculations           `json:"layout_calculations"`

	// Dimension validation
	DimensionMismatch bool   `json:"dimension_mismatch"`
	MismatchDetails   string `json:"mismatch_details,omitempty"`

	Timestamp string `json:"timestamp"`
}

// LayoutResponse represents layout diagnostics data for compatibility
type LayoutResponse struct {
	Terminal     [2]int                       `json:"terminal"`
	Components   map[string]ComponentPosition `json:"components"`
	Warnings     []string                     `json:"warnings"`
	Calculations LayoutCalculations           `json:"calculations"`
}

// captureSnapshot captures current application snapshot data
func captureSnapshot(ds *DebugServer, raw bool) (*SnapshotData, error) {
	model := ds.GetModel()
	if model == nil {
		return nil, fmt.Errorf("model not available")
	}

	width, height := getTerminalDimensions()
	content := getViewContent(ds, model)

	if raw {
		content = stripANSICodes(content)
	}

	cursorPos := estimateCursorPosition(content)
	layoutData := extractLayoutDiagnostics(model)
	renderedWidth, renderedHeight := calculateContentDimensions(content)
	dimensionMismatch, mismatchDetails := checkDimensionMismatch(
		width, height, renderedWidth, renderedHeight, model)

	return &SnapshotData{
		Content:        content,
		Width:          width,
		Height:         height,
		CursorPosition: cursorPos,
		Raw:            raw,

		Terminal:           layoutData.Terminal,
		Components:         layoutData.Components,
		LayoutWarnings:     layoutData.Warnings,
		LayoutCalculations: layoutData.Calculations,

		DimensionMismatch: dimensionMismatch,
		MismatchDetails:   mismatchDetails,

		Timestamp: getCurrentTimestamp(),
	}, nil
}

// getViewContent gets the rendered view content from ViewProvider or model fallback
func getViewContent(ds *DebugServer, model *types.Model) string {
	if ds.viewProvider != nil {
		return ds.viewProvider.GetView()
	}

	model.Mutex.RLock()
	defer model.Mutex.RUnlock()
	return fmt.Sprintf("Permissions: %d, Duplicates: %d",
		len(model.Permissions), len(model.Duplicates))
}

// calculateContentDimensions calculates rendered content width and height
func calculateContentDimensions(content string) (width, height int) {
	contentLines := strings.Split(content, "\n")
	height = len(contentLines)
	width = 0
	for _, line := range contentLines {
		lineWidth := visualWidth(line)
		if lineWidth > width {
			width = lineWidth
		}
	}
	return width, height
}

// checkDimensionMismatch checks for mismatches between terminal and rendered dimensions
func checkDimensionMismatch(
	termWidth, termHeight, renderedWidth, renderedHeight int,
	model *types.Model,
) (bool, string) {
	if renderedWidth != termWidth || renderedHeight != termHeight {
		return true, fmt.Sprintf("Terminal: %dx%d, Rendered: %dx%d, Model: %dx%d",
			termWidth, termHeight, renderedWidth, renderedHeight, model.Width, model.Height)
	}
	return false, ""
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

// visualWidth calculates the visual width of a string, accounting for ANSI codes
func visualWidth(s string) int {
	// Strip ANSI codes first
	cleaned := stripANSICodes(s)
	// Return the rune count (not byte count) for proper Unicode support
	return utf8.RuneCountInString(cleaned)
}

// extractLayoutDiagnostics creates layout diagnostics for the pure lipgloss architecture
func extractLayoutDiagnostics(model *types.Model) *LayoutResponse {
	model.Mutex.RLock()
	defer model.Mutex.RUnlock()

	response := &LayoutResponse{
		Terminal:   [2]int{model.Width, model.Height},
		Components: make(map[string]ComponentPosition),
		Warnings:   []string{"pure_lipgloss_architecture"},
		Calculations: LayoutCalculations{
			FrameOverhead:  make(map[string]int),
			ComponentSizes: make(map[string]interface{}),
		},
	}

	// Create simplified component positions for pure lipgloss layout
	headerHeight := 3
	footerHeight := 1
	contentHeight := model.Height - headerHeight - footerHeight

	response.Components["header"] = ComponentPosition{
		X: 0, Y: 0, W: model.Width, H: headerHeight,
	}
	response.Components["content"] = ComponentPosition{
		X: 0, Y: headerHeight, W: model.Width, H: contentHeight,
	}
	response.Components["footer"] = ComponentPosition{
		X: 0, Y: headerHeight + contentHeight, W: model.Width, H: footerHeight,
	}

	response.Calculations = LayoutCalculations{
		AvailableHeight: contentHeight,
		FixedHeight:     headerHeight + footerHeight,
		FrameOverhead: map[string]int{
			"height": 2,
			"width":  4,
		},
		ComponentSizes: map[string]interface{}{
			"content_area": map[string]int{
				"width":  model.Width - 0,  // ContentWidthBuffer from ui/components.go
				"height": model.Height - 8, // Approximate content height accounting for header/footer/status
			},
			"duplicates_table": map[string]int{
				"width":  model.DuplicatesTable.Width(),
				"height": model.DuplicatesTable.Height(),
			},
		},
	}

	return response
}

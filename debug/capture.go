package debug

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"claude-permissions/types"

	"golang.org/x/term"
)

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

// SnapshotResponse represents the combined screen snapshot and layout data
type SnapshotResponse struct {
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

// No interfaces needed anymore - we use concrete types.Model

// handleSnapshot handles the GET /snapshot endpoint
//
//nolint:funlen // Debug function complexity is acceptable for testing/debugging purposes
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

	// Get layout diagnostics
	layoutData := extractLayoutDiagnostics(model)

	// Calculate rendered content dimensions
	contentLines := strings.Split(content, "\n")
	renderedHeight := len(contentLines)
	renderedWidth := 0
	for _, line := range contentLines {
		// Always use visual width (accounting for ANSI codes and Unicode)
		lineWidth := visualWidth(line)
		if lineWidth > renderedWidth {
			renderedWidth = lineWidth
		}
	}

	// Check for dimension mismatches
	var dimensionMismatch bool
	var mismatchDetails string

	if renderedWidth != width || renderedHeight != height {
		dimensionMismatch = true
		mismatchDetails = fmt.Sprintf("Terminal: %dx%d, Rendered: %dx%d, Model: %dx%d",
			width, height, renderedWidth, renderedHeight, model.Width, model.Height)
	}

	response := SnapshotResponse{
		// Rendered content
		Content:        content,
		Width:          width,
		Height:         height,
		CursorPosition: cursorPos,
		Raw:            raw,

		// Layout diagnostics
		Terminal:           layoutData.Terminal,
		Components:         layoutData.Components,
		LayoutWarnings:     layoutData.Warnings,
		LayoutCalculations: layoutData.Calculations,

		// Dimension validation
		DimensionMismatch: dimensionMismatch,
		MismatchDetails:   mismatchDetails,

		Timestamp: getCurrentTimestamp(),
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
			"permissions_list": map[string]int{
				"width":  model.PermissionsList.Width(),
				"height": model.PermissionsList.Height(),
			},
			"duplicates_table": map[string]int{
				"width":  model.DuplicatesTable.Width(),
				"height": model.DuplicatesTable.Height(),
			},
			"actions_view": map[string]int{
				"width":  model.ActionsView.Width,
				"height": model.ActionsView.Height,
			},
		},
	}

	return response
}

// LayoutResponse represents layout diagnostics data for compatibility
type LayoutResponse struct {
	Terminal     [2]int                       `json:"terminal"`
	Components   map[string]ComponentPosition `json:"components"`
	Warnings     []string                     `json:"warnings"`
	Calculations LayoutCalculations           `json:"calculations"`
}

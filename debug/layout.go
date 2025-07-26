package debug

import (
	"net/http"

	"claude-permissions/layout"
	"claude-permissions/types"
)

// LayoutResponse represents the layout diagnostics data
type LayoutResponse struct {
	Terminal     [2]int                       `json:"terminal"`
	Components   map[string]ComponentPosition `json:"components"`
	Warnings     []string                     `json:"warnings"`
	Calculations LayoutCalculations           `json:"calculations"`
	Timestamp    string                       `json:"timestamp"`
}

// ComponentPosition represents the position and dimensions of a component
type ComponentPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// LayoutCalculations represents layout calculation details
type LayoutCalculations struct {
	AvailableHeight int                    `json:"available_height"`
	FixedHeight     int                    `json:"fixed_height"`
	FrameOverhead   map[string]int         `json:"frame_overhead"`
	ComponentSizes  map[string]interface{} `json:"component_sizes"`
}

// handleLayout handles the GET /layout endpoint
func (ds *DebugServer) handleLayout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	model := ds.GetModel()
	if model == nil {
		ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
		return
	}

	response := extractLayoutDiagnostics(model)
	response.Timestamp = getCurrentTimestamp()

	ds.logger.LogEvent("layout_extracted", map[string]interface{}{
		"terminal_width":   response.Terminal[0],
		"terminal_height":  response.Terminal[1],
		"components_count": len(response.Components),
		"warnings_count":   len(response.Warnings),
	})

	ds.writeJSONResponse(w, response)
}

// extractLayoutDiagnostics extracts layout information from the model using direct field access
func extractLayoutDiagnostics(model *types.Model) LayoutResponse {
	model.Mutex.RLock()
	defer model.Mutex.RUnlock()

	response := LayoutResponse{
		Terminal:     [2]int{80, 24}, // Default values
		Components:   make(map[string]ComponentPosition),
		Warnings:     []string{},
		Calculations: LayoutCalculations{
			FrameOverhead:  make(map[string]int),
			ComponentSizes: make(map[string]interface{}),
		},
	}

	if model.LayoutEngine == nil { // Direct field access
		response.Warnings = append(response.Warnings, "layout_engine_not_found")
		return response
	}

	// Direct method calls instead of reflection
	width, height := model.LayoutEngine.GetTerminalSize()
	response.Terminal = [2]int{width, height}

	if result := model.LayoutEngine.GetLastResult(); result != nil {
		response.Components = extractComponentsFromResult(result)
		response.Warnings = result.Warnings
	}

	response.Calculations = extractLayoutCalculations(model, height)
	return response
}

// extractComponentsFromResult extracts components from a layout result using direct access
func extractComponentsFromResult(result *layout.LayoutResult) map[string]ComponentPosition {
	components := make(map[string]ComponentPosition)

	for id, layout := range result.Components {
		components[id] = ComponentPosition{
			X: layout.X,
			Y: layout.Y,
			W: layout.Width,
			H: layout.Height,
		}
	}

	return components
}

// extractLayoutCalculations extracts calculation details using direct field access
func extractLayoutCalculations(model *types.Model, terminalHeight int) LayoutCalculations {
	calc := LayoutCalculations{
		FrameOverhead:  make(map[string]int),
		ComponentSizes: make(map[string]interface{}),
	}

	calc.AvailableHeight, calc.FixedHeight = calculateHeightDistribution(terminalHeight)

	// Extract component sizes using direct method calls
	calc.FrameOverhead = map[string]int{
		"width":  4, // Default frame width overhead
		"height": 2, // Default frame height overhead
	}

	calc.ComponentSizes = map[string]interface{}{
		"permissions_list": map[string]int{
			"width":  model.PermissionsList.Width(),  // Direct method call
			"height": model.PermissionsList.Height(), // Direct method call
		},
		"duplicates_table": map[string]int{
			"width":  model.DuplicatesTable.Width(),  // Direct method call
			"height": model.DuplicatesTable.Height(), // Direct method call
		},
		"actions_view": map[string]int{
			"width":  model.ActionsView.Width,  // Direct field access
			"height": model.ActionsView.Height, // Direct field access
		},
	}

	return calc
}

// calculateHeightDistribution calculates how height is distributed among components
func calculateHeightDistribution(terminalHeight int) (int, int) {
	// This is a simplified calculation
	// In a real implementation, this would analyze the actual component constraints
	fixedHeight := 8 // Estimated fixed height (header + footer)
	availableHeight := terminalHeight - fixedHeight

	if availableHeight < 0 {
		availableHeight = 0
	}

	return availableHeight, fixedHeight
}

package layout

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// DebugMode represents different debug modes
type DebugMode int

const (
	statusFailed = "✗"
)

const (
	// DebugModeOff disables debug output
	DebugModeOff DebugMode = iota
	// DebugModeBasic enables basic debug output
	DebugModeBasic
	// DebugModeVerbose enables verbose debug output
	DebugModeVerbose
	// DebugModeDetailed enables detailed debug output
	DebugModeDetailed
)

// DebugInfo contains comprehensive debug information about the layout system
type DebugInfo struct {
	Timestamp       time.Time
	EngineInfo      EngineDebugInfo
	ComponentsInfo  []ComponentDebugInfo
	LayoutResult    *LayoutResult
	ConstraintsInfo []ConstraintDebugInfo
	WarningsInfo    []WarningInfo
	PerformanceInfo PerformanceDebugInfo
}

// EngineDebugInfo contains debug information about the layout engine
type EngineDebugInfo struct {
	TerminalWidth       int
	TerminalHeight      int
	ComponentCount      int
	NeedsRecalculation  bool
	Config              EngineConfig
	LastCalculationTime time.Time
}

// ComponentDebugInfo contains debug information about a component
type ComponentDebugInfo struct {
	ID              string
	Type            string
	Width           int
	Height          int
	Position        Position
	Constraints     []string
	IsDirty         bool
	HasValidLayout  bool
	OverflowsX      bool
	OverflowsY      bool
	ConstraintCount int
}

// ConstraintDebugInfo contains debug information about constraints
type ConstraintDebugInfo struct {
	ComponentID    string
	ConstraintType string
	Value          string
	IsValid        bool
	ErrorMessage   string
}

// WarningInfo contains information about layout warnings
type WarningInfo struct {
	Type        string
	Message     string
	ComponentID string
	Severity    WarningSeverity
}

// WarningSeverity represents the severity of a warning
type WarningSeverity int

const (
	// SeverityInfo represents informational warnings
	SeverityInfo WarningSeverity = iota
	// SeverityWarning represents warning-level messages
	SeverityWarning
	// SeverityError represents error-level messages
	SeverityError
	// SeverityCritical represents critical error messages
	SeverityCritical
)

// PerformanceDebugInfo contains performance-related debug information
type PerformanceDebugInfo struct {
	LastCalculationDuration time.Duration
	AverageCalculationTime  time.Duration
	TotalCalculations       int
	CacheHitRate            float64
	MemoryUsage             int64
}

// LayoutDebugger provides debugging and validation utilities for the layout system
type LayoutDebugger struct {
	engine          *LayoutEngine
	mode            DebugMode
	history         []DebugInfo
	maxHistory      int
	enableProfiling bool
}

// NewLayoutDebugger creates a new layout debugger
func NewLayoutDebugger(engine *LayoutEngine, mode DebugMode) *LayoutDebugger {
	return &LayoutDebugger{
		engine:          engine,
		mode:            mode,
		history:         []DebugInfo{},
		maxHistory:      100,
		enableProfiling: false,
	}
}

// SetMode sets the debug mode
func (ld *LayoutDebugger) SetMode(mode DebugMode) {
	ld.mode = mode
}

// EnableProfiling enables performance profiling
func (ld *LayoutDebugger) EnableProfiling(enable bool) {
	ld.enableProfiling = enable
}

// CaptureDebugInfo captures comprehensive debug information
func (ld *LayoutDebugger) CaptureDebugInfo() DebugInfo {
	info := DebugInfo{
		Timestamp:       time.Now(),
		EngineInfo:      ld.captureEngineInfo(),
		ComponentsInfo:  ld.captureComponentsInfo(),
		LayoutResult:    ld.engine.GetLastCalculation(),
		ConstraintsInfo: ld.captureConstraintsInfo(),
		WarningsInfo:    ld.captureWarningsInfo(),
		PerformanceInfo: ld.capturePerformanceInfo(),
	}

	// Add to history
	ld.addToHistory(info)

	return info
}

// captureEngineInfo captures debug information about the engine
func (ld *LayoutDebugger) captureEngineInfo() EngineDebugInfo {
	width, height := ld.engine.GetTerminalDimensions()

	return EngineDebugInfo{
		TerminalWidth:       width,
		TerminalHeight:      height,
		ComponentCount:      ld.engine.GetComponentCount(),
		NeedsRecalculation:  ld.engine.NeedsRecalculation(),
		Config:              ld.engine.GetConfig(),
		LastCalculationTime: time.Now(), // Placeholder
	}
}

// captureComponentsInfo captures debug information about all components
func (ld *LayoutDebugger) captureComponentsInfo() []ComponentDebugInfo {
	componentsInfo := make([]ComponentDebugInfo, 0, len(ld.engine.components.GetAll()))

	for _, wrapper := range ld.engine.components.GetAll() {
		info := ComponentDebugInfo{
			ID:              wrapper.ID(),
			Type:            ld.getComponentType(wrapper),
			Width:           wrapper.width,
			Height:          wrapper.height,
			Position:        wrapper.position,
			Constraints:     ld.getConstraintStrings(wrapper.Constraints()),
			IsDirty:         wrapper.IsDirty(),
			HasValidLayout:  ld.validateComponentLayout(wrapper),
			OverflowsX:      ld.checkXOverflow(wrapper),
			OverflowsY:      ld.checkYOverflow(wrapper),
			ConstraintCount: len(wrapper.Constraints().All()),
		}

		componentsInfo = append(componentsInfo, info)
	}

	return componentsInfo
}

// captureConstraintsInfo captures debug information about constraints
func (ld *LayoutDebugger) captureConstraintsInfo() []ConstraintDebugInfo {
	var constraintsInfo []ConstraintDebugInfo

	for _, wrapper := range ld.engine.components.GetAll() {
		for _, constraint := range wrapper.Constraints().All() {
			info := ConstraintDebugInfo{
				ComponentID:    wrapper.ID(),
				ConstraintType: string(constraint.Type()),
				Value:          constraint.String(),
				IsValid:        constraint.Validate() == nil,
				ErrorMessage:   ld.getConstraintError(constraint),
			}

			constraintsInfo = append(constraintsInfo, info)
		}
	}

	return constraintsInfo
}

// captureWarningsInfo captures warning information
func (ld *LayoutDebugger) captureWarningsInfo() []WarningInfo {
	var warningsInfo []WarningInfo

	if result := ld.engine.GetLastCalculation(); result != nil {
		for _, warning := range result.Warnings {
			info := WarningInfo{
				Type:        "layout",
				Message:     warning,
				ComponentID: "unknown",
				Severity:    SeverityWarning,
			}
			warningsInfo = append(warningsInfo, info)
		}
	}

	// Add constraint validation warnings
	for _, wrapper := range ld.engine.components.GetAll() {
		if err := wrapper.Constraints().Validate(); err != nil {
			info := WarningInfo{
				Type:        "constraint",
				Message:     err.Error(),
				ComponentID: wrapper.ID(),
				Severity:    SeverityError,
			}
			warningsInfo = append(warningsInfo, info)
		}
	}

	return warningsInfo
}

// capturePerformanceInfo captures performance information
func (ld *LayoutDebugger) capturePerformanceInfo() PerformanceDebugInfo {
	return PerformanceDebugInfo{
		LastCalculationDuration: 0,   // Placeholder
		AverageCalculationTime:  0,   // Placeholder
		TotalCalculations:       0,   // Placeholder
		CacheHitRate:            0.0, // Placeholder
		MemoryUsage:             0,   // Placeholder
	}
}

// Helper functions for debug info capture

// getComponentType returns the type of a component
func (ld *LayoutDebugger) getComponentType(wrapper *ComponentWrapper) string {
	switch wrapper.GetComponent().(type) {
	case *SimpleComponent:
		return "SimpleComponent"
	default:
		return fmt.Sprintf("%T", wrapper.GetComponent())
	}
}

// getConstraintStrings returns string representations of constraints
func (ld *LayoutDebugger) getConstraintStrings(constraints ConstraintSet) []string {
	strings := make([]string, 0, len(constraints.All()))
	for _, constraint := range constraints.All() {
		strings = append(strings, constraint.String())
	}
	return strings
}

// validateComponentLayout validates a component's layout
func (ld *LayoutDebugger) validateComponentLayout(wrapper *ComponentWrapper) bool {
	if wrapper.width <= 0 || wrapper.height <= 0 {
		return false
	}

	if wrapper.position.X < 0 || wrapper.position.Y < 0 {
		return false
	}

	return true
}

// checkXOverflow checks if a component overflows horizontally
func (ld *LayoutDebugger) checkXOverflow(wrapper *ComponentWrapper) bool {
	terminalWidth, _ := ld.engine.GetTerminalDimensions()
	return wrapper.position.X+wrapper.width > terminalWidth
}

// checkYOverflow checks if a component overflows vertically
func (ld *LayoutDebugger) checkYOverflow(wrapper *ComponentWrapper) bool {
	_, terminalHeight := ld.engine.GetTerminalDimensions()
	return wrapper.position.Y+wrapper.height > terminalHeight
}

// getConstraintError returns the error message for a constraint
func (ld *LayoutDebugger) getConstraintError(constraint Constraint) string {
	if err := constraint.Validate(); err != nil {
		return err.Error()
	}
	return ""
}

// addToHistory adds debug info to the history
func (ld *LayoutDebugger) addToHistory(info DebugInfo) {
	ld.history = append(ld.history, info)

	// Keep only the last maxHistory entries
	if len(ld.history) > ld.maxHistory {
		ld.history = ld.history[len(ld.history)-ld.maxHistory:]
	}
}

// Formatting and output functions

// FormatDebugInfo formats debug information for display
func (ld *LayoutDebugger) FormatDebugInfo(info DebugInfo) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("=== Layout Debug Info (%s) ===\n",
		info.Timestamp.Format("15:04:05.000")))

	// Engine information
	output.WriteString(ld.formatEngineInfo(info.EngineInfo))

	// Components information
	if ld.mode >= DebugModeBasic {
		output.WriteString(ld.formatComponentsInfo(info.ComponentsInfo))
	}

	// Constraints information
	if ld.mode >= DebugModeVerbose {
		output.WriteString(ld.formatConstraintsInfo(info.ConstraintsInfo))
	}

	// Warnings information
	if len(info.WarningsInfo) > 0 {
		output.WriteString(ld.formatWarningsInfo(info.WarningsInfo))
	}

	// Performance information
	if ld.mode >= DebugModeDetailed {
		output.WriteString(ld.formatPerformanceInfo(info.PerformanceInfo))
	}

	return output.String()
}

// formatEngineInfo formats engine debug information
func (ld *LayoutDebugger) formatEngineInfo(info EngineDebugInfo) string {
	var output strings.Builder

	output.WriteString("Engine Information:\n")
	output.WriteString(fmt.Sprintf("  Terminal: %dx%d\n", info.TerminalWidth, info.TerminalHeight))
	output.WriteString(fmt.Sprintf("  Components: %d\n", info.ComponentCount))
	output.WriteString(fmt.Sprintf("  Needs Recalculation: %t\n", info.NeedsRecalculation))
	output.WriteString(fmt.Sprintf("  Auto Recalculate: %t\n", info.Config.AutoRecalculate))
	output.WriteString("\n")

	return output.String()
}

// formatComponentsInfo formats components debug information
func (ld *LayoutDebugger) formatComponentsInfo(info []ComponentDebugInfo) string {
	var output strings.Builder

	output.WriteString("Components Information:\n")

	// Sort by ID for consistent output
	sort.Slice(info, func(i, j int) bool {
		return info[i].ID < info[j].ID
	})

	for _, comp := range info {
		status := "✓"
		if !comp.HasValidLayout {
			status = statusFailed
		}
		if comp.IsDirty {
			status += " (dirty)"
		}
		if comp.OverflowsX || comp.OverflowsY {
			status += " (overflow)"
		}

		output.WriteString(fmt.Sprintf("  %s %s: %dx%d at (%d,%d) [%d constraints]\n",
			status, comp.ID, comp.Width, comp.Height, comp.Position.X, comp.Position.Y, comp.ConstraintCount))

		if ld.mode >= DebugModeVerbose {
			for _, constraint := range comp.Constraints {
				output.WriteString(fmt.Sprintf("    - %s\n", constraint))
			}
		}
	}

	output.WriteString("\n")
	return output.String()
}

// formatConstraintsInfo formats constraints debug information
func (ld *LayoutDebugger) formatConstraintsInfo(info []ConstraintDebugInfo) string {
	var output strings.Builder

	output.WriteString("Constraints Information:\n")

	// Group by component
	componentConstraints := make(map[string][]ConstraintDebugInfo)
	for _, constraint := range info {
		componentConstraints[constraint.ComponentID] = append(
			componentConstraints[constraint.ComponentID], constraint)
	}

	for componentID, constraints := range componentConstraints {
		output.WriteString(fmt.Sprintf("  %s:\n", componentID))

		for _, constraint := range constraints {
			status := "✓"
			if !constraint.IsValid {
				status = statusFailed
			}

			output.WriteString(fmt.Sprintf("    %s %s: %s\n",
				status, constraint.ConstraintType, constraint.Value))

			if constraint.ErrorMessage != "" {
				output.WriteString(fmt.Sprintf("      Error: %s\n", constraint.ErrorMessage))
			}
		}
	}

	output.WriteString("\n")
	return output.String()
}

// formatWarningsInfo formats warnings debug information
func (ld *LayoutDebugger) formatWarningsInfo(info []WarningInfo) string {
	var output strings.Builder

	output.WriteString("Warnings:\n")

	for _, warning := range info {
		severityIcon := ld.getSeverityIcon(warning.Severity)
		output.WriteString(fmt.Sprintf("  %s [%s] %s: %s\n",
			severityIcon, warning.Type, warning.ComponentID, warning.Message))
	}

	output.WriteString("\n")
	return output.String()
}

// formatPerformanceInfo formats performance debug information
func (ld *LayoutDebugger) formatPerformanceInfo(info PerformanceDebugInfo) string {
	var output strings.Builder

	output.WriteString("Performance Information:\n")
	output.WriteString(fmt.Sprintf("  Last Calculation: %v\n", info.LastCalculationDuration))
	output.WriteString(fmt.Sprintf("  Average Calculation: %v\n", info.AverageCalculationTime))
	output.WriteString(fmt.Sprintf("  Total Calculations: %d\n", info.TotalCalculations))
	output.WriteString(fmt.Sprintf("  Cache Hit Rate: %.2f%%\n", info.CacheHitRate*100))
	output.WriteString(fmt.Sprintf("  Memory Usage: %d bytes\n", info.MemoryUsage))
	output.WriteString("\n")

	return output.String()
}

// getSeverityIcon returns an icon for the warning severity
func (ld *LayoutDebugger) getSeverityIcon(severity WarningSeverity) string {
	switch severity {
	case SeverityInfo:
		return "ℹ"
	case SeverityWarning:
		return "⚠"
	case SeverityError:
		return statusFailed
	case SeverityCritical:
		return "❌"
	default:
		return "?"
	}
}

// Analysis functions

// AnalyzeLayout performs comprehensive layout analysis
func (ld *LayoutDebugger) AnalyzeLayout() LayoutAnalysis {
	info := ld.CaptureDebugInfo()

	analysis := LayoutAnalysis{
		Timestamp:         info.Timestamp,
		OverallHealth:     ld.calculateOverallHealth(info),
		ComponentIssues:   ld.findComponentIssues(info),
		ConstraintIssues:  ld.findConstraintIssues(info),
		PerformanceIssues: ld.findPerformanceIssues(info),
		Recommendations:   ld.generateRecommendations(info),
	}

	return analysis
}

// LayoutAnalysis contains the results of layout analysis
type LayoutAnalysis struct {
	Timestamp         time.Time
	OverallHealth     HealthStatus
	ComponentIssues   []ComponentIssue
	ConstraintIssues  []ConstraintIssue
	PerformanceIssues []PerformanceIssue
	Recommendations   []Recommendation
}

// HealthStatus represents the overall health of the layout
type HealthStatus int

const (
	// HealthGood indicates good layout health
	HealthGood HealthStatus = iota
	// HealthWarning indicates layout health warnings
	HealthWarning
	// HealthError indicates layout health errors
	HealthError
	// HealthCritical indicates critical layout health issues
	HealthCritical
)

// ComponentIssue represents an issue with a component
type ComponentIssue struct {
	ComponentID string
	IssueType   string
	Description string
	Severity    WarningSeverity
}

// ConstraintIssue represents an issue with constraints
type ConstraintIssue struct {
	ComponentID    string
	ConstraintType string
	IssueType      string
	Description    string
	Severity       WarningSeverity
}

// PerformanceIssue represents a performance issue
type PerformanceIssue struct {
	IssueType   string
	Description string
	Impact      string
	Severity    WarningSeverity
}

// Recommendation represents a recommended action
type Recommendation struct {
	Category    string
	Description string
	Priority    int
	Action      string
}

// Analysis implementation functions

// calculateOverallHealth calculates the overall health status
func (ld *LayoutDebugger) calculateOverallHealth(info DebugInfo) HealthStatus {
	errorCount := 0
	warningCount := 0

	for _, warning := range info.WarningsInfo {
		switch warning.Severity {
		case SeverityError, SeverityCritical:
			errorCount++
		case SeverityWarning:
			warningCount++
		}
	}

	if errorCount > 0 {
		return HealthError
	}

	if warningCount > 0 {
		return HealthWarning
	}

	return HealthGood
}

// findComponentIssues finds issues with components
func (ld *LayoutDebugger) findComponentIssues(info DebugInfo) []ComponentIssue {
	var issues []ComponentIssue

	for _, comp := range info.ComponentsInfo {
		if !comp.HasValidLayout {
			issues = append(issues, ComponentIssue{
				ComponentID: comp.ID,
				IssueType:   "invalid_layout",
				Description: "Component has invalid layout dimensions or position",
				Severity:    SeverityError,
			})
		}

		if comp.OverflowsX || comp.OverflowsY {
			issues = append(issues, ComponentIssue{
				ComponentID: comp.ID,
				IssueType:   "overflow",
				Description: "Component extends beyond terminal boundaries",
				Severity:    SeverityWarning,
			})
		}

		if comp.IsDirty {
			issues = append(issues, ComponentIssue{
				ComponentID: comp.ID,
				IssueType:   "dirty",
				Description: "Component needs update",
				Severity:    SeverityInfo,
			})
		}
	}

	return issues
}

// findConstraintIssues finds issues with constraints
func (ld *LayoutDebugger) findConstraintIssues(info DebugInfo) []ConstraintIssue {
	var issues []ConstraintIssue

	for _, constraint := range info.ConstraintsInfo {
		if !constraint.IsValid {
			issues = append(issues, ConstraintIssue{
				ComponentID:    constraint.ComponentID,
				ConstraintType: constraint.ConstraintType,
				IssueType:      "invalid_constraint",
				Description:    constraint.ErrorMessage,
				Severity:       SeverityError,
			})
		}
	}

	return issues
}

// findPerformanceIssues finds performance issues
func (ld *LayoutDebugger) findPerformanceIssues(info DebugInfo) []PerformanceIssue {
	var issues []PerformanceIssue

	// Check for frequent recalculations
	if info.PerformanceInfo.TotalCalculations > 100 {
		issues = append(issues, PerformanceIssue{
			IssueType:   "frequent_recalculation",
			Description: "Layout is being recalculated very frequently",
			Impact:      "May cause performance degradation",
			Severity:    SeverityWarning,
		})
	}

	// Check for slow calculations
	if info.PerformanceInfo.AverageCalculationTime > 100*time.Millisecond {
		issues = append(issues, PerformanceIssue{
			IssueType:   "slow_calculation",
			Description: "Layout calculations are taking too long",
			Impact:      "May cause UI responsiveness issues",
			Severity:    SeverityWarning,
		})
	}

	return issues
}

// generateRecommendations generates recommendations for improvement
func (ld *LayoutDebugger) generateRecommendations(info DebugInfo) []Recommendation {
	var recommendations []Recommendation

	// Check for too many components
	if len(info.ComponentsInfo) > 20 {
		recommendations = append(recommendations, Recommendation{
			Category:    "performance",
			Description: "Consider reducing the number of components",
			Priority:    2,
			Action:      "Combine similar components or use virtualization",
		})
	}

	// Check for complex constraints
	complexConstraintCount := 0
	for _, comp := range info.ComponentsInfo {
		if comp.ConstraintCount > 5 {
			complexConstraintCount++
		}
	}

	if complexConstraintCount > 0 {
		recommendations = append(recommendations, Recommendation{
			Category:    "maintainability",
			Description: "Some components have complex constraint sets",
			Priority:    1,
			Action:      "Simplify constraints or use layout presets",
		})
	}

	return recommendations
}

// Utility functions

// PrintDebugInfo prints debug information to stdout
func (ld *LayoutDebugger) PrintDebugInfo() {
	if ld.mode == DebugModeOff {
		return
	}

	// info := ld.CaptureDebugInfo()
	// fmt.Print(ld.FormatDebugInfo(info))
}

// GetHistory returns the debug history
func (ld *LayoutDebugger) GetHistory() []DebugInfo {
	return ld.history
}

// ClearHistory clears the debug history
func (ld *LayoutDebugger) ClearHistory() {
	ld.history = []DebugInfo{}
}

// ValidateLayout performs comprehensive layout validation
func (ld *LayoutDebugger) ValidateLayout() error {
	info := ld.CaptureDebugInfo()

	// Check for critical issues
	for _, warning := range info.WarningsInfo {
		if warning.Severity == SeverityCritical {
			return fmt.Errorf("critical layout issue: %s", warning.Message)
		}
	}

	// Check for invalid components
	for _, comp := range info.ComponentsInfo {
		if !comp.HasValidLayout {
			return fmt.Errorf("component '%s' has invalid layout", comp.ID)
		}
	}

	// Check for invalid constraints
	for _, constraint := range info.ConstraintsInfo {
		if !constraint.IsValid {
			return fmt.Errorf("component '%s' has invalid constraint: %s",
				constraint.ComponentID, constraint.ErrorMessage)
		}
	}

	return nil
}

// String returns a string representation of the health status
func (hs HealthStatus) String() string {
	switch hs {
	case HealthGood:
		return "Good"
	case HealthWarning:
		return "Warning"
	case HealthError:
		return "Error"
	case HealthCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

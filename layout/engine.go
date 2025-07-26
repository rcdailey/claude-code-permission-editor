package layout

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LayoutEngine is the main orchestrator for the declarative layout system
type LayoutEngine struct {
	// Terminal dimensions
	terminalWidth  int
	terminalHeight int

	// Component management
	components *ComponentRegistry

	// Layout state
	needsRecalculation bool
	lastCalculation    *LayoutResult

	// Configuration
	config EngineConfig
}

// EngineConfig contains configuration options for the layout engine
type EngineConfig struct {
	// Minimum terminal dimensions
	MinTerminalWidth  int
	MinTerminalHeight int

	// Spacing configuration
	DefaultSpacing int

	// Performance settings
	AutoRecalculate bool // Whether to automatically recalculate on changes

	// Debug settings
	DebugMode bool
}

// DefaultEngineConfig returns the default engine configuration
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MinTerminalWidth:  40,
		MinTerminalHeight: 10,
		DefaultSpacing:    1,
		AutoRecalculate:   true,
		DebugMode:         false,
	}
}

// NewLayoutEngine creates a new layout engine with default dimensions (80x24)
func NewLayoutEngine(args ...interface{}) *LayoutEngine {
	switch len(args) {
	case 0:
		return newLayoutEngineWithDimensions(80, 24)
	case 1:
		return handleSingleArg(args[0])
	case 2:
		return handleTwoArgs(args[0], args[1])
	case 3:
		return handleThreeArgs(args[0], args[1], args[2])
	}
	return newLayoutEngineWithDimensions(80, 24)
}

// handleSingleArg handles single argument case for NewLayoutEngine
func handleSingleArg(arg interface{}) *LayoutEngine {
	if config, ok := arg.(EngineConfig); ok {
		return newLayoutEngineWithDimensions(80, 24, config)
	}
	return newLayoutEngineWithDimensions(80, 24)
}

// handleTwoArgs handles two argument case for NewLayoutEngine
func handleTwoArgs(arg1, arg2 interface{}) *LayoutEngine {
	if width, ok := arg1.(int); ok {
		if height, ok := arg2.(int); ok {
			return newLayoutEngineWithDimensions(width, height)
		}
	}
	return newLayoutEngineWithDimensions(80, 24)
}

// handleThreeArgs handles three argument case for NewLayoutEngine
func handleThreeArgs(arg1, arg2, arg3 interface{}) *LayoutEngine {
	if width, ok := arg1.(int); ok {
		if height, ok := arg2.(int); ok {
			if config, ok := arg3.(EngineConfig); ok {
				return newLayoutEngineWithDimensions(width, height, config)
			}
		}
	}
	return newLayoutEngineWithDimensions(80, 24)
}

// newLayoutEngineWithDimensions creates a new layout engine with the given terminal dimensions
func newLayoutEngineWithDimensions(width, height int, config ...EngineConfig) *LayoutEngine {
	var engineConfig EngineConfig
	if len(config) > 0 {
		engineConfig = config[0]
	} else {
		engineConfig = DefaultEngineConfig()
	}

	return &LayoutEngine{
		terminalWidth:      width,
		terminalHeight:     height,
		components:         NewComponentRegistry(),
		needsRecalculation: true,
		lastCalculation:    nil,
		config:             engineConfig,
	}
}

// AddComponent adds a component to the layout engine with the given constraints
func (le *LayoutEngine) AddComponent(id string, component interface{}, constraints ...Constraint) error {
	if err := le.components.Register(id, component, constraints...); err != nil {
		return fmt.Errorf("failed to add component '%s': %w", id, err)
	}

	le.markNeedsRecalculation()

	if le.config.AutoRecalculate {
		return le.Recalculate()
	}

	return nil
}

// RemoveComponent removes a component from the layout engine
func (le *LayoutEngine) RemoveComponent(id string) error {
	if err := le.components.Unregister(id); err != nil {
		return fmt.Errorf("failed to remove component '%s': %w", id, err)
	}

	le.markNeedsRecalculation()

	if le.config.AutoRecalculate {
		return le.Recalculate()
	}

	return nil
}

// GetComponent returns a component by ID
func (le *LayoutEngine) GetComponent(id string) (interface{}, bool) {
	wrapper, exists := le.components.Get(id)
	if !exists {
		return nil, false
	}
	return wrapper.GetComponent(), true
}

// UpdateConstraints updates the constraints for a component
func (le *LayoutEngine) UpdateConstraints(id string, constraints ...Constraint) error {
	wrapper, exists := le.components.Get(id)
	if !exists {
		return fmt.Errorf("component '%s' not found", id)
	}

	wrapper.SetConstraints(constraints...)
	le.markNeedsRecalculation()

	if le.config.AutoRecalculate {
		return le.Recalculate()
	}

	return nil
}

// HandleResize updates the terminal dimensions and recalculates layout
func (le *LayoutEngine) HandleResize(width, height int) error {
	if le.terminalWidth == width && le.terminalHeight == height {
		return nil // No change
	}

	le.terminalWidth = width
	le.terminalHeight = height
	le.markNeedsRecalculation()

	return le.Recalculate()
}

// Recalculate performs layout calculation and updates all components
func (le *LayoutEngine) Recalculate() error {
	// Validate terminal dimensions
	if le.terminalWidth < le.config.MinTerminalWidth || le.terminalHeight < le.config.MinTerminalHeight {
		return fmt.Errorf("terminal dimensions too small: %dx%d (min: %dx%d)",
			le.terminalWidth, le.terminalHeight,
			le.config.MinTerminalWidth, le.config.MinTerminalHeight)
	}

	// Validate all component constraints
	if err := le.components.ValidateAll(); err != nil {
		return fmt.Errorf("constraint validation failed: %w", err)
	}

	// Perform layout calculation
	result := le.calculateLayout()

	// Apply the calculated layout to components
	if err := le.applyLayout(result); err != nil {
		return fmt.Errorf("failed to apply layout: %w", err)
	}

	le.lastCalculation = result
	le.needsRecalculation = false

	return nil
}

// Update handles Bubble Tea update messages and forwards them to components
func (le *LayoutEngine) Update(msg tea.Msg) []tea.Cmd {
	// Handle window resize messages automatically
	if resizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		if err := le.HandleResize(resizeMsg.Width, resizeMsg.Height); err != nil {
			// Log error but don't stop the application
			slog.Error("Layout engine resize failed", "error", err, "width", resizeMsg.Width, "height", resizeMsg.Height)
		}
	}

	// Forward update message to all components
	return le.components.UpdateAll(msg)
}

// View renders all components in their calculated positions using absolute positioning
func (le *LayoutEngine) View() string {
	// Ensure layout is calculated
	if le.needsRecalculation {
		if err := le.Recalculate(); err != nil {
			return fmt.Sprintf("Layout error: %v", err)
		}
	}

	// Use absolute positioning if layout has been calculated
	if le.lastCalculation != nil {
		return le.renderWithAbsolutePositioning()
	}

	// Fallback to simple vertical layout (should rarely happen)
	return le.renderVerticalFallback()
}

// renderWithAbsolutePositioning renders components at their calculated absolute positions
func (le *LayoutEngine) renderWithAbsolutePositioning() string {
	screen := le.createEmptyScreen()
	sortedComponents := le.getSortedComponentsByPosition()
	le.placeComponentsOnScreen(screen, sortedComponents)
	return le.joinScreenLines(screen)
}

// createEmptyScreen creates an empty screen buffer
func (le *LayoutEngine) createEmptyScreen() []string {
	screen := make([]string, le.terminalHeight)
	for i := range screen {
		screen[i] = ""
	}
	return screen
}

// getSortedComponentsByPosition returns components sorted by Y position for proper layering
func (le *LayoutEngine) getSortedComponentsByPosition() []*ComponentWrapper {
	components := le.components.GetInOrder()

	// Sort by Y position (top to bottom)
	for i := 0; i < len(components)-1; i++ {
		for j := i + 1; j < len(components); j++ {
			layout1, exists1 := le.lastCalculation.Components[components[i].ID()]
			layout2, exists2 := le.lastCalculation.Components[components[j].ID()]

			if exists1 && exists2 && layout1.Y > layout2.Y {
				components[i], components[j] = components[j], components[i]
			}
		}
	}

	return components
}

// placeComponentsOnScreen places all components on the screen buffer at their calculated positions
func (le *LayoutEngine) placeComponentsOnScreen(screen []string, components []*ComponentWrapper) {
	for _, component := range components {
		if layout, exists := le.lastCalculation.Components[component.ID()]; exists {
			content := component.View()
			if content != "" {
				le.placeComponentContentAtPosition(screen, content, layout.Y)
			}
		}
	}
}

// placeComponentContentAtPosition places component content at the specified Y position
func (le *LayoutEngine) placeComponentContentAtPosition(screen []string, content string, y int) {
	lines := strings.Split(content, "\n")

	for lineIdx, line := range lines {
		screenY := y + lineIdx
		if screenY >= 0 && screenY < len(screen) {
			screen[screenY] = line
		}
	}
}

// joinScreenLines joins screen lines into final output
func (le *LayoutEngine) joinScreenLines(screen []string) string {
	var resultBuilder strings.Builder
	for i, line := range screen {
		resultBuilder.WriteString(line)
		if i < len(screen)-1 {
			resultBuilder.WriteRune('\n')
		}
	}
	return resultBuilder.String()
}

// renderVerticalFallback provides fallback vertical layout (for compatibility)
func (le *LayoutEngine) renderVerticalFallback() string {
	result := ""
	for _, component := range le.components.GetInOrder() {
		content := component.View()
		if content != "" {
			result += content
			if result != "" && result[len(result)-1] != '\n' {
				result += "\n"
			}
		}
	}
	return result
}

// calculateLayout performs the actual layout calculation
func (le *LayoutEngine) calculateLayout() *LayoutResult {
	result := le.initializeLayoutResult()
	components := le.components.GetInOrder()

	if len(components) == 0 {
		return result
	}

	le.processComponentsLayout(components, result)

	return result
}

// initializeLayoutResult creates a new layout result with default values
func (le *LayoutEngine) initializeLayoutResult() *LayoutResult {
	return &LayoutResult{
		TerminalWidth:  le.terminalWidth,
		TerminalHeight: le.terminalHeight,
		Components:     make(map[string]ComponentLayout),
		Warnings:       []string{},
	}
}

// processComponentsLayout processes all components and calculates their layout
func (le *LayoutEngine) processComponentsLayout(components []*ComponentWrapper, result *LayoutResult) {
	// Separate anchored components from sequential components
	anchoredComponents, sequentialComponents := le.separateAnchoredComponents(components)

	// Layout anchored components first
	reservedSpace := le.layoutAnchoredComponents(anchoredComponents, result)

	// Layout sequential components in remaining space
	fixedHeight, flexComponents, totalFlexWeight := le.categorizeComponents(sequentialComponents)
	availableHeight := le.calculateAvailableHeight(sequentialComponents, fixedHeight) - reservedSpace
	flexHeights := le.calculateFlexHeights(flexComponents, totalFlexWeight, availableHeight)
	le.layoutAllComponents(sequentialComponents, flexHeights, availableHeight, result)
}

// categorizeComponents separates fixed and flex components
func (le *LayoutEngine) categorizeComponents(components []*ComponentWrapper) (int, []*ComponentWrapper, float64) {
	fixedHeight := 0
	flexComponents := []*ComponentWrapper{}
	totalFlexWeight := 0.0

	for _, component := range components {
		sizeConstraint, valid := le.getHeightConstraint(component)
		if !valid {
			continue
		}

		if sizeConstraint.Value().IsFlexible() {
			le.handleFlexComponent(component, *sizeConstraint, &flexComponents, &totalFlexWeight)
		} else {
			le.handleFixedComponent(*sizeConstraint, &fixedHeight)
		}
	}

	return fixedHeight, flexComponents, totalFlexWeight
}

// getHeightConstraint extracts height constraint from component
func (le *LayoutEngine) getHeightConstraint(component *ComponentWrapper) (*SizeConstraint, bool) {
	constraints := component.Constraints()
	heightConstraint, exists := constraints.Get(ConstraintHeight)
	if !exists {
		return nil, false
	}

	if sizeConstraint, ok := heightConstraint.(SizeConstraint); ok {
		return &sizeConstraint, true
	}

	return nil, false
}

// handleFlexComponent processes a flex component
func (le *LayoutEngine) handleFlexComponent(component *ComponentWrapper, sizeConstraint SizeConstraint,
	flexComponents *[]*ComponentWrapper, totalFlexWeight *float64) {
	if flexSize, ok := sizeConstraint.Value().(FlexSize); ok {
		*flexComponents = append(*flexComponents, component)
		*totalFlexWeight += flexSize.Weight()
	}
}

// handleFixedComponent processes a fixed component
func (le *LayoutEngine) handleFixedComponent(sizeConstraint SizeConstraint, fixedHeight *int) {
	if fixedSize, ok := sizeConstraint.Value().(FixedSize); ok {
		*fixedHeight += fixedSize.Calculate(le.terminalHeight)
	}
}

// calculateAvailableHeight calculates space available for flex components
func (le *LayoutEngine) calculateAvailableHeight(components []*ComponentWrapper, fixedHeight int) int {
	spacingHeight := 0
	if len(components) > 1 {
		spacingHeight = (len(components) - 1) * le.config.DefaultSpacing
	}
	availableHeight := le.terminalHeight - fixedHeight - spacingHeight
	if availableHeight < 0 {
		availableHeight = 0
	}
	return availableHeight
}

// calculateFlexHeights pre-calculates heights for flex components
func (le *LayoutEngine) calculateFlexHeights(
	flexComponents []*ComponentWrapper, totalFlexWeight float64, availableHeight int) map[string]int {
	flexHeights := make(map[string]int)
	if totalFlexWeight <= 0 {
		return flexHeights
	}

	for _, component := range flexComponents {
		if flexSize, valid := le.getFlexSize(component); valid {
			height := int(float64(availableHeight) * (flexSize.Weight() / totalFlexWeight))
			flexHeights[component.ID()] = height
		}
	}
	return flexHeights
}

// getFlexSize extracts flex size from component
func (le *LayoutEngine) getFlexSize(component *ComponentWrapper) (*FlexSize, bool) {
	sizeConstraint, valid := le.getHeightConstraint(component)
	if !valid {
		return nil, false
	}
	if flexSize, ok := sizeConstraint.Value().(FlexSize); ok {
		return &flexSize, true
	}
	return nil, false
}

// layoutAllComponents layouts all components with calculated dimensions
func (le *LayoutEngine) layoutAllComponents(
	components []*ComponentWrapper, flexHeights map[string]int, availableHeight int, result *LayoutResult) {
	currentY := 0
	for _, component := range components {
		var width, height int
		if flexHeight, isFlexComponent := flexHeights[component.ID()]; isFlexComponent {
			width = le.calculateComponentWidth(component.Constraints())
			height = flexHeight
		} else {
			width, height = le.calculateComponentDimensionsWithContext(component, availableHeight)
		}
		currentY = le.layoutComponent(component, width, height, currentY, result)
	}
}


// calculateComponentDimensionsWithContext calculates dimensions with available space context
func (le *LayoutEngine) calculateComponentDimensionsWithContext(
	component *ComponentWrapper, availableHeight int) (int, int) {
	constraints := component.Constraints()

	width := le.calculateComponentWidth(constraints)
	height := le.calculateComponentHeightWithContext(constraints, availableHeight)

	return width, height
}

// calculateComponentWidth calculates the width for a component based on constraints
func (le *LayoutEngine) calculateComponentWidth(constraints ConstraintSet) int {
	// Use full terminal width - let's debug what's actually happening
	width := le.terminalWidth

	if widthConstraint, exists := constraints.Get(ConstraintWidth); exists {
		if sizeConstraint, ok := widthConstraint.(SizeConstraint); ok {
			width = sizeConstraint.Value().Calculate(le.terminalWidth)
		}
	}

	return width
}


// calculateComponentHeightWithContext calculates height using available space for percentages
func (le *LayoutEngine) calculateComponentHeightWithContext(constraints ConstraintSet, availableHeight int) int {
	height := 10 // Default height

	if heightConstraint, exists := constraints.Get(ConstraintHeight); exists {
		if sizeConstraint, ok := heightConstraint.(SizeConstraint); ok {
			// Handle different size types appropriately
			switch sizeValue := sizeConstraint.Value().(type) {
			case FlexSize:
				// For flex sizes, calculate based on weight and available space
				// This is a simplified approach - ideally we'd collect all flex components first
				height = int(float64(availableHeight) * sizeValue.Weight())
			default:
				// Use terminalHeight for fixed sizes
				height = sizeValue.Calculate(le.terminalHeight)
			}
		}
	}

	return le.applyMinHeightConstraintWithContext(constraints, height)
}


// applyMinHeightConstraintWithContext applies minimum height constraints with available space context
func (le *LayoutEngine) applyMinHeightConstraintWithContext(
	constraints ConstraintSet, height int) int {
	minHeightConstraint, exists := constraints.Get(ConstraintMinHeight)
	if !exists {
		return height
	}

	sizeConstraint, ok := minHeightConstraint.(SizeConstraint)
	if !ok {
		return height
	}

	// Calculate minimum height - use terminalHeight for all size types
	minHeight := sizeConstraint.Value().Calculate(le.terminalHeight)

	if height < minHeight {
		return minHeight
	}
	return height
}

// layoutComponent positions a component and updates the layout result
func (le *LayoutEngine) layoutComponent(
	component *ComponentWrapper, width, height, currentY int, result *LayoutResult) int {
	le.checkLayoutBounds(component, currentY, height, result)

	result.Components[component.ID()] = ComponentLayout{
		X:      0,
		Y:      currentY,
		Width:  width,
		Height: height,
	}

	return currentY + height + le.config.DefaultSpacing
}

// separateAnchoredComponents separates components with anchor constraints from sequential ones
func (le *LayoutEngine) separateAnchoredComponents(
	components []*ComponentWrapper) ([]*ComponentWrapper, []*ComponentWrapper) {
	var anchored, sequential []*ComponentWrapper

	for _, component := range components {
		if le.hasAnchorConstraint(component) {
			anchored = append(anchored, component)
		} else {
			sequential = append(sequential, component)
		}
	}

	return anchored, sequential
}

// hasAnchorConstraint checks if a component has an anchor constraint
func (le *LayoutEngine) hasAnchorConstraint(component *ComponentWrapper) bool {
	constraints := component.Constraints()
	_, exists := constraints.Get(ConstraintAnchor)
	return exists
}

// layoutAnchoredComponents positions anchored components and returns reserved space
func (le *LayoutEngine) layoutAnchoredComponents(components []*ComponentWrapper, result *LayoutResult) int {
	reservedSpace := 0

	for _, component := range components {
		constraints := component.Constraints()
		anchorConstraint, exists := constraints.Get(ConstraintAnchor)
		if !exists {
			continue
		}

		anchor, ok := anchorConstraint.(AnchorConstraint)
		if !ok {
			continue
		}

		// Calculate component dimensions
		width := le.calculateComponentWidth(constraints)
		height := le.calculateComponentHeightWithContext(constraints, le.terminalHeight)

		// Position based on anchor
		x, y := le.calculateAnchoredPosition(anchor.Position(), height)

		result.Components[component.ID()] = ComponentLayout{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		}

		// Reserve space for bottom-anchored components
		if anchor.Position() == AnchorBottom {
			reservedSpace += height
		}
	}

	return reservedSpace
}

// calculateAnchoredPosition calculates position based on anchor constraint
//
//nolint:unparam // X coordinate is always 0 for now, but function is prepared for future horizontal anchoring
func (le *LayoutEngine) calculateAnchoredPosition(anchor AnchorPosition, height int) (int, int) {
	const leftX = 0 // All components start at left edge for now

	switch anchor {
	case AnchorBottom:
		return leftX, le.terminalHeight - height
	case AnchorTop:
		return leftX, 0
	case AnchorCenter:
		return leftX, (le.terminalHeight - height) / 2
	// Add more anchor positions as needed
	default:
		return leftX, 0
	}
}

// checkLayoutBounds checks if the component fits within terminal bounds
func (le *LayoutEngine) checkLayoutBounds(component *ComponentWrapper, currentY, height int, result *LayoutResult) {
	if currentY+height > le.terminalHeight {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Component '%s' may exceed terminal height", component.ID()))
	}
}

// applyLayout applies the calculated layout to all components
func (le *LayoutEngine) applyLayout(result *LayoutResult) error {
	for id, layout := range result.Components {
		wrapper, exists := le.components.Get(id)
		if !exists {
			return fmt.Errorf("component '%s' not found during layout application", id)
		}

		// Apply calculated dimensions to components
		// Note: Style-aware sizing is currently handled in the UI layer
		// Future enhancement: integrate with BubbleTeaIntegration for automatic style-aware sizing
		wrapper.SetSize(layout.Width, layout.Height)

		wrapper.SetPosition(layout.X, layout.Y)
	}

	return nil
}

// markNeedsRecalculation marks that the layout needs recalculation
func (le *LayoutEngine) markNeedsRecalculation() {
	le.needsRecalculation = true
}

// GetTerminalDimensions returns the current terminal dimensions
func (le *LayoutEngine) GetTerminalDimensions() (width, height int) {
	return le.terminalWidth, le.terminalHeight
}

// GetComponentCount returns the number of registered components
func (le *LayoutEngine) GetComponentCount() int {
	return le.components.Count()
}

// GetComponentIDs returns all registered component IDs
func (le *LayoutEngine) GetComponentIDs() []string {
	return le.components.GetIDs()
}

// GetLastCalculation returns the last layout calculation result
func (le *LayoutEngine) GetLastCalculation() *LayoutResult {
	return le.lastCalculation
}

// NeedsRecalculation returns true if the layout needs recalculation
func (le *LayoutEngine) NeedsRecalculation() bool {
	return le.needsRecalculation
}

// SetConfig updates the engine configuration
func (le *LayoutEngine) SetConfig(config EngineConfig) {
	le.config = config
	le.markNeedsRecalculation()
}

// GetConfig returns the current engine configuration
func (le *LayoutEngine) GetConfig() EngineConfig {
	return le.config
}

// GetLastResult returns the last calculated layout result
func (le *LayoutEngine) GetLastResult() *LayoutResult {
	return le.lastCalculation
}

// GetTerminalSize returns the current terminal dimensions
func (le *LayoutEngine) GetTerminalSize() (width, height int) {
	return le.terminalWidth, le.terminalHeight
}

// HasComponents returns true if any components are registered
func (le *LayoutEngine) HasComponents() bool {
	return le.components.Count() > 0
}

// SetTerminalSize updates the terminal dimensions and triggers recalculation if needed
func (le *LayoutEngine) SetTerminalSize(width, height int) {
	if le.terminalWidth != width || le.terminalHeight != height {
		le.terminalWidth = width
		le.terminalHeight = height
		le.markNeedsRecalculation()

		if le.config.AutoRecalculate {
			_ = le.Recalculate() // Ignore error for auto-recalculate
		}
	}
}

// Layout performs a layout calculation and returns any error
func (le *LayoutEngine) Layout() error {
	return le.Recalculate()
}

// AddComponentWrapper accepts a pre-configured ComponentWrapper
func (le *LayoutEngine) AddComponentWrapper(wrapper *ComponentWrapper) error {
	if err := le.components.RegisterWrapper(wrapper); err != nil {
		return fmt.Errorf("failed to add component '%s': %w", wrapper.ID(), err)
	}

	le.markNeedsRecalculation()

	if le.config.AutoRecalculate {
		return le.Recalculate()
	}

	return nil
}

// SetComponentContent sets the rendered content for a component (used by UI layer)
func (le *LayoutEngine) SetComponentContent(componentID string, content string) error {
	wrapper, exists := le.components.Get(componentID)
	if !exists {
		return fmt.Errorf("component '%s' not found", componentID)
	}

	wrapper.SetContent(content)
	return nil
}

// SetAllComponentContent sets content for multiple components at once
func (le *LayoutEngine) SetAllComponentContent(contentMap map[string]string) error {
	for componentID, content := range contentMap {
		if err := le.SetComponentContent(componentID, content); err != nil {
			return fmt.Errorf("failed to set content for component '%s': %w", componentID, err)
		}
	}
	return nil
}

// LayoutResult represents the result of a layout calculation
type LayoutResult struct {
	TerminalWidth  int
	TerminalHeight int
	Components     map[string]ComponentLayout
	Warnings       []string
}

// ComponentLayout represents the calculated layout for a single component
type ComponentLayout struct {
	X, Y          int
	Width, Height int
}

// IsValid checks if the layout result is valid
func (lr *LayoutResult) IsValid() bool {
	return len(lr.Warnings) == 0
}

// GetComponent returns the layout for a specific component
func (lr *LayoutResult) GetComponent(id string) (ComponentLayout, bool) {
	layout, exists := lr.Components[id]
	return layout, exists
}

// GetComponentsSorted returns all components sorted by Y position (top to bottom)
func (lr *LayoutResult) GetComponentsSorted() []ComponentLayoutInfo {
	components := make([]ComponentLayoutInfo, 0, len(lr.Components))
	for id, layout := range lr.Components {
		components = append(components, ComponentLayoutInfo{
			ID:     id,
			Layout: layout,
		})
	}

	sort.Slice(components, func(i, j int) bool {
		return components[i].Layout.Y < components[j].Layout.Y
	})

	return components
}

// ComponentLayoutInfo combines component ID with its layout
type ComponentLayoutInfo struct {
	ID     string
	Layout ComponentLayout
}

// Validate validates the layout result
func (lr *LayoutResult) Validate() error {
	for id, layout := range lr.Components {
		if layout.Width <= 0 || layout.Height <= 0 {
			return fmt.Errorf("component '%s' has invalid dimensions: %dx%d",
				id, layout.Width, layout.Height)
		}

		if layout.X < 0 || layout.Y < 0 {
			return fmt.Errorf("component '%s' has invalid position: (%d,%d)",
				id, layout.X, layout.Y)
		}
	}
	return nil
}

// Helper functions for common layout operations

// CreateSimpleVerticalLayout creates a simple vertical layout with the given components
func CreateSimpleVerticalLayout(width, height int, components ...string) *LayoutEngine {
	engine := NewLayoutEngine(width, height)

	for _, id := range components {
		// Create a simple component wrapper for demonstration
		component := &SimpleComponent{id: id}

		// Add with basic constraints
		_ = engine.AddComponent(id, component,
			Height(Flex(1.0)),
		)
	}

	return engine
}

// SimpleComponent is a basic component implementation for testing
type SimpleComponent struct {
	id     string
	width  int
	height int
}

// SetWidth implements component width setting
func (sc *SimpleComponent) SetWidth(width int) {
	sc.width = width
}

// SetHeight implements component height setting
func (sc *SimpleComponent) SetHeight(height int) {
	sc.height = height
}

// Width returns the component width
func (sc *SimpleComponent) Width() int {
	return sc.width
}

// Height returns the component height
func (sc *SimpleComponent) Height() int {
	return sc.height
}

// Init implements Bubble Tea init interface
func (sc *SimpleComponent) Init() tea.Cmd {
	return nil
}

// Update implements Bubble Tea update interface
func (sc *SimpleComponent) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return sc, nil
}

// View implements Bubble Tea view interface
func (sc *SimpleComponent) View() string {
	return fmt.Sprintf("Component: %s (%dx%d)", sc.id, sc.width, sc.height)
}

// GetSize returns the component size
func (sc *SimpleComponent) GetSize() (int, int) {
	return sc.width, sc.height
}

// SetSize sets the component size
func (sc *SimpleComponent) SetSize(width, height int) {
	sc.width = width
	sc.height = height
}

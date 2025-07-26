// Package layout provides a declarative constraint-based layout system for terminal applications.
// This file contains the layout calculation engine that handles space distribution,
// constraint resolution, and positioning algorithms.
package layout

import (
	"fmt"
	"math"
)

// LayoutCalculator handles the sophisticated layout calculation algorithms
type LayoutCalculator struct {
	terminalWidth  int
	terminalHeight int
	spacing        int
	components     []*ComponentWrapper
}

// NewLayoutCalculator creates a new layout calculator
func NewLayoutCalculator(terminalWidth, terminalHeight, spacing int) *LayoutCalculator {
	return &LayoutCalculator{
		terminalWidth:  terminalWidth,
		terminalHeight: terminalHeight,
		spacing:        spacing,
		components:     []*ComponentWrapper{},
	}
}

// AddComponent adds a component to the calculation
func (lc *LayoutCalculator) AddComponent(component *ComponentWrapper) {
	lc.components = append(lc.components, component)
}

// Calculate performs the full layout calculation
func (lc *LayoutCalculator) Calculate() (*LayoutResult, error) {
	if len(lc.components) == 0 {
		return &LayoutResult{
			TerminalWidth:  lc.terminalWidth,
			TerminalHeight: lc.terminalHeight,
			Components:     make(map[string]ComponentLayout),
			Warnings:       []string{},
		}, nil
	}

	// Step 1: Resolve relationships and create dependency graph
	dependencyGraph, err := lc.buildDependencyGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Step 2: Topological sort to determine layout order
	layoutOrder, err := lc.topologicalSort(dependencyGraph)
	if err != nil {
		return nil, fmt.Errorf("failed to sort components: %w", err)
	}

	// Step 3: Calculate dimensions for each component
	dimensions, warnings := lc.calculateDimensions(layoutOrder)

	// Step 4: Calculate positions based on dimensions and relationships
	positions, positionWarnings := lc.calculatePositions(layoutOrder, dimensions)

	warnings = append(warnings, positionWarnings...)

	// Step 5: Build final result
	result := &LayoutResult{
		TerminalWidth:  lc.terminalWidth,
		TerminalHeight: lc.terminalHeight,
		Components:     make(map[string]ComponentLayout),
		Warnings:       warnings,
	}

	for _, component := range lc.components {
		id := component.ID()
		dim := dimensions[id]
		pos := positions[id]

		result.Components[id] = ComponentLayout{
			X:      pos.X,
			Y:      pos.Y,
			Width:  dim.Width,
			Height: dim.Height,
		}
	}

	return result, nil
}

// ComponentDimensions represents calculated dimensions for a component
type ComponentDimensions struct {
	Width  int
	Height int
}

// buildDependencyGraph creates a dependency graph based on relationship constraints
func (lc *LayoutCalculator) buildDependencyGraph() (map[string][]string, error) {
	graph := make(map[string][]string)

	// Initialize graph with all components
	for _, component := range lc.components {
		graph[component.ID()] = []string{}
	}

	// Add dependencies based on relationship constraints
	for _, component := range lc.components {
		constraints := component.Constraints()

		// Check for relationship constraints
		for _, constraint := range constraints.All() {
			if relConstraint, ok := constraint.(RelationshipConstraint); ok {
				targetID := relConstraint.TargetID()

				// Verify target exists
				if _, exists := graph[targetID]; !exists {
					return nil, fmt.Errorf("component '%s' references non-existent target '%s'",
						component.ID(), targetID)
				}

				// Add dependency based on relationship type
				switch relConstraint.Type() {
				case ConstraintBelow:
					// Component depends on target (target must be positioned first)
					graph[component.ID()] = append(graph[component.ID()], targetID)
				case ConstraintAbove:
					// Target depends on component (component must be positioned first)
					graph[targetID] = append(graph[targetID], component.ID())
				case ConstraintRight:
					// Component depends on target for horizontal positioning
					graph[component.ID()] = append(graph[component.ID()], targetID)
				case ConstraintLeft:
					// Target depends on component for horizontal positioning
					graph[targetID] = append(graph[targetID], component.ID())
				}
			}
		}
	}

	return graph, nil
}

// topologicalSort performs topological sorting on the dependency graph
func (lc *LayoutCalculator) topologicalSort(graph map[string][]string) ([]*ComponentWrapper, error) {
	inDegree := lc.calculateInDegrees(graph)
	queue := lc.findZeroInDegreeNodes(inDegree)
	result, processed := lc.processTopologicalQueue(graph, inDegree, queue)

	if processed != len(lc.components) {
		return nil, fmt.Errorf("circular dependency detected in layout constraints")
	}

	return result, nil
}

// calculateInDegrees calculates in-degree count for each node in the graph
func (lc *LayoutCalculator) calculateInDegrees(graph map[string][]string) map[string]int {
	inDegree := make(map[string]int)

	for id := range graph {
		inDegree[id] = 0
	}

	for _, deps := range graph {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	return inDegree
}

// findZeroInDegreeNodes finds nodes with no incoming edges
func (lc *LayoutCalculator) findZeroInDegreeNodes(inDegree map[string]int) []string {
	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	return queue
}

// processTopologicalQueue processes nodes in topological order
func (lc *LayoutCalculator) processTopologicalQueue(
	graph map[string][]string, inDegree map[string]int, queue []string) ([]*ComponentWrapper, int) {
	result := []*ComponentWrapper{}
	processed := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentComponent := lc.findComponentByID(current)
		if currentComponent == nil {
			continue
		}

		result = append(result, currentComponent)
		processed++

		queue = lc.updateInDegrees(graph, inDegree, current, queue)
	}

	return result, processed
}

// findComponentByID finds a component wrapper by its ID
func (lc *LayoutCalculator) findComponentByID(id string) *ComponentWrapper {
	for _, comp := range lc.components {
		if comp.ID() == id {
			return comp
		}
	}
	return nil
}

// updateInDegrees reduces in-degree for dependent nodes and updates queue
func (lc *LayoutCalculator) updateInDegrees(
	graph map[string][]string, inDegree map[string]int, current string, queue []string) []string {
	for _, dep := range graph[current] {
		inDegree[dep]--
		if inDegree[dep] == 0 {
			queue = append(queue, dep)
		}
	}
	return queue
}

// calculateDimensions calculates width and height for all components
func (lc *LayoutCalculator) calculateDimensions(layoutOrder []*ComponentWrapper) (
	map[string]ComponentDimensions, []string) {
	dimensions := make(map[string]ComponentDimensions)
	warnings := []string{}

	// First pass: calculate non-flexible dimensions
	flexComponents, totalFlexHeight, usedHeight := lc.calculateFixedDimensions(layoutOrder, dimensions)

	// Add spacing between components
	if len(layoutOrder) > 1 {
		usedHeight += (len(layoutOrder) - 1) * lc.spacing
	}

	// Second pass: distribute remaining space among flexible components
	flexWarnings := lc.calculateFlexibleDimensions(flexComponents, totalFlexHeight, usedHeight, dimensions)
	warnings = append(warnings, flexWarnings...)

	return dimensions, warnings
}

// calculateFixedDimensions processes components with fixed dimensions
func (lc *LayoutCalculator) calculateFixedDimensions(layoutOrder []*ComponentWrapper,
	dimensions map[string]ComponentDimensions) ([]*ComponentWrapper, float64, int) {
	flexComponents := []*ComponentWrapper{}
	totalFlexHeight := 0.0
	usedHeight := 0

	for _, component := range layoutOrder {
		constraints := component.Constraints()
		id := component.ID()

		width := lc.calculateWidth(constraints)
		height, isFlexible := lc.calculateHeight(constraints)

		if isFlexible {
			flexComponents = append(flexComponents, component)
			totalFlexHeight += lc.getFlexWeight(constraints)
		} else {
			dimensions[id] = ComponentDimensions{Width: width, Height: height}
			usedHeight += height
		}
	}

	return flexComponents, totalFlexHeight, usedHeight
}

// calculateFlexibleDimensions processes components with flexible dimensions
func (lc *LayoutCalculator) calculateFlexibleDimensions(flexComponents []*ComponentWrapper,
	totalFlexHeight float64, usedHeight int, dimensions map[string]ComponentDimensions) []string {
	warnings := []string{}
	remainingHeight := lc.terminalHeight - usedHeight

	if remainingHeight < 0 {
		warnings = append(warnings, fmt.Sprintf("Components exceed terminal height by %d pixels", -remainingHeight))
		remainingHeight = 0
	}

	for _, component := range flexComponents {
		constraints := component.Constraints()
		id := component.ID()

		width := lc.calculateWidth(constraints)
		height := lc.calculateFlexHeight(constraints, remainingHeight, totalFlexHeight)
		height = lc.applyHeightConstraints(constraints, height)

		dimensions[id] = ComponentDimensions{Width: width, Height: height}
	}

	return warnings
}

// getFlexWeight extracts the flex weight from constraints
func (lc *LayoutCalculator) getFlexWeight(constraints ConstraintSet) float64 {
	if flexConstraint, exists := constraints.Get(ConstraintHeight); exists {
		if sizeConstraint, ok := flexConstraint.(SizeConstraint); ok {
			if flexSize, ok := sizeConstraint.Value().(FlexSize); ok {
				return flexSize.Weight()
			}
		}
	}
	return 0.0
}

// calculateFlexHeight calculates height for flexible components
func (lc *LayoutCalculator) calculateFlexHeight(constraints ConstraintSet, remainingHeight int,
	totalFlexHeight float64) int {
	heightConstraint, exists := constraints.Get(ConstraintHeight)
	if !exists {
		return 0
	}

	sizeConstraint, ok := heightConstraint.(SizeConstraint)
	if !ok {
		return 0
	}

	flexSize, ok := sizeConstraint.Value().(FlexSize)
	if !ok {
		return 0
	}

	if totalFlexHeight <= 0 {
		return 0
	}

	return int(math.Round(float64(remainingHeight) * (flexSize.Weight() / totalFlexHeight)))
}

// applyHeightConstraints applies min/max height constraints
func (lc *LayoutCalculator) applyHeightConstraints(constraints ConstraintSet, height int) int {
	// Apply minimum height constraints
	if minHeightConstraint, exists := constraints.Get(ConstraintMinHeight); exists {
		if sizeConstraint, ok := minHeightConstraint.(SizeConstraint); ok {
			minHeight := sizeConstraint.Value().Calculate(lc.terminalHeight)
			if height < minHeight {
				height = minHeight
			}
		}
	}

	// Apply maximum height constraints
	if maxHeightConstraint, exists := constraints.Get(ConstraintMaxHeight); exists {
		if sizeConstraint, ok := maxHeightConstraint.(SizeConstraint); ok {
			maxHeight := sizeConstraint.Value().Calculate(lc.terminalHeight)
			if height > maxHeight {
				height = maxHeight
			}
		}
	}

	return height
}

// calculateWidth calculates the width for a component
func (lc *LayoutCalculator) calculateWidth(constraints ConstraintSet) int {
	width := lc.terminalWidth // Default to full width

	if widthConstraint, exists := constraints.Get(ConstraintWidth); exists {
		if sizeConstraint, ok := widthConstraint.(SizeConstraint); ok {
			width = sizeConstraint.Value().Calculate(lc.terminalWidth)
		}
	}

	// Apply minimum width constraints
	if minWidthConstraint, exists := constraints.Get(ConstraintMinWidth); exists {
		if sizeConstraint, ok := minWidthConstraint.(SizeConstraint); ok {
			minWidth := sizeConstraint.Value().Calculate(lc.terminalWidth)
			if width < minWidth {
				width = minWidth
			}
		}
	}

	// Apply maximum width constraints
	if maxWidthConstraint, exists := constraints.Get(ConstraintMaxWidth); exists {
		if sizeConstraint, ok := maxWidthConstraint.(SizeConstraint); ok {
			maxWidth := sizeConstraint.Value().Calculate(lc.terminalWidth)
			if width > maxWidth {
				width = maxWidth
			}
		}
	}

	return width
}

// calculateHeight calculates the height for a component and returns whether it's flexible
func (lc *LayoutCalculator) calculateHeight(constraints ConstraintSet) (int, bool) {
	height := 10 // Default height

	if heightConstraint, exists := constraints.Get(ConstraintHeight); exists {
		if sizeConstraint, ok := heightConstraint.(SizeConstraint); ok {
			if sizeConstraint.Value().IsFlexible() {
				return 0, true // Actual height calculated later
			}
			height = sizeConstraint.Value().Calculate(lc.terminalHeight)
		}
	}

	return height, false
}

// calculatePositions calculates positions for all components
func (lc *LayoutCalculator) calculatePositions(layoutOrder []*ComponentWrapper,
	dimensions map[string]ComponentDimensions) (map[string]Position, []string) {
	positions := make(map[string]Position)
	warnings := []string{}
	currentY := 0

	for _, component := range layoutOrder {
		constraints := component.Constraints()
		id := component.ID()
		dim := dimensions[id]

		x, y := lc.calculateComponentPosition(component, dim, currentY, positions, dimensions)

		boundsWarnings := lc.checkBounds(id, x, y, dim)
		warnings = append(warnings, boundsWarnings...)

		positions[id] = Position{X: x, Y: y}

		if !lc.hasSpecificPositioning(constraints) {
			currentY = y + dim.Height + lc.spacing
		}
	}

	return positions, warnings
}

// calculateComponentPosition calculates position for a single component
func (lc *LayoutCalculator) calculateComponentPosition(component *ComponentWrapper,
	dim ComponentDimensions, currentY int, positions map[string]Position,
	dimensions map[string]ComponentDimensions) (int, int) {
	constraints := component.Constraints()
	x, y := 0, currentY

	// Check for anchor constraints
	if anchorConstraint, exists := constraints.Get(ConstraintAnchor); exists {
		if anchor, ok := anchorConstraint.(AnchorConstraint); ok {
			x, y = lc.calculateAnchorPosition(anchor.Position(), dim, currentY)
		}
	}

	// Check for relationship constraints
	relX, relY, err := lc.calculateRelationshipPosition(component, dim, positions, dimensions)
	if err == nil {
		x, y = relX, relY
	}

	return x, y
}

// calculateRelationshipPosition calculates position based on relationship constraints
func (lc *LayoutCalculator) calculateRelationshipPosition(component *ComponentWrapper,
	dim ComponentDimensions, positions map[string]Position,
	dimensions map[string]ComponentDimensions) (int, int, error) {
	constraints := component.Constraints()

	for _, constraint := range constraints.All() {
		if relConstraint, ok := constraint.(RelationshipConstraint); ok {
			targetID := relConstraint.TargetID()
			targetPos, exists := positions[targetID]
			if !exists {
				return 0, 0, fmt.Errorf("target component '%s' not positioned before '%s'", targetID, component.ID())
			}

			targetDim := dimensions[targetID]
			x, y := lc.calculateRelativePosition(relConstraint, targetPos, targetDim, dim)
			return x, y, nil
		}
	}

	return 0, 0, fmt.Errorf("no relationship constraints found")
}

// calculateRelativePosition calculates position relative to another component
func (lc *LayoutCalculator) calculateRelativePosition(relConstraint RelationshipConstraint,
	targetPos Position, targetDim, dim ComponentDimensions) (int, int) {
	switch relConstraint.Type() {
	case ConstraintBelow:
		return targetPos.X, targetPos.Y + targetDim.Height + lc.spacing + relConstraint.Offset()
	case ConstraintAbove:
		return targetPos.X, targetPos.Y - dim.Height - lc.spacing - relConstraint.Offset()
	case ConstraintRight:
		return targetPos.X + targetDim.Width + lc.spacing + relConstraint.Offset(), targetPos.Y
	case ConstraintLeft:
		return targetPos.X - dim.Width - lc.spacing - relConstraint.Offset(), targetPos.Y
	default:
		return targetPos.X, targetPos.Y
	}
}

// checkBounds validates component position within terminal bounds
func (lc *LayoutCalculator) checkBounds(id string, x, y int, dim ComponentDimensions) []string {
	var warnings []string

	if x < 0 {
		warnings = append(warnings, fmt.Sprintf("Component '%s' positioned outside left boundary", id))
	}
	if y < 0 {
		warnings = append(warnings, fmt.Sprintf("Component '%s' positioned outside top boundary", id))
	}
	if x+dim.Width > lc.terminalWidth {
		warnings = append(warnings, fmt.Sprintf("Component '%s' extends beyond right boundary", id))
	}
	if y+dim.Height > lc.terminalHeight {
		warnings = append(warnings, fmt.Sprintf("Component '%s' extends beyond bottom boundary", id))
	}

	return warnings
}

// calculateAnchorPosition calculates position based on anchor constraint
func (lc *LayoutCalculator) calculateAnchorPosition(anchor AnchorPosition,
	dim ComponentDimensions, defaultY int) (int, int) {
	switch anchor {
	case AnchorTopLeft, AnchorLeft, AnchorBottomLeft:
		return lc.calculateLeftAnchor(anchor, dim, defaultY)
	case AnchorTop, AnchorCenter, AnchorBottom:
		return lc.calculateCenterAnchor(anchor, dim, defaultY)
	case AnchorTopRight, AnchorRight, AnchorBottomRight:
		return lc.calculateRightAnchor(anchor, dim, defaultY)
	default:
		return 0, defaultY
	}
}

// calculateLeftAnchor calculates position for left-aligned anchors
func (lc *LayoutCalculator) calculateLeftAnchor(anchor AnchorPosition,
	dim ComponentDimensions, defaultY int) (int, int) {
	switch anchor {
	case AnchorTopLeft:
		return 0, 0
	case AnchorLeft:
		return 0, (lc.terminalHeight - dim.Height) / 2
	case AnchorBottomLeft:
		return 0, lc.terminalHeight - dim.Height
	default:
		return 0, defaultY
	}
}

// calculateCenterAnchor calculates position for center-aligned anchors
func (lc *LayoutCalculator) calculateCenterAnchor(anchor AnchorPosition,
	dim ComponentDimensions, defaultY int) (int, int) {
	centerX := (lc.terminalWidth - dim.Width) / 2
	switch anchor {
	case AnchorTop:
		return centerX, 0
	case AnchorCenter:
		return centerX, (lc.terminalHeight - dim.Height) / 2
	case AnchorBottom:
		return centerX, lc.terminalHeight - dim.Height
	default:
		return centerX, defaultY
	}
}

// calculateRightAnchor calculates position for right-aligned anchors
func (lc *LayoutCalculator) calculateRightAnchor(anchor AnchorPosition,
	dim ComponentDimensions, defaultY int) (int, int) {
	rightX := lc.terminalWidth - dim.Width
	switch anchor {
	case AnchorTopRight:
		return rightX, 0
	case AnchorRight:
		return rightX, (lc.terminalHeight - dim.Height) / 2
	case AnchorBottomRight:
		return rightX, lc.terminalHeight - dim.Height
	default:
		return rightX, defaultY
	}
}

// hasSpecificPositioning checks if a component has specific positioning constraints
func (lc *LayoutCalculator) hasSpecificPositioning(constraints ConstraintSet) bool {
	// Check for anchor constraints
	if _, exists := constraints.Get(ConstraintAnchor); exists {
		return true
	}

	// Check for relationship constraints
	for _, constraint := range constraints.All() {
		if _, ok := constraint.(RelationshipConstraint); ok {
			return true
		}
	}

	return false
}

// LayoutAlgorithm represents different layout algorithms
type LayoutAlgorithm int

const (
	// AlgorithmVerticalStack stacks components vertically
	AlgorithmVerticalStack LayoutAlgorithm = iota
	// AlgorithmHorizontalStack stacks components horizontally
	AlgorithmHorizontalStack
	// AlgorithmGrid arranges components in a grid
	AlgorithmGrid
	// AlgorithmAbsolute allows absolute positioning
	AlgorithmAbsolute
)

// AdvancedLayoutCalculator provides more sophisticated layout algorithms
type AdvancedLayoutCalculator struct {
	*LayoutCalculator
	algorithm LayoutAlgorithm
}

// NewAdvancedLayoutCalculator creates a new advanced layout calculator
func NewAdvancedLayoutCalculator(terminalWidth, terminalHeight, spacing int,
	algorithm LayoutAlgorithm) *AdvancedLayoutCalculator {
	return &AdvancedLayoutCalculator{
		LayoutCalculator: NewLayoutCalculator(terminalWidth, terminalHeight, spacing),
		algorithm:        algorithm,
	}
}

// Calculate performs advanced layout calculation based on the selected algorithm
func (alc *AdvancedLayoutCalculator) Calculate() (*LayoutResult, error) {
	switch alc.algorithm {
	case AlgorithmVerticalStack:
		return alc.calculateVerticalStack()
	case AlgorithmHorizontalStack:
		return alc.calculateHorizontalStack()
	case AlgorithmGrid:
		return alc.calculateGrid()
	case AlgorithmAbsolute:
		return alc.calculateAbsolute()
	default:
		return alc.LayoutCalculator.Calculate()
	}
}

// calculateVerticalStack implements vertical stacking algorithm
func (alc *AdvancedLayoutCalculator) calculateVerticalStack() (*LayoutResult, error) {
	// Use the base implementation for now
	return alc.LayoutCalculator.Calculate()
}

// calculateHorizontalStack implements horizontal stacking algorithm
func (alc *AdvancedLayoutCalculator) calculateHorizontalStack() (*LayoutResult, error) {
	// TODO: Implement horizontal stacking
	return nil, fmt.Errorf("horizontal stacking not yet implemented")
}

// calculateGrid implements grid-based layout algorithm
func (alc *AdvancedLayoutCalculator) calculateGrid() (*LayoutResult, error) {
	// TODO: Implement grid layout
	return nil, fmt.Errorf("grid layout not yet implemented")
}

// calculateAbsolute implements absolute positioning algorithm
func (alc *AdvancedLayoutCalculator) calculateAbsolute() (*LayoutResult, error) {
	// TODO: Implement absolute positioning
	return nil, fmt.Errorf("absolute positioning not yet implemented")
}

// Helper functions for layout calculations

// DistributeSpace distributes available space among flexible components
func DistributeSpace(availableSpace int, flexComponents []FlexComponent) []int {
	if len(flexComponents) == 0 {
		return []int{}
	}

	totalWeight := 0.0
	for _, comp := range flexComponents {
		totalWeight += comp.Weight
	}

	if totalWeight == 0 {
		// Equal distribution if no weights
		size := availableSpace / len(flexComponents)
		result := make([]int, len(flexComponents))
		for i := range result {
			result[i] = size
		}
		return result
	}

	// Weighted distribution
	result := make([]int, len(flexComponents))
	for i, comp := range flexComponents {
		result[i] = int(math.Round(float64(availableSpace) * (comp.Weight / totalWeight)))
	}

	return result
}

// FlexComponent represents a flexible component with weight
type FlexComponent struct {
	ID     string
	Weight float64
}

// SortComponentsByZIndex sorts components by their z-index for proper rendering order
func SortComponentsByZIndex(components []ComponentLayoutInfo) []ComponentLayoutInfo {
	// For now, just return the original order
	// TODO: Implement z-index sorting when z-index constraints are added
	return components
}

// CheckOverlaps checks for overlapping components and returns warnings
func CheckOverlaps(layouts map[string]ComponentLayout) []string {
	warnings := []string{}

	// Convert to slice for easier iteration
	components := make([]struct {
		ID     string
		Layout ComponentLayout
	}, 0, len(layouts))

	for id, layout := range layouts {
		components = append(components, struct {
			ID     string
			Layout ComponentLayout
		}{ID: id, Layout: layout})
	}

	// Check each pair of components
	for i := 0; i < len(components); i++ {
		for j := i + 1; j < len(components); j++ {
			comp1 := components[i]
			comp2 := components[j]

			// Check if rectangles overlap
			if rectanglesOverlap(comp1.Layout, comp2.Layout) {
				warnings = append(warnings,
					fmt.Sprintf("Components '%s' and '%s' overlap", comp1.ID, comp2.ID))
			}
		}
	}

	return warnings
}

// rectanglesOverlap checks if two rectangles overlap
func rectanglesOverlap(rect1, rect2 ComponentLayout) bool {
	// No overlap if one rectangle is to the left of the other
	if rect1.X >= rect2.X+rect2.Width || rect2.X >= rect1.X+rect1.Width {
		return false
	}

	// No overlap if one rectangle is above the other
	if rect1.Y >= rect2.Y+rect2.Height || rect2.Y >= rect1.Y+rect1.Height {
		return false
	}

	return true
}

package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// SpacingCalculator provides Lipgloss-backed spacing calculations for the layout engine
type SpacingCalculator struct {
	// Internal Lipgloss style for calculations
	style lipgloss.Style
}

// NewSpacingCalculator creates a new spacing calculator
func NewSpacingCalculator() *SpacingCalculator {
	return &SpacingCalculator{
		style: lipgloss.NewStyle(),
	}
}

// CalculateMargins calculates margin spacing using Lipgloss logic
func (sc *SpacingCalculator) CalculateMargins(top, right, bottom, left int) (int, int, int, int) {
	// Use Lipgloss style to validate and normalize margin values
	tempStyle := sc.style.Margin(top, right, bottom, left)
	return tempStyle.GetMargin()
}

// CalculatePadding calculates padding spacing using Lipgloss logic
func (sc *SpacingCalculator) CalculatePadding(top, right, bottom, left int) (int, int, int, int) {
	// Use Lipgloss style to validate and normalize padding values
	tempStyle := sc.style.Padding(top, right, bottom, left)
	return tempStyle.GetPadding()
}

// GetContentArea calculates the content area after applying margins and padding
func (sc *SpacingCalculator) GetContentArea(containerWidth, containerHeight int,
	marginTop, marginRight, marginBottom, marginLeft int,
	paddingTop, paddingRight, paddingBottom, paddingLeft int) (width, height int) {
	// Create a style with the given margins and padding
	tempStyle := sc.style.
		Margin(marginTop, marginRight, marginBottom, marginLeft).
		Padding(paddingTop, paddingRight, paddingBottom, paddingLeft)

	// Use Lipgloss GetFrameSize to calculate total frame overhead
	frameWidth, frameHeight := tempStyle.GetFrameSize()

	// Calculate content area
	contentWidth := containerWidth - frameWidth
	contentHeight := containerHeight - frameHeight

	// Ensure minimum content area
	if contentWidth < 1 {
		contentWidth = 1
	}
	if contentHeight < 1 {
		contentHeight = 1
	}

	return contentWidth, contentHeight
}

// GetHorizontalSpacing calculates total horizontal spacing (margins + padding)
func (sc *SpacingCalculator) GetHorizontalSpacing(marginLeft, marginRight, paddingLeft, paddingRight int) int {
	return marginLeft + marginRight + paddingLeft + paddingRight
}

// GetVerticalSpacing calculates total vertical spacing (margins + padding)
func (sc *SpacingCalculator) GetVerticalSpacing(marginTop, marginBottom, paddingTop, paddingBottom int) int {
	return marginTop + marginBottom + paddingTop + paddingBottom
}

// ValidateSpacing validates spacing values using Lipgloss constraints
func (sc *SpacingCalculator) ValidateSpacing(top, right, bottom, left int) error {
	// Create a temporary style to validate spacing values
	tempStyle := sc.style.Margin(top, right, bottom, left)

	// Lipgloss will normalize negative values to 0, so we can check if values changed
	actualTop, actualRight, actualBottom, actualLeft := tempStyle.GetMargin()

	// For now, we'll be permissive and let Lipgloss handle normalization
	// In the future, we could return an error here if strict validation is needed
	_ = actualTop
	_ = actualRight
	_ = actualBottom
	_ = actualLeft

	return nil
}

// CreateLipglossStyle creates a Lipgloss style from layout constraints
func (sc *SpacingCalculator) CreateLipglossStyle(constraints ConstraintSet) lipgloss.Style {
	style := lipgloss.NewStyle()

	style = sc.applyMarginConstraints(style, constraints)
	style = sc.applyPaddingConstraints(style, constraints)
	style = sc.applyWidthConstraints(style, constraints)
	style = sc.applyHeightConstraints(style, constraints)

	return style
}

// applyMarginConstraints applies margin constraints to a style
func (sc *SpacingCalculator) applyMarginConstraints(style lipgloss.Style, constraints ConstraintSet) lipgloss.Style {
	if marginConstraint, exists := constraints.Get(ConstraintMargin); exists {
		if spacingConstraint, ok := marginConstraint.(SpacingConstraint); ok {
			top, right, bottom, left := spacingConstraint.Values()
			style = style.Margin(top, right, bottom, left)
		}
	}
	return style
}

// applyPaddingConstraints applies padding constraints to a style
func (sc *SpacingCalculator) applyPaddingConstraints(style lipgloss.Style, constraints ConstraintSet) lipgloss.Style {
	if paddingConstraint, exists := constraints.Get(ConstraintPadding); exists {
		if spacingConstraint, ok := paddingConstraint.(SpacingConstraint); ok {
			top, right, bottom, left := spacingConstraint.Values()
			style = style.Padding(top, right, bottom, left)
		}
	}
	return style
}

// applyWidthConstraints applies width constraints to a style
func (sc *SpacingCalculator) applyWidthConstraints(style lipgloss.Style, constraints ConstraintSet) lipgloss.Style {
	if widthConstraint, exists := constraints.Get(ConstraintWidth); exists {
		if sizeConstraint, ok := widthConstraint.(SizeConstraint); ok {
			// For fixed sizes, we can set the width directly
			if fixedSize, ok := sizeConstraint.Value().(FixedSize); ok {
				// Note: We need terminal width context for accurate calculation
				// This is a simplified approach - ideally we'd pass terminal dimensions
				width := fixedSize.Calculate(100) // Placeholder terminal width
				style = style.Width(width)
			}
		}
	}
	return style
}

// applyHeightConstraints applies height constraints to a style
func (sc *SpacingCalculator) applyHeightConstraints(style lipgloss.Style, constraints ConstraintSet) lipgloss.Style {
	if heightConstraint, exists := constraints.Get(ConstraintHeight); exists {
		if sizeConstraint, ok := heightConstraint.(SizeConstraint); ok {
			// For fixed sizes, we can set the height directly
			if fixedSize, ok := sizeConstraint.Value().(FixedSize); ok {
				// Note: We need terminal height context for accurate calculation
				height := fixedSize.Calculate(30) // Placeholder terminal height
				style = style.Height(height)
			}
		}
	}
	return style
}

// ApplySpacingToStyle applies spacing constraints to an existing Lipgloss style
func (sc *SpacingCalculator) ApplySpacingToStyle(style lipgloss.Style, constraints ConstraintSet) lipgloss.Style {
	// Apply margin constraints
	if marginConstraint, exists := constraints.Get(ConstraintMargin); exists {
		if spacingConstraint, ok := marginConstraint.(SpacingConstraint); ok {
			top, right, bottom, left := spacingConstraint.Values()
			style = style.Margin(top, right, bottom, left)
		}
	}

	// Apply padding constraints
	if paddingConstraint, exists := constraints.Get(ConstraintPadding); exists {
		if spacingConstraint, ok := paddingConstraint.(SpacingConstraint); ok {
			top, right, bottom, left := spacingConstraint.Values()
			style = style.Padding(top, right, bottom, left)
		}
	}

	return style
}

// CalculateFrameSize calculates the total frame size (borders + margins + padding) for a style
func (sc *SpacingCalculator) CalculateFrameSize(style lipgloss.Style) (width, height int) {
	return style.GetFrameSize()
}

// DefaultSpacingCalculator is the global spacing calculator instance for convenient access
var DefaultSpacingCalculator = NewSpacingCalculator()

// Convenience functions that use the default calculator

// CalculateMargins calculates margin spacing using the default calculator
func CalculateMargins(top, right, bottom, left int) (int, int, int, int) {
	return DefaultSpacingCalculator.CalculateMargins(top, right, bottom, left)
}

// CalculatePadding calculates padding spacing using the default calculator
func CalculatePadding(top, right, bottom, left int) (int, int, int, int) {
	return DefaultSpacingCalculator.CalculatePadding(top, right, bottom, left)
}

// GetContentArea calculates content area using the default calculator
func GetContentArea(containerWidth, containerHeight int,
	marginTop, marginRight, marginBottom, marginLeft int,
	paddingTop, paddingRight, paddingBottom, paddingLeft int) (width, height int) {
	return DefaultSpacingCalculator.GetContentArea(containerWidth, containerHeight,
		marginTop, marginRight, marginBottom, marginLeft,
		paddingTop, paddingRight, paddingBottom, paddingLeft)
}

// CreateLipglossStyleFromConstraints creates a Lipgloss style from constraints using the default calculator
func CreateLipglossStyleFromConstraints(constraints ConstraintSet) lipgloss.Style {
	return DefaultSpacingCalculator.CreateLipglossStyle(constraints)
}

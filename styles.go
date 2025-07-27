package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Layout constants for consistent spacing and sizing
const (
	ColumnSplitRatio = 0.7 // 70% left column, 30% right column
)

// Essential styles that are still used by legacy code
var (
	// Base style for text styling
	baseTextStyle = lipgloss.NewStyle().Bold(true)

	// Level styles used by delegate.go
	levelUserStyle  = baseTextStyle.Foreground(lipgloss.Color("#10B981")) // Green
	levelRepoStyle  = baseTextStyle.Foreground(lipgloss.Color("#3B82F6")) // Blue
	levelLocalStyle = baseTextStyle.Foreground(lipgloss.Color("#8B5CF6")) // Purple

	// Column styles used by delegate.go
	leftColumnStyle  = lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left)
	rightColumnStyle = lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Right)

	// Highlight style for search matches used by delegate.go
	highlightedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FCD34D")). // Bright yellow
				Foreground(lipgloss.Color("#1F2937"))  // Dark text

	// Arrow style for move operations
	moveArrowStyle = CreateLevelStyle("9") // Orange
)

// CreateLevelStyle creates a style for different permission levels
func CreateLevelStyle(color string) lipgloss.Style {
	return baseTextStyle.Foreground(lipgloss.Color(color))
}

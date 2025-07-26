package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	// Base styles for common patterns
	baseTextStyle = lipgloss.NewStyle().
			Bold(true)

	// Color scheme foundation using inheritance
	headerTextStyle = baseTextStyle.
			Foreground(lipgloss.Color("15")) // Bright white text

	statusGoodStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#059669", // Dark green for light themes
				Dark:  "#10B981", // Bright green for dark themes
			})

	statusBadStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#DC2626", // Dark red for light themes
				Dark:  "#EF4444", // Bright red for dark themes
			})

	accentStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#0EA5E9", // Blue for light themes
				Dark:  "#38BDF8", // Cyan for dark themes
			})

	countStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // Yellow for counts

	// Column header styling (consistent between panels)
	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // Bright white text
				Background(lipgloss.Color("8")).  // Dark gray background
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 1, 0)

	// Base dark theme style
	baseDarkStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("0")). // Dark background
			Padding(0, 1)

	// Updated header style with dark theme using inheritance
	headerStyle = baseDarkStyle.
			Margin(0, 0, 1, 0) // Add bottom margin to ensure separation

	// Header with border variant for enhanced visual separation
	headerWithBorderStyle = headerStyle.
			BorderTop(true).
			BorderStyle(lipgloss.DoubleBorder())

	// Base panel style with common properties
	basePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	// Panel style variations using inheritance
	panelStyle = basePanelStyle.
			BorderForeground(lipgloss.Color("8")) // Gray

	activePanelStyle = basePanelStyle.
			BorderForeground(lipgloss.AdaptiveColor{
				Light: "#0EA5E9", // Blue for light themes
				Dark:  "#38BDF8", // Cyan for dark themes
			})

	// selectedStyle = lipgloss.NewStyle().
	// 		Background(lipgloss.Color("3")). // Yellow
	// 		Foreground(lipgloss.Color("0"))  // Black

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")) // Cyan

	// Level styles with adaptive colors for better theme compatibility
	levelUserStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#059669", // Dark green for light themes
				Dark:  "#10B981", // Bright green for dark themes
			})

	levelRepoStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#2563EB", // Dark blue for light themes
				Dark:  "#3B82F6", // Bright blue for dark themes
			})

	levelLocalStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#7C3AED", // Dark purple for light themes
				Dark:  "#8B5CF6", // Bright purple for dark themes
			})

	moveArrowStyle = CreateLevelStyle("9") // Orange

	// duplicateStyle = lipgloss.NewStyle().
	// 		Foreground(lipgloss.Color("9")) // Red

	// Updated footer style with dark theme using inheritance
	footerStyle = baseDarkStyle
	// Uses base dark style without additional margin to save space

	// Highlight style for search matches with adaptive colors
	highlightedItemStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{
				Light: "#FEF3C7", // Light yellow for light themes
				Dark:  "#FCD34D", // Bright yellow for dark themes
			}).
			Foreground(lipgloss.AdaptiveColor{
				Light: "#92400E", // Dark orange text for light themes
				Dark:  "#1F2937", // Dark text for dark themes
			})

	// Status message style for UI notifications with adaptive colors
	statusMessageStyle = baseTextStyle.
			Foreground(lipgloss.AdaptiveColor{
				Light: "#059669", // Dark green for light themes
				Dark:  "#10B981", // Bright green for dark themes
			}).
			Padding(0, 1).
			Margin(1, 0)

	// Dialog style for centered confirmations
	dialogStyle = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center)

	// Base column style with common properties
	baseColumnStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Column styles for two-column layouts using inheritance
	leftColumnStyle = baseColumnStyle.
			AlignHorizontal(lipgloss.Left)

	rightColumnStyle = baseColumnStyle.
			AlignHorizontal(lipgloss.Right)
)

// Layout constants for consistent spacing and sizing
const (
	// Column layout ratios
	ColumnSplitRatio = 0.7 // 70% left column, 30% right column

	// Minimum content dimensions
	MinContentWidth  = 40
	MinListHeight    = 4  // Minimum height for permission list
	MinTableHeight   = 2  // Minimum height for duplicates table
)

// Style utility functions for dynamic style creation

// CreateLevelStyle creates a style for different permission levels
func CreateLevelStyle(color string) lipgloss.Style {
	return baseTextStyle.Foreground(lipgloss.Color(color))
}

// CreatePanelVariant creates a panel style variant with custom border color
func CreatePanelVariant(borderColor string) lipgloss.Style {
	return basePanelStyle.BorderForeground(lipgloss.Color(borderColor))
}

// CreateColumnVariant creates a column style variant with custom width and alignment
func CreateColumnVariant(alignment lipgloss.Position, width int) lipgloss.Style {
	return baseColumnStyle.AlignHorizontal(alignment).Width(width)
}

// Key bindings
var keyMap = struct {
	up, down, pageUp, pageDown, tab key.Binding
	space, selectAll                key.Binding
	moveUser, moveRepo, moveLocal   key.Binding
	edit, enter, clear              key.Binding
	escape, quit                    key.Binding
}{
	up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	pageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
	pageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
	tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),

	space:     key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	selectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle all")),

	moveUser:  key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "move to user")),
	moveRepo:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "move to repo")),
	moveLocal: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "move to local")),

	edit:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	clear: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "clear queue")),

	escape: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

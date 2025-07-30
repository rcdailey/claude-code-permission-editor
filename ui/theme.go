package ui

import "github.com/charmbracelet/lipgloss/v2"

// Application color palette - centralized theme constants
const (
	// Primary colors
	ColorAccent     = "#38BDF8" // Cyan - for interactive elements, focused borders
	ColorTitle      = "15"      // Bright white - for titles and headers
	ColorText       = "7"       // Light gray - for normal text
	ColorBackground = "0"       // Black - for backgrounds

	// Status colors
	ColorSuccess = "#10B981" // Green - for success states, user level
	ColorError   = "#EF4444" // Red - for error states, missing files
	ColorWarning = "#FbbF24" // Amber - for warnings, local level
	ColorInfo    = "#38BDF8" // Cyan - for info, repo level
	ColorCount   = "11"      // Yellow - for count displays

	// UI element colors
	ColorBorderNormal        = "8" // Gray - for unfocused borders
	ColorBorderFocused       = ColorAccent
	ColorBackgroundSecondary = "240" // Dark gray - for status bars, selections, interactive backgrounds
	ColorTextSecondary       = "244" // Lighter gray - for secondary text, indicators
)

// Pre-configured styles for common UI patterns
var (
	// Base styles
	AccentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccent)).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitle)).
			Bold(true)

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	CountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCount))

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo)).
			Bold(true)

	// Border styles
	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorderFocused))

	NormalBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorderNormal))

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorBackgroundSecondary)).
			Padding(0, 1)

	// Selection highlighting styles for currently selected item
	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorBackgroundSecondary)).
				Foreground(lipgloss.Color(ColorAccent)).
				Bold(true).
				Padding(0, 1)

	// Origin indicator styles for moved permissions
	OriginIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorTextSecondary)).
				Italic(true)
)

// Level-specific styles for consistent color coding
var (
	LocalLevelStyle = WarningStyle // Amber for Local
	RepoLevelStyle  = InfoStyle    // Cyan for Repo
	UserLevelStyle  = SuccessStyle // Green for User
)

// Darker level styles for origin indicators to match gray text contrast
var (
	LocalOriginStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("136")).
				Italic(true)
	RepoOriginStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("31")).
			Italic(true)
	UserOriginStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")).
			Italic(true)
)

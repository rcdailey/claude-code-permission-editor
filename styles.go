package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")). // Bright White
			Background(lipgloss.Color("4")).  // Blue
			Padding(0, 1)
		// Removed margins to save space

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")). // Gray
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("14")). // Cyan
				Padding(0, 1)

	// selectedStyle = lipgloss.NewStyle().
	// 		Background(lipgloss.Color("3")). // Yellow
	// 		Foreground(lipgloss.Color("0"))  // Black

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")) // Cyan

	levelUserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")) // Green

	levelRepoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")) // Blue

	levelLocalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")) // Purple

	moveArrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")) // Orange

	// duplicateStyle = lipgloss.NewStyle().
	// 		Foreground(lipgloss.Color("9")) // Red

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")). // Black text for better contrast
			Background(lipgloss.Color("7")). // Light gray background
			Padding(0, 1)
	// Removed margin to save space
)

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

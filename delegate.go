package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PermissionDelegate handles rendering of permission items in the list
type PermissionDelegate struct{}

// Height returns the height of each permission item in the list.
func (d PermissionDelegate) Height() int { return 1 }

// Spacing returns the spacing between permission items in the list.
func (d PermissionDelegate) Spacing() int { return 0 }

// Update handles messages for the permission delegate.
func (d PermissionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders a permission item in the list with styling and selection indicators.
func (d PermissionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	perm, ok := listItem.(Permission)
	if !ok {
		return
	}

	// Selection indicator
	selected := " "
	if perm.Selected {
		selected = "x"
	}

	// Cursor indicator - use list's built-in cursor tracking
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}

	// Level styling
	levelStyle := levelLocalStyle
	switch perm.CurrentLevel {
	case "User":
		levelStyle = levelUserStyle
	case "Repo":
		levelStyle = levelRepoStyle
	}

	// Pending move arrow
	moveArrow := ""
	if perm.PendingMove != "" {
		moveArrow = moveArrowStyle.Render(fmt.Sprintf(" → [%s]", perm.PendingMove))
	}

	// Use built-in list highlighting for filter matches
	permName := perm.Name
	if matches := m.MatchesForItem(index); len(matches) > 0 {
		// Apply highlighting to matched characters
		highlightStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("11")). // Yellow background
			Foreground(lipgloss.Color("0"))   // Black text

		permName = highlightMatches(perm.Name, matches, highlightStyle)
	}

	// Create two-column layout
	leftColumn := fmt.Sprintf("%s[%s] %s", cursor, selected, permName)
	rightColumn := levelStyle.Render(fmt.Sprintf("[%s]", perm.CurrentLevel)) + moveArrow

	// Calculate column widths based on list width
	totalWidth := m.Width()
	leftWidth := int(float64(totalWidth) * 0.7)
	rightWidth := totalWidth - leftWidth

	// Create columns with proper alignment
	leftColumnStyled := lipgloss.NewStyle().
		Width(leftWidth).
		AlignHorizontal(lipgloss.Left).
		Render(leftColumn)

	rightColumnStyled := lipgloss.NewStyle().
		Width(rightWidth).
		AlignHorizontal(lipgloss.Right).
		Render(rightColumn)

	line := lipgloss.JoinHorizontal(lipgloss.Top, leftColumnStyled, rightColumnStyled)

	// Apply cursor highlighting if this item is selected
	if index == m.Index() {
		line = cursorStyle.Render(line)
	}

	_, _ = fmt.Fprint(w, line)
}

// highlightMatches highlights character positions based on list filter matches
func highlightMatches(text string, matches []int, highlightStyle lipgloss.Style) string {
	if len(matches) == 0 {
		return text
	}

	// Convert string to runes for proper unicode handling
	runes := []rune(text)
	result := strings.Builder{}

	matchSet := make(map[int]bool)
	for _, pos := range matches {
		if pos >= 0 && pos < len(runes) {
			matchSet[pos] = true
		}
	}

	for i, r := range runes {
		if matchSet[i] {
			result.WriteString(highlightStyle.Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

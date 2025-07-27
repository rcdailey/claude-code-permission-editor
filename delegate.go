package main

import (
	"fmt"
	"io"
	"strings"

	"claude-permissions/types"

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
	perm, ok := listItem.(types.Permission)
	if !ok {
		return
	}

	leftColumn := d.formatLeftColumn(perm, m, index)
	rightColumn := d.formatRightColumn(perm)
	output := d.formatTwoColumnLayout(leftColumn, rightColumn, m.Width())

	_, _ = fmt.Fprint(w, output)
}

// formatLeftColumn creates the left column content with cursor and selection indicators
func (d PermissionDelegate) formatLeftColumn(
	perm types.Permission,
	m list.Model,
	index int,
) string {
	selected := " "
	if perm.Selected {
		selected = "x"
	}

	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}

	permName := perm.Name
	if matches := m.MatchesForItem(index); len(matches) > 0 {
		permName = highlightMatches(perm.Name, matches, highlightedItemStyle)
	}

	return fmt.Sprintf("%s[%s] %s", cursor, selected, permName)
}

// formatRightColumn creates the right column with level and move indicators
func (d PermissionDelegate) formatRightColumn(perm types.Permission) string {
	levelStyle := d.getLevelStyle(perm.CurrentLevel)
	rightColumn := levelStyle.Render(fmt.Sprintf("[%s]", perm.CurrentLevel))

	if perm.PendingMove != "" {
		rightColumn += moveArrowStyle.Render(fmt.Sprintf(" → [%s]", perm.PendingMove))
	}

	return rightColumn
}

// getLevelStyle returns the appropriate style for a permission level
func (d PermissionDelegate) getLevelStyle(level string) lipgloss.Style {
	switch level {
	case "User":
		return levelUserStyle
	case "Repo":
		return levelRepoStyle
	default:
		return levelLocalStyle
	}
}

// formatTwoColumnLayout creates a two-column layout for the permission item
func (d PermissionDelegate) formatTwoColumnLayout(
	leftColumn, rightColumn string,
	totalWidth int,
) string {
	// Account for panel border/padding overhead
	contentWidth := totalWidth - 4
	if contentWidth < 40 {
		contentWidth = 40
	}

	leftWidth := int(float64(contentWidth) * ColumnSplitRatio)

	leftColumnFormatted := leftColumnStyle.Render(leftColumn)
	rightColumnFormatted := rightColumnStyle.Render(rightColumn)

	paddingNeeded := leftWidth - lipgloss.Width(leftColumnFormatted)
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}

	return leftColumnFormatted + strings.Repeat(" ", paddingNeeded) + rightColumnFormatted
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

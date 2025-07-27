package ui

import (
	"fmt"
	"strings"

	"claude-permissions/types"

	"github.com/charmbracelet/lipgloss"
)

// HeaderComponent represents the top header section
type HeaderComponent struct {
	width   int
	content string
}

// NewHeaderComponent creates a new header component
func NewHeaderComponent(width int) *HeaderComponent {
	return &HeaderComponent{
		width: width,
	}
}

// SetContent updates the header content
func (h *HeaderComponent) SetContent(content string) {
	h.content = content
}

// View renders the header preserving the enhanced styling from renderHeaderContent
func (h *HeaderComponent) View() string {
	if h.content == "" {
		return ""
	}
	// Return content as-is since renderHeaderContent already applies styling
	// Just ensure consistent width
	style := lipgloss.NewStyle().Width(h.width)
	return style.Render(h.content)
}

// FooterComponent represents the bottom footer section
type FooterComponent struct {
	width   int
	content string
}

// NewFooterComponent creates a new footer component
func NewFooterComponent(width int) *FooterComponent {
	return &FooterComponent{
		width: width,
	}
}

// SetContent updates the footer content
func (f *FooterComponent) SetContent(content string) {
	f.content = content
}

// View renders the footer preserving the enhanced styling from renderFooterContent
func (f *FooterComponent) View() string {
	if f.content == "" {
		return ""
	}
	// Return content as-is since renderFooterContent already applies styling
	// Just ensure consistent width
	style := lipgloss.NewStyle().Width(f.width)
	return style.Render(f.content)
}

// ContentComponent represents the main content area
type ContentComponent struct {
	width  int
	height int
	model  *types.Model
}

// NewContentComponent creates a new content component
func NewContentComponent(width, height int, model *types.Model) *ContentComponent {
	return &ContentComponent{
		width:  width,
		height: height,
		model:  model,
	}
}

// SetDimensions updates the content dimensions
func (c *ContentComponent) SetDimensions(width, height int) {
	c.width = width
	c.height = height
}

// View renders the appropriate content based on current screen
func (c *ContentComponent) View() string {
	switch c.model.CurrentScreen {
	case types.ScreenDuplicates:
		return c.renderDuplicatesContent()
	case types.ScreenOrganization:
		return c.renderOrganizationContent()
	default:
		return c.renderDuplicatesContent()
	}
}

// renderDuplicatesContent renders the duplicates screen content
func (c *ContentComponent) renderDuplicatesContent() string {
	if c.width <= 0 || c.height <= 0 {
		return ""
	}

	// Render the actual duplicates table with proper width accounting for border/padding
	// Calculate content width by subtracting border and padding overhead
	borderPaddingOverhead := 4 // 2 for border + 2 for padding (left+right)
	contentWidth := c.width - borderPaddingOverhead
	if contentWidth < 20 { // Minimum usable width
		contentWidth = 20
	}

	tableStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(c.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorderFocused)). // Use centralized theme
		Padding(1)

	if len(c.model.Duplicates) == 0 {
		emptyMessage := "No duplicate permissions found across levels"
		return tableStyle.Render(emptyMessage)
	}

	// Use the actual duplicates table from the model
	tableContent := c.model.DuplicatesTable.View()
	return tableStyle.Render(tableContent)
}

// renderOrganizationContent renders the three-column organization screen
func (c *ContentComponent) renderOrganizationContent() string {
	if c.width <= 0 || c.height <= 0 {
		return ""
	}

	// Calculate column width accounting for border and padding overhead
	// Each column has border (2 chars) + padding (2 chars) = 4 chars overhead
	borderPaddingPerColumn := 4
	totalOverhead := borderPaddingPerColumn * 3 // 3 columns
	availableContentWidth := c.width - totalOverhead
	columnWidth := availableContentWidth / 3

	// Render each column
	localColumn := c.renderPermissionColumn("Local", columnWidth, 0)
	repoColumn := c.renderPermissionColumn("Repo", columnWidth, 1)
	userColumn := c.renderPermissionColumn("User", columnWidth, 2)

	// Join horizontally using pure lipgloss
	return lipgloss.JoinHorizontal(lipgloss.Top, localColumn, repoColumn, userColumn)
}

// renderPermissionColumn renders a single permission column
func (c *ContentComponent) renderPermissionColumn(level string, width int, columnIndex int) string {
	focused := c.model.FocusedColumn == columnIndex
	style := c.getColumnStyle(focused, width)
	header := c.renderColumnHeader(level)
	content := c.renderColumnContent(level, columnIndex, focused)
	columnContent := lipgloss.JoinVertical(lipgloss.Left, header, "", content)
	return style.Render(columnContent)
}

// getColumnStyle returns the appropriate style for focused/unfocused columns
func (c *ContentComponent) getColumnStyle(focused bool, width int) lipgloss.Style {
	if focused {
		return FocusedBorderStyle.Width(width).Height(c.height).Padding(1)
	}
	return NormalBorderStyle.Width(width).Height(c.height).Padding(1)
}

// renderColumnHeader creates the styled header for a column
func (c *ContentComponent) renderColumnHeader(level string) string {
	var headerStyle lipgloss.Style
	var count int

	switch level {
	case "Local":
		count = len(c.model.LocalLevel.Permissions)
		headerStyle = LocalLevelStyle.
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 1).
			Margin(0, 0, 1, 0)
	case "Repo":
		count = len(c.model.RepoLevel.Permissions)
		headerStyle = RepoLevelStyle.
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 1).
			Margin(0, 0, 1, 0)
	case "User":
		count = len(c.model.UserLevel.Permissions)
		headerStyle = UserLevelStyle.
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 1).
			Margin(0, 0, 1, 0)
	}

	headerText := level + " " + CountStyle.Render(fmt.Sprintf("(%d)", count))
	return headerStyle.Render(headerText)
}

// renderColumnContent creates the content for a column
func (c *ContentComponent) renderColumnContent(level string, columnIndex int, focused bool) string {
	levelPermissions := c.getLevelPermissions(level)

	var permissionItems []string
	if len(levelPermissions) == 0 {
		permissionItems = append(permissionItems, "No permissions")
	} else {
		for i, perm := range levelPermissions {
			prefix := " "
			if focused && i == c.model.ColumnSelections[columnIndex] {
				prefix = ">"
			}
			permissionItems = append(permissionItems, prefix+" "+perm)
		}
	}

	return strings.Join(permissionItems, "\n")
}

// getLevelPermissions returns permissions for the specified level
func (c *ContentComponent) getLevelPermissions(level string) []string {
	switch level {
	case "Local":
		return c.model.LocalLevel.Permissions
	case "Repo":
		return c.model.RepoLevel.Permissions
	case "User":
		return c.model.UserLevel.Permissions
	default:
		return []string{}
	}
}

package ui

import (
	"fmt"
	"strings"

	"claude-permissions/types"

	"github.com/charmbracelet/lipgloss/v2"
)

// Level display constants to avoid goconst warnings
const (
	levelDisplayLocal = "Local"
	levelDisplayRepo  = "Repo"
	levelDisplayUser  = "User"
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

// Layout constants for consistent width calculations across all screens
const (
	// ContentWidthBuffer is the buffer space subtracted from terminal width
	// to ensure all content fits within terminal boundaries consistently.
	//
	// This value is critical for visual consistency when switching between screens.
	// All screens must use getConsistentContentWidth() to ensure their rightmost
	// borders align at the same position, preventing visual "jumps" during navigation.
	//
	// Value determined through testing to provide optimal balance between
	// maximizing usable space and preventing terminal overflow.
	ContentWidthBuffer = 0
)

// NewContentComponent creates a new content component
func NewContentComponent(width, height int, model *types.Model) *ContentComponent {
	return &ContentComponent{
		width:  width,
		height: height,
		model:  model,
	}
}

// getConsistentContentWidth returns the standardized content width used across all screens
// This ensures visual consistency when switching between screens with TAB
func (c *ContentComponent) getConsistentContentWidth() int {
	return c.width - ContentWidthBuffer
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

	// Use centralized width calculation for consistency across all screens
	contentWidth := c.getConsistentContentWidth()
	if contentWidth < 20 { // Minimum usable width
		contentWidth = 20
	}

	if len(c.model.Duplicates) == 0 {
		emptyMessage := "No duplicate permissions found across levels"
		return BlockingMessageStyle.
			Width(contentWidth).
			Height(c.height).
			Render(emptyMessage)
	}

	tableStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(c.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorderFocused)). // Use centralized theme
		Padding(1)

	// Use the actual duplicates table from the model
	tableContent := c.model.DuplicatesTable.View()
	return tableStyle.Render(tableContent)
}

// renderOrganizationContent renders the three-column organization screen or blocking message
func (c *ContentComponent) renderOrganizationContent() string {
	if c.width <= 0 || c.height <= 0 {
		return ""
	}

	// Check if there are unresolved duplicates - if so, show blocking message
	if hasUnresolvedDuplicates(c.model) {
		return c.renderBlockingMessage()
	}

	// Use centralized width calculation and divide among columns
	totalContentWidth := c.getConsistentContentWidth()
	baseColumnWidth := totalContentWidth / 3
	remainder := totalContentWidth % 3

	// Distribute remainder to columns to use full width
	columnWidths := []int{baseColumnWidth, baseColumnWidth, baseColumnWidth}
	for i := 0; i < remainder; i++ {
		columnWidths[i]++
	}

	// Render each column
	localColumn := c.renderPermissionColumn(levelDisplayLocal, columnWidths[0], 0)
	repoColumn := c.renderPermissionColumn(levelDisplayRepo, columnWidths[1], 1)
	userColumn := c.renderPermissionColumn(levelDisplayUser, columnWidths[2], 2)

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
	case levelDisplayLocal:
		count = len(c.model.LocalLevel.Permissions)
		headerStyle = LocalLevelStyle.
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 1).
			Margin(0, 0, 1, 0)
	case levelDisplayRepo:
		count = len(c.model.RepoLevel.Permissions)
		headerStyle = RepoLevelStyle.
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 1).
			Margin(0, 0, 1, 0)
	case levelDisplayUser:
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
	levelPermissions := c.getColumnPermissionStructs(level)

	var permissionItems []string
	if len(levelPermissions) == 0 {
		permissionItems = append(permissionItems, "No permissions")
	} else {
		for i, perm := range levelPermissions {
			isSelected := focused && i == c.model.ColumnSelections[columnIndex]
			permItem := c.renderPermissionItem(perm, isSelected)
			permissionItems = append(permissionItems, permItem)
		}
	}

	return strings.Join(permissionItems, "\n")
}

// getColumnPermissionStructs returns Permission structs for the specified level
func (c *ContentComponent) getColumnPermissionStructs(level string) []types.Permission {
	var targetLevel string
	switch level {
	case levelDisplayLocal:
		targetLevel = types.LevelLocal
	case levelDisplayRepo:
		targetLevel = types.LevelRepo
	case levelDisplayUser:
		targetLevel = types.LevelUser
	default:
		return []types.Permission{}
	}

	var columnPerms []types.Permission
	for _, perm := range c.model.Permissions {
		if perm.CurrentLevel == targetLevel {
			columnPerms = append(columnPerms, perm)
		}
	}
	return columnPerms
}

// renderPermissionItem renders a single permission with selection highlighting and origin indicator
func (c *ContentComponent) renderPermissionItem(perm types.Permission, isSelected bool) string {
	// Build origin indicator text if moved
	var originText string
	if perm.CurrentLevel != perm.OriginalLevel {
		originStyle := c.getOriginStyle(perm.OriginalLevel)
		// Only color the level name, not the whole "(from X)" text
		coloredLevel := originStyle.Render(perm.OriginalLevel)
		originText = OriginIndicatorStyle.Render(
			" (",
		) + coloredLevel + OriginIndicatorStyle.Render(
			")",
		)
	}

	// Add selection highlighting if this item is selected
	if isSelected {
		// Highlight only the permission name, not the origin indicator
		highlightedName := SelectedItemStyle.Render("> " + perm.Name)
		return highlightedName + originText
	}

	return "  " + perm.Name + originText
}

// getOriginStyle returns the appropriate style for the origin level indicator
func (c *ContentComponent) getOriginStyle(level string) lipgloss.Style {
	switch level {
	case types.LevelLocal:
		return LocalOriginStyle
	case types.LevelRepo:
		return RepoOriginStyle
	case types.LevelUser:
		return UserOriginStyle
	default:
		return OriginIndicatorStyle
	}
}

// renderBlockingMessage renders the blocking message when duplicates need to be resolved
func (c *ContentComponent) renderBlockingMessage() string {
	contentWidth := c.getConsistentContentWidth()
	if contentWidth < 20 {
		contentWidth = 20
	}

	message := "Duplicate permissions must be resolved before organizing permissions.\n\n" +
		"Use TAB to switch to the Duplicates panel and resolve conflicts first."

	return BlockingMessageStyle.
		Width(contentWidth).
		Height(c.height).
		Render(message)
}

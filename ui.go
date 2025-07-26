package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"claude-permissions/layout"
	"claude-permissions/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the model
func Init(_ *types.Model) tea.Cmd {
	// Request initial terminal size from Bubble Tea
	return tea.WindowSize()
}

// initializeLayoutEngine creates and configures the layout engine with given dimensions
func initializeLayoutEngine(m *types.Model, width, height int) {
	// Create layout engine with the provided terminal dimensions
	m.LayoutEngine = layout.NewLayoutEngine(width, height)
	initializeLayout(m)
}

// initializeLayout sets up the layout engine with constraints
func initializeLayout(m *types.Model) {
	// Create components for two-panel layout that fills terminal width

	// Header component (fixed height, full width)
	headerComponent := layout.NewBasicComponent("header")
	headerComponent.SetConstraints(
		layout.Height(layout.Fixed(4)),
	)

	// Permissions panel (70% of remaining height, full width)
	permissionsComponent := layout.NewBasicComponent("permissions")
	permissionsComponent.SetConstraints(
		layout.Height(layout.Flex(0.7)),
	)

	// Duplicates panel (30% of remaining height, full width)
	duplicatesComponent := layout.NewBasicComponent("duplicates")
	duplicatesComponent.SetConstraints(
		layout.Height(layout.Flex(0.3)),
	)

	// Footer component (fixed height, full width, anchored to bottom)
	footerComponent := layout.NewBasicComponent("footer")
	footerComponent.SetConstraints(
		layout.Height(layout.Fixed(2)),
		layout.Anchor(layout.AnchorBottom),
	)

	// Register components with layout engine
	if err := m.LayoutEngine.AddComponentWrapper(headerComponent); err != nil {
		slog.Error("Failed to add layout component", "component", "header", "error", err)
	}
	if err := m.LayoutEngine.AddComponentWrapper(permissionsComponent); err != nil {
		slog.Error("Failed to add layout component", "component", "permissions", "error", err)
	}
	if err := m.LayoutEngine.AddComponentWrapper(duplicatesComponent); err != nil {
		slog.Error("Failed to add layout component", "component", "duplicates", "error", err)
	}
	if err := m.LayoutEngine.AddComponentWrapper(footerComponent); err != nil {
		slog.Error("Failed to add layout component", "component", "footer", "error", err)
	}

	// Apply initial layout
	if err := m.LayoutEngine.Layout(); err != nil {
		slog.Error("Failed to calculate initial layout", "error", err)
	} else {
		slog.Debug("Initial layout calculated successfully")
	}

	// Update component sizes after layout calculation
	updateComponentSizes(m)
}

// Update handles all input events
func Update(m *types.Model, msg tea.Msg) (*types.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg)
	case timer.TickMsg, timer.StartStopMsg:
		return handleTimerUpdate(m, msg)
	case timer.TimeoutMsg:
		return handleTimerTimeout(m)
	case tea.KeyMsg:
		return handleKeyPress(m, msg)
	}
	return m, nil
}

// handleWindowResize handles window resize events
func handleWindowResize(m *types.Model, msg tea.WindowSizeMsg) (*types.Model, tea.Cmd) {
	slog.Debug("Window resize event", "width", msg.Width, "height", msg.Height)

	if m.LayoutEngine == nil {
		// Create layout engine using the unified initialization function
		initializeLayoutEngine(m, msg.Width, msg.Height)
	} else {
		// Update existing layout engine size and ensure components are registered
		m.LayoutEngine.SetTerminalSize(msg.Width, msg.Height)
		// Only initialize components if they haven't been registered yet
		if !m.LayoutEngine.HasComponents() {
			initializeLayout(m)
		}
	}

	if err := m.LayoutEngine.Layout(); err != nil {
		slog.Error("Failed to calculate layout on resize", "error", err, "width", msg.Width, "height", msg.Height)
	}
	updateComponentSizes(m)

	return m, nil
}

// handleTimerUpdate handles timer tick and start/stop events
func handleTimerUpdate(m *types.Model, msg tea.Msg) (*types.Model, tea.Cmd) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	var cmd tea.Cmd
	m.StatusTimer, cmd = m.StatusTimer.Update(msg)
	return m, cmd
}

// handleTimerTimeout handles timer timeout events
func handleTimerTimeout(m *types.Model) (*types.Model, tea.Cmd) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.StatusMessage = ""
	return m, nil
}

// handleKeyPress handles keyboard input
func handleKeyPress(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	m.Mutex.RLock()
	confirmMode := m.ConfirmMode
	m.Mutex.RUnlock()
	if confirmMode {
		return updateConfirmMode(m, msg)
	}
	return handleGlobalKeys(m, msg)
}

// handleGlobalKeys handles global keyboard shortcuts
func handleGlobalKeys(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keyMap.quit):
		return m, tea.Quit
	case key.Matches(msg, keyMap.tab):
		m.Mutex.Lock()
		m.ActivePanel = (m.ActivePanel + 1) % 2 // Only 2 panels now
		m.Mutex.Unlock()
		return m, nil
	case key.Matches(msg, keyMap.enter):
		return handleSubmit(m)
	case key.Matches(msg, keyMap.clear):
		m.Mutex.Lock()
		m.Actions = []types.Action{}
		m.Mutex.Unlock()
		clearPendingMoves(m)
		return m, nil
	}

	// Panel-specific updates
	m.Mutex.RLock()
	activePanel := m.ActivePanel
	m.Mutex.RUnlock()
	switch activePanel {
	case 0: // Permissions
		return updatePermissionsPanel(m, msg)
	case 1: // Duplicates
		return updateDuplicatesPanel(m, msg)
	}
	return m, nil
}

// updateConfirmMode handles confirmation dialog input
func updateConfirmMode(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keyMap.escape):
		m.Mutex.Lock()
		m.ConfirmMode = false
		m.Mutex.Unlock()
		return m, nil

	case key.Matches(msg, keyMap.enter):
		// Execute actions
		if err := executeActions(m); err != nil {
			// Show error message
			m.Mutex.Lock()
			m.ConfirmMode = false
			m.Mutex.Unlock()
			return m, showStatusMessage(m, fmt.Sprintf("Error: %v", err))
		}
		// Show success message and quit after timer
		m.Mutex.Lock()
		m.ConfirmMode = false
		actionCount := len(m.Actions)
		m.Actions = []types.Action{} // Clear actions
		m.Mutex.Unlock()
		clearPendingMoves(m)

		// Show success message for 2 seconds then quit
		cmd := showStatusMessage(m, fmt.Sprintf("✓ Successfully applied %d actions!", actionCount))
		return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return tea.Quit()
		}))

	default:
		return m, nil
	}
}

// updatePermissionsPanel handles permissions panel input
func updatePermissionsPanel(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	// Let the list handle its own keys (navigation, filtering, etc.) first
	var cmd tea.Cmd
	m.Mutex.Lock()
	m.PermissionsList, cmd = m.PermissionsList.Update(msg)
	m.Mutex.Unlock()

	// Then handle our custom keys
	switch {
	case key.Matches(msg, keyMap.space):
		// Toggle selection of current item
		m.Mutex.Lock()
		currentIdx := m.PermissionsList.Index()
		if currentIdx < len(m.Permissions) {
			m.Permissions[currentIdx].Selected = !m.Permissions[currentIdx].Selected
			// Update the list item
			m.PermissionsList.SetItem(currentIdx, m.Permissions[currentIdx])
		}
		m.Mutex.Unlock()
		return m, cmd

	case key.Matches(msg, keyMap.selectAll):
		// Toggle between all selected and none selected
		m.Mutex.Lock()
		allSelected := true
		for _, perm := range m.Permissions {
			if !perm.Selected {
				allSelected = false
				break
			}
		}

		// If all are selected, deselect all; otherwise select all
		for i := range m.Permissions {
			m.Permissions[i].Selected = !allSelected
			// Update the list item
			m.PermissionsList.SetItem(i, m.Permissions[i])
		}
		m.Mutex.Unlock()
		return m, cmd

	case key.Matches(msg, keyMap.moveUser):
		moveSelectedPermissions(m, types.LevelUser)
		return m, cmd

	case key.Matches(msg, keyMap.moveRepo):
		moveSelectedPermissions(m, types.LevelRepo)
		return m, cmd

	case key.Matches(msg, keyMap.moveLocal):
		moveSelectedPermissions(m, types.LevelLocal)
		return m, cmd
	}

	return m, cmd
}

// updateDuplicatesPanel handles duplicates panel input
func updateDuplicatesPanel(m *types.Model, msg tea.KeyMsg) (*types.Model, tea.Cmd) {
	// Let the table handle navigation first
	var cmd tea.Cmd
	m.Mutex.Lock()
	m.DuplicatesTable, cmd = m.DuplicatesTable.Update(msg)
	m.Mutex.Unlock()

	// Handle custom keys
	switch {
	case key.Matches(msg, keyMap.space):
		m.Mutex.Lock()
		cursor := m.DuplicatesTable.Cursor()
		if cursor < len(m.Duplicates) {
			m.Duplicates[cursor].Selected = !m.Duplicates[cursor].Selected
			updateDuplicatesTableRows(m)
		}
		m.Mutex.Unlock()
		return m, cmd

	case key.Matches(msg, keyMap.moveUser):
		setDuplicateKeepLevel(m, types.LevelUser)
		return m, cmd

	case key.Matches(msg, keyMap.moveRepo):
		setDuplicateKeepLevel(m, types.LevelRepo)
		return m, cmd

	case key.Matches(msg, keyMap.moveLocal):
		setDuplicateKeepLevel(m, types.LevelLocal)
		return m, cmd
	}

	return m, cmd
}

// showStatusMessage displays a status message for a few seconds
func showStatusMessage(m *types.Model, message string) tea.Cmd {
	m.Mutex.Lock()
	m.StatusMessage = message
	cmd := m.StatusTimer.Init()
	m.Mutex.Unlock()
	return cmd
}

// updateDuplicatesTableRows refreshes the table rows with current duplicate data
func updateDuplicatesTableRows(m *types.Model) {
	rows := []table.Row{}
	for _, dup := range m.Duplicates {
		levelsStr := strings.Join(dup.Levels, ", ")
		keepLevel := dup.KeepLevel
		if keepLevel == "" {
			keepLevel = "None"
		}
		// Add selection indicator
		name := dup.Name
		if dup.Selected {
			name = "[x] " + name
		} else {
			name = "[ ] " + name
		}
		rows = append(rows, table.Row{name, levelsStr, keepLevel})
	}
	m.DuplicatesTable.SetRows(rows)
}

// handleSubmit handles the submit action
func handleSubmit(m *types.Model) (*types.Model, tea.Cmd) {
	m.Mutex.RLock()
	actionsLen := len(m.Actions)
	m.Mutex.RUnlock()
	if actionsLen == 0 {
		return m, nil
	}

	// Generate confirmation text
	m.Mutex.Lock()
	m.ConfirmText = generateConfirmationText(m)
	m.ConfirmMode = true
	m.Mutex.Unlock()

	return m, nil
}

// updateComponentSizes updates component sizes using layout engine
// Enhanced with Lipgloss style-aware calculations
func updateComponentSizes(m *types.Model) {
	if m.LayoutEngine == nil {
		return
	}

	result := m.LayoutEngine.GetLastResult()
	if result == nil {
		return
	}

	// Get frame size from panel style for consistent calculations
	frameWidth, frameHeight := panelStyle.GetFrameSize()

	// Get layout information for each component
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	if permLayout, exists := result.Components["permissions"]; exists {
		// Use Lipgloss introspection and centralized constants for proper content area calculation
		contentWidth := maxInt(MinContentWidth, permLayout.Width-frameWidth)
		// Account for title (1 line) + column headers (1 line) = 2 lines overhead
		availableContentHeight := maxInt(MinListHeight, permLayout.Height-frameHeight-2)

		// Use the layout-allocated height to fill the allocated space properly
		slog.Debug("updateComponentSizes permissions",
			"layoutHeight", permLayout.Height,
			"frameHeight", frameHeight,
			"availableContentHeight", availableContentHeight,
			"usingHeight", availableContentHeight)
		m.PermissionsList.SetWidth(contentWidth)
		m.PermissionsList.SetHeight(availableContentHeight)
	}

	if dupLayout, exists := result.Components["duplicates"]; exists {
		// Use Lipgloss introspection and centralized constants for proper content area calculation
		contentWidth := maxInt(MinContentWidth, dupLayout.Width-frameWidth)
		// Account for title (1 line) = 1 line overhead
		availableContentHeight := maxInt(MinTableHeight, dupLayout.Height-frameHeight-1)

		// Use the layout-allocated height to fill the allocated space properly
		slog.Debug("updateComponentSizes duplicates",
			"layoutHeight", dupLayout.Height,
			"frameHeight", frameHeight,
			"availableContentHeight", availableContentHeight,
			"usingHeight", availableContentHeight)
		m.DuplicatesTable.SetWidth(contentWidth)
		m.DuplicatesTable.SetHeight(availableContentHeight)
	}
}

// View renders the entire UI
func View(m *types.Model) string {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	if m.ConfirmMode {
		return renderConfirmation(m)
	}

	// Handle case when layout engine hasn't been initialized yet
	if m.LayoutEngine == nil {
		return "Initializing layout... (waiting for terminal size)"
	}

	// Header
	header := renderHeader(m)

	// Panels (only 2 panels in main screen now)
	permissionsPanel := renderPermissionsPanel(m)
	duplicatesPanel := renderDuplicatesPanel(m)

	// Footer
	footer := renderFooter(m)

	// Status message (if any)
	var statusPanel string
	if m.StatusMessage != "" {
		statusPanel = statusMessageStyle.Render(m.StatusMessage)
	}

	// Set component content in layout engine and let it handle positioning
	contentMap := map[string]string{
		"header":      header,
		"permissions": permissionsPanel,
		"duplicates":  duplicatesPanel,
		"footer":      footer,
	}

	// Register content with layout engine
	if err := m.LayoutEngine.SetAllComponentContent(contentMap); err != nil {
		// Fallback to simple layout if registration fails
		elements := []string{header, permissionsPanel, duplicatesPanel}
		if statusPanel != "" {
			elements = append(elements, statusPanel)
		}
		elements = append(elements, footer)
		return lipgloss.JoinVertical(lipgloss.Left, elements...)
	}

	// Let layout engine handle absolute positioning and rendering
	content := m.LayoutEngine.View()

	// Add status panel as overlay if present
	if statusPanel != "" {
		// For now, prepend status panel - could be enhanced to overlay properly
		content = statusPanel + "\n" + content
	}

	return content
}


// renderConfirmation renders the confirmation dialog
func renderConfirmation(m *types.Model) string {
	width, height := m.LayoutEngine.GetTerminalSize()
	return dialogStyle.
		Width(width).
		Height(height).
		Render(m.ConfirmText)
}

// renderHeader renders the header section
func renderHeader(m *types.Model) string {
	// File status indicators with colored styling
	userStatus := "X"
	userStatusStyle := statusBadStyle
	if m.UserLevel.Exists {
		userStatus = "OK"
		userStatusStyle = statusGoodStyle
	}

	repoStatus := "X"
	repoStatusStyle := statusBadStyle
	if m.RepoLevel.Exists {
		repoStatus = "OK"
		repoStatusStyle = statusGoodStyle
	}

	localStatus := "X"
	localStatusStyle := statusBadStyle
	if m.LocalLevel.Exists {
		localStatus = "OK"
		localStatusStyle = statusGoodStyle
	}

	// Build file info with multi-color styling
	fileInfo := fmt.Sprintf("Files: User:%s%s Repo:%s%s Local:%s%s",
		userStatusStyle.Render(userStatus), countStyle.Render(fmt.Sprintf("(%d)", len(m.UserLevel.Permissions))),
		repoStatusStyle.Render(repoStatus), countStyle.Render(fmt.Sprintf("(%d)", len(m.RepoLevel.Permissions))),
		localStatusStyle.Render(localStatus), countStyle.Render(fmt.Sprintf("(%d)", len(m.LocalLevel.Permissions))))

	// Current working directory with accent color
	cwd, _ := os.Getwd()
	currentDir := fmt.Sprintf("%s %s", accentStyle.Render("Current:"), cwd)

	// Build header text with multi-color styling
	title := headerTextStyle.Render("Claude Tool Permission Editor")
	headerText := fmt.Sprintf("%s\n%s | %s", title, fileInfo, currentDir)

	// Create header with Lipgloss border rendering using centralized style
	width, _ := m.LayoutEngine.GetTerminalSize()

	// Use centralized header style with built-in border
	headerWithBorder := headerWithBorderStyle.
		Width(width).
		Render(headerText)

	return headerWithBorder
}

// renderPermissionsPanel renders the permissions panel
func renderPermissionsPanel(m *types.Model) string {
	title := fmt.Sprintf("Permissions (%d total)", len(m.Permissions))

	// Create column headers that match the two-column layout
	// Calculate column widths based on layout width, not list width
	totalWidth := m.PermissionsList.Width()
	if result := m.LayoutEngine.GetLastResult(); result != nil {
		if permLayout, exists := result.Components["permissions"]; exists {
			// Use the layout width minus borders/padding using Lipgloss introspection
			frameWidth, _ := panelStyle.GetFrameSize()
			totalWidth = permLayout.Width - frameWidth
		}
	}
	leftWidth := int(float64(totalWidth) * ColumnSplitRatio)
	rightWidth := totalWidth - leftWidth

	leftHeader := leftColumnStyle.
		Width(leftWidth).
		Render(columnHeaderStyle.Render("Permission Name"))

	rightHeader := rightColumnStyle.
		Width(rightWidth).
		Render(columnHeaderStyle.Render("Current Level"))

	columnHeaders := lipgloss.JoinHorizontal(lipgloss.Top, leftHeader, rightHeader)

	// Get list content
	content := m.PermissionsList.View()

	style := panelStyle
	if m.ActivePanel == 0 {
		style = activePanelStyle
	}

	// Calculate actual content height needed and use that instead of layout height
	rawContent := fmt.Sprintf("%s\n%s\n%s", title, columnHeaders, content)
	rawLines := strings.Split(rawContent, "\n")
	actualContentHeight := len(rawLines)

	// Apply height constraint from layout, let width be natural
	if result := m.LayoutEngine.GetLastResult(); result != nil {
		if permLayout, exists := result.Components["permissions"]; exists {
			slog.Debug("permissions panel sizing",
				"layoutHeight", permLayout.Height,
				"actualContentHeight", actualContentHeight,
				"allocatedWidth", permLayout.Width,
				"usingHeight", actualContentHeight,
				"rawContentLength", len(strings.Split(rawContent, "\n")[0]))
			// Only set height, let width be determined by content
			style = style.Height(actualContentHeight)
		}
	}

	return style.Render(rawContent)
}

// renderDuplicatesPanel renders the duplicates panel using table component
func renderDuplicatesPanel(m *types.Model) string {
	title := fmt.Sprintf("Duplicates (%d conflicts)", len(m.Duplicates))

	// Set table focus based on active panel
	if m.ActivePanel == 1 {
		m.DuplicatesTable.Focus()
	} else {
		m.DuplicatesTable.Blur()
	}

	content := m.DuplicatesTable.View()

	style := panelStyle
	if m.ActivePanel == 1 {
		style = activePanelStyle
	}

	// Calculate actual content height needed and use that instead of layout height
	rawContent := fmt.Sprintf("%s\n%s", title, content)
	rawLines := strings.Split(rawContent, "\n")
	actualContentHeight := len(rawLines)

	// Apply height constraint from layout, let width be natural
	if result := m.LayoutEngine.GetLastResult(); result != nil {
		if dupLayout, exists := result.Components["duplicates"]; exists {
			slog.Debug("duplicates panel sizing",
				"layoutHeight", dupLayout.Height,
				"actualContentHeight", actualContentHeight,
				"allocatedWidth", dupLayout.Width,
				"usingHeight", actualContentHeight)
			// Only set height, let width be determined by content
			style = style.Height(actualContentHeight)
		}
	}

	return style.Render(rawContent)
}

// renderFooter renders the footer with context-sensitive hotkeys in a fixed 2-line layout
func renderFooter(m *types.Model) string {
	var row1Keys []string
	var row2Keys []string

	// Row 1: Panel-specific keys with multi-color styling
	switch m.ActivePanel {
	case 0: // Permissions
		row1Keys = []string{
			headerTextStyle.Render("↑↓:") + " Navigate",
			headerTextStyle.Render("SPACE:") + " Select",
			headerTextStyle.Render("A:") + " Toggle All",
			headerTextStyle.Render("E:") + " Edit",
			headerTextStyle.Render("U/R/L:") + " Move",
			headerTextStyle.Render("/:") + " Filter",
		}
	case 1: // Duplicates
		row1Keys = []string{
			headerTextStyle.Render("↑↓:") + " Navigate",
			headerTextStyle.Render("SPACE:") + " Select",
			headerTextStyle.Render("U/R/L:") + " Keep Level",
		}
	}

	// Row 2: Common action keys with multi-color styling
	row2Keys = []string{
		accentStyle.Render("ENTER:") + " Submit",
		accentStyle.Render("C:") + " Clear",
		accentStyle.Render("TAB:") + " Switch",
		accentStyle.Render("Q:") + " Quit",
	}

	// Create 2-line footer with consistent styling
	footerText := strings.Join(row1Keys, "  |  ") + "\n" + strings.Join(row2Keys, "  |  ")

	width, _ := m.LayoutEngine.GetTerminalSize()
	return footerStyle.
		Width(width).
		Render(footerText)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}


// Note: Init, Update, View functions are now called directly by AppModel wrapper in main.go

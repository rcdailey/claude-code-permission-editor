package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all input events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)
	case timer.TickMsg, timer.StartStopMsg:
		return m.handleTimerUpdate(msg)
	case timer.TimeoutMsg:
		return m.handleTimerTimeout()
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

// handleWindowResize handles window resize events
func (m Model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.updateViewports()
	return m, nil
}

// handleTimerUpdate handles timer tick and start/stop events
func (m Model) handleTimerUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.statusTimer, cmd = m.statusTimer.Update(msg)
	return m, cmd
}

// handleTimerTimeout handles timer timeout events
func (m Model) handleTimerTimeout() (tea.Model, tea.Cmd) {
	m.statusMessage = ""
	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmMode {
		return m.updateConfirmMode(msg)
	}
	return m.handleGlobalKeys(msg)
}

// handleGlobalKeys handles global keyboard shortcuts
func (m Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keyMap.quit):
		return m, tea.Quit
	case key.Matches(msg, keyMap.tab):
		m.activePanel = (m.activePanel + 1) % 2 // Only 2 panels now
		return m, nil
	case key.Matches(msg, keyMap.enter):
		return m.handleSubmit()
	case key.Matches(msg, keyMap.clear):
		m.actions = []Action{}
		m.clearPendingMoves()
		return m, nil
	}

	// Panel-specific updates
	switch m.activePanel {
	case 0: // Permissions
		return m.updatePermissionsPanel(msg)
	case 1: // Duplicates
		return m.updateDuplicatesPanel(msg)
	}
	return m, nil
}

// updateConfirmMode handles confirmation dialog input
func (m Model) updateConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keyMap.escape):
		m.confirmMode = false
		return m, nil

	case key.Matches(msg, keyMap.enter):
		// Execute actions
		if err := m.executeActions(); err != nil {
			// Show error message
			m.confirmMode = false
			return m, m.showStatusMessage(fmt.Sprintf("Error: %v", err))
		}
		// Show success message and quit after timer
		m.confirmMode = false
		actionCount := len(m.actions)
		m.actions = []Action{} // Clear actions
		m.clearPendingMoves()

		// Show success message for 2 seconds then quit
		cmd := m.showStatusMessage(fmt.Sprintf("✓ Successfully applied %d actions!", actionCount))
		return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return tea.Quit()
		}))

	default:
		return m, nil
	}
}

// updatePermissionsPanel handles permissions panel input
func (m Model) updatePermissionsPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Let the list handle its own keys (navigation, filtering, etc.) first
	var cmd tea.Cmd
	m.permissionsList, cmd = m.permissionsList.Update(msg)

	// Then handle our custom keys
	switch {
	case key.Matches(msg, keyMap.space):
		// Toggle selection of current item
		currentIdx := m.permissionsList.Index()
		if currentIdx < len(m.permissions) {
			m.permissions[currentIdx].Selected = !m.permissions[currentIdx].Selected
			// Update the list item
			m.permissionsList.SetItem(currentIdx, m.permissions[currentIdx])
		}
		return m, cmd

	case key.Matches(msg, keyMap.selectAll):
		// Toggle between all selected and none selected
		allSelected := true
		for _, perm := range m.permissions {
			if !perm.Selected {
				allSelected = false
				break
			}
		}

		// If all are selected, deselect all; otherwise select all
		for i := range m.permissions {
			m.permissions[i].Selected = !allSelected
			// Update the list item
			m.permissionsList.SetItem(i, m.permissions[i])
		}
		return m, cmd

	case key.Matches(msg, keyMap.moveUser):
		m.moveSelectedPermissions(LevelUser)
		return m, cmd

	case key.Matches(msg, keyMap.moveRepo):
		m.moveSelectedPermissions(LevelRepo)
		return m, cmd

	case key.Matches(msg, keyMap.moveLocal):
		m.moveSelectedPermissions(LevelLocal)
		return m, cmd
	}

	return m, cmd
}

// updateDuplicatesPanel handles duplicates panel input
func (m Model) updateDuplicatesPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Let the table handle navigation first
	var cmd tea.Cmd
	m.duplicatesTable, cmd = m.duplicatesTable.Update(msg)

	// Handle custom keys
	switch {
	case key.Matches(msg, keyMap.space):
		cursor := m.duplicatesTable.Cursor()
		if cursor < len(m.duplicates) {
			m.duplicates[cursor].Selected = !m.duplicates[cursor].Selected
			m.updateDuplicatesTableRows()
		}
		return m, cmd

	case key.Matches(msg, keyMap.moveUser):
		m.setDuplicateKeepLevel(LevelUser)
		return m, cmd

	case key.Matches(msg, keyMap.moveRepo):
		m.setDuplicateKeepLevel(LevelRepo)
		return m, cmd

	case key.Matches(msg, keyMap.moveLocal):
		m.setDuplicateKeepLevel(LevelLocal)
		return m, cmd
	}

	return m, cmd
}

// showStatusMessage displays a status message for a few seconds
func (m *Model) showStatusMessage(message string) tea.Cmd {
	m.statusMessage = message
	return m.statusTimer.Init()
}

// updateDuplicatesTableRows refreshes the table rows with current duplicate data
func (m *Model) updateDuplicatesTableRows() {
	rows := []table.Row{}
	for _, dup := range m.duplicates {
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
	m.duplicatesTable.SetRows(rows)
}

// handleSubmit handles the submit action
func (m Model) handleSubmit() (tea.Model, tea.Cmd) {
	if len(m.actions) == 0 {
		return m, nil
	}

	// Generate confirmation text
	m.confirmText = m.generateConfirmationText()
	m.confirmMode = true

	return m, nil
}

// updateViewports updates viewport sizes based on terminal dimensions
func (m *Model) updateViewports() {
	// Don't update if dimensions aren't set yet
	if m.width == 0 || m.height == 0 {
		return
	}

	// Fixed heights to ensure header is always visible
	headerHeight := 2 // Header is always 2 lines
	footerHeight := 2 // Footer is fixed at 2 lines
	panelBorders := 4 // Each panel has 2 border lines (top/bottom) = 2 panels * 2 = 4

	// Calculate available height for viewport content with conservative approach
	availableHeight := m.height - headerHeight - footerHeight - panelBorders

	// Ensure we have enough space, prioritize header visibility
	if availableHeight < 6 {
		availableHeight = 6 // Absolute minimum for content
	}

	// For very small terminals, reduce panel content but keep header
	if m.height < 25 {
		availableHeight = maxInt(4, m.height-headerHeight-footerHeight-panelBorders)
	}

	// Allocate space between 2 panels (content area only, not including borders)
	// Give more space to permissions since it's the main panel
	permissionsHeight := maxInt(4, availableHeight*70/100)           // Reduced from 75% to 70%
	duplicatesHeight := maxInt(2, availableHeight-permissionsHeight) // Minimum 2 lines

	// Set dimensions (account for panel borders and padding)
	contentWidth := maxInt(40, m.width-4) // Minimum width of 40, account for borders

	// Update list dimensions
	m.permissionsList.SetWidth(contentWidth)
	m.permissionsList.SetHeight(permissionsHeight)

	// Update duplicates table
	m.duplicatesTable.SetWidth(contentWidth)
	m.duplicatesTable.SetHeight(duplicatesHeight)
}

// View renders the entire UI
func (m Model) View() string {
	if m.confirmMode {
		return m.renderConfirmation()
	}

	// Handle case when terminal dimensions haven't been set yet
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Header
	header := m.renderHeader()

	// Panels (only 2 panels in main screen now)
	permissionsPanel := m.renderPermissionsPanel()
	duplicatesPanel := m.renderDuplicatesPanel()

	// Footer
	footer := m.renderFooter()

	// Status message (if any)
	var statusPanel string
	if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")). // Green
			Bold(true).
			Padding(0, 1).
			Margin(1, 0)
		statusPanel = statusStyle.Render(m.statusMessage)
	}

	// Build UI elements (2-panel layout)
	elements := []string{header, permissionsPanel, duplicatesPanel}
	if statusPanel != "" {
		elements = append(elements, statusPanel)
	}
	elements = append(elements, footer)

	ui := lipgloss.JoinVertical(lipgloss.Left, elements...)

	// In debug mode, add layout information
	if *debugMode {
		debugInfo := "\n=== DEBUG INFO ===\n"
		debugInfo += fmt.Sprintf("Terminal: %dx%d\n", m.width, m.height)
		debugInfo += fmt.Sprintf("Permissions viewport: %dx%d\n", m.permissionsView.Width, m.permissionsView.Height)
		debugInfo += fmt.Sprintf("Duplicates table: %dx%d\n", m.duplicatesTable.Width(), m.duplicatesTable.Height())
		debugInfo += fmt.Sprintf("Actions viewport: %dx%d\n", m.actionsView.Width, m.actionsView.Height)
		debugInfo += fmt.Sprintf("Header height: %d lines\n", lipgloss.Height(header))
		debugInfo += fmt.Sprintf("Footer height: %d lines\n", lipgloss.Height(footer))
		debugInfo += fmt.Sprintf("Total UI height: %d lines\n", lipgloss.Height(ui))
		debugInfo += "=== UI OUTPUT ===\n"

		return debugInfo + ui
	}

	return ui
}

// renderConfirmation renders the confirmation dialog
func (m Model) renderConfirmation() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(m.confirmText)
}

// renderHeader renders the header section
func (m Model) renderHeader() string {
	// File status indicators (ASCII for better compatibility)
	userStatus := "X"
	if m.userLevel.Exists {
		userStatus = "OK"
	}

	repoStatus := "X"
	if m.repoLevel.Exists {
		repoStatus = "OK"
	}

	localStatus := "X"
	if m.localLevel.Exists {
		localStatus = "OK"
	}

	fileInfo := fmt.Sprintf("Files: User:%s(%d) Repo:%s(%d) Local:%s(%d)",
		userStatus, len(m.userLevel.Permissions),
		repoStatus, len(m.repoLevel.Permissions),
		localStatus, len(m.localLevel.Permissions))

	// Current working directory
	cwd, _ := os.Getwd()

	headerText := fmt.Sprintf("Claude Tool Permission Editor\n%s | Current: %s", fileInfo, cwd)

	// Add extra spacing to ensure header is visible
	header := headerStyle.
		Width(m.width).
		Render(headerText)

	// Just return the header without extra spacing
	return header
}

// renderPermissionsPanel renders the permissions panel
func (m Model) renderPermissionsPanel() string {
	title := fmt.Sprintf("Permissions (%d total)", len(m.permissions))

	// Use the optimized list component for rendering
	content := m.permissionsList.View()

	style := panelStyle
	if m.activePanel == 0 {
		style = activePanelStyle
	}

	return style.Render(fmt.Sprintf("%s\n%s", title, content))
}

// renderDuplicatesPanel renders the duplicates panel using table component
func (m Model) renderDuplicatesPanel() string {
	title := fmt.Sprintf("Duplicates (%d conflicts)", len(m.duplicates))

	// Set table focus based on active panel
	if m.activePanel == 1 {
		m.duplicatesTable.Focus()
	} else {
		m.duplicatesTable.Blur()
	}

	content := m.duplicatesTable.View()

	style := panelStyle
	if m.activePanel == 1 {
		style = activePanelStyle
	}

	return style.Render(fmt.Sprintf("%s\n%s", title, content))
}

// renderFooter renders the footer with context-sensitive hotkeys in a fixed 2-line layout
func (m Model) renderFooter() string {
	var row1Keys []string
	var row2Keys []string

	// Row 1: Panel-specific keys
	switch m.activePanel {
	case 0: // Permissions
		row1Keys = []string{"↑↓: Navigate", "SPACE: Select", "A: Toggle All", "E: Edit", "U/R/L: Move", "/: Filter"}
	case 1: // Duplicates
		row1Keys = []string{"↑↓: Navigate", "SPACE: Select", "U/R/L: Keep Level"}
	}

	// Row 2: Common action keys (same for all panels)
	row2Keys = []string{"ENTER: Submit", "C: Clear", "TAB: Switch", "Q: Quit"}

	// Create 2-line footer
	footerText := strings.Join(row1Keys, "  |  ") + "\n" + strings.Join(row2Keys, "  |  ")

	return footerStyle.
		Width(m.width).
		Render(footerText)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

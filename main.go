package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"claude-permissions/debug"
	"claude-permissions/types"
	"claude-permissions/ui"

	"github.com/charmbracelet/bubbles/v2/table"
	"github.com/charmbracelet/bubbles/v2/timer"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// Command line flags for testing
var (
	userFile    = flag.String("user-file", "", "Override user level settings file path")
	repoFile    = flag.String("repo-file", "", "Override repo level settings file path")
	localFile   = flag.String("local-file", "", "Override local level settings file path")
	debugServer = flag.Bool("debug-server", false, "Start HTTP debug server alongside TUI")
	debugPort   = flag.Int("debug-port", 8080, "Port for debug server")
)

// AppModel wraps types.Model and implements tea.Model interface
type AppModel struct {
	*types.Model
}

// Init implements tea.Model interface
func (a *AppModel) Init() tea.Cmd {
	return ui.Init(a.Model)
}

// Update implements tea.Model interface
func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := ui.Update(a.Model, msg)
	a.Model = newModel
	return a, cmd
}

// View implements tea.Model interface
func (a *AppModel) View() string {
	return ui.View(a.Model)
}

// GetView implements debug.ViewProvider interface
func (a *AppModel) GetView() string {
	return a.View()
}

// setupLogger configures the global slog logger based on debug server availability
func setupLogger(debugSrv *debug.DebugServer) {
	var handler slog.Handler

	if debugSrv != nil {
		// Debug server enabled - route logs to debug server
		handler = debug.NewDebugSlogHandler(debugSrv.Logger())
	} else {
		// Debug server disabled - use no-op handler for zero overhead
		handler = NoOpHandler{}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	flag.Parse()

	dataModel, err := initialModel()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Wrap the data model with AppModel to implement tea.Model
	appModel := &AppModel{Model: dataModel}

	// Normal mode: interactive TUI
	p := tea.NewProgram(appModel, tea.WithAltScreen())

	// Start debug server if requested
	var debugSrv *debug.DebugServer
	if *debugServer {
		debugSrv = debug.NewDebugServer(*debugPort, p, dataModel, appModel)
		if err := debugSrv.Start(); err != nil {
			fmt.Printf("Warning: Failed to start debug server: %v\n", err)
		} else {
			fmt.Printf("Debug server started on port %d\n", *debugPort)
		}
	}

	// Setup logging system based on debug server availability
	setupLogger(debugSrv)

	// Run the TUI program
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Stop debug server if it was started
	if debugSrv != nil {
		if err := debugSrv.Stop(); err != nil {
			fmt.Printf("Warning: Failed to stop debug server: %v\n", err)
		}
	}

	// Update debug server with final model if needed
	if debugSrv != nil {
		if finalAppModel, ok := finalModel.(*AppModel); ok {
			debugSrv.UpdateModel(finalAppModel.Model)
		}
	}
}

// loadAllLevels loads settings from all three levels
func loadAllLevels() (types.SettingsLevel, types.SettingsLevel, types.SettingsLevel, int, error) {
	userLevel, err := loadUserLevel()
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, fmt.Errorf(
			"failed to load user level: %w",
			err,
		)
	}

	repoLevel, err := loadRepoLevel()
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, fmt.Errorf(
			"failed to load repo level: %w",
			err,
		)
	}

	localLevel, err := loadLocalLevel()
	if err != nil {
		return types.SettingsLevel{}, types.SettingsLevel{}, types.SettingsLevel{}, 0, fmt.Errorf(
			"failed to load local level: %w",
			err,
		)
	}

	// Auto-resolve same-level duplicates and track statistics
	userCleaned := autoResolveSameLevelDuplicates(&userLevel)
	repoCleaned := autoResolveSameLevelDuplicates(&repoLevel)
	localCleaned := autoResolveSameLevelDuplicates(&localLevel)
	totalSameLevelCleaned := userCleaned + repoCleaned + localCleaned

	return userLevel, repoLevel, localLevel, totalSameLevelCleaned, nil
}

// createUIComponents creates the UI components
func createUIComponents(duplicates []types.Duplicate) table.Model {
	// Create table for duplicates panel
	duplicatesTable := createDuplicatesTable(duplicates)

	return duplicatesTable
}

func initialModel() (*types.Model, error) {
	userLevel, repoLevel, localLevel, totalSameLevelCleaned, err := loadAllLevels()
	if err != nil {
		return nil, err
	}

	// Create consolidated permissions list
	permissions := consolidatePermissions(userLevel, repoLevel, localLevel)

	// Detect cross-level duplicates
	duplicates := detectDuplicates(userLevel, repoLevel, localLevel)

	duplicatesTable := createUIComponents(duplicates)

	// Determine starting screen based on duplicates
	startingScreen := types.ScreenOrganization
	if len(duplicates) > 0 {
		startingScreen = types.ScreenDuplicates
	}

	model := &types.Model{
		UserLevel:     userLevel,
		RepoLevel:     repoLevel,
		LocalLevel:    localLevel,
		Permissions:   permissions,
		Duplicates:    duplicates,
		ActivePanel:   0,
		CurrentScreen: startingScreen,
		CleanupStats: struct {
			DuplicatesResolved int
			SameLevelCleaned   int
		}{
			DuplicatesResolved: 0,
			SameLevelCleaned:   totalSameLevelCleaned,
		},
		FocusedColumn:    0, // Start with LOCAL column
		SelectedItem:     0,
		ColumnSelections: [3]int{0, 0, 0},
		Width:            0, // Will be set by terminal size message
		Height:           0, // Will be set by terminal size message
		DuplicatesTable:  duplicatesTable,
		ConfirmMode:      false,
		StatusMessage:    "",
		StatusTimer:      timer.New(3 * time.Second),
	}

	return model, nil
}

func createDuplicatesTable(duplicates []types.Duplicate) table.Model {
	columns := []table.Column{
		{Title: "Permission", Width: 30},
		{Title: "Found In", Width: 25},
		{Title: "Keep Level", Width: 15},
	}

	rows := []table.Row{}
	for _, dup := range duplicates {
		levelsStr := strings.Join(dup.Levels, ", ")
		keepLevel := dup.KeepLevel
		if keepLevel == "" {
			keepLevel = "None"
		}
		rows = append(rows, table.Row{dup.Name, levelsStr, keepLevel})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Apply consistent styling to match permissions panel headers
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(ui.ColorBorderNormal)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(ui.ColorTitle)) // Bright white text, no background

	tableStyle.Selected = tableStyle.Selected.
		Foreground(lipgloss.Color(ui.ColorAccent)).              // Use theme accent color
		Background(lipgloss.Color(ui.ColorBackgroundSecondary)). // Use theme secondary background
		Bold(true)
		// Match permissions screen selection style

	t.SetStyles(tableStyle)

	return t
}

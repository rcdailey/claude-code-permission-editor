package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Command line flags for testing
var (
	userFile  = flag.String("user-file", "", "Override user level settings file path")
	repoFile  = flag.String("repo-file", "", "Override repo level settings file path")
	localFile = flag.String("local-file", "", "Override local level settings file path")
	debugMode = flag.Bool("debug", false, "Print UI layout to stdout and exit")
)

func main() {
	flag.Parse()

	model, err := initialModel()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if *debugMode {
		// Debug mode: print debug info and UI snapshot
		fmt.Printf("=== Claude Permission Editor Debug Output ===\n")
		fmt.Printf("User Level: %s (exists: %t, permissions: %d)\n",
			model.userLevel.Path, model.userLevel.Exists, len(model.userLevel.Permissions))
		fmt.Printf("Repo Level: %s (exists: %t, permissions: %d)\n",
			model.repoLevel.Path, model.repoLevel.Exists, len(model.repoLevel.Permissions))
		fmt.Printf("Local Level: %s (exists: %t, permissions: %d)\n",
			model.localLevel.Path, model.localLevel.Exists, len(model.localLevel.Permissions))
		fmt.Printf("Total Permissions: %d\n", len(model.permissions))
		fmt.Printf("Duplicates Found: %d\n", len(model.duplicates))

		// Set dimensions for debug view
		model.width = 120
		model.height = 40
		model.updateViewports()

		fmt.Println("\n=== UI Layout ===")
		fmt.Println(model.View())
		return
	}

	// Normal mode: interactive TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() (Model, error) {
	// Load settings from all levels
	userLevel, err := loadUserLevel()
	if err != nil {
		return Model{}, fmt.Errorf("failed to load user level: %w", err)
	}

	repoLevel, err := loadRepoLevel()
	if err != nil {
		return Model{}, fmt.Errorf("failed to load repo level: %w", err)
	}

	localLevel, err := loadLocalLevel()
	if err != nil {
		return Model{}, fmt.Errorf("failed to load local level: %w", err)
	}

	// Create consolidated permissions list
	permissions := consolidatePermissions(userLevel, repoLevel, localLevel)

	// Detect duplicates
	duplicates := detectDuplicates(userLevel, repoLevel, localLevel)

	// Create list items for permissions
	listItems := make([]list.Item, len(permissions))
	for i, perm := range permissions {
		listItems[i] = perm
	}

	// Create list with custom delegate
	delegate := PermissionDelegate{}
	permissionsList := list.New(listItems, delegate, 0, 0)
	permissionsList.SetShowStatusBar(false)
	permissionsList.SetShowHelp(false)
	permissionsList.SetFilteringEnabled(true)
	permissionsList.SetShowTitle(false)

	// Create table for duplicates panel
	duplicatesTable := createDuplicatesTable(duplicates)

	// Create viewport for actions panel
	actionsView := viewport.New(0, 0)

	model := Model{
		userLevel:       userLevel,
		repoLevel:       repoLevel,
		localLevel:      localLevel,
		permissions:     permissions,
		duplicates:      duplicates,
		actions:         []Action{},
		activePanel:     0,
		permissionsList: permissionsList,
		permissionsView: viewport.New(0, 0), // Keep for compatibility
		duplicatesTable: duplicatesTable,
		actionsView:     actionsView,
		width:           0, // Will be set by first WindowSizeMsg
		height:          0, // Will be set by first WindowSizeMsg
		confirmMode:     false,
		statusMessage:   "",
		statusTimer:     timer.New(3 * time.Second),
	}

	return model, nil
}

func createDuplicatesTable(duplicates []Duplicate) table.Model {
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
		table.WithFocused(false),
		table.WithHeight(7),
	)

	return t
}

package types

import (
	"sync"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
)

// Constants for settings levels
const (
	LevelUser  = "User"
	LevelRepo  = "Repo"
	LevelLocal = "Local"
)

// Constants for action types
const (
	ActionDuplicate = "duplicate"
	ActionMove      = "move"
)

// Constants for screen states
const (
	ScreenDuplicates = iota
	ScreenOrganization
)

// Settings represents the structure of Claude settings.json
type Settings struct {
	Allow []string `json:"allow"`
}

// SettingsLevel represents a level of settings (User, Repo, Local)
type SettingsLevel struct {
	Name        string
	Path        string
	Permissions []string
	Exists      bool
}

// Permission represents a permission with its current level and pending operations
type Permission struct {
	Name          string
	CurrentLevel  string
	OriginalLevel string // Track the original level for moved permissions
	Selected      bool
	Edited        bool
	NewName       string
}

// FilterValue implements the list.Item interface for Permission.
func (p Permission) FilterValue() string {
	return p.Name
}

// Duplicate represents a duplicate permission across levels
type Duplicate struct {
	Name      string
	Levels    []string
	KeepLevel string
	Selected  bool
}

// Action represents a queued action
type Action struct {
	Type       string // "move", "edit", "duplicate"
	Permission string
	FromLevel  string
	ToLevel    string
	NewName    string
}

// Model represents the application state
type Model struct {
	// Thread safety
	Mutex sync.RWMutex // Changed from: mutex sync.RWMutex

	// Settings data
	UserLevel  SettingsLevel // Changed from: userLevel
	RepoLevel  SettingsLevel // Changed from: repoLevel
	LocalLevel SettingsLevel // Changed from: localLevel

	// UI state
	Permissions []Permission // Changed from: permissions
	Duplicates  []Duplicate  // Changed from: duplicates
	Actions     []Action     // Changed from: actions
	ActivePanel int          // Changed from: activePanel

	// Screen management
	CurrentScreen int
	CleanupStats  struct {
		DuplicatesResolved int
		SameLevelCleaned   int
	}

	// Terminal dimensions (for pure lipgloss layout)
	Width  int
	Height int

	// Three-column organization state
	FocusedColumn    int    // 0=LOCAL, 1=REPO, 2=USER
	SelectedItem     int    // Index within focused column
	ColumnSelections [3]int // Selection index for each column

	// UI components
	DuplicatesTable table.Model    // Changed from: duplicatesTable
	ActionsView     viewport.Model // Changed from: actionsView

	// Confirmation state
	ConfirmMode bool   // Changed from: confirmMode
	ConfirmText string // Changed from: confirmText

	// Modal state
	ShowModal   bool
	ModalTitle  string
	ModalBody   string
	ModalAction string // "continue" or "exit"

	// Status message state
	StatusMessage string      // Changed from: statusMessage
	StatusTimer   timer.Model // Changed from: statusTimer
}

// Note: tea.Model interface methods are now implemented by AppModel wrapper in main package

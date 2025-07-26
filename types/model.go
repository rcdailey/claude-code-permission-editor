package types

import (
	"sync"

	"claude-permissions/layout"

	"github.com/charmbracelet/bubbles/list"
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
	Name         string
	CurrentLevel string
	PendingMove  string
	Selected     bool
	Edited       bool
	NewName      string
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

	// Layout engine
	LayoutEngine *layout.LayoutEngine // Changed from: layoutEngine

	// UI components
	PermissionsList list.Model    // Changed from: permissionsList
	DuplicatesTable table.Model   // Changed from: duplicatesTable
	ActionsView     viewport.Model // Changed from: actionsView

	// Confirmation state
	ConfirmMode bool   // Changed from: confirmMode
	ConfirmText string // Changed from: confirmText

	// Status message state
	StatusMessage string      // Changed from: statusMessage
	StatusTimer   timer.Model // Changed from: statusTimer
}

// Note: tea.Model interface methods are now implemented by AppModel wrapper in main package

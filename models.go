package main

import (
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
	// Settings data
	userLevel  SettingsLevel
	repoLevel  SettingsLevel
	localLevel SettingsLevel

	// UI state
	permissions []Permission
	duplicates  []Duplicate
	actions     []Action
	activePanel int // 0=permissions, 1=duplicates, 2=actions

	// Permissions panel
	permissionsList list.Model
	permissionsView viewport.Model // Keep for duplicates/actions panels

	// Duplicates panel
	duplicatesTable table.Model

	// Actions panel
	actionsView viewport.Model

	// UI dimensions
	width  int
	height int

	// Confirmation state
	confirmMode bool
	confirmText string

	// Status message state
	statusMessage string
	statusTimer   timer.Model
}

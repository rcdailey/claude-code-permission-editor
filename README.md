# Claude Code Permission Editor

An interactive TUI tool for managing Claude Code tool permissions across different settings levels.

## Overview

This tool provides a visual interface to manage Claude Code permissions across three levels:
- **User level**: `~/.claude/settings.json` (with chezmoi support)
- **Repo level**: `{REPO}/.claude/settings.json`
- **Local level**: `{REPO}/.claude/settings.local.json`

## Features

- Interactive terminal UI with streamlined two-panel layout
- Two-phase workflow: queue actions, then review and execute
- Consolidated permissions view across all levels
- Automatic duplicate detection and resolution
- Move permissions between levels
- Full-screen confirmation with comprehensive action summary
- Search functionality with highlighting
- Chezmoi dotfiles integration
- Git repository auto-detection

## Quick Start

### Build and Run

```bash
# Quick test with sample data
./scripts/test.sh

# Build for production use
go build -o claude-permissions .
```

### Basic Usage

```bash
# Interactive mode (normal use)
./claude-permissions

# Debug mode (see loaded data and UI layout)
./claude-permissions --debug

# Override file paths for testing
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
```

## Interface

### Navigation
- `↑↓`: Navigate lists
- `TAB`: Switch between panels
- `SPACE`: Select/deselect items
- `A`: Select all
- `N`: Deselect all

### Actions
- `U/R/L`: Move selected permissions to User/Repo/Local level
- `/`: Search permissions
- `ENTER`: Preview and apply changes
- `C`: Clear all pending actions
- `Q`: Quit

## User Interface

### Two-Phase Workflow

The application follows a two-phase approach to ensure safe permission management:

**Phase 1 - Action Planning (Main Screen):**
- **Permissions Panel**: Browse and select permissions to move between levels
- **Duplicates Panel**: Resolve conflicts by choosing which level to keep
- Navigate with TAB, select items, queue up changes
- No files are modified during this phase

**Phase 2 - Review & Execute (Confirmation Screen):**
- Full-screen summary of all queued actions
- Shows exactly what will be changed before applying
- Confirm with ENTER to execute, or ESC to return and modify

## Requirements

- Go 1.21+
- Terminal supporting ANSI colors

## Testing

```bash
# Run comprehensive tests
./scripts/verify.sh
```

## Implementation Notes

- Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Only modifies "allow" permissions arrays in JSON files
- Preserves all other settings in JSON files
- Supports chezmoi dotfiles when `chezmoi` command is available
- Auto-detects git repositories for repo/local level settings

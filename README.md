# Claude Code Permission Editor

> [!WARNING]
> This application is currently a work in progress and may not be fully functional. Use at your
> own risk and consider backing up your settings files before use.

An interactive terminal application for managing Claude Code tool permissions across different
settings levels.

## Overview

While Claude Code has built-in permission management, it's clunky and limited. This tool provides a
better editing workflow that's quicker and easier to use.

Manage Claude Code permissions across three levels:

- **User level**: `~/.claude/settings.json` (with chezmoi support)
- **Repo level**: `{REPO}/.claude/settings.json`
- **Local level**: `{REPO}/.claude/settings.local.json`

## Features

- Interactive terminal interface for permission management
- Move permissions between User/Repo/Local levels
- Automatic duplicate detection and resolution
- Search functionality
- Chezmoi dotfiles integration
- Safe: only modifies permission arrays, preserves other settings

## Installation

### Download Binary

Download the latest release from GitHub releases page, or build from source:

```bash
go install github.com/rcdailey/claude-code-permission-editor@latest
```

### Build from Source

```bash
git clone https://github.com/rcdailey/claude-code-permission-editor.git
cd claude-code-permission-editor
go build -o claude-permissions .
```

## Usage

### Basic Usage

```bash
# Start the application
./claude-permissions

# Test with sample data
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
```

## How to Use

The application provides context-sensitive help in the footer that shows available keys for each
screen.

### Duplicates Screen

- `↑↓`: Navigate between duplicate conflicts
- `1/2/3`: Keep permission in LOCAL/REPO/USER level
- `TAB`: Switch to organization screen
- `ENTER`: Save changes and continue
- `ESC`: Cancel/exit (if there are pending changes)

### Organization Screen

- `↑↓`: Navigate within current column
- `←→`: Switch between columns (Local/Repo/User)
- `1/2/3`: Move selected permission to LOCAL/REPO/USER level
- `TAB`: Switch to duplicates screen
- `ENTER`: Save changes and exit
- `ESC`: Reset all pending changes

### Global Keys

- `Q`: Quit application
- `Ctrl+C`: Force quit

## Requirements

- Go 1.21+
- Terminal supporting ANSI colors

## How It Works

The application loads your Claude Code settings from all three levels and presents them in a unified
interface. You can:

1. **Browse permissions** across User, Repo, and Local levels
2. **Move permissions** between levels using keyboard shortcuts
3. **Resolve duplicates** when the same permission exists at multiple levels
4. **Save changes** to update your settings files

Only permission arrays are modified - all other settings remain untouched.

## Support

For issues and feature requests, please visit the [GitHub
repository](https://github.com/rcdailey/claude-code-permission-editor).

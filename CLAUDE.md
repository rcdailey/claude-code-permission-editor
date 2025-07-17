# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

## Project Overview

This is a Go-based interactive TUI tool for managing Claude Code tool permissions across different
settings levels. The application uses the Bubble Tea framework to provide a two-panel interface
for viewing and managing permissions from:

- User level: `~/.claude/settings.json` (with chezmoi dotfiles support)
- Repo level: `{REPO}/.claude/settings.json`
- Local level: `{REPO}/.claude/settings.local.json`

## Development Requirements

- NEVER reinvent the wheel; always look for built-in or library-provided functionality to solve a
  problem.

### 🚨 MANDATORY: Code Quality Validation
**CRITICAL STEP**: After making ANY code changes, you MUST immediately run:
```bash
pre-commit run --files <modified-files>
```
- This step is **NON-NEGOTIABLE** and must be completed before considering any task finished
- Fix any formatting, linting, or quality issues that arise
- **IMPORTANT**: Use `--files` option to specify modified files directly
- **NEVER run mutating git commands** (like `git add`) without explicit user approval
- If pre-commit requires staged files, ask user for permission to stage them first
- **Failure to run this violates project requirements**

## Build and Development Commands

```bash
# Full verification and testing (for Claude Code automated testing)
# Uses --debug to avoid TTY issues, validates all functionality
scripts/verify.sh

# Interactive testing with TUI (for user/manual testing)
# Note: requires terminal TTY, not suitable for automated testing
scripts/test.sh

# Code quality and linting
golangci-lint run

# MANDATORY: Run after ANY code changes (specify modified files)
pre-commit run --files <modified-files>

# Manual build for production
go build -o claude-permissions .

# Run with debug output (shows loaded data and UI layout)
./claude-permissions --debug

# Run with custom test files
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
```

## UX Design & Workflow

### Two-Phase User Workflow
The application is designed around a **two-phase user workflow** that separates action planning from
execution:

1. **Phase 1 - Action Queuing** (Main Screen):
   - **Permissions Panel**: Select permissions to move between levels (User/Repo/Local)
   - **Duplicates Panel**: Resolve conflicts by choosing which level to keep
   - User navigates with TAB, selects items, queues up moves and resolutions
   - No immediate file modifications - everything is staged

2. **Phase 2 - Review & Execute** (Confirmation Screen):
   - **Full-screen summary**: Clean, comprehensive view of all queued actions
   - Shows duplicates to be removed, permissions to be moved, settings changes
   - User confirms with ENTER to execute all actions, or ESC to return
   - Only at this point are JSON files actually modified

### Design Rationale
- **Separation of concerns**: Planning vs execution are distinct mental tasks
- **Safety**: User reviews all changes before they happen
- **Clarity**: Dedicated confirmation screen reduces cognitive load
- **Vertical space**: Two-panel main screen fits better in terminal constraints

## Testing Scripts

### scripts/verify.sh
- **Purpose**: Automated testing and verification for Claude Code
- **Features**: Uses `--debug` flag to avoid TTY errors, tests data loading, error handling, and
  code structure
- **Usage**: Run from repo root with `./scripts/verify.sh`
- **Output**: Validates build, data parsing, error handling, and UI improvements

### scripts/test.sh
- **Purpose**: Interactive TUI testing for user/manual testing
- **Features**: Launches full interactive interface with sample data
- **Usage**: Run from any directory - auto-detects repo root
- **Note**: Requires TTY, not suitable for automated/CI testing

## Architecture

### Core Components

- **main.go**: Entry point, command-line parsing, and model initialization
- **models.go**: Core data structures (Settings, Permission, Duplicate, Action, Model)
- **settings.go**: Settings file loading, parsing, and git repository detection
- **ui.go**: Bubble Tea UI implementation with two-panel layout
- **actions.go**: Action queue system for permission moves/edits
- **styles.go**: Lipgloss styling definitions

### Key Data Flow

1. **Startup**: Load settings from all three levels, consolidate permissions, detect duplicates
2. **UI State**: Two panels (permissions, duplicates) with keyboard navigation
3. **Action Queue**: Queue moves/edits before applying, with preview and confirmation
4. **File Operations**: Only modify "allow" arrays in JSON files, preserve other settings

### TUI Design Patterns

- Two-panel layout with tab navigation
- Viewport scrolling for each panel
- Search mode with highlighting
- Confirmation dialogs for destructive operations
- ASCII status indicators for terminal compatibility

## Special Features

### Chezmoi Integration
The tool automatically detects and works with chezmoi dotfiles if:
1. `chezmoi` command is available on PATH
2. `chezmoi source-path ~/.claude/settings.json` returns valid path

### Git Repository Detection
Traverses parent directories from current working directory to find `.git/config` and determine
repo-level settings paths.

### Settings File Format
Only modifies the "allow" permission arrays in Claude Code settings JSON files. All other settings
remain untouched.

## Testing

The project includes comprehensive testing via `scripts/verify.sh` which checks:
- Data loading and parsing
- Error handling (invalid JSON, missing files)
- UI improvement implementations
- Code structure validations

Test data is available in `testdata/` directory with sample settings files for all three levels.

## Go Documentation

- ALWAYS use the godoc MCP for Go package documentation
- Use `mcp__godoc__get_doc` instead of web search for Go APIs
- Prefer godoc MCP for understanding Go standard library and third-party packages

## Go Project Layout Guidelines

Based on [appliedgo.com/blog/go-project-layout](https://appliedgo.com/blog/go-project-layout):

### Core Principles
- **Start simple** - Begin with minimal structure, add complexity only when needed
- **No universal standard** - Adapt to your project's specific needs
- **Name by functionality** - Avoid generic names like "util" or "helpers"

### Directory Structure
- **Avoid `src/`** - Don't use this directory
- **`internal/`** - For packages not accessible outside the module
- **`testdata/`** - For ancillary test data (ignored by Go toolchain)
- **`vendor/`** - Optional for downloaded dependencies
- **`cmd/`** - Consider for executables in larger projects
- **`pkg/`** - Optional for public packages (use with caution)

### Project Types
- **Small/CLI tools**: Single directory with `main.go`
- **Libraries**: Root-level packages, separate dirs for sub-packages
- **Large projects**: Consider `cmd/`, potentially `pkg/`, use `internal/`

### When to Add Structure
- Group code by responsibility when root becomes cluttered
- Add subdirectories only when project complexity demands it
- Prioritize clarity and simplicity over rigid conventions

## Requirements

- Go 1.21+
- Terminal with ANSI color support
- Optional: chezmoi for dotfiles integration

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

## Project Overview

This is a Go-based interactive TUI tool for managing Claude Code tool permissions across different
settings levels. The application uses the Bubble Tea framework to provide a two-panel interface for
viewing and managing permissions from:

- User level: `~/.claude/settings.json` (with chezmoi dotfiles support)
- Repo level: `{REPO}/.claude/settings.json`
- Local level: `{REPO}/.claude/settings.local.json`

## Development Requirements

- Implement go code in an idiomatic way.
- NEVER reinvent the wheel; always look for built-in or library-provided functionality to solve a
  problem.
- Do not make changes to `.golangci.yml` without searching the docs first:
  <https://golangci-lint.run/usage/linters/#revive>
- Assume user has very little to no knowledge of golang and related tooling.
- Claude Code will be implementing the entirety of this tool, so it shall not be assumed the user is
  familiar with current architecture or implementation. Explanations and background will need to be
  provided.
- `pre-commit run --files` MUST be executed at stopping points during or after iterations of work.
  Any issues found MUST be corrected. This command will be re-run until there are zero issues
  remaining.

## Build and Development Commands

```bash
# Manual build for production
go build -o claude-permissions .

# Run with debug server for development/debugging
./claude-permissions --debug-server --debug-port=8080

# Run with custom test files
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"

# Hot reload development (human/user only - requires TTY)
scripts/dev.sh
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

## Architecture

### Core Components

- **main.go**: Entry point, command-line parsing, model initialization, and tea.Model wrapper
- **types/model.go**: Core data structures (Settings, Permission, Duplicate, Action, Model)
- **settings.go**: Settings file loading, parsing, and git repository detection
- **ui.go**: Bubble Tea UI implementation with two-panel layout
- **actions.go**: Action queue system for permission moves/edits
- **styles.go**: Lipgloss styling definitions
- **interfaces.go**: Interface definitions for application components
- **delegate.go**: Custom list delegate for permissions display
- **logging.go**: Logging utilities and no-op handler
- **debug/**: HTTP debug server package for development and debugging
  - `server.go`: HTTP server setup and endpoint routing
  - `state.go`: Application state inspection and model access
  - `capture.go`: Screen content capture and snapshot functionality
  - `input.go`: Input simulation for testing and debugging
  - `layout.go`: Layout diagnostics and inspection
  - `logging.go`: Debug event logging system
  - `slog_handler.go`: Custom slog handler for debug integration
- **layout/**: Layout engine for responsive terminal UI components
  - `engine.go`: Core layout calculation engine
  - `calculator.go`: Dimension and positioning calculations
  - `components.go`: UI component layout definitions
  - `constraints.go`: Layout constraint system
  - `events.go`: Layout event handling
  - `integration.go`: Integration with Bubble Tea framework
  - `debug.go`: Layout debugging utilities

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

### Terminal UI Development Notes

- When calculating line widths, use `wc -m` not `wc -c`. The latter counts bytes not characters, which causes issues with Unicode characters used in TUI rendering
- `wc -m` includes the final carriage return, so subtract 1 from the result

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

## Debug Server

The application includes an HTTP debug server for development and debugging:

- **Purpose**: Real-time inspection of application state, layout diagnostics, and input simulation
- **Activation**: Use `--debug-server` flag with optional `--debug-port` (default: 8080)
- **Thread-safe**: Uses direct field access with proper mutex locking, no reflection
- **Interface**: Use `scripts/debug-api.sh` script (designed for Claude Code usage)
- **Status**: **EXPERIMENTAL/WIP** - Debug endpoints are new and may contain bugs; always consider that debug tooling itself might be faulty when diagnosing issues

### Development Workflow with Debug API

**CRITICAL: Claude Code Debug Workflow Protocol**

Claude Code MUST follow this exact protocol when debugging or developing:

1. **NEVER run `scripts/dev.sh` directly** - This requires TTY and is user-only
2. **ALWAYS check server status first** - Run `scripts/debug-api.sh health` before any debug operations
3. **If health check fails** - Ask user to run `scripts/dev.sh` and wait for confirmation
4. **Only proceed with debug operations after health check passes**

**Standard Development Process:**
1. **User runs `scripts/dev.sh`** to start live reload development server
2. **Claude checks health** - ALWAYS run `scripts/debug-api.sh health` first
3. **Claude makes code changes** - dev.sh automatically rebuilds and restarts application
4. **Claude verifies live reload** - Check `scripts/debug-api.sh health` after each change
5. **Claude uses debug API** - Inspect application state, test functionality, and diagnose issues
6. **Iterate** - repeat steps 2-5 until task is complete

**Debug Server Dependency Rule:**
- All debug operations (state, logs, snapshot, input) require the debug server to be running
- If any debug API call fails, immediately check health and ask user to restart dev.sh if needed
- Never attempt to start the server yourself - always request user to do it

This workflow ensures rapid development with real-time feedback and debugging capabilities.

### Debug API Script

Use `scripts/debug-api.sh` for easy API access:

```bash
# Check debug server status
scripts/debug-api.sh health

# Get complete application state
scripts/debug-api.sh state

# Get layout diagnostics
scripts/debug-api.sh layout

# Capture screen content (with ANSI codes stripped)
scripts/debug-api.sh snapshot --raw

# Get debug event logs (returns all entries and clears buffer)
scripts/debug-api.sh logs

# Send key inputs to the application
scripts/debug-api.sh input tab
scripts/debug-api.sh input enter
scripts/debug-api.sh input up
scripts/debug-api.sh input a

# Reset application state
scripts/debug-api.sh reset

# Use custom host/port
scripts/debug-api.sh state --host localhost --port 8081
```

**Supported key inputs**: `tab`, `enter`, `escape`/`esc`, `up`, `down`, `left`, `right`, `space`,
`a`, `u`, `r`, `l`, `e`, `c`, `q`, `/`

**Architecture**: Performance-optimized with compile-time type checking and proper thread
synchronization for concurrent access to model state.

## Testing

Test data is available in `testdata/` directory with sample settings files for all three levels
(user-settings.json, repo-settings.json, local-settings.json). Use these files with the
`--user-file`, `--repo-file`, and `--local-file` flags for testing different scenarios.

## Go Documentation

- ALWAYS use the godoc MCP for Go package documentation
- Use `mcp__godoc__get_doc` instead of web search for Go APIs
- Prefer godoc MCP for understanding Go standard library and third-party packages

## Go Project Layout

This project follows a simple CLI tool structure with `main.go` in the root and specialized packages
(`debug/`, `layout/`, `types/`) for focused functionality. The `testdata/` directory contains test
files, following Go toolchain conventions.

## Requirements

- Go 1.24+
- Terminal with ANSI color support
- Optional: chezmoi for dotfiles integration

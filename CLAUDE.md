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

1. **Phase 1 - Change Planning** (Main Screen):
   - **Permissions Panel**: Move permissions between levels (User/Repo/Local) immediately
   - **Duplicates Panel**: Resolve conflicts by choosing which level to keep
   - User navigates with TAB, selects items, makes immediate changes to model state
   - Changes are applied to in-memory model immediately

2. **Phase 2 - Review & Save** (Confirmation Screen):
   - **Full-screen summary**: Clean, comprehensive view of all pending changes
   - Shows duplicates to be removed, permissions that were moved, settings changes
   - User confirms with ENTER to save changes to disk, or ESC to return
   - Only at this point are JSON files actually written

### Design Rationale

- **Separation of concerns**: Planning vs execution are distinct mental tasks
- **Safety**: User reviews all changes before they happen
- **Clarity**: Dedicated confirmation screen reduces cognitive load
- **Vertical space**: Two-panel main screen fits better in terminal constraints

## CRITICAL: Pure Lipgloss Architecture

**ALWAYS use pure Bubble Tea + Lipgloss patterns.** All UI rendering uses industry-standard TUI patterns:

- ✅ REQUIRED: `lipgloss.JoinVertical()`, `lipgloss.JoinHorizontal()` for layout composition
- ✅ REQUIRED: Dynamic sizing using `lipgloss.Width()` and `lipgloss.Height()` best practices
- ✅ REQUIRED: Centralized theme system in `ui/theme.go` for consistent styling
- ❌ FORBIDDEN: Custom layout engines, manual dimension calculations, reinventing the wheel

**Modals/overlays:** Use `lipgloss.Place()` for absolute positioning

**This architecture follows industry-standard TUI application patterns.**

## Architecture

### Core Components

- **main.go**: Entry point, command-line parsing, model initialization, and tea.Model wrapper
- **types/model.go**: Core data structures (Settings, Permission, Duplicate, Model)
- **settings.go**: Settings file loading, parsing, and git repository detection
- **ui/**: Pure Bubble Tea + Lipgloss UI module using industry-standard patterns
  - `main.go`: Core UI rendering logic with `lipgloss.JoinVertical()` composition
  - `components.go`: UI components (header, footer, content) with dynamic sizing
  - `helpers.go`: Key handling and modal rendering using pure state management
  - `theme.go`: Centralized color palette and style definitions
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

### Key Data Flow

1. **Startup**: Load settings from all three levels, consolidate permissions, detect duplicates
2. **UI State**: Two panels (permissions, duplicates) with keyboard navigation
3. **Immediate Changes**: Permission moves and duplicate resolutions happen immediately in memory
4. **File Operations**: Only modify "allow" arrays in JSON files, preserve other settings

### TUI Design Patterns

- **Pure Lipgloss Composition**: Using `lipgloss.JoinVertical()` and `lipgloss.JoinHorizontal()`
- **Dynamic Sizing**: `lipgloss.Width()` and `lipgloss.Height()` for responsive layouts
- **Centralized Theming**: All colors and styles defined in `ui/theme.go`
- **Component Architecture**: Header, content, status bar, footer as separate components
- **Two-panel navigation**: TAB switching between duplicates and organization screens
- **Three-column layout**: Local/Repo/User permission organization
- **Context-sensitive UI**: Hotkeys and status information change based on current screen

### Terminal UI Development Notes

- **Use Dynamic Sizing**: Always use `lipgloss.Width()` and `lipgloss.Height()` instead of manual calculations
- **Account for Borders/Padding**: When calculating available space, subtract border + padding overhead
- **Centralized Colors**: Use theme constants from `ui/theme.go` instead of hardcoded color values
- **Component-based**: Each UI section (header, content, status, footer) is a separate component
- **Responsive Layout**: Columns automatically adjust to terminal width using `c.width / 3` pattern

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
- **Status**: **EXPERIMENTAL/WIP** - Debug endpoints are new and may contain bugs; always consider
  that debug tooling itself might be faulty when diagnosing issues

### Development Workflow with Debug API

**CRITICAL: Claude Code Debug Workflow Protocol**

Claude Code MUST follow this exact protocol when debugging or developing:

1. **NEVER run `scripts/dev.sh` directly** - This requires TTY and is user-only
1. **NEVER** run the `claude-permissions` executable directly -- it requires TTY which Claude Code
   has no access to.
3. **If any debug endpoint fails** - Ask user to run `scripts/dev.sh` and wait for confirmation

**Standard Development Process:**
1. **User runs `scripts/dev.sh`** to start live reload development server
3. **Claude makes code changes** - dev.sh automatically rebuilds and restarts application
5. **Claude uses debug API** - Inspect application state, test functionality, and diagnose issues
6. **Iterate** - repeat steps 2-5 until task is complete

**Debug Server Dependency Rule:**
- All debug operations (state, logs, snapshot, input) require the debug server to be running
- If any debug API call fails, immediately ask user to restart dev.sh if needed
- Never attempt to start the server yourself - always request user to do it

This workflow ensures rapid development with real-time feedback and debugging capabilities.

### Debug API Script

Use `scripts/debug-api.sh` for easy API access:

```bash
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
(`debug/`, `ui/`, `types/`) for focused functionality. The `testdata/` directory contains test
files, following Go toolchain conventions.

### UI Architecture

The `ui/` package implements pure Bubble Tea + Lipgloss patterns:
- **Industry-standard composition**: Uses `lipgloss.JoinVertical()` and `lipgloss.JoinHorizontal()`
- **Component-based**: Header, content, status bar, footer as separate components
- **Centralized theming**: All colors and styles in `ui/theme.go`
- **Dynamic sizing**: Responsive layouts using lipgloss best practices

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

### Library-First Development Mandate

- **CRITICAL**: NEVER reinvent the wheel; always look for built-in or library-provided functionality
  to solve a problem.
- **ALWAYS** check library documentation first before implementing any functionality
- **PREFER** composition over custom abstraction - wrap library functions rather than replacing them
- **CREATE** thin helper functions for consistency only, never full custom implementations

### Code Quality & Standards

- Implement go code in an idiomatic way.
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

## Duplicate Resolution Workflow

### Auto-Selection by Priority

Duplicates are automatically pre-selected based on level priority (User > Repo > Local):

- "Bash" at LOCAL and USER → AUTO-SELECT USER
- "Bash" at LOCAL and REPO → AUTO-SELECT REPO
- "Bash" at LOCAL, REPO, and USER → AUTO-SELECT USER

This makes duplicate resolution "hands-free" by default, requiring minimal user intervention.

### Two-Phase Resolution Process

1. **Selection Phase**: Duplicates show with auto-selected KeepLevel, user can change with 1/2/3
2. **Commitment Phase**: User hits ENTER → confirmation modal → actual file updates

### Blocking Logic

- Organization screen is BLOCKED while `len(m.Duplicates) > 0`
- Duplicates are considered "unresolved" until committed to files (not just assigned KeepLevel)
- After successful commit, duplicates are removed from model and organization screen becomes
  accessible

### State Functions

- `hasUnresolvedDuplicates()`: Returns `true` if ANY duplicates exist in model (need commitment)
- `hasPendingChanges()`: Returns `true` if duplicates have assigned KeepLevel (ready for commit)

## CRITICAL: Library-First Architecture

**ALWAYS leverage existing library functionality before implementing custom solutions.** This
project follows a strict library-first approach:

### Core Principles

- ✅ REQUIRED: Always use library-provided functionality before implementing custom solutions
- ✅ REQUIRED: Leverage existing APIs as the primary interface (e.g., lipgloss for layout, Bubble Tea
  for state)
- ✅ REQUIRED: Create thin wrapper functions only for consistency, not custom abstractions
- ❌ FORBIDDEN: Builder patterns or complex abstractions that duplicate library functionality
- ❌ FORBIDDEN: Custom implementations of functionality already provided by dependencies

### Lipgloss-Specific Requirements

**ALWAYS use pure Bubble Tea + Lipgloss patterns.** All UI rendering uses industry-standard TUI
patterns:

- ✅ REQUIRED: `lipgloss.JoinVertical()`, `lipgloss.JoinHorizontal()` for layout composition
- ✅ REQUIRED: Dynamic sizing using `lipgloss.Width()` and `lipgloss.Height()` best practices
- ✅ REQUIRED: Centralized theme system in `ui/theme.go` for consistent styling
- ❌ FORBIDDEN: Custom layout engines, manual dimension calculations, reinventing the wheel

**Modals/overlays:** Use `lipgloss.Place()` for absolute positioning

### Theme Architecture: Nuanced Centralization

**CRITICAL: Follow nuanced centralization principles - not everything belongs in theme.go.**

#### Centralization Decision Rules

**✅ CENTRALIZE in `ui/theme.go`:**

- **Color palette constants** - used across multiple components
- **Typography scales** - font sizes, weights used by 2+ components
- **Spacing tokens** - standard margins, padding, border widths
- **Core interaction states** - focused, selected, disabled styles used across components
- **Genuinely reusable patterns** - used by 2+ unrelated components

**❌ KEEP LOCAL in component files:**

- **Component-specific styling** - unique to one screen/component
- **Complex layout logic** - doesn't generalize to other components
- **Single-use variations** - not part of the design system

#### The Key Principle

**Centralize design tokens and genuinely reusable patterns, not every style declaration.**

**Decision rule**: A style earns centralization when it's used by 2+ unrelated components OR
represents a core design decision (like color palette).

#### Examples of Proper Theme Architecture

**✅ Good - Proper Centralization:**

```go
// ui/theme.go - Centralized design tokens
const (
    ColorAccent = "#38BDF8"  // Used across multiple components
    ColorTitle  = "15"       // Used in headers, modals, etc.
)

var AccentStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color(ColorAccent)).
    Bold(true)

// ui/modals.go - Component uses centralized tokens
titleStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color(ColorTitle)).  // ✅ Uses theme constant
    Align(lipgloss.Center).
    Width(contentWidth - 4)                  // ❌ Component-specific calculation
```

**❌ Bad - Over-Centralization:**

```go
// ui/theme.go - Don't put single-use styles here
var SpecificModalTitleStyle = lipgloss.NewStyle().  // ❌ Single-use style
    Width(56).                                      // ❌ Component-specific width
    Padding(1, 2).                                  // ❌ Component-specific padding
    Bold(true)

// ui/modals.go - Component forced to use over-specific style
title := SpecificModalTitleStyle.Render("Title")   // ❌ Not flexible
```

**✅ Good - Component-Specific Styling:**

```go
// ui/modals.go - Keep component-specific styling local
modalStyle := lipgloss.NewStyle().
    Width(contentWidth).                            // ✅ Component-specific
    Foreground(lipgloss.Color(ColorTitle)).         // ✅ Uses theme token
    Padding(1, 2)                                   // ✅ Component-specific
```

#### Claude Code Directive

**PROACTIVELY look for centralization opportunities during feature work:**

- Spot duplicated styling patterns across components
- Extract genuinely reusable styles to theme.go
- Replace hardcoded colors with theme constants
- Document reasoning when keeping styles local vs centralizing

#### Real-World Centralization Exemplar

**✅ Excellent - BlockingMessageStyle Implementation:**

```go
// ui/theme.go - Centralized core styling for blocking messages
var BlockingMessageStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color(ColorBorderFocused)).
    Padding(1).           // Core layout shared across components
    Align(lipgloss.Center, lipgloss.Center)

// ui/components.go - Components customize layout but reuse core styling
return BlockingMessageStyle.
    Width(contentWidth).   // ✅ Component-specific dimensions
    Height(c.height).      // ✅ Component-specific dimensions
    Render(message)        // ✅ Component-specific content
```

**Why this works well:**

- **Core styling centralized**: Border, colors, padding, alignment consistent across components
- **Layout flexibility preserved**: Components control width, height, content
- **Genuinely reusable**: Used by duplicates panel and organization panel for similar purposes
- **Easy to maintain**: Change border style once, affects all blocking messages

### Examples of Library-First Approach

**✅ Good - Library-First:**

```go
// Use lipgloss composition directly
content := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)
footer := lipgloss.JoinVertical(lipgloss.Left, row1, row2)

// Use lipgloss styling directly
style := lipgloss.NewStyle().Align(lipgloss.Center).Width(width)
```

**❌ Bad - Custom Abstraction:**

```go
// Don't create custom layout managers
layoutManager := NewLayoutManager()
layoutManager.AddColumn(col1).AddColumn(col2)
content := layoutManager.Build()

// Don't create custom styling systems
footer := NewFooterBuilder().AddAction("ENTER", "Save").Build()
```

**This architecture follows industry-standard TUI application patterns with zero custom layout
logic.**

## Architecture

### Core Components

- **main.go**: Entry point, command-line parsing, model initialization, and tea.Model wrapper
- **types/**: Core data structures and modal definitions
  - `model.go`: Core data structures (Settings, Permission, Duplicate, Model)
  - `modal.go`: Modal state and type definitions
- **settings.go**: Settings file loading, parsing, and git repository detection
- **ui/**: Pure Bubble Tea + Lipgloss UI module using industry-standard patterns
  - `main.go`: Core UI rendering logic with `lipgloss.JoinVertical()` composition
  - `components.go`: UI components (header, footer, content) with dynamic sizing
  - `helpers.go`: Key handling and navigation utilities
  - `modals.go`: Modal rendering and interaction logic
  - `theme.go`: Centralized color palette and style definitions
- **logging.go**: Logging utilities and no-op handler
- **debug/**: HTTP debug server package for development and debugging
  - `server.go`: HTTP server setup and endpoint registration system
  - `utils.go`: Shared utilities (JSON responses, timestamps, conversions)
  - `logger.go`: Logging infrastructure and event tracking
  - Endpoint files (self-registering pattern):
    - `endpoint-health.go`: Health check endpoint
    - `endpoint-state.go`: Application state inspection
    - `endpoint-snapshot.go`: Screen capture and layout diagnostics
    - `endpoint-input.go`: Input injection with state analysis
    - `endpoint-logs.go`: Debug event logs retrieval
    - `endpoint-reset.go`: Application state reset
    - `endpoint-launch-confirm-changes.go`: Screen testing with mock data

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

- **Use Dynamic Sizing**: Always use `lipgloss.Width()` and `lipgloss.Height()` instead of manual
  calculations
- **Account for Borders/Padding**: When calculating available space, subtract border + padding
  overhead
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

#### Claude Code Debug Workflow Protocol

Claude Code MUST follow this exact protocol when debugging or developing:

1. **NEVER run `scripts/dev.sh` directly** - This requires TTY and is user-only
1. **NEVER** run the `claude-permissions` executable directly -- it requires TTY which Claude Code
   has no access to.
1. **ASSUME the debug server is always running** - Proceed directly with debug API calls
1. **ONLY if a debug endpoint fails with connection error** - Ask user to run `scripts/dev.sh`

#### Standard Development Process

1. **Claude makes code changes** - Assume dev.sh is running and will auto-rebuild
1. **Claude uses debug API immediately** - Don't ask permission, just use the debug endpoints
1. **If endpoint fails** - Then and only then ask user to start/restart `scripts/dev.sh`
1. **Iterate** - repeat steps 1-3 until task is complete

#### Debug Server Dependency Rule

- **Default assumption**: Debug server is running and ready for API calls
- **Never preemptively ask** user to start the server before making API calls
- **Only request server start** if debug API calls return connection failures
- Never attempt to start the server yourself - always request user to do it

This workflow ensures rapid development with real-time feedback and debugging capabilities.

### Debug API Script

Use `scripts/debug-api.sh` for easy API access:

```bash
# Get complete application state
scripts/debug-api.sh state

# Capture screen content (ANSI codes stripped by default)
scripts/debug-api.sh snapshot

# Capture screen content with ANSI color codes
scripts/debug-api.sh snapshot --color

# Get debug event logs (returns all entries and clears buffer)
scripts/debug-api.sh logs

# Send key inputs to the application
scripts/debug-api.sh input tab
scripts/debug-api.sh input enter
scripts/debug-api.sh input up
scripts/debug-api.sh input a

# Reset application state
scripts/debug-api.sh reset

# Load settings from different files
scripts/debug-api.sh load-settings --user-file testdata/user-no-duplicates.json --repo-file testdata/repo-no-duplicates.json --local-file testdata/local-no-duplicates.json

# Use custom host/port
scripts/debug-api.sh state --host localhost --port 8081
```

**Supported key inputs**: `tab`, `enter`, `escape`/`esc`, `up`, `down`, `left`, `right`, `space`,
`a`, `u`, `r`, `l`, `e`, `c`, `q`, `/`

**Architecture**: Performance-optimized with compile-time type checking and proper thread
synchronization for concurrent access to model state.

### Debug File Organization

The debug package follows a self-registering endpoint pattern for clean separation:

- **Endpoint files**: `endpoint-*.go` - Self-registering HTTP endpoint handlers
- **Infrastructure files**: Core debug server components (`server.go`, `utils.go`, `logger.go`)

See `debug/CLAUDE.md` for detailed debug package architecture and patterns.

### Debug Endpoint Patterns

Debug endpoints follow consistent naming patterns:

- **Screen testing**: `/launch-<screen_name>` - Launch specific screens with mock data for testing
  - `/launch-confirm-changes` - Launch confirmation screen with mock permission moves and duplicate
    resolutions
  - Future: `/launch-duplicates`, `/launch-organization`, etc.
- **State inspection**: `/state`, `/snapshot`, `/logs` - Inspect current application state
- **Interaction**: `/input`, `/reset` - Send inputs or reset application state
- **Data loading**: `/load-settings` - Dynamically load settings from specified file paths

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
(`debug/`, `ui/`, `types/`) for focused functionality. Additional support includes:

- `testdata/`: Sample settings files for testing different scenarios
- `scripts/`: Development and debugging utilities (`dev.sh`, `debug-api.sh`)
- `docs/`: Project documentation and guides

### UI Architecture

The `ui/` package implements pure Bubble Tea + Lipgloss patterns:

- **Industry-standard composition**: Uses `lipgloss.JoinVertical()` and `lipgloss.JoinHorizontal()`
- **Component-based**: Header, content, status bar, footer as separate components
- **Centralized theming**: All colors and styles in `ui/theme.go`
- **Dynamic sizing**: Responsive layouts using lipgloss best practices

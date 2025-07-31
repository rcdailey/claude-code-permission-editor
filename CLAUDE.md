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

## Essential Commands

```bash
# CRITICAL: Claude Code Development - NEVER run the executable directly
# The executable requires TTY which Claude Code cannot provide
# ALWAYS use debug API exclusively through scripts/debug-api.sh

# Build (for user to run)
go build -o claude-permissions .

# IMPORTANT: Run pre-commit at stopping points - YOU MUST fix all issues
pre-commit run --files

# Debug API - Claude's ONLY feedback mechanism (assume server is running)
scripts/debug-api.sh state          # Get application state
scripts/debug-api.sh snapshot       # Screen capture
scripts/debug-api.sh input tab      # Send input
scripts/debug-api.sh logs           # Get debug events
scripts/debug-api.sh reset          # Reset state

# Development server (user only - requires TTY)
scripts/dev.sh

# Search with ripgrep (ALWAYS use rg instead of grep)
rg "pattern" --type go
rg "function.*Name" --type go -A 3 -B 1
```

## CRITICAL: Library-First Development Mandate

**YOU MUST NEVER reinvent the wheel - always use existing library functionality.**

### Core Principles

- **IMPORTANT**: ALWAYS check library documentation first before implementing any functionality
- **NEVER** create custom implementations of functionality already provided by dependencies
- **ALWAYS** use library-provided functionality as the primary interface
- **PREFER** composition over custom abstraction - wrap library functions, don't replace them
- **CREATE** thin helper functions for consistency only, never full custom implementations

### Lipgloss-Specific Requirements

**YOU MUST use pure Bubble Tea + Lipgloss patterns for all UI rendering:**

- **REQUIRED**: `lipgloss.JoinVertical()`, `lipgloss.JoinHorizontal()` for layout composition
- **REQUIRED**: Dynamic sizing using `lipgloss.Width()` and `lipgloss.Height()` best practices
- **REQUIRED**: Centralized theme system in `ui/theme.go` for consistent styling
- **FORBIDDEN**: Custom layout engines, manual dimension calculations, reinventing the wheel

**For modals/overlays:** ALWAYS use `lipgloss.Place()` for absolute positioning

### Theme Architecture: Nuanced Centralization

**IMPORTANT: Follow nuanced centralization principles - not everything belongs in theme.go.**

**CENTRALIZE in `ui/theme.go`:**

- Color palette constants used across multiple components
- Typography scales and spacing tokens used by 2+ components
- Core interaction states (focused, selected, disabled) used across components
- Genuinely reusable patterns used by 2+ unrelated components

**KEEP LOCAL in component files:**

- Component-specific styling unique to one screen/component
- Complex layout logic that doesn't generalize
- Single-use variations not part of the design system

**Decision rule**: A style earns centralization when it's used by 2+ unrelated components OR
represents a core design decision (like color palette).

## Code Quality & Standards

- **IMPORTANT**: Implement Go code in idiomatic way
- **NEVER** make changes to `.golangci.yml` without searching docs first:
  <https://golangci-lint.run/usage/linters/#revive>
- **YOU MUST** run `pre-commit run --files` at stopping points - fix ALL issues before continuing
- Assume user has minimal Go knowledge - provide explanations and background
- **ALWAYS** use `mcp__godoc__get_doc` for Go package documentation instead of web search

## UX Design & Workflow

### Two-Phase User Workflow

**IMPORTANT**: The application uses a two-phase workflow that separates action planning from
execution:

1. **Phase 1 - Change Planning** (Main Screen):
   - Permissions Panel: Move permissions between levels immediately
   - Duplicates Panel: Resolve conflicts by choosing which level to keep
   - Changes applied to in-memory model immediately

2. **Phase 2 - Review & Save** (Confirmation Screen):
   - Full-screen summary of all pending changes
   - User confirms with ENTER to save to disk, or ESC to return
   - **IMPORTANT**: Only at this point are JSON files actually written

### Design Rationale: Safety, clarity, separation of concerns, optimal terminal space usage

## Duplicate Resolution Workflow

**IMPORTANT**: Duplicates auto-selected by priority (User > Repo > Local) for hands-free resolution.

### Two-Phase Resolution

1. **Selection Phase**: Auto-selected KeepLevel, user can change with 1/2/3
2. **Commitment Phase**: ENTER → confirmation modal → file updates

### Blocking Logic

- **Organization screen BLOCKED while `len(m.Duplicates) > 0`**
- Duplicates unresolved until committed to files (not just assigned KeepLevel)
- After commit, duplicates removed and organization screen accessible

### Key State Functions

- `hasUnresolvedDuplicates()`: ANY duplicates exist (need commitment)
- `hasPendingChanges()`: Duplicates have assigned KeepLevel (ready for commit)

## Architecture Overview

### Core Components

- **main.go**: Entry point, CLI parsing, model initialization
- **types/**: Core data structures (Settings, Permission, Duplicate, Model)
- **settings.go**: Settings file loading, parsing, git repository detection
- **ui/**: Pure Bubble Tea + Lipgloss UI module
  - `main.go`: Core UI rendering with `lipgloss.JoinVertical()` composition
  - `components.go`: UI components with dynamic sizing
  - `theme.go`: Centralized color palette and style definitions
- **debug/**: HTTP debug server (self-registering endpoint pattern)

### Key Data Flow

1. Load settings from all levels → consolidate permissions → detect duplicates
2. Two panels (permissions, duplicates) with keyboard navigation
3. Immediate in-memory changes → confirmation screen → file operations

### TUI Design Patterns

- **Pure Lipgloss Composition**: `lipgloss.JoinVertical()` and `lipgloss.JoinHorizontal()`
- **Dynamic Sizing**: `lipgloss.Width()` and `lipgloss.Height()` for responsive layouts
- **Centralized Theming**: All colors in `ui/theme.go`
- **Component Architecture**: Header, content, status bar, footer as separate components

## Special Features

- **Chezmoi Integration**: Auto-detects chezmoi dotfiles for user settings
- **Git Repository Detection**: Traverses parent directories to find `.git/config`
- **Settings File Format**: **IMPORTANT** - Only modifies "allow" arrays, preserves other settings

## Debug Server

**CRITICAL**: Claude Code MUST use debug API exclusively - the executable requires TTY which Claude
Code cannot provide.

### Claude Code Debug Protocol

**YOU MUST follow this exact workflow:**

1. **NEVER** run the executable directly - it requires TTY which Claude Code lacks
2. **NEVER** run `scripts/dev.sh` directly - this is user-only
3. **ALWAYS** assume debug server is running - proceed directly with API calls
4. **ONLY** ask user to run `scripts/dev.sh` if debug endpoints fail with connection error

### REQUIRED Development Process

**Claude Code Development Loop:**

1. Make code changes (assume `scripts/dev.sh` auto-rebuilds)
2. **IMMEDIATELY** use `scripts/debug-api.sh` for feedback
3. Iterate based on debug API responses
4. **NEVER** attempt direct executable interaction

**Status**: EXPERIMENTAL - debug tooling may contain bugs

## Testing & Project Layout

- **Test data**: `testdata/` directory with sample settings files for all levels
- **Scripts**: `scripts/dev.sh` (development), `scripts/debug-api.sh` (API access)
- **Simple CLI structure**: `main.go` in root, specialized packages (`debug/`, `ui/`, `types/`)

**IMPORTANT**: See `debug/CLAUDE.md` for detailed debug package architecture and self-registering
endpoint patterns.

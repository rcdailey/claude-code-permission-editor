# Contributing to Claude Code Permission Editor

## Development Setup

### Prerequisites

- Go 1.21+
- Terminal supporting ANSI colors

### Build and Development Commands

```bash
# Build for production
go build -o claude-permissions .

# Hot reload development (requires TTY)
scripts/dev.sh

# Run with debug server for development/debugging
./claude-permissions --debug-server --debug-port=8080
```

## Development Workflow

### Hot Reload Development

For the best development experience, use the hot reload script:

```bash
# Start development server with automatic rebuilds
scripts/dev.sh
```

This script:

- Watches for file changes
- Automatically rebuilds the application
- Restarts the TUI when changes are detected
- Provides real-time feedback during development

### Manual Development

If you prefer manual control:

```bash
# Build and run manually
go build -o claude-permissions .
./claude-permissions

# Or with test data
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
```

### Debug Server

The application includes an HTTP debug server for inspecting application state during development:

```bash
# Start with debug server enabled
./claude-permissions --debug-server --debug-port=8080
```

The debug server provides endpoints for:

- Real-time application state inspection
- Layout diagnostics and debugging
- Input simulation for testing
- Screen content capture

#### Debug API

Use the `scripts/debug-api.sh` helper script to interact with the debug server:

```bash
# View current application state
scripts/debug-api.sh state

# Get layout diagnostics
scripts/debug-api.sh layout

# Capture current screen content
scripts/debug-api.sh snapshot --raw

# View debug event logs
scripts/debug-api.sh logs

# Simulate key inputs for testing
scripts/debug-api.sh input tab
scripts/debug-api.sh input enter
scripts/debug-api.sh input up

# Reset application state
scripts/debug-api.sh reset
```

**Note**: The debug server is experimental and primarily useful for development and automated testing.

## Architecture

### Core Components

- **main.go**: Entry point, command-line parsing, model initialization
- **types/model.go**: Core data structures (Settings, Permission, Duplicate, Model)
- **settings.go**: Settings file loading, parsing, git repository detection
- **ui/**: Pure Bubble Tea + Lipgloss UI module
  - `main.go`: Core UI rendering logic
  - `components.go`: UI components (header, footer, content)
  - `helpers.go`: Key handling and modal rendering
  - `theme.go`: Centralized color palette and style definitions
- **debug/**: HTTP debug server package

### UI Architecture

The `ui/` package implements pure Bubble Tea + Lipgloss patterns:

- **Industry-standard composition**: Uses `lipgloss.JoinVertical()` and `lipgloss.JoinHorizontal()`
- **Component-based**: Header, content, status bar, footer as separate components
- **Centralized theming**: All colors and styles in `ui/theme.go`
- **Dynamic sizing**: Responsive layouts using lipgloss best practices

### Key Data Flow

1. **Startup**: Load settings from all three levels, consolidate permissions, detect duplicates
2. **UI State**: Two panels (permissions, duplicates) with keyboard navigation
3. **Immediate Changes**: Permission moves and duplicate resolutions happen immediately in memory
4. **File Operations**: Only modify "allow" arrays in JSON files, preserve other settings

## Testing

Test data is available in `testdata/` directory with sample settings files for all three levels. Use
these files with the `--user-file`, `--repo-file`, and `--local-file` flags for testing different
scenarios.

```bash
# Test with sample data
./claude-permissions \
  --user-file="testdata/user-settings.json" \
  --repo-file="testdata/repo-settings.json" \
  --local-file="testdata/local-settings.json"
```

## Code Quality

This project uses pre-commit hooks to maintain code quality:

```bash
# Run pre-commit checks manually
pre-commit run --files <file1> <file2>

# Run all checks
pre-commit run --all-files
```

The pre-commit configuration includes:

- Go formatting (gofumpt)
- Linting (golangci-lint)
- Line length formatting (golines)
- Markdown linting (markdownlint-cli2)
- General file cleanup

Please ensure all pre-commit checks pass before submitting pull requests.

## Special Features

### Chezmoi Integration

The tool automatically detects and works with chezmoi dotfiles if:

1. `chezmoi` command is available on PATH
2. `chezmoi source-path ~/.claude/settings.json` returns valid path

### Git Repository Detection

Traverses parent directories from current working directory to find `.git/config` and determine
repo-level settings paths.

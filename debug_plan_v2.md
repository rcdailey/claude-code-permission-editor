# Debug Server Refactor Plan v2: Remove Reflection, Use Direct Field Access

## Context and Background

### Current State
The debug server feature was implemented using Go reflection to extract state from the main application's `Model` struct. This was done to avoid circular imports between the `debug` package and the `main` package.

### Problem with Current Approach
- **Performance overhead**: Reflection is slower than direct field access
- **Runtime errors**: Field access failures only discovered at runtime, not compile time
- **Complex code**: High cyclomatic complexity due to reflection logic
- **Poor IDE support**: No autocomplete, go-to-definition, or refactoring support
- **Maintenance burden**: Difficult to debug and modify reflection-based code

### Why Reflection Was Used Initially
The debug server needed to access the `Model` struct fields, but:
- Debug package can't import main package (circular dependency)
- Model struct is defined in main package
- Original solution: Use `interface{}` and reflection to inspect at runtime

### New Approach: Shared Types Package
Move all types to a shared `types` package that both `main` and `debug` can import:
- `main` imports `types` and uses `types.Model`
- `debug` imports `types` and uses `types.Model`
- No circular dependency: both import the same shared package
- No reflection needed: direct field access with `model.ActivePanel`

## Project Structure Analysis

### Current Project Layout
```
claude-permissions/
├── main.go              # Main entry point
├── models.go            # Contains Model, Permission, Action, etc.
├── actions.go           # Action handling logic
├── ui.go               # TUI rendering and event handling
├── settings.go         # Settings file loading/saving
├── delegate.go         # List item delegate
├── interfaces.go       # Interface definitions
├── styles.go           # Styling definitions
├── debug/              # Debug server package
│   ├── server.go       # HTTP server setup
│   ├── state.go        # State extraction (uses reflection)
│   ├── layout.go       # Layout diagnostics (uses reflection)
│   ├── input.go        # Input injection (uses reflection)
│   ├── capture.go      # Screen capture
│   └── logging.go      # Event logging
├── layout/             # Layout engine package
└── testdata/           # Test data files
```

### Target Project Layout
```
claude-permissions/
├── main.go              # Main entry point (imports types)
├── actions.go           # Action handling (imports types)
├── ui.go               # TUI rendering (imports types)
├── settings.go         # Settings operations (imports types)
├── delegate.go         # List delegate (imports types)
├── interfaces.go       # Interfaces (imports types)
├── styles.go           # Styling (imports types)
├── types/              # Shared types package
│   └── model.go        # All types: Model, Permission, Action, etc.
├── debug/              # Debug server package (imports types)
│   ├── server.go       # HTTP server (uses *types.Model)
│   ├── state.go        # State extraction (direct field access)
│   ├── layout.go       # Layout diagnostics (direct field access)
│   ├── input.go        # Input injection (direct field access)
│   ├── capture.go      # Screen capture
│   └── logging.go      # Event logging
├── layout/             # Layout engine package
└── testdata/           # Test data files
```

## Detailed Implementation Plan

### Phase 1: Create Shared Types Package

#### 1.1 Create Directory Structure
```bash
mkdir -p types/
```

#### 1.2 Create types/model.go
Move all content from `models.go` to `types/model.go` with these changes:

**Package Declaration:**
```go
package types  // Changed from: package main
```

**Imports:**
```go
import (
    "sync"
    "claude-permissions/layout"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/bubbles/timer"
    "github.com/charmbracelet/bubbles/viewport"
)
```

**Export All Constants:**
```go
const (
    LevelUser  = "User"    // Already exported
    LevelRepo  = "Repo"    // Already exported
    LevelLocal = "Local"   // Already exported
)

const (
    ActionDuplicate = "duplicate"  // Already exported
    ActionMove      = "move"       // Already exported
)
```

**Export All Types:**
```go
// Settings, SettingsLevel, Permission, Duplicate, Action - already exported

// Model struct - EXPORT ALL FIELDS:
type Model struct {
    // Thread safety
    Mutex sync.RWMutex  // Changed from: mutex sync.RWMutex

    // Settings data
    UserLevel  SettingsLevel  // Changed from: userLevel
    RepoLevel  SettingsLevel  // Changed from: repoLevel
    LocalLevel SettingsLevel  // Changed from: localLevel

    // UI state
    Permissions []Permission  // Changed from: permissions
    Duplicates  []Duplicate   // Changed from: duplicates
    Actions     []Action      // Changed from: actions
    ActivePanel int           // Changed from: activePanel

    // Layout engine
    LayoutEngine *layout.LayoutEngine  // Changed from: layoutEngine

    // UI components
    PermissionsList list.Model      // Changed from: permissionsList
    DuplicatesTable table.Model     // Changed from: duplicatesTable
    ActionsView     viewport.Model  // Changed from: actionsView

    // Confirmation state
    ConfirmMode bool    // Changed from: confirmMode
    ConfirmText string  // Changed from: confirmText

    // Status message state
    StatusMessage string      // Changed from: statusMessage
    StatusTimer   timer.Model // Changed from: statusTimer
}
```

**Remove Thread-Safe Getters:**
Delete all the getter methods (GetActivePanel, GetConfirmMode, etc.) since we'll use direct field access with manual mutex locking.

#### 1.3 Delete models.go
After verifying types/model.go is complete, delete the original `models.go` file.

### Phase 2: Update Main Package Files

All files in the main package need to import the types package and update their Model references.

#### 2.1 Update main.go
**Add import:**
```go
import (
    // existing imports...
    "claude-permissions/types"
)
```

**Update function signatures:**
```go
func initialModel() (*types.Model, error) {  // Changed from: (*Model, error)
    // ... existing logic ...

    model := &types.Model{  // Changed from: &Model{
        UserLevel:       userLevel,   // Changed from: userLevel:
        RepoLevel:       repoLevel,   // Changed from: repoLevel:
        LocalLevel:      localLevel,  // Changed from: localLevel:
        Permissions:     permissions, // Changed from: permissions:
        Duplicates:      duplicates,  // Changed from: duplicates:
        Actions:         []types.Action{},  // Changed from: []Action{}
        ActivePanel:     0,           // Changed from: activePanel:
        LayoutEngine:    layout.NewLayoutEngine(),  // Changed from: layoutEngine:
        PermissionsList: permissionsList,  // Changed from: permissionsList:
        DuplicatesTable: duplicatesTable,  // Changed from: duplicatesTable:
        ActionsView:     actionsView,      // Changed from: actionsView:
        ConfirmMode:     false,            // Changed from: confirmMode:
        StatusMessage:   "",               // Changed from: statusMessage:
        StatusTimer:     timer.New(3 * time.Second),  // Changed from: statusTimer:
    }

    return model, nil
}
```

**Update createDuplicatesTable signature:**
```go
func createDuplicatesTable(duplicates []types.Duplicate) table.Model {  // Changed from: []Duplicate
    // ... existing logic unchanged ...
}
```

#### 2.2 Update actions.go
**Add import:**
```go
import (
    // existing imports...
    "claude-permissions/types"
)
```

**Update all function signatures that use Model:**
```go
func (m *types.Model) generateConfirmationText() string {  // Changed from: (m Model)
    m.Mutex.RLock()  // Add explicit locking
    defer m.Mutex.RUnlock()

    // Update field access:
    for _, action := range m.Actions {  // Changed from: m.actions
        // ... existing logic ...
    }
}

func (m *types.Model) executeActions() error {  // Changed from: (m Model)
    m.Mutex.Lock()  // Add explicit locking
    defer m.Mutex.Unlock()

    // Update field access:
    userSettings, repoSettings, localSettings, err := m.loadAllSettings()
    // ... existing logic ...
}

// Similar updates for all other Model methods:
// - loadAllSettings()
// - applyActionsToSettings()
// - removePermissionFromLevel()
// - addPermissionToLevel()
// - saveAllSettings()
```

**Update type references:**
```go
func (m *types.Model) loadAllSettings() (types.Settings, types.Settings, types.Settings, error) {
    // Access exported fields:
    userSettings, err := loadSettingsFromPath(m.UserLevel.Path)   // Changed from: m.userLevel.Path
    repoSettings, err := loadSettingsFromPath(m.RepoLevel.Path)   // Changed from: m.repoLevel.Path
    localSettings, err := loadSettingsFromPath(m.LocalLevel.Path) // Changed from: m.localLevel.Path
    // ... existing logic ...
}
```

#### 2.3 Update ui.go
**Add import:**
```go
import (
    // existing imports...
    "claude-permissions/types"
)
```

**Update all function signatures:**
```go
func (m *types.Model) Init() tea.Cmd {  // Changed from: (m Model)
    // ... existing logic unchanged ...
}

func (m *types.Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {  // Changed from: (m Model)
    // Add mutex locking for state changes:
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        return m.handleWindowResize(msg)
    case timer.TickMsg:
        return m.handleTimerUpdate(msg)
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    }
    return m, nil
}

func (m *types.Model) handleWindowResize(msg tea.WindowSizeMsg) (*types.Model, tea.Cmd) {
    m.Mutex.Lock()  // Add explicit locking
    defer m.Mutex.Unlock()

    // Update field access:
    m.LayoutEngine.HandleResize(msg.Width, msg.Height)  // Changed from: m.layoutEngine
    // ... existing logic ...
}

func (m *types.Model) View() string {  // Changed from: (m Model)
    m.Mutex.RLock()  // Add explicit locking
    defer m.Mutex.RUnlock()

    if m.ConfirmMode {  // Changed from: m.confirmMode
        return m.renderConfirmation()
    }
    // ... existing logic ...
}
```

**Update all other UI methods similarly:**
- handleTimerUpdate, handleTimerTimeout, handleKeyPress
- handleGlobalKeys, updateConfirmMode, updatePermissionsPanel, updateDuplicatesPanel
- handleSubmit, renderConfirmation, renderHeader, renderPermissionsPanel, renderDuplicatesPanel, renderFooter

#### 2.4 Update settings.go
**Add import:**
```go
import (
    // existing imports...
    "claude-permissions/types"
)
```

**Update function signatures that use Settings types:**
```go
func loadUserLevel() (types.SettingsLevel, error) {  // Changed from: (SettingsLevel, error)
    // ... existing logic unchanged ...
}

func loadRepoLevel() (types.SettingsLevel, error) {  // Changed from: (SettingsLevel, error)
    // ... existing logic unchanged ...
}

func loadLocalLevel() (types.SettingsLevel, error) {  // Changed from: (SettingsLevel, error)
    // ... existing logic unchanged ...
}

func consolidatePermissions(userLevel, repoLevel, localLevel types.SettingsLevel) []types.Permission {
    // Changed from: (userLevel, repoLevel, localLevel SettingsLevel) []Permission
    // ... existing logic unchanged ...
}

func detectDuplicates(userLevel, repoLevel, localLevel types.SettingsLevel) []types.Duplicate {
    // Changed from: (userLevel, repoLevel, localLevel SettingsLevel) []Duplicate
    // ... existing logic unchanged ...
}
```

#### 2.5 Update delegate.go, interfaces.go, styles.go
Check each file for Model/Permission/Action references and add types import + update references as needed.

### Phase 3: Rewrite Debug Package

Complete rewrite of debug package to eliminate all reflection and use direct field access.

#### 3.1 Update debug/server.go
**Add import:**
```go
import (
    // existing imports...
    "claude-permissions/types"
)
```

**Update DebugServer struct:**
```go
type DebugServer struct {
    server   *http.Server
    program  *tea.Program
    model    *types.Model        // Changed from: interface{}
    mutex    sync.RWMutex
    logger   *Logger
    shutdown chan struct{}
}
```

**Update constructor and methods:**
```go
func NewDebugServer(port int, program *tea.Program, model *types.Model) *DebugServer {
    // Changed from: model interface{}
    // ... existing logic unchanged ...
}

func (ds *DebugServer) UpdateModel(model *types.Model) {  // Changed from: interface{}
    ds.mutex.Lock()
    defer ds.mutex.Unlock()
    ds.model = model
}

func (ds *DebugServer) GetModel() *types.Model {  // Changed from: interface{}
    ds.mutex.RLock()
    defer ds.mutex.RUnlock()
    return ds.model
}
```

#### 3.2 Rewrite debug/state.go
**Complete rewrite - remove all reflection:**

```go
package debug

import (
    "net/http"
    "claude-permissions/types"
)

// StateResponse represents the complete application state
type StateResponse struct {
    UI        UIState        `json:"ui"`
    Data      DataState      `json:"data"`
    Files     FilesState     `json:"files"`
    Errors    []string       `json:"errors"`
    Timestamp string         `json:"timestamp"`
}

// UIState, DataState, FilesState structs remain the same...

func (ds *DebugServer) handleState(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    model := ds.GetModel()
    if model == nil {
        ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
        return
    }

    response := extractApplicationState(model)
    response.Timestamp = getCurrentTimestamp()

    ds.logger.LogEvent("state_extracted", map[string]interface{}{
        "active_panel":      response.UI.ActivePanel,
        "permissions_count": response.Data.PermissionsCount,
        "duplicates_count":  response.Data.DuplicatesCount,
        "actions_queued":    response.Data.ActionsQueued,
    })

    ds.writeJSONResponse(w, response)
}

func extractApplicationState(model *types.Model) StateResponse {
    model.Mutex.RLock()  // Add explicit locking
    defer model.Mutex.RUnlock()

    return StateResponse{
        UI:     extractUIState(model),
        Data:   extractDataState(model),
        Files:  extractFilesState(model),
        Errors: []string{}, // No more reflection errors
    }
}

func extractUIState(model *types.Model) UIState {
    return UIState{
        ActivePanel:   panelNumberToName(model.ActivePanel),  // Direct field access
        SelectedItems: []string{}, // TODO: Extract from PermissionsList if needed
        ListPosition:  0,          // TODO: Extract from PermissionsList if needed
        FilterText:    "",         // TODO: Extract from PermissionsList if needed
        ConfirmMode:   model.ConfirmMode,    // Direct field access
        ConfirmText:   model.ConfirmText,    // Direct field access
        StatusMessage: model.StatusMessage,  // Direct field access
    }
}

func extractDataState(model *types.Model) DataState {
    return DataState{
        PermissionsCount: len(model.Permissions),  // Direct field access
        DuplicatesCount:  len(model.Duplicates),   // Direct field access
        ActionsQueued:    len(model.Actions),      // Direct field access
        PendingMoves:     extractPendingMoves(model.Actions),
        PendingEdits:     extractPendingEdits(model.Actions),
    }
}

func extractFilesState(model *types.Model) FilesState {
    return FilesState{
        UserExists:  model.UserLevel.Exists,   // Direct field access
        RepoExists:  model.RepoLevel.Exists,   // Direct field access
        LocalExists: model.LocalLevel.Exists,  // Direct field access
        UserPath:    model.UserLevel.Path,     // Direct field access
        RepoPath:    model.RepoLevel.Path,     // Direct field access
        LocalPath:   model.LocalLevel.Path,    // Direct field access
    }
}

func extractPendingMoves(actions []types.Action) []string {
    var moves []string
    for _, action := range actions {
        if action.Type == types.ActionMove {  // Direct field access
            moves = append(moves, action.Permission+"→"+action.ToLevel)
        }
    }
    return moves
}

func extractPendingEdits(actions []types.Action) []string {
    var edits []string
    for _, action := range actions {
        if action.Type == "edit" {  // Direct field access
            edits = append(edits, action.Permission+"→"+action.NewName)
        }
    }
    return edits
}

func panelNumberToName(panel int) string {
    switch panel {
    case 0:
        return "permissions"
    case 1:
        return "duplicates"
    case 2:
        return "actions"
    default:
        return "unknown"
    }
}
```

#### 3.3 Rewrite debug/layout.go
**Remove all reflection, use direct access:**

```go
package debug

import (
    "net/http"
    "claude-permissions/types"
)

// LayoutResponse and other types remain the same...

func (ds *DebugServer) handleLayout(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    model := ds.GetModel()
    if model == nil {
        ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
        return
    }

    response := extractLayoutDiagnostics(model)
    response.Timestamp = getCurrentTimestamp()

    ds.logger.LogEvent("layout_extracted", map[string]interface{}{
        "terminal_width":    response.Terminal[0],
        "terminal_height":   response.Terminal[1],
        "components_count":  len(response.Components),
        "warnings_count":    len(response.Warnings),
    })

    ds.writeJSONResponse(w, response)
}

func extractLayoutDiagnostics(model *types.Model) LayoutResponse {
    model.Mutex.RLock()  // Add explicit locking
    defer model.Mutex.RUnlock()

    response := LayoutResponse{
        Terminal:     [2]int{80, 24}, // Default values
        Components:   make(map[string]ComponentPosition),
        Warnings:     []string{},
        Calculations: LayoutCalculations{
            FrameOverhead:  make(map[string]int),
            ComponentSizes: make(map[string]interface{}),
        },
    }

    if model.LayoutEngine == nil {  // Direct field access
        response.Warnings = append(response.Warnings, "layout_engine_not_found")
        return response
    }

    // Direct method calls instead of reflection
    width, height := model.LayoutEngine.GetTerminalSize()
    response.Terminal = [2]int{width, height}

    if result := model.LayoutEngine.GetLastResult(); result != nil {
        response.Components = extractComponentsFromResult(result)
        response.Warnings = result.Warnings
    }

    response.Calculations = extractLayoutCalculations(model, height)
    return response
}

func extractComponentsFromResult(result *layout.LayoutResult) map[string]ComponentPosition {
    components := make(map[string]ComponentPosition)

    for id, layout := range result.Components {
        components[id] = ComponentPosition{
            X: layout.X,
            Y: layout.Y,
            W: layout.Width,
            H: layout.Height,
        }
    }

    return components
}

func extractLayoutCalculations(model *types.Model, terminalHeight int) LayoutCalculations {
    calc := LayoutCalculations{
        FrameOverhead:  make(map[string]int),
        ComponentSizes: make(map[string]interface{}),
    }

    calc.AvailableHeight, calc.FixedHeight = calculateHeightDistribution(terminalHeight)

    // Extract component sizes using direct method calls
    calc.FrameOverhead = map[string]int{
        "width":  4, // Default frame width overhead
        "height": 2, // Default frame height overhead
    }

    calc.ComponentSizes = map[string]interface{}{
        "permissions_list": map[string]int{
            "width":  model.PermissionsList.Width(),   // Direct method call
            "height": model.PermissionsList.Height(),  // Direct method call
        },
        "duplicates_table": map[string]int{
            "width":  model.DuplicatesTable.Width(),   // Direct method call
            "height": model.DuplicatesTable.Height(),  // Direct method call
        },
        "actions_view": map[string]int{
            "width":  model.ActionsView.Width,   // Direct field access
            "height": model.ActionsView.Height,  // Direct field access
        },
    }

    return calc
}
```

#### 3.4 Rewrite debug/input.go
**Remove all reflection, use direct access:**

```go
package debug

import (
    "encoding/json"
    "net/http"
    "time"
    "claude-permissions/types"
)

// InputRequest, InputResponse types remain the same...

type ModelStateCapture struct {
    ActivePanel     int      `json:"active_panel"`
    SelectedItems   []string `json:"selected_items"`
    FilterText      string   `json:"filter_text"`
    ConfirmMode     bool     `json:"confirm_mode"`
    ActionsCount    int      `json:"actions_count"`
    StatusMessage   string   `json:"status_message"`
}

func (ds *DebugServer) handleInput(w http.ResponseWriter, r *http.Request) {
    // ... request parsing logic unchanged ...

    // Capture state before input
    beforeState := ds.captureModelState()

    // Send the input to the application
    err := ds.SendInput(request.Key)

    // Give the application a moment to process the input
    time.Sleep(50 * time.Millisecond)

    // Capture state after input
    afterState := ds.captureModelState()

    // Build response - logic unchanged
    // ...
}

func (ds *DebugServer) captureModelState() ModelStateCapture {
    model := ds.GetModel()
    if model == nil {
        return ModelStateCapture{}
    }

    model.Mutex.RLock()  // Add explicit locking
    defer model.Mutex.RUnlock()

    return ModelStateCapture{
        ActivePanel:   model.ActivePanel,          // Direct field access
        SelectedItems: []string{},                 // TODO: Extract from lists if needed
        FilterText:    "",                         // TODO: Extract from lists if needed
        ConfirmMode:   model.ConfirmMode,          // Direct field access
        ActionsCount:  len(model.Actions),         // Direct field access
        StatusMessage: model.StatusMessage,        // Direct field access
    }
}

// analyzeStateChanges function remains the same...
```

#### 3.5 Update debug/capture.go (minimal changes)
**Add types import and update interface:**

```go
package debug

import (
    "net/http"
    "regexp"
    "strings"
    "claude-permissions/types"
    "golang.org/x/term"
)

// Update interface to use concrete type
type ModelViewProvider interface {
    View() string
}

func (ds *DebugServer) handleSnapshot(w http.ResponseWriter, r *http.Request) {
    // ... existing logic ...

    model := ds.GetModel()
    if model == nil {
        ds.writeErrorResponse(w, "Model not available", http.StatusInternalServerError)
        return
    }

    // Get the view content from the model
    var content string
    if viewProvider, ok := model.(ModelViewProvider); ok {
        content = viewProvider.View()  // Direct method call
    } else {
        ds.writeErrorResponse(w, "Model does not support view capture", http.StatusInternalServerError)
        return
    }

    // ... rest of logic unchanged ...
}
```

### Phase 4: Testing and Validation

#### 4.1 Build Verification
```bash
go mod tidy
go build -o claude-permissions .
```

#### 4.2 Import Verification
Check for circular imports:
```bash
go list -deps . | grep claude-permissions
```

#### 4.3 Functionality Testing
- Start debug server: `./claude-permissions --debug-server`
- Test all endpoints:
  - `curl http://localhost:8080/health`
  - `curl http://localhost:8080/state`
  - `curl http://localhost:8080/layout`
  - `curl http://localhost:8080/snapshot`
  - `curl -X POST http://localhost:8080/input -d '{"key":"tab"}'`
  - `curl http://localhost:8080/logs`

#### 4.4 Pre-commit Validation
```bash
pre-commit run --files types/model.go main.go actions.go ui.go settings.go debug/state.go debug/layout.go debug/input.go debug/server.go
```

## Expected Benefits

### Performance Improvements
- **Eliminate reflection overhead**: Direct field access is orders of magnitude faster
- **Reduced memory allocations**: No reflection-based temporary objects
- **Faster debug server responses**: State extraction becomes much faster

### Code Quality Improvements
- **Compile-time safety**: Type errors caught at build time, not runtime
- **Reduced complexity**: Eliminate high cyclomatic complexity from reflection
- **Better IDE support**: Autocomplete, go-to-definition, refactoring all work
- **Easier debugging**: Simple field access instead of complex reflection stack traces

### Maintainability Improvements
- **Cleaner code**: Direct field access is much easier to read and understand
- **Easier modifications**: Adding/removing fields doesn't require reflection updates
- **Type safety**: Changes to Model struct cause compile errors instead of runtime failures

## Risk Mitigation

### Thread Safety
- **Challenge**: Manual mutex management instead of getter methods
- **Mitigation**: Clear locking patterns, consistent mutex usage
- **Testing**: Verify no race conditions with concurrent debug server access

### Large Refactor Risk
- **Challenge**: Touching many files increases chance of errors
- **Mitigation**: Incremental testing, comprehensive validation, detailed plan
- **Rollback**: Keep original code until fully verified

### Import Cycles
- **Challenge**: Shared package could create circular dependencies
- **Mitigation**: Careful package design, types package has minimal dependencies
- **Testing**: Explicit import cycle checking

## Success Criteria

1. **All files compile successfully** - No build errors
2. **Debug server endpoints work** - All API endpoints return correct data
3. **TUI application works** - Main application functionality unchanged
4. **Performance improvement** - Debug endpoints respond faster
5. **No reflection code** - All reflection imports and usage removed
6. **Pre-commit passes** - No linting errors or complexity warnings
7. **Thread safety maintained** - No race conditions in concurrent access

## Implementation Notes

### Manual Mutex Management
With direct field access, the debug server needs to manually manage the mutex:

```go
func extractUIState(model *types.Model) UIState {
    model.Mutex.RLock()  // Lock before reading
    defer model.Mutex.RUnlock()

    // Safe to access fields while locked
    activePanel := model.ActivePanel
    confirmMode := model.ConfirmMode
    // ... other field access ...

    return UIState{
        ActivePanel: panelNumberToName(activePanel),
        ConfirmMode: confirmMode,
        // ...
    }
    // Mutex automatically unlocked by defer
}
```

### Field Naming Convention
All Model fields must be exported (capitalized) for debug package access:
- `activePanel` → `ActivePanel`
- `confirmMode` → `ConfirmMode`
- `statusMessage` → `StatusMessage`
- etc.

### Error Handling Simplification
With direct field access, many error conditions disappear:
- No more "field not found" errors
- No more type assertion failures
- No more reflection panic conditions

The debug server becomes much more robust and predictable.

# Debug Package Architecture

## Overview

The debug package implements a self-registering endpoint system where each HTTP endpoint is
completely isolated in its own file. This architecture ensures zero-modification
addition/removal of debug endpoints.

## Core Principles

### 1. One Endpoint, One File

- Each debug endpoint lives in exactly one `.go` file
- File name matches endpoint path exactly (e.g., `/health` → `health.go`,
  `/launch-confirm-changes` → `launch-confirm-changes.go`)
- Complete isolation: endpoint handler, types, helper functions, and registration

### 2. Self-Registration Pattern

Every endpoint file follows this pattern:

```go
package debug

func init() {
    RegisterEndpoint("/endpoint-name", handleEndpointName)
}

func handleEndpointName(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}

// Endpoint-specific types and helpers go here
```

### 3. Zero-Modification Guarantee

- **Adding endpoint**: Create single new `.go` file, no other files modified
- **Removing endpoint**: Delete single `.go` file, no other files modified
- **No dangling code**: Deleting an endpoint file removes all related functionality

## File Structure

```txt
debug/
├── server.go                    # Core server + registration system
├── utils.go                     # Shared utilities only
├── [endpoint-name].go           # Individual endpoint files
└── CLAUDE.md                    # This architecture guide
```

## Implementation Guidelines

### Endpoint File Template

```go
package debug

import (
    // Only imports needed for this specific endpoint
)

func init() {
    RegisterEndpoint("/your-endpoint", handleYourEndpoint)
}

// Endpoint-specific types
type YourEndpointRequest struct {
    // Request fields
}

type YourEndpointResponse struct {
    // Response fields
}

// Main handler (keep under 60 lines for linting)
func handleYourEndpoint(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
    // Implementation
}

// Endpoint-specific helper functions
func yourEndpointHelper() {
    // Helper implementation
}
```

### Shared vs. Endpoint-Specific Code

**Put in `utils.go` (shared):**

- JSON response writing (`writeJSONResponse`, `writeErrorResponse`)
- Query parameter parsing (`getQueryParamBool`)
- Timestamp utilities (`getCurrentTimestamp`, `getTimestamp`)
- Type conversion utilities (`panelNumberToName`, `screenNumberToName`)

**Put in endpoint file (specific):**

- Endpoint handler function
- Request/response types for that endpoint
- Helper functions used only by that endpoint
- Endpoint-specific constants or variables

### Code Quality Requirements

- **Function length**: Keep handlers under 60 lines (extract helper functions if needed)
- **Line length**: Maximum 120 characters (use golines formatting)
- **Imports**: Only import what you actually use
- **Error handling**: Use `writeErrorResponse` for consistent error responses
- **Logging**: Use `ds.logger.LogEvent()` for endpoint activity

## Current Endpoints

| Endpoint | File | Purpose |
|----------|------|---------|
| `/health` | `health.go` | Health check |
| `/state` | `state.go` | Application state inspection |
| `/snapshot` | `snapshot.go` | Screen capture + layout diagnostics |
| `/input` | `input.go` | Input injection + state change analysis |
| `/logs` | `logs.go` | Debug event logs + slog handler |
| `/reset` | `reset.go` | Application state reset |
| `/launch-confirm-changes` | `launch-confirm-changes.go` | Screen testing with mock data |

## Development Workflow

### Adding New Endpoint

1. Create `debug/[endpoint-name].go`
2. Follow the endpoint file template
3. Implement handler and types
4. Test with `go build .`
5. Run `pre-commit run --files debug/[endpoint-name].go`

### Removing Endpoint

1. Delete `debug/[endpoint-name].go`
2. Verify build: `go build .`
3. Confirm no dangling references

### Modifying Endpoint

1. Edit only the specific endpoint file
2. Move shared code to `utils.go` if it becomes reusable
3. Extract helper functions if handler gets too long

## Architecture Benefits

- **Scalability**: Easy to add new debug endpoints without touching existing code
- **Maintainability**: Clear separation of concerns, easy to find endpoint-specific code
- **Testability**: Each endpoint can be tested in isolation
- **Code review**: Changes are localized to single files
- **Debugging**: Easy to temporarily disable endpoints by renaming files

## Integration Points

- **Server registration**: Automatic via `init()` functions
- **Logger access**: Available via `ds.logger` parameter
- **Model access**: Use `ds.GetModel()` with proper mutex locking
- **Program interaction**: Use `ds.program.Send()` for TUI messages

## Common Patterns

### Request validation

```go
if r.Method != http.MethodPost {
    writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
    return
}
```

### JSON request parsing

```go
var request YourRequest
if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    writeErrorResponse(w, "Invalid JSON in request body", http.StatusBadRequest, ds.logger)
    return
}
```

### Safe model access

```go
model := ds.GetModel()
if model == nil {
    writeErrorResponse(w, "Model not available", http.StatusInternalServerError, ds.logger)
    return
}

model.Mutex.RLock()
// Access model fields
model.Mutex.RUnlock()
```

### Response logging and writing

```go
ds.logger.LogEvent("endpoint_accessed", map[string]interface{}{
    "key": "value",
})

writeJSONResponse(w, response, ds.logger)
```

This architecture ensures maintainable, scalable debug endpoints that follow the
zero-modification principle.

# Debug Package Architecture

## Overview

Self-registering endpoint system: one endpoint per file, zero-modification addition/removal.

## Essential Debug Commands

```bash
# Debug API usage (ALWAYS assume server is running)
scripts/debug-api.sh state          # Get application state
scripts/debug-api.sh snapshot       # Screen capture (no ANSI)
scripts/debug-api.sh snapshot --color  # Screen capture with ANSI
scripts/debug-api.sh logs           # Get debug events (clears buffer)
scripts/debug-api.sh reset          # Reset application state

# Input simulation
scripts/debug-api.sh input tab      # Send TAB key
scripts/debug-api.sh input enter    # Send ENTER key
scripts/debug-api.sh input up       # Navigation keys
scripts/debug-api.sh input a        # Letter keys

# Settings loading
scripts/debug-api.sh load-settings --user-file testdata/user.json --repo-file testdata/repo.json

# IMPORTANT: Supported keys - tab, enter, escape/esc, up, down, left, right, space, a, u, r, l, e, c, q, /
```

## Core Principles

### IMPORTANT: One Endpoint, One File

- **YOU MUST** follow file name = endpoint path (e.g., `/health` → `health.go`)
- **REQUIRED**: Complete isolation - handler, types, helpers, registration in single file

### CRITICAL: Self-Registration Pattern

**YOU MUST use this exact pattern for all endpoints:**

```go
package debug

func init() {
    RegisterEndpoint("/endpoint-name", handleEndpointName)
}

func handleEndpointName(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### Zero-Modification Guarantee

- **Add endpoint**: Create one file, modify nothing else
- **Remove endpoint**: Delete one file, modify nothing else

## Implementation

### REQUIRED File Template

**YOU MUST use this exact structure for all endpoint files:**

```go
package debug

func init() {
    RegisterEndpoint("/your-endpoint", handleYourEndpoint)
}

type YourEndpointRequest struct {
    // Request fields
}

type YourEndpointResponse struct {
    // Response fields
}

func handleYourEndpoint(ds *DebugServer, w http.ResponseWriter, r *http.Request) {
    // IMPORTANT: Keep under 60 lines, extract helpers if needed
}
```

### Code Organization Rules

- **utils.go (shared)**: JSON responses, query parsing, timestamps, type conversions
- **endpoint files**: Handler, types, helpers specific to that endpoint

### CRITICAL Quality Requirements

- **YOU MUST** keep handlers under 60 lines
- **NEVER** exceed 120 character lines
- **ALWAYS** use `writeErrorResponse`, `ds.logger.LogEvent()`

## Current Endpoints

- `/health` → `endpoint-health.go` - Health check
- `/state` → `endpoint-state.go` - Application state
- `/snapshot` → `endpoint-snapshot.go` - Screen capture
- `/input` → `endpoint-input.go` - Input injection
- `/logs` → `endpoint-logs.go` - Debug events
- `/reset` → `endpoint-reset.go` - State reset
- `/launch-confirm-changes` → `endpoint-launch-confirm-changes.go` - Screen testing
- `/load-settings` → `endpoint-load-settings.go` - Dynamic settings loading

## CRITICAL Common Patterns

**ALWAYS use these patterns in every endpoint:**

**Request validation (REQUIRED):**

```go
if r.Method != http.MethodPost {
    writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
    return
}
```

**Safe model access (YOU MUST use mutex):**

```go
model := ds.GetModel()
if model == nil {
    writeErrorResponse(w, "Model not available", http.StatusInternalServerError, ds.logger)
    return
}
model.Mutex.RLock()
// Access fields
model.Mutex.RUnlock()
```

**Response (ALWAYS log events):**

```go
ds.logger.LogEvent("endpoint_accessed", map[string]interface{}{"key": "value"})
writeJSONResponse(w, response, ds.logger)
```

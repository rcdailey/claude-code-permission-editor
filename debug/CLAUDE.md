# Debug Package Architecture

## Overview

Self-registering endpoint system: one endpoint per file, zero-modification addition/removal.

## Core Principles

### One Endpoint, One File

- File name = endpoint path (e.g., `/health` → `health.go`)
- Complete isolation: handler, types, helpers, registration in single file

### Self-Registration Pattern

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

- Add endpoint: Create one file, modify nothing else
- Remove endpoint: Delete one file, modify nothing else

## Implementation

### File Template

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
    // Keep under 60 lines, extract helpers if needed
}
```

### Code Organization

**utils.go (shared):** JSON responses, query parsing, timestamps, type conversions
**endpoint files:** Handler, types, helpers specific to that endpoint

### Quality Requirements

- Handlers under 60 lines
- Max 120 character lines
- Use `writeErrorResponse`, `ds.logger.LogEvent()`

## Current Endpoints

- `/health` → `health.go` - Health check
- `/state` → `state.go` - Application state
- `/snapshot` → `snapshot.go` - Screen capture
- `/input` → `input.go` - Input injection
- `/logs` → `logs.go` - Debug events
- `/reset` → `reset.go` - State reset
- `/launch-confirm-changes` → `launch-confirm-changes.go` - Screen testing

## Common Patterns

**Request validation:**

```go
if r.Method != http.MethodPost {
    writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed, ds.logger)
    return
}
```

**Safe model access:**

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

**Response:**

```go
ds.logger.LogEvent("endpoint_accessed", map[string]interface{}{"key": "value"})
writeJSONResponse(w, response, ds.logger)
```

# Debug API Architecture Plan

## Overview

This document outlines the comprehensive plan for implementing a real-time debug API that allows Claude to inspect and interact with the running TUI application. This replaces the current non-functional debug mode with a proper feedback loop mechanism.

## Problem Statement

The current debug mode shows a static snapshot at 80x24 that doesn't reflect the actual interactive TUI experience. This breaks the feedback loop needed for effective debugging, particularly for viewport utilization and layout issues.

## Solution Architecture

### Core Concept
- User runs the TUI application in their terminal and keeps it running
- Claude can query the running application for real-time information
- Claude can send input to the application and observe state changes
- All communication happens through HTTP API endpoints optimized for AI consumption

### Technology Stack
- **Transport**: HTTP server embedded in the TUI application
- **Format**: JSON responses optimized for AI consumption (clear keys, no redundancy)
- **Future Compatibility**: Designed to work with MCP via separate translation layer

## File Structure

```
debug/
├── server.go       // HTTP server setup and routing
├── capture.go      // Screen snapshot logic + SnapshotResponse type
├── state.go        // Application state extraction + StateResponse type
├── layout.go       // Layout engine diagnostics + LayoutResponse type
├── input.go        // Input injection and key handling + InputResponse type
└── logging.go      // Event logging system + LogResponse type
```

**Design Principle**: Each file owns its functionality and response types for clean feature isolation.

## API Endpoints

### 1. Screen Snapshot
**`GET /snapshot`**
- Returns exact terminal buffer content as seen by user
- Includes ANSI color information by default
- Query parameter `?raw=true` for plain text without colors
- Response includes terminal dimensions

```json
{
  "content": "\033[1;32mPermissions\033[0m (11 total)\n...",
  "width": 120,
  "height": 40,
  "cursor_position": [15, 8]
}
```

**Use Cases**:
- Color needed: Debugging visual styling, focus indicators, status colors
- Raw needed: Text parsing, layout analysis, automated testing

### 2. Application State
**`GET /state`**
- Comprehensive dump of current application behavior state
- Non-UI, non-layout information focused on data and interactions
- Designed as info dump for maximum debugging utility

```json
{
  "ui": {
    "active_panel": "permissions",
    "selected_items": ["Bash", "Edit"],
    "list_position": 2,
    "filter_text": "",
    "confirm_mode": false
  },
  "data": {
    "permissions_count": 11,
    "duplicates_count": 3,
    "actions_queued": 2,
    "pending_moves": ["Edit→User", "Read→Repo"]
  },
  "files": {
    "user_exists": true,
    "repo_exists": true,
    "local_exists": false
  },
  "errors": ["validation_failed"]
}
```

### 3. Layout Diagnostics
**`GET /layout`**
- Layout engine specific information
- Component dimensions, positioning, calculations
- Warnings and overflow detection

```json
{
  "terminal": [120, 40],
  "components": {
    "permissions": {"x": 0, "y": 5, "w": 120, "h": 15},
    "duplicates": {"x": 0, "y": 21, "w": 120, "h": 8},
    "header": {"x": 0, "y": 0, "w": 120, "h": 4},
    "footer": {"x": 0, "y": 36, "w": 120, "h": 2}
  },
  "warnings": ["component_overflow"],
  "calculations": {
    "available_height": 32,
    "fixed_height": 8,
    "frame_overhead": {"width": 4, "height": 2}
  }
}
```

### 4. Input Injection
**`POST /input`**
- Send keystrokes to the application
- Returns structured information about state changes
- Supports all application hotkeys

```json
// Request
{"key": "tab"}

// Response
{
  "previous_panel": "permissions",
  "new_panel": "duplicates",
  "state_changes": ["panel_switched"],
  "success": true
}
```

**Supported Keys**: ↑↓←→, Tab, Space, Enter, Esc, A, U/R/L, E, C, Q, /

### 5. Event Logging
**`GET /logs`**
- Real-time feed of application events
- Incremental reads with `?since=<id>` parameter
- In-memory circular buffer (1000 entries max)

```json
{
  "entries": [
    {
      "id": 124,
      "timestamp": "2025-01-19T16:30:15Z",
      "level": "info",
      "event": "panel_switch",
      "data": {"from": "permissions", "to": "duplicates"}
    },
    {
      "id": 125,
      "timestamp": "2025-01-19T16:30:16Z",
      "level": "info",
      "event": "item_selected",
      "data": {"item": "Bash", "panel": "permissions"}
    }
  ],
  "next_id": 126
}
```

**Logged Events**:
- Panel switches
- Item selections/deselections
- Layout recalculations
- Input events
- Error conditions
- Action queue changes

### 6. State Reset
**`POST /reset`**
- Reset application to initial state
- Clear action queues, selections, filters
- Useful for test setup

## Implementation Details

### Server Integration
- Add `--debug-server` flag to enable HTTP server
- Server runs in goroutine alongside Bubble Tea application
- Default port: 8080 (configurable)
- Only enabled in debug builds/mode

### Thread Safety
- Mutex protection around shared state access
- Safe concurrent access between TUI and HTTP handlers
- Atomic operations for simple state reads

### Response Optimization
- Clear, semantic key names for AI readability
- No redundant data in responses
- Compact but not cryptic JSON structure
- Consistent error response format

### Error Handling
- Graceful degradation if debug server fails to start
- Structured error responses with helpful context
- Logging of debug server issues

## Future Considerations

### MCP Integration
The HTTP API is designed to be MCP-compatible through a translation layer:

```
Claude ↔ MCP Server ↔ TUI Debug API
         (stdio)    (http)
```

- HTTP endpoints map directly to MCP tool definitions
- Same response structures work for both transports
- MCP server would be thin translation layer

### Extensibility
- Easy to add new endpoints as debugging needs arise
- Modular design allows feature-specific enhancements
- Response schemas can evolve independently

### Testing Strategy
- Each debug module testable independently
- Mock HTTP responses for automated testing
- Integration tests using actual TUI instance

## Success Criteria

1. **Real Feedback Loop**: Claude can see exactly what user sees
2. **Interactive Control**: Claude can drive the application and observe responses
3. **Comprehensive Visibility**: All relevant state accessible for debugging
4. **Performance**: Debug API doesn't interfere with TUI responsiveness
5. **Maintainability**: Clean separation from core TUI functionality

## Implementation Priority

1. **Phase 1**: Basic HTTP server + snapshot endpoint
2. **Phase 2**: State and layout endpoints
3. **Phase 3**: Input injection with response feedback
4. **Phase 4**: Event logging system
5. **Phase 5**: Polish and optimization

This architecture transforms TUI debugging from "guess and check" to "see and control", providing Claude with the real-time visibility needed for effective issue resolution.

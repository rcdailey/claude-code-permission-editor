package debug

import (
	"net/http"
	"sync"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        int64                  `json:"id"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// LogResponse represents the logs endpoint response
type LogResponse struct {
	Entries []LogEntry `json:"entries"`
}

// Logger manages event logging for the debug system
type Logger struct {
	mutex      sync.RWMutex
	entries    []LogEntry
	nextID     int64
	maxEntries int
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		entries:    make([]LogEntry, 0),
		nextID:     1,
		maxEntries: 1000, // Circular buffer of 1000 entries
	}
}

// LogEvent logs an event with optional data
func (l *Logger) LogEvent(event string, data map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	entry := LogEntry{
		ID:        l.nextID,
		Timestamp: getTimestamp(),
		Level:     "info",
		Event:     event,
		Data:      data,
	}

	l.addEntry(entry)
}

// LogError logs an error event
func (l *Logger) LogError(event string, err error, data map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if data == nil {
		data = make(map[string]interface{})
	}
	data["error"] = err.Error()

	entry := LogEntry{
		ID:        l.nextID,
		Timestamp: getTimestamp(),
		Level:     "error",
		Event:     event,
		Data:      data,
	}

	l.addEntry(entry)
}

// LogWarning logs a warning event
func (l *Logger) LogWarning(event string, message string, data map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if data == nil {
		data = make(map[string]interface{})
	}
	data["message"] = message

	entry := LogEntry{
		ID:        l.nextID,
		Timestamp: getTimestamp(),
		Level:     "warning",
		Event:     event,
		Data:      data,
	}

	l.addEntry(entry)
}

// LogDebug logs a debug event
func (l *Logger) LogDebug(event string, data map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	entry := LogEntry{
		ID:        l.nextID,
		Timestamp: getTimestamp(),
		Level:     "debug",
		Event:     event,
		Data:      data,
	}

	l.addEntry(entry)
}

// addEntry adds an entry to the circular buffer
func (l *Logger) addEntry(entry LogEntry) {
	l.entries = append(l.entries, entry)
	l.nextID++

	// Maintain circular buffer size
	if len(l.entries) > l.maxEntries {
		// Remove the oldest entry
		l.entries = l.entries[1:]
	}
}

// GetAndClearEntries returns all current entries and clears the buffer
func (l *Logger) GetAndClearEntries() []LogEntry {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Return all entries
	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)

	// Clear the buffer
	l.entries = make([]LogEntry, 0)

	return result
}

// GetAllEntries returns all current entries
func (l *Logger) GetAllEntries() []LogEntry {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// GetNextID returns the next ID that will be assigned
func (l *Logger) GetNextID() int64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.nextID
}

// Clear clears all log entries
func (l *Logger) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.entries = make([]LogEntry, 0)
	l.nextID = 1
}

// SetMaxEntries sets the maximum number of entries to keep
func (l *Logger) SetMaxEntries(max int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.maxEntries = max

	// Trim existing entries if necessary
	if len(l.entries) > max {
		startIndex := len(l.entries) - max
		l.entries = l.entries[startIndex:]
	}
}

// GetStats returns logging statistics
func (l *Logger) GetStats() map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_entries": len(l.entries),
		"max_entries":   l.maxEntries,
		"next_id":       l.nextID,
	}

	// Count entries by level
	levelCounts := make(map[string]int)
	for _, entry := range l.entries {
		levelCounts[entry.Level]++
	}
	stats["entries_by_level"] = levelCounts

	// Count entries by event type
	eventCounts := make(map[string]int)
	for _, entry := range l.entries {
		eventCounts[entry.Event]++
	}
	stats["entries_by_event"] = eventCounts

	return stats
}

// handleLogs handles the GET /logs endpoint
func (ds *DebugServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ds.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all entries and clear the buffer
	entries := ds.logger.GetAndClearEntries()

	response := LogResponse{
		Entries: entries,
	}

	ds.writeJSONResponse(w, response)
}

// getTimestamp returns the current timestamp in RFC3339 format
func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Common event types that can be logged
const (
	EventServerStart        = "server_start"
	EventServerStop         = "server_stop"
	EventSnapshotCaptured   = "snapshot_captured"
	EventStateExtracted     = "state_extracted"
	EventLayoutExtracted    = "layout_extracted"
	EventInputSent          = "input_sent"
	EventInputProcessed     = "input_processed"
	EventResetRequested     = "reset_requested"
	EventPanelSwitch        = "panel_switch"
	EventItemSelected       = "item_selected"
	EventItemDeselected     = "item_deselected"
	EventLayoutRecalculated = "layout_recalculated"
	EventErrorOccurred      = "error_occurred"
	EventFilterActivated    = "filter_activated"
	EventFilterCleared      = "filter_cleared"
	EventConfirmModeEntered = "confirm_mode_entered"
	EventConfirmModeExited  = "confirm_mode_exited"
	EventStatusMessageSet   = "status_message_set"
	EventJSONEncodeError    = "json_encode_error"
	EventErrorResponse      = "error_response"
)

// LogPanelSwitch logs a panel switch event
func (l *Logger) LogPanelSwitch(from, to string) {
	l.LogEvent(EventPanelSwitch, map[string]interface{}{
		"from": from,
		"to":   to,
	})
}

// LogItemSelected logs an item selection event
func (l *Logger) LogItemSelected(item, panel string) {
	l.LogEvent(EventItemSelected, map[string]interface{}{
		"item":  item,
		"panel": panel,
	})
}

// LogItemDeselected logs an item deselection event
func (l *Logger) LogItemDeselected(item, panel string) {
	l.LogEvent(EventItemDeselected, map[string]interface{}{
		"item":  item,
		"panel": panel,
	})
}

// LogLayoutRecalculated logs a layout recalculation event
func (l *Logger) LogLayoutRecalculated(width, height int, componentCount int) {
	l.LogEvent(EventLayoutRecalculated, map[string]interface{}{
		"width":           width,
		"height":          height,
		"component_count": componentCount,
	})
}

// LogFilterChange logs filter activation or clearing
func (l *Logger) LogFilterChange(activated bool, filterText string) {
	if activated {
		l.LogEvent(EventFilterActivated, map[string]interface{}{
			"filter_text": filterText,
		})
	} else {
		l.LogEvent(EventFilterCleared, nil)
	}
}

// LogConfirmModeChange logs confirm mode changes
func (l *Logger) LogConfirmModeChange(entered bool, confirmText string) {
	if entered {
		l.LogEvent(EventConfirmModeEntered, map[string]interface{}{
			"confirm_text": confirmText,
		})
	} else {
		l.LogEvent(EventConfirmModeExited, nil)
	}
}

// LogStatusMessage logs status message changes
func (l *Logger) LogStatusMessage(message string) {
	l.LogEvent(EventStatusMessageSet, map[string]interface{}{
		"message": message,
	})
}

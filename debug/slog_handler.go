package debug

import (
	"context"
	"log/slog"
)

// DebugSlogHandler implements slog.Handler to route logs to the debug server
type DebugSlogHandler struct {
	logger *Logger
}

// NewDebugSlogHandler creates a new slog handler that routes to the debug server
func NewDebugSlogHandler(debugLogger *Logger) *DebugSlogHandler {
	return &DebugSlogHandler{
		logger: debugLogger,
	}
}

// Enabled returns true if the given level should be logged
func (h *DebugSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Always enabled - let the debug logger decide what to capture
	return true
}

// Handle processes a log record and routes it to the debug server
func (h *DebugSlogHandler) Handle(_ context.Context, r slog.Record) error {
	// Convert slog attributes to map for debug server
	data := make(map[string]interface{})
	r.Attrs(func(attr slog.Attr) bool {
		data[attr.Key] = attr.Value.Any()
		return true
	})

	// Route to appropriate debug logger method based on slog level
	switch r.Level {
	case slog.LevelDebug:
		h.logger.LogDebug(r.Message, data)
	case slog.LevelInfo:
		h.logger.LogEvent(r.Message, data)
	case slog.LevelWarn:
		h.logger.LogWarning(r.Message, r.Message, data)
	case slog.LevelError:
		// For errors, try to extract error from data
		if errVal, exists := data["error"]; exists {
			if err, ok := errVal.(error); ok {
				h.logger.LogError(r.Message, err, data)
				return nil
			}
		}
		// If no error in data, log as regular event
		h.logger.LogEvent(r.Message, data)
	default:
		h.logger.LogEvent(r.Message, data)
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes added
func (h *DebugSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler
	// In a more sophisticated implementation, we'd maintain attribute context
	return h
}

// WithGroup returns a new handler with the given group name
func (h *DebugSlogHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler
	// In a more sophisticated implementation, we'd maintain group context
	return h
}

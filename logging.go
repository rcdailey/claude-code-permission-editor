package main

import (
	"context"
	"log/slog"
)

// NoOpHandler implements slog.Handler with no-op methods for zero overhead
// when debug server is disabled
type NoOpHandler struct{}

// Enabled always returns false for NoOpHandler since we don't want to process anything
func (NoOpHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

// Handle is a no-op for NoOpHandler
func (NoOpHandler) Handle(context.Context, slog.Record) error {
	return nil
}

// WithAttrs returns the same NoOpHandler since attributes don't matter
func (n NoOpHandler) WithAttrs([]slog.Attr) slog.Handler {
	return n
}

// WithGroup returns the same NoOpHandler since groups don't matter
func (n NoOpHandler) WithGroup(string) slog.Handler {
	return n
}

package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"claude-permissions/types"

	tea "github.com/charmbracelet/bubbletea/v2"
)

// ViewProvider interface for getting the rendered view
type ViewProvider interface {
	GetView() string
}

// DebugServer represents the HTTP debug server
type DebugServer struct {
	server       *http.Server
	program      *tea.Program
	model        *types.Model
	viewProvider ViewProvider
	mutex        sync.RWMutex
	logger       *Logger
	shutdown     chan struct{}
}

// NewDebugServer creates a new debug server instance
func NewDebugServer(
	port int,
	program *tea.Program,
	model *types.Model,
	viewProvider ViewProvider,
) *DebugServer {
	logger := NewLogger()

	ds := &DebugServer{
		program:      program,
		model:        model,
		viewProvider: viewProvider,
		logger:       logger,
		shutdown:     make(chan struct{}),
	}

	mux := http.NewServeMux()

	// Register all endpoints
	mux.HandleFunc("/snapshot", ds.handleSnapshot)
	mux.HandleFunc("/state", ds.handleState)
	// Layout diagnostics are included in the snapshot endpoint
	mux.HandleFunc("/input", ds.handleInput)
	mux.HandleFunc("/logs", ds.handleLogs)
	mux.HandleFunc("/reset", ds.handleReset)

	// Health check endpoint
	mux.HandleFunc("/health", ds.handleHealth)

	ds.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return ds
}

// Start starts the debug server in a goroutine
func (ds *DebugServer) Start() error {
	go func() {
		ds.logger.LogEvent("server_start", map[string]interface{}{
			"port": ds.server.Addr,
		})

		if err := ds.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Debug server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Stop gracefully stops the debug server
func (ds *DebugServer) Stop() error {
	close(ds.shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ds.logger.LogEvent("server_stop", nil)
	return ds.server.Shutdown(ctx)
}

// UpdateModel safely updates the model reference
func (ds *DebugServer) UpdateModel(model *types.Model) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()
	ds.model = model
}

// GetModel safely retrieves the current model
func (ds *DebugServer) GetModel() *types.Model {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.model
}

// Logger returns the debug server's logger instance
func (ds *DebugServer) Logger() *Logger {
	return ds.logger
}

// SendInput sends a key input to the TUI program
func (ds *DebugServer) SendInput(key string) error {
	if ds.program == nil {
		return fmt.Errorf("no program instance available")
	}

	msg, err := convertKeyToMessage(key)
	if err != nil {
		return err
	}

	ds.program.Send(msg)
	ds.logger.LogEvent("input_sent", map[string]interface{}{
		"key": key,
	})

	return nil
}

// convertKeyToMessage converts a string key to a tea.Msg
func convertKeyToMessage(key string) (tea.Msg, error) {
	switch key {
	case "up", "arrow-up":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}), nil
	case "down", "arrow-down":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}), nil
	case "left", "arrow-left":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}), nil
	case "right", "arrow-right":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}), nil
	case "tab":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}), nil
	case "enter":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}), nil
	case "escape", "esc":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}), nil
	case "space":
		return tea.KeyPressMsg(tea.Key{Code: tea.KeySpace, Text: " "}), nil
	default:
		return convertRuneKeyToMessage(key)
	}
}

// keyMappings maps key strings to their corresponding rune
var keyMappings = map[string]rune{
	"a": 'a', "A": 'a',
	"u": 'u', "U": 'u',
	"r": 'r', "R": 'r',
	"l": 'l', "L": 'l',
	"e": 'e', "E": 'e',
	"c": 'c', "C": 'c',
	"q": 'q', "Q": 'q',
	"y": 'y', "Y": 'y',
	"n": 'n', "N": 'n',
	"/": '/',
	"1": '1',
	"2": '2',
	"3": '3',
}

// convertRuneKeyToMessage converts single character keys to messages
func convertRuneKeyToMessage(key string) (tea.Msg, error) {
	if r, ok := keyMappings[key]; ok {
		return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)}), nil
	}
	return nil, fmt.Errorf("unsupported key: %s", key)
}

// handleHealth provides a health check endpoint
func (ds *DebugServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		ds.logger.LogError("health_endpoint_error", err, nil)
	}
}

// writeJSONResponse writes a JSON response with proper headers
func (ds *DebugServer) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		ds.logger.LogEvent("json_encode_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// writeErrorResponse writes a structured error response
func (ds *DebugServer) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ds.logger.LogError("error_response_encode_failed", err, nil)
	}
	ds.logger.LogEvent("error_response", map[string]interface{}{
		"message":     message,
		"status_code": statusCode,
	})
}

// getQueryParamBool safely gets a boolean query parameter with a default value
func getQueryParamBool(r *http.Request, key string, defaultValue bool) bool {
	if value := r.URL.Query().Get(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

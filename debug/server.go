package debug

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

// EndpointHandler represents a handler function for debug endpoints
type EndpointHandler func(*DebugServer, http.ResponseWriter, *http.Request)

// Endpoint registry for self-registering endpoints
var (
	endpointRegistry = make(map[string]EndpointHandler)
	registryMutex    sync.RWMutex
)

// RegisterEndpoint allows endpoints to register themselves
func RegisterEndpoint(path string, handler EndpointHandler) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	endpointRegistry[path] = handler
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

	// Register all self-registered endpoints
	registryMutex.RLock()
	for path, handler := range endpointRegistry {
		// Create a closure to capture the handler and ds
		capturedHandler := handler
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			capturedHandler(ds, w, r)
		})
	}
	registryMutex.RUnlock()

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

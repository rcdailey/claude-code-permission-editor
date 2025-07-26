package layout

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// EventType represents different types of layout events
type EventType int

const (
	// EventResize represents a window resize event
	EventResize EventType = iota
	// EventRecalculate represents a layout recalculation event
	EventRecalculate
	// EventComponentUpdate represents a component update event
	EventComponentUpdate
	// EventConstraintChange represents a constraint change event
	EventConstraintChange
	// EventConfigChange represents a configuration change event
	EventConfigChange
)

// LayoutEvent represents a layout system event
type LayoutEvent struct {
	Type      EventType
	Timestamp time.Time
	Data      interface{}
}

// EventHandler is a function that handles layout events
type EventHandler func(event LayoutEvent) error

// EventManager manages layout events and provides automatic handling
type EventManager struct {
	engine   *LayoutEngine
	handlers map[EventType][]EventHandler
	mutex    sync.RWMutex

	// Event queue for batching
	eventQueue []LayoutEvent
	queueMutex sync.Mutex

	// Performance optimization
	debounceDelay time.Duration
	lastResize    time.Time
	resizeTimer   *time.Timer

	// Configuration
	config EventConfig
}

// EventConfig contains configuration for event handling
type EventConfig struct {
	// Debounce settings
	EnableDebounce bool
	DebounceDelay  time.Duration

	// Batching settings
	EnableBatching bool
	BatchSize      int
	BatchTimeout   time.Duration

	// Performance settings
	EnableAsync   bool
	MaxConcurrent int

	// Debug settings
	LogEvents      bool
	LogPerformance bool
}

// DefaultEventConfig returns the default event configuration
func DefaultEventConfig() EventConfig {
	return EventConfig{
		EnableDebounce: true,
		DebounceDelay:  100 * time.Millisecond,
		EnableBatching: true,
		BatchSize:      10,
		BatchTimeout:   50 * time.Millisecond,
		EnableAsync:    false, // Keep synchronous for now
		MaxConcurrent:  4,
		LogEvents:      false,
		LogPerformance: false,
	}
}

// NewEventManager creates a new event manager
func NewEventManager(engine *LayoutEngine, config ...EventConfig) *EventManager {
	var eventConfig EventConfig
	if len(config) > 0 {
		eventConfig = config[0]
	} else {
		eventConfig = DefaultEventConfig()
	}

	em := &EventManager{
		engine:        engine,
		handlers:      make(map[EventType][]EventHandler),
		eventQueue:    []LayoutEvent{},
		debounceDelay: eventConfig.DebounceDelay,
		config:        eventConfig,
	}

	// Register default handlers
	em.registerDefaultHandlers()

	return em
}

// registerDefaultHandlers registers the default event handlers
func (em *EventManager) registerDefaultHandlers() {
	// Resize handler
	em.AddHandler(EventResize, func(event LayoutEvent) error {
		if resizeData, ok := event.Data.(ResizeEventData); ok {
			return em.handleResize(resizeData)
		}
		return fmt.Errorf("invalid resize event data")
	})

	// Recalculate handler
	em.AddHandler(EventRecalculate, func(_ LayoutEvent) error {
		return em.handleRecalculate()
	})

	// Component update handler
	em.AddHandler(EventComponentUpdate, func(event LayoutEvent) error {
		if updateData, ok := event.Data.(ComponentUpdateData); ok {
			return em.handleComponentUpdate(updateData)
		}
		return fmt.Errorf("invalid component update event data")
	})

	// Constraint change handler
	em.AddHandler(EventConstraintChange, func(event LayoutEvent) error {
		if constraintData, ok := event.Data.(ConstraintChangeData); ok {
			return em.handleConstraintChange(constraintData)
		}
		return fmt.Errorf("invalid constraint change event data")
	})
}

// AddHandler adds an event handler for a specific event type
func (em *EventManager) AddHandler(eventType EventType, handler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.handlers[eventType] = append(em.handlers[eventType], handler)
}

// RemoveHandler removes an event handler (Note: removes all handlers of that type)
func (em *EventManager) RemoveHandler(eventType EventType) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	delete(em.handlers, eventType)
}

// EmitEvent emits an event to all registered handlers
func (em *EventManager) EmitEvent(eventType EventType, data interface{}) error {
	event := LayoutEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}


	if em.config.EnableBatching {
		return em.queueEvent(event)
	}

	return em.processEvent(event)
}

// queueEvent adds an event to the processing queue
func (em *EventManager) queueEvent(event LayoutEvent) error {
	em.queueMutex.Lock()
	defer em.queueMutex.Unlock()

	em.eventQueue = append(em.eventQueue, event)

	// Process queue if it's full or after timeout
	if len(em.eventQueue) >= em.config.BatchSize {
		return em.processBatch()
	}

	// Set timer for batch timeout if not already set
	if len(em.eventQueue) == 1 {
		time.AfterFunc(em.config.BatchTimeout, func() {
			em.queueMutex.Lock()
			defer em.queueMutex.Unlock()
			if len(em.eventQueue) > 0 {
				_ = em.processBatch()
			}
		})
	}

	return nil
}

// processBatch processes all events in the queue
func (em *EventManager) processBatch() error {
	if len(em.eventQueue) == 0 {
		return nil
	}

	events := make([]LayoutEvent, len(em.eventQueue))
	copy(events, em.eventQueue)
	em.eventQueue = em.eventQueue[:0] // Clear queue

	var lastError error
	for _, event := range events {
		if err := em.processEvent(event); err != nil {
			lastError = err
		}
	}

	return lastError
}

// processEvent processes a single event
func (em *EventManager) processEvent(event LayoutEvent) error {
	em.mutex.RLock()
	handlers := em.handlers[event.Type]
	em.mutex.RUnlock()

	var lastError error
	for _, handler := range handlers {
		if err := handler(event); err != nil {
			lastError = err
		}
	}

	return lastError
}

// HandleBubbleTeaMessage handles incoming Bubble Tea messages and converts them to layout events
func (em *EventManager) HandleBubbleTeaMessage(msg tea.Msg) ([]tea.Cmd, error) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		return em.handleWindowSizeMsg(m)
	case tea.KeyMsg:
		return em.handleKeyMsg(m)
	default:
		// Forward other messages to components
		return em.forwardToComponents(msg), nil
	}
}

// handleWindowSizeMsg handles window resize messages
func (em *EventManager) handleWindowSizeMsg(msg tea.WindowSizeMsg) ([]tea.Cmd, error) {
	if em.config.EnableDebounce {
		return em.handleDebouncedResize(msg)
	}

	err := em.EmitEvent(EventResize, ResizeEventData{
		Width:  msg.Width,
		Height: msg.Height,
	})

	return []tea.Cmd{}, err
}

// handleDebouncedResize handles debounced resize events
func (em *EventManager) handleDebouncedResize(msg tea.WindowSizeMsg) ([]tea.Cmd, error) {
	em.lastResize = time.Now()

	if em.resizeTimer != nil {
		em.resizeTimer.Stop()
	}

	em.resizeTimer = time.AfterFunc(em.debounceDelay, func() {
		_ = em.EmitEvent(EventResize, ResizeEventData{
			Width:  msg.Width,
			Height: msg.Height,
		})
	})

	return []tea.Cmd{}, nil
}

// handleKeyMsg handles keyboard messages
func (em *EventManager) handleKeyMsg(msg tea.KeyMsg) ([]tea.Cmd, error) {
	// For now, just forward to components
	return em.forwardToComponents(msg), nil
}

// forwardToComponents forwards messages to all registered components
func (em *EventManager) forwardToComponents(msg tea.Msg) []tea.Cmd {
	return em.engine.Update(msg)
}

// Event data structures

// ResizeEventData contains data for resize events
type ResizeEventData struct {
	Width  int
	Height int
}

// ComponentUpdateData contains data for component update events
type ComponentUpdateData struct {
	ComponentID string
	Message     tea.Msg
}

// ConstraintChangeData contains data for constraint change events
type ConstraintChangeData struct {
	ComponentID    string
	OldConstraints ConstraintSet
	NewConstraints ConstraintSet
}

// Default event handlers

// handleResize handles resize events
func (em *EventManager) handleResize(data ResizeEventData) error {
	err := em.engine.HandleResize(data.Width, data.Height)

	// Note: Performance logging now handled by slog system

	return err
}

// handleRecalculate handles recalculation events
func (em *EventManager) handleRecalculate() error {
	err := em.engine.Recalculate()

	return err
}

// handleComponentUpdate handles component update events
func (em *EventManager) handleComponentUpdate(data ComponentUpdateData) error {
	// Get the component and forward the update message
	if wrapper, exists := em.engine.components.Get(data.ComponentID); exists {
		wrapper.Update(data.Message)
	}

	return nil
}

// handleConstraintChange handles constraint change events
func (em *EventManager) handleConstraintChange(_ ConstraintChangeData) error {
	// Trigger recalculation when constraints change
	return em.EmitEvent(EventRecalculate, nil)
}

// Utility functions


// GetEventStats returns statistics about event handling
func (em *EventManager) GetEventStats() EventStats {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	return EventStats{
		QueueSize:      len(em.eventQueue),
		HandlerCounts:  em.getHandlerCounts(),
		LastResize:     em.lastResize,
		DebounceActive: em.resizeTimer != nil,
	}
}

// getHandlerCounts returns the number of handlers for each event type
func (em *EventManager) getHandlerCounts() map[EventType]int {
	counts := make(map[EventType]int)
	for eventType, handlers := range em.handlers {
		counts[eventType] = len(handlers)
	}
	return counts
}

// EventStats contains statistics about event handling
type EventStats struct {
	QueueSize      int
	HandlerCounts  map[EventType]int
	LastResize     time.Time
	DebounceActive bool
}

// String returns a string representation of event stats
func (es EventStats) String() string {
	result := "Event Stats:\n"
	result += fmt.Sprintf("  Queue Size: %d\n", es.QueueSize)
	result += fmt.Sprintf("  Debounce Active: %t\n", es.DebounceActive)
	result += fmt.Sprintf("  Last Resize: %s\n", es.LastResize.Format("15:04:05.000"))
	result += "  Handler Counts:\n"

	for eventType, count := range es.HandlerCounts {
		result += fmt.Sprintf("    %s: %d\n", eventTypeToString(eventType), count)
	}

	return result
}

// eventTypeToString converts event type to string (standalone function)
func eventTypeToString(eventType EventType) string {
	switch eventType {
	case EventResize:
		return "resize"
	case EventRecalculate:
		return "recalculate"
	case EventComponentUpdate:
		return "component_update"
	case EventConstraintChange:
		return "constraint_change"
	case EventConfigChange:
		return "config_change"
	default:
		return "unknown"
	}
}

// Performance monitoring

// PerformanceMonitor monitors layout performance
type PerformanceMonitor struct {
	eventManager *EventManager
	metrics      PerformanceMetrics
	mutex        sync.RWMutex
}

// PerformanceMetrics contains performance metrics
type PerformanceMetrics struct {
	ResizeCount            int
	RecalculateCount       int
	AverageResizeTime      time.Duration
	AverageRecalculateTime time.Duration
	LastResizeTime         time.Duration
	LastRecalculateTime    time.Duration
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(eventManager *EventManager) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		eventManager: eventManager,
		metrics:      PerformanceMetrics{},
	}

	// Add performance monitoring handlers
	eventManager.AddHandler(EventResize, pm.monitorResize)
	eventManager.AddHandler(EventRecalculate, pm.monitorRecalculate)

	return pm
}

// monitorResize monitors resize performance
func (pm *PerformanceMonitor) monitorResize(_ LayoutEvent) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.metrics.ResizeCount++
	// Performance timing would be added here

	return nil
}

// monitorRecalculate monitors recalculation performance
func (pm *PerformanceMonitor) monitorRecalculate(_ LayoutEvent) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.metrics.RecalculateCount++
	// Performance timing would be added here

	return nil
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.metrics
}

// ResetMetrics resets all performance metrics
func (pm *PerformanceMonitor) ResetMetrics() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.metrics = PerformanceMetrics{}
}

// Helper functions for integration

// CreateManagedLayoutEngine creates a layout engine with automatic event management
func CreateManagedLayoutEngine(width, height int, eventConfig ...EventConfig) (*LayoutEngine, *EventManager) {
	engine := NewLayoutEngine(width, height)
	manager := NewEventManager(engine, eventConfig...)

	return engine, manager
}

// AutoResizeHandler creates a handler that automatically handles resize events
func AutoResizeHandler(engine *LayoutEngine) EventHandler {
	return func(event LayoutEvent) error {
		if resizeData, ok := event.Data.(ResizeEventData); ok {
			return engine.HandleResize(resizeData.Width, resizeData.Height)
		}
		return fmt.Errorf("invalid resize event data")
	}
}

// BatchedUpdateHandler creates a handler that batches component updates
func BatchedUpdateHandler(engine *LayoutEngine) EventHandler {
	return func(_ LayoutEvent) error {
		// Implement batched updates
		return engine.Recalculate()
	}
}

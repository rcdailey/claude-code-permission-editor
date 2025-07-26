package layout

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Component represents a UI component that can be managed by the layout engine
type Component interface {
	// SetSize updates the component's dimensions
	SetSize(width, height int)

	// GetSize returns the component's current dimensions
	GetSize() (width, height int)

	// Update handles Bubble Tea update messages
	Update(msg tea.Msg) tea.Cmd

	// View renders the component
	View() string
}

// ComponentWrapper wraps existing Bubble Tea components to make them layout-compatible
type ComponentWrapper struct {
	id          string
	component   interface{}
	width       int
	height      int
	constraints ConstraintSet
	position    Position
	dirty       bool // true if component needs update
	content     string // rendered content for absolute positioning
}

// Position represents the calculated position of a component
type Position struct {
	X, Y int
}

// NewBasicComponent creates a basic component wrapper with the given ID
func NewBasicComponent(id string) *ComponentWrapper {
	return &ComponentWrapper{
		id:          id,
		component:   nil, // Basic components don't wrap anything
		width:       0,
		height:      0,
		constraints: NewConstraintSet(),
		position:    Position{X: 0, Y: 0},
		dirty:       true,
	}
}

// NewComponentWrapper creates a new component wrapper
func NewComponentWrapper(id string, component interface{}, constraints ...Constraint) *ComponentWrapper {
	return &ComponentWrapper{
		id:          id,
		component:   component,
		width:       0,
		height:      0,
		constraints: NewConstraintSet(constraints...),
		position:    Position{X: 0, Y: 0},
		dirty:       true,
	}
}

// ID returns the component's unique identifier
func (cw *ComponentWrapper) ID() string {
	return cw.id
}

// SetSize updates the component's dimensions
func (cw *ComponentWrapper) SetSize(width, height int) {
	if cw.width != width || cw.height != height {
		cw.width = width
		cw.height = height
		cw.dirty = true
		cw.applyDimensions()
	}
}

// GetSize returns the component's current dimensions
func (cw *ComponentWrapper) GetSize() (width, height int) {
	return cw.width, cw.height
}

// SetPosition updates the component's position
func (cw *ComponentWrapper) SetPosition(x, y int) {
	if cw.position.X != x || cw.position.Y != y {
		cw.position.X = x
		cw.position.Y = y
		cw.dirty = true
	}
}

// GetPosition returns the component's current position
func (cw *ComponentWrapper) GetPosition() Position {
	return cw.position
}

// Constraints returns the component's constraint set
func (cw *ComponentWrapper) Constraints() ConstraintSet {
	return cw.constraints
}

// SetConstraints updates the component's constraints
func (cw *ComponentWrapper) SetConstraints(constraints ...Constraint) {
	cw.constraints = NewConstraintSet(constraints...)
	cw.dirty = true
}

// AddConstraint adds a constraint to the component
func (cw *ComponentWrapper) AddConstraint(constraint Constraint) {
	cw.constraints.AddConstraint(constraint)
	cw.dirty = true
}

// IsDirty returns true if the component needs to be updated
func (cw *ComponentWrapper) IsDirty() bool {
	return cw.dirty
}

// MarkClean marks the component as clean (up to date)
func (cw *ComponentWrapper) MarkClean() {
	cw.dirty = false
}

// Update handles Bubble Tea update messages
func (cw *ComponentWrapper) Update(msg tea.Msg) tea.Cmd {
	// Try to update the wrapped component if it supports the Update method
	if updatable, ok := cw.component.(interface {
		Update(tea.Msg) (tea.Model, tea.Cmd)
	}); ok {
		model, cmd := updatable.Update(msg)
		cw.component = model
		return cmd
	}

	// If the component doesn't support Update, return nil
	return nil
}

// View renders the component
func (cw *ComponentWrapper) View() string {
	// Return stored content if available (for absolute positioning)
	if cw.content != "" {
		return cw.content
	}

	// Try to render the wrapped component if it supports the View method
	if viewable, ok := cw.component.(interface {
		View() string
	}); ok {
		return viewable.View()
	}

	// If the component doesn't support View, return empty string
	return ""
}

// SetContent sets the rendered content for this component (used for absolute positioning)
func (cw *ComponentWrapper) SetContent(content string) {
	if cw.content != content {
		cw.content = content
		cw.dirty = true
	}
}

// GetContent returns the stored content
func (cw *ComponentWrapper) GetContent() string {
	return cw.content
}

// HasContent returns true if content has been set
func (cw *ComponentWrapper) HasContent() bool {
	return cw.content != ""
}

// GetComponent returns the underlying component
func (cw *ComponentWrapper) GetComponent() interface{} {
	return cw.component
}

// applyDimensions applies the current dimensions to the underlying component
func (cw *ComponentWrapper) applyDimensions() {
	// Try different common methods to set dimensions
	switch comp := cw.component.(type) {
	case interface{ SetWidth(int) }:
		comp.SetWidth(cw.width)
	case interface{ SetHeight(int) }:
		comp.SetHeight(cw.height)
	case interface{ SetSize(int, int) }:
		comp.SetSize(cw.width, cw.height)
	}

	// Handle specific Bubble Tea component types
	cw.applyBubbleTeaComponentDimensions()
}

// applyBubbleTeaComponentDimensions applies dimensions to known Bubble Tea components
func (cw *ComponentWrapper) applyBubbleTeaComponentDimensions() {
	// This will be implemented in integration.go to avoid circular dependencies
	// For now, we'll use reflection or type assertions for common components

	// Try to set width and height separately if available
	if widthSetter, ok := cw.component.(interface{ SetWidth(int) }); ok {
		widthSetter.SetWidth(cw.width)
	}

	if heightSetter, ok := cw.component.(interface{ SetHeight(int) }); ok {
		heightSetter.SetHeight(cw.height)
	}
}

// ComponentRegistry manages a collection of components
type ComponentRegistry struct {
	components map[string]*ComponentWrapper
	order      []string // Track order of component registration
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]*ComponentWrapper),
		order:      []string{},
	}
}

// Register adds a component to the registry
func (cr *ComponentRegistry) Register(id string, component interface{}, constraints ...Constraint) error {
	if _, exists := cr.components[id]; exists {
		return fmt.Errorf("component with ID '%s' already exists", id)
	}

	wrapper := NewComponentWrapper(id, component, constraints...)
	cr.components[id] = wrapper
	cr.order = append(cr.order, id)

	return nil
}

// RegisterWrapper adds a pre-configured ComponentWrapper to the registry
func (cr *ComponentRegistry) RegisterWrapper(wrapper *ComponentWrapper) error {
	if _, exists := cr.components[wrapper.ID()]; exists {
		return fmt.Errorf("component with ID '%s' already exists", wrapper.ID())
	}

	cr.components[wrapper.ID()] = wrapper
	cr.order = append(cr.order, wrapper.ID())
	return nil
}

// Unregister removes a component from the registry
func (cr *ComponentRegistry) Unregister(id string) error {
	if _, exists := cr.components[id]; !exists {
		return fmt.Errorf("component with ID '%s' not found", id)
	}

	delete(cr.components, id)

	// Remove from order slice
	for i, orderID := range cr.order {
		if orderID == id {
			cr.order = append(cr.order[:i], cr.order[i+1:]...)
			break
		}
	}

	return nil
}

// Get returns a component by ID
func (cr *ComponentRegistry) Get(id string) (*ComponentWrapper, bool) {
	component, exists := cr.components[id]
	return component, exists
}

// GetAll returns all components
func (cr *ComponentRegistry) GetAll() map[string]*ComponentWrapper {
	return cr.components
}

// GetInOrder returns components in the order they were registered
func (cr *ComponentRegistry) GetInOrder() []*ComponentWrapper {
	result := make([]*ComponentWrapper, 0, len(cr.order))
	for _, id := range cr.order {
		if component, exists := cr.components[id]; exists {
			result = append(result, component)
		}
	}
	return result
}

// GetIDs returns all component IDs
func (cr *ComponentRegistry) GetIDs() []string {
	return cr.order
}

// Count returns the number of registered components
func (cr *ComponentRegistry) Count() int {
	return len(cr.components)
}

// UpdateAll updates all components with the given message
func (cr *ComponentRegistry) UpdateAll(msg tea.Msg) []tea.Cmd {
	commands := make([]tea.Cmd, 0, len(cr.components))

	for _, component := range cr.components {
		if cmd := component.Update(msg); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// MarkAllClean marks all components as clean
func (cr *ComponentRegistry) MarkAllClean() {
	for _, component := range cr.components {
		component.MarkClean()
	}
}

// GetDirtyComponents returns all components that need updating
func (cr *ComponentRegistry) GetDirtyComponents() []*ComponentWrapper {
	var dirty []*ComponentWrapper
	for _, component := range cr.components {
		if component.IsDirty() {
			dirty = append(dirty, component)
		}
	}
	return dirty
}

// ValidateAll validates all component constraints
func (cr *ComponentRegistry) ValidateAll() error {
	for id, component := range cr.components {
		if err := component.Constraints().Validate(); err != nil {
			return fmt.Errorf("component '%s' validation failed: %w", id, err)
		}
	}
	return nil
}

// ComponentInfo provides information about a component for debugging
type ComponentInfo struct {
	ID          string
	Width       int
	Height      int
	Position    Position
	Constraints []string
	IsDirty     bool
}

// GetComponentInfo returns debugging information about a component
func (cw *ComponentWrapper) GetComponentInfo() ComponentInfo {
	constraintStrings := make([]string, 0, len(cw.constraints.All()))
	for _, constraint := range cw.constraints.All() {
		constraintStrings = append(constraintStrings, constraint.String())
	}

	return ComponentInfo{
		ID:          cw.id,
		Width:       cw.width,
		Height:      cw.height,
		Position:    cw.position,
		Constraints: constraintStrings,
		IsDirty:     cw.dirty,
	}
}

// GetAllComponentInfo returns debugging information about all components
func (cr *ComponentRegistry) GetAllComponentInfo() []ComponentInfo {
	info := make([]ComponentInfo, 0, len(cr.components))
	for _, component := range cr.components {
		info = append(info, component.GetComponentInfo())
	}
	return info
}

// ComponentUpdate represents an update operation for a component
type ComponentUpdate struct {
	ID       string
	Width    int
	Height   int
	Position Position
}

// ApplyUpdate applies an update to a component
func (cw *ComponentWrapper) ApplyUpdate(update ComponentUpdate) {
	cw.SetSize(update.Width, update.Height)
	cw.SetPosition(update.Position.X, update.Position.Y)
}

// BatchUpdate represents a batch of component updates
type BatchUpdate struct {
	Updates []ComponentUpdate
}

// ApplyBatchUpdate applies multiple updates to components
func (cr *ComponentRegistry) ApplyBatchUpdate(batch BatchUpdate) error {
	for _, update := range batch.Updates {
		component, exists := cr.components[update.ID]
		if !exists {
			return fmt.Errorf("component with ID '%s' not found for update", update.ID)
		}
		component.ApplyUpdate(update)
	}
	return nil
}

// CreateBatchUpdate creates a batch update from component updates
func CreateBatchUpdate(updates ...ComponentUpdate) BatchUpdate {
	return BatchUpdate{Updates: updates}
}

// Helper functions for common component operations

// SetComponentSize is a helper function to set a component's size
func SetComponentSize(component interface{}, width, height int) {
	switch comp := component.(type) {
	case interface{ SetWidth(int) }:
		comp.SetWidth(width)
	case interface{ SetHeight(int) }:
		comp.SetHeight(height)
	case interface{ SetSize(int, int) }:
		comp.SetSize(width, height)
	}
}

// GetComponentSize is a helper function to get a component's size
func GetComponentSize(component interface{}) (width, height int) {
	switch comp := component.(type) {
	case interface{ Width() int }:
		return comp.Width(), 0
	case interface{ Height() int }:
		return 0, comp.Height()
	case interface{ GetSize() (int, int) }:
		return comp.GetSize()
	}
	return 0, 0
}

package layout

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BubbleTeaIntegration provides helpers for integrating with Bubble Tea components
type BubbleTeaIntegration struct {
	engine *LayoutEngine
}

// NewBubbleTeaIntegration creates a new Bubble Tea integration helper
func NewBubbleTeaIntegration(engine *LayoutEngine) *BubbleTeaIntegration {
	return &BubbleTeaIntegration{engine: engine}
}

// AddList adds a Bubble Tea list component to the layout
func (bti *BubbleTeaIntegration) AddList(id string, listModel list.Model, constraints ...Constraint) error {
	wrapper := &BubbleTeaListWrapper{
		id:        id,
		listModel: listModel,
		wrapper:   NewComponentWrapper(id, listModel, constraints...),
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// AddListWithStyle adds a Bubble Tea list component to the layout with style-aware sizing
func (bti *BubbleTeaIntegration) AddListWithStyle(id string, listModel list.Model,
	style lipgloss.Style, constraints ...Constraint) error {
	wrapper := &BubbleTeaListWrapper{
		id:           id,
		listModel:    listModel,
		wrapper:      NewComponentWrapper(id, listModel, constraints...),
		contentStyle: style,
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// AddTable adds a Bubble Tea table component to the layout
func (bti *BubbleTeaIntegration) AddTable(id string, tableModel table.Model, constraints ...Constraint) error {
	wrapper := &BubbleTeaTableWrapper{
		id:         id,
		tableModel: tableModel,
		wrapper:    NewComponentWrapper(id, tableModel, constraints...),
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// AddTableWithStyle adds a Bubble Tea table component to the layout with style-aware sizing
func (bti *BubbleTeaIntegration) AddTableWithStyle(id string, tableModel table.Model,
	style lipgloss.Style, constraints ...Constraint) error {
	wrapper := &BubbleTeaTableWrapper{
		id:           id,
		tableModel:   tableModel,
		wrapper:      NewComponentWrapper(id, tableModel, constraints...),
		contentStyle: style,
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// AddViewport adds a Bubble Tea viewport component to the layout
func (bti *BubbleTeaIntegration) AddViewport(id string, viewportModel viewport.Model, constraints ...Constraint) error {
	wrapper := &BubbleTeaViewportWrapper{
		id:            id,
		viewportModel: viewportModel,
		wrapper:       NewComponentWrapper(id, viewportModel, constraints...),
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// AddCustomBubbleTeaComponent adds a custom Bubble Tea component to the layout
func (bti *BubbleTeaIntegration) AddCustomBubbleTeaComponent(id string, component tea.Model,
	constraints ...Constraint) error {
	wrapper := &CustomBubbleTeaWrapper{
		id:        id,
		component: component,
		wrapper:   NewComponentWrapper(id, component, constraints...),
	}

	return bti.engine.components.Register(id, wrapper, constraints...)
}

// Specialized wrappers for common Bubble Tea components

// BubbleTeaListWrapper wraps a Bubble Tea list component
type BubbleTeaListWrapper struct {
	id           string
	listModel    list.Model
	wrapper      *ComponentWrapper
	contentStyle lipgloss.Style // Associated style for content area calculation
}

// SetSize applies dimensions to the list component
func (btlw *BubbleTeaListWrapper) SetSize(width, height int) {
	btlw.listModel.SetWidth(width)
	btlw.listModel.SetHeight(height)
	btlw.wrapper.SetSize(width, height)
}

// SetSizeWithStyle applies dimensions accounting for style frame size
func (btlw *BubbleTeaListWrapper) SetSizeWithStyle(containerWidth, containerHeight int, style lipgloss.Style) {
	// Calculate content area by subtracting frame size (borders, padding, margins)
	frameWidth, frameHeight := style.GetFrameSize()

	// Apply minimum content dimensions
	contentWidth := maxInt(40, containerWidth-frameWidth)
	contentHeight := maxInt(4, containerHeight-frameHeight)

	// Set the actual content size on the list component
	btlw.listModel.SetWidth(contentWidth)
	btlw.listModel.SetHeight(contentHeight)

	// Set the container size on the wrapper
	btlw.wrapper.SetSize(containerWidth, containerHeight)

	// Store the style for future reference
	btlw.contentStyle = style
}

// Helper function for minimum integer calculation
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetSize returns the current dimensions
func (btlw *BubbleTeaListWrapper) GetSize() (width, height int) {
	return btlw.listModel.Width(), btlw.listModel.Height()
}

// Update handles Bubble Tea messages
func (btlw *BubbleTeaListWrapper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	btlw.listModel, cmd = btlw.listModel.Update(msg)
	return cmd
}

// View renders the list component
func (btlw *BubbleTeaListWrapper) View() string {
	return btlw.listModel.View()
}

// GetList returns the underlying list model
func (btlw *BubbleTeaListWrapper) GetList() list.Model {
	return btlw.listModel
}

// BubbleTeaTableWrapper wraps a Bubble Tea table component
type BubbleTeaTableWrapper struct {
	id           string
	tableModel   table.Model
	wrapper      *ComponentWrapper
	contentStyle lipgloss.Style // Associated style for content area calculation
}

// SetSize applies dimensions to the table component
func (bttw *BubbleTeaTableWrapper) SetSize(width, height int) {
	bttw.tableModel.SetWidth(width)
	bttw.tableModel.SetHeight(height)
	bttw.wrapper.SetSize(width, height)
}

// SetSizeWithStyle applies dimensions accounting for style frame size
func (bttw *BubbleTeaTableWrapper) SetSizeWithStyle(containerWidth, containerHeight int, style lipgloss.Style) {
	// Calculate content area by subtracting frame size (borders, padding, margins)
	frameWidth, frameHeight := style.GetFrameSize()

	// Apply minimum content dimensions (tables need smaller minimum height)
	contentWidth := maxInt(40, containerWidth-frameWidth)
	contentHeight := maxInt(2, containerHeight-frameHeight)

	// Set the actual content size on the table component
	bttw.tableModel.SetWidth(contentWidth)
	bttw.tableModel.SetHeight(contentHeight)

	// Set the container size on the wrapper
	bttw.wrapper.SetSize(containerWidth, containerHeight)

	// Store the style for future reference
	bttw.contentStyle = style
}

// GetSize returns the current dimensions
func (bttw *BubbleTeaTableWrapper) GetSize() (width, height int) {
	return bttw.tableModel.Width(), bttw.tableModel.Height()
}

// Update handles Bubble Tea messages
func (bttw *BubbleTeaTableWrapper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	bttw.tableModel, cmd = bttw.tableModel.Update(msg)
	return cmd
}

// View renders the table component
func (bttw *BubbleTeaTableWrapper) View() string {
	return bttw.tableModel.View()
}

// GetTable returns the underlying table model
func (bttw *BubbleTeaTableWrapper) GetTable() table.Model {
	return bttw.tableModel
}

// BubbleTeaViewportWrapper wraps a Bubble Tea viewport component
type BubbleTeaViewportWrapper struct {
	id            string
	viewportModel viewport.Model
	wrapper       *ComponentWrapper
}

// SetSize applies dimensions to the viewport component
func (btvw *BubbleTeaViewportWrapper) SetSize(width, height int) {
	btvw.viewportModel.Width = width
	btvw.viewportModel.Height = height
	btvw.wrapper.SetSize(width, height)
}

// GetSize returns the current dimensions
func (btvw *BubbleTeaViewportWrapper) GetSize() (width, height int) {
	return btvw.viewportModel.Width, btvw.viewportModel.Height
}

// Update handles Bubble Tea messages
func (btvw *BubbleTeaViewportWrapper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	btvw.viewportModel, cmd = btvw.viewportModel.Update(msg)
	return cmd
}

// View renders the viewport component
func (btvw *BubbleTeaViewportWrapper) View() string {
	return btvw.viewportModel.View()
}

// GetViewport returns the underlying viewport model
func (btvw *BubbleTeaViewportWrapper) GetViewport() viewport.Model {
	return btvw.viewportModel
}

// CustomBubbleTeaWrapper wraps custom Bubble Tea components
type CustomBubbleTeaWrapper struct {
	id        string
	component tea.Model
	wrapper   *ComponentWrapper
}

// SetSize attempts to apply dimensions to the custom component
func (cbtw *CustomBubbleTeaWrapper) SetSize(width, height int) {
	// Try to apply dimensions using reflection or type assertion
	ApplyDimensionsToComponent(cbtw.component, width, height)
	cbtw.wrapper.SetSize(width, height)
}

// GetSize attempts to get dimensions from the custom component
func (cbtw *CustomBubbleTeaWrapper) GetSize() (width, height int) {
	return GetDimensionsFromComponent(cbtw.component)
}

// Update handles Bubble Tea messages
func (cbtw *CustomBubbleTeaWrapper) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	cbtw.component, cmd = cbtw.component.Update(msg)
	return cmd
}

// View renders the custom component
func (cbtw *CustomBubbleTeaWrapper) View() string {
	return cbtw.component.View()
}

// GetComponent returns the underlying component
func (cbtw *CustomBubbleTeaWrapper) GetComponent() tea.Model {
	return cbtw.component
}

// Helper functions for working with Bubble Tea components

// ApplyDimensionsToComponent attempts to apply dimensions to a component using various methods
func ApplyDimensionsToComponent(component interface{}, width, height int) {
	// Try specific Bubble Tea component types first
	switch comp := component.(type) {
	case *list.Model:
		comp.SetWidth(width)
		comp.SetHeight(height)
		return
	case list.Model:
		comp.SetWidth(width)
		comp.SetHeight(height)
		return
	case *table.Model:
		comp.SetWidth(width)
		comp.SetHeight(height)
		return
	case table.Model:
		comp.SetWidth(width)
		comp.SetHeight(height)
		return
	case *viewport.Model:
		comp.Width = width
		comp.Height = height
		return
	case viewport.Model:
		comp.Width = width
		comp.Height = height
		return
	}

	// Try common interface patterns
	if widthSetter, ok := component.(interface{ SetWidth(int) }); ok {
		widthSetter.SetWidth(width)
	}

	if heightSetter, ok := component.(interface{ SetHeight(int) }); ok {
		heightSetter.SetHeight(height)
	}

	if sizeSetter, ok := component.(interface{ SetSize(int, int) }); ok {
		sizeSetter.SetSize(width, height)
	}

	// Try using reflection as a last resort
	applyDimensionsWithReflection(component, width, height)
}

// GetDimensionsFromComponent attempts to get dimensions from a component
func GetDimensionsFromComponent(component interface{}) (width, height int) {
	// Try specific Bubble Tea component types first
	switch comp := component.(type) {
	case *list.Model:
		return comp.Width(), comp.Height()
	case list.Model:
		return comp.Width(), comp.Height()
	case *table.Model:
		return comp.Width(), comp.Height()
	case table.Model:
		return comp.Width(), comp.Height()
	case *viewport.Model:
		return comp.Width, comp.Height
	case viewport.Model:
		return comp.Width, comp.Height
	}

	// Try common interface patterns
	if widthGetter, ok := component.(interface{ Width() int }); ok {
		return widthGetter.Width(), 0
	}

	if heightGetter, ok := component.(interface{ Height() int }); ok {
		return 0, heightGetter.Height()
	}

	if sizeGetter, ok := component.(interface{ GetSize() (int, int) }); ok {
		return sizeGetter.GetSize()
	}

	// Try using reflection as a last resort
	return getDimensionsWithReflection(component)
}

// applyDimensionsWithReflection uses reflection to apply dimensions
func applyDimensionsWithReflection(component interface{}, width, height int) {
	value := reflect.ValueOf(component)

	// Dereference pointer if necessary
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Look for Width and Height fields
	if value.Kind() == reflect.Struct {
		widthField := value.FieldByName("Width")
		if widthField.IsValid() && widthField.CanSet() && widthField.Kind() == reflect.Int {
			widthField.SetInt(int64(width))
		}

		heightField := value.FieldByName("Height")
		if heightField.IsValid() && heightField.CanSet() && heightField.Kind() == reflect.Int {
			heightField.SetInt(int64(height))
		}
	}
}

// getDimensionsWithReflection uses reflection to get dimensions
func getDimensionsWithReflection(component interface{}) (width, height int) {
	value := reflect.ValueOf(component)

	// Dereference pointer if necessary
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Look for Width and Height fields
	if value.Kind() == reflect.Struct {
		widthField := value.FieldByName("Width")
		if widthField.IsValid() && widthField.Kind() == reflect.Int {
			width = int(widthField.Int())
		}

		heightField := value.FieldByName("Height")
		if heightField.IsValid() && heightField.Kind() == reflect.Int {
			height = int(heightField.Int())
		}
	}

	return width, height
}

// Lipgloss integration helpers

// ApplyLipglossStyle applies a lipgloss style to a component's content
func ApplyLipglossStyle(content string, style lipgloss.Style, width, height int) string {
	return style.Width(width).Height(height).Render(content)
}

// CreateStyledPanel creates a styled panel with the given content and constraints
func CreateStyledPanel(content string, style lipgloss.Style, constraints ConstraintSet) string {
	// Apply width and height from constraints if available
	width := 0
	height := 0

	if widthConstraint, exists := constraints.Get(ConstraintWidth); exists {
		if sizeConstraint, ok := widthConstraint.(SizeConstraint); ok {
			// For styling, we need actual pixel values
			// This is a simplified approach - in practice, you'd need the container dimensions
			width = sizeConstraint.Value().Calculate(100) // Assume 100 as placeholder
		}
	}

	if heightConstraint, exists := constraints.Get(ConstraintHeight); exists {
		if sizeConstraint, ok := heightConstraint.(SizeConstraint); ok {
			height = sizeConstraint.Value().Calculate(30) // Assume 30 as placeholder
		}
	}

	if width > 0 && height > 0 {
		style = style.Width(width).Height(height)
	}

	return style.Render(content)
}

// Layout builders for common patterns

// TwoPanelLayoutBuilder creates a common two-panel layout
type TwoPanelLayoutBuilder struct {
	engine      *LayoutEngine
	integration *BubbleTeaIntegration
}

// NewTwoPanelLayoutBuilder creates a new two-panel layout builder
func NewTwoPanelLayoutBuilder(width, height int) *TwoPanelLayoutBuilder {
	engine := NewLayoutEngine(width, height)
	integration := NewBubbleTeaIntegration(engine)

	return &TwoPanelLayoutBuilder{
		engine:      engine,
		integration: integration,
	}
}

// AddHeader adds a header component to the layout
func (tplb *TwoPanelLayoutBuilder) AddHeader(component tea.Model, height int) *TwoPanelLayoutBuilder {
	_ = tplb.integration.AddCustomBubbleTeaComponent("header", component,
		Height(Fixed(height)),
		Anchor(AnchorTop),
	)
	return tplb
}

// AddMainPanel adds the main panel component
func (tplb *TwoPanelLayoutBuilder) AddMainPanel(component tea.Model, weight float64) *TwoPanelLayoutBuilder {
	_ = tplb.integration.AddCustomBubbleTeaComponent("main", component,
		Height(Flex(weight)),
		Below("header", 1),
	)
	return tplb
}

// AddSecondaryPanel adds the secondary panel component
func (tplb *TwoPanelLayoutBuilder) AddSecondaryPanel(component tea.Model, weight float64) *TwoPanelLayoutBuilder {
	_ = tplb.integration.AddCustomBubbleTeaComponent("secondary", component,
		Height(Flex(weight)),
		Below("main", 1),
	)
	return tplb
}

// AddFooter adds a footer component to the layout
func (tplb *TwoPanelLayoutBuilder) AddFooter(component tea.Model, height int) *TwoPanelLayoutBuilder {
	_ = tplb.integration.AddCustomBubbleTeaComponent("footer", component,
		Height(Fixed(height)),
		Anchor(AnchorBottom),
	)
	return tplb
}

// Build returns the configured layout engine
func (tplb *TwoPanelLayoutBuilder) Build() *LayoutEngine {
	return tplb.engine
}

// Helper functions for specific use cases

// CreatePermissionEditorLayout creates a layout specifically for the permission editor
func CreatePermissionEditorLayout(width, height int,
	permissionsList list.Model,
	duplicatesTable table.Model,
	panelStyle lipgloss.Style) (*LayoutEngine, error) {
	engine := NewLayoutEngine(width, height)
	integration := NewBubbleTeaIntegration(engine)

	// Add header (you would create this component)
	headerComponent := &SimpleComponent{id: "header"}
	if err := integration.AddCustomBubbleTeaComponent("header", headerComponent,
		Height(Fixed(4)),
		Anchor(AnchorTop),
	); err != nil {
		return nil, fmt.Errorf("failed to add header: %w", err)
	}

	// Add permissions list with style-aware sizing
	if err := integration.AddListWithStyle("permissions", permissionsList, panelStyle,
		Height(Flex(0.7)),
		Below("header", 1),
	); err != nil {
		return nil, fmt.Errorf("failed to add permissions list: %w", err)
	}

	// Add duplicates table with style-aware sizing
	if err := integration.AddTableWithStyle("duplicates", duplicatesTable, panelStyle,
		Height(Flex(0.3)),
		Below("permissions", 1),
	); err != nil {
		return nil, fmt.Errorf("failed to add duplicates table: %w", err)
	}

	// Add footer (you would create this component)
	footerComponent := &SimpleComponent{id: "footer"}
	if err := integration.AddCustomBubbleTeaComponent("footer", footerComponent,
		Height(Fixed(2)),
		Anchor(AnchorBottom),
	); err != nil {
		return nil, fmt.Errorf("failed to add footer: %w", err)
	}

	return engine, nil
}

// MigrationHelper helps migrate existing code to use the layout engine
type MigrationHelper struct {
	engine *LayoutEngine
}

// NewMigrationHelper creates a new migration helper
func NewMigrationHelper(engine *LayoutEngine) *MigrationHelper {
	return &MigrationHelper{engine: engine}
}

// ReplaceManualLayout replaces manual layout calculations with engine-managed layout
func (mh *MigrationHelper) ReplaceManualLayout(componentID string,
	component interface{},
	oldWidth, oldHeight int,
	constraints ...Constraint) error {
	// If no constraints provided, try to infer from old dimensions
	if len(constraints) == 0 {
		// This is a simplified approach - in practice, you'd analyze the old layout logic
		constraints = []Constraint{
			Width(Fixed(oldWidth)),
			Height(Fixed(oldHeight)),
		}
	}

	return mh.engine.AddComponent(componentID, component, constraints...)
}

// ConvertViewportToManaged converts a manual viewport to engine-managed
func (mh *MigrationHelper) ConvertViewportToManaged(id string,
	viewport viewport.Model,
	constraints ...Constraint) error {
	integration := NewBubbleTeaIntegration(mh.engine)
	return integration.AddViewport(id, viewport, constraints...)
}

// Compatibility functions for existing code

// CreateCompatibleEngine creates an engine that's compatible with existing updateViewports patterns
func CreateCompatibleEngine(width, height int) *LayoutEngine {
	config := DefaultEngineConfig()
	config.AutoRecalculate = true // Enable automatic recalculation

	return NewLayoutEngine(width, height, config)
}

// UpdateViewportsReplacement provides a drop-in replacement for updateViewports function
func UpdateViewportsReplacement(engine *LayoutEngine, width, height int) error {
	return engine.HandleResize(width, height)
}

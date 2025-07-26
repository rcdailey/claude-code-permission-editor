package layout

import (
	"fmt"
)

// Constraint represents a layout constraint (width, height, position, etc.)
type Constraint interface {
	// Type returns the constraint type for identification
	Type() ConstraintType

	// Validate checks if the constraint is valid
	Validate() error

	// String returns a string representation for debugging
	String() string
}

// ConstraintType identifies the type of constraint
type ConstraintType string

const (
	// ConstraintWidth represents a width constraint
	ConstraintWidth ConstraintType = "width"
	// ConstraintHeight represents a height constraint
	ConstraintHeight ConstraintType = "height"
	// ConstraintMinWidth represents a minimum width constraint
	ConstraintMinWidth ConstraintType = "min_width"
	// ConstraintMinHeight represents a minimum height constraint
	ConstraintMinHeight ConstraintType = "min_height"
	// ConstraintMaxWidth represents a maximum width constraint
	ConstraintMaxWidth ConstraintType = "max_width"
	// ConstraintMaxHeight represents a maximum height constraint
	ConstraintMaxHeight ConstraintType = "max_height"

	// ConstraintAnchor represents an anchor position constraint
	ConstraintAnchor ConstraintType = "anchor"
	// ConstraintMargin represents a margin constraint
	ConstraintMargin ConstraintType = "margin"
	// ConstraintPadding represents a padding constraint
	ConstraintPadding ConstraintType = "padding"

	// ConstraintAbove represents a positioning constraint above another component
	ConstraintAbove ConstraintType = "above"
	// ConstraintBelow represents a positioning constraint below another component
	ConstraintBelow ConstraintType = "below"
	// ConstraintLeft represents a positioning constraint left of another component
	ConstraintLeft ConstraintType = "left"
	// ConstraintRight represents a positioning constraint right of another component
	ConstraintRight ConstraintType = "right"

	// ConstraintDependsOn represents a dependency constraint
	ConstraintDependsOn ConstraintType = "depends_on"
)

// SizeConstraint represents a size-based constraint (width, height, etc.)
type SizeConstraint struct {
	constraintType ConstraintType
	value          SizeValue
}

// Type returns the constraint type
func (sc SizeConstraint) Type() ConstraintType {
	return sc.constraintType
}

// Validate checks if the constraint is valid
func (sc SizeConstraint) Validate() error {
	if sc.value == nil {
		return fmt.Errorf("size constraint cannot have nil value")
	}
	return sc.value.Validate()
}

// String returns a string representation
func (sc SizeConstraint) String() string {
	return fmt.Sprintf("%s: %s", sc.constraintType, sc.value.String())
}

// Value returns the size value
func (sc SizeConstraint) Value() SizeValue {
	return sc.value
}

// DependencyConstraint represents a dependency constraint between components
type DependencyConstraint struct {
	constraintType ConstraintType
	dependencies   []string
}

// Type returns the constraint type
func (dc DependencyConstraint) Type() ConstraintType {
	return dc.constraintType
}

// Validate validates the dependency constraint
func (dc DependencyConstraint) Validate() error {
	if len(dc.dependencies) == 0 {
		return fmt.Errorf("dependency constraint must have at least one dependency")
	}
	return nil
}

// String returns a string representation
func (dc DependencyConstraint) String() string {
	return fmt.Sprintf("%s: %v", dc.constraintType, dc.dependencies)
}

// Dependencies returns the list of dependencies
func (dc DependencyConstraint) Dependencies() []string {
	return dc.dependencies
}

// SizeValue represents different ways to specify size
type SizeValue interface {
	// Calculate returns the actual pixel value given available space
	Calculate(availableSpace int) int

	// Validate checks if the size value is valid
	Validate() error

	// String returns a string representation
	String() string

	// IsFlexible returns true if this size can be adjusted during layout
	IsFlexible() bool
}

// FixedSize represents a fixed pixel size
type FixedSize struct {
	pixels int
}

// Calculate returns the fixed pixel value
func (fs FixedSize) Calculate(_ int) int {
	return fs.pixels
}

// Validate checks if the fixed size is valid
func (fs FixedSize) Validate() error {
	if fs.pixels < 0 {
		return fmt.Errorf("fixed size cannot be negative: %d", fs.pixels)
	}
	return nil
}

// String returns a string representation
func (fs FixedSize) String() string {
	return fmt.Sprintf("%dpx", fs.pixels)
}

// IsFlexible returns false for fixed sizes
func (fs FixedSize) IsFlexible() bool {
	return false
}


// FlexSize represents a flexible size based on weight
type FlexSize struct {
	weight float64
}

// Calculate returns a weighted portion of available space
// Note: This requires additional context about other flex items, so it returns 0 here
// The actual calculation happens in the layout engine
func (fs FlexSize) Calculate(_ int) int {
	// This is a placeholder - actual calculation happens in the layout engine
	// with knowledge of all flex items and their weights
	return 0
}

// Validate checks if the flex weight is valid
func (fs FlexSize) Validate() error {
	if fs.weight < 0 {
		return fmt.Errorf("flex weight cannot be negative: %.2f", fs.weight)
	}
	return nil
}

// String returns a string representation
func (fs FlexSize) String() string {
	return fmt.Sprintf("flex(%.2f)", fs.weight)
}

// IsFlexible returns true for flex sizes
func (fs FlexSize) IsFlexible() bool {
	return true
}

// Weight returns the flex weight
func (fs FlexSize) Weight() float64 {
	return fs.weight
}

// AnchorConstraint represents positioning anchor points
type AnchorConstraint struct {
	anchor AnchorPosition
}

// Type returns the constraint type
func (ac AnchorConstraint) Type() ConstraintType {
	return ConstraintAnchor
}

// Validate checks if the anchor is valid
func (ac AnchorConstraint) Validate() error {
	// All anchor positions are valid
	return nil
}

// String returns a string representation
func (ac AnchorConstraint) String() string {
	return fmt.Sprintf("anchor: %s", ac.anchor)
}

// Position returns the anchor position
func (ac AnchorConstraint) Position() AnchorPosition {
	return ac.anchor
}

// AnchorPosition represents different anchor positions
type AnchorPosition string

const (
	// AnchorTopLeft anchors component to top-left corner
	AnchorTopLeft AnchorPosition = "top_left"
	// AnchorTop anchors component to top center
	AnchorTop AnchorPosition = "top"
	// AnchorTopRight anchors component to top-right corner
	AnchorTopRight AnchorPosition = "top_right"
	// AnchorLeft anchors component to left center
	AnchorLeft AnchorPosition = "left"
	// AnchorCenter anchors component to center
	AnchorCenter AnchorPosition = "center"
	// AnchorRight anchors component to right center
	AnchorRight AnchorPosition = "right"
	// AnchorBottomLeft anchors component to bottom-left corner
	AnchorBottomLeft AnchorPosition = "bottom_left"
	// AnchorBottom anchors component to bottom center
	AnchorBottom AnchorPosition = "bottom"
	// AnchorBottomRight anchors component to bottom-right corner
	AnchorBottomRight AnchorPosition = "bottom_right"
)

// SpacingConstraint represents margin or padding
type SpacingConstraint struct {
	constraintType           ConstraintType
	top, right, bottom, left int
}

// Type returns the constraint type
func (sc SpacingConstraint) Type() ConstraintType {
	return sc.constraintType
}

// Validate checks if the spacing values are valid
func (sc SpacingConstraint) Validate() error {
	if sc.top < 0 || sc.right < 0 || sc.bottom < 0 || sc.left < 0 {
		return fmt.Errorf("spacing values cannot be negative: top=%d, right=%d, bottom=%d, left=%d",
			sc.top, sc.right, sc.bottom, sc.left)
	}
	return nil
}

// String returns a string representation
func (sc SpacingConstraint) String() string {
	if sc.top == sc.right && sc.right == sc.bottom && sc.bottom == sc.left {
		return fmt.Sprintf("%s: %d", sc.constraintType, sc.top)
	}
	return fmt.Sprintf("%s: %d %d %d %d", sc.constraintType, sc.top, sc.right, sc.bottom, sc.left)
}

// Values returns the spacing values
func (sc SpacingConstraint) Values() (top, right, bottom, left int) {
	return sc.top, sc.right, sc.bottom, sc.left
}

// RelationshipConstraint represents positioning relative to other components
type RelationshipConstraint struct {
	constraintType ConstraintType
	targetID       string
	offset         int
}

// Type returns the constraint type
func (rc RelationshipConstraint) Type() ConstraintType {
	return rc.constraintType
}

// Validate checks if the relationship is valid
func (rc RelationshipConstraint) Validate() error {
	if rc.targetID == "" {
		return fmt.Errorf("relationship constraint requires a target component ID")
	}
	if rc.offset < 0 {
		return fmt.Errorf("relationship offset cannot be negative: %d", rc.offset)
	}
	return nil
}

// String returns a string representation
func (rc RelationshipConstraint) String() string {
	if rc.offset == 0 {
		return fmt.Sprintf("%s: %s", rc.constraintType, rc.targetID)
	}
	return fmt.Sprintf("%s: %s+%d", rc.constraintType, rc.targetID, rc.offset)
}

// TargetID returns the target component ID
func (rc RelationshipConstraint) TargetID() string {
	return rc.targetID
}

// Offset returns the offset value
func (rc RelationshipConstraint) Offset() int {
	return rc.offset
}

// Constructor functions for constraints (CSS-like API)

// NewSizeConstraint creates a size constraint for any constraint type
func NewSizeConstraint(value SizeValue) SizeConstraint {
	return SizeConstraint{
		constraintType: ConstraintWidth, // Default to width
		value:          value,
	}
}

// Width creates a width constraint
func Width(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintWidth, value: value}
}

// Height creates a height constraint
func Height(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintHeight, value: value}
}

// MinWidth creates a minimum width constraint
func MinWidth(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintMinWidth, value: value}
}

// MinHeight creates a minimum height constraint
func MinHeight(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintMinHeight, value: value}
}

// MaxWidth creates a maximum width constraint
func MaxWidth(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintMaxWidth, value: value}
}

// MaxHeight creates a maximum height constraint
func MaxHeight(value SizeValue) SizeConstraint {
	return SizeConstraint{constraintType: ConstraintMaxHeight, value: value}
}

// NewDependencyConstraint creates a dependency constraint
func NewDependencyConstraint(dependencies []string) DependencyConstraint {
	return DependencyConstraint{
		constraintType: ConstraintDependsOn,
		dependencies:   dependencies,
	}
}

// Anchor creates an anchor constraint
func Anchor(position AnchorPosition) AnchorConstraint {
	return AnchorConstraint{anchor: position}
}

// Margin creates a margin constraint
func Margin(top, right, bottom, left int) SpacingConstraint {
	return SpacingConstraint{constraintType: ConstraintMargin, top: top, right: right, bottom: bottom, left: left}
}

// MarginAll creates a margin constraint with the same value for all sides
func MarginAll(value int) SpacingConstraint {
	return Margin(value, value, value, value)
}

// Padding creates a padding constraint
func Padding(top, right, bottom, left int) SpacingConstraint {
	return SpacingConstraint{constraintType: ConstraintPadding, top: top, right: right, bottom: bottom, left: left}
}

// PaddingAll creates a padding constraint with the same value for all sides
func PaddingAll(value int) SpacingConstraint {
	return Padding(value, value, value, value)
}

// Above creates a relationship constraint to position above another component
func Above(targetID string, offset int) RelationshipConstraint {
	return RelationshipConstraint{constraintType: ConstraintAbove, targetID: targetID, offset: offset}
}

// Below creates a relationship constraint to position below another component
func Below(targetID string, offset int) RelationshipConstraint {
	return RelationshipConstraint{constraintType: ConstraintBelow, targetID: targetID, offset: offset}
}

// LeftOf creates a relationship constraint to position left of another component
func LeftOf(targetID string, offset int) RelationshipConstraint {
	return RelationshipConstraint{constraintType: ConstraintLeft, targetID: targetID, offset: offset}
}

// RightOf creates a relationship constraint to position right of another component
func RightOf(targetID string, offset int) RelationshipConstraint {
	return RelationshipConstraint{constraintType: ConstraintRight, targetID: targetID, offset: offset}
}

// Constructor functions for size values (CSS-like API)

// Fixed creates a fixed size value
func Fixed(pixels int) SizeValue {
	return FixedSize{pixels: pixels}
}


// Flex creates a flexible size value with weight
func Flex(weight float64) SizeValue {
	return FlexSize{weight: weight}
}

// ConstraintSet represents a collection of constraints for a component
type ConstraintSet struct {
	constraints []Constraint
}

// NewConstraintSet creates a new constraint set
func NewConstraintSet(constraints ...Constraint) ConstraintSet {
	return ConstraintSet{constraints: constraints}
}

// NewConstraintSetPtr creates a new constraint set pointer for chaining
func NewConstraintSetPtr(constraints ...Constraint) *ConstraintSet {
	return &ConstraintSet{constraints: constraints}
}

// Add adds a constraint to the set and returns the set for chaining
func (cs *ConstraintSet) Add(constraintType ConstraintType, constraint Constraint) *ConstraintSet {
	cs.constraints = append(cs.constraints, constraint)
	return cs
}

// AddConstraint adds a constraint to the set (for backwards compatibility)
func (cs *ConstraintSet) AddConstraint(constraint Constraint) {
	cs.constraints = append(cs.constraints, constraint)
}

// Get returns the first constraint of the specified type
func (cs ConstraintSet) Get(constraintType ConstraintType) (Constraint, bool) {
	for _, constraint := range cs.constraints {
		if constraint.Type() == constraintType {
			return constraint, true
		}
	}
	return nil, false
}

// GetAll returns all constraints of the specified type
func (cs ConstraintSet) GetAll(constraintType ConstraintType) []Constraint {
	var result []Constraint
	for _, constraint := range cs.constraints {
		if constraint.Type() == constraintType {
			result = append(result, constraint)
		}
	}
	return result
}

// All returns all constraints
func (cs ConstraintSet) All() []Constraint {
	return cs.constraints
}

// Validate checks if all constraints in the set are valid
func (cs ConstraintSet) Validate() error {
	for _, constraint := range cs.constraints {
		if err := constraint.Validate(); err != nil {
			return fmt.Errorf("constraint validation failed: %w", err)
		}
	}
	return nil
}

// String returns a string representation of all constraints
func (cs ConstraintSet) String() string {
	if len(cs.constraints) == 0 {
		return "no constraints"
	}

	result := "constraints: ["
	for i, constraint := range cs.constraints {
		if i > 0 {
			result += ", "
		}
		result += constraint.String()
	}
	result += "]"
	return result
}

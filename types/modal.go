package types

// Modal represents a renderable modal dialog that can be overlaid on any screen
type Modal interface {
	// RenderModal renders the modal content for the given terminal dimensions
	RenderModal(width, height int) string

	// HandleInput processes keyboard input for this modal
	// Returns (handled, result) where handled indicates if the modal processed the input
	// and result contains any data to return to the caller (nil if modal should stay open)
	HandleInput(key string) (handled bool, result interface{})
}

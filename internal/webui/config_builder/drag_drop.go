package config_builder

// DragDropState holds the state of the drag-and-drop canvas
// This is a minimal scaffold for now

type DragDropState struct {
	Components []string // component names on canvas
	Links      []struct {
		From string
		To   string
	}
}

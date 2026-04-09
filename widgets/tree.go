//go:build windows

package widgets

// Walk traverses a widget tree depth-first. Returning false stops traversal.
func Walk(root Widget, visit func(Widget) bool) bool {
	if root == nil || visit == nil {
		return true
	}
	if !visit(root) {
		return false
	}
	container, ok := root.(Container)
	if !ok {
		return true
	}
	for _, child := range container.Children() {
		if !Walk(child, visit) {
			return false
		}
	}
	return true
}

// FindByID looks up the first widget whose ID matches the provided value.
func FindByID(root Widget, id string) Widget {
	if root == nil || id == "" {
		return nil
	}
	var found Widget
	Walk(root, func(widget Widget) bool {
		if widget.ID() != id {
			return true
		}
		found = widget
		return false
	})
	return found
}

// FindByIDAs looks up a widget by ID and type-asserts it to the requested type.
func FindByIDAs[T Widget](root Widget, id string) (T, bool) {
	var zero T
	widget := FindByID(root, id)
	if widget == nil {
		return zero, false
	}
	typed, ok := widget.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

//go:build windows

package jsonui

import (
	"errors"

	"github.com/AzureIvory/winui/widgets"
)

// WindowHost mounts one JSON UI window into an existing scene and supports hot replacement.
type WindowHost struct {
	scene  *widgets.Scene
	window *Window
}

// MountWindow attaches a JSON UI window to an existing scene.
func MountWindow(scene *widgets.Scene, window *Window) (*WindowHost, error) {
	if scene == nil {
		return nil, errors.New("scene is nil")
	}
	if window == nil {
		return nil, errors.New("window is nil")
	}
	if err := window.Attach(scene); err != nil {
		return nil, err
	}
	return &WindowHost{scene: scene, window: window}, nil
}

// Window returns the currently mounted window.
func (h *WindowHost) Window() *Window {
	if h == nil {
		return nil
	}
	return h.window
}

// ReplaceWindow swaps the mounted window while preserving the existing bound data source when needed.
func (h *WindowHost) ReplaceWindow(next *Window) error {
	if h == nil {
		return errors.New("window host is nil")
	}
	if next == nil {
		return errors.New("replacement window is nil")
	}
	if h.scene == nil {
		return errors.New("window host scene is nil")
	}
	if current := h.window; current != nil {
		if next.Data() == nil {
			next.SetData(current.Data())
		}
		if err := current.Detach(); err != nil {
			return err
		}
	}
	if err := next.Attach(h.scene); err != nil {
		return err
	}
	h.window = next
	return nil
}

// Detach unmounts the current window from the host scene.
func (h *WindowHost) Detach() error {
	if h == nil || h.window == nil {
		return nil
	}
	return h.window.Detach()
}

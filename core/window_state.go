//go:build windows

package core

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procSetWindowLongW      = user32.NewProc("SetWindowLongW")
	procSetWindowLongPtrW   = user32.NewProc("SetWindowLongPtrW")
	procSetLastError        = kernel32.NewProc("SetLastError")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

// callOnUI runs a window operation synchronously on its owning UI thread.
func (a *App) callOnUI(fn func() error) error {
	if a == nil || a.Handle() == 0 {
		return ErrNotInitialized
	}
	invoke := func() error {
		if a.closed.Load() {
			return ErrAppClosed
		}
		return fn()
	}
	if a.IsUIThread() {
		return invoke()
	}
	done := make(chan error, 1)
	if err := a.Post(func() { done <- invoke() }); err != nil {
		return err
	}
	select {
	case err := <-done:
		return err
	case <-a.stopped:
		return ErrAppClosed
	}
}

// ShowNoActivate shows the window in its current position without taking focus.
// The window retains normal activation when the user clicks it.
func (a *App) ShowNoActivate() error {
	return a.callOnUI(func() error {
		procShowWindow.Call(uintptr(a.Handle()), 8) // SW_SHOWNA preserves size and position.
		return nil
	})
}

// ShowAndActivate restores and shows the window, then requests foreground focus.
// Use it for an explicit user action; automatic visibility should use
// ShowNoActivate. Windows may deny foreground activation, in which case the
// window is still shown and an error is returned.
func (a *App) ShowAndActivate() error {
	return a.callOnUI(func() error {
		procShowWindow.Call(uintptr(a.Handle()), 9) // SW_RESTORE
		ok, _, _ := procSetForegroundWindow.Call(uintptr(a.Handle()))
		if ok == 0 {
			return fmt.Errorf("SetForegroundWindow declined activation")
		}
		return nil
	})
}

// Hide hides the window without destroying its scene or native controls.
func (a *App) Hide() error {
	return a.callOnUI(func() error {
		procShowWindow.Call(uintptr(a.Handle()), 0) // SW_HIDE
		return nil
	})
}

// RaiseNoActivate raises the window to the top of its existing Z-order group
// without changing visibility, size or keyboard focus. It does not make a
// normal window topmost and does not create an owner lifetime relationship.
func (a *App) RaiseNoActivate() error {
	return a.callOnUI(func() error {
		ok, _, err := procSetWindowPos.Call(uintptr(a.Handle()), 0, 0, 0, 0, 0,
			0x0001|0x0002|swpNoActivate) // SWP_NOSIZE | SWP_NOMOVE
		if ok == 0 {
			return wrapError("SetWindowPos", err)
		}
		return nil
	})
}

// SetTopmost changes the window's topmost Z-order group without changing its
// visibility, bounds, owner or keyboard focus. Win32 ownership rules also apply
// to related top-level windows. Reapply the owner's current state after attaching
// a new owned window; SetOwner alone does not apply that Z-order state.
func (a *App) SetTopmost(enabled bool) error {
	return a.callOnUI(func() error {
		insertAfter := ^uintptr(1) // HWND_NOTOPMOST (-2)
		if enabled {
			insertAfter = ^uintptr(0) // HWND_TOPMOST (-1)
		}
		ok, _, err := procSetWindowPos.Call(
			uintptr(a.Handle()), insertAfter, 0, 0, 0, 0,
			0x0001|0x0002|swpNoActivate, // SWP_NOSIZE | SWP_NOMOVE
		)
		if ok == 0 {
			return wrapError("SetWindowPos", err)
		}
		return nil
	})
}

// SetScreenBounds moves and resizes the outer window in physical screen pixels.
// It preserves visibility, focus and Z order; Width/Height in Options remain DP.
func (a *App) SetScreenBounds(bounds Rect) error {
	if bounds.Empty() {
		return fmt.Errorf("window bounds must have positive width and height")
	}
	return a.callOnUI(func() error {
		a.placingScreenBounds = true
		defer func() { a.placingScreenBounds = false }()
		ok, _, err := procSetWindowPos.Call(uintptr(a.Handle()), 0,
			uintptr(bounds.X), uintptr(bounds.Y), uintptr(bounds.W), uintptr(bounds.H),
			swpNoZOrder|swpNoActivate)
		if ok == 0 {
			return wrapError("SetWindowPos", err)
		}
		return nil
	})
}

// ScreenBounds returns the outer window rectangle in physical screen pixels.
func (a *App) ScreenBounds() (Rect, error) {
	var bounds Rect
	err := a.callOnUI(func() error {
		var rect winRect
		ok, _, err := procGetWindowRect.Call(uintptr(a.Handle()), uintptr(unsafe.Pointer(&rect)))
		if ok == 0 {
			return wrapError("GetWindowRect", err)
		}
		bounds = rectFromWinRect(rect)
		return nil
	})
	return bounds, err
}

// SetOwner makes this an owned window without activating either window.
// Zero removes the owner. It does not refresh the owner's topmost Z-order group;
// reapply SetTopmost on the owner after attaching a new owned window.
// Windows hides an owned window when its owner is minimized, and destroys it
// when its owner is destroyed. The owner must be another top-level window.
func (a *App) SetOwner(owner windows.Handle) error {
	return a.callOnUI(func() error {
		if owner == a.Handle() {
			return fmt.Errorf("a window cannot own itself")
		}
		proc := procSetWindowLongPtrW
		if unsafe.Sizeof(uintptr(0)) == 4 {
			proc = procSetWindowLongW
		}
		procSetLastError.Call(0)
		previous, _, err := proc.Call(uintptr(a.Handle()), ^uintptr(7), uintptr(owner)) // GWLP_HWNDPARENT
		if previous == 0 && err != windows.ERROR_SUCCESS {
			return wrapError("SetWindowLongPtrW", err)
		}
		return nil
	})
}

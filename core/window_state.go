//go:build windows

package core

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procSetWindowLongW    = user32.NewProc("SetWindowLongW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procSetLastError      = kernel32.NewProc("SetLastError")
)

// callOnUI runs a window operation synchronously on its owning UI thread.
func (a *App) callOnUI(fn func() error) error {
	if a == nil || a.hwnd == 0 {
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
		procShowWindow.Call(uintptr(a.hwnd), 8) // SW_SHOWNA preserves size and position.
		return nil
	})
}

// Hide hides the window without destroying its scene or native controls.
func (a *App) Hide() error {
	return a.callOnUI(func() error {
		procShowWindow.Call(uintptr(a.hwnd), 0) // SW_HIDE
		return nil
	})
}

// RaiseNoActivate raises the window to the top of its existing Z-order group
// without changing visibility, size or keyboard focus. It does not make a
// normal window topmost and does not create an owner lifetime relationship.
func (a *App) RaiseNoActivate() error {
	return a.callOnUI(func() error {
		ok, _, err := procSetWindowPos.Call(uintptr(a.hwnd), 0, 0, 0, 0, 0,
			0x0001|0x0002|swpNoActivate) // SWP_NOSIZE | SWP_NOMOVE
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
		ok, _, err := procSetWindowPos.Call(uintptr(a.hwnd), 0,
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
		ok, _, err := procGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&rect)))
		if ok == 0 {
			return wrapError("GetWindowRect", err)
		}
		bounds = rectFromWinRect(rect)
		return nil
	})
	return bounds, err
}

// SetOwner makes this a normal owned window above owner, without making it
// globally topmost or activating either window. Zero removes the owner.
// Windows hides an owned window when its owner is minimized, and destroys it
// when its owner is destroyed. The owner must be another top-level window.
func (a *App) SetOwner(owner windows.Handle) error {
	return a.callOnUI(func() error {
		if owner == a.hwnd {
			return fmt.Errorf("a window cannot own itself")
		}
		proc := procSetWindowLongPtrW
		if unsafe.Sizeof(uintptr(0)) == 4 {
			proc = procSetWindowLongW
		}
		procSetLastError.Call(0)
		previous, _, err := proc.Call(uintptr(a.hwnd), ^uintptr(7), uintptr(owner)) // GWLP_HWNDPARENT
		if previous == 0 && err != windows.ERROR_SUCCESS {
			return wrapError("SetWindowLongPtrW", err)
		}
		return nil
	})
}

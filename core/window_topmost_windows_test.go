//go:build windows

package core

import (
	"errors"
	"testing"
)

func newHiddenTopmostTestApp(t *testing.T, name string, owner *App) *App {
	t.Helper()
	app, err := NewApp(Options{
		Title: "WinUI isolated topmost test: " + name, StartHidden: true,
		Width: 320, Height: 200,
		OnCreate: func(created *App) error {
			if owner != nil {
				return created.SetOwner(owner.Handle())
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	return app
}

type topmostWindowSnapshot struct {
	topmost bool
	visible bool
	owner   uintptr
	focus   uintptr
	bounds  Rect
}

func snapshotTopmostWindow(t *testing.T, app *App) topmostWindowSnapshot {
	t.Helper()
	var snapshot topmostWindowSnapshot
	// Observe each HWND on its own UI thread after queued position changes,
	// rather than racing cross-thread propagation from the owner's operation.
	if err := app.callOnUI(func() error {
		style, _, _ := user32.NewProc("GetWindowLongW").Call(uintptr(app.Handle()), ^uintptr(19)) // GWL_EXSTYLE
		visible, _, _ := user32.NewProc("IsWindowVisible").Call(uintptr(app.Handle()))
		owner, _, _ := user32.NewProc("GetWindow").Call(uintptr(app.Handle()), 4) // GW_OWNER
		focus, _, _ := user32.NewProc("GetFocus").Call()
		bounds, err := app.ScreenBounds()
		if err != nil {
			return err
		}
		snapshot = topmostWindowSnapshot{
			topmost: style&0x00000008 != 0, visible: visible != 0, // WS_EX_TOPMOST
			owner: owner, focus: focus, bounds: bounds,
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestSetTopmostHiddenOwnedWindowsFollowOwner(t *testing.T) {
	foreground, _, _ := user32.NewProc("GetForegroundWindow").Call()
	owner := newHiddenTopmostTestApp(t, "owner", nil)
	existing := newHiddenTopmostTestApp(t, "existing-owned", owner)
	attachLater := newHiddenTopmostTestApp(t, "attach-later", nil)
	unrelated := newHiddenTopmostTestApp(t, "unrelated", nil)
	apps := map[string]*App{
		"owner": owner, "existing-owned": existing, "attach-later": attachLater, "unrelated": unrelated,
	}
	initial := make(map[*App]topmostWindowSnapshot)
	for name, app := range apps {
		if err := app.SetScreenBounds(Rect{X: 40, Y: 60, W: 360, H: 220}); err != nil {
			t.Fatal(err)
		}
		initial[app] = snapshotTopmostWindow(t, app)
		if initial[app].topmost || initial[app].visible {
			t.Fatalf("%s did not begin hidden/non-topmost: %+v", name, initial[app])
		}
	}
	assertState := func(stage string, wantTopmost map[*App]bool) {
		t.Helper()
		for name, app := range apps {
			got := snapshotTopmostWindow(t, app)
			want := initial[app]
			want.topmost = wantTopmost[app]
			if got != want {
				t.Errorf("%s %s state=%+v want=%+v", stage, name, got, want)
			}
			t.Logf("%s %s: topmost=%v hidden=%v owner=%x", stage, name, got.topmost, !got.visible, got.owner)
		}
		after, _, _ := user32.NewProc("GetForegroundWindow").Call()
		if after != foreground {
			for name, app := range apps {
				if after == uintptr(app.Handle()) {
					t.Errorf("%s activated hidden fixture %s", stage, name)
				}
			}
		}
	}
	if err := owner.SetTopmost(true); err != nil {
		t.Fatal(err)
	}
	assertState("pin-existing", map[*App]bool{owner: true, existing: true})
	if err := attachLater.SetOwner(owner.Handle()); err != nil {
		t.Fatal(err)
	}
	initialState := initial[attachLater]
	initialState.owner = uintptr(owner.Handle())
	initial[attachLater] = initialState
	assertState("attach-after-pin-needs-projection", map[*App]bool{owner: true, existing: true})
	createdLater := newHiddenTopmostTestApp(t, "create-after-pin", owner)
	apps["create-after-pin"] = createdLater
	initial[createdLater] = snapshotTopmostWindow(t, createdLater)
	if initial[createdLater].visible {
		t.Fatal("creating an owned window after pin made it visible")
	}
	assertState("create-after-pin-needs-projection", map[*App]bool{owner: true, existing: true})
	for _, enabled := range []bool{true, false, false, true, false} {
		// Only the owner changes topmost. Independently demoting an owned
		// popup would also demote its owner under Win32 ownership rules.
		if err := owner.SetTopmost(enabled); err != nil {
			t.Fatal(err)
		}
		assertState("owner-toggle", map[*App]bool{
			owner: enabled, existing: enabled, attachLater: enabled, createdLater: enabled,
		})
	}
	if initial[existing].owner != uintptr(owner.Handle()) ||
		initial[createdLater].owner != uintptr(owner.Handle()) {
		t.Fatal("fixture ownership was not established")
	}
}

func TestSetTopmostRunsInlineOnUIThread(t *testing.T) {
	app := newHiddenTopmostTestApp(t, "inline", nil)
	if err := app.callOnUI(func() error {
		if err := app.SetTopmost(true); err != nil {
			return err
		}
		return app.SetTopmost(false)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestWindowActivationAndTopmostRejectClosedApp(t *testing.T) {
	app, err := NewApp(Options{StartHidden: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	app.Close()
	app.Run()
	for name, operation := range map[string]func() error{
		"activate": app.ShowAndActivate,
		"pin":      func() error { return app.SetTopmost(true) },
		"unpin":    func() error { return app.SetTopmost(false) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := operation(); !errors.Is(err, ErrAppClosed) {
				t.Fatalf("closed operation returned %v; want ErrAppClosed", err)
			}
		})
	}
}

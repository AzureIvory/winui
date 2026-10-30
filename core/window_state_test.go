//go:build windows

package core

import (
	"errors"
	"testing"
	"time"
)

func TestWindowPlacementPreservesVisibilityAndForeground(t *testing.T) {
	foreground := user32.NewProc("GetForegroundWindow")
	visible := user32.NewProc("IsWindowVisible")
	before, _, _ := foreground.Call()
	app, err := NewApp(Options{Title: "WinUI isolated placement test", StartHidden: true, Width: 320, Height: 220})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	isVisible := func() bool { v, _, _ := visible.Call(uintptr(app.Handle())); return v != 0 }
	assertForeground := func() {
		t.Helper()
		after, _, _ := foreground.Call()
		// The user's desktop may legitimately change while this isolated test
		// runs. The tested window must never become foreground as a result.
		if before != after && after == uintptr(app.Handle()) {
			t.Fatalf("nonactivating operation activated test window %x (previous foreground %x)", after, before)
		}
	}
	if isVisible() {
		t.Fatal("StartHidden created a visible window")
	}
	assertForeground()
	want := Rect{X: -370, Y: 25, W: 480, H: 250}
	if err := app.SetScreenBounds(want); err != nil {
		t.Fatal(err)
	}
	if got, err := app.ScreenBounds(); err != nil || got != want {
		t.Fatalf("screen bounds = %+v, %v; want %+v", got, err, want)
	}
	if isVisible() {
		t.Fatal("moving a hidden window showed it")
	}
	if err := app.RaiseNoActivate(); err != nil {
		t.Fatal(err)
	}
	if isVisible() {
		t.Fatal("raising a hidden window showed it")
	}
	if err := app.ShowNoActivate(); err != nil {
		t.Fatal(err)
	}
	if !isVisible() {
		t.Fatal("ShowNoActivate did not show window")
	}
	assertForeground()
	if got, err := app.ScreenBounds(); err != nil || got != want {
		t.Fatalf("show changed bounds = %+v, %v", got, err)
	}
	if err := app.Hide(); err != nil {
		t.Fatal(err)
	}
	if isVisible() {
		t.Fatal("Hide left window visible")
	}
	assertForeground()
	if err := app.SetScreenBounds(Rect{W: -1, H: 10}); err == nil {
		t.Fatal("negative width accepted")
	}
}

func TestQueuedWindowOperationReturnsWhenMessageLoopStops(t *testing.T) {
	app, err := NewApp(Options{Title: "WinUI isolated shutdown test", StartHidden: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	if err := app.Post(func() {
		close(entered)
		<-release
		procDestroyWindow.Call(uintptr(app.Handle()))
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("UI callback did not start")
	}
	// Post succeeds while the UI callback is blocked. Destroying its window
	// before returning invalidates the queued HWND invocation, then WM_QUIT
	// ends the loop. Do not post an extra quit message: it could survive on a
	// reusable Go OS thread and terminate the next test window's message loop.
	result := make(chan error, 1)
	go func() { result <- app.Hide() }()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	for {
		app.postMu.Lock()
		queued := len(app.postQueue) > 0
		app.postMu.Unlock()
		if queued {
			break
		}
		select {
		case <-deadline.C:
			close(release)
			t.Fatal("window operation did not queue")
		case <-time.After(time.Millisecond):
		}
	}
	close(release)
	select {
	case err := <-result:
		if !errors.Is(err, ErrAppClosed) {
			t.Fatalf("shutdown operation returned %v; want ErrAppClosed", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("synchronous window operation hung after message loop stopped")
	}
	select {
	case <-app.stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("UI thread did not signal shutdown")
	}
	app.Run()
}

func TestWindowOperationsRejectUninitializedApp(t *testing.T) {
	var app *App
	for name, operation := range map[string]func() error{
		"show": app.ShowNoActivate, "activate": app.ShowAndActivate, "hide": app.Hide, "raise": app.RaiseNoActivate,
		"pin": func() error { return app.SetTopmost(true) }, "unpin": func() error { return app.SetTopmost(false) },
	} {
		t.Run(name, func(t *testing.T) {
			if !errors.Is(operation(), ErrNotInitialized) {
				t.Fatal("expected ErrNotInitialized")
			}
		})
	}
}

func TestKeyPreviewFiltersMessagesAndRecognizesProcessKey(t *testing.T) {
	calls := 0
	var received KeyEvent
	processKeyComposing := false
	app, err := NewApp(Options{StartHidden: true, OnKeyPreview: func(_ *App, event KeyEvent) bool {
		calls++
		received = event
		if event.Key == 0xE5 {
			processKeyComposing = event.Composing
		}
		return true
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	err = app.callOnUI(func() error {
		ordinary := keyEvent(app.Handle(), uintptr(KeyReturn), 0)
		if ordinary.Key != KeyReturn || ordinary.Composing || ordinary.Handle != app.Handle() {
			t.Errorf("ordinary Return/source identity lost: %+v", ordinary)
		}
		for _, message := range []msg{
			{HWnd: app.Handle(), Message: 0x0102, WParam: 'K'},
			{Message: wmKeyDown, WParam: 'K'},
			{HWnd: app.Handle(), Message: wmKeyDown, WParam: 0xE5},
		} {
			if app.previewKeyMessage(&message) {
				t.Error("unrelated/character/IME process message was consumed")
			}
		}
		message := msg{HWnd: app.Handle(), Message: wmKeyDown, WParam: 'K', LParam: 1 << 30}
		if !app.previewKeyMessage(&message) {
			t.Error("preview did not consume window key")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || !processKeyComposing || received.Key != 'K' || !received.Repeat {
		t.Fatalf("preview calls=%d event=%+v", calls, received)
	}
}

func TestIMEProcessKeyNormalizationKeepsOrdinaryReturnUsable(t *testing.T) {
	for _, test := range []struct {
		name                string
		key, original, want uint32
	}{
		{"IME wraps ordinary Return", 0xE5, KeyReturn, KeyReturn},
		{"IME wraps shortcut", 0xE5, 'K', 'K'},
		{"unresolved zero stays process key", 0xE5, 0, 0xE5},
		{"unresolved process key stays process key", 0xE5, 0xE5, 0xE5},
		{"ordinary Return is unchanged", KeyReturn, 'K', KeyReturn},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeIMEVirtualKey(test.key, test.original); got != test.want {
				t.Fatalf("normalized key=%x want=%x", got, test.want)
			}
		})
	}
}

func TestFailedNativePostNeverRetainsItsCallback(t *testing.T) {
	app, err := NewApp(Options{StartHidden: true})
	if err != nil {
		t.Fatal(err)
	}
	// Use a verified nonexistent ordinary HWND, excluding all special values.
	// No native window is created or valid application's queue is targeted.
	app.hwnd.Store(0x7ffe7ffe)
	if valid, _, _ := user32.NewProc("IsWindow").Call(uintptr(app.Handle())); valid != 0 {
		t.Fatal("test HWND unexpectedly exists")
	}
	prior, failed := 0, 0
	app.postQueue = append(app.postQueue, func() { prior++ })
	if err := app.Post(func() { failed++ }); err == nil {
		t.Fatal("invalid HWND unexpectedly accepted PostMessageW")
	}
	app.drainPosted()
	if prior != 1 || failed != 0 {
		t.Fatalf("Post error retained callback or removed earlier work: prior=%d failed=%d", prior, failed)
	}
}

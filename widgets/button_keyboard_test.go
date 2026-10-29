//go:build windows

package widgets

import (
	"testing"
	"time"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func TestCustomButtonParticipatesInFocusTraversalAndKeyboardActivation(t *testing.T) {
	scene := NewScene(nil)
	input := NewEditBox("input", ModeCustom)
	button := NewButton("save", "保存", ModeCustom)
	clicks := 0
	button.SetOnClick(func() { clicks++ })
	scene.Root().AddAll(input, button)
	if !scene.FocusNext(false) || scene.Focus() != input {
		t.Fatal("initial input focus failed")
	}
	if !scene.FocusNext(false) || scene.Focus() != button {
		t.Fatal("custom save button is unreachable through focus traversal")
	}
	if !button.Focused {
		t.Fatal("keyboard-focused button did not expose focused state for painting")
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn})
	scene.DispatchChar('\r')
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Repeat: true})
	if clicks != 1 {
		t.Fatalf("Enter/transformed character/repeat activated %d times; want1", clicks)
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeySpace})
	scene.DispatchChar(' ')
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeySpace, Repeat: true})
	if clicks != 2 {
		t.Fatalf("Space/transformed character/repeat activated %d times total; want2", clicks)
	}
	for _, key := range []core.KeyEvent{{Key: core.KeyReturn, Alt: true}, {Key: core.KeyReturn, Ctrl: true}, {Key: core.KeyReturn, Shift: true}, {Key: core.KeyReturn, Flags: 1 << 29}, {Key: core.KeySpace, Alt: true}, {Key: core.KeySpace, Ctrl: true}, {Key: core.KeySpace, Shift: true}, {Key: core.KeyReturn, Composing: true}} {
		scene.DispatchKeyDown(key)
	}
	if clicks != 2 {
		t.Fatal("modified or composing key activated custom button")
	}
	button.SetEnabled(false)
	if button.Focused {
		t.Fatal("disabled button retained visible focus state")
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn})
	if clicks != 2 {
		t.Fatal("disabled focused button activated")
	}
	scene.FocusWidget(input)
	if !scene.FocusNext(false) || scene.Focus() != input {
		t.Fatal("traversal did not skip disabled button")
	}
	button.SetEnabled(true)
	scene.FocusWidget(button)
	if !button.Focused {
		t.Fatal("re-enabled button did not regain keyboard focus")
	}
	scene.FocusWidget(input)
	if button.Focused {
		t.Fatal("blurred button retained focus indicator state")
	}
	button.SetVisible(false)
	if !scene.FocusNext(false) || scene.Focus() != input {
		t.Fatal("traversal did not skip hidden button")
	}
}

func TestCustomButtonNativePretranslationDoesNotLeakAfterFocusChange(t *testing.T) {
	var scene *Scene
	var button *Button
	var editor *EditBox
	clicks := 0
	opts := core.Options{Title: "WinUI isolated button keyboard test", StartHidden: true}
	BindScene(&opts, SceneHooks{OnCreate: func(_ *core.App, created *Scene) error {
		scene = created
		button = NewButton("save", "保存", ModeCustom)
		editor = NewEditBox("editor", ModeNative)
		button.SetBounds(Rect{X: 10, Y: 10, W: 100, H: 40})
		editor.SetBounds(Rect{X: 10, Y: 60, W: 250, H: 40})
		button.SetOnClick(func() { clicks++; scene.FocusWidget(editor) })
		scene.Root().AddAll(button, editor)
		scene.FocusWidget(button)
		return nil
	}})
	app, err := core.NewApp(opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	onUI := func(fn func()) {
		t.Helper()
		done := make(chan struct{})
		if err := app.Post(func() { fn(); close(done) }); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("button pretranslation callback timed out")
		}
	}
	post := windows.NewLazySystemDLL("user32.dll").NewProc("PostMessageW")
	for index, key := range []uint32{core.KeySpace, core.KeyReturn} {
		onUI(func() { scene.FocusWidget(button) })
		if ok, _, err := post.Call(uintptr(app.Handle()), 0x0100, uintptr(key), 1); ok == 0 {
			t.Fatal(err)
		}
		onUI(func() {})
		onUI(func() {
			if clicks != index+1 || scene.Focus() != editor || editor.TextValue() != "" {
				t.Errorf("button activation leaked translated input: clicks=%d text=%q", clicks, editor.TextValue())
			}
		})
	}
}

func TestMouseClickCustomButtonTransfersRealFocusFromNativeEditor(t *testing.T) {
	var scene *Scene
	var button *Button
	var editor *EditBox
	clicks := 0
	opts := core.Options{Title: "WinUI isolated mouse-focus regression", StartHidden: true}
	BindScene(&opts, SceneHooks{OnCreate: func(_ *core.App, created *Scene) error {
		scene = created
		button = NewButton("save", "保存", ModeCustom)
		editor = NewEditBox("native", ModeNative)
		button.SetBounds(Rect{X: 10, Y: 10, W: 100, H: 40})
		editor.SetBounds(Rect{X: 10, Y: 60, W: 250, H: 40})
		button.SetOnClick(func() { clicks++ })
		scene.Root().AddAll(button, editor)
		return nil
	}})
	app, err := core.NewApp(opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	onUI := func(fn func()) {
		t.Helper()
		done := make(chan struct{})
		if err := app.Post(func() { fn(); close(done) }); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("mouse-focus callback timed out")
		}
	}
	getFocus := windows.NewLazySystemDLL("user32.dll").NewProc("GetFocus")
	post := windows.NewLazySystemDLL("user32.dll").NewProc("PostMessageW")
	var actualFocus uintptr
	onUI(func() {
		scene.FocusWidget(editor)
		before, _, _ := getFocus.Call()
		if before != uintptr(scene.nativeHandleFor(editor)) {
			t.Error("native editor did not own actual initial focus")
		}
		mouse := core.MouseEvent{Point: core.Point{X: 20, Y: 20}, Button: core.MouseButtonLeft}
		scene.DispatchMouseDown(mouse)
		scene.DispatchMouseUp(mouse)
		actualFocus, _, _ = getFocus.Call()
		if scene.Focus() != button || !button.Focused || actualFocus != uintptr(app.Handle()) {
			t.Errorf("mouse focus diverged: logicalButton=%v focusedState=%v actualHWND=%x wantHost=%x", scene.Focus() == button, button.Focused, actualFocus, app.Handle())
		}
	})
	for index, key := range []uint32{core.KeySpace, core.KeyReturn} {
		if ok, _, err := post.Call(actualFocus, 0x0100, uintptr(key), 1); ok == 0 {
			t.Fatal(err)
		}
		onUI(func() {})
		onUI(func() {
			if clicks != index+2 || editor.TextValue() != "" {
				t.Errorf("key followed stale native focus: clicks=%d text=%q", clicks, editor.TextValue())
			}
		})
	}
}

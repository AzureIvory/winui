//go:build windows

package widgets

import (
	"testing"
	"time"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func TestFocusWidgetHonorsSceneAndAncestors(t *testing.T) {
	scene := NewScene(nil)
	parent := NewPanel("container")
	edit := NewEditBox("input", ModeCustom)
	parent.Add(edit)
	scene.Root().Add(parent)
	scene.FocusWidget(edit)
	if scene.Focus() != edit {
		t.Fatal("attached editor did not receive focus")
	}
	scene.FocusWidget(nil)
	parent.SetVisible(false)
	scene.FocusWidget(edit)
	if scene.Focus() != nil {
		t.Fatal("hidden descendant received focus")
	}
	parent.SetVisible(true)
	parent.SetEnabled(false)
	scene.FocusWidget(edit)
	if scene.Focus() != nil {
		t.Fatal("disabled descendant received focus")
	}
	other := NewScene(nil)
	foreign := NewEditBox("foreign", ModeCustom)
	other.Root().Add(foreign)
	scene.FocusWidget(foreign)
	if scene.Focus() != nil {
		t.Fatal("foreign scene editor received focus")
	}
}

func TestMultilineEnterAndComposition(t *testing.T) {
	scene := NewScene(nil)
	edit := NewEditBox("draft", ModeCustom)
	edit.SetMultiline(true)
	edit.SetAcceptReturn(true)
	edit.SetText("售后")
	submits := 0
	edit.SetOnSubmit(func(string) { submits++ })
	scene.Root().Add(edit)
	scene.FocusWidget(edit)
	previews := 0
	scene.SetKeyPreview(func(core.KeyEvent) bool { previews++; return false })
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Composing: true})
	if edit.TextValue() != "售后" || submits != 0 || previews != 0 {
		t.Fatal("IME candidate confirmation changed draft or triggered shortcut")
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn})
	if edit.TextValue() != "售后\n" || submits != 0 {
		t.Fatalf("plain Enter: text=%q submits=%d", edit.TextValue(), submits)
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Ctrl: true})
	if edit.TextValue() != "售后\n" || submits != 1 {
		t.Fatalf("Ctrl+Enter: text=%q submits=%d", edit.TextValue(), submits)
	}
}

func TestNativeDefaultSubmitIgnoresAltEnter(t *testing.T) {
	var scene *Scene
	var edit *EditBox
	submits := 0
	opts := core.Options{Title: "WinUI isolated Alt+Enter test", StartHidden: true}
	BindScene(&opts, SceneHooks{OnCreate: func(_ *core.App, created *Scene) error {
		scene = created
		edit = NewEditBox("input", ModeNative)
		edit.SetOnSubmit(func(string) { submits++ })
		scene.Root().Add(edit)
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
	done := make(chan struct{})
	if err := app.Post(func() {
		defer close(done)
		handle := scene.nativeHandleFor(edit)
		for _, key := range []core.KeyEvent{{Handle: handle, Key: core.KeyReturn, Alt: true}, {Handle: handle, Key: core.KeyReturn, Alt: true, Ctrl: true}, {Handle: handle, Key: core.KeyReturn, Flags: 1 << 29}} {
			if opts.OnKeyPreview(app, key) {
				t.Errorf("Alt+Enter incorrectly consumed as native submit: %+v", key)
			}
		}
		if submits != 0 {
			t.Errorf("Alt+Enter invoked default submit %d times", submits)
		}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Alt+Enter test callback timed out")
	}
}

func TestNativeChildPreviewAndIMEEnter(t *testing.T) {
	var scene *Scene
	var edit *EditBox
	previews, submits := 0, 0
	opts := core.Options{Title: "WinUI isolated native keyboard test", StartHidden: true, Width: 320, Height: 220}
	BindScene(&opts, SceneHooks{
		OnCreate: func(_ *core.App, created *Scene) error {
			scene = created
			edit = NewEditBox("native", ModeNative)
			edit.SetMultiline(true)
			edit.SetAcceptReturn(false)
			edit.SetBounds(Rect{X: 10, Y: 10, W: 250, H: 40})
			edit.SetOnSubmit(func(string) { submits++ })
			scene.Root().Add(edit)
			scene.FocusWidget(edit)
			return nil
		},
		OnKeyPreview: func(_ *core.App, _ *Scene, key core.KeyEvent) bool { previews++; return key.Key == 'K' },
	})
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
			t.Fatal("UI callback timed out")
		}
	}
	var handle windows.Handle
	onUI(func() { handle = scene.nativeHandleFor(edit) })
	if handle == 0 {
		t.Fatal("native EDIT child missing")
	}
	post := windows.NewLazySystemDLL("user32.dll").NewProc("PostMessageW")
	send := windows.NewLazySystemDLL("user32.dll").NewProc("SendMessageW")
	postKey := func(key uintptr) {
		t.Helper()
		if ok, _, err := post.Call(uintptr(handle), uintptr(nativeWindowKeyDown), key, 1); ok == 0 {
			t.Fatal(err)
		}
	}
	postKey('K')
	onUI(func() {
		if previews != 1 || edit.TextValue() != "" {
			t.Errorf("consumed native shortcut: previews=%d text=%q", previews, edit.TextValue())
		}
	})
	// Synthetic native composition messages exercise the widget's composition
	// lifecycle guard without depending on the desktop user's installed IME.
	send.Call(uintptr(handle), 0x010D, 0, 0)
	postKey(uintptr(core.KeyReturn))
	onUI(func() {
		if submits != 0 || previews != 1 {
			t.Errorf("composition Enter submitted=%d previews=%d", submits, previews)
		}
	})
	// Translation can end the composition before dispatching this same Return.
	// Reproduce that order through the bound preview and real native subclass.
	onUI(func() {
		if opts.OnKeyPreview(app, core.KeyEvent{Handle: handle, Key: core.KeyReturn, Composing: true}) {
			t.Error("composition preview consumed input")
		}
		send.Call(uintptr(handle), 0x010E, 0, 0)
		send.Call(uintptr(handle), uintptr(nativeWindowKeyDown), uintptr(core.KeyReturn), 1)
		if submits != 0 || previews != 1 {
			t.Errorf("composition ending before dispatch submitted=%d previews=%d", submits, previews)
		}
	})
	send.Call(uintptr(handle), 0x010E, 0, 0)
	postKey(uintptr(core.KeyReturn))
	onUI(func() {
		if submits != 1 || previews != 2 {
			t.Errorf("normal Enter submitted=%d previews=%d", submits, previews)
		}
	})
	onUI(func() {
		if !opts.OnKeyPreview(app, core.KeyEvent{Handle: handle, Key: core.KeyReturn}) {
			t.Error("native default Return was not consumed before translation")
		}
		edit.SetAcceptReturn(true)
		handle = scene.nativeHandleFor(edit)
		if !opts.OnKeyPreview(app, core.KeyEvent{Handle: handle, Key: core.KeyReturn, Ctrl: true}) {
			t.Error("native Ctrl+Enter was not consumed before translation")
		}
		if opts.OnKeyPreview(app, core.KeyEvent{Handle: handle, Key: core.KeyReturn}) {
			t.Error("multiline ordinary Enter was consumed instead of reaching native editor")
		}
		if submits != 3 {
			t.Errorf("native pretranslation submit count=%d want=3", submits)
		}
		edit.SetAcceptReturn(false)
		handle = scene.nativeHandleFor(edit)
	})
	onUI(func() { edit.SetText("积压键盘消息时保留正文") })
	entered, release := make(chan struct{}), make(chan struct{})
	if err := app.Post(func() { close(entered); <-release }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("keyboard backlog barrier timed out")
	}
	postKey(uintptr(core.KeyReturn))
	postKey(uintptr(core.KeyReturn))
	close(release)
	onUI(func() {})
	// TranslateMessage enqueues characters behind already queued keydowns.
	// A second barrier ensures both translated WM_CHAR messages have run.
	onUI(func() {
		if submits != 5 || edit.TextValue() != "积压键盘消息时保留正文" {
			t.Errorf("queued Return submit/character pairing: submits=%d text=%q", submits, edit.TextValue())
		}
	})
	// Reproduce the reviewed backlog: Ctrl+Enter followed by Right in a
	// WANTRETURN native editor, before translated characters are dispatched.
	getState := windows.NewLazySystemDLL("user32.dll").NewProc("GetKeyboardState")
	setState := windows.NewLazySystemDLL("user32.dll").NewProc("SetKeyboardState")
	var originalState [256]byte
	onUI(func() {
		edit.SetAcceptReturn(true)
		handle = scene.nativeHandleFor(edit)
		edit.SetText("Ctrl Enter 与导航积压")
		getState.Call(uintptr(unsafe.Pointer(&originalState[0])))
		ctrlState := originalState
		ctrlState[0x11] |= 0x80
		setState.Call(uintptr(unsafe.Pointer(&ctrlState[0])))
	})
	entered, release = make(chan struct{}), make(chan struct{})
	if err := app.Post(func() { close(entered); <-release }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("Ctrl+Enter backlog barrier timed out")
	}
	postKey(uintptr(core.KeyReturn))
	postKey(uintptr(core.KeyRight))
	close(release)
	onUI(func() {})
	onUI(func() {
		setState.Call(uintptr(unsafe.Pointer(&originalState[0])))
		if submits != 6 || edit.TextValue() != "Ctrl Enter 与导航积压" {
			t.Errorf("Ctrl+Enter/navigation inserted translated newline: submits=%d text=%q", submits, edit.TextValue())
		}
	})
}

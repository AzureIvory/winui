//go:build windows

package widgets

import (
	"testing"
	"time"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func TestFocusNextDepthFirstWrapAndEligibility(t *testing.T) {
	scene := NewScene(nil)
	first := NewEditBox("first", ModeCustom)
	group := NewPanel("group")
	second := NewEditBox("second", ModeCustom)
	hidden := NewPanel("hidden")
	hidden.Add(NewEditBox("hidden-input", ModeCustom))
	hidden.SetVisible(false)
	disabled := NewPanel("disabled")
	disabled.Add(NewEditBox("disabled-input", ModeCustom))
	disabled.SetEnabled(false)
	group.AddAll(NewLabel("not-focusable", "label"), second, hidden, disabled)
	last := NewEditBox("last", ModeCustom)
	scene.Root().AddAll(first, group, last)
	foreign := NewEditBox("foreign", ModeCustom)
	scene.Root().Add(foreign)
	foreign.setScene(NewScene(nil))
	for _, want := range []Widget{first, second, last, first} {
		if !scene.FocusNext(false) || scene.Focus() != want {
			t.Fatalf("forward traversal focused %v want %s", scene.Focus(), want.ID())
		}
	}
	for _, want := range []Widget{last, second, first} {
		if !scene.FocusNext(true) || scene.Focus() != want {
			t.Fatalf("reverse traversal want %s", want.ID())
		}
	}
	scene.FocusWidget(nil)
	if !scene.FocusNext(true) || scene.Focus() != last {
		t.Fatal("initial reverse focus did not start at last eligible widget")
	}
	detached := NewEditBox("detached", ModeCustom)
	scene.FocusWidget(detached)
	if scene.Focus() != last {
		t.Fatal("detached target accepted by shared focus qualification")
	}
	first.SetEnabled(false)
	group.SetVisible(false)
	last.SetVisible(false)
	if scene.FocusNext(false) {
		t.Fatal("traversal accepted foreign/hidden/disabled candidates")
	}
	var nilScene *Scene
	if nilScene.FocusNext(false) {
		t.Fatal("nil scene reported focus traversal")
	}
}

func TestFocusNextStaysWithinTopmostNestedModal(t *testing.T) {
	scene := NewScene(nil)
	outside := NewEditBox("outside", ModeCustom)
	lower := NewModal("lower")
	a, b := NewEditBox("a", ModeCustom), NewEditBox("b", ModeCustom)
	lower.AddAll(a, b)
	scene.Root().AddAll(outside, lower)
	scene.FocusWidget(outside)
	if !scene.FocusNext(false) || scene.Focus() != a {
		t.Fatal("modal traversal escaped to background")
	}
	if !scene.FocusNext(true) || scene.Focus() != b {
		t.Fatal("reverse modal wrap escaped modal")
	}
	nested := NewModal("nested")
	nestedInput := NewEditBox("nested-input", ModeCustom)
	nested.Add(nestedInput)
	lower.Add(nested)
	if !scene.FocusNext(false) || scene.Focus() != nestedInput {
		t.Fatal("nested topmost modal was not focus scope")
	}
	upper := NewModal("upper")
	upperInput := NewEditBox("upper-input", ModeCustom)
	upper.Add(upperInput)
	scene.Root().Add(upper)
	if !scene.FocusNext(false) || scene.Focus() != upperInput {
		t.Fatal("topmost sibling modal did not override lower nested modal")
	}
	upper.SetEnabled(false)
	if !scene.FocusNext(false) || scene.Focus() != nestedInput {
		t.Fatal("disabled modal did not reveal eligible lower scope")
	}
	nested.SetVisible(false)
	if !scene.FocusNext(false) || scene.Focus() != a {
		t.Fatal("hidden nested modal did not reveal parent scope")
	}
	empty := NewModal("empty-topmost")
	empty.Add(NewLabel("message", "no input"))
	scene.Root().Add(empty)
	scene.FocusWidget(outside)
	if scene.FocusNext(false) || scene.Focus() != outside {
		t.Fatal("no-focusable topmost modal incorrectly fell back to background focus")
	}
}

func TestFocusNextRequiresUIThreadAndTransfersNativeFocus(t *testing.T) {
	var scene *Scene
	var first, second *EditBox
	opts := core.Options{Title: "WinUI isolated focus traversal test", StartHidden: true}
	BindScene(&opts, SceneHooks{OnCreate: func(_ *core.App, created *Scene) error {
		scene = created
		first = NewEditBox("native-first", ModeNative)
		second = NewEditBox("native-second", ModeNative)
		first.SetBounds(Rect{X: 10, Y: 10, W: 200, H: 40})
		second.SetBounds(Rect{X: 10, Y: 60, W: 200, H: 40})
		scene.Root().AddAll(first, second)
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
	if scene.FocusNext(false) {
		t.Fatal("off-thread FocusNext scheduled or claimed synchronous transfer")
	}
	done := make(chan struct{})
	if err := app.Post(func() {
		defer close(done)
		getFocus := windows.NewLazySystemDLL("user32.dll").NewProc("GetFocus")
		for _, want := range []Widget{first, second, first} {
			if !scene.FocusNext(false) || scene.Focus() != want {
				t.Error("UI-thread traversal failed")
			}
			actual, _, _ := getFocus.Call()
			if actual != uintptr(scene.nativeHandleFor(want)) {
				t.Errorf("scene and actual native focus differ: actual=%x expected=%x", actual, scene.nativeHandleFor(want))
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("native focus traversal callback timed out")
	}
}

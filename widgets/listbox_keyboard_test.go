//go:build windows

package widgets

import (
	"testing"
	"time"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func TestListBoxShowCheckSpaceTogglesOnceWithoutActivatingSelection(t *testing.T) {
	scene := NewScene(nil)
	list := NewListBox("products")
	list.SetItems([]ListItem{{Value: "stable-product-id", Text: "型号"}})
	list.SetShowCheck(true)
	list.SetSelected(0)
	checks, changes, activations := 0, 0, 0
	list.SetOnCheckChange(func(index int, item ListItem) {
		checks++
		if index != 0 || item.Value != "stable-product-id" || item.Checked != list.IsItemChecked(0) {
			t.Error("check callback did not expose the actual selected item's new state")
		}
	})
	list.SetOnChange(func(int, ListItem) { changes++ })
	list.SetOnActivate(func(int, ListItem) { activations++ })
	scene.Root().Add(list)
	scene.FocusWidget(list)
	if !scene.DispatchKeyDown(core.KeyEvent{Key: core.KeySpace}) || !list.IsItemChecked(0) || checks != 1 {
		t.Errorf("plain Space did not check selected item exactly once: checked=%v callbacks=%d", list.IsItemChecked(0), checks)
	}
	scene.DispatchChar(' ')
	for _, key := range []core.KeyEvent{
		{Key: core.KeySpace, Repeat: true}, {Key: core.KeySpace, Flags: 1 << 30},
		{Key: core.KeySpace, Ctrl: true}, {Key: core.KeySpace, Shift: true},
		{Key: core.KeySpace, Alt: true}, {Key: core.KeySpace, Flags: 1 << 29},
		{Key: core.KeySpace, Composing: true},
	} {
		scene.DispatchKeyDown(key)
	}
	if !list.IsItemChecked(0) || checks != 1 || changes != 0 || activations != 0 {
		t.Errorf("Space char/repeat/modifiers/IME changed selection or repeated check: checked=%v checks=%d changes=%d activations=%d", list.IsItemChecked(0), checks, changes, activations)
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeySpace})
	if list.IsItemChecked(0) || checks != 2 {
		t.Errorf("second plain Space did not uncheck once: checked=%v callbacks=%d", list.IsItemChecked(0), checks)
	}
	// Return keeps its existing independent activation semantics.
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn})
	if activations != 1 || list.IsItemChecked(0) || checks != 2 {
		t.Error("Return changed check state or lost normal list activation")
	}
}

func TestListBoxShowCheckSpaceSkipsIneligibleTargets(t *testing.T) {
	for _, name := range []string{"disabled-item", "no-selection", "disabled-widget", "hidden-ancestor", "checks-off"} {
		t.Run(name, func(t *testing.T) {
			scene := NewScene(nil)
			parent := NewPanel("parent")
			list := NewListBox("products")
			list.SetItems([]ListItem{{Value: "id", Text: "型号"}})
			list.SetShowCheck(true)
			list.SetSelected(0)
			parent.Add(list)
			scene.Root().Add(parent)
			scene.FocusWidget(list)
			switch name {
			case "disabled-item":
				list.SetItems([]ListItem{{Value: "id", Text: "型号", Disabled: true}})
			case "no-selection":
				list.ClearSelection()
			case "disabled-widget":
				list.SetEnabled(false)
			case "hidden-ancestor":
				parent.SetVisible(false)
			case "checks-off":
				list.SetShowCheck(false)
			}
			checks := 0
			list.SetOnCheckChange(func(int, ListItem) { checks++ })
			scene.DispatchKeyDown(core.KeyEvent{Key: core.KeySpace})
			if checks != 0 || list.IsItemChecked(0) {
				t.Error("Space checked an ineligible item/widget")
			}
		})
	}
}

func TestListBoxShowCheckNativePretranslationDoesNotLeakAfterFocusChange(t *testing.T) {
	var scene *Scene
	var list *ListBox
	var editor *EditBox
	checks, chars := 0, 0
	opts := core.Options{Title: "WinUI isolated list check keyboard test", StartHidden: true, OnChar: func(*core.App, rune) { chars++ }}
	BindScene(&opts, SceneHooks{OnCreate: func(_ *core.App, created *Scene) error {
		scene = created
		list = NewListBox("products")
		list.SetBounds(Rect{X: 10, Y: 10, W: 250, H: 90})
		list.SetItems([]ListItem{{Value: "stable-id", Text: "型号"}})
		list.SetShowCheck(true)
		list.SetSelected(0)
		editor = NewEditBox("draft", ModeCustom)
		editor.SetBounds(Rect{X: 10, Y: 110, W: 250, H: 40})
		list.SetOnCheckChange(func(int, ListItem) {
			checks++
			scene.FocusWidget(editor)
		})
		scene.Root().AddAll(list, editor)
		scene.FocusWidget(list)
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
			t.Fatal("list check UI callback timed out")
		}
	}
	post := windows.NewLazySystemDLL("user32.dll").NewProc("PostMessageW")
	for i := 0; i < 2; i++ {
		onUI(func() { scene.FocusWidget(list) })
		if ok, _, err := post.Call(uintptr(app.Handle()), 0x0100, uintptr(core.KeySpace), 1); ok == 0 {
			t.Fatal(err)
		}
		// WM_CHAR, if TranslateMessage wrongly creates one, is posted behind
		// the first callback. The second barrier observes that whole sequence.
		onUI(func() {})
		onUI(func() {
			actualFocus, _, _ := nativeUser32.NewProc("GetFocus").Call()
			if checks != i+1 || chars != 0 || scene.Focus() != editor || actualFocus != uintptr(editor.native.handle) || editor.TextValue() != "" {
				t.Errorf("check pretranslation/focus transfer failed: checks=%d chars=%d editorFocus=%v nativeFocus=%x text=%q", checks, chars, scene.Focus() == editor, actualFocus, editor.TextValue())
			}
		})
	}
}

//go:build windows

package widgets

import (
	"testing"
	"time"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func scrollTestNativeApp(t *testing.T, create func(*core.App, *Scene)) (*core.App, func(func())) {
	return scrollTestNativeAppMode(t, core.RenderModeAuto, create)
}

func scrollTestNativeAppMode(t *testing.T, mode core.RenderMode, create func(*core.App, *Scene)) (*core.App, func(func())) {
	t.Helper()
	opts := core.Options{Title: "WinUI isolated native clipping test", StartHidden: true, Width: 640, Height: 480, RenderMode: mode}
	BindScene(&opts, SceneHooks{OnCreate: func(app *core.App, scene *Scene) error { create(app, scene); return nil }})
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
			t.Fatal("native clipping UI callback timed out")
		}
	}
	return app, onUI
}

func assertNativeClip(t *testing.T, editor *EditBox, viewports ...Rect) {
	t.Helper()
	host := editor.nativeHostBounds()
	visible := host
	for _, viewport := range viewports {
		left, top := max(visible.X, viewport.X), max(visible.Y, viewport.Y)
		right, bottom := min(visible.X+visible.W, viewport.X+viewport.W), min(visible.Y+visible.H, viewport.Y+viewport.H)
		visible = Rect{X: left, Y: top, W: max(int32(0), right-left), H: max(int32(0), bottom-top)}
	}
	var box nativeRect
	kind, _, _ := nativeUser32.NewProc("GetWindowRgnBox").Call(uintptr(editor.native.handle), uintptr(unsafe.Pointer(&box)))
	if visible.Empty() {
		if kind != 1 {
			t.Errorf("fully clipped native host has region kind=%d box=%+v; want NULLREGION", kind, box)
		}
		return
	}
	if visible == host {
		if kind != 0 {
			t.Errorf("fully visible native host retained clipping region kind=%d box=%+v", kind, box)
		}
		return
	}
	want := nativeRect{Left: visible.X - host.X, Top: visible.Y - host.Y, Right: visible.X - host.X + visible.W, Bottom: visible.Y - host.Y + visible.H}
	if kind != 2 || box != want {
		t.Errorf("native region kind=%d box=%+v want=%+v host=%+v", kind, box, want, host)
	}
}

func TestNativeEditorPreservesTextSetBeforeSceneAttachment(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode ControlMode
	}{{"custom", ModeCustom}, {"native", ModeNative}} {
		t.Run(tc.name, func(t *testing.T) {
			_, onUI := scrollTestNativeApp(t, func(app *core.App, scene *Scene) {
				editor := NewEditBox("preseed", tc.mode)
				editor.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(200), H: app.DP(60)})
				editor.SetText("abcdef")
				if got := editor.TextValue(); got != "abcdef" {
					t.Errorf("before attachment: text=%q", got)
				}
				scene.Root().Add(editor)
				if got := editor.TextValue(); got != "abcdef" {
					t.Errorf("after attachment before focus: text=%q", got)
				}
				scene.FocusWidget(editor)
				if got := editor.TextValue(); got != "abcdef" {
					t.Errorf("after focus: text=%q", got)
				}
			})
			onUI(func() {})
		})
	}
}

func TestFocusedNativeEditorClipsAcrossScrollAndViewportResize(t *testing.T) {
	var scene *Scene
	var scroll *ScrollView
	var editor *EditBox
	app, onUI := scrollTestNativeApp(t, func(app *core.App, created *Scene) {
		scene = created
		scroll = NewScrollView("viewport")
		scroll.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(300), H: app.DP(160)})
		content := scrollTestContent("content", 600, 600)
		editor = NewEditBox("custom-input", ModeCustom)
		scrollTestPlace(editor, 20, 220, 200, 60)
		editor.SetText("abcdef")
		content.Add(editor)
		scroll.SetContent(content)
		scene.Root().Add(scroll)
		scene.FocusWidget(editor)
	})
	onUI(func() {
		handle := editor.native.handle
		if handle == 0 {
			t.Error("focused ModeCustom editor has no native backend")
			return
		}
		editor.SetText("abcdef")
		if got := editor.TextValue(); got != "abcdef" {
			t.Errorf("native editor text precondition: %q", got)
			return
		}
		setNativeSelection(handle, 2, 4)
		for _, offset := range []int32{150, 80, 0} {
			scroll.ScrollTo(0, app.DP(offset))
			assertNativeClip(t, editor, scroll.Bounds())
		}
		scroll.ScrollTo(0, app.DP(80))
		scroll.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(300), H: app.DP(240)})
		assertNativeClip(t, editor, scroll.Bounds())
		scroll.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(300), H: app.DP(120)})
		assertNativeClip(t, editor, scroll.Bounds())
		if editor.native.handle != handle || editor.TextValue() != "abcdef" {
			t.Errorf("clipping recreated native input or changed text: original=%x actual=%x currentText=%q", handle, editor.native.handle, editor.TextValue())
		}
		if start, end := getNativeSelection(handle); start != 2 || end != 4 {
			t.Errorf("native selection lost across clipping: %d,%d", start, end)
		}
		// Same visible crop, now a larger HWND: a previous full region must no
		// longer be treated as full merely because its cached crop is unchanged.
		scroll.ScrollTo(0, 0)
		viewport := scroll.Bounds()
		bounds := Rect{X: viewport.X + app.DP(20), Y: viewport.Y + app.DP(20), W: app.DP(100), H: app.DP(60)}
		editor.SetBounds(bounds)
		host := editor.nativeHostBounds()
		viewport.H = host.Y + host.H - viewport.Y
		scroll.SetBounds(viewport)
		editor.SetBounds(bounds) // absolute-layout metadata is deliberately unchanged
		assertNativeClip(t, editor, scroll.Bounds())
		fullCrop := nativeHostClip(editor, editor.nativeHostBounds())
		if fullCrop != (Rect{W: host.W, H: host.H}) {
			t.Error("same-crop cache test did not start with a fully visible host")
		}
		bounds.H += app.DP(150)
		editor.SetBounds(bounds)
		if got := nativeHostClip(editor, editor.nativeHostBounds()); got != fullCrop {
			t.Errorf("host growth changed crop precondition: before=%+v after=%+v", fullCrop, got)
		}
		assertNativeClip(t, editor, scroll.Bounds())

		// An invisible ancestor must remove the input surface even while the
		// focused custom editor retains its own native WS_VISIBLE style.
		scroll.SetVisible(false)
		var box nativeRect
		if kind, _, _ := nativeUser32.NewProc("GetWindowRgnBox").Call(uintptr(handle), uintptr(unsafe.Pointer(&box))); kind != 1 {
			t.Errorf("hidden ancestor retained native input region: kind=%d box=%+v", kind, box)
		}
		scroll.SetVisible(true)
		assertNativeClip(t, editor, scroll.Bounds())
		button := NewButton("focus-away", "Away", ModeCustom)
		button.SetBounds(Rect{X: app.DP(400), Y: app.DP(20), W: app.DP(100), H: app.DP(40)})
		scene.Root().Add(button)
		scene.FocusWidget(button)
		styleIndex := int32(-16) // GWL_STYLE; IsWindowVisible also tests our hidden top-level parent.
		style, _, _ := nativeUser32.NewProc("GetWindowLongW").Call(uintptr(handle), uintptr(styleIndex))
		if style&uintptr(nativeWindowVisible) != 0 {
			t.Error("custom editor native backend stayed visible after blur")
		}
		scene.FocusWidget(editor)
		style, _, _ = nativeUser32.NewProc("GetWindowLongW").Call(uintptr(handle), uintptr(styleIndex))
		if style&uintptr(nativeWindowVisible) == 0 || editor.native.handle != handle || editor.TextValue() != "abcdef" {
			t.Error("custom editor backend failed to restore without losing HWND/text")
		}
		assertNativeClip(t, editor, scroll.Bounds())
	})
}

func TestNativeEditorClipsToIntersectionOfNestedViewportsWithoutRegionLeak(t *testing.T) {
	var outer, inner *ScrollView
	var editor *EditBox
	app, onUI := scrollTestNativeApp(t, func(app *core.App, scene *Scene) {
		outer = NewScrollView("outer")
		outer.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(300), H: app.DP(180)})
		outerContent := scrollTestContent("outer-content", 700, 800)
		inner = NewScrollView("inner")
		scrollTestPlace(inner, 20, 120, 240, 180)
		innerContent := scrollTestContent("inner-content", 500, 500)
		editor = NewEditBox("native-input", ModeNative)
		scrollTestPlace(editor, 20, 30, 200, 90)
		innerContent.Add(editor)
		inner.SetContent(innerContent)
		outerContent.Add(inner)
		outer.SetContent(outerContent)
		scene.Root().Add(outer)
	})
	onUI(func() {
		assertNativeClip(t, editor, outer.Bounds(), inner.Bounds())
		inner.ScrollTo(0, app.DP(60))
		assertNativeClip(t, editor, outer.Bounds(), inner.Bounds())
		outer.ScrollTo(0, app.DP(100))
		assertNativeClip(t, editor, outer.Bounds(), inner.Bounds())
		process, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetCurrentProcess").Call()
		resources := nativeUser32.NewProc("GetGuiResources")
		before, _, _ := resources.Call(process, 0)
		for i := 0; i < 80; i++ {
			outer.ScrollTo(0, app.DP(int32((i%3)*100)))
			assertNativeClip(t, editor, outer.Bounds(), inner.Bounds())
		}
		outer.ScrollTo(0, app.DP(100))
		after, _, _ := resources.Call(process, 0)
		if before != 0 && after > before+2 {
			t.Errorf("native region changes leaked GDI resources: before=%d after=%d", before, after)
		}
	})
}

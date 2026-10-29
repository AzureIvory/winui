//go:build windows

package widgets

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/AzureIvory/winui/core"
)

func editScrollTestText() string {
	var text strings.Builder
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&text, "测试行%02d AaBbCcDdEeFfGgHhIiJjKk 0123456789\n", i)
	}
	return text.String()
}

func TestMultilineNativeEditorViewportSurvivesScrollbarAndLayoutChanges(t *testing.T) {
	for _, mode := range []ControlMode{ModeCustom, ModeNative} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			var edit *EditBox
			app, onUI := scrollTestNativeApp(t, func(app *core.App, scene *Scene) {
				edit = NewEditBox("viewport", mode)
				edit.SetMultiline(true)
				edit.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(320), H: app.DP(117)})
				edit.SetText(editScrollTestText())
				scene.Root().Add(edit)
				scene.FocusWidget(edit)
			})
			onUI(func() {
				for _, horizontal := range []bool{false, true, false} {
					edit.SetHorizontalScroll(horizontal)
					edit.SetWordWrap(!horizontal)
					handle := edit.native.handle
					// Text changes can show/hide native scrollbars without a
					// widget resize. The remaining client is still authoritative.
					for _, text := range []string{"短文本", editScrollTestText(), "短文本", editScrollTestText()} {
						edit.SetText(text)
						client, ok := getNativeClientRect(handle)
						var format nativeRect
						sendNativeMessage(handle, 0x00B2, 0, uintptr(unsafe.Pointer(&format)))
						if !ok || format != client {
							t.Errorf("text-driven scrollbar change left format=%+v client=%+v horizontal=%v", format, client, horizontal)
							return
						}
					}
					setNativeSelection(handle, 2, 5)
					for _, padding := range []int32{6, 12} {
						edit.SetStyle(EditStyle{PaddingDP: padding, CornerRadius: 6, Font: FontSpec{SizeDP: 13}})
						for _, height := range []int32{91, 113, 137} {
							edit.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(320), H: app.DP(height)})
							client, ok := getNativeClientRect(handle)
							if !ok || client.Right <= 0 || client.Bottom <= 0 {
								t.Errorf("invalid native client: %+v", client)
								return
							}
							var format nativeRect
							sendNativeMessage(handle, 0x00B2, 0, uintptr(unsafe.Pointer(&format))) // EM_GETRECT
							if format != client {
								t.Errorf("formatting viewport %+v differs from actual client %+v (horizontal=%v padding=%d height=%d)", format, client, horizontal, padding, height)
							}
							// Hit testing and painting must use the same physical text
							// viewport; padding must not remain in RichEdit's client.
							bounds := edit.Bounds()
							wantHost := Rect{X: bounds.X + app.DP(padding), Y: bounds.Y + app.DP(padding), W: bounds.W - app.DP(padding)*2, H: bounds.H - app.DP(padding)*2}
							if got := edit.native.geometry.bounds; got != wantHost {
								t.Errorf("native viewport %+v want %+v", got, wantHost)
							}
							sendNativeMessage(handle, 0x0115, 7, 0) // WM_VSCROLL / SB_BOTTOM
							if edit.native.handle != handle || edit.TextValue() != editScrollTestText() {
								t.Error("geometry change recreated the native host or changed text")
								return
							}
							if start, end := getNativeSelection(handle); start != 2 || end != 5 {
								t.Errorf("geometry change lost native selection: %d,%d", start, end)
								return
							}
						}
					}
				}
			})
		})
	}
}

// This opt-in check samples the real child window surface after RichEdit's own
// scroll blit, before any complete repaint. WM_PRINTCLIENT would repaint it and
// hide the bug. It uses a separate non-activating window and never sends input
// to an application under test. Run on an unobscured interactive Windows desktop.
func TestMultilineNativeEditorScrollPixelsMatchFreshPaint(t *testing.T) {
	if os.Getenv("WINUI_TEST_NATIVE_PAINT") != "1" {
		t.Skip("set WINUI_TEST_NATIVE_PAINT=1 for the real-window pixel regression")
	}
	for _, renderer := range []core.RenderMode{core.RenderModeAuto, core.RenderModeGDI} {
		for _, mode := range []ControlMode{ModeCustom, ModeNative} {
			t.Run(fmt.Sprintf("%v/%v", renderer, mode), func(t *testing.T) {
				var edit *EditBox
				app, onUI := scrollTestNativeAppMode(t, renderer, func(app *core.App, scene *Scene) {
					edit = NewEditBox("scroll-pixels", mode)
					edit.SetBounds(Rect{X: 20, Y: 20, W: 396, H: 113})
					edit.SetStyle(EditStyle{PaddingDP: 8, CornerRadius: 6, Font: FontSpec{SizeDP: 13}})
					edit.SetMultiline(true)
					edit.SetText(editScrollTestText())
					scene.Root().Add(edit)
					scene.FocusWidget(edit)
				})
				if err := app.SetScreenBounds(core.Rect{X: 20, Y: 20, W: 480, H: 240}); err != nil {
					t.Fatal(err)
				}
				if err := app.ShowNoActivate(); err != nil {
					t.Fatal(err)
				}
				if err := app.RaiseNoActivate(); err != nil {
					t.Fatal(err)
				}
				onUI(func() {
					handle := edit.native.handle
					setNativeSelection(handle, 0, 0)
					edit.ScrollToCaret()
					redraw := func() {
						procRedrawWindow.Call(uintptr(app.Handle()), 0, 0, 0x0001|0x0004|0x0080|0x0100) // INVALIDATE | ERASE | ALLCHILDREN | UPDATENOW
					}
					redraw()
					if renderer == core.RenderModeGDI && app.RenderBackend() != core.RenderBackendGDI {
						t.Errorf("GDI render request resolved to %v", app.RenderBackend())
						return
					}
					t.Logf("actual backend: %v", app.RenderBackend())
					for _, horizontal := range []bool{false, true} {
						edit.SetHorizontalScroll(horizontal)
						edit.SetWordWrap(!horizontal)
						handle = edit.native.handle
						setNativeSelection(handle, 0, 0)
						edit.ScrollToCaret()
						redraw()
						messages := []uint32{0x0115} // WM_VSCROLL
						if horizontal {
							messages = append(messages, 0x0114) // WM_HSCROLL
						}
						for _, message := range messages {
							moved := false
							for _, action := range []uintptr{1, 3, 7, 0, 2, 6} { // line/page down, bottom, line/page up, top
								var before, after core.Point
								sendNativeMessage(handle, 0x04DD, 0, uintptr(unsafe.Pointer(&before))) // EM_GETSCROLLPOS
								sendNativeMessage(handle, message, action, 0)
								sendNativeMessage(handle, 0x04DD, 0, uintptr(unsafe.Pointer(&after)))
								moved = moved || before != after
								nativeUser32.NewProc("UpdateWindow").Call(uintptr(handle))
								scrolled := captureNativeEditPixels(t, uintptr(handle))
								redraw()
								fresh := captureNativeEditPixels(t, uintptr(handle))
								if scrolled == nil || fresh == nil {
									return
								}
								if !bytes.Equal(scrolled.Pix, fresh.Pix) {
									t.Errorf("native scroll message %#x action %d left pixels differing from a complete repaint (horizontal=%v)", message, action, horizontal)
									return
								}
							}
							if !moved {
								t.Errorf("native scroll message %#x never moved the document", message)
							}
						}
					}
				})
			})
		}
	}
}

func captureNativeEditPixels(t *testing.T, hwnd uintptr) *image.RGBA {
	t.Helper()
	var rect nativeRect
	nativeUser32.NewProc("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	w, h := int(rect.Right), int(rect.Bottom)
	if w <= 0 || h <= 0 {
		t.Errorf("invalid capture size %d,%d", w, h)
		return nil
	}
	dc, _, _ := nativeUser32.NewProc("GetDC").Call(hwnd)
	if dc == 0 {
		t.Error("GetDC failed")
		return nil
	}
	defer nativeUser32.NewProc("ReleaseDC").Call(hwnd, dc)
	mem, _, _ := nativeGdi32.NewProc("CreateCompatibleDC").Call(dc)
	if mem == 0 {
		t.Error("CreateCompatibleDC failed")
		return nil
	}
	defer nativeGdi32.NewProc("DeleteDC").Call(mem)
	bitmap, _, _ := nativeGdi32.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(w), uintptr(h))
	if bitmap == 0 {
		t.Error("CreateCompatibleBitmap failed")
		return nil
	}
	defer procDeleteObject.Call(bitmap)
	old, _, _ := nativeGdi32.NewProc("SelectObject").Call(mem, bitmap)
	ok, _, _ := nativeGdi32.NewProc("BitBlt").Call(mem, 0, 0, uintptr(w), uintptr(h), dc, 0, 0, 0x00CC0020)
	nativeGdi32.NewProc("SelectObject").Call(mem, old)
	if ok == 0 {
		t.Error("BitBlt failed")
		return nil
	}
	info := struct {
		Size                         uint32
		Width, Height                int32
		Planes, BitCount             uint16
		Compression, SizeImage       uint32
		XPelsPerMeter, YPelsPerMeter int32
		ClrUsed, ClrImportant        uint32
	}{Size: 40, Width: int32(w), Height: -int32(h), Planes: 1, BitCount: 32}
	pixels := make([]byte, w*h*4)
	rows, _, err := nativeGdi32.NewProc("GetDIBits").Call(dc, bitmap, 0, uintptr(h), uintptr(unsafe.Pointer(&pixels[0])), uintptr(unsafe.Pointer(&info)), 0)
	if rows != uintptr(h) {
		t.Errorf("GetDIBits: rows %d: %v", rows, err)
		return nil
	}
	for i := 0; i < len(pixels); i += 4 {
		pixels[i], pixels[i+2], pixels[i+3] = pixels[i+2], pixels[i], 255
	}
	return &image.RGBA{Pix: pixels, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
}

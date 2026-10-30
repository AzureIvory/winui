//go:build windows

package widgets

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestListBoxItemActionHiddenNativeDPIAndRenderBackends(t *testing.T) {
	for _, renderer := range []core.RenderMode{core.RenderModeAuto, core.RenderModeGDI} {
		t.Run(fmt.Sprint(renderer), func(t *testing.T) {
			source := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					source.SetRGBA(x, y, color.RGBA{R: 40, G: 120, B: 200, A: 255})
				}
			}
			icon, err := core.NewImageFromDecoded(source)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = icon.Close() })
			var list *ListBox
			var scene *Scene
			actions := 0
			app, onUI := scrollTestNativeAppMode(t, renderer, func(_ *core.App, created *Scene) {
				scene = created
				list = NewListBox("action-dpi")
				list.SetStyle(ListStyle{ItemHeightDP: 40, PaddingDP: 4})
				list.SetItems([]ListItem{{Value: "first"}, {Value: "second"}, {Value: "third"}})
				list.SetShowCheck(true)
				list.SetItemAction(icon, func(int, ListItem) { actions++ })
				scene.Root().Add(list)
			})
			onUI(func() {
				for _, dpi := range []uintptr{96, 144, 192} {
					// Change only this hidden fixture's DPI; never inject desktop input.
					sendNativeMessage(app.Handle(), 0x02E0, dpi|dpi<<16, 0) // WM_DPICHANGED
					list.SetBounds(Rect{X: app.DP(20), Y: app.DP(30), W: app.DP(320), H: app.DP(100)})
					style := list.resolveStyle(nil)
					row := list.rowRect(0, &PaintCtx{scene: scene}, style)
					text, check, action := list.rowContentRects(row, style)
					if app.DP(32) != action.W || text.X+text.W > check.X || check.X+check.W > action.X {
						t.Errorf("DPI %d overlaps action geometry: text=%+v check=%+v action=%+v", dpi, text, check, action)
					}
					point := core.Point{X: action.X + action.W/2, Y: action.Y + action.H/2}
					before := actions
					listActionMouse(scene, point, true)
					listActionMouse(scene, point, false)
					if actions != before+1 {
						t.Errorf("DPI %d action hit does not match paint geometry", dpi)
					}
					// Execute the real scene/image paint path while the fixture stays hidden.
					for _, enabled := range []bool{true, false} {
						list.SetItemActionEnabled(enabled)
						sendNativeMessage(app.Handle(), 0x000F, 0, 0) // WM_PAINT
					}
					list.SetItemActionEnabled(true)
					for _, state := range []ActionVisualState{
						ActionVisualDefault, ActionVisualWorking, ActionVisualSuccess, ActionVisualError,
					} {
						presentation := ItemActionPresentation{Text: "发送中", State: state}
						provider := func(int, ListItem) ItemActionPresentation { return presentation }
						list.SetItemActionPresentation(provider)
						text, check, action = list.rowContentRects(row, style)
						if action.W != app.DP(76) || text.X+text.W > check.X || check.X+check.W > action.X {
							t.Errorf("DPI %d presentation columns overlap: text=%+v check=%+v action=%+v", dpi, text, check, action)
						}
						point = core.Point{X: action.X + action.W/2, Y: action.Y + action.H/2}
						for _, disabled := range []bool{false, true} {
							presentation.Disabled = disabled
							list.SetItemActionPresentation(provider)
							before := actions
							listActionMouse(scene, point, true)
							listActionMouse(scene, point, false)
							want := before
							if !disabled {
								want++
							}
							if actions != want {
								t.Errorf("DPI %d state %d disabled=%v actions=%d want=%d", dpi, state, disabled, actions, want)
							}
							sendNativeMessage(app.Handle(), 0x000F, 0, 0) // WM_PAINT
						}
					}
					list.SetItemActionPresentation(nil)
				}
				backend := app.RenderBackend()
				if backend == core.RenderBackendUnknown || renderer == core.RenderModeGDI && backend != core.RenderBackendGDI {
					t.Errorf("unexpected render backend: requested=%v actual=%v", renderer, backend)
				}
				t.Logf("actual backend: %v", backend)
				list.SetItemAction(nil, nil)
				scene.Root().Remove(list.ID())
				if icon.NaturalSize() != (core.Size{Width: 32, Height: 32}) {
					t.Error("list closed a host-owned image")
				}
			})
		})
	}
}

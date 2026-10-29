//go:build windows

package widgets

import (
	"fmt"
	"testing"
	"time"

	"github.com/AzureIvory/winui/core"
)

type comboInputWindow struct {
	app                                         *core.App
	scene                                       *Scene
	options                                     core.Options
	outer, inner                                *ScrollView
	combo                                       *ComboBox
	prime                                       *EditBox
	previewCalls, previousPreviewCalls, changes int
	onUI                                        func(func())
}

func newComboInputWindow(t *testing.T) *comboInputWindow {
	t.Helper()
	f := &comboInputWindow{}
	f.options = core.Options{Title: "WinUI isolated combo input acceptance", StartHidden: true, Width: 640, Height: 500,
		OnKeyPreview: func(_ *core.App, ev core.KeyEvent) bool { f.previousPreviewCalls++; return ev.Key == core.KeyEscape },
	}
	BindScene(&f.options, SceneHooks{
		OnCreate: func(app *core.App, scene *Scene) error {
			f.app, f.scene = app, scene
			f.outer = NewScrollView("combo-input-form")
			f.outer.SetWheelStep(7)
			scrollTestPlace(f.outer, 20, 20, 400, 350)
			outerContent := scrollTestContent("combo-input-form-content", 380, 1000)
			f.inner = NewScrollView("combo-input-section")
			f.inner.SetWheelStep(11)
			scrollTestPlace(f.inner, 20, 60, 320, 180)
			innerContent := scrollTestContent("combo-input-section-content", 300, 800)
			f.combo = NewComboBox("combo-input-model", ModeCustom)
			f.combo.SetStyle(ComboStyle{MaxVisibleItems: 5})
			items := make([]ListItem, 60)
			for i := range items {
				items[i] = ListItem{Text: fmt.Sprintf("model-%02d", i), Value: fmt.Sprint(i)}
			}
			f.combo.SetItems(items)
			f.combo.SetSelected(4)
			f.combo.SetOnChange(func(int, ListItem) { f.changes++ })
			scrollTestPlace(f.combo, 15, 20, 260, 30)
			innerContent.Add(f.combo)
			f.inner.SetContent(innerContent)
			outerContent.Add(f.inner)
			f.outer.SetContent(outerContent)
			scene.Root().Add(f.outer)
			f.outer.ScrollTo(0, 30)
			f.inner.ScrollTo(0, 10)
			f.prime = NewEditBox("combo-input-native-focus-prime", ModeCustom)
			scrollTestPlace(f.prime, 450, 20, 100, 30)
			scene.Root().Add(f.prime)
			return nil
		},
		OnKeyPreview: func(_ *core.App, _ *Scene, ev core.KeyEvent) bool { f.previewCalls++; return ev.Key == core.KeyEscape },
	})
	app, err := core.NewApp(f.options)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	f.onUI = func(fn func()) {
		t.Helper()
		done := make(chan struct{})
		if err := app.Post(func() { defer close(done); fn() }); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("combo input hidden UI callback timed out")
		}
	}
	return f
}

func (f *comboInputWindow) clickCombo(t *testing.T) {
	t.Helper()
	// Establish focus in the hidden native host before transferring to a custom
	// widget. The fixture never activates a visible window.
	f.scene.FocusWidget(f.prime)
	b := f.combo.Bounds()
	mouse := core.MouseEvent{Point: core.Point{X: b.X + b.W/2, Y: b.Y + b.H/2}, Button: core.MouseButtonLeft}
	f.scene.DispatchMouseDown(mouse)
	f.scene.DispatchMouseUp(mouse)
	if !f.combo.open {
		t.Fatal("actual Scene click did not open the custom popup")
	}
	if f.scene.focus != f.combo {
		t.Fatal("actual Scene click did not focus the custom combo")
	}
}

func (f *comboInputWindow) boundKey(ev core.KeyEvent) bool {
	ev.Handle = f.app.Handle()
	if f.options.OnKeyPreview(f.app, ev) {
		return true
	}
	f.options.OnKeyDown(f.app, ev)
	return false
}

func TestComboInputPopupWheelScrollsOptionsWithoutChangingSelectionOrForm(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.clickCombo(t)
		_, outerY := f.outer.ScrollOffset()
		_, innerY := f.inner.ScrollOffset()
		selection := f.combo.SelectedIndex()
		start := f.combo.popupLayout().start
		wheel := func(delta int32) bool {
			popup := f.combo.popupLayout().rect
			return f.scene.DispatchMouseWheel(core.MouseEvent{Point: core.Point{X: popup.X + 20, Y: popup.Y + popup.H/2}, Delta: delta})
		}
		if !wheel(-120) || f.combo.popupLayout().start <= start {
			t.Error("wheel inside actual popup did not scroll the option viewport")
		}
		for i := 0; i < 100; i++ {
			wheel(-120)
		}
		bottom := f.combo.popupLayout()
		if bottom.end != len(f.combo.Items()) {
			t.Errorf("popup cannot reach its final option: visible=[%d,%d)", bottom.start, bottom.end)
		}
		if !wheel(-120) {
			t.Error("popup boundary wheel escaped instead of being consumed")
		}
		for i := 0; i < 100; i++ {
			wheel(120)
		}
		if f.combo.popupLayout().start != 0 || !wheel(120) {
			t.Error("popup top boundary was not reachable/contained")
		}
		_, gotOuter := f.outer.ScrollOffset()
		_, gotInner := f.inner.ScrollOffset()
		if gotOuter != outerY || gotInner != innerY || f.combo.SelectedIndex() != selection || f.changes != 0 || !f.combo.open {
			t.Errorf("popup wheel changed form/selection: offsets=(%d,%d) want=(%d,%d) selected=%d changes=%d open=%t", gotOuter, gotInner, outerY, innerY, f.combo.SelectedIndex(), f.changes, f.combo.open)
		}
	})
}

func TestComboInputNavigationAndEscapePrecedeGlobalShortcuts(t *testing.T) {
	for _, route := range []string{"scene", "bound-pretranslate"} {
		t.Run(route, func(t *testing.T) {
			f := newComboInputWindow(t)
			f.onUI(func() {
				f.clickCombo(t)
				_, outerY := f.outer.ScrollOffset()
				_, innerY := f.inner.ScrollOffset()
				send := func(ev core.KeyEvent) bool {
					if route == "scene" {
						return f.scene.DispatchKeyDown(ev)
					}
					return f.boundKey(ev)
				}
				for _, key := range []struct {
					code uint32
					want int
				}{{core.KeyEnd, 59}, {core.KeyHome, 0}, {core.KeyDown, 1}, {core.KeyUp, 0}} {
					send(core.KeyEvent{Key: key.code})
					if f.combo.SelectedIndex() != key.want {
						t.Errorf("navigation key %d reached parent before combo: selected=%d want=%d", key.code, f.combo.SelectedIndex(), key.want)
					}
				}
				_, gotOuter := f.outer.ScrollOffset()
				_, gotInner := f.inner.ScrollOffset()
				if gotOuter != outerY || gotInner != innerY {
					t.Errorf("combo navigation scrolled parent containers: (%d,%d) want (%d,%d)", gotOuter, gotInner, outerY, innerY)
				}
				before, previous := f.previewCalls, f.previousPreviewCalls
				if !send(core.KeyEvent{Key: core.KeyEscape}) || f.combo.open || f.previewCalls != before || f.previousPreviewCalls != previous {
					t.Error("first Escape did not close only the popup before both global previews")
				}
				if !send(core.KeyEvent{Key: core.KeyEscape}) || f.previewCalls != before+1 {
					t.Error("second Escape did not reach the global shortcut exactly once")
				}
			})
		})
	}
}

func TestComboInputComposingKeysBypassPopupAndGlobalRoutes(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.clickCombo(t)
		selected, changes, previews, previous := f.combo.SelectedIndex(), f.changes, f.previewCalls, f.previousPreviewCalls
		_, outerY := f.outer.ScrollOffset()
		_, innerY := f.inner.ScrollOffset()
		for _, key := range []uint32{core.KeyEscape, core.KeyReturn, core.KeyHome, core.KeyEnd, core.KeyDown, core.KeyUp} {
			if f.scene.DispatchKeyDown(core.KeyEvent{Key: key, Composing: true}) {
				t.Error("Scene consumed an IME composition key")
			}
			if f.boundKey(core.KeyEvent{Key: key, Composing: true}) {
				t.Error("BindScene consumed an IME composition key")
			}
		}
		_, gotOuter := f.outer.ScrollOffset()
		_, gotInner := f.inner.ScrollOffset()
		if !f.combo.open || f.combo.SelectedIndex() != selected || f.changes != changes || f.previewCalls != previews || f.previousPreviewCalls != previous || gotOuter != outerY || gotInner != innerY {
			t.Error("IME composition changed combo, scroll position or global shortcut state")
		}
	})
}

func TestComboInputUnavailableAncestorDoesNotStealGlobalEscape(t *testing.T) {
	for _, state := range []string{"hidden", "disabled"} {
		for _, route := range []string{"scene", "bound-pretranslate"} {
			t.Run(state+"/"+route, func(t *testing.T) {
				f := newComboInputWindow(t)
				f.onUI(func() {
					f.clickCombo(t)
					if state == "hidden" {
						f.inner.Content().SetVisible(false)
					} else {
						f.inner.Content().SetEnabled(false)
					}
					before, previous := f.previewCalls, f.previousPreviewCalls
					var consumed bool
					if route == "scene" {
						consumed = f.scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyEscape})
					} else {
						consumed = f.boundKey(core.KeyEvent{Key: core.KeyEscape})
					}
					if !consumed || f.previewCalls != before+1 || f.previousPreviewCalls != previous {
						t.Error("a popup in an unavailable ancestor intercepted the active page's global Escape")
					}
					selected, changes := f.combo.SelectedIndex(), f.changes
					outerX, outerY := f.outer.ScrollOffset()
					innerX, innerY := f.inner.ScrollOffset()
					f.scene.SetKeyPreview(func(core.KeyEvent) bool { f.previewCalls++; return false })
					for _, key := range []uint32{core.KeyHome, core.KeyDown} {
						if route == "scene" {
							consumed = f.scene.DispatchKeyDown(core.KeyEvent{Key: key})
						} else {
							consumed = f.boundKey(core.KeyEvent{Key: key})
						}
						if consumed {
							t.Error("inactive focused descendant consumed a key after the global preview declined it")
						}
					}
					gotX, gotY := f.outer.ScrollOffset()
					gotInnerX, gotInnerY := f.inner.ScrollOffset()
					if f.combo.SelectedIndex() != selected || f.changes != changes || gotX != outerX || gotY != outerY || gotInnerX != innerX || gotInnerY != innerY || f.previewCalls != before+3 {
						t.Error("declined global Home/Down fell through to a combo or scroll ancestor inside a hidden/disabled panel")
					}
				})
			})
		}
	}
}

func TestComboInputOrdinaryNestedWheelStartsAtInnermostConsumer(t *testing.T) {
	for _, kind := range []string{"scroll", "list"} {
		t.Run(kind, func(t *testing.T) {
			f := newComboInputWindow(t)
			f.onUI(func() {
				var target Widget = f.inner
				var list *ListBox
				if kind == "list" {
					list = NewListBox("nested-wheel-list")
					items := make([]ListItem, 100)
					for i := range items {
						items[i] = ListItem{Text: fmt.Sprint(i)}
					}
					list.SetItems(items)
					scrollTestPlace(list, 0, 0, 280, 130)
					f.inner.SetContent(list)
					f.inner.ScrollTo(0, 0)
					target = list
				}
				_, outerY := f.outer.ScrollOffset()
				_, innerY := f.inner.ScrollOffset()
				b := target.Bounds()
				point := core.Point{X: b.X + 30, Y: b.Y + 30}
				if !f.scene.DispatchMouseWheel(core.MouseEvent{Point: point, Delta: -120}) {
					t.Error("nested wheel was not handled")
				}
				_, gotOuter := f.outer.ScrollOffset()
				_, gotInner := f.inner.ScrollOffset()
				if kind == "scroll" && (gotInner != innerY+11 || gotOuter != outerY) {
					t.Errorf("nested scroll did not consume exactly one step first: inner=%d outer=%d", gotInner, gotOuter)
				}
				if kind == "list" && (list.scroll != 3 || gotOuter != outerY || gotInner != innerY) {
					t.Errorf("nested list did not consume exactly one wheel step first: list=%d inner=%d outer=%d", list.scroll, gotInner, gotOuter)
				}
				if kind == "list" {
					list.SetSelected(len(list.Items()) - 1)
				} else {
					f.inner.ScrollTo(0, 1<<30)
				}
				_, innerEnd := f.inner.ScrollOffset()
				beforeList := 0
				if list != nil {
					beforeList = list.scroll
				}
				f.outer.ScrollTo(0, outerY)
				b = target.Bounds()
				point = core.Point{X: b.X + 30, Y: b.Y + 30}
				f.scene.DispatchMouseWheel(core.MouseEvent{Point: point, Delta: -120})
				_, gotOuter = f.outer.ScrollOffset()
				_, gotInner = f.inner.ScrollOffset()
				if gotOuter != outerY+7 || gotInner != innerEnd || list != nil && list.scroll != beforeList {
					t.Errorf("nested boundary did not bubble exactly once: inner=%d outer=%d wantOuter=%d", gotInner, gotOuter, outerY+7)
				}
			})
		})
	}
}

type comboInputCountingPanel struct {
	*Panel
	wheels int
}

func (p *comboInputCountingPanel) Add(child Widget) {
	p.Panel.Add(child)
	asWidgetNode(child).setParent(p)
}
func (p *comboInputCountingPanel) OnEvent(ev Event) bool {
	if ev.Type == EventMouseWheel {
		p.wheels++
	}
	return false
}

func TestComboInputUnhandledWheelVisitsEachAncestorOnce(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.outer.ScrollTo(0, 0)
		f.inner.ScrollTo(0, 0)
		probe := &comboInputCountingPanel{Panel: NewPanel("wheel-ancestor")}
		leaf := NewLabel("wheel-leaf", "unhandled target")
		scrollTestPlace(leaf, 0, 0, 200, 80)
		probe.Add(leaf)
		f.inner.SetContent(probe)
		b := leaf.Bounds()
		f.scene.DispatchMouseWheel(core.MouseEvent{Point: core.Point{X: b.X + 10, Y: b.Y + 10}, Delta: 120})
		if probe.wheels != 1 {
			t.Errorf("unhandled wheel visited the same ancestor %d times; want one target-to-root pass", probe.wheels)
		}
	})
}

func TestComboInputScrolledPopupClicksTheVisibleItemAndRevealsSameSelection(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.clickCombo(t)
		wheelDown := func() {
			popup := f.combo.popupLayout().rect
			f.scene.DispatchMouseWheel(core.MouseEvent{Point: core.Point{X: popup.X + 20, Y: popup.Y + popup.H/2}, Delta: -120})
		}
		for i := 0; i < 100; i++ {
			wheelDown()
		}
		layout := f.combo.popupLayout()
		if layout.end != len(f.combo.Items()) {
			t.Fatal("fixture could not reach the final popup option")
		}
		point := core.Point{X: layout.rect.X + 20, Y: layout.rect.Y + layout.rect.H - layout.padding - 1}
		f.scene.DispatchMouseDown(core.MouseEvent{Point: point, Button: core.MouseButtonLeft})
		f.scene.DispatchMouseUp(core.MouseEvent{Point: point, Button: core.MouseButtonLeft})
		if f.combo.open || f.combo.SelectedIndex() != 59 || f.changes != 1 {
			t.Errorf("scrolled popup click used the wrong item identity: selected=%d changes=%d open=%t", f.combo.SelectedIndex(), f.changes, f.combo.open)
		}
		f.clickCombo(t)
		f.boundKey(core.KeyEvent{Key: core.KeyHome})
		if f.combo.SelectedIndex() != 0 || f.combo.popupLayout().start != 0 {
			t.Fatal("Home did not select and reveal the first option")
		}
		for i := 0; i < 20; i++ {
			wheelDown()
		}
		if f.combo.popupLayout().start == 0 || f.combo.SelectedIndex() != 0 {
			t.Fatal("fixture did not independently scroll away from its unchanged selection")
		}
		before := f.changes
		f.boundKey(core.KeyEvent{Key: core.KeyHome})
		if f.combo.SelectedIndex() != 0 || f.combo.popupLayout().start != 0 || f.changes != before {
			t.Error("same-index Home failed to reveal the selection or emitted a duplicate change")
		}
	})
}

func TestComboInputActivationRepeatAndModifiersRespectGlobalOwnership(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.clickCombo(t)
		f.scene.SetKeyPreview(func(ev core.KeyEvent) bool { f.previewCalls++; return ev.Ctrl || ev.Alt || ev.Shift })
		for _, modified := range []core.KeyEvent{{Key: core.KeyReturn, Ctrl: true}, {Key: core.KeySpace, Alt: true}, {Key: core.KeyHome, Shift: true}} {
			before := f.previewCalls
			if !f.boundKey(modified) || !f.combo.open || f.combo.SelectedIndex() != 4 || f.previewCalls != before+1 {
				t.Error("modified combo key did not remain owned by the global shortcut")
			}
		}
		if !f.boundKey(core.KeyEvent{Key: core.KeyReturn}) || f.combo.open {
			t.Error("first Enter did not close popup in pretranslation")
		}
		for _, repeat := range []core.KeyEvent{{Key: core.KeyReturn, Repeat: true}, {Key: core.KeyReturn, Flags: 1 << 30}} {
			if !f.boundKey(repeat) || f.combo.open {
				t.Error("repeated Enter toggled the popup again")
			}
		}
		if !f.boundKey(core.KeyEvent{Key: core.KeySpace}) || !f.combo.open {
			t.Error("Space did not open popup in pretranslation")
		}
		if !f.boundKey(core.KeyEvent{Key: core.KeySpace, Repeat: true}) || !f.combo.open {
			t.Error("repeated Space closed the popup")
		}
	})
}

func TestComboInputScrollThumbCaptureReleasesOutsideViewport(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		_, thumb, _, _ := f.outer.scrollbarRectsForWidget()
		if thumb.Empty() {
			t.Fatal("fixture has no actual outer scrollbar thumb")
		}
		point := core.Point{X: thumb.X + thumb.W/2, Y: thumb.Y + thumb.H/2}
		_, before := f.outer.ScrollOffset()
		f.scene.DispatchMouseDown(core.MouseEvent{Point: point, Button: core.MouseButtonLeft})
		outside := core.Point{X: f.outer.Bounds().X + f.outer.Bounds().W + 50, Y: point.Y + 40}
		f.scene.DispatchMouseMove(core.MouseEvent{Point: outside, Button: core.MouseButtonLeft})
		_, dragged := f.outer.ScrollOffset()
		if dragged == before {
			t.Error("Scene capture did not drag the scrollbar outside its viewport")
		}
		f.scene.DispatchMouseUp(core.MouseEvent{Point: outside, Button: core.MouseButtonLeft})
		f.scene.DispatchMouseMove(core.MouseEvent{Point: core.Point{X: outside.X, Y: outside.Y + 80}})
		_, after := f.outer.ScrollOffset()
		if after != dragged || f.scene.capture != nil {
			t.Error("released outside thumb retained capture or continued scrolling")
		}
	})
}

func TestComboInputHorizontalPopupWheelBubblesWithoutVerticalConversion(t *testing.T) {
	f := newComboInputWindow(t)
	f.onUI(func() {
		f.outer.SetHorizontalScroll(true)
		extent := NewLabel("combo-input-horizontal-extent", "")
		scrollTestPlace(extent, 1000, 0, 1, 1)
		f.outer.Content().(*Panel).Add(extent)
		f.clickCombo(t)
		start := f.combo.popupLayout().start
		outerX, outerY := f.outer.ScrollOffset()
		innerX, innerY := f.inner.ScrollOffset()
		popup := f.combo.popupLayout().rect
		if !f.scene.DispatchMouseWheel(core.MouseEvent{Point: core.Point{X: popup.X + 20, Y: popup.Y + popup.H/2}, Delta: 120, Horizontal: true}) {
			t.Error("horizontal popup wheel failed to bubble to its capable ancestor")
		}
		gotX, gotY := f.outer.ScrollOffset()
		gotInnerX, gotInnerY := f.inner.ScrollOffset()
		if gotX != outerX+7 || gotY != outerY || gotInnerX != innerX || gotInnerY != innerY || f.combo.popupLayout().start != start || f.combo.SelectedIndex() != 4 || f.changes != 0 {
			t.Errorf("horizontal popup wheel converted/repeated: outer=(%d,%d) inner=(%d,%d) first=%d", gotX, gotY, gotInnerX, gotInnerY, f.combo.popupLayout().start)
		}
	})
}

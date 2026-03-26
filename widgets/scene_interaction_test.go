//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestComboBoxPopupHitPriority(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 400, H: 300})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 20, W: 160, H: 40})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
	})

	under := NewButton("under", "Under", ModeCustom)
	under.SetBounds(Rect{X: 20, Y: 66, W: 160, H: 40})

	scene.Root().Add(combo)
	scene.Root().Add(under)

	combo.open = true
	point := comboPopupPoint(combo, 0)

	if raw := scene.hitTest(scene.root, point.X, point.Y); raw != under {
		t.Fatalf("expected regular hit test to see underlying widget, got %T", raw)
	}
	if target := scene.mouseTargetAt(point.X, point.Y, nil); target != combo {
		t.Fatalf("expected popup overlay target to win, got %T", target)
	}
}

func TestComboBoxPopupClickDoesNotPassThrough(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 400, H: 300})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 20, W: 160, H: 40})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
	})

	clicked := 0
	under := NewButton("under", "Under", ModeCustom)
	under.SetBounds(Rect{X: 20, Y: 66, W: 160, H: 40})
	under.SetOnClick(func() {
		clicked++
	})

	scene.Root().Add(combo)
	scene.Root().Add(under)

	basePoint := core.Point{X: 30, Y: 30}
	scene.DispatchMouseDown(core.MouseEvent{Point: basePoint, Button: core.MouseButtonLeft})
	scene.DispatchMouseUp(core.MouseEvent{Point: basePoint, Button: core.MouseButtonLeft})

	if !combo.open {
		t.Fatalf("expected combo popup to open")
	}

	itemPoint := comboPopupPoint(combo, 0)
	scene.DispatchMouseDown(core.MouseEvent{Point: itemPoint, Button: core.MouseButtonLeft})
	scene.DispatchMouseUp(core.MouseEvent{Point: itemPoint, Button: core.MouseButtonLeft})

	if combo.SelectedIndex() != 0 {
		t.Fatalf("expected first popup item selected, got %d", combo.SelectedIndex())
	}
	if combo.open {
		t.Fatalf("expected combo popup to close after selection")
	}
	if clicked != 0 {
		t.Fatalf("expected popup click not to reach underlying widget, got %d clicks", clicked)
	}
}

func TestEditBoxHoverStaysStableOnArrowFallback(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 300, H: 200})

	edit := NewEditBox("edit", ModeCustom)
	edit.SetBounds(Rect{X: 20, Y: 20, W: 160, H: 36})

	flaky := newFlakyArrowWidget("flaky")
	flaky.SetBounds(edit.Bounds())

	scene.Root().Add(edit)
	scene.Root().Add(flaky)

	point := core.Point{X: 24, Y: 24}
	scene.DispatchMouseMove(core.MouseEvent{Point: point})
	scene.DispatchMouseMove(core.MouseEvent{Point: point})
	scene.DispatchMouseMove(core.MouseEvent{Point: point})

	if scene.hover != edit {
		t.Fatalf("expected hover target to stay on edit box, got %T", scene.hover)
	}
	if !edit.Hover {
		t.Fatalf("expected edit box hover state to remain true")
	}
	if cursorFor(scene.hover) != core.CursorIBeam {
		t.Fatalf("expected stable IBeam cursor target, got %v", cursorFor(scene.hover))
	}
}

func newTestScene(bounds Rect) *Scene {
	root := NewPanel("root")
	scene := &Scene{
		root:          root,
		theme:         DefaultTheme(),
		fonts:         make(map[FontSpec]*core.Font),
		timers:        make(map[uintptr]Widget),
		nativeTargets: make(map[uintptr]nativeCommandHandler),
	}
	root.setScene(scene)
	root.SetBounds(bounds)
	return scene
}

func comboPopupPoint(combo *ComboBox, index int) core.Point {
	style := mergeComboStyle(DefaultTheme().ComboBox, combo.Style)
	popup := combo.popupRect()
	padding := combo.dp(style.PaddingDP)
	itemHeight := combo.dp(style.ItemHeightDP)
	return core.Point{
		X: popup.X + 10,
		Y: popup.Y + padding + int32(index)*itemHeight + itemHeight/2,
	}
}

type flakyArrowWidget struct {
	widgetBase
	nextHit bool
}

func newFlakyArrowWidget(id string) *flakyArrowWidget {
	return &flakyArrowWidget{
		widgetBase: newWidgetBase(id, "flaky"),
		nextHit:    false,
	}
}

func (w *flakyArrowWidget) SetBounds(rect Rect) {
	w.widgetBase.setBounds(w, rect)
}

func (w *flakyArrowWidget) SetVisible(visible bool) {
	w.widgetBase.setVisible(w, visible)
}

func (w *flakyArrowWidget) SetEnabled(enabled bool) {
	w.widgetBase.setEnabled(w, enabled)
}

func (w *flakyArrowWidget) HitTest(x, y int32) bool {
	if !w.widgetBase.HitTest(x, y) {
		return false
	}
	hit := w.nextHit
	w.nextHit = !w.nextHit
	return hit
}

func (w *flakyArrowWidget) OnEvent(Event) bool {
	return false
}

func (w *flakyArrowWidget) Paint(*PaintCtx) {}

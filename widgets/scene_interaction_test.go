//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestComboBoxPopupHitPriority(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 400, H: 240})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 180, W: 160, H: 40})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
		{Value: "c", Text: "Gamma"},
	})

	under := NewButton("under", "Under", ModeCustom)
	under.SetBounds(Rect{X: 20, Y: 96, W: 160, H: 40})

	scene.Root().Add(combo)
	scene.Root().Add(under)

	combo.open = true
	point := comboPopupPoint(combo, 1)

	if raw := scene.hitTest(scene.root, point.X, point.Y); raw != under {
		t.Fatalf("expected regular hit test to see underlying widget, got %T", raw)
	}
	if overlay := scene.hitTestOverlay(scene.root, point.X, point.Y); overlay != combo {
		t.Fatalf("expected overlay hit test to resolve combo popup, got %T", overlay)
	}
	if target := scene.mouseTargetAt(point.X, point.Y, under); target != combo {
		t.Fatalf("expected popup overlay target to win, got %T", target)
	}
}

func TestComboBoxPopupOpensUpNearBottom(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 320, H: 220})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 170, W: 160, H: 36})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
		{Value: "c", Text: "Gamma"},
	})

	scene.Root().Add(combo)
	combo.open = true

	popup := combo.popupRect()
	if popup.Empty() {
		t.Fatalf("expected popup rect")
	}
	if popup.Y >= combo.Bounds().Y {
		t.Fatalf("expected popup to open upward, got popup=%#v combo=%#v", popup, combo.Bounds())
	}

	point := comboPopupPoint(combo, 0)
	row := combo.popupRowRect(0, nil, ComboStyle{})
	if row.Empty() {
		t.Fatalf("expected first popup row rect")
	}
	if !popup.Contains(point.X, point.Y) {
		t.Fatalf("expected popup point inside popup, got popup=%#v point=%#v", popup, point)
	}
	if !combo.overlayHitTest(point.X, point.Y) {
		t.Fatalf("expected overlay hit test for visible popup row")
	}
	if index := combo.popupIndexAt(point); index != 0 {
		t.Fatalf("expected popup index 0, got %d", index)
	}
}

func TestComboBoxPopupClampsToSceneViewport(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 320, H: 160})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 70, W: 160, H: 36})
	combo.SetItems([]ListItem{
		{Value: "0", Text: "Zero"},
		{Value: "1", Text: "One"},
		{Value: "2", Text: "Two"},
		{Value: "3", Text: "Three"},
		{Value: "4", Text: "Four"},
		{Value: "5", Text: "Five"},
		{Value: "6", Text: "Six"},
		{Value: "7", Text: "Seven"},
	})

	scene.Root().Add(combo)
	combo.open = true

	style := combo.popupStyle()
	fullHeight := combo.dp(style.PaddingDP)*2 + int32(min(int(style.MaxVisibleItems), len(combo.items)))*combo.dp(style.ItemHeightDP)
	popup := combo.popupRect()
	viewport := scene.Root().Bounds()

	if popup.Empty() {
		t.Fatalf("expected popup rect")
	}
	if popup.Y < viewport.Y || popup.Y+popup.H > viewport.Y+viewport.H {
		t.Fatalf("expected popup to stay inside viewport, got popup=%#v viewport=%#v", popup, viewport)
	}
	if popup.H >= fullHeight {
		t.Fatalf("expected constrained popup height, got popup=%#v fullHeight=%d", popup, fullHeight)
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

func TestComboBoxPopupClickDoesNotPassThroughNearBottom(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 400, H: 240})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 180, W: 160, H: 40})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
		{Value: "c", Text: "Gamma"},
	})

	clicked := 0
	under := NewButton("under", "Under", ModeCustom)
	under.SetBounds(Rect{X: 20, Y: 96, W: 160, H: 40})
	under.SetOnClick(func() {
		clicked++
	})

	scene.Root().Add(combo)
	scene.Root().Add(under)

	basePoint := core.Point{X: 30, Y: 190}
	scene.DispatchMouseDown(core.MouseEvent{Point: basePoint, Button: core.MouseButtonLeft})
	scene.DispatchMouseUp(core.MouseEvent{Point: basePoint, Button: core.MouseButtonLeft})

	if !combo.open {
		t.Fatalf("expected combo popup to open")
	}

	itemPoint := comboPopupPoint(combo, 1)
	scene.DispatchMouseDown(core.MouseEvent{Point: itemPoint, Button: core.MouseButtonLeft})
	scene.DispatchMouseUp(core.MouseEvent{Point: itemPoint, Button: core.MouseButtonLeft})

	if combo.SelectedIndex() != 1 {
		t.Fatalf("expected second popup item selected, got %d", combo.SelectedIndex())
	}
	if clicked != 0 {
		t.Fatalf("expected upward popup click not to reach underlying widget, got %d clicks", clicked)
	}
}

func TestEditBoxHoverStaysStableOnArrowFallback(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 300, H: 200})

	edit := NewEditBox("edit", ModeCustom)
	edit.SetBounds(Rect{X: 20, Y: 20, W: 160, H: 36})

	flaky := newFlakyCursorWidget("flaky-arrow", core.CursorArrow)
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

func TestEditBoxHoverStaysStableOnNonArrowSibling(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 300, H: 200})

	edit := NewEditBox("edit", ModeCustom)
	edit.SetBounds(Rect{X: 20, Y: 20, W: 160, H: 36})

	flaky := newFlakyCursorWidget("flaky-hand", core.CursorHand)
	flaky.SetBounds(edit.Bounds())

	scene.Root().Add(edit)
	scene.Root().Add(flaky)

	point := core.Point{X: 178, Y: 24}
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

func TestComboBoxDirtyRectIncludesPopupOverlay(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 320, H: 220})

	combo := NewComboBox("combo", ModeCustom)
	combo.SetBounds(Rect{X: 20, Y: 170, W: 160, H: 36})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
		{Value: "c", Text: "Gamma"},
	})

	scene.Root().Add(combo)
	combo.open = true

	dirty := widgetDirtyRect(combo)
	popup := combo.popupRect()
	if popup.Empty() {
		t.Fatalf("expected popup rect")
	}
	if !rectContainsRect(dirty, popup) {
		t.Fatalf("expected dirty rect to include popup overlay, got dirty=%#v popup=%#v", dirty, popup)
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
	row := combo.popupRowRect(index, nil, ComboStyle{})
	return core.Point{
		X: row.X + max32(1, min32(10, row.W-1)),
		Y: row.Y + max32(0, row.H/2),
	}
}

func rectContainsRect(outer, inner Rect) bool {
	if inner.Empty() {
		return true
	}
	return inner.X >= outer.X &&
		inner.Y >= outer.Y &&
		inner.X+inner.W <= outer.X+outer.W &&
		inner.Y+inner.H <= outer.Y+outer.H
}

type flakyCursorWidget struct {
	widgetBase
	nextHit  bool
	cursorID CursorID
}

func newFlakyCursorWidget(id string, cursorID CursorID) *flakyCursorWidget {
	return &flakyCursorWidget{
		widgetBase: newWidgetBase(id, "flaky"),
		cursorID:   cursorID,
	}
}

func (w *flakyCursorWidget) SetBounds(rect Rect) {
	w.widgetBase.setBounds(w, rect)
}

func (w *flakyCursorWidget) SetVisible(visible bool) {
	w.widgetBase.setVisible(w, visible)
}

func (w *flakyCursorWidget) SetEnabled(enabled bool) {
	w.widgetBase.setEnabled(w, enabled)
}

func (w *flakyCursorWidget) HitTest(x, y int32) bool {
	if !w.widgetBase.HitTest(x, y) {
		return false
	}
	hit := w.nextHit
	w.nextHit = !w.nextHit
	return hit
}

func (w *flakyCursorWidget) OnEvent(Event) bool {
	return false
}

func (w *flakyCursorWidget) Paint(*PaintCtx) {}

func (w *flakyCursorWidget) cursor() CursorID {
	return w.cursorID
}

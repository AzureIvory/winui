//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func scrollTestPlace(widget Widget, x, y, w, h int32) {
	widget.SetLayoutData(AbsoluteLayoutData{Left: x, Top: y, Width: w, Height: h, HasLeft: true, HasTop: true, HasWidth: true, HasHeight: true})
}
func scrollTestContent(id string, w, h int32) *Panel {
	panel := NewPanel(id)
	panel.SetLayout(AbsoluteLayout{})
	extent := NewLabel(id+"-extent", "")
	scrollTestPlace(extent, 0, h-1, w, 1)
	panel.Add(extent)
	return panel
}

func TestFocusRevealUsesMinimumScrollAndKeepsOversizedTargetsStable(t *testing.T) {
	scene := NewScene(nil)
	scroll := NewScrollView("viewport")
	scroll.SetHorizontalScroll(true)
	scroll.SetBounds(Rect{W: 100, H: 100})
	content := scrollTestContent("content", 500, 600)
	inside := NewEditBox("inside", ModeCustom)
	scrollTestPlace(inside, 20, 20, 30, 30)
	partial := NewEditBox("partial", ModeCustom)
	scrollTestPlace(partial, 20, 90, 30, 30)
	far := NewEditBox("far", ModeCustom)
	scrollTestPlace(far, 300, 400, 40, 40)
	large := NewEditBox("large", ModeCustom)
	scrollTestPlace(large, 160, 200, 180, 250)
	content.AddAll(inside, partial, far, large)
	scroll.SetContent(content)
	scene.Root().Add(scroll)
	assertOffset := func(x, y int32) {
		t.Helper()
		gotX, gotY := scroll.ScrollOffset()
		if gotX != x || gotY != y {
			t.Fatalf("scroll offset=(%d,%d) want=(%d,%d)", gotX, gotY, x, y)
		}
	}
	scene.FocusWidget(inside)
	assertOffset(0, 0)
	scene.FocusWidget(partial)
	assertOffset(0, 20)
	scene.FocusWidget(partial)
	assertOffset(0, 20)
	scene.FocusWidget(far)
	assertOffset(240, 340)
	scene.FocusWidget(far)
	assertOffset(240, 340)
	scroll.ScrollTo(0, 0)
	scene.FocusWidget(large)
	assertOffset(160, 200)
	scene.FocusWidget(large)
	assertOffset(160, 200)
	scroll.ScrollTo(180, 230)
	scene.FocusWidget(large)
	assertOffset(180, 230)
	scroll.ScrollTo(310, 400)
	scene.FocusWidget(large)
	assertOffset(240, 350)
	scene.FocusWidget(large)
	assertOffset(240, 350)
	if got := large.Bounds(); got.X > 0 || got.Y > 0 || got.X+got.W < 100 || got.Y+got.H < 100 {
		t.Fatalf("oversized target failed to cover viewport: %+v", got)
	}
}

func TestFocusWidgetAndFocusNextRevealThroughNestedScrollViews(t *testing.T) {
	scene := NewScene(nil)
	outer := NewScrollView("outer")
	outer.SetHorizontalScroll(true)
	outer.SetBounds(Rect{W: 200, H: 120})
	outerContent := scrollTestContent("outer-content", 800, 800)
	inner := NewScrollView("inner")
	inner.SetHorizontalScroll(true)
	scrollTestPlace(inner, 300, 400, 150, 100)
	innerContent := scrollTestContent("inner-content", 500, 400)
	target := NewEditBox("target", ModeCustom)
	scrollTestPlace(target, 300, 200, 40, 30)
	innerContent.Add(target)
	inner.SetContent(innerContent)
	outerContent.Add(inner)
	outer.SetContent(outerContent)
	scene.Root().Add(outer)
	scene.FocusWidget(target)
	for _, viewport := range []*ScrollView{inner, outer} {
		r, v := target.Bounds(), viewport.Bounds()
		if r.X < v.X || r.Y < v.Y || r.X+r.W > v.X+v.W || r.Y+r.H > v.Y+v.H {
			t.Fatalf("focused target %+v not revealed through %s %+v", r, viewport.ID(), v)
		}
	}
	ox, oy := outer.ScrollOffset()
	ix, iy := inner.ScrollOffset()
	scene.FocusWidget(target)
	ox2, oy2 := outer.ScrollOffset()
	ix2, iy2 := inner.ScrollOffset()
	if ox != ox2 || oy != oy2 || ix != ix2 || iy != iy2 {
		t.Fatal("repeated nested focus oscillated scroll offsets")
	}
	outer.ScrollTo(0, 0)
	inner.ScrollTo(0, 0)
	scene.FocusWidget(inner)
	if !scene.FocusNext(false) || scene.Focus() != target {
		t.Fatal("FocusNext did not reach nested off-viewport descendant")
	}
	for _, viewport := range []*ScrollView{inner, outer} {
		r, v := target.Bounds(), viewport.Bounds()
		if r.X < v.X || r.Y < v.Y || r.X+r.W > v.X+v.W || r.Y+r.H > v.Y+v.H {
			t.Fatal("FocusNext left descendant clipped")
		}
	}
}

func TestMouseFocusKeepsPartiallyVisibleEditorAtPointerCoordinates(t *testing.T) {
	scene := NewScene(nil)
	scroll := NewScrollView("viewport")
	scroll.SetBounds(Rect{W: 200, H: 100})
	content := scrollTestContent("content", 300, 500)
	editor := NewEditBox("multiline", ModeCustom)
	editor.SetMultiline(true)
	editor.SetText("one\ntwo\nthree\nfour")
	scrollTestPlace(editor, 20, 80, 140, 80)
	content.Add(editor)
	scroll.SetContent(content)
	scene.Root().Add(scroll)
	before := editor.Bounds()
	point := core.Point{X: before.X + 1, Y: before.Y + 10}
	if !scroll.Bounds().Contains(point.X, point.Y) || editor.caretAtPoint(point) != 0 {
		t.Fatal("mouse regression requires a visible first-line click")
	}
	scene.DispatchMouseDown(core.MouseEvent{Point: point, Button: core.MouseButtonLeft})
	scene.DispatchMouseUp(core.MouseEvent{Point: point, Button: core.MouseButtonLeft})
	if scene.Focus() != editor {
		t.Fatal("click did not focus the partially visible editor")
	}
	if x, y := scroll.ScrollOffset(); x != 0 || y != 0 || editor.Bounds() != before {
		t.Errorf("mouse focus moved target away from original pointer: offset=(%d,%d), before=%+v after=%+v", x, y, before, editor.Bounds())
	}
	if editor.caret != 0 {
		t.Errorf("mouse focus changed the clicked first line: caret=%d want=0", editor.caret)
	}
}

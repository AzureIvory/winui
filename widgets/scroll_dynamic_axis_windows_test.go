//go:build windows

package widgets

import (
	"strings"
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestWarmScrollContentMutationsUpdateBothRangesWithoutViewportResize(t *testing.T) {
	scene := NewScene(nil)
	scroll := NewScrollView("warm-scroll")
	scroll.SetHorizontalScroll(true)
	scene.Root().Add(scroll)
	content := NewPanel("live-content")
	content.SetLayout(ColumnLayout{})
	scroll.SetBounds(Rect{W: 100, H: 100})
	scroll.SetContent(content) // the production path attaches an empty panel first
	viewport := scroll.Bounds()
	check := func(wantX, wantY int32, last Widget) {
		t.Helper()
		scroll.ScrollTo(1<<30, 1<<30)
		x, y := scroll.ScrollOffset()
		if scroll.Bounds() != viewport || x != wantX || y != wantY {
			t.Errorf("warm content range without resize=(%d,%d) want=(%d,%d) viewport=%+v", x, y, wantX, wantY, scroll.Bounds())
		}
		if last != nil {
			bounds := last.Bounds()
			if bounds.X+bounds.W > viewport.X+viewport.W || bounds.Y+bounds.H > viewport.Y+viewport.H {
				t.Errorf("last content edge is unreachable: bounds=%+v viewport=%+v", bounds, viewport)
			}
		}
	}
	first := NewLabel("first", "first row")
	SetPreferredSize(first, core.Size{Width: 260, Height: 40})
	content.Add(first)
	check(160, 0, first)
	last := NewLabel("last", "last row")
	SetPreferredSize(last, core.Size{Width: 260, Height: 280})
	content.Add(last)
	check(160, 220, last)
	SetPreferredSize(first, core.Size{Width: 420, Height: 40})
	check(320, 220, last)
	content.Remove(last.ID())
	check(320, 0, first)
}

func wheelAxisMouse(t *testing.T, horizontal bool, delta int32, flags uintptr) core.MouseEvent {
	t.Helper()
	return core.MouseEvent{Point: core.Point{X: 30, Y: 30}, Horizontal: horizontal, Delta: delta, Flags: flags}
}

func TestWheelAxisPropagationAndScrollDirectionsPreserveExistingSemantics(t *testing.T) {
	for _, test := range []struct {
		name         string
		horizontal   bool
		delta        int32
		flags        uintptr
		wantX, wantY int32
	}{
		{"vertical-up", false, 120, 0x0011, 100, 52},
		{"vertical-down", false, -120, 0x0011, 100, 148},
		{"shift-vertical-right", false, -120, 0x0015, 148, 100},
		{"shift-vertical-left", false, 120, 0x0015, 52, 100},
		{"native-horizontal-right", true, 120, 0x0011, 148, 100},
		{"native-horizontal-left", true, -120, 0x0011, 52, 100},
		{"native-horizontal-hires-right", true, 30, 0x0011, 112, 100},
		{"native-horizontal-hires-left", true, -30, 0x0011, 88, 100},
		{"native-horizontal-shift-stays-right", true, 120, 0x0015, 148, 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			scene := NewScene(nil)
			scroll := NewScrollView("both-axes")
			scroll.SetHorizontalScroll(true)
			scroll.SetBounds(Rect{W: 100, H: 100})
			scroll.SetContent(scrollTestContent("wide-tall", 500, 500))
			scene.Root().Add(scroll)
			scroll.ScrollTo(100, 100)
			mouse := wheelAxisMouse(t, test.horizontal, test.delta, test.flags)
			converted := eventFromMouse(EventMouseWheel, mouse)
			if converted.Delta != test.delta || converted.Flags != test.flags {
				t.Error("scene conversion changed the original delta or raw mouse flags")
			}
			if converted.Horizontal != test.horizontal {
				t.Error("scene conversion did not preserve the independent wheel axis")
			}
			if !scene.DispatchMouseWheel(mouse) {
				t.Error("wheel event did not reach the scrollable scene target")
			}
			x, y := scroll.ScrollOffset()
			if x != test.wantX || y != test.wantY {
				t.Errorf("wheel moved wrong axis/direction: got=(%d,%d) want=(%d,%d)", x, y, test.wantX, test.wantY)
			}
		})
	}
}

func TestHorizontalWheelDoesNotBecomeVerticalInOtherConsumers(t *testing.T) {
	for _, kind := range []string{"list", "edit", "vertical-only-scroll"} {
		t.Run(kind, func(t *testing.T) {
			scene := NewScene(nil)
			outer := NewScrollView("horizontal-parent")
			outer.SetHorizontalScroll(true)
			outer.SetBounds(Rect{W: 100, H: 100})
			content := scrollTestContent("content", 500, 500)
			var child Widget
			var unchanged func() bool
			switch kind {
			case "list":
				list := NewListBox("inner-list")
				items := make([]ListItem, 30)
				for i := range items {
					items[i] = ListItem{Text: "row"}
				}
				list.SetItems(items)
				list.scroll = 4
				child, unchanged = list, func() bool { return list.scroll == 4 }
			case "edit":
				edit := NewEditBox("inner-edit", ModeCustom)
				edit.SetMultiline(true)
				edit.SetVerticalScroll(true)
				edit.SetText(strings.Repeat("line\n", 40))
				edit.scrollY = 60
				child, unchanged = edit, func() bool { return edit.scrollY == 60 }
			default:
				inner := NewScrollView("vertical-only")
				inner.SetContent(scrollTestContent("tall", 80, 500))
				inner.ScrollTo(0, 100)
				child = inner
				unchanged = func() bool { _, y := inner.ScrollOffset(); return y == 100 }
			}
			scrollTestPlace(child, 0, 0, 80, 80)
			content.Add(child)
			outer.SetContent(content)
			scene.Root().Add(outer)
			switch inner := child.(type) {
			case *ListBox:
				inner.scroll = 4
			case *EditBox:
				inner.scrollY = 60
			case *ScrollView:
				inner.ScrollTo(0, 100)
			}
			mouse := wheelAxisMouse(t, true, 120, 0x0015)
			if !scene.DispatchMouseWheel(mouse) {
				t.Error("horizontal event did not bubble to its capable parent")
			}
			x, y := outer.ScrollOffset()
			if !unchanged() || x != 48 || y != 0 {
				t.Errorf("horizontal wheel changed vertical consumer or failed parent horizontal scroll: outer=(%d,%d) child=%s", x, y, kind)
			}
		})
	}
}

func TestDefaultScrollbarPaletteSeparatesTrackThumbAndBackground(t *testing.T) {
	scroll := NewScrollView("default-scrollbar-palette")
	background := core.RGB(255, 255, 255)
	track, idle := scroll.scrollbarColors()
	scroll.hoverHorizontalThumb = true
	_, hover := scroll.scrollbarColors()
	scroll.dragAxis = scrollDragHorizontal
	_, drag := scroll.scrollbarColors()
	if track == background || idle == background || track == idle || hover == idle || drag == hover || drag == idle {
		t.Errorf("default scrollbar palette has indistinguishable states: background=%#x track=%#x idle=%#x hover=%#x drag=%#x", background, track, idle, hover, drag)
	}
}

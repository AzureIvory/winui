//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestFlexNaturalMeasurementFollowsVisibility(t *testing.T) {
	for _, test := range []struct {
		name   string
		layout Layout
		want   core.Size
	}{
		{"row", RowLayout{Gap: 6, Padding: UniformInsets(4)}, core.Size{Width: 94, Height: 38}},
		{"column", ColumnLayout{Gap: 6, Padding: UniformInsets(4)}, core.Size{Width: 48, Height: 74}},
		{"linear-row", LinearLayout{Axis: AxisHorizontal, Gap: 6, Padding: 4}, core.Size{Width: 94, Height: 38}},
		{"linear-column", LinearLayout{Axis: AxisVertical, Gap: 6, Padding: 4}, core.Size{Width: 48, Height: 74}},
	} {
		t.Run(test.name, func(t *testing.T) {
			panel := NewPanel("content")
			panel.SetLayout(test.layout)
			first, hidden, last := NewLabel("first", ""), NewLabel("hidden", ""), NewLabel("last", "")
			SetPreferredSize(first, core.Size{Width: 40, Height: 30})
			SetPreferredSize(hidden, core.Size{Width: 900, Height: 800})
			SetPreferredSize(last, core.Size{Width: 40, Height: 30})
			panel.AddAll(first, hidden, last)
			hidden.SetVisible(false)
			if got := MeasureNatural(panel); got != test.want {
				t.Errorf("hidden child added size or gap: got=%+v want=%+v", got, test.want)
			}
			if got := MeasureForWidth(panel, 100); got.Height != test.want.Height {
				t.Errorf("width-constrained measurement counted hidden child: got=%+v", got)
			}
			first.SetVisible(false)
			last.SetVisible(false)
			if got := MeasureNatural(panel); got != (core.Size{Width: 8, Height: 8}) {
				t.Errorf("all-hidden flex content should retain only padding: %+v", got)
			}
		})
	}
}

func TestWarmScrollHiddenLinearChildrenDoNotKeepRanges(t *testing.T) {
	scene := NewScene(nil)
	scroll := NewScrollView("scroll")
	scroll.SetHorizontalScroll(true)
	scroll.SetBounds(Rect{W: 100, H: 100})
	scene.Root().Add(scroll)
	content := NewPanel("content")
	content.SetLayout(ColumnLayout{})
	first, hidden, last := NewLabel("first", ""), NewLabel("hidden", ""), NewLabel("last", "")
	SetPreferredSize(first, core.Size{Width: 80, Height: 40})
	SetPreferredSize(hidden, core.Size{Width: 260, Height: 240})
	SetPreferredSize(last, core.Size{Width: 80, Height: 40})
	content.AddAll(first, hidden, last)
	scroll.SetContent(content)
	check := func(wantX, wantY int32) {
		t.Helper()
		scroll.ScrollTo(1<<30, 1<<30)
		x, y := scroll.ScrollOffset()
		if x != wantX || y != wantY {
			t.Fatalf("scroll range=(%d,%d) want=(%d,%d)", x, y, wantX, wantY)
		}
	}
	check(160, 220)
	hidden.SetVisible(false)
	if bounds := hidden.Bounds(); bounds.W != 260 || bounds.H != 240 {
		t.Fatalf("test needs a hidden widget with retained wide/tall geometry: %+v", bounds)
	}
	check(0, 0)
	if content.Bounds().W != 100 || content.Bounds().H != 100 {
		t.Fatalf("hidden geometry still expanded the content: %+v", content.Bounds())
	}
	hidden.SetVisible(true)
	check(160, 220)
}

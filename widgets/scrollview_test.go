//go:build windows

package widgets

import (
	"fmt"
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestScrollViewScrollRangeAndOffset(t *testing.T) {
	scroll := NewScrollView("scroll")
	scroll.SetBounds(Rect{W: 120, H: 80})

	content := NewPanel("content")
	content.SetLayout(ColumnLayout{Gap: 4})
	for i := 0; i < 6; i++ {
		label := NewLabel(fmt.Sprintf("row-%d", i), fmt.Sprintf("Row %d", i))
		label.SetBounds(Rect{W: 100, H: 28})
		content.Add(label)
	}

	scroll.SetContent(content)
	scroll.ScrollTo(0, 999)
	_, offsetY := scroll.ScrollOffset()
	if scroll.maxOffsetY <= 0 {
		t.Fatalf("expected positive vertical scroll range, got %d", scroll.maxOffsetY)
	}
	if offsetY != scroll.maxOffsetY {
		t.Fatalf("expected ScrollTo to clamp to max offset %d, got %d", scroll.maxOffsetY, offsetY)
	}
	if content.Bounds().Y != -scroll.maxOffsetY {
		t.Fatalf("expected content Y to track scroll offset, got %d", content.Bounds().Y)
	}
}

func TestScrollViewMouseWheelScrollsContent(t *testing.T) {
	scroll := NewScrollView("scroll-wheel")
	scroll.SetBounds(Rect{W: 120, H: 80})

	content := NewPanel("content-wheel")
	content.SetLayout(ColumnLayout{Gap: 4})
	for i := 0; i < 8; i++ {
		label := NewLabel(fmt.Sprintf("item-%d", i), fmt.Sprintf("Item %d", i))
		label.SetBounds(Rect{W: 100, H: 24})
		content.Add(label)
	}
	scroll.SetContent(content)

	if handled := scroll.OnEvent(Event{Type: EventMouseWheel, Delta: -120}); !handled {
		t.Fatalf("expected wheel event to be handled when content overflows")
	}
	_, offsetY := scroll.ScrollOffset()
	if offsetY <= 0 {
		t.Fatalf("expected mouse wheel to increase vertical offset, got %d", offsetY)
	}
}

func TestScrollViewHitTestClipsViewport(t *testing.T) {
	scene := newTestScene(Rect{X: 0, Y: 0, W: 240, H: 180})

	scroll := NewScrollView("scroll-hit")
	scroll.SetBounds(Rect{X: 20, Y: 20, W: 120, H: 80})

	content := NewPanel("content-hit")
	content.SetLayout(ColumnLayout{})

	spacer := NewLabel("spacer", "Spacer")
	spacer.SetBounds(Rect{W: 100, H: 100})
	hidden := NewButton("hidden", "Hidden", ModeCustom)
	hidden.SetBounds(Rect{W: 100, H: 32})
	content.Add(spacer)
	content.Add(hidden)

	scroll.SetContent(content)
	scene.Root().Add(scroll)

	point := core.Point{X: hidden.Bounds().X + 10, Y: hidden.Bounds().Y + hidden.Bounds().H/2}
	if scroll.Bounds().Contains(point.X, point.Y) {
		t.Fatalf("expected test point outside scroll viewport, got bounds=%#v point=%#v", scroll.Bounds(), point)
	}
	if raw := scene.hitTest(scene.root, point.X, point.Y); raw == hidden {
		t.Fatalf("expected clipped child not to win hit test outside viewport")
	}
}

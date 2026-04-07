//go:build windows

package widgets

import (
	"fmt"
	"testing"
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

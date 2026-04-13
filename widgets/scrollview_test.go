//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestScrollViewScrollbarRectsAppearForOverflowContent(t *testing.T) {
	scroll := NewScrollView("scroll")
	scroll.SetBounds(Rect{X: 20, Y: 30, W: 200, H: 120})

	content := NewPanel("content")
	SetPreferredSize(content, core.Size{Width: 320, Height: 480})
	scroll.SetContent(content)
	scroll.ScrollTo(0, 80)

	vTrack, vThumb, hTrack, hThumb := scroll.scrollbarRects(func(v int32) int32 { return v })
	if vTrack.W <= 0 || vTrack.H <= 0 {
		t.Fatalf("vertical track = %+v, want non-empty", vTrack)
	}
	if vThumb.W <= 0 || vThumb.H <= 0 {
		t.Fatalf("vertical thumb = %+v, want non-empty", vThumb)
	}
	if vThumb.Y < vTrack.Y || vThumb.Y+vThumb.H > vTrack.Y+vTrack.H {
		t.Fatalf("vertical thumb %+v should stay inside track %+v", vThumb, vTrack)
	}
	if hTrack.W != 0 || hTrack.H != 0 || hThumb.W != 0 || hThumb.H != 0 {
		t.Fatalf("horizontal scrollbars should stay hidden when horizontal scrolling is disabled: track=%+v thumb=%+v", hTrack, hThumb)
	}

	scroll.SetHorizontalScroll(true)
	scroll.ScrollTo(48, 80)
	_, _, hTrack, hThumb = scroll.scrollbarRects(func(v int32) int32 { return v })
	if hTrack.W <= 0 || hTrack.H <= 0 {
		t.Fatalf("horizontal track = %+v, want non-empty", hTrack)
	}
	if hThumb.W <= 0 || hThumb.H <= 0 {
		t.Fatalf("horizontal thumb = %+v, want non-empty", hThumb)
	}
	if hThumb.X < hTrack.X || hThumb.X+hThumb.W > hTrack.X+hTrack.W {
		t.Fatalf("horizontal thumb %+v should stay inside track %+v", hThumb, hTrack)
	}
}

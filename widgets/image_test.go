//go:build windows

package widgets

import (
	"github.com/AzureIvory/winui/core"
	"image"
	"image/color"
	"testing"
	"time"
)

// TestImageLoadBytes 测试图像字节加载�?func TestImageLoadBytes(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	bitmap, err := bitmapFromImage(img)
	if err != nil {
		t.Fatalf("bitmapFromImage failed: %v", err)
	}
	defer bitmap.Close()

	widget := NewImage("image")
	widget.SetBitmapOwned(bitmap)

	size := widget.NaturalSize()
	if size.Width != 8 || size.Height != 6 {
		t.Fatalf("unexpected natural size: %+v", size)
	}
}

// TestAnimatedImageTimerAdvance 测试动画图像定时推进�?func TestAnimatedImageTimerAdvance(t *testing.T) {
	anim := NewAnimatedImage("anim")
	anim.SetFrames([]core.AnimatedFrame{
		{Width: 10, Height: 10, Delay: 10 * time.Millisecond},
		{Width: 10, Height: 10, Delay: 10 * time.Millisecond},
	})
	anim.timerID = 7

	handled := anim.OnEvent(Event{Type: EventTimer, TimerID: 7})
	if !handled {
		t.Fatalf("expected timer event handled")
	}
	if anim.CurrentFrame() != 1 {
		t.Fatalf("expected frame index 1, got %d", anim.CurrentFrame())
	}
}

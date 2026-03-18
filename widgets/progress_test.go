//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

// TestProgressStyleDefaults 测试进度条默认样式。
func TestProgressStyleDefaults(t *testing.T) {
	progress := NewProgressBar("progress")
	style := progress.resolveStyle(nil)

	if style.FillColor != core.RGB(16, 185, 129) {
		t.Fatalf("expected default fill color updated, got %#08x", uint32(style.FillColor))
	}
	if style.BubbleColor != core.RGB(5, 150, 105) {
		t.Fatalf("expected default bubble color updated, got %#08x", uint32(style.BubbleColor))
	}
	if style.TextColor != core.RGB(255, 255, 255) {
		t.Fatalf("expected default text color updated, got %#08x", uint32(style.TextColor))
	}
}

// TestProgressStyleOverride 测试进度条样式覆盖。
func TestProgressStyleOverride(t *testing.T) {
	progress := NewProgressBar("progress")
	progress.Style = ProgressStyle{
		FillColor:   core.RGB(12, 34, 56),
		BubbleColor: core.RGB(65, 43, 21),
		TextColor:   core.RGB(240, 240, 240),
	}

	style := progress.resolveStyle(nil)
	if style.FillColor != core.RGB(12, 34, 56) {
		t.Fatalf("expected fill override, got %#08x", uint32(style.FillColor))
	}
	if style.BubbleColor != core.RGB(65, 43, 21) {
		t.Fatalf("expected bubble override, got %#08x", uint32(style.BubbleColor))
	}
	if style.TextColor != core.RGB(240, 240, 240) {
		t.Fatalf("expected text override, got %#08x", uint32(style.TextColor))
	}
}

// TestProgressDirtyRect 测试进度条脏区会覆盖百分比气泡。
func TestProgressDirtyRect(t *testing.T) {
	progress := NewProgressBar("progress")
	progress.SetBounds(Rect{X: 20, Y: 100, W: 200, H: 16})
	progress.SetStyle(ProgressStyle{ShowPercent: true})

	dirty := widgetDirtyRect(progress)
	if dirty.Y >= 100 {
		t.Fatalf("expected dirty rect to extend above progress bar, got %#v", dirty)
	}
	if dirty.W <= 200 {
		t.Fatalf("expected dirty rect width to include bubble sweep, got %#v", dirty)
	}
	if dirty.H <= 16 {
		t.Fatalf("expected dirty rect height to include bubble area, got %#v", dirty)
	}
}

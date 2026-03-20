//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

// TestButtonStateMachine 测试按钮状态机。
func TestButtonStateMachine(t *testing.T) {
	button := NewButton("button", "demo", ModeCustom)
	button.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 40})

	clicked := 0
	button.OnClick = func() {
		clicked++
	}

	button.OnEvent(Event{Type: EventMouseEnter})
	if !button.Hover {
		t.Fatalf("expected hover state after mouse enter")
	}

	button.OnEvent(Event{Type: EventMouseDown})
	if !button.Down {
		t.Fatalf("expected down state after mouse down")
	}

	button.OnEvent(Event{Type: EventMouseUp})
	if button.Down {
		t.Fatalf("expected down state cleared after mouse up")
	}

	button.OnEvent(Event{Type: EventClick})
	if clicked != 1 {
		t.Fatalf("expected click callback once, got %d", clicked)
	}

	button.OnEvent(Event{Type: EventMouseLeave})
	if button.Hover {
		t.Fatalf("expected hover state cleared after mouse leave")
	}
}

// TestButtonStyleOverride 测试按钮样式覆盖和布局模式。
func TestButtonStyleOverride(t *testing.T) {
	button := NewButton("button", "demo", ModeCustom)
	button.SetKind(BtnLeft)
	button.SetStyle(ButtonStyle{
		TextColor: core.RGB(1, 2, 3),
		DownText:  core.RGB(4, 5, 6),
		Pressed:   core.RGB(7, 8, 9),
		GapDP:     9,
		PadDP:     14,
	})

	style := button.resolveStyle(nil)
	if button.Kind() != BtnLeft {
		t.Fatalf("expected left button kind")
	}
	if style.TextColor != core.RGB(1, 2, 3) {
		t.Fatalf("expected text color override, got %#08x", uint32(style.TextColor))
	}
	if style.DownText != core.RGB(4, 5, 6) {
		t.Fatalf("expected down text override, got %#08x", uint32(style.DownText))
	}
	if style.Pressed != core.RGB(7, 8, 9) {
		t.Fatalf("expected pressed color override, got %#08x", uint32(style.Pressed))
	}
	if style.GapDP != 9 || style.PadDP != 14 {
		t.Fatalf("expected layout spacing override, got gap=%d pad=%d", style.GapDP, style.PadDP)
	}
}

// TestProgressClamp 测试进度值钳制。
func TestProgressClamp(t *testing.T) {
	progress := NewProgressBar("progress")
	progress.SetValue(120)
	if progress.Value() != 100 {
		t.Fatalf("expected clamp to 100, got %d", progress.Value())
	}

	progress.SetValue(-5)
	if progress.Value() != 0 {
		t.Fatalf("expected clamp to 0, got %d", progress.Value())
	}
}

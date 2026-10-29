//go:build windows

package core

import (
	"testing"
	"time"
)

func TestNativeWheelMessagesPreserveAxisDeltaFlagsAndClientPoint(t *testing.T) {
	events := make(chan MouseEvent, 8)
	app, err := NewApp(Options{Title: "WinUI isolated wheel decoding test", StartHidden: true, Width: 320, Height: 220,
		OnMouseWheel: func(_ *App, event MouseEvent) { events <- event }})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.Close(); app.Run() })
	for _, test := range []struct {
		name       string
		message    uint32
		delta      int32
		horizontal bool
	}{
		{"vertical-positive", 0x020A, 120, false},
		{"vertical-negative-hires", 0x020A, -30, false},
		{"horizontal-positive", 0x020E, 120, true},
		{"horizontal-negative-hires", 0x020E, -30, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			const flags = uintptr(0x0015) // unchanged MK_LBUTTON | MK_SHIFT | MK_MBUTTON
			screenPoint := Point{X: 333, Y: 444}
			wantPoint := app.screenToClient(screenPoint)
			wParam := flags | uintptr(uint16(int16(test.delta)))<<16
			lParam := uintptr(uint16(int16(screenPoint.X))) | uintptr(uint16(int16(screenPoint.Y)))<<16
			done := make(chan struct{})
			if err := app.Post(func() {
				defer close(done)
				// Send a real native message to this owned hidden HWND only.
				procSendMessageW.Call(uintptr(app.Handle()), uintptr(test.message), wParam, lParam)
			}); err != nil {
				t.Fatal(err)
			}
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("isolated native wheel message timed out")
			}
			select {
			case event := <-events:
				if event.Delta != test.delta || event.Flags != flags || event.Point != wantPoint {
					t.Errorf("native message lost signed delta/raw flags/client point: got=%+v want delta=%d flags=%#x point=%+v", event, test.delta, flags, wantPoint)
				}
				if event.Horizontal != test.horizontal {
					t.Errorf("native wheel callback has no correct independent horizontal axis: got=%+v want horizontal=%v", event, test.horizontal)
				}
			default:
				t.Errorf("native message %#x did not enter the existing OnMouseWheel callback", test.message)
			}
		})
	}
}

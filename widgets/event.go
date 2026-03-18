//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

type EventType int

const (
	EventMouseMove EventType = iota + 1
	EventMouseEnter
	EventMouseLeave
	EventMouseDown
	EventMouseUp
	EventMouseWheel
	EventClick
	EventFocus
	EventBlur
	EventKeyDown
	EventChar
	EventTimer
	EventPaint
	EventResize
)

type Event struct {
	Type    EventType
	Point   core.Point
	Button  core.MouseButton
	Key     core.KeyEvent
	Rune    rune
	Flags   uintptr
	Delta   int32
	TimerID uintptr
	Bounds  Rect
	Ctx     *PaintCtx
	Source  Widget
}

// eventFromMouse 将鼠标事件转换为控件事件。
func eventFromMouse(t EventType, ev core.MouseEvent) Event {
	return Event{
		Type:   t,
		Point:  ev.Point,
		Button: ev.Button,
		Flags:  ev.Flags,
		Delta:  ev.Delta,
	}
}

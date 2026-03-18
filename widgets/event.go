//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// EventType 表示场景和控件可处理的事件类别。
type EventType int

const (
	// EventMouseMove 表示鼠标移动事件。
	EventMouseMove EventType = iota + 1
	// EventMouseEnter 表示鼠标进入控件事件。
	EventMouseEnter
	// EventMouseLeave 表示鼠标离开控件事件。
	EventMouseLeave
	// EventMouseDown 表示鼠标按下事件。
	EventMouseDown
	// EventMouseUp 表示鼠标抬起事件。
	EventMouseUp
	// EventMouseWheel 表示鼠标滚轮事件。
	EventMouseWheel
	// EventClick 表示点击事件。
	EventClick
	// EventFocus 表示获得焦点事件。
	EventFocus
	// EventBlur 表示失去焦点事件。
	EventBlur
	// EventKeyDown 表示按键按下事件。
	EventKeyDown
	// EventChar 表示字符输入事件。
	EventChar
	// EventTimer 表示定时器事件。
	EventTimer
	// EventPaint 表示显式绘制事件。
	EventPaint
	// EventResize 表示尺寸变化事件。
	EventResize
)

// Event 封装控件事件分发时可能携带的上下文信息。
type Event struct {
	// Type 表示事件类型。
	Type EventType
	// Point 表示事件坐标。
	Point core.Point
	// Button 表示事件关联的鼠标按键。
	Button core.MouseButton
	// Key 表示键盘事件数据。
	Key core.KeyEvent
	// Rune 表示字符输入内容。
	Rune rune
	// Flags 保存原始消息标志。
	Flags uintptr
	// Delta 表示滚轮增量。
	Delta int32
	// TimerID 表示触发事件的定时器标识。
	TimerID uintptr
	// Bounds 表示与尺寸变化相关的矩形范围。
	Bounds Rect
	// Ctx 表示绘制事件使用的上下文。
	Ctx *PaintCtx
	// Source 表示事件来源控件。
	Source Widget
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

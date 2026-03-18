//go:build windows

package core

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procTrackMouseEvent = user32.NewProc("TrackMouseEvent")
	procSetCapture      = user32.NewProc("SetCapture")
	procReleaseCapture  = user32.NewProc("ReleaseCapture")
	procLoadCursorW     = user32.NewProc("LoadCursorW")
	procSetCursor       = user32.NewProc("SetCursor")
)

// trackMouseEvent 对应 Win32 的 TRACKMOUSEEVENT 结构。
type trackMouseEvent struct {
	// CbSize 表示结构体大小。
	CbSize uint32
	// DwFlags 表示跟踪选项。
	DwFlags uint32
	// HWndTrack 表示被跟踪的窗口。
	HWndTrack windows.Handle
	// DwHoverTime 表示悬停超时时间。
	DwHoverTime uint32
}

// MouseTarget 描述输入路由器可命中的目标控件。
type MouseTarget struct {
	// Bounds 返回目标的命中区域。
	Bounds func() Rect
	// Visible 返回目标是否可见。
	Visible func() bool
	// Enabled 返回目标是否可用。
	Enabled func() bool
	// Cursor 指定悬停时使用的光标。
	Cursor CursorID

	// OnEnter 在鼠标进入时触发。
	OnEnter func()
	// OnLeave 在鼠标离开时触发。
	OnLeave func()
	// OnMove 在鼠标移动时触发。
	OnMove func(MouseEvent)
	// OnDown 在鼠标按下时触发。
	OnDown func(MouseEvent)
	// OnUp 在鼠标抬起时触发。
	OnUp func(MouseEvent)
	// OnClick 在点击完成时触发。
	OnClick func(MouseEvent)
}

// InputRouter 负责把鼠标事件路由到多个目标。
type InputRouter struct {
	// app 指向关联的应用实例。
	app *App
	// targets 保存当前可路由目标集合。
	targets []*MouseTarget
	// hovered 保存当前悬停目标。
	hovered *MouseTarget
	// captured 保存当前捕获鼠标的目标。
	captured *MouseTarget
}

// NewInputRouter 创建一个新的输入路由器。
func NewInputRouter(app *App) *InputRouter {
	return &InputRouter{app: app}
}

// SetTargets 更新输入路由器的目标集合。
func (r *InputRouter) SetTargets(targets ...*MouseTarget) {
	r.targets = append(r.targets[:0], targets...)
	r.hovered = nil
	r.captured = nil
}

// HandleMove 处理输入路由器的移动事件。
func (r *InputRouter) HandleMove(ev MouseEvent) bool {
	if r == nil {
		return false
	}
	r.trackLeave()

	target := r.hit(ev.Point)
	changed := r.hovered != target
	if changed {
		if r.hovered != nil && r.hovered.OnLeave != nil {
			r.hovered.OnLeave()
		}
		r.hovered = target
		if r.hovered != nil && r.hovered.OnEnter != nil {
			r.hovered.OnEnter()
		}
	}
	if target != nil && target.OnMove != nil {
		target.OnMove(ev)
	}
	r.applyCursor(target)
	return changed
}

// HandleLeave 处理输入路由器的离开事件。
func (r *InputRouter) HandleLeave() bool {
	if r == nil || r.hovered == nil {
		r.applyCursor(nil)
		return false
	}
	if r.hovered.OnLeave != nil {
		r.hovered.OnLeave()
	}
	r.hovered = nil
	r.applyCursor(nil)
	return true
}

// HandleDown 处理输入路由器的按下事件。
func (r *InputRouter) HandleDown(ev MouseEvent) bool {
	if r == nil {
		return false
	}
	target := r.hit(ev.Point)
	if target == nil {
		return false
	}
	r.captured = target
	if r.app != nil {
		r.app.captureMouse()
	}
	if target.OnDown != nil {
		target.OnDown(ev)
	}
	return true
}

// HandleUp 处理输入路由器的抬起事件。
func (r *InputRouter) HandleUp(ev MouseEvent) bool {
	if r == nil {
		return false
	}
	if r.app != nil {
		r.app.releaseMouse()
	}
	if r.captured == nil {
		return false
	}

	target := r.captured
	r.captured = nil
	if target.OnUp != nil {
		target.OnUp(ev)
	}
	if target == r.hit(ev.Point) && target.OnClick != nil {
		target.OnClick(ev)
	}
	return true
}

// hit 命中测试输入路由器的目标。
func (r *InputRouter) hit(pt Point) *MouseTarget {
	for _, target := range r.targets {
		if target == nil || target.Bounds == nil {
			continue
		}
		if target.Visible != nil && !target.Visible() {
			continue
		}
		if target.Enabled != nil && !target.Enabled() {
			continue
		}
		if target.Bounds().Contains(pt.X, pt.Y) {
			return target
		}
	}
	return nil
}

// trackLeave 为输入路由器启用离开跟踪。
func (r *InputRouter) trackLeave() {
	if r == nil || r.app == nil || r.app.hwnd == 0 {
		return
	}
	event := trackMouseEvent{
		CbSize:    uint32(unsafe.Sizeof(trackMouseEvent{})),
		DwFlags:   tmeLeave,
		HWndTrack: r.app.hwnd,
	}
	procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&event)))
}

// applyCursor 为输入路由器应用光标。
func (r *InputRouter) applyCursor(target *MouseTarget) {
	cursor := CursorArrow
	if target != nil && target.Cursor != 0 {
		cursor = target.Cursor
	}
	if r.app != nil {
		r.app.setCursor(cursor)
	}
}

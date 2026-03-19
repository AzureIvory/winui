//go:build windows

package core

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// procTrackMouseEvent 指向 Win32 的 TrackMouseEvent 过程。
	procTrackMouseEvent = user32.NewProc("TrackMouseEvent")
	// procSetCapture 指向 Win32 的 SetCapture 过程。
	procSetCapture = user32.NewProc("SetCapture")
	// procReleaseCapture 指向 Win32 的 ReleaseCapture 过程。
	procReleaseCapture = user32.NewProc("ReleaseCapture")
	// procLoadCursorW 指向 Win32 的 LoadCursorW 过程。
	procLoadCursorW = user32.NewProc("LoadCursorW")
	// procSetCursor 指向 Win32 的 SetCursor 过程。
	procSetCursor = user32.NewProc("SetCursor")
)

// trackMouseEvent 映射 Win32 的 TRACKMOUSEEVENT 结构体。
type trackMouseEvent struct {
	// CbSize 表示结构体自身大小。
	CbSize uint32
	// DwFlags 表示跟踪标志。
	DwFlags uint32
	// HWndTrack 表示需要跟踪的窗口句柄。
	HWndTrack windows.Handle
	// DwHoverTime 表示悬停检测时间。
	DwHoverTime uint32
}

// trackMouseLeave 请求当前窗口接收 WM_MOUSELEAVE 消息。
func (a *App) trackMouseLeave() {
	if a == nil || a.hwnd == 0 {
		return
	}
	event := trackMouseEvent{
		CbSize:    uint32(unsafe.Sizeof(trackMouseEvent{})),
		DwFlags:   tmeLeave,
		HWndTrack: a.hwnd,
	}
	procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&event)))
}

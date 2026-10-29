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
	procSetCursor                = user32.NewProc("SetCursor")
	procGetKeyState              = user32.NewProc("GetKeyState")
	imm32                        = windows.NewLazySystemDLL("imm32.dll")
	procImmGetContext            = imm32.NewProc("ImmGetContext")
	procImmReleaseContext        = imm32.NewProc("ImmReleaseContext")
	procImmGetCompositionStringW = imm32.NewProc("ImmGetCompositionStringW")
	procImmGetVirtualKey         = imm32.NewProc("ImmGetVirtualKey")
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

func keyEvent(hwnd windows.Handle, key, flags uintptr) KeyEvent {
	virtualKey := uint32(key)
	if virtualKey == 0xE5 { // VK_PROCESSKEY wraps keys even when the IME is not composing.
		original, _, _ := procImmGetVirtualKey.Call(uintptr(hwnd))
		virtualKey = normalizeIMEVirtualKey(virtualKey, uint32(original))
	}
	return KeyEvent{
		Key: virtualKey, Handle: hwnd, Flags: flags,
		Ctrl: keyPressed(0x11), Shift: keyPressed(0x10), Alt: keyPressed(0x12),
		Repeat:    flags&(1<<30) != 0,
		Composing: virtualKey == 0xE5 || imeComposing(hwnd),
	}
}

func normalizeIMEVirtualKey(key, original uint32) uint32 {
	if key == 0xE5 && original != 0 && original != 0xE5 {
		return original
	}
	return key
}

func keyPressed(key uintptr) bool {
	state, _, _ := procGetKeyState.Call(key)
	return int16(uint16(state)) < 0
}

func imeComposing(hwnd windows.Handle) bool {
	context, _, _ := procImmGetContext.Call(uintptr(hwnd))
	if context == 0 {
		return false
	}
	defer procImmReleaseContext.Call(uintptr(hwnd), context)
	length, _, _ := procImmGetCompositionStringW.Call(context, 0x0008, 0, 0) // GCS_COMPSTR
	return int32(length) > 0
}

// previewKeyMessage runs before TranslateMessage so a consumed shortcut cannot
// insert a character in a native child. Native subclasses never preview again.
func (a *App) previewKeyMessage(message *msg) bool {
	if a.opts.OnKeyPreview == nil || (message.Message != wmKeyDown && message.Message != 0x0104) {
		return false
	}
	if message.HWnd != a.hwnd && !isChildWindow(a.hwnd, message.HWnd) {
		return false
	}
	ev := keyEvent(message.HWnd, message.WParam, message.LParam)
	consumed := a.opts.OnKeyPreview(a, ev)
	return !ev.Composing && consumed
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

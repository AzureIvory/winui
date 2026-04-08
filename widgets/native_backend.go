//go:build windows

package widgets

import (
	"sync"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

var (
	// nativeUser32 表示 user32.dll 的延迟加载句柄。
	nativeUser32 = windows.NewLazySystemDLL("user32.dll")
	// nativeGdi32 表示 gdi32.dll 的延迟加载句柄。
	nativeGdi32 = windows.NewLazySystemDLL("gdi32.dll")
)

var (
	// procCreateWindowExW 表示 CreateWindowExW 过程入口。
	procCreateWindowExW = nativeUser32.NewProc("CreateWindowExW")
	// procDestroyWindow 表示 DestroyWindow 过程入口。
	procDestroyWindow = nativeUser32.NewProc("DestroyWindow")
	// procSetWindowPos 表示 SetWindowPos 过程入口。
	procSetWindowPos = nativeUser32.NewProc("SetWindowPos")
	// procShowWindow 表示 ShowWindow 过程入口。
	procShowWindow = nativeUser32.NewProc("ShowWindow")
	// procEnableWindow 表示 EnableWindow 过程入口。
	procEnableWindow = nativeUser32.NewProc("EnableWindow")
	// procSetWindowTextW 表示 SetWindowTextW 过程入口。
	procSetWindowTextW = nativeUser32.NewProc("SetWindowTextW")
	// procGetWindowTextW 表示 GetWindowTextW 过程入口。
	procGetWindowTextW = nativeUser32.NewProc("GetWindowTextW")
	// procGetWindowTextLenW 表示 GetWindowTextLengthW 过程入口。
	procGetWindowTextLenW = nativeUser32.NewProc("GetWindowTextLengthW")
	// procSendMessageW 表示 SendMessageW 过程入口。
	procSendMessageW = nativeUser32.NewProc("SendMessageW")
	// procSetFocus 表示 SetFocus 过程入口。
	procSetFocus    = nativeUser32.NewProc("SetFocus")
	procGetKeyState = nativeUser32.NewProc("GetKeyState")
	// procSetWindowLongPtrW 表示 SetWindowLongPtrW 过程入口。
	procSetWindowLongPtrW = nativeUser32.NewProc("SetWindowLongPtrW")
	// procCallWindowProcW 表示 CallWindowProcW 过程入口。
	procCallWindowProcW = nativeUser32.NewProc("CallWindowProcW")
	// procGetStockObject 表示 GetStockObject 过程入口。
	procGetStockObject = nativeGdi32.NewProc("GetStockObject")
)

var (
	// nativeEditWndProc 保存编辑框子类化后使用的窗口过程回调。
	nativeEditWndProc = windows.NewCallback(nativeEditProc)
	// nativeEditRegistry 保存原生编辑框句柄到控件实例的映射。
	nativeEditRegistry sync.Map
)

// ControlMode 表示控件创建时选择的后端模式。
type ControlMode uint8

const (
	// ModeCustom 表示使用库内自绘控件后端。
	ModeCustom ControlMode = iota
	// ModeNative 表示使用原生系统子控件后端。
	ModeNative
)

const (
	// nativeWindowChild 表示创建子窗口。
	nativeWindowChild uint32 = 0x40000000
	// nativeWindowVisible 表示窗口初始可见。
	nativeWindowVisible uint32 = 0x10000000
	// nativeWindowTabStop 表示参与系统 Tab 焦点切换。
	nativeWindowTabStop uint32 = 0x00010000
	nativeWindowHScroll uint32 = 0x00100000
	// nativeWindowVScroll 表示启用垂直滚动条样式。
	nativeWindowVScroll uint32 = 0x00200000
	// nativeWindowBorder 表示启用标准边框。
	nativeWindowBorder uint32 = 0x00800000
)

const (
	// nativeButtonPush 表示普通按钮样式。
	nativeButtonPush uint32 = 0x00000000
	// nativeButtonCheckBox 表示自动复选框样式。
	nativeButtonCheckBox uint32 = 0x00000003
	// nativeButtonRadio 表示普通单选按钮样式。
	nativeButtonRadio uint32 = 0x00000004
)

const (
	// nativeEditMultiline ??????????
	nativeEditMultiline uint32 = 0x0004
	nativeEditPassword  uint32 = 0x0020
	// nativeEditAutoVScroll ?????????
	nativeEditAutoVScroll uint32 = 0x0040
	// nativeEditAutoHScroll ??????????????
	nativeEditAutoHScroll uint32 = 0x0080
	// nativeEditWantReturn ??????????
	nativeEditWantReturn uint32 = 0x1000
)

const (
	// nativeComboDropDownList 表示不可编辑的下拉列表组合框。
	nativeComboDropDownList uint32 = 0x0003
)

const (
	// nativeDefaultGUIFont 表示系统默认 GUI 字体对象。
	nativeDefaultGUIFont = 17
)

const (
	// nativeShowHide 表示隐藏窗口。
	nativeShowHide = 0
	// nativeShowShow 表示显示窗口。
	nativeShowShow = 5
)

const (
	// nativeWindowNoZOrder 表示移动窗口时保持 Z 序不变。
	nativeWindowNoZOrder = 0x0004
	// nativeWindowNoActivate 表示移动窗口时不激活窗口。
	nativeWindowNoActivate = 0x0010
)

const (
	// nativeMessageSetFont 表示为原生控件设置字体。
	nativeMessageSetFont = 0x0030
)

const (
	// nativeEditSetReadOnly 表示切换编辑框只读状态。
	nativeEditSetReadOnly = 0x00CF
	// nativeEditSetCueBanner 表示设置编辑框占位提示。
	nativeEditSetCueBanner = 0x1501
	nativeEditScrollCaret  = 0x00B7
)

const (
	// nativeButtonGetCheck 表示读取按钮勾选状态。
	nativeButtonGetCheck = 0x00F0
	// nativeButtonSetCheck 表示设置按钮勾选状态。
	nativeButtonSetCheck = 0x00F1
)

const (
	// nativeButtonStateUnchecked 表示未选中状态。
	nativeButtonStateUnchecked = 0
	// nativeButtonStateChecked 表示已选中状态。
	nativeButtonStateChecked = 1
)

const (
	// nativeComboResetContent 表示清空组合框内容。
	nativeComboResetContent = 0x014B
	// nativeComboAddString 表示向组合框追加字符串项。
	nativeComboAddString = 0x0143
	// nativeComboSetCurSel 表示设置组合框当前选择。
	nativeComboSetCurSel = 0x014E
	// nativeComboGetCurSel 表示读取组合框当前选择。
	nativeComboGetCurSel = 0x0147
)

const (
	// nativeButtonClicked 表示按钮点击通知。
	nativeButtonClicked uint16 = 0
	// nativeButtonSetFocus 表示按钮获得焦点通知。
	nativeButtonSetFocus uint16 = 6
	// nativeEditChanged 表示编辑框文本变化通知。
	nativeEditChanged uint16 = 0x0300
	// nativeEditSetFocus 表示编辑框获得焦点通知。
	nativeEditSetFocus uint16 = 0x0100
	// nativeComboSelectionChanged 表示组合框选项变化通知。
	nativeComboSelectionChanged uint16 = 1
	// nativeComboSetFocus 表示组合框获得焦点通知。
	nativeComboSetFocus uint16 = 3
)

const (
	// nativeWindowKeyDown 表示键盘按下消息。
	nativeWindowKeyDown uint32 = 0x0100
)

const (
	// nativeLongWndProc 表示窗口过程指针槽位索引。
	nativeLongWndProc       uintptr = ^uintptr(3)
	nativeVirtualKeyControl uintptr = 0x11
)

// nativeCommandHandler 表示可处理原生命令通知的内部接口。
type nativeCommandHandler interface {
	// handleNativeCommand 处理给定的通知码。
	handleNativeCommand(code uint16) bool
}

// nativeControlState 保存原生子控件的运行时状态。
type nativeControlState struct {
	// handle 保存当前原生子控件句柄。
	handle windows.Handle
	// commandID 保存分配给原生子控件的命令标识。
	commandID uint16
	// oldWndProc 保存子类化前的原始窗口过程指针。
	oldWndProc uintptr
}

// normalizeControlMode 将传入模式规整为受支持的枚举值。
func normalizeControlMode(mode ControlMode) ControlMode {
	if mode == ModeNative {
		return ModeNative
	}
	return ModeCustom
}

// isNativeMode 返回给定模式是否为原生后端。
func isNativeMode(mode ControlMode) bool {
	return normalizeControlMode(mode) == ModeNative
}

// valid 返回当前原生控件状态是否已经持有句柄。
func (s *nativeControlState) valid() bool {
	return s != nil && s.handle != 0
}

// HandleNativeCommand 将原生命令通知路由到已注册的控件实例。
func (s *Scene) HandleNativeCommand(evt core.CommandEvent) bool {
	if s == nil || evt.Handle == 0 {
		return false
	}
	s.nativeMu.RLock()
	handler := s.nativeTargets[uintptr(evt.Handle)]
	s.nativeMu.RUnlock()
	if handler == nil {
		return false
	}
	return handler.handleNativeCommand(evt.Code)
}

// allocateNativeCommandID 分配一个新的原生命令标识。
func (s *Scene) allocateNativeCommandID() uint16 {
	if s == nil {
		return 0
	}
	s.nativeMu.Lock()
	defer s.nativeMu.Unlock()
	s.nextNativeID++
	if s.nextNativeID == 0 {
		s.nextNativeID++
	}
	return s.nextNativeID
}

// registerNativeControl 将原生句柄注册到场景命令路由表。
func (s *Scene) registerNativeControl(handle windows.Handle, handler nativeCommandHandler) {
	if s == nil || handle == 0 || handler == nil {
		return
	}
	s.nativeMu.Lock()
	s.nativeTargets[uintptr(handle)] = handler
	s.nativeMu.Unlock()
}

// unregisterNativeControl 从场景命令路由表中移除原生句柄。
func (s *Scene) unregisterNativeControl(handle windows.Handle) {
	if s == nil || handle == 0 {
		return
	}
	s.nativeMu.Lock()
	delete(s.nativeTargets, uintptr(handle))
	s.nativeMu.Unlock()
}

// createNativeControl 在场景宿主窗口下创建原生子控件。
func createNativeControl(scene *Scene, className, text string, style uint32, bounds Rect, commandID uint16) (windows.Handle, error) {
	if scene == nil || scene.app == nil || scene.app.Handle() == 0 {
		return 0, core.ErrNotInitialized
	}

	classPtr, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	textPtr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return 0, err
	}

	handle, _, callErr := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(classPtr)),
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(style),
		uintptr(bounds.X),
		uintptr(bounds.Y),
		uintptr(bounds.W),
		uintptr(bounds.H),
		uintptr(scene.app.Handle()),
		uintptr(commandID),
		0,
		0,
	)
	if handle == 0 {
		return 0, callErr
	}

	child := windows.Handle(handle)
	applyNativeDefaultFont(child)
	return child, nil
}

// destroyNativeControl 销毁给定原生子控件。
func destroyNativeControl(handle windows.Handle) {
	if handle == 0 {
		return
	}
	procDestroyWindow.Call(uintptr(handle))
}

// setNativeBounds 同步原生子控件边界。
func setNativeBounds(handle windows.Handle, bounds Rect) {
	if handle == 0 {
		return
	}
	procSetWindowPos.Call(
		uintptr(handle),
		0,
		uintptr(bounds.X),
		uintptr(bounds.Y),
		uintptr(bounds.W),
		uintptr(bounds.H),
		nativeWindowNoZOrder|nativeWindowNoActivate,
	)
}

// setNativeVisible 同步原生子控件可见性。
func setNativeVisible(handle windows.Handle, visible bool) {
	if handle == 0 {
		return
	}
	cmd := nativeShowHide
	if visible {
		cmd = nativeShowShow
	}
	procShowWindow.Call(uintptr(handle), uintptr(cmd))
}

// setNativeEnabled 同步原生子控件启用状态。
func setNativeEnabled(handle windows.Handle, enabled bool) {
	if handle == 0 {
		return
	}
	flag := uintptr(0)
	if enabled {
		flag = 1
	}
	procEnableWindow.Call(uintptr(handle), flag)
}

// setNativeText 同步原生子控件文本。
func setNativeText(handle windows.Handle, text string) {
	if handle == 0 {
		return
	}
	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	procSetWindowTextW.Call(uintptr(handle), uintptr(unsafe.Pointer(ptr)))
}

// getNativeText 读取原生子控件当前文本。
func getNativeText(handle windows.Handle) string {
	if handle == 0 {
		return ""
	}
	length, _, _ := procGetWindowTextLenW.Call(uintptr(handle))
	buffer := make([]uint16, length+1)
	procGetWindowTextW.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	return windows.UTF16ToString(buffer)
}

// sendNativeMessage 向原生子控件发送窗口消息。
func sendNativeMessage(handle windows.Handle, message uint32, wParam, lParam uintptr) uintptr {
	if handle == 0 {
		return 0
	}
	result, _, _ := procSendMessageW.Call(uintptr(handle), uintptr(message), wParam, lParam)
	return result
}

// applyNativeDefaultFont 为原生子控件应用系统默认 GUI 字体。
func applyNativeDefaultFont(handle windows.Handle) {
	if handle == 0 {
		return
	}
	font, _, _ := procGetStockObject.Call(nativeDefaultGUIFont)
	if font == 0 {
		return
	}
	procSendMessageW.Call(uintptr(handle), nativeMessageSetFont, font, 1)
}

// setNativeCueBanner 设置原生编辑框的占位提示。
func setNativeCueBanner(handle windows.Handle, text string) {
	if handle == 0 {
		return
	}
	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	sendNativeMessage(handle, nativeEditSetCueBanner, 0, uintptr(unsafe.Pointer(ptr)))
}

// setNativeReadOnly 设置原生编辑框的只读状态。
func setNativeReadOnly(handle windows.Handle, readOnly bool) {
	if handle == 0 {
		return
	}
	value := uintptr(0)
	if readOnly {
		value = 1
	}
	sendNativeMessage(handle, nativeEditSetReadOnly, value, 0)
}

// setNativeFocus 将系统焦点切换到给定原生子控件。
func setNativeFocus(handle windows.Handle) {
	if handle == 0 {
		return
	}
	procSetFocus.Call(uintptr(handle))
}

// subclassNativeEdit 为原生编辑框安装用于处理回车提交的子类窗口过程。
func subclassNativeEdit(edit *EditBox) {
	if edit == nil || !edit.native.valid() || edit.native.oldWndProc != 0 {
		return
	}
	nativeEditRegistry.Store(edit.native.handle, edit)
	oldProc, _, _ := procSetWindowLongPtrW.Call(
		uintptr(edit.native.handle),
		nativeLongWndProc,
		nativeEditWndProc,
	)
	edit.native.oldWndProc = oldProc
}

// unsubclassNativeEdit 还原原生编辑框的原始窗口过程。
func unsubclassNativeEdit(edit *EditBox) {
	if edit == nil || !edit.native.valid() {
		return
	}
	nativeEditRegistry.Delete(edit.native.handle)
	if edit.native.oldWndProc != 0 {
		procSetWindowLongPtrW.Call(
			uintptr(edit.native.handle),
			nativeLongWndProc,
			edit.native.oldWndProc,
		)
		edit.native.oldWndProc = 0
	}
}

func keyHasCtrlState() bool {
	state, _, _ := procGetKeyState.Call(nativeVirtualKeyControl)
	return int16(uint16(state)) < 0
}

// nativeEditProc 处理原生编辑框子类化后的窗口消息。
func nativeEditProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	value, ok := nativeEditRegistry.Load(windows.Handle(hwnd))
	if !ok {
		return 0
	}
	edit, _ := value.(*EditBox)
	if edit == nil {
		return 0
	}
	if msg == nativeWindowKeyDown && uint32(wParam) == core.KeyReturn {
		shouldSubmit := !edit.multiline || !edit.acceptReturn || keyHasCtrlState()
		if shouldSubmit {
			if edit.OnSubmit != nil {
				edit.OnSubmit(edit.TextValue())
			}
			return 0
		}
	}
	if edit.native.oldWndProc == 0 {
		return 0
	}
	result, _, _ := procCallWindowProcW.Call(edit.native.oldWndProc, hwnd, uintptr(msg), wParam, lParam)
	return result
}

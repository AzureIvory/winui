//go:build windows

package widgets

import (
	"sync"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

var (
	nativeUser32   = windows.NewLazySystemDLL("user32.dll")
	nativeGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	nativeMsftedit = windows.NewLazySystemDLL("Msftedit.dll")
)

var (
	procCreateWindowExW   = nativeUser32.NewProc("CreateWindowExW")
	procDestroyWindow     = nativeUser32.NewProc("DestroyWindow")
	procSetWindowPos      = nativeUser32.NewProc("SetWindowPos")
	procShowWindow        = nativeUser32.NewProc("ShowWindow")
	procEnableWindow      = nativeUser32.NewProc("EnableWindow")
	procSetWindowTextW    = nativeUser32.NewProc("SetWindowTextW")
	procGetWindowTextW    = nativeUser32.NewProc("GetWindowTextW")
	procGetWindowTextLenW = nativeUser32.NewProc("GetWindowTextLengthW")
	procSendMessageW      = nativeUser32.NewProc("SendMessageW")
	procInvalidateRect    = nativeUser32.NewProc("InvalidateRect")
	procSetFocus          = nativeUser32.NewProc("SetFocus")
	procGetFocus          = nativeUser32.NewProc("GetFocus")
	procGetKeyState       = nativeUser32.NewProc("GetKeyState")
	procSetWindowLongW    = nativeUser32.NewProc("SetWindowLongW")
	procSetWindowLongPtrW = nativeUser32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW   = nativeUser32.NewProc("CallWindowProcW")
	procGetStockObject    = nativeGdi32.NewProc("GetStockObject")
	procLoadCursorW       = nativeUser32.NewProc("LoadCursorW")
	procSetCursor         = nativeUser32.NewProc("SetCursor")
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
	// nativeEditMultiline 表示多行编辑样式。
	nativeEditMultiline uint32 = 0x0004
	// nativeEditAutoVScroll 表示自动垂直滚动。
	nativeEditAutoVScroll uint32 = 0x0040
	// nativeEditAutoHScroll 表示自动水平滚动。
	nativeEditAutoHScroll uint32 = 0x0080
	// nativeEditWantReturn 表示回车写入文本。
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
	// nativeEditSetPasswordChar 表示切换密码掩码字符。
	nativeEditSetPasswordChar = 0x00CC
	// nativeEditSetReadOnly 表示切换编辑框只读状态。
	nativeEditSetReadOnly = 0x00CF
	// nativeEditGetSel 表示读取选区。
	nativeEditGetSel = 0x00B0
	// nativeEditSetSel 表示设置选区。
	nativeEditSetSel = 0x00B1
	// nativeEditScrollCaret 表示将光标滚动到可见区域。
	nativeEditScrollCaret = 0x00B7
	// nativeEditSetCueBanner 表示设置编辑框占位提示。
	nativeEditSetCueBanner = 0x1501
	// nativeRichEditSetEventMask 表示配置 RichEdit 事件掩码。
	nativeRichEditSetEventMask = 0x0445
	// nativeRichEditSetBackgroundColor 表示设置 RichEdit 背景色。
	nativeRichEditSetBackgroundColor = 0x0443
	// nativeRichEditSetCharFormat 表示设置 RichEdit 字符格式。
	nativeRichEditSetCharFormat = 0x0444
	// nativeRichEditSetTextMode 表示设置 RichEdit 文本模式。
	nativeRichEditSetTextMode = 0x045C
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
	// nativeEditSetFocus 表示编辑框获得焦点通知。
	nativeEditSetFocus uint16 = 0x0100
	// nativeEditKillFocus 表示编辑框失去焦点通知。
	nativeEditKillFocus uint16 = 0x0200
	// nativeEditChanged 表示编辑框文本变化通知。
	nativeEditChanged uint16 = 0x0300
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

const (
	nativeRichEditClass = "RICHEDIT50W"
	// nativeRichEditTextModePlainText 表示纯文本模式。
	nativeRichEditTextModePlainText uintptr = 0x0001
	// nativeRichEditTextModeMultiLevelUndo 表示启用多级撤销。
	nativeRichEditTextModeMultiLevelUndo uintptr = 0x0004
	// nativeRichEditTextModeMultiCodepage 表示允许多代码页。
	nativeRichEditTextModeMultiCodepage uintptr = 0x0010
	// nativeRichEditEventMaskChange 表示订阅 EN_CHANGE。
	nativeRichEditEventMaskChange uintptr = 0x00000001
	// nativeRichEditSCFDefault 表示默认字符格式。
	nativeRichEditSCFDefault uintptr = 0
	// nativeRichEditSCFAll 表示应用到全部文本。
	nativeRichEditSCFAll uintptr = 4
	// nativeRichEditCFMColor 表示文本颜色字段有效。
	nativeRichEditCFMColor uint32 = 0x40000000
)
const (
	nativeWindowSetCursor uint32  = 0x0020
	nativeEditSetMargins  uint32  = 0x00D3
	nativeEditLeftMargin  uintptr = 0x0001
	nativeEditRightMargin uintptr = 0x0002
	nativeEditSetRectNP   uint32  = 0x00B4
)

type nativeRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type nativeCharFormatW struct {
	CbSize         uint32
	Mask           uint32
	Effects        uint32
	Height         int32
	Offset         int32
	TextColor      uint32
	CharSet        byte
	PitchAndFamily byte
	FaceName       [32]uint16
}

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

func createNativeRichEditControl(scene *Scene, style uint32, bounds Rect, commandID uint16) (windows.Handle, error) {
	if err := nativeMsftedit.Load(); err != nil {
		return 0, err
	}
	handle, err := createNativeControl(scene, nativeRichEditClass, "", style, bounds, commandID)
	if err != nil {
		return 0, err
	}
	sendNativeMessage(handle, nativeRichEditSetTextMode, nativeRichEditTextModePlainText|nativeRichEditTextModeMultiLevelUndo|nativeRichEditTextModeMultiCodepage, 0)
	sendNativeMessage(handle, nativeRichEditSetEventMask, 0, nativeRichEditEventMaskChange)
	return handle, nil
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

func setNativePassword(handle windows.Handle, enabled bool) {
	if handle == 0 {
		return
	}
	ch := uintptr(0)
	if enabled {
		ch = uintptr('•')
	}
	sendNativeMessage(handle, nativeEditSetPasswordChar, ch, 0)
	procInvalidateRect.Call(uintptr(handle), 0, 1)
}

func setNativeSelection(handle windows.Handle, start, end int) {
	if handle == 0 {
		return
	}
	sendNativeMessage(handle, nativeEditSetSel, uintptr(start), uintptr(end))
}

func getNativeSelection(handle windows.Handle) (int, int) {
	if handle == 0 {
		return 0, 0
	}
	var start uint32
	var end uint32
	sendNativeMessage(handle, nativeEditGetSel, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return int(start), int(end)
}

func setNativeRichEditBackgroundColor(handle windows.Handle, color core.Color) {
	if handle == 0 {
		return
	}
	sendNativeMessage(handle, nativeRichEditSetBackgroundColor, 0, uintptr(color))
}

func setNativeRichEditTextColor(handle windows.Handle, color core.Color) {
	if handle == 0 {
		return
	}
	cf := nativeCharFormatW{
		CbSize:    uint32(unsafe.Sizeof(nativeCharFormatW{})),
		Mask:      nativeRichEditCFMColor,
		TextColor: uint32(color),
	}
	sendNativeMessage(handle, nativeRichEditSetCharFormat, nativeRichEditSCFDefault, uintptr(unsafe.Pointer(&cf)))
	sendNativeMessage(handle, nativeRichEditSetCharFormat, nativeRichEditSCFAll, uintptr(unsafe.Pointer(&cf)))
}

// setNativeFocus 将系统焦点切换到给定原生子控件。
func setNativeFocus(handle windows.Handle) {
	if handle == 0 {
		return
	}
	procSetFocus.Call(uintptr(handle))
}

func nativeHasFocus(handle windows.Handle) bool {
	if handle == 0 {
		return false
	}
	focused, _, _ := procGetFocus.Call()
	return windows.Handle(focused) == handle
}

func setWindowLong(pointerSize uintptr) *windows.LazyProc {
	if pointerSize == 4 {
		return procSetWindowLongW
	}
	return procSetWindowLongPtrW
}

func callSetWindowLong(handle windows.Handle, value uintptr) uintptr {
	oldProc, _, _ := setWindowLong(unsafe.Sizeof(uintptr(0))).Call(
		uintptr(handle),
		nativeLongWndProc,
		value,
	)
	return oldProc
}

// subclassNativeEdit 为原生编辑框安装用于处理回车提交的子类窗口过程。
func subclassNativeEdit(edit *EditBox) {
	if edit == nil || !edit.native.valid() || edit.native.oldWndProc != 0 {
		return
	}
	nativeEditRegistry.Store(edit.native.handle, edit)
	edit.native.oldWndProc = callSetWindowLong(edit.native.handle, nativeEditWndProc)
}

// unsubclassNativeEdit 还原原生编辑框的原始窗口过程。
func unsubclassNativeEdit(edit *EditBox) {
	if edit == nil || !edit.native.valid() {
		return
	}
	nativeEditRegistry.Delete(edit.native.handle)
	if edit.native.oldWndProc != 0 {
		callSetWindowLong(edit.native.handle, edit.native.oldWndProc)
		edit.native.oldWndProc = 0
	}
}

func keyHasCtrlState() bool {
	state, _, _ := procGetKeyState.Call(nativeVirtualKeyControl)
	return int16(uint16(state)) < 0
}

// nativeEditProc 处理原生编辑框子类化后的窗口消息。
func nativeEditProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == nativeWindowSetCursor {
		h, _, _ := procLoadCursorW.Call(0, uintptr(core.CursorIBeam))
		if h != 0 {
			procSetCursor.Call(h)
			return 1
		}
	}
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
func setNativeEditMargins(handle windows.Handle, left, right int32) {
	if handle == 0 {
		return
	}
	lr := uintptr(uint32(left)&0xFFFF | (uint32(right) & 0xFFFF << 16))
	sendNativeMessage(
		handle,
		nativeEditSetMargins,
		nativeEditLeftMargin|nativeEditRightMargin,
		lr,
	)
}
func setNativeEditRect(handle windows.Handle, rc nativeRect) {
	if handle == 0 {
		return
	}
	sendNativeMessage(handle, nativeEditSetRectNP, 0, uintptr(unsafe.Pointer(&rc)))
	procInvalidateRect.Call(uintptr(handle), 0, 1)
}

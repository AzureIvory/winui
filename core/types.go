//go:build windows

package core

import "golang.org/x/sys/windows"

// Color 表示 Win32 常用的 0x00BBGGRR 颜色值。
type Color uint32

// CursorID 表示系统预定义光标的标识。
type CursorID uintptr

// FontQuality 表示 GDI 字体平滑质量选项。
type FontQuality byte

// DPIAwareness 表示进程或线程的 DPI 感知级别。
type DPIAwareness int

// MouseButton 表示鼠标按键类型。
type MouseButton uint8

// RenderMode 表示应用请求的渲染模式。
type RenderMode uint8

// RenderBackend 表示窗口当前实际启用的绘制后端。
type RenderBackend uint8

// Point 表示二维平面中的整数坐标点。
type Point struct {
	// X 表示水平坐标。
	X int32
	// Y 表示垂直坐标。
	Y int32
}

// Size 表示宽高尺寸。
type Size struct {
	// Width 表示宽度。
	Width int32
	// Height 表示高度。
	Height int32
}

// Rect 表示以左上角和宽高定义的矩形区域。
type Rect struct {
	// X 表示左上角横坐标。
	X int32
	// Y 表示左上角纵坐标。
	Y int32
	// W 表示矩形宽度。
	W int32
	// H 表示矩形高度。
	H int32
}

// DPIInfo 保存当前窗口或进程的 DPI 缩放信息。
type DPIInfo struct {
	// X 表示水平 DPI。
	X int32
	// Y 表示垂直 DPI。
	Y int32
	// Scale 表示相对 96 DPI 的缩放比例。
	Scale float64
	// Awareness 表示当前 DPI 感知级别。
	Awareness DPIAwareness
}

// MouseEvent 描述鼠标位置、按键和滚轮等输入信息。
type MouseEvent struct {
	// Point 表示事件发生时的客户区坐标。
	Point Point
	// Button 表示本次事件关联的按键。
	Button MouseButton
	// Flags 保存原始 Win32 鼠标标志位。
	Flags uintptr
	// Delta 表示滚轮事件的滚动增量。
	Delta int32
}

// KeyEvent 描述键盘按键与底层消息标志。
type KeyEvent struct {
	// Key 表示虚拟键码。
	Key uint32
	// Flags 保存原始消息中的位标志。
	Flags uintptr
}

// Options 定义创建 App 时使用的窗口参数和事件回调。
type Options struct {
	// ClassName 指定 Win32 窗口类名。
	ClassName string
	// Title 指定窗口标题。
	Title string
	// Width 指定初始客户区宽度。
	Width int32
	// Height 指定初始客户区高度。
	Height int32
	// Style 指定窗口样式标志。
	Style uint32
	// ExStyle 指定窗口扩展样式标志。
	ExStyle uint32
	// Cursor 指定默认光标。
	Cursor CursorID
	// Icon 指定窗口图标。
	Icon *Icon
	// Background 指定默认背景色。
	Background Color
	// DoubleBuffered 控制是否启用 GDI 双缓冲。
	DoubleBuffered bool
	// RenderMode 控制绘制后端选择。Auto 会优先尝试 Direct2D，失败时回退到 GDI。
	RenderMode RenderMode

	// OnCreate 在窗口创建完成后触发。
	OnCreate func(*App) error
	// OnPaint 在窗口需要重绘时触发。
	OnPaint func(*App, *Canvas)
	// OnResize 在客户区尺寸变化时触发。
	OnResize func(*App, Size)
	// OnMouseMove 在鼠标移动时触发。
	OnMouseMove func(*App, MouseEvent)
	// OnMouseLeave 在鼠标离开窗口时触发。
	OnMouseLeave func(*App)
	// OnMouseDown 在鼠标按下时触发。
	OnMouseDown func(*App, MouseEvent)
	// OnMouseUp 在鼠标抬起时触发。
	OnMouseUp func(*App, MouseEvent)
	// OnMouseWheel 在鼠标滚轮滚动时触发。
	OnMouseWheel func(*App, MouseEvent)
	// OnKeyDown 在按键按下时触发。
	OnKeyDown func(*App, KeyEvent)
	// OnChar 在字符输入时触发。
	OnChar func(*App, rune)
	// OnFocus 在窗口焦点变化时触发。
	OnFocus func(*App, bool)
	// OnTimer 在窗口定时器触发时调用。
	OnTimer func(*App, uintptr)
	// OnDPIChanged 在窗口 DPI 变化时触发。
	OnDPIChanged func(*App, DPIInfo)
	// OnDestroy 在窗口销毁前触发。
	OnDestroy func(*App)
}

const (
	// DPIAwarenessUnknown 表示尚未确定 DPI 感知模式。
	DPIAwarenessUnknown DPIAwareness = iota
	// DPIAwarenessSystem 表示系统级 DPI 感知。
	DPIAwarenessSystem
	// DPIAwarenessPerMonitor 表示按显示器感知 DPI。
	DPIAwarenessPerMonitor
	// DPIAwarenessPerMonitorV2 表示按显示器感知 DPI V2。
	DPIAwarenessPerMonitorV2
)

const (
	// RenderModeAuto 优先尝试 Direct2D，初始化或运行失败时自动回退到 GDI。
	RenderModeAuto RenderMode = iota + 1
	// RenderModeGDI 强制使用 GDI。
	RenderModeGDI
)

const (
	// RenderBackendUnknown 表示后端尚未初始化。
	RenderBackendUnknown RenderBackend = iota
	// RenderBackendGDI 表示当前使用 GDI 绘制。
	RenderBackendGDI
	// RenderBackendDirect2D 表示当前使用 Direct2D/DirectWrite/WIC 绘制。
	RenderBackendDirect2D
)

const (
	// MouseButtonLeft 表示鼠标左键。
	MouseButtonLeft MouseButton = 1
	// MouseButtonRight 表示鼠标右键。
	MouseButtonRight MouseButton = 2
)

const (
	// WSCaption 启用标准标题栏。
	WSCaption uint32 = 0x00C00000
	// WSSysMenu 启用系统菜单。
	WSSysMenu uint32 = 0x00080000
	// WSMinimizeBox 启用最小化按钮。
	WSMinimizeBox uint32 = 0x00020000
	// WSThickFrame 启用可调整大小的边框。
	WSThickFrame uint32 = 0x00040000
	// WSMaximizeBox 启用最大化按钮。
	WSMaximizeBox uint32 = 0x00010000
	// WSClipSiblings 避免同级窗口互相覆盖绘制。
	WSClipSiblings uint32 = 0x04000000
	// WSClipChildren 避免父窗口覆盖子窗口绘制。
	WSClipChildren uint32 = 0x02000000

	// WSExAppWindow 强制窗口显示在任务栏。
	WSExAppWindow uint32 = 0x00040000
	// WSExLayered 启用分层窗口能力。
	WSExLayered uint32 = 0x00080000
)

const (
	// DefaultWindowStyle 表示默认窗口样式组合。
	DefaultWindowStyle = WSCaption | WSSysMenu | WSMinimizeBox | WSClipChildren | WSClipSiblings
	// DefaultWindowExStyle 表示默认扩展窗口样式组合。
	DefaultWindowExStyle = WSExAppWindow
)

const (
	// CursorArrow 表示标准箭头光标。
	CursorArrow CursorID = 32512
	// CursorIBeam 表示文本输入光标。
	CursorIBeam CursorID = 32513
	// CursorHand 表示手型链接光标。
	CursorHand CursorID = 32649
)

const (
	// DTCenter 让文本在水平方向居中。
	DTCenter uint32 = 0x00000001
	// DTVCenter 让文本在垂直方向居中。
	DTVCenter uint32 = 0x00000004
	// DTSingleLine 强制文本单行绘制。
	DTSingleLine uint32 = 0x00000020
	// DTEndEllipsis 在末尾显示省略号。
	DTEndEllipsis uint32 = 0x00008000
)

const (
	// MessageBoxOKCancel 表示确定和取消按钮组合。
	MessageBoxOKCancel uint32 = 0x00000001
	// MessageBoxRetryCancel 表示重试和取消按钮组合。
	MessageBoxRetryCancel uint32 = 0x00000005
)

const (
	// MessageBoxResultOK 表示用户点击了确定。
	MessageBoxResultOK = 1
	// MessageBoxResultRetry 表示用户点击了重试。
	MessageBoxResultRetry = 4
)

const (
	// FontQualityAntialiased 表示普通抗锯齿字体质量。
	FontQualityAntialiased FontQuality = 4
	// FontQualityClearType 表示 ClearType 字体质量。
	FontQualityClearType FontQuality = 5
)

const (
	// drawTextAutoLen 让 DrawTextW 自动计算字符串长度。
	drawTextAutoLen = uintptr(^uint32(0))

	// showWindowNormal 以正常状态显示窗口。
	showWindowNormal = 5

	// wmDestroy 表示窗口销毁消息。
	wmDestroy = 0x0002
	// wmSize 表示客户区尺寸变化消息。
	wmSize = 0x0005
	// wmSetFocus 表示窗口获得焦点消息。
	wmSetFocus = 0x0007
	// wmKillFocus 表示窗口失去焦点消息。
	wmKillFocus = 0x0008
	// wmPaint 表示窗口重绘消息。
	wmPaint = 0x000F
	// wmClose 表示窗口关闭请求消息。
	wmClose = 0x0010
	// wmSetIcon 表示设置窗口图标消息。
	wmSetIcon = 0x0080
	// wmKeyDown 表示按键按下消息。
	wmKeyDown = 0x0100
	// wmChar 表示字符输入消息。
	wmChar = 0x0102
	// wmTimer 表示定时器消息。
	wmTimer = 0x0113
	// wmMouseMove 表示鼠标移动消息。
	wmMouseMove = 0x0200
	// wmLButtonDown 表示鼠标左键按下消息。
	wmLButtonDown = 0x0201
	// wmLButtonUp 表示鼠标左键抬起消息。
	wmLButtonUp = 0x0202
	// wmRButtonDown 表示鼠标右键按下消息。
	wmRButtonDown = 0x0204
	// wmRButtonUp 表示鼠标右键抬起消息。
	wmRButtonUp = 0x0205
	// wmMouseWheel 表示鼠标滚轮消息。
	wmMouseWheel = 0x020A
	// wmMouseLeave 表示鼠标离开窗口消息。
	wmMouseLeave = 0x02A3
	// wmDPICHanged 表示 DPI 变化消息。
	wmDPICHanged = 0x02E0
	// wmApp 表示应用自定义消息起始值。
	wmApp = 0x8000
	// wmNcCreate 表示非客户区创建消息。
	wmNcCreate = 0x0081
	// wmNcDestroy 表示非客户区销毁消息。
	wmNcDestroy = 0x0082
	// wmAppInvoke 表示执行投递回调的自定义消息。
	wmAppInvoke = wmApp + 0x240

	// iconSmall 表示小图标槽位。
	iconSmall = 0
	// iconBig 表示大图标槽位。
	iconBig = 1

	// psSolid 表示实线画笔样式。
	psSolid = 0

	// srccopy 表示位块传输的直接复制模式。
	srccopy = 0x00CC0020

	// bkModeTransparent 表示文本背景透明模式。
	bkModeTransparent = 1

	// tmeLeave 表示请求鼠标离开跟踪。
	tmeLeave = 0x00000002

	// diNormal 表示图标正常绘制模式。
	diNormal = 0x0003

	// acSrcOver 表示 AlphaBlend 的普通叠加模式。
	acSrcOver = 0x00
	// acSrcAlpha 表示使用源 Alpha 通道。
	acSrcAlpha = 0x01

	// logPixelsX 表示设备水平 DPI 项。
	logPixelsX = 88
	// logPixelsY 表示设备垂直 DPI 项。
	logPixelsY = 90

	// swpNoZOrder 表示保持原有 Z 序。
	swpNoZOrder = 0x0004
	// swpNoActivate 表示调整窗口时不激活。
	swpNoActivate = 0x0010

	// spiGetWorkArea 表示查询桌面工作区。
	spiGetWorkArea = 0x0030
)

const (
	// KeyBack 表示退格键。
	KeyBack uint32 = 0x08
	// KeyTab 表示 Tab 键。
	KeyTab uint32 = 0x09
	// KeyReturn 表示回车键。
	KeyReturn uint32 = 0x0D
	// KeyEscape 表示 Esc 键。
	KeyEscape uint32 = 0x1B
	// KeySpace 表示空格键。
	KeySpace uint32 = 0x20
	// KeyHome 表示 Home 键。
	KeyHome uint32 = 0x24
	// KeyLeft 表示左方向键。
	KeyLeft uint32 = 0x25
	// KeyUp 表示上方向键。
	KeyUp uint32 = 0x26
	// KeyRight 表示右方向键。
	KeyRight uint32 = 0x27
	// KeyDown 表示下方向键。
	KeyDown uint32 = 0x28
	// KeyEnd 表示 End 键。
	KeyEnd uint32 = 0x23
	// KeyDelete 表示 Delete 键。
	KeyDelete uint32 = 0x2E
)

// wndClassEx 对应 Win32 的 WNDCLASSEXW 结构。
type wndClassEx struct {
	// CbSize 表示结构体自身大小。
	CbSize uint32
	// Style 表示窗口类样式。
	Style uint32
	// LpfnWndProc 表示窗口过程回调地址。
	LpfnWndProc uintptr
	// CbClsExtra 表示额外类内存大小。
	CbClsExtra int32
	// CbWndExtra 表示额外窗口内存大小。
	CbWndExtra int32
	// HInstance 表示模块实例句柄。
	HInstance windows.Handle
	// HIcon 表示大图标句柄。
	HIcon windows.Handle
	// HCursor 表示默认光标句柄。
	HCursor windows.Handle
	// HbrBackground 表示背景画刷句柄。
	HbrBackground windows.Handle
	// LpszMenuName 表示菜单资源名称。
	LpszMenuName *uint16
	// LpszClassName 表示窗口类名。
	LpszClassName *uint16
	// HIconSm 表示小图标句柄。
	HIconSm windows.Handle
}

// msg 对应 Win32 的 MSG 结构。
type msg struct {
	// HWnd 表示消息关联的窗口句柄。
	HWnd windows.Handle
	// Message 表示消息编号。
	Message uint32
	// WParam 表示消息的 WPARAM 数据。
	WParam uintptr
	// LParam 表示消息的 LPARAM 数据。
	LParam uintptr
	// Time 表示消息时间戳。
	Time uint32
	// Pt 表示消息关联的屏幕坐标。
	Pt point
}

// point 对应 Win32 的 POINT 结构。
type point struct {
	// X 表示水平坐标。
	X int32
	// Y 表示垂直坐标。
	Y int32
}

// winRect 对应 Win32 的 RECT 结构。
type winRect struct {
	// Left 表示左边界。
	Left int32
	// Top 表示上边界。
	Top int32
	// Right 表示右边界。
	Right int32
	// Bottom 表示下边界。
	Bottom int32
}

// RGB 根据红、绿、蓝通道值构造 Color。
func RGB(r, g, b byte) Color {
	return Color(uint32(r) | (uint32(g) << 8) | (uint32(b) << 16))
}

// String 返回渲染模式的可读名称。
func (m RenderMode) String() string {
	switch m {
	case RenderModeAuto:
		return "Auto"
	case RenderModeGDI:
		return "GDI"
	default:
		return "Unknown"
	}
}

// String 返回实际渲染后端的可读名称。
func (b RenderBackend) String() string {
	switch b {
	case RenderBackendGDI:
		return "GDI"
	case RenderBackendDirect2D:
		return "Direct2D"
	default:
		return "Unknown"
	}
}

// Contains 返回矩形是否包含指定点。
func (r Rect) Contains(x, y int32) bool {
	return x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
}

// Empty 返回矩形是否没有可绘制区域。
func (r Rect) Empty() bool {
	return r.W <= 0 || r.H <= 0
}

// toWinRect 把 Rect 转成兼容 Win32 RECT 的结构。
func (r Rect) toWinRect() winRect {
	return winRect{
		Left:   r.X,
		Top:    r.Y,
		Right:  r.X + r.W,
		Bottom: r.Y + r.H,
	}
}

// rectFromWinRect 把 Win32 RECT 转成库内使用的 Rect。
func rectFromWinRect(r winRect) Rect {
	return Rect{
		X: r.Left,
		Y: r.Top,
		W: r.Right - r.Left,
		H: r.Bottom - r.Top,
	}
}

// pointFromLParam 从 Win32 LPARAM 的鼠标坐标中提取 Point。
func pointFromLParam(lParam uintptr) Point {
	return Point{
		X: int32(int16(uint16(lParam & 0xFFFF))),
		Y: int32(int16(uint16((lParam >> 16) & 0xFFFF))),
	}
}

// dpiFromWParam 从 WM_DPICHANGED 的 WPARAM 中提取水平和垂直 DPI。
func dpiFromWParam(wParam uintptr) (int32, int32) {
	return int32(uint16(wParam & 0xFFFF)), int32(uint16((wParam >> 16) & 0xFFFF))
}

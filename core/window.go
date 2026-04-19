//go:build windows

package core

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	msimg32  = windows.NewLazySystemDLL("msimg32.dll")
	shcore   = windows.NewLazySystemDLL("shcore.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
)

var (
	procRegisterClassExW      = user32.NewProc("RegisterClassExW")
	procCreateWindowExW       = user32.NewProc("CreateWindowExW")
	procDefWindowProcW        = user32.NewProc("DefWindowProcW")
	procDestroyWindow         = user32.NewProc("DestroyWindow")
	procShowWindow            = user32.NewProc("ShowWindow")
	procUpdateWindow          = user32.NewProc("UpdateWindow")
	procGetMessageW           = user32.NewProc("GetMessageW")
	procTranslateMessage      = user32.NewProc("TranslateMessage")
	procDispatchMessageW      = user32.NewProc("DispatchMessageW")
	procPostQuitMessage       = user32.NewProc("PostQuitMessage")
	procPostMessageW          = user32.NewProc("PostMessageW")
	procSendMessageW          = user32.NewProc("SendMessageW")
	procInvalidateRect        = user32.NewProc("InvalidateRect")
	procGetClientRect         = user32.NewProc("GetClientRect")
	procAdjustWindowRectEx    = user32.NewProc("AdjustWindowRectEx")
	procSetWindowTextW        = user32.NewProc("SetWindowTextW")
	procMessageBoxW           = user32.NewProc("MessageBoxW")
	procMessageBoxTimeoutW    = user32.NewProc("MessageBoxTimeoutW")
	procMessageBeep           = user32.NewProc("MessageBeep")
	procGetDC                 = user32.NewProc("GetDC")
	procReleaseDC             = user32.NewProc("ReleaseDC")
	procScreenToClient        = user32.NewProc("ScreenToClient")
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
	procIsChild               = user32.NewProc("IsChild")
	procGetModuleHandleW      = kernel32.NewProc("GetModuleHandleW")
	procGetCurrentThreadID    = kernel32.NewProc("GetCurrentThreadId")
	procRtlMoveMemory         = kernel32.NewProc("RtlMoveMemory")
)

var (
	globalWndProc  = windows.NewCallback(appWndProc)
	classSequence  atomic.Uint64
	windowRegistry sync.Map
	createRegistry sync.Map
)

// App 表示一个原生 Win32 应用窗口及其运行时状态。
type App struct {
	// opts 保存创建应用时的配置。
	opts Options
	// hwnd 保存底层窗口句柄。
	hwnd windows.Handle
	// threadID 保存 UI 线程标识。
	threadID uint32
	// className 保存注册到 Win32 的窗口类名。
	className string

	// dpiMu 保护 DPI 信息。
	dpiMu sync.RWMutex
	// dpi 保存当前 DPI 状态。
	dpi DPIInfo

	// sizeMu 保护客户区尺寸。
	sizeMu sync.RWMutex
	// clientSize 保存最新客户区尺寸。
	clientSize Size

	// initOnce 保证初始化只执行一次。
	initOnce sync.Once
	// ready 用于通知初始化结果已就绪。
	ready chan struct{}
	// done 用于返回消息循环退出码。
	done chan int
	// initErr 保存初始化错误。
	initErr error

	// closed 标记窗口是否已经关闭。
	closed atomic.Bool
	// currentCursor 保存当前应用显式设置的鼠标光标。
	currentCursor atomic.Uint64

	// postMu 保护投递回调队列。
	postMu sync.Mutex
	// postQueue 保存待在 UI 线程执行的回调。
	postQueue []func()

	// timerMu 保护定时器集合。
	timerMu sync.Mutex
	// activeTimers 保存当前激活的原生定时器。
	activeTimers map[uintptr]struct{}

	// renderMu 保护渲染后端状态。
	renderMu sync.RWMutex
	// renderBackend 保存当前实际后端。
	renderBackend RenderBackend
	// renderFallback 保存回退到 GDI 的原因。
	renderFallback string
	// d2dRenderer 保存 Direct2D 渲染器实例。
	d2dRenderer *d2dRenderer
}

// NewApp 创建一个新的应用实例。
func NewApp(opts Options) (*App, error) {
	if opts.ClassName == "" {
		opts.ClassName = fmt.Sprintf("WinUICore_%d", classSequence.Add(1))
	}
	if opts.Width <= 0 {
		opts.Width = 600
	}
	if opts.Height <= 0 {
		opts.Height = 400
	}
	if opts.MinWidth > 0 && opts.Width < opts.MinWidth {
		opts.Width = opts.MinWidth
	}
	if opts.MinHeight > 0 && opts.Height < opts.MinHeight {
		opts.Height = opts.MinHeight
	}
	if opts.Style == 0 {
		opts.Style = DefaultWindowStyle
	}
	if opts.ExStyle == 0 {
		opts.ExStyle = DefaultWindowExStyle
	}
	if opts.Cursor == 0 {
		opts.Cursor = CursorArrow
	}
	if opts.Background == 0 {
		opts.Background = RGB(255, 255, 255)
	}
	if opts.RenderMode == 0 {
		opts.RenderMode = RenderModeAuto
	}

	app := &App{
		opts:         opts,
		className:    opts.ClassName,
		ready:        make(chan struct{}),
		done:         make(chan int, 1),
		activeTimers: make(map[uintptr]struct{}),
	}
	app.currentCursor.Store(uint64(opts.Cursor))
	return app, nil
}

// Init 启动应用的 UI 线程，并在需要时创建原生窗口。
func (a *App) Init() error {
	if a == nil {
		return ErrNotInitialized
	}
	a.initOnce.Do(func() {
		go a.runUIThread()
		<-a.ready
	})
	return a.initErr
}

// Run 在需要时初始化应用，并进入原生消息循环。
func (a *App) Run() int {
	if a == nil {
		return -1
	}
	if err := a.Init(); err != nil {
		return -1
	}
	return <-a.done
}

// Handle 返回应用底层的原生句柄。
func (a *App) Handle() windows.Handle {
	if a == nil {
		return 0
	}
	return a.hwnd
}

// ClientSize 返回应用的客户区尺寸。
func (a *App) ClientSize() Size {
	a.sizeMu.RLock()
	defer a.sizeMu.RUnlock()
	return a.clientSize
}

// IsUIThread 判断当前是否为应用的 UI 线程。
func (a *App) IsUIThread() bool {
	if a == nil {
		return false
	}
	return currentThreadID() == a.threadID
}

// Post 将回调调度到应用的 UI 线程执行。
func (a *App) Post(fn func()) error {
	if a == nil || a.hwnd == 0 {
		return ErrNotInitialized
	}
	if fn == nil {
		return nil
	}
	if a.closed.Load() {
		return ErrAppClosed
	}
	if currentThreadID() == a.threadID {
		fn()
		return nil
	}

	a.postMu.Lock()
	if a.closed.Load() {
		a.postMu.Unlock()
		return ErrAppClosed
	}
	a.postQueue = append(a.postQueue, fn)
	a.postMu.Unlock()

	r1, _, err := procPostMessageW.Call(uintptr(a.hwnd), wmAppInvoke, 0, 0)
	if r1 == 0 {
		return wrapError("PostMessageW", err)
	}
	return nil
}

// Invalidate 标记区域或控件需要重绘。
func (a *App) Invalidate(rect *Rect) {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}

	if rect == nil {
		procInvalidateRect.Call(uintptr(a.hwnd), 0, 0)
		return
	}

	local := rect.toWinRect()
	procInvalidateRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&local)), 0)
}

// SetTitle 更新应用窗口标题。
func (a *App) SetTitle(title string) {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}

	text := title
	_ = a.Post(func() {
		ptr, err := windows.UTF16PtrFromString(text)
		if err != nil {
			return
		}
		procSetWindowTextW.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(ptr)))
	})
}

func iconHandle(icon *Icon) windows.Handle {
	if icon == nil {
		return 0
	}
	return icon.Handle()
}

func (a *App) windowImageBaseSizeDP() int32 {
	if a == nil || a.opts.WindowImageSizeDP <= 0 {
		return 16
	}
	return a.opts.WindowImageSizeDP
}

func (a *App) resolveWindowImageIcons() (small *Icon, big *Icon, err error) {
	if a == nil {
		return nil, nil, nil
	}
	if a.opts.WindowImage == nil {
		return a.opts.Icon, a.opts.Icon, nil
	}

	baseDP := a.windowImageBaseSizeDP()
	smallSize := a.DP(baseDP)
	bigSize := a.DP(baseDP * 2)
	if smallSize < 16 {
		smallSize = 16
	}
	if bigSize < smallSize {
		bigSize = smallSize
	}

	small, err = a.opts.WindowImage.IconFor(smallSize)
	if err != nil {
		return nil, nil, err
	}
	big, err = a.opts.WindowImage.IconFor(bigSize)
	if err != nil {
		return nil, nil, err
	}
	return small, big, nil
}

func (a *App) applyResolvedWindowIcons(small *Icon, big *Icon) {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}
	procSendMessageW.Call(uintptr(a.hwnd), wmSetIcon, iconSmall, uintptr(iconHandle(small)))
	procSendMessageW.Call(uintptr(a.hwnd), wmSetIcon, iconBig, uintptr(iconHandle(big)))
}

func (a *App) applyWindowImage() {
	small, big, err := a.resolveWindowImageIcons()
	if err != nil {
		return
	}
	a.applyResolvedWindowIcons(small, big)
}

// SetWindowImage 设置窗口图片资源，并在内部生成 HICON。
func (a *App) SetWindowImage(img *Image) {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}
	a.opts.WindowImage = img
	a.opts.Icon = nil

	apply := func() {
		a.applyWindowImage()
	}
	if a.IsUIThread() {
		apply()
		return
	}
	_ = a.Post(apply)
}

// SetIcon 保留旧的 HICON 设置路径。
func (a *App) SetIcon(icon *Icon) {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}
	a.opts.Icon = icon
	apply := func() {
		var handle windows.Handle
		if icon != nil {
			handle = icon.Handle()
		}
		procSendMessageW.Call(uintptr(a.hwnd), wmSetIcon, iconSmall, uintptr(handle))
		procSendMessageW.Call(uintptr(a.hwnd), wmSetIcon, iconBig, uintptr(handle))
	}
	if a.IsUIThread() {
		apply()
		return
	}
	_ = a.Post(apply)
}

// MessageBox 显示一个由应用窗口拥有的原生 Windows 消息框。
func (a *App) MessageBox(title, text string, flags uint32, timeout time.Duration) (int, error) {
	if a == nil || a.hwnd == 0 {
		return 0, ErrNotInitialized
	}

	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}
	textPtr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return 0, err
	}

	if timeout > 0 && procMessageBoxTimeoutW.Find() == nil {
		r1, _, _ := procMessageBoxTimeoutW.Call(
			uintptr(a.hwnd),
			uintptr(unsafe.Pointer(textPtr)),
			uintptr(unsafe.Pointer(titlePtr)),
			uintptr(flags),
			0,
			uintptr(timeout/time.Millisecond),
		)
		return int(r1), nil
	}

	r1, _, callErr := procMessageBoxW.Call(
		uintptr(a.hwnd),
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
	if r1 == 0 {
		return 0, wrapError("MessageBoxW", callErr)
	}
	return int(r1), nil
}

// MessageBeep 播放默认的 Windows 消息提示音。
func MessageBeep() error {
	r1, _, err := procMessageBeep.Call(0)
	if r1 == 0 {
		return wrapError("MessageBeep", err)
	}
	return nil
}

// runUIThread 锁定到操作系统线程，创建窗口并持有 UI 循环。
func (a *App) runUIThread() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer a.closeRenderer()

	a.threadID = currentThreadID()
	a.setDPI(initProcessDPIAwareness())

	if err := a.createWindow(); err != nil {
		a.initErr = err
		close(a.ready)
		a.done <- -1
		return
	}

	close(a.ready)
	a.done <- a.runLoop()
}

// createWindow 注册窗口类并创建原生窗口实例。
func (a *App) createWindow() error {
	hInst, _, err := procGetModuleHandleW.Call(0)
	if hInst == 0 {
		return wrapError("GetModuleHandleW", err)
	}

	classNamePtr, err := windows.UTF16PtrFromString(a.className)
	if err != nil {
		return err
	}

	cursor, err := loadCursor(a.opts.Cursor)
	if err != nil {
		return err
	}

	smallIcon, bigIcon, err := a.resolveWindowImageIcons()
	if err != nil {
		return err
	}

	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   globalWndProc,
		HInstance:     windows.Handle(hInst),
		HCursor:       cursor,
		HIcon:         iconHandle(bigIcon),
		HIconSm:       iconHandle(smallIcon),
		LpszClassName: classNamePtr,
	}

	if atom, _, regErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		if last := windows.GetLastError(); last != windows.ERROR_CLASS_ALREADY_EXISTS {
			return wrapError("RegisterClassExW", regErr)
		}
	}

	wa, err := workArea()
	if err != nil {
		return err
	}

	clientW := a.DP(a.opts.Width)
	clientH := a.DP(a.opts.Height)
	frame := winRect{Right: clientW, Bottom: clientH}
	if r1, _, adjustErr := procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&frame)),
		uintptr(a.opts.Style),
		0,
		uintptr(a.opts.ExStyle),
	); r1 == 0 {
		return wrapError("AdjustWindowRectEx", adjustErr)
	}

	winW := frame.Right - frame.Left
	winH := frame.Bottom - frame.Top
	x := wa.Left + (wa.Right-wa.Left-winW)/2
	y := wa.Top + (wa.Bottom-wa.Top-winH)/2

	title := a.opts.Title
	if title == "" {
		title = a.className
	}
	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	createRegistry.Store(a.threadID, a)
	hwnd, _, createErr := procCreateWindowExW.Call(
		uintptr(a.opts.ExStyle),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(a.opts.Style),
		uintptr(x),
		uintptr(y),
		uintptr(winW),
		uintptr(winH),
		0,
		0,
		hInst,
		0,
	)
	createRegistry.Delete(a.threadID)
	if hwnd == 0 {
		return wrapError("CreateWindowExW", createErr)
	}

	a.hwnd = windows.Handle(hwnd)
	windowRegistry.Store(a.hwnd, a)
	a.refreshWindowDPI()
	a.initRenderer()

	a.applyResolvedWindowIcons(smallIcon, bigIcon)

	size := Size{}
	hasSize := false
	if rect, err := clientRect(a.hwnd); err == nil {
		size = Size{Width: rect.W, Height: rect.H}
		hasSize = true
	}
	if err := a.finishCreate(size, hasSize); err != nil {
		procDestroyWindow.Call(hwnd)
		return err
	}

	procShowWindow.Call(hwnd, showWindowNormal)
	procUpdateWindow.Call(hwnd)
	a.scheduleInitialRelayout()
	return nil
}

// clientRect 返回指定窗口句柄当前的客户区矩形。
func clientRect(hwnd windows.Handle) (Rect, error) {
	var rc winRect
	r1, _, err := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	if r1 == 0 {
		return Rect{}, wrapError("GetClientRect", err)
	}
	return rectFromWinRect(rc), nil
}

// workArea 返回操作系统报告的桌面工作区。
func workArea() (winRect, error) {
	var rc winRect
	r1, _, err := procSystemParametersInfoW.Call(
		spiGetWorkArea,
		0,
		uintptr(unsafe.Pointer(&rc)),
		0,
	)
	if r1 == 0 {
		return winRect{}, wrapError("SystemParametersInfoW", err)
	}
	return rc, nil
}

// loadCursor 按标识加载预定义系统光标。
func loadCursor(cursor CursorID) (windows.Handle, error) {
	h, _, err := procLoadCursorW.Call(0, uintptr(cursor))
	if h == 0 {
		return 0, wrapError("LoadCursorW", err)
	}
	return windows.Handle(h), nil
}

// setCursor 更新应用使用的光标。
func (a *App) setCursor(cursor CursorID) {
	h, err := loadCursor(cursor)
	if err != nil {
		return
	}
	a.currentCursor.Store(uint64(cursor))
	procSetCursor.Call(uintptr(h))
}

func (a *App) effectiveCursor() CursorID {
	if a == nil {
		return CursorArrow
	}
	if cursor := CursorID(a.currentCursor.Load()); cursor != 0 {
		return cursor
	}
	if a.opts.Cursor != 0 {
		return a.opts.Cursor
	}
	return CursorArrow
}

// SetCursor 更新应用使用的光标。
func (a *App) SetCursor(cursor CursorID) {
	if a == nil || a.hwnd == 0 {
		return
	}
	if a.IsUIThread() {
		a.setCursor(cursor)
		return
	}
	_ = a.Post(func() {
		a.setCursor(cursor)
	})
}

// captureMouse 在 UI 线程为应用窗口捕获鼠标输入。
func (a *App) captureMouse() {
	if a == nil || a.hwnd == 0 {
		return
	}
	procSetCapture.Call(uintptr(a.hwnd))
}

// CaptureMouse 在 UI 线程为应用窗口捕获鼠标输入。
func (a *App) CaptureMouse() {
	if a == nil || a.hwnd == 0 {
		return
	}
	if a.IsUIThread() {
		a.captureMouse()
		return
	}
	_ = a.Post(func() {
		a.captureMouse()
	})
}

// releaseMouse 在 UI 线程释放鼠标捕获。
func (a *App) releaseMouse() {
	procReleaseCapture.Call()
}

// ReleaseMouse 在 UI 线程释放鼠标捕获。
func (a *App) ReleaseMouse() {
	if a == nil || a.hwnd == 0 {
		return
	}
	if a.IsUIThread() {
		a.releaseMouse()
		return
	}
	_ = a.Post(func() {
		a.releaseMouse()
	})
}

func isChildWindow(parent, child windows.Handle) bool {
	if parent == 0 || child == 0 {
		return false
	}
	result, _, _ := procIsChild.Call(uintptr(parent), uintptr(child))
	return result != 0
}

// updateClientSize 保存应用窗口最新的客户区尺寸。
func (a *App) updateClientSize(size Size) {
	a.sizeMu.Lock()
	a.clientSize = size
	a.sizeMu.Unlock()
}

func (a *App) finishCreate(size Size, hasSize bool) error {
	if hasSize {
		a.updateClientSize(size)
	}
	if a.opts.OnCreate != nil {
		if err := a.opts.OnCreate(a); err != nil {
			return err
		}
	}
	if hasSize && a.opts.OnResize != nil {
		a.opts.OnResize(a, size)
	}
	return nil
}

func (a *App) scheduleInitialRelayout() {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}
	_ = a.Post(func() {
		if a == nil || a.hwnd == 0 || a.closed.Load() {
			return
		}
		rect, err := clientRect(a.hwnd)
		if err != nil {
			return
		}
		size := Size{Width: rect.W, Height: rect.H}
		a.updateClientSize(size)
		if a.opts.OnResize != nil {
			a.opts.OnResize(a, size)
		}
	})
}

// minTrackSize 返回窗口最小可拖拽外框尺寸。
func (a *App) minTrackSize() Size {
	if a == nil {
		return Size{}
	}
	minW := a.opts.MinWidth
	minH := a.opts.MinHeight
	if minW <= 0 && minH <= 0 {
		return Size{}
	}
	if minW < 0 {
		minW = 0
	}
	if minH < 0 {
		minH = 0
	}

	clientW := a.DP(minW)
	clientH := a.DP(minH)
	frame := winRect{Right: clientW, Bottom: clientH}
	if r1, _, _ := procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&frame)),
		uintptr(a.opts.Style),
		0,
		uintptr(a.opts.ExStyle),
	); r1 == 0 {
		return Size{Width: clientW, Height: clientH}
	}
	width := frame.Right - frame.Left
	height := frame.Bottom - frame.Top
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return Size{Width: width, Height: height}
}

// currentThreadID 返回调用线程的标识。
// Close 请求关闭当前应用窗口。
func (a *App) Close() {
	if a == nil || a.hwnd == 0 || a.closed.Load() {
		return
	}
	if a.IsUIThread() {
		procDestroyWindow.Call(uintptr(a.hwnd))
		return
	}
	_ = a.Post(func() {
		if a.hwnd != 0 && !a.closed.Load() {
			procDestroyWindow.Call(uintptr(a.hwnd))
		}
	})
}

// currentThreadID 返回调用线程的标识。
func currentThreadID() uint32 {
	r1, _, _ := procGetCurrentThreadID.Call()
	return uint32(r1)
}

// screenToClient 将屏幕坐标转换为窗口客户区坐标。
func (a *App) screenToClient(pt Point) Point {
	if a == nil || a.hwnd == 0 {
		return pt
	}
	wp := point{X: pt.X, Y: pt.Y}
	procScreenToClient.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&wp)))
	return Point{X: wp.X, Y: wp.Y}
}

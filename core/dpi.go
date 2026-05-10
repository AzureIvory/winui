//go:build windows

package core

import (
	"math"
	"sync"
	"unsafe"
)

var (
	procSetProcessDPIAware              = user32.NewProc("SetProcessDPIAware")
	procSetProcessDpiAwarenessContext   = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetThreadDpiAwarenessContext    = user32.NewProc("GetThreadDpiAwarenessContext")
	procGetAwarenessFromDpiAwarenessCtx = user32.NewProc("GetAwarenessFromDpiAwarenessContext")
	procGetDpiForWindow                 = user32.NewProc("GetDpiForWindow")
	procSetProcessDpiAwareness          = shcore.NewProc("SetProcessDpiAwareness")
	procGetDeviceCaps                   = gdi32.NewProc("GetDeviceCaps")
)

var dpiInitOnce sync.Once

const (
	// processDpiSystemAware 表示系统级 DPI 感知常量值。
	processDpiSystemAware = 1
	// processDpiPerMonitor 表示按显示器 DPI 感知常量值。
	processDpiPerMonitor = 2
)

const (
	// hResultOK 表示调用成功的 HRESULT。
	hResultOK uint32 = 0
	// hResultDenied 表示访问被拒绝但可视为已设置的 HRESULT。
	hResultDenied uint32 = 0x80070005
)

const (
	// dpiAwareContextSystemAware 表示系统级 DPI 感知上下文。
	dpiAwareContextSystemAware = ^uintptr(1)
	// dpiAwareContextPerMonitorAware 表示按显示器 DPI 感知上下文。
	dpiAwareContextPerMonitorAware = ^uintptr(2)
	// dpiAwareContextPerMonitorAwareV2 表示按显示器 DPI 感知 V2 上下文。
	dpiAwareContextPerMonitorAwareV2 = ^uintptr(3)
)

// initProcessDPIAwareness 配置进程的 DPI 模式，并返回初始 DPI 信息。
func initProcessDPIAwareness() DPIInfo {
	awareness := DPIAwarenessUnknown

	dpiInitOnce.Do(func() {
		switch {
		case trySetDpiAwarenessContext(dpiAwareContextPerMonitorAwareV2):
			awareness = DPIAwarenessPerMonitorV2
		case trySetDpiAwarenessContext(dpiAwareContextPerMonitorAware):
			awareness = DPIAwarenessPerMonitor
		case trySetProcessDpiAwareness(processDpiPerMonitor):
			awareness = DPIAwarenessPerMonitor
		case trySetProcessDpiAwareness(processDpiSystemAware):
			awareness = DPIAwarenessSystem
		case trySetProcessDPIAware():
			awareness = DPIAwarenessSystem
		default:
			awareness = currentDPIAwareness()
			if awareness == DPIAwarenessUnknown {
				awareness = DPIAwarenessSystem
			}
		}
	})

	if current := currentDPIAwareness(); current != DPIAwarenessUnknown {
		awareness = current
	} else if awareness == DPIAwarenessUnknown {
		awareness = DPIAwarenessSystem
	}

	// 先启用 DPI 感知，再查询一次主屏 DPI，避免初次窗口尺寸仍按 96 DPI 创建。
	return resolveInitialDPIInfo(queryScreenDPI, awareness)
}

// resolveInitialDPIInfo 在 DPI 感知配置完成后重新查询主屏 DPI，
// 并补齐 Awareness 与兜底值，供窗口初始尺寸计算使用。
func resolveInitialDPIInfo(query func() DPIInfo, awareness DPIAwareness) DPIInfo {
	info := DPIInfo{X: 96, Y: 96, Scale: 1, Awareness: awareness}
	if query == nil {
		return info
	}

	queried := query()
	if queried.X > 0 {
		info.X = queried.X
	}
	if queried.Y > 0 {
		info.Y = queried.Y
	}
	if queried.Scale > 0 {
		info.Scale = queried.Scale
	} else if info.X > 0 {
		info.Scale = float64(info.X) / 96.0
	}
	if info.Y <= 0 {
		info.Y = info.X
	}
	if info.Scale <= 0 {
		info.Scale = 1
	}
	return info
}

// trySetDpiAwarenessContext 尝试设置进程的 DPI 感知上下文。
func trySetDpiAwarenessContext(ctx uintptr) bool {
	if err := procSetProcessDpiAwarenessContext.Find(); err != nil {
		return false
	}
	r1, _, _ := procSetProcessDpiAwarenessContext.Call(ctx)
	return r1 != 0
}

// trySetProcessDpiAwareness 尝试通过 shcore.dll 设置进程的 DPI 感知级别。
func trySetProcessDpiAwareness(level uintptr) bool {
	if err := procSetProcessDpiAwareness.Find(); err != nil {
		return false
	}
	hr, _, _ := procSetProcessDpiAwareness.Call(level)
	return uint32(hr) == hResultOK || uint32(hr) == hResultDenied
}

// trySetProcessDPIAware 尝试为当前进程启用旧版系统 DPI 感知。
func trySetProcessDPIAware() bool {
	if err := procSetProcessDPIAware.Find(); err != nil {
		return false
	}
	r1, _, _ := procSetProcessDPIAware.Call()
	return r1 != 0
}

// currentDPIAwareness 返回当前线程的 DPI 感知模式。
func currentDPIAwareness() DPIAwareness {
	if err := procGetThreadDpiAwarenessContext.Find(); err != nil {
		return DPIAwarenessUnknown
	}
	ctx, _, _ := procGetThreadDpiAwarenessContext.Call()
	switch ctx {
	case dpiAwareContextPerMonitorAwareV2:
		return DPIAwarenessPerMonitorV2
	case dpiAwareContextPerMonitorAware:
		return DPIAwarenessPerMonitor
	case dpiAwareContextSystemAware:
		return DPIAwarenessSystem
	}

	if err := procGetAwarenessFromDpiAwarenessCtx.Find(); err != nil {
		return DPIAwarenessUnknown
	}
	r1, _, _ := procGetAwarenessFromDpiAwarenessCtx.Call(ctx)
	switch r1 {
	case processDpiSystemAware:
		return DPIAwarenessSystem
	case processDpiPerMonitor:
		return DPIAwarenessPerMonitor
	default:
		return DPIAwarenessUnknown
	}
}

// queryScreenDPI 查询主屏幕 DPI，并以 DPIInfo 返回。
func queryScreenDPI() DPIInfo {
	info := DPIInfo{X: 96, Y: 96, Scale: 1}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return info
	}
	defer procReleaseDC.Call(0, hdc)

	dx, _, _ := procGetDeviceCaps.Call(hdc, logPixelsX)
	dy, _, _ := procGetDeviceCaps.Call(hdc, logPixelsY)
	if dx > 0 {
		info.X = int32(dx)
		info.Scale = float64(dx) / 96.0
	}
	if dy > 0 {
		info.Y = int32(dy)
	}
	return info
}

// ScreenDPI returns the current primary-screen DPI information.
func ScreenDPI() DPIInfo {
	return queryScreenDPI()
}

// refreshWindowDPI 刷新应用窗口的 DPI 信息。
func (a *App) refreshWindowDPI() {
	if a == nil || a.hwnd == 0 {
		return
	}
	if err := procGetDpiForWindow.Find(); err == nil {
		dpiX, _, _ := procGetDpiForWindow.Call(uintptr(a.hwnd))
		if dpiX > 0 {
			info := a.DPI()
			info.X = int32(dpiX)
			info.Y = int32(dpiX)
			info.Scale = float64(info.X) / 96.0
			a.setDPI(info)
			return
		}
	}
	info := queryScreenDPI()
	info.Awareness = a.DPI().Awareness
	a.setDPI(info)
}

// setDPI 更新应用当前的 DPI 信息。
func (a *App) setDPI(info DPIInfo) {
	a.dpiMu.Lock()
	a.dpi = info
	a.dpiMu.Unlock()
}

// DPI 返回应用缓存的 DPI 信息。
func (a *App) DPI() DPIInfo {
	a.dpiMu.RLock()
	defer a.dpiMu.RUnlock()
	return a.dpi
}

// DP 按应用当前 DPI 缩放设备无关值。
func (a *App) DP(value int32) int32 {
	scale := a.DPI().Scale
	if scale <= 0 {
		scale = 1
	}
	return int32(math.Round(float64(value) * scale))
}

// Scale 按应用当前 DPI 缩放整数值。
func (a *App) Scale(value int) int {
	return int(a.DP(int32(value)))
}

// dpiChangeFromMessage 将 WM_DPICHANGED 消息解析为 DPI 信息和建议边界。
func dpiChangeFromMessage(wParam, lParam uintptr, current DPIInfo) (DPIInfo, *winRect) {
	x, y := dpiFromWParam(wParam)
	if x <= 0 {
		x = current.X
	}
	if y <= 0 {
		y = current.Y
	}
	if x <= 0 {
		x = 96
	}
	if y <= 0 {
		y = x
	}

	info := DPIInfo{
		X:         x,
		Y:         y,
		Scale:     float64(x) / 96.0,
		Awareness: current.Awareness,
	}
	if lParam == 0 {
		return info, nil
	}
	suggested := &winRect{}
	size := uintptr(unsafe.Sizeof(*suggested))
	procRtlMoveMemory.Call(uintptr(unsafe.Pointer(suggested)), lParam, size)
	return info, suggested
}

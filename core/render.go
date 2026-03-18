//go:build windows

package core

// RenderMode 返回应用请求的渲染模式。
func (a *App) RenderMode() RenderMode {
	if a == nil {
		return RenderModeAuto
	}
	return a.opts.RenderMode
}

// RenderBackend 返回应用当前实际使用的渲染后端。
func (a *App) RenderBackend() RenderBackend {
	if a == nil {
		return RenderBackendUnknown
	}
	a.renderMu.RLock()
	defer a.renderMu.RUnlock()
	return a.renderBackend
}

// RenderFallbackReason 返回后端回退到 GDI 的原因。
func (a *App) RenderFallbackReason() string {
	if a == nil {
		return ""
	}
	a.renderMu.RLock()
	defer a.renderMu.RUnlock()
	return a.renderFallback
}

// setRenderBackend 更新当前激活的后端和回退原因。
func (a *App) setRenderBackend(backend RenderBackend, reason string) {
	if a == nil {
		return
	}
	a.renderMu.Lock()
	a.renderBackend = backend
	a.renderFallback = reason
	a.renderMu.Unlock()
}

// initRenderer 按配置初始化渲染后端。
func (a *App) initRenderer() {
	if a == nil {
		return
	}
	if a.opts.RenderMode == RenderModeGDI {
		a.setRenderBackend(RenderBackendGDI, "")
		return
	}

	renderer, err := newD2DRenderer()
	if err != nil {
		a.setRenderBackend(RenderBackendGDI, err.Error())
		return
	}

	a.renderMu.Lock()
	a.d2dRenderer = renderer
	a.renderBackend = RenderBackendDirect2D
	a.renderFallback = ""
	a.renderMu.Unlock()
}

// fallbackToGDI 关闭 Direct2D 并切回 GDI。
func (a *App) fallbackToGDI(err error) {
	if a == nil {
		return
	}

	var reason string
	if err != nil {
		reason = err.Error()
	}

	a.renderMu.Lock()
	if a.d2dRenderer != nil {
		a.d2dRenderer.Close()
		a.d2dRenderer = nil
	}
	a.renderBackend = RenderBackendGDI
	a.renderFallback = reason
	a.renderMu.Unlock()
}

// closeRenderer 释放当前持有的渲染器实例。
func (a *App) closeRenderer() {
	if a == nil {
		return
	}
	a.renderMu.Lock()
	renderer := a.d2dRenderer
	a.d2dRenderer = nil
	a.renderMu.Unlock()
	if renderer != nil {
		renderer.Close()
	}
}

// rendererForPaint 返回本次绘制可用的 Direct2D 实例及后端状态。
func (a *App) rendererForPaint() (*d2dRenderer, RenderBackend) {
	if a == nil {
		return nil, RenderBackendUnknown
	}
	a.renderMu.RLock()
	defer a.renderMu.RUnlock()
	return a.d2dRenderer, a.renderBackend
}

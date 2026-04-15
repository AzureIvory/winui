//go:build windows

package jsonui

import (
	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type ReloadReason uint8

const (
	ReloadReasonAttach ReloadReason = iota + 1
	ReloadReasonDPIChanged
	ReloadReasonHotSwap
)

type resourceReloadContext struct {
	Reason ReloadReason
	Window *Window
	Scene  *widgets.Scene
	App    *core.App
	Scale  float64
}

type resourceReloader struct {
	reload func(resourceReloadContext) error
}

// ReloadResources 刷新资源
func (w *Window) ReloadResources(reason ReloadReason) error {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	reloaders := append([]resourceReloader(nil), w.reloaders...)
	scene := w.scene
	w.mu.Unlock()

	ctx := resourceReloadContext{
		Reason: reason,
		Window: w,
		Scene:  scene,
	}
	if scene != nil {
		ctx.App = scene.App()
	}
	ctx.Scale = resourceReloadScale(ctx.App)

	for _, reloader := range reloaders {
		if reloader.reload == nil {
			continue
		}
		if err := reloader.reload(ctx); err != nil {
			return err
		}
	}
	return nil
}

func resourceReloadScale(app *core.App) float64 {
	if app != nil {
		if scale := app.DPI().Scale; scale > 0 {
			return scale
		}
	}
	if scale := core.ScreenDPI().Scale; scale > 0 {
		return scale
	}
	return 1
}

//go:build windows

package jsonui

import (
	"sync"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// HostedWindow binds one JSON UI window to a core.App and SceneRef pair.
type HostedWindow struct {
	ID       string
	Window   *Window
	App      *core.App
	SceneRef *widgets.SceneRef
}

// Scene returns the hosted window's Scene after OnCreate has run.
func (w *HostedWindow) Scene() *widgets.Scene {
	if w == nil || w.SceneRef == nil {
		return nil
	}
	return w.SceneRef.Scene()
}

// NewApps creates one core.App per JSON UI window using a shared base options template.
func (d *Document) NewApps(base core.Options) ([]*HostedWindow, error) {
	if d == nil || len(d.Windows) == 0 {
		return nil, nil
	}

	out := make([]*HostedWindow, 0, len(d.Windows))
	for _, window := range d.Windows {
		if window == nil {
			continue
		}
		window := window
		opts := base
		window.ApplyOptions(&opts)
		ref := widgets.BindScene(&opts, widgets.SceneHooks{
			Theme: window.theme,
			OnCreate: func(_ *core.App, scene *widgets.Scene) error {
				return window.Attach(scene)
			},
			OnDPIChanged: func(_ *core.App, _ *widgets.Scene, _ core.DPIInfo) {
				_ = window.ReloadResources(ReloadReasonDPIChanged)
			},
		})
		app, err := core.NewApp(opts)
		if err != nil {
			return nil, err
		}
		out = append(out, &HostedWindow{
			ID:       window.ID,
			Window:   window,
			App:      app,
			SceneRef: ref,
		})
	}
	return out, nil
}

// RunApps starts every hosted window and waits for all message loops to exit.
func RunApps(windows []*HostedWindow) int {
	if len(windows) == 0 {
		return 0
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		firstCode int
		hasCode   bool
	)
	for _, window := range windows {
		if window == nil || window.App == nil {
			continue
		}
		wg.Add(1)
		go func(app *core.App) {
			defer wg.Done()
			code := app.Run()
			mu.Lock()
			if !hasCode || (firstCode == 0 && code != 0) {
				firstCode = code
				hasCode = true
			}
			mu.Unlock()
		}(window.App)
	}
	wg.Wait()
	return firstCode
}

//go:build windows

package jsonui

import (
	"fmt"
	"math"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type iconPolicy string

const (
	iconPolicyAuto  iconPolicy = "auto"
	iconPolicyFixed iconPolicy = "fixed"
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
	policy iconPolicy
	reload func(resourceReloadContext) error
}

func normalizeIconPolicy(policy iconPolicy) iconPolicy {
	switch policy {
	case iconPolicyFixed:
		return iconPolicyFixed
	default:
		return iconPolicyAuto
	}
}

func parseIconPolicy(text string) (iconPolicy, error) {
	switch text {
	case "", string(iconPolicyAuto):
		return iconPolicyAuto, nil
	case string(iconPolicyFixed):
		return iconPolicyFixed, nil
	default:
		return "", fmt.Errorf("unsupported iconPolicy %q", text)
	}
}

func resolveIconLoadSize(sizeDP int32, scale float64, policies ...iconPolicy) int32 {
	policy := iconPolicyAuto
	if len(policies) > 0 {
		policy = normalizeIconPolicy(policies[0])
	}
	if sizeDP <= 0 {
		sizeDP = 32
	}
	if policy == iconPolicyFixed {
		if sizeDP < 1 {
			return 1
		}
		return sizeDP
	}
	if scale <= 0 {
		scale = 1
	}
	size := int32(math.Round(float64(sizeDP) * scale))
	if size < 1 {
		return 1
	}
	return size
}

// ReloadResources refreshes resources whose density or window attachment depends on runtime state.
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
		if reason == ReloadReasonDPIChanged && normalizeIconPolicy(reloader.policy) == iconPolicyFixed {
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

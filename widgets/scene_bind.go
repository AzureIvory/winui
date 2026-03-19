//go:build windows

package widgets

import (
	"sync/atomic"

	"github.com/AzureIvory/winui/core"
)

// SceneRef 持有延迟创建的场景引用，便于在 App 初始化后读取当前场景。
type SceneRef struct {
	// scene 保存当前绑定到应用实例的场景对象。
	scene atomic.Pointer[Scene]
}

// Scene 返回当前已经绑定到 App 的场景实例。
func (r *SceneRef) Scene() *Scene {
	if r == nil {
		return nil
	}
	return r.scene.Load()
}

// SceneHooks 描述把 Scene 绑定到 core.Options 时可选的扩展钩子。
type SceneHooks struct {
	// Theme 指定场景创建后立即应用的主题。
	Theme *Theme
	// OnCreate 在场景创建完成后执行自定义初始化。
	OnCreate func(*core.App, *Scene) error
	// BeforePaint 在场景绘制前执行额外绘制逻辑。
	BeforePaint func(*core.App, *Scene, *core.Canvas)
	// AfterPaint 在场景绘制后执行额外绘制逻辑。
	AfterPaint func(*core.App, *Scene, *core.Canvas)
	// OnResize 在场景同步尺寸后执行额外逻辑。
	OnResize func(*core.App, *Scene, core.Size)
	// OnFocus 在场景处理焦点切换后执行额外逻辑。
	OnFocus func(*core.App, *Scene, bool)
	// OnTimer 在场景处理定时器后执行额外逻辑。
	OnTimer func(*core.App, *Scene, uintptr)
	// OnDPIChanged 在场景刷新 DPI 相关资源后执行额外逻辑。
	OnDPIChanged func(*core.App, *Scene, core.DPIInfo)
	// OnDestroy 在场景释放前执行额外清理逻辑。
	OnDestroy func(*core.App, *Scene)
}

// BindScene 把 Scene 所需的绘制、输入和生命周期回调绑定到 core.Options。
func BindScene(opts *core.Options, hooks SceneHooks) *SceneRef {
	ref := &SceneRef{}
	if opts == nil {
		return ref
	}

	prevCreate := opts.OnCreate
	prevPaint := opts.OnPaint
	prevResize := opts.OnResize
	prevMouseMove := opts.OnMouseMove
	prevMouseLeave := opts.OnMouseLeave
	prevMouseDown := opts.OnMouseDown
	prevMouseUp := opts.OnMouseUp
	prevMouseWheel := opts.OnMouseWheel
	prevKeyDown := opts.OnKeyDown
	prevChar := opts.OnChar
	prevFocus := opts.OnFocus
	prevTimer := opts.OnTimer
	prevDPIChanged := opts.OnDPIChanged
	prevDestroy := opts.OnDestroy

	opts.OnCreate = func(app *core.App) error {
		scene := NewScene(app)
		ref.scene.Store(scene)
		if hooks.Theme != nil {
			scene.SetTheme(hooks.Theme)
		}
		if prevCreate != nil {
			if err := prevCreate(app); err != nil {
				ref.scene.Store(nil)
				_ = scene.Close()
				return err
			}
		}
		if hooks.OnCreate != nil {
			if err := hooks.OnCreate(app, scene); err != nil {
				ref.scene.Store(nil)
				_ = scene.Close()
				return err
			}
		}
		return nil
	}

	opts.OnPaint = func(app *core.App, canvas *core.Canvas) {
		if prevPaint != nil {
			prevPaint(app, canvas)
		}
		scene := ref.Scene()
		if scene == nil {
			return
		}
		if hooks.BeforePaint != nil {
			hooks.BeforePaint(app, scene, canvas)
		}
		scene.PaintCore(canvas)
		if hooks.AfterPaint != nil {
			hooks.AfterPaint(app, scene, canvas)
		}
	}

	opts.OnResize = func(app *core.App, size core.Size) {
		scene := ref.Scene()
		if scene != nil {
			scene.Resize(core.Rect{X: 0, Y: 0, W: size.Width, H: size.Height})
		}
		if prevResize != nil {
			prevResize(app, size)
		}
		if scene != nil && hooks.OnResize != nil {
			hooks.OnResize(app, scene, size)
		}
	}

	opts.OnMouseMove = func(app *core.App, ev core.MouseEvent) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchMouseMove(ev)
		}
		if prevMouseMove != nil {
			prevMouseMove(app, ev)
		}
	}

	opts.OnMouseLeave = func(app *core.App) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchMouseLeave()
		}
		if prevMouseLeave != nil {
			prevMouseLeave(app)
		}
	}

	opts.OnMouseDown = func(app *core.App, ev core.MouseEvent) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchMouseDown(ev)
		}
		if prevMouseDown != nil {
			prevMouseDown(app, ev)
		}
	}

	opts.OnMouseUp = func(app *core.App, ev core.MouseEvent) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchMouseUp(ev)
		}
		if prevMouseUp != nil {
			prevMouseUp(app, ev)
		}
	}

	opts.OnMouseWheel = func(app *core.App, ev core.MouseEvent) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchMouseWheel(ev)
		}
		if prevMouseWheel != nil {
			prevMouseWheel(app, ev)
		}
	}

	opts.OnKeyDown = func(app *core.App, ev core.KeyEvent) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchKeyDown(ev)
		}
		if prevKeyDown != nil {
			prevKeyDown(app, ev)
		}
	}

	opts.OnChar = func(app *core.App, ch rune) {
		scene := ref.Scene()
		if scene != nil {
			scene.DispatchChar(ch)
		}
		if prevChar != nil {
			prevChar(app, ch)
		}
	}

	opts.OnFocus = func(app *core.App, focused bool) {
		scene := ref.Scene()
		if scene != nil && !focused {
			scene.Blur()
		}
		if prevFocus != nil {
			prevFocus(app, focused)
		}
		if scene != nil && hooks.OnFocus != nil {
			hooks.OnFocus(app, scene, focused)
		}
	}

	opts.OnTimer = func(app *core.App, id uintptr) {
		scene := ref.Scene()
		if scene != nil {
			scene.HandleTimer(id)
		}
		if prevTimer != nil {
			prevTimer(app, id)
		}
		if scene != nil && hooks.OnTimer != nil {
			hooks.OnTimer(app, scene, id)
		}
	}

	opts.OnDPIChanged = func(app *core.App, info core.DPIInfo) {
		scene := ref.Scene()
		if scene != nil {
			scene.ReloadResources()
		}
		if prevDPIChanged != nil {
			prevDPIChanged(app, info)
		}
		if scene != nil && hooks.OnDPIChanged != nil {
			hooks.OnDPIChanged(app, scene, info)
		}
	}

	opts.OnDestroy = func(app *core.App) {
		scene := ref.Scene()
		if prevDestroy != nil {
			prevDestroy(app)
		}
		if scene != nil && hooks.OnDestroy != nil {
			hooks.OnDestroy(app, scene)
		}
		if scene != nil {
			_ = scene.Close()
			ref.scene.Store(nil)
		}
	}

	return ref
}

//go:build windows

package widgets

import (
	"fmt"
	"sync/atomic"

	"github.com/AzureIvory/winui/core"
)

// Rect 复用 core 包中的矩形类型。
type Rect = core.Rect

// Color 复用 core 包中的颜色类型。
type Color = core.Color

// CursorID 复用 core 包中的光标标识类型。
type CursorID = core.CursorID

// Widget 定义所有控件都要实现的基础行为。
type Widget interface {
	// ID 返回控件标识。
	ID() string
	// Bounds 返回控件边界。
	Bounds() Rect
	// SetBounds 更新控件边界。
	SetBounds(Rect)
	// Visible 返回控件是否可见。
	Visible() bool
	// SetVisible 更新控件可见状态。
	SetVisible(bool)
	// Enabled 返回控件是否可用。
	Enabled() bool
	// SetEnabled 更新控件可用状态。
	SetEnabled(bool)
	// HitTest 判断点是否命中控件。
	HitTest(x, y int32) bool
	// OnEvent 处理分发到控件的事件。
	OnEvent(evt Event) bool
	// Paint 使用绘制上下文渲染控件。
	Paint(ctx *PaintCtx)
}

// widgetNode 定义场景树内部节点需要实现的附加行为。
type widgetNode interface {
	Widget
	// setScene 绑定控件所属场景。
	setScene(*Scene)
	// scene 返回控件所属场景。
	scene() *Scene
	// setParent 绑定父容器。
	setParent(Container)
	// parent 返回父容器。
	parent() Container
	// cursor 返回悬停时应显示的光标。
	cursor() CursorID
}

// focusableWidget 表示可接收键盘焦点的控件。
type focusableWidget interface {
	Widget
	// acceptsFocus 返回控件是否允许获得焦点。
	acceptsFocus() bool
}

// overlayWidget 表示需要在覆盖层阶段追加绘制的控件。
type overlayWidget interface {
	Widget
	// PaintOverlay 在覆盖层阶段绘制额外内容。
	PaintOverlay(ctx *PaintCtx)
}

// dirtyWidget 表示可提供自定义脏区的控件。
type dirtyWidget interface {
	// dirtyRect 返回控件当前需要刷新的矩形区域。
	dirtyRect() Rect
}

// widgetSequence 用于生成自动控件标识。
var widgetSequence atomic.Uint64

// newWidgetID 使用给定前缀生成唯一的控件标识。
func newWidgetID(prefix string) string {
	if prefix == "" {
		prefix = "widget"
	}
	return fmt.Sprintf("%s-%d", prefix, widgetSequence.Add(1))
}

// widgetBase 保存大多数控件共享的基础状态。
type widgetBase struct {
	// id 保存控件唯一标识。
	id string
	// bounds 保存控件边界。
	bounds Rect
	// visible 记录控件是否可见。
	visible bool
	// enabled 记录控件是否可用。
	enabled bool
	// sceneRef 指向控件所属场景。
	sceneRef *Scene
	// parentRef 指向控件父容器。
	parentRef Container
}

// newWidgetBase 初始化具体控件共享的基础状态。
func newWidgetBase(id, prefix string) widgetBase {
	if id == "" {
		id = newWidgetID(prefix)
	}
	return widgetBase{
		id:      id,
		visible: true,
		enabled: true,
	}
}

// ID 返回控件基础对象的标识。
func (b *widgetBase) ID() string {
	return b.id
}

// Bounds 返回控件基础对象的绘制边界。
func (b *widgetBase) Bounds() Rect {
	return b.bounds
}

// Visible 返回控件基础对象的可见状态。
func (b *widgetBase) Visible() bool {
	return b.visible
}

// Enabled 返回控件基础对象的可用状态。
func (b *widgetBase) Enabled() bool {
	return b.enabled
}

// HitTest 判断给定点是否命中当前控件。
func (b *widgetBase) HitTest(x, y int32) bool {
	return b.visible && b.bounds.Contains(x, y)
}

// setScene 更新控件基础对象的场景引用。
func (b *widgetBase) setScene(scene *Scene) {
	b.sceneRef = scene
}

// scene 返回控件基础对象关联的场景。
func (b *widgetBase) scene() *Scene {
	return b.sceneRef
}

// setParent 更新控件基础对象的父容器。
func (b *widgetBase) setParent(parent Container) {
	b.parentRef = parent
}

// parent 返回控件基础对象的父容器。
func (b *widgetBase) parent() Container {
	return b.parentRef
}

// cursor 返回悬停控件时应使用的光标。
func (b *widgetBase) cursor() CursorID {
	return core.CursorArrow
}

// setBounds 更新控件基础对象的边界。
func (b *widgetBase) setBounds(owner Widget, rect Rect) {
	if b.bounds == rect {
		return
	}
	var oldRect Rect
	if b.sceneRef != nil && owner != nil {
		oldRect = widgetDirtyRect(owner)
	}
	b.bounds = rect
	if b.sceneRef != nil {
		if !oldRect.Empty() {
			b.sceneRef.invalidateRect(oldRect)
		}
		b.sceneRef.Invalidate(owner)
	}
}

// setVisible 更新控件基础对象的可见状态。
func (b *widgetBase) setVisible(owner Widget, visible bool) {
	if b.visible == visible {
		return
	}
	b.visible = visible
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// setEnabled 更新控件基础对象的可用状态。
func (b *widgetBase) setEnabled(owner Widget, enabled bool) {
	if b.enabled == enabled {
		return
	}
	b.enabled = enabled
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// runOnUI 在 UI 线程执行回调。
func (b *widgetBase) runOnUI(fn func()) {
	if fn == nil {
		return
	}
	if b.sceneRef == nil {
		fn()
		return
	}
	b.sceneRef.runOnUI(fn)
}

// invalidate 标记区域或控件需要重绘。
func (b *widgetBase) invalidate(owner Widget) {
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// asWidgetNode 在可用时将 Widget 转换为内部节点表示。
func asWidgetNode(widget Widget) widgetNode {
	if widget == nil {
		return nil
	}
	node, _ := widget.(widgetNode)
	return node
}

//go:build windows

package widgets

import (
	"fmt"
	"github.com/yourname/winui/core"
	"sync/atomic"
)

type Rect = core.Rect
type Color = core.Color
type CursorID = core.CursorID

type Widget interface {
	ID() string
	Bounds() Rect
	SetBounds(Rect)
	Visible() bool
	SetVisible(bool)
	Enabled() bool
	SetEnabled(bool)
	HitTest(x, y int32) bool
	OnEvent(evt Event) bool
	Paint(ctx *PaintCtx)
}

type widgetNode interface {
	Widget
	setScene(*Scene)
	scene() *Scene
	setParent(Container)
	parent() Container
	cursor() CursorID
}

type focusableWidget interface {
	Widget
	acceptsFocus() bool
}

type overlayWidget interface {
	Widget
	PaintOverlay(ctx *PaintCtx)
}

var widgetSequence atomic.Uint64

// newWidgetID 使用给定前缀生成唯一的控件标识。
func newWidgetID(prefix string) string {
	if prefix == "" {
		prefix = "widget"
	}
	return fmt.Sprintf("%s-%d", prefix, widgetSequence.Add(1))
}

type widgetBase struct {
	id        string
	bounds    Rect
	visible   bool
	enabled   bool
	sceneRef  *Scene
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
	b.bounds = rect
	if b.sceneRef != nil {
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

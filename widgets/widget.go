//go:build windows

package widgets

import (
	"fmt"
	"sync/atomic"

	"github.com/AzureIvory/winui/core"
)

// Rect 复用 core 中的矩形类型。
type Rect = core.Rect

// Color 复用 core 中的颜色类型。
type Color = core.Color

// CursorID 复用 core 中的光标标识类型。
type CursorID = core.CursorID

// Widget 定义所有控件都需要实现的基础行为。
type Widget interface {
	// ID 返回控件标识。
	ID() string
	// Bounds 返回控件当前边界。
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
	// LayoutData 返回布局附加数据。
	LayoutData() any
	// SetLayoutData 更新布局附加数据。
	SetLayoutData(any)
	// HitTest 判断指定坐标是否命中控件。
	HitTest(x, y int32) bool
	// OnEvent 处理控件事件。
	OnEvent(evt Event) bool
	// Paint 绘制控件内容。
	Paint(ctx *PaintCtx)
}

// widgetNode 定义控件树内部使用的场景和父子关系能力。
type widgetNode interface {
	Widget
	// setScene 绑定控件所在场景。
	setScene(*Scene)
	// scene 返回控件所在场景。
	scene() *Scene
	// setParent 绑定控件父容器。
	setParent(Container)
	// parent 返回控件父容器。
	parent() Container
	// cursor 返回控件当前希望使用的光标。
	cursor() CursorID
}

// focusableWidget 表示支持接收键盘焦点的控件。
type focusableWidget interface {
	Widget
	// acceptsFocus 返回控件是否允许被聚焦。
	acceptsFocus() bool
}

// overlayWidget 表示支持在常规绘制后追加覆盖层的控件。
type overlayWidget interface {
	Widget
	// PaintOverlay 绘制控件覆盖层内容。
	PaintOverlay(ctx *PaintCtx)
}

// dirtyWidget 表示支持自定义脏区域的控件。
type dirtyWidget interface {
	// dirtyRect 返回控件的实际脏区域。
	dirtyRect() Rect
}

// widgetSequence 为自动生成的控件标识提供递增序号。
var widgetSequence atomic.Uint64

// layoutPassDepth 记录当前布局过程的嵌套深度。
var layoutPassDepth atomic.Int32

// newWidgetID 返回一个稳定的自动生成控件标识。
func newWidgetID(prefix string) string {
	if prefix == "" {
		prefix = "widget"
	}
	return fmt.Sprintf("%s-%d", prefix, widgetSequence.Add(1))
}

// widgetBase 保存大部分控件都会复用的基础状态。
type widgetBase struct {
	// id 表示控件标识。
	id string
	// bounds 表示控件当前边界。
	bounds Rect
	// preferred 表示控件在布局前声明的首选尺寸。
	preferred core.Size
	// visible 表示控件是否可见。
	visible bool
	// enabled 表示控件是否可用。
	enabled bool
	// sceneRef 表示控件当前所属场景。
	sceneRef *Scene
	// parentRef 表示控件当前所属父容器。
	parentRef Container
	// layoutData 表示控件附带的布局数据。
	layoutData any
}

// newWidgetBase 初始化控件公共基础状态。
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

// ID 返回控件标识。
func (b *widgetBase) ID() string {
	return b.id
}

// Bounds 返回控件当前边界。
func (b *widgetBase) Bounds() Rect {
	return b.bounds
}

// preferredSize 返回控件在布局前声明的首选尺寸。
func (b *widgetBase) preferredSize() core.Size {
	return b.preferred
}

// Visible 返回控件是否可见。
func (b *widgetBase) Visible() bool {
	return b.visible
}

// Enabled 返回控件是否可用。
func (b *widgetBase) Enabled() bool {
	return b.enabled
}

// LayoutData 返回控件附带的布局数据。
func (b *widgetBase) LayoutData() any {
	return b.layoutData
}

// SetLayoutData 更新控件附带的布局数据并触发布局刷新。
func (b *widgetBase) SetLayoutData(data any) {
	b.layoutData = data
	if panel, ok := b.parentRef.(*Panel); ok {
		panel.applyLayout()
		panel.invalidate(panel)
		return
	}
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(nil)
	}
}

// HitTest 判断给定坐标是否命中当前控件。
func (b *widgetBase) HitTest(x, y int32) bool {
	return b.visible && b.bounds.Contains(x, y)
}

// setScene 更新控件关联的场景引用。
func (b *widgetBase) setScene(scene *Scene) {
	b.sceneRef = scene
}

// scene 返回控件当前关联的场景引用。
func (b *widgetBase) scene() *Scene {
	return b.sceneRef
}

// setParent 更新控件关联的父容器。
func (b *widgetBase) setParent(parent Container) {
	b.parentRef = parent
}

// parent 返回控件当前的父容器。
func (b *widgetBase) parent() Container {
	return b.parentRef
}

// cursor 返回控件默认使用的光标。
func (b *widgetBase) cursor() CursorID {
	return core.CursorArrow
}

// setBounds 更新控件边界并在必要时同步首选尺寸。
func (b *widgetBase) setBounds(owner Widget, rect Rect) {
	if b.bounds == rect {
		return
	}
	if layoutPassDepth.Load() == 0 {
		b.preferred = core.Size{Width: rect.W, Height: rect.H}
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

// setVisible 更新控件可见状态并请求重绘。
func (b *widgetBase) setVisible(owner Widget, visible bool) {
	if b.visible == visible {
		return
	}
	b.visible = visible
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// setEnabled 更新控件可用状态并请求重绘。
func (b *widgetBase) setEnabled(owner Widget, enabled bool) {
	if b.enabled == enabled {
		return
	}
	b.enabled = enabled
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// runOnUI 在 UI 线程上执行给定回调。
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

// invalidate 请求当前控件重绘。
func (b *widgetBase) invalidate(owner Widget) {
	if b.sceneRef != nil {
		b.sceneRef.Invalidate(owner)
	}
}

// asWidgetNode 把普通控件转换为内部使用的节点接口。
func asWidgetNode(widget Widget) widgetNode {
	if widget == nil {
		return nil
	}
	node, _ := widget.(widgetNode)
	return node
}

// beginLayoutPass 标记开始一次布局过程。
func beginLayoutPass() {
	layoutPassDepth.Add(1)
}

// endLayoutPass 标记结束一次布局过程。
func endLayoutPass() {
	layoutPassDepth.Add(-1)
}

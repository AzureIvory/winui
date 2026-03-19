//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// Container 定义可以持有子控件的容器行为。
type Container interface {
	Widget
	// Add 向容器添加子控件。
	Add(child Widget)
	// Remove 按标识移除子控件。
	Remove(id string)
	// Children 返回当前子控件列表。
	Children() []Widget
}

// PanelStyle 描述面板控件的外观样式。
type PanelStyle struct {
	// Background 指定面板背景色。
	Background core.Color
	// BorderColor 指定面板边框颜色。
	BorderColor core.Color
	// CornerRadius 指定面板圆角半径。
	CornerRadius int32
	// BorderWidth 指定边框宽度。
	BorderWidth int32
}

// Panel 表示可承载其他控件的基础面板。
type Panel struct {
	// widgetBase 提供面板共享的基础控件能力。
	widgetBase
	// children 保存面板当前的子控件。
	children []Widget
	// layout 保存子控件布局策略。
	layout Layout
	// Style 保存面板的样式覆盖。
	Style PanelStyle
	// OnClick 保存面板点击回调。
	OnClick func()
}

// NewPanel 创建一个新的面板。
func NewPanel(id string) *Panel {
	return &Panel{
		widgetBase: newWidgetBase(id, "panel"),
		layout:     AbsoluteLayout{},
	}
}

// SetBounds 更新面板的边界。
func (p *Panel) SetBounds(rect Rect) {
	p.widgetBase.setBounds(p, rect)
	p.applyLayout()
}

// SetVisible 更新面板的可见状态。
func (p *Panel) SetVisible(visible bool) {
	p.widgetBase.setVisible(p, visible)
}

// SetEnabled 更新面板的可用状态。
func (p *Panel) SetEnabled(enabled bool) {
	p.widgetBase.setEnabled(p, enabled)
}

// Add 向面板添加子控件。
func (p *Panel) Add(child Widget) {
	if child == nil {
		return
	}
	node := asWidgetNode(child)
	if node == nil {
		return
	}

	p.children = append(p.children, child)
	node.setParent(p)
	node.setScene(p.scene())
	if container, ok := child.(Container); ok {
		attachSceneRecursive(container, p.scene())
	}
	p.applyLayout()
	p.invalidate(p)
}

// AddAll 按顺序向面板追加多个子控件。
func (p *Panel) AddAll(children ...Widget) {
	for _, child := range children {
		p.Add(child)
	}
}

// Remove 从面板移除子控件。
func (p *Panel) Remove(id string) {
	for i, child := range p.children {
		if child.ID() != id {
			continue
		}
		if scene := p.scene(); scene != nil {
			scene.disposeTree(child)
		}
		if node := asWidgetNode(child); node != nil {
			node.setParent(nil)
			node.setScene(nil)
		}
		p.children = append(p.children[:i], p.children[i+1:]...)
		p.applyLayout()
		p.invalidate(p)
		return
	}
}

// Children 返回面板的子控件列表。
func (p *Panel) Children() []Widget {
	out := make([]Widget, len(p.children))
	copy(out, p.children)
	return out
}

// SetLayout 更新面板的布局。
func (p *Panel) SetLayout(layout Layout) {
	if layout == nil {
		layout = AbsoluteLayout{}
	}
	p.layout = layout
	p.applyLayout()
	p.invalidate(p)
}

// SetStyle 更新面板背景和边框样式。
func (p *Panel) SetStyle(style PanelStyle) {
	p.Style = style
	p.invalidate(p)
}

// SetOnClick 注册面板点击回调。
func (p *Panel) SetOnClick(fn func()) {
	p.OnClick = fn
}

// OnEvent 处理输入事件或生命周期事件。
func (p *Panel) OnEvent(evt Event) bool {
	if evt.Type == EventClick && p.OnClick != nil {
		p.OnClick()
		return true
	}
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (p *Panel) Paint(ctx *PaintCtx) {
	if !p.Visible() {
		return
	}
	if ctx != nil {
		radius := ctx.DP(p.Style.CornerRadius)
		if p.Style.Background != 0 {
			if radius > 0 {
				_ = ctx.FillRoundRect(p.Bounds(), radius, p.Style.Background)
			} else {
				_ = ctx.FillRect(p.Bounds(), p.Style.Background)
			}
		}
		if p.Style.BorderColor != 0 {
			width := p.Style.BorderWidth
			if width <= 0 {
				width = 1
			}
			if radius > 0 {
				_ = ctx.StrokeRoundRect(p.Bounds(), radius, p.Style.BorderColor, width)
			}
		}
	}
	for _, child := range p.children {
		if child.Visible() {
			child.Paint(ctx)
		}
	}
}

// setScene 更新面板关联的场景。
func (p *Panel) setScene(scene *Scene) {
	p.widgetBase.setScene(scene)
	for _, child := range p.children {
		if node := asWidgetNode(child); node != nil {
			node.setScene(scene)
		}
		if container, ok := child.(Container); ok {
			attachSceneRecursive(container, scene)
		}
	}
}

// applyLayout 应用面板布局。
func (p *Panel) applyLayout() {
	if p.layout == nil {
		return
	}
	beginLayoutPass()
	defer endLayoutPass()
	p.layout.Apply(p.Bounds(), p.children)
}

// attachSceneRecursive 递归关联场景引用。
func attachSceneRecursive(container Container, scene *Scene) {
	node := asWidgetNode(container)
	if node != nil {
		node.setScene(scene)
	}
	for _, child := range container.Children() {
		if childNode := asWidgetNode(child); childNode != nil {
			childNode.setScene(scene)
		}
		if childContainer, ok := child.(Container); ok {
			attachSceneRecursive(childContainer, scene)
		}
	}
}

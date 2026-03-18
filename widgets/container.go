//go:build windows

package widgets

type Container interface {
	Widget
	Add(child Widget)
	Remove(id string)
	Children() []Widget
}

type Panel struct {
	widgetBase
	children []Widget
	layout   Layout
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

// OnEvent 处理输入事件或生命周期事件。
func (p *Panel) OnEvent(Event) bool {
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (p *Panel) Paint(ctx *PaintCtx) {
	if !p.Visible() {
		return
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

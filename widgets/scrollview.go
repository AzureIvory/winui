//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ScrollView 表示带裁剪视口的滚动容器。
type ScrollView struct {
	// widgetBase 提供滚动容器共享的基础控件能力。
	widgetBase
	// content 保存当前内容节点。
	content Widget
	// Style 保存滚动容器样式覆盖。
	Style PanelStyle
	// Focused 记录滚动容器当前是否拥有焦点。
	Focused bool
	// offsetX 保存当前水平滚动偏移。
	offsetX int32
	// offsetY 保存当前垂直滚动偏移。
	offsetY int32
	// maxOffsetX 保存当前可滚动的最大水平偏移。
	maxOffsetX int32
	// maxOffsetY 保存当前可滚动的最大垂直偏移。
	maxOffsetY int32
	// wheelStep 保存单次滚轮滚动步长。
	wheelStep int32
	// verticalScroll 控制是否处理垂直滚动输入。
	verticalScroll bool
	// horizontalScroll 控制是否处理水平滚动输入。
	horizontalScroll bool
}

// NewScrollView 创建一个新的滚动容器。
func NewScrollView(id string) *ScrollView {
	return &ScrollView{
		widgetBase:       newWidgetBase(id, "scroll"),
		wheelStep:        48,
		verticalScroll:   true,
		horizontalScroll: false,
	}
}

// SetBounds 更新滚动容器的边界。
func (s *ScrollView) SetBounds(rect Rect) {
	s.runOnUI(func() {
		s.widgetBase.setBounds(s, rect)
		s.layoutContent()
		s.invalidate(s)
	})
}

// SetVisible 更新滚动容器可见状态。
func (s *ScrollView) SetVisible(visible bool) {
	s.runOnUI(func() {
		s.widgetBase.setVisible(s, visible)
	})
}

// SetEnabled 更新滚动容器可用状态。
func (s *ScrollView) SetEnabled(enabled bool) {
	s.runOnUI(func() {
		s.widgetBase.setEnabled(s, enabled)
	})
}

// SetStyle 更新滚动容器样式覆盖。
func (s *ScrollView) SetStyle(style PanelStyle) {
	s.runOnUI(func() {
		s.Style = style
		s.invalidate(s)
	})
}

// SetContent 更新滚动容器的内容节点。
func (s *ScrollView) SetContent(widget Widget) {
	s.runOnUI(func() {
		s.replaceContent(widget)
	})
}

// Content 返回当前内容节点。
func (s *ScrollView) Content() Widget {
	return s.content
}

// SetScrollOffset 更新滚动偏移。
func (s *ScrollView) SetScrollOffset(x, y int32) {
	s.ScrollTo(x, y)
}

// ScrollOffset 返回当前滚动偏移。
func (s *ScrollView) ScrollOffset() (int32, int32) {
	return s.offsetX, s.offsetY
}

// ScrollBy 按增量滚动内容。
func (s *ScrollView) ScrollBy(dx, dy int32) {
	s.runOnUI(func() {
		s.layoutContent()
		s.applyOffset(s.offsetX+dx, s.offsetY+dy)
	})
}

// ScrollTo 滚动到给定偏移。
func (s *ScrollView) ScrollTo(x, y int32) {
	s.runOnUI(func() {
		s.layoutContent()
		s.applyOffset(x, y)
	})
}

// SetWheelStep 更新滚轮步长。
func (s *ScrollView) SetWheelStep(step int32) {
	s.runOnUI(func() {
		if step < 0 {
			step = 0
		}
		s.wheelStep = step
	})
}

// SetVerticalScroll 更新垂直滚动支持状态。
func (s *ScrollView) SetVerticalScroll(enabled bool) {
	s.runOnUI(func() {
		if s.verticalScroll == enabled {
			return
		}
		s.verticalScroll = enabled
		s.invalidate(s)
	})
}

// VerticalScroll 返回是否启用垂直滚动。
func (s *ScrollView) VerticalScroll() bool {
	return s.verticalScroll
}

// SetHorizontalScroll 更新水平滚动支持状态。
func (s *ScrollView) SetHorizontalScroll(enabled bool) {
	s.runOnUI(func() {
		if s.horizontalScroll == enabled {
			return
		}
		s.horizontalScroll = enabled
		s.invalidate(s)
	})
}

// HorizontalScroll 返回是否启用水平滚动。
func (s *ScrollView) HorizontalScroll() bool {
	return s.horizontalScroll
}

// Add 将子控件加入滚动容器内容中。
func (s *ScrollView) Add(child Widget) {
	if child == nil {
		return
	}
	s.runOnUI(func() {
		if s.content == nil {
			s.replaceContent(child)
			return
		}
		if container, ok := s.content.(Container); ok {
			container.Add(child)
			s.layoutContent()
			s.invalidate(s)
			return
		}
		existing := s.content
		s.content = nil
		if node := asWidgetNode(existing); node != nil {
			node.setParent(nil)
		}
		wrapper := NewPanel(s.ID() + "-content")
		wrapper.SetLayout(ColumnLayout{Gap: 8})
		wrapper.Add(existing)
		wrapper.Add(child)
		s.replaceContent(wrapper)
	})
}

// Remove 从滚动容器中移除指定 ID 的内容或其后代。
func (s *ScrollView) Remove(id string) {
	if id == "" || s.content == nil {
		return
	}
	s.runOnUI(func() {
		if s.content == nil {
			return
		}
		if s.content.ID() == id {
			s.replaceContent(nil)
			return
		}
		if container, ok := s.content.(Container); ok {
			container.Remove(id)
			s.layoutContent()
			s.invalidate(s)
		}
	})
}

// Children 返回滚动容器内容节点切片。
func (s *ScrollView) Children() []Widget {
	if s.content == nil {
		return nil
	}
	return []Widget{s.content}
}

// HitTest 判断给定点是否命中滚动容器视口。
func (s *ScrollView) HitTest(x, y int32) bool {
	return s.Visible() && s.Bounds().Contains(x, y)
}

// OnEvent 处理输入和焦点事件。
func (s *ScrollView) OnEvent(evt Event) bool {
	s.layoutContent()
	switch evt.Type {
	case EventFocus:
		s.Focused = true
		return false
	case EventBlur:
		s.Focused = false
		return false
	case EventMouseWheel:
		return s.handleWheel(evt)
	case EventKeyDown:
		switch evt.Key.Key {
		case 0x21: // VK_PRIOR / PageUp
			if s.maxOffsetY > 0 {
				s.ScrollBy(0, -max32(1, s.Bounds().H))
				return true
			}
		case 0x22: // VK_NEXT / PageDown
			if s.maxOffsetY > 0 {
				s.ScrollBy(0, max32(1, s.Bounds().H))
				return true
			}
		case core.KeyHome:
			if s.maxOffsetY > 0 && s.Focused {
				s.ScrollTo(s.offsetX, 0)
				return true
			}
		case core.KeyEnd:
			if s.maxOffsetY > 0 && s.Focused {
				s.ScrollTo(s.offsetX, s.maxOffsetY)
				return true
			}
		}
	}
	return false
}

// Paint 使用给定绘制上下文完成滚动容器绘制。
func (s *ScrollView) Paint(ctx *PaintCtx) {
	if !s.Visible() || ctx == nil {
		return
	}
	bounds := s.Bounds()
	if bounds.Empty() {
		return
	}
	radius := ctx.DP(s.Style.CornerRadius)
	if s.Style.Background != 0 {
		if radius > 0 {
			_ = ctx.FillRoundRect(bounds, radius, s.Style.Background)
		} else {
			_ = ctx.FillRect(bounds, s.Style.Background)
		}
	}
	if s.content != nil {
		s.layoutContent()
		restore := ctx.PushClipRect(s.viewportRect())
		s.content.Paint(ctx)
		restore()
	}
	if s.Style.BorderColor != 0 {
		width := s.Style.BorderWidth
		if width <= 0 {
			width = 1
		}
		_ = ctx.StrokeRoundRect(bounds, max32(0, radius), s.Style.BorderColor, width)
	}
}

// dirtyRect 返回滚动容器脏区。
func (s *ScrollView) dirtyRect() Rect {
	return s.Bounds()
}

// acceptsFocus 表示滚动容器可接收键盘焦点。
func (s *ScrollView) acceptsFocus() bool {
	return true
}

// cursor 返回滚动容器希望使用的光标。
func (s *ScrollView) cursor() CursorID {
	return core.CursorArrow
}

// clipBounds 返回滚动容器对子树生效的裁剪区域。
func (s *ScrollView) clipBounds() Rect {
	return s.viewportRect()
}

// setScene 更新滚动容器关联的场景。
func (s *ScrollView) setScene(scene *Scene) {
	s.widgetBase.setScene(scene)
	if node := asWidgetNode(s.content); node != nil {
		node.setScene(scene)
	}
	if container, ok := s.content.(Container); ok {
		attachSceneRecursive(container, scene)
	}
}

func (s *ScrollView) replaceContent(widget Widget) {
	if s.content == widget {
		s.layoutContent()
		s.invalidate(s)
		return
	}
	if old := s.content; old != nil {
		if scene := s.scene(); scene != nil {
			scene.disposeTree(old)
		}
		if node := asWidgetNode(old); node != nil {
			node.setParent(nil)
			node.setScene(nil)
		}
	}
	s.content = nil
	if widget != nil {
		node := asWidgetNode(widget)
		if node != nil {
			node.setParent(s)
			node.setScene(s.scene())
		}
		if container, ok := widget.(Container); ok {
			attachSceneRecursive(container, s.scene())
		}
		s.content = widget
	}
	s.offsetX = 0
	s.offsetY = 0
	s.maxOffsetX = 0
	s.maxOffsetY = 0
	s.layoutContent()
	s.invalidate(s)
}

func (s *ScrollView) viewportRect() Rect {
	return s.Bounds()
}

func (s *ScrollView) layoutContent() {
	viewport := s.viewportRect()
	if viewport.Empty() || s.content == nil {
		s.maxOffsetX = 0
		s.maxOffsetY = 0
		return
	}
	size := measureWidgetNatural(s.content)
	if size.Width < viewport.W {
		size.Width = viewport.W
	}
	if size.Height < viewport.H {
		size.Height = viewport.H
	}
	s.maxOffsetX = max32(0, size.Width-viewport.W)
	s.maxOffsetY = max32(0, size.Height-viewport.H)
	s.offsetX = clampValue(s.offsetX, 0, s.maxOffsetX)
	s.offsetY = clampValue(s.offsetY, 0, s.maxOffsetY)
	target := Rect{
		X: viewport.X - s.offsetX,
		Y: viewport.Y - s.offsetY,
		W: size.Width,
		H: size.Height,
	}
	if s.content.Bounds() != target {
		s.content.SetBounds(target)
		return
	}
	if panel, ok := s.content.(*Panel); ok {
		panel.applyLayout()
	}
}

func (s *ScrollView) applyOffset(x, y int32) {
	x = clampValue(x, 0, s.maxOffsetX)
	y = clampValue(y, 0, s.maxOffsetY)
	if s.offsetX == x && s.offsetY == y {
		return
	}
	s.offsetX = x
	s.offsetY = y
	s.layoutContent()
	s.invalidate(s)
}

func (s *ScrollView) handleWheel(evt Event) bool {
	step := s.wheelStep
	if step <= 0 {
		step = 48
	}
	if s.verticalScroll && s.maxOffsetY > 0 {
		delta := -evt.Delta * step / 120
		if delta == 0 {
			if evt.Delta > 0 {
				delta = -step
			} else if evt.Delta < 0 {
				delta = step
			}
		}
		s.ScrollBy(0, delta)
		return true
	}
	if s.horizontalScroll && s.maxOffsetX > 0 {
		delta := -evt.Delta * step / 120
		if delta == 0 {
			if evt.Delta > 0 {
				delta = -step
			} else if evt.Delta < 0 {
				delta = step
			}
		}
		s.ScrollBy(delta, 0)
		return true
	}
	return false
}

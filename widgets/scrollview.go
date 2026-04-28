//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

type scrollDragAxis uint8

const (
	scrollDragNone scrollDragAxis = iota
	scrollDragVertical
	scrollDragHorizontal
)

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
	// hoverVerticalThumb 记录垂直滚动条滑块是否处于悬停状态。
	hoverVerticalThumb bool
	// hoverHorizontalThumb 记录水平滚动条滑块是否处于悬停状态。
	hoverHorizontalThumb bool
	// dragAxis 记录当前拖拽的滚动轴。
	dragAxis scrollDragAxis
	// dragPointer 保存拖拽开始时的鼠标坐标。
	dragPointer int32
	// dragOffset 保存拖拽开始时的滚动偏移。
	dragOffset int32
	// layoutDirty 标记内容自然尺寸是否需要重新测量。
	layoutDirty bool
	// measuredViewport 保存最近一次测量使用的视口尺寸。
	measuredViewport core.Size
	// measuredContent 保存最近一次稳定测量得到的内容尺寸。
	measuredContent core.Size
}

// NewScrollView 创建一个新的滚动容器。
func NewScrollView(id string) *ScrollView {
	return &ScrollView{
		widgetBase:       newWidgetBase(id, "scroll"),
		wheelStep:        48,
		verticalScroll:   true,
		horizontalScroll: false,
		layoutDirty:      true,
	}
}

// SetBounds 更新滚动容器的边界。
func (s *ScrollView) SetBounds(rect Rect) {
	s.runOnUI(func() {
		if bounds := s.Bounds(); bounds.W != rect.W || bounds.H != rect.H {
			s.invalidateLayoutCache()
		}
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
		s.applyOffset(s.offsetX+dx, s.offsetY+dy)
	})
}

// ScrollTo 滚动到给定偏移。
func (s *ScrollView) ScrollTo(x, y int32) {
	s.runOnUI(func() {
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
		if !enabled {
			s.hoverVerticalThumb = false
			if s.dragAxis == scrollDragVertical {
				s.dragAxis = scrollDragNone
				s.dragPointer = 0
				s.dragOffset = 0
			}
			s.syncCursor()
		}
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
		if !enabled {
			s.hoverHorizontalThumb = false
			if s.dragAxis == scrollDragHorizontal {
				s.dragAxis = scrollDragNone
				s.dragPointer = 0
				s.dragOffset = 0
			}
			s.syncCursor()
		}
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
			s.invalidateLayoutCache()
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
			s.invalidateLayoutCache()
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
	switch evt.Type {
	case EventFocus:
		s.Focused = true
		return false
	case EventBlur:
		s.Focused = false
		return false
	case EventMouseMove:
		return s.handleMouseMove(evt)
	case EventMouseLeave:
		return s.handleMouseLeave()
	case EventMouseDown:
		return s.handleMouseDown(evt)
	case EventMouseUp:
		return s.handleMouseUp(evt)
	case EventMouseWheel:
		return s.handleWheel(evt)
	case EventKeyDown:
		s.layoutContent()
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
	s.paintScrollbars(ctx)
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
	if s.dragAxis != scrollDragNone || s.hoverVerticalThumb || s.hoverHorizontalThumb {
		return core.CursorHand
	}
	return core.CursorArrow
}

func (s *ScrollView) overlayHitTest(x, y int32) bool {
	vTrack, vThumb, hTrack, hThumb := s.scrollbarRectsForWidget()
	if !vTrack.Empty() && vTrack.Contains(x, y) {
		return true
	}
	if !vThumb.Empty() && vThumb.Contains(x, y) {
		return true
	}
	if !hTrack.Empty() && hTrack.Contains(x, y) {
		return true
	}
	if !hThumb.Empty() && hThumb.Contains(x, y) {
		return true
	}
	return false
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
	s.invalidateLayoutCache()
	s.hoverVerticalThumb = false
	s.hoverHorizontalThumb = false
	s.dragAxis = scrollDragNone
	s.dragPointer = 0
	s.dragOffset = 0
	s.layoutContent()
	s.invalidate(s)
}

func (s *ScrollView) viewportRect() Rect {
	return s.Bounds()
}

func (s *ScrollView) layoutContent() {
	s.layoutContentForOffsets(s.offsetX, s.offsetY)
}

func (s *ScrollView) layoutContentForOffsets(requestedX, requestedY int32) {
	viewport := s.viewportRect()
	if viewport.Empty() || s.content == nil {
		s.offsetX = 0
		s.offsetY = 0
		s.maxOffsetX = 0
		s.maxOffsetY = 0
		s.invalidateLayoutCache()
		return
	}
	if s.canReuseMeasuredLayout(viewport) {
		s.applyMeasuredLayout(viewport, requestedX, requestedY, s.measuredContent)
		return
	}

	// Width-constrained children can grow the content height after the first layout pass.
	// Iterate until the measured content size stabilizes so max offsets stay in sync.
	size := s.contentSizeForViewport(viewport)
	for pass := 0; pass < 4; pass++ {
		s.applyMeasuredLayout(viewport, requestedX, requestedY, size)
		measured := s.measuredContentSize(viewport, size)
		if measured == size {
			s.measuredViewport = core.Size{Width: viewport.W, Height: viewport.H}
			s.measuredContent = size
			s.layoutDirty = false
			return
		}
		size = measured
	}

	s.applyMeasuredLayout(viewport, requestedX, requestedY, size)
	s.measuredViewport = core.Size{Width: viewport.W, Height: viewport.H}
	s.measuredContent = size
	s.layoutDirty = false
}

func (s *ScrollView) positionContent(target Rect) {
	if s == nil || s.content == nil {
		return
	}
	current := s.content.Bounds()
	if current == target {
		return
	}
	if current.W == target.W && current.H == target.H {
		translateWidgetTree(s.content, target.X-current.X, target.Y-current.Y)
		return
	}
	beginSuppressedInvalidation()
	s.content.SetBounds(target)
	endSuppressedInvalidation()
}

func (s *ScrollView) contentSizeForViewport(viewport Rect) core.Size {
	size := measureWidgetNatural(s.content)
	if size.Width < viewport.W {
		size.Width = viewport.W
	}
	if size.Height < viewport.H {
		size.Height = viewport.H
	}
	return size
}

func (s *ScrollView) measuredContentSize(viewport Rect, size core.Size) core.Size {
	container, ok := s.content.(Container)
	if !ok {
		return size
	}
	bounds := s.content.Bounds()
	right := bounds.X
	bottom := bounds.Y
	for _, child := range container.Children() {
		if child == nil {
			continue
		}
		childBounds := child.Bounds()
		if edge := childBounds.X + childBounds.W; edge > right {
			right = edge
		}
		if edge := childBounds.Y + childBounds.H; edge > bottom {
			bottom = edge
		}
	}
	if width := right - bounds.X; width > size.Width {
		size.Width = width
	}
	if height := bottom - bounds.Y; height > size.Height {
		size.Height = height
	}
	return size
}

func (s *ScrollView) applyOffset(x, y int32) {
	oldX := s.offsetX
	oldY := s.offsetY
	s.layoutContentForOffsets(x, y)
	if s.offsetX == oldX && s.offsetY == oldY {
		return
	}
	s.invalidate(s)
}

func (s *ScrollView) handleMouseMove(evt Event) bool {
	if s.dragAxis != scrollDragNone {
		handled := s.updateDrag(evt.Point)
		s.updateScrollbarHover(evt.Point)
		s.syncCursor()
		return handled
	}
	if s.updateScrollbarHover(evt.Point) {
		s.invalidate(s)
		s.syncCursor()
	}
	return s.hoverVerticalThumb || s.hoverHorizontalThumb
}

func (s *ScrollView) handleMouseLeave() bool {
	if s.dragAxis != scrollDragNone {
		return true
	}
	if !s.hoverVerticalThumb && !s.hoverHorizontalThumb {
		return false
	}
	s.hoverVerticalThumb = false
	s.hoverHorizontalThumb = false
	s.invalidate(s)
	s.syncCursor()
	return true
}

func (s *ScrollView) handleMouseDown(evt Event) bool {
	if evt.Button != core.MouseButtonLeft {
		return false
	}
	vTrack, vThumb, hTrack, hThumb := s.scrollbarRectsForWidget()
	if !vThumb.Empty() && vThumb.Contains(evt.Point.X, evt.Point.Y) {
		s.dragAxis = scrollDragVertical
		s.dragPointer = evt.Point.Y
		s.dragOffset = s.offsetY
		s.updateScrollbarHover(evt.Point)
		s.invalidate(s)
		s.syncCursor()
		return true
	}
	if !hThumb.Empty() && hThumb.Contains(evt.Point.X, evt.Point.Y) {
		s.dragAxis = scrollDragHorizontal
		s.dragPointer = evt.Point.X
		s.dragOffset = s.offsetX
		s.updateScrollbarHover(evt.Point)
		s.invalidate(s)
		s.syncCursor()
		return true
	}
	if !vTrack.Empty() && vTrack.Contains(evt.Point.X, evt.Point.Y) {
		page := max32(1, s.viewportRect().H)
		if evt.Point.Y < vThumb.Y {
			s.applyOffset(s.offsetX, s.offsetY-page)
		} else if evt.Point.Y > vThumb.Y+vThumb.H {
			s.applyOffset(s.offsetX, s.offsetY+page)
		}
		s.updateScrollbarHover(evt.Point)
		s.syncCursor()
		return true
	}
	if !hTrack.Empty() && hTrack.Contains(evt.Point.X, evt.Point.Y) {
		page := max32(1, s.viewportRect().W)
		if evt.Point.X < hThumb.X {
			s.applyOffset(s.offsetX-page, s.offsetY)
		} else if evt.Point.X > hThumb.X+hThumb.W {
			s.applyOffset(s.offsetX+page, s.offsetY)
		}
		s.updateScrollbarHover(evt.Point)
		s.syncCursor()
		return true
	}
	return false
}

func (s *ScrollView) handleMouseUp(evt Event) bool {
	if evt.Button != core.MouseButtonLeft {
		return false
	}
	if s.dragAxis == scrollDragNone {
		return false
	}
	s.dragAxis = scrollDragNone
	s.dragPointer = 0
	s.dragOffset = 0
	s.updateScrollbarHover(evt.Point)
	s.invalidate(s)
	s.syncCursor()
	return true
}

func (s *ScrollView) updateDrag(point core.Point) bool {
	vTrack, vThumb, hTrack, hThumb := s.scrollbarRectsForWidget()
	switch s.dragAxis {
	case scrollDragVertical:
		if vTrack.Empty() || vThumb.Empty() || s.maxOffsetY <= 0 {
			return true
		}
		travel := max32(0, vTrack.H-vThumb.H)
		if travel <= 0 {
			return true
		}
		delta := point.Y - s.dragPointer
		target := clampValue(s.dragOffset+int32(int64(delta)*int64(s.maxOffsetY)/int64(travel)), 0, s.maxOffsetY)
		if target != s.offsetY {
			s.applyOffset(s.offsetX, target)
		}
		return true
	case scrollDragHorizontal:
		if hTrack.Empty() || hThumb.Empty() || s.maxOffsetX <= 0 {
			return true
		}
		travel := max32(0, hTrack.W-hThumb.W)
		if travel <= 0 {
			return true
		}
		delta := point.X - s.dragPointer
		target := clampValue(s.dragOffset+int32(int64(delta)*int64(s.maxOffsetX)/int64(travel)), 0, s.maxOffsetX)
		if target != s.offsetX {
			s.applyOffset(target, s.offsetY)
		}
		return true
	default:
		return false
	}
}

func (s *ScrollView) updateScrollbarHover(point core.Point) bool {
	_, vThumb, _, hThumb := s.scrollbarRectsForWidget()
	hoverV := !vThumb.Empty() && vThumb.Contains(point.X, point.Y)
	hoverH := !hThumb.Empty() && hThumb.Contains(point.X, point.Y)
	if s.hoverVerticalThumb == hoverV && s.hoverHorizontalThumb == hoverH {
		return false
	}
	s.hoverVerticalThumb = hoverV
	s.hoverHorizontalThumb = hoverH
	return true
}

func (s *ScrollView) syncCursor() {
	if scene := s.scene(); scene != nil && scene.app != nil {
		scene.app.SetCursor(s.cursor())
	}
}

func (s *ScrollView) handleWheel(evt Event) bool {
	step := s.wheelStep
	if step <= 0 {
		step = 48
	}
	delta := -evt.Delta * step / 120
	if delta == 0 {
		if evt.Delta > 0 {
			delta = -step
		} else if evt.Delta < 0 {
			delta = step
		}
	}
	const mouseKeyShift = 0x0004
	preferHorizontal := evt.Flags&mouseKeyShift != 0
	if preferHorizontal {
		if s.scrollAlongAxis(delta, true) {
			return true
		}
		if s.scrollAlongAxis(delta, false) {
			return true
		}
		return false
	}
	if s.scrollAlongAxis(delta, false) {
		return true
	}
	if s.scrollAlongAxis(delta, true) {
		return true
	}
	return false
}

func (s *ScrollView) scrollAlongAxis(delta int32, horizontal bool) bool {
	if delta == 0 {
		return false
	}
	if horizontal {
		if !s.horizontalScroll || s.maxOffsetX <= 0 {
			return false
		}
		target := clampValue(s.offsetX+delta, 0, s.maxOffsetX)
		if target == s.offsetX {
			return false
		}
		s.applyOffset(target, s.offsetY)
		return true
	}
	if !s.verticalScroll || s.maxOffsetY <= 0 {
		return false
	}
	target := clampValue(s.offsetY+delta, 0, s.maxOffsetY)
	if target == s.offsetY {
		return false
	}
	s.applyOffset(s.offsetX, target)
	return true
}

func (s *ScrollView) paintScrollbars(ctx *PaintCtx) {
	if ctx == nil {
		return
	}
	vTrack, vThumb, hTrack, hThumb := s.scrollbarRects(ctx.DP)
	trackColor, thumbColor := s.scrollbarColors()
	radius := int32(0)
	if vTrack.W > 0 {
		radius = max32(1, vTrack.W/2)
	}
	if vTrack.W > 0 && vTrack.H > 0 {
		_ = ctx.FillRoundRect(vTrack, radius, trackColor)
	}
	if vThumb.W > 0 && vThumb.H > 0 {
		_ = ctx.FillRoundRect(vThumb, radius, thumbColor)
	}
	if hTrack.H > 0 {
		radius = max32(1, hTrack.H/2)
	}
	if hTrack.W > 0 && hTrack.H > 0 {
		_ = ctx.FillRoundRect(hTrack, radius, trackColor)
	}
	if hThumb.W > 0 && hThumb.H > 0 {
		_ = ctx.FillRoundRect(hThumb, radius, thumbColor)
	}
}

func (s *ScrollView) scrollbarRectsForWidget() (Rect, Rect, Rect, Rect) {
	s.layoutContent()
	return s.scrollbarRects(func(value int32) int32 {
		return widgetDP(s, value)
	})
}

func (s *ScrollView) invalidateLayoutCache() {
	s.layoutDirty = true
	s.measuredViewport = core.Size{}
	s.measuredContent = core.Size{}
}

func (s *ScrollView) canReuseMeasuredLayout(viewport Rect) bool {
	return !s.layoutDirty &&
		s.measuredViewport.Width == viewport.W &&
		s.measuredViewport.Height == viewport.H &&
		s.measuredContent.Width > 0 &&
		s.measuredContent.Height > 0
}

func (s *ScrollView) applyMeasuredLayout(viewport Rect, requestedX, requestedY int32, size core.Size) {
	s.maxOffsetX = max32(0, size.Width-viewport.W)
	s.maxOffsetY = max32(0, size.Height-viewport.H)
	s.offsetX = clampValue(requestedX, 0, s.maxOffsetX)
	s.offsetY = clampValue(requestedY, 0, s.maxOffsetY)
	s.positionContent(Rect{
		X: viewport.X - s.offsetX,
		Y: viewport.Y - s.offsetY,
		W: size.Width,
		H: size.Height,
	})
}

func (s *ScrollView) scrollbarRects(scale func(int32) int32) (Rect, Rect, Rect, Rect) {
	if s == nil || scale == nil {
		return Rect{}, Rect{}, Rect{}, Rect{}
	}
	viewport := s.viewportRect()
	if viewport.Empty() {
		return Rect{}, Rect{}, Rect{}, Rect{}
	}

	thickness := max32(6, scale(8))
	margin := max32(2, scale(4))
	minThumb := max32(18, scale(24))

	showV := s.verticalScroll && s.maxOffsetY > 0
	showH := s.horizontalScroll && s.maxOffsetX > 0

	var verticalTrack Rect
	if showV {
		verticalTrack = Rect{
			X: viewport.X + viewport.W - margin - thickness,
			Y: viewport.Y + margin,
			W: thickness,
			H: max32(0, viewport.H-margin*2),
		}
		if showH {
			verticalTrack.H = max32(0, verticalTrack.H-thickness-margin)
		}
	}

	var horizontalTrack Rect
	if showH {
		horizontalTrack = Rect{
			X: viewport.X + margin,
			Y: viewport.Y + viewport.H - margin - thickness,
			W: max32(0, viewport.W-margin*2),
			H: thickness,
		}
		if showV {
			horizontalTrack.W = max32(0, horizontalTrack.W-thickness-margin)
		}
	}

	verticalThumb := scrollbarThumbRect(verticalTrack, minThumb, viewport.H+s.maxOffsetY, viewport.H, s.offsetY, s.maxOffsetY, true)
	horizontalThumb := scrollbarThumbRect(horizontalTrack, minThumb, viewport.W+s.maxOffsetX, viewport.W, s.offsetX, s.maxOffsetX, false)
	return verticalTrack, verticalThumb, horizontalTrack, horizontalThumb
}

func scrollbarThumbRect(track Rect, minThumb, contentSize, viewportSize, offset, maxOffset int32, vertical bool) Rect {
	if track.Empty() || contentSize <= 0 || viewportSize <= 0 {
		return Rect{}
	}
	trackSize := track.W
	if vertical {
		trackSize = track.H
	}
	if trackSize <= 0 {
		return Rect{}
	}
	thumbSize := trackSize
	if contentSize > viewportSize {
		thumbSize = trackSize * viewportSize / contentSize
	}
	if thumbSize < minThumb {
		thumbSize = minThumb
	}
	if thumbSize > trackSize {
		thumbSize = trackSize
	}
	travel := max32(0, trackSize-thumbSize)
	pos := int32(0)
	if maxOffset > 0 && travel > 0 {
		pos = travel * clampValue(offset, 0, maxOffset) / maxOffset
	}
	if vertical {
		return Rect{X: track.X, Y: track.Y + pos, W: track.W, H: thumbSize}
	}
	return Rect{X: track.X + pos, Y: track.Y, W: thumbSize, H: track.H}
}

func (s *ScrollView) scrollbarColors() (core.Color, core.Color) {
	background := s.Style.Background
	if background == 0 {
		background = core.RGB(255, 255, 255)
	}
	border := s.Style.BorderColor
	if border == 0 {
		border = core.RGB(203, 213, 225)
	}
	thumbAlpha := byte(112)
	if s.hoverVerticalThumb || s.hoverHorizontalThumb {
		thumbAlpha = 144
	}
	if s.dragAxis != scrollDragNone {
		thumbAlpha = 176
	}
	return blendScrollColor(background, border, 28), blendScrollColor(background, border, thumbAlpha)
}

func blendScrollColor(background, foreground core.Color, alpha byte) core.Color {
	bgR, bgG, bgB := scrollColorChannels(background)
	fgR, fgG, fgB := scrollColorChannels(foreground)
	return core.RGB(
		blendScrollChannel(bgR, fgR, alpha),
		blendScrollChannel(bgG, fgG, alpha),
		blendScrollChannel(bgB, fgB, alpha),
	)
}

func blendScrollChannel(background, foreground, alpha byte) byte {
	const scale = 255
	value := int(background)*(scale-int(alpha)) + int(foreground)*int(alpha)
	return byte((value + scale/2) / scale)
}

func scrollColorChannels(color core.Color) (byte, byte, byte) {
	return byte(color), byte(color >> 8), byte(color >> 16)
}

//go:build windows

package widgets

import (
	"fmt"
	"github.com/AzureIvory/winui/core"
	"sync"
	"time"
)

// PaintCtx 是带场景信息的绘制上下文包装器。
type PaintCtx struct {
	// canvas 保存底层 core 绘制上下文。
	canvas *core.PaintCtx
	// scene 指向当前正在绘制的场景。
	scene *Scene
}

// newPaintCtx 将 core 绘制上下文封装为场景感知的绘制上下文。
func newPaintCtx(scene *Scene, canvas *core.PaintCtx) *PaintCtx {
	return &PaintCtx{
		canvas: canvas,
		scene:  scene,
	}
}

// Canvas 返回底层的 core 绘制上下文。
func (p *PaintCtx) Canvas() *core.PaintCtx {
	if p == nil {
		return nil
	}
	return p.canvas
}

// Bounds 返回绘制上下文的边界。
func (p *PaintCtx) Bounds() Rect {
	if p == nil || p.canvas == nil {
		return Rect{}
	}
	return p.canvas.Bounds()
}

// DP 按应用当前 DPI 缩放设备无关值。
func (p *PaintCtx) DP(value int32) int32 {
	if p == nil || p.scene == nil || p.scene.app == nil {
		return value
	}
	return p.scene.app.DP(value)
}

// FillRoundRect 在当前画布上填充指定圆角矩形。
func (p *PaintCtx) FillRoundRect(rect Rect, radius int32, color core.Color) error {
	if p == nil || p.canvas == nil {
		return nil
	}
	return p.canvas.FillRoundRect(rect, radius, color)
}

// FillRect 在当前画布上填充指定矩形。
func (p *PaintCtx) FillRect(rect Rect, color core.Color) error {
	if p == nil || p.canvas == nil {
		return nil
	}
	return p.canvas.FillRect(rect, color)
}

// StrokeRoundRect 在当前画布上描边指定圆角矩形。
func (p *PaintCtx) StrokeRoundRect(rect Rect, radius int32, color core.Color, width int32) error {
	if p == nil || p.canvas == nil {
		return nil
	}
	return p.canvas.StrokeRoundRect(rect, radius, color, width)
}

// DrawIcon 在当前画布上绘制图标。
func (p *PaintCtx) DrawIcon(icon *core.Icon, rect Rect) error {
	if p == nil || p.canvas == nil {
		return nil
	}
	return p.canvas.DrawIcon(icon, rect)
}

// DrawText 在当前画布上绘制文本。
func (p *PaintCtx) DrawText(text string, rect Rect, style TextStyle) error {
	if p == nil || p.canvas == nil {
		return nil
	}
	font, err := p.scene.font(style.Font)
	if err != nil {
		return err
	}
	return p.canvas.DrawText(text, rect, font, style.Color, style.Format)
}

// MeasureText 测量渲染给定文本所需的尺寸。
func (p *PaintCtx) MeasureText(text string, spec FontSpec) (core.Size, error) {
	if p == nil || p.canvas == nil || p.scene == nil {
		return core.Size{}, nil
	}
	font, err := p.scene.font(spec)
	if err != nil {
		return core.Size{}, err
	}
	return p.canvas.MeasureText(text, font)
}

// DrawProgress 在当前画布上绘制进度条。
func (p *PaintCtx) DrawProgress(rect Rect, value int32, style ProgressStyle) error {
	if p == nil || p.canvas == nil || rect.Empty() {
		return nil
	}
	value = clampValue(value, 0, 100)
	trackRadius := p.DP(style.CornerRadius)
	if trackRadius <= 0 {
		trackRadius = max32(1, rect.H/2)
	}

	if err := p.canvas.FillRoundRect(rect, trackRadius, style.TrackColor); err != nil {
		return err
	}

	fillW := rect.W * value / 100
	if fillW > 0 {
		fillW = clampValue(fillW, 1, rect.W)
		if err := p.canvas.FillRoundRect(
			Rect{X: rect.X, Y: rect.Y, W: fillW, H: rect.H},
			min32(trackRadius, max32(1, fillW)),
			style.FillColor,
		); err != nil {
			return err
		}
	}

	if !style.ShowPercent {
		return nil
	}

	bubbleColor := style.BubbleColor
	if bubbleColor == 0 {
		bubbleColor = style.FillColor
	}
	textColor := style.TextColor
	if textColor == 0 {
		textColor = core.RGB(255, 255, 255)
	}

	bubbleW := p.DP(52)
	bubbleH := p.DP(28)
	bubbleGap := p.DP(6)
	pointerH := p.DP(6)
	pointerHalfW := p.DP(6)
	pointerInset := p.DP(10)
	bubbleRadius := min32(max32(1, trackRadius), max32(1, bubbleH/2))

	anchorX := rect.X + fillW
	if anchorX < rect.X {
		anchorX = rect.X
	}
	if anchorX > rect.X+rect.W {
		anchorX = rect.X + rect.W
	}

	bubbleX := anchorX - bubbleW/2
	if rect.W > bubbleW {
		bubbleX = clampValue(bubbleX, rect.X, rect.X+rect.W-bubbleW)
	} else {
		bubbleX = rect.X + (rect.W-bubbleW)/2
	}

	tipY := rect.Y - bubbleGap
	baseY := tipY - pointerH
	bubbleRect := Rect{
		X: bubbleX,
		Y: baseY - bubbleH,
		W: bubbleW,
		H: bubbleH,
	}

	pointerX := anchorX
	minPointerX := bubbleRect.X + pointerInset
	maxPointerX := bubbleRect.X + bubbleRect.W - pointerInset
	if minPointerX > maxPointerX {
		pointerX = bubbleRect.X + bubbleRect.W/2
	} else {
		pointerX = clampValue(pointerX, minPointerX, maxPointerX)
	}

	if err := p.canvas.FillPolygon([]core.Point{
		{X: pointerX, Y: tipY},
		{X: pointerX - pointerHalfW, Y: baseY},
		{X: pointerX + pointerHalfW, Y: baseY},
	}, bubbleColor); err != nil {
		return err
	}
	if err := p.canvas.FillRoundRect(bubbleRect, bubbleRadius, bubbleColor); err != nil {
		return err
	}

	return p.DrawText(
		fmt.Sprintf("%d%%", value),
		bubbleRect,
		TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
		},
	)
}

// Scene 管理控件树、主题、焦点和定时器等运行时状态。
type Scene struct {
	// app 指向场景绑定的底层应用实例。
	app *core.App
	// root 保存场景根面板。
	root *Panel
	// theme 保存当前主题。
	theme *Theme
	// hover 记录当前鼠标悬停的控件。
	hover Widget
	// capture 记录当前捕获鼠标的控件。
	capture Widget
	// focus 记录当前拥有键盘焦点的控件。
	focus Widget

	// fontMu 保护字体缓存。
	fontMu sync.Mutex
	// fonts 缓存已创建的字体资源。
	fonts map[FontSpec]*core.Font

	// timerMu 保护定时器映射。
	timerMu sync.Mutex
	// timers 保存定时器与控件的映射关系。
	timers map[uintptr]Widget
	// nextTimerID 保存下一个自动分配的定时器标识。
	nextTimerID uintptr
}

// NewScene 创建一个新的场景。
func NewScene(coreApp *core.App) *Scene {
	root := NewPanel("scene-root")
	scene := &Scene{
		app:    coreApp,
		root:   root,
		theme:  DefaultTheme(),
		fonts:  make(map[FontSpec]*core.Font),
		timers: make(map[uintptr]Widget),
	}
	root.setScene(scene)
	root.SetBounds(Rect{W: coreApp.ClientSize().Width, H: coreApp.ClientSize().Height})
	return scene
}

// Root 返回场景的根面板。
func (s *Scene) Root() *Panel {
	return s.root
}

// Theme 返回场景当前使用的主题。
func (s *Scene) Theme() *Theme {
	return s.theme
}

// SetTheme 更新场景当前主题。
func (s *Scene) SetTheme(theme *Theme) {
	if theme == nil {
		return
	}
	s.runOnUI(func() {
		s.theme = theme
		s.resetFonts()
		s.Invalidate(nil)
	})
}

// ReloadResources 重新加载场景资源。
func (s *Scene) ReloadResources() {
	s.runOnUI(func() {
		s.resetFonts()
		s.Invalidate(nil)
	})
}

// Resize 按给定边界调整场景尺寸。
func (s *Scene) Resize(bounds Rect) {
	s.runOnUI(func() {
		s.root.SetBounds(bounds)
		s.Dispatch(Event{Type: EventResize, Bounds: bounds, Source: s.root})
		s.Invalidate(nil)
	})
}

// Paint 使用给定的绘制上下文完成绘制。
func (s *Scene) Paint(ctx *PaintCtx) {
	if s == nil || s.root == nil || ctx == nil {
		return
	}
	s.root.Paint(ctx)
	s.paintOverlays(s.root, ctx)
}

// PaintCore 使用 core 画布绘制场景。
func (s *Scene) PaintCore(canvas *core.PaintCtx) {
	s.Paint(newPaintCtx(s, canvas))
}

// Dispatch 将事件分发到场景中。
func (s *Scene) Dispatch(evt Event) bool {
	if s == nil || s.root == nil {
		return false
	}

	switch evt.Type {
	case EventMouseMove:
		return s.dispatchMouseMove(evt)
	case EventMouseLeave:
		return s.dispatchMouseLeave(evt)
	case EventMouseDown:
		return s.dispatchMouseDown(evt)
	case EventMouseUp:
		return s.dispatchMouseUp(evt)
	case EventMouseWheel:
		return s.dispatchMouseWheel(evt)
	case EventResize:
		return s.routeEvent(s.root, evt)
	case EventKeyDown, EventChar:
		return s.dispatchKeyboard(evt)
	case EventTimer, EventPaint:
		return s.routeEvent(s.root, evt)
	default:
		if evt.Source != nil {
			return s.routeEvent(evt.Source, evt)
		}
	}
	return false
}

// DispatchMouseMove 在场景中分发鼠标移动事件。
func (s *Scene) DispatchMouseMove(ev core.MouseEvent) bool {
	return s.Dispatch(eventFromMouse(EventMouseMove, ev))
}

// DispatchMouseLeave 在场景中分发鼠标离开事件。
func (s *Scene) DispatchMouseLeave() bool {
	return s.Dispatch(Event{Type: EventMouseLeave})
}

// DispatchMouseDown 在场景中分发鼠标按下事件。
func (s *Scene) DispatchMouseDown(ev core.MouseEvent) bool {
	return s.Dispatch(eventFromMouse(EventMouseDown, ev))
}

// DispatchMouseUp 在场景中分发鼠标抬起事件。
func (s *Scene) DispatchMouseUp(ev core.MouseEvent) bool {
	return s.Dispatch(eventFromMouse(EventMouseUp, ev))
}

// DispatchMouseWheel 在场景中分发鼠标滚轮事件。
func (s *Scene) DispatchMouseWheel(ev core.MouseEvent) bool {
	return s.Dispatch(eventFromMouse(EventMouseWheel, ev))
}

// DispatchKeyDown 在场景中分发按键按下事件。
func (s *Scene) DispatchKeyDown(ev core.KeyEvent) bool {
	return s.Dispatch(Event{Type: EventKeyDown, Key: ev, Source: s.focus})
}

// DispatchChar 在场景中分发字符输入事件。
func (s *Scene) DispatchChar(ch rune) bool {
	return s.Dispatch(Event{Type: EventChar, Rune: ch, Source: s.focus})
}

// Focus 返回当前持有键盘焦点的控件。
func (s *Scene) Focus() Widget {
	return s.focus
}

// Blur 清除场景中的键盘焦点。
func (s *Scene) Blur() {
	if s == nil {
		return
	}
	s.runOnUI(func() {
		s.setFocus(nil)
	})
}

// Invalidate 标记区域或控件需要重绘。
func (s *Scene) Invalidate(widget Widget) {
	if s == nil || s.app == nil {
		return
	}
	if widget == nil {
		s.app.Invalidate(nil)
		return
	}
	s.invalidateRect(widgetDirtyRect(widget))
}

// invalidateRect 让指定矩形区域失效并等待重绘。
func (s *Scene) invalidateRect(rect Rect) {
	if s == nil || s.app == nil || rect.Empty() {
		return
	}
	s.app.Invalidate(&rect)
}

// HandleTimer 处理场景的定时器事件。
func (s *Scene) HandleTimer(timerID uintptr) bool {
	if s == nil || timerID == 0 {
		return false
	}

	s.timerMu.Lock()
	target := s.timers[timerID]
	s.timerMu.Unlock()
	if target == nil {
		return false
	}

	return s.routeEvent(target, Event{
		Type:    EventTimer,
		TimerID: timerID,
		Source:  target,
	})
}

// Close 释放场景持有的资源。
func (s *Scene) Close() error {
	s.timerMu.Lock()
	timerIDs := make([]uintptr, 0, len(s.timers))
	for id := range s.timers {
		timerIDs = append(timerIDs, id)
	}
	s.timerMu.Unlock()

	for _, id := range timerIDs {
		_ = s.app.KillTimer(id)
	}

	s.disposeTree(s.root)
	s.focus = nil
	s.hover = nil
	s.capture = nil

	s.fontMu.Lock()
	defer s.fontMu.Unlock()
	for key, font := range s.fonts {
		_ = font.Close()
		delete(s.fonts, key)
	}
	return nil
}

// runOnUI 在场景的 UI 线程执行回调。
func (s *Scene) runOnUI(fn func()) {
	if fn == nil {
		return
	}
	if s == nil || s.app == nil || s.app.IsUIThread() {
		fn()
		return
	}
	_ = s.app.Post(fn)
}

// font 返回场景中对应规格的字体资源。
func (s *Scene) font(spec FontSpec) (*core.Font, error) {
	s.fontMu.Lock()
	defer s.fontMu.Unlock()

	if font := s.fonts[spec]; font != nil {
		return font, nil
	}
	font, err := s.app.NewFont(spec.Face, spec.SizeDP, spec.Weight)
	if err != nil {
		return nil, err
	}
	s.fonts[spec] = font
	return font, nil
}

// resetFonts 重置场景缓存的字体资源。
func (s *Scene) resetFonts() {
	s.fontMu.Lock()
	defer s.fontMu.Unlock()
	for key, font := range s.fonts {
		_ = font.Close()
		delete(s.fonts, key)
	}
}

// startTimer 启动场景定时器。
func (s *Scene) startTimer(owner Widget, timerID uintptr, interval time.Duration) (uintptr, error) {
	if s == nil || s.app == nil || owner == nil {
		return 0, nil
	}

	s.timerMu.Lock()
	if timerID == 0 {
		s.nextTimerID++
		timerID = s.nextTimerID
	}
	s.timers[timerID] = owner
	s.timerMu.Unlock()

	if err := s.app.SetTimer(timerID, interval); err != nil {
		s.timerMu.Lock()
		delete(s.timers, timerID)
		s.timerMu.Unlock()
		return 0, err
	}
	return timerID, nil
}

// updateTimer 更新场景定时器。
func (s *Scene) updateTimer(timerID uintptr, interval time.Duration) error {
	if s == nil || s.app == nil || timerID == 0 {
		return nil
	}
	return s.app.SetTimer(timerID, interval)
}

// stopTimer 停止场景定时器。
func (s *Scene) stopTimer(timerID uintptr) error {
	if s == nil || s.app == nil || timerID == 0 {
		return nil
	}
	s.timerMu.Lock()
	delete(s.timers, timerID)
	s.timerMu.Unlock()
	return s.app.KillTimer(timerID)
}

// disposeTree 释放场景控件树。
func (s *Scene) disposeTree(widget Widget) {
	if widget == nil {
		return
	}
	if container, ok := widget.(Container); ok {
		for _, child := range container.Children() {
			s.disposeTree(child)
		}
	}
	if disposable, ok := widget.(interface{ Close() error }); ok {
		_ = disposable.Close()
	}
}

// routeEvent 沿目标控件的祖先链路由事件。
func (s *Scene) routeEvent(target Widget, evt Event) bool {
	if target == nil {
		return false
	}
	path := s.pathTo(target)
	if len(path) == 0 {
		return false
	}

	for _, widget := range path {
		if widget.OnEvent(evt) {
			return true
		}
	}
	for i := len(path) - 2; i >= 0; i-- {
		if path[i].OnEvent(evt) {
			return true
		}
	}
	return false
}

// dispatchKeyboard 在场景中分发键盘事件。
func (s *Scene) dispatchKeyboard(evt Event) bool {
	if s.focus == nil {
		return false
	}
	evt.Source = s.focus
	return s.routeEvent(s.focus, evt)
}

// pathTo 构建从根节点到目标控件的祖先路径。
func (s *Scene) pathTo(target Widget) []Widget {
	node := asWidgetNode(target)
	if node == nil {
		return nil
	}

	var reverse []Widget
	for current := node; current != nil; {
		reverse = append(reverse, current)
		parent := current.parent()
		if parent == nil {
			break
		}
		current = asWidgetNode(parent)
	}

	path := make([]Widget, 0, len(reverse))
	for i := len(reverse) - 1; i >= 0; i-- {
		path = append(path, reverse[i])
	}
	return path
}

// dispatchMouseMove 在场景中分发鼠标移动事件。
func (s *Scene) dispatchMouseMove(evt Event) bool {
	hit := s.hitTest(s.root, evt.Point.X, evt.Point.Y)
	eventTarget := hit
	if s.capture != nil {
		eventTarget = s.capture
	}

	if s.hover != hit {
		if s.hover != nil {
			s.routeEvent(s.hover, Event{Type: EventMouseLeave, Point: evt.Point, Source: s.hover})
		}
		s.hover = hit
		if s.hover != nil {
			s.routeEvent(s.hover, Event{Type: EventMouseEnter, Point: evt.Point, Source: s.hover})
		}
	}

	if hit != nil {
		s.app.SetCursor(cursorFor(hit))
	} else {
		s.app.SetCursor(core.CursorArrow)
	}

	if eventTarget != nil {
		return s.routeEvent(eventTarget, evt)
	}
	return false
}

// dispatchMouseLeave 在场景中分发鼠标离开事件。
func (s *Scene) dispatchMouseLeave(evt Event) bool {
	if s.hover == nil {
		s.app.SetCursor(core.CursorArrow)
		return false
	}
	target := s.hover
	s.hover = nil
	s.app.SetCursor(core.CursorArrow)
	return s.routeEvent(target, Event{Type: EventMouseLeave, Point: evt.Point, Source: target})
}

// dispatchMouseDown 在场景中分发鼠标按下事件。
func (s *Scene) dispatchMouseDown(evt Event) bool {
	target := s.hitTest(s.root, evt.Point.X, evt.Point.Y)
	s.setFocus(target)
	if target == nil {
		return false
	}
	s.capture = target
	s.app.CaptureMouse()
	return s.routeEvent(target, evt)
}

// dispatchMouseUp 在场景中分发鼠标抬起事件。
func (s *Scene) dispatchMouseUp(evt Event) bool {
	target := s.capture
	s.capture = nil
	s.app.ReleaseMouse()

	hit := s.hitTest(s.root, evt.Point.X, evt.Point.Y)
	if target == nil {
		target = hit
	}
	if target == nil {
		return false
	}

	handled := s.routeEvent(target, evt)
	if evt.Button == core.MouseButtonLeft && hit != nil && target == hit {
		click := evt
		click.Type = EventClick
		click.Source = hit
		if s.routeEvent(hit, click) {
			handled = true
		}
	}
	return handled
}

// dispatchMouseWheel 在场景中分发滚轮事件。
func (s *Scene) dispatchMouseWheel(evt Event) bool {
	target := s.capture
	if target == nil {
		target = s.hitTest(s.root, evt.Point.X, evt.Point.Y)
	}
	if target == nil {
		return false
	}
	return s.routeEvent(target, evt)
}

// hitTest 返回给定点命中的最上层可见控件。
func (s *Scene) hitTest(widget Widget, x, y int32) Widget {
	if widget == nil || !widget.Visible() || !widget.Enabled() {
		return nil
	}
	if container, ok := widget.(Container); ok {
		children := container.Children()
		for i := len(children) - 1; i >= 0; i-- {
			if hit := s.hitTest(children[i], x, y); hit != nil {
				return hit
			}
		}
	}
	if widget.HitTest(x, y) {
		return widget
	}
	return nil
}

// setFocus 更新场景焦点。
func (s *Scene) setFocus(target Widget) {
	if target != nil {
		focusable, ok := target.(focusableWidget)
		if !ok || !focusable.acceptsFocus() || !target.Visible() || !target.Enabled() {
			target = nil
		}
	}
	if s.focus == target {
		return
	}

	old := s.focus
	s.focus = target
	if old != nil {
		s.routeEvent(old, Event{Type: EventBlur, Source: old})
	}
	if s.focus != nil {
		s.routeEvent(s.focus, Event{Type: EventFocus, Source: s.focus})
	}
}

// paintOverlays 在常规内容之后绘制覆盖层控件。
func (s *Scene) paintOverlays(widget Widget, ctx *PaintCtx) {
	if widget == nil || !widget.Visible() {
		return
	}
	if container, ok := widget.(Container); ok {
		for _, child := range container.Children() {
			s.paintOverlays(child, ctx)
		}
	}
	if overlay, ok := widget.(overlayWidget); ok {
		overlay.PaintOverlay(ctx)
	}
}

// cursorFor 返回悬停控件时应显示的光标。
func cursorFor(widget Widget) core.CursorID {
	if node := asWidgetNode(widget); node != nil {
		return node.cursor()
	}
	return core.CursorArrow
}

// clampValue 将值限制在给定的闭区间内。
func clampValue(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// min32 返回两个 32 位整数中较小的值。
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// max32 返回两个 32 位整数中较大的值。
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// widgetDirtyRect 返回控件声明的脏区或其边界。
func widgetDirtyRect(widget Widget) Rect {
	if widget == nil {
		return Rect{}
	}
	if dirty, ok := widget.(dirtyWidget); ok {
		return dirty.dirtyRect()
	}
	return widget.Bounds()
}

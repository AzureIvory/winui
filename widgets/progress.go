//go:build windows

package widgets

type ProgressBar struct {
	widgetBase
	value int32
	Style ProgressStyle
}

// NewProgressBar 创建一个新的进度条。
func NewProgressBar(id string) *ProgressBar {
	return &ProgressBar{
		widgetBase: newWidgetBase(id, "progress"),
	}
}

// SetBounds 更新进度条边界。
func (p *ProgressBar) SetBounds(rect Rect) {
	p.runOnUI(func() {
		p.widgetBase.setBounds(p, rect)
	})
}

// SetVisible 更新进度条可见状态。
func (p *ProgressBar) SetVisible(visible bool) {
	p.runOnUI(func() {
		p.widgetBase.setVisible(p, visible)
	})
}

// SetEnabled 更新进度条可用状态。
func (p *ProgressBar) SetEnabled(enabled bool) {
	p.runOnUI(func() {
		p.widgetBase.setEnabled(p, enabled)
	})
}

// SetValue 更新进度条当前值。
func (p *ProgressBar) SetValue(value int32) {
	p.runOnUI(func() {
		value = clampValue(value, 0, 100)
		if p.value == value {
			return
		}
		p.value = value
		p.invalidate(p)
	})
}

// SetStyle 更新进度条样式覆盖。
func (p *ProgressBar) SetStyle(style ProgressStyle) {
	p.runOnUI(func() {
		oldRect := widgetDirtyRect(p)
		p.Style = style
		if scene := p.scene(); scene != nil {
			scene.invalidateRect(oldRect)
		}
		p.invalidate(p)
	})
}

// Value 返回进度条当前值。
func (p *ProgressBar) Value() int32 {
	return p.value
}

// OnEvent 处理输入事件或生命周期事件。
func (p *ProgressBar) OnEvent(Event) bool {
	return false
}

// Paint 使用给定绘制上下文完成绘制。
func (p *ProgressBar) Paint(ctx *PaintCtx) {
	if !p.Visible() || ctx == nil {
		return
	}
	style := p.resolveStyle(ctx)
	_ = ctx.DrawProgress(p.Bounds(), p.value, style)
}

// resolveStyle 解析进度条最终样式。
func (p *ProgressBar) resolveStyle(ctx *PaintCtx) ProgressStyle {
	style := DefaultTheme().Progress
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.Progress
	}
	if p.Style.Font.Face != "" {
		style.Font = p.Style.Font
	}
	if p.Style.TextColor != 0 {
		style.TextColor = p.Style.TextColor
	}
	if p.Style.TrackColor != 0 {
		style.TrackColor = p.Style.TrackColor
	}
	if p.Style.FillColor != 0 {
		style.FillColor = p.Style.FillColor
	}
	if p.Style.BubbleColor != 0 {
		style.BubbleColor = p.Style.BubbleColor
	}
	if p.Style.CornerRadius != 0 {
		style.CornerRadius = p.Style.CornerRadius
	}
	if p.Style.ShowPercent != style.ShowPercent {
		style.ShowPercent = p.Style.ShowPercent
	}
	return style
}

// dirtyRect 返回进度条及其百分比气泡可能覆盖到的脏区。
func (p *ProgressBar) dirtyRect() Rect {
	rect := p.Bounds()
	if rect.Empty() {
		return rect
	}

	style := p.resolveStyle(&PaintCtx{scene: p.scene()})
	if !style.ShowPercent {
		return rect
	}

	bubbleW := p.dp(52)
	bubbleH := p.dp(28)
	bubbleGap := p.dp(6)
	pointerH := p.dp(6)

	padX := bubbleW / 2
	if bubbleW > rect.W {
		padX = bubbleW
	}

	return Rect{
		X: rect.X - padX,
		Y: rect.Y - bubbleGap - pointerH - bubbleH,
		W: rect.W + padX*2,
		H: rect.H + bubbleGap + pointerH + bubbleH,
	}
}

// dp 按控件所在场景的 DPI 缩放尺寸。
func (p *ProgressBar) dp(value int32) int32 {
	if scene := p.scene(); scene != nil && scene.app != nil {
		return scene.app.DP(value)
	}
	return value
}

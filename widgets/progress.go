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

// SetBounds 更新进度条的边界。
func (p *ProgressBar) SetBounds(rect Rect) {
	p.widgetBase.setBounds(p, rect)
}

// SetVisible 更新进度条的可见状态。
func (p *ProgressBar) SetVisible(visible bool) {
	p.widgetBase.setVisible(p, visible)
}

// SetEnabled 更新进度条的可用状态。
func (p *ProgressBar) SetEnabled(enabled bool) {
	p.widgetBase.setEnabled(p, enabled)
}

// SetValue 更新进度条的当前值。
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

// SetStyle 更新进度条的样式覆盖。
func (p *ProgressBar) SetStyle(style ProgressStyle) {
	p.runOnUI(func() {
		p.Style = style
		p.invalidate(p)
	})
}

// Value 返回进度条的当前值。
func (p *ProgressBar) Value() int32 {
	return p.value
}

// OnEvent 处理输入事件或生命周期事件。
func (p *ProgressBar) OnEvent(Event) bool {
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (p *ProgressBar) Paint(ctx *PaintCtx) {
	if !p.Visible() {
		return
	}
	style := p.resolveStyle(ctx)
	_ = ctx.DrawProgress(p.Bounds(), p.value, style)
}

// resolveStyle 解析进度条的最终样式。
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

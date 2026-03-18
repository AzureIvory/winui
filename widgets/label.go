//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

type Label struct {
	widgetBase
	Text  string
	Style TextStyle
}

// NewLabel 创建一个新的标签控件�?func NewLabel(id, text string) *Label {
	return &Label{
		widgetBase: newWidgetBase(id, "label"),
		Text:       text,
	}
}

// SetBounds 更新标签的边界�?func (l *Label) SetBounds(rect Rect) {
	l.widgetBase.setBounds(l, rect)
}

// SetVisible 更新标签的可见状态�?func (l *Label) SetVisible(visible bool) {
	l.widgetBase.setVisible(l, visible)
}

// SetEnabled 更新标签的可用状态�?func (l *Label) SetEnabled(enabled bool) {
	l.widgetBase.setEnabled(l, enabled)
}

// SetText 更新标签的显示文本�?func (l *Label) SetText(text string) {
	l.runOnUI(func() {
		if l.Text == text {
			return
		}
		l.Text = text
		l.invalidate(l)
	})
}

// SetStyle 更新标签的样式覆盖�?func (l *Label) SetStyle(style TextStyle) {
	l.runOnUI(func() {
		l.Style = style
		l.invalidate(l)
	})
}

// OnEvent 处理输入事件或生命周期事件�?func (l *Label) OnEvent(Event) bool {
	return false
}

// Paint 使用给定的绘制上下文完成绘制�?func (l *Label) Paint(ctx *PaintCtx) {
	if !l.Visible() || l.Text == "" {
		return
	}
	style := l.resolveStyle(ctx)
	_ = ctx.DrawText(l.Text, l.Bounds(), style)
}

// resolveStyle 解析标签的最终样式�?func (l *Label) resolveStyle(ctx *PaintCtx) TextStyle {
	style := TextStyle{
		Font: FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 16,
		},
		Color:  core.RGB(16, 16, 16),
		Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
	}
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.Text
	}
	if l.Style.Font.Face != "" {
		style.Font = l.Style.Font
	}
	if l.Style.Color != 0 {
		style.Color = l.Style.Color
	}
	if l.Style.Format != 0 {
		style.Format = l.Style.Format
	}
	return style
}

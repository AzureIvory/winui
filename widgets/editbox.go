//go:build windows

package widgets

import "github.com/yourname/winui/core"

// EditBox 表示单行可编辑文本控件。
type EditBox struct {
	widgetBase
	Text        string
	Placeholder string
	ReadOnly    bool
	Hover       bool
	Focused     bool
	caret       int
	Style       EditStyle
	OnChange    func(string)
	OnSubmit    func(string)
}

// NewEditBox 创建一个新的编辑框。
func NewEditBox(id string) *EditBox {
	return &EditBox{
		widgetBase: newWidgetBase(id, "edit"),
	}
}

// SetBounds 更新编辑框的边界。
func (e *EditBox) SetBounds(rect Rect) {
	e.runOnUI(func() {
		e.widgetBase.setBounds(e, rect)
	})
}

// SetVisible 更新编辑框的可见状态。
func (e *EditBox) SetVisible(visible bool) {
	e.runOnUI(func() {
		e.widgetBase.setVisible(e, visible)
	})
}

// SetEnabled 更新编辑框的可用状态。
func (e *EditBox) SetEnabled(enabled bool) {
	e.runOnUI(func() {
		e.widgetBase.setEnabled(e, enabled)
	})
}

// SetText 更新编辑框的显示文本。
func (e *EditBox) SetText(text string) {
	e.runOnUI(func() {
		if e.Text == text {
			return
		}
		e.Text = text
		e.caret = len([]rune(text))
		e.invalidate(e)
	})
}

// TextValue 返回编辑框当前保存的文本。
func (e *EditBox) TextValue() string {
	return e.Text
}

// SetPlaceholder 更新编辑框的占位文本。
func (e *EditBox) SetPlaceholder(text string) {
	e.runOnUI(func() {
		if e.Placeholder == text {
			return
		}
		e.Placeholder = text
		e.invalidate(e)
	})
}

// SetReadOnly 更新编辑框的只读状态。
func (e *EditBox) SetReadOnly(readOnly bool) {
	e.runOnUI(func() {
		if e.ReadOnly == readOnly {
			return
		}
		e.ReadOnly = readOnly
		e.invalidate(e)
	})
}

// SetStyle 更新编辑框的样式覆盖。
func (e *EditBox) SetStyle(style EditStyle) {
	e.runOnUI(func() {
		e.Style = style
		e.invalidate(e)
	})
}

// SetOnChange 注册编辑框的变更回调。
func (e *EditBox) SetOnChange(fn func(string)) {
	e.runOnUI(func() {
		e.OnChange = fn
	})
}

// SetOnSubmit 注册编辑框的提交回调。
func (e *EditBox) SetOnSubmit(fn func(string)) {
	e.runOnUI(func() {
		e.OnSubmit = fn
	})
}

// OnEvent 处理输入事件或生命周期事件。
func (e *EditBox) OnEvent(evt Event) bool {
	switch evt.Type {
	case EventMouseEnter:
		if !e.Hover {
			e.Hover = true
			e.invalidate(e)
		}
	case EventMouseLeave:
		if e.Hover {
			e.Hover = false
			e.invalidate(e)
		}
	case EventClick:
		if e.Enabled() {
			e.caret = len([]rune(e.Text))
			e.invalidate(e)
			return true
		}
	case EventFocus:
		if !e.Focused {
			e.Focused = true
			e.caret = len([]rune(e.Text))
			e.invalidate(e)
		}
	case EventBlur:
		if e.Focused {
			e.Focused = false
			e.invalidate(e)
		}
	case EventKeyDown:
		return e.handleKey(evt.Key)
	case EventChar:
		return e.handleChar(evt.Rune)
	}
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (e *EditBox) Paint(ctx *PaintCtx) {
	if !e.Visible() || ctx == nil {
		return
	}

	style := e.resolveStyle(ctx)
	bounds := e.Bounds()
	if bounds.Empty() {
		return
	}

	background := style.Background
	borderColor := style.BorderColor
	textColor := style.TextColor
	if !e.Enabled() {
		background = style.DisabledBg
		textColor = style.DisabledText
	} else if e.Focused {
		borderColor = style.FocusBorder
	} else if e.Hover {
		borderColor = style.HoverBorder
	}

	padding := ctx.DP(style.PaddingDP)
	radius := ctx.DP(style.CornerRadius)
	_ = ctx.FillRoundRect(bounds, radius, background)
	_ = ctx.StrokeRoundRect(bounds, radius, borderColor, 1)

	textRect := Rect{
		X: bounds.X + padding,
		Y: bounds.Y,
		W: max32(0, bounds.W-padding*2),
		H: bounds.H,
	}

	displayText := e.Text
	if displayText == "" {
		displayText = e.Placeholder
		textColor = style.PlaceholderColor
	}
	_ = ctx.DrawText(displayText, textRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	if !e.Focused {
		return
	}

	prefix := []rune(e.Text)
	if e.caret < 0 {
		e.caret = 0
	}
	if e.caret > len(prefix) {
		e.caret = len(prefix)
	}

	prefixText := string(prefix[:e.caret])
	size, _ := ctx.MeasureText(prefixText, style.Font)
	caretX := textRect.X + size.Width
	maxX := bounds.X + bounds.W - padding
	if caretX > maxX {
		caretX = maxX
	}
	caretRect := Rect{
		X: caretX,
		Y: bounds.Y + ctx.DP(8),
		W: max32(1, ctx.DP(2)),
		H: max32(1, bounds.H-ctx.DP(16)),
	}
	_ = ctx.FillRect(caretRect, style.CaretColor)
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (e *EditBox) acceptsFocus() bool {
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (e *EditBox) cursor() CursorID {
	if !e.Enabled() {
		return core.CursorArrow
	}
	return core.CursorIBeam
}

// resolveStyle 解析编辑框的最终样式。
func (e *EditBox) resolveStyle(ctx *PaintCtx) EditStyle {
	style := DefaultTheme().Edit
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.Edit
	}
	return mergeEditStyle(style, e.Style)
}

// handleKey 处理编辑框的按键事件。
func (e *EditBox) handleKey(key core.KeyEvent) bool {
	if !e.Enabled() {
		return false
	}

	runes := []rune(e.Text)
	switch key.Key {
	case core.KeyLeft:
		if e.caret > 0 {
			e.caret--
			e.invalidate(e)
		}
		return true
	case core.KeyRight:
		if e.caret < len(runes) {
			e.caret++
			e.invalidate(e)
		}
		return true
	case core.KeyHome:
		if e.caret != 0 {
			e.caret = 0
			e.invalidate(e)
		}
		return true
	case core.KeyEnd:
		if e.caret != len(runes) {
			e.caret = len(runes)
			e.invalidate(e)
		}
		return true
	case core.KeyBack:
		if e.ReadOnly || e.caret == 0 || len(runes) == 0 {
			return true
		}
		e.Text = string(append(runes[:e.caret-1], runes[e.caret:]...))
		e.caret--
		e.notifyChanged()
		return true
	case core.KeyDelete:
		if e.ReadOnly || e.caret >= len(runes) || len(runes) == 0 {
			return true
		}
		e.Text = string(append(runes[:e.caret], runes[e.caret+1:]...))
		e.notifyChanged()
		return true
	case core.KeyReturn:
		if e.OnSubmit != nil {
			e.OnSubmit(e.Text)
		}
		return true
	}
	return false
}

// handleChar 处理编辑框的字符输入。
func (e *EditBox) handleChar(ch rune) bool {
	if !e.Enabled() || e.ReadOnly {
		return false
	}
	if ch < 32 || ch == '\r' || ch == '\n' || ch == '\t' {
		return false
	}

	runes := []rune(e.Text)
	if e.caret < 0 {
		e.caret = 0
	}
	if e.caret > len(runes) {
		e.caret = len(runes)
	}
	runes = append(runes[:e.caret], append([]rune{ch}, runes[e.caret:]...)...)
	e.Text = string(runes)
	e.caret++
	e.notifyChanged()
	return true
}

// notifyChanged 使控件失效并触发变更回调。
func (e *EditBox) notifyChanged() {
	e.invalidate(e)
	if e.OnChange != nil {
		e.OnChange(e.Text)
	}
}

// mergeEditStyle 将编辑框样式覆盖合并到基础样式上。
func mergeEditStyle(base, override EditStyle) EditStyle {
	if override.Font.Face != "" {
		base.Font = override.Font
	}
	if override.TextColor != 0 {
		base.TextColor = override.TextColor
	}
	if override.PlaceholderColor != 0 {
		base.PlaceholderColor = override.PlaceholderColor
	}
	if override.Background != 0 {
		base.Background = override.Background
	}
	if override.BorderColor != 0 {
		base.BorderColor = override.BorderColor
	}
	if override.HoverBorder != 0 {
		base.HoverBorder = override.HoverBorder
	}
	if override.FocusBorder != 0 {
		base.FocusBorder = override.FocusBorder
	}
	if override.DisabledText != 0 {
		base.DisabledText = override.DisabledText
	}
	if override.DisabledBg != 0 {
		base.DisabledBg = override.DisabledBg
	}
	if override.CaretColor != 0 {
		base.CaretColor = override.CaretColor
	}
	if override.PaddingDP != 0 {
		base.PaddingDP = override.PaddingDP
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	return base
}

//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// EditBox 表示单行可编辑文本控件。
type EditBox struct {
	// widgetBase 提供编辑框共享的基础控件能力。
	widgetBase
	// mode 表示编辑框当前使用的后端模式。
	mode ControlMode
	// native 保存编辑框在原生后端下的运行时状态。
	native nativeControlState
	// Text 保存当前文本内容。
	Text string
	// Placeholder 保存占位文本。
	Placeholder string
	// ReadOnly 记录是否为只读模式。
	ReadOnly bool
	// Hover 记录当前是否处于悬停状态。
	Hover bool
	// Focused 记录当前是否拥有焦点。
	Focused bool
	// caret 保存光标所在字符索引。
	caret int
	// Style 保存样式覆盖。
	Style EditStyle
	// OnChange 保存文本变更回调。
	OnChange func(string)
	// OnSubmit 保存提交回调。
	OnSubmit func(string)
}

// NewEditBox 创建一个新的编辑框。
func NewEditBox(id string, mode ControlMode) *EditBox {
	return &EditBox{
		widgetBase: newWidgetBase(id, "edit"),
		mode:       normalizeControlMode(mode),
	}
}

// SetBounds 更新编辑框的边界。
func (e *EditBox) SetBounds(rect Rect) {
	e.runOnUI(func() {
		e.widgetBase.setBounds(e, rect)
		e.syncNativeBounds()
	})
}

// SetVisible 更新编辑框的可见状态。
func (e *EditBox) SetVisible(visible bool) {
	e.runOnUI(func() {
		e.widgetBase.setVisible(e, visible)
		e.syncNativeVisible()
	})
}

// SetEnabled 更新编辑框的可用状态。
func (e *EditBox) SetEnabled(enabled bool) {
	e.runOnUI(func() {
		e.widgetBase.setEnabled(e, enabled)
		e.syncNativeEnabled()
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
		e.syncNativeText()
		e.invalidate(e)
	})
}

// TextValue 返回编辑框当前保存的文本。
func (e *EditBox) TextValue() string {
	if e.native.valid() {
		e.Text = getNativeText(e.native.handle)
		e.caret = len([]rune(e.Text))
	}
	return e.Text
}

// SetPlaceholder 更新编辑框的占位文本。
func (e *EditBox) SetPlaceholder(text string) {
	e.runOnUI(func() {
		if e.Placeholder == text {
			return
		}
		e.Placeholder = text
		e.syncNativePlaceholder()
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
		e.syncNativeReadOnly()
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

// HitTest 判断给定点是否命中当前编辑框。
func (e *EditBox) HitTest(x, y int32) bool {
	if isNativeMode(e.mode) {
		return false
	}
	return e.widgetBase.HitTest(x, y)
}

// OnEvent 处理输入事件或生命周期事件。
func (e *EditBox) OnEvent(evt Event) bool {
	if isNativeMode(e.mode) {
		return false
	}
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
	if isNativeMode(e.mode) || !e.Visible() || ctx == nil {
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

// setScene 更新编辑框关联的场景，并在原生模式下同步子控件生命周期。
func (e *EditBox) setScene(scene *Scene) {
	current := e.scene()
	if current != scene {
		e.destroyNativeControl(current)
	}
	e.widgetBase.setScene(scene)
	e.ensureNativeControl(scene)
}

// Close 释放编辑框持有的原生后端资源。
func (e *EditBox) Close() error {
	e.runOnUI(func() {
		e.destroyNativeControl(e.scene())
	})
	return nil
}

// handleNativeCommand 处理原生编辑框发送的命令通知。
func (e *EditBox) handleNativeCommand(code uint16) bool {
	if !isNativeMode(e.mode) {
		return false
	}
	switch code {
	case nativeEditSetFocus:
		if scene := e.scene(); scene != nil {
			scene.Blur()
		}
		return true
	case nativeEditChanged:
		text := e.Text
		if e.native.valid() {
			text = getNativeText(e.native.handle)
		}
		if e.Text == text {
			return true
		}
		e.Text = text
		e.caret = len([]rune(text))
		e.invalidate(e)
		if e.OnChange != nil {
			e.OnChange(text)
		}
		return true
	default:
		return false
	}
}

// ensureNativeControl 确保编辑框在原生模式下已创建系统子控件。
func (e *EditBox) ensureNativeControl(scene *Scene) {
	if !isNativeMode(e.mode) || scene == nil || scene.app == nil {
		return
	}
	if e.native.valid() {
		e.syncNativeBounds()
		e.syncNativeVisible()
		e.syncNativeEnabled()
		e.syncNativeText()
		e.syncNativePlaceholder()
		e.syncNativeReadOnly()
		return
	}
	commandID := scene.allocateNativeCommandID()
	handle, err := createNativeControl(
		scene,
		"EDIT",
		e.Text,
		nativeWindowChild|nativeWindowVisible|nativeWindowTabStop|nativeWindowBorder|nativeEditAutoHScroll,
		e.Bounds(),
		commandID,
	)
	if err != nil {
		return
	}
	e.native.handle = handle
	e.native.commandID = commandID
	scene.registerNativeControl(handle, e)
	subclassNativeEdit(e)
	e.syncNativeBounds()
	e.syncNativeVisible()
	e.syncNativeEnabled()
	e.syncNativeText()
	e.syncNativePlaceholder()
	e.syncNativeReadOnly()
}

// destroyNativeControl 销毁编辑框对应的原生系统子控件。
func (e *EditBox) destroyNativeControl(scene *Scene) {
	if !e.native.valid() {
		e.native.commandID = 0
		return
	}
	unsubclassNativeEdit(e)
	if scene != nil {
		scene.unregisterNativeControl(e.native.handle)
	}
	destroyNativeControl(e.native.handle)
	e.native.handle = 0
	e.native.commandID = 0
	e.native.oldWndProc = 0
}

// syncNativeBounds 同步编辑框原生控件边界。
func (e *EditBox) syncNativeBounds() {
	if e.native.valid() {
		setNativeBounds(e.native.handle, e.Bounds())
	}
}

// syncNativeVisible 同步编辑框原生控件可见性。
func (e *EditBox) syncNativeVisible() {
	if e.native.valid() {
		setNativeVisible(e.native.handle, e.Visible())
	}
}

// syncNativeEnabled 同步编辑框原生控件启用状态。
func (e *EditBox) syncNativeEnabled() {
	if e.native.valid() {
		setNativeEnabled(e.native.handle, e.Enabled())
	}
}

// syncNativeText 同步编辑框原生控件文本。
func (e *EditBox) syncNativeText() {
	if e.native.valid() {
		setNativeText(e.native.handle, e.Text)
	}
}

// syncNativePlaceholder 同步编辑框原生控件占位提示。
func (e *EditBox) syncNativePlaceholder() {
	if e.native.valid() {
		setNativeCueBanner(e.native.handle, e.Placeholder)
	}
}

// syncNativeReadOnly 同步编辑框原生控件只读状态。
func (e *EditBox) syncNativeReadOnly() {
	if e.native.valid() {
		setNativeReadOnly(e.native.handle, e.ReadOnly)
	}
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (e *EditBox) acceptsFocus() bool {
	if isNativeMode(e.mode) {
		return false
	}
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (e *EditBox) cursor() CursorID {
	if isNativeMode(e.mode) {
		return core.CursorArrow
	}
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
	base.Font = mergeFontSpec(base.Font, override.Font)
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

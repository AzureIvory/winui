//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// CheckBox 表示可切换状态的多选框控件。
// CheckBox 表示可切换状态的多选框控件。
type CheckBox struct {
	// widgetBase 提供选项控件共享的基础控件能力。
	widgetBase
	// mode 表示复选框当前使用的后端模式。
	mode ControlMode
	// native 保存复选框在原生后端下的运行时状态。
	native nativeControlState
	// Text 保存复选框文本。
	Text string
	// Checked 记录当前是否已选中。
	Checked bool
	// Hover 记录当前是否处于悬停状态。
	Hover bool
	// Down 记录当前是否处于按下状态。
	Down bool
	// Focused 记录当前是否拥有焦点。
	Focused bool
	// Style 保存样式覆盖。
	Style ChoiceStyle
	// OnChange 保存状态变更回调。
	OnChange func(bool)
}

// NewCheckBox 创建一个新的多选框。
func NewCheckBox(id, text string, mode ControlMode) *CheckBox {
	return &CheckBox{
		widgetBase: newWidgetBase(id, "checkbox"),
		mode:       normalizeControlMode(mode),
		Text:       text,
	}
}

// SetBounds 更新多选框边界。
func (c *CheckBox) SetBounds(rect Rect) {
	c.runOnUI(func() {
		c.widgetBase.setBounds(c, rect)
		c.syncNativeBounds()
	})
}

// SetVisible 更新多选框可见状态。
func (c *CheckBox) SetVisible(visible bool) {
	c.runOnUI(func() {
		c.widgetBase.setVisible(c, visible)
		c.syncNativeVisible()
	})
}

// SetEnabled 更新多选框可用状态。
func (c *CheckBox) SetEnabled(enabled bool) {
	c.runOnUI(func() {
		c.widgetBase.setEnabled(c, enabled)
		c.syncNativeEnabled()
	})
}

// SetText 更新多选框文本。
func (c *CheckBox) SetText(text string) {
	c.runOnUI(func() {
		if c.Text == text {
			return
		}
		c.Text = text
		c.syncNativeText()
		c.invalidate(c)
	})
}

// SetChecked 更新多选框选中状态。
func (c *CheckBox) SetChecked(checked bool) {
	c.runOnUI(func() {
		c.setChecked(checked, false)
	})
}

// IsChecked 返回多选框是否选中。
func (c *CheckBox) IsChecked() bool {
	return c.Checked
}

// SetStyle 更新多选框样式覆盖。
func (c *CheckBox) SetStyle(style ChoiceStyle) {
	c.runOnUI(func() {
		c.Style = style
		c.invalidate(c)
	})
}

// SetOnChange 注册多选框变更回调。
func (c *CheckBox) SetOnChange(fn func(bool)) {
	c.runOnUI(func() {
		c.OnChange = fn
	})
}

// HitTest 判断给定点是否命中当前复选框。
func (c *CheckBox) HitTest(x, y int32) bool {
	if isNativeMode(c.mode) {
		return false
	}
	return c.widgetBase.HitTest(x, y)
}

// OnEvent 处理输入事件或生命周期事件。
func (c *CheckBox) OnEvent(evt Event) bool {
	if isNativeMode(c.mode) {
		return false
	}
	switch evt.Type {
	case EventMouseEnter:
		if !c.Hover {
			c.Hover = true
			c.invalidate(c)
		}
	case EventMouseLeave:
		changed := c.Hover || c.Down
		c.Hover = false
		c.Down = false
		if changed {
			c.invalidate(c)
		}
	case EventMouseDown:
		if c.Enabled() {
			c.Down = true
			c.invalidate(c)
			return true
		}
	case EventMouseUp:
		if c.Down {
			c.Down = false
			c.invalidate(c)
			return true
		}
	case EventClick:
		if c.Enabled() {
			c.setChecked(!c.Checked, true)
			return true
		}
	case EventFocus:
		if !c.Focused {
			c.Focused = true
			c.invalidate(c)
		}
	case EventBlur:
		if c.Focused {
			c.Focused = false
			c.Down = false
			c.invalidate(c)
		}
	}
	return false
}

// Paint 使用给定绘制上下文完成绘制。
func (c *CheckBox) Paint(ctx *PaintCtx) {
	if isNativeMode(c.mode) || !c.Visible() || ctx == nil {
		return
	}

	style := c.resolveStyle(ctx)
	content := c.Bounds()
	if content.Empty() {
		return
	}

	boxSize := ctx.DP(style.IndicatorSizeDP)
	if boxSize <= 0 {
		boxSize = ctx.DP(18)
	}
	gap := ctx.DP(style.IndicatorGapDP)
	if gap <= 0 {
		gap = ctx.DP(10)
	}

	boxRect := Rect{
		X: content.X,
		Y: content.Y + (content.H-boxSize)/2,
		W: boxSize,
		H: boxSize,
	}
	wrapRect := Rect{X: content.X, Y: content.Y, W: content.W, H: content.H}

	if c.Hover || c.Focused {
		_ = ctx.FillRoundRect(wrapRect, choiceWrapRadius(ctx, style), style.HoverBackground)
	}

	background := style.Background
	borderColor := style.BorderColor
	textColor := style.TextColor
	if !c.Enabled() {
		background = style.DisabledBg
		borderColor = style.DisabledBorder
		textColor = style.DisabledText
	} else if c.Focused {
		borderColor = style.FocusBorder
	} else if c.Hover {
		borderColor = style.HoverBorder
	}

	_ = ctx.FillRoundRect(boxRect, ctx.DP(style.CornerRadius), background)
	if c.Checked {
		_ = ctx.FillRoundRect(boxRect, ctx.DP(style.CornerRadius), style.IndicatorColor)
		drawChoiceMark(ctx, boxRect, style, false)
	}
	_ = ctx.StrokeRoundRect(boxRect, ctx.DP(style.CornerRadius), borderColor, 1)

	textRect := Rect{
		X: boxRect.X + boxRect.W + gap,
		Y: content.Y,
		W: max32(0, content.W-boxRect.W-gap),
		H: content.H,
	}
	_ = ctx.DrawText(c.Text, textRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})
}

// setScene 更新复选框关联的场景，并在原生模式下同步子控件生命周期。
func (c *CheckBox) setScene(scene *Scene) {
	current := c.scene()
	if current != scene {
		c.destroyNativeControl(current)
	}
	c.widgetBase.setScene(scene)
	c.ensureNativeControl(scene)
}

// Close 释放复选框持有的原生后端资源。
func (c *CheckBox) Close() error {
	c.runOnUI(func() {
		c.destroyNativeControl(c.scene())
	})
	return nil
}

// handleNativeCommand 处理原生复选框发送的命令通知。
func (c *CheckBox) handleNativeCommand(code uint16) bool {
	if !isNativeMode(c.mode) {
		return false
	}
	switch code {
	case nativeButtonSetFocus:
		if scene := c.scene(); scene != nil {
			scene.Blur()
		}
		return true
	case nativeButtonClicked:
		if !c.Enabled() {
			return true
		}
		checked := sendNativeMessage(c.native.handle, nativeButtonGetCheck, 0, 0) == nativeButtonStateChecked
		c.setChecked(checked, true)
		return true
	default:
		return false
	}
}

// ensureNativeControl 确保复选框在原生模式下已创建系统子控件。
func (c *CheckBox) ensureNativeControl(scene *Scene) {
	if !isNativeMode(c.mode) || scene == nil || scene.app == nil {
		return
	}
	if c.native.valid() {
		c.syncNativeBounds()
		c.syncNativeVisible()
		c.syncNativeEnabled()
		c.syncNativeText()
		c.syncNativeChecked()
		return
	}
	commandID := scene.allocateNativeCommandID()
	handle, err := createNativeControl(
		scene,
		"BUTTON",
		c.Text,
		nativeWindowChild|nativeWindowVisible|nativeWindowTabStop|nativeButtonCheckBox,
		c.Bounds(),
		commandID,
	)
	if err != nil {
		return
	}
	c.native.handle = handle
	c.native.commandID = commandID
	scene.registerNativeControl(handle, c)
	c.syncNativeBounds()
	c.syncNativeVisible()
	c.syncNativeEnabled()
	c.syncNativeText()
	c.syncNativeChecked()
}

// destroyNativeControl 销毁复选框对应的原生系统子控件。
func (c *CheckBox) destroyNativeControl(scene *Scene) {
	if !c.native.valid() {
		c.native.commandID = 0
		return
	}
	if scene != nil {
		scene.unregisterNativeControl(c.native.handle)
	}
	destroyNativeControl(c.native.handle)
	c.native.handle = 0
	c.native.commandID = 0
	c.native.oldWndProc = 0
}

// syncNativeBounds 同步复选框原生控件边界。
func (c *CheckBox) syncNativeBounds() {
	if c.native.valid() {
		setNativeBounds(c.native.handle, c.Bounds())
	}
}

// syncNativeVisible 同步复选框原生控件可见性。
func (c *CheckBox) syncNativeVisible() {
	if c.native.valid() {
		setNativeVisible(c.native.handle, c.Visible())
	}
}

// syncNativeEnabled 同步复选框原生控件启用状态。
func (c *CheckBox) syncNativeEnabled() {
	if c.native.valid() {
		setNativeEnabled(c.native.handle, c.Enabled())
	}
}

// syncNativeText 同步复选框原生控件文本。
func (c *CheckBox) syncNativeText() {
	if c.native.valid() {
		setNativeText(c.native.handle, c.Text)
	}
}

// syncNativeChecked 同步复选框原生控件勾选状态。
func (c *CheckBox) syncNativeChecked() {
	if !c.native.valid() {
		return
	}
	state := uintptr(nativeButtonStateUnchecked)
	if c.Checked {
		state = nativeButtonStateChecked
	}
	sendNativeMessage(c.native.handle, nativeButtonSetCheck, state, 0)
}

// acceptsFocus 返回控件是否可接受键盘焦点。
func (c *CheckBox) acceptsFocus() bool {
	if isNativeMode(c.mode) {
		return false
	}
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (c *CheckBox) cursor() CursorID {
	if !c.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// resolveStyle 解析多选框最终样式。
func (c *CheckBox) resolveStyle(ctx *PaintCtx) ChoiceStyle {
	style := DefaultTheme().CheckBox
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.CheckBox
	}
	return mergeChoiceStyle(style, c.Style)
}

// setChecked 更新多选框选中状态。
func (c *CheckBox) setChecked(checked bool, notify bool) {
	if c.Checked == checked {
		return
	}
	c.Checked = checked
	c.syncNativeChecked()
	c.invalidate(c)
	if notify && c.OnChange != nil {
		c.OnChange(checked)
	}
}

// RadioButton 表示互斥选择的单选按钮控件。
// RadioButton 表示互斥选择的单选按钮控件。
type RadioButton struct {
	// widgetBase 提供选项控件共享的基础控件能力。
	widgetBase
	// mode 表示单选按钮当前使用的后端模式。
	mode ControlMode
	// native 保存单选按钮在原生后端下的运行时状态。
	native nativeControlState
	// Text 保存单选按钮文本。
	Text string
	// Group 指定互斥分组名称。
	Group string
	// Checked 记录当前是否已选中。
	Checked bool
	// Hover 记录当前是否处于悬停状态。
	Hover bool
	// Down 记录当前是否处于按下状态。
	Down bool
	// Focused 记录当前是否拥有焦点。
	Focused bool
	// Style 保存样式覆盖。
	Style ChoiceStyle
	// OnChange 保存状态变更回调。
	OnChange func(bool)
}

// NewRadioButton 创建一个新的单选按钮。
func NewRadioButton(id, text string, mode ControlMode) *RadioButton {
	return &RadioButton{
		widgetBase: newWidgetBase(id, "radio"),
		mode:       normalizeControlMode(mode),
		Text:       text,
	}
}

// SetBounds 更新单选按钮边界。
func (r *RadioButton) SetBounds(rect Rect) {
	r.runOnUI(func() {
		r.widgetBase.setBounds(r, rect)
		r.syncNativeBounds()
	})
}

// SetVisible 更新单选按钮可见状态。
func (r *RadioButton) SetVisible(visible bool) {
	r.runOnUI(func() {
		r.widgetBase.setVisible(r, visible)
		r.syncNativeVisible()
	})
}

// SetEnabled 更新单选按钮可用状态。
func (r *RadioButton) SetEnabled(enabled bool) {
	r.runOnUI(func() {
		r.widgetBase.setEnabled(r, enabled)
		r.syncNativeEnabled()
	})
}

// SetText 更新单选按钮文本。
func (r *RadioButton) SetText(text string) {
	r.runOnUI(func() {
		if r.Text == text {
			return
		}
		r.Text = text
		r.syncNativeText()
		r.invalidate(r)
	})
}

// SetGroup 更新单选按钮分组。
func (r *RadioButton) SetGroup(group string) {
	r.runOnUI(func() {
		if r.Group == group {
			return
		}
		r.Group = group
		if r.Checked {
			r.syncGroup(false)
		}
	})
}

// SetChecked 更新单选按钮选中状态。
func (r *RadioButton) SetChecked(checked bool) {
	r.runOnUI(func() {
		r.setChecked(checked, false)
	})
}

// IsChecked 返回单选按钮是否选中。
func (r *RadioButton) IsChecked() bool {
	return r.Checked
}

// SetStyle 更新单选按钮样式覆盖。
func (r *RadioButton) SetStyle(style ChoiceStyle) {
	r.runOnUI(func() {
		r.Style = style
		r.invalidate(r)
	})
}

// SetOnChange 注册单选按钮变更回调。
func (r *RadioButton) SetOnChange(fn func(bool)) {
	r.runOnUI(func() {
		r.OnChange = fn
	})
}

// HitTest 判断给定点是否命中当前单选按钮。
func (r *RadioButton) HitTest(x, y int32) bool {
	if isNativeMode(r.mode) {
		return false
	}
	return r.widgetBase.HitTest(x, y)
}

// OnEvent 处理输入事件或生命周期事件。
func (r *RadioButton) OnEvent(evt Event) bool {
	if isNativeMode(r.mode) {
		return false
	}
	switch evt.Type {
	case EventMouseEnter:
		if !r.Hover {
			r.Hover = true
			r.invalidate(r)
		}
	case EventMouseLeave:
		changed := r.Hover || r.Down
		r.Hover = false
		r.Down = false
		if changed {
			r.invalidate(r)
		}
	case EventMouseDown:
		if r.Enabled() {
			r.Down = true
			r.invalidate(r)
			return true
		}
	case EventMouseUp:
		if r.Down {
			r.Down = false
			r.invalidate(r)
			return true
		}
	case EventClick:
		if r.Enabled() {
			r.setChecked(true, true)
			return true
		}
	case EventFocus:
		if !r.Focused {
			r.Focused = true
			r.invalidate(r)
		}
	case EventBlur:
		if r.Focused {
			r.Focused = false
			r.Down = false
			r.invalidate(r)
		}
	}
	return false
}

// Paint 使用给定绘制上下文完成绘制。
func (r *RadioButton) Paint(ctx *PaintCtx) {
	if isNativeMode(r.mode) || !r.Visible() || ctx == nil {
		return
	}

	style := r.resolveStyle(ctx)
	content := r.Bounds()
	if content.Empty() {
		return
	}

	boxSize := ctx.DP(style.IndicatorSizeDP)
	if boxSize <= 0 {
		boxSize = ctx.DP(18)
	}
	gap := ctx.DP(style.IndicatorGapDP)
	if gap <= 0 {
		gap = ctx.DP(10)
	}

	boxRect := Rect{
		X: content.X,
		Y: content.Y + (content.H-boxSize)/2,
		W: boxSize,
		H: boxSize,
	}
	wrapRect := Rect{X: content.X, Y: content.Y, W: content.W, H: content.H}

	if r.Hover || r.Focused {
		_ = ctx.FillRoundRect(wrapRect, choiceWrapRadius(ctx, style), style.HoverBackground)
	}

	background := style.Background
	borderColor := style.BorderColor
	textColor := style.TextColor
	if !r.Enabled() {
		background = style.DisabledBg
		borderColor = style.DisabledBorder
		textColor = style.DisabledText
	} else if r.Focused {
		borderColor = style.FocusBorder
	} else if r.Hover {
		borderColor = style.HoverBorder
	}

	radius := max32(1, boxSize/2)
	_ = ctx.FillRoundRect(boxRect, radius, background)
	if r.Checked {
		if resolveChoiceIndicatorStyle(style, true) == ChoiceIndicatorCheck {
			_ = ctx.FillRoundRect(boxRect, radius, style.IndicatorColor)
		}
		drawChoiceMark(ctx, boxRect, style, true)
	}
	_ = ctx.StrokeRoundRect(boxRect, radius, borderColor, 1)

	textRect := Rect{
		X: boxRect.X + boxRect.W + gap,
		Y: content.Y,
		W: max32(0, content.W-boxRect.W-gap),
		H: content.H,
	}
	_ = ctx.DrawText(r.Text, textRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})
}

// setScene 更新单选按钮关联的场景，并在原生模式下同步子控件生命周期。
func (r *RadioButton) setScene(scene *Scene) {
	current := r.scene()
	if current != scene {
		r.destroyNativeControl(current)
	}
	r.widgetBase.setScene(scene)
	r.ensureNativeControl(scene)
}

// Close 释放单选按钮持有的原生后端资源。
func (r *RadioButton) Close() error {
	r.runOnUI(func() {
		r.destroyNativeControl(r.scene())
	})
	return nil
}

// handleNativeCommand 处理原生单选按钮发送的命令通知。
func (r *RadioButton) handleNativeCommand(code uint16) bool {
	if !isNativeMode(r.mode) {
		return false
	}
	switch code {
	case nativeButtonSetFocus:
		if scene := r.scene(); scene != nil {
			scene.Blur()
		}
		return true
	case nativeButtonClicked:
		if r.Enabled() {
			r.setChecked(true, true)
		}
		return true
	default:
		return false
	}
}

// ensureNativeControl 确保单选按钮在原生模式下已创建系统子控件。
func (r *RadioButton) ensureNativeControl(scene *Scene) {
	if !isNativeMode(r.mode) || scene == nil || scene.app == nil {
		return
	}
	if r.native.valid() {
		r.syncNativeBounds()
		r.syncNativeVisible()
		r.syncNativeEnabled()
		r.syncNativeText()
		r.syncNativeChecked()
		return
	}
	commandID := scene.allocateNativeCommandID()
	handle, err := createNativeControl(
		scene,
		"BUTTON",
		r.Text,
		nativeWindowChild|nativeWindowVisible|nativeWindowTabStop|nativeButtonRadio,
		r.Bounds(),
		commandID,
	)
	if err != nil {
		return
	}
	r.native.handle = handle
	r.native.commandID = commandID
	scene.registerNativeControl(handle, r)
	r.syncNativeBounds()
	r.syncNativeVisible()
	r.syncNativeEnabled()
	r.syncNativeText()
	r.syncNativeChecked()
}

// destroyNativeControl 销毁单选按钮对应的原生系统子控件。
func (r *RadioButton) destroyNativeControl(scene *Scene) {
	if !r.native.valid() {
		r.native.commandID = 0
		return
	}
	if scene != nil {
		scene.unregisterNativeControl(r.native.handle)
	}
	destroyNativeControl(r.native.handle)
	r.native.handle = 0
	r.native.commandID = 0
	r.native.oldWndProc = 0
}

// syncNativeBounds 同步单选按钮原生控件边界。
func (r *RadioButton) syncNativeBounds() {
	if r.native.valid() {
		setNativeBounds(r.native.handle, r.Bounds())
	}
}

// syncNativeVisible 同步单选按钮原生控件可见性。
func (r *RadioButton) syncNativeVisible() {
	if r.native.valid() {
		setNativeVisible(r.native.handle, r.Visible())
	}
}

// syncNativeEnabled 同步单选按钮原生控件启用状态。
func (r *RadioButton) syncNativeEnabled() {
	if r.native.valid() {
		setNativeEnabled(r.native.handle, r.Enabled())
	}
}

// syncNativeText 同步单选按钮原生控件文本。
func (r *RadioButton) syncNativeText() {
	if r.native.valid() {
		setNativeText(r.native.handle, r.Text)
	}
}

// syncNativeChecked 同步单选按钮原生控件勾选状态。
func (r *RadioButton) syncNativeChecked() {
	if !r.native.valid() {
		return
	}
	state := uintptr(nativeButtonStateUnchecked)
	if r.Checked {
		state = nativeButtonStateChecked
	}
	sendNativeMessage(r.native.handle, nativeButtonSetCheck, state, 0)
}

// acceptsFocus 返回控件是否可接受键盘焦点。
func (r *RadioButton) acceptsFocus() bool {
	if isNativeMode(r.mode) {
		return false
	}
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (r *RadioButton) cursor() CursorID {
	if !r.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// resolveStyle 解析单选按钮最终样式。
func (r *RadioButton) resolveStyle(ctx *PaintCtx) ChoiceStyle {
	style := DefaultTheme().RadioButton
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.RadioButton
	}
	return mergeChoiceStyle(style, r.Style)
}

// setChecked 更新单选按钮选中状态。
func (r *RadioButton) setChecked(checked bool, notify bool) {
	if r.Checked == checked {
		return
	}
	r.Checked = checked
	r.syncNativeChecked()
	if checked {
		r.syncGroup(notify)
	}
	r.invalidate(r)
	if notify && r.OnChange != nil {
		r.OnChange(checked)
	}
}

// syncGroup 同步当前分组内其他单选按钮状态。
func (r *RadioButton) syncGroup(notify bool) {
	parent := r.parent()
	if parent == nil || r.Group == "" {
		return
	}
	for _, child := range parent.Children() {
		peer, ok := child.(*RadioButton)
		if !ok || peer == r || peer.Group != r.Group || !peer.Checked {
			continue
		}
		peer.Checked = false
		peer.syncNativeChecked()
		peer.invalidate(peer)
		if notify && peer.OnChange != nil {
			peer.OnChange(false)
		}
	}
}

// choiceWrapRadius 返回选择类控件外层悬停区域应使用的圆角半径。
func choiceWrapRadius(ctx *PaintCtx, style ChoiceStyle) int32 {
	if ctx == nil {
		return 0
	}
	radius := ctx.DP(style.CornerRadius)
	if radius <= 0 {
		return 0
	}
	return radius + ctx.DP(4)
}

// resolveChoiceIndicatorStyle 返回当前选择类控件应使用的选中标记样式。
func resolveChoiceIndicatorStyle(style ChoiceStyle, isRadio bool) ChoiceIndicatorStyle {
	if style.IndicatorStyle != ChoiceIndicatorAuto {
		return style.IndicatorStyle
	}
	if isRadio {
		return ChoiceIndicatorDot
	}
	return ChoiceIndicatorCheck
}

// drawChoiceMark 按给定样式绘制选择类控件的选中标记。
func drawChoiceMark(ctx *PaintCtx, boxRect Rect, style ChoiceStyle, isRadio bool) {
	switch resolveChoiceIndicatorStyle(style, isRadio) {
	case ChoiceIndicatorCheck:
		drawChoiceCheck(ctx, boxRect, style.CheckColor)
	default:
		color := style.CheckColor
		if isRadio {
			color = style.IndicatorColor
		}
		drawChoiceDot(ctx, boxRect, color)
	}
}

// drawChoiceDot 在选择框内部绘制圆点选中标记。
func drawChoiceDot(ctx *PaintCtx, boxRect Rect, color core.Color) {
	if ctx == nil || boxRect.Empty() {
		return
	}
	dotSize := max32(ctx.DP(6), boxRect.W/2)
	dotRect := Rect{
		X: boxRect.X + (boxRect.W-dotSize)/2,
		Y: boxRect.Y + (boxRect.H-dotSize)/2,
		W: dotSize,
		H: dotSize,
	}
	_ = ctx.FillRoundRect(dotRect, max32(1, dotSize/2), color)
}

// drawChoiceCheck 在选择框内部绘制打钩选中标记。
func drawChoiceCheck(ctx *PaintCtx, boxRect Rect, color core.Color) {
	if ctx == nil || boxRect.Empty() {
		return
	}

	segments := []Rect{
		{
			X: boxRect.X + boxRect.W/5,
			Y: boxRect.Y + boxRect.H/2,
			W: max32(1, boxRect.W/6),
			H: max32(1, boxRect.H/4),
		},
		{
			X: boxRect.X + boxRect.W/3,
			Y: boxRect.Y + boxRect.H/2 + boxRect.H/6,
			W: max32(1, boxRect.W/6),
			H: max32(1, boxRect.H/5),
		},
		{
			X: boxRect.X + boxRect.W/2,
			Y: boxRect.Y + boxRect.H/3,
			W: max32(1, boxRect.W/6),
			H: max32(1, boxRect.H/2),
		},
		{
			X: boxRect.X + boxRect.W*2/3,
			Y: boxRect.Y + boxRect.H/5,
			W: max32(1, boxRect.W/6),
			H: max32(1, boxRect.H/2),
		},
	}
	for _, segment := range segments {
		_ = ctx.FillRect(segment, color)
	}
}

// mergeChoiceStyle 把多选框或单选按钮样式覆盖合并到基础样式中。
func mergeChoiceStyle(base, override ChoiceStyle) ChoiceStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.TextColor != 0 {
		base.TextColor = override.TextColor
	}
	if override.DisabledText != 0 {
		base.DisabledText = override.DisabledText
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
	if override.IndicatorColor != 0 {
		base.IndicatorColor = override.IndicatorColor
	}
	if override.CheckColor != 0 {
		base.CheckColor = override.CheckColor
	}
	if override.IndicatorStyle != ChoiceIndicatorAuto {
		base.IndicatorStyle = override.IndicatorStyle
	}
	if override.HoverBackground != 0 {
		base.HoverBackground = override.HoverBackground
	}
	if override.DisabledBg != 0 {
		base.DisabledBg = override.DisabledBg
	}
	if override.DisabledBorder != 0 {
		base.DisabledBorder = override.DisabledBorder
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	if override.IndicatorSizeDP != 0 {
		base.IndicatorSizeDP = override.IndicatorSizeDP
	}
	if override.IndicatorGapDP != 0 {
		base.IndicatorGapDP = override.IndicatorGapDP
	}
	return base
}

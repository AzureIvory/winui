//go:build windows

package widgets

import (
	"github.com/AzureIvory/winui/core"
)

// Toggle 表示可在开与关之间切换的滑动开关控件。
// 视觉为带 ON/OFF 文案的胶囊轨道与圆形滑块，仅支持自绘模式。
type Toggle struct {
	// widgetBase 提供开关控件共享的基础控件能力。
	widgetBase
	// Checked 记录当前是否处于开启状态。
	Checked bool
	// Hover 记录当前是否处于悬停状态。
	Hover bool
	// Down 记录当前是否处于按下状态。
	Down bool
	// Focused 记录当前是否拥有焦点。
	Focused bool
	// LabelOn 保存开启状态下显示的文案，为空时回退到 "ON"。
	LabelOn string
	// LabelOff 保存关闭状态下显示的文案，为空时回退到 "OFF"。
	LabelOff string
	// Style 保存样式覆盖。
	Style ToggleStyle
	// OnChange 保存状态变更回调。
	OnChange func(bool)
}

// NewToggle 创建一个新的开关控件，并以 checked 指定初始状态。
func NewToggle(id string, checked bool) *Toggle {
	return &Toggle{
		widgetBase: newWidgetBase(id, "toggle"),
		Checked:    checked,
	}
}

// SetBounds 更新开关边界。
func (t *Toggle) SetBounds(rect Rect) {
	t.runOnUI(func() {
		t.widgetBase.setBounds(t, rect)
	})
}

// SetVisible 更新开关可见状态。
func (t *Toggle) SetVisible(visible bool) {
	t.runOnUI(func() {
		t.widgetBase.setVisible(t, visible)
	})
}

// SetEnabled 更新开关可用状态。
func (t *Toggle) SetEnabled(enabled bool) {
	t.runOnUI(func() {
		t.widgetBase.setEnabled(t, enabled)
	})
}

// SetChecked 更新开关选中状态。
func (t *Toggle) SetChecked(checked bool) {
	t.runOnUI(func() {
		t.setChecked(checked, false)
	})
}

// IsChecked 返回开关是否开启。
func (t *Toggle) IsChecked() bool {
	return t.Checked
}

// SetStyle 更新开关样式覆盖。
func (t *Toggle) SetStyle(style ToggleStyle) {
	t.runOnUI(func() {
		t.Style = style
		t.invalidate(t)
	})
}

// SetLabelOn 更新开启状态显示的文案。
func (t *Toggle) SetLabelOn(text string) {
	t.runOnUI(func() {
		if t.LabelOn == text {
			return
		}
		t.LabelOn = text
		t.invalidate(t)
	})
}

// SetLabelOff 更新关闭状态显示的文案。
func (t *Toggle) SetLabelOff(text string) {
	t.runOnUI(func() {
		if t.LabelOff == text {
			return
		}
		t.LabelOff = text
		t.invalidate(t)
	})
}

// SetOnChange 注册开关状态变更回调。
func (t *Toggle) SetOnChange(fn func(bool)) {
	t.runOnUI(func() {
		t.OnChange = fn
	})
}

// HitTest 判断给定点是否命中当前开关。
func (t *Toggle) HitTest(x, y int32) bool {
	return t.widgetBase.HitTest(x, y)
}

// OnEvent 处理输入事件或生命周期事件。
func (t *Toggle) OnEvent(evt Event) bool {
	switch evt.Type {
	case EventMouseEnter:
		if !t.Hover {
			t.Hover = true
			t.invalidate(t)
		}
	case EventMouseLeave:
		changed := t.Hover || t.Down
		t.Hover = false
		t.Down = false
		if changed {
			t.invalidate(t)
		}
	case EventMouseDown:
		if t.Enabled() {
			t.Down = true
			t.invalidate(t)
			return true
		}
	case EventMouseUp:
		if t.Down {
			t.Down = false
			t.invalidate(t)
			return true
		}
	case EventClick:
		if t.Enabled() {
			t.setChecked(!t.Checked, true)
			return true
		}
	case EventFocus:
		if !t.Focused {
			t.Focused = true
			t.invalidate(t)
		}
	case EventBlur:
		if t.Focused {
			t.Focused = false
			t.Down = false
			t.invalidate(t)
		}
	case EventKeyDown:
		// 键盘空格或回车也可切换开关，保证可访问性。
		if t.Enabled() && (evt.Key.Key == core.KeySpace || evt.Key.Key == core.KeyReturn) {
			t.setChecked(!t.Checked, true)
			return true
		}
	}
	return false
}

// Paint 使用给定绘制上下文完成绘制。
func (t *Toggle) Paint(ctx *PaintCtx) {
	if !t.Visible() || ctx == nil {
		return
	}
	bounds := t.Bounds()
	if bounds.Empty() {
		return
	}

	style := t.resolveStyle(ctx)

	// 轨道几何按样式尺寸缩放，并在控件边界内居中。
	trackW := ctx.DP(style.TrackWidthDP)
	trackH := ctx.DP(style.TrackHeightDP)
	trackRect := Rect{
		X: bounds.X + max32(0, (bounds.W-trackW)/2),
		Y: bounds.Y + max32(0, (bounds.H-trackH)/2),
		W: trackW,
		H: trackH,
	}
	radius := trackH / 2
	if style.CornerRadius > 0 {
		radius = ctx.DP(style.CornerRadius)
	}

	// 根据开关状态选择轨道色、文案色与文案。
	trackColor := style.TrackOff
	textColor := style.TextOff
	label := toggleLabelText(t.LabelOff, "OFF")
	if t.Checked {
		trackColor = style.TrackOn
		textColor = style.TextOn
		label = toggleLabelText(t.LabelOn, "ON")
	}
	// 禁用态整体压暗，提示不可交互。
	if !t.Enabled() {
		trackColor = BlendColor(trackColor, core.RGB(150, 150, 150), 110)
		textColor = BlendColor(textColor, core.RGB(150, 150, 150), 110)
	}

	_ = ctx.FillRoundRect(trackRect, radius, trackColor)

	// 文案在轨道两端：开启态靠左、关闭态靠右，恰好与滑块错开。
	pad := ctx.DP(style.LabelPaddingDP)
	labelRect := Rect{
		X: trackRect.X + pad,
		Y: trackRect.Y,
		W: max32(0, trackW-pad*2),
		H: trackH,
	}
	// 默认水平左对齐（左对齐标志为 0），关闭态叠加右对齐。
	format := core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis
	if !t.Checked {
		format |= core.DTRight
	}
	_ = ctx.DrawWidgetText(t, label, labelRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: format,
	})

	// 滑块沿水平方向在轨道两端之间移动。
	knobSize := ctx.DP(style.KnobSizeDP)
	inset := ctx.DP(style.KnobInsetDP)
	travel := max32(0, trackW-knobSize-inset*2)
	knobX := trackRect.X + inset
	if t.Checked {
		knobX = trackRect.X + inset + travel
	}
	knobColor := style.Knob
	if !t.Enabled() {
		knobColor = BlendColor(knobColor, core.RGB(150, 150, 150), 110)
	}
	_ = ctx.FillRoundRect(
		Rect{
			X: knobX,
			Y: trackRect.Y + max32(0, (trackH-knobSize)/2),
			W: knobSize,
			H: knobSize,
		},
		knobSize/2,
		knobColor,
	)
}

// Close 释放开关持有的资源（开关仅自绘，无需释放原生句柄）。
func (t *Toggle) Close() error {
	return nil
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (t *Toggle) acceptsFocus() bool {
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (t *Toggle) cursor() CursorID {
	if !t.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// resolveStyle 解析开关最终样式。
func (t *Toggle) resolveStyle(ctx *PaintCtx) ToggleStyle {
	style := DefaultTheme().Toggle
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.Toggle
	} else if scene := t.scene(); scene != nil && scene.theme != nil {
		style = scene.theme.Toggle
	}
	return mergeToggleStyle(style, t.Style)
}

// setChecked 更新开关选中状态，并按需触发回调。
func (t *Toggle) setChecked(checked bool, notify bool) {
	if t.Checked == checked {
		return
	}
	t.Checked = checked
	t.invalidate(t)
	if notify && t.OnChange != nil {
		t.OnChange(checked)
	}
}

// toggleLabelText 返回开关文案，为空时回退到默认值。
func toggleLabelText(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// mergeToggleStyle 把开关样式覆盖合并到基础样式上。
func mergeToggleStyle(base, override ToggleStyle) ToggleStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.TrackOff != 0 {
		base.TrackOff = override.TrackOff
	}
	if override.TrackOn != 0 {
		base.TrackOn = override.TrackOn
	}
	if override.Knob != 0 {
		base.Knob = override.Knob
	}
	if override.TextOn != 0 {
		base.TextOn = override.TextOn
	}
	if override.TextOff != 0 {
		base.TextOff = override.TextOff
	}
	if override.TrackWidthDP != 0 {
		base.TrackWidthDP = override.TrackWidthDP
	}
	if override.TrackHeightDP != 0 {
		base.TrackHeightDP = override.TrackHeightDP
	}
	if override.KnobSizeDP != 0 {
		base.KnobSizeDP = override.KnobSizeDP
	}
	if override.KnobInsetDP != 0 {
		base.KnobInsetDP = override.KnobInsetDP
	}
	if override.LabelPaddingDP != 0 {
		base.LabelPaddingDP = override.LabelPaddingDP
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	return base
}

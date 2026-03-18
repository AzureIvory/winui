//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// BtnKind 表示按钮里图标和文本的排布方式。
type BtnKind uint8

const (
	// BtnAuto 自动选择布局；当同时有图标和文本时默认使用上下布局。
	BtnAuto BtnKind = iota
	// BtnTop 表示图标在上、文本在下。
	BtnTop
	// BtnLeft 表示左侧小图标、右侧文本。
	BtnLeft
)

type Button struct {
	widgetBase
	Text    string
	Icon    *core.Icon
	Hover   bool
	Down    bool
	OnClick func()
	Style   ButtonStyle
	kind    BtnKind
}

// NewButton 创建一个新的按钮控件。
func NewButton(id, text string) *Button {
	return &Button{
		widgetBase: newWidgetBase(id, "button"),
		Text:       text,
	}
}

// SetBounds 更新按钮边界。
func (b *Button) SetBounds(rect Rect) {
	b.runOnUI(func() {
		b.widgetBase.setBounds(b, rect)
	})
}

// SetVisible 更新按钮可见状态。
func (b *Button) SetVisible(visible bool) {
	b.runOnUI(func() {
		b.widgetBase.setVisible(b, visible)
	})
}

// SetEnabled 更新按钮可用状态。
func (b *Button) SetEnabled(enabled bool) {
	b.runOnUI(func() {
		b.widgetBase.setEnabled(b, enabled)
	})
}

// SetText 更新按钮文本。
func (b *Button) SetText(text string) {
	b.runOnUI(func() {
		if b.Text == text {
			return
		}
		b.Text = text
		b.invalidate(b)
	})
}

// SetIcon 更新按钮图标。
func (b *Button) SetIcon(icon *core.Icon) {
	b.runOnUI(func() {
		if b.Icon == icon {
			return
		}
		b.Icon = icon
		b.invalidate(b)
	})
}

// SetStyle 更新按钮样式覆盖。
func (b *Button) SetStyle(style ButtonStyle) {
	b.runOnUI(func() {
		b.Style = style
		b.invalidate(b)
	})
}

// SetKind 设置按钮内容布局。
func (b *Button) SetKind(kind BtnKind) {
	b.runOnUI(func() {
		kind = normalizeBtnKind(kind)
		if b.kind == kind {
			return
		}
		b.kind = kind
		b.invalidate(b)
	})
}

// Kind 返回按钮当前的布局方式。
func (b *Button) Kind() BtnKind {
	return normalizeBtnKind(b.kind)
}

// SetOnClick 注册按钮点击回调。
func (b *Button) SetOnClick(fn func()) {
	b.runOnUI(func() {
		b.OnClick = fn
	})
}

// cursor 返回悬停按钮时应使用的光标。
func (b *Button) cursor() CursorID {
	if !b.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// OnEvent 处理输入事件或生命周期事件。
func (b *Button) OnEvent(evt Event) bool {
	switch evt.Type {
	case EventMouseEnter:
		if !b.Hover {
			b.Hover = true
			b.invalidate(b)
		}
	case EventMouseLeave:
		changed := b.Hover || b.Down
		b.Hover = false
		b.Down = false
		if changed {
			b.invalidate(b)
		}
	case EventMouseDown:
		if b.Enabled() {
			b.Down = true
			b.invalidate(b)
			return true
		}
	case EventMouseUp:
		if b.Down {
			b.Down = false
			b.invalidate(b)
			return true
		}
	case EventClick:
		if b.Enabled() && b.OnClick != nil {
			b.OnClick()
			return true
		}
	}
	return false
}

// Paint 使用给定绘制上下文完成按钮绘制。
func (b *Button) Paint(ctx *PaintCtx) {
	if !b.Visible() || ctx == nil {
		return
	}

	bounds := b.Bounds()
	if bounds.Empty() {
		return
	}

	style := b.resolveStyle(ctx)
	bgColor := style.Background
	textColor := style.TextColor
	borderColor := style.Border

	switch {
	case !b.Enabled():
		bgColor = style.Disabled
		textColor = style.DisabledText
	case b.Down:
		bgColor = style.Pressed
		if style.DownText != 0 {
			textColor = style.DownText
		}
	case b.Hover:
		bgColor = style.Hover
	}

	radius := ctx.DP(style.CornerRadius)
	if radius <= 0 {
		radius = ctx.DP(10)
	}

	_ = ctx.FillRoundRect(bounds, radius, bgColor)
	if borderColor != 0 {
		_ = ctx.StrokeRoundRect(bounds, radius, borderColor, 1)
	}

	kind := b.Kind()
	if kind == BtnAuto && b.Icon != nil && b.Text != "" {
		kind = BtnTop
	}

	switch {
	case b.Icon == nil:
		b.drawButtonText(ctx, bounds, style.Font, textColor)
	case b.Text == "":
		b.drawCenteredIcon(ctx, bounds, style, kind)
	case kind == BtnLeft:
		b.drawLeftIconButton(ctx, bounds, style, textColor)
	default:
		b.drawTopIconButton(ctx, bounds, style, textColor)
	}
}

// resolveStyle 解析按钮最终样式。
func (b *Button) resolveStyle(ctx *PaintCtx) ButtonStyle {
	style := DefaultTheme().Button
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.Button
	}
	if b.Style.Font.Face != "" {
		style.Font = b.Style.Font
	}
	if b.Style.TextColor != 0 {
		style.TextColor = b.Style.TextColor
	}
	if b.Style.DownText != 0 {
		style.DownText = b.Style.DownText
	}
	if b.Style.DisabledText != 0 {
		style.DisabledText = b.Style.DisabledText
	}
	if b.Style.Background != 0 {
		style.Background = b.Style.Background
	}
	if b.Style.Hover != 0 {
		style.Hover = b.Style.Hover
	}
	if b.Style.Pressed != 0 {
		style.Pressed = b.Style.Pressed
	}
	if b.Style.Disabled != 0 {
		style.Disabled = b.Style.Disabled
	}
	if b.Style.Border != 0 {
		style.Border = b.Style.Border
	}
	if b.Style.CornerRadius != 0 {
		style.CornerRadius = b.Style.CornerRadius
	}
	if b.Style.IconSizeDP != 0 {
		style.IconSizeDP = b.Style.IconSizeDP
	}
	if b.Style.TextInsetDP != 0 {
		style.TextInsetDP = b.Style.TextInsetDP
	}
	if b.Style.GapDP != 0 {
		style.GapDP = b.Style.GapDP
	}
	if b.Style.PadDP != 0 {
		style.PadDP = b.Style.PadDP
	}
	return style
}

func (b *Button) drawButtonText(ctx *PaintCtx, rect Rect, font FontSpec, color core.Color) {
	if b.Text == "" {
		return
	}
	if b.Down {
		rect = offsetRect(rect, 1, 1)
	}
	_ = ctx.DrawText(
		b.Text,
		rect,
		TextStyle{
			Font:   font,
			Color:  color,
			Format: core.DTCenter | core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		},
	)
}

func (b *Button) drawCenteredIcon(ctx *PaintCtx, rect Rect, style ButtonStyle, kind BtnKind) {
	if b.Icon == nil {
		return
	}
	iconSize := buttonIconSize(ctx, rect, style, kind)
	iconRect := Rect{
		X: rect.X + (rect.W-iconSize)/2,
		Y: rect.Y + (rect.H-iconSize)/2,
		W: iconSize,
		H: iconSize,
	}
	if b.Down {
		iconRect = offsetRect(iconRect, 1, 1)
	}
	_ = ctx.DrawIcon(b.Icon, iconRect)
}

func (b *Button) drawLeftIconButton(ctx *PaintCtx, rect Rect, style ButtonStyle, textColor core.Color) {
	if b.Icon == nil {
		b.drawButtonText(ctx, rect, style.Font, textColor)
		return
	}

	pad := ctx.DP(style.PadDP)
	if pad <= 0 {
		pad = ctx.DP(12)
	}
	gap := ctx.DP(style.GapDP)
	if gap <= 0 {
		gap = ctx.DP(8)
	}
	iconSize := buttonIconSize(ctx, rect, style, BtnLeft)

	iconRect := Rect{
		X: rect.X + pad,
		Y: rect.Y + (rect.H-iconSize)/2,
		W: iconSize,
		H: iconSize,
	}
	textRect := Rect{
		X: iconRect.X + iconRect.W + gap,
		Y: rect.Y,
		W: max32(0, rect.W-pad*2-iconSize-gap),
		H: rect.H,
	}

	if b.Down {
		iconRect = offsetRect(iconRect, 1, 1)
		textRect = offsetRect(textRect, 1, 1)
	}

	_ = ctx.DrawIcon(b.Icon, iconRect)
	_ = ctx.DrawText(
		b.Text,
		textRect,
		TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		},
	)
}

func (b *Button) drawTopIconButton(ctx *PaintCtx, rect Rect, style ButtonStyle, textColor core.Color) {
	if b.Icon == nil {
		b.drawButtonText(ctx, rect, style.Font, textColor)
		return
	}

	pad := ctx.DP(style.PadDP)
	if pad <= 0 {
		pad = ctx.DP(8)
	}
	gap := ctx.DP(style.GapDP)
	if gap <= 0 {
		gap = ctx.DP(6)
	}
	labelH := ctx.DP(style.TextInsetDP)
	if labelH <= 0 {
		labelH = ctx.DP(18)
	}

	iconSize := buttonIconSize(ctx, rect, style, BtnTop)
	maxIconSize := rect.H - pad*2 - labelH - gap
	if maxIconSize > 0 {
		iconSize = min32(iconSize, maxIconSize)
	}
	iconSize = max32(iconSize, ctx.DP(12))

	iconRect := Rect{
		X: rect.X + (rect.W-iconSize)/2,
		Y: rect.Y + pad,
		W: iconSize,
		H: iconSize,
	}
	textRect := Rect{
		X: rect.X + pad,
		Y: iconRect.Y + iconRect.H + gap,
		W: max32(0, rect.W-pad*2),
		H: max32(0, rect.Y+rect.H-pad-(iconRect.Y+iconRect.H+gap)),
	}

	if b.Down {
		iconRect = offsetRect(iconRect, 1, 1)
		textRect = offsetRect(textRect, 1, 1)
	}

	_ = ctx.DrawIcon(b.Icon, iconRect)
	_ = ctx.DrawText(
		b.Text,
		textRect,
		TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTCenter | core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		},
	)
}

func buttonIconSize(ctx *PaintCtx, rect Rect, style ButtonStyle, kind BtnKind) int32 {
	if ctx == nil {
		return 0
	}
	if style.IconSizeDP > 0 {
		return ctx.DP(style.IconSizeDP)
	}
	if kind == BtnLeft {
		return clampValue(rect.H-ctx.DP(20), ctx.DP(14), ctx.DP(18))
	}
	return clampValue(rect.H-ctx.DP(22), ctx.DP(16), ctx.DP(28))
}

func normalizeBtnKind(kind BtnKind) BtnKind {
	switch kind {
	case BtnTop, BtnLeft:
		return kind
	default:
		return BtnAuto
	}
}

func offsetRect(rect Rect, dx, dy int32) Rect {
	rect.X += dx
	rect.Y += dy
	return rect
}

//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

func mergeListBadgeStyle(base, override ListBadgeStyle) ListBadgeStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.TextColor != 0 {
		base.TextColor = override.TextColor
	}
	if override.Background != 0 {
		base.Background = override.Background
	}
	if override.PaddingDP != 0 {
		base.PaddingDP = override.PaddingDP
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	return base
}

// paintRowBadge consumes only the existing title content rectangle. The shared
// rowContentRects remains authoritative for action/check/scrollbar geometry and
// all pointer routes. The existing row/ancestor clips also apply to this label.
func (l *ListBox) paintRowBadge(ctx *PaintCtx, item ListItem, title Rect, style ListStyle) Rect {
	if item.Badge == "" || title.Empty() {
		return title
	}
	badge := mergeListBadgeStyle(style.Badge, item.BadgeStyle)
	badge.Font = mergeFontSpec(itemSubtitleFont(style.Font), badge.Font)
	size, err := ctx.MeasureWidgetText(l, item.Badge, badge.Font)
	if err != nil || size.Width <= 0 || size.Height <= 0 {
		return title
	}
	pad := max32(0, ctx.DP(badge.PaddingDP))
	gap := ctx.DP(6)
	// Keep at least half the title width for its original text. Long labels
	// ellipsize; tiny rows omit the badge rather than overlap action/check slots.
	width := min32(size.Width+2*pad, max32(0, title.W/2-gap))
	if width <= 2*pad || title.H < size.Height {
		return title
	}
	height := min32(title.H, size.Height+ctx.DP(2))
	box := Rect{X: title.X, Y: title.Y + (title.H-height)/2, W: width, H: height}
	textColor := badge.TextColor
	if textColor == 0 {
		textColor = style.TextColor
	}
	if item.Disabled || !l.Enabled() {
		textColor = style.DisabledText
	}
	if badge.Background != 0 {
		_ = ctx.FillRoundRect(box, max32(0, ctx.DP(badge.CornerRadius)), badge.Background)
	}
	_ = ctx.DrawWidgetText(l, item.Badge, Rect{X: box.X + pad, Y: box.Y, W: max32(0, box.W-2*pad), H: box.H}, TextStyle{
		Font: badge.Font, Color: textColor, Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})
	title.X += width + gap
	title.W = max32(0, title.W-width-gap)
	return title
}

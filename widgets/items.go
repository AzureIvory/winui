//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ListItem 描述列表类控件中的一个选项。
type ListItem struct {
	// Value 保存选项值。
	Value string
	// Text 保存主显示文本；为空时回退到 Value。
	Text string
	// Subtitle 保存富样式下的副标题文本。
	Subtitle string
	// Image 保存富样式下的左侧图片资源。
	Image *core.Image
	// ImagePath 保存 JSON 字面量声明的图片路径。
	ImagePath string
	// Disabled 表示该项是否禁用。
	Disabled bool
	// Background 指定该项单独的背景色；为零值时回退到列表默认背景。
	Background core.Color
	// TextColor 指定该项单独的文本颜色；为零值时回退到列表默认文本色。
	TextColor core.Color
	// Checked 记录该项侧边打勾列的勾选状态，仅在列表开启打勾列时生效。
	Checked bool
}

// displayText 返回列表项最合适的显示文本。
func (i ListItem) displayText() string {
	if i.Text != "" {
		return i.Text
	}
	return i.Value
}

func (i ListItem) hasRichContent() bool {
	return i.Subtitle != "" || i.Image != nil || i.ImagePath != ""
}

// cloneItems 返回列表项切片的浅拷贝。
func cloneItems(items []ListItem) []ListItem {
	cloned := make([]ListItem, 0, len(items))
	for _, item := range items {
		if item.Text == "" {
			item.Text = item.Value
		}
		cloned = append(cloned, item)
	}
	return cloned
}

func itemSubtitleFont(font FontSpec) FontSpec {
	if font.SizeDP > 12 {
		font.SizeDP--
	}
	if font.Weight > 500 {
		font.Weight = 500
	}
	return font
}

func itemTextLineHeight(ctx *PaintCtx, spec FontSpec, fallback int32) int32 {
	if ctx == nil || spec.SizeDP <= 0 {
		return fallback
	}
	return max32(fallback, ctx.DP(spec.SizeDP+6))
}

func itemTextLineRects(ctx *PaintCtx, bounds Rect, titleFont, subtitleFont FontSpec) (Rect, Rect) {
	if ctx == nil || bounds.Empty() {
		return Rect{}, Rect{}
	}
	titleH := itemTextLineHeight(ctx, titleFont, ctx.DP(20))
	subtitleH := itemTextLineHeight(ctx, subtitleFont, ctx.DP(16))
	gap := max32(0, ctx.DP(1))
	if total := titleH + subtitleH + gap; total > bounds.H {
		available := max32(2, bounds.H-gap)
		titleH = max32(1, available*11/20)
		subtitleH = max32(1, available-titleH)
	}
	top := max32(bounds.Y, bounds.Y+(bounds.H-titleH-subtitleH-gap)/2)
	return Rect{X: bounds.X, Y: top, W: bounds.W, H: titleH},
		Rect{X: bounds.X, Y: top + titleH + gap, W: bounds.W, H: subtitleH}
}

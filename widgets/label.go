//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// Label 表示只读文本标签控件。
type Label struct {
	// widgetBase 提供标签共享的基础控件能力。
	widgetBase
	// Text 保存标签文本。
	Text string
	// Style 保存文本样式覆盖。
	Style TextStyle
	// multiline controls whether the label allows multiple lines.
	multiline bool
	// wordWrap controls whether the label wraps long lines.
	wordWrap bool
}

// NewLabel 创建一个新的标签控件。
func NewLabel(id, text string) *Label {
	return &Label{
		widgetBase: newWidgetBase(id, "label"),
		Text:       text,
	}
}

// SetBounds 更新标签的边界。
func (l *Label) SetBounds(rect Rect) {
	l.widgetBase.setBounds(l, rect)
}

// SetVisible 更新标签的可见状态。
func (l *Label) SetVisible(visible bool) {
	l.widgetBase.setVisible(l, visible)
}

// SetEnabled 更新标签的可用状态。
func (l *Label) SetEnabled(enabled bool) {
	l.widgetBase.setEnabled(l, enabled)
}

// SetText 更新标签的显示文本。
func (l *Label) SetText(text string) {
	l.runOnUI(func() {
		if l.Text == text {
			return
		}
		l.Text = text
		l.invalidate(l)
	})
}

// SetMultiline updates whether the label allows multiple lines.
func (l *Label) SetMultiline(multiline bool) {
	l.runOnUI(func() {
		if l.multiline == multiline {
			return
		}
		l.multiline = multiline
		if panel, ok := l.parent().(*Panel); ok {
			panel.applyLayout()
			panel.invalidate(panel)
			return
		}
		l.invalidate(l)
	})
}

// Multiline reports whether the label allows multiple lines.
func (l *Label) Multiline() bool {
	return l.multiline
}

// SetWordWrap updates whether the label wraps long lines.
func (l *Label) SetWordWrap(wordWrap bool) {
	l.runOnUI(func() {
		if l.wordWrap == wordWrap {
			return
		}
		l.wordWrap = wordWrap
		if panel, ok := l.parent().(*Panel); ok {
			panel.applyLayout()
			panel.invalidate(panel)
			return
		}
		l.invalidate(l)
	})
}

// WordWrap reports whether the label wraps long lines.
func (l *Label) WordWrap() bool {
	return l.wordWrap
}

// SetStyle 更新标签的样式覆盖。
func (l *Label) SetStyle(style TextStyle) {
	l.runOnUI(func() {
		l.Style = style
		l.invalidate(l)
	})
}

// OnEvent 处理输入事件或生命周期事件。
func (l *Label) OnEvent(Event) bool {
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (l *Label) Paint(ctx *PaintCtx) {
	if !l.Visible() || l.Text == "" {
		return
	}
	style := l.resolveStyle(ctx)
	_ = ctx.DrawWidgetText(l, l.Text, l.Bounds(), style)
}

func (l *Label) preferredSize() core.Size {
	return l.preferredSizeForWidth(0)
}

func (l *Label) preferredSizeForWidth(width int32) core.Size {
	size, logical := l.preferredSizeInfo()
	if logical {
		size = scaleSizeForWidget(l, size)
	}
	if size.Width > 0 || size.Height > 0 {
		if !l.multiline || width <= 0 {
			return size
		}
		measured := l.measureTextSize(width)
		if size.Width > 0 {
			measured.Width = size.Width
		}
		if size.Height > 0 {
			measured.Height = size.Height
		}
		return measured
	}
	return l.measureTextSize(width)
}

// resolveStyle 解析标签的最终样式。
func (l *Label) resolveStyle(ctx *PaintCtx) TextStyle {
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
	style = mergeTextStyle(style, l.Style)
	style.Format = l.normalizedFormat(style.Format)
	return style
}

func (l *Label) normalizedFormat(format uint32) uint32 {
	format &^= core.DTSingleLine | core.DTEndEllipsis | core.DTVCenter | core.DTWordBreak
	if !l.multiline {
		return format | core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis
	}
	if l.wordWrap {
		format |= core.DTWordBreak
	}
	return format
}

func (l *Label) measureTextSize(width int32) core.Size {
	if l.Text == "" {
		return core.Size{}
	}
	style := l.resolveStyle(&PaintCtx{scene: l.scene()})
	layout := l.textLayout(style, width)
	if len(layout.lines) == 0 {
		return core.Size{}
	}
	measuredWidth := layout.maxLineWidth
	if width > 0 && l.multiline {
		measuredWidth = min32(width, max32(measuredWidth, 1))
	}
	return core.Size{Width: measuredWidth, Height: layout.contentHeight}
}

type labelVisualLine struct {
	width int32
}

type labelVisualLayout struct {
	lines         []labelVisualLine
	lineHeight    int32
	maxLineWidth  int32
	contentHeight int32
}

func (l *Label) textLayout(style TextStyle, contentWidth int32) labelVisualLayout {
	layout := labelVisualLayout{lineHeight: l.textLineHeight(style)}
	runes := []rune(l.Text)
	if !l.multiline {
		width := l.measureRunes(style, runes)
		layout.lines = []labelVisualLine{{width: width}}
		layout.maxLineWidth = width
		layout.contentHeight = layout.lineHeight
		return layout
	}

	appendLine := func(width int32) {
		layout.lines = append(layout.lines, labelVisualLine{width: width})
		if width > layout.maxLineWidth {
			layout.maxLineWidth = width
		}
	}

	logicalStart := 0
	for idx := 0; idx <= len(runes); idx++ {
		if idx < len(runes) && runes[idx] != '\n' {
			continue
		}
		segment := runes[logicalStart:idx]
		if l.wordWrap && contentWidth > 0 {
			for _, lineWidth := range l.wrapLogicalLine(style, segment, contentWidth) {
				appendLine(lineWidth)
			}
		} else {
			appendLine(l.measureRunes(style, segment))
		}
		logicalStart = idx + 1
	}
	if len(layout.lines) == 0 {
		layout.lines = []labelVisualLine{{}}
	}
	layout.contentHeight = int32(len(layout.lines)) * layout.lineHeight
	return layout
}

func (l *Label) wrapLogicalLine(style TextStyle, runes []rune, maxWidth int32) []int32 {
	if len(runes) == 0 {
		return []int32{0}
	}
	if maxWidth <= 0 {
		return []int32{l.measureRunes(style, runes)}
	}

	lines := make([]int32, 0, 4)
	start := 0
	for start < len(runes) {
		pos := start
		var width int32
		lastBreak := -1
		lastBreakWidth := int32(0)
		for pos < len(runes) {
			w := l.measureRune(style, runes[pos])
			if width+w > maxWidth && pos > start {
				break
			}
			width += w
			if isWrapBreak(runes[pos]) {
				lastBreak = pos + 1
				lastBreakWidth = width
			}
			pos++
		}
		end := pos
		lineWidth := width
		if pos < len(runes) && lastBreak > start {
			end = lastBreak
			lineWidth = lastBreakWidth
		}
		if end <= start {
			end = start + 1
			lineWidth = l.measureRune(style, runes[start])
		}
		lines = append(lines, lineWidth)
		start = end
	}
	return lines
}

func (l *Label) textLineHeight(style TextStyle) int32 {
	size := scaledFontHeightForWidget(l, style.Font)
	return max32(size+6, 18)
}

func (l *Label) measureRune(style TextStyle, ch rune) int32 {
	size := scaledFontHeightForWidget(l, style.Font)
	switch ch {
	case '\t':
		return max32(size*2, 8)
	case ' ':
		return max32(size/3, 4)
	case 'i', 'l', 'I', '!', '.', ',', ';', ':', '\'', '"', '`', '|':
		return max32(size/3, 4)
	case 'M', 'W', '@', '#', '%', '&':
		return max32(size*9/10, size/2)
	default:
		if ch >= 0x2E80 {
			return max32(size, 12)
		}
		return max32(size*3/5, 6)
	}
}

func (l *Label) measureRunes(style TextStyle, runes []rune) int32 {
	var width int32
	for _, ch := range runes {
		width += l.measureRune(style, ch)
	}
	return width
}

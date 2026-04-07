//go:build windows

package widgets

import (
	"github.com/AzureIvory/winui/core"
)

type editVisualLine struct {
	text  string
	start int
	end   int
	width int32
}

type editVisualLayout struct {
	lines         []editVisualLine
	lineHeight    int32
	maxLineWidth  int32
	contentHeight int32
}

func (e *EditBox) effectiveWordWrap() bool {
	return e.multiline && (e.wordWrap || !e.horizontalScroll)
}

func (e *EditBox) textLayout(style EditStyle, contentWidth int32) editVisualLayout {
	layout := editVisualLayout{lineHeight: e.textLineHeight(style)}
	runes := []rune(e.Text)
	if !e.multiline {
		width := e.measureRunes(style, runes)
		layout.lines = []editVisualLine{{text: string(runes), start: 0, end: len(runes), width: width}}
		layout.maxLineWidth = width
		layout.contentHeight = layout.lineHeight
		return layout
	}

	appendLine := func(line editVisualLine) {
		layout.lines = append(layout.lines, line)
		if line.width > layout.maxLineWidth {
			layout.maxLineWidth = line.width
		}
	}

	logicalStart := 0
	for idx := 0; idx <= len(runes); idx++ {
		if idx < len(runes) && runes[idx] != '\n' {
			continue
		}
		segment := runes[logicalStart:idx]
		if e.effectiveWordWrap() && contentWidth > 0 {
			for _, line := range e.wrapLogicalLine(style, segment, logicalStart, contentWidth) {
				appendLine(line)
			}
		} else {
			appendLine(editVisualLine{text: string(segment), start: logicalStart, end: idx, width: e.measureRunes(style, segment)})
		}
		logicalStart = idx + 1
	}
	if len(layout.lines) == 0 {
		layout.lines = []editVisualLine{{start: 0, end: 0}}
	}
	layout.contentHeight = int32(len(layout.lines)) * layout.lineHeight
	return layout
}

func (e *EditBox) wrapLogicalLine(style EditStyle, runes []rune, globalStart int, maxWidth int32) []editVisualLine {
	if len(runes) == 0 {
		return []editVisualLine{{text: "", start: globalStart, end: globalStart, width: 0}}
	}
	if maxWidth <= 0 {
		return []editVisualLine{{text: string(runes), start: globalStart, end: globalStart + len(runes), width: e.measureRunes(style, runes)}}
	}

	lines := make([]editVisualLine, 0, 4)
	start := 0
	for start < len(runes) {
		pos := start
		var width int32
		lastBreak := -1
		lastBreakWidth := int32(0)
		for pos < len(runes) {
			w := e.measureRune(style, runes[pos])
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
			lineWidth = e.measureRune(style, runes[start])
		}
		lines = append(lines, editVisualLine{
			text:  string(runes[start:end]),
			start: globalStart + start,
			end:   globalStart + end,
			width: lineWidth,
		})
		start = end
	}
	return lines
}

func (e *EditBox) textLineHeight(style EditStyle) int32 {
	size := style.Font.SizeDP
	if size <= 0 {
		size = 15
	}
	if scene := e.scene(); scene != nil && scene.app != nil {
		size = scene.app.DP(size)
	}
	return max32(size+6, 18)
}

func (e *EditBox) measureRune(style EditStyle, ch rune) int32 {
	size := style.Font.SizeDP
	if size <= 0 {
		size = 15
	}
	if scene := e.scene(); scene != nil && scene.app != nil {
		size = scene.app.DP(size)
	}
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

func (e *EditBox) measureRunes(style EditStyle, runes []rune) int32 {
	var width int32
	for _, ch := range runes {
		width += e.measureRune(style, ch)
	}
	return width
}

func (e *EditBox) caretRect(style EditStyle, textRect Rect) Rect {
	layout := e.textLayout(style, textRect.W)
	lineIndex, line := e.findLineForCaret(layout)
	lineX := e.caretXInLine(style, line, e.caret)
	caretX := textRect.X - e.scrollX + lineX
	caretY := textRect.Y - e.scrollY
	if e.multiline {
		caretY += int32(lineIndex) * layout.lineHeight
	} else {
		caretY = e.Bounds().Y + max32(0, (e.Bounds().H-layout.lineHeight)/2)
	}
	return Rect{
		X: caretX,
		Y: caretY,
		W: max32(1, e.dp(2)),
		H: max32(1, layout.lineHeight),
	}
}

func (e *EditBox) caretAtPoint(point core.Point) int {
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	layout := e.textLayout(style, textRect.W)
	if !e.multiline {
		line := layout.lines[0]
		localX := point.X - textRect.X + e.scrollX
		if localX <= 0 {
			return 0
		}
		return e.caretFromLineX(style, line, localX)
	}
	if point.Y <= textRect.Y {
		return layout.lines[0].start
	}
	localY := point.Y - textRect.Y + e.scrollY
	lineIndex := int(localY / layout.lineHeight)
	lineIndex = clampInt(lineIndex, 0, len(layout.lines)-1)
	line := layout.lines[lineIndex]
	localX := point.X - textRect.X + e.scrollX
	if localX <= 0 {
		return line.start
	}
	return e.caretFromLineX(style, line, localX)
}

func (e *EditBox) moveCaretVertical(delta int) (int, bool) {
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	layout := e.textLayout(style, textRect.W)
	lineIndex, line := e.findLineForCaret(layout)
	target := lineIndex + delta
	if target < 0 || target >= len(layout.lines) {
		return e.caret, false
	}
	x := e.caretXInLine(style, line, e.caret)
	caret := e.caretFromLineX(style, layout.lines[target], x)
	return caret, caret != e.caret
}

func (e *EditBox) lineEdgeCaret(end bool) (int, bool) {
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	layout := e.textLayout(style, textRect.W)
	_, line := e.findLineForCaret(layout)
	caret := line.start
	if end {
		caret = line.end
	}
	return caret, caret != e.caret
}

func (e *EditBox) findLineForCaret(layout editVisualLayout) (int, editVisualLine) {
	if len(layout.lines) == 0 {
		return 0, editVisualLine{}
	}
	for idx, line := range layout.lines {
		if e.caret < line.start {
			return idx, line
		}
		if e.caret >= line.start && e.caret <= line.end {
			if idx+1 < len(layout.lines) && layout.lines[idx+1].start > line.end && e.caret == layout.lines[idx+1].start {
				continue
			}
			return idx, line
		}
		if idx+1 < len(layout.lines) && e.caret < layout.lines[idx+1].start {
			return idx, line
		}
	}
	return len(layout.lines) - 1, layout.lines[len(layout.lines)-1]
}

func (e *EditBox) caretXInLine(style EditStyle, line editVisualLine, caret int) int32 {
	if caret <= line.start {
		return 0
	}
	if caret >= line.end {
		return line.width
	}
	runes := []rune(line.text)
	count := clampInt(caret-line.start, 0, len(runes))
	return e.measureRunes(style, runes[:count])
}

func (e *EditBox) caretFromLineX(style EditStyle, line editVisualLine, x int32) int {
	runes := []rune(line.text)
	if len(runes) == 0 || x <= 0 {
		return line.start
	}
	var width int32
	for idx, ch := range runes {
		cw := e.measureRune(style, ch)
		if x <= width+cw/2 {
			return line.start + idx
		}
		width += cw
	}
	return line.end
}

func (e *EditBox) dp(value int32) int32 {
	if scene := e.scene(); scene != nil && scene.app != nil {
		return scene.app.DP(value)
	}
	return value
}

func isWrapBreak(ch rune) bool {
	switch ch {
	case ' ', '\t', '-', '/', '\\':
		return true
	default:
		return false
	}
}

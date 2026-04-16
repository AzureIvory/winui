//go:build windows

package widgets

import (
	"strings"
	"unicode/utf8"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

const (
	// editKeyFlagCtrl 为测试和跨层路由保留的 Control 修饰位。
	editKeyFlagCtrl uint64 = 1 << 63
)

// EditBox 表示支持单行和多行模式的文本输入控件。
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

	// multiline 记录当前是否启用多行模式。
	multiline bool
	// wordWrap 记录当前是否显式启用自动换行。
	wordWrap bool
	// verticalScroll 记录当前是否允许垂直滚轮滚动。
	verticalScroll bool
	// horizontalScroll 记录当前是否允许水平滚动。
	horizontalScroll bool
	// acceptReturn 记录多行模式下回车是否插入换行。
	acceptReturn bool
	// password 记录当前是否启用密码掩码显示。
	password bool
	// scrollX 保存当前水平滚动偏移。
	scrollX int32
	// scrollY 保存当前垂直滚动偏移。
	scrollY int32
}

// NewEditBox 创建一个新的编辑框。
func NewEditBox(id string, mode ControlMode) *EditBox {
	return &EditBox{
		widgetBase:       newWidgetBase(id, "edit"),
		mode:             normalizeControlMode(mode),
		wordWrap:         true,
		verticalScroll:   true,
		horizontalScroll: false,
		acceptReturn:     true,
	}
}

// SetBounds 更新编辑框的边界。
func (e *EditBox) SetBounds(rect Rect) {
	e.runOnUI(func() {
		if e.Bounds() == rect {
			return
		}
		e.widgetBase.setBounds(e, rect)
		e.clampScroll()
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
		normalized := e.normalizeText(text)
		if e.Text == normalized {
			return
		}
		e.Text = normalized
		e.caret = len([]rune(normalized))
		e.scrollX = 0
		e.scrollY = 0
		e.syncNativeText()
		e.ScrollToCaret()
		e.invalidate(e)
	})
}

// TextValue 返回编辑框当前保存的文本。
func (e *EditBox) TextValue() string {
	if e.native.valid() {
		e.Text = e.normalizeText(getNativeText(e.native.handle))
		_, end := getNativeSelection(e.native.handle)
		e.caret = clampInt(end, 0, len([]rune(e.Text)))
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

// SetPassword 更新编辑框的密码模式。启用后显示层使用掩码字符，内部文本仍保留真实值。
func (e *EditBox) SetPassword(password bool) {
	e.runOnUI(func() {
		if e.password == password {
			return
		}
		e.password = password
		e.syncNativePassword()
		e.invalidate(e)
	})
}

// Password 返回编辑框是否启用密码模式。
func (e *EditBox) Password() bool {
	return e.password
}

// SetMultiline 更新编辑框的多行模式。
func (e *EditBox) SetMultiline(v bool) {
	e.runOnUI(func() {
		if e.multiline == v {
			return
		}
		e.multiline = v
		e.Text = e.normalizeText(e.Text)
		e.caret = clampInt(e.caret, 0, len([]rune(e.Text)))
		e.scrollX = 0
		e.scrollY = 0
		e.recreateNativeControl()
		e.invalidate(e)
	})
}

// Multiline 返回编辑框是否启用多行模式。
func (e *EditBox) Multiline() bool {
	return e.multiline
}

// SetWordWrap 更新编辑框的自动换行偏好。
func (e *EditBox) SetWordWrap(v bool) {
	e.runOnUI(func() {
		if e.wordWrap == v {
			return
		}
		e.wordWrap = v
		e.scrollX = 0
		e.clampScroll()
		e.recreateNativeControl()
		e.invalidate(e)
	})
}

// WordWrap 返回编辑框是否启用自动换行偏好。
func (e *EditBox) WordWrap() bool {
	return e.wordWrap
}

// SetVerticalScroll 更新编辑框的垂直滚动支持状态。
func (e *EditBox) SetVerticalScroll(v bool) {
	e.runOnUI(func() {
		if e.verticalScroll == v {
			return
		}
		e.verticalScroll = v
		e.recreateNativeControl()
		e.invalidate(e)
	})
}

// VerticalScroll 返回编辑框是否启用垂直滚动支持。
func (e *EditBox) VerticalScroll() bool {
	return e.verticalScroll
}

// SetHorizontalScroll 更新编辑框的水平滚动支持状态。
func (e *EditBox) SetHorizontalScroll(v bool) {
	e.runOnUI(func() {
		if e.horizontalScroll == v {
			return
		}
		e.horizontalScroll = v
		if !v {
			e.scrollX = 0
		}
		e.clampScroll()
		e.recreateNativeControl()
		e.invalidate(e)
	})
}

// HorizontalScroll 返回编辑框是否启用水平滚动支持。
func (e *EditBox) HorizontalScroll() bool {
	return e.horizontalScroll
}

// SetAcceptReturn 更新多行模式下回车是否插入换行。
func (e *EditBox) SetAcceptReturn(v bool) {
	e.runOnUI(func() {
		if e.acceptReturn == v {
			return
		}
		e.acceptReturn = v
		e.recreateNativeControl()
		e.invalidate(e)
	})
}

// AcceptReturn 返回多行模式下回车是否插入换行。
func (e *EditBox) AcceptReturn() bool {
	return e.acceptReturn
}

// ScrollToCaret 将当前光标滚动到可见区域内。
func (e *EditBox) ScrollToCaret() {
	e.runOnUI(func() {
		if e.native.valid() {
			sendNativeMessage(e.native.handle, nativeEditScrollCaret, 0, 0)
			return
		}
		e.ensureCaretVisible()
	})
}

// LineCount 返回当前逻辑行数。
func (e *EditBox) LineCount() int {
	text := e.TextValue()
	if text == "" {
		return 1
	}
	return strings.Count(text, "\n") + 1
}

// SetStyle 更新编辑框的样式覆盖。
func (e *EditBox) SetStyle(style EditStyle) {
	e.runOnUI(func() {
		e.Style = style
		e.syncNativeBounds()
		e.syncNativeFont()
		e.syncNativeTheme()
		e.invalidate(e)
		e.syncNativeInsets()
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
	case EventMouseDown:
		if e.Enabled() {
			e.caret = e.caretAtPoint(evt.Point)
			e.ensureNativeControl(e.scene())
			e.syncNativeBounds()
			e.syncNativeVisible()
			if e.native.valid() {
				setNativeSelection(e.native.handle, e.caret, e.caret)
				setNativeFocus(e.native.handle)
				e.syncNativeVisible()
			}
			e.ensureCaretVisible()
			e.invalidate(e)
			return true
		}
	case EventClick:
		if e.Enabled() {
			return true
		}
	case EventFocus:
		if !e.Focused {
			e.Focused = true
			e.caret = clampInt(e.caret, 0, len([]rune(e.Text)))
			e.ensureNativeControl(e.scene())
			e.syncNativeBounds()
			e.syncNativeVisible()
			e.syncNativeEnabled()
			e.syncNativeTheme()
			e.syncNativeText()
			e.syncNativeReadOnly()
			e.syncNativePassword()
			if e.native.valid() {
				setNativeSelection(e.native.handle, e.caret, e.caret)
				setNativeFocus(e.native.handle)
				e.syncNativeVisible()
			}
			e.ensureCaretVisible()
			e.invalidate(e)
		}
	case EventBlur:
		if e.Focused {
			if e.native.valid() {
				text := e.normalizeText(getNativeText(e.native.handle))
				if e.Text != text {
					e.Text = text
					if e.OnChange != nil {
						e.OnChange(text)
					}
				}
				_, end := getNativeSelection(e.native.handle)
				e.caret = clampInt(end, 0, len([]rune(e.Text)))
			}
			e.Focused = false
			e.syncNativeVisible()
			e.invalidate(e)
		}
	case EventMouseWheel:
		return e.handleWheel(evt)
	case EventKeyDown:
		if e.native.valid() && e.Focused {
			return true
		}
		return e.handleKey(evt.Key)
	case EventChar:
		if e.native.valid() && e.Focused {
			return true
		}
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
		Y: bounds.Y + padding,
		W: max32(0, bounds.W-padding*2),
		H: max32(0, bounds.H-padding*2),
	}
	if textRect.Empty() {
		return
	}

	if e.native.valid() && e.Focused {
		return
	}

	restore := ctx.PushClipRect(textRect)
	defer restore()

	displayText := e.visualText()
	if displayText == "" {
		displayText = e.Placeholder
		textColor = style.PlaceholderColor
	}

	if !e.multiline {
		textRect.Y = bounds.Y
		textRect.H = bounds.H
		drawRect := textRect
		drawRect.X -= e.scrollX
		_ = ctx.DrawText(displayText, drawRect, TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: alignTextFormat(style.TextAlign, core.DTVCenter|core.DTSingleLine|core.DTEndEllipsis),
		})
	} else {
		layout := e.textLayout(style, textRect.W)
		lineHeight := layout.lineHeight
		baseY := textRect.Y - e.scrollY
		for idx, line := range layout.lines {
			lineY := baseY + int32(idx)*lineHeight
			lineRect := Rect{X: textRect.X - e.scrollX, Y: lineY, W: max32(layout.maxLineWidth, textRect.W), H: lineHeight}
			if lineRect.Y+lineRect.H < textRect.Y || lineRect.Y > textRect.Y+textRect.H {
				continue
			}
			_ = ctx.DrawText(line.text, lineRect, TextStyle{
				Font:   style.Font,
				Color:  textColor,
				Format: alignTextFormat(style.TextAlign, 0),
			})
		}
		if displayText == e.Placeholder {
			_ = ctx.DrawText(displayText, Rect{X: textRect.X, Y: textRect.Y, W: textRect.W, H: lineHeight}, TextStyle{
				Font:   style.Font,
				Color:  textColor,
				Format: alignTextFormat(style.TextAlign, 0),
			})
		}
	}

	if !e.Focused {
		return
	}

	caretRect := e.caretRect(style, textRect)
	if caretRect.Empty() {
		return
	}
	visibleCaret := intersectRect(caretRect, textRect)
	if visibleCaret.Empty() {
		return
	}
	_ = ctx.FillRect(visibleCaret, style.CaretColor)
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
	switch code {
	case nativeEditSetFocus:
		if scene := e.scene(); scene != nil {
			if isNativeMode(e.mode) {
				scene.Blur()
			} else {
				scene.setFocus(e)
			}
		}
		e.syncNativeVisible()
		e.invalidate(e)
		return true
	case nativeEditKillFocus:
		if scene := e.scene(); scene != nil && scene.Focus() == e {
			scene.setFocus(nil)
			return true
		}
		e.syncNativeVisible()
		e.invalidate(e)
		return true
	case nativeEditChanged:
		text := e.Text
		if e.native.valid() {
			text = e.normalizeText(getNativeText(e.native.handle))
			_, end := getNativeSelection(e.native.handle)
			e.caret = clampInt(end, 0, len([]rune(text)))
		}
		if e.Text == text {
			e.invalidate(e)
			return true
		}
		e.Text = text
		e.clampScroll()
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
	if scene == nil || scene.app == nil {
		return
	}
	if e.native.valid() {
		e.syncNativeBounds()
		e.syncNativeFont()
		e.syncNativeVisible()
		e.syncNativeEnabled()
		e.syncNativeTheme()
		e.syncNativeText()
		e.syncNativePlaceholder()
		e.syncNativeReadOnly()
		e.syncNativePassword()
		return
	}
	commandID := scene.allocateNativeCommandID()
	handle, err := createNativeRichEditControl(
		scene,
		e.nativeEditStyle(),
		e.nativeHostBounds(),
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
	e.syncNativeTheme()
	e.syncNativeText()
	e.syncNativePlaceholder()
	e.syncNativeReadOnly()
	e.syncNativePassword()
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
		setNativeBounds(e.native.handle, e.nativeHostBounds())
		e.syncNativeInsets()
	}
}

// syncNativeVisible 同步编辑框原生控件可见性。
func (e *EditBox) syncNativeVisible() {
	if !e.native.valid() {
		return
	}
	visible := e.Visible()
	if !isNativeMode(e.mode) {
		visible = visible && e.Focused
	}
	setNativeVisible(e.native.handle, visible)
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
		setNativeText(e.native.handle, e.nativeTextValue())
		setNativeSelection(e.native.handle, e.caret, e.caret)
	}
}

// syncNativePlaceholder 同步编辑框原生控件占位提示。
func (e *EditBox) syncNativePlaceholder() {
	if e.native.valid() {
		// RichEdit 不支持 EM_SETCUEBANNER；自绘模式下占位文本由外层壳负责绘制。
	}
}

// syncNativeReadOnly 同步编辑框原生控件只读状态。
func (e *EditBox) syncNativeReadOnly() {
	if e.native.valid() {
		setNativeReadOnly(e.native.handle, e.ReadOnly)
	}
}

func (e *EditBox) syncNativePassword() {
	if e.native.valid() {
		setNativePassword(e.native.handle, e.password)
	}
}

func (e *EditBox) syncNativeTheme() {
	if !e.native.valid() {
		return
	}
	e.syncNativeFont()
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	background := style.Background
	textColor := style.TextColor
	if !e.Enabled() {
		background = style.DisabledBg
		textColor = style.DisabledText
	}
	setNativeRichEditBackgroundColor(e.native.handle, background)
	setNativeRichEditTextColor(e.native.handle, textColor)
}

func (e *EditBox) syncNativeInsets() {
	if !e.native.valid() {
		return
	}
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	padding := style.PaddingDP
	if scene := e.scene(); scene != nil && scene.app != nil {
		padding = scene.app.DP(padding)
	}

	b := e.Bounds()
	if e.multiline {
		setNativeEditRect(e.native.handle, nativeRect{
			Left:   padding,
			Top:    padding,
			Right:  max32(padding, b.W-padding),
			Bottom: max32(padding, b.H-padding),
		})
		return
	}

	lineHeight := e.textLineHeight(style)
	top := max32(0, (b.H-lineHeight)/2)
	setNativeEditRect(e.native.handle, nativeRect{
		Left:   padding,
		Top:    top,
		Right:  max32(padding, b.W-padding),
		Bottom: max32(top, b.H-top),
	})
}

func (e *EditBox) nativeHostBounds() Rect {
	return e.Bounds()
}

func (e *EditBox) syncNativeFont() {
	if !e.native.valid() {
		return
	}
	scene := e.scene()
	if scene == nil {
		return
	}
	style := e.resolveStyle(&PaintCtx{scene: scene})
	font, err := scene.font(style.Font)
	if err != nil || font == nil {
		return
	}
	setNativeControlFont(e.native.handle, font.Handle())
}

func setNativeControlFont(handle windows.Handle, font windows.Handle) {
	if handle == 0 || font == 0 {
		return
	}
	sendNativeMessage(handle, nativeMessageSetFont, uintptr(font), 1)
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

func (e *EditBox) stableMouseHit(x, y int32) bool {
	if isNativeMode(e.mode) || !e.Visible() || !e.Enabled() {
		return false
	}
	return e.Bounds().Contains(x, y)
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
			e.ensureCaretVisible()
			e.invalidate(e)
		}
		return true
	case core.KeyRight:
		if e.caret < len(runes) {
			e.caret++
			e.ensureCaretVisible()
			e.invalidate(e)
		}
		return true
	case core.KeyUp:
		if e.multiline {
			if caret, changed := e.moveCaretVertical(-1); changed {
				e.caret = caret
				e.ensureCaretVisible()
				e.invalidate(e)
			}
			return true
		}
	case core.KeyDown:
		if e.multiline {
			if caret, changed := e.moveCaretVertical(1); changed {
				e.caret = caret
				e.ensureCaretVisible()
				e.invalidate(e)
			}
			return true
		}
	case core.KeyHome:
		if e.multiline {
			if caret, changed := e.lineEdgeCaret(false); changed {
				e.caret = caret
				e.ensureCaretVisible()
				e.invalidate(e)
			}
			return true
		}
		if e.caret != 0 {
			e.caret = 0
			e.ensureCaretVisible()
			e.invalidate(e)
		}
		return true
	case core.KeyEnd:
		if e.multiline {
			if caret, changed := e.lineEdgeCaret(true); changed {
				e.caret = caret
				e.ensureCaretVisible()
				e.invalidate(e)
			}
			return true
		}
		if e.caret != len(runes) {
			e.caret = len(runes)
			e.ensureCaretVisible()
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
		e.ensureCaretVisible()
		return true
	case core.KeyDelete:
		if e.ReadOnly || e.caret >= len(runes) || len(runes) == 0 {
			return true
		}
		e.Text = string(append(runes[:e.caret], runes[e.caret+1:]...))
		e.notifyChanged()
		e.ensureCaretVisible()
		return true
	case core.KeyReturn:
		if e.multiline {
			if keyHasCtrl(key) {
				e.submit()
				return true
			}
			if e.acceptReturn && !e.ReadOnly {
				e.insertRune('\n')
				return true
			}
			e.submit()
			return true
		}
		e.submit()
		return true
	}
	return false
}

// handleChar 处理编辑框的字符输入。
func (e *EditBox) handleChar(ch rune) bool {
	if !e.Enabled() || e.ReadOnly {
		return false
	}
	if ch == '\r' || ch == '\n' {
		return false
	}
	if ch < 32 && ch != '\t' {
		return false
	}
	e.insertRune(ch)
	return true
}

// notifyChanged 使控件失效并触发变更回调。
func (e *EditBox) notifyChanged() {
	e.Text = e.normalizeText(e.Text)
	e.syncNativeText()
	e.clampScroll()
	e.invalidate(e)
	if e.OnChange != nil {
		e.OnChange(e.Text)
	}
}

// mergeEditStyle 将编辑框样式覆盖合并到基础样式上。
func mergeEditStyle(base, override EditStyle) EditStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.TextAlign != 0 {
		base.TextAlign = override.TextAlign
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

func (e *EditBox) normalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if !e.multiline {
		text = strings.ReplaceAll(text, "\n", " ")
	}
	return text
}

func (e *EditBox) nativeTextValue() string {
	if !e.multiline {
		return e.Text
	}
	return strings.ReplaceAll(e.Text, "\n", "\r\n")
}

func (e *EditBox) visualText() string {
	if !e.password {
		return e.Text
	}
	count := utf8.RuneCountInString(e.Text)
	if count <= 0 {
		return ""
	}
	return strings.Repeat("•", count)
}

func (e *EditBox) recreateNativeControl() {
	if !e.native.valid() {
		return
	}
	scene := e.scene()
	e.destroyNativeControl(scene)
	e.ensureNativeControl(scene)
}

func (e *EditBox) nativeEditStyle() uint32 {
	style := nativeWindowChild | nativeWindowTabStop
	if isNativeMode(e.mode) {
		style |= nativeWindowVisible | nativeWindowBorder
	}
	if !e.multiline {
		style |= nativeEditAutoHScroll
		return style
	}
	style |= nativeEditMultiline | nativeEditAutoVScroll
	if e.acceptReturn {
		style |= nativeEditWantReturn
	}
	if e.verticalScroll {
		style |= nativeWindowVScroll
	}
	if e.horizontalScroll || !e.effectiveWordWrap() {
		style |= nativeWindowHScroll | nativeEditAutoHScroll
	}
	return style
}

func (e *EditBox) submit() {
	if e.OnSubmit != nil {
		e.OnSubmit(e.Text)
	}
}

func (e *EditBox) insertRune(ch rune) {
	runes := []rune(e.Text)
	e.caret = clampInt(e.caret, 0, len(runes))
	runes = append(runes[:e.caret], append([]rune{ch}, runes[e.caret:]...)...)
	e.Text = string(runes)
	e.caret++
	e.notifyChanged()
	e.ensureCaretVisible()
}

func (e *EditBox) handleWheel(evt Event) bool {
	if !e.multiline || !e.verticalScroll {
		return false
	}
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	layout := e.textLayout(style, textRect.W)
	maxY := max32(0, layout.contentHeight-textRect.H)
	if maxY <= 0 {
		return false
	}
	step := layout.lineHeight
	if step <= 0 {
		step = 20
	}
	delta := -evt.Delta * step / 120
	if delta == 0 {
		if evt.Delta > 0 {
			delta = -step
		} else if evt.Delta < 0 {
			delta = step
		}
	}
	old := e.scrollY
	e.scrollY = clampValue(e.scrollY+delta, 0, maxY)
	if old != e.scrollY {
		e.invalidate(e)
		return true
	}
	return false
}

func (e *EditBox) ensureCaretVisible() {
	if !e.multiline && !e.horizontalScroll {
		e.scrollX = 0
		return
	}
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	if textRect.Empty() {
		return
	}
	layout := e.textLayout(style, textRect.W)
	caretRect := e.caretRect(style, textRect)
	if e.multiline {
		maxY := max32(0, layout.contentHeight-textRect.H)
		if caretRect.Y < textRect.Y {
			e.scrollY = clampValue(e.scrollY-(textRect.Y-caretRect.Y), 0, maxY)
		} else if caretRect.Y+caretRect.H > textRect.Y+textRect.H {
			e.scrollY = clampValue(e.scrollY+(caretRect.Y+caretRect.H-(textRect.Y+textRect.H)), 0, maxY)
		}
	} else {
		e.scrollY = 0
	}
	if e.horizontalScroll {
		maxX := max32(0, layout.maxLineWidth-textRect.W)
		if caretRect.X < textRect.X {
			e.scrollX = clampValue(e.scrollX-(textRect.X-caretRect.X), 0, maxX)
		} else if caretRect.X+caretRect.W > textRect.X+textRect.W {
			e.scrollX = clampValue(e.scrollX+(caretRect.X+caretRect.W-(textRect.X+textRect.W)), 0, maxX)
		}
	} else {
		e.scrollX = 0
	}
}

func (e *EditBox) clampScroll() {
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	textRect := e.contentRect(style)
	if textRect.Empty() {
		e.scrollX = 0
		e.scrollY = 0
		return
	}
	layout := e.textLayout(style, textRect.W)
	if !e.horizontalScroll {
		e.scrollX = 0
	} else {
		e.scrollX = clampValue(e.scrollX, 0, max32(0, layout.maxLineWidth-textRect.W))
	}
	if !e.multiline {
		e.scrollY = 0
	} else {
		e.scrollY = clampValue(e.scrollY, 0, max32(0, layout.contentHeight-textRect.H))
	}
}

func (e *EditBox) contentRect(style EditStyle) Rect {
	padding := style.PaddingDP
	if scene := e.scene(); scene != nil && scene.app != nil {
		padding = scene.app.DP(padding)
	}
	bounds := e.Bounds()
	return Rect{
		X: bounds.X + padding,
		Y: bounds.Y + padding,
		W: max32(0, bounds.W-padding*2),
		H: max32(0, bounds.H-padding*2),
	}
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func alignTextFormat(align Alignment, base uint32) uint32 {
	base &^= core.DTCenter | core.DTRight
	switch normalizeAlignment(align, AlignStart) {
	case AlignCenter:
		return base | core.DTCenter
	case AlignEnd:
		return base | core.DTRight
	default:
		return base
	}
}

func keyHasCtrl(key core.KeyEvent) bool {
	return uint64(key.Flags)&editKeyFlagCtrl != 0 || keyHasCtrlState()
}

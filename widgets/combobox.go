//go:build windows

package widgets

import (
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

// ComboBox 表示带弹出列表的选择控件。
type ComboBox struct {
	// widgetBase 提供组合框共享的基础控件能力。
	widgetBase
	// mode 表示组合框当前使用的后端模式。
	mode ControlMode
	// native 保存组合框在原生后端下的运行时状态。
	native nativeControlState
	// items 保存可选项目集合。
	items []ListItem
	// selected 保存当前选中索引。
	selected int
	// hover 保存弹出层当前悬停索引。
	hover int
	// focused 记录控件是否拥有焦点。
	focused bool
	// open 记录弹出层是否展开。
	open bool
	// Placeholder 保存未选中时显示的占位文本。
	Placeholder string
	// Style 保存样式覆盖。
	Style ComboStyle
	// OnChange 保存选择变更回调。
	OnChange func(int, ListItem)
}

// NewComboBox 创建一个新的组合框。
func NewComboBox(id string, mode ControlMode) *ComboBox {
	return &ComboBox{
		widgetBase: newWidgetBase(id, "combobox"),
		mode:       normalizeControlMode(mode),
		selected:   -1,
		hover:      -1,
	}
}

// SetBounds 更新组合框的边界。
func (c *ComboBox) SetBounds(rect Rect) {
	c.runOnUI(func() {
		c.widgetBase.setBounds(c, rect)
		c.syncNativeBounds()
	})
}

// SetVisible 更新组合框的可见状态。
func (c *ComboBox) SetVisible(visible bool) {
	c.runOnUI(func() {
		oldRect := widgetDirtyRect(c)
		changed := c.Visible() != visible
		if !visible && (c.open || c.hover != -1) {
			c.open = false
			c.hover = -1
			changed = true
		}
		if !changed {
			return
		}
		c.widgetBase.setVisible(c, visible)
		c.syncNativeVisible()
		c.invalidateStateChange(oldRect)
	})
}

// SetEnabled 更新组合框的可用状态。
func (c *ComboBox) SetEnabled(enabled bool) {
	c.runOnUI(func() {
		oldRect := widgetDirtyRect(c)
		changed := c.Enabled() != enabled
		if !enabled && (c.open || c.hover != -1) {
			c.open = false
			c.hover = -1
			changed = true
		}
		if !changed {
			return
		}
		c.widgetBase.setEnabled(c, enabled)
		c.syncNativeEnabled()
		c.invalidateStateChange(oldRect)
	})
}

// SetItems 更新组合框的项目集合。
func (c *ComboBox) SetItems(items []ListItem) {
	c.runOnUI(func() {
		c.updateState(func() bool {
			c.items = cloneItems(items)
			if len(c.items) == 0 {
				c.selected = -1
				c.hover = -1
				c.open = false
			} else {
				if c.selected >= len(c.items) {
					c.selected = len(c.items) - 1
				}
				if c.hover >= len(c.items) {
					c.hover = -1
				}
			}
			return true
		})
		c.syncNativeItems()
		c.syncNativeSelection()
	})
}

// Items 返回组合框所管理项目的副本。
func (c *ComboBox) Items() []ListItem {
	return cloneItems(c.items)
}

// SetSelected 更新组合框的当前选择。
func (c *ComboBox) SetSelected(index int) {
	c.runOnUI(func() {
		c.selectIndex(index, false)
	})
}

// SelectedIndex 返回组合框当前选中的索引。
func (c *ComboBox) SelectedIndex() int {
	return c.selected
}

// SelectedItem 返回组合框当前选中的项目。
func (c *ComboBox) SelectedItem() (ListItem, bool) {
	if c.selected < 0 || c.selected >= len(c.items) {
		return ListItem{}, false
	}
	return c.items[c.selected], true
}

// SetPlaceholder 更新组合框的占位文本。
func (c *ComboBox) SetPlaceholder(text string) {
	c.runOnUI(func() {
		c.updateState(func() bool {
			if c.Placeholder == text {
				return false
			}
			c.Placeholder = text
			return true
		})
	})
}

// SetStyle 更新组合框的样式覆盖。
func (c *ComboBox) SetStyle(style ComboStyle) {
	c.runOnUI(func() {
		c.updateState(func() bool {
			c.Style = style
			return true
		})
	})
}

// SetOnChange 注册组合框的变更回调。
func (c *ComboBox) SetOnChange(fn func(int, ListItem)) {
	c.runOnUI(func() {
		c.OnChange = fn
	})
}

// HitTest 判断给定点是否命中当前控件。
func (c *ComboBox) HitTest(x, y int32) bool {
	if isNativeMode(c.mode) {
		return false
	}
	if !c.Visible() {
		return false
	}
	if c.widgetBase.HitTest(x, y) {
		return true
	}
	if c.open {
		return c.popupRect().Contains(x, y)
	}
	return false
}

func (c *ComboBox) overlayHitTest(x, y int32) bool {
	if isNativeMode(c.mode) || !c.Visible() || !c.open {
		return false
	}
	return c.popupRect().Contains(x, y)
}

// OnEvent 处理输入事件或生命周期事件。
func (c *ComboBox) OnEvent(evt Event) bool {
	if isNativeMode(c.mode) {
		return false
	}
	switch evt.Type {
	case EventMouseMove:
		if c.open {
			index := c.popupIndexAt(evt.Point)
			c.updateState(func() bool {
				if c.hover == index {
					return false
				}
				c.hover = index
				return true
			})
			return index >= 0
		}
	case EventMouseLeave:
		c.updateState(func() bool {
			if c.hover == -1 {
				return false
			}
			c.hover = -1
			return true
		})
	case EventMouseDown:
		if c.Enabled() {
			return true
		}
	case EventMouseUp:
		if c.Enabled() {
			return true
		}
	case EventClick:
		if !c.Enabled() {
			return false
		}
		if c.bounds.Contains(evt.Point.X, evt.Point.Y) {
			c.updateState(func() bool {
				open := !c.open && len(c.items) > 0
				if c.open == open && (open || c.hover == -1) {
					return false
				}
				c.open = open
				if !c.open {
					c.hover = -1
				}
				return true
			})
			return true
		}
		if c.open {
			index := c.popupIndexAt(evt.Point)
			if index >= 0 {
				c.selectIndex(index, true)
				c.updateState(func() bool {
					changed := c.open || c.hover != -1
					c.open = false
					c.hover = -1
					return changed
				})
				return true
			}
		}
	case EventFocus:
		c.updateState(func() bool {
			if c.focused {
				return false
			}
			c.focused = true
			return true
		})
	case EventBlur:
		c.updateState(func() bool {
			if !c.focused && !c.open && c.hover == -1 {
				return false
			}
			c.focused = false
			c.open = false
			c.hover = -1
			return true
		})
	case EventKeyDown:
		if c.handleKey(evt.Key) {
			return true
		}
	}
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (c *ComboBox) Paint(ctx *PaintCtx) {
	if isNativeMode(c.mode) || !c.Visible() || ctx == nil {
		return
	}

	style := c.resolveStyle(ctx)
	bounds := c.Bounds()
	if bounds.Empty() {
		return
	}

	borderColor := style.BorderColor
	if c.focused || c.open {
		borderColor = style.FocusBorder
	} else if c.hover >= 0 {
		borderColor = style.HoverBorder
	}

	_ = ctx.FillRoundRect(bounds, ctx.DP(style.CornerRadius), style.Background)
	_ = ctx.StrokeRoundRect(bounds, ctx.DP(style.CornerRadius), borderColor, 1)

	if c.usesRichLayout(style) {
		c.paintRichField(ctx, style, bounds)
		return
	}

	text := c.Placeholder
	textColor := style.PlaceholderColor
	if item, ok := c.SelectedItem(); ok {
		text = item.displayText()
		textColor = style.TextColor
	}

	padding := ctx.DP(style.PaddingDP)
	arrowW := ctx.DP(28)
	textRect := Rect{
		X: bounds.X + padding,
		Y: bounds.Y,
		W: max32(0, bounds.W-padding*2-arrowW),
		H: bounds.H,
	}
	arrowRect := Rect{
		X: bounds.X + bounds.W - arrowW - padding/2,
		Y: bounds.Y,
		W: arrowW,
		H: bounds.H,
	}
	_ = ctx.DrawWidgetText(c, text, textRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})
	arrow := "v"
	if c.open {
		arrow = "^"
	}
	_ = ctx.DrawWidgetText(c, arrow, arrowRect, TextStyle{
		Font: FontSpec{
			Face:   style.Font.Face,
			SizeDP: style.Font.SizeDP,
			Weight: 700,
		},
		Color:  style.ArrowColor,
		Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
	})
}

// PaintOverlay 在常规控件树绘制完成后绘制覆盖层内容。
func (c *ComboBox) PaintOverlay(ctx *PaintCtx) {
	if isNativeMode(c.mode) || !c.Visible() || !c.open || ctx == nil {
		return
	}

	style := c.resolveStyle(ctx)
	layout := c.popupLayout()
	if layout.rect.Empty() || layout.start >= layout.end {
		return
	}
	popup := layout.rect

	radius := ctx.DP(style.CornerRadius)
	_ = ctx.FillRoundRect(popup, radius, style.PopupBackground)
	_ = ctx.StrokeRoundRect(popup, radius, style.FocusBorder, 1)

	if c.usesRichLayout(style) {
		c.paintRichPopup(ctx, style, layout)
		return
	}

	itemRadius := radius - ctx.DP(2)
	if itemRadius < 0 {
		itemRadius = 0
	}

	for index := layout.start; index < layout.end; index++ {
		item := c.items[index]
		rowRect := c.popupRowRectForLayout(index, layout)
		if rowRect.Empty() {
			continue
		}
		textColor := style.TextColor
		if item.Disabled {
			textColor = style.PlaceholderColor
		}
		if index == c.selected {
			_ = ctx.FillRoundRect(rowRect, itemRadius, style.ItemSelectedColor)
			textColor = style.ItemTextColor
		} else if index == c.hover {
			_ = ctx.FillRoundRect(rowRect, itemRadius, style.ItemHoverColor)
		}

		textRect := Rect{
			X: rowRect.X + ctx.DP(10),
			Y: rowRect.Y,
			W: max32(0, rowRect.W-ctx.DP(20)),
			H: rowRect.H,
		}
		_ = ctx.DrawWidgetText(c, item.displayText(), textRect, TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
	}
}

func (c *ComboBox) usesRichLayout(style ComboStyle) bool {
	return !isNativeMode(c.mode) && style.Layout == ComboLayoutRich
}

// paintRichField 负责绘制左图标和双行文本的组合框主区域。
func (c *ComboBox) paintRichField(ctx *PaintCtx, style ComboStyle, bounds Rect) {
	padding := ctx.DP(style.PaddingDP)
	arrowW := ctx.DP(28)
	contentRect := Rect{
		X: bounds.X + padding,
		Y: bounds.Y + padding/2,
		W: max32(0, bounds.W-padding*2-arrowW),
		H: max32(0, bounds.H-padding),
	}
	arrowRect := Rect{
		X: bounds.X + bounds.W - arrowW - padding/2,
		Y: bounds.Y,
		W: arrowW,
		H: bounds.H,
	}

	slotRect, textRect := comboRichContentRects(ctx, contentRect, 0)
	item, hasItem := c.SelectedItem()
	if slotRect.W > 0 && slotRect.H > 0 {
		_ = ctx.FillRoundRect(slotRect, max32(1, slotRect.W/3), style.IconBackground)
		if hasItem {
			_ = drawComboItemImage(ctx, item, comboRichImageRect(ctx, style, slotRect))
		}
	}

	titleColor := style.PlaceholderColor
	titleText := c.Placeholder
	subtitleText := ""
	if hasItem {
		titleColor = style.TextColor
		titleText = item.displayText()
		subtitleText = item.Subtitle
	}

	if subtitleText == "" {
		_ = ctx.DrawWidgetText(c, titleText, textRect, TextStyle{
			Font:   comboRichTitleFont(style),
			Color:  titleColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
	} else {
		titleRect, subtitleRect := comboRichTextLineRects(ctx, style, textRect)
		_ = ctx.DrawWidgetText(c, titleText, titleRect, TextStyle{
			Font:   comboRichTitleFont(style),
			Color:  titleColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
		_ = ctx.DrawWidgetText(c, subtitleText, subtitleRect, TextStyle{
			Font:   comboRichSubtitleFont(style),
			Color:  style.SubtitleColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
	}

	arrow := "v"
	if c.open {
		arrow = "^"
	}
	_ = ctx.DrawWidgetText(c, arrow, arrowRect, TextStyle{
		Font: FontSpec{
			Face:   style.Font.Face,
			SizeDP: style.Font.SizeDP,
			Weight: 700,
		},
		Color:  style.ArrowColor,
		Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
	})
}

// paintRichPopup 负责绘制富样式下拉层，包含右侧选中勾标记。
func (c *ComboBox) paintRichPopup(ctx *PaintCtx, style ComboStyle, layout comboPopupLayout) {
	itemRadius := max32(0, ctx.DP(style.CornerRadius)-ctx.DP(2))
	for index := layout.start; index < layout.end; index++ {
		item := c.items[index]
		rowRect := c.popupRowRectForLayout(index, layout)
		if rowRect.Empty() {
			continue
		}

		titleColor := style.TextColor
		subtitleColor := style.SubtitleColor
		if item.Disabled {
			titleColor = style.PlaceholderColor
			subtitleColor = style.PlaceholderColor
		}
		if index == c.selected {
			_ = ctx.FillRoundRect(rowRect, itemRadius, style.ItemSelectedColor)
			if style.ItemTextColor != 0 {
				titleColor = style.ItemTextColor
			}
		} else if index == c.hover {
			_ = ctx.FillRoundRect(rowRect, itemRadius, style.ItemHoverColor)
		}

		slotRect, textRect := comboRichContentRects(ctx, insetRect(rowRect, ctx.DP(10), 0), ctx.DP(34))
		if slotRect.W > 0 && slotRect.H > 0 {
			_ = ctx.FillRoundRect(slotRect, max32(1, slotRect.W/3), style.IconBackground)
			_ = drawComboItemImage(ctx, item, comboRichImageRect(ctx, style, slotRect))
		}

		if item.Subtitle == "" {
			_ = ctx.DrawWidgetText(c, item.displayText(), textRect, TextStyle{
				Font:   comboRichTitleFont(style),
				Color:  titleColor,
				Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
			})
		} else {
			titleRect, subtitleRect := comboRichTextLineRects(ctx, style, textRect)
			_ = ctx.DrawWidgetText(c, item.displayText(), titleRect, TextStyle{
				Font:   comboRichTitleFont(style),
				Color:  titleColor,
				Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
			})
			_ = ctx.DrawWidgetText(c, item.Subtitle, subtitleRect, TextStyle{
				Font:   comboRichSubtitleFont(style),
				Color:  subtitleColor,
				Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
			})
		}
		if index == c.selected {
			markRect := comboSelectedMarkRect(ctx, rowRect)
			_ = ctx.FillRoundRect(markRect, max32(1, markRect.W/2), style.SelectedMarkColor)
			drawComboSelectedMark(ctx, markRect, core.RGB(255, 255, 255))
		}
	}
}

// setScene 更新组合框关联的场景，并在原生模式下同步子控件生命周期。
func (c *ComboBox) setScene(scene *Scene) {
	current := c.scene()
	if current != scene {
		c.destroyNativeControl(current)
	}
	c.widgetBase.setScene(scene)
	c.ensureNativeControl(scene)
}

// Close 释放组合框持有的原生后端资源。
func (c *ComboBox) Close() error {
	c.runOnUI(func() {
		c.destroyNativeControl(c.scene())
	})
	return nil
}

// handleNativeCommand 处理原生组合框发送的命令通知。
func (c *ComboBox) handleNativeCommand(code uint16) bool {
	if !isNativeMode(c.mode) {
		return false
	}
	switch code {
	case nativeComboSetFocus:
		if scene := c.scene(); scene != nil {
			scene.Blur()
		}
		return true
	case nativeComboSelectionChanged:
		index := int(int32(sendNativeMessage(c.native.handle, nativeComboGetCurSel, 0, 0)))
		if index < 0 || index >= len(c.items) || c.items[index].Disabled {
			c.syncNativeSelection()
			return true
		}
		if c.selected == index {
			return true
		}
		c.updateState(func() bool {
			c.selected = index
			return true
		})
		if c.OnChange != nil {
			c.OnChange(index, c.items[index])
		}
		return true
	default:
		return false
	}
}

// ensureNativeControl 确保组合框在原生模式下已创建系统子控件。
func (c *ComboBox) ensureNativeControl(scene *Scene) {
	if !isNativeMode(c.mode) || scene == nil || scene.app == nil {
		return
	}
	if c.native.valid() {
		c.syncNativeBounds()
		c.syncNativeVisible()
		c.syncNativeEnabled()
		c.syncNativeItems()
		c.syncNativeSelection()
		return
	}
	commandID := scene.allocateNativeCommandID()
	handle, err := createNativeControl(
		scene,
		"COMBOBOX",
		"",
		nativeWindowChild|nativeWindowVisible|nativeWindowTabStop|nativeWindowVScroll|nativeComboDropDownList,
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
	c.syncNativeItems()
	c.syncNativeSelection()
}

// destroyNativeControl 销毁组合框对应的原生系统子控件。
func (c *ComboBox) destroyNativeControl(scene *Scene) {
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

// syncNativeBounds 同步组合框原生控件边界。
func (c *ComboBox) syncNativeBounds() {
	if c.native.valid() {
		setNativeBounds(c.native.handle, c.Bounds())
	}
}

// syncNativeVisible 同步组合框原生控件可见性。
func (c *ComboBox) syncNativeVisible() {
	if c.native.valid() {
		setNativeVisible(c.native.handle, c.Visible())
	}
}

// syncNativeEnabled 同步组合框原生控件启用状态。
func (c *ComboBox) syncNativeEnabled() {
	if c.native.valid() {
		setNativeEnabled(c.native.handle, c.Enabled())
	}
}

// syncNativeItems 同步组合框原生控件的项目列表。
func (c *ComboBox) syncNativeItems() {
	if !c.native.valid() {
		return
	}
	sendNativeMessage(c.native.handle, nativeComboResetContent, 0, 0)
	for _, item := range c.items {
		ptr, err := windows.UTF16PtrFromString(item.displayText())
		if err != nil {
			continue
		}
		sendNativeMessage(c.native.handle, nativeComboAddString, 0, uintptr(unsafe.Pointer(ptr)))
	}
}

// syncNativeSelection 同步组合框原生控件的当前选择。
func (c *ComboBox) syncNativeSelection() {
	if !c.native.valid() {
		return
	}
	if c.selected < 0 || c.selected >= len(c.items) || c.items[c.selected].Disabled {
		sendNativeMessage(c.native.handle, nativeComboSetCurSel, ^uintptr(0), 0)
		return
	}
	sendNativeMessage(c.native.handle, nativeComboSetCurSel, uintptr(c.selected), 0)
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (c *ComboBox) acceptsFocus() bool {
	return !isNativeMode(c.mode)
}

// cursor 返回悬停控件时应使用的光标。
func (c *ComboBox) cursor() CursorID {
	if isNativeMode(c.mode) {
		return core.CursorArrow
	}
	if !c.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// resolveStyle 解析组合框的最终样式。
func (c *ComboBox) resolveStyle(ctx *PaintCtx) ComboStyle {
	style := DefaultTheme().ComboBox
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.ComboBox
	}
	return mergeComboStyle(style, c.Style)
}

// handleKey 处理组合框的按键事件。
func (c *ComboBox) handleKey(key core.KeyEvent) bool {
	if !c.Enabled() {
		return false
	}

	switch key.Key {
	case core.KeyReturn, core.KeySpace:
		if len(c.items) == 0 {
			return true
		}
		c.updateState(func() bool {
			c.open = !c.open
			if !c.open {
				c.hover = -1
			}
			return true
		})
		return true
	case core.KeyEscape:
		if c.open {
			c.updateState(func() bool {
				c.open = false
				c.hover = -1
				return true
			})
			return true
		}
	case core.KeyDown:
		if len(c.items) == 0 {
			return true
		}
		c.updateState(func() bool {
			if c.open {
				return false
			}
			c.open = true
			return true
		})
		c.selectRelative(1)
		c.updateState(func() bool {
			if c.hover == c.selected {
				return false
			}
			c.hover = c.selected
			return true
		})
		return true
	case core.KeyUp:
		if len(c.items) == 0 {
			return true
		}
		c.updateState(func() bool {
			if c.open {
				return false
			}
			c.open = true
			return true
		})
		c.selectRelative(-1)
		c.updateState(func() bool {
			if c.hover == c.selected {
				return false
			}
			c.hover = c.selected
			return true
		})
		return true
	case core.KeyHome:
		if len(c.items) == 0 {
			return true
		}
		c.selectIndex(0, true)
		c.updateState(func() bool {
			if c.hover == c.selected {
				return false
			}
			c.hover = c.selected
			return true
		})
		return true
	case core.KeyEnd:
		if len(c.items) == 0 {
			return true
		}
		c.selectIndex(len(c.items)-1, true)
		c.updateState(func() bool {
			if c.hover == c.selected {
				return false
			}
			c.hover = c.selected
			return true
		})
		return true
	}
	return false
}

// selectRelative 按给定步长移动当前选择，并跳过禁用项。
func (c *ComboBox) selectRelative(step int) {
	index := c.selected
	if index < 0 {
		if step >= 0 {
			index = -1
		} else {
			index = len(c.items)
		}
	}
	for {
		index += step
		if index < 0 || index >= len(c.items) {
			return
		}
		if !c.items[index].Disabled {
			c.selectIndex(index, true)
			return
		}
	}
}

// selectIndex 将当前选择设置为指定项索引。
func (c *ComboBox) selectIndex(index int, notify bool) {
	if index < 0 || index >= len(c.items) || c.items[index].Disabled {
		return
	}
	if c.selected == index {
		return
	}
	c.updateState(func() bool {
		c.selected = index
		c.syncNativeSelection()
		return true
	})
	if notify && c.OnChange != nil {
		c.OnChange(index, c.items[index])
	}
}

// popupRect 返回组合框弹出层的边界。
func (c *ComboBox) popupRect() Rect {
	return c.popupLayout().rect
}

// popupIndexAt 返回弹出层指定位置对应的项索引。
func (c *ComboBox) popupIndexAt(point core.Point) int {
	layout := c.popupLayout()
	if layout.rect.Empty() || !layout.rect.Contains(point.X, point.Y) || layout.start >= layout.end {
		return -1
	}
	if point.Y < layout.rect.Y+layout.padding {
		return -1
	}
	index := int((point.Y - layout.rect.Y - layout.padding) / layout.itemHeight)
	if index < 0 || layout.start+index >= layout.end {
		return -1
	}
	rowRect := c.popupRowRectForLayout(layout.start+index, layout)
	if rowRect.Empty() || !rowRect.Contains(point.X, point.Y) {
		return -1
	}
	return layout.start + index
}

// dp 按应用当前 DPI 缩放设备无关值。
func (c *ComboBox) dp(value int32) int32 {
	if scene := c.scene(); scene != nil && scene.app != nil {
		return scene.app.DP(value)
	}
	return value
}

// mergeComboStyle 将组合框样式覆盖合并到基础样式上。
type comboPopupLayout struct {
	rect       Rect
	start      int
	end        int
	upward     bool
	itemHeight int32
	padding    int32
}

func (c *ComboBox) popupRowRectForLayout(index int, layout comboPopupLayout) Rect {
	if layout.rect.Empty() || index < layout.start || index >= layout.end {
		return Rect{}
	}
	offset := int32(index - layout.start)
	y := layout.rect.Y + layout.padding + offset*layout.itemHeight
	height := min32(layout.itemHeight, max32(0, layout.rect.Y+layout.rect.H-y-layout.padding))
	return Rect{
		X: layout.rect.X + layout.padding,
		Y: y,
		W: max32(0, layout.rect.W-layout.padding*2),
		H: height,
	}
}

func (c *ComboBox) popupLayout() comboPopupLayout {
	layout := comboPopupLayout{}
	if !c.open || len(c.items) == 0 {
		return layout
	}

	style := c.popupStyle()
	layout.itemHeight = comboPopupItemHeight(c, style)
	layout.padding = max32(0, c.dp(style.PaddingDP))

	maxVisible := int(style.MaxVisibleItems)
	if maxVisible <= 0 || maxVisible > len(c.items) {
		maxVisible = len(c.items)
	}
	if maxVisible == 0 {
		return layout
	}

	fullHeight := layout.padding*2 + int32(maxVisible)*layout.itemHeight
	gap := max32(0, c.dp(6))
	viewport := c.popupViewport()
	if viewport.Empty() {
		layout.start, layout.end = c.popupRangeForVisibleCount(maxVisible)
		layout.rect = Rect{
			X: c.bounds.X,
			Y: c.bounds.Y + c.bounds.H + gap,
			W: c.bounds.W,
			H: fullHeight,
		}
		return layout
	}

	downSpace := max32(0, viewport.Y+viewport.H-(c.bounds.Y+c.bounds.H+gap))
	upSpace := max32(0, c.bounds.Y-gap-viewport.Y)
	downCount := c.popupVisibleCountForSpace(downSpace, layout.padding, layout.itemHeight)
	upCount := c.popupVisibleCountForSpace(upSpace, layout.padding, layout.itemHeight)

	visibleCount := maxVisible
	availableSpace := downSpace

	switch {
	case fullHeight <= downSpace:
	case upSpace > downSpace:
		layout.upward = true
		availableSpace = upSpace
		visibleCount = min(maxVisible, upCount)
	default:
		visibleCount = min(maxVisible, downCount)
	}

	if !layout.upward && visibleCount == 0 && upCount > 0 {
		layout.upward = true
		availableSpace = upSpace
		visibleCount = min(maxVisible, upCount)
	}
	if layout.upward && visibleCount == 0 && downCount > 0 {
		layout.upward = false
		availableSpace = downSpace
		visibleCount = min(maxVisible, downCount)
	}
	if visibleCount <= 0 {
		if availableSpace > layout.padding*2 {
			visibleCount = 1
		} else {
			return layout
		}
	}
	if visibleCount <= 0 {
		return layout
	}

	layout.start, layout.end = c.popupRangeForVisibleCount(visibleCount)
	layout.rect = Rect{
		X: c.bounds.X,
		W: c.bounds.W,
		H: layout.padding*2 + int32(visibleCount)*layout.itemHeight,
	}
	if layout.rect.H > availableSpace {
		layout.rect.H = availableSpace
	}
	if layout.upward {
		layout.rect.Y = c.bounds.Y - gap - layout.rect.H
		if layout.rect.Y < viewport.Y {
			layout.rect.Y = viewport.Y
		}
	} else {
		layout.rect.Y = c.bounds.Y + c.bounds.H + gap
		maxY := viewport.Y + viewport.H - layout.rect.H
		if layout.rect.Y > maxY {
			layout.rect.Y = maxY
		}
	}
	return layout
}

func (c *ComboBox) popupStyle() ComboStyle {
	return c.resolveStyle(&PaintCtx{scene: c.scene()})
}

func (c *ComboBox) popupViewport() Rect {
	scene := c.scene()
	if scene == nil {
		return Rect{}
	}
	if scene.root != nil {
		rect := scene.root.Bounds()
		if !rect.Empty() {
			return rect
		}
	}
	if scene.app != nil {
		size := scene.app.ClientSize()
		return Rect{W: size.Width, H: size.Height}
	}
	return Rect{}
}

func (c *ComboBox) popupRangeForVisibleCount(visible int) (int, int) {
	if len(c.items) == 0 || visible <= 0 {
		return 0, 0
	}
	if visible > len(c.items) {
		visible = len(c.items)
	}
	start := 0
	if c.selected >= visible {
		start = c.selected - visible + 1
	}
	end := start + visible
	if end > len(c.items) {
		end = len(c.items)
		start = end - visible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func (c *ComboBox) popupVisibleCountForSpace(space, padding, itemHeight int32) int {
	contentHeight := space - padding*2
	if contentHeight <= 0 || itemHeight <= 0 {
		return 0
	}
	return int(contentHeight / itemHeight)
}

func comboPopupItemHeight(widget *ComboBox, style ComboStyle) int32 {
	height := max32(1, widget.dp(style.ItemHeightDP))
	if style.Layout == ComboLayoutRich {
		height = max32(height, widget.dp(60))
	}
	return height
}

func comboRichTitleFont(style ComboStyle) FontSpec {
	font := style.Font
	if font.Weight < 700 {
		font.Weight = 700
	}
	return font
}

func comboRichSubtitleFont(style ComboStyle) FontSpec {
	font := style.Font
	if font.SizeDP > 12 {
		font.SizeDP--
	}
	if font.Weight > 500 {
		font.Weight = 500
	}
	return font
}

func comboRichContentRects(ctx *PaintCtx, bounds Rect, markWidth int32) (Rect, Rect) {
	if ctx == nil || bounds.Empty() {
		return Rect{}, Rect{}
	}
	slotSize := clampValue(
		bounds.H-ctx.DP(10),
		ctx.DP(30),
		ctx.DP(36),
	)
	slotRect := Rect{
		X: bounds.X,
		Y: bounds.Y + (bounds.H-slotSize)/2,
		W: slotSize,
		H: slotSize,
	}
	textRect := Rect{
		X: slotRect.X + slotRect.W + ctx.DP(12),
		Y: bounds.Y,
		W: max32(0, bounds.W-slotRect.W-ctx.DP(12)-markWidth),
		H: bounds.H,
	}
	return slotRect, textRect
}

func comboRichTextLineRects(ctx *PaintCtx, style ComboStyle, bounds Rect) (Rect, Rect) {
	if ctx == nil || bounds.Empty() {
		return Rect{}, Rect{}
	}
	titleH := comboRichLineHeight(ctx, comboRichTitleFont(style), ctx.DP(20))
	subtitleH := comboRichLineHeight(ctx, comboRichSubtitleFont(style), ctx.DP(16))
	gap := max32(0, ctx.DP(1))
	if total := titleH + subtitleH + gap; total > bounds.H {
		available := max32(2, bounds.H-gap)
		titleH = max32(1, available*11/20)
		subtitleH = max32(1, available-titleH)
	}
	totalH := titleH + gap + subtitleH
	top := bounds.Y + (bounds.H-totalH)/2
	if top < bounds.Y {
		top = bounds.Y
	}
	return Rect{
			X: bounds.X,
			Y: top,
			W: bounds.W,
			H: titleH,
		}, Rect{
			X: bounds.X,
			Y: top + titleH + gap,
			W: bounds.W,
			H: subtitleH,
		}
}

func comboRichLineHeight(ctx *PaintCtx, spec FontSpec, fallback int32) int32 {
	if ctx == nil {
		return fallback
	}
	if spec.SizeDP <= 0 {
		return fallback
	}
	return max32(fallback, ctx.DP(spec.SizeDP+6))
}

func comboRichImageRect(ctx *PaintCtx, style ComboStyle, slot Rect) Rect {
	if ctx == nil || slot.Empty() {
		return Rect{}
	}
	size := ctx.DP(style.ImageSizeDP)
	if size <= 0 {
		size = ctx.DP(18)
	}
	size = min32(size, min32(slot.W, slot.H))
	return Rect{
		X: slot.X + (slot.W-size)/2,
		Y: slot.Y + (slot.H-size)/2,
		W: size,
		H: size,
	}
}

func drawComboItemImage(ctx *PaintCtx, item ListItem, rect Rect) error {
	if ctx == nil || rect.Empty() || item.Image == nil {
		return nil
	}
	canvas := ctx.Canvas()
	if canvas == nil {
		return nil
	}
	src := item.Image.NaturalSize()
	fitted := core.FitContain(src, rect.W, rect.H)
	if fitted.Width <= 0 || fitted.Height <= 0 {
		return nil
	}
	bitmap, err := item.Image.BitmapFor(
		fitted.Width,
		fitted.Height,
		core.ChooseScaleQuality(src, fitted),
	)
	if err != nil {
		return err
	}
	return canvas.DrawBitmapAlpha(bitmap, Rect{
		X: rect.X + (rect.W-fitted.Width)/2,
		Y: rect.Y + (rect.H-fitted.Height)/2,
		W: fitted.Width,
		H: fitted.Height,
	}, 255)
}

func comboSelectedMarkRect(ctx *PaintCtx, rowRect Rect) Rect {
	size := clampValue(rowRect.H-ctx.DP(30), ctx.DP(18), ctx.DP(22))
	return Rect{
		X: rowRect.X + rowRect.W - ctx.DP(14) - size,
		Y: rowRect.Y + (rowRect.H-size)/2,
		W: size,
		H: size,
	}
}

func drawComboSelectedMark(ctx *PaintCtx, rect Rect, color core.Color) {
	if ctx == nil || rect.Empty() {
		return
	}
	stroke := max32(ctx.DP(2), rect.W/7)
	geometry := choiceCheckMarkGeometry{
		Start: core.Point{
			X: rect.X + rect.W*5/22,
			Y: rect.Y + rect.H*11/22,
		},
		Mid: core.Point{
			X: rect.X + rect.W*9/22,
			Y: rect.Y + rect.H*15/22,
		},
		End: core.Point{
			X: rect.X + rect.W*16/22,
			Y: rect.Y + rect.H*7/22,
		},
		Stroke: stroke,
	}
	canvas := ctx.Canvas()
	if canvas == nil {
		return
	}
	for _, quad := range [][]core.Point{
		choiceStrokeQuad(geometry.Start, geometry.Mid, geometry.Stroke),
		choiceStrokeQuad(geometry.Mid, geometry.End, geometry.Stroke),
	} {
		if len(quad) == 0 {
			continue
		}
		_ = canvas.FillPolygon(quad, color)
	}
	for _, point := range []core.Point{geometry.Start, geometry.Mid, geometry.End} {
		_ = ctx.FillRoundRect(
			choiceCheckCapRect(point, geometry.Stroke),
			max32(1, geometry.Stroke/2),
			color,
		)
	}
}

func insetRect(rect Rect, dx, dy int32) Rect {
	return Rect{
		X: rect.X + dx,
		Y: rect.Y + dy,
		W: max32(0, rect.W-dx*2),
		H: max32(0, rect.H-dy*2),
	}
}

func (c *ComboBox) updateState(fn func() bool) {
	if fn == nil {
		return
	}
	oldRect := widgetDirtyRect(c)
	if !fn() {
		return
	}
	c.invalidateStateChange(oldRect)
}

func (c *ComboBox) invalidateStateChange(oldRect Rect) {
	if scene := c.scene(); scene != nil {
		if !oldRect.Empty() {
			scene.invalidateRect(oldRect)
		}
		scene.Invalidate(c)
	}
}

func (c *ComboBox) dirtyRect() Rect {
	rect := unionRect(c.Bounds(), c.popupRect())
	if !c.open || rect.Empty() {
		return rect
	}
	// The popup overlay border is antialiased and can touch pixels just outside
	// popupRect, so keep one extra device pixel dirty while the popup is open.
	return insetRect(rect, -1, -1)
}

func mergeComboStyle(base, override ComboStyle) ComboStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.Layout != ComboLayoutAuto {
		base.Layout = override.Layout
	}
	if override.TextColor != 0 {
		base.TextColor = override.TextColor
	}
	if override.SubtitleColor != 0 {
		base.SubtitleColor = override.SubtitleColor
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
	if override.ArrowColor != 0 {
		base.ArrowColor = override.ArrowColor
	}
	if override.PopupBackground != 0 {
		base.PopupBackground = override.PopupBackground
	}
	if override.IconBackground != 0 {
		base.IconBackground = override.IconBackground
	}
	if override.ItemHoverColor != 0 {
		base.ItemHoverColor = override.ItemHoverColor
	}
	if override.ItemSelectedColor != 0 {
		base.ItemSelectedColor = override.ItemSelectedColor
	}
	if override.ItemTextColor != 0 {
		base.ItemTextColor = override.ItemTextColor
	}
	if override.SelectedMarkColor != 0 {
		base.SelectedMarkColor = override.SelectedMarkColor
	}
	if override.ItemHeightDP != 0 {
		base.ItemHeightDP = override.ItemHeightDP
	}
	if override.PaddingDP != 0 {
		base.PaddingDP = override.PaddingDP
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	if override.ImageSizeDP != 0 {
		base.ImageSizeDP = override.ImageSizeDP
	}
	if override.MaxVisibleItems != 0 {
		base.MaxVisibleItems = override.MaxVisibleItems
	}
	return base
}

//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ComboBox 表示带弹出列表的选择控件。
type ComboBox struct {
	widgetBase
	items       []ListItem
	selected    int
	hover       int
	focused     bool
	open        bool
	Placeholder string
	Style       ComboStyle
	OnChange    func(int, ListItem)
}

// NewComboBox 创建一个新的组合框。
func NewComboBox(id string) *ComboBox {
	return &ComboBox{
		widgetBase: newWidgetBase(id, "combobox"),
		selected:   -1,
		hover:      -1,
	}
}

// SetBounds 更新组合框的边界。
func (c *ComboBox) SetBounds(rect Rect) {
	c.runOnUI(func() {
		c.widgetBase.setBounds(c, rect)
		c.invalidateAll()
	})
}

// SetVisible 更新组合框的可见状态。
func (c *ComboBox) SetVisible(visible bool) {
	c.runOnUI(func() {
		if !visible {
			c.open = false
		}
		c.widgetBase.setVisible(c, visible)
		c.invalidateAll()
	})
}

// SetEnabled 更新组合框的可用状态。
func (c *ComboBox) SetEnabled(enabled bool) {
	c.runOnUI(func() {
		c.widgetBase.setEnabled(c, enabled)
		c.invalidateAll()
	})
}

// SetItems 更新组合框的项目集合。
func (c *ComboBox) SetItems(items []ListItem) {
	c.runOnUI(func() {
		c.items = cloneItems(items)
		if len(c.items) == 0 {
			c.selected = -1
			c.hover = -1
			c.open = false
		} else if c.selected >= len(c.items) {
			c.selected = len(c.items) - 1
		}
		c.invalidateAll()
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
		if c.Placeholder == text {
			return
		}
		c.Placeholder = text
		c.invalidateAll()
	})
}

// SetStyle 更新组合框的样式覆盖。
func (c *ComboBox) SetStyle(style ComboStyle) {
	c.runOnUI(func() {
		c.Style = style
		c.invalidateAll()
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

// OnEvent 处理输入事件或生命周期事件。
func (c *ComboBox) OnEvent(evt Event) bool {
	switch evt.Type {
	case EventMouseMove:
		if c.open {
			index := c.popupIndexAt(evt.Point)
			if c.hover != index {
				c.hover = index
				c.invalidateAll()
			}
			return index >= 0
		}
	case EventMouseLeave:
		if c.hover != -1 {
			c.hover = -1
			c.invalidateAll()
		}
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
			c.open = !c.open && len(c.items) > 0
			if !c.open {
				c.hover = -1
			}
			c.invalidateAll()
			return true
		}
		if c.open {
			index := c.popupIndexAt(evt.Point)
			if index >= 0 {
				c.selectIndex(index, true)
				c.open = false
				c.hover = -1
				c.invalidateAll()
				return true
			}
		}
	case EventFocus:
		if !c.focused {
			c.focused = true
			c.invalidateAll()
		}
	case EventBlur:
		if c.focused || c.open || c.hover != -1 {
			c.focused = false
			c.open = false
			c.hover = -1
			c.invalidateAll()
		}
	case EventKeyDown:
		if c.handleKey(evt.Key) {
			return true
		}
	}
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (c *ComboBox) Paint(ctx *PaintCtx) {
	if !c.Visible() || ctx == nil {
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
	_ = ctx.DrawText(text, textRect, TextStyle{
		Font:   style.Font,
		Color:  textColor,
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})
	arrow := "v"
	if c.open {
		arrow = "^"
	}
	_ = ctx.DrawText(arrow, arrowRect, TextStyle{
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
	if !c.Visible() || !c.open || ctx == nil {
		return
	}

	style := c.resolveStyle(ctx)
	popup := c.popupRect()
	if popup.Empty() {
		return
	}

	radius := ctx.DP(style.CornerRadius)
	_ = ctx.FillRoundRect(popup, radius, style.PopupBackground)
	_ = ctx.StrokeRoundRect(popup, radius, style.FocusBorder, 1)

	start, end := c.popupRange()
	for index := start; index < end; index++ {
		item := c.items[index]
		rowRect := c.popupRowRect(index, ctx, style)
		textColor := style.TextColor
		if item.Disabled {
			textColor = style.PlaceholderColor
		}
		if index == c.selected {
			_ = ctx.FillRoundRect(rowRect, max32(1, radius-ctx.DP(2)), style.ItemSelectedColor)
			textColor = style.ItemTextColor
		} else if index == c.hover {
			_ = ctx.FillRoundRect(rowRect, max32(1, radius-ctx.DP(2)), style.ItemHoverColor)
		}

		textRect := Rect{
			X: rowRect.X + ctx.DP(10),
			Y: rowRect.Y,
			W: max32(0, rowRect.W-ctx.DP(20)),
			H: rowRect.H,
		}
		_ = ctx.DrawText(item.displayText(), textRect, TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
	}
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (c *ComboBox) acceptsFocus() bool {
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (c *ComboBox) cursor() CursorID {
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
		c.open = !c.open
		if !c.open {
			c.hover = -1
		}
		c.invalidateAll()
		return true
	case core.KeyEscape:
		if c.open {
			c.open = false
			c.hover = -1
			c.invalidateAll()
			return true
		}
	case core.KeyDown:
		if len(c.items) == 0 {
			return true
		}
		c.open = true
		c.selectRelative(1)
		c.hover = c.selected
		c.invalidateAll()
		return true
	case core.KeyUp:
		if len(c.items) == 0 {
			return true
		}
		c.open = true
		c.selectRelative(-1)
		c.hover = c.selected
		c.invalidateAll()
		return true
	case core.KeyHome:
		if len(c.items) == 0 {
			return true
		}
		c.selectIndex(0, true)
		c.hover = c.selected
		c.invalidateAll()
		return true
	case core.KeyEnd:
		if len(c.items) == 0 {
			return true
		}
		c.selectIndex(len(c.items)-1, true)
		c.hover = c.selected
		c.invalidateAll()
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
	c.selected = index
	c.invalidateAll()
	if notify && c.OnChange != nil {
		c.OnChange(index, c.items[index])
	}
}

// popupRect 返回组合框弹出层的边界。
func (c *ComboBox) popupRect() Rect {
	if !c.open || len(c.items) == 0 {
		return Rect{}
	}
	style := mergeComboStyle(DefaultTheme().ComboBox, c.Style)
	itemHeight := c.dp(style.ItemHeightDP)
	padding := c.dp(style.PaddingDP)
	start, end := c.popupRange()
	visibleCount := end - start
	return Rect{
		X: c.bounds.X,
		Y: c.bounds.Y + c.bounds.H + c.dp(6),
		W: c.bounds.W,
		H: padding*2 + int32(visibleCount)*itemHeight,
	}
}

// popupRange 返回组合框弹出层的可见项范围。
func (c *ComboBox) popupRange() (int, int) {
	if len(c.items) == 0 {
		return 0, 0
	}
	style := mergeComboStyle(DefaultTheme().ComboBox, c.Style)
	visible := int(style.MaxVisibleItems)
	if visible <= 0 || visible > len(c.items) {
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

// popupIndexAt 返回弹出层指定位置对应的项索引。
func (c *ComboBox) popupIndexAt(point core.Point) int {
	popup := c.popupRect()
	if !popup.Contains(point.X, point.Y) {
		return -1
	}
	style := mergeComboStyle(DefaultTheme().ComboBox, c.Style)
	itemHeight := max32(1, c.dp(style.ItemHeightDP))
	padding := max32(0, c.dp(style.PaddingDP))
	start, end := c.popupRange()
	index := int((point.Y - popup.Y - padding) / itemHeight)
	if point.Y < popup.Y+padding || start+index >= end || index < 0 {
		return -1
	}
	return start + index
}

// popupRowRect 返回组合框弹出层某一行的绘制矩形。
func (c *ComboBox) popupRowRect(index int, ctx *PaintCtx, style ComboStyle) Rect {
	popup := c.popupRect()
	start, _ := c.popupRange()
	padding := ctx.DP(style.PaddingDP)
	itemHeight := ctx.DP(style.ItemHeightDP)
	offset := int32(index - start)
	return Rect{
		X: popup.X + padding,
		Y: popup.Y + padding + offset*itemHeight,
		W: max32(0, popup.W-padding*2),
		H: itemHeight,
	}
}

// invalidateAll 使整个场景失效，以刷新弹出层或覆盖层状态。
func (c *ComboBox) invalidateAll() {
	if scene := c.scene(); scene != nil {
		scene.Invalidate(nil)
	}
}

// dp 按应用当前 DPI 缩放设备无关值。
func (c *ComboBox) dp(value int32) int32 {
	if scene := c.scene(); scene != nil && scene.app != nil {
		return scene.app.DP(value)
	}
	return value
}

// mergeComboStyle 将组合框样式覆盖合并到基础样式上。
func mergeComboStyle(base, override ComboStyle) ComboStyle {
	if override.Font.Face != "" {
		base.Font = override.Font
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
	if override.ArrowColor != 0 {
		base.ArrowColor = override.ArrowColor
	}
	if override.PopupBackground != 0 {
		base.PopupBackground = override.PopupBackground
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
	if override.ItemHeightDP != 0 {
		base.ItemHeightDP = override.ItemHeightDP
	}
	if override.PaddingDP != 0 {
		base.PaddingDP = override.PaddingDP
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	if override.MaxVisibleItems != 0 {
		base.MaxVisibleItems = override.MaxVisibleItems
	}
	return base
}

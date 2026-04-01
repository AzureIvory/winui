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
		_ = ctx.DrawText(item.displayText(), textRect, TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})
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
	if isNativeMode(c.mode) {
		return false
	}
	return true
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

// popupRange 返回组合框弹出层的可见项范围。
func (c *ComboBox) popupRange() (int, int) {
	layout := c.popupLayout()
	return layout.start, layout.end
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

// popupRowRect 返回组合框弹出层某一行的绘制矩形。
func (c *ComboBox) popupRowRect(index int, ctx *PaintCtx, style ComboStyle) Rect {
	return c.popupRowRectForLayout(index, c.popupLayout())
}

// invalidateAll 使整个场景失效，以刷新弹出层或覆盖层状态。
func (c *ComboBox) invalidateAll() {
	c.invalidateStateChange(widgetDirtyRect(c))
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
	layout.itemHeight = max32(1, c.dp(style.ItemHeightDP))
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
	return unionRect(c.Bounds(), c.popupRect())
}

func mergeComboStyle(base, override ComboStyle) ComboStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
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

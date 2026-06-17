//go:build windows

package widgets

import (
	"time"

	"github.com/AzureIvory/winui/core"
)

// ListBox 表示单选列表控件。
// ListBox 表示单选列表控件。
type ListBox struct {
	// widgetBase 提供列表框共享的基础控件能力。
	widgetBase
	// items 保存列表项集合。
	items []ListItem
	// selected 保存当前选中索引。
	selected int
	// hover 保存当前悬停索引。
	hover int
	// pressed 保存当前按下索引。
	pressed int
	// scroll 保存顶部可见行索引。
	scroll int
	// focused 记录控件是否拥有焦点。
	focused bool
	// lastClickIndex 保存上次点击的索引，用于双击检测。
	lastClickIndex int
	// lastClickAt 保存上次点击时间。
	lastClickAt time.Time
	// Style 保存样式覆盖。
	Style ListStyle
	// OnChange 保存选择变更回调。
	OnChange func(int, ListItem)
	// OnActivate 保存项激活回调。
	OnActivate func(int, ListItem)
	// OnRightClick 保存右键回调。
	OnRightClick func(int, ListItem, core.Point)
	// showCheck 控制是否在每行右侧渲染打勾列。
	showCheck bool
	// OnCheckChange 保存打勾列勾选状态变更回调。
	OnCheckChange func(int, ListItem)
}

// NewListBox 创建一个新的列表框。
func NewListBox(id string) *ListBox {
	return &ListBox{
		widgetBase:     newWidgetBase(id, "listbox"),
		selected:       -1,
		hover:          -1,
		pressed:        -1,
		lastClickIndex: -1,
	}
}

// SetBounds 更新列表框的边界。
func (l *ListBox) SetBounds(rect Rect) {
	l.runOnUI(func() {
		l.widgetBase.setBounds(l, rect)
		l.clampScroll()
	})
}

// SetVisible 更新列表框的可见状态。
func (l *ListBox) SetVisible(visible bool) {
	l.runOnUI(func() {
		l.widgetBase.setVisible(l, visible)
	})
}

// SetEnabled 更新列表框的可用状态。
func (l *ListBox) SetEnabled(enabled bool) {
	l.runOnUI(func() {
		l.widgetBase.setEnabled(l, enabled)
	})
}

// SetItems 更新列表框的项目集合。
func (l *ListBox) SetItems(items []ListItem) {
	l.runOnUI(func() {
		l.items = cloneItems(items)
		if len(l.items) == 0 {
			l.selected = -1
			l.hover = -1
			l.pressed = -1
			l.scroll = 0
			l.lastClickIndex = -1
			l.lastClickAt = time.Time{}
		} else if l.selected >= len(l.items) {
			l.selected = len(l.items) - 1
		}
		l.clampScroll()
		l.invalidate(l)
	})
}

// Items 返回列表框所管理项目的副本。
func (l *ListBox) Items() []ListItem {
	return cloneItems(l.items)
}

// SetSelected 更新列表框的当前选择。
func (l *ListBox) SetSelected(index int) {
	l.runOnUI(func() {
		if index < 0 {
			if l.selected != -1 {
				l.selected = -1
				l.invalidate(l)
			}
			return
		}
		l.selectIndex(index, false)
	})
}

// ClearSelection 清除当前选择。
func (l *ListBox) ClearSelection() {
	l.SetSelected(-1)
}

// SelectedIndex 返回列表框当前选中的索引。
func (l *ListBox) SelectedIndex() int {
	return l.selected
}

// SelectedItem 返回列表框当前选中的项目。
func (l *ListBox) SelectedItem() (ListItem, bool) {
	if l.selected < 0 || l.selected >= len(l.items) {
		return ListItem{}, false
	}
	return l.items[l.selected], true
}

// SetStyle 更新列表框的样式覆盖。
func (l *ListBox) SetStyle(style ListStyle) {
	l.runOnUI(func() {
		l.Style = style
		l.invalidate(l)
	})
}

// SetOnChange 注册列表框的变更回调。
func (l *ListBox) SetOnChange(fn func(int, ListItem)) {
	l.runOnUI(func() {
		l.OnChange = fn
	})
}

// SetOnActivate 注册列表项激活回调，例如双击或按下 Enter。
func (l *ListBox) SetOnActivate(fn func(int, ListItem)) {
	l.runOnUI(func() {
		l.OnActivate = fn
	})
}

// SetOnRightClick 注册列表项右键回调。
func (l *ListBox) SetOnRightClick(fn func(int, ListItem, core.Point)) {
	l.runOnUI(func() {
		l.OnRightClick = fn
	})
}

// SetShowCheck 控制是否在每行右侧渲染打勾列。
func (l *ListBox) SetShowCheck(show bool) {
	l.runOnUI(func() {
		if l.showCheck == show {
			return
		}
		l.showCheck = show
		l.invalidate(l)
	})
}

// ShowCheck 返回是否已开启打勾列。
func (l *ListBox) ShowCheck() bool {
	return l.showCheck
}

// SetOnCheckChange 注册打勾列勾选状态变更回调。
func (l *ListBox) SetOnCheckChange(fn func(int, ListItem)) {
	l.runOnUI(func() {
		l.OnCheckChange = fn
	})
}

// SetItemChecked 更新指定项的勾选状态。
func (l *ListBox) SetItemChecked(index int, checked bool) {
	l.runOnUI(func() {
		if index < 0 || index >= len(l.items) {
			return
		}
		if l.items[index].Checked == checked {
			return
		}
		l.items[index].Checked = checked
		l.invalidate(l)
	})
}

// IsItemChecked 返回指定项是否处于勾选状态。
func (l *ListBox) IsItemChecked(index int) bool {
	if index < 0 || index >= len(l.items) {
		return false
	}
	return l.items[index].Checked
}

// OnEvent 处理输入事件或生命周期事件。
func (l *ListBox) OnEvent(evt Event) bool {
	switch evt.Type {
	case EventMouseMove:
		index := l.indexAt(evt.Point)
		if l.hover != index {
			l.hover = index
			l.invalidate(l)
		}
	case EventMouseLeave:
		if l.hover != -1 {
			l.hover = -1
			l.invalidate(l)
		}
	case EventMouseDown:
		if l.Enabled() {
			index := l.indexAt(evt.Point)
			switch evt.Button {
			case core.MouseButtonLeft:
				l.pressed = index
				l.invalidate(l)
				return true
			case core.MouseButtonRight:
				if index >= 0 {
					l.selectIndex(index, true)
					if l.OnRightClick != nil {
						l.OnRightClick(index, l.items[index], evt.Point)
					}
					return true
				}
			}
		}
	case EventMouseUp:
		if l.pressed != -1 {
			l.pressed = -1
			l.invalidate(l)
			return true
		}
	case EventClick:
		if !l.Enabled() {
			return false
		}
		index := l.indexAt(evt.Point)
		if index < 0 {
			return false
		}
		// 点击落在打勾列时切换勾选，与单选选中相互独立。
		if l.showCheck && l.pointInCheckArea(evt.Point, index) {
			l.toggleItemCheck(index, true)
			return true
		}
		now := time.Now()
		activate := index == l.lastClickIndex && now.Sub(l.lastClickAt) <= 450*time.Millisecond
		l.lastClickIndex = index
		l.lastClickAt = now
		l.selectIndex(index, true)
		if activate && l.OnActivate != nil {
			l.OnActivate(index, l.items[index])
		}
		return true
	case EventMouseWheel:
		if !l.Enabled() {
			return false
		}
		if l.scrollBy(wheelSteps(evt.Delta) * -3) {
			return true
		}
	case EventFocus:
		if !l.focused {
			l.focused = true
			l.invalidate(l)
		}
	case EventBlur:
		if l.focused || l.hover != -1 {
			l.focused = false
			l.hover = -1
			l.pressed = -1
			l.invalidate(l)
		}
	case EventKeyDown:
		if l.handleKey(evt.Key) {
			return true
		}
	}
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (l *ListBox) Paint(ctx *PaintCtx) {
	if !l.Visible() || ctx == nil {
		return
	}

	style := l.resolveStyle(ctx)
	bounds := l.Bounds()
	if bounds.Empty() {
		return
	}

	borderColor := style.BorderColor
	if l.focused {
		borderColor = style.FocusBorder
	} else if l.hover >= 0 {
		borderColor = style.HoverBorder
	}

	radius := ctx.DP(style.CornerRadius)
	_ = ctx.FillRoundRect(bounds, radius, style.Background)
	_ = ctx.StrokeRoundRect(bounds, radius, borderColor, 1)
	itemRadius := radius - ctx.DP(2)
	if itemRadius < 0 {
		itemRadius = 0
	}
	scrollRadius := ctx.DP(2)
	if radius <= 0 {
		scrollRadius = 0
	}

	start := l.scroll
	end := len(l.items)
	if visible := l.visibleRows(style); visible > 0 && start+visible+1 < end {
		end = start + visible + 1
	}

	// 打勾列开启时，在每行右侧预留指示器宽度。
	leftPad := ctx.DP(10)
	rightPad := leftPad
	checkSize := int32(0)
	if l.showCheck {
		checkSize = ctx.DP(style.CheckSizeDP)
		if checkSize <= 0 {
			checkSize = ctx.DP(18)
		}
		rightPad += checkSize + ctx.DP(10)
	}

	// 行绘制裁剪到列表框内部区域，避免最后一行（部分可见）的圆角背景
	// 溢出到列表框底部边界之外。
	padding := max32(0, ctx.DP(style.PaddingDP))
	contentRect := Rect{
		X: bounds.X + padding,
		Y: bounds.Y + padding,
		W: max32(0, bounds.W-padding*2),
		H: max32(0, bounds.H-padding*2),
	}
	restore := ctx.PushClipRect(contentRect)
	defer restore()

	for index := start; index < end; index++ {
		item := l.items[index]
		rowRect := l.rowRect(index, ctx, style)
		if rowRect.Y >= bounds.Y+bounds.H {
			break
		}
		if rowRect.Y+rowRect.H <= bounds.Y {
			continue
		}

		// 行底色：优先使用项自定义底色，否则透出控件背景（不填）。
		rowBg := item.Background
		baseBg := rowBg
		if baseBg == 0 {
			baseBg = style.Background
		}
		// 选中/悬停态在行底色上按强度混合，保证自定义底色依旧可见。
		switch {
		case index == l.selected:
			rowBg = BlendColor(baseBg, style.ItemSelectedColor, listSelectionAlpha)
		case index == l.hover:
			rowBg = BlendColor(baseBg, style.ItemHoverColor, listHoverAlpha)
		}
		if rowBg != 0 {
			_ = ctx.FillRoundRect(rowRect, itemRadius, rowBg)
		}

		// 文字色：项自定义最优先，其次禁用、选中、默认。
		textColor := style.TextColor
		switch {
		case item.TextColor != 0:
			textColor = item.TextColor
		case item.Disabled:
			textColor = style.DisabledText
		case index == l.selected:
			textColor = style.ItemTextColor
		}

		textRect := Rect{
			X: rowRect.X + leftPad,
			Y: rowRect.Y,
			W: max32(0, rowRect.W-leftPad-rightPad),
			H: rowRect.H,
		}
		_ = ctx.DrawWidgetText(l, item.displayText(), textRect, TextStyle{
			Font:   style.Font,
			Color:  textColor,
			Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
		})

		// 打勾指示器画在行右侧，垂直居中。
		if checkSize > 0 {
			boxRect := Rect{
				X: rowRect.X + rowRect.W - leftPad - checkSize,
				Y: rowRect.Y + (rowRect.H-checkSize)/2,
				W: checkSize,
				H: checkSize,
			}
			l.paintRowCheck(ctx, boxRect, item.Checked, style)
		}
	}

	if maxScroll := l.maxScroll(style); maxScroll > 0 {
		track := Rect{
			X: bounds.X + bounds.W - ctx.DP(8),
			Y: bounds.Y + ctx.DP(8),
			W: ctx.DP(4),
			H: max32(0, bounds.H-ctx.DP(16)),
		}
		if track.H > 0 {
			thumbH := max32(ctx.DP(24), track.H*int32(l.visibleRows(style))/int32(len(l.items)))
			thumbRange := max32(1, track.H-thumbH)
			thumbY := track.Y
			if maxScroll > 0 {
				thumbY += thumbRange * int32(l.scroll) / int32(maxScroll)
			}
			_ = ctx.FillRoundRect(track, scrollRadius, core.RGB(241, 245, 249))
			_ = ctx.FillRoundRect(
				Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH},
				scrollRadius,
				core.RGB(148, 163, 184),
			)
		}
	}
}

// acceptsFocus 返回控件是否可接收键盘焦点。
func (l *ListBox) acceptsFocus() bool {
	return true
}

// cursor 返回悬停控件时应使用的光标。
func (l *ListBox) cursor() CursorID {
	if !l.Enabled() {
		return core.CursorArrow
	}
	return core.CursorHand
}

// resolveStyle 解析列表框的最终样式。
func (l *ListBox) resolveStyle(ctx *PaintCtx) ListStyle {
	style := DefaultTheme().ListBox
	if ctx != nil && ctx.scene != nil && ctx.scene.theme != nil {
		style = ctx.scene.theme.ListBox
	} else if scene := l.scene(); scene != nil && scene.theme != nil {
		style = scene.theme.ListBox
	}
	return mergeListStyle(style, l.Style)
}

// handleKey 处理列表框的按键事件。
func (l *ListBox) handleKey(key core.KeyEvent) bool {
	if len(l.items) == 0 || !l.Enabled() {
		return false
	}

	switch key.Key {
	case core.KeyUp:
		l.selectRelative(-1)
		return true
	case core.KeyDown:
		l.selectRelative(1)
		return true
	case core.KeyHome:
		l.selectIndex(0, true)
		return true
	case core.KeyEnd:
		l.selectIndex(len(l.items)-1, true)
		return true
	case core.KeyReturn, core.KeySpace:
		if l.selected >= 0 && l.selected < len(l.items) && l.OnChange != nil {
			l.OnChange(l.selected, l.items[l.selected])
		}
		if key.Key == core.KeyReturn && l.selected >= 0 && l.selected < len(l.items) && l.OnActivate != nil {
			l.OnActivate(l.selected, l.items[l.selected])
		}
		return true
	}
	return false
}

// selectRelative 按给定步长移动当前选择，并跳过禁用项。
func (l *ListBox) selectRelative(step int) {
	index := l.selected
	if index < 0 {
		if step >= 0 {
			index = -1
		} else {
			index = len(l.items)
		}
	}
	for {
		index += step
		if index < 0 || index >= len(l.items) {
			return
		}
		if !l.items[index].Disabled {
			l.selectIndex(index, true)
			return
		}
	}
}

// selectIndex 将当前选择设置为指定项索引。
func (l *ListBox) selectIndex(index int, notify bool) {
	if index < 0 || index >= len(l.items) || l.items[index].Disabled {
		return
	}
	if l.selected == index {
		return
	}
	l.selected = index
	l.ensureVisible(index)
	l.invalidate(l)
	if notify && l.OnChange != nil {
		l.OnChange(index, l.items[index])
	}
}

// indexAt 返回指定位置对应的项索引。
func (l *ListBox) indexAt(point core.Point) int {
	rect := l.Bounds()
	if !rect.Contains(point.X, point.Y) {
		return -1
	}
	style := l.resolveStyle(nil)
	itemHeight := max32(1, l.dp(style.ItemHeightDP))
	padding := max32(0, l.dp(style.PaddingDP))
	index := int((point.Y-rect.Y-padding)/itemHeight) + l.scroll
	if point.Y < rect.Y+padding || index < 0 || index >= len(l.items) {
		return -1
	}
	return index
}

// rowRect 返回列表行的绘制矩形。
func (l *ListBox) rowRect(index int, ctx *PaintCtx, style ListStyle) Rect {
	padding := ctx.DP(style.PaddingDP)
	itemHeight := ctx.DP(style.ItemHeightDP)
	return Rect{
		X: l.bounds.X + padding,
		Y: l.bounds.Y + padding + int32(index-l.scroll)*itemHeight,
		W: max32(0, l.bounds.W-padding*2),
		H: itemHeight,
	}
}

// dp 按应用当前 DPI 缩放设备无关值。
func (l *ListBox) dp(value int32) int32 {
	if scene := l.scene(); scene != nil && scene.app != nil {
		return scene.app.DP(value)
	}
	return value
}

// mergeListStyle 将列表框样式覆盖合并到基础样式上。
func mergeListStyle(base, override ListStyle) ListStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.TextColor != 0 {
		base.TextColor = override.TextColor
	}
	if override.DisabledText != 0 {
		base.DisabledText = override.DisabledText
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
	if override.CheckColor != 0 {
		base.CheckColor = override.CheckColor
	}
	if override.CheckMarkColor != 0 {
		base.CheckMarkColor = override.CheckMarkColor
	}
	if override.CheckBorder != 0 {
		base.CheckBorder = override.CheckBorder
	}
	if override.CheckSizeDP != 0 {
		base.CheckSizeDP = override.CheckSizeDP
	}
	return base
}

// visibleRows 计算当前样式下可见的行数。
func (l *ListBox) visibleRows(style ListStyle) int {
	padding := max32(0, l.dp(style.PaddingDP))
	itemHeight := max32(1, l.dp(style.ItemHeightDP))
	height := l.bounds.H - padding*2
	if height <= 0 {
		return 0
	}
	rows := int(height / itemHeight)
	if rows <= 0 {
		rows = 1
	}
	return rows
}

// maxScroll 计算当前列表允许的最大滚动偏移。
func (l *ListBox) maxScroll(style ListStyle) int {
	rows := l.visibleRows(style)
	if rows <= 0 || len(l.items) <= rows {
		return 0
	}
	return len(l.items) - rows
}

// clampScroll 把滚动位置限制在合法范围内。
func (l *ListBox) clampScroll() {
	style := l.resolveStyle(nil)
	maxScroll := l.maxScroll(style)
	if l.scroll < 0 {
		l.scroll = 0
	}
	if l.scroll > maxScroll {
		l.scroll = maxScroll
	}
}

// scrollBy 按给定偏移滚动列表。
func (l *ListBox) scrollBy(delta int) bool {
	if delta == 0 {
		return false
	}
	old := l.scroll
	l.scroll += delta
	l.clampScroll()
	if l.scroll == old {
		return false
	}
	l.invalidate(l)
	return true
}

// ensureVisible 调整滚动位置以确保指定项可见。
func (l *ListBox) ensureVisible(index int) {
	if index < 0 || index >= len(l.items) {
		return
	}
	style := l.resolveStyle(nil)
	rows := l.visibleRows(style)
	if rows <= 0 {
		return
	}
	if index < l.scroll {
		l.scroll = index
		return
	}
	last := l.scroll + rows - 1
	if index > last {
		l.scroll = index - rows + 1
	}
	l.clampScroll()
}

// wheelSteps 把鼠标滚轮增量转换为滚动步数。
func wheelSteps(delta int32) int {
	if delta == 0 {
		return 0
	}
	steps := int(delta / 120)
	if steps == 0 {
		if delta > 0 {
			return 1
		}
		return -1
	}
	return steps
}

// listSelectionAlpha 控制选中行高亮叠加强度（0-255），值越小自定义行底色越明显。
const listSelectionAlpha byte = 180

// listHoverAlpha 控制悬停行高亮叠加强度（0-255）。
const listHoverAlpha byte = 85

// paintRowCheck 在指定矩形内绘制打勾列指示器。
func (l *ListBox) paintRowCheck(ctx *PaintCtx, box Rect, checked bool, style ListStyle) {
	radius := ctx.DP(4)
	if checked {
		_ = ctx.FillRoundRect(box, radius, style.CheckColor)
		// 复用复选框的打勾几何，保证视觉风格统一。
		drawChoiceCheck(ctx, l, box, style.CheckMarkColor)
		return
	}
	_ = ctx.StrokeRoundRect(box, radius, style.CheckBorder, 1)
}

// pointInCheckArea 判断指定坐标是否落在给定行的打勾列区域内。
func (l *ListBox) pointInCheckArea(point core.Point, index int) bool {
	if index < 0 || index >= len(l.items) {
		return false
	}
	style := l.resolveStyle(nil)
	checkSize := l.dp(style.CheckSizeDP)
	if checkSize <= 0 {
		checkSize = l.dp(18)
	}
	padding := max32(0, l.dp(style.PaddingDP))
	itemHeight := max32(1, l.dp(style.ItemHeightDP))
	rowW := max32(0, l.bounds.W-padding*2)
	rowY := l.bounds.Y + padding + int32(index-l.scroll)*itemHeight
	// 与 Paint 中保持一致：打勾区距行右边缘 leftPad（10dp）。
	checkLeft := l.bounds.X + padding + rowW - l.dp(10) - checkSize
	return point.X >= checkLeft && point.X <= checkLeft+checkSize &&
		point.Y >= rowY && point.Y <= rowY+itemHeight
}

// toggleItemCheck 翻转指定项的勾选状态，并按需触发回调。
func (l *ListBox) toggleItemCheck(index int, notify bool) {
	if index < 0 || index >= len(l.items) {
		return
	}
	l.items[index].Checked = !l.items[index].Checked
	l.invalidate(l)
	if notify && l.OnCheckChange != nil {
		l.OnCheckChange(index, l.items[index])
	}
}

//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// MeasureNatural returns the natural size a widget would like to occupy.
func MeasureNatural(widget Widget) core.Size {
	return measureWidgetNatural(widget)
}

func measureWidgetNatural(widget Widget) core.Size {
	if widget == nil {
		return core.Size{}
	}
	switch typed := widget.(type) {
	case *Panel:
		return measurePanelNatural(typed)
	case *ScrollView:
		return measureScrollViewNatural(typed)
	default:
		return preferredSizeOf(widget)
	}
}

func measureScrollViewNatural(scroll *ScrollView) core.Size {
	if scroll == nil {
		return core.Size{}
	}
	size := preferredSizeOf(scroll)
	if size.Width > 0 || size.Height > 0 {
		return size
	}
	if scroll.content != nil {
		return measureWidgetNatural(scroll.content)
	}
	bounds := scroll.Bounds()
	return core.Size{Width: bounds.W, Height: bounds.H}
}

func measurePanelNatural(panel *Panel) core.Size {
	if panel == nil {
		return core.Size{}
	}
	if len(panel.children) == 0 {
		return preferredSizeOf(panel)
	}
	switch layout := panel.layout.(type) {
	case RowLayout:
		return measureFlexNatural(panel.children, AxisHorizontal, layout.Gap, layout.Padding, layout.ItemSize)
	case ColumnLayout:
		return measureFlexNatural(panel.children, AxisVertical, layout.Gap, layout.Padding, layout.ItemSize)
	case GridLayout:
		return measureGridNatural(panel.children, layout)
	case FormLayout:
		return measureFormNatural(panel.children, layout)
	case LinearLayout:
		padding := UniformInsets(layout.Padding)
		axis := layout.Axis
		if axis == 0 {
			axis = AxisHorizontal
		}
		return measureFlexNatural(panel.children, axis, layout.Gap, padding, layout.ItemSize)
	default:
		return measureAbsoluteNatural(panel.children)
	}
}

func measureFlexNatural(children []Widget, axis Axis, gap int32, padding Insets, itemSize int32) core.Size {
	metricWidget := layoutMetricWidget(children)
	gap = widgetDP(metricWidget, gap)
	padding = scaleInsetsForWidget(metricWidget, padding)
	itemSize = widgetDP(metricWidget, itemSize)
	main := int32(0)
	cross := int32(0)
	count := 0
	for _, child := range children {
		if child == nil {
			continue
		}
		size := preferredSizeOf(child)
		childMain := size.Width
		childCross := size.Height
		if axis == AxisVertical {
			childMain = size.Height
			childCross = size.Width
		}
		if itemSize > 0 {
			childMain = itemSize
		}
		if childMain < 0 {
			childMain = 0
		}
		if childCross < 0 {
			childCross = 0
		}
		main += childMain
		if childCross > cross {
			cross = childCross
		}
		count++
	}
	if count > 1 {
		main += gap * int32(count-1)
	}
	if axis == AxisHorizontal {
		return core.Size{Width: main + padding.horizontal(), Height: cross + padding.vertical()}
	}
	return core.Size{Width: cross + padding.horizontal(), Height: main + padding.vertical()}
}

func measureGridNatural(children []Widget, layout GridLayout) core.Size {
	columns := layout.Columns
	if columns <= 0 {
		columns = 1
	}
	metricWidget := layoutMetricWidget(children)
	padding := scaleInsetsForWidget(metricWidget, layout.Padding)
	columnGap := widgetDP(metricWidget, layout.ColumnGap)
	if columnGap == 0 {
		columnGap = widgetDP(metricWidget, layout.Gap)
	}
	rowGap := widgetDP(metricWidget, layout.RowGap)
	if rowGap == 0 {
		rowGap = widgetDP(metricWidget, layout.Gap)
	}
	colWidths := make([]int32, columns)
	rowHeights := make([]int32, 0)
	for index, child := range children {
		if child == nil {
			continue
		}
		size := preferredSizeOf(child)
		row := index / columns
		col := index % columns
		for len(rowHeights) <= row {
			rowHeights = append(rowHeights, 0)
		}
		if size.Width > colWidths[col] {
			colWidths[col] = size.Width
		}
		if size.Height > rowHeights[row] {
			rowHeights[row] = size.Height
		}
	}
	width := padding.horizontal()
	for _, colWidth := range colWidths {
		width += colWidth
	}
	height := padding.vertical()
	for _, rowHeight := range rowHeights {
		height += rowHeight
	}
	if len(colWidths) > 1 {
		width += columnGap * int32(len(colWidths)-1)
	}
	if len(rowHeights) > 1 {
		height += rowGap * int32(len(rowHeights)-1)
	}
	return core.Size{Width: width, Height: height}
}

func measureFormNatural(children []Widget, layout FormLayout) core.Size {
	metricWidget := layoutMetricWidget(children)
	padding := scaleInsetsForWidget(metricWidget, layout.Padding)
	labelWidth := widgetDP(metricWidget, layout.LabelWidth)
	rowGap := widgetDP(metricWidget, layout.RowGap)
	columnGap := widgetDP(metricWidget, layout.ColumnGap)
	fieldWidth := int32(0)
	height := padding.vertical()
	rows := 0
	for index, child := range children {
		if child == nil {
			continue
		}
		size := preferredSizeOf(child)
		if index%2 == 0 {
			if size.Width > labelWidth {
				labelWidth = size.Width
			}
			height += size.Height
		} else {
			if size.Width > fieldWidth {
				fieldWidth = size.Width
			}
			if size.Height > 0 {
				height = height - preferredSizeOf(children[index-1]).Height + max32(preferredSizeOf(children[index-1]).Height, size.Height)
			}
			rows++
		}
	}
	if rows > 1 {
		height += rowGap * int32(rows-1)
	}
	width := padding.horizontal() + labelWidth + fieldWidth
	if labelWidth > 0 && fieldWidth > 0 {
		width += columnGap
	}
	return core.Size{Width: width, Height: height}
}

func measureAbsoluteNatural(children []Widget) core.Size {
	width := int32(0)
	height := int32(0)
	for _, child := range children {
		if child == nil {
			continue
		}
		size := preferredSizeOf(child)
		bounds := child.Bounds()
		x := bounds.X
		y := bounds.Y
		w := bounds.W
		h := bounds.H
		if data, ok := absoluteDataOf(child.LayoutData()); ok {
			if data.HasLeft {
				x = widgetDP(child, data.Left)
			}
			if data.HasTop {
				y = widgetDP(child, data.Top)
			}
			if data.HasWidth {
				w = widgetDP(child, data.Width)
			}
			if data.HasHeight {
				h = widgetDP(child, data.Height)
			}
		}
		if w <= 0 {
			w = size.Width
		}
		if h <= 0 {
			h = size.Height
		}
		right := x + w
		bottom := y + h
		if right > width {
			width = right
		}
		if bottom > height {
			height = bottom
		}
	}
	return core.Size{Width: width, Height: height}
}

//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// Layout 描述容器对子控件进行排布的策略。
type Layout interface {
	// Apply 按父容器边界对子控件进行排布。
	Apply(parent Rect, children []Widget)
}

// AbsoluteLayout 表示绝对布局，不主动调整子控件边界。
type AbsoluteLayout struct{}

// AbsoluteLayoutData describes optional absolute positioning constraints.
type AbsoluteLayoutData struct {
	Left      int32
	Top       int32
	Right     int32
	Bottom    int32
	Width     int32
	Height    int32
	HasLeft   bool
	HasTop    bool
	HasRight  bool
	HasBottom bool
	HasWidth  bool
	HasHeight bool
}

// Apply resolves absolute constraints against the parent rect.
func (AbsoluteLayout) Apply(parent Rect, children []Widget) {
	for _, child := range children {
		if child == nil {
			continue
		}
		data, ok := absoluteDataOf(child.LayoutData())
		if !ok {
			continue
		}
		size := preferredSizeOf(child)
		width := size.Width
		height := size.Height
		if data.HasWidth {
			width = widgetDP(child, data.Width)
		} else if data.HasLeft && data.HasRight {
			width = parent.W - widgetDP(child, data.Left) - widgetDP(child, data.Right)
		}
		size = preferredSizeForWidth(child, width)
		if !data.HasWidth && width <= 0 {
			width = size.Width
		}
		if data.HasHeight {
			height = widgetDP(child, data.Height)
		} else if data.HasTop && data.HasBottom {
			height = parent.H - widgetDP(child, data.Top) - widgetDP(child, data.Bottom)
		} else if size.Height > 0 {
			height = size.Height
		}
		if width < 0 {
			width = 0
		}
		if height < 0 {
			height = 0
		}

		rect := child.Bounds()
		rect.W = width
		rect.H = height
		if data.HasLeft {
			rect.X = parent.X + widgetDP(child, data.Left)
		} else if data.HasRight {
			rect.X = parent.X + parent.W - widgetDP(child, data.Right) - rect.W
		}
		if data.HasTop {
			rect.Y = parent.Y + widgetDP(child, data.Top)
		} else if data.HasBottom {
			rect.Y = parent.Y + parent.H - widgetDP(child, data.Bottom) - rect.H
		}
		child.SetBounds(rect)
	}
}

// Axis 表示线性布局的主轴方向。
type Axis int

const (
	// AxisHorizontal 表示主轴沿水平方向展开。
	AxisHorizontal Axis = iota + 1
	// AxisVertical 表示主轴沿垂直方向展开。
	AxisVertical
)

// Alignment 表示控件在某个轴上的对齐方式。
type Alignment uint8

const (
	// AlignDefault 表示使用布局提供的默认对齐方式。
	AlignDefault Alignment = iota
	// AlignStart 表示靠起始边对齐。
	AlignStart
	// AlignCenter 表示居中对齐。
	AlignCenter
	// AlignEnd 表示靠结束边对齐。
	AlignEnd
	// AlignStretch 表示拉伸到可用空间大小。
	AlignStretch
)

// Insets 描述矩形四个方向的内边距。
type Insets struct {
	// Top 表示上内边距。
	Top int32
	// Right 表示右内边距。
	Right int32
	// Bottom 表示下内边距。
	Bottom int32
	// Left 表示左内边距。
	Left int32
}

// UniformInsets 返回四个方向相同的内边距配置。
func UniformInsets(value int32) Insets {
	return Insets{Top: value, Right: value, Bottom: value, Left: value}
}

// SymmetricInsets 返回水平和垂直分别对称的内边距配置。
func SymmetricInsets(horizontal, vertical int32) Insets {
	return Insets{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}

// horizontal 返回左右内边距之和。
func (i Insets) horizontal() int32 {
	return i.Left + i.Right
}

// vertical 返回上下内边距之和。
func (i Insets) vertical() int32 {
	return i.Top + i.Bottom
}

// inset 返回扣除内边距后的内容区域。
func (i Insets) inset(rect Rect) Rect {
	rect.X += i.Left
	rect.Y += i.Top
	rect.W -= i.horizontal()
	rect.H -= i.vertical()
	if rect.W < 0 {
		rect.W = 0
	}
	if rect.H < 0 {
		rect.H = 0
	}
	return rect
}

// FlexLayoutData 描述行布局和列布局中单个子控件的附加参数。
type FlexLayoutData struct {
	// Grow 表示主轴剩余空间分配时的权重。
	Grow int32
	// Align 表示交叉轴上的对齐方式。
	Align Alignment
}

// GridLayoutData 描述网格布局中单个子控件的附加参数。
type GridLayoutData struct {
	// ColumnSpan 表示控件横向跨越的列数。
	ColumnSpan int
	// RowSpan 表示控件纵向跨越的行数。
	RowSpan int
	// HorizontalAlign 表示控件在单元格中的水平对齐方式。
	HorizontalAlign Alignment
	// VerticalAlign 表示控件在单元格中的垂直对齐方式。
	VerticalAlign Alignment
}

// FormLayoutData 描述表单布局中字段控件的附加参数。
type FormLayoutData struct {
	// Grow 表示字段是否沿水平方向占满可用空间。
	Grow int32
	// Align 表示控件在行高内的垂直对齐方式。
	Align Alignment
}

// RowLayout 按从左到右的顺序排布子控件。
type RowLayout struct {
	// Gap 表示相邻子控件之间的水平间距。
	Gap int32
	// Padding 表示容器内容区域四周的内边距。
	Padding Insets
	// ItemSize 表示所有子控件在主轴上的固定尺寸，零值表示使用首选尺寸。
	ItemSize int32
	// CrossAlign 表示所有子控件在交叉轴上的默认对齐方式。
	CrossAlign Alignment
}

// ColumnLayout 按从上到下的顺序排布子控件。
type ColumnLayout struct {
	// Gap 表示相邻子控件之间的垂直间距。
	Gap int32
	// Padding 表示容器内容区域四周的内边距。
	Padding Insets
	// ItemSize 表示所有子控件在主轴上的固定尺寸，零值表示使用首选尺寸。
	ItemSize int32
	// CrossAlign 表示所有子控件在交叉轴上的默认对齐方式。
	CrossAlign Alignment
}

// GridLayout 按网格方式排布子控件。
type GridLayout struct {
	// Columns 表示网格总列数。
	Columns int
	// Gap 表示行列都未单独指定时的默认间距。
	Gap int32
	// RowGap 表示行与行之间的间距。
	RowGap int32
	// ColumnGap 表示列与列之间的间距。
	ColumnGap int32
	// Padding 表示容器内容区域四周的内边距。
	Padding Insets
	// ColumnWeights 表示额外宽度在各列之间的分配权重。
	ColumnWeights []int32
	// RowWeights 表示额外高度在各行之间的分配权重。
	RowWeights []int32
	// HorizontalAlign 表示单元格内默认的水平对齐方式。
	HorizontalAlign Alignment
	// VerticalAlign 表示单元格内默认的垂直对齐方式。
	VerticalAlign Alignment
}

// FormLayout 按标签列和字段列的形式排布子控件。
type FormLayout struct {
	// Padding 表示表单内容区域四周的内边距。
	Padding Insets
	// RowGap 表示表单行与行之间的间距。
	RowGap int32
	// ColumnGap 表示标签列和字段列之间的间距。
	ColumnGap int32
	// LabelWidth 表示标签列宽度，零值表示按首选宽度自动推导。
	LabelWidth int32
	// CrossAlign 表示控件在行高内的默认垂直对齐方式。
	CrossAlign Alignment
}

// LinearLayout 兼容原有线性布局接口，并委托给行布局或列布局。
type LinearLayout struct {
	// Axis 表示主轴方向。
	Axis Axis
	// Gap 表示子控件之间的间距。
	Gap int32
	// Padding 表示四周统一的内边距。
	Padding int32
	// ItemSize 表示主轴上的固定尺寸。
	ItemSize int32
}

// Apply 按行布局规则排布子控件。
func (l RowLayout) Apply(parent Rect, children []Widget) {
	applyFlexLayout(parent, children, flexOptions{
		axis:       AxisHorizontal,
		gap:        l.Gap,
		padding:    l.Padding,
		itemSize:   l.ItemSize,
		crossAlign: l.CrossAlign,
	})
}

// Apply 按列布局规则排布子控件。
func (l ColumnLayout) Apply(parent Rect, children []Widget) {
	applyFlexLayout(parent, children, flexOptions{
		axis:       AxisVertical,
		gap:        l.Gap,
		padding:    l.Padding,
		itemSize:   l.ItemSize,
		crossAlign: l.CrossAlign,
	})
}

// Apply 按网格布局规则排布子控件。
func (l GridLayout) Apply(parent Rect, children []Widget) {
	columns := l.Columns
	if columns <= 0 {
		columns = 1
	}

	metricWidget := layoutMetricWidget(children)
	padding := scaleInsetsForWidget(metricWidget, l.Padding)
	content := padding.inset(parent)
	if content.Empty() {
		return
	}

	columnGap := widgetDP(metricWidget, l.ColumnGap)
	if columnGap == 0 {
		columnGap = widgetDP(metricWidget, l.Gap)
	}
	rowGap := widgetDP(metricWidget, l.RowGap)
	if rowGap == 0 {
		rowGap = widgetDP(metricWidget, l.Gap)
	}

	items := make([]gridItem, 0, len(children))
	occupied := make([][]bool, 0, 4)
	rowCount := 0
	for _, child := range children {
		if child == nil {
			continue
		}
		data := gridDataOf(child.LayoutData())
		if data.ColumnSpan < 1 {
			data.ColumnSpan = 1
		}
		if data.RowSpan < 1 {
			data.RowSpan = 1
		}
		if data.ColumnSpan > columns {
			data.ColumnSpan = columns
		}

		row, col := findGridSlot(&occupied, columns, data.ColumnSpan, data.RowSpan)
		markGridSlot(&occupied, columns, row, col, data.ColumnSpan, data.RowSpan)
		if row+data.RowSpan > rowCount {
			rowCount = row + data.RowSpan
		}

		size := preferredSizeOf(child)
		items = append(items, gridItem{
			widget: child,
			data:   data,
			row:    row,
			col:    col,
			w:      max32(0, size.Width),
			h:      max32(0, size.Height),
		})
	}
	if len(items) == 0 {
		return
	}

	columnWidths := make([]int32, columns)
	for _, item := range items {
		if item.data.ColumnSpan == 1 && item.w > columnWidths[item.col] {
			columnWidths[item.col] = item.w
		}
	}
	for _, item := range items {
		if item.data.ColumnSpan <= 1 {
			continue
		}
		required := max32(0, item.w-columnGap*int32(item.data.ColumnSpan-1))
		current := sumInt32(columnWidths[item.col : item.col+item.data.ColumnSpan])
		if required > current {
			growEvenly(columnWidths[item.col:item.col+item.data.ColumnSpan], required-current)
		}
	}

	totalColumnGap := columnGap * int32(maxInt(0, columns-1))
	extraWidth := content.W - sumInt32(columnWidths) - totalColumnGap
	if extraWidth > 0 {
		distributeWeighted(columnWidths, l.ColumnWeights, extraWidth, 1)
	}

	rowHeights := make([]int32, rowCount)
	for _, item := range items {
		required := max32(0, item.h-rowGap*int32(item.data.RowSpan-1))
		current := sumInt32(rowHeights[item.row : item.row+item.data.RowSpan])
		if required > current {
			growEvenly(rowHeights[item.row:item.row+item.data.RowSpan], required-current)
		}
	}

	totalRowGap := rowGap * int32(maxInt(0, rowCount-1))
	extraHeight := content.H - sumInt32(rowHeights) - totalRowGap
	if extraHeight > 0 {
		distributeWeighted(rowHeights, l.RowWeights, extraHeight, 0)
	}

	columnPositions := positionsFromSizes(content.X, columnWidths, columnGap)
	rowPositions := positionsFromSizes(content.Y, rowHeights, rowGap)

	defaultHAlign := normalizeAlignment(l.HorizontalAlign, AlignStretch)
	defaultVAlign := normalizeAlignment(l.VerticalAlign, AlignStretch)

	for _, item := range items {
		cell := Rect{
			X: columnPositions[item.col],
			Y: rowPositions[item.row],
			W: spanLength(columnWidths[item.col:item.col+item.data.ColumnSpan], columnGap),
			H: spanLength(rowHeights[item.row:item.row+item.data.RowSpan], rowGap),
		}
		item.widget.SetBounds(alignedRect(
			cell,
			item.w,
			item.h,
			normalizeAlignment(item.data.HorizontalAlign, defaultHAlign),
			normalizeAlignment(item.data.VerticalAlign, defaultVAlign),
		))
	}
}

// Apply 按表单布局规则排布子控件。
func (l FormLayout) Apply(parent Rect, children []Widget) {
	metricWidget := layoutMetricWidget(children)
	padding := scaleInsetsForWidget(metricWidget, l.Padding)
	content := padding.inset(parent)
	if content.Empty() {
		return
	}

	labelWidth := widgetDP(metricWidget, l.LabelWidth)
	if labelWidth <= 0 {
		for index := 0; index < len(children); index += 2 {
			if children[index] == nil {
				continue
			}
			width := max32(0, preferredSizeOf(children[index]).Width)
			if width > labelWidth {
				labelWidth = width
			}
		}
	}
	if labelWidth > content.W {
		labelWidth = content.W
	}

	rowGap := widgetDP(metricWidget, l.RowGap)
	columnGap := widgetDP(metricWidget, l.ColumnGap)
	defaultAlign := normalizeAlignment(l.CrossAlign, AlignCenter)

	y := content.Y
	for index := 0; index < len(children); index += 2 {
		label := children[index]
		var field Widget
		if index+1 < len(children) {
			field = children[index+1]
		}

		if field == nil {
			if label == nil {
				continue
			}
			data := formDataOf(label.LayoutData())
			labelSize := preferredSizeForWidth(label, content.W)
			cell := Rect{X: content.X, Y: y, W: content.W, H: max32(0, labelSize.Height)}
			label.SetBounds(alignedRect(
				cell,
				labelSize.Width,
				labelSize.Height,
				AlignStretch,
				normalizeAlignment(data.Align, defaultAlign),
			))
			y += cell.H + rowGap
			continue
		}
		if label == nil {
			fieldData := formDataOf(field.LayoutData())
			fieldSize := preferredSizeForWidth(field, content.W)
			rowHeight := max32(0, fieldSize.Height)
			field.SetBounds(formFieldRect(
				Rect{X: content.X, Y: y, W: content.W, H: rowHeight},
				fieldSize.Width,
				fieldSize.Height,
				defaultFormData(fieldData),
				defaultAlign,
			))
			y += rowHeight + rowGap
			continue
		}

		labelData := formDataOf(label.LayoutData())
		fieldData := defaultFormData(formDataOf(field.LayoutData()))
		labelSize := preferredSizeForWidth(label, labelWidth)
		fieldWidth := max32(0, content.W-labelWidth-columnGap)
		fieldSize := preferredSizeForWidth(field, fieldWidth)
		rowHeight := max32(labelSize.Height, fieldSize.Height)

		labelCell := Rect{X: content.X, Y: y, W: labelWidth, H: rowHeight}
		fieldCell := Rect{
			X: content.X + labelWidth + columnGap,
			Y: y,
			W: max32(0, content.W-labelWidth-columnGap),
			H: rowHeight,
		}

		label.SetBounds(alignedRect(
			labelCell,
			labelWidth,
			labelSize.Height,
			AlignStretch,
			normalizeAlignment(labelData.Align, defaultAlign),
		))
		field.SetBounds(formFieldRect(
			fieldCell,
			fieldSize.Width,
			fieldSize.Height,
			fieldData,
			defaultAlign,
		))

		y += rowHeight + rowGap
	}
}

// Apply 兼容旧接口并委托给行布局或列布局。
func (l LinearLayout) Apply(parent Rect, children []Widget) {
	padding := UniformInsets(l.Padding)
	if l.Axis == AxisHorizontal {
		RowLayout{
			Gap:        l.Gap,
			Padding:    padding,
			ItemSize:   l.ItemSize,
			CrossAlign: AlignStretch,
		}.Apply(parent, children)
		return
	}
	ColumnLayout{
		Gap:        l.Gap,
		Padding:    padding,
		ItemSize:   l.ItemSize,
		CrossAlign: AlignStretch,
	}.Apply(parent, children)
}

// flexOptions 描述线性布局内部使用的归一化配置。
type flexOptions struct {
	// axis 表示线性布局的主轴方向。
	axis Axis
	// gap 表示相邻子控件之间的间距。
	gap int32
	// padding 表示内容区域的内边距。
	padding Insets
	// itemSize 表示主轴上的固定尺寸。
	itemSize int32
	// crossAlign 表示交叉轴上的默认对齐方式。
	crossAlign Alignment
}

// flexItem 表示线性布局中的单个排版项。
type flexItem struct {
	// widget 表示关联的控件实例。
	widget Widget
	// data 表示控件附带的线性布局参数。
	data FlexLayoutData
	// main 表示主轴上的基础尺寸。
	main int32
	// cross 表示交叉轴上的基础尺寸。
	cross int32
}

// preferredSizer 描述可直接返回首选尺寸的控件能力。
type preferredSizer interface {
	// preferredSize ????????????????
	preferredSize() core.Size
}

type widthConstrainedSizer interface {
	preferredSizeForWidth(width int32) core.Size
}

// applyFlexLayout 根据统一逻辑执行行布局或列布局。
func applyFlexLayout(parent Rect, children []Widget, opts flexOptions) {
	metricWidget := layoutMetricWidget(children)
	opts.padding = scaleInsetsForWidget(metricWidget, opts.padding)
	opts.gap = widgetDP(metricWidget, opts.gap)
	opts.itemSize = widgetDP(metricWidget, opts.itemSize)
	content := opts.padding.inset(parent)
	if content.Empty() {
		return
	}

	items := make([]flexItem, 0, len(children))
	totalBase := int32(0)
	totalGrow := int32(0)
	availableMain := content.W
	availableCross := content.H
	if opts.axis == AxisVertical {
		availableMain = content.H
		availableCross = content.W
	}
	for _, child := range children {
		if child == nil {
			continue
		}
		size := preferredSizeOf(child)
		if opts.axis == AxisVertical {
			size = preferredSizeForWidth(child, availableCross)
		}
		item := flexItem{
			widget: child,
			data:   flexDataOf(child.LayoutData()),
		}
		if opts.axis == AxisHorizontal {
			item.main = max32(0, size.Width)
			item.cross = max32(0, size.Height)
		} else {
			item.main = max32(0, size.Height)
			item.cross = max32(0, size.Width)
		}
		if opts.itemSize > 0 {
			item.main = opts.itemSize
		}
		if item.data.Grow > 0 {
			totalGrow += item.data.Grow
		}
		totalBase += item.main
		items = append(items, item)
	}
	if len(items) == 0 {
		return
	}

	totalGap := opts.gap * int32(maxInt(0, len(items)-1))

	extra := availableMain - totalBase - totalGap
	if extra < 0 {
		extra = 0
	}

	cursor := content.X
	crossStart := content.Y
	if opts.axis == AxisVertical {
		cursor = content.Y
		crossStart = content.X
	}

	remainingExtra := extra
	remainingGrow := totalGrow
	defaultAlign := normalizeAlignment(opts.crossAlign, AlignStretch)

	for _, item := range items {
		main := item.main
		if item.data.Grow > 0 && remainingGrow > 0 && remainingExtra > 0 {
			share := remainingExtra * item.data.Grow / remainingGrow
			main += share
			remainingExtra -= share
			remainingGrow -= item.data.Grow
		}

		align := normalizeAlignment(item.data.Align, defaultAlign)
		cross := item.cross
		if align == AlignStretch || cross <= 0 {
			cross = availableCross
		}
		if cross > availableCross {
			cross = availableCross
		}
		crossPos := alignedOffset(crossStart, availableCross, cross, align)

		if opts.axis == AxisHorizontal {
			item.widget.SetBounds(Rect{X: cursor, Y: crossPos, W: main, H: cross})
		} else {
			item.widget.SetBounds(Rect{X: crossPos, Y: cursor, W: cross, H: main})
		}
		cursor += main + opts.gap
	}
}

// preferredSizeOf 返回控件在布局前声明的首选尺寸。
func preferredSizeOf(widget Widget) core.Size {
	if widget == nil {
		return core.Size{}
	}
	if info, ok := widget.(interface {
		preferredSizeInfo() (core.Size, bool)
	}); ok {
		size, logical := info.preferredSizeInfo()
		if size.Width > 0 || size.Height > 0 {
			if logical {
				return scaleSizeForWidget(widget, size)
			}
			return size
		}
	}
	if sized, ok := widget.(preferredSizer); ok {
		return sized.preferredSize()
	}
	bounds := widget.Bounds()
	return core.Size{Width: bounds.W, Height: bounds.H}
}

func preferredSizeForWidth(widget Widget, width int32) core.Size {
	if widget == nil {
		return core.Size{}
	}
	if width > 0 {
		if sized, ok := widget.(widthConstrainedSizer); ok {
			return sized.preferredSizeForWidth(width)
		}
	}
	return preferredSizeOf(widget)
}

// gridItem 表示网格布局中的单个排版项。
type gridItem struct {
	// widget 表示关联的控件实例。
	widget Widget
	// data 表示控件附带的网格布局参数。
	data GridLayoutData
	// row 表示控件所在的起始行。
	row int
	// col 表示控件所在的起始列。
	col int
	// w 表示控件首选宽度。
	w int32
	// h 表示控件首选高度。
	h int32
}

// flexDataOf 将布局数据转换为线性布局参数。
func flexDataOf(data any) FlexLayoutData {
	switch typed := data.(type) {
	case FlexLayoutData:
		return typed
	case *FlexLayoutData:
		if typed != nil {
			return *typed
		}
	}
	return FlexLayoutData{}
}

// gridDataOf 将布局数据转换为网格布局参数。
func gridDataOf(data any) GridLayoutData {
	switch typed := data.(type) {
	case GridLayoutData:
		return typed
	case *GridLayoutData:
		if typed != nil {
			return *typed
		}
	}
	return GridLayoutData{}
}

// formDataOf 将布局数据转换为表单布局参数。
func formDataOf(data any) FormLayoutData {
	switch typed := data.(type) {
	case FormLayoutData:
		return typed
	case *FormLayoutData:
		if typed != nil {
			return *typed
		}
	}
	return FormLayoutData{}
}

// defaultFormData 补齐表单布局的默认参数。
func defaultFormData(data FormLayoutData) FormLayoutData {
	if data.Grow <= 0 {
		data.Grow = 1
	}
	if data.Align == AlignDefault {
		data.Align = AlignStretch
	}
	return data
}

// defaultWeight 返回指定下标的权重或兜底权重。
func defaultWeight(weights []int32, index int, fallback int32) int32 {
	if index >= 0 && index < len(weights) && weights[index] > 0 {
		return weights[index]
	}
	return fallback
}

// distributeWeighted 按权重把剩余空间分配给一组尺寸。
func distributeWeighted(sizes []int32, weights []int32, extra int32, fallbackWeight int32) {
	if len(sizes) == 0 || extra <= 0 {
		return
	}
	totalWeight := int32(0)
	for index := range sizes {
		totalWeight += defaultWeight(weights, index, fallbackWeight)
	}
	if totalWeight <= 0 {
		return
	}

	remaining := extra
	remainingWeight := totalWeight
	for index := range sizes {
		weight := defaultWeight(weights, index, fallbackWeight)
		if weight <= 0 {
			continue
		}
		share := remaining * weight / remainingWeight
		sizes[index] += share
		remaining -= share
		remainingWeight -= weight
	}
}

// growEvenly 把额外空间平均分配给一组尺寸。
func growEvenly(sizes []int32, extra int32) {
	if len(sizes) == 0 || extra <= 0 {
		return
	}
	remaining := extra
	for index := range sizes {
		share := remaining / int32(len(sizes)-index)
		sizes[index] += share
		remaining -= share
	}
}

// positionsFromSizes 根据尺寸和间距计算各项起始坐标。
func positionsFromSizes(start int32, sizes []int32, gap int32) []int32 {
	positions := make([]int32, len(sizes))
	cursor := start
	for index, size := range sizes {
		positions[index] = cursor
		cursor += size + gap
	}
	return positions
}

// spanLength 返回一组尺寸加上间距后的总长度。
func spanLength(sizes []int32, gap int32) int32 {
	if len(sizes) == 0 {
		return 0
	}
	return sumInt32(sizes) + gap*int32(len(sizes)-1)
}

// sumInt32 计算整型切片的总和。
func sumInt32(values []int32) int32 {
	total := int32(0)
	for _, value := range values {
		total += value
	}
	return total
}

// maxInt 返回两个整数中的较大值。
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ceilDiv 返回向上取整后的整除结果。
func ceilDiv(value, divisor int32) int32 {
	if divisor <= 0 {
		return value
	}
	return (value + divisor - 1) / divisor
}

// normalizeAlignment 把默认对齐值替换成给定的回退值。
func normalizeAlignment(value, fallback Alignment) Alignment {
	if value == AlignDefault {
		return fallback
	}
	return value
}

// alignedOffset 计算元素在某个轴上的对齐起点。
func alignedOffset(start, available, size int32, align Alignment) int32 {
	switch align {
	case AlignEnd:
		return start + max32(0, available-size)
	case AlignCenter:
		return start + max32(0, (available-size)/2)
	default:
		return start
	}
}

// alignedRect 根据对齐方式在单元格内放置一个矩形。
func alignedRect(cell Rect, prefW, prefH int32, hAlign, vAlign Alignment) Rect {
	width := prefW
	height := prefH
	if hAlign == AlignStretch || width <= 0 {
		width = cell.W
	}
	if vAlign == AlignStretch || height <= 0 {
		height = cell.H
	}
	if width > cell.W {
		width = cell.W
	}
	if height > cell.H {
		height = cell.H
	}
	return Rect{
		X: alignedOffset(cell.X, cell.W, width, hAlign),
		Y: alignedOffset(cell.Y, cell.H, height, vAlign),
		W: width,
		H: height,
	}
}

// formFieldRect 计算表单字段控件的目标边界。
func formFieldRect(cell Rect, prefW, prefH int32, data FormLayoutData, fallbackAlign Alignment) Rect {
	width := prefW
	if data.Grow > 0 || width <= 0 {
		width = cell.W
	}
	if width > cell.W {
		width = cell.W
	}
	align := normalizeAlignment(data.Align, fallbackAlign)
	return Rect{
		X: cell.X,
		Y: alignedOffset(cell.Y, cell.H, min32(cell.H, max32(0, prefH)), align),
		W: width,
		H: heightForAlign(cell.H, prefH, align),
	}
}

func absoluteDataOf(data any) (AbsoluteLayoutData, bool) {
	switch typed := data.(type) {
	case AbsoluteLayoutData:
		return typed, true
	case *AbsoluteLayoutData:
		if typed != nil {
			return *typed, true
		}
	}
	return AbsoluteLayoutData{}, false
}

func layoutMetricWidget(children []Widget) Widget {
	for _, child := range children {
		if child != nil {
			return child
		}
	}
	return nil
}

func widgetDP(widget Widget, value int32) int32 {
	if value == 0 {
		return 0
	}
	node := asWidgetNode(widget)
	if node == nil {
		return value
	}
	scene := node.scene()
	if scene == nil || scene.app == nil {
		return value
	}
	return scene.app.DP(value)
}

func scaleInsetsForWidget(widget Widget, value Insets) Insets {
	return Insets{
		Top:    widgetDP(widget, value.Top),
		Right:  widgetDP(widget, value.Right),
		Bottom: widgetDP(widget, value.Bottom),
		Left:   widgetDP(widget, value.Left),
	}
}

func scaleSizeForWidget(widget Widget, value core.Size) core.Size {
	return core.Size{
		Width:  widgetDP(widget, value.Width),
		Height: widgetDP(widget, value.Height),
	}
}

// heightForAlign 根据对齐方式计算实际高度。
func heightForAlign(cellH, prefH int32, align Alignment) int32 {
	if align == AlignStretch || prefH <= 0 {
		return cellH
	}
	if prefH > cellH {
		return cellH
	}
	return prefH
}

// findGridSlot 在占用表中寻找一个可放置控件的位置。
func findGridSlot(occupied *[][]bool, columns, colSpan, rowSpan int) (int, int) {
	for row := 0; ; row++ {
		ensureGridRows(occupied, row+rowSpan, columns)
		for col := 0; col <= columns-colSpan; col++ {
			if gridSlotFree(*occupied, row, col, colSpan, rowSpan) {
				return row, col
			}
		}
	}
}

// ensureGridRows 确保占用表至少拥有指定数量的行。
func ensureGridRows(occupied *[][]bool, rows, columns int) {
	for len(*occupied) < rows {
		*occupied = append(*occupied, make([]bool, columns))
	}
}

// gridSlotFree 判断某个网格区域是否全部空闲。
func gridSlotFree(occupied [][]bool, row, col, colSpan, rowSpan int) bool {
	for y := row; y < row+rowSpan; y++ {
		for x := col; x < col+colSpan; x++ {
			if occupied[y][x] {
				return false
			}
		}
	}
	return true
}

// markGridSlot 把某个网格区域标记为已占用。
func markGridSlot(occupied *[][]bool, columns, row, col, colSpan, rowSpan int) {
	ensureGridRows(occupied, row+rowSpan, columns)
	for y := row; y < row+rowSpan; y++ {
		for x := col; x < col+colSpan; x++ {
			(*occupied)[y][x] = true
		}
	}
}

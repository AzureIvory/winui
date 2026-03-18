//go:build windows

package widgets

// Layout 定义容器对子控件执行布局的行为。
type Layout interface {
	// Apply 根据父区域和子控件集合应用布局。
	Apply(parent Rect, children []Widget)
}

// AbsoluteLayout 表示不主动调整子控件位置的绝对布局。
type AbsoluteLayout struct{}

// Apply 将布局应用到给定子控件。
func (AbsoluteLayout) Apply(Rect, []Widget) {}

// Axis 表示线性布局的主轴方向。
type Axis int

const (
	// AxisHorizontal 表示水平主轴。
	AxisHorizontal Axis = iota + 1
	// AxisVertical 表示垂直主轴。
	AxisVertical
)

// LinearLayout 表示按同一主轴顺序排布子控件的布局。
type LinearLayout struct {
	// Axis 指定布局主轴方向。
	Axis Axis
	// Gap 指定相邻子控件间距。
	Gap int32
	// Padding 指定容器内边距。
	Padding int32
	// ItemSize 指定主轴方向上的固定尺寸。
	ItemSize int32
}

// Apply 将布局应用到给定子控件。
func (l LinearLayout) Apply(parent Rect, children []Widget) {
	if len(children) == 0 {
		return
	}
	if l.Axis == 0 {
		l.Axis = AxisVertical
	}

	cursorX := parent.X + l.Padding
	cursorY := parent.Y + l.Padding
	availableW := parent.W - l.Padding*2
	availableH := parent.H - l.Padding*2

	for _, child := range children {
		if child == nil {
			continue
		}
		bounds := child.Bounds()
		switch l.Axis {
		case AxisHorizontal:
			width := bounds.W
			if l.ItemSize > 0 {
				width = l.ItemSize
			}
			if width <= 0 {
				width = availableW / int32(len(children))
			}
			height := bounds.H
			if height <= 0 {
				height = availableH
			}
			child.SetBounds(Rect{X: cursorX, Y: cursorY, W: width, H: height})
			cursorX += width + l.Gap
		default:
			height := bounds.H
			if l.ItemSize > 0 {
				height = l.ItemSize
			}
			if height <= 0 {
				height = availableH / int32(len(children))
			}
			width := bounds.W
			if width <= 0 {
				width = availableW
			}
			child.SetBounds(Rect{X: cursorX, Y: cursorY, W: width, H: height})
			cursorY += height + l.Gap
		}
	}
}

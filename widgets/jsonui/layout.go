//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/AzureIvory/winui/widgets"
)

type layoutFrame struct {
	X exprSource
	Y exprSource
	R exprSource
	B exprSource
	W exprSource
	H exprSource
}

type absoluteLayoutData struct {
	frame  layoutFrame
	window *Window
}

type absoluteLayout struct {
	window *Window
}

func (l absoluteLayout) Apply(parent widgets.Rect, children []widgets.Widget) {
	for _, child := range children {
		if child == nil {
			continue
		}
		data, ok := child.LayoutData().(absoluteLayoutData)
		if !ok {
			continue
		}

		ctx := data.exprContext(parent)
		rect := child.Bounds()
		size := widgets.PreferredSize(child)
		width := size.Width
		height := size.Height

		if data.frame.W.Has {
			width = data.resolve(data.frame.W, AxisWidth, ctx)
		} else if data.frame.X.Has && data.frame.R.Has {
			width = ctx.ParentW - data.resolve(data.frame.X, AxisX, ctx) - data.resolve(data.frame.R, AxisWidth, ctx)
		}
		if data.frame.H.Has {
			height = data.resolve(data.frame.H, AxisHeight, ctx)
		} else if data.frame.Y.Has && data.frame.B.Has {
			height = ctx.ParentH - data.resolve(data.frame.Y, AxisY, ctx) - data.resolve(data.frame.B, AxisHeight, ctx)
		}

		if width < 0 {
			width = 0
		}
		if height < 0 {
			height = 0
		}

		rect.W = data.scale(width)
		rect.H = data.scale(height)
		if data.frame.X.Has {
			rect.X = parent.X + data.scale(data.resolve(data.frame.X, AxisX, ctx))
		} else if data.frame.R.Has {
			rect.X = parent.X + parent.W - data.scale(data.resolve(data.frame.R, AxisWidth, ctx)) - rect.W
		}
		if data.frame.Y.Has {
			rect.Y = parent.Y + data.scale(data.resolve(data.frame.Y, AxisY, ctx))
		} else if data.frame.B.Has {
			rect.Y = parent.Y + parent.H - data.scale(data.resolve(data.frame.B, AxisHeight, ctx)) - rect.H
		}
		child.SetBounds(rect)
	}
}

func (d absoluteLayoutData) resolve(source exprSource, axis ExprAxis, ctx ExprContext) int32 {
	if !source.Has {
		return 0
	}
	return source.Literal.Resolve(axis, ctx)
}

func (d absoluteLayoutData) exprContext(parent widgets.Rect) ExprContext {
	scale := d.scaleFactor()
	windowSize := d.window.sceneClientSize()
	return ExprContext{
		WindowW: logicalPixels(windowSize.Width, scale),
		WindowH: logicalPixels(windowSize.Height, scale),
		ParentW: logicalPixels(parent.W, scale),
		ParentH: logicalPixels(parent.H, scale),
	}
}

func (d absoluteLayoutData) scale(value int32) int32 {
	if value == 0 {
		return 0
	}
	if d.window == nil {
		return value
	}
	d.window.mu.Lock()
	scene := d.window.scene
	d.window.mu.Unlock()
	if scene == nil || scene.App() == nil {
		return value
	}
	return scene.App().DP(value)
}

func (d absoluteLayoutData) scaleFactor() float64 {
	if d.window == nil {
		return 1
	}
	d.window.mu.Lock()
	scene := d.window.scene
	d.window.mu.Unlock()
	if scene == nil || scene.App() == nil {
		return 1
	}
	scale := scene.App().DPI().Scale
	if scale <= 0 {
		return 1
	}
	return scale
}

func logicalPixels(value int32, scale float64) int32 {
	if value == 0 {
		return 0
	}
	if scale <= 0 {
		scale = 1
	}
	return int32(math.Round(float64(value) / scale))
}

func buildLayout(window *Window, raw json.RawMessage) (widgets.Layout, string, error) {
	if len(raw) == 0 {
		return absoluteLayout{window: window}, "abs", nil
	}

	var kind string
	if err := json.Unmarshal(raw, &kind); err == nil {
		switch kind {
		case "", "abs":
			return absoluteLayout{window: window}, "abs", nil
		case "row":
			return widgets.RowLayout{}, "row", nil
		case "col":
			return widgets.ColumnLayout{}, "col", nil
		case "grid":
			return widgets.GridLayout{Columns: 1}, "grid", nil
		case "form":
			return widgets.FormLayout{}, "form", nil
		default:
			return nil, "", fmt.Errorf("unsupported layout %q", kind)
		}
	}

	var spec struct {
		Type   string          `json:"type"`
		Gap    json.RawMessage `json:"gap"`
		Pad    json.RawMessage `json:"pad"`
		Item   json.RawMessage `json:"item"`
		Cross  string          `json:"cross"`
		Cols   int             `json:"cols"`
		RowGap json.RawMessage `json:"rowGap"`
		ColGap json.RawMessage `json:"colGap"`
		LabelW json.RawMessage `json:"labelW"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, "", err
	}

	switch spec.Type {
	case "", "abs":
		return absoluteLayout{window: window}, "abs", nil
	case "row":
		layout := widgets.RowLayout{}
		assignLayoutInt(spec.Gap, &layout.Gap)
		assignLayoutInt(spec.Item, &layout.ItemSize)
		assignInsets(spec.Pad, &layout.Padding)
		if spec.Cross != "" {
			if align, ok, err := parseAlignmentValue(spec.Cross); err != nil {
				return nil, "", err
			} else if ok {
				layout.CrossAlign = align
			}
		}
		return layout, "row", nil
	case "col":
		layout := widgets.ColumnLayout{}
		assignLayoutInt(spec.Gap, &layout.Gap)
		assignLayoutInt(spec.Item, &layout.ItemSize)
		assignInsets(spec.Pad, &layout.Padding)
		if spec.Cross != "" {
			if align, ok, err := parseAlignmentValue(spec.Cross); err != nil {
				return nil, "", err
			} else if ok {
				layout.CrossAlign = align
			}
		}
		return layout, "col", nil
	case "grid":
		layout := widgets.GridLayout{Columns: spec.Cols}
		if layout.Columns <= 0 {
			layout.Columns = 1
		}
		assignLayoutInt(spec.Gap, &layout.Gap)
		assignLayoutInt(spec.RowGap, &layout.RowGap)
		assignLayoutInt(spec.ColGap, &layout.ColumnGap)
		assignInsets(spec.Pad, &layout.Padding)
		return layout, "grid", nil
	case "form":
		layout := widgets.FormLayout{}
		assignLayoutInt(spec.RowGap, &layout.RowGap)
		assignLayoutInt(spec.ColGap, &layout.ColumnGap)
		assignLayoutInt(spec.LabelW, &layout.LabelWidth)
		assignInsets(spec.Pad, &layout.Padding)
		return layout, "form", nil
	default:
		return nil, "", fmt.Errorf("unsupported layout %q", spec.Type)
	}
}

func assignLayoutInt(raw json.RawMessage, target *int32) error {
	if len(raw) == 0 || target == nil {
		return nil
	}
	value, err := decodeInt32Literal(raw)
	if err != nil {
		return err
	}
	*target = value
	return nil
}

func assignInsets(raw json.RawMessage, target *widgets.Insets) error {
	if len(raw) == 0 || target == nil {
		return nil
	}
	var values []int32
	if err := json.Unmarshal(raw, &values); err == nil {
		switch len(values) {
		case 1:
			*target = widgets.UniformInsets(values[0])
		case 2:
			*target = widgets.Insets{Top: values[0], Right: values[1], Bottom: values[0], Left: values[1]}
		case 4:
			*target = widgets.Insets{Top: values[0], Right: values[1], Bottom: values[2], Left: values[3]}
		default:
			return fmt.Errorf("padding must contain 1, 2, or 4 values")
		}
		return nil
	}

	value, err := decodeInt32Literal(raw)
	if err != nil {
		return err
	}
	*target = widgets.UniformInsets(value)
	return nil
}

func bindingExprValue(value any) (ScalarExpr, bool) {
	expr, err := ParseScalarExpr(value)
	if err != nil {
		return ScalarExpr{}, false
	}
	return expr, true
}

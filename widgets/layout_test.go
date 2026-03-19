//go:build windows

package widgets

import "testing"

// TestRowLayoutGrowAndAlign 验证行布局的扩展和交叉轴对齐行为。
func TestRowLayoutGrowAndAlign(t *testing.T) {
	panel := NewPanel("root")
	panel.SetBounds(Rect{X: 0, Y: 0, W: 300, H: 100})
	panel.SetLayout(RowLayout{
		Gap:        10,
		Padding:    UniformInsets(10),
		CrossAlign: AlignCenter,
	})

	left := NewPanel("left")
	left.SetBounds(Rect{W: 50, H: 20})
	right := NewPanel("right")
	right.SetBounds(Rect{W: 40, H: 30})
	right.SetLayoutData(FlexLayoutData{Grow: 1})

	panel.AddAll(left, right)

	if got := left.Bounds(); got != (Rect{X: 10, Y: 40, W: 50, H: 20}) {
		t.Fatalf("unexpected left bounds: %+v", got)
	}
	if got := right.Bounds(); got != (Rect{X: 70, Y: 35, W: 220, H: 30}) {
		t.Fatalf("unexpected right bounds: %+v", got)
	}
}

// TestGridLayoutSpanPlacement 验证网格布局中的跨列摆放结果。
func TestGridLayoutSpanPlacement(t *testing.T) {
	panel := NewPanel("grid")
	panel.SetBounds(Rect{X: 0, Y: 0, W: 340, H: 120})
	panel.SetLayout(GridLayout{
		Columns: 3,
		Gap:     10,
		Padding: UniformInsets(10),
	})

	first := NewPanel("first")
	first.SetBounds(Rect{H: 20})
	second := NewPanel("second")
	second.SetBounds(Rect{H: 20})
	second.SetLayoutData(GridLayoutData{ColumnSpan: 2})
	third := NewPanel("third")
	third.SetBounds(Rect{H: 20})
	third.SetLayoutData(GridLayoutData{ColumnSpan: 3})

	panel.AddAll(first, second, third)

	if got := first.Bounds(); got != (Rect{X: 10, Y: 10, W: 100, H: 20}) {
		t.Fatalf("unexpected first bounds: %+v", got)
	}
	if got := second.Bounds(); got != (Rect{X: 120, Y: 10, W: 210, H: 20}) {
		t.Fatalf("unexpected second bounds: %+v", got)
	}
	if got := third.Bounds(); got != (Rect{X: 10, Y: 40, W: 320, H: 20}) {
		t.Fatalf("unexpected third bounds: %+v", got)
	}
}

// TestFormLayoutFieldGrow 验证表单布局中的字段扩展行为。
func TestFormLayoutFieldGrow(t *testing.T) {
	panel := NewPanel("form")
	panel.SetBounds(Rect{X: 0, Y: 0, W: 300, H: 120})
	panel.SetLayout(FormLayout{
		Padding:    UniformInsets(10),
		RowGap:     10,
		ColumnGap:  10,
		LabelWidth: 80,
		CrossAlign: AlignCenter,
	})

	label := NewPanel("label")
	label.SetBounds(Rect{H: 20})
	field := NewPanel("field")
	field.SetBounds(Rect{W: 60, H: 30})

	panel.AddAll(label, field)

	if got := label.Bounds(); got != (Rect{X: 10, Y: 15, W: 80, H: 20}) {
		t.Fatalf("unexpected label bounds: %+v", got)
	}
	if got := field.Bounds(); got != (Rect{X: 100, Y: 10, W: 190, H: 30}) {
		t.Fatalf("unexpected field bounds: %+v", got)
	}
}

// TestTextStyleFontMergesByField 验证文本样式中的字体会按字段合并。
func TestTextStyleFontMergesByField(t *testing.T) {
	label := NewLabel("label", "demo")
	label.SetStyle(TextStyle{
		Font: FontSpec{
			SizeDP: 24,
		},
	})

	style := label.resolveStyle(nil)
	if style.Font.Face == "" {
		t.Fatalf("expected default face preserved")
	}
	if style.Font.SizeDP != 24 {
		t.Fatalf("expected override size applied, got %d", style.Font.SizeDP)
	}
}

// TestChoiceStyleFontMergesByField 验证选择类控件样式中的字体会按字段合并。
func TestChoiceStyleFontMergesByField(t *testing.T) {
	base := DefaultTheme().CheckBox
	override := ChoiceStyle{
		Font: FontSpec{
			Weight: 700,
		},
	}

	merged := mergeChoiceStyle(base, override)
	if merged.Font.Face == "" {
		t.Fatalf("expected face preserved")
	}
	if merged.Font.Weight != 700 {
		t.Fatalf("expected weight override applied, got %d", merged.Font.Weight)
	}
	if merged.Font.SizeDP != base.Font.SizeDP {
		t.Fatalf("expected size preserved, got %d", merged.Font.SizeDP)
	}
}

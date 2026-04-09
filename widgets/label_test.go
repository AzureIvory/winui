//go:build windows

package widgets

import "testing"

func TestAbsoluteLayoutMeasuresMultilineLabelHeightFromWidth(t *testing.T) {
	root := NewPanel("root")
	root.SetLayout(AbsoluteLayout{})
	root.SetBounds(Rect{W: 220, H: 200})

	label := NewLabel("body", "This is a long line that should wrap into multiple rows.")
	label.SetMultiline(true)
	label.SetWordWrap(true)
	label.SetLayoutData(AbsoluteLayoutData{
		Left:     0,
		Top:      0,
		Width:    90,
		HasLeft:  true,
		HasTop:   true,
		HasWidth: true,
	})

	root.Add(label)

	if got := label.Bounds().H; got <= 28 {
		t.Fatalf("label.Bounds().H = %d, want > 28 for wrapped multiline label", got)
	}
}

func TestColumnLayoutMeasuresMultilineLabelWithinAvailableWidth(t *testing.T) {
	root := NewPanel("root")
	root.SetLayout(ColumnLayout{})
	root.SetBounds(Rect{W: 120, H: 220})

	label := NewLabel("body", "This is a long line that should wrap into multiple rows.")
	label.SetMultiline(true)
	label.SetWordWrap(true)

	root.Add(label)

	if got := label.Bounds().H; got <= 28 {
		t.Fatalf("label.Bounds().H = %d, want > 28 for wrapped multiline label", got)
	}
}

//go:build windows

package widgets

import "testing"

func TestColumnLayoutMeasuresNestedPanelNaturalHeight(t *testing.T) {
	root := NewPanel("root")
	root.SetLayout(ColumnLayout{Gap: 12})
	root.SetBounds(Rect{W: 320, H: 480})

	card := NewPanel("card")
	card.SetLayout(ColumnLayout{
		Gap:     8,
		Padding: UniformInsets(12),
	})
	card.Add(NewLabel("title", "Nested panel title"))

	body := NewLabel("body", "This nested panel should contribute a positive natural height to the parent column layout.")
	body.SetMultiline(true)
	body.SetWordWrap(true)
	card.Add(body)

	root.Add(card)

	if got := card.Bounds().H; got <= 40 {
		t.Fatalf("card.Bounds().H = %d, want > 40 for nested panel natural height", got)
	}
	if got := body.Bounds().H; got <= 28 {
		t.Fatalf("body.Bounds().H = %d, want > 28 for wrapped content", got)
	}
}

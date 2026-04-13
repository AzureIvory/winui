//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestChoiceIndicatorCheckFillUsesIndicatorTint(t *testing.T) {
	background := core.RGB(255, 255, 255)
	indicator := core.RGB(14, 165, 233)

	got := choiceIndicatorCheckFill(background, indicator)
	want := core.RGB(217, 241, 252)
	if got != want {
		t.Fatalf("choiceIndicatorCheckFill() = %#08x, want %#08x", got, want)
	}
	if got == background {
		t.Fatal("choiceIndicatorCheckFill() returned the unchanged background color")
	}
	if got == indicator {
		t.Fatal("choiceIndicatorCheckFill() returned the solid indicator color")
	}
}

func TestResolveChoiceCheckMarkColor(t *testing.T) {
	tests := []struct {
		name  string
		style ChoiceStyle
		want  core.Color
	}{
		{
			name: "white check falls back to indicator color",
			style: ChoiceStyle{
				IndicatorColor: core.RGB(14, 165, 233),
				CheckColor:     core.RGB(255, 255, 255),
			},
			want: core.RGB(14, 165, 233),
		},
		{
			name: "custom check color is preserved",
			style: ChoiceStyle{
				IndicatorColor: core.RGB(14, 165, 233),
				CheckColor:     core.RGB(17, 24, 39),
			},
			want: core.RGB(17, 24, 39),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := resolveChoiceCheckMarkColor(tt.style)
			if got != tt.want {
				t.Fatalf("resolveChoiceCheckMarkColor() = %#08x, want %#08x", got, tt.want)
			}
		})
	}
}

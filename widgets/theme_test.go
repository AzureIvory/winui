//go:build windows

package widgets

import "testing"

// TestChoiceStyleIndicatorStyleMerges 验证选择类控件的标记样式可以通过覆盖值合并。
func TestChoiceStyleIndicatorStyleMerges(t *testing.T) {
	base := DefaultTheme().CheckBox
	merged := mergeChoiceStyle(base, ChoiceStyle{IndicatorStyle: ChoiceIndicatorCheck})

	if merged.IndicatorStyle != ChoiceIndicatorCheck {
		t.Fatalf("expected indicator style override applied, got %v", merged.IndicatorStyle)
	}
}

// TestNewThemeHardModeUsesSharpDefaults 验证硬核模式会切换到更接近系统原生的默认外观。
func TestNewThemeHardModeUsesSharpDefaults(t *testing.T) {
	theme := NewTheme(ThemeOptions{HardMode: true})

	if theme.Button.CornerRadius != 0 {
		t.Fatalf("expected button corner radius 0, got %d", theme.Button.CornerRadius)
	}
	if theme.Progress.CornerRadius != 0 {
		t.Fatalf("expected progress corner radius 0, got %d", theme.Progress.CornerRadius)
	}
	if theme.Progress.ShowPercent {
		t.Fatalf("expected hard mode progress to hide percent bubble")
	}
	if theme.CheckBox.IndicatorStyle != ChoiceIndicatorCheck {
		t.Fatalf("expected hard mode checkbox to use check style, got %v", theme.CheckBox.IndicatorStyle)
	}
	if theme.RadioButton.IndicatorStyle != ChoiceIndicatorDot {
		t.Fatalf("expected hard mode radio button to keep dot style, got %v", theme.RadioButton.IndicatorStyle)
	}
	if theme.ComboBox.CornerRadius != 0 {
		t.Fatalf("expected combo box corner radius 0, got %d", theme.ComboBox.CornerRadius)
	}
	if theme.Edit.CornerRadius != 0 {
		t.Fatalf("expected edit box corner radius 0, got %d", theme.Edit.CornerRadius)
	}
}

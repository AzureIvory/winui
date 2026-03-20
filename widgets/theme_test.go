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

// TestControlModeNormalization 验证控件后端模式会规范化为已知值。
func TestControlModeNormalization(t *testing.T) {
	if got := normalizeControlMode(ModeNative); got != ModeNative {
		t.Fatalf("expected native mode preserved, got %v", got)
	}
	if got := normalizeControlMode(ControlMode(99)); got != ModeCustom {
		t.Fatalf("expected unknown mode fallback to custom, got %v", got)
	}
}

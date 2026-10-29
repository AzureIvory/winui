//go:build windows

package widgets

import (
	"strings"
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestMeasureForWidthPreservesNaturalAndExplicitSizing(t *testing.T) {
	if got := MeasureForWidth(nil, 180); got != (core.Size{}) {
		t.Fatalf("nil measurement=%+v", got)
	}
	label := NewLabel("measure-explicit", strings.Repeat("完整中文内容", 12))
	label.SetMultiline(true)
	label.SetWordWrap(true)
	for _, width := range []int32{0, -1} {
		if got, want := MeasureForWidth(label, width), MeasureNatural(label); got != want {
			t.Errorf("nonpositive width%d changed natural sizing: got=%+v want=%+v", width, got, want)
		}
	}
	SetPreferredSize(label, core.Size{Width: 55, Height: 37})
	if got, want := MeasureForWidth(label, 180), preferredSizeForWidth(label, 180); got != want || got != (core.Size{Width: 55, Height: 37}) {
		t.Errorf("constrained API ignored explicit preferred size or forked layout SSOT: got=%+v existing=%+v", got, want)
	}
	plain := NewLabel("measure-single-line", strings.Repeat("MODEL202", 15))
	before := plain.Bounds()
	if got := MeasureForWidth(plain, 180); got != preferredSizeForWidth(plain, 180) || got.Width <= 180 || plain.Multiline() || plain.WordWrap() || plain.Bounds() != before {
		t.Error("measurement silently enabled wrapping, clamped a single line or changed geometry")
	}
}

func TestMeasureForWidthReusesSceneTextWrapAndColumnGeometry(t *testing.T) {
	for _, item := range []struct{ name, text string }{
		{"chinese", strings.Repeat("需要完整阅读的中文标题与内容", 8)},
		{"continuous-model", "MODEL202" + strings.Repeat("MW0123456789", 5) + "ENDTOKEN202"},
		{"continuous-digits", strings.Repeat("1234567890", 8) + ".123456"},
	} {
		t.Run(item.name, func(t *testing.T) {
			scene := NewScene(nil)
			theme := DefaultTheme()
			theme.Text.Font.SizeDP = 13
			scene.SetTheme(theme)
			panel := NewPanel("measure-panel")
			panel.SetLayout(ColumnLayout{})
			scene.Root().Add(panel)
			label := NewLabel("measure-label", item.text)
			label.SetMultiline(true)
			label.SetWordWrap(true)
			panel.Add(label)
			panel.SetBounds(Rect{W: 180, H: 1 << 20})
			before := label.Bounds()
			natural := MeasureForWidth(label, 0)
			got := MeasureForWidth(label, 180)
			if got != preferredSizeForWidth(label, 180) || got.Height != before.H || got.Height <= natural.Height || got.Width > 180 || got.Width <= 0 {
				t.Errorf("shared constrained measurement/actual column geometry diverged: got=%+v natural=%+v bounds=%+v", got, natural, before)
			}
			if wide := MeasureForWidth(label, 720); wide.Height > got.Height {
				t.Errorf("increasing available width increased wrapped height: narrow=%+v wide=%+v", got, wide)
			}
			style := label.resolveStyle(&PaintCtx{scene: scene})
			if style.Format&(core.DTSingleLine|core.DTEndEllipsis|core.DTVCenter) != 0 || style.Format&core.DTWordBreak == 0 {
				t.Error("multiline/wrap label still passes single-line/ellipsis flags to its real paint path")
			}
			if label.Text != item.text || label.Bounds() != before {
				t.Error("measurement mutated original text or layout bounds")
			}
			label.SetStyle(TextStyle{Font: FontSpec{SizeDP: 24}})
			if large := MeasureForWidth(label, 180); large.Height <= got.Height {
				t.Errorf("actual scene font override was ignored: small=%+v large=%+v", got, large)
			}
		})
	}
	scene := NewScene(nil)
	short := NewLabel("measure-short", "短值")
	short.SetMultiline(true)
	short.SetWordWrap(true)
	scene.Root().Add(short)
	if got, natural := MeasureForWidth(short, 180), MeasureForWidth(short, 0); got != natural {
		t.Errorf("ordinary short label was inflated by width-constrained measurement: got=%+v natural=%+v", got, natural)
	}
}

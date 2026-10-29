//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestListSubtitleUsesConfiguredRowHitGeometry(t *testing.T) {
	for _, test := range []struct {
		name   string
		height int32
	}{{"compact_default", DefaultTheme().ListBox.ItemHeightDP}, {"subtitle_tall", 68}} {
		t.Run(test.name, func(t *testing.T) {
			list := NewListBox("phrases")
			list.SetBounds(Rect{X: 20, Y: 30, W: 300, H: 300})
			list.SetStyle(ListStyle{ItemHeightDP: test.height})
			list.SetItems([]ListItem{{Text: "标题", Subtitle: "正文"}, {Text: "第二条", Subtitle: "正文二"}, {Text: "第三条"}})
			style := list.resolveStyle(nil)
			ctx := &PaintCtx{}
			for index := 0; index < 3; index++ {
				rect := list.rowRect(index, ctx, style)
				if rect.H != test.height {
					t.Fatalf("row height=%d want=%d", rect.H, test.height)
				}
				if got := list.indexAt(core.Point{X: rect.X + 1, Y: rect.Y + rect.H/2}); got != index {
					t.Errorf("painted row %d hits row %d", index, got)
				}
				if index > 0 {
					previous := list.rowRect(index-1, ctx, style)
					if rect.Y != previous.Y+previous.H {
						t.Error("adjacent rows overlap or have gaps")
					}
				}
			}
			bounds := list.rowRect(0, ctx, style)
			title, subtitle := itemTextLineRects(ctx, bounds, style.Font, itemSubtitleFont(style.Font))
			if title.Y < bounds.Y || title.Y+title.H > subtitle.Y || subtitle.Y+subtitle.H > bounds.Y+bounds.H {
				t.Fatalf("subtitle text escapes row: row=%+v title=%+v subtitle=%+v", bounds, title, subtitle)
			}
		})
	}
}

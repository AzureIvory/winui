//go:build windows

package jsonui

import (
	"testing"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

func TestLoadDocumentStringBuildsWindows(t *testing.T) {
	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "title": "Main Demo",
      "w": 980,
      "h": 720,
      "minW": 900,
      "minH": 640,
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "label",
            "id": "title",
            "text": "JSON Demo",
            "frame": { "x": 20, "y": 20, "w": 320, "h": 28 }
          },
          {
            "type": "button",
            "id": "saveBtn",
            "text": "Save",
            "style": {
              "fg": "#0f172a",
              "downFg": "#ffffff",
              "bg": "#f5f9ff",
              "hoverBg": "#e0f2fe",
              "pressedBg": "#2563eb",
              "border": "#9dbfe8",
              "radius": 10,
              "pad": 12,
              "gap": 8
            },
            "frame": { "x": "50%-70", "y": 56, "w": 140, "h": 40 }
          }
        ]
      }
    },
    {
      "id": "tools",
      "title": "Tools",
      "w": 420,
      "h": 280,
      "root": {
        "type": "panel",
        "layout": "col",
        "children": [
          { "type": "label", "id": "hint", "text": "Second window" }
        ]
      }
    }
  ]
}`, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	if len(doc.Windows) != 2 {
		t.Fatalf("len(doc.Windows) = %d, want 2", len(doc.Windows))
	}

	mainWin := doc.Window("main")
	if mainWin == nil {
		t.Fatal("Window(main) returned nil")
	}
	if mainWin.Meta.Title != "Main Demo" {
		t.Fatalf("mainWin.Meta.Title = %q, want %q", mainWin.Meta.Title, "Main Demo")
	}
	if mainWin.Meta.Width != 980 || mainWin.Meta.Height != 720 {
		t.Fatalf("main window size = %dx%d, want 980x720", mainWin.Meta.Width, mainWin.Meta.Height)
	}
	if mainWin.Meta.MinWidth != 900 || mainWin.Meta.MinHeight != 640 {
		t.Fatalf("main window min size = %dx%d, want 900x640", mainWin.Meta.MinWidth, mainWin.Meta.MinHeight)
	}

	rootPanel, ok := mainWin.Root.(*widgets.Panel)
	if !ok {
		t.Fatalf("mainWin.Root type = %T, want *widgets.Panel", mainWin.Root)
	}
	children := rootPanel.Children()
	if len(children) != 2 {
		t.Fatalf("len(root children) = %d, want 2", len(children))
	}

	title, ok := children[0].(*widgets.Label)
	if !ok {
		t.Fatalf("children[0] type = %T, want *widgets.Label", children[0])
	}
	if title.Text != "JSON Demo" {
		t.Fatalf("title.Text = %q, want %q", title.Text, "JSON Demo")
	}

	button, ok := children[1].(*widgets.Button)
	if !ok {
		t.Fatalf("children[1] type = %T, want *widgets.Button", children[1])
	}
	if button.Text != "Save" {
		t.Fatalf("button.Text = %q, want %q", button.Text, "Save")
	}
	if button.Style.Pressed == 0 {
		t.Fatal("button.Style.Pressed was not populated from JSON style")
	}

	toolsWin := doc.Window("tools")
	if toolsWin == nil {
		t.Fatal("Window(tools) returned nil")
	}
	if doc.PrimaryWindow() == nil || doc.PrimaryWindow().ID != "main" {
		t.Fatalf("PrimaryWindow() = %#v, want main window", doc.PrimaryWindow())
	}
}

func TestLoadDocumentStringRejectsInvalidSchema(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "missing wins",
			json: `{}`,
		},
		{
			name: "empty wins",
			json: `{"wins":[]}`,
		},
		{
			name: "missing root",
			json: `{"wins":[{"id":"main","title":"oops"}]}`,
		},
		{
			name: "duplicate window id",
			json: `{"wins":[{"id":"main","root":{"type":"panel"}},{"id":"main","root":{"type":"panel"}}]}`,
		},
		{
			name: "unsupported widget type",
			json: `{"wins":[{"id":"main","root":{"type":"unknown"}}]}`,
		},
		{
			name: "duplicate widget id in window",
			json: `{"wins":[{"id":"main","root":{"type":"panel","layout":"col","children":[{"type":"label","id":"dup","text":"A"},{"type":"button","id":"dup","text":"B"}]}}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := LoadDocumentString(tt.json, LoadOptions{}); err == nil {
				t.Fatalf("LoadDocumentString unexpectedly succeeded for %s", tt.name)
			}
		})
	}
}

func TestDocumentNewAppsBuildsOneAppPerWindow(t *testing.T) {
	doc, err := LoadDocumentString(`{
  "wins": [
    { "id": "main", "title": "Main", "w": 640, "h": 480, "root": { "type": "panel" } },
    { "id": "tools", "title": "Tools", "w": 360, "h": 240, "root": { "type": "panel" } }
  ]
}`, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	windows, err := doc.NewApps(core.Options{
		ClassName:      "JSONUITest",
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	})
	if err != nil {
		t.Fatalf("NewApps returned error: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("len(windows) = %d, want 2", len(windows))
	}
	for _, window := range windows {
		if window.App == nil {
			t.Fatalf("window %q App is nil", window.ID)
		}
		if window.SceneRef == nil {
			t.Fatalf("window %q SceneRef is nil", window.ID)
		}
		if window.Window == nil {
			t.Fatalf("window %q Window is nil", window.ID)
		}
	}
}

func TestLoadDocumentStringBuildsScrollView(t *testing.T) {
	store := NewStore(map[string]any{
		"scroll": map[string]any{
			"vertical":   false,
			"horizontal": true,
		},
	})

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "scrollview",
            "id": "scroller",
            "layout": "col",
            "verticalScroll": { "bind": "scroll.vertical" },
            "horizontalScroll": { "bind": "scroll.horizontal" },
            "style": {
              "bg": "#ffffff",
              "border": "#cbd5e1",
              "radius": 12,
              "borderW": 1
            },
            "frame": { "x": 20, "y": 20, "w": 240, "h": 180 },
            "children": [
              { "type": "label", "id": "insideLabel", "text": "Inside scroll" },
              { "type": "button", "id": "insideButton", "text": "Run" }
            ]
          }
        ]
      }
    }
  ]
}`, LoadOptions{Data: store})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	win := doc.PrimaryWindow()
	if win == nil {
		t.Fatal("PrimaryWindow() returned nil")
	}

	widget := win.FindWidget("scroller")
	scroller, ok := widget.(*widgets.ScrollView)
	if !ok {
		t.Fatalf("scroller type = %T, want *widgets.ScrollView", widget)
	}
	if scroller.VerticalScroll() {
		t.Fatal("scroller.VerticalScroll() = true, want false")
	}
	if !scroller.HorizontalScroll() {
		t.Fatal("scroller.HorizontalScroll() = false, want true")
	}
	if scroller.Style.Background == 0 || scroller.Style.BorderColor == 0 {
		t.Fatal("scrollview style was not populated from JSON")
	}

	content, ok := scroller.Content().(*widgets.Panel)
	if !ok {
		t.Fatalf("scroller.Content() type = %T, want *widgets.Panel", scroller.Content())
	}
	children := content.Children()
	if len(children) != 2 {
		t.Fatalf("len(scroll content children) = %d, want 2", len(children))
	}

	store.Patch(map[string]any{
		"scroll.vertical":   true,
		"scroll.horizontal": false,
	})

	if !scroller.VerticalScroll() {
		t.Fatal("updated scroller.VerticalScroll() = false, want true")
	}
	if scroller.HorizontalScroll() {
		t.Fatal("updated scroller.HorizontalScroll() = true, want false")
	}
}

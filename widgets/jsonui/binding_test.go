//go:build windows

package jsonui

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestLoadDocumentStringAppliesBindingsFromStore(t *testing.T) {
	store := NewStore(map[string]any{
		"page": map[string]any{
			"title": "Initial Title",
			"body":  "Initial Body",
		},
	})

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "title": { "bind": "page.title", "default": "Fallback Title" },
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "label",
            "id": "title",
            "text": { "bind": "page.title" },
            "frame": { "x": 20, "y": 20, "w": 220, "h": 28 }
          },
          {
            "type": "label",
            "id": "body",
            "text": { "bind": "page.body", "default": "Fallback Body" },
            "frame": { "x": 20, "y": 60, "w": 320, "h": 28 }
          }
        ]
      }
    }
  ]
}`, LoadOptions{
		Data: store,
	})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	win := doc.PrimaryWindow()
	if win == nil {
		t.Fatal("PrimaryWindow() returned nil")
	}
	if win.Meta.Title != "Initial Title" {
		t.Fatalf("win.Meta.Title = %q, want %q", win.Meta.Title, "Initial Title")
	}

	title := findTextWidget(t, win, "title")
	body := findTextWidget(t, win, "body")
	if title != "Initial Title" {
		t.Fatalf("title text = %q, want %q", title, "Initial Title")
	}
	if body != "Initial Body" {
		t.Fatalf("body text = %q, want %q", body, "Initial Body")
	}

	store.Patch(map[string]any{
		"page.title": "Updated Title",
		"page.body":  "Updated Body",
	})

	if win.Meta.Title != "Updated Title" {
		t.Fatalf("updated win.Meta.Title = %q, want %q", win.Meta.Title, "Updated Title")
	}
	if got := findTextWidget(t, win, "title"); got != "Updated Title" {
		t.Fatalf("updated title text = %q, want %q", got, "Updated Title")
	}
	if got := findTextWidget(t, win, "body"); got != "Updated Body" {
		t.Fatalf("updated body text = %q, want %q", got, "Updated Body")
	}
}

func findTextWidget(t *testing.T, win *Window, id string) string {
	t.Helper()

	widget := win.FindWidget(id)
	switch typed := widget.(type) {
	case *widgets.Label:
		return typed.Text
	case *widgets.Button:
		return typed.Text
	case *widgets.CheckBox:
		return typed.Text
	case *widgets.RadioButton:
		return typed.Text
	default:
		t.Fatalf("widget %q type = %T, does not expose text", id, widget)
		return ""
	}
}

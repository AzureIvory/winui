//go:build windows

package jsonui

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestWindowHostReplaceReusesBoundData(t *testing.T) {
	store := NewStore(map[string]any{
		"page": map[string]any{
			"title": "First Title",
		},
	})

	firstDoc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "col",
        "children": [
          { "type": "label", "id": "title", "text": { "bind": "page.title" } }
        ]
      }
    }
  ]
}`, LoadOptions{Data: store})
	if err != nil {
		t.Fatalf("LoadDocumentString(first) returned error: %v", err)
	}

	secondDoc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "col",
        "children": [
          { "type": "label", "id": "title2", "text": { "bind": "page.title" } }
        ]
      }
    }
  ]
}`, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadDocumentString(second) returned error: %v", err)
	}

	scene := widgets.NewScene(nil)
	scene.Root().SetBounds(widgets.Rect{W: 320, H: 200})

	host, err := MountWindow(scene, firstDoc.PrimaryWindow())
	if err != nil {
		t.Fatalf("MountWindow returned error: %v", err)
	}
	if got := len(scene.Root().Children()); got != 1 {
		t.Fatalf("len(scene.Root().Children()) = %d, want 1", got)
	}

	if err := host.ReplaceWindow(secondDoc.PrimaryWindow()); err != nil {
		t.Fatalf("ReplaceWindow returned error: %v", err)
	}
	if got := len(scene.Root().Children()); got != 1 {
		t.Fatalf("len(scene.Root().Children()) after replace = %d, want 1", got)
	}
	if host.Window() != secondDoc.PrimaryWindow() {
		t.Fatal("host.Window() was not updated to the replacement window")
	}

	store.Set("page.title", "Updated Title")

	title, ok := host.Window().FindWidget("title2").(*widgets.Label)
	if !ok {
		t.Fatalf("replacement title widget type = %T, want *widgets.Label", host.Window().FindWidget("title2"))
	}
	if title.Text != "Updated Title" {
		t.Fatalf("title.Text = %q, want %q", title.Text, "Updated Title")
	}

	if err := host.Detach(); err != nil {
		t.Fatalf("Detach returned error: %v", err)
	}
	if got := len(scene.Root().Children()); got != 0 {
		t.Fatalf("len(scene.Root().Children()) after detach = %d, want 0", got)
	}
}

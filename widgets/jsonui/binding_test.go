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

func TestLoadDocumentStringBoolDefaultsMatchWidgetSemantics(t *testing.T) {
	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": { "type": "col", "gap": 8 },
        "children": [
          {
            "type": "button",
            "id": "action",
            "text": "Run"
          },
          {
            "type": "checkbox",
            "id": "check",
            "text": "Unchecked by default"
          },
          {
            "type": "radio",
            "id": "radioA",
            "group": "mode",
            "text": "A",
            "checked": true
          },
          {
            "type": "radio",
            "id": "radioB",
            "group": "mode",
            "text": "B"
          }
        ]
      }
    }
  ]
}`, LoadOptions{})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	win := doc.PrimaryWindow()
	if win == nil {
		t.Fatal("PrimaryWindow() returned nil")
	}

	button, ok := win.FindWidget("action").(*widgets.Button)
	if !ok {
		t.Fatalf("action type = %T, want *widgets.Button", win.FindWidget("action"))
	}
	if !button.Visible() {
		t.Fatal("button without visible should default to visible")
	}
	if !button.Enabled() {
		t.Fatal("button without enabled should default to enabled")
	}

	check, ok := win.FindWidget("check").(*widgets.CheckBox)
	if !ok {
		t.Fatalf("check type = %T, want *widgets.CheckBox", win.FindWidget("check"))
	}
	if check.IsChecked() {
		t.Fatal("checkbox without checked should default to unchecked")
	}

	radioA, ok := win.FindWidget("radioA").(*widgets.RadioButton)
	if !ok {
		t.Fatalf("radioA type = %T, want *widgets.RadioButton", win.FindWidget("radioA"))
	}
	radioB, ok := win.FindWidget("radioB").(*widgets.RadioButton)
	if !ok {
		t.Fatalf("radioB type = %T, want *widgets.RadioButton", win.FindWidget("radioB"))
	}
	if !radioA.IsChecked() {
		t.Fatal("radioA should keep explicit checked=true")
	}
	if radioB.IsChecked() {
		t.Fatal("radioB without checked should default to unchecked")
	}
}

func TestLoadDocumentStringBoolBindingsUseSemanticFallbacks(t *testing.T) {
	store := NewStore(map[string]any{
		"page": map[string]any{},
	})

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": { "type": "col", "gap": 8 },
        "children": [
          {
            "type": "button",
            "id": "action",
            "text": "Run",
            "visible": { "bind": "page.visible" },
            "enabled": { "bind": "page.enabled" }
          },
          {
            "type": "checkbox",
            "id": "check",
            "text": "Unchecked by default",
            "checked": { "bind": "page.checked" }
          },
          {
            "type": "scrollview",
            "id": "scroll",
            "verticalScroll": { "bind": "page.vertical" },
            "horizontalScroll": { "bind": "page.horizontal" },
            "children": [
              {
                "type": "label",
                "id": "inside",
                "text": "content"
              }
            ]
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

	button, ok := win.FindWidget("action").(*widgets.Button)
	if !ok {
		t.Fatalf("action type = %T, want *widgets.Button", win.FindWidget("action"))
	}
	if !button.Visible() {
		t.Fatal("button visible binding without value should stay visible")
	}
	if !button.Enabled() {
		t.Fatal("button enabled binding without value should stay enabled")
	}

	check, ok := win.FindWidget("check").(*widgets.CheckBox)
	if !ok {
		t.Fatalf("check type = %T, want *widgets.CheckBox", win.FindWidget("check"))
	}
	if check.IsChecked() {
		t.Fatal("checkbox checked binding without value should stay unchecked")
	}

	scroll, ok := win.FindWidget("scroll").(*widgets.ScrollView)
	if !ok {
		t.Fatalf("scroll type = %T, want *widgets.ScrollView", win.FindWidget("scroll"))
	}
	if !scroll.VerticalScroll() {
		t.Fatal("verticalScroll binding without value should default to true")
	}
	if scroll.HorizontalScroll() {
		t.Fatal("horizontalScroll binding without value should default to false")
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

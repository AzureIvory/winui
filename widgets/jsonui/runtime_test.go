//go:build windows

package jsonui

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestLoadDocumentStringConfiguresEditBoxCapabilities(t *testing.T) {
	store := NewStore(map[string]any{
		"form": map[string]any{
			"nameReadOnly": false,
			"logMultiline": true,
			"logReadOnly":  false,
		},
	})

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "col",
        "children": [
          {
            "type": "input",
            "id": "name",
            "value": "Azure",
            "readOnly": { "bind": "form.nameReadOnly" }
          },
          {
            "type": "input",
            "id": "log",
            "value": "Line 1\nLine 2",
            "multiline": { "bind": "form.logMultiline" },
            "readOnly": { "bind": "form.logReadOnly" },
            "wordWrap": false,
            "horizontalScroll": true,
            "acceptReturn": false
          },
          {
            "type": "textarea",
            "id": "notes",
            "value": "Read only notes",
            "readOnly": true
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

	nameWidget := win.FindWidget("name")
	name, ok := nameWidget.(*widgets.EditBox)
	if !ok {
		t.Fatalf("name widget type = %T, want *widgets.EditBox", nameWidget)
	}
	if name.ReadOnly {
		t.Fatal("name.ReadOnly = true, want false")
	}
	if name.Multiline() {
		t.Fatal("name.Multiline() = true, want false")
	}

	logWidget := win.FindWidget("log")
	logBox, ok := logWidget.(*widgets.EditBox)
	if !ok {
		t.Fatalf("log widget type = %T, want *widgets.EditBox", logWidget)
	}
	if !logBox.Multiline() {
		t.Fatal("log.Multiline() = false, want true")
	}
	if logBox.WordWrap() {
		t.Fatal("log.WordWrap() = true, want false")
	}
	if !logBox.HorizontalScroll() {
		t.Fatal("log.HorizontalScroll() = false, want true")
	}
	if logBox.AcceptReturn() {
		t.Fatal("log.AcceptReturn() = true, want false")
	}
	if logBox.ReadOnly {
		t.Fatal("log.ReadOnly = true, want false")
	}

	notesWidget := win.FindWidget("notes")
	notes, ok := notesWidget.(*widgets.EditBox)
	if !ok {
		t.Fatalf("notes widget type = %T, want *widgets.EditBox", notesWidget)
	}
	if !notes.Multiline() {
		t.Fatal("notes.Multiline() = false, want true")
	}
	if !notes.ReadOnly {
		t.Fatal("notes.ReadOnly = false, want true")
	}

	store.Patch(map[string]any{
		"form.nameReadOnly": true,
		"form.logMultiline": false,
		"form.logReadOnly":  true,
	})

	if !name.ReadOnly {
		t.Fatal("updated name.ReadOnly = false, want true")
	}
	if logBox.Multiline() {
		t.Fatal("updated log.Multiline() = true, want false")
	}
	if !logBox.ReadOnly {
		t.Fatal("updated log.ReadOnly = false, want true")
	}
}

func TestWindowLookupHelpersAndActionContext(t *testing.T) {
	var captured ActionContext

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "col",
        "children": [
          {
            "type": "button",
            "id": "saveBtn",
            "text": "Save",
            "onClick": "save"
          }
        ]
      }
    }
  ]
}`, LoadOptions{
		ActionHandlers: map[string]func(ActionContext){
			"save": func(ctx ActionContext) {
				captured = ctx
			},
		},
	})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	win := doc.PrimaryWindow()
	if win == nil {
		t.Fatal("PrimaryWindow() returned nil")
	}

	widget := win.FindWidget("saveBtn")
	button, ok := widget.(*widgets.Button)
	if !ok {
		t.Fatalf("saveBtn type = %T, want *widgets.Button", widget)
	}
	if got := doc.FindWidget("main", "saveBtn"); got != button {
		t.Fatalf("doc.FindWidget(main, saveBtn) = %T(%v), want button", got, got)
	}
	if handled := button.OnEvent(widgets.Event{Type: widgets.EventClick}); !handled {
		t.Fatal("button click event was not handled")
	}

	if captured.Name != "save" {
		t.Fatalf("captured.Name = %q, want save", captured.Name)
	}
	if captured.Window != win {
		t.Fatalf("captured.Window = %p, want %p", captured.Window, win)
	}
	if captured.Widget != button {
		t.Fatalf("captured.Widget = %p, want %p", captured.Widget, button)
	}
	if captured.ID != "saveBtn" {
		t.Fatalf("captured.ID = %q, want saveBtn", captured.ID)
	}
}

func TestResolveIconLoadSizeScalesLogicalPixels(t *testing.T) {
	tests := []struct {
		name   string
		sizeDP int32
		scale  float64
		wantPx int32
	}{
		{name: "default size", sizeDP: 0, scale: 1, wantPx: 32},
		{name: "fractional scale", sizeDP: 32, scale: 1.5, wantPx: 48},
		{name: "large scale", sizeDP: 32, scale: 2, wantPx: 64},
		{name: "custom size", sizeDP: 20, scale: 1.25, wantPx: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveIconLoadSize(tt.sizeDP, tt.scale); got != tt.wantPx {
				t.Fatalf("resolveIconLoadSize(%d, %v) = %d, want %d", tt.sizeDP, tt.scale, got, tt.wantPx)
			}
		})
	}
}

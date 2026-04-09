//go:build windows

package jsonui

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestLoadDocumentStringBuildsModalAndDispatchesDismiss(t *testing.T) {
	var dismissed ActionContext

	doc, err := LoadDocumentString(`{
  "wins": [
    {
      "id": "main",
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "modal",
            "id": "dialog",
            "backdrop": {
              "color": "#000000",
              "opacity": 96,
              "blur": 8,
              "dismissOnClick": true
            },
            "onDismiss": "dismissDialog",
            "children": [
              {
                "type": "panel",
                "id": "card",
                "frame": { "x": 20, "y": 20, "w": 120, "h": 80 }
              }
            ]
          }
        ]
      }
    }
  ]
}`, LoadOptions{
		ActionHandlers: map[string]func(ActionContext){
			"dismissDialog": func(ctx ActionContext) {
				dismissed = ctx
			},
		},
	})
	if err != nil {
		t.Fatalf("LoadDocumentString returned error: %v", err)
	}

	win := doc.PrimaryWindow()
	modalWidget := win.FindWidget("dialog")
	modal, ok := modalWidget.(*widgets.Modal)
	if !ok {
		t.Fatalf("modal widget type = %T, want *widgets.Modal", modalWidget)
	}
	if modal.BackdropOpacity() != 96 {
		t.Fatalf("modal.BackdropOpacity() = %d, want 96", modal.BackdropOpacity())
	}
	if modal.BlurRadiusDP() != 8 {
		t.Fatalf("modal.BlurRadiusDP() = %d, want 8", modal.BlurRadiusDP())
	}
	if !modal.DismissOnBackdrop() {
		t.Fatal("modal.DismissOnBackdrop() = false, want true")
	}

	if handled := modal.OnEvent(widgets.Event{Type: widgets.EventClick, Source: modal}); !handled {
		t.Fatal("modal dismiss click was not handled")
	}
	if dismissed.Name != "dismissDialog" {
		t.Fatalf("dismissed.Name = %q, want dismissDialog", dismissed.Name)
	}
	if dismissed.Widget != modal {
		t.Fatalf("dismissed.Widget = %T, want modal", dismissed.Widget)
	}
}

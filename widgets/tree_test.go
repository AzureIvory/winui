//go:build windows

package widgets

import "testing"

func TestFindByIDWalksWidgetTree(t *testing.T) {
	root := NewPanel("root")
	form := NewPanel("form")
	title := NewLabel("title", "JSON Demo")
	field := NewEditBox("field", ModeCustom)

	form.AddAll(title, field)
	root.Add(form)

	if got := FindByID(root, "title"); got != title {
		t.Fatalf("FindByID(title) = %T(%v), want title label", got, got)
	}

	typed, ok := FindByIDAs[*EditBox](root, "field")
	if !ok {
		t.Fatal("FindByIDAs(field) returned ok=false")
	}
	if typed != field {
		t.Fatalf("FindByIDAs(field) = %p, want %p", typed, field)
	}

	if got := FindByID(root, "missing"); got != nil {
		t.Fatalf("FindByID(missing) = %T, want nil", got)
	}
}

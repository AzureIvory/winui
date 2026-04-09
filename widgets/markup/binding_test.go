//go:build windows

package markup

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestDocumentBindingsRefreshTextVisibilityAndWindowTitle(t *testing.T) {
	t.Parallel()

	state := NewState(map[string]any{
		"page": map[string]any{
			"title":   "Initial Title",
			"visible": false,
		},
	})

	doc, err := LoadDocumentString(`
<window bind-title="page.title">
  <body>
    <label id="title" bind-text="page.title" bind-visible="page.visible"></label>
  </body>
</window>
`, "", LoadOptions{State: state})
	if err != nil {
		t.Fatalf("LoadDocumentString() error = %v", err)
	}

	if got := doc.Meta.Title; got != "Initial Title" {
		t.Fatalf("doc.Meta.Title = %q, want %q", got, "Initial Title")
	}

	label, ok := findWidgetByID(doc.Root, "title").(*widgets.Label)
	if !ok {
		t.Fatalf("expected label widget with id %q", "title")
	}
	if got := label.Text; got != "Initial Title" {
		t.Fatalf("label.Text = %q, want %q", got, "Initial Title")
	}
	if label.Visible() {
		t.Fatalf("label.Visible() = true, want false")
	}

	state.Set("page.title", "Updated Title")
	state.Set("page.visible", true)

	if got := doc.Meta.Title; got != "Updated Title" {
		t.Fatalf("doc.Meta.Title = %q, want %q after update", got, "Updated Title")
	}
	if got := label.Text; got != "Updated Title" {
		t.Fatalf("label.Text = %q, want %q after update", got, "Updated Title")
	}
	if !label.Visible() {
		t.Fatalf("label.Visible() = false, want true")
	}
}

func TestDocumentBindingsRefreshAbsoluteLayoutAndPreferredSize(t *testing.T) {
	t.Parallel()

	state := NewState(map[string]any{
		"box": map[string]any{
			"text":   "Alpha",
			"x":      10,
			"y":      20,
			"width":  160,
			"height": 40,
		},
	})

	doc, err := LoadDocumentString(`
<body style="display:absolute">
  <label
    id="box"
    bind-text="box.text"
    bind-left="box.x"
    bind-top="box.y"
    bind-width="box.width"
    bind-height="box.height"></label>
</body>
`, "", LoadOptions{State: state})
	if err != nil {
		t.Fatalf("LoadDocumentString() error = %v", err)
	}

	scene := widgets.NewScene(nil)
	scene.Resize(widgets.Rect{W: 800, H: 600})
	if err := doc.Attach(scene); err != nil {
		t.Fatalf("doc.Attach() error = %v", err)
	}

	label, ok := findWidgetByID(doc.Root, "box").(*widgets.Label)
	if !ok {
		t.Fatalf("expected label widget with id %q", "box")
	}

	assertRect(t, label.Bounds(), widgets.Rect{X: 10, Y: 20, W: 160, H: 40})

	data, ok := label.LayoutData().(widgets.AbsoluteLayoutData)
	if !ok {
		t.Fatalf("label.LayoutData() type = %T, want widgets.AbsoluteLayoutData", label.LayoutData())
	}
	if !data.HasWidth || !data.HasHeight {
		t.Fatalf("absolute layout data missing width/height flags: %+v", data)
	}

	state.Set("box.text", "Beta")
	state.Set("box.x", 32)
	state.Set("box.y", 48)
	state.Set("box.width", 220)
	state.Set("box.height", 60)

	if got := label.Text; got != "Beta" {
		t.Fatalf("label.Text = %q, want %q after update", got, "Beta")
	}
	assertRect(t, label.Bounds(), widgets.Rect{X: 32, Y: 48, W: 220, H: 60})
}

func TestDocumentBindingsRefreshListItemsAndSelection(t *testing.T) {
	t.Parallel()

	state := NewState(map[string]any{
		"users": []map[string]any{
			{"id": "u1", "name": "Alice"},
			{"id": "u2", "name": "Bob"},
		},
		"selected_user": "u2",
	})

	doc, err := LoadDocumentString(`
<body>
  <listbox
    id="users"
    bind-items="users"
    bind-selected="selected_user"
    item-text-field="name"
    item-value-field="id"></listbox>
</body>
`, "", LoadOptions{State: state})
	if err != nil {
		t.Fatalf("LoadDocumentString() error = %v", err)
	}

	list, ok := findWidgetByID(doc.Root, "users").(*widgets.ListBox)
	if !ok {
		t.Fatalf("expected listbox widget with id %q", "users")
	}

	assertListItems(t, list.Items(), []widgets.ListItem{
		{Value: "u1", Text: "Alice"},
		{Value: "u2", Text: "Bob"},
	})
	if got := list.SelectedIndex(); got != 1 {
		t.Fatalf("list.SelectedIndex() = %d, want 1", got)
	}

	state.Set("users", []map[string]any{
		{"id": "u3", "name": "Carol"},
		{"id": "u4", "name": "Dora"},
	})
	state.Set("selected_user", "u4")

	assertListItems(t, list.Items(), []widgets.ListItem{
		{Value: "u3", Text: "Carol"},
		{Value: "u4", Text: "Dora"},
	})
	if got := list.SelectedIndex(); got != 1 {
		t.Fatalf("list.SelectedIndex() = %d, want 1 after update", got)
	}
}

func TestDocumentBindingsRefreshInputValueAndProgress(t *testing.T) {
	t.Parallel()

	state := NewState(map[string]any{
		"form": map[string]any{
			"query":    "alpha",
			"progress": 25,
		},
	})

	doc, err := LoadDocumentString(`
<body>
  <input id="query" bind-value="form.query" />
  <progress id="progress" bind-value="form.progress"></progress>
</body>
`, "", LoadOptions{State: state})
	if err != nil {
		t.Fatalf("LoadDocumentString() error = %v", err)
	}

	input, ok := findWidgetByID(doc.Root, "query").(*widgets.EditBox)
	if !ok {
		t.Fatalf("expected edit widget with id %q", "query")
	}
	progress, ok := findWidgetByID(doc.Root, "progress").(*widgets.ProgressBar)
	if !ok {
		t.Fatalf("expected progress widget with id %q", "progress")
	}

	if got := input.TextValue(); got != "alpha" {
		t.Fatalf("input.TextValue() = %q, want %q", got, "alpha")
	}
	if got := progress.Value(); got != 25 {
		t.Fatalf("progress.Value() = %d, want 25", got)
	}

	state.Set("form.query", "beta")
	state.Set("form.progress", 80)

	if got := input.TextValue(); got != "beta" {
		t.Fatalf("input.TextValue() = %q, want %q after update", got, "beta")
	}
	if got := progress.Value(); got != 80 {
		t.Fatalf("progress.Value() = %d, want 80 after update", got)
	}
}

func assertRect(t *testing.T, got widgets.Rect, want widgets.Rect) {
	t.Helper()
	if got != want {
		t.Fatalf("rect = %+v, want %+v", got, want)
	}
}

func assertListItems(t *testing.T, got []widgets.ListItem, want []widgets.ListItem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(items) = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("items[%d] = %+v, want %+v", index, got[index], want[index])
		}
	}
}

func findWidgetByID(root widgets.Widget, id string) widgets.Widget {
	if root == nil {
		return nil
	}
	if root.ID() == id {
		return root
	}
	container, ok := root.(widgets.Container)
	if !ok {
		return nil
	}
	for _, child := range container.Children() {
		if match := findWidgetByID(child, id); match != nil {
			return match
		}
	}
	return nil
}

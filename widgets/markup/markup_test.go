//go:build windows

package markup

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestCSSSpecificityMerge(t *testing.T) {
	doc, err := parseHTMLDocument(`<body><button id="submitBtn" class="primary" style="color:#444444">Save</button></body>`)
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}
	rules, err := parseCSS(`button { color: #111111; } .primary { color: #222222; } #submitBtn { color: #333333; }`)
	if err != nil {
		t.Fatalf("parse css: %v", err)
	}
	if err := applyCSS(doc, rules); err != nil {
		t.Fatalf("apply css: %v", err)
	}
	child := doc.elementChildren()[0]
	if got := child.Styles["color"]; got != "#444444" {
		t.Fatalf("expected inline style to win, got %q", got)
	}
}

func TestLoadHTMLStringBuildsWidgetTree(t *testing.T) {
	html := `
<body id="page">
  <row id="toolbar">
    <button id="saveBtn" onclick="save">Save</button>
  </row>
  <textarea id="notes">Line 1
Line 2</textarea>
  <scroll id="listHost">
    <column id="items">
      <label id="row1">One</label>
      <label id="row2">Two</label>
    </column>
  </scroll>
</body>`
	css := `#page { display: column; gap: 12px; } #toolbar { gap: 8px; }`
	clicked := 0
	root, err := LoadHTMLString(html, css, LoadOptions{Actions: map[string]func(){"save": func() { clicked++ }}})
	if err != nil {
		t.Fatalf("load html: %v", err)
	}
	page, ok := root.(*widgets.Panel)
	if !ok {
		t.Fatalf("expected root panel, got %T", root)
	}
	if _, ok := page.Layout().(widgets.ColumnLayout); !ok {
		t.Fatalf("expected body to map to ColumnLayout, got %T", page.Layout())
	}
	children := page.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 direct children, got %d", len(children))
	}
	toolbar, ok := children[0].(*widgets.Panel)
	if !ok {
		t.Fatalf("expected toolbar panel, got %T", children[0])
	}
	if _, ok := toolbar.Layout().(widgets.RowLayout); !ok {
		t.Fatalf("expected row tag to map to RowLayout, got %T", toolbar.Layout())
	}
	button, ok := toolbar.Children()[0].(*widgets.Button)
	if !ok {
		t.Fatalf("expected button in toolbar, got %T", toolbar.Children()[0])
	}
	button.OnEvent(widgets.Event{Type: widgets.EventClick})
	if clicked != 1 {
		t.Fatalf("expected onclick action to run once, got %d", clicked)
	}
	textarea, ok := children[1].(*widgets.EditBox)
	if !ok {
		t.Fatalf("expected textarea to map to EditBox, got %T", children[1])
	}
	if !textarea.Multiline() {
		t.Fatalf("expected textarea EditBox to be multiline")
	}
	if textarea.TextValue() != "Line 1\nLine 2" {
		t.Fatalf("expected textarea text to preserve newlines, got %q", textarea.TextValue())
	}
	scroll, ok := children[2].(*widgets.ScrollView)
	if !ok {
		t.Fatalf("expected scroll tag to map to ScrollView, got %T", children[2])
	}
	content, ok := scroll.Content().(*widgets.Panel)
	if !ok {
		t.Fatalf("expected scroll content to be a Panel, got %T", scroll.Content())
	}
	if _, ok := content.Layout().(widgets.ColumnLayout); !ok {
		t.Fatalf("expected scroll content wrapper to use ColumnLayout, got %T", content.Layout())
	}
}

func TestTextareaScrollButtonRowColumnMapping(t *testing.T) {
	html := `
<body>
  <div id="shell" style="display: column; gap: 10px;">
    <div id="toolbar" style="display: row; gap: 6px;">
      <button id="go">Go</button>
    </div>
    <textarea id="memo" style="overflow-y: scroll;">memo</textarea>
    <div id="wrapped" style="overflow-y: auto; height: 80px;">
      <label>First</label>
      <label>Second</label>
    </div>
  </div>
</body>`
	root, err := LoadHTMLString(html, "", LoadOptions{})
	if err != nil {
		t.Fatalf("load html: %v", err)
	}
	page := root.(*widgets.Panel)
	shell, ok := page.Children()[0].(*widgets.Panel)
	if !ok {
		t.Fatalf("expected shell panel, got %T", page.Children()[0])
	}
	if _, ok := shell.Layout().(widgets.ColumnLayout); !ok {
		t.Fatalf("expected shell to use ColumnLayout, got %T", shell.Layout())
	}
	toolbar, ok := shell.Children()[0].(*widgets.Panel)
	if !ok {
		t.Fatalf("expected toolbar panel, got %T", shell.Children()[0])
	}
	if _, ok := toolbar.Layout().(widgets.RowLayout); !ok {
		t.Fatalf("expected toolbar to use RowLayout, got %T", toolbar.Layout())
	}
	if _, ok := toolbar.Children()[0].(*widgets.Button); !ok {
		t.Fatalf("expected toolbar child to be Button, got %T", toolbar.Children()[0])
	}
	memo, ok := shell.Children()[1].(*widgets.EditBox)
	if !ok {
		t.Fatalf("expected textarea to map to EditBox, got %T", shell.Children()[1])
	}
	if !memo.Multiline() {
		t.Fatalf("expected textarea to be multiline")
	}
	wrapped, ok := shell.Children()[2].(*widgets.ScrollView)
	if !ok {
		t.Fatalf("expected overflow container to map to ScrollView, got %T", shell.Children()[2])
	}
	if !wrapped.VerticalScroll() {
		t.Fatalf("expected overflow-y:auto to enable vertical scrolling")
	}
}

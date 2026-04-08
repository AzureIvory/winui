//go:build windows

package markup

import (
	"encoding/binary"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AzureIvory/winui/core"
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

func TestLoadDocumentWindowMetaAndBodyCompat(t *testing.T) {
	dir := t.TempDir()
	iconPath := filepath.Join(dir, "app.ico")
	if err := os.WriteFile(iconPath, buildICOForTest(), 0o600); err != nil {
		t.Fatalf("write icon: %v", err)
	}
	htmlPath := filepath.Join(dir, "demo.ui.html")
	html := `<window title="Markup Demo" icon="app.ico" min-width="900" min-height="640"><body><label id="title">ok</label></body></window>`
	if err := os.WriteFile(htmlPath, []byte(html), 0o600); err != nil {
		t.Fatalf("write html: %v", err)
	}
	doc, err := LoadDocumentFile(htmlPath, LoadOptions{AssetsDir: dir})
	if err != nil {
		t.Fatalf("load document: %v", err)
	}
	if doc.Meta.Title != "Markup Demo" {
		t.Fatalf("unexpected title: %q", doc.Meta.Title)
	}
	if doc.Meta.Icon == nil {
		t.Fatalf("expected window icon to be loaded")
	}
	if doc.Meta.IconPath != "app.ico" {
		t.Fatalf("unexpected icon path: %q", doc.Meta.IconPath)
	}
	if doc.Meta.MinWidth != 900 || doc.Meta.MinHeight != 640 {
		t.Fatalf("unexpected min size: %+v", doc.Meta)
	}
	root, err := LoadHTMLString(`<body id="plain"><label>ok</label></body>`, "", LoadOptions{})
	if err != nil {
		t.Fatalf("load old body root: %v", err)
	}
	if _, ok := root.(*widgets.Panel); !ok {
		t.Fatalf("expected old body root to stay compatible, got %T", root)
	}
}

func TestPasswordAbsoluteListBoxAnimatedAndButtonIcon(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "save.ico"), buildICOForTest(), 0o600); err != nil {
		t.Fatalf("write icon: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spinner.gif"), tinyGIFData(), 0o600); err != nil {
		t.Fatalf("write gif: %v", err)
	}
	html := `<body style="display:absolute;">
	<input id="pwd" type="password" style="left:20px; top:20px; width:220px; height:36px;" />
	<button id="saveBtn" icon="save.ico" icon-position="left" style="left:260px; top:20px; width:140px; height:40px;">保存</button>
	<listbox id="cityList" style="left:20px; top:80px; width:220px; height:160px;">
		<option value="bj">北京</option>
		<option value="sh" selected>上海</option>
	</listbox>
	<animated-img id="spinner" src="spinner.gif" autoplay="true" style="left:260px; top:90px; width:48px; height:48px;" />
</body>`
	root, err := LoadHTMLString(html, "", LoadOptions{AssetsDir: dir})
	if err != nil {
		t.Fatalf("load html: %v", err)
	}
	pwd, ok := findWidgetByID(root, "pwd").(*widgets.EditBox)
	if !ok {
		t.Fatalf("pwd not mapped to EditBox")
	}
	if !pwd.Password() {
		t.Fatalf("password input should enable password mode")
	}
	if bounds := pwd.Bounds(); bounds.X != 20 || bounds.Y != 20 || bounds.W != 220 || bounds.H != 36 {
		t.Fatalf("unexpected pwd bounds: %+v", bounds)
	}
	btn, ok := findWidgetByID(root, "saveBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("saveBtn not mapped to Button")
	}
	if btn.Icon == nil {
		t.Fatalf("button icon should be loaded")
	}
	if btn.Kind() != widgets.BtnLeft {
		t.Fatalf("button icon-position should map to BtnLeft, got %v", btn.Kind())
	}
	list, ok := findWidgetByID(root, "cityList").(*widgets.ListBox)
	if !ok {
		t.Fatalf("cityList not mapped to ListBox")
	}
	if items := list.Items(); len(items) != 2 {
		t.Fatalf("unexpected listbox items: %d", len(items))
	}
	if item, ok := list.SelectedItem(); !ok || item.Value != "sh" {
		t.Fatalf("unexpected selected listbox item: %+v, ok=%v", item, ok)
	}
	if bounds := list.Bounds(); bounds.X != 20 || bounds.Y != 80 || bounds.W != 220 || bounds.H != 160 {
		t.Fatalf("unexpected listbox bounds: %+v", bounds)
	}
	animated, ok := findWidgetByID(root, "spinner").(*widgets.AnimatedImage)
	if !ok {
		t.Fatalf("spinner not mapped to AnimatedImage")
	}
	natural := animated.NaturalSize()
	if natural.Width <= 0 || natural.Height <= 0 {
		t.Fatalf("animated gif should be loaded, got natural size %+v", natural)
	}
	if bounds := animated.Bounds(); bounds.X != 260 || bounds.Y != 90 || bounds.W != 48 || bounds.H != 48 {
		t.Fatalf("unexpected animated-img bounds: %+v", bounds)
	}
}

func TestActionContextAndLegacyFallback(t *testing.T) {
	html := `<body>
	<input id="txt" value="init" onchange="inputChanged" onsubmit="inputSubmit" />
	<textarea id="memo" onchange="memoChanged" onsubmit="memoSubmit">memo</textarea>
	<select id="city" onchange="cityChanged"><option value="bj" selected>北京</option></select>
	<listbox id="cityList" onchange="listChanged" onactivate="listActivate"><option value="sh" selected>上海</option></listbox>
	<checkbox id="agree" onchange="agreeChanged">同意</checkbox>
	<radio id="plan" onchange="planChanged">方案A</radio>
	<button id="save" onclick="save">保存</button>
	<button id="legacy" onclick="legacyOnly">仅旧动作</button>
</body>`
	captured := map[string]ActionContext{}
	legacyCalls := map[string]int{}
	handlers := map[string]func(ActionContext){}
	for _, name := range []string{"inputChanged", "inputSubmit", "memoChanged", "memoSubmit", "cityChanged", "listChanged", "listActivate", "agreeChanged", "planChanged", "save"} {
		nameCopy := name
		handlers[nameCopy] = func(ctx ActionContext) {
			captured[nameCopy] = ctx
		}
	}
	actions := map[string]func(){}
	for _, name := range []string{"inputChanged", "save", "legacyOnly"} {
		nameCopy := name
		actions[nameCopy] = func() {
			legacyCalls[nameCopy]++
		}
	}
	root, err := LoadHTMLString(html, "", LoadOptions{ActionHandlers: handlers, Actions: actions})
	if err != nil {
		t.Fatalf("load html: %v", err)
	}
	input := findWidgetByID(root, "txt").(*widgets.EditBox)
	textarea := findWidgetByID(root, "memo").(*widgets.EditBox)
	combo := findWidgetByID(root, "city").(*widgets.ComboBox)
	list := findWidgetByID(root, "cityList").(*widgets.ListBox)
	agree := findWidgetByID(root, "agree").(*widgets.CheckBox)
	radio := findWidgetByID(root, "plan").(*widgets.RadioButton)
	save := findWidgetByID(root, "save").(*widgets.Button)
	legacy := findWidgetByID(root, "legacy").(*widgets.Button)

	input.SetText("abc")
	if input.OnChange != nil {
		input.OnChange("ignored")
	}
	if input.OnSubmit != nil {
		input.OnSubmit("ignored")
	}
	textarea.SetText("memo2")
	if textarea.OnChange != nil {
		textarea.OnChange("ignored")
	}
	if textarea.OnSubmit != nil {
		textarea.OnSubmit("ignored")
	}
	if combo.OnChange != nil {
		combo.OnChange(0, widgets.ListItem{Value: "bj", Text: "北京"})
	}
	if list.OnChange != nil {
		list.OnChange(0, widgets.ListItem{Value: "sh", Text: "上海"})
	}
	if list.OnActivate != nil {
		list.OnActivate(0, widgets.ListItem{Value: "sh", Text: "上海"})
	}
	if agree.OnChange != nil {
		agree.OnChange(true)
	}
	if radio.OnChange != nil {
		radio.OnChange(true)
	}
	if save.OnClick != nil {
		save.OnClick()
	}
	if legacy.OnClick != nil {
		legacy.OnClick()
	}

	ctx := captured["inputChanged"]
	if ctx.ID != "txt" || ctx.Value != "abc" || ctx.Widget != input {
		t.Fatalf("unexpected input context: %+v", ctx)
	}
	ctx = captured["cityChanged"]
	if ctx.ID != "city" || ctx.Index != 0 || ctx.Item.Value != "bj" || ctx.Value != "bj" {
		t.Fatalf("unexpected select context: %+v", ctx)
	}
	ctx = captured["listActivate"]
	if ctx.ID != "cityList" || ctx.Index != 0 || ctx.Item.Value != "sh" || ctx.Value != "sh" {
		t.Fatalf("unexpected listbox activate context: %+v", ctx)
	}
	ctx = captured["agreeChanged"]
	if ctx.ID != "agree" || !ctx.Checked {
		t.Fatalf("unexpected checkbox context: %+v", ctx)
	}
	ctx = captured["planChanged"]
	if ctx.ID != "plan" || !ctx.Checked {
		t.Fatalf("unexpected radio context: %+v", ctx)
	}
	ctx = captured["save"]
	if ctx.ID != "save" || ctx.Value != "保存" {
		t.Fatalf("unexpected button context: %+v", ctx)
	}
	if legacyCalls["inputChanged"] != 0 || legacyCalls["save"] != 0 {
		t.Fatalf("ActionHandlers should override legacy Actions, got %+v", legacyCalls)
	}
	if legacyCalls["legacyOnly"] != 1 {
		t.Fatalf("legacy action fallback should still work, got %+v", legacyCalls)
	}
}

func TestLoadIntoSceneAppliesTheme(t *testing.T) {
	scene := widgets.NewScene(nil)
	theme := widgets.DefaultTheme()
	theme.Button.TextColor = core.RGB(255, 0, 0)
	doc, err := LoadIntoScene(scene, `<body><button id="ok">OK</button></body>`, "", LoadOptions{Theme: theme})
	if err != nil {
		t.Fatalf("load into scene: %v", err)
	}
	if doc == nil || doc.Root == nil {
		t.Fatalf("expected document root")
	}
	if scene.Theme() != theme {
		t.Fatalf("expected scene theme to be applied")
	}
	if got := len(scene.Root().Children()); got != 1 {
		t.Fatalf("expected 1 root child after attach, got %d", got)
	}
}

func TestAbsoluteLayoutRightBottomReturnsError(t *testing.T) {
	_, err := LoadHTMLString(`<body style="display:absolute;"><label id="x" style="right:10px; top:10px; width:80px; height:20px;">x</label></body>`, "", LoadOptions{})
	if err == nil {
		t.Fatalf("expected right/bottom unsupported error")
	}
	if !strings.Contains(err.Error(), "does not support right") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func findWidgetByID(root widgets.Widget, id string) widgets.Widget {
	if root == nil || id == "" {
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
		if found := findWidgetByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

func buildICOForTest() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 40, G: 120, B: 220, A: 255})
		}
	}
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	maskStride := ((w + 31) / 32) * 4
	maskSize := maskStride * h
	bmpSize := 40 + w*h*4 + maskSize

	data := make([]byte, 6+16+bmpSize)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint16(data[4:], 1)

	entry := data[6:22]
	entry[0] = byte(w)
	entry[1] = byte(h)
	binary.LittleEndian.PutUint16(entry[4:], 1)
	binary.LittleEndian.PutUint16(entry[6:], 32)
	binary.LittleEndian.PutUint32(entry[8:], uint32(bmpSize))
	binary.LittleEndian.PutUint32(entry[12:], 22)

	bmp := data[22:]
	binary.LittleEndian.PutUint32(bmp[0:], 40)
	binary.LittleEndian.PutUint32(bmp[4:], uint32(w))
	binary.LittleEndian.PutUint32(bmp[8:], uint32(h*2))
	binary.LittleEndian.PutUint16(bmp[12:], 1)
	binary.LittleEndian.PutUint16(bmp[14:], 32)
	binary.LittleEndian.PutUint32(bmp[20:], uint32(w*h*4))

	pixelOffset := 40
	index := 0
	for y := h - 1; y >= 0; y-- {
		row := img.Pix[y*img.Stride:]
		for x := 0; x < w; x++ {
			src := x * 4
			dst := pixelOffset + index*4
			data[22+dst] = row[src+2]
			data[22+dst+1] = row[src+1]
			data[22+dst+2] = row[src]
			data[22+dst+3] = row[src+3]
			index++
		}
	}
	return data
}

func tinyGIFData() []byte {
	return []byte{
		'G', 'I', 'F', '8', '9', 'a',
		0x01, 0x00, 0x01, 0x00,
		0x80, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0xFF, 0xFF, 0xFF,
		0x21, 0xF9, 0x04, 0x01, 0x0A, 0x00, 0x01, 0x00,
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0x02, 0x02, 0x4C, 0x01, 0x00,
		0x3B,
	}
}

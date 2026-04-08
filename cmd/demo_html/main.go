//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/markup"
)

var (
	demoDoc     *markup.Document
	demoRoot    widgets.Widget
	statusLabel *widgets.Label
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	assetsDir := filepath.Join(baseDir, "assets")
	if err := ensureDemoAssets(assetsDir); err != nil {
		panic(err)
	}

	var app *core.App
	actionHandlers := map[string]func(markup.ActionContext){
		"pwdChanged": func(ctx markup.ActionContext) { showActionStatus(app, "密码变化", ctx) },
		"pwdSubmit":  func(ctx markup.ActionContext) { showActionStatus(app, "密码提交", ctx) },
		"save":       func(ctx markup.ActionContext) { showActionStatus(app, "保存点击", ctx) },
		"cityChanged": func(ctx markup.ActionContext) {
			showActionStatus(app, "城市变化", ctx)
		},
		"cityOpen": func(ctx markup.ActionContext) {
			showActionStatus(app, "城市激活", ctx)
		},
	}
	legacyActions := map[string]func(){
		"legacyOnly": func() {
			if statusLabel != nil {
				statusLabel.SetText("legacyOnly: 使用旧版 Actions map[string]func() 回调")
			}
			if app != nil {
				app.SetTitle("Markup Demo - legacyOnly")
			}
		},
	}

	theme := demoTheme()
	doc, err := markup.LoadDocumentFile(filepath.Join(baseDir, "demo.ui.html"), markup.LoadOptions{
		ActionHandlers: actionHandlers,
		Actions:        legacyActions,
		AssetsDir:      baseDir,
		DefaultMode:    widgets.ModeCustom,
		Theme:          theme,
	})
	if err != nil {
		panic(err)
	}
	demoDoc = doc

	opts := core.Options{
		ClassName:      "WinUIMarkupDemo",
		Title:          "winui markup demo",
		Width:          900,
		Height:         640,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(242, 246, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	doc.ApplyWindowMeta(&opts)

	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate: func(createdApp *core.App, scene *widgets.Scene) error {
			app = createdApp
			if err := demoDoc.Attach(scene); err != nil {
				return err
			}
			demoRoot = demoDoc.Root
			if demoRoot != nil {
				demoRoot.SetBounds(widgets.Rect{W: app.ClientSize().Width, H: app.ClientSize().Height})
			}
			statusLabel, _ = findWidgetByID(demoRoot, "status").(*widgets.Label)
			showActionStatus(app, "初始化完成", markup.ActionContext{Name: "init", ID: "page", Index: -1})
			return nil
		},
		OnResize: func(_ *core.App, _ *widgets.Scene, size core.Size) {
			if demoRoot != nil {
				demoRoot.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
		},
		OnDestroy: func(_ *core.App, _ *widgets.Scene) {
			demoRoot = nil
			statusLabel = nil
			demoDoc = nil
		},
	})

	app, err = core.NewApp(opts)
	if err != nil {
		panic(err)
	}
	if err := app.Init(); err != nil {
		panic(err)
	}
	app.Run()
}

func showActionStatus(app *core.App, title string, ctx markup.ActionContext) {
	if ctx.Index == 0 && ctx.Item.Value == "" && ctx.Item.Text == "" {
		ctx.Index = -1
	}
	itemText := strings.TrimSpace(ctx.Item.Text)
	if itemText == "" {
		itemText = strings.TrimSpace(ctx.Item.Value)
	}
	if itemText == "" {
		itemText = "-"
	}
	valueText := strings.TrimSpace(ctx.Value)
	if valueText == "" {
		valueText = "-"
	}
	text := fmt.Sprintf("%s: action=%s id=%s value=%s checked=%v index=%d item=%s", title, ctx.Name, fallbackText(ctx.ID, "-"), valueText, ctx.Checked, ctx.Index, itemText)
	if statusLabel != nil {
		statusLabel.SetText(text)
	}
	if app != nil {
		app.SetTitle("Markup Demo - " + fallbackText(ctx.Name, title))
	}
}

func fallbackText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func demoTheme() *widgets.Theme {
	theme := widgets.DefaultTheme()
	theme.BackgroundColor = core.RGB(242, 246, 251)
	theme.Button.Background = core.RGB(245, 249, 255)
	theme.Button.Hover = core.RGB(227, 242, 255)
	theme.Button.Pressed = core.RGB(33, 118, 215)
	theme.Button.Border = core.RGB(155, 191, 232)
	theme.Edit.FocusBorder = core.RGB(48, 126, 223)
	theme.ListBox.ItemSelectedColor = core.RGB(210, 232, 255)
	theme.ListBox.ItemHoverColor = core.RGB(232, 243, 255)
	return theme
}

func ensureDemoAssets(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	files := map[string][]byte{
		"app.ico":     buildICO(color.RGBA{R: 52, G: 120, B: 220, A: 255}),
		"save.ico":    buildICO(color.RGBA{R: 50, G: 168, B: 97, A: 255}),
		"spinner.gif": tinyGIFData(),
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func buildICO(fill color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, fill)
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

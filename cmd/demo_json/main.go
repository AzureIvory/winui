//go:build windows

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	assetsDir := filepath.Join(baseDir, "assets")
	if err := ensureDemoAssets(assetsDir); err != nil {
		panic(err)
	}

	store := jsonui.NewStore(map[string]any{
		"overlay": map[string]any{
			"helpVisible": false,
		},
	})

	var demoWindow *jsonui.Window
	actionHandlers := map[string]func(jsonui.ActionContext){
		"pwdChanged": func(ctx jsonui.ActionContext) { showActionStatus("Password changed", ctx) },
		"pwdSubmit":  func(ctx jsonui.ActionContext) { showActionStatus("Password submitted", ctx) },
		"save":       func(ctx jsonui.ActionContext) { showActionStatus("Save button clicked", ctx) },
		"cityChanged": func(ctx jsonui.ActionContext) {
			showActionStatus("City changed", ctx)
		},
		"cityOpen": func(ctx jsonui.ActionContext) {
			showActionStatus("City activated", ctx)
		},
		"openPicked": func(ctx jsonui.ActionContext) {
			showActionStatus("Open dialog selected", ctx)
		},
		"savePicked": func(ctx jsonui.ActionContext) {
			showActionStatus("Save dialog selected", ctx)
		},
		"folderPicked": func(ctx jsonui.ActionContext) {
			showActionStatus("Folder dialog selected", ctx)
		},
		"multiPicked": func(ctx jsonui.ActionContext) {
			showActionStatus("Multi-file dialog selected", ctx)
		},
		"showHelpDialog": func(ctx jsonui.ActionContext) {
			store.Set("overlay.helpVisible", true)
			showActionStatus("Modal opened", ctx)
		},
		"dismissHelpDialog": func(ctx jsonui.ActionContext) {
			store.Set("overlay.helpVisible", false)
			showActionStatus("Modal dismissed", ctx)
		},
	}
	legacyActions := map[string]func(){
		"legacyOnly": func() {
			if demoWindow != nil {
				if statusLabel, ok := demoWindow.FindWidget("status").(*widgets.Label); ok {
					statusLabel.SetText("legacyOnly: using the older Actions map[string]func() callback")
				}
				if app := demoWindow.App(); app != nil {
					app.SetTitle("JSON Demo - legacyOnly")
				}
			}
		},
	}

	theme := demoTheme()
	doc, err := jsonui.LoadDocumentFile(filepath.Join(baseDir, "demo.ui.json"), jsonui.LoadOptions{
		ActionHandlers: actionHandlers,
		Actions:        legacyActions,
		AssetsDir:      baseDir,
		DefaultMode:    widgets.ModeCustom,
		Data:           store,
		Theme:          theme,
	})
	if err != nil {
		panic(err)
	}
	demoWindow = doc.PrimaryWindow()
	if demoWindow == nil {
		panic("demo window is nil")
	}

	opts := core.Options{
		ClassName:      "WinUIJSONDemo",
		Title:          "winui json demo",
		Width:          980,
		Height:         720,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(242, 246, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	demoWindow.ApplyOptions(&opts)

	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate: func(createdApp *core.App, scene *widgets.Scene) error {
			if err := demoWindow.Attach(scene); err != nil {
				return err
			}
			if demoWindow.Root != nil {
				size := createdApp.ClientSize()
				demoWindow.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
			showActionStatus("Ready", jsonui.ActionContext{
				Name:   "init",
				Window: demoWindow,
				ID:     "page",
				Index:  -1,
			})
			return nil
		},
		OnResize: func(_ *core.App, _ *widgets.Scene, size core.Size) {
			if demoWindow.Root != nil {
				demoWindow.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
		},
		OnDestroy: func(_ *core.App, _ *widgets.Scene) {
		},
	})

	app, err := core.NewApp(opts)
	if err != nil {
		panic(err)
	}
	if err := app.Init(); err != nil {
		panic(err)
	}
	app.Run()
}

func showActionStatus(title string, ctx jsonui.ActionContext) {
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
	pathsText := "-"
	if len(ctx.Paths) > 0 {
		pathsText = strings.Join(ctx.Paths, " | ")
	}
	text := fmt.Sprintf(
		"%s\naction=%s id=%s value=%s checked=%v index=%d item=%s\npaths=%s",
		title,
		ctx.Name,
		fallbackText(ctx.ID, "-"),
		valueText,
		ctx.Checked,
		ctx.Index,
		itemText,
		pathsText,
	)
	if ctx.Window != nil {
		if statusLabel, ok := ctx.Window.FindWidget("status").(*widgets.Label); ok {
			statusLabel.SetText(text)
		}
	}
	if ctx.Window != nil {
		if app := ctx.Window.App(); app != nil {
			app.SetTitle("JSON Demo - " + fallbackText(ctx.Name, title))
		}
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
		"app.png":     buildPNG(color.RGBA{R: 52, G: 120, B: 220, A: 255}),
		"save.png":    buildPNG(color.RGBA{R: 50, G: 168, B: 97, A: 255}),
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

func buildPNG(fill color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
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

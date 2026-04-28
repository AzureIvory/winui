//go:build windows

package main

import (
	"path/filepath"
	"runtime"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)

	doc, store, err := loadDemoDocument(baseDir)
	if err != nil {
		panic(err)
	}
	window := doc.PrimaryWindow()
	if window == nil {
		panic("primary window is nil")
	}

	controller := newDemoController(baseDir, store, doc, window, widgets.ModeCustom)

	opts := core.Options{
		ClassName:      "WinUIJSONFullDemo",
		Title:          "WinUI JSON Full Demo",
		Width:          1380,
		Height:         940,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(244, 247, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	window.ApplyOptions(&opts)

	widgets.BindScene(&opts, widgets.SceneHooks{
		Theme: demoTheme(),
		OnCreate: func(app *core.App, scene *widgets.Scene) error {
			if err := window.Attach(scene); err != nil {
				return err
			}
			if window.Root != nil {
				size := app.ClientSize()
				window.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
			controller.ensureSpinnerPlaying()
			controller.setStatus(controller.tr("status.uiReady", "UI ready"))
			return nil
		},
		OnResize: func(_ *core.App, _ *widgets.Scene, size core.Size) {
			if controller.window != nil && controller.window.Root != nil {
				controller.window.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
				controller.restartSpinnerAfterLayout()
			}
		},
		OnDPIChanged: func(_ *core.App, _ *widgets.Scene, _ core.DPIInfo) {
			if controller.window != nil {
				_ = controller.window.ReloadResources(jsonui.ReloadReasonDPIChanged)
			}
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

func loadDemoDocument(baseDir string) (*jsonui.Document, *jsonui.Store, error) {
	return loadDemoDocumentWithMode(baseDir, widgets.ModeCustom, nil)
}

func loadDemoDocumentWithMode(baseDir string, mode widgets.ControlMode, store *jsonui.Store) (*jsonui.Document, *jsonui.Store, error) {
	if store == nil {
		store = newDemoStore()
	}
	doc, err := jsonui.LoadDocumentFile(filepath.Join(baseDir, "demo.ui.json"), jsonui.LoadOptions{
		AssetsDir:   baseDir,
		DefaultMode: mode,
		Data:        store,
		Theme:       demoTheme(),
	})
	if err != nil {
		return nil, nil, err
	}
	return doc, store, nil
}

func newDemoStore() *jsonui.Store {
	return jsonui.NewStore(map[string]any{
		"demo": map[string]any{
			"windowTitle":          "WinUI JSON Full Demo",
			"paletteName":          "Ocean Blue",
			"report":               "",
			"reportSummary":        defaultReportSummary(),
			"reportPath":           defaultReportPath(),
			"lastAction":           "Ready",
			"modalVisible":         false,
			"showVerticalScroll":   true,
			"showHorizontalScroll": true,
		},
	})
}

func demoTheme() *widgets.Theme {
	theme := widgets.DefaultTheme()
	theme.BackgroundColor = core.RGB(244, 247, 251)
	theme.Text.Color = core.RGB(22, 31, 47)
	theme.Title.Color = core.RGB(15, 23, 42)
	theme.Button.Background = core.RGB(245, 249, 255)
	theme.Button.Hover = core.RGB(226, 240, 255)
	theme.Button.Pressed = core.RGB(37, 99, 235)
	theme.Button.Border = core.RGB(173, 201, 236)
	theme.Edit.FocusBorder = core.RGB(37, 99, 235)
	theme.ComboBox.FocusBorder = core.RGB(37, 99, 235)
	theme.ListBox.FocusBorder = core.RGB(37, 99, 235)
	return theme
}

func defaultReportSummary() string {
	return "No API check has been run yet."
}

func defaultReportPath() string {
	return "output\\latest-api-check.txt"
}

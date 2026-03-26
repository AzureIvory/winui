//go:build windows

package main

//go:generate rsrc -manifest main.manifest -o manifest.syso

import (
	"os"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

var ui demoUI

// main 启动并运行演示程序。
func main() {
	opts := core.Options{
		ClassName:      "WinUIDemo",
		Title:          "winui demo",
		Width:          760,
		Height:         560,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(248, 250, 252),
		DoubleBuffered: true,
		RenderMode:     demoRenderMode(),
	}
	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate:  buildDemoUI,
		OnResize:  resizeDemoUI,
		OnDestroy: destroyDemoUI,
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

// demoRenderMode 从环境变量读取演示程序期望使用的渲染模式。
func demoRenderMode() core.RenderMode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("WINUI_RENDER_MODE"))) {
	case "gdi":
		return core.RenderModeGDI
	default:
		return core.RenderModeAuto
	}
}

# winui

`winui` 是一个仅支持 Windows 的 Go UI 工具包，直接构建在 Win32 API 之上。

它面向小型原生桌面工具，适合希望完整掌控窗口循环、绘制、DPI 行为和控件渲染的场景，不依赖 WebView 或 XAML。

## 当前状态

- 平台：仅 Windows
- 语言：Go
- 渲染：自定义 Win32 绘制
- 控件模型：带自定义控件的保留式场景树

## 包说明

- `core`：窗口生命周期、绘制、DPI、定时器、图标、字体与输入
- `widgets`：场景树、主题、布局辅助和可复用控件

## 内置控件

- `Button`
- `CheckBox`
- `RadioButton`
- `ComboBox`
- `EditBox`
- `Image`
- `AnimatedImage`
- `Label`
- `ListBox`
- `Panel`
- `ProgressBar`
- `Scene`

## 快速开始

初始化应用、创建场景、添加控件，并将 Win32 事件转发给场景：

```go
package main

import (
	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

var scene *widgets.Scene

func main() {
	app, err := core.NewApp(core.Options{
		ClassName:      "ExampleApp",
		Title:          "winui example",
		Width:          800,
		Height:         600,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(255, 255, 255),
		DoubleBuffered: true,
		OnCreate: func(app *core.App) error {
			scene = widgets.NewScene(app)

			label := widgets.NewLabel("title", "Hello winui")
			label.SetBounds(core.Rect{X: 24, Y: 24, W: 240, H: 32})
			scene.Root().Add(label)
			return nil
		},
		OnPaint: func(_ *core.App, canvas *core.Canvas) {
			scene.PaintCore(canvas)
		},
		OnResize: func(_ *core.App, size core.Size) {
			scene.Resize(core.Rect{X: 0, Y: 0, W: size.Width, H: size.Height})
		},
		OnMouseMove: func(_ *core.App, ev core.MouseEvent) {
			scene.DispatchMouseMove(ev)
		},
		OnMouseLeave: func(_ *core.App) {
			scene.DispatchMouseLeave()
		},
		OnMouseDown: func(_ *core.App, ev core.MouseEvent) {
			scene.DispatchMouseDown(ev)
		},
		OnMouseUp: func(_ *core.App, ev core.MouseEvent) {
			scene.DispatchMouseUp(ev)
		},
		OnKeyDown: func(_ *core.App, ev core.KeyEvent) {
			scene.DispatchKeyDown(ev)
		},
		OnChar: func(_ *core.App, ch rune) {
			scene.DispatchChar(ch)
		},
		OnFocus: func(_ *core.App, focused bool) {
			if !focused && scene != nil {
				scene.Blur()
			}
		},
		OnTimer: func(_ *core.App, id uintptr) {
			scene.HandleTimer(id)
		},
	})
	if err != nil {
		panic(err)
	}

	if err := app.Init(); err != nil {
		panic(err)
	}
	app.Run()
}
```

## 示例

在模块根目录运行附带示例：

```powershell
go run ./cmd/demo
```

## 开发

开发流程和仓库约定见 [`DEVELOPING.md`](./DEVELOPING.md)。

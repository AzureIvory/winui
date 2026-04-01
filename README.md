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
- `widgets`：场景树、`BindScene` 接线辅助、主题、`Absolute/Linear/Row/Column/Grid/Form` 布局和可复用控件

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

初始化应用，使用 `widgets.BindScene` 自动接入场景的绘制、输入和生命周期回调：

```go
package main

import (
	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

func main() {
	opts := core.Options{
		ClassName:      "ExampleApp",
		Title:          "winui example",
		Width:          800,
		Height:         600,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(255, 255, 255),
		DoubleBuffered: true,
	}
	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate: func(_ *core.App, scene *widgets.Scene) error {
			label := widgets.NewLabel("title", "Hello winui")
			label.SetBounds(core.Rect{X: 24, Y: 24, W: 240, H: 32})
			scene.Root().Add(label)
			return nil
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
```

## 示例

在模块根目录运行附带示例：

```powershell
go run ./cmd/demo
```

## 文档

组件用法、`BindScene` 接法、布局系统、`LayoutData` 约定、构造参数和样式字段说明见 [`WIDGETS.zh-CN.md`](./WIDGETS.zh-CN.md)。

近期交互行为修复和后续 AI/代理协作所需的实现约束摘要见 [`AI_CHANGELOG.md`](./AI_CHANGELOG.md)。

`Button`、`EditBox`、`CheckBox`、`RadioButton`、`ComboBox` 在构造时支持通过 `mode` 参数切换自绘后端或原生系统 API 控件后端。

如果你要让 `ModeNative` 控件显示为带 Win10/Win11 visual styles 的系统控件，最终 `main` 可执行程序还需要嵌入 `Microsoft.Windows.Common-Controls` v6 manifest；只改库代码不够。

源码中的函数、自定义类型、结构体字段和常量也统一补充了中文注释，便于直接从代码阅读 API 与内部行为。

## 开发

开发流程和仓库约定见 [`DEVELOPING.md`](./DEVELOPING.md)。




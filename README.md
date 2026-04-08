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




## HTML 映射增强（`widgets/markup`）

当前版本补齐了 HTML -> WinUI 映射中的关键缺口，保持旧入口兼容：

- 新增文档级入口：
  - `markup.LoadDocumentFile(...)`
  - `markup.LoadDocumentString(...)`
  - 返回 `*markup.Document`，包含：
    - `Root widgets.Widget`
    - `Meta markup.WindowMeta`
- 旧入口仍可用且不破坏：
  - `markup.LoadHTMLFile(...)`
  - `markup.LoadHTMLString(...)`
  - 内部复用文档级入口，只返回 `Root`

### `<window>` 元数据

支持：

- `title`
- `icon`（仅本地 `.ico`，通过 `core.LoadIconFromICO`）
- `min-width`
- `min-height`

示例：

```html
<window title="Markup Demo" icon="assets/app.ico" min-width="900" min-height="640">
  <body>...</body>
</window>
```

应用方式：

- `doc.ApplyWindowMeta(&opts)`：把标题、图标、最小尺寸写入 `core.Options`
- `doc.Attach(scene)` 或 `markup.LoadIntoScene(...)`：挂载控件树并应用主题

### 交互与控件映射新增

- `input[type=password]` -> `widgets.EditBox` 密码模式（真实值仍由 `TextValue()` 返回）
- `display:absolute`：对子项生效 `left/top/width/height`（支持 `x/y` 别名）
  - `right/bottom` 目前明确报错，不做静默忽略
- `listbox` + `option` -> `widgets.ListBox`
  - 支持 `value`、`selected`、`onchange`、`onactivate`
- `animated-img` -> `widgets.AnimatedImage`
  - `src` 仅支持本地 `.gif`
  - `autoplay`、`object-fit` 生效
- `button` 新增图标属性：
  - `icon="..."`（仅本地 `.ico`）
  - `icon-position="left|top|auto"`

### 动作上下文（兼容旧回调）

- 保留：`LoadOptions.Actions map[string]func()`
- 新增：`LoadOptions.ActionHandlers map[string]func(markup.ActionContext)`
- 分发优先级：`ActionHandlers` > `Actions`
- `ActionContext` 提供：`Name`、`Widget`、`ID`、`Value`、`Checked`、`Index`、`Item`

### Theme 生效路径

`LoadOptions.Theme` 不再是占位字段。通过文档入口可真实应用：

- `doc.Attach(scene)`
- `markup.LoadIntoScene(scene, html, css, opts)`
- `markup.LoadFileIntoScene(scene, path, opts)`

旧 `LoadHTML*` 入口仍保持“只构建 Root、不直接操作 Scene”的语义。

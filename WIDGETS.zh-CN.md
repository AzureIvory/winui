# winui 组件文档

本文档面向 `github.com/AzureIvory/winui/widgets` 的公开组件 API，记录每个内置组件的用途、构造参数、常用方法和样式参数。

## 1. 快速接入

最小可运行接线方式如下，`widgets.BindScene` 会自动完成场景的绘制、输入、焦点、定时器、DPI 和销毁接线：

```go
package main

import (
	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

func main() {
	opts := core.Options{
		ClassName:      "ExampleApp",
		Title:          "winui demo",
		Width:          800,
		Height:         600,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(255, 255, 255),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate: func(_ *core.App, scene *widgets.Scene) error {
			label := widgets.NewLabel("title", "Hello winui")
			label.SetBounds(core.Rect{X: 20, Y: 20, W: 240, H: 32})
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

### 1.1 渲染后端选择

从当前版本开始，`core.Canvas` 支持双后端：

- `core.RenderModeAuto`
  - 默认值。
  - 优先尝试 `Direct2D + DirectWrite + WIC`。
  - 初始化失败或运行时失败时自动回退到 GDI。
- `core.RenderModeGDI`
  - 强制使用 GDI。
  - 适合需要兼容 WinPE、禁用 cgo 构建，或需要排查 Direct2D 环境问题时使用。

运行时可以通过以下接口查看实际结果：

- `app.RenderMode()`
  - 返回请求的模式。
- `app.RenderBackend()`
  - 返回当前实际激活的后端，值为 `GDI` 或 `Direct2D`。
- `app.RenderFallbackReason()`
  - 当 `Auto` 模式回退到 GDI 时，返回回退原因。

示例程序 `cmd/demo` 已经接入这个能力：

```powershell
go run ./cmd/demo
```

强制 GDI：

```powershell
$env:WINUI_RENDER_MODE='gdi'
go run ./cmd/demo
```

## 2. 通用约定

### 2.1 所有控件共有的概念

- `id string`
  - 控件标识。
  - 传空字符串时会自动生成，例如 `button-1`。
- `bounds core.Rect`
  - 控件边界，字段为 `X`、`Y`、`W`、`H`。
  - 坐标相对窗口客户区。
- `visible bool`
  - 是否可见。
- `enabled bool`
  - 是否可交互。

大多数组件都有以下通用方法：

- `SetBounds(rect core.Rect)`
- `SetVisible(visible bool)`
- `SetEnabled(enabled bool)`

### 2.2 列表项类型 `ListItem`

```go
type ListItem struct {
    Value    string
    Text     string
    Disabled bool
}
```

- `Value`
  - 实际值。
- `Text`
  - 显示文本。
  - 为空时回退为 `Value`。
- `Disabled`
  - 是否禁用该项。
  - 禁用项不会被 `ListBox` 或 `ComboBox` 选中。

### 2.3 样式覆盖的零值规则

当前组件样式合并逻辑会把很多字段的零值视为“未设置”：

- 颜色字段为 `0` 时，通常表示“不覆盖默认值”。
- 尺寸字段为 `0` 时，通常表示“沿用默认值”。

这意味着：

- `core.RGB(0, 0, 0)` 这样的纯黑颜色，作为覆盖值时可能会被视为“未设置”。
- 如果你需要完全控制颜色和尺寸，优先通过 `Theme` 先改默认样式，再做局部覆盖。

## 3. 场景与容器

### 3.1 `Scene`

用途：管理控件树、焦点、鼠标命中、重绘、主题和定时器。

构造函数：

```go
scene := widgets.NewScene(app)
```

参数：

- `app *core.App`
  - 已创建的应用实例。

常用方法：

- `Root() *Panel`
  - 返回根面板。
- `Theme() *Theme`
  - 读取当前主题。
- `SetTheme(theme *Theme)`
  - 替换主题并触发刷新。
- `ReloadResources()`
  - 重新创建字体等缓存资源。
- `Resize(bounds core.Rect)`
  - 调整场景边界，通常在 `OnResize` 中调用。
- `PaintCore(canvas *core.PaintCtx)`
  - 使用 `core.Canvas` 绘制整棵控件树。
- `DispatchMouseMove(ev core.MouseEvent) bool`
- `DispatchMouseLeave() bool`
- `DispatchMouseDown(ev core.MouseEvent) bool`
- `DispatchMouseUp(ev core.MouseEvent) bool`
- `DispatchMouseWheel(ev core.MouseEvent) bool`
- `DispatchKeyDown(ev core.KeyEvent) bool`
- `DispatchChar(ch rune) bool`
  - 把窗口输入事件转发给场景。
- `HandleTimer(timerID uintptr) bool`
  - 把窗口定时器事件转发给场景。
- `Focus() Widget`
  - 返回当前拥有键盘焦点的控件。
- `Blur()`
  - 清除焦点。
- `Invalidate(widget Widget)`
  - 刷新整个场景或某个控件。
- `Close() error`
  - 释放场景缓存的字体、定时器和可释放控件资源。

示例：

```go
scene := widgets.NewScene(app)
scene.Root().Add(widgets.NewLabel("title", "控制面板"))
```

### 3.2 `Panel`

用途：容器组件，用于承载子控件和组织布局。

构造函数：

```go
panel := widgets.NewPanel("content")
```

参数：

- `id string`
  - 面板 ID，空字符串时自动生成。

常用方法：

- `Add(child Widget)`
  - 添加子控件。
- `Remove(id string)`
  - 按控件 ID 删除子控件。
- `Children() []Widget`
  - 返回子控件切片副本。
- `SetLayout(layout Layout)`
  - 设置布局器。
- `SetStyle(style PanelStyle)`
  - 设置面板背景、边框、圆角等样式。
- `SetOnClick(fn func())`
  - 注册面板点击回调。

示例：

```go
panel := widgets.NewPanel("form")
panel.SetBounds(core.Rect{X: 20, Y: 20, W: 320, H: 240})
panel.Add(widgets.NewLabel("nameLabel", "名称"))
scene.Root().Add(panel)
```

### 3.3 `Layout`

#### `AbsoluteLayout`

用途：绝对布局，不自动调整子控件。

```go
panel.SetLayout(widgets.AbsoluteLayout{})
```

#### `LinearLayout`

用途：线性排列子控件。

```go
panel.SetLayout(widgets.LinearLayout{
    Axis:     widgets.AxisVertical,
    Gap:      8,
    Padding:  12,
    ItemSize: 36,
})
```

参数：

- `Axis`
  - `widgets.AxisVertical` 或 `widgets.AxisHorizontal`。
- `Gap int32`
  - 子项间距。
- `Padding int32`
  - 容器内边距。
- `ItemSize int32`
  - 固定项目高度或宽度。
  - 为 `0` 时按剩余空间自动分配。

## 4. 显示类组件

### 4.1 `Label`

用途：显示静态文本。

构造函数：

```go
label := widgets.NewLabel("title", "欢迎使用")
```

参数：

- `id string`
  - 标签 ID。
- `text string`
  - 显示文本。

常用方法：

- `SetText(text string)`
  - 更新文本。
- `SetStyle(style TextStyle)`
  - 覆盖文本样式。

示例：

```go
label := widgets.NewLabel("title", "欢迎使用")
label.SetBounds(core.Rect{X: 20, Y: 20, W: 240, H: 32})
label.SetStyle(widgets.TextStyle{
    Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 18, Weight: 700},
    Color: core.RGB(16, 16, 16),
    Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
})
scene.Root().Add(label)
```

### 4.2 `Button`

用途：点击触发动作，可显示文本和图标。

构造函数：

```go
btn := widgets.NewButton("save", "保存")
```

参数：

- `id string`
  - 按钮 ID。
- `text string`
  - 按钮标题。

常用方法：

- `SetText(text string)`
  - 更新按钮文本。
- `SetIcon(icon *core.Icon)`
  - 设置图标。
- `SetKind(kind widgets.BtnKind)`
  - 设置图标和文本布局。
  - `widgets.BtnAuto`：自动布局，有图标和文本时默认图标在上。
  - `widgets.BtnTop`：图标在上、文本在下。
  - `widgets.BtnLeft`：左侧小图标、右侧文本。
- `SetOnClick(fn func())`
  - 注册点击回调。
- `SetStyle(style ButtonStyle)`
  - 覆盖按钮样式。

示例：

```go
btn := widgets.NewButton("save", "保存")
btn.SetBounds(core.Rect{X: 20, Y: 70, W: 120, H: 44})
btn.SetKind(widgets.BtnLeft)
btn.SetOnClick(func() {
    app.SetTitle("已点击保存")
})
scene.Root().Add(btn)
```

### 4.3 `ProgressBar`

用途：显示 `0..100` 的进度。

构造函数：

```go
progress := widgets.NewProgressBar("progress")
```

参数：

- `id string`
  - 进度条 ID。

常用方法：

- `SetValue(value int32)`
  - 设置进度值，内部会自动限制到 `0..100`。
- `Value() int32`
  - 读取当前值。
- `SetStyle(style ProgressStyle)`
  - 覆盖样式。

行为说明：

- 当 `ShowPercent` 为 `true` 时，会在进度条上方绘制百分比气泡。
- demo 里额外演示了一个独立的百分比文本标签，方便测试文字更新。

示例：

```go
progress := widgets.NewProgressBar("progress")
progress.SetBounds(core.Rect{X: 20, Y: 130, W: 260, H: 18})
progress.SetValue(65)
scene.Root().Add(progress)
```

## 5. 选择类组件

### 5.1 `CheckBox`

用途：布尔开关。

构造函数：

```go
check := widgets.NewCheckBox("agree", "同意协议")
```

参数：

- `id string`
  - 复选框 ID。
- `text string`
  - 标题文本。

常用方法：

- `SetText(text string)`
- `SetChecked(checked bool)`
- `IsChecked() bool`
- `SetStyle(style ChoiceStyle)`
- `SetOnChange(fn func(bool))`

行为说明：

- 点击会在选中和未选中之间切换。
- 选中时会绘制居中的圆点标记，不再使用 `"X"` 字符。
- 支持键盘焦点。

示例：

```go
check := widgets.NewCheckBox("agree", "同意协议")
check.SetBounds(core.Rect{X: 20, Y: 170, W: 200, H: 32})
check.SetOnChange(func(v bool) {
    if v {
        app.SetTitle("已勾选")
    }
})
scene.Root().Add(check)
```

### 5.2 `RadioButton`

用途：互斥单选。

构造函数：

```go
radio := widgets.NewRadioButton("planA", "方案 A")
```

参数：

- `id string`
  - 单选按钮 ID。
- `text string`
  - 标题文本。

常用方法：

- `SetText(text string)`
- `SetGroup(group string)`
  - 设置分组名。
- `SetChecked(checked bool)`
- `IsChecked() bool`
- `SetStyle(style ChoiceStyle)`
- `SetOnChange(fn func(bool))`

行为说明：

- 同一父容器下、`Group` 相同的单选按钮互斥。
- 不在同一个 `Panel` 下时，即使分组名相同也不会互斥。

示例：

```go
radioA := widgets.NewRadioButton("planA", "方案 A")
radioB := widgets.NewRadioButton("planB", "方案 B")
radioA.SetGroup("plan")
radioB.SetGroup("plan")
radioA.SetChecked(true)
panel.Add(radioA)
panel.Add(radioB)
```

### 5.3 `ListBox`

用途：单选列表。

构造函数：

```go
list := widgets.NewListBox("city")
```

参数：

- `id string`
  - 列表框 ID。

常用方法：

- `SetItems(items []ListItem)`
- `Items() []ListItem`
- `SetSelected(index int)`
- `SelectedIndex() int`
- `SelectedItem() (ListItem, bool)`
- `SetStyle(style ListStyle)`
- `SetOnChange(fn func(int, ListItem))`
- `SetOnActivate(fn func(int, ListItem))`
- `SetOnRightClick(fn func(int, ListItem, core.Point))`
- `ClearSelection()`

行为说明：

- 只支持单选。
- 禁用项不会被选中。
- 支持滚轮滚动长列表。
- 双击列表项或按 `Enter` 会触发激活回调。
- 右键点击列表项会先选中该项，再触发右键回调。
- 支持键盘：
  - `Up` / `Down`
  - `Home` / `End`
  - `Enter` / `Space`

示例：

```go
list := widgets.NewListBox("city")
list.SetBounds(core.Rect{X: 20, Y: 220, W: 220, H: 140})
list.SetItems([]widgets.ListItem{
    {Value: "sh", Text: "上海"},
    {Value: "sz", Text: "深圳"},
    {Value: "gz", Text: "广州", Disabled: true},
})
list.SetOnChange(func(index int, item widgets.ListItem) {
    app.SetTitle(item.Value)
})
scene.Root().Add(list)
```

### 5.4 `ComboBox`

用途：带下拉弹层的单选框。

构造函数：

```go
combo := widgets.NewComboBox("city")
```

参数：

- `id string`
  - 组合框 ID。

常用方法：

- `SetItems(items []ListItem)`
- `Items() []ListItem`
- `SetSelected(index int)`
- `SelectedIndex() int`
- `SelectedItem() (ListItem, bool)`
- `SetPlaceholder(text string)`
- `SetStyle(style ComboStyle)`
- `SetOnChange(fn func(int, ListItem))`

行为说明：

- 只支持单选。
- 禁用项不会被选中。
- 展开后会绘制覆盖层。
- 支持键盘：
  - `Enter` / `Space` 打开或关闭下拉层
  - `Esc` 关闭下拉层
  - `Up` / `Down` 选择上一项或下一项
  - `Home` / `End` 跳到首项或末项

示例：

```go
combo := widgets.NewComboBox("city")
combo.SetBounds(core.Rect{X: 260, Y: 220, W: 180, H: 36})
combo.SetPlaceholder("请选择城市")
combo.SetItems([]widgets.ListItem{
    {Value: "bj", Text: "北京"},
    {Value: "hz", Text: "杭州"},
    {Value: "cd", Text: "成都"},
})
combo.SetOnChange(func(index int, item widgets.ListItem) {
    app.SetTitle(item.Text)
})
scene.Root().Add(combo)
```

## 6. 输入类组件

### 6.1 `EditBox`

用途：单行文本输入框。

构造函数：

```go
edit := widgets.NewEditBox("keyword")
```

参数：

- `id string`
  - 输入框 ID。

常用方法：

- `SetText(text string)`
- `TextValue() string`
- `SetPlaceholder(text string)`
- `SetReadOnly(readOnly bool)`
- `SetStyle(style EditStyle)`
- `SetOnChange(fn func(string))`
- `SetOnSubmit(fn func(string))`

行为说明：

- 只支持单行输入。
- 只读模式下仍可获得焦点，但不会修改内容。
- 支持键盘：
  - `Left` / `Right`
  - `Home` / `End`
  - `Backspace`
  - `Delete`
  - `Enter` 触发 `OnSubmit`

示例：

```go
edit := widgets.NewEditBox("keyword")
edit.SetBounds(core.Rect{X: 20, Y: 380, W: 240, H: 36})
edit.SetPlaceholder("输入关键字")
edit.SetOnChange(func(text string) {
    app.SetTitle(text)
})
scene.Root().Add(edit)
```

## 7. 图像类组件

### 7.1 `Image`

用途：显示静态位图。

构造函数：

```go
img := widgets.NewImage("logo")
```

参数：

- `id string`
  - 图像控件 ID。

常用方法：

- `SetScaleMode(mode ImageScaleMode)`
  - 缩放模式：
  - `widgets.ImageScaleStretch`
  - `widgets.ImageScaleContain`
  - `widgets.ImageScaleCenter`
- `SetOpacity(alpha byte)`
  - 设置透明度，`0` 为全透明，`255` 为不透明。
- `SetBitmap(bitmap *core.Bitmap)`
  - 使用外部位图，不接管释放。
- `SetBitmapOwned(bitmap *core.Bitmap)`
  - 使用位图并由控件负责 `Close()`。
- `LoadBytes(data []byte) error`
  - 从字节加载图像。
  - 当前已注册 PNG、JPEG、GIF 解码。
- `NaturalSize() core.Size`
  - 返回原始图像尺寸。
- `Bitmap() *core.Bitmap`
  - 返回当前位图。
- `Close() error`
  - 释放控件持有的位图。

资源所有权说明：

- 如果位图由你自己创建并自己管理，使用 `SetBitmap`。
- 如果希望控件负责释放，使用 `SetBitmapOwned` 或 `LoadBytes`。

示例：

```go
img := widgets.NewImage("logo")
img.SetBounds(core.Rect{X: 280, Y: 20, W: 128, H: 128})
img.SetScaleMode(widgets.ImageScaleContain)
if err := img.LoadBytes(pngBytes); err != nil {
    panic(err)
}
scene.Root().Add(img)
```

### 7.2 `AnimatedImage`

用途：显示 GIF 或逐帧动画。

构造函数：

```go
anim := widgets.NewAnimatedImage("spinner")
```

参数：

- `id string`
  - 动画图像 ID。

常用方法：

- `SetScaleMode(mode ImageScaleMode)`
- `SetOpacity(alpha byte)`
- `SetPlaying(playing bool)`
  - 控制播放和暂停。
- `LoadGIF(data []byte) error`
  - 从 GIF 字节生成帧。
- `SetFrames(frames []core.AnimatedFrame)`
  - 使用外部帧，不接管帧位图生命周期。
- `SetFramesOwned(frames []core.AnimatedFrame)`
  - 由控件接管帧位图释放。
- `NaturalSize() core.Size`
- `CurrentFrame() int`
- `Close() error`

行为说明：

- 动画依赖 `scene.HandleTimer(id)` 接收定时器回调。
- 当控件不可见、暂停播放或帧数小于等于 1 时，会自动停掉内部定时器。

示例：

```go
anim := widgets.NewAnimatedImage("spinner")
anim.SetBounds(core.Rect{X: 280, Y: 170, W: 96, H: 96})
anim.SetPlaying(true)
if err := anim.LoadGIF(gifBytes); err != nil {
    panic(err)
}
scene.Root().Add(anim)
```

## 8. 主题与样式

### 8.1 `FontSpec`

```go
type FontSpec struct {
    Face   string
    SizeDP int32
    Weight int32
}
```

- `Face`
  - 字体名称。
- `SizeDP`
  - 设备无关字号。
- `Weight`
  - 字重，例如 `400`、`700`。

### 8.2 `TextStyle`

```go
type TextStyle struct {
    Font   FontSpec
    Color  core.Color
    Format uint32
}
```

- `Font`
  - 字体设置。
- `Color`
  - 文本颜色。
- `Format`
  - 文本排版标志，例如 `core.DTCenter | core.DTVCenter | core.DTSingleLine`。

### 8.3 `ButtonStyle`

- `Font`
  - 按钮文本字体。
- `TextColor`
  - 普通文本颜色。
- `DownText`
  - 按下时的文本颜色。
- `DisabledText`
  - 禁用文本颜色。
- `Background`
  - 普通背景色。
- `Hover`
  - 悬停背景色。
- `Pressed`
  - 按下背景色。
- `Disabled`
  - 禁用背景色。
- `Border`
  - 边框色。
- `CornerRadius`
  - 圆角半径。
- `IconSizeDP`
  - 图标绘制尺寸。
- `TextInsetDP`
  - 图标与文本混排时的预留偏移。
- `GapDP`
  - 图标与文本之间的间距。
- `PadDP`
  - 按钮内容内边距。

### 8.4 `ProgressStyle`

- `Font`
  - 百分比文本字体。
- `TextColor`
  - 百分比文本颜色。
- `TrackColor`
  - 轨道颜色。
- `FillColor`
  - 进度填充色。
- `BubbleColor`
  - 百分比气泡底色。
- `CornerRadius`
  - 轨道圆角。
- `ShowPercent`
  - 是否显示百分比。
  - 为 `true` 时会显示跟随进度位置移动的百分比气泡。

### 8.5 `ChoiceStyle`

适用于 `CheckBox` 和 `RadioButton`。

- `Font`
  - 文本字体。
- `TextColor`
  - 正常文本颜色。
- `DisabledText`
  - 禁用文本颜色。
- `Background`
  - 指示器背景色。
- `BorderColor`
  - 正常边框色。
- `HoverBorder`
  - 悬停边框色。
- `FocusBorder`
  - 焦点边框色。
- `IndicatorColor`
  - 选中标记颜色。
- `CheckColor`
  - 复选框圆心点颜色。
- `HoverBackground`
  - 整行悬停背景色。
- `DisabledBg`
  - 禁用背景色。
- `DisabledBorder`
  - 禁用边框色。
- `CornerRadius`
  - 圆角半径。
- `IndicatorSizeDP`
  - 指示器尺寸。
- `IndicatorGapDP`
  - 指示器与文本间距。

### 8.6 `ListStyle`

- `Font`
  - 列表文本字体。
- `TextColor`
  - 正常文本颜色。
- `DisabledText`
  - 禁用项文本颜色。
- `Background`
  - 背景色。
- `BorderColor`
  - 普通边框色。
- `HoverBorder`
  - 悬停边框色。
- `FocusBorder`
  - 焦点边框色。
- `ItemHoverColor`
  - 悬停项底色。
- `ItemSelectedColor`
  - 选中项底色。
- `ItemTextColor`
  - 选中项文本颜色。
- `ItemHeightDP`
  - 行高。
- `PaddingDP`
  - 列表内边距。
- `CornerRadius`
  - 圆角半径。

### 8.7 `ComboStyle`

- `Font`
  - 主文本和下拉项字体。
- `TextColor`
  - 选中值文本颜色。
- `PlaceholderColor`
  - 占位文本颜色。
- `Background`
  - 输入框背景色。
- `BorderColor`
  - 普通边框色。
- `HoverBorder`
  - 悬停边框色。
- `FocusBorder`
  - 焦点和展开态边框色。
- `ArrowColor`
  - 箭头颜色。
- `PopupBackground`
  - 弹层背景色。
- `ItemHoverColor`
  - 悬停项底色。
- `ItemSelectedColor`
  - 选中项底色。
- `ItemTextColor`
  - 选中项文本颜色。
- `ItemHeightDP`
  - 下拉项高度。
- `PaddingDP`
  - 内边距。
- `CornerRadius`
  - 圆角半径。
- `MaxVisibleItems`
  - 展开时最多显示多少项。

### 8.8 `EditStyle`

- `Font`
  - 输入框字体。
- `TextColor`
  - 正常文本颜色。
- `PlaceholderColor`
  - 占位文本颜色。
- `Background`
  - 背景色。
- `BorderColor`
  - 普通边框色。
- `HoverBorder`
  - 悬停边框色。
- `FocusBorder`
  - 聚焦边框色。
- `DisabledText`
  - 禁用文本颜色。
- `DisabledBg`
  - 禁用背景色。
- `CaretColor`
  - 光标颜色。
- `PaddingDP`
  - 内边距。
- `CornerRadius`
  - 圆角半径。

### 8.9 `Theme`

```go
theme := widgets.DefaultTheme()
theme.Button.Background = core.RGB(30, 41, 59)
theme.Button.TextColor = core.RGB(255, 255, 255)
scene.SetTheme(theme)
```

`Theme` 包含以下字段：

- `BackgroundColor`
- `Text`
- `Title`
- `Button`
- `Progress`
- `CheckBox`
- `RadioButton`
- `ListBox`
- `ComboBox`
- `Edit`

其中：

- `Text` 和 `Title` 用于普通文本和标题文本样式。
- 其余字段分别对应同名组件的默认样式。

## 9. 常见问题

### 9.1 为什么组件不响应点击？

通常有三个原因：

- 没有把窗口事件转发给 `Scene`。
- 控件 `Visible` 为 `false`。
- 控件 `Enabled` 为 `false`。

### 9.2 为什么动画图片不动？

确认以下接线已经存在：

```go
OnTimer: func(_ *core.App, id uintptr) {
    scene.HandleTimer(id)
}
```

### 9.3 为什么 `RadioButton` 没有互斥？

确认：

- 两个单选按钮都设置了相同的 `Group`。
- 它们位于同一个父容器下面。

### 9.4 为什么样式覆盖有时不生效？

当前实现里，很多样式字段的零值表示“不覆盖默认值”。如果你传入的是零值颜色或零尺寸，可能会被忽略。

### 9.5 如何确认是否已经回退到 GDI？

可以在运行时读取：

```go
backend := app.RenderBackend()
reason := app.RenderFallbackReason()
```

如果 `backend == core.RenderBackendGDI` 且 `reason != ""`，说明应用在 `RenderModeAuto` 下尝试过 Direct2D，但最终回退到了 GDI。


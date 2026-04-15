# WIDGETS 指南

## 1. 核心模型

`widgets` 建立在 `core` 之上，`Scene` 是运行时协调者。

- `widgets.BindScene(&opts, hooks)` 把控件生命周期接到 `core.Options`
- `scene.Root()` 返回根面板
- `scene.Theme()` 返回当前主题
- `Scene` 负责焦点、悬停、鼠标捕获、定时器、事件分发和重绘协调

典型流程：

1. 配置 `core.Options`
2. 调用 `widgets.BindScene`
3. 在 `OnCreate` 中创建控件
4. 把控件加到 `scene.Root()`
5. 调用 `app.Init()` 和 `app.Run()`

## 2. 控件模式

以下控件支持 `mode`：

- `widgets.ModeCustom`
- `widgets.ModeNative`

常见控件：

- `Button`
- `EditBox`
- `CheckBox`
- `RadioButton`
- `ComboBox`
- `FilePicker`

如果需要 Win10 / Win11 视觉样式，最终可执行文件仍然需要 `Microsoft.Windows.Common-Controls` v6 manifest。

## 3. 布局

`Panel` 是基础容器，使用 `SetLayout(...)` 指定布局。

内置布局：

- `AbsoluteLayout`
- `RowLayout`
- `ColumnLayout`
- `GridLayout`
- `FormLayout`

常用子项布局数据：

- `FlexLayoutData`
- `GridLayoutData`
- `FormLayoutData`

## 4. 常用控件

- `Panel`
- `Label`
- `Button`
- `CheckBox`
- `RadioButton`
- `ComboBox`
- `EditBox`
- `FilePicker`
- `Image`
- `AnimatedImage`
- `ListBox`
- `ProgressBar`
- `ScrollView`

`Button` 支持文本 + 图片内容，图片槽位尺寸由 `ButtonStyle.ImageSizeDP` 控制。

## 5. JSON UI

声明式 UI 现在统一放在 `widgets/jsonui`。

核心 API：

- `jsonui.LoadDocumentFile(...)`
- `jsonui.LoadDocumentString(...)`
- `jsonui.LoadIntoScene(...)`
- `jsonui.LoadFileIntoScene(...)`
- `doc.PrimaryWindow()`
- `doc.Window(id)`
- `win.FindWidget(id)`
- `doc.FindWidget(winID, widgetID)`
- `widgets.FindByID(root, id)`
- `widgets.FindByIDAs[T](root, id)`
- `doc.NewApps(baseOpts)`
- `jsonui.RunApps(...)`

`widgets/jsonui` 里的窗口和按钮图片统一使用 `image` / `imagePos` / `imageSizeDP` 语义，详情见 `JSONUI.zh-CN.md`。

## 6. JSON 结构

顶层使用 `wins` 声明一个或多个窗口：

```json
{
  "wins": [
    {
      "id": "main",
      "title": "Demo",
      "w": 980,
      "h": 720,
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "label",
            "id": "title",
            "text": "Hello",
            "frame": { "x": 20, "y": 20, "w": 240, "h": 28 }
          }
        ]
      }
    }
  ]
}
```

常用键：

- `wins`
- `id`
- `title`
- `w` / `h` / `minW` / `minH`
- `root`
- `type`
- `layout`
- `children`
- `frame`
- `style`
- `text`
- `value`
- `readOnly`
- `multiline`
- `wordWrap`
- `acceptReturn`
- `verticalScroll`
- `horizontalScroll`
- `items`
- `sel`

布尔字段需要保持控件语义默认值：`visible` / `enabled` 缺省时保持 `true`，`checked` 缺省时保持 `false`，`ScrollView` 缺省时保持 `verticalScroll=true` / `horizontalScroll=false`。

## 7. 绑定模型

JSON 只声明绑定关系，不在 JSON 文本里做数据增删改查。

宿主层负责数据：

- 自己实现 `jsonui.DataSource`
- 或直接使用 `jsonui.Store`

绑定写法：

```json
{
  "title": { "bind": "page.title", "default": "Fallback" }
}
```

Go 侧更新：

```go
store := jsonui.NewStore(map[string]any{
	"page": map[string]any{
		"title": "Initial",
	},
})

store.Set("page.title", "Updated")
```

绑定同样可以用于编辑框的 `readOnly`、`multiline` 等开关，方便在宿主层和局部 imperative 代码之间切换。

当前支持的常见绑定目标：

- 窗口标题
- 文本内容
- 输入值
- 可见状态
- 可用状态
- 选中状态
- 列表项
- 当前选择
- 绝对布局 `frame`

## 8. 表达式与 DPI

`frame` 的数值默认按逻辑 DP 处理，并在布局时按 DPI 缩放。

表达式现在支持整数四则运算：

- `+`
- `-`
- `*`
- `/`
- `()`

可用变量仅限：

- `winW`
- `winH`
- `parentW`
- `parentH`

`%` 保留为百分比字面量，不是取模运算。例如：

- `"50%"`
- `"50%+12"`
- `"(parentW - 12*3 - 20*2 - 108) / 4"`
- `"(parentW-184)/4"`

其中：

- `50%` 仍表示窗口当前宽度或高度的 50%
- `x` / `w` 轴按窗口宽度取百分比
- `y` / `h` 轴按窗口高度取百分比
- 其他数字字面量仍按逻辑 DP 处理

`frame` 支持：

- `x`
- `y`
- `r`
- `b`
- `w`
- `h`

常见组合：

- `x + y + w + h`
- `x + r + h`
- `y + b + w`
- `x + r + y + b`

## 9. 样式映射

JSON 样式直接映射到现有控件样式结构，而不是再造一套渲染系统。

覆盖范围包括：

- `button`
- `progress`
- `checkbox`
- `radio`
- `select`
- `listbox`
- `input`
- `textarea`
- `panel`

常用键示例：

- `fg`
- `bg`
- `ph`
- `border`
- `hoverBg`
- `pressedBg`
- `hoverBorder`
- `focusBorder`
- `radius`
- `pad`
- `gap`
- `imageSize`
- `itemH`
- `indicatorStyle`

按钮图片槽位按 contain 方式缩放，不会强行拉伸成正方形。

## 10. 文件对话框控件

`type: "file"` 映射到 `widgets.FilePicker`。

常用字段：

- `dialog: "open" | "save" | "folder"`
- `multiple`
- `accept`
- `filters`
- `buttonText`
- `dialogTitle`
- `defaultExt`
- `valueSep`

## 11. Demo 与验证

```powershell
go run ./cmd/demo
go run ./cmd/demo_json
go run ./cmd/demo_json_full
go test ./...
go test -v ./cmd/demo_json_full
go vet ./...
```

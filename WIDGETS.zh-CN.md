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

缁戝畾鍚屾牱鍙互鐢ㄥ湪缂栬緫妗嗙殑 `readOnly`銆?`multiline` 绛夊紑鍏充笂锛屾柟渚垮湪瀹夸富灞傚拰灞€閮?imperative 浠ｇ爜涔嬮棿鍒囨崲銆?

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

支持的表达式：

- `100`
- `"50%"`
- `"50%-100"`
- `"winW-100"`
- `"winH-100"`
- `"parentW-100"`
- `"parentH-100"`

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
- `itemH`
- `indicatorStyle`

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
go test ./...
go vet ./...
```

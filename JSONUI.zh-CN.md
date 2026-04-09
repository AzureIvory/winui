# JSON UI 使用指南

这份文档面向直接使用 `widgets/jsonui` 的业务代码，重点说明 JSON DSL、绑定方式、表达式语义、多窗口，以及宿主层应该如何组织数据。

## 1. 设计原则

- JSON 只负责描述 UI 结构、样式、动作和绑定关系
- 数据的增删改查全部发生在宿主层
- JSON 默认使用逻辑 DP，布局时自动处理 DPI
- JSON UI 不再兼容旧的 HTML / CSS DSL

## 2. 最小示例

```json
{
  "wins": [
    {
      "id": "main",
      "title": { "bind": "page.title", "default": "Demo" },
      "w": 980,
      "h": 720,
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "label",
            "id": "title",
            "text": { "bind": "page.title" },
            "frame": { "x": 20, "y": 20, "w": 320, "h": 28 }
          },
          {
            "type": "button",
            "id": "saveBtn",
            "text": "Save",
            "frame": { "x": "50%-70", "y": 56, "w": 140, "h": 40 }
          }
        ]
      }
    }
  ]
}
```

Go 侧：

```go
store := jsonui.NewStore(map[string]any{
	"page": map[string]any{
		"title": "Initial Title",
	},
})

doc, err := jsonui.LoadDocumentFile("demo.ui.json", jsonui.LoadOptions{
	Data: store,
})
if err != nil {
	panic(err)
}

win := doc.PrimaryWindow()
_ = win

store.Set("page.title", "Updated Title")
```

## 3. 顶层结构

顶层必须是一个对象，并包含 `wins`。

每个窗口常用字段：

- `id`: 窗口标识，必填
- `title`: 窗口标题，支持字面量或绑定
- `icon`: `.ico` 路径
- `w` / `h`: 初始客户区尺寸
- `minW` / `minH`: 最小客户区尺寸
- `bg`: 窗口背景色
- `root`: 根节点，必填

## 4. 常用节点类型

- `panel`
- `label`
- `button`
- `input`
- `textarea`
- `progress`
- `checkbox`
- `radio`
- `select`
- `listbox`
- `file`
- `image`
- `animimg`

常用公共字段：

- `id`
- `type`
- `style`
- `visible`
- `enabled`
- `frame`

不同控件的常用业务字段：

- `text`
- `value`
- `placeholder`
- `items`
- `sel`
- `group`
- `src`
- `dialog`
- `multiple`
- `accept`
- `filters`
- `buttonText`
- `dialogTitle`
- `defaultExt`
- `valueSep`

## 5. 绑定写法

绑定统一使用对象：

```json
{ "bind": "page.title", "default": "Fallback Title" }
```

适合绑定的字段：

- `title`
- `text`
- `value`
- `visible`
- `enabled`
- `checked`
- `items`
- `sel`
- `frame.x`
- `frame.y`
- `frame.r`
- `frame.b`
- `frame.w`
- `frame.h`

## 6. 宿主数据源

宿主层有两种方式：

1. 直接实现 `jsonui.DataSource`
2. 使用内置的 `jsonui.Store`

`jsonui.Store` 常用方法：

- `Get(path)`
- `Set(path, value)`
- `Patch(map[string]any)`
- `Replace(snapshot)`
- `Subscribe(fn)`

建议：

- 小型页面用 `Set` / `Patch`
- 数据结构变化很大时用 `Replace`
- 如果你已经有自己的状态容器，直接实现 `DataSource` 即可

## 7. 表达式语义

`frame` 内的数值默认都按逻辑 DP 处理，并在布局阶段自动按 DPI 缩放。

支持的表达式：

- `100`
- `"50%"`
- `"50%-100"`
- `"winW-100"`
- `"winH-100"`
- `"parentW-100"`
- `"parentH-100"`

说明：

- `50%` 表示窗口当前宽度或高度的 50%
- `winW` / `winH` 表示窗口客户区逻辑尺寸
- `parentW` / `parentH` 表示父容器逻辑尺寸
- `-100` 里的 `100` 也是逻辑 DP

`frame` 字段：

- `x`
- `y`
- `r`
- `b`
- `w`
- `h`

常用组合：

- `x + y + w + h`
- `x + r + h`
- `y + b + w`
- `x + r + y + b`

## 8. 样式字段

JSON 样式不是一套新的渲染系统，而是直接映射到现有控件样式结构。

常见键：

- `fg`
- `bg`
- `ph`
- `border`
- `hoverBg`
- `pressedBg`
- `disabledBg`
- `hoverBorder`
- `focusBorder`
- `radius`
- `borderW`
- `pad`
- `gap`
- `size`
- `weight`
- `align`
- `itemH`
- `itemFg`
- `indicator`
- `indicatorStyle`

文件选择器支持嵌套按钮样式：

```json
{
  "style": {
    "fg": "#1f2937",
    "bg": "#ffffff",
    "border": "#cbd5e1",
    "btn": {
      "bg": "#f5f9ff",
      "hoverBg": "#e0f2fe",
      "pressedBg": "#2563eb",
      "border": "#9dbfe8"
    }
  }
}
```

## 9. 文件选择器

`type: "file"` 映射到 `widgets.FilePicker`。

常见字段：

- `dialog: "open" | "save" | "folder"`
- `multiple`
- `accept`
- `filters`
- `buttonText`
- `dialogTitle`
- `defaultExt`
- `valueSep`

动作回调里可以拿到：

- `ActionContext.Value`
- `ActionContext.Paths`

## 10. 多窗口

一个 JSON 文档可以声明多个窗口。

常用 API：

- `doc.PrimaryWindow()`
- `doc.Window("tools")`
- `doc.NewApps(baseOpts)`
- `jsonui.RunApps(hosted)`

典型用法：

```go
doc, err := jsonui.LoadDocumentFile("app.ui.json", jsonui.LoadOptions{})
if err != nil {
	panic(err)
}

hosted, err := doc.NewApps(core.Options{
	ClassName:      "WinUIJSONApp",
	DoubleBuffered: true,
	RenderMode:     core.RenderModeAuto,
})
if err != nil {
	panic(err)
}

jsonui.RunApps(hosted)
```

## 11. 迁移建议

如果你之前在用旧的 HTML / CSS DSL：

1. 先把窗口拆成 `wins`
2. 把标签改成 JSON 节点
3. 把 style 改成 JSON 对象
4. 把绝对定位改成 `frame`
5. 把 `bind-*` 改成 `{ "bind": "...", "default": ... }`

## 12. 当前边界

- 绝对布局仍然是约束式布局，不是完整 CSS 盒模型
- JSON 已支持多窗口、样式映射、文件对话框、宿主绑定和 DPI 表达式
- 如果需要更复杂的模板、循环或条件渲染，建议在宿主层生成 JSON 或直接用 Go 代码构建控件树

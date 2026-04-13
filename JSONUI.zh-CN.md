# JSON UI 使用指南

本文面向直接使用 `widgets/jsonui` 的宿主代码，重点说明 JSON DSL、绑定方式、表达式语义、多窗口支持，以及少量 imperative 运行时代码应该如何和声明式 UI 协作。

## 1. 设计原则

- JSON 只负责描述 UI 结构、样式、动作和绑定关系
- 数据的增删改查全部发生在宿主层
- JSON 默认使用逻辑 DP，布局阶段自动做 DPI 缩放
- `widgets/jsonui` 不承担业务状态管理；宿主通过 `jsonui.DataSource` 驱动 UI

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
            "frame": { "x": "50%-70", "y": 56, "w": 140, "h": 40 },
            "onClick": "save"
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
title := win.FindWidget("title")
_ = title

store.Set("page.title", "Updated Title")
```

## 3. 顶层结构

顶层必须是一个对象，并包含 `wins`。

每个窗口常用字段：

- `id`: 窗口标识，必填
- `title`: 窗口标题，支持字面量或绑定
- `icon`: `.ico` 路径，默认按当前屏幕 DPI 把逻辑 `32dp` 缩放后加载，也可以通过 `LoadOptions.IconSizeDP` 覆盖
- `w` / `h`: 初始客户区尺寸
- `minW` / `minH`: 最小客户区尺寸
- `bg`: 窗口背景色
- `root`: 根节点，必填

同一窗口内的 widget `id` 必须唯一。loader 会按窗口建立运行时索引，供 `win.FindWidget(...)` 和 `doc.FindWidget(...)` 使用。

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

文本输入相关字段，`input` 和 `textarea` 都可使用：

- `value`
- `placeholder`
- `readOnly`
- `multiline`
- `wordWrap`
- `acceptReturn`
- `verticalScroll`
- `horizontalScroll`

其他常见业务字段：

- `text`
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

说明：

- `textarea` 只是多行编辑框的语义别名，默认 `multiline: true`
- `input` 也可以显式写 `multiline: true`，适合需要统一节点类型时使用

## 5. 绑定写法

绑定统一使用对象：

```json
{ "bind": "page.title", "default": "Fallback Title" }
```

适合绑定的字段包括：

- `title`
- `text`
- `value`
- `readOnly`
- `multiline`
- `wordWrap`
- `acceptReturn`
- `verticalScroll`
- `horizontalScroll`
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

这意味着你可以在宿主层临时切换输入框能力，例如把日志框切成只读、把输入框切成多行，而不需要重建整棵 UI 树。

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
- 如果已经有自己的状态容器，直接实现 `DataSource` 即可

## 7. 表达式语义

`frame` 内的数值默认都按逻辑 DP 处理，并在布局阶段自动按 DPI 缩放。

现在支持完整的整数四则运算：

- `+`
- `-`
- `*`
- `/`
- `()`

允许的变量只有：

- `winW`
- `winH`
- `parentW`
- `parentH`

百分号保留原来的“百分比字面量”语义，不是取模运算：

- `50%` 表示窗口当前宽度或高度的 50%
- 百分比基准仍然由 `frame` 所在轴决定：`x` / `w` 走窗口宽度，`y` / `h` 走窗口高度
- `50%+12` 这类写法合法，后面的 `12` 仍然是逻辑 DP

示例：

- `100`
- `"50%"`
- `"50%+12"`
- `"(parentW - 12*3 - 20*2 - 108) / 4"`
- `"(parentW-184)/4"`

说明：

- `winW` / `winH` 表示窗口客户区逻辑尺寸
- `parentW` / `parentH` 表示父容器逻辑尺寸
- 普通数字字面量始终按逻辑 DP 处理

`frame` 字段：

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

## 8. 样式字段

JSON 样式不是另一套渲染系统，而是直接映射到现有控件样式结构。

常见字段：

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

- `ActionContext.Window`
- `ActionContext.Value`
- `ActionContext.Paths`

## 10. 运行时辅助 API

当宿主代码里存在少量不适合绑定的临时 imperative 操作时，优先使用这些 helper，而不是自己递归遍历控件树：

- `win.FindWidget("status")`
- `doc.FindWidget("main", "status")`
- `widgets.FindByID(root, "status")`
- `widgets.FindByIDAs[*widgets.Label](root, "status")`

典型用途：

- 更新某个状态标签
- 在动作回调里拿当前窗口并修改标题
- 针对一两个控件做局部 imperative 行为，而不把整页改成纯手写 UI

## 11. 多窗口

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

## 12. 迁移建议

如果之前在用旧的 HTML / CSS DSL：

1. 先把窗口拆成 `wins`
2. 把标签改成 JSON 节点
3. 把 `style` 改成 JSON 对象
4. 把绝对定位改成 `frame`
5. 把 `bind-*` 改成 `{ "bind": "...", "default": ... }`

## 13. 当前边界

- 绝对布局仍然是约束式布局，不是完整的 CSS 盒模型
- JSON 已支持多窗口、样式映射、文件对话框、宿主绑定和 DPI 表达式
- 如果需要更复杂的模板、循环或条件渲染，建议在宿主层生成 JSON，或直接用 Go 代码构建控件树

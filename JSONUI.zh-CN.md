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
- `image`: 窗口图片路径，默认按当前屏幕 DPI 把逻辑 `32dp` 缩放后加载，也可以通过 `LoadOptions.ImageSizeDP` 覆盖
- `w` / `h`: 初始客户区尺寸
- `minW` / `minH`: 最小客户区尺寸
- `bg`: 窗口背景色
- `root`: 根节点，必填

同一窗口内的 widget `id` 必须唯一。loader 会按窗口建立运行时索引，供 `win.FindWidget(...)` 和 `doc.FindWidget(...)` 使用。

### 3.1 组件模板

顶层可选增加 `components` 注册表，把组件名映射到外部模板文件：

```json
{
  "components": {
    "button": "components/button.json",
    "summaryCard": "components/summary_card.json"
  },
  "wins": [...]
}
```

路径相对于 `LoadOptions.AssetsDir`；如果使用 `LoadDocumentFile(...)`，默认就是主 JSON 所在目录。

主 JSON 在“节点位置”可以直接实例化组件：

```json
{
  "component": "button",
  "args": {
    "id": "saveBtn",
    "text": "保存",
    "action": "save",
    "x": 20,
    "y": 20
  }
}
```

对应模板文件示例：

```json
{
  "component": "button",
  "params": {
    "id": { "type": "string" },
    "text": { "type": "string", "default": "OK" },
    "action": { "type": "string" },
    "x": { "type": "int" },
    "y": { "type": "int" }
  },
  "node": {
    "type": "button",
    "id": "${id}",
    "text": "${text}",
    "onClick": "${action}",
    "frame": { "x": "${x}", "y": "${y}", "w": 120, "h": 36 }
  }
}
```

规则：

- 一个模板文件只定义一个组件
- 一个组件实例只展开成一棵节点子树，不会直接产出多个并列 sibling
- `args` 只允许传给模板里声明过的 `params`
- 组件实例还可通过 `slots` 传命名子节点数组，例如 `{ "slots": { "content": [ ... ] } }`
- 缺少无默认值的参数会报错
- 支持的参数类型：`string` / `int` / `bool` / `object` / `array`
- `type` 可省略；若存在 `default`，会从默认值推导类型；若两者都省略，则接受任意 JSON 值
- 占位符只允许整值替换，例如 `"${title}"`、`"${style}"`、`"${items}"`；不支持 `"btn-${id}"` 这种字符串拼接
- 会检测递归展开和循环引用
- 当前只支持显式命名的 `children` slot，占位节点写成 `{ "slot": "content" }`；不支持条件分支、循环渲染，也不支持在组件实例节点上追加额外 override 字段

### 图片资源

窗口和按钮统一使用 `image` 语义，而不是旧的 `icon` 语义。

- 窗口字段：`image`、`imageSizeDP`
- 按钮字段：`image`、`imagePos`、`imageSizeDP`
- 样式字段：`style.imageSize`
- `LoadOptions.ImageSizeDP` 可作为窗口和按钮图片尺寸的默认值
- 支持格式：`png`、`jpg`、`jpeg`、`gif`（只取首帧）
- 不保证：`bmp`、`svg`、`webp`、`avif`，以及把 `.ico` 作为专用唯一格式的旧语义
- `window.image` 最终仍会生成 Win32 `HICON`
- `window.image` 和 `button.image` 只做静态显示；需要播放 GIF 时请使用 `animimg`
- 图片按原始宽高比以 contain 方式缩放，不会强行拉伸成正方形
- 图片会按目标像素尺寸和缩放质量缓存；Direct2D bitmaps 优先，必要时回退到 GDI

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
- `shape`
- `radius`
- `borderW`
- `pad`
- `gap`
- `imageSize`
- `size`
- `weight`
- `align`
- `itemH`
- `itemFg`
- `indicator`
- `indicatorStyle`
- `imageSize` 表示按钮图片槽位尺寸，按 contain 方式适配，不会强制拉伸成正方形。
- `shape` 支持 `default` 和 `pill`；`pill` 会自动使用按钮高度一半作为圆角，适合胶囊按钮。
- `border` 支持 `transparent`；将其设为 `"transparent"` 可声明无边框按钮。

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

### `listbox` 双行条目与宿主交互

`listbox` 也支持已有条目字段 `subtitle`。配置足够的 `style.itemH`（默认字体建议 56 DP）后显示双行条目；紧凑行高保留单行行为，无需新增 JSON 节点或布局类型：

```json
{
  "type": "listbox",
  "id": "entries",
  "style": { "itemH": 56 },
  "items": [
    { "value": "one", "text": "主标题", "subtitle": "副标题预览" }
  ]
}
```

窗口隐藏启动、物理屏幕定位和快捷键消费属于宿主运行时能力：在 `core.Options` 设置 `StartHidden`，通过 `App.SetScreenBounds` / `ShowNoActivate` / `Hide` / `RaiseNoActivate` 控制窗口，通过 `Scene.FocusWidget` / `SetKeyPreview` 操作挂载后的场景。JSON 的 `frame` 仍按原有布局缩放规则解释；没有新增窗口定位或快捷键业务 DSL。

表单需要 Tab / Shift+Tab 时，宿主可在当前场景的 UI 线程按键预览中调用 `Scene.FocusNext(key.Shift)`；资格判定、子控件树顺序和模态范围由场景统一管理，不需要在产品中维护另一份控件顺序，也没有新增 JSON Tab 配置。

嵌套 `scrollview` 内的原生编辑输入层也按祖先视口交集裁剪，不会盖住视口外的固定页脚。宿主调用 `FocusWidget` / `FocusNext` 时会用既有滚动偏移显露离屏控件，无需另写 JSON 滚动动作或产品级原生窗口覆盖逻辑。

鼠标点击仍按原始控件位置解释，不会在处理点击坐标前自动滚动。编辑控件的初始文本在原生输入层变更通知启用前写入，JSON/宿主预先设置的正文可在挂载后保留。

自绘 `button` 同样参与这个焦点顺序，聚焦后无修饰键的 Enter / Space 触发其原有 `onClick` 动作一次，并显示按钮内部的主题文字色焦点轮廓；不需要添加单独的键盘动作或焦点样式字段。

### `select` 富样式扩展

- `style.layout` 支持 `default` 和 `rich`。
- 当 `layout = "rich"` 且控件运行在 `ModeCustom` 时，组合框按「左图标 + 主标题 + 副标题」绘制。
- `items` 数组新增可选字段：
  - `subtitle`：副标题文本
  - `image`：条目左侧图标路径（相对 `assetsDir`）
- 富样式常用样式键：
  - `subtitleFg`：副标题颜色
  - `iconBg`：图标槽背景色
  - `mark`：选中项右侧勾标记底色
  - `imageSize`：图标绘制尺寸（DP）

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
## 14. `scale` 缩放策略

默认情况下，JSON 仍然沿用旧版行为：布局尺寸、字体、图片槽位、padding、gap、radius 都按逻辑 DP 参与 DPI 缩放。

如果某个节点需要混合策略，可以声明 `scale`：

```json
{
  "type": "button",
  "id": "fixedBtn",
  "scale": {
    "layout": "px",
    "font": "px",
    "image": "px",
    "padding": "px",
    "gap": "px",
    "radius": "dp"
  }
}
```

也可以写成简写：

```json
{
  "type": "panel",
  "scale": "px"
}
```

约定如下：

- `mode` 是节点默认策略，可向后代继承
- `layout` 控制该节点自己的布局尺寸，例如 preferred size、绝对布局 `frame`，以及容器自己的 `gap` / `pad` / `item` / `labelW`
- `font` / `image` / `padding` / `gap` / `radius` 控制渲染侧样式尺寸
- `frame` 默认仍按逻辑 DP 解释；当节点 `scale.layout = "px"` 时，`frame` 表达式里的 `winW` / `winH` / `parentW` / `parentH` 会改用物理像素，结果也不再二次按 DPI 缩放
- 不写 `scale` 时，旧页面行为保持不变

## 紧凑知识工作台键盘语义（2026-10-29）

`ListBox.SetShowCheck(true)` 时，聚焦的有效选中行可用不带修饰键的 Space 切换勾选。勾选通过现有 `OnCheckChange` 回调报告，不改变单选项，不触发 `OnChange` 或 `OnActivate`。按键自动重复只消费，不重复切换；Ctrl/Alt/Shift、IME 组合、禁用/隐藏项不参与勾选。该操作复用 Scene 的 TranslateMessage 前默认键阶段；即使回调转焦点到原生编辑器，旧 Space 也不会作为字符漏入新编辑器。应用不应模拟第二套勾选键盘循环。

`EditBox.Composing() bool` 是 UI 线程上的只读原生 IME 状态查询。需要延迟搜索的宿主可在去抖回调中检查它，组合期间等待，组合结束后再按需发起后台工作。它复用已有原生组合状态，不执行网络/文件访问，也不启动额外消息循环。JSON UI 的 ListBox `showCheck` 自动采用上述通用 Space 行为；组合状态查询是 Go 宿主 API，无需新的 JSON 字段。

## 宿主受宽测量

Go 宿主可在文档已挂到 scene 后，于 UI 线程调用 `widgets.MeasureForWidth(widget, width)`。宽度和结果均为物理像素，不能直接作为逻辑 DP 的 JSON frame 或 `SetPreferredSize` 值；需要按实际 theme、字体、DPI 测量并换算单位。Label 自动增长时使用既有多行/换行能力并留空显式 preferred height；显式高度仍优先。该 API 仅公开原有测量路径，不新增 JSON 字段、绑定或布局模式。完整语义见 [受宽测量](WIDGETS.zh-CN.md#受宽测量)。

行、列和线性布局的隐藏子项不贡献测量尺寸或间距；滚动边界也忽略隐藏节点保留的旧几何。通过既有 `visible` 或 Go setter 切换显隐即可更新范围，无需新增 JSON 属性或清空节点边界。

ScrollView 的动态后代通过现有 Go setter 更新内容、布局、preferred size 等度量后，会自动使祖先滚动测量失效，在相同视口下重新派生横纵范围；纯滚动和悬停继续复用缓存。原生横滚沿现有事件链进入启用了 `horizontalScroll` 的容器，正值向右；Shift + 纵滚保持既有横移优先行为。Go 事件的 `Horizontal` 字段与原始 `Flags`、有符号 `Delta` 独立，不新增 JSON 轴字段或第二套回调。白色背景省略边框色即可使用默认滚条配色；将背景和边框都设白会同时影响派生滚条颜色。

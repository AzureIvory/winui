# WIDGETS 指南

## 1. 核心模型

`widgets` 建立在 `core` 之上，`Scene` 是运行时协调者。

- `widgets.BindScene(&opts, hooks)` 把控件生命周期接到 `core.Options`
- `scene.Root()` 返回根面板
- `scene.Theme()` 返回当前主题
- `Scene` 负责焦点、悬停、鼠标捕获、定时器、事件分发和重绘协调
- `scene.FocusWidget(widget)` 显式聚焦当前场景的控件；会检查控件及祖先的可见、启用状态，原生控件会同步系统焦点，传 `nil` 清除场景焦点
- `scene.FocusNext(reverse bool) bool` 在 UI 线程按实际子控件的深度优先顺序循环聚焦，`reverse=true` 反向；复用 `FocusWidget` 的资格检查与原生焦点转移，当前最上层可见启用的 `Modal` 内循环，没有候选或非 UI 线程调用时返回 `false`
- `scene.SetKeyPreview(func(core.KeyEvent) bool)` 或 `SceneHooks.OnKeyPreview` 注册可消费的快捷键预览；返回 `true` 后该按键不会再分发或生成字符

使用 `BindScene` 时，预览同时覆盖窗口和 RichEdit 等原生子控件，发生在 `TranslateMessage` 之前，只执行一次。`KeyEvent` 提供 `Ctrl` / `Shift` / `Alt` / `Repeat` / `Composing`。IME 组合输入及候选确认期间跳过快捷键与编辑器提交；未消费的按键继续沿用原有控件和回调。`SetKeyPreview` 会替换当前场景的预览处理器（包括创建时的 hook）。

需要表单 Tab / Shift+Tab 时，在应用现有预览中对无 Ctrl/Alt 的 Tab 返回 `scene.FocusNext(key.Shift)` 即可。库不自动接管 Tab，也不改变既有控件的文本输入语义。原生编辑器的默认提交排除 Alt+Enter / Ctrl+Alt+Enter；应用仍可在优先执行的预览中显式定义这些快捷键。

`FocusWidget` / `FocusNext` 会先通过祖先 `ScrollView` 的原有滚动偏移显露目标，再转移焦点；嵌套视口由内向外处理，只移动必要距离。目标已完整可见时不动；目标大于视口且已覆盖视口时也不动，重复聚焦不会来回跳动。鼠标聚焦复用同一条真实 HWND 焦点转移；`FocusWidget(nil)` 把系统键盘焦点交回窗口宿主，而窗口失焦用的 `Blur` 不主动抢回系统焦点。

鼠标聚焦只转移焦点，不在原指针事件分发前自动滚动；部分可见输入框的原始坐标和光标落点因此保持一致。`SetText` 可在编辑框加入场景前设置初始草稿，两个模式都会在注册原生变更通知前先写入新建输入层，避免挂载过程清空正文。

原生编辑框的默认 `OnSubmit` 在外部预览未消费后，也在同一个翻译前入口处理并消费提交键，不生成尾随 CR/LF。多个积压的提交按键仍逐个提交，不会把换行留在草稿中。`KeyEvent.Handle` 标记实际输入窗口；IME 包装的普通按键在翻译前恢复原键码，未开始组合输入时仍可正常 Enter 提交。直接 `SendMessage(WM_KEYDOWN)` 保留一次默认提交语义，另外发送的 `WM_CHAR` 仍属于普通字符输入。

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
- `Toggle`
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
- `Toggle`
- `ComboBox`
- `EditBox`
- `FilePicker`
- `Image`
- `AnimatedImage`
- `ListBox`
- `ProgressBar`
- `ScrollView`

`Button` 支持文本 + 图片内容，图片槽位尺寸由 `ButtonStyle.ImageSizeDP` 控制。

`ModeCustom` 按钮支持 `FocusWidget` / `FocusNext`；`Button.Focused` 表示当前焦点，焦点时在按钮内部按当前文字色和圆角绘制轮廓，适配浅色和实色背景。聚焦且可用时，无 Ctrl/Alt/Shift 的 Enter 或 Space 调用一次原有 `OnClick`。自动重复、组合输入和不可见/禁用祖先不会触发；隐藏或禁用按钮会清除其焦点。默认键处理沿用场景翻译前入口，`OnClick` 切换焦点后也不会把翻译字符插入新编辑框。原生按钮保留 Windows 默认键盘行为。

`ScrollView` 的祖先视口裁剪同样覆盖原生控件 HWND，包括自绘 EditBox 内的 RichEdit 输入层。部分可见时裁剪绘制及鼠标命中，完全离开视口时区域为空；恢复可见或调整视口会重新同步，不缩小编辑器内部文档尺寸、不重建 HWND、不重置选择。裁剪交集复用场景的 `clipBounds`，缓存实际区域避免每帧创建 GDI region。

`Toggle` 是滑动开关控件，仅自绘模式，通过 `ToggleStyle` 可自定义开关轨道色、滑块色以及 ON/OFF 文字颜色；JSON 中用 `type: "toggle"` 声明，样式键含 `trackOn` / `trackOff` / `knob` / `textOn` / `textOff`。

`ListBox` 支持每行独立配色（`ListItem.Background` / `TextColor`）与右侧打勾列（`SetShowCheck` 开启，`ListItem.Checked` 记录勾选，回调 `OnCheckChange`）；勾选与单选选中互不影响，选中/悬停态以半透明叠加，自定义行底色依旧可见。JSON 侧 listbox 用 `check: true` 开启打勾列，每项可声明 `bg` / `fg` / `checked`。

`ListItem.Subtitle` 可在列表行第二行显示。例如设置 `ListStyle{ItemHeightDP: 56}`，默认字体下即可同时显示主标题和副标题。行高不足以容纳两行时仍用原有单行显示；没有副标题的条目垂直居中。副标题复用列表的可见行裁剪、选择、滚动、禁用色、打勾列和双击/Enter 激活，不影响命中区域。

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

`jsonui` 现在还支持组件模板：

- 顶层 `components` 注册组件名到外部 JSON 模板文件
- 节点位置可用 `{ "component": "...", "args": { ... } }` 实例化
- 组件实例还可传 `{ "slots": { "content": [ ... ] } }`，由模板在 `children` 位置显式插入
- 模板文件用 `component` / `params` / `node` 描述一棵可复用子树
- 占位符只支持整值替换，不支持字符串拼接
- 当前只支持命名 `children` slot，不支持条件、循环，组件实例节点也不支持额外 override 字段

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
- `title`
- `tag`
- `value`
- `readOnly`
- `multiline`
- `wordWrap`
- `acceptReturn`
- `verticalScroll`
- `horizontalScroll`
- `items`
- `buttons`
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
- `shape`
- `radius`
- `pad`
- `gap`
- `imageSize`
- `itemH`
- `indicatorStyle`

按钮图片槽位按 contain 方式缩放，不会强行拉伸成正方形。
按钮轮廓可通过 `ButtonStyle.Shape` 指定，当前支持 `ButtonShapeDefault` 和 `ButtonShapePill`；`pill` 会自动使用按钮高度一半作为圆角，适合高曲率胶囊按钮。
按钮边框可通过 `ButtonStyle.Border` 控制；将其显式设为 `0` 可绘制无边框按钮。

### `ComboBox` 富样式扩展

- 新增 `ComboStyle.Layout`，支持 `ComboLayoutDefault` 和 `ComboLayoutRich`。
- `ComboLayoutRich` 只在 `ModeCustom` 下生效，绘制为「左图标 + 主标题 + 副标题 + 下拉选中勾标记」。
- `ListItem` 新增可选字段：
  - `Subtitle`：副标题文本
  - `Image`：条目图标资源
  - `ImagePath`：JSON 字面量声明的图标路径
- `ComboStyle` 新增富样式配套字段：
  - `SubtitleColor`
  - `IconBackground`
  - `SelectedMarkColor`
  - `ImageSizeDP`

## 10. 文件对话框控件

`type: "file"` 映射到 `widgets.FilePicker`。

## 11. DPI 缩放策略

默认情况下，控件行为与旧版本一致：几何尺寸、字体、图片槽位、padding、gap、radius 都按逻辑 DP 参与 DPI 缩放。

如果需要局部混合策略，可以在 Go 代码里调用：

```go
widgets.SetScalePolicy(widget, widgets.ScalePolicy{
    Mode:    widgets.ScaleDP,
    Layout:  widgets.ScalePX,
    Font:    widgets.ScalePX,
    Image:   widgets.ScalePX,
    Padding: widgets.ScalePX,
    Gap:     widgets.ScalePX,
    Radius:  widgets.ScaleDP,
})
```

规则约定：

- `Mode` 是节点默认策略，可向后代继承
- `Layout` 控制该节点自己的布局尺寸：例如 preferred size、绝对布局 `frame`，以及容器自己的 `gap` / `padding` / `itemSize` / `labelWidth`
- `Font` / `Image` / `Padding` / `Gap` / `Radius` 控制渲染侧样式尺寸
- 不设置策略时仍然走原来的 DP 路径

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
go run ./demo/demo_json_full
go run ./demo/demo_go_full
go test ./...
go test -v ./demo/demo_json_full
go test -v ./demo/demo_go_full
go vet ./...
```

## 紧凑知识工作台键盘语义（2026-10-29）

`ListBox.SetShowCheck(true)` 时，聚焦的有效选中行可用不带修饰键的 Space 切换勾选。勾选通过现有 `OnCheckChange` 回调报告，不改变单选项，不触发 `OnChange` 或 `OnActivate`。按键自动重复只消费，不重复切换；Ctrl/Alt/Shift、IME 组合、禁用/隐藏项不参与勾选。该操作复用 Scene 的 TranslateMessage 前默认键阶段；即使回调转焦点到原生编辑器，旧 Space 也不会作为字符漏入新编辑器。应用不应模拟第二套勾选键盘循环。

`EditBox.Composing() bool` 是 UI 线程上的只读原生 IME 状态查询。需要延迟搜索的宿主可在去抖回调中检查它，组合期间等待，组合结束后再按需发起后台工作。它复用已有原生组合状态，不执行网络/文件访问，也不启动额外消息循环。JSON UI 的 ListBox `showCheck` 自动采用上述通用 Space 行为；组合状态查询是 Go 宿主 API，无需新的 JSON 字段。

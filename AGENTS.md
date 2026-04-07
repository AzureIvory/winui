# AGENTS.md

本文件面向在此仓库中协作的 AI 编码代理，目标是帮助其在最短时间内理解项目边界、修改方式和验证要求。

## 1. 项目定位

- 这是一个仅支持 Windows 的 Go UI 工具包，直接构建在 Win32 API 之上。
- 项目目标是提供可复用、可嵌入的小型原生桌面 UI 能力，不承载任何具体业务逻辑。
- 不引入 WebView、XAML 或跨平台抽象层；核心关注 Win32 窗口、绘制、DPI、输入和控件场景树。

## 2. 仓库结构

- `core/`：底层 Win32 封装。负责窗口生命周期、消息循环、绘制、字体、位图、图标、DPI、定时器和输入。
- `widgets/`：构建在 `core` 之上的控件层。负责场景树、事件分发、主题、布局和控件实现。
- `cmd/demo/`：手工验证用的示例应用，是理解 API 用法和回归验证的重要入口。
- `scripts/`：维护脚本，目前包含模块路径调整脚本。
- `README.md`：面向使用者的项目简介和快速开始。
- `DEVELOPING.md`：面向维护者的开发原则和常用命令。
- `WIDGETS.zh-CN.md`：控件、布局、`BindScene`、`LayoutData` 等详细用法文档。

## 3. 技术边界与硬约束

- 只支持 Windows。新增源码文件通常需要显式保留 `//go:build windows`。
- `core` 只能处理 Win32 原语和底层能力，不应引入控件语义。
- `widgets` 不能耦合任何业务逻辑；它应该保持为通用 UI 层。
- 公共 API 设计优先清晰稳定，不要为了抽象而抽象。
- 已有代码和文档以中文注释为主；新增导出 API、关键行为和不直观的辅助函数应继续补充中文注释。

## 4. 核心架构

### 4.1 `core.App`

- `core.NewApp(opts)` 创建应用实例。
- `app.Init()` 启动 UI 线程并创建原生窗口。
- `app.Run()` 进入消息循环。
- `core.Options` 是底层窗口配置与回调集合，包含 `OnCreate`、`OnPaint`、`OnResize`、输入事件、DPI 变化和销毁回调。

### 4.2 `widgets.Scene`

- `widgets.BindScene(&opts, hooks)` 会把场景树接到 `core.Options` 的窗口生命周期和输入回调上。
- `Scene` 是 `widgets` 的运行时核心，管理：
  - 根面板 `Root()`
  - 主题 `Theme()`
  - 焦点、hover、capture
  - 字体缓存
  - 场景级定时器
  - 控件事件分发
- 典型调用链：
  1. `core` 接收 Win32 消息
  2. `BindScene` 包装 `Options` 回调
  3. `Scene` 接收鼠标、键盘、定时器、重绘事件
  4. 事件被分发到目标控件或其祖先链

### 4.3 控件模型

- 所有控件实现 `widgets.Widget`。
- 大部分控件复用 `widgetBase`，它提供：
  - `ID/Bounds/Visible/Enabled/LayoutData`
  - 失效重绘触发
  - 场景和父容器引用
  - UI 线程切换辅助
- 容器控件实现 `Container`，当前最核心的是 `Panel`。
- `Panel` 持有子控件和布局策略，是场景树中的基础容器。

## 5. 布局系统

- `Panel` 通过 `SetLayout(layout)` 应用布局。
- 当前内置布局：
  - `AbsoluteLayout`
  - `RowLayout`
  - `ColumnLayout`
  - `GridLayout`
  - `FormLayout`
  - 兼容接口 `LinearLayout`
- 子控件通过 `SetLayoutData(...)` 提供布局附加参数，常见类型有：
  - `FlexLayoutData`
  - `GridLayoutData`
  - `FormLayoutData`
- 修改布局行为时，优先保持现有 `LayoutData` 约定稳定，并同步检查 `widgets/layout_test.go` 与 `WIDGETS.zh-CN.md`。

## 6. 绘制与渲染

- 默认渲染模式是 `RenderModeAuto`：
  - 优先尝试 Direct2D
  - 初始化失败或运行中失败时自动回退到 GDI
- Direct2D 仅在 `windows && cgo` 下可用，对应：
  - `core/render_d2d.go`
  - `core/render_d2d_bridge.c`
- 当 `cgo` 关闭时，`core/render_d2d_stub.go` 会让后端自动回退到 GDI。
- 因此：
  - 不要假设 Direct2D 必然存在
  - 涉及绘制的变更必须兼容 GDI fallback
  - 若新增底层绘制能力，应评估两个后端的行为是否一致

## 7. DPI、资源与生命周期

- 所有尺寸优先使用 DP 语义，经 `app.DP(...)` 或 `PaintCtx.DP(...)` 转成实际像素。
- `Scene` 会缓存字体，并在主题变化或 DPI 变化后重建资源。
- `Scene.Close()` 负责释放：
  - 场景定时器
  - 控件树中实现了 `Close() error` 的对象
  - 字体缓存
- 新增持有原生资源的控件时，必须考虑：
  - 是否需要实现 `Close() error`
  - 资源是否由控件拥有
  - 在替换资源和场景销毁时是否会泄漏

## 8. UI 线程约束

- `core.App` 维护自己的 UI 线程，跨线程更新需要通过 `app.Post(...)`。
- `widgets` 中多数可变更状态的方法最终会通过 `runOnUI(...)` 切回 UI 线程。
- 新增或修改控件时：
  - 会改动 UI 状态的方法，优先保持和现有控件一致的 `runOnUI(...)` 模式
  - 不要在任意 goroutine 中直接操作依赖窗口句柄或场景状态的对象
  - 改动状态后通常需要调用 `invalidate(...)` 或 `Scene.Invalidate(...)`

## 9. 事件分发约定

- 鼠标事件由 `Scene` 命中测试后路由到目标控件。
- 键盘事件只分发给当前焦点控件。
- `EventClick` 通常由按下/抬起流程组合产生，控件内部状态机要保持自洽。
- 若新增交互控件，至少要检查：
  - hover
  - pressed/down
  - focus/blur
  - enabled/disabled
  - 失效重绘是否正确

## 10. 修改建议

### 10.1 改 `core/` 时

- 先确认修改是否真属于底层原语，而不是应该落在 `widgets/`。
- 不要把控件样式、控件状态机或业务行为塞进 `core`。
- 涉及 Win32 常量、消息、句柄和资源释放时，优先保持显式、直接和可追踪。

### 10.2 改 `widgets/` 时

- 优先复用 `widgetBase`、`Scene`、主题体系和现有事件模型。
- 风格覆盖应采用“默认主题 + 控件局部覆盖”的模式，不要绕开 `resolveStyle` 一类合并逻辑。
- 容器行为变更要注意对子控件的：
  - `scene` 关联
  - `parent` 关联
  - `applyLayout()`
  - `Invalidate()`

### 10.3 新增控件时

- 参考现有控件模式，例如：
  - `button.go`
  - `choice.go`
  - `editbox.go`
  - `listbox.go`
- 通常需要补齐：
  - 构造函数
  - `SetBounds/SetVisible/SetEnabled`
  - 状态更新方法
  - `OnEvent`
  - `Paint`
  - 主题样式结构或样式合并逻辑
  - 必要测试
  - 文档更新

## 11. 测试与验证

在仓库根目录运行：

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
```

建议：

- 纯控件状态机、布局或样式逻辑，优先补 `widgets/*_test.go` 单元测试。
- 涉及视觉效果、事件链路、DPI 或渲染回退时，用 `cmd/demo` 做手工回归。
- 涉及 Direct2D 的改动，要考虑在启用和禁用 `cgo` 两种情况下的行为。

## 12. 文档同步要求

以下改动通常需要同步文档：

- 新增或删除公共控件
- 修改 `BindScene`、布局系统或 `LayoutData` 约定
- 修改公开样式字段
- 修改快速开始示例或推荐用法

同步目标通常是：

- `README.md`
- `WIDGETS.zh-CN.md`
- 必要时为导出符号补充源码注释

## 13. 对 AI 的工作建议

- 先看 `README.md` 和 `DEVELOPING.md`，再定位到具体包。
- 做 API 或结构性变更前，先检查 `cmd/demo/` 是否能体现该能力。
- 优先小步修改，避免在 `core` 和 `widgets` 同时做无关重构。
- 如果需要新增抽象，先确认它是否真的减少重复，且不会让 Win32 行为变得更难追踪。
- 若不确定某个行为是否属于公开约定，先查看：
  - 现有测试
  - `cmd/demo`
  - `WIDGETS.zh-CN.md`

## 14. 当前项目画像

- Go 模块：`github.com/AzureIvory/winui`
- Go 版本：`1.24.0`
- 目标平台：Windows
- 主要包：`core`、`widgets`
- 示例入口：`go run ./cmd/demo`


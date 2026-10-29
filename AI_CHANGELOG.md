# AI Change Log

## 2026-10-29

- 修复线性布局的自然尺寸、受宽度约束测量与实际排版对隐藏控件处理不一致的问题；隐藏项不再增加尺寸或间距。
- 滚动容器测量子节点边界时忽略隐藏项保留的旧几何尺寸，显隐切换后会重新计算双轴范围，无需调用方清空隐藏项边界。
- 新增行、列及线性布局显隐测量和热滚动范围回归，覆盖隐藏后缩回视口及重新显示后恢复范围。

## 2026-10-28

- Added backward-compatible hidden window startup and synchronous, UI-thread
  visibility, physical outer-window bounds, owner and non-activating raise APIs.
  Explicit physical placement is preserved across a resulting DPI change.
  UI-loop shutdown releases pending synchronous operations with `ErrAppClosed`.
- Documented the minimize/destruction lifetime of native owned windows; independent
  companion windows can raise without activation or permanent ownership.
- Added explicit scene widget focus and one consumable keyboard preview path
  before message translation, including native children. Key events expose
  modifiers, repeat and IME composition information; composition candidate input
  bypasses shortcuts and native Enter submission. Submitted Enter no longer also
  inserts its translated newline.
- Added `ListItem.Subtitle` rendering in sufficiently tall `ListBox` rows, sharing
  rich combo text geometry and preserving compact rows and existing interaction.
- Updated README, development, widget and JSON UI documentation. No application
  behavior, business constants or JSON schema extensions were added.
- Review F1 follow-up: moved native default submission after external preview in
  the existing pre-translation path. Removed the single pending-character bool;
  direct native key callers reuse the same submit logic without affecting later
  characters. Native keys route by `KeyEvent.Handle` through the existing registry.
- Resolved IME-wrapped ordinary keys with `ImmGetVirtualKey` before translation;
  active composition and unknown process keys retain IME protection. This fixes
  the source-level path where `VK_PROCESSKEY` could hide an ordinary Enter from
  preview and native submission while the input method was merely open.
- Review F2 follow-up: default native submission preserves Alt+Enter and
  Ctrl+Alt+Enter for ordinary system/native processing unless an external preview
  explicitly consumes them.
- Added opt-in `Scene.FocusNext(reverse bool) bool` for UI-thread form navigation.
  It shares focus eligibility/transfer, follows the existing child tree, wraps
  within the topmost visible enabled modal, and reuses overlay traversal order.
  No default Tab key behavior or product-owned focus order was introduced.
- Made `App.Post` enqueue and notification atomic: a failed PostMessage rolls back
  that callback while holding the queue lock, so an error cannot leave executable
  work behind for a later drain. Documented the failure/acceptance contract.
- Added custom Button focus, inset theme-contrast indication and one-shot
  unmodified Enter/Space activation using existing scene focus and pre-translation
  default-key handling. Repeats/modifiers/IME/inactive ancestors cannot activate;
  native buttons keep Windows handling. No additional key-up loop was introduced.
- Added shared cached native-host region clipping from ancestor viewport bounds
  and visibility, including partial/empty/nested clips and viewport-only changes.
  Hosts are clipped before first show; regions transfer to Windows on success.
- Explicit and traversal focus now reveal targets through existing inner-to-outer
  ScrollView offsets with nearest-edge, stable oversized-target behavior.
- Unified mouse focus with the same real HWND transfer as explicit/traversal
  focus, preventing a custom button's logical focus from leaving keyboard input
  at the previous native editor. External-window Blur remains non-activating.
- Scroll review S1 follow-up: explicit/traversal focus reveals targets, while
  mouse focus uses the same native transfer without moving the clicked geometry
  before pointer-coordinate interpretation.
- Seeded new editor hosts with preset text/caret before registering native change
  notifications, preserving pre-attachment drafts in custom and native modes.

## 紧凑知识工作台键盘语义（2026-10-29）

`ListBox.SetShowCheck(true)` 时，聚焦的有效选中行可用不带修饰键的 Space 切换勾选。勾选通过现有 `OnCheckChange` 回调报告，不改变单选项，不触发 `OnChange` 或 `OnActivate`。按键自动重复只消费，不重复切换；Ctrl/Alt/Shift、IME 组合、禁用/隐藏项不参与勾选。该操作复用 Scene 的 TranslateMessage 前默认键阶段；即使回调转焦点到原生编辑器，旧 Space 也不会作为字符漏入新编辑器。应用不应模拟第二套勾选键盘循环。

`EditBox.Composing() bool` 是 UI 线程上的只读原生 IME 状态查询。需要延迟搜索的宿主可在去抖回调中检查它，组合期间等待，组合结束后再按需发起后台工作。它复用已有原生组合状态，不执行网络/文件访问，也不启动额外消息循环。JSON UI 的 ListBox `showCheck` 自动采用上述通用 Space 行为；组合状态查询是 Go 宿主 API，无需新的 JSON 字段。

## 受宽测量公开入口（2026-10-29）

- 新增 `widgets.MeasureForWidth(widget Widget, width int32) core.Size`，直接复用 `preferredSizeForWidth`，没有第二套换行或测量实现。
- 契约：UI 线程、挂载到 scene 后使用当前 theme/font/DPI；约束与结果均为物理像素；nil 为零，非正宽度为自然测量；显式 preferred 尺寸保持原语义，Label 自动高度应留空 preferred height。
- README、开发指南、控件指南和 JSON UI 指南同步单位换算与使用边界；未新增控件、HWND、JSON 字段或线程调度行为。

## 动态滚动范围与原生横轴（2026-10-29）

- 修正已布局 ScrollView 的后代增删、布局/布局数据、显式几何、preferred size、可见性及 Label 文本/换行/字体格式变更未使祖先测量失效的问题，复用现有父树和 `layoutDirty`；下次使用重新派生横纵范围并夹紧偏移。
- 测量失效时即使内容外框未变也重新安排后代；内部布局及滚动位置不写回 preferred size，纯滚动、悬停和绘制继续复用测量缓存。
- 在既有 `OnMouseWheel` 路径分发 `WM_MOUSEHWHEEL`，`core.MouseEvent` 和 `widgets.Event` 新增默认 false 的 `Horizontal bool`，保留原始 Flags、有符号 Delta 与客户区坐标。横轴正值向右，纵轴正值向上；Shift + 纵滚沿用既有行为。
- 真实横滚只处理横轴，未支持或到边界时冒泡；ListBox/EditBox 的纵向滚轮处理器不再误消费横轴。同步说明默认滚条派生色及白色样式覆盖边界，未新增控件、第二套回调或 JSON 字段。

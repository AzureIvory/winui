# AI Change Log

本文档面向后续协作的 AI / 代理，记录最近一次会影响交互推理、测试设计和局部重绘策略的行为变更。

## 2026-04-01

范围：

- [widgets/combobox.go](./widgets/combobox.go)
- [widgets/scene.go](./widgets/scene.go)
- [widgets/editbox.go](./widgets/editbox.go)
- [widgets/scene_interaction_test.go](./widgets/scene_interaction_test.go)

### ComboBox popup 行为

- `ModeCustom` 下不再固定向下展开。
- popup 位置和高度统一由内部 `popupLayout()` 计算。
- 选择方向规则：
  - 下方空间足够时向下展开。
  - 下方不足且上方空间更大时向上展开。
  - 上下都不足时限制高度到可用空间。
- `popupRect()`、`popupRange()`、`popupIndexAt()`、`popupRowRect()`、`PaintOverlay()`、`overlayHitTest()` 必须和 `popupLayout()` 保持一致。
- 如果继续修改 popup 行为，不要只改绘制或只改命中；这几处是联动点。

### ComboBox overlay 命中优先级

- popup overlay 现在必须始终优先于底层兄弟控件命中。
- `Scene.hitTestOverlay()` 先递归子节点，再检查当前节点 overlay，保证更上层子控件优先。
- `Scene.dispatchMouseDown()` 和 `Scene.dispatchMouseUp()` 会先尝试 `overlayTargetAt()`，避免点击穿透。
- `Scene.mouseTargetAt()` 也会先返回 overlay 命中结果，再做普通 hit test 和稳定化逻辑。
- 后续如果引入新的 overlay 控件，优先复用 `overlayHitWidget` 和 `overlayTargetAt()` 这条路径。

### EditBox hover / IBeam 稳定化

- `Scene.stabilizeMouseTarget()` 不再只依赖“preferred 光标非 Arrow 且 hit 光标为 Arrow”。
- 新增 `stableMouseWidget` 内部协议。
- `EditBox.stableMouseHit()` 表示：只要鼠标仍在当前编辑框有效区域内，优先保持这个 hover 目标。
- 这条规则是为了抑制边缘区域瞬时命中波动导致的 `Arrow` / `IBeam` 抖动。

### dirtyRect / 局部重绘

- `ComboBox` 现在实现了 `dirtyRect()`，返回主框与 popup overlay 的并集。
- `ComboBox` 的状态变更统一通过 `updateState()` + `invalidateStateChange()` 处理。
- 目标：
  - 旧 popup 区域会被单独失效。
  - 新 popup 区域会被单独失效。
  - 避免 open / close / hover 时整窗 `Invalidate(nil)`。
- 后续改 `ComboBox` 状态字段时，优先走 `updateState()`，不要绕回整窗重绘。

### 测试基线

当前回归测试至少覆盖：

- 窗口底部空间不足时 `ComboBox` 自动向上展开。
- popup 限高后不超出 `Scene` 客户区。
- popup 打开后点击可见项不会穿透到底层控件。
- `EditBox` 边缘 hover 在重复 mouse move 下保持稳定。
- `ComboBox` 的 dirty rect 包含 popup overlay。

对应测试文件：

- [widgets/scene_interaction_test.go](./widgets/scene_interaction_test.go)

### 兼容性提示

- 没有引入新的公开 API。
- 变更以 `ModeCustom` 为主。
- 可能影响的仅是更严格的交互行为：
  - 隐藏或禁用中的 `ComboBox` 会同步关闭 popup。
  - 重叠控件场景中，popup overlay 与编辑框 hover 的优先级更稳定。

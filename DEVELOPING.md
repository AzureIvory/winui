# Developing winui

## Goal

`winui` is a Windows-only Go UI toolkit. Changes should keep the project reusable, embeddable, free of app logic, and easy to reason about at the Win32 boundary.

## Package Boundaries

- `core/`: Win32 primitives, window lifecycle, paint, DPI, input, timer
- `sysapi/`: Windows system API helpers, including native file dialogs
- `widgets/`: scene tree, event routing, theme, layout, controls
- `widgets/jsonui/`: declarative JSON loader, bindings, expressions, multi-window helpers
- `demo/demo_json_full/`: full-surface JSON UI regression entry point
- `demo/demo_go_full/`: full-surface Go UI regression entry point

## Rules

- New platform-specific files should usually keep `//go:build windows`
- Do not move widget semantics into `core`
- Do not move app logic into `widgets`
- Keep native open / save / folder COM interop in `sysapi`
- Reuse existing `Scene`, `widgetBase`, layout, and style-merge patterns
- Follow existing UI-thread patterns such as `runOnUI(...)` and `app.Post(...)`
- Trigger invalidation after state changes when needed
- Treat JSON UI frame values as logical DP units, not raw device pixels
- Keep the default DPI path backward compatible: no scale policy should behave exactly like the legacy all-DP model
- Keep DPI policy ownership on widgets; `Mode` may cascade, while a widget's own `Layout` slot controls its preferred size / frame metrics and a container's own `Layout` slot controls that container's gap / padding / item metrics
- Keep JSON frame expressions compatible with the arithmetic parser: `+`, `-`, `*`, `/`, `()`, `%`, and only `winW`, `winH`, `parentW`, `parentH`
- Keep per-window JSON widget ids unique; the loader indexes them for runtime lookup
- Keep JSON image loading DPI-aware through logical DP sizing instead of fixed pixel constants, and keep image slots using contain-style aspect ratio preservation
- Keep image reload policy in runtime hooks such as `Window.ReloadResources(...)`; do not declare callback names in JSON
- Keep hot-reload APIs coarse-grained: whole-window mount / replace / detach, with data reuse but without promising transient widget-state migration
- Keep JSON style fields aligned with existing widget style structs instead of inventing a parallel rendering layer
- Keep data mutation in the host through `jsonui.DataSource`; JSON only declares binding relationships

## Rendering

- Default mode is `RenderModeAuto`
- Prefer Direct2D when available
- Fall back to GDI when Direct2D is unavailable
- Check drawing changes against both paths

## Validation

```powershell
go test ./...
go test -v ./demo/demo_json_full
go test -v ./demo/demo_go_full
go vet ./...
powershell -ExecutionPolicy Bypass -File .\scripts\staticcheck.ps1 -CgoEnabled 0
powershell -ExecutionPolicy Bypass -File .\scripts\staticcheck.ps1 -CgoEnabled 1
go run ./demo/demo_json_full
go run ./demo/demo_go_full
```

GitHub Actions mirrors the build checks on Windows in `.github/workflows/ci.yml` with both `CGO_ENABLED=0` and `CGO_ENABLED=1`, including dedicated `staticcheck`, `demo/demo_json_full`, and `demo/demo_go_full` regression steps, plus the `demo/demo_json_full/output/latest-api-check.txt` artifact upload.

Recommended:

- Use the demos after layout, painting, hit testing, or input-routing changes
- Think about both `cgo` enabled and disabled render paths
- After JSON UI changes, verify that `demo/demo_json_full/demo.ui.json` still loads
- Keep bool field defaults aligned with widget semantics when a JSON field is omitted or a binding has no value: `visible` / `enabled` stay `true`, `checked` / `multiple` / `autoplay` stay `false`
- If you touch DPI-sensitive layout code, verify that JSON absolute expressions still reflow on resize and DPI changes
- If you touch `widgets/jsonui/expr.go`, verify precedence, parentheses, percent semantics, and whitespace-insensitive parsing
- If you touch `sysapi/` or `widgets.FilePicker`, manually verify open, save, folder, and multi-select flows in `demo/demo_json_full`
- If you touch binding code, verify title, text, value, visibility, selection, and frame refresh behavior
- If you touch JSON runtime helpers or action dispatch, verify `Window.FindWidget`, `Document.FindWidget`, `ActionContext.Window`, `MountWindow`, `ReplaceWindow`, and `Detach` behavior
- If you touch modal / overlay behavior, verify backdrop hit testing, card hit testing, and Direct2D fallback behavior together

## Documentation Sync

### Window and keyboard contracts

- `App.Post` returning an error guarantees its callback has not run and will
  never run. Enqueue and nonblocking `PostMessageW` execute under the same queue
  lock; failed notification rolls back only the current callback before a drain
  can see it. A nil return accepts work, including inline execution on the UI
  thread; pump termination may cancel accepted queued callbacks.

- `Options.StartHidden` prevents the creation path from showing the window, even
  when the supplied style includes `WS_VISIBLE`; `OnCreate` and initial layout
  still run.
- `SetScreenBounds` / `ScreenBounds` describe the outer window in physical screen
  pixels. Explicit placement remains authoritative during a resulting DPI change;
  ordinary user moves still apply the recommended `WM_DPICHANGED` rectangle.
- `ShowNoActivate`, `Hide`, `RaiseNoActivate`, `SetScreenBounds`, `ScreenBounds`
  and `SetOwner` execute inline on the UI thread or synchronously through `Post`.
  Call them after `Init`, which starts the pump; `Run` only waits for exit. Keep
  UI callbacks pumping, and do not hold locks that those callbacks require. The
  `stopped` broadcast releases synchronous callers if the UI loop exits before
  their queued operation completes. Win32 BOOL-returning operations propagate
  failures; `ShowWindow` returns previous visibility, rather than an error status.
- `SetOwner` intentionally inherits Win32 owner minimize/destruction behavior.
  Independent companion windows should use `RaiseNoActivate` instead. No helper
  calls `SetForegroundWindow` or makes a normal window topmost.
- Keyboard preview belongs to the core message loop, before `TranslateMessage`,
  for the app HWND and its children. A consumed key is never translated or
  dispatched. Native edit subclasses must not run preview again.
- `BindScene` runs the scene preview before a previously installed core preview;
  `SetKeyPreview` replaces the scene handler, including an initial scene hook.
  If neither preview consumes a native editor submit key, the existing default
  `OnSubmit` runs in this same pre-translation path and consumes it. No translated
  submit character is generated, including with multiple queued keydowns. Native
  subclasses keep the direct `SendMessage(WM_KEYDOWN)` fallback without retaining
  character-suppression state. A separately sent `WM_CHAR` remains ordinary input.
  Unconsumed parent keys retain existing widget and `OnKeyDown` routing. Manual
  `Scene.DispatchKeyDown` also previews once. Synthetic `SendMessage(WM_KEYDOWN)`
  bypasses the queue preview by design; use posted key messages to test this path.
- IMM composition text, unresolved `VK_PROCESSKEY` and native RichEdit composition lifetime
  prevent shortcuts from consuming candidate input. Native Enter submission also
  preserves pre-translation composition state and bypasses composition. Before
  translation, `ImmGetVirtualKey` resolves an IME-wrapped ordinary key so an open
  IME alone does not disable Enter or shortcuts. `KeyEvent.Handle` identifies the
  source HWND; the scene's existing native registry owns native-key routing.
  Scene shortcut handlers bypass composition; low-level
  `Options.OnKeyPreview` receives `Composing` for native coordination and must
  skip its shortcut actions when set (its return cannot consume IME input). Actual IME candidate
  selection should additionally be checked with the user's installed input method.
- `FocusWidget` follows scene scheduling, validates scene membership and visible,
  enabled ancestors, and transfers system focus to registered native controls.
- Explicit, traversal and mouse focus use the same transfer operation: eligible
  native widgets receive HWND focus, custom widgets receive app-host focus.
  Reveal is a separate policy: explicit/traversal callers enable it, mouse callers
  disable it so original pointer coordinates still refer to the clicked geometry.
  `FocusWidget(nil)` clears widget focus and returns system focus to the host;
  `Scene.Blur` remains logical-only for external window focus loss.
- Focus reveal prepares ancestor scroll layouts, then uses existing offsets from
  inner to outer. Each outer reveal receives the portion exposed by inner clips.
  Fitting targets reveal only the hidden edge; oversized targets already covering
  a viewport do not move, and otherwise align the nearest exposed edge. Repeated
  focus is stable. Focus order, eligibility, modal scope and Tab policy stay shared.
- Native host geometry intersects existing ancestor `clipBounds` and visibility,
  in physical scene coordinates, then installs a local outer-HWND rectangle using
  `SetWindowRgn`. Empty regions remove drawing/hit areas without recreating the
  editor or editing selection/text; fully visible hosts clear the region. Native
  creation stays hidden until placement/clipping is synchronized.
- The native state caches bounds, local crop and whether a region is required;
  compare region mode as well as crop (resizing a full host can make the same crop
  partial). Successful `SetWindowRgn` transfers ownership to Windows; delete a
  new region only when installation fails. Never retain an HRGN after success.
  Unchanged geometry creates no region and performs no SetWindowPos/redraw work.
  Recreated hosts reset the value cache. Viewport-only and ancestor visibility
  changes synchronize native descendants even if their own bounds did not move.
- Region/position changes defer drawing until both are current, then invalidate
  client and frame. Existing movement hooks synchronize all native host types;
  editor move-only refreshes preserve its format rectangle, IME and selection.
- New editor hosts receive their cached text/caret before native command routing
  is registered. Geometry/theme setup can emit EN_CHANGE; an empty, registered
  host must never overwrite the preset draft during initialization. Both native
  and custom editor modes use this same creation sequence.
- `FocusNext(reverse bool) bool` is an opt-in helper for UI-thread key previews.
  It shares qualification and native focus transfer with `FocusWidget`, follows
  depth-first actual child order, wraps, and skips foreign/detached widgets and
  invisible/disabled ancestors. Topmost visible enabled `Modal` selection shares
  the same reverse-child traversal as overlay hit testing; no candidates in that
  modal means false, without falling through to background controls. Off-thread
  calls return false without scheduling work. No default Tab binding is changed.
- Default native submit excludes Alt-modified Enter and raw Alt-context flags,
  including Ctrl+Alt+Enter. External scene/core previews still run first and can
  explicitly define those shortcuts. Ordinary Enter/Ctrl+Enter rules are retained.
- Custom Button owns its focus state, keyboard eligibility and activation. It
  participates in the same optional pre-translation default-key stage as native
  editor submission, after external previews. Enter/Space therefore generate no
  trailing character if OnClick changes focus. Direct scene KeyDown uses the same
  handler; Char events never activate. Repeats are consumed without activation;
  Ctrl/Alt/Shift, IME and inactive ancestors cannot activate. Native Button bypasses
  this custom path and keeps Windows defaults. The focus outline stays inside the
  bounds and uses current resolved text color and corner radius, without another
  style source. Hiding/disabling a custom button clears its focus.
- List and rich combo rows share text-line geometry. List subtitles only appear
  when the configured row fits both fonts plus padding; compact rows keep their
  previous rendering. Do not add a second list/input dispatch implementation.

Update docs when public behavior changes, especially for:

- new or removed controls
- `BindScene`, layout, `LayoutData`, or JSON UI DSL changes
- public style fields
- demo usage

Typical files:

- `README.md`
- `WIDGETS.zh-CN.md`
- `JSONUI.zh-CN.md`
- `AGENTS.md`
- `AI_CHANGELOG.md`

## 紧凑知识工作台键盘语义（2026-10-29）

`ListBox.SetShowCheck(true)` 时，聚焦的有效选中行可用不带修饰键的 Space 切换勾选。勾选通过现有 `OnCheckChange` 回调报告，不改变单选项，不触发 `OnChange` 或 `OnActivate`。按键自动重复只消费，不重复切换；Ctrl/Alt/Shift、IME 组合、禁用/隐藏项不参与勾选。该操作复用 Scene 的 TranslateMessage 前默认键阶段；即使回调转焦点到原生编辑器，旧 Space 也不会作为字符漏入新编辑器。应用不应模拟第二套勾选键盘循环。

`EditBox.Composing() bool` 是 UI 线程上的只读原生 IME 状态查询。需要延迟搜索的宿主可在去抖回调中检查它，组合期间等待，组合结束后再按需发起后台工作。它复用已有原生组合状态，不执行网络/文件访问，也不启动额外消息循环。JSON UI 的 ListBox `showCheck` 自动采用上述通用 Space 行为；组合状态查询是 Go 宿主 API，无需新的 JSON 字段。

## Width-constrained measurement

`widgets.MeasureForWidth(Widget, int32) core.Size` is a thin public wrapper over `preferredSizeForWidth`; keep that private path as the single measurement implementation. The constraint and result are physical pixels, including the widget's scene/theme/DPI policy. It is a UI-thread query, not a marshalled operation; attach the widget before measuring when scene metrics matter.

Keep existing natural, nil and explicit preferred-size semantics. In particular, a Label's explicit preferred height wins over measured wrapped height; automatic-height callers must leave it unset. Logical `SetPreferredSize` values must not receive measured pixels without unit conversion. Cover constrained/natural/nil inputs, same-scene metrics, long wrapped text and explicit preferences; measurement evidence alone does not prove backend paint output.

## Scroll measurement invalidation and wheel axes

Use `widgetBase.invalidateMeasurement` for intrinsic content, child-tree, layout or explicit geometry changes. It walks the existing parent tree, invalidates each ancestor ScrollView's existing `layoutDirty` cache and repaints its viewport. Do not route paint-only invalidation through this helper: cached layout must survive hover, paint and offset-only movement. Internal layout/suppressed positioning must not record measured geometry as a new preferred size or invalidate ancestor measurement. A dirty ScrollView reapplies content layout even when the measured outer size is unchanged, so descendant arrangements can update.

Keep `WM_MOUSEWHEEL` and `WM_MOUSEHWHEEL` in the same `OnMouseWheel` path. Copy the independent `Horizontal` field through scene conversion, preserve raw low-word flags, signed delta and client coordinates, and do not encode the axis using Shift. Horizontal positive delta is right; vertical positive delta is up. True horizontal input has no vertical fallback; unsupported consumers return false so existing ancestor routing can handle it. Preserve the existing Shift + vertical-wheel preference and ordinary vertical behavior.

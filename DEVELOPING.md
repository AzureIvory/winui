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
go run ./demo/demo_json_full
go run ./demo/demo_go_full
```

GitHub Actions mirrors the build checks on Windows in `.github/workflows/ci.yml` with both `CGO_ENABLED=0` and `CGO_ENABLED=1`, including dedicated `demo/demo_json_full` and `demo/demo_go_full` regression steps, plus the `demo/demo_json_full/output/latest-api-check.txt` artifact upload.

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

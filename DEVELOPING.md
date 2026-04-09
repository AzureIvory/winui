# Developing winui

## Goal

`winui` is a Windows-only Go UI toolkit. Changes should keep the project:

- reusable
- embeddable
- free of app-specific logic
- easy to reason about at the Win32 boundary

## Package Boundaries

- `core/`: Win32 primitives, window lifecycle, paint, DPI, input, timer
- `sysapi/`: Windows system API helpers, including native file dialogs on top of Win32 COM
- `widgets/`: scene tree, event routing, theme, layout, controls, markup
- `cmd/demo/`: manual regression entry point
- `cmd/demo_html/`: manual markup regression entry point

## Rules

- New platform-specific source files should usually keep `//go:build windows`
- Do not move widget semantics into `core`
- Do not move app logic into `widgets`
- Reuse existing `Scene`, `widgetBase`, layout, and style-merge patterns
- Keep native open/save/folder dialog details in `sysapi/` rather than scattering COM calls through widgets or demos
- Follow existing UI-thread patterns such as `runOnUI(...)` and `app.Post(...)`
- Trigger invalidation after state changes when needed
- Treat markup lengths as logical DP units, not raw device pixels
- Keep markup style fields aligned with existing widget style structs instead of inventing parallel style paths
- Keep declarative markup refresh driven by `markup.State` and `bind-*` attributes instead of per-widget custom observers

## Rendering

- Default mode is `RenderModeAuto`
- Prefer Direct2D when available
- Fall back to GDI when Direct2D is unavailable
- Drawing changes should be checked against both paths

## Validation

The repository now keeps a small set of regression tests, so validation is build checks plus manual demos:

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./cmd/demo_html
```

Recommended:

- Use both demos after layout, painting, hit-testing, or input-routing changes
- Think about both `cgo` enabled and disabled paths for rendering changes
- After markup changes, verify that `cmd/demo_html/demo.ui.html` still loads
- If you touch DPI-sensitive layout code, verify a DPI change still reflows markup-created UIs
- If you touch `sysapi/` or `input type="file"`, manually verify open, save, folder, and multi-select flows in `cmd/demo_html`
- If you touch markup binding code, verify `markup.State` updates still refresh title, value, layout, and list bindings

## Documentation Sync

Update docs when public behavior changes, especially for:

- new or removed controls
- `BindScene`, layout, `LayoutData`, or markup DSL changes
- public style fields
- demo usage

Typical files:

- `README.md`
- `WIDGETS.zh-CN.md`
- `MARKUP_BINDING.zh-CN.md`
- `AGENTS.md`
- `AI_CHANGELOG.md`

## Module Path

If the module path changes:

```powershell
.\scripts\Set-ModulePath.ps1 -ModulePath github.com/<your-org>/<your-repo>
```

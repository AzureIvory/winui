# Developing winui

## Goal

`winui` is a Windows-only Go UI toolkit. Changes should keep the project:

- reusable
- embeddable
- free of app-specific logic
- easy to reason about at the Win32 boundary

## Package Boundaries

- `core/`: Win32 primitives, window lifecycle, paint, DPI, input, timer
- `widgets/`: scene tree, event routing, theme, layout, controls, markup
- `cmd/demo/`: manual regression entry point
- `widgets/cmd/demo_html/`: manual markup regression entry point

## Rules

- New platform-specific source files should usually keep `//go:build windows`
- Do not move widget semantics into `core`
- Do not move app logic into `widgets`
- Reuse existing `Scene`, `widgetBase`, layout, and style-merge patterns
- Follow existing UI-thread patterns such as `runOnUI(...)` and `app.Post(...)`
- Trigger invalidation after state changes when needed

## Rendering

- Default mode is `RenderModeAuto`
- Prefer Direct2D when available
- Fall back to GDI when Direct2D is unavailable
- Drawing changes should be checked against both paths

## Validation

The repository does not keep `*_test.go` files now, so validation is mainly build checks plus manual demos:

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./widgets/cmd/demo_html
```

Recommended:

- Use both demos after layout, painting, hit-testing, or input-routing changes
- Think about both `cgo` enabled and disabled paths for rendering changes
- After markup changes, verify that `widgets/cmd/demo_html/demo.ui.html` still loads

## Documentation Sync

Update docs when public behavior changes, especially for:

- new or removed controls
- `BindScene`, layout, `LayoutData`, or markup DSL changes
- public style fields
- demo usage

Typical files:

- `README.md`
- `WIDGETS.zh-CN.md`
- `AGENTS.md`
- `AI_CHANGELOG.md`

## Module Path

If the module path changes:

```powershell
.\scripts\Set-ModulePath.ps1 -ModulePath github.com/<your-org>/<your-repo>
```

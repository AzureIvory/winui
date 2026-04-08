# AGENTS.md

## Scope

Windows-only Go UI toolkit on top of Win32. No WebView, no XAML, no app logic.

## Packages

- `core`: low-level Win32 window, paint, DPI, input, timer, icon, font
- `widgets`: scene tree, theme, layout, controls, markup
- `cmd/demo`: manual regression demo
- `widgets/cmd/demo_html`: markup demo

## Hard Rules

- New platform files should usually keep `//go:build windows`
- Keep `core` free of widget semantics
- Keep `widgets` free of app/business logic
- Prefer simple, stable public APIs over extra abstraction
- Reuse existing `Scene`, `widgetBase`, layout, and style-merge patterns

## Runtime Facts

- `RenderModeAuto` prefers Direct2D and falls back to GDI
- Changes must work with `cgo` on and off
- UI state changes should follow existing UI-thread patterns: `app.Post(...)` or `runOnUI(...)`
- State mutations usually need invalidation

## Interaction Facts

- `Scene` owns focus, hover, capture, timers, and dispatch
- Keyboard events go to the focused widget
- Mouse routing depends on scene hit testing and overlay handling
- If you change hit testing, check scroll clipping, overlays, and ancestor bounds together

## Markup Facts

- Document API: `LoadDocumentFile`, `LoadDocumentString`, `LoadIntoScene`, `LoadFileIntoScene`
- Legacy API: `LoadHTMLFile`, `LoadHTMLString`
- `<window><body>...</body></window>` is supported
- `WindowMeta` can set title, icon, min width, min height

## Validation

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./widgets/cmd/demo_html
```

`go test` is currently a build check. The repo does not keep `*_test.go` files.

## Docs To Update

- `README.md`
- `DEVELOPING.md`
- `WIDGETS.zh-CN.md`
- `AI_CHANGELOG.md`

## Current Open Risk

- `ScrollView` clip bounds are not yet part of scene hit testing. Changes around pointer routing should account for ancestor clip propagation.

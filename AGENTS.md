# AGENTS.md

## Scope

Windows-only Go UI toolkit on top of Win32. No WebView, no XAML, no app logic.

## Packages

- `core`: low-level Win32 window, paint, DPI, input, timer, icon, font
- `dialogs`: native open/save/folder dialogs on top of Win32 COM
- `widgets`: scene tree, theme, layout, controls, markup
- `cmd/demo`: manual regression demo
- `cmd/demo_html`: markup demo

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
- `dialogs` should stay as the single owner of Win32 file dialog COM interop
- Markup-created preferred sizes are stored as logical DP values and scaled during layout
- `Scene.ReloadResources()` re-applies layout, so DPI changes can move and resize markup-created widgets

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
- `input type="file"` maps to `widgets.FilePicker`
- File input actions surface full selections through `markup.ActionContext.Paths`
- File input supports `dialog="open|save|folder"`, `multiple`, `accept`, `filters`, `button-text`, `dialog-title`, and `default-extension`
- Markup absolute layout supports `left`, `top`, `right`, `bottom`, `width`, `height`, plus `x` / `y`
- Markup style mapping should target existing widget style structs for button, progress, choice, combo, list, edit, and panel controls

## Validation

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./cmd/demo_html
```

`go test` is currently a build check. The repo does not keep `*_test.go` files.

## Docs To Update

- `README.md`
- `DEVELOPING.md`
- `WIDGETS.zh-CN.md`
- `AI_CHANGELOG.md`

## Current Open Risk

- Markup absolute positioning is constraint-based for the supported fields above; it is still not a full CSS box model.

# AGENTS.md

## Scope

Windows-only Go UI toolkit on top of Win32. No WebView, no XAML, no app logic.

## Packages

- `core`: low-level Win32 window, paint, DPI, input, timer, icon, font
- `sysapi`: Windows system API helpers, including native open/save/folder dialogs on top of Win32 COM
- `widgets`: scene tree, theme, layout, controls
- `widgets/jsonui`: JSON schema loader, bindings, expressions, and multi-window helpers
- `cmd/demo`: manual regression demo
- `cmd/demo_json`: JSON UI demo

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
- `sysapi` should stay as the single owner of Win32 file dialog COM interop
- JSON UI frame values are stored as logical DP expressions and resolved during layout
- DPI changes should continue to reflow JSON absolute layouts

## Interaction Facts

- `Scene` owns focus, hover, capture, timers, and dispatch
- Keyboard events go to the focused widget
- Mouse routing depends on scene hit testing and overlay handling
- If you change hit testing, check scroll clipping, overlays, and ancestor bounds together

## JSON UI Facts

- Document API: `LoadDocumentFile`, `LoadDocumentString`, `LoadIntoScene`, `LoadFileIntoScene`
- Multi-window helpers: `Document.NewApps(...)`, `RunApps(...)`
- top-level JSON uses `wins`
- `WindowMeta` can set title, icon, size, and minimum size
- `type: "file"` maps to `widgets.FilePicker`
- file actions surface full selections through `jsonui.ActionContext.Paths`
- absolute `frame` supports `x`, `y`, `r`, `b`, `w`, `h`
- frame expressions support `100`, `50%`, `50%-100`, `winW-100`, `winH-100`, `parentW-100`, `parentH-100`
- JSON style mapping should target existing widget style structs for button, progress, choice, combo, list, edit, and panel controls
- JSON only declares binding relationships; host-side mutation lives in `jsonui.DataSource`

## Validation

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./cmd/demo_json
```

## Docs To Update

- `README.md`
- `DEVELOPING.md`
- `WIDGETS.zh-CN.md`
- `JSONUI.zh-CN.md`
- `AI_CHANGELOG.md`

## Current Open Risk

- JSON absolute layout is still constraint-based for the supported `frame` fields above; it is not a full flexbox or CSS box model replacement.

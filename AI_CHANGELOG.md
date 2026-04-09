# AI Change Log

Only keep behavior here that changes how future edits should be reasoned about.

## 2026-04-09

### JSON UI replaces markup

- `widgets/jsonui` is now the only declarative UI package
- the old HTML / CSS-style `widgets/markup` package was removed
- `cmd/demo_json` replaces `cmd/demo_html`
- declarative documents are now JSON-only

### JSON UI document model

- top-level `wins` declares one or more windows
- `Document.PrimaryWindow()` returns the first window
- `Document.Window(id)` looks up a specific window
- `Document.NewApps(baseOpts)` creates one `core.App` per declared window
- `RunApps(...)` runs hosted windows concurrently and waits for all loops to exit

### Binding model

- JSON only declares binding relationships
- host-side data mutation lives behind `jsonui.DataSource`
- `jsonui.Store` is the default in-memory implementation
- common binding targets include window title, text, value, visible, enabled, checked, items, selection, and absolute `frame`

### DPI-aware frame expressions

- JSON absolute layout is based on logical DP units
- supported expressions are:
  - `100`
  - `50%`
  - `50%-100`
  - `winW-100`
  - `winH-100`
  - `parentW-100`
  - `parentH-100`
- `frame` uses `x`, `y`, `r`, `b`, `w`, `h`

### System API ownership

- `sysapi` remains the single owner of native file dialog COM interop
- JSON `type: "file"` maps to `widgets.FilePicker`, which still delegates to `sysapi`

### JSON UI text/runtime helpers

- `input` and `textarea` now expose `readOnly`, `multiline`, `wordWrap`, `acceptReturn`, `verticalScroll`, and `horizontalScroll`
- widget ids are indexed per window and must stay unique within one window
- `Window.FindWidget(...)`, `Document.FindWidget(...)`, and `widgets.FindByID(...)` are the supported imperative lookup helpers
- `ActionContext.Window` points back to the runtime window that dispatched the action
- JSON `.ico` loading now scales a logical `32dp` default by the current screen DPI and can be overridden with `LoadOptions.IconSizeDP`

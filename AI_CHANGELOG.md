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
- expressions now use an arithmetic parser with `+`, `-`, `*`, `/`, and parentheses
- allowed variables are limited to `winW`, `winH`, `parentW`, and `parentH`
- percent literals such as `50%` keep the legacy axis-based window percentage semantics
- representative expressions include `50%+12` and `(parentW - 12*3 - 20*2 - 108) / 4`
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
- window and widget icon declarations also accept `iconSizeDP` and `iconPolicy: auto | fixed`, and runtime reloads skip fixed-policy icons on DPI changes
- `label` now supports declarative `multiline` / `wordWrap`, and layout can measure multiline height from constrained width
- `Window.ReloadResources(...)`, `MountWindow(...)`, `WindowHost.ReplaceWindow(...)`, and `WindowHost.Detach()` are the supported coarse-grained runtime reload helpers
- JSON `type: "modal"` supports `backdrop.color`, `backdrop.opacity`, `backdrop.blur`, `backdrop.dismissOnClick`, and `onDismiss`; blur is a Direct2D-only soft backdrop pass

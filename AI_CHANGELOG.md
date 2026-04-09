# AI Change Log

Only keep behavior here that changes how an agent should reason about edits or validation.

## 2026-04-09

### Markup state bindings

- `widgets/markup` now supports declarative `bind-*` attributes backed by `markup.State`
- `LoadOptions` includes `State *markup.State`
- `Document` now supports `SetState(...)`, `State()`, and `RefreshBindings(...)`
- supported bindings include:
  - `bind-title` on `<window>`
  - `bind-text`
  - `bind-value`
  - `bind-visible`
  - `bind-enabled`
  - `bind-checked`
  - `bind-width` / `bind-height`
  - `bind-left` / `bind-top` / `bind-right` / `bind-bottom`
  - aliases `bind-x` / `bind-y`
  - `bind-items` / `bind-selected` on `select` and `listbox`
- list bindings accept `[]string`, `[]widgets.ListItem`, and slices of structs or maps with `item-text-field`, `item-value-field`, and `item-disabled-field`
- `markup.State.Set(...)` is map-oriented; prefer `State.Replace(...)` when the source snapshot is easier to rebuild as a struct
- markup text updates do not automatically recompute natural widget size; bind width/height when size must follow data

## 2026-04-08

### Markup input normalization

- `widgets/markup/parser.go` strips a UTF-8 BOM before XML parsing
- `cmd/demo_html/demo.ui.html` may be saved with BOM and should still load

### Markup document API

- Prefer document APIs when window metadata matters:
  - `LoadDocumentFile`
  - `LoadDocumentString`
  - `LoadIntoScene`
  - `LoadFileIntoScene`
- `LoadHTMLFile` and `LoadHTMLString` still work, but only return `Root`
- `<window>` supports `title`, `icon`, `min-width`, `min-height`

### Overlay / popup routing

- ComboBox popup routing depends on overlay-first dispatch in `Scene`
- Popup geometry logic must stay consistent across layout, paint, and hit testing

### Markup layout and DPI

- Markup length values are logical DP units, not raw device pixels
- `widgets.SetPreferredSize(...)` preserves logical preferred sizes for layout-time scaling
- `Scene.ReloadResources()` now re-applies layout so markup UIs reflow after DPI changes
- Markup absolute layout supports `left`, `top`, `right`, `bottom`, `width`, `height`, plus `x` / `y`

### Native file dialogs

- `sysapi` is the Windows system API package; native file dialog COM interop lives there
- `widgets.FilePicker` wraps a readonly field plus browse button and uses `sysapi.ShowFileDialog(...)`
- Markup `input type="file"` supports `dialog="open|save|folder"`, `multiple`, `accept`, `filters`, `button-text`, `dialog-title`, `dialog-button-text`, `default-extension`, `value-separator`, and `initial-path`
- File input selections surface in `markup.ActionContext.Paths`

### Scroll clip routing

- `Scene.hitTest()` now propagates clip bounds from ancestors such as `ScrollView`
- Scrolled-off children should not receive pointer input outside the visible viewport

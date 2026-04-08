# AI Change Log

Only keep behavior here that changes how an agent should reason about edits or validation.

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

### Scroll clip routing

- `Scene.hitTest()` now propagates clip bounds from ancestors such as `ScrollView`
- Scrolled-off children should not receive pointer input outside the visible viewport

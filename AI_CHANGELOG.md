# AI Change Log

Only keep behavior here that changes how an agent should reason about edits or validation.

## 2026-04-08

### Markup input normalization

- `widgets/markup/parser.go` strips a UTF-8 BOM before XML parsing
- `widgets/cmd/demo_html/demo.ui.html` may be saved with BOM and should still load

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

### Open interaction issue

- `ScrollView` clip bounds are exposed, but scene hit testing still does not enforce ancestor clip regions
- Scrolled-off children may still receive pointer input outside the visible viewport

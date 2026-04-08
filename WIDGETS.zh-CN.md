# WIDGETS Guide

This file keeps the legacy name `WIDGETS.zh-CN.md`, but the content is now English for encoding stability and easier maintenance.

## 1. Core Model

`widgets` sits on top of `core`, and `Scene` is the runtime coordinator.

- `widgets.BindScene(&opts, hooks)` connects widget lifecycle to `core.Options`
- `scene.Root()` returns the root panel
- `scene.Theme()` returns the active theme
- `Scene` owns focus, hover, capture, timers, event dispatch, and repaint coordination

Typical flow:

1. Configure `core.Options`
2. Call `widgets.BindScene`
3. Create widgets in `OnCreate`
4. Add them to `scene.Root()`
5. Call `app.Init()` and `app.Run()`

## 2. Control Modes

Some controls accept a `mode` argument:

- `widgets.ModeCustom`: custom-drawn control
- `widgets.ModeNative`: system-native control

Common controls with mode support:

- `Button`
- `EditBox`
- `CheckBox`
- `RadioButton`
- `ComboBox`

If you need themed native controls, the final executable still needs a `Microsoft.Windows.Common-Controls` v6 manifest.

## 3. Layout

`Panel` is the base container. Use `SetLayout(...)` to assign a layout.

Built-in layouts:

- `AbsoluteLayout`
- `RowLayout`
- `ColumnLayout`
- `GridLayout`
- `FormLayout`

Per-child layout data:

- `FlexLayoutData`
- `GridLayoutData`
- `FormLayoutData`

Use cases:

- use `AbsoluteLayout` for explicit coordinates
- use `RowLayout` / `ColumnLayout` for linear toolbars or lists
- use `GridLayout` / `FormLayout` for structured forms

## 4. Common Controls

- `Panel`
- `Label`
- `Button`
- `CheckBox`
- `RadioButton`
- `ComboBox`
- `EditBox`
- `Image`
- `AnimatedImage`
- `ListBox`
- `ProgressBar`
- `ScrollView`

Most control state changes already invalidate automatically. Composite widgets still need to invalidate explicitly when they cache extra state.

## 5. ScrollView

`ScrollView` hosts scrollable content.

```go
content := widgets.NewPanel("content")
content.SetLayout(widgets.ColumnLayout{Gap: 8})

scroll := widgets.NewScrollView("list")
scroll.SetContent(content)
scroll.SetVerticalScroll(true)
scroll.SetHorizontalScroll(false)
```

Current behavior:

- viewport clipping affects both painting and scene hit testing
- scrolled-off children do not receive pointer input outside the visible viewport
- if you change hit testing, keep overlay routing and ancestor clip propagation consistent

## 6. Markup

`widgets/markup` provides an HTML/CSS-style DSL for building UI.

Document APIs:

- `markup.LoadDocumentFile(...)`
- `markup.LoadDocumentString(...)`
- `markup.LoadIntoScene(...)`
- `markup.LoadFileIntoScene(...)`

Legacy APIs:

- `markup.LoadHTMLFile(...)`
- `markup.LoadHTMLString(...)`

`Document` contains:

- `Root widgets.Widget`
- `Meta markup.WindowMeta`

### 6.1 Window Metadata

Supported on `<window>`:

- `title`
- `icon`
- `min-width`
- `min-height`

Example:

```html
<window title="Markup Demo" icon="assets/app.ico" min-width="900" min-height="640">
  <body>
    <label id="title">Hello</label>
  </body>
</window>
```

### 6.2 Common Tag Mapping

- `body` / `div` / `section` / `panel`
- `row` / `column` / `form`
- `scroll`
- `label`
- `button`
- `input`
- `textarea`
- `checkbox`
- `radio`
- `select`
- `listbox`
- `img`
- `animated-img`
- `progress`

### 6.3 Length and DPI

Markup length values such as `20` and `20px` are treated as logical DP units.

- the loader stores logical preferred sizes for markup-created widgets
- layout converts those values with the active scene DPI
- `Scene.ReloadResources()` now re-applies layout, so markup UIs reflow after DPI changes

### 6.4 Absolute Layout

`display:absolute` is a constraint-based layout, not full CSS positioning.

Supported absolute properties:

- `left`
- `top`
- `right`
- `bottom`
- `width`
- `height`
- aliases `x` and `y`

Common combinations:

- `left + top + width + height`
- `left + right + height`
- `top + bottom + width`
- `left + right + top + bottom`

### 6.5 Style Mapping

Markup style fields map into the existing widget theme structs instead of using a separate render layer.

High-coverage mappings exist for:

- `button`
- `progress`
- `checkbox`
- `radio`
- `select`
- `listbox`
- `input`
- `textarea`
- `panel`

Examples of mapped fields include:

- text and placeholder colors
- hover, focus, pressed, and disabled colors
- border and corner radius
- item height, padding, gap, indicator size, and max visible items

### 6.6 Action Routing

`LoadOptions` supports:

- `Actions map[string]func()`
- `ActionHandlers map[string]func(markup.ActionContext)`

Priority:

- `ActionHandlers`
- `Actions`

`ActionContext` includes:

- action name
- widget instance
- widget ID
- current value
- checked state
- index and list item

## 7. Demos

Run both demos when you touch interaction or rendering:

```powershell
go run ./cmd/demo
go run ./cmd/demo_html
```

- `cmd/demo`: core controls, layout, theme, rendering
- `cmd/demo_html`: markup, document metadata, assets, actions, DPI-aware layout, style mapping

## 8. Validation

The repository no longer keeps `*_test.go` files:

```powershell
go test ./...
go vet ./...
```

Treat `go test` as a build check and use the demos for manual regression.

# winui

`winui` is a Windows-only Go UI toolkit built directly on top of Win32.

It targets small native desktop tools that need direct control over window lifecycle, painting, DPI behavior, input, and reusable widgets without relying on WebView, XAML, or cross-platform layers.

## Features

- Windows only
- Clear `core` / `widgets` split
- `RenderModeAuto`: prefer Direct2D, fall back to GDI
- Retained widget scene tree with themes and layouts
- Reusable built-in controls
- Native file dialog API in `sysapi`
- `FilePicker` widget for readonly path display + browse button flows
- HTML/CSS-style markup loader with window metadata support
- Declarative markup bindings via `markup.State` and `bind-*` attributes
- Markup lengths are logical DP values and reflow on DPI changes
- Absolute markup layout supports `left` / `top` / `right` / `bottom` / `width` / `height`
- Markup style mapping covers button, progress, choice, combo, list, edit, and panel styles
- Markup `<input type="file">` supports open / save / folder / multi-select flows
- Native demo app in `cmd/demo`
- Markup demo app in `cmd/demo_html`

## Packages

- `core/`: window lifecycle, paint, DPI, input, timer, icon, font
- `sysapi/`: Windows system API helpers, including native open/save/folder dialogs on top of Win32 COM
- `widgets/`: scene tree, event routing, layout, theme, controls, markup
- `cmd/demo/`: manual regression demo
- `cmd/demo_html/`: markup and document-loading demo
- `scripts/`: maintenance scripts

## Built-in Controls

- `Panel`
- `Label`
- `Button`
- `CheckBox`
- `RadioButton`
- `ComboBox`
- `EditBox`
- `FilePicker`
- `Image`
- `AnimatedImage`
- `ListBox`
- `ProgressBar`
- `ScrollView`

## Requirements

- Windows
- Go `1.24.0` or newer
- Optional: `cgo` for Direct2D support

## Quick Start

```go
package main

import (
	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

func main() {
	opts := core.Options{
		ClassName:      "ExampleApp",
		Title:          "winui example",
		Width:          800,
		Height:         600,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(255, 255, 255),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}

	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate: func(_ *core.App, scene *widgets.Scene) error {
			label := widgets.NewLabel("title", "Hello winui")
			label.SetBounds(widgets.Rect{X: 24, Y: 24, W: 240, H: 32})
			scene.Root().Add(label)
			return nil
		},
	})

	app, err := core.NewApp(opts)
	if err != nil {
		panic(err)
	}
	if err := app.Init(); err != nil {
		panic(err)
	}
app.Run()
}
```

## File Dialogs

Go API:

```go
import (
	"fmt"

	"github.com/AzureIvory/winui/sysapi"
)

result, err := sysapi.ShowFileDialog(app, sysapi.Options{
	Mode:        sysapi.DialogOpen,
	Title:       "Open a file",
	Filters:     []sysapi.FileFilter{{Name: "Text Files", Pattern: "*.txt;*.md"}},
	MultiSelect: true,
})
if err != nil {
	panic(err)
}
fmt.Println(result.Paths)
```

Markup:

```html
<input type="file"
       dialog="save"
       default-extension="txt"
       filters="Text Files=*.txt,All Files=*.*"
       button-text="Save..."
       onchange="savePicked" />
```

## Markup Bindings

`widgets/markup` also supports state-driven bindings for common UI updates such
as window titles, text content, visibility, enabled state, preferred size,
absolute positioning, list items, and selection.

Go:

```go
state := markup.NewState(map[string]any{
	"page": map[string]any{
		"title":   "Search",
		"visible": true,
	},
	"form": map[string]any{
		"query": "initial",
	},
})

doc, err := markup.LoadIntoScene(scene, htmlText, "", markup.LoadOptions{
	State: state,
})
if err != nil {
	panic(err)
}

state.Set("page.title", "Updated Search")
state.Set("form.query", "next value")
_ = doc
```

Markup:

```html
<window bind-title="page.title">
  <body>
    <label bind-text="page.title" bind-visible="page.visible"></label>
    <input bind-value="form.query" />
  </body>
</window>
```

For a Chinese end-user guide focused on markup bindings, see
[MARKUP_BINDING.zh-CN.md](./MARKUP_BINDING.zh-CN.md).

## Run Demos

Core widget demo:

```powershell
go run ./cmd/demo
```

Markup demo:

```powershell
go run ./cmd/demo_html
```

## Validation

The repository now keeps a small set of regression tests. `go test` is still primarily a build-level check plus helper regression coverage.

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./cmd/demo_html
```

GitHub Actions runs the Windows validation path in `.github/workflows/ci.yml`,
including `gofmt`, `go test ./...`, and `go vet ./...` with `CGO_ENABLED=0`
and `CGO_ENABLED=1`.

## Docs

- [DEVELOPING.md](./DEVELOPING.md): maintainer rules, architecture boundaries, validation
- [WIDGETS.zh-CN.md](./WIDGETS.zh-CN.md): widget, layout, `BindScene`, and markup guide
- [MARKUP_BINDING.zh-CN.md](./MARKUP_BINDING.zh-CN.md): Chinese guide for `markup.State` and `bind-*`
- [AGENTS.md](./AGENTS.md): compact agent-facing repository guide
- [AI_CHANGELOG.md](./AI_CHANGELOG.md): recent behavior changes that affect agent reasoning

## Native Mode Note

`Button`, `EditBox`, `CheckBox`, `RadioButton`, `ComboBox`, and `FilePicker` can switch between custom-drawn and native-system backends via the `mode` parameter.

If you want `ModeNative` controls to render with Win10/Win11 visual styles, the final executable still needs a `Microsoft.Windows.Common-Controls` v6 manifest.

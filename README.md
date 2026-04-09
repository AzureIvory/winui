# winui

`winui` is a Windows-only Go UI toolkit built directly on top of Win32.

It targets native desktop tools that need explicit control over window lifecycle, painting, DPI, input, reusable widgets, and a declarative JSON UI layer without WebView, XAML, or cross-platform wrappers.

## Features

- Windows only
- Clear `core` / `widgets` / `sysapi` split
- `RenderModeAuto`: prefer Direct2D, fall back to GDI
- Retained widget scene tree with themes and layouts
- Reusable built-in controls
- Native open / save / folder dialogs in `sysapi`
- Declarative JSON UI loader in `widgets/jsonui`
- DPI-aware JSON frame expressions such as `100`, `50%`, `50%-100`, `winW-100`, `parentW-100`
- State-driven bindings through `jsonui.DataSource`
- Single-window and multi-window helpers
- Demo apps in `cmd/demo` and `cmd/demo_json`

## Packages

- `core/`: window lifecycle, paint, DPI, input, timer, icon, font
- `sysapi/`: Windows system API helpers, including native file dialogs
- `widgets/`: scene tree, event routing, layout, theme, controls
- `widgets/jsonui/`: JSON schema loader, bindings, expressions, multi-window helpers
- `cmd/demo/`: manual widget regression demo
- `cmd/demo_json/`: manual JSON UI regression demo

## Quick Start

Imperative widgets:

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

JSON UI:

```go
store := jsonui.NewStore(map[string]any{
	"page": map[string]any{
		"title": "JSON Demo",
	},
})

win, err := jsonui.LoadFileIntoScene(scene, "demo.ui.json", jsonui.LoadOptions{
	Data: store,
})
if err != nil {
	panic(err)
}

store.Set("page.title", "Updated Title")
_ = win
```

```json
{
  "wins": [
    {
      "id": "main",
      "title": { "bind": "page.title", "default": "Fallback" },
      "w": 980,
      "h": 720,
      "root": {
        "type": "panel",
        "layout": "abs",
        "children": [
          {
            "type": "label",
            "id": "title",
            "text": { "bind": "page.title" },
            "frame": { "x": 20, "y": 20, "w": 320, "h": 28 }
          }
        ]
      }
    }
  ]
}
```

## JSON UI Model

- Top-level `wins` declares one or more windows
- JSON declares widget trees, styles, actions, and bindings
- Host code owns all data mutation through `jsonui.DataSource`
- Frame values are logical DP by default and scale with DPI
- `frame` supports `x`, `y`, `r`, `b`, `w`, `h`
- Expressions support:
  - `100`
  - `"50%"`
  - `"50%-100"`
  - `"winW-100"`
  - `"winH-100"`
  - `"parentW-100"`
  - `"parentH-100"`

## File Dialogs

Go API:

```go
result, err := sysapi.ShowFileDialog(app, sysapi.Options{
	Mode:        sysapi.DialogOpen,
	Title:       "Open a file",
	Filters:     []sysapi.FileFilter{{Name: "Text Files", Pattern: "*.txt;*.md"}},
	MultiSelect: true,
})
if err != nil {
	panic(err)
}
_ = result.Paths
```

JSON UI:

```json
{
  "type": "file",
  "id": "openFile",
  "dialog": "open",
  "buttonText": "Open...",
  "dialogTitle": "Open a source file",
  "accept": ["*.txt", "*.md", "*.go"]
}
```

## Multi-Window

`jsonui.Document` keeps every declared window.

- `doc.PrimaryWindow()` returns the first one
- `doc.Window("tools")` looks up by id
- `doc.NewApps(baseOpts)` creates one `core.App` per window
- `jsonui.RunApps(...)` starts every hosted window and waits for all loops to exit

## Run Demos

```powershell
go run ./cmd/demo
go run ./cmd/demo_json
```

## Validation

```powershell
go test ./...
go vet ./...
go run ./cmd/demo
go run ./cmd/demo_json
```

GitHub Actions mirrors the Windows validation path in `.github/workflows/ci.yml` with `gofmt`, `go test ./...`, and `go vet ./...` under both `CGO_ENABLED=0` and `CGO_ENABLED=1`.

## Docs

- [DEVELOPING.md](./DEVELOPING.md): maintainer rules, architecture boundaries, validation
- [WIDGETS.zh-CN.md](./WIDGETS.zh-CN.md): widget and JSON UI overview
- [JSONUI.zh-CN.md](./JSONUI.zh-CN.md): Chinese end-user guide for the JSON DSL
- [AGENTS.md](./AGENTS.md): compact repository guide
- [AI_CHANGELOG.md](./AI_CHANGELOG.md): behavior changes that affect future edits

## Native Mode Note

`Button`, `EditBox`, `CheckBox`, `RadioButton`, `ComboBox`, and `FilePicker` can switch between custom-drawn and native-system backends via the `mode` parameter.

If you want `ModeNative` controls to render with Win10/Win11 visual styles, the final executable still needs a `Microsoft.Windows.Common-Controls` v6 manifest.

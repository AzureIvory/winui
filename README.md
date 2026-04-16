# winui

`winui` is a Windows-only Go UI toolkit built directly on top of Win32.

It targets native desktop tools that need explicit control over window lifecycle, painting, DPI, input, reusable widgets, and a declarative JSON UI layer without WebView, XAML, or cross-platform wrappers.

## Features

- Windows only
- Clear `core` / `widgets` / `sysapi` split
- `RenderModeAuto`: prefer Direct2D, fall back to GDI
- Direct2D text rendering keeps Windows color fonts such as emoji when `cgo` is enabled
- Window and button image resources accept PNG / JPG / JPEG / GIF (GIF uses the first frame only); window images still become native `HICON` internally
- Retained widget scene tree with themes and layouts
- Reusable built-in controls
- Native open / save / folder dialogs in `sysapi`
- Declarative JSON UI loader in `widgets/jsonui`
- DPI-aware JSON frame expressions with `+`, `-`, `*`, `/`, `()`, `%`, and window/parent size variables
- State-driven bindings through `jsonui.DataSource`
- JSON text inputs with `readOnly`, `multiline`, wrapping, and scroll flags
- Declarative multiline `label` with width-constrained auto-height measurement
- Declarative `modal` / backdrop support with Direct2D-only blur tint
- Declarative `scrollview` for JSON-authored nested scrolling surfaces
- Runtime lookup helpers such as `Window.FindWidget`, `Document.FindWidget`, and `widgets.FindByID`
- Single-window and multi-window helpers
- Demo apps in `demo/demo_json_full` and `demo/demo_go_full`

## Packages

- `core/`: window lifecycle, paint, DPI, input, timer, image, icon, font
- `sysapi/`: Windows system API helpers, including native file dialogs
- `widgets/`: scene tree, event routing, layout, theme, controls
- `widgets/jsonui/`: JSON schema loader, bindings, expressions, multi-window helpers
- `demo/demo_json_full/`: full-surface JSON UI demo with palette switching and API checks
- `demo/demo_go_full/`: full-surface Go UI demo with palette switching, mode toggling, and API checks

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

title := win.FindWidget("title")
_ = title

store.Set("page.title", "Updated Title")
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
- Widget ids must stay unique within each declared window
- Bool fields fall back to widget semantics when omitted or when a bound value is missing: `visible` / `enabled` stay `true`, while `checked`, `multiple`, and `autoplay` stay `false`
- `input` / `textarea` support `readOnly`, `multiline`, `wordWrap`, `acceptReturn`, `verticalScroll`, and `horizontalScroll`
- `window.image`, `button.image`, and `image` controls accept static PNG / JPG / JPEG / GIF input; GIF uses the first frame only
- `LoadOptions.ImageSizeDP`, per-window `imageSizeDP`, per-node `imageSizeDP`, and `style.imageSize` control the image slot size
- Image rendering keeps the original aspect ratio and fits with contain semantics instead of stretching to a square
- Images are cached by target pixel size and quality; Direct2D bitmaps are preferred, with GDI as fallback when needed
- Use `animimg` when you need animated GIF playback
- The loader does not guarantee BMP, SVG, WEBP, AVIF, or ico-specific loading semantics
- `label` supports `multiline` and `wordWrap`, including auto-measured height when width is constrained
- `modal` supports `backdrop.color`, `backdrop.opacity`, `backdrop.blur`, `backdrop.dismissOnClick`, and `onDismiss`
- `frame` supports `x`, `y`, `r`, `b`, `w`, `h`
- Expressions support integer arithmetic with `+`, `-`, `*`, `/`, and parentheses
- Variables are limited to `winW`, `winH`, `parentW`, and `parentH`
- Percent literals such as `"50%"` keep the legacy axis-based window percentage semantics
- Example expressions include:
  - `100`
  - `"50%+12"`
  - `"(parentW - 12*3 - 20*2 - 108) / 4"`
  - `"(parentW-184)/4"`

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
- `win.FindWidget("status")` looks up a widget inside one window
- `doc.FindWidget("main", "status")` looks up across windows
- `ActionContext.Window` points back to the runtime window in action handlers
- `doc.NewApps(baseOpts)` creates one `core.App` per window
- `jsonui.MountWindow(scene, win)` creates a `WindowHost` for mount / replace / detach hot reload flows
- `jsonui.RunApps(...)` starts every hosted window and waits for all loops to exit

## Run Demos

```powershell
go run ./demo/demo_json_full
go run ./demo/demo_go_full
```

## Validation

```powershell
go test ./...
go test -v ./demo/demo_json_full
go test -v ./demo/demo_go_full
go vet ./...
go run ./demo/demo_json_full
go run ./demo/demo_go_full
```

GitHub Actions mirrors the Windows validation path in `.github/workflows/ci.yml` with `gofmt`, dedicated `go test -v ./demo/demo_json_full` and `go test -v ./demo/demo_go_full` regression steps, an upload of `demo/demo_json_full/output/latest-api-check.txt`, plus `go test ./...` and `go vet ./...` under both `CGO_ENABLED=0` and `CGO_ENABLED=1`.

## Docs

- [DEVELOPING.md](./DEVELOPING.md): maintainer rules, architecture boundaries, validation
- [WIDGETS.zh-CN.md](./WIDGETS.zh-CN.md): widget and JSON UI overview
- [JSONUI.zh-CN.md](./JSONUI.zh-CN.md): Chinese end-user guide for the JSON DSL
- [AGENTS.md](./AGENTS.md): compact repository guide
- [AI_CHANGELOG.md](./AI_CHANGELOG.md): behavior changes that affect future edits

## Native Mode Note

`Button`, `EditBox`, `CheckBox`, `RadioButton`, `ComboBox`, and `FilePicker` can switch between custom-drawn and native-system backends via the `mode` parameter.

If you want `ModeNative` controls to render with Win10/Win11 visual styles, the final executable still needs a `Microsoft.Windows.Common-Controls` v6 manifest.

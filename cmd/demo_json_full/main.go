//go:build windows

package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)

	doc, store, err := loadDemoDocument(baseDir)
	if err != nil {
		panic(err)
	}
	window := doc.PrimaryWindow()
	if window == nil {
		panic("primary window is nil")
	}

	controller := newDemoController(baseDir, store, doc, window)

	opts := core.Options{
		ClassName:      "WinUIJSONFullDemo",
		Title:          "WinUI JSON Full Demo",
		Width:          1380,
		Height:         940,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(244, 247, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	window.ApplyOptions(&opts)

	widgets.BindScene(&opts, widgets.SceneHooks{
		Theme: demoTheme(),
		OnCreate: func(app *core.App, scene *widgets.Scene) error {
			if err := window.Attach(scene); err != nil {
				return err
			}
			if window.Root != nil {
				size := app.ClientSize()
				window.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
			controller.setStatus("UI ready")
			return nil
		},
		OnResize: func(_ *core.App, _ *widgets.Scene, size core.Size) {
			if window.Root != nil {
				window.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
			}
		},
		OnDPIChanged: func(_ *core.App, _ *widgets.Scene, _ core.DPIInfo) {
			_ = window.ReloadResources(jsonui.ReloadReasonDPIChanged)
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

func loadDemoDocument(baseDir string) (*jsonui.Document, *jsonui.Store, error) {
	assetsDir := filepath.Join(baseDir, "assets")
	if err := ensureDemoAssets(assetsDir); err != nil {
		return nil, nil, err
	}

	store := newDemoStore()
	doc, err := jsonui.LoadDocumentFile(filepath.Join(baseDir, "demo.ui.json"), jsonui.LoadOptions{
		AssetsDir:   baseDir,
		DefaultMode: widgets.ModeCustom,
		Data:        store,
		Theme:       demoTheme(),
	})
	if err != nil {
		return nil, nil, err
	}
	return doc, store, nil
}

func newDemoStore() *jsonui.Store {
	return jsonui.NewStore(map[string]any{
		"demo": map[string]any{
			"windowTitle":          "WinUI JSON Full Demo",
			"paletteName":          "Ocean Blue",
			"report":               "",
			"reportSummary":        defaultReportSummary(),
			"reportPath":           defaultReportPath(),
			"lastAction":           "Ready",
			"modalVisible":         false,
			"showVerticalScroll":   true,
			"showHorizontalScroll": true,
		},
	})
}

func demoTheme() *widgets.Theme {
	theme := widgets.DefaultTheme()
	theme.BackgroundColor = core.RGB(244, 247, 251)
	theme.Text.Color = core.RGB(22, 31, 47)
	theme.Title.Color = core.RGB(15, 23, 42)
	theme.Button.Background = core.RGB(245, 249, 255)
	theme.Button.Hover = core.RGB(226, 240, 255)
	theme.Button.Pressed = core.RGB(37, 99, 235)
	theme.Button.Border = core.RGB(173, 201, 236)
	theme.Edit.FocusBorder = core.RGB(37, 99, 235)
	theme.ComboBox.FocusBorder = core.RGB(37, 99, 235)
	theme.ListBox.FocusBorder = core.RGB(37, 99, 235)
	return theme
}

func defaultReportSummary() string {
	return "No API check has been run yet."
}

func defaultReportPath() string {
	return "output\\latest-api-check.txt"
}

func ensureDemoAssets(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	files := map[string][]byte{
		"app.ico":     buildICO(color.RGBA{R: 47, G: 109, B: 214, A: 255}),
		"save.ico":    buildICO(color.RGBA{R: 20, G: 184, B: 166, A: 255}),
		"palette.ico": buildICO(color.RGBA{R: 107, G: 114, B: 128, A: 255}),
		"preview.png": buildPreviewPNG(),
		"spinner.gif": buildSpinnerGIF(),
	}

	for name, data := range files {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func buildPreviewPNG() []byte {
	const width = 128
	const height = 96

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(212 + x/6)
			g := uint8(230 - y/4)
			b := uint8(246 - x/7)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	for y := 18; y < 78; y++ {
		for x := 18; x < 110; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	for y := 28; y < 68; y++ {
		for x := 28; x < 58; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 47, G: 109, B: 214, A: 255})
		}
	}
	for y := 34; y < 44; y++ {
		for x := 70; x < 100; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 120, G: 144, B: 181, A: 255})
		}
	}
	for y := 50; y < 58; y++ {
		for x := 70; x < 92; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 181, G: 195, B: 216, A: 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func buildSpinnerGIF() []byte {
	palette := color.Palette{
		color.RGBA{0, 0, 0, 0},
		color.RGBA{47, 109, 214, 255},
		color.RGBA{20, 184, 166, 255},
		color.RGBA{148, 163, 184, 255},
	}

	frames := make([]*image.Paletted, 0, 3)
	delays := []int{8, 8, 8}
	for index := 0; index < 3; index++ {
		frame := image.NewPaletted(image.Rect(0, 0, 24, 24), palette)
		drawSpinnerFrame(frame, uint8(index+1))
		frames = append(frames, frame)
	}

	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0,
	})
	return buf.Bytes()
}

func drawSpinnerFrame(img *image.Paletted, highlight uint8) {
	if img == nil {
		return
	}
	ring := []struct {
		x int
		y int
	}{
		{11, 2},
		{18, 6},
		{21, 12},
		{18, 18},
		{11, 21},
		{5, 18},
		{2, 12},
		{5, 6},
	}
	for index, point := range ring {
		colorIndex := uint8(3)
		if uint8(index%3)+1 == highlight {
			colorIndex = 1
		} else if uint8((index+1)%3)+1 == highlight {
			colorIndex = 2
		}
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				x := point.x + dx
				y := point.y + dy
				if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
					continue
				}
				if dx*dx+dy*dy <= 4 {
					img.SetColorIndex(x, y, colorIndex)
				}
			}
		}
	}
}

func buildICO(fill color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	for y := 3; y < 13; y++ {
		for x := 3; x < 13; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: fill.R + (255-fill.R)/3,
				G: fill.G + (255-fill.G)/3,
				B: fill.B + (255-fill.B)/3,
				A: 255,
			})
		}
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	maskStride := ((w + 31) / 32) * 4
	maskSize := maskStride * h
	bmpSize := 40 + w*h*4 + maskSize

	data := make([]byte, 6+16+bmpSize)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint16(data[4:], 1)

	entry := data[6:22]
	entry[0] = byte(w)
	entry[1] = byte(h)
	binary.LittleEndian.PutUint16(entry[4:], 1)
	binary.LittleEndian.PutUint16(entry[6:], 32)
	binary.LittleEndian.PutUint32(entry[8:], uint32(bmpSize))
	binary.LittleEndian.PutUint32(entry[12:], 22)

	bmp := data[22:]
	binary.LittleEndian.PutUint32(bmp[0:], 40)
	binary.LittleEndian.PutUint32(bmp[4:], uint32(w))
	binary.LittleEndian.PutUint32(bmp[8:], uint32(h*2))
	binary.LittleEndian.PutUint16(bmp[12:], 1)
	binary.LittleEndian.PutUint16(bmp[14:], 32)
	binary.LittleEndian.PutUint32(bmp[20:], uint32(w*h*4))

	pixelOffset := 40
	index := 0
	for y := h - 1; y >= 0; y-- {
		row := img.Pix[y*img.Stride:]
		for x := 0; x < w; x++ {
			src := x * 4
			dst := pixelOffset + index*4
			data[22+dst] = row[src+2]
			data[22+dst+1] = row[src+1]
			data[22+dst+2] = row[src]
			data[22+dst+3] = row[src+3]
			index++
		}
	}
	return data
}

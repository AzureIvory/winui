//go:build windows

package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type demoUI struct {
	app   *core.App
	scene *widgets.Scene

	title       *widgets.Label
	renderer    *widgets.Label
	status      *widgets.Label
	progressPct *widgets.Label
	progress    *widgets.ProgressBar
	check       *widgets.CheckBox
	radioA      *widgets.RadioButton
	radioB      *widgets.RadioButton
	list        *widgets.ListBox
	combo       *widgets.ComboBox
	edit        *widgets.EditBox
	btnStep     *widgets.Button
	btnReset    *widgets.Button

	stepIcon *core.Icon
}

var ui demoUI

func main() {
	app, err := core.NewApp(core.Options{
		ClassName:      "WinUIDemo",
		Title:          "winui demo",
		Width:          760,
		Height:         560,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(248, 250, 252),
		DoubleBuffered: true,
		RenderMode:     demoRenderMode(),
		OnCreate:       onCreate,
		OnPaint:        onPaint,
		OnResize:       onResize,
		OnMouseMove:    onMouseMove,
		OnMouseLeave:   onMouseLeave,
		OnMouseDown:    onMouseDown,
		OnMouseUp:      onMouseUp,
		OnKeyDown:      onKeyDown,
		OnChar:         onChar,
		OnFocus:        onFocus,
		OnTimer:        onTimer,
		OnDPIChanged:   onDPIChanged,
		OnDestroy:      onDestroy,
	})
	if err != nil {
		panic(err)
	}

	ui.app = app
	if err := app.Init(); err != nil {
		panic(err)
	}
	app.Run()
}

// onCreate 初始化示例界面。
func onCreate(app *core.App) error {
	ui.app = app
	ui.scene = widgets.NewScene(app)

	ui.title = widgets.NewLabel("title", "winui widgets demo")
	ui.title.SetStyle(widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 22,
			Weight: 700,
		},
		Color:  core.RGB(15, 23, 42),
		Format: core.DTVCenter | core.DTSingleLine,
	})

	ui.renderer = widgets.NewLabel("renderer", renderSummary(app))
	ui.renderer.SetStyle(widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 13,
			Weight: 700,
		},
		Color:  core.RGB(37, 99, 235),
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	ui.status = widgets.NewLabel("status", "Ready")
	ui.status.SetStyle(widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		Color:  core.RGB(71, 85, 105),
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	ui.progressPct = widgets.NewLabel("progress-text", "")
	ui.progressPct.SetStyle(widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 14,
			Weight: 700,
		},
		Color:  core.RGB(5, 150, 105),
		Format: core.DTVCenter | core.DTSingleLine,
	})

	ui.progress = widgets.NewProgressBar("progress")
	ui.progress.SetStyle(widgets.ProgressStyle{
		FillColor:    core.RGB(16, 185, 129),
		BubbleColor:  core.RGB(5, 150, 105),
		TrackColor:   core.RGB(229, 231, 235),
		TextColor:    core.RGB(255, 255, 255),
		CornerRadius: 10,
		ShowPercent:  true,
	})
	setProgress(35)

	ui.check = widgets.NewCheckBox("check", "Enable safe mode")
	ui.check.SetStyle(widgets.ChoiceStyle{
		HoverBorder:     core.RGB(20, 184, 166),
		FocusBorder:     core.RGB(13, 148, 136),
		IndicatorColor:  core.RGB(13, 148, 136),
		HoverBackground: core.RGB(240, 253, 250),
	})
	ui.check.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Safe mode enabled")
			return
		}
		ui.status.SetText("Safe mode disabled")
	})

	ui.radioA = widgets.NewRadioButton("radio-quick", "Quick install")
	ui.radioA.SetGroup("mode")
	ui.radioA.SetStyle(widgets.ChoiceStyle{
		HoverBorder:     core.RGB(249, 115, 22),
		FocusBorder:     core.RGB(234, 88, 12),
		IndicatorColor:  core.RGB(234, 88, 12),
		HoverBackground: core.RGB(255, 247, 237),
	})
	ui.radioA.SetChecked(true)
	ui.radioA.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Quick install selected")
		}
	})

	ui.radioB = widgets.NewRadioButton("radio-full", "Full install")
	ui.radioB.SetGroup("mode")
	ui.radioB.SetStyle(ui.radioA.Style)
	ui.radioB.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Full install selected")
		}
	})

	ui.list = widgets.NewListBox("list")
	ui.list.SetStyle(widgets.ListStyle{
		HoverBorder:       core.RGB(96, 165, 250),
		FocusBorder:       core.RGB(37, 99, 235),
		ItemHoverColor:    core.RGB(239, 246, 255),
		ItemSelectedColor: core.RGB(37, 99, 235),
	})
	ui.list.SetItems([]widgets.ListItem{
		{Value: "windows-10", Text: "Windows 10"},
		{Value: "windows-11", Text: "Windows 11"},
		{Value: "windows-pe", Text: "Windows PE"},
	})
	ui.list.SetSelected(1)
	ui.list.SetOnChange(func(_ int, item widgets.ListItem) {
		ui.status.SetText("List item: " + displayItem(item))
	})

	ui.combo = widgets.NewComboBox("combo")
	ui.combo.SetStyle(widgets.ComboStyle{
		HoverBorder:       core.RGB(251, 191, 36),
		FocusBorder:       core.RGB(249, 115, 22),
		ArrowColor:        core.RGB(249, 115, 22),
		ItemHoverColor:    core.RGB(255, 247, 237),
		ItemSelectedColor: core.RGB(249, 115, 22),
	})
	ui.combo.SetPlaceholder("Select accent color")
	ui.combo.SetItems([]widgets.ListItem{
		{Value: "blue", Text: "Blue"},
		{Value: "green", Text: "Green"},
		{Value: "orange", Text: "Orange"},
	})
	ui.combo.SetSelected(0)
	ui.combo.SetOnChange(func(_ int, item widgets.ListItem) {
		ui.status.SetText("Accent color: " + displayItem(item))
	})

	ui.edit = widgets.NewEditBox("edit")
	ui.edit.SetStyle(widgets.EditStyle{
		HoverBorder: core.RGB(125, 211, 252),
		FocusBorder: core.RGB(14, 165, 233),
		CaretColor:  core.RGB(14, 165, 233),
	})
	ui.edit.SetPlaceholder("Type a custom environment label")
	ui.edit.SetOnChange(func(text string) {
		if text == "" {
			ui.status.SetText("Input cleared")
			return
		}
		ui.status.SetText("Label: " + text)
	})

	ui.stepIcon = demoBadge(
		color.RGBA{R: 37, G: 99, B: 235, A: 255},
		color.RGBA{R: 191, G: 219, B: 254, A: 255},
	)

	ui.btnStep = widgets.NewButton("step", "Advance")
	ui.btnStep.SetKind(widgets.BtnLeft)
	ui.btnStep.SetIcon(ui.stepIcon)
	ui.btnStep.SetStyle(widgets.ButtonStyle{
		TextColor:    core.RGB(30, 64, 175),
		DownText:     core.RGB(30, 64, 175),
		Background:   core.RGB(239, 246, 255),
		Hover:        core.RGB(219, 234, 254),
		Pressed:      core.RGB(191, 219, 254),
		Border:       core.RGB(96, 165, 250),
		CornerRadius: 10,
		IconSizeDP:   18,
		GapDP:        10,
		PadDP:        14,
	})
	ui.btnStep.SetOnClick(func() {
		next := ui.progress.Value() + 10
		if next > 100 {
			next = 100
		}
		setProgress(next)
		ui.status.SetText("Progress updated")
	})

	ui.btnReset = widgets.NewButton("reset", "Reset")
	ui.btnReset.SetStyle(widgets.ButtonStyle{
		TextColor:    core.RGB(194, 65, 12),
		DownText:     core.RGB(255, 255, 255),
		Background:   core.RGB(255, 247, 237),
		Hover:        core.RGB(254, 215, 170),
		Pressed:      core.RGB(249, 115, 22),
		Border:       core.RGB(251, 146, 60),
		CornerRadius: 10,
		PadDP:        14,
	})
	ui.btnReset.SetOnClick(func() {
		setProgress(0)
		ui.status.SetText("Progress reset")
	})

	root := ui.scene.Root()
	root.Add(ui.title)
	root.Add(ui.renderer)
	root.Add(ui.status)
	root.Add(ui.progressPct)
	root.Add(ui.progress)
	root.Add(ui.check)
	root.Add(ui.radioA)
	root.Add(ui.radioB)
	root.Add(ui.list)
	root.Add(ui.combo)
	root.Add(ui.edit)
	root.Add(ui.btnStep)
	root.Add(ui.btnReset)

	size := app.ClientSize()
	layout(size.Width, size.Height)
	return nil
}

// onPaint 绘制当前场景。
func onPaint(_ *core.App, canvas *core.Canvas) {
	if ui.scene != nil {
		ui.scene.PaintCore(canvas)
	}
}

// onResize 按窗口尺寸重新布局。
func onResize(_ *core.App, size core.Size) {
	if ui.scene == nil {
		return
	}
	ui.scene.Resize(core.Rect{X: 0, Y: 0, W: size.Width, H: size.Height})
	layout(size.Width, size.Height)
}

// onMouseMove 分发鼠标移动事件。
func onMouseMove(_ *core.App, ev core.MouseEvent) {
	if ui.scene != nil {
		ui.scene.DispatchMouseMove(ev)
	}
}

// onMouseLeave 分发鼠标离开事件。
func onMouseLeave(_ *core.App) {
	if ui.scene != nil {
		ui.scene.DispatchMouseLeave()
	}
}

// onMouseDown 分发鼠标按下事件。
func onMouseDown(_ *core.App, ev core.MouseEvent) {
	if ui.scene != nil {
		ui.scene.DispatchMouseDown(ev)
	}
}

// onMouseUp 分发鼠标释放事件。
func onMouseUp(_ *core.App, ev core.MouseEvent) {
	if ui.scene != nil {
		ui.scene.DispatchMouseUp(ev)
	}
}

// onKeyDown 分发按键按下事件。
func onKeyDown(_ *core.App, ev core.KeyEvent) {
	if ui.scene != nil {
		ui.scene.DispatchKeyDown(ev)
	}
}

// onChar 分发字符输入事件。
func onChar(_ *core.App, ch rune) {
	if ui.scene != nil {
		ui.scene.DispatchChar(ch)
	}
}

// onFocus 处理焦点切换。
func onFocus(_ *core.App, focused bool) {
	if ui.scene != nil && !focused {
		ui.scene.Blur()
	}
}

// onTimer 处理定时器事件。
func onTimer(_ *core.App, id uintptr) {
	if ui.scene != nil {
		ui.scene.HandleTimer(id)
	}
}

// onDPIChanged 响应 DPI 变化。
func onDPIChanged(_ *core.App, _ core.DPIInfo) {
	if ui.scene != nil {
		ui.scene.ReloadResources()
	}
	size := ui.app.ClientSize()
	layout(size.Width, size.Height)
}

// onDestroy 清理示例资源。
func onDestroy(_ *core.App) {
	if ui.stepIcon != nil {
		_ = ui.stepIcon.Close()
		ui.stepIcon = nil
	}
	if ui.scene != nil {
		_ = ui.scene.Close()
	}
}

// layout 按当前窗口尺寸摆放示例控件。
func layout(w, h int32) {
	if ui.app == nil {
		return
	}

	margin := ui.app.DP(24)
	columnGap := ui.app.DP(28)
	leftW := (w - margin*2 - columnGap) / 2
	rightX := margin + leftW + columnGap
	fieldH := ui.app.DP(42)
	rowGap := ui.app.DP(16)

	ui.title.SetBounds(core.Rect{X: margin, Y: ui.app.DP(20), W: w - margin*2, H: ui.app.DP(36)})
	ui.renderer.SetBounds(core.Rect{X: margin, Y: ui.app.DP(56), W: w - margin*2, H: ui.app.DP(22)})
	ui.status.SetBounds(core.Rect{X: margin, Y: ui.app.DP(84), W: w - margin*2, H: ui.app.DP(24)})

	ui.progressPct.SetBounds(core.Rect{X: margin, Y: ui.app.DP(118), W: leftW, H: ui.app.DP(20)})
	ui.progress.SetBounds(core.Rect{X: margin, Y: ui.app.DP(144), W: leftW, H: ui.app.DP(16)})
	ui.check.SetBounds(core.Rect{X: margin, Y: ui.app.DP(178), W: leftW, H: ui.app.DP(30)})
	ui.radioA.SetBounds(core.Rect{X: margin, Y: ui.app.DP(216), W: leftW, H: ui.app.DP(30)})
	ui.radioB.SetBounds(core.Rect{X: margin, Y: ui.app.DP(254), W: leftW, H: ui.app.DP(30)})
	ui.edit.SetBounds(core.Rect{X: margin, Y: ui.app.DP(302), W: leftW, H: fieldH})

	ui.list.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(122), W: leftW, H: ui.app.DP(176)})
	ui.combo.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(314), W: leftW, H: fieldH})

	buttonY := h - margin - ui.app.DP(48)
	buttonW := ui.app.DP(132)
	ui.btnStep.SetBounds(core.Rect{X: margin, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
	ui.btnReset.SetBounds(core.Rect{X: margin + buttonW + rowGap, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
}

// displayItem 返回示例列表项对应显示的文本。
func displayItem(item widgets.ListItem) string {
	if item.Text != "" {
		return item.Text
	}
	return item.Value
}

// setProgress 同步更新进度条和百分比文本。
func setProgress(value int32) {
	value = demoClamp(value, 0, 100)
	if ui.progress != nil {
		ui.progress.SetValue(value)
	}
	if ui.progressPct != nil {
		ui.progressPct.SetText("Progress " + strconv.Itoa(int(value)) + "%")
	}
}

// demoRenderMode 从环境变量读取示例期望使用的渲染模式。
func demoRenderMode() core.RenderMode {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("WINUI_RENDER_MODE"))) {
	case "gdi":
		return core.RenderModeGDI
	default:
		return core.RenderModeAuto
	}
}

// renderSummary 汇总示例请求的模式、实际后端和回退原因。
func renderSummary(app *core.App) string {
	if app == nil {
		return "Renderer: unavailable"
	}
	summary := "Requested: " + app.RenderMode().String() + " | Active: " + app.RenderBackend().String()
	if reason := app.RenderFallbackReason(); reason != "" {
		summary += " | Fallback: " + reason
	}
	return summary
}

// demoBadge 生成一个简单的演示图标，用于左图标按钮样式。
func demoBadge(fill, accent color.RGBA) *core.Icon {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fillCircle(img, 16, 16, 14, fill)
	fillCircle(img, 11, 11, 5, accent)
	fillCircle(img, 20, 21, 4, accent)

	icon, err := core.LoadIconFromICO(buildICO(img), 32)
	if err != nil {
		return nil
	}
	return icon
}

func buildICO(img *image.RGBA) []byte {
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

func fillCircle(img *image.RGBA, cx, cy, radius int, clr color.RGBA) {
	if img == nil || radius <= 0 {
		return
	}
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if !image.Pt(x, y).In(img.Bounds()) {
				continue
			}
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(x, y, clr)
			}
		}
	}
}

func demoClamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

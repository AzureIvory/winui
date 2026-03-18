//go:build windows

package main

import (
	"github.com/yourname/winui/core"
	"github.com/yourname/winui/widgets"
)

type demoUI struct {
	app   *core.App
	scene *widgets.Scene

	title    *widgets.Label
	status   *widgets.Label
	progress *widgets.ProgressBar
	check    *widgets.CheckBox
	radioA   *widgets.RadioButton
	radioB   *widgets.RadioButton
	list     *widgets.ListBox
	combo    *widgets.ComboBox
	edit     *widgets.EditBox
	btnStep  *widgets.Button
	btnReset *widgets.Button
}

var ui demoUI

// main 是程序入口。
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

	ui.status = widgets.NewLabel("status", "Ready")
	ui.status.SetStyle(widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		Color:  core.RGB(71, 85, 105),
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	ui.progress = widgets.NewProgressBar("progress")
	ui.progress.SetValue(35)

	ui.check = widgets.NewCheckBox("check", "Enable safe mode")
	ui.check.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Safe mode enabled")
			return
		}
		ui.status.SetText("Safe mode disabled")
	})

	ui.radioA = widgets.NewRadioButton("radio-quick", "Quick install")
	ui.radioA.SetGroup("mode")
	ui.radioA.SetChecked(true)
	ui.radioA.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Quick install selected")
		}
	})

	ui.radioB = widgets.NewRadioButton("radio-full", "Full install")
	ui.radioB.SetGroup("mode")
	ui.radioB.SetOnChange(func(checked bool) {
		if checked {
			ui.status.SetText("Full install selected")
		}
	})

	ui.list = widgets.NewListBox("list")
	ui.list.SetItems([]widgets.ListItem{
		{Value: "windows-10", Text: "Windows 10"},
		{Value: "windows-11", Text: "Windows 11"},
		{Value: "windows-server", Text: "Windows Server"},
	})
	ui.list.SetSelected(1)
	ui.list.SetOnChange(func(_ int, item widgets.ListItem) {
		ui.status.SetText("List item: " + displayItem(item))
	})

	ui.combo = widgets.NewComboBox("combo")
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
	ui.edit.SetPlaceholder("Type a custom environment label")
	ui.edit.SetOnChange(func(text string) {
		if text == "" {
			ui.status.SetText("Input cleared")
			return
		}
		ui.status.SetText("Label: " + text)
	})

	ui.btnStep = widgets.NewButton("step", "Advance")
	ui.btnStep.SetOnClick(func() {
		next := ui.progress.Value() + 10
		if next > 100 {
			next = 100
		}
		ui.progress.SetValue(next)
		ui.status.SetText("Progress updated")
	})

	ui.btnReset = widgets.NewButton("reset", "Reset")
	ui.btnReset.SetOnClick(func() {
		ui.progress.SetValue(0)
		ui.status.SetText("Progress reset")
	})

	root := ui.scene.Root()
	root.Add(ui.title)
	root.Add(ui.status)
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
	ui.status.SetBounds(core.Rect{X: margin, Y: ui.app.DP(58), W: w - margin*2, H: ui.app.DP(26)})

	ui.progress.SetBounds(core.Rect{X: margin, Y: ui.app.DP(106), W: leftW, H: ui.app.DP(14)})
	ui.check.SetBounds(core.Rect{X: margin, Y: ui.app.DP(146), W: leftW, H: ui.app.DP(30)})
	ui.radioA.SetBounds(core.Rect{X: margin, Y: ui.app.DP(184), W: leftW, H: ui.app.DP(30)})
	ui.radioB.SetBounds(core.Rect{X: margin, Y: ui.app.DP(222), W: leftW, H: ui.app.DP(30)})
	ui.edit.SetBounds(core.Rect{X: margin, Y: ui.app.DP(270), W: leftW, H: fieldH})

	ui.list.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(106), W: leftW, H: ui.app.DP(170)})
	ui.combo.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(292), W: leftW, H: fieldH})

	buttonY := h - margin - ui.app.DP(48)
	buttonW := ui.app.DP(132)
	ui.btnStep.SetBounds(core.Rect{X: margin, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
	ui.btnReset.SetBounds(core.Rect{X: margin + buttonW + rowGap, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
}

// displayItem 返回示例列表项应显示的文本。
func displayItem(item widgets.ListItem) string {
	if item.Text != "" {
		return item.Text
	}
	return item.Value
}

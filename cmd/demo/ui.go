//go:build windows

package main

import (
	"image/color"
	"strconv"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// demoUI 聚合演示程序中会反复访问的控件引用。
type demoUI struct {
	// app 保存演示程序的应用实例。
	app *core.App

	// title 保存标题标签。
	title *widgets.Label
	// renderer 保存渲染器摘要标签。
	renderer *widgets.Label
	// status 保存状态标签。
	status *widgets.Label
	// progressPct 保存进度文本标签。
	progressPct *widgets.Label
	// progress 保存进度条控件。
	progress *widgets.ProgressBar
	// check 保存复选框控件。
	check *widgets.CheckBox
	// radioA 保存第一个单选按钮。
	radioA *widgets.RadioButton
	// radioB 保存第二个单选按钮。
	radioB *widgets.RadioButton
	// list 保存列表框控件。
	list *widgets.ListBox
	// combo 保存组合框控件。
	combo *widgets.ComboBox
	// edit 保存编辑框控件。
	edit *widgets.EditBox
	// btnStep 保存推进按钮。
	btnStep *widgets.Button
	// btnReset 保存重置按钮。
	btnReset *widgets.Button

	// stepIcon 保存推进按钮使用的图标资源。
	stepIcon *core.Icon
}

// buildDemoUI 初始化演示界面。
func buildDemoUI(app *core.App, scene *widgets.Scene) error {
	ui = demoUI{app: app}

	ui.initLabels(app)
	ui.initProgress()
	ui.initChoices()
	ui.initSelectors()
	ui.initButtons()

	scene.Root().AddAll(
		ui.title,
		ui.renderer,
		ui.status,
		ui.progressPct,
		ui.progress,
		ui.check,
		ui.radioA,
		ui.radioB,
		ui.list,
		ui.combo,
		ui.edit,
		ui.btnStep,
		ui.btnReset,
	)

	size := app.ClientSize()
	layout(size.Width, size.Height)
	return nil
}

// resizeDemoUI 按当前窗口尺寸重新布局演示控件。
func resizeDemoUI(_ *core.App, _ *widgets.Scene, size core.Size) {
	layout(size.Width, size.Height)
}

// destroyDemoUI 清理演示资源。
func destroyDemoUI(_ *core.App, _ *widgets.Scene) {
	if ui.stepIcon != nil {
		_ = ui.stepIcon.Close()
		ui.stepIcon = nil
	}
}

// initLabels 初始化顶部文本信息。
func (d *demoUI) initLabels(app *core.App) {
	d.title = newLabel("title", "winui widgets demo", widgets.TextStyle{
		Font:   demoFont(22, 700),
		Color:  core.RGB(15, 23, 42),
		Format: core.DTVCenter | core.DTSingleLine,
	})

	d.renderer = newLabel("renderer", renderSummary(app), widgets.TextStyle{
		Font:   demoFont(13, 700),
		Color:  core.RGB(37, 99, 235),
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	d.status = newLabel("status", "Ready", widgets.TextStyle{
		Font:   demoFont(15, 400),
		Color:  core.RGB(71, 85, 105),
		Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis,
	})

	d.progressPct = newLabel("progress-text", "", widgets.TextStyle{
		Font:   demoFont(14, 700),
		Color:  core.RGB(5, 150, 105),
		Format: core.DTVCenter | core.DTSingleLine,
	})
}

// initProgress 初始化进度条区域。
func (d *demoUI) initProgress() {
	d.progress = widgets.NewProgressBar("progress")
	d.progress.SetStyle(widgets.ProgressStyle{
		FillColor:    core.RGB(16, 185, 129),
		BubbleColor:  core.RGB(5, 150, 105),
		TrackColor:   core.RGB(229, 231, 235),
		TextColor:    core.RGB(255, 255, 255),
		CornerRadius: 10,
		ShowPercent:  true,
	})
	setProgress(35)
}

// initChoices 初始化复选与单选控件。
func (d *demoUI) initChoices() {
	d.check = widgets.NewCheckBox("check", "Enable safe mode")
	d.check.SetStyle(widgets.ChoiceStyle{
		HoverBorder:     core.RGB(20, 184, 166),
		FocusBorder:     core.RGB(13, 148, 136),
		IndicatorColor:  core.RGB(13, 148, 136),
		HoverBackground: core.RGB(240, 253, 250),
	})
	d.check.SetOnChange(func(checked bool) {
		if checked {
			d.status.SetText("Safe mode enabled")
			return
		}
		d.status.SetText("Safe mode disabled")
	})

	d.radioA = widgets.NewRadioButton("radio-quick", "Quick install")
	d.radioA.SetGroup("mode")
	d.radioA.SetStyle(widgets.ChoiceStyle{
		HoverBorder:     core.RGB(249, 115, 22),
		FocusBorder:     core.RGB(234, 88, 12),
		IndicatorColor:  core.RGB(234, 88, 12),
		HoverBackground: core.RGB(255, 247, 237),
	})
	d.radioA.SetChecked(true)
	d.radioA.SetOnChange(func(checked bool) {
		if checked {
			d.status.SetText("Quick install selected")
		}
	})

	d.radioB = widgets.NewRadioButton("radio-full", "Full install")
	d.radioB.SetGroup("mode")
	d.radioB.SetStyle(d.radioA.Style)
	d.radioB.SetOnChange(func(checked bool) {
		if checked {
			d.status.SetText("Full install selected")
		}
	})
}

// initSelectors 初始化列表、组合框和输入框。
func (d *demoUI) initSelectors() {
	d.list = widgets.NewListBox("list")
	d.list.SetStyle(widgets.ListStyle{
		HoverBorder:       core.RGB(96, 165, 250),
		FocusBorder:       core.RGB(37, 99, 235),
		ItemHoverColor:    core.RGB(239, 246, 255),
		ItemSelectedColor: core.RGB(37, 99, 235),
	})
	d.list.SetItems([]widgets.ListItem{
		{Value: "windows-10", Text: "Windows 10"},
		{Value: "windows-11", Text: "Windows 11"},
		{Value: "windows-pe", Text: "Windows PE"},
	})
	d.list.SetSelected(1)
	d.list.SetOnChange(func(_ int, item widgets.ListItem) {
		d.status.SetText("List item: " + displayItem(item))
	})

	d.combo = widgets.NewComboBox("combo")
	d.combo.SetStyle(widgets.ComboStyle{
		HoverBorder:       core.RGB(251, 191, 36),
		FocusBorder:       core.RGB(249, 115, 22),
		ArrowColor:        core.RGB(249, 115, 22),
		ItemHoverColor:    core.RGB(255, 247, 237),
		ItemSelectedColor: core.RGB(249, 115, 22),
	})
	d.combo.SetPlaceholder("Select accent color")
	d.combo.SetItems([]widgets.ListItem{
		{Value: "blue", Text: "Blue"},
		{Value: "green", Text: "Green"},
		{Value: "orange", Text: "Orange"},
	})
	d.combo.SetSelected(0)
	d.combo.SetOnChange(func(_ int, item widgets.ListItem) {
		d.status.SetText("Accent color: " + displayItem(item))
	})

	d.edit = widgets.NewEditBox("edit")
	d.edit.SetStyle(widgets.EditStyle{
		HoverBorder: core.RGB(125, 211, 252),
		FocusBorder: core.RGB(14, 165, 233),
		CaretColor:  core.RGB(14, 165, 233),
	})
	d.edit.SetPlaceholder("Type a custom environment label")
	d.edit.SetOnChange(func(text string) {
		if text == "" {
			d.status.SetText("Input cleared")
			return
		}
		d.status.SetText("Label: " + text)
	})
}

// initButtons 初始化演示按钮。
func (d *demoUI) initButtons() {
	d.stepIcon = demoBadge(
		color.RGBA{R: 37, G: 99, B: 235, A: 255},
		color.RGBA{R: 191, G: 219, B: 254, A: 255},
	)

	d.btnStep = widgets.NewButton("step", "Advance")
	d.btnStep.SetKind(widgets.BtnLeft)
	d.btnStep.SetIcon(d.stepIcon)
	d.btnStep.SetStyle(widgets.ButtonStyle{
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
	d.btnStep.SetOnClick(func() {
		next := d.progress.Value() + 10
		if next > 100 {
			next = 100
		}
		setProgress(next)
		d.status.SetText("Progress updated")
	})

	d.btnReset = widgets.NewButton("reset", "Reset")
	d.btnReset.SetStyle(widgets.ButtonStyle{
		TextColor:    core.RGB(194, 65, 12),
		DownText:     core.RGB(255, 255, 255),
		Background:   core.RGB(255, 247, 237),
		Hover:        core.RGB(254, 215, 170),
		Pressed:      core.RGB(249, 115, 22),
		Border:       core.RGB(251, 146, 60),
		CornerRadius: 10,
		PadDP:        14,
	})
	d.btnReset.SetOnClick(func() {
		setProgress(0)
		d.status.SetText("Progress reset")
	})
}

// newLabel 创建并初始化一个演示标签。
func newLabel(id, text string, style widgets.TextStyle) *widgets.Label {
	label := widgets.NewLabel(id, text)
	label.SetStyle(style)
	return label
}

// demoFont 返回演示中统一使用的字体规格。
func demoFont(sizeDP, weight int32) widgets.FontSpec {
	return widgets.FontSpec{
		Face:   "Microsoft YaHei UI",
		SizeDP: sizeDP,
		Weight: weight,
	}
}

// renderSummary 汇总演示请求的模式、实际后端和回退原因。
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

// displayItem 返回列表项对应的展示文本。
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

// demoClamp 将演示中的整数值限制在闭区间内。
func demoClamp(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

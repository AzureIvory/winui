//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/sysapi"
	"github.com/AzureIvory/winui/widgets"
)

type goDemo struct {
	baseDir   string
	assetsDir string

	app   *core.App
	scene *widgets.Scene

	lang   *demoLang
	locale string
	mode   widgets.ControlMode

	root       *widgets.Panel
	headerCard *widgets.Panel
	leftCard   *widgets.Panel
	rightCard  *widgets.Panel

	titleLabel    *widgets.Label
	subtitleLabel *widgets.Label
	statusLabel   *widgets.Label
	emojiLabel    *widgets.Label

	modeButton *widgets.Button
	langButton *widgets.Button
	runButton  *widgets.Button

	nameInput     *widgets.EditBox
	passwordInput *widgets.EditBox
	notesBox      *widgets.EditBox
	dotStyleBox   *widgets.CheckBox
	radioDotA     *widgets.RadioButton
	radioDotB     *widgets.RadioButton
	citySelect    *widgets.ComboBox
	cityList      *widgets.ListBox
	openFile      *widgets.FilePicker
	uploadBar     *widgets.ProgressBar
	previewImage  *widgets.Image
	spinnerImage  *widgets.AnimatedImage

	appIcon    *core.Image
	runIcon    *core.Image
	previewPNG *core.Image
	spinnerGIF []core.AnimatedFrame
}

func main() {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)

	demo := newGoDemo(baseDir)

	opts := core.Options{
		ClassName:      "WinUIGoFullDemo",
		Title:          "WinUI Go Full Demo",
		Width:          1380,
		Height:         940,
		MinWidth:       1240,
		MinHeight:      820,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(244, 247, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	if demo.appIcon != nil {
		opts.WindowImage = demo.appIcon
		opts.WindowImageSizeDP = 32
	}

	widgets.BindScene(&opts, widgets.SceneHooks{
		Theme: demoTheme(),
		OnCreate: func(app *core.App, scene *widgets.Scene) error {
			return demo.onCreate(app, scene)
		},
		OnResize: func(_ *core.App, _ *widgets.Scene, size core.Size) {
			demo.layout(size)
		},
		OnDestroy: func(_ *core.App, _ *widgets.Scene) {
			demo.onDestroy()
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

func newGoDemo(baseDir string) *goDemo {
	assetsDir := filepath.Join(baseDir, "..", "demo_json_full", "assets")
	demo := &goDemo{
		baseDir:   baseDir,
		assetsDir: assetsDir,
		locale:    "en",
		mode:      widgets.ModeCustom,
	}
	if lang, err := loadDemoLang(filepath.Join(assetsDir, "lang.json")); err == nil {
		demo.lang = lang
	}
	demo.loadSharedResources()
	return demo
}

func (d *goDemo) onCreate(app *core.App, scene *widgets.Scene) error {
	d.app = app
	d.scene = scene
	d.rebuildUI()
	d.setStatus(d.tr("status.uiReady", "UI ready"))
	return nil
}

func (d *goDemo) onDestroy() {
	d.closeSharedResources()
}

func (d *goDemo) loadSharedResources() {
	if icon, err := core.LoadImageFile(filepath.Join(d.assetsDir, "app.png")); err == nil {
		d.appIcon = icon
	}
	if icon, err := core.LoadImageFile(filepath.Join(d.assetsDir, "save.png")); err == nil {
		d.runIcon = icon
	}
	if image, err := core.LoadImageFile(filepath.Join(d.assetsDir, "preview.png")); err == nil {
		d.previewPNG = image
	}
	if data, err := os.ReadFile(filepath.Join(d.assetsDir, "spinner.gif")); err == nil {
		if frames, decodeErr := core.DecodeGIF(data); decodeErr == nil {
			d.spinnerGIF = frames
		}
	}
}

func (d *goDemo) closeSharedResources() {
	if d.appIcon != nil {
		_ = d.appIcon.Close()
		d.appIcon = nil
	}
	if d.runIcon != nil {
		_ = d.runIcon.Close()
		d.runIcon = nil
	}
	if d.previewPNG != nil {
		_ = d.previewPNG.Close()
		d.previewPNG = nil
	}
	for i := range d.spinnerGIF {
		if d.spinnerGIF[i].Bitmap != nil {
			_ = d.spinnerGIF[i].Bitmap.Close()
		}
	}
	d.spinnerGIF = nil
}

func (d *goDemo) rebuildUI() {
	if d.scene == nil {
		return
	}
	if d.root != nil {
		d.scene.Root().Remove(d.root.ID())
	}

	d.root = d.buildRoot()
	d.scene.Root().Add(d.root)
	d.applyLocalizedTexts()
	d.applySharedVisuals()
	d.layout(d.app.ClientSize())
}

func (d *goDemo) buildRoot() *widgets.Panel {
	root := widgets.NewPanel("goPageRoot")
	root.SetStyle(widgets.PanelStyle{Background: core.RGB(244, 247, 251)})
	root.SetLayout(widgets.AbsoluteLayout{})

	d.headerCard = widgets.NewPanel("goHeaderCard")
	d.headerCard.SetStyle(widgets.PanelStyle{
		Background:   core.RGB(255, 255, 255),
		BorderColor:  core.RGB(214, 222, 234),
		BorderWidth:  1,
		CornerRadius: 18,
	})
	d.headerCard.SetLayout(widgets.AbsoluteLayout{})

	d.titleLabel = widgets.NewLabel("goTitle", "")
	d.titleLabel.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 22, Weight: 700}, Color: core.RGB(15, 23, 42), Format: core.DTVCenter | core.DTSingleLine})
	d.subtitleLabel = widgets.NewLabel("goSubtitle", "")
	d.subtitleLabel.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 13}, Color: core.RGB(100, 116, 139), Format: core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis})

	d.modeButton = widgets.NewButton("toggleControlModeBtn", "", d.mode)
	d.modeButton.SetOnClick(func() {
		if d.mode == widgets.ModeNative {
			d.mode = widgets.ModeCustom
		} else {
			d.mode = widgets.ModeNative
		}
		d.rebuildUI()
		d.setStatus(d.tr("status.controlModeSwitched", "Control mode switched to %s", goModeLabel(d.lang, d.locale, d.mode)))
	})

	d.langButton = widgets.NewButton("languageToggleBtn", "", d.mode)
	d.langButton.SetOnClick(func() {
		if normalizeDemoLocale(d.locale) == "en" {
			d.locale = "zh"
		} else {
			d.locale = "en"
		}
		d.applyLocalizedTexts()
		d.setStatus(d.tr("status.uiReady", "UI ready"))
	})

	d.runButton = widgets.NewButton("runAllBtn", "", d.mode)
	d.runButton.SetKind(widgets.BtnLeft)
	d.runButton.SetOnClick(func() {
		next := d.uploadBar.Value() + 17
		if next > 100 {
			next = 12
		}
		d.uploadBar.SetValue(next)
		summaryFormat := "PASS=%d FAIL=%d"
		if d.lang != nil {
			summaryFormat = d.lang.text(d.locale, "report.summaryFormat", summaryFormat)
		}
		summary := fmt.Sprintf(summaryFormat, int(next/20)+1, int((100-next)/33))
		d.setStatus(d.tr("status.apiCheckComplete", "API check complete: %s", summary))
	})

	d.headerCard.AddAll(d.titleLabel, d.subtitleLabel, d.modeButton, d.langButton, d.runButton)

	d.leftCard = widgets.NewPanel("goLeftCard")
	d.leftCard.SetStyle(widgets.PanelStyle{
		Background:   core.RGB(255, 255, 255),
		BorderColor:  core.RGB(214, 222, 234),
		BorderWidth:  1,
		CornerRadius: 16,
	})
	d.leftCard.SetLayout(widgets.ColumnLayout{Gap: 10, Padding: widgets.UniformInsets(16), CrossAlign: widgets.AlignStretch})

	leftTitle := widgets.NewLabel("leftTitle", "")
	leftTitle.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 18, Weight: 700}, Color: core.RGB(15, 23, 42), Format: core.DTVCenter | core.DTSingleLine})

	d.nameInput = widgets.NewEditBox("nameInput", d.mode)
	d.nameInput.SetOnChange(func(value string) {
		d.setStatus(d.tr("status.nameChanged", "Name input changed: %s", value))
	})
	widgets.SetPreferredSize(d.nameInput, core.Size{Height: 40})

	d.passwordInput = widgets.NewEditBox("passwordInput", d.mode)
	d.passwordInput.SetPassword(true)
	d.passwordInput.SetOnSubmit(func(value string) {
		d.setStatus(d.tr("status.passwordSubmitted", "Password submitted, rune length: %d", len([]rune(value))))
	})
	widgets.SetPreferredSize(d.passwordInput, core.Size{Height: 40})

	d.notesBox = widgets.NewEditBox("notesBox", d.mode)
	d.notesBox.SetMultiline(true)
	d.notesBox.SetAcceptReturn(true)
	d.notesBox.SetReadOnly(true)
	widgets.SetPreferredSize(d.notesBox, core.Size{Height: 88})

	d.dotStyleBox = widgets.NewCheckBox("dotStyleBox", "", d.mode)
	d.dotStyleBox.SetOnChange(func(checked bool) {
		d.setStatus(d.tr("status.choiceChanged", "%s changed to %s", d.dotStyleBox.Text, d.checkedLabel(checked)))
	})

	d.radioDotA = widgets.NewRadioButton("radioDotA", "", d.mode)
	d.radioDotA.SetGroup("go-demo-mode")
	d.radioDotA.SetChecked(true)
	d.radioDotA.SetOnChange(func(checked bool) {
		if checked {
			d.setStatus(d.tr("status.radioSelected", "%s selected", d.radioDotA.Text))
		}
	})

	d.radioDotB = widgets.NewRadioButton("radioDotB", "", d.mode)
	d.radioDotB.SetGroup("go-demo-mode")
	d.radioDotB.SetOnChange(func(checked bool) {
		if checked {
			d.setStatus(d.tr("status.radioSelected", "%s selected", d.radioDotB.Text))
		}
	})

	d.openFile = widgets.NewFilePicker("openFile", d.mode)
	d.openFile.SetOnChange(func(paths []string) {
		if len(paths) == 0 {
			d.setStatus(d.tr("status.pickerCleared", "%s cleared", "openFile"))
			return
		}
		d.setStatus(d.tr("status.pickerSelected", "%s selected %d path(s)", "openFile", len(paths)))
	})
	widgets.SetPreferredSize(d.openFile, core.Size{Height: 40})

	d.uploadBar = widgets.NewProgressBar("uploadProgress")
	d.uploadBar.SetValue(58)
	d.uploadBar.SetStyle(widgets.ProgressStyle{
		TrackColor:   core.RGB(219, 234, 254),
		FillColor:    core.RGB(37, 99, 235),
		BubbleColor:  core.RGB(29, 78, 216),
		TextColor:    core.RGB(255, 255, 255),
		CornerRadius: 10,
		ShowPercent:  true,
	})
	widgets.SetPreferredSize(d.uploadBar, core.Size{Height: 22})

	d.leftCard.AddAll(leftTitle, d.nameInput, d.passwordInput, d.notesBox, d.dotStyleBox, d.radioDotA, d.radioDotB, d.openFile, d.uploadBar)

	d.rightCard = widgets.NewPanel("goRightCard")
	d.rightCard.SetStyle(widgets.PanelStyle{
		Background:   core.RGB(255, 255, 255),
		BorderColor:  core.RGB(214, 222, 234),
		BorderWidth:  1,
		CornerRadius: 16,
	})
	d.rightCard.SetLayout(widgets.ColumnLayout{Gap: 10, Padding: widgets.UniformInsets(16), CrossAlign: widgets.AlignStretch})

	rightTitle := widgets.NewLabel("rightTitle", "")
	rightTitle.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 18, Weight: 700}, Color: core.RGB(15, 23, 42), Format: core.DTVCenter | core.DTSingleLine})

	d.citySelect = widgets.NewComboBox("citySelect", d.mode)
	d.citySelect.SetOnChange(func(index int, item widgets.ListItem) {
		d.setStatus(d.tr("status.comboSelected", "%s selected: %d / %s", d.tr("i18n.data.comboLabel", "City combo box"), index, item.Value))
	})
	widgets.SetPreferredSize(d.citySelect, core.Size{Height: 40})

	d.cityList = widgets.NewListBox("cityList")
	d.cityList.SetOnChange(func(index int, item widgets.ListItem) {
		d.setStatus(d.tr("status.listChanged", "%s changed: %d / %s", d.tr("i18n.data.listLabel", "Rollout list"), index, item.Value))
	})
	widgets.SetPreferredSize(d.cityList, core.Size{Height: 140})

	mediaTitle := widgets.NewLabel("mediaTitle", "")
	mediaTitle.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 16, Weight: 700}, Color: core.RGB(15, 23, 42), Format: core.DTVCenter | core.DTSingleLine})

	mediaRow := widgets.NewPanel("mediaRow")
	mediaRow.SetLayout(widgets.RowLayout{Gap: 12, CrossAlign: widgets.AlignStretch})

	d.previewImage = widgets.NewImage("previewImage")
	d.previewImage.SetScaleMode(widgets.ImageScaleContain)
	widgets.SetPreferredSize(d.previewImage, core.Size{Width: 210, Height: 148})

	d.spinnerImage = widgets.NewAnimatedImage("spinnerImage")
	d.spinnerImage.SetScaleMode(widgets.ImageScaleContain)
	widgets.SetPreferredSize(d.spinnerImage, core.Size{Width: 210, Height: 148})
	mediaRow.AddAll(d.previewImage, d.spinnerImage)
	widgets.SetPreferredSize(mediaRow, core.Size{Height: 154})

	d.emojiLabel = widgets.NewLabel("emojiBurstLabel", "")
	d.emojiLabel.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 16, Weight: 700}, Color: core.RGB(37, 99, 235), Format: core.DTVCenter | core.DTSingleLine})

	d.statusLabel = widgets.NewLabel("statusValue", "")
	d.statusLabel.SetMultiline(true)
	d.statusLabel.SetWordWrap(true)
	d.statusLabel.SetStyle(widgets.TextStyle{Font: widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 14}, Color: core.RGB(51, 65, 85), Format: core.DTWordBreak})
	widgets.SetPreferredSize(d.statusLabel, core.Size{Height: 58})

	d.rightCard.AddAll(rightTitle, d.citySelect, d.cityList, mediaTitle, mediaRow, d.emojiLabel, d.statusLabel)

	root.AddAll(d.headerCard, d.leftCard, d.rightCard)
	return root
}

func (d *goDemo) applySharedVisuals() {
	if d.runButton != nil && d.runIcon != nil {
		d.runButton.SetImage(d.runIcon)
	}
	if d.previewImage != nil && d.previewPNG != nil {
		d.previewImage.SetBitmap(d.previewPNG.MasterBitmap())
	}
	if d.spinnerImage != nil && len(d.spinnerGIF) > 0 {
		d.spinnerImage.SetFrames(d.spinnerGIF)
		d.spinnerImage.SetPlaying(true)
	}
}

func (d *goDemo) applyLocalizedTexts() {
	if d.app != nil {
		d.app.SetTitle(d.tr("windowTitle", "WinUI Go Full Demo"))
	}

	if d.titleLabel != nil {
		d.titleLabel.SetText(d.tr("i18n.header.titleLabel", "WinUI Go Full Demo"))
	}
	if d.subtitleLabel != nil {
		d.subtitleLabel.SetText(d.tr("i18n.header.subtitleLabel", "One native Go page that showcases every public control, mode switching, and emoji-rich content."))
	}
	if d.modeButton != nil {
		d.modeButton.SetText(goModeButtonText(d.lang, d.locale, d.mode))
	}
	if d.langButton != nil {
		d.langButton.SetText(d.tr("i18n.header.languageToggleBtn", "中文"))
	}
	if d.runButton != nil {
		d.runButton.SetText("🧪 " + d.tr("i18n.header.runAllBtn", "Run widget API tests"))
	}

	if d.nameInput != nil {
		d.nameInput.SetPlaceholder(d.tr("i18n.inputs.namePlaceholder", "Project name"))
	}
	if d.passwordInput != nil {
		d.passwordInput.SetPlaceholder(d.tr("i18n.inputs.passwordPlaceholder", "Type a secret"))
	}
	if d.notesBox != nil {
		d.notesBox.SetPlaceholder(d.tr("i18n.inputs.notesPlaceholder", "Notes"))
		d.notesBox.SetText(d.tr("i18n.inputs.notesValue", "Go builds this UI entirely at runtime, while language data is still reused from lang.json."))
	}
	if d.dotStyleBox != nil {
		d.dotStyleBox.SetText(d.tr("i18n.choices.dotStyleBox", "Dot-style checkbox"))
	}
	if d.radioDotA != nil {
		d.radioDotA.SetText(d.tr("i18n.choices.radioDotA", "Notify immediately"))
	}
	if d.radioDotB != nil {
		d.radioDotB.SetText(d.tr("i18n.choices.radioDotB", "Notify later"))
	}

	comboItems := []widgets.ListItem{
		{Value: "bj", Text: "Beijing"},
		{Value: "sh", Text: "Shanghai"},
		{Value: "hz", Text: "Hangzhou"},
		{Value: "cd", Text: "Chengdu"},
	}
	if d.lang != nil {
		comboItems = d.lang.listItems(d.locale, "i18n.data.comboItems", comboItems)
	}
	if d.citySelect != nil {
		d.citySelect.SetItems(comboItems)
		d.citySelect.SetPlaceholder(d.tr("i18n.data.comboPlaceholder", "Select a city"))
		d.citySelect.SetSelected(0)
	}

	listItems := []widgets.ListItem{
		{Value: "alpha", Text: "Alpha rollout"},
		{Value: "beta", Text: "Beta rollout"},
		{Value: "release", Text: "Release rollout"},
		{Value: "hold", Text: "Hold rollout", Disabled: true},
	}
	if d.lang != nil {
		listItems = d.lang.listItems(d.locale, "i18n.data.listItems", listItems)
	}
	if d.cityList != nil {
		d.cityList.SetItems(listItems)
		d.cityList.SetSelected(0)
	}

	if d.openFile != nil {
		d.openFile.SetButtonText(d.tr("i18n.files.openFile.buttonText", "Open file"))
		d.openFile.SetPlaceholder(d.tr("i18n.files.openFile.placeholder", "No file selected"))
		opts := d.openFile.DialogOptions()
		opts.Mode = sysapi.DialogOpen
		opts.Title = d.tr("i18n.files.openFile.dialogTitle", "Select a source file")
		opts.Filters = []sysapi.FileFilter{{Name: "Source Files", Pattern: "*.md;*.txt;*.go"}}
		d.openFile.SetDialogOptions(opts)
	}

	if d.emojiLabel != nil {
		d.emojiLabel.SetText(d.tr("i18n.media.emojiBurstLabel", "😀 🚀 🧠 🧩 ✨ 🎯 🔥 🛠️"))
	}

	if d.statusLabel != nil && strings.TrimSpace(d.statusLabel.Text) == "" {
		d.statusLabel.SetText(d.tr("status.ready", "Ready"))
	}
}

func (d *goDemo) checkedLabel(checked bool) string {
	if checked {
		return d.tr("common.checked", "checked")
	}
	return d.tr("common.unchecked", "unchecked")
}

func (d *goDemo) tr(path string, fallback string, args ...any) string {
	if d.lang == nil {
		if len(args) == 0 {
			return fallback
		}
		return fmt.Sprintf(fallback, args...)
	}
	return d.lang.text(d.locale, path, fallback, args...)
}

func (d *goDemo) setStatus(text string) {
	if d.statusLabel != nil {
		d.statusLabel.SetText(text)
	}
}

func (d *goDemo) layout(size core.Size) {
	if d.root == nil || d.app == nil {
		return
	}
	margin := d.app.DP(20)
	gap := d.app.DP(16)
	headerH := d.app.DP(86)

	d.root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
	d.headerCard.SetBounds(widgets.Rect{X: margin, Y: margin, W: size.Width - margin*2, H: headerH})

	contentTop := margin + headerH + gap
	contentH := size.Height - contentTop - margin
	if contentH < d.app.DP(320) {
		contentH = d.app.DP(320)
	}
	leftW := (size.Width - margin*2 - gap) / 2
	if leftW < d.app.DP(420) {
		leftW = d.app.DP(420)
	}
	rightW := size.Width - margin*2 - gap - leftW
	if rightW < d.app.DP(420) {
		rightW = d.app.DP(420)
		leftW = size.Width - margin*2 - gap - rightW
	}
	if leftW < 0 {
		leftW = 0
	}
	if rightW < 0 {
		rightW = 0
	}

	d.leftCard.SetBounds(widgets.Rect{X: margin, Y: contentTop, W: leftW, H: contentH})
	d.rightCard.SetBounds(widgets.Rect{X: margin + leftW + gap, Y: contentTop, W: rightW, H: contentH})

	headerInnerX := d.headerCard.Bounds().X + d.app.DP(20)
	headerInnerY := d.headerCard.Bounds().Y + d.app.DP(16)
	headerRight := d.headerCard.Bounds().X + d.headerCard.Bounds().W - d.app.DP(20)
	buttonH := d.app.DP(40)
	runW := d.app.DP(206)
	langW := d.app.DP(96)
	modeW := d.app.DP(170)

	runX := headerRight - runW
	langX := runX - d.app.DP(12) - langW
	modeX := langX - d.app.DP(12) - modeW

	d.titleLabel.SetBounds(widgets.Rect{X: headerInnerX, Y: headerInnerY - d.app.DP(4), W: d.app.DP(440), H: d.app.DP(30)})
	subtitleW := modeX - d.app.DP(16) - headerInnerX
	if subtitleW < d.app.DP(280) {
		subtitleW = d.app.DP(280)
	}
	d.subtitleLabel.SetBounds(widgets.Rect{X: headerInnerX, Y: headerInnerY + d.app.DP(26), W: subtitleW, H: d.app.DP(20)})

	d.modeButton.SetBounds(widgets.Rect{X: modeX, Y: headerInnerY, W: modeW, H: buttonH})
	d.langButton.SetBounds(widgets.Rect{X: langX, Y: headerInnerY, W: langW, H: buttonH})
	d.runButton.SetBounds(widgets.Rect{X: runX, Y: headerInnerY, W: runW, H: buttonH})
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

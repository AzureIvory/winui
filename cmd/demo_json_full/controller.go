//go:build windows

package main

import (
	"fmt"
	"path/filepath"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

type demoController struct {
	baseDir   string
	assetsDir string
	store     *jsonui.Store
	doc       *jsonui.Document
	window    *jsonui.Window
	useAlt    bool
}

type demoPalette struct {
	name          string
	page          widgets.PanelStyle
	card          widgets.PanelStyle
	cardMuted     widgets.PanelStyle
	scroll        widgets.PanelStyle
	buttonPrimary widgets.ButtonStyle
	buttonNeutral widgets.ButtonStyle
	buttonGhost   widgets.ButtonStyle
	edit          widgets.EditStyle
	reportEdit    widgets.EditStyle
	choiceDot     widgets.ChoiceStyle
	choiceCheck   widgets.ChoiceStyle
	radioDot      widgets.ChoiceStyle
	radioCheck    widgets.ChoiceStyle
	combo         widgets.ComboStyle
	list          widgets.ListStyle
	progressMain  widgets.ProgressStyle
	progressAlt   widgets.ProgressStyle
	title         widgets.TextStyle
	subtitle      widgets.TextStyle
	sectionTitle  widgets.TextStyle
	body          widgets.TextStyle
	accentText    widgets.TextStyle
	modalSurface  widgets.PanelStyle
	modalBackdrop core.Color
	modalOpacity  byte
	modalBlurDP   int32
}

func newDemoController(baseDir string, store *jsonui.Store, doc *jsonui.Document, window *jsonui.Window) *demoController {
	controller := &demoController{
		baseDir:   baseDir,
		assetsDir: filepath.Join(baseDir, "assets"),
		store:     store,
		doc:       doc,
		window:    window,
	}
	controller.bindCallbacks()
	controller.applyPalette(oceanPalette())
	controller.setStatus("Ready")
	controller.setReportSummary(defaultReportSummary())
	controller.setReportPath(defaultReportPath())
	return controller
}

func (c *demoController) setReportSummary(summary string) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Set("demo.reportSummary", summary)
}

func (c *demoController) setReportPath(path string) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Set("demo.reportPath", path)
}

func (c *demoController) setStatus(text string) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Set("demo.lastAction", text)
}

func (c *demoController) togglePalette() {
	if c == nil {
		return
	}
	c.useAlt = !c.useAlt
	palette := oceanPalette()
	if c.useAlt {
		palette = graphitePalette()
	}
	c.applyPalette(palette)
	c.store.Set("demo.paletteName", palette.name)
	c.setStatus(fmt.Sprintf("Palette switched to %s", palette.name))
}
func (c *demoController) showModal(visible bool) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Set("demo.modalVisible", visible)
	if visible {
		c.setStatus("Modal opened")
		return
	}
	c.setStatus("Modal closed")
}
func (c *demoController) bindCallbacks() {
	c.mustButton("togglePaletteBtn").SetOnClick(func() {
		c.togglePalette()
	})
	c.mustButton("runAllBtn").SetOnClick(func() {
		c.runFunctionTests()
	})
	c.mustButton("openModalBtn").SetOnClick(func() {
		c.showModal(true)
	})
	c.mustButton("closeModalBtn").SetOnClick(func() {
		c.showModal(false)
	})

	c.mustPanel("interactivePanel").SetOnClick(func() {
		c.setStatus("Interactive panel clicked")
	})
	c.mustModal("helpModal").SetOnDismiss(func() {
		c.showModal(false)
	})

	c.mustEdit("nameInput").SetOnChange(func(value string) {
		c.setStatus("Name input changed: " + value)
	})
	c.mustEdit("passwordInput").SetOnSubmit(func(value string) {
		c.setStatus("Password submitted, rune length: " + fmt.Sprint(len([]rune(value))))
	})

	c.mustCheckBox("dotStyleBox").SetOnChange(func(checked bool) {
		c.setStatus(fmt.Sprintf("dotStyleBox changed to %v", checked))
	})
	c.mustCheckBox("checkStyleBox").SetOnChange(func(checked bool) {
		c.setStatus(fmt.Sprintf("checkStyleBox changed to %v", checked))
	})
	c.mustRadio("radioDotA").SetOnChange(func(checked bool) {
		if checked {
			c.setStatus("radioDotA selected")
		}
	})
	c.mustRadio("radioCheckA").SetOnChange(func(checked bool) {
		if checked {
			c.setStatus("radioCheckA selected")
		}
	})

	c.mustCombo("citySelect").SetOnChange(func(index int, item widgets.ListItem) {
		c.setStatus(fmt.Sprintf("ComboBox selected: %d / %s", index, item.Value))
	})
	c.mustListBox("cityList").SetOnChange(func(index int, item widgets.ListItem) {
		c.setStatus(fmt.Sprintf("ListBox changed: %d / %s", index, item.Value))
	})
	c.mustListBox("cityList").SetOnActivate(func(index int, item widgets.ListItem) {
		c.setStatus(fmt.Sprintf("ListBox activated: %d / %s", index, item.Value))
	})
	c.mustListBox("cityList").SetOnRightClick(func(index int, item widgets.ListItem, _ core.Point) {
		c.setStatus(fmt.Sprintf("ListBox right-click: %d / %s", index, item.Value))
	})

	for _, id := range []string{"openFile", "saveFile", "folderPick", "multiFiles"} {
		pickerID := id
		picker := c.mustFilePicker(pickerID)
		picker.SetOnChange(func(paths []string) {
			if len(paths) == 0 {
				c.setStatus(pickerID + " cleared")
				return
			}
			c.setStatus(fmt.Sprintf("%s selected %d path(s)", pickerID, len(paths)))
		})
	}
}
func (c *demoController) applyPalette(p demoPalette) {
	c.mustPanel("pageRoot").SetStyle(p.page)
	for _, id := range []string{
		"headerCard",
		"introCard",
		"buttonsCard",
		"inputCard",
		"choiceCard",
		"dataCard",
		"fileCard",
		"mediaCard",
		"scrollCard",
		"panelCard",
		"resultCard",
	} {
		c.mustPanel(id).SetStyle(p.card)
	}
	c.mustPanel("interactivePanel").SetStyle(p.cardMuted)
	c.mustScroll("showcaseScroll").SetStyle(p.scroll)
	c.mustScroll("miniScroll").SetStyle(p.cardMuted)
	c.mustModal("helpModal").SetBackdropColor(p.modalBackdrop)
	c.mustModal("helpModal").SetBackdropOpacity(p.modalOpacity)
	c.mustModal("helpModal").SetBlurRadiusDP(p.modalBlurDP)
	c.mustPanel("modalSurface").SetStyle(p.modalSurface)

	c.mustButton("togglePaletteBtn").SetStyle(p.buttonNeutral)
	c.mustButton("runAllBtn").SetStyle(p.buttonPrimary)
	c.mustButton("openModalBtn").SetStyle(p.buttonGhost)
	c.mustButton("closeModalBtn").SetStyle(p.buttonNeutral)
	c.mustButton("saveBtn").SetStyle(p.buttonPrimary)
	c.mustButton("iconTopBtn").SetStyle(p.buttonNeutral)
	c.mustButton("ghostBtn").SetStyle(p.buttonGhost)
	c.mustButton("disabledBtn").SetStyle(p.buttonNeutral)
	for _, id := range []string{
		"miniScrollBtn1",
		"miniScrollBtn2",
		"miniScrollBtn3",
		"miniScrollBtn4",
		"miniScrollBtn5",
		"miniScrollBtn6",
	} {
		c.mustButton(id).SetStyle(p.buttonNeutral)
	}

	for _, id := range []string{"nameInput", "passwordInput", "notesBox"} {
		c.mustEdit(id).SetStyle(p.edit)
	}

	for _, id := range []string{"openFile", "saveFile", "folderPick", "multiFiles"} {
		picker := c.mustFilePicker(id)
		picker.SetFieldStyle(p.edit)
		picker.SetButtonStyle(p.buttonNeutral)
	}

	c.mustCheckBox("dotStyleBox").SetStyle(p.choiceDot)
	c.mustCheckBox("checkStyleBox").SetStyle(p.choiceCheck)
	c.mustCheckBox("disabledRule").SetStyle(p.choiceCheck)
	c.mustRadio("radioDotA").SetStyle(p.radioDot)
	c.mustRadio("radioDotB").SetStyle(p.radioDot)
	c.mustRadio("radioCheckA").SetStyle(p.radioCheck)
	c.mustRadio("radioCheckB").SetStyle(p.radioCheck)

	c.mustCombo("citySelect").SetStyle(p.combo)
	c.mustListBox("cityList").SetStyle(p.list)
	c.mustProgress("uploadProgress").SetStyle(p.progressMain)
	c.mustProgress("syncProgress").SetStyle(p.progressAlt)

	c.mustLabel("titleLabel").SetStyle(p.title)
	c.mustLabel("subtitleLabel").SetStyle(p.subtitle)
	c.mustLabel("paletteValue").SetStyle(p.accentText)
	c.mustLabel("resultsTitle").SetStyle(p.sectionTitle)
	c.mustLabel("statusCaption").SetStyle(p.subtitle)
	c.mustLabel("statusValue").SetStyle(p.body)
	c.mustLabel("resultsHint").SetStyle(p.subtitle)
	c.mustLabel("reportSummaryCaption").SetStyle(p.subtitle)
	c.mustLabel("reportSummaryValue").SetStyle(p.body)
	c.mustLabel("reportPathCaption").SetStyle(p.subtitle)
	c.mustLabel("reportPathValue").SetStyle(p.body)
	c.mustLabel("modalTitle").SetStyle(p.title)
	c.mustLabel("modalBody").SetStyle(p.body)
	c.mustLabel("modalStatus").SetStyle(p.subtitle)
	for _, id := range []string{
		"introTitle",
		"buttonsTitle",
		"inputTitle",
		"choiceTitle",
		"dataTitle",
		"fileTitle",
		"mediaTitle",
		"scrollTitle",
		"panelTitle",
	} {
		c.mustLabel(id).SetStyle(p.sectionTitle)
	}
	for _, id := range []string{
		"introBody",
		"buttonsHint",
		"inputHint",
		"choiceHint",
		"dataHint",
		"fileHint",
		"mediaHint",
		"scrollHint",
		"panelHint",
		"panelInfo",
		"auxLabel",
	} {
		if widget := c.doc.FindWidget("main", id); widget != nil {
			c.mustLabel(id).SetStyle(p.body)
			continue
		}
		c.mustAuxLabel(id).SetStyle(p.body)
	}
}

func oceanPalette() demoPalette {
	accent := core.RGB(37, 99, 235)
	accentSoft := core.RGB(226, 240, 255)
	panelBorder := core.RGB(214, 222, 234)
	pageBg := core.RGB(244, 247, 251)
	return demoPalette{
		name:          "Ocean Blue",
		page:          widgets.PanelStyle{Background: pageBg},
		card:          widgets.PanelStyle{Background: core.RGB(255, 255, 255), BorderColor: panelBorder, BorderWidth: 1, CornerRadius: 18},
		cardMuted:     widgets.PanelStyle{Background: core.RGB(248, 250, 252), BorderColor: core.RGB(205, 216, 230), BorderWidth: 1, CornerRadius: 16},
		scroll:        widgets.PanelStyle{Background: core.RGB(255, 255, 255), BorderColor: panelBorder, BorderWidth: 1, CornerRadius: 18},
		buttonPrimary: makeButtonStyle(core.RGB(245, 249, 255), core.RGB(173, 201, 236), accentSoft, accent, core.RGB(15, 23, 42), core.RGB(255, 255, 255)),
		buttonNeutral: makeButtonStyle(core.RGB(255, 255, 255), core.RGB(205, 216, 230), core.RGB(240, 247, 255), core.RGB(58, 116, 214), core.RGB(30, 41, 59), core.RGB(255, 255, 255)),
		buttonGhost:   makeButtonStyle(core.RGB(255, 255, 255), core.RGB(205, 216, 230), accentSoft, core.RGB(30, 64, 175), core.RGB(37, 99, 235), core.RGB(255, 255, 255)),
		edit:          makeEditStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(96, 165, 250), accent, core.RGB(31, 41, 55), core.RGB(148, 163, 184), accent),
		reportEdit:    makeReportStyle(core.RGB(248, 250, 252), panelBorder, core.RGB(96, 165, 250), accent, core.RGB(15, 23, 42)),
		choiceDot:     makeChoiceStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(56, 189, 248), accent, accent, widgets.ChoiceIndicatorDot, 6),
		choiceCheck:   makeChoiceStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(56, 189, 248), accent, accent, widgets.ChoiceIndicatorCheck, 6),
		radioDot:      makeChoiceStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(56, 189, 248), accent, accent, widgets.ChoiceIndicatorDot, 9),
		radioCheck:    makeChoiceStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(56, 189, 248), accent, accent, widgets.ChoiceIndicatorCheck, 9),
		combo:         makeComboStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(96, 165, 250), accent, accent),
		list:          makeListStyle(core.RGB(255, 255, 255), panelBorder, core.RGB(96, 165, 250), accent),
		progressMain:  makeProgressStyle(core.RGB(219, 234, 254), accent, core.RGB(29, 78, 216)),
		progressAlt:   makeProgressStyle(core.RGB(209, 250, 229), core.RGB(16, 185, 129), core.RGB(5, 150, 105)),
		title:         makeTextStyle("Microsoft YaHei UI", 22, 700, core.RGB(15, 23, 42)),
		subtitle:      makeTextStyle("Microsoft YaHei UI", 13, 400, core.RGB(100, 116, 139)),
		sectionTitle:  makeTextStyle("Microsoft YaHei UI", 18, 700, core.RGB(15, 23, 42)),
		body:          makeTextStyle("Microsoft YaHei UI", 14, 400, core.RGB(71, 85, 105)),
		accentText:    makeTextStyle("Microsoft YaHei UI", 14, 700, accent),
		modalSurface:  widgets.PanelStyle{Background: core.RGB(255, 255, 255), BorderColor: panelBorder, BorderWidth: 1, CornerRadius: 18},
		modalBackdrop: core.RGB(15, 23, 42),
		modalOpacity:  96,
		modalBlurDP:   8,
	}
}

func graphitePalette() demoPalette {
	accent := core.RGB(107, 114, 128)
	border := core.RGB(201, 206, 214)
	pageBg := core.RGB(236, 239, 242)
	return demoPalette{
		name:          "Graphite Gray",
		page:          widgets.PanelStyle{Background: pageBg},
		card:          widgets.PanelStyle{Background: core.RGB(250, 251, 252), BorderColor: border, BorderWidth: 1, CornerRadius: 18},
		cardMuted:     widgets.PanelStyle{Background: core.RGB(242, 244, 246), BorderColor: core.RGB(214, 218, 224), BorderWidth: 1, CornerRadius: 16},
		scroll:        widgets.PanelStyle{Background: core.RGB(249, 250, 251), BorderColor: border, BorderWidth: 1, CornerRadius: 18},
		buttonPrimary: makeButtonStyle(core.RGB(244, 245, 246), border, core.RGB(233, 236, 240), accent, core.RGB(31, 41, 55), core.RGB(255, 255, 255)),
		buttonNeutral: makeButtonStyle(core.RGB(248, 249, 250), border, core.RGB(239, 241, 243), core.RGB(85, 94, 109), core.RGB(31, 41, 55), core.RGB(255, 255, 255)),
		buttonGhost:   makeButtonStyle(core.RGB(250, 251, 252), border, core.RGB(239, 241, 243), core.RGB(85, 94, 109), accent, core.RGB(255, 255, 255)),
		edit:          makeEditStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, core.RGB(31, 41, 55), core.RGB(148, 163, 184), accent),
		reportEdit:    makeReportStyle(core.RGB(244, 245, 246), border, core.RGB(156, 163, 175), accent, core.RGB(17, 24, 39)),
		choiceDot:     makeChoiceStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, accent, widgets.ChoiceIndicatorDot, 6),
		choiceCheck:   makeChoiceStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, accent, widgets.ChoiceIndicatorCheck, 6),
		radioDot:      makeChoiceStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, accent, widgets.ChoiceIndicatorDot, 9),
		radioCheck:    makeChoiceStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, accent, widgets.ChoiceIndicatorCheck, 9),
		combo:         makeComboStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent, accent),
		list:          makeListStyle(core.RGB(251, 252, 253), border, core.RGB(156, 163, 175), accent),
		progressMain:  makeProgressStyle(core.RGB(223, 227, 232), accent, core.RGB(75, 85, 99)),
		progressAlt:   makeProgressStyle(core.RGB(231, 233, 236), core.RGB(120, 125, 134), core.RGB(82, 89, 100)),
		title:         makeTextStyle("Microsoft YaHei UI", 22, 700, core.RGB(17, 24, 39)),
		subtitle:      makeTextStyle("Microsoft YaHei UI", 13, 400, core.RGB(107, 114, 128)),
		sectionTitle:  makeTextStyle("Microsoft YaHei UI", 18, 700, core.RGB(31, 41, 55)),
		body:          makeTextStyle("Microsoft YaHei UI", 14, 400, core.RGB(75, 85, 99)),
		accentText:    makeTextStyle("Microsoft YaHei UI", 14, 700, accent),
		modalSurface:  widgets.PanelStyle{Background: core.RGB(250, 251, 252), BorderColor: border, BorderWidth: 1, CornerRadius: 18},
		modalBackdrop: core.RGB(31, 41, 55),
		modalOpacity:  84,
		modalBlurDP:   6,
	}
}

func makeTextStyle(face string, size int32, weight int32, color core.Color) widgets.TextStyle {
	return widgets.TextStyle{
		Font: widgets.FontSpec{
			Face:   face,
			SizeDP: size,
			Weight: weight,
		},
		Color: color,
	}
}

func makeButtonStyle(bg, border, hover, pressed, fg, downFG core.Color) widgets.ButtonStyle {
	return widgets.ButtonStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
			Weight: 600,
		},
		TextAlign:    widgets.AlignCenter,
		TextColor:    fg,
		DownText:     downFG,
		DisabledText: core.RGB(148, 163, 184),
		Background:   bg,
		Hover:        hover,
		Pressed:      pressed,
		Disabled:     core.RGB(229, 231, 235),
		Border:       border,
		CornerRadius: 12,
		IconSizeDP:   18,
		TextInsetDP:  18,
		GapDP:        8,
		PadDP:        12,
	}
}

func makeEditStyle(bg, border, hoverBorder, focusBorder, fg, ph, caret core.Color) widgets.EditStyle {
	return widgets.EditStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		TextColor:        fg,
		PlaceholderColor: ph,
		Background:       bg,
		BorderColor:      border,
		HoverBorder:      hoverBorder,
		FocusBorder:      focusBorder,
		DisabledText:     core.RGB(156, 163, 175),
		DisabledBg:       core.RGB(243, 244, 246),
		CaretColor:       caret,
		PaddingDP:        10,
		CornerRadius:     12,
	}
}

func makeReportStyle(bg, border, hoverBorder, focusBorder, fg core.Color) widgets.EditStyle {
	style := makeEditStyle(bg, border, hoverBorder, focusBorder, fg, core.RGB(148, 163, 184), focusBorder)
	style.Font = widgets.FontSpec{Face: "Cascadia Mono", SizeDP: 13}
	return style
}

func makeChoiceStyle(bg, border, hoverBorder, accent, check core.Color, indicator widgets.ChoiceIndicatorStyle, radius int32) widgets.ChoiceStyle {
	return widgets.ChoiceStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		TextColor:       core.RGB(31, 41, 55),
		DisabledText:    core.RGB(156, 163, 175),
		Background:      bg,
		BorderColor:     border,
		HoverBorder:     hoverBorder,
		FocusBorder:     accent,
		IndicatorColor:  accent,
		CheckColor:      check,
		IndicatorStyle:  indicator,
		HoverBackground: core.RGB(241, 245, 249),
		DisabledBg:      core.RGB(243, 244, 246),
		DisabledBorder:  core.RGB(209, 213, 219),
		CornerRadius:    radius,
		IndicatorSizeDP: 18,
		IndicatorGapDP:  10,
	}
}

func makeComboStyle(bg, border, hoverBorder, focusBorder, accent core.Color) widgets.ComboStyle {
	return widgets.ComboStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		TextColor:         core.RGB(31, 41, 55),
		PlaceholderColor:  core.RGB(148, 163, 184),
		Background:        bg,
		BorderColor:       border,
		HoverBorder:       hoverBorder,
		FocusBorder:       focusBorder,
		ArrowColor:        accent,
		PopupBackground:   bg,
		ItemHoverColor:    core.RGB(239, 246, 255),
		ItemSelectedColor: accent,
		ItemTextColor:     core.RGB(255, 255, 255),
		CornerRadius:      12,
		PaddingDP:         10,
		ItemHeightDP:      34,
		MaxVisibleItems:   6,
	}
}

func makeListStyle(bg, border, hoverBorder, accent core.Color) widgets.ListStyle {
	return widgets.ListStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 15,
		},
		TextColor:         core.RGB(31, 41, 55),
		DisabledText:      core.RGB(156, 163, 175),
		Background:        bg,
		BorderColor:       border,
		HoverBorder:       hoverBorder,
		FocusBorder:       accent,
		ItemHoverColor:    core.RGB(239, 246, 255),
		ItemSelectedColor: accent,
		ItemTextColor:     core.RGB(255, 255, 255),
		CornerRadius:      12,
		PaddingDP:         8,
		ItemHeightDP:      34,
	}
}

func makeProgressStyle(track, fill, bubble core.Color) widgets.ProgressStyle {
	return widgets.ProgressStyle{
		Font: widgets.FontSpec{
			Face:   "Microsoft YaHei UI",
			SizeDP: 13,
			Weight: 700,
		},
		TextColor:    core.RGB(255, 255, 255),
		TrackColor:   track,
		FillColor:    fill,
		BubbleColor:  bubble,
		CornerRadius: 12,
		ShowPercent:  true,
	}
}

func mustWidget[T any](window *jsonui.Window, id string) T {
	if window == nil {
		panic("window is nil")
	}
	widget := window.FindWidget(id)
	typed, ok := widget.(T)
	if !ok {
		panic(fmt.Sprintf("widget %q has type %T", id, widget))
	}
	return typed
}

func (c *demoController) mustPanel(id string) *widgets.Panel {
	return mustWidget[*widgets.Panel](c.window, id)
}
func (c *demoController) mustModal(id string) *widgets.Modal {
	return mustWidget[*widgets.Modal](c.window, id)
}
func (c *demoController) mustLabel(id string) *widgets.Label {
	return mustWidget[*widgets.Label](c.window, id)
}
func (c *demoController) mustButton(id string) *widgets.Button {
	return mustWidget[*widgets.Button](c.window, id)
}
func (c *demoController) mustEdit(id string) *widgets.EditBox {
	return mustWidget[*widgets.EditBox](c.window, id)
}
func (c *demoController) mustCheckBox(id string) *widgets.CheckBox {
	return mustWidget[*widgets.CheckBox](c.window, id)
}
func (c *demoController) mustRadio(id string) *widgets.RadioButton {
	return mustWidget[*widgets.RadioButton](c.window, id)
}
func (c *demoController) mustCombo(id string) *widgets.ComboBox {
	return mustWidget[*widgets.ComboBox](c.window, id)
}
func (c *demoController) mustListBox(id string) *widgets.ListBox {
	return mustWidget[*widgets.ListBox](c.window, id)
}
func (c *demoController) mustFilePicker(id string) *widgets.FilePicker {
	return mustWidget[*widgets.FilePicker](c.window, id)
}
func (c *demoController) mustImage(id string) *widgets.Image {
	return mustWidget[*widgets.Image](c.window, id)
}
func (c *demoController) mustAnimated(id string) *widgets.AnimatedImage {
	return mustWidget[*widgets.AnimatedImage](c.window, id)
}
func (c *demoController) mustProgress(id string) *widgets.ProgressBar {
	return mustWidget[*widgets.ProgressBar](c.window, id)
}
func (c *demoController) mustScroll(id string) *widgets.ScrollView {
	return mustWidget[*widgets.ScrollView](c.window, id)
}

func (c *demoController) mustAuxLabel(id string) *widgets.Label {
	return mustWidget[*widgets.Label](c.doc.Window("aux"), id)
}

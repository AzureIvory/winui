//go:build windows

package main

import (
	"fmt"
	"path/filepath"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/sysapi"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

type demoController struct {
	baseDir   string
	assetsDir string
	store     *jsonui.Store
	doc       *jsonui.Document
	window    *jsonui.Window
	lang      *demoLang
	locale    string
	useAlt    bool
	mode      widgets.ControlMode
}

type demoPalette struct {
	name          string
	nameKey       string
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

func newDemoController(baseDir string, store *jsonui.Store, doc *jsonui.Document, window *jsonui.Window, mode widgets.ControlMode) *demoController {
	if mode != widgets.ModeNative {
		mode = widgets.ModeCustom
	}
	controller := &demoController{
		baseDir:   baseDir,
		assetsDir: filepath.Join(baseDir, "assets"),
		store:     store,
		doc:       doc,
		window:    window,
		locale:    "en",
		mode:      mode,
	}
	if lang, err := loadDemoLang(filepath.Join(controller.assetsDir, "lang.json")); err == nil {
		controller.lang = lang
	}
	controller.bindCallbacks()
	controller.applyPalette(oceanPalette())
	controller.applyLanguage("en")
	controller.setStatus(controller.tr("status.ready", "Ready"))
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

func (c *demoController) tr(path string, fallback string) string {
	if c == nil {
		return fallback
	}
	if c.lang == nil {
		return fallback
	}
	return c.lang.text(c.locale, path, fallback)
}

func (c *demoController) trf(path string, fallback string, args ...any) string {
	return fmt.Sprintf(c.tr(path, fallback), args...)
}

func (c *demoController) toggleLanguage() {
	if c == nil {
		return
	}
	if normalizeDemoLocale(c.locale) == "en" {
		c.applyLanguage("zh")
		return
	}
	c.applyLanguage("en")
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
	displayName := c.paletteDisplayName(palette)
	c.store.Set("demo.paletteName", displayName)
	c.setStatus(c.trf("status.paletteSwitched", "Palette switched to %s", displayName))
}

func (c *demoController) controlModeLabel() string {
	if c != nil && c.mode == widgets.ModeNative {
		return c.tr("mode.native", "Native controls")
	}
	return c.tr("mode.custom", "Custom draw")
}

func (c *demoController) controlModeButtonText() string {
	return c.trf("i18n.header.toggleControlModeBtn", "Mode: %s", c.controlModeLabel())
}

func (c *demoController) nextControlMode() widgets.ControlMode {
	if c != nil && c.mode == widgets.ModeNative {
		return widgets.ModeCustom
	}
	return widgets.ModeNative
}

func (c *demoController) toggleControlMode() {
	if c == nil {
		return
	}
	nextMode := c.nextControlMode()
	if err := c.reloadWindowForControlMode(nextMode); err != nil {
		c.setStatus(c.trf("status.controlModeSwitchFailed", "Failed to switch control mode: %v", err))
		return
	}
	c.setStatus(c.trf("status.controlModeSwitched", "Control mode switched to %s", c.controlModeLabel()))
}

func (c *demoController) reloadWindowForControlMode(mode widgets.ControlMode) error {
	if c == nil {
		return fmt.Errorf("controller is nil")
	}
	if mode != widgets.ModeNative {
		mode = widgets.ModeCustom
	}

	oldDoc := c.doc
	oldWindow := c.window
	var scene *widgets.Scene
	var app *core.App
	if oldWindow != nil {
		scene = oldWindow.Scene()
		app = oldWindow.App()
	}

	newDoc, store, err := loadDemoDocumentWithMode(c.baseDir, mode, c.store)
	if err != nil {
		return err
	}
	newWindow := newDoc.PrimaryWindow()
	if newWindow == nil {
		return fmt.Errorf("primary window is nil")
	}

	if scene != nil && oldWindow != nil {
		_ = oldWindow.Detach()
		if err := newWindow.Attach(scene); err != nil {
			_ = oldWindow.Attach(scene)
			return err
		}
	}

	c.doc = newDoc
	c.window = newWindow
	c.store = store
	c.mode = mode

	c.bindCallbacks()
	c.applyPalette(c.currentPalette())
	c.applyLanguage(c.locale)
	c.ensureSpinnerPlaying()

	if app != nil && c.window != nil && c.window.Root != nil {
		size := app.ClientSize()
		c.window.Root.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
	}

	if oldDoc != nil && oldDoc != newDoc {
		for _, win := range oldDoc.Windows {
			if win == nil || win == oldWindow {
				continue
			}
			_ = win.Detach()
		}
	}
	return nil
}

func (c *demoController) showModal(visible bool) {
	if c == nil || c.store == nil {
		return
	}
	c.store.Set("demo.modalVisible", visible)
	if visible {
		c.setStatus(c.tr("status.modalOpened", "Modal opened"))
		return
	}
	c.setStatus(c.tr("status.modalClosed", "Modal closed"))
}

func (c *demoController) bindCallbacks() {
	c.mustButton("togglePaletteBtn").SetOnClick(func() {
		c.togglePalette()
	})
	c.mustButton("toggleControlModeBtn").SetOnClick(func() {
		c.toggleControlMode()
	})
	c.mustButton("languageToggleBtn").SetOnClick(func() {
		c.toggleLanguage()
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
		c.setStatus(c.tr("status.interactivePanelClicked", "Interactive panel clicked"))
	})
	c.mustModal("helpModal").SetOnDismiss(func() {
		c.showModal(false)
	})

	c.mustEdit("nameInput").SetOnChange(func(value string) {
		c.setStatus(c.trf("status.nameChanged", "Name input changed: %s", value))
	})
	c.mustEdit("passwordInput").SetOnSubmit(func(value string) {
		c.setStatus(c.trf("status.passwordSubmitted", "Password submitted, rune length: %d", len([]rune(value))))
	})

	c.mustCheckBox("dotStyleBox").SetOnChange(func(checked bool) {
		c.setStatus(c.trf(
			"status.choiceChanged",
			"%s changed to %s",
			c.mustCheckBox("dotStyleBox").Text,
			c.checkedLabel(checked),
		))
	})
	c.mustCheckBox("checkStyleBox").SetOnChange(func(checked bool) {
		c.setStatus(c.trf(
			"status.choiceChanged",
			"%s changed to %s",
			c.mustCheckBox("checkStyleBox").Text,
			c.checkedLabel(checked),
		))
	})
	c.mustRadio("radioDotA").SetOnChange(func(checked bool) {
		if checked {
			c.setStatus(c.trf("status.radioSelected", "%s selected", c.mustRadio("radioDotA").Text))
		}
	})
	c.mustRadio("radioCheckA").SetOnChange(func(checked bool) {
		if checked {
			c.setStatus(c.trf("status.radioSelected", "%s selected", c.mustRadio("radioCheckA").Text))
		}
	})

	c.mustCombo("citySelect").SetOnChange(func(index int, item widgets.ListItem) {
		c.setStatus(c.trf("status.comboSelected", "%s selected: %d / %s", c.tr("i18n.data.comboLabel", "City combo box"), index, item.Value))
	})
	c.mustListBox("cityList").SetOnChange(func(index int, item widgets.ListItem) {
		c.setStatus(c.trf("status.listChanged", "%s changed: %d / %s", c.tr("i18n.data.listLabel", "Rollout list"), index, item.Value))
	})
	c.mustListBox("cityList").SetOnActivate(func(index int, item widgets.ListItem) {
		c.setStatus(c.trf("status.listActivated", "%s activated: %d / %s", c.tr("i18n.data.listLabel", "Rollout list"), index, item.Value))
	})
	c.mustListBox("cityList").SetOnRightClick(func(index int, item widgets.ListItem, _ core.Point) {
		c.setStatus(c.trf("status.listRightClick", "%s right-click: %d / %s", c.tr("i18n.data.listLabel", "Rollout list"), index, item.Value))
	})

	for _, id := range []string{"openFile", "saveFile", "folderPick", "multiFiles"} {
		pickerID := id
		picker := c.mustFilePicker(pickerID)
		picker.SetOnChange(func(paths []string) {
			if len(paths) == 0 {
				c.setStatus(c.trf("status.pickerCleared", "%s cleared", pickerID))
				return
			}
			c.setStatus(c.trf("status.pickerSelected", "%s selected %d path(s)", pickerID, len(paths)))
		})
	}
}

func (c *demoController) ensureSpinnerPlaying() {
	if c == nil {
		return
	}
	spinner := c.mustAnimated("spinnerImage")
	spinner.SetPlaying(false)
	spinner.SetPlaying(true)
}

func (c *demoController) checkedLabel(checked bool) string {
	if checked {
		return c.tr("common.checked", "checked")
	}
	return c.tr("common.unchecked", "unchecked")
}

func (c *demoController) paletteDisplayName(p demoPalette) string {
	if p.nameKey != "" {
		return c.tr("paletteNames."+p.nameKey, p.name)
	}
	return p.name
}

func (c *demoController) currentPalette() demoPalette {
	if c != nil && c.useAlt {
		return graphitePalette()
	}
	return oceanPalette()
}

func (c *demoController) applyLanguage(locale string) {
	if c == nil {
		return
	}
	c.locale = normalizeDemoLocale(locale)

	c.setTextOnWindowWidget(c.window, "titleLabel", c.tr("i18n.header.titleLabel", "WinUI JSON Full Demo"))
	c.setTextOnWindowWidget(c.window, "subtitleLabel", c.tr("i18n.header.subtitleLabel", "One native JSON page that showcases every public control, palette switching, and API checks."))
	c.setTextOnWindowWidget(c.window, "togglePaletteBtn", c.tr("i18n.header.togglePaletteBtn", "Switch palette"))
	c.setTextOnWindowWidget(c.window, "toggleControlModeBtn", c.controlModeButtonText())
	c.setTextOnWindowWidget(c.window, "runAllBtn", c.tr("i18n.header.runAllBtn", "Run widget API tests"))
	c.setTextOnWindowWidget(c.window, "languageToggleBtn", c.tr("i18n.header.languageToggleBtn", "中文"))
	c.setTextOnWindowWidget(c.window, "openModalBtn", c.tr("i18n.header.openModalBtn", "Show modal"))

	c.setTextOnWindowWidget(c.window, "introTitle", c.tr("i18n.intro.title", "Coverage"))
	c.setTextOnWindowWidget(c.window, "introBody", c.tr("i18n.intro.body", "This page covers Panel, Modal, Label, Button, EditBox, ProgressBar, CheckBox, RadioButton, ComboBox, ListBox, FilePicker, Image, AnimatedImage, and ScrollView."))

	c.setTextOnWindowWidget(c.window, "buttonsTitle", c.tr("i18n.buttons.title", "Buttons"))
	c.setTextOnWindowWidget(c.window, "buttonsHint", c.tr("i18n.buttons.hint", "Primary, image-left, image-top, ghost, and disabled states are all included in one section."))
	c.setTextOnWindowWidget(c.window, "saveBtn", c.tr("i18n.buttons.saveBtn", "Save changes"))
	c.setTextOnWindowWidget(c.window, "imageTopBtn", c.tr("i18n.buttons.imageTopBtn", "Top image"))
	c.setTextOnWindowWidget(c.window, "ghostBtn", c.tr("i18n.buttons.ghostBtn", "Ghost action"))
	c.setTextOnWindowWidget(c.window, "disabledBtn", c.tr("i18n.buttons.disabledBtn", "Disabled action"))

	c.setTextOnWindowWidget(c.window, "inputTitle", c.tr("i18n.inputs.title", "Inputs"))
	c.setTextOnWindowWidget(c.window, "inputHint", c.tr("i18n.inputs.hint", "Single-line, password, and read-only multiline input states are all present here."))
	c.mustEdit("nameInput").SetPlaceholder(c.tr("i18n.inputs.namePlaceholder", "Project name"))
	c.mustEdit("passwordInput").SetPlaceholder(c.tr("i18n.inputs.passwordPlaceholder", "Type a secret"))
	c.mustEdit("notesBox").SetPlaceholder(c.tr("i18n.inputs.notesPlaceholder", "Notes"))
	c.mustEdit("notesBox").SetText(c.tr("i18n.inputs.notesValue", "JSON declares structure and bindings. The host runtime still owns mutations, actions, and API checks."))

	c.setTextOnWindowWidget(c.window, "choiceTitle", c.tr("i18n.choices.title", "Checks and radios"))
	c.setTextOnWindowWidget(c.window, "choiceHint", c.tr("i18n.choices.hint", "Checkboxes show both dot and check indicators. Radios include the default dot and explicit check styles."))
	c.setTextOnWindowWidget(c.window, "dotStyleBox", c.tr("i18n.choices.dotStyleBox", "Dot-style checkbox"))
	c.setTextOnWindowWidget(c.window, "checkStyleBox", c.tr("i18n.choices.checkStyleBox", "Check-style checkbox"))
	c.setTextOnWindowWidget(c.window, "disabledRule", c.tr("i18n.choices.disabledRule", "Disabled selected rule"))
	c.setTextOnWindowWidget(c.window, "radioDotA", c.tr("i18n.choices.radioDotA", "Notify immediately"))
	c.setTextOnWindowWidget(c.window, "radioDotB", c.tr("i18n.choices.radioDotB", "Notify later"))
	c.setTextOnWindowWidget(c.window, "radioCheckA", c.tr("i18n.choices.radioCheckA", "Approve with check style"))
	c.setTextOnWindowWidget(c.window, "radioCheckB", c.tr("i18n.choices.radioCheckB", "Escalate with check style"))

	c.setTextOnWindowWidget(c.window, "dataTitle", c.tr("i18n.data.title", "Progress and selection"))
	c.setTextOnWindowWidget(c.window, "dataHint", c.tr("i18n.data.hint", "Two progress bars, one combo box, and one list box cover the selection and data-display widgets."))
	c.localizeSelectionWidgets()

	c.setTextOnWindowWidget(c.window, "fileTitle", c.tr("i18n.files.title", "File pickers"))
	c.setTextOnWindowWidget(c.window, "fileHint", c.tr("i18n.files.hint", "Open, save, folder, and multi-select file pickers are all present in one section."))
	c.localizeFilePickers()

	c.setTextOnWindowWidget(c.window, "mediaTitle", c.tr("i18n.media.title", "Images"))
	c.setTextOnWindowWidget(c.window, "mediaHint", c.tr("i18n.media.hint", "A static PNG preview and an animated GIF prove that both image controls work from the same JSON document."))
	c.setTextOnWindowWidget(c.window, "emojiBurstLabel", c.tr("i18n.media.emojiBurstLabel", "😀 🚀 🧠 🧩 ✨ 🎯 🔥 🛠️"))

	c.setTextOnWindowWidget(c.window, "scrollTitle", c.tr("i18n.scroll.title", "ScrollView"))
	c.setTextOnWindowWidget(c.window, "scrollHint", c.tr("i18n.scroll.hint", "The outer document uses a vertical ScrollView, and this card contains a horizontal ScrollView for Add, Remove, SetContent, and scrolling APIs."))
	scrollButtonIDs := []string{"miniScrollBtn1", "miniScrollBtn2", "miniScrollBtn3", "miniScrollBtn4", "miniScrollBtn5", "miniScrollBtn6"}
	scrollButtonFallbacks := []string{"Card A", "Card B", "Card C", "Card D", "Card E", "Card F"}
	scrollTexts := c.langStringSlice("i18n.scroll.buttons", scrollButtonFallbacks)
	for idx, id := range scrollButtonIDs {
		if idx < len(scrollTexts) {
			c.setTextOnWindowWidget(c.window, id, scrollTexts[idx])
		}
	}

	c.setTextOnWindowWidget(c.window, "panelTitle", c.tr("i18n.panel.title", "Panel and modal hooks"))
	c.setTextOnWindowWidget(c.window, "panelHint", c.tr("i18n.panel.hint", "Click the panel below to exercise the panel click callback. The modal above is used for backdrop, blur, and dismiss APIs."))
	c.setTextOnWindowWidget(c.window, "panelInfo", c.tr("i18n.panel.info", "Click this surface to update the status area. The function test also exercises Panel.Add, Panel.Remove, and Panel.SetLayout on runtime-created content."))

	c.setTextOnWindowWidget(c.window, "resultsTitle", c.tr("i18n.results.title", "API report"))
	c.setTextOnWindowWidget(c.window, "statusCaption", c.tr("i18n.results.statusCaption", "Status"))
	c.setTextOnWindowWidget(c.window, "resultsHint", c.tr("i18n.results.hint", "The button writes the full PASS / FAIL detail report to a local txt file. The UI only shows the latest summary and file path."))
	c.setTextOnWindowWidget(c.window, "reportSummaryCaption", c.tr("i18n.results.reportSummaryCaption", "Last summary"))
	c.setTextOnWindowWidget(c.window, "reportPathCaption", c.tr("i18n.results.reportPathCaption", "Report file"))

	c.setTextOnWindowWidget(c.window, "modalTitle", c.tr("i18n.modal.title", "Native modal panel"))
	c.setTextOnWindowWidget(c.window, "modalBody", c.tr("i18n.modal.body", "This modal is still declared by JSON. Visibility is bound through the store, while the host wires actions and API tests at runtime."))
	c.setTextOnWindowWidget(c.window, "modalStatus", c.tr("i18n.modal.status", "Backdrop click will dismiss the modal."))
	c.setTextOnWindowWidget(c.window, "closeModalBtn", c.tr("i18n.modal.closeButton", "Close modal"))

	if aux := c.doc.Window("aux"); aux != nil {
		aux.Meta.Title = c.tr("auxWindowTitle", "WinUI JSON Aux")
		if app := aux.App(); app != nil {
			app.SetTitle(aux.Meta.Title)
		}
	}
	if auxLabel := c.doc.FindWidget("aux", "auxLabel"); auxLabel != nil {
		if label, ok := auxLabel.(*widgets.Label); ok {
			label.SetText(c.tr("i18n.aux.body", "A second window exists so the demo can exercise jsonui Window.Attach, Window.Detach, and document window lookup helpers."))
		}
	}

	if c.store != nil {
		c.store.Set("demo.windowTitle", c.tr("windowTitle", "WinUI JSON Full Demo"))
		c.store.Set("demo.paletteName", c.paletteDisplayName(c.currentPalette()))
		if raw, ok := c.store.Get("demo.report"); !ok || fmt.Sprint(raw) == "" {
			c.setReportSummary(c.tr("report.noRun", defaultReportSummary()))
		}
	}

	c.setStatus(c.tr("status.ready", "Ready"))
}

func (c *demoController) setTextOnWindowWidget(window *jsonui.Window, id string, text string) {
	if window == nil || id == "" {
		return
	}
	widget := window.FindWidget(id)
	switch typed := widget.(type) {
	case *widgets.Label:
		typed.SetText(text)
	case *widgets.Button:
		typed.SetText(text)
	case *widgets.CheckBox:
		typed.SetText(text)
	case *widgets.RadioButton:
		typed.SetText(text)
	}
}

func (c *demoController) langStringSlice(path string, fallback []string) []string {
	if c == nil || c.lang == nil {
		return append([]string(nil), fallback...)
	}
	raw, ok := c.lang.value(c.locale, path)
	if !ok {
		return append([]string(nil), fallback...)
	}
	nodes, ok := raw.([]any)
	if !ok {
		return append([]string(nil), fallback...)
	}
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if text, ok := node.(string); ok {
			values = append(values, text)
		}
	}
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return values
}

func (c *demoController) localizeSelectionWidgets() {
	combo := c.mustCombo("citySelect")
	comboSelectedValue := ""
	if item, ok := combo.SelectedItem(); ok {
		comboSelectedValue = item.Value
	}
	combo.SetPlaceholder(c.tr("i18n.data.comboPlaceholder", "Select a city"))
	comboItems := combo.Items()
	if c.lang != nil {
		comboItems = c.lang.listItems(c.locale, "i18n.data.comboItems", comboItems)
	}
	combo.SetItems(comboItems)
	_ = setComboSelectionByValue(combo, comboSelectedValue)

	list := c.mustListBox("cityList")
	listSelectedValue := ""
	if item, ok := list.SelectedItem(); ok {
		listSelectedValue = item.Value
	}
	listItems := list.Items()
	if c.lang != nil {
		listItems = c.lang.listItems(c.locale, "i18n.data.listItems", listItems)
	}
	list.SetItems(listItems)
	_ = setListSelectionByValue(list, listSelectedValue)
}

func (c *demoController) localizeFilePickers() {
	c.localizeFilePicker(
		c.mustFilePicker("openFile"),
		"i18n.files.openFile",
		sysapi.DialogOpen,
		false,
		"",
		[]sysapi.FileFilter{{Name: "Source Files", Pattern: "*.md;*.txt;*.go"}},
		"No file selected",
		"Open file",
		"Select a source file",
	)
	c.localizeFilePicker(
		c.mustFilePicker("saveFile"),
		"i18n.files.saveFile",
		sysapi.DialogSave,
		false,
		"txt",
		[]sysapi.FileFilter{
			{Name: "Text Files", Pattern: "*.txt"},
			{Name: "Markdown", Pattern: "*.md"},
			{Name: "All Files", Pattern: "*.*"},
		},
		"Choose an output path",
		"Save report",
		"Export the API report",
	)
	c.localizeFilePicker(
		c.mustFilePicker("folderPick"),
		"i18n.files.folderPick",
		sysapi.DialogFolder,
		false,
		"",
		nil,
		"No folder selected",
		"Pick folder",
		"Select a workspace folder",
	)
	multi := c.mustFilePicker("multiFiles")
	multi.SetSeparator(" | ")
	c.localizeFilePicker(
		multi,
		"i18n.files.multiFiles",
		sysapi.DialogOpen,
		true,
		"",
		[]sysapi.FileFilter{{Name: "Project Files", Pattern: "*.go;*.json;*.md"}},
		"No files selected",
		"Pick files",
		"Select multiple project files",
	)
}

func (c *demoController) localizeFilePicker(
	picker *widgets.FilePicker,
	basePath string,
	mode sysapi.DialogMode,
	multiple bool,
	defaultExt string,
	fallbackFilters []sysapi.FileFilter,
	fallbackPlaceholder string,
	fallbackButton string,
	fallbackDialogTitle string,
) {
	if picker == nil {
		return
	}
	buttonText := c.tr(basePath+".buttonText", fallbackButton)
	dialogTitle := c.tr(basePath+".dialogTitle", fallbackDialogTitle)
	placeholder := c.tr(basePath+".placeholder", fallbackPlaceholder)

	options := picker.DialogOptions()
	options.Mode = mode
	options.MultiSelect = multiple
	options.DefaultExtension = defaultExt
	options.Title = dialogTitle
	options.ButtonLabel = buttonText
	if c.lang != nil {
		options.Filters = c.lang.fileFilters(c.locale, basePath+".filters", fallbackFilters)
	} else {
		options.Filters = append([]sysapi.FileFilter(nil), fallbackFilters...)
	}
	picker.SetDialogOptions(options)
	picker.SetButtonText(buttonText)
	picker.SetPlaceholder(placeholder)
}

func setComboSelectionByValue(combo *widgets.ComboBox, value string) bool {
	if combo == nil || value == "" {
		return false
	}
	items := combo.Items()
	for idx, item := range items {
		if item.Value == value {
			combo.SetSelected(idx)
			return true
		}
	}
	return false
}

func setListSelectionByValue(list *widgets.ListBox, value string) bool {
	if list == nil || value == "" {
		return false
	}
	items := list.Items()
	for idx, item := range items {
		if item.Value == value {
			list.SetSelected(idx)
			return true
		}
	}
	return false
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
	c.mustButton("toggleControlModeBtn").SetStyle(p.buttonNeutral)
	c.mustButton("languageToggleBtn").SetStyle(p.buttonNeutral)
	c.mustButton("runAllBtn").SetStyle(p.buttonPrimary)
	c.mustButton("openModalBtn").SetStyle(p.buttonGhost)
	c.mustButton("closeModalBtn").SetStyle(p.buttonNeutral)
	c.mustButton("saveBtn").SetStyle(p.buttonPrimary)
	c.mustButton("imageTopBtn").SetStyle(p.buttonNeutral)
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
		"emojiBurstLabel",
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
	if c.store != nil {
		c.store.Set("demo.paletteName", c.paletteDisplayName(p))
	}
}

func oceanPalette() demoPalette {
	accent := core.RGB(37, 99, 235)
	accentSoft := core.RGB(226, 240, 255)
	panelBorder := core.RGB(214, 222, 234)
	pageBg := core.RGB(244, 247, 251)
	return demoPalette{
		name:          "Ocean Blue",
		nameKey:       "ocean",
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
		nameKey:       "graphite",
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
		ImageSizeDP:  18,
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

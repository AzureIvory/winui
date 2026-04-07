//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/markup"
)

var (
	demoPage    widgets.Widget
	statusLabel *widgets.Label
	keywordEdit *widgets.EditBox
	notesEdit   *widgets.EditBox
	advancedBox *widgets.CheckBox
	progressBar *widgets.ProgressBar
	modeCombo   *widgets.ComboBox
)

func main() {
	opts := core.Options{
		ClassName:      "WinUIMarkupDemo",
		Title:          "winui markup demo",
		Width:          900,
		Height:         680,
		Style:          core.DefaultWindowStyle,
		ExStyle:        core.DefaultWindowExStyle,
		Cursor:         core.CursorArrow,
		Background:     core.RGB(246, 248, 251),
		DoubleBuffered: true,
		RenderMode:     core.RenderModeAuto,
	}
	widgets.BindScene(&opts, widgets.SceneHooks{
		OnCreate:  buildDemo,
		OnResize:  resizeDemo,
		OnDestroy: destroyDemo,
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

func buildDemo(app *core.App, scene *widgets.Scene) error {
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	actions := map[string]func(){
		"applyDemo": func() {
			applyDemoState(app)
		},
	}

	root, err := markup.LoadHTMLFile(filepath.Join(baseDir, "demo.ui.html"), markup.LoadOptions{
		Actions:     actions,
		AssetsDir:   baseDir,
		DefaultMode: widgets.ModeCustom,
	})
	if err != nil {
		return err
	}
	if page, ok := root.(*widgets.Panel); ok {
		page.SetStyle(mergePanelStyle(page.Style, widgets.PanelStyle{Background: core.RGB(246, 248, 251)}))
	}

	demoPage = root
	scene.Root().Add(root)
	root.SetBounds(widgets.Rect{W: app.ClientSize().Width, H: app.ClientSize().Height})

	statusLabel, _ = findWidget(root, "status").(*widgets.Label)
	keywordEdit, _ = findWidget(root, "keyword").(*widgets.EditBox)
	notesEdit, _ = findWidget(root, "notes").(*widgets.EditBox)
	advancedBox, _ = findWidget(root, "advanced").(*widgets.CheckBox)
	progressBar, _ = findWidget(root, "progress").(*widgets.ProgressBar)
	modeCombo, _ = findWidget(root, "modeSelect").(*widgets.ComboBox)

	applyDemoState(app)
	return nil
}

func resizeDemo(_ *core.App, _ *widgets.Scene, size core.Size) {
	if demoPage != nil {
		demoPage.SetBounds(widgets.Rect{W: size.Width, H: size.Height})
	}
}

func destroyDemo(_ *core.App, _ *widgets.Scene) {
	demoPage = nil
	statusLabel = nil
	keywordEdit = nil
	notesEdit = nil
	advancedBox = nil
	progressBar = nil
	modeCombo = nil
}

func applyDemoState(app *core.App) {
	modeText := "未选择"
	if modeCombo != nil {
		if item, ok := modeCombo.SelectedItem(); ok {
			modeText = item.Text
		}
	}
	keyword := ""
	if keywordEdit != nil {
		keyword = keywordEdit.TextValue()
	}
	lineCount := 0
	if notesEdit != nil {
		lineCount = notesEdit.LineCount()
	}
	advanced := false
	if advancedBox != nil {
		advanced = advancedBox.IsChecked()
	}
	progress := int32(35)
	if advanced {
		progress += 25
	}
	if keyword != "" {
		progress += 20
	}
	if lineCount >= 3 {
		progress += 20
		if progress > 100 {
			progress = 100
		}
	}
	if progressBar != nil {
		progressBar.SetValue(progress)
	}
	if statusLabel != nil {
		statusLabel.SetText(fmt.Sprintf("状态：mode=%s，keyword=%s，notes=%d 行，advanced=%v", modeText, displayText(keyword, "（空）"), lineCount, advanced))
	}
	if app != nil {
		app.SetTitle("winui markup demo - " + strings.TrimSpace(modeText))
	}
}

func findWidget(root widgets.Widget, id string) widgets.Widget {
	if root == nil || id == "" {
		return nil
	}
	if root.ID() == id {
		return root
	}
	container, ok := root.(widgets.Container)
	if !ok {
		return nil
	}
	for _, child := range container.Children() {
		if found := findWidget(child, id); found != nil {
			return found
		}
	}
	return nil
}

func displayText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func mergePanelStyle(base, override widgets.PanelStyle) widgets.PanelStyle {
	if override.Background != 0 {
		base.Background = override.Background
	}
	if override.BorderColor != 0 {
		base.BorderColor = override.BorderColor
	}
	if override.CornerRadius != 0 {
		base.CornerRadius = override.CornerRadius
	}
	if override.BorderWidth != 0 {
		base.BorderWidth = override.BorderWidth
	}
	return base
}

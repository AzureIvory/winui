//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
	"github.com/AzureIvory/winui/widgets/jsonui"
)

func TestDemoControllerTogglePaletteAppliesGraphiteStyles(t *testing.T) {
	controller, window := loadDemoControllerForTest(t)

	controller.togglePalette()

	rootPanel, ok := window.FindWidget("pageRoot").(*widgets.Panel)
	if !ok {
		t.Fatalf("pageRoot type = %T, want *widgets.Panel", window.FindWidget("pageRoot"))
	}
	if rootPanel.Style.Background != core.RGB(236, 239, 242) {
		t.Fatalf("pageRoot background = %#08x, want %#08x", rootPanel.Style.Background, core.RGB(236, 239, 242))
	}

	primaryButton, ok := window.FindWidget("saveBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("saveBtn type = %T, want *widgets.Button", window.FindWidget("saveBtn"))
	}
	if primaryButton.Style.Background != core.RGB(244, 245, 246) {
		t.Fatalf("saveBtn background = %#08x, want %#08x", primaryButton.Style.Background, core.RGB(244, 245, 246))
	}

	checkStyleWidget, ok := window.FindWidget("checkStyleBox").(*widgets.CheckBox)
	if !ok {
		t.Fatalf("checkStyleBox type = %T, want *widgets.CheckBox", window.FindWidget("checkStyleBox"))
	}
	if checkStyleWidget.Style.IndicatorColor != core.RGB(107, 114, 128) {
		t.Fatalf("checkStyleBox indicator = %#08x, want %#08x", checkStyleWidget.Style.IndicatorColor, core.RGB(107, 114, 128))
	}
}

func TestDemoControllerRunFunctionTestsBuildsDetailedReport(t *testing.T) {
	controller, _ := loadDemoControllerForTest(t)

	report := controller.runFunctionTests()
	reportPath := filepath.Join(controller.baseDir, "output", "latest-api-check.txt")

	required := []string{
		"Button.SetText",
		"CheckBox.SetChecked",
		"ScrollView.SetScrollOffset",
		"Window.ApplyOptions",
		"Document.FindWidget",
	}
	for _, needle := range required {
		if !strings.Contains(report, needle) {
			t.Fatalf("runFunctionTests report missing %q\n%s", needle, report)
		}
	}
	if !strings.Contains(report, "PASS") {
		t.Fatalf("runFunctionTests report did not contain PASS entries\n%s", report)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) returned error: %v", reportPath, err)
	}
	if !strings.Contains(string(data), "Button.SetText") {
		t.Fatalf("saved report missing method details\n%s", string(data))
	}
	if got, ok := controller.store.Get("demo.reportSummary"); !ok || !strings.Contains(got.(string), "PASS=") {
		t.Fatalf("demo.reportSummary = %#v, ok=%v, want PASS summary", got, ok)
	}
	if got, ok := controller.store.Get("demo.reportPath"); !ok || !strings.Contains(got.(string), "latest-api-check.txt") {
		t.Fatalf("demo.reportPath = %#v, ok=%v, want latest-api-check.txt path", got, ok)
	} else if strings.Contains(got.(string), controller.baseDir) {
		t.Fatalf("demo.reportPath should stay UI-friendly and relative, got %q", got.(string))
	}
	if got, ok := controller.store.Get("demo.report"); ok && strings.Contains(got.(string), "Button.SetText") {
		t.Fatalf("demo.report should not contain detailed method lines in the UI, got %q", got.(string))
	}
}

func TestDemoWindowLaysOutShowcaseCardsWithVisibleHeights(t *testing.T) {
	_, window := loadDemoControllerForTest(t)
	if window.Root == nil {
		t.Fatal("window.Root is nil")
	}
	window.Root.SetBounds(widgets.Rect{W: 1380, H: 940})

	introCard, ok := window.FindWidget("introCard").(*widgets.Panel)
	if !ok {
		t.Fatalf("introCard type = %T, want *widgets.Panel", window.FindWidget("introCard"))
	}
	if got := introCard.Bounds().H; got <= 60 {
		t.Fatalf("introCard.Bounds().H = %d, want > 60", got)
	}

	buttonsCard, ok := window.FindWidget("buttonsCard").(*widgets.Panel)
	if !ok {
		t.Fatalf("buttonsCard type = %T, want *widgets.Panel", window.FindWidget("buttonsCard"))
	}
	if got := buttonsCard.Bounds().H; got <= 80 {
		t.Fatalf("buttonsCard.Bounds().H = %d, want > 80", got)
	}

	saveBtn, ok := window.FindWidget("saveBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("saveBtn type = %T, want *widgets.Button", window.FindWidget("saveBtn"))
	}
	if got := saveBtn.Bounds().H; got <= 0 {
		t.Fatalf("saveBtn.Bounds().H = %d, want > 0", got)
	}

	imageTopBtn, ok := window.FindWidget("imageTopBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("imageTopBtn type = %T, want *widgets.Button", window.FindWidget("imageTopBtn"))
	}
	if got := imageTopBtn.Bounds().H; got <= 56 {
		t.Fatalf("imageTopBtn.Bounds().H = %d, want > 56 for top-image layout", got)
	}
}

func TestDemoRadioGroupsStartExclusiveAndTogglePeers(t *testing.T) {
	controller, _ := loadDemoControllerForTest(t)

	radioDotA := controller.mustRadio("radioDotA")
	radioDotB := controller.mustRadio("radioDotB")
	radioCheckA := controller.mustRadio("radioCheckA")
	radioCheckB := controller.mustRadio("radioCheckB")

	if !radioDotA.IsChecked() {
		t.Fatal("radioDotA should start checked")
	}
	if radioDotB.IsChecked() {
		t.Fatal("radioDotB should start unchecked")
	}
	if !radioCheckA.IsChecked() {
		t.Fatal("radioCheckA should start checked")
	}
	if radioCheckB.IsChecked() {
		t.Fatal("radioCheckB should start unchecked")
	}

	if handled := radioDotB.OnEvent(widgets.Event{Type: widgets.EventClick, Source: radioDotB}); !handled {
		t.Fatal("radioDotB click was not handled")
	}
	if radioDotA.IsChecked() {
		t.Fatal("radioDotA should be cleared after selecting radioDotB")
	}
	if !radioDotB.IsChecked() {
		t.Fatal("radioDotB should be checked after click")
	}

	if handled := radioCheckB.OnEvent(widgets.Event{Type: widgets.EventClick, Source: radioCheckB}); !handled {
		t.Fatal("radioCheckB click was not handled")
	}
	if radioCheckA.IsChecked() {
		t.Fatal("radioCheckA should be cleared after selecting radioCheckB")
	}
	if !radioCheckB.IsChecked() {
		t.Fatal("radioCheckB should be checked after click")
	}
}

func TestDemoHeaderSubtitleDoesNotOverlapHeaderButtons(t *testing.T) {
	_, window := loadDemoControllerForTest(t)
	if window.Root == nil {
		t.Fatal("window.Root is nil")
	}
	window.Root.SetBounds(widgets.Rect{W: 1380, H: 940})

	subtitle, ok := window.FindWidget("subtitleLabel").(*widgets.Label)
	if !ok {
		t.Fatalf("subtitleLabel type = %T, want *widgets.Label", window.FindWidget("subtitleLabel"))
	}
	togglePalette, ok := window.FindWidget("togglePaletteBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("togglePaletteBtn type = %T, want *widgets.Button", window.FindWidget("togglePaletteBtn"))
	}
	if subtitle.Bounds().X+subtitle.Bounds().W > togglePalette.Bounds().X {
		t.Fatalf(
			"subtitleLabel right edge (%d) overlaps togglePaletteBtn left edge (%d)",
			subtitle.Bounds().X+subtitle.Bounds().W,
			togglePalette.Bounds().X,
		)
	}
}

func TestDemoLanguageToggleSwitchesBetweenEnglishAndChinese(t *testing.T) {
	controller, window := loadDemoControllerForTest(t)

	runAllButton, ok := window.FindWidget("runAllBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("runAllBtn type = %T, want *widgets.Button", window.FindWidget("runAllBtn"))
	}
	langButton, ok := window.FindWidget("languageToggleBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("languageToggleBtn type = %T, want *widgets.Button", window.FindWidget("languageToggleBtn"))
	}

	if runAllButton.Text != "Run widget API tests" {
		t.Fatalf("runAllBtn default text = %q, want %q", runAllButton.Text, "Run widget API tests")
	}
	if langButton.Text != "中文" {
		t.Fatalf("languageToggleBtn default text = %q, want %q", langButton.Text, "中文")
	}

	controller.mustButton("languageToggleBtn").OnEvent(widgets.Event{
		Type:   widgets.EventClick,
		Source: langButton,
	})

	if runAllButton.Text != "测试所有控件函数" {
		t.Fatalf("runAllBtn Chinese text = %q, want %q", runAllButton.Text, "测试所有控件函数")
	}
	if langButton.Text != "English" {
		t.Fatalf("languageToggleBtn Chinese text = %q, want %q", langButton.Text, "English")
	}

	controller.mustButton("languageToggleBtn").OnEvent(widgets.Event{
		Type:   widgets.EventClick,
		Source: langButton,
	})

	if runAllButton.Text != "Run widget API tests" {
		t.Fatalf("runAllBtn toggled back text = %q, want %q", runAllButton.Text, "Run widget API tests")
	}
	if langButton.Text != "中文" {
		t.Fatalf("languageToggleBtn toggled back text = %q, want %q", langButton.Text, "中文")
	}
}

func TestDemoHeaderContainsControlModeToggleAndEmojiLabel(t *testing.T) {
	_, window := loadDemoControllerForTest(t)

	modeButton, ok := window.FindWidget("toggleControlModeBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("toggleControlModeBtn type = %T, want *widgets.Button", window.FindWidget("toggleControlModeBtn"))
	}
	if strings.TrimSpace(modeButton.Text) == "" {
		t.Fatal("toggleControlModeBtn text should not be empty")
	}

	emojiLabel, ok := window.FindWidget("emojiBurstLabel").(*widgets.Label)
	if !ok {
		t.Fatalf("emojiBurstLabel type = %T, want *widgets.Label", window.FindWidget("emojiBurstLabel"))
	}
	if !strings.Contains(emojiLabel.Text, "🚀") {
		t.Fatalf("emojiBurstLabel text = %q, want at least one rocket emoji", emojiLabel.Text)
	}
}

func TestDemoControlModeToggleUpdatesButtonTextAndStatus(t *testing.T) {
	controller, _ := loadDemoControllerForTest(t)

	modeButton := controller.mustButton("toggleControlModeBtn")
	before := modeButton.Text
	if strings.TrimSpace(before) == "" {
		t.Fatal("toggleControlModeBtn initial text should not be empty")
	}

	if handled := modeButton.OnEvent(widgets.Event{Type: widgets.EventClick, Source: modeButton}); !handled {
		t.Fatal("toggleControlModeBtn click was not handled")
	}

	after := controller.mustButton("toggleControlModeBtn").Text
	if strings.TrimSpace(after) == "" {
		t.Fatal("toggleControlModeBtn text after toggle should not be empty")
	}
	if before == after {
		t.Fatalf("toggleControlModeBtn text did not change after toggle: before=%q after=%q", before, after)
	}

	rawStatus, ok := controller.store.Get("demo.lastAction")
	if !ok {
		t.Fatal("demo.lastAction was not updated")
	}
	status := strings.ToLower(fmt.Sprint(rawStatus))
	if !strings.Contains(status, "mode") && !strings.Contains(fmt.Sprint(rawStatus), "模式") {
		t.Fatalf("demo.lastAction = %q, want mode-switch status", fmt.Sprint(rawStatus))
	}
}

func loadDemoControllerForTest(t *testing.T) (*demoController, *jsonui.Window) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	baseDir := filepath.Dir(currentFile)

	doc, store, err := loadDemoDocument(baseDir)
	if err != nil {
		t.Fatalf("loadDemoDocument returned error: %v", err)
	}
	window := doc.PrimaryWindow()
	if window == nil {
		t.Fatal("PrimaryWindow() returned nil")
	}

	return newDemoController(baseDir, store, doc, window, widgets.ModeCustom), window
}

//go:build windows

package main

import (
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

	iconTopBtn, ok := window.FindWidget("iconTopBtn").(*widgets.Button)
	if !ok {
		t.Fatalf("iconTopBtn type = %T, want *widgets.Button", window.FindWidget("iconTopBtn"))
	}
	if got := iconTopBtn.Bounds().H; got <= 56 {
		t.Fatalf("iconTopBtn.Bounds().H = %d, want > 56 for top-icon layout", got)
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

	return newDemoController(baseDir, store, doc, window), window
}

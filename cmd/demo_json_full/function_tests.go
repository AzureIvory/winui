//go:build windows

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/sysapi"
	"github.com/AzureIvory/winui/widgets"
)

type methodInvoker func() (string, error)

type panelAPI interface {
	widgets.Widget
	Add(widgets.Widget)
	AddAll(...widgets.Widget)
	Remove(string)
	Children() []widgets.Widget
	SetLayout(widgets.Layout)
	Layout() widgets.Layout
	SetStyle(widgets.PanelStyle)
	SetOnClick(func())
	OnEvent(widgets.Event) bool
}

func (c *demoController) runFunctionTests() string {
	lines := []string{
		"=== Widget API Check ===",
		"",
	}

	passes := 0
	failures := 0
	for _, suite := range c.methodSuites() {
		lines = append(lines, "## "+suite.name)
		suitePass, suiteFail := c.runMethodSuite(&lines, suite.name, suite.target, suite.invokers)
		passes += suitePass
		failures += suiteFail
		lines = append(lines, "")
	}

	summary := fmt.Sprintf("PASS=%d FAIL=%d", passes, failures)
	lines = append(lines, summary)
	report := strings.Join(lines, "\n")
	reportPath, err := c.writeFunctionReport(report)
	if err != nil {
		c.setReportSummary(summary + " | save failed")
		c.setReportPath(err.Error())
		c.store.Set("demo.report", summary)
		c.setStatus("API check complete, but saving report failed: " + err.Error())
		return report
	}
	uiReportPath := reportPath
	if rel, relErr := filepath.Rel(c.baseDir, reportPath); relErr == nil && rel != "" {
		uiReportPath = rel
	}
	c.setReportSummary(summary)
	c.setReportPath(uiReportPath)
	c.store.Set("demo.report", summary)
	c.setStatus("API check complete: " + summary)
	return report
}

func (c *demoController) writeFunctionReport(report string) (string, error) {
	outputDir := filepath.Join(c.baseDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	reportPath := filepath.Join(outputDir, "latest-api-check.txt")
	if err := os.WriteFile(reportPath, []byte(report), 0o600); err != nil {
		return "", err
	}
	return reportPath, nil
}

type methodSuite struct {
	name     string
	target   any
	invokers map[string]methodInvoker
}

func (c *demoController) runMethodSuite(lines *[]string, suiteName string, target any, invokers map[string]methodInvoker) (int, int) {
	methods := exportedMethods(target)
	passCount := 0
	failCount := 0
	for _, method := range methods {
		invoker := invokers[method]
		if invoker == nil {
			failCount++
			*lines = append(*lines, fmt.Sprintf("[FAIL] %s.%s - missing invoker", suiteName, method))
			continue
		}
		if c.recordMethod(lines, suiteName+"."+method, invoker) {
			passCount++
		} else {
			failCount++
		}
	}
	return passCount, failCount
}

func (c *demoController) recordMethod(lines *[]string, name string, invoker methodInvoker) (ok bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			ok = false
			*lines = append(*lines, fmt.Sprintf("[FAIL] %s - panic: %v", name, recovered))
		}
	}()

	detail, err := invoker()
	if err != nil {
		*lines = append(*lines, fmt.Sprintf("[FAIL] %s - %v", name, err))
		return false
	}
	if strings.TrimSpace(detail) == "" {
		detail = "ok"
	}
	*lines = append(*lines, fmt.Sprintf("[PASS] %s - %s", name, detail))
	return true
}

func exportedMethods(target any) []string {
	typ := reflect.TypeOf(target)
	if typ == nil {
		return nil
	}
	names := make([]string, 0, typ.NumMethod())
	for index := 0; index < typ.NumMethod(); index++ {
		method := typ.Method(index)
		if method.PkgPath != "" {
			continue
		}
		names = append(names, method.Name)
	}
	sort.Strings(names)
	return names
}

func mergeInvokers(parts ...map[string]methodInvoker) map[string]methodInvoker {
	out := map[string]methodInvoker{}
	for _, part := range parts {
		for key, value := range part {
			out[key] = value
		}
	}
	return out
}

func widgetInvokers(widget widgets.Widget, rect widgets.Rect) map[string]methodInvoker {
	return map[string]methodInvoker{
		"ID": func() (string, error) {
			if widget.ID() == "" {
				return "", fmt.Errorf("empty widget id")
			}
			return widget.ID(), nil
		},
		"Bounds": func() (string, error) {
			bounds := widget.Bounds()
			return fmt.Sprintf("(%d,%d,%d,%d)", bounds.X, bounds.Y, bounds.W, bounds.H), nil
		},
		"SetBounds": func() (string, error) {
			original := widget.Bounds()
			defer widget.SetBounds(original)
			widget.SetBounds(rect)
			if widget.Bounds() != rect {
				return "", fmt.Errorf("bounds=%v want %v", widget.Bounds(), rect)
			}
			return fmt.Sprintf("-> (%d,%d,%d,%d)", rect.X, rect.Y, rect.W, rect.H), nil
		},
		"Visible": func() (string, error) {
			return fmt.Sprintf("%v", widget.Visible()), nil
		},
		"SetVisible": func() (string, error) {
			original := widget.Visible()
			defer widget.SetVisible(original)
			widget.SetVisible(!original)
			if widget.Visible() != !original {
				return "", fmt.Errorf("visible=%v want %v", widget.Visible(), !original)
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"Enabled": func() (string, error) {
			return fmt.Sprintf("%v", widget.Enabled()), nil
		},
		"SetEnabled": func() (string, error) {
			original := widget.Enabled()
			defer widget.SetEnabled(original)
			widget.SetEnabled(!original)
			if widget.Enabled() != !original {
				return "", fmt.Errorf("enabled=%v want %v", widget.Enabled(), !original)
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"LayoutData": func() (string, error) {
			return fmt.Sprintf("%T", widget.LayoutData()), nil
		},
		"SetLayoutData": func() (string, error) {
			original := widget.LayoutData()
			defer widget.SetLayoutData(original)
			const marker = "demo-layout-marker"
			widget.SetLayoutData(marker)
			if widget.LayoutData() != marker {
				return "", fmt.Errorf("layoutData=%v want %v", widget.LayoutData(), marker)
			}
			return marker, nil
		},
		"HitTest": func() (string, error) {
			original := widget.Bounds()
			defer widget.SetBounds(original)
			widget.SetBounds(rect)
			hit := widget.HitTest(rect.X+1, rect.Y+1)
			if !hit {
				return "", fmt.Errorf("hit test returned false")
			}
			return "inside=true", nil
		},
		"OnEvent": func() (string, error) {
			handled := widget.OnEvent(widgets.Event{Type: widgets.EventPaint, Source: widget})
			return fmt.Sprintf("handled=%v", handled), nil
		},
		"Paint": func() (string, error) {
			widget.Paint(nil)
			return "paint(nil)", nil
		},
	}
}

func panelInvokers(panel panelAPI) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Add": func() (string, error) {
			temp := widgets.NewLabel("panel-add-temp", "temp")
			panel.Add(temp)
			defer panel.Remove(temp.ID())
			if len(panel.Children()) == 0 {
				return "", fmt.Errorf("child was not added")
			}
			return fmt.Sprintf("children=%d", len(panel.Children())), nil
		},
		"AddAll": func() (string, error) {
			tempA := widgets.NewLabel("panel-addall-a", "a")
			tempB := widgets.NewLabel("panel-addall-b", "b")
			panel.AddAll(tempA, tempB)
			defer panel.Remove(tempA.ID())
			defer panel.Remove(tempB.ID())
			return fmt.Sprintf("children=%d", len(panel.Children())), nil
		},
		"Remove": func() (string, error) {
			temp := widgets.NewLabel("panel-remove-temp", "temp")
			panel.Add(temp)
			panel.Remove(temp.ID())
			for _, child := range panel.Children() {
				if child != nil && child.ID() == temp.ID() {
					return "", fmt.Errorf("child %q still present", temp.ID())
				}
			}
			return temp.ID(), nil
		},
		"Children": func() (string, error) {
			return fmt.Sprintf("count=%d", len(panel.Children())), nil
		},
		"SetLayout": func() (string, error) {
			original := panel.Layout()
			defer panel.SetLayout(original)
			panel.SetLayout(widgets.ColumnLayout{Gap: 6})
			if _, ok := panel.Layout().(widgets.ColumnLayout); !ok {
				return "", fmt.Errorf("layout=%T want ColumnLayout", panel.Layout())
			}
			return "ColumnLayout", nil
		},
		"Layout": func() (string, error) {
			return fmt.Sprintf("%T", panel.Layout()), nil
		},
		"SetStyle": func() (string, error) {
			original := currentPanelStyle(panel)
			defer panel.SetStyle(original)
			style := widgets.PanelStyle{
				Background:   core.RGB(250, 250, 250),
				BorderColor:  core.RGB(180, 184, 192),
				BorderWidth:  1,
				CornerRadius: 10,
			}
			panel.SetStyle(style)
			if currentPanelStyle(panel) != style {
				return "", fmt.Errorf("style did not apply")
			}
			return fmt.Sprintf("bg=%#08x", style.Background), nil
		},
		"SetOnClick": func() (string, error) {
			original := currentPanelOnClick(panel)
			restoreDismiss := func() {}
			if modal, ok := panel.(*widgets.Modal); ok {
				previous := modal.DismissOnBackdrop()
				modal.SetDismissOnBackdrop(false)
				restoreDismiss = func() { modal.SetDismissOnBackdrop(previous) }
			}
			defer restoreDismiss()
			defer panel.SetOnClick(original)
			clicked := false
			panel.SetOnClick(func() {
				clicked = true
			})
			panel.OnEvent(widgets.Event{Type: widgets.EventClick, Source: panel})
			if !clicked {
				return "", fmt.Errorf("callback was not invoked")
			}
			return "callback invoked", nil
		},
	}
}

func currentPanelStyle(panel panelAPI) widgets.PanelStyle {
	switch typed := panel.(type) {
	case *widgets.Panel:
		return typed.Style
	case *widgets.Modal:
		return typed.Style
	case *widgets.FilePicker:
		return typed.Style
	default:
		return widgets.PanelStyle{}
	}
}

func currentPanelOnClick(panel panelAPI) func() {
	switch typed := panel.(type) {
	case *widgets.Panel:
		return typed.OnClick
	case *widgets.Modal:
		return typed.OnClick
	case *widgets.FilePicker:
		return typed.OnClick
	default:
		return nil
	}
}

func loadIconFromAsset(dir string, name string) (*core.Icon, error) {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	return core.LoadIconFromICO(data, 16)
}

func loadPNGAsset(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func buildBitmap(fill color.RGBA) (*core.Bitmap, error) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, fill)
		}
	}
	return core.BitmapFromRGBA(img)
}

func buildFrames() ([]core.AnimatedFrame, error) {
	first, err := buildBitmap(color.RGBA{R: 47, G: 109, B: 214, A: 255})
	if err != nil {
		return nil, err
	}
	second, err := buildBitmap(color.RGBA{R: 107, G: 114, B: 128, A: 255})
	if err != nil {
		_ = first.Close()
		return nil, err
	}
	return []core.AnimatedFrame{
		{Bitmap: first, Width: 16, Height: 16, Delay: 80 * time.Millisecond},
		{Bitmap: second, Width: 16, Height: 16, Delay: 80 * time.Millisecond},
	}, nil
}

func closeFrames(frames []core.AnimatedFrame) {
	for _, frame := range frames {
		if frame.Bitmap != nil {
			_ = frame.Bitmap.Close()
		}
	}
}

func (c *demoController) methodSuites() []methodSuite {
	panel := c.mustPanel("interactivePanel")
	modal := c.mustModal("helpModal")
	label := c.mustLabel("panelInfo")
	button := c.mustButton("saveBtn")
	edit := c.mustEdit("notesBox")
	check := c.mustCheckBox("checkStyleBox")
	radio := c.mustRadio("radioCheckA")
	combo := c.mustCombo("citySelect")
	list := c.mustListBox("cityList")
	picker := c.mustFilePicker("openFile")
	imageWidget := c.mustImage("previewImage")
	animated := c.mustAnimated("spinnerImage")
	progress := c.mustProgress("uploadProgress")
	scroll := c.mustScroll("miniScroll")

	return []methodSuite{
		{
			name:   "Panel",
			target: panel,
			invokers: mergeInvokers(
				widgetInvokers(panel, widgets.Rect{X: 24, Y: 24, W: 260, H: 96}),
				panelInvokers(panel),
				map[string]methodInvoker{
					"OnEvent": func() (string, error) {
						clicked := false
						original := panel.OnClick
						defer panel.SetOnClick(original)
						panel.SetOnClick(func() { clicked = true })
						handled := panel.OnEvent(widgets.Event{Type: widgets.EventClick, Source: panel})
						if !handled || !clicked {
							return "", fmt.Errorf("handled=%v clicked=%v", handled, clicked)
						}
						return "click dispatched", nil
					},
				},
			),
		},
		{
			name:   "Modal",
			target: modal,
			invokers: mergeInvokers(
				widgetInvokers(modal, widgets.Rect{X: 0, Y: 0, W: 440, H: 320}),
				panelInvokers(modal),
				c.modalInvokers(modal),
			),
		},
		{
			name:   "Label",
			target: label,
			invokers: mergeInvokers(
				widgetInvokers(label, widgets.Rect{X: 20, Y: 20, W: 260, H: 72}),
				c.labelInvokers(label),
			),
		},
		{
			name:   "Button",
			target: button,
			invokers: mergeInvokers(
				widgetInvokers(button, widgets.Rect{X: 20, Y: 20, W: 132, H: 40}),
				c.buttonInvokers(button),
			),
		},
		{
			name:   "EditBox",
			target: edit,
			invokers: mergeInvokers(
				widgetInvokers(edit, widgets.Rect{X: 20, Y: 20, W: 320, H: 110}),
				c.editInvokers(edit),
			),
		},
		{
			name:   "CheckBox",
			target: check,
			invokers: mergeInvokers(
				widgetInvokers(check, widgets.Rect{X: 20, Y: 20, W: 220, H: 28}),
				c.checkBoxInvokers(check),
			),
		},
		{
			name:   "RadioButton",
			target: radio,
			invokers: mergeInvokers(
				widgetInvokers(radio, widgets.Rect{X: 20, Y: 20, W: 220, H: 28}),
				c.radioInvokers(radio),
			),
		},
		{
			name:   "ComboBox",
			target: combo,
			invokers: mergeInvokers(
				widgetInvokers(combo, widgets.Rect{X: 20, Y: 20, W: 220, H: 36}),
				c.comboInvokers(combo),
			),
		},
		{
			name:   "ListBox",
			target: list,
			invokers: mergeInvokers(
				widgetInvokers(list, widgets.Rect{X: 20, Y: 20, W: 220, H: 160}),
				c.listInvokers(list),
			),
		},
		{
			name:   "FilePicker",
			target: picker,
			invokers: mergeInvokers(
				widgetInvokers(picker, widgets.Rect{X: 20, Y: 20, W: 360, H: 36}),
				panelInvokers(picker),
				c.filePickerInvokers(picker),
			),
		},
		{
			name:   "Image",
			target: imageWidget,
			invokers: mergeInvokers(
				widgetInvokers(imageWidget, widgets.Rect{X: 20, Y: 20, W: 160, H: 120}),
				c.imageInvokers(imageWidget),
			),
		},
		{
			name:   "AnimatedImage",
			target: animated,
			invokers: mergeInvokers(
				widgetInvokers(animated, widgets.Rect{X: 20, Y: 20, W: 96, H: 96}),
				c.animatedInvokers(animated),
			),
		},
		{
			name:   "ProgressBar",
			target: progress,
			invokers: mergeInvokers(
				widgetInvokers(progress, widgets.Rect{X: 20, Y: 20, W: 240, H: 24}),
				c.progressInvokers(progress),
			),
		},
		{
			name:   "ScrollView",
			target: scroll,
			invokers: mergeInvokers(
				widgetInvokers(scroll, widgets.Rect{X: 20, Y: 20, W: 280, H: 80}),
				c.scrollInvokers(scroll),
			),
		},
		{
			name:     "Document",
			target:   c.doc,
			invokers: c.documentInvokers(),
		},
		{
			name:     "Window",
			target:   c.window,
			invokers: c.windowInvokers(),
		},
	}
}

func (c *demoController) modalInvokers(modal *widgets.Modal) map[string]methodInvoker {
	return map[string]methodInvoker{
		"BackdropColor": func() (string, error) {
			return fmt.Sprintf("%#08x", modal.BackdropColor()), nil
		},
		"BackdropOpacity": func() (string, error) {
			return fmt.Sprintf("%d", modal.BackdropOpacity()), nil
		},
		"BlurRadiusDP": func() (string, error) {
			return fmt.Sprintf("%d", modal.BlurRadiusDP()), nil
		},
		"DismissOnBackdrop": func() (string, error) {
			return fmt.Sprintf("%v", modal.DismissOnBackdrop()), nil
		},
		"SetBackdropColor": func() (string, error) {
			original := modal.BackdropColor()
			defer modal.SetBackdropColor(original)
			modal.SetBackdropColor(core.RGB(51, 65, 85))
			if modal.BackdropColor() != core.RGB(51, 65, 85) {
				return "", fmt.Errorf("backdrop color did not apply")
			}
			return "updated", nil
		},
		"SetBackdropOpacity": func() (string, error) {
			original := modal.BackdropOpacity()
			defer modal.SetBackdropOpacity(original)
			modal.SetBackdropOpacity(72)
			if modal.BackdropOpacity() != 72 {
				return "", fmt.Errorf("opacity=%d", modal.BackdropOpacity())
			}
			return "72", nil
		},
		"SetBlurRadiusDP": func() (string, error) {
			original := modal.BlurRadiusDP()
			defer modal.SetBlurRadiusDP(original)
			modal.SetBlurRadiusDP(4)
			if modal.BlurRadiusDP() != 4 {
				return "", fmt.Errorf("blur=%d", modal.BlurRadiusDP())
			}
			return "4", nil
		},
		"SetDismissOnBackdrop": func() (string, error) {
			original := modal.DismissOnBackdrop()
			defer modal.SetDismissOnBackdrop(original)
			modal.SetDismissOnBackdrop(!original)
			if modal.DismissOnBackdrop() != !original {
				return "", fmt.Errorf("dismissOnBackdrop=%v", modal.DismissOnBackdrop())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"SetOnDismiss": func() (string, error) {
			triggered := false
			originalDismiss := modal.DismissOnBackdrop()
			defer modal.SetDismissOnBackdrop(originalDismiss)
			modal.SetDismissOnBackdrop(true)
			modal.SetOnDismiss(func() { triggered = true })
			modal.OnEvent(widgets.Event{Type: widgets.EventClick, Source: modal})
			if !triggered {
				return "", fmt.Errorf("dismiss callback not triggered")
			}
			return "callback invoked", nil
		},
		"Close": func() (string, error) {
			return "closed", modal.Close()
		},
	}
}

func (c *demoController) labelInvokers(label *widgets.Label) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Multiline": func() (string, error) { return fmt.Sprintf("%v", label.Multiline()), nil },
		"SetMultiline": func() (string, error) {
			original := label.Multiline()
			defer label.SetMultiline(original)
			label.SetMultiline(!original)
			if label.Multiline() != !original {
				return "", fmt.Errorf("multiline=%v", label.Multiline())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"WordWrap": func() (string, error) { return fmt.Sprintf("%v", label.WordWrap()), nil },
		"SetWordWrap": func() (string, error) {
			original := label.WordWrap()
			defer label.SetWordWrap(original)
			label.SetWordWrap(!original)
			if label.WordWrap() != !original {
				return "", fmt.Errorf("wordWrap=%v", label.WordWrap())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"SetText": func() (string, error) {
			original := label.Text
			defer label.SetText(original)
			label.SetText("Label.SetText")
			if label.Text != "Label.SetText" {
				return "", fmt.Errorf("text=%q", label.Text)
			}
			return label.Text, nil
		},
		"SetStyle": func() (string, error) {
			original := label.Style
			defer label.SetStyle(original)
			style := makeTextStyle("Microsoft YaHei UI", 14, 700, core.RGB(30, 41, 59))
			label.SetStyle(style)
			if label.Style.Color != style.Color {
				return "", fmt.Errorf("style color=%#08x", label.Style.Color)
			}
			return "updated", nil
		},
	}
}

func (c *demoController) buttonInvokers(button *widgets.Button) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Kind": func() (string, error) { return fmt.Sprintf("%d", button.Kind()), nil },
		"SetKind": func() (string, error) {
			original := button.Kind()
			defer button.SetKind(original)
			button.SetKind(widgets.BtnTop)
			if button.Kind() != widgets.BtnTop {
				return "", fmt.Errorf("kind=%d", button.Kind())
			}
			return "BtnTop", nil
		},
		"SetText": func() (string, error) {
			original := button.Text
			defer button.SetText(original)
			button.SetText("Button.SetText")
			if button.Text != "Button.SetText" {
				return "", fmt.Errorf("text=%q", button.Text)
			}
			return button.Text, nil
		},
		"SetIcon": func() (string, error) {
			original := button.Icon
			icon, err := loadIconFromAsset(c.assetsDir, "palette.ico")
			if err != nil {
				return "", err
			}
			defer icon.Close()
			defer button.SetIcon(original)
			button.SetIcon(icon)
			if button.Icon != icon {
				return "", fmt.Errorf("icon pointer did not update")
			}
			return "palette.ico", nil
		},
		"SetOnClick": func() (string, error) {
			original := button.OnClick
			defer button.SetOnClick(original)
			clicked := false
			button.SetOnClick(func() { clicked = true })
			button.OnEvent(widgets.Event{Type: widgets.EventClick, Source: button})
			if !clicked {
				return "", fmt.Errorf("click callback not triggered")
			}
			return "callback invoked", nil
		},
		"SetStyle": func() (string, error) {
			original := button.Style
			defer button.SetStyle(original)
			style := makeButtonStyle(core.RGB(248, 249, 250), core.RGB(203, 213, 225), core.RGB(239, 246, 255), core.RGB(59, 130, 246), core.RGB(15, 23, 42), core.RGB(255, 255, 255))
			button.SetStyle(style)
			if button.Style.Background != style.Background {
				return "", fmt.Errorf("background=%#08x", button.Style.Background)
			}
			return "updated", nil
		},
		"Close": func() (string, error) { return "closed", button.Close() },
	}
}

func (c *demoController) editInvokers(edit *widgets.EditBox) map[string]methodInvoker {
	return map[string]methodInvoker{
		"TextValue": func() (string, error) { return edit.TextValue(), nil },
		"SetText": func() (string, error) {
			original := edit.TextValue()
			defer edit.SetText(original)
			edit.SetText("first\nsecond\nthird")
			if edit.TextValue() != "first\nsecond\nthird" {
				return "", fmt.Errorf("text=%q", edit.TextValue())
			}
			return "3 lines", nil
		},
		"SetPlaceholder": func() (string, error) {
			original := edit.Placeholder
			defer edit.SetPlaceholder(original)
			edit.SetPlaceholder("placeholder")
			if edit.Placeholder != "placeholder" {
				return "", fmt.Errorf("placeholder=%q", edit.Placeholder)
			}
			return edit.Placeholder, nil
		},
		"SetReadOnly": func() (string, error) {
			original := edit.ReadOnly
			defer edit.SetReadOnly(original)
			edit.SetReadOnly(!original)
			if edit.ReadOnly != !original {
				return "", fmt.Errorf("readOnly=%v", edit.ReadOnly)
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"Password": func() (string, error) { return fmt.Sprintf("%v", edit.Password()), nil },
		"SetPassword": func() (string, error) {
			original := edit.Password()
			defer edit.SetPassword(original)
			edit.SetPassword(!original)
			if edit.Password() != !original {
				return "", fmt.Errorf("password=%v", edit.Password())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"Multiline": func() (string, error) { return fmt.Sprintf("%v", edit.Multiline()), nil },
		"SetMultiline": func() (string, error) {
			original := edit.Multiline()
			defer edit.SetMultiline(original)
			edit.SetMultiline(!original)
			if edit.Multiline() != !original {
				return "", fmt.Errorf("multiline=%v", edit.Multiline())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"WordWrap": func() (string, error) { return fmt.Sprintf("%v", edit.WordWrap()), nil },
		"SetWordWrap": func() (string, error) {
			original := edit.WordWrap()
			defer edit.SetWordWrap(original)
			edit.SetWordWrap(!original)
			if edit.WordWrap() != !original {
				return "", fmt.Errorf("wordWrap=%v", edit.WordWrap())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"VerticalScroll": func() (string, error) { return fmt.Sprintf("%v", edit.VerticalScroll()), nil },
		"SetVerticalScroll": func() (string, error) {
			original := edit.VerticalScroll()
			defer edit.SetVerticalScroll(original)
			edit.SetVerticalScroll(!original)
			if edit.VerticalScroll() != !original {
				return "", fmt.Errorf("verticalScroll=%v", edit.VerticalScroll())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"HorizontalScroll": func() (string, error) { return fmt.Sprintf("%v", edit.HorizontalScroll()), nil },
		"SetHorizontalScroll": func() (string, error) {
			original := edit.HorizontalScroll()
			defer edit.SetHorizontalScroll(original)
			edit.SetHorizontalScroll(!original)
			if edit.HorizontalScroll() != !original {
				return "", fmt.Errorf("horizontalScroll=%v", edit.HorizontalScroll())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"AcceptReturn": func() (string, error) { return fmt.Sprintf("%v", edit.AcceptReturn()), nil },
		"SetAcceptReturn": func() (string, error) {
			original := edit.AcceptReturn()
			defer edit.SetAcceptReturn(original)
			edit.SetAcceptReturn(!original)
			if edit.AcceptReturn() != !original {
				return "", fmt.Errorf("acceptReturn=%v", edit.AcceptReturn())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"ScrollToCaret": func() (string, error) {
			edit.ScrollToCaret()
			return "called", nil
		},
		"LineCount": func() (string, error) {
			originalText := edit.TextValue()
			originalMultiline := edit.Multiline()
			defer edit.SetText(originalText)
			defer edit.SetMultiline(originalMultiline)
			edit.SetMultiline(true)
			edit.SetText("a\nb\nc")
			if edit.LineCount() < 3 {
				return "", fmt.Errorf("lineCount=%d", edit.LineCount())
			}
			return fmt.Sprintf("%d", edit.LineCount()), nil
		},
		"SetStyle": func() (string, error) {
			original := edit.Style
			defer edit.SetStyle(original)
			style := makeEditStyle(core.RGB(250, 251, 252), core.RGB(203, 213, 225), core.RGB(148, 163, 184), core.RGB(37, 99, 235), core.RGB(31, 41, 55), core.RGB(148, 163, 184), core.RGB(37, 99, 235))
			edit.SetStyle(style)
			if edit.Style.Background != style.Background {
				return "", fmt.Errorf("background=%#08x", edit.Style.Background)
			}
			return "updated", nil
		},
		"SetOnChange": func() (string, error) {
			original := edit.OnChange
			defer edit.SetOnChange(original)
			edit.SetOnChange(func(string) {})
			return "callback assigned", nil
		},
		"SetOnSubmit": func() (string, error) {
			original := edit.OnSubmit
			defer edit.SetOnSubmit(original)
			edit.SetOnSubmit(func(string) {})
			return "callback assigned", nil
		},
		"Close": func() (string, error) { return "closed", edit.Close() },
	}
}

func (c *demoController) checkBoxInvokers(check *widgets.CheckBox) map[string]methodInvoker {
	return map[string]methodInvoker{
		"IsChecked": func() (string, error) { return fmt.Sprintf("%v", check.IsChecked()), nil },
		"SetChecked": func() (string, error) {
			original := check.IsChecked()
			defer check.SetChecked(original)
			check.SetChecked(!original)
			if check.IsChecked() != !original {
				return "", fmt.Errorf("checked=%v", check.IsChecked())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"SetText": func() (string, error) {
			original := check.Text
			defer check.SetText(original)
			check.SetText("CheckBox.SetText")
			if check.Text != "CheckBox.SetText" {
				return "", fmt.Errorf("text=%q", check.Text)
			}
			return check.Text, nil
		},
		"SetStyle": func() (string, error) {
			original := check.Style
			defer check.SetStyle(original)
			style := makeChoiceStyle(core.RGB(255, 255, 255), core.RGB(203, 213, 225), core.RGB(96, 165, 250), core.RGB(59, 130, 246), core.RGB(59, 130, 246), widgets.ChoiceIndicatorCheck, 6)
			check.SetStyle(style)
			if check.Style.IndicatorColor != style.IndicatorColor {
				return "", fmt.Errorf("indicator=%#08x", check.Style.IndicatorColor)
			}
			return "updated", nil
		},
		"SetOnChange": func() (string, error) {
			original := check.OnChange
			defer check.SetOnChange(original)
			triggered := false
			originalChecked := check.IsChecked()
			defer check.SetChecked(originalChecked)
			check.SetOnChange(func(bool) { triggered = true })
			check.OnEvent(widgets.Event{Type: widgets.EventClick, Source: check})
			if !triggered {
				return "", fmt.Errorf("change callback not triggered")
			}
			return "callback invoked", nil
		},
		"Close": func() (string, error) { return "closed", check.Close() },
	}
}

func (c *demoController) radioInvokers(radio *widgets.RadioButton) map[string]methodInvoker {
	return map[string]methodInvoker{
		"IsChecked": func() (string, error) { return fmt.Sprintf("%v", radio.IsChecked()), nil },
		"SetChecked": func() (string, error) {
			original := radio.IsChecked()
			defer radio.SetChecked(original)
			radio.SetChecked(true)
			if !radio.IsChecked() {
				return "", fmt.Errorf("radio not checked")
			}
			return "checked", nil
		},
		"SetGroup": func() (string, error) {
			original := radio.Group
			defer radio.SetGroup(original)
			radio.SetGroup("function-tests")
			if radio.Group != "function-tests" {
				return "", fmt.Errorf("group=%q", radio.Group)
			}
			return radio.Group, nil
		},
		"SetText": func() (string, error) {
			original := radio.Text
			defer radio.SetText(original)
			radio.SetText("RadioButton.SetText")
			if radio.Text != "RadioButton.SetText" {
				return "", fmt.Errorf("text=%q", radio.Text)
			}
			return radio.Text, nil
		},
		"SetStyle": func() (string, error) {
			original := radio.Style
			defer radio.SetStyle(original)
			style := makeChoiceStyle(core.RGB(255, 255, 255), core.RGB(203, 213, 225), core.RGB(96, 165, 250), core.RGB(59, 130, 246), core.RGB(59, 130, 246), widgets.ChoiceIndicatorCheck, 9)
			radio.SetStyle(style)
			if radio.Style.IndicatorColor != style.IndicatorColor {
				return "", fmt.Errorf("indicator=%#08x", radio.Style.IndicatorColor)
			}
			return "updated", nil
		},
		"SetOnChange": func() (string, error) {
			original := radio.OnChange
			defer radio.SetOnChange(original)
			triggered := false
			radio.SetChecked(false)
			radio.SetOnChange(func(bool) { triggered = true })
			radio.OnEvent(widgets.Event{Type: widgets.EventClick, Source: radio})
			if !triggered {
				return "", fmt.Errorf("change callback not triggered")
			}
			return "callback invoked", nil
		},
		"Close": func() (string, error) { return "closed", radio.Close() },
	}
}

func (c *demoController) comboInvokers(combo *widgets.ComboBox) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Items": func() (string, error) { return fmt.Sprintf("%d", len(combo.Items())), nil },
		"SetItems": func() (string, error) {
			original := combo.Items()
			defer combo.SetItems(original)
			items := []widgets.ListItem{{Value: "a", Text: "A"}, {Value: "b", Text: "B"}}
			combo.SetItems(items)
			if len(combo.Items()) != 2 {
				return "", fmt.Errorf("items=%d", len(combo.Items()))
			}
			return "2 items", nil
		},
		"SetSelected": func() (string, error) {
			original := combo.SelectedIndex()
			defer combo.SetSelected(original)
			combo.SetSelected(1)
			if combo.SelectedIndex() != 1 {
				return "", fmt.Errorf("selected=%d", combo.SelectedIndex())
			}
			return "index=1", nil
		},
		"SelectedIndex": func() (string, error) { return fmt.Sprintf("%d", combo.SelectedIndex()), nil },
		"SelectedItem": func() (string, error) {
			item, ok := combo.SelectedItem()
			if !ok {
				return "none", nil
			}
			return item.Value, nil
		},
		"SetPlaceholder": func() (string, error) {
			original := combo.Placeholder
			defer combo.SetPlaceholder(original)
			combo.SetPlaceholder("ComboBox.SetPlaceholder")
			if combo.Placeholder != "ComboBox.SetPlaceholder" {
				return "", fmt.Errorf("placeholder=%q", combo.Placeholder)
			}
			return combo.Placeholder, nil
		},
		"SetStyle": func() (string, error) {
			original := combo.Style
			defer combo.SetStyle(original)
			style := makeComboStyle(core.RGB(250, 251, 252), core.RGB(203, 213, 225), core.RGB(148, 163, 184), core.RGB(59, 130, 246), core.RGB(59, 130, 246))
			combo.SetStyle(style)
			if combo.Style.BorderColor != style.BorderColor {
				return "", fmt.Errorf("border=%#08x", combo.Style.BorderColor)
			}
			return "updated", nil
		},
		"SetOnChange": func() (string, error) {
			original := combo.OnChange
			defer combo.SetOnChange(original)
			combo.SetOnChange(func(int, widgets.ListItem) {})
			return "callback assigned", nil
		},
		"Close": func() (string, error) { return "closed", combo.Close() },
	}
}

func (c *demoController) listInvokers(list *widgets.ListBox) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Items": func() (string, error) { return fmt.Sprintf("%d", len(list.Items())), nil },
		"SetItems": func() (string, error) {
			original := list.Items()
			defer list.SetItems(original)
			items := []widgets.ListItem{{Value: "a", Text: "A"}, {Value: "b", Text: "B"}, {Value: "c", Text: "C"}}
			list.SetItems(items)
			if len(list.Items()) != len(items) {
				return "", fmt.Errorf("items=%d", len(list.Items()))
			}
			return "3 items", nil
		},
		"SetSelected": func() (string, error) {
			original := list.SelectedIndex()
			defer list.SetSelected(original)
			list.SetSelected(1)
			if list.SelectedIndex() != 1 {
				return "", fmt.Errorf("selected=%d", list.SelectedIndex())
			}
			return "index=1", nil
		},
		"ClearSelection": func() (string, error) {
			original := list.SelectedIndex()
			defer list.SetSelected(original)
			list.ClearSelection()
			if list.SelectedIndex() != -1 {
				return "", fmt.Errorf("selected=%d", list.SelectedIndex())
			}
			return "cleared", nil
		},
		"SelectedIndex": func() (string, error) { return fmt.Sprintf("%d", list.SelectedIndex()), nil },
		"SelectedItem": func() (string, error) {
			item, ok := list.SelectedItem()
			if !ok {
				return "none", nil
			}
			return item.Value, nil
		},
		"SetStyle": func() (string, error) {
			original := list.Style
			defer list.SetStyle(original)
			style := makeListStyle(core.RGB(250, 251, 252), core.RGB(203, 213, 225), core.RGB(148, 163, 184), core.RGB(59, 130, 246))
			list.SetStyle(style)
			if list.Style.BorderColor != style.BorderColor {
				return "", fmt.Errorf("border=%#08x", list.Style.BorderColor)
			}
			return "updated", nil
		},
		"SetOnChange": func() (string, error) {
			original := list.OnChange
			defer list.SetOnChange(original)
			list.SetOnChange(func(int, widgets.ListItem) {})
			return "callback assigned", nil
		},
		"SetOnActivate": func() (string, error) {
			original := list.OnActivate
			defer list.SetOnActivate(original)
			list.SetOnActivate(func(int, widgets.ListItem) {})
			return "callback assigned", nil
		},
		"SetOnRightClick": func() (string, error) {
			original := list.OnRightClick
			defer list.SetOnRightClick(original)
			list.SetOnRightClick(func(int, widgets.ListItem, core.Point) {})
			return "callback assigned", nil
		},
	}
}

func (c *demoController) filePickerInvokers(picker *widgets.FilePicker) map[string]methodInvoker {
	return map[string]methodInvoker{
		"DialogOptions": func() (string, error) {
			return string(picker.DialogOptions().Mode), nil
		},
		"SetDialogOptions": func() (string, error) {
			original := picker.DialogOptions()
			defer picker.SetDialogOptions(original)
			opts := sysapi.Options{
				Mode:             sysapi.DialogSave,
				Title:            "Save test",
				DefaultExtension: "txt",
			}
			picker.SetDialogOptions(opts)
			if picker.DialogOptions().Mode != sysapi.DialogSave {
				return "", fmt.Errorf("mode=%s", picker.DialogOptions().Mode)
			}
			return string(picker.DialogOptions().Mode), nil
		},
		"SetFieldStyle": func() (string, error) {
			picker.SetFieldStyle(makeEditStyle(core.RGB(250, 251, 252), core.RGB(203, 213, 225), core.RGB(148, 163, 184), core.RGB(59, 130, 246), core.RGB(31, 41, 55), core.RGB(148, 163, 184), core.RGB(59, 130, 246)))
			return "applied", nil
		},
		"SetButtonStyle": func() (string, error) {
			picker.SetButtonStyle(makeButtonStyle(core.RGB(248, 249, 250), core.RGB(203, 213, 225), core.RGB(239, 246, 255), core.RGB(59, 130, 246), core.RGB(15, 23, 42), core.RGB(255, 255, 255)))
			return "applied", nil
		},
		"SetSeparator": func() (string, error) {
			picker.SetSeparator(" | ")
			return "\" | \"", nil
		},
		"SetButtonText": func() (string, error) {
			picker.SetButtonText("FilePicker.SetButtonText")
			return "updated", nil
		},
		"SetPlaceholder": func() (string, error) {
			picker.SetPlaceholder("FilePicker.SetPlaceholder")
			return "updated", nil
		},
		"SetPaths": func() (string, error) {
			paths := []string{"README.md", "go.mod"}
			picker.SetPaths(paths)
			got := picker.Paths()
			if len(got) != len(paths) {
				return "", fmt.Errorf("paths=%v", got)
			}
			return strings.Join(got, ","), nil
		},
		"Paths": func() (string, error) {
			return strings.Join(picker.Paths(), ","), nil
		},
		"SetOnChange": func() (string, error) {
			picker.SetOnChange(func([]string) {})
			return "callback assigned", nil
		},
	}
}

func (c *demoController) imageInvokers(imageWidget *widgets.Image) map[string]methodInvoker {
	assetPath := filepath.Join(c.assetsDir, "preview.png")
	return map[string]methodInvoker{
		"SetScaleMode": func() (string, error) {
			imageWidget.SetScaleMode(widgets.ImageScaleCenter)
			return "ImageScaleCenter", nil
		},
		"SetOpacity": func() (string, error) {
			imageWidget.SetOpacity(180)
			return "180", nil
		},
		"SetBitmap": func() (string, error) {
			original := imageWidget.Bitmap()
			temp, err := buildBitmap(color.RGBA{R: 47, G: 109, B: 214, A: 255})
			if err != nil {
				return "", err
			}
			defer imageWidget.SetBitmap(original)
			defer temp.Close()
			imageWidget.SetBitmap(temp)
			if imageWidget.Bitmap() != temp {
				return "", fmt.Errorf("bitmap pointer did not update")
			}
			return "non-owned bitmap", nil
		},
		"SetBitmapOwned": func() (string, error) {
			temp, err := buildBitmap(color.RGBA{R: 107, G: 114, B: 128, A: 255})
			if err != nil {
				return "", err
			}
			imageWidget.SetBitmapOwned(temp)
			if imageWidget.Bitmap() != temp {
				return "", fmt.Errorf("owned bitmap did not update")
			}
			data, err := loadPNGAsset(assetPath)
			if err != nil {
				return "", err
			}
			if err := imageWidget.LoadBytes(data); err != nil {
				return "", err
			}
			return "owned bitmap", nil
		},
		"LoadBytes": func() (string, error) {
			data, err := loadPNGAsset(assetPath)
			if err != nil {
				return "", err
			}
			if err := imageWidget.LoadBytes(data); err != nil {
				return "", err
			}
			size := imageWidget.NaturalSize()
			if size.Width == 0 || size.Height == 0 {
				return "", fmt.Errorf("naturalSize=%v", size)
			}
			return fmt.Sprintf("%dx%d", size.Width, size.Height), nil
		},
		"NaturalSize": func() (string, error) {
			size := imageWidget.NaturalSize()
			return fmt.Sprintf("%dx%d", size.Width, size.Height), nil
		},
		"Bitmap": func() (string, error) {
			return fmt.Sprintf("%v", imageWidget.Bitmap() != nil), nil
		},
		"Close": func() (string, error) {
			if err := imageWidget.Close(); err != nil {
				return "", err
			}
			data, err := loadPNGAsset(assetPath)
			if err != nil {
				return "", err
			}
			if err := imageWidget.LoadBytes(data); err != nil {
				return "", err
			}
			return "closed and restored", nil
		},
	}
}

func (c *demoController) animatedInvokers(animated *widgets.AnimatedImage) map[string]methodInvoker {
	assetPath := filepath.Join(c.assetsDir, "spinner.gif")
	return map[string]methodInvoker{
		"SetScaleMode": func() (string, error) {
			animated.SetScaleMode(widgets.ImageScaleCenter)
			return "ImageScaleCenter", nil
		},
		"SetOpacity": func() (string, error) {
			animated.SetOpacity(190)
			return "190", nil
		},
		"SetPlaying": func() (string, error) {
			animated.SetPlaying(false)
			animated.SetPlaying(true)
			return "false -> true", nil
		},
		"LoadGIF": func() (string, error) {
			data, err := os.ReadFile(assetPath)
			if err != nil {
				return "", err
			}
			if err := animated.LoadGIF(data); err != nil {
				return "", err
			}
			size := animated.NaturalSize()
			return fmt.Sprintf("%dx%d", size.Width, size.Height), nil
		},
		"SetFrames": func() (string, error) {
			frames, err := buildFrames()
			if err != nil {
				return "", err
			}
			defer closeFrames(frames)
			animated.SetFrames(frames)
			if animated.NaturalSize().Width == 0 {
				return "", fmt.Errorf("natural size did not update")
			}
			data, err := os.ReadFile(assetPath)
			if err == nil {
				_ = animated.LoadGIF(data)
			}
			return fmt.Sprintf("frames=%d", len(frames)), nil
		},
		"SetFramesOwned": func() (string, error) {
			frames, err := buildFrames()
			if err != nil {
				return "", err
			}
			animated.SetFramesOwned(frames)
			data, err := os.ReadFile(assetPath)
			if err != nil {
				return "", err
			}
			if err := animated.LoadGIF(data); err != nil {
				return "", err
			}
			return fmt.Sprintf("frames=%d", len(frames)), nil
		},
		"NaturalSize": func() (string, error) {
			size := animated.NaturalSize()
			return fmt.Sprintf("%dx%d", size.Width, size.Height), nil
		},
		"CurrentFrame": func() (string, error) {
			return fmt.Sprintf("%d", animated.CurrentFrame()), nil
		},
		"Close": func() (string, error) {
			if err := animated.Close(); err != nil {
				return "", err
			}
			data, err := os.ReadFile(assetPath)
			if err != nil {
				return "", err
			}
			if err := animated.LoadGIF(data); err != nil {
				return "", err
			}
			animated.SetPlaying(true)
			return "closed and restored", nil
		},
	}
}

func (c *demoController) progressInvokers(progress *widgets.ProgressBar) map[string]methodInvoker {
	return map[string]methodInvoker{
		"Value": func() (string, error) { return fmt.Sprintf("%d", progress.Value()), nil },
		"SetValue": func() (string, error) {
			original := progress.Value()
			defer progress.SetValue(original)
			progress.SetValue(61)
			if progress.Value() != 61 {
				return "", fmt.Errorf("value=%d", progress.Value())
			}
			return "61", nil
		},
		"SetStyle": func() (string, error) {
			original := progress.Style
			defer progress.SetStyle(original)
			style := makeProgressStyle(core.RGB(223, 227, 232), core.RGB(59, 130, 246), core.RGB(30, 64, 175))
			progress.SetStyle(style)
			if progress.Style.FillColor != style.FillColor {
				return "", fmt.Errorf("fill=%#08x", progress.Style.FillColor)
			}
			return "updated", nil
		},
	}
}

func (c *demoController) scrollInvokers(scroll *widgets.ScrollView) map[string]methodInvoker {
	return map[string]methodInvoker{
		"SetStyle": func() (string, error) {
			original := scroll.Style
			defer scroll.SetStyle(original)
			style := widgets.PanelStyle{Background: core.RGB(248, 249, 250), BorderColor: core.RGB(203, 213, 225), BorderWidth: 1, CornerRadius: 12}
			scroll.SetStyle(style)
			if scroll.Style.Background != style.Background {
				return "", fmt.Errorf("background=%#08x", scroll.Style.Background)
			}
			return "updated", nil
		},
		"SetContent": func() (string, error) {
			original := scroll.Content()
			replacement := widgets.NewPanel("scroll-replacement")
			replacement.SetLayout(widgets.RowLayout{Gap: 8})
			replacement.Add(widgets.NewButton("scroll-replacement-btn", "Replacement", widgets.ModeCustom))
			scroll.SetContent(replacement)
			defer scroll.SetContent(original)
			if scroll.Content() != replacement {
				return "", fmt.Errorf("content did not update")
			}
			return "replacement panel", nil
		},
		"Content": func() (string, error) { return fmt.Sprintf("%T", scroll.Content()), nil },
		"SetScrollOffset": func() (string, error) {
			originalContent := scroll.Content()
			scroll.SetBounds(widgets.Rect{X: 0, Y: 0, W: 160, H: 40})
			scroll.SetScrollOffset(36, 0)
			x, y := scroll.ScrollOffset()
			defer scroll.SetContent(originalContent)
			if x == 0 && y == 0 {
				return "", fmt.Errorf("scroll offset unchanged")
			}
			return fmt.Sprintf("%d,%d", x, y), nil
		},
		"ScrollOffset": func() (string, error) {
			x, y := scroll.ScrollOffset()
			return fmt.Sprintf("%d,%d", x, y), nil
		},
		"ScrollBy": func() (string, error) {
			scroll.SetBounds(widgets.Rect{X: 0, Y: 0, W: 160, H: 40})
			scroll.ScrollBy(20, 0)
			x, y := scroll.ScrollOffset()
			return fmt.Sprintf("%d,%d", x, y), nil
		},
		"ScrollTo": func() (string, error) {
			scroll.SetBounds(widgets.Rect{X: 0, Y: 0, W: 160, H: 40})
			scroll.ScrollTo(10, 0)
			x, y := scroll.ScrollOffset()
			return fmt.Sprintf("%d,%d", x, y), nil
		},
		"SetWheelStep": func() (string, error) {
			scroll.SetWheelStep(32)
			return "32", nil
		},
		"SetVerticalScroll": func() (string, error) {
			original := scroll.VerticalScroll()
			defer scroll.SetVerticalScroll(original)
			scroll.SetVerticalScroll(!original)
			if scroll.VerticalScroll() != !original {
				return "", fmt.Errorf("verticalScroll=%v", scroll.VerticalScroll())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"VerticalScroll": func() (string, error) { return fmt.Sprintf("%v", scroll.VerticalScroll()), nil },
		"SetHorizontalScroll": func() (string, error) {
			original := scroll.HorizontalScroll()
			defer scroll.SetHorizontalScroll(original)
			scroll.SetHorizontalScroll(!original)
			if scroll.HorizontalScroll() != !original {
				return "", fmt.Errorf("horizontalScroll=%v", scroll.HorizontalScroll())
			}
			return fmt.Sprintf("-> %v", !original), nil
		},
		"HorizontalScroll": func() (string, error) { return fmt.Sprintf("%v", scroll.HorizontalScroll()), nil },
		"Add": func() (string, error) {
			temp := widgets.NewButton("scroll-add-temp", "Temp Add", widgets.ModeCustom)
			scroll.Add(temp)
			defer scroll.Remove(temp.ID())
			return fmt.Sprintf("children=%d", len(scroll.Children())), nil
		},
		"Remove": func() (string, error) {
			temp := widgets.NewButton("scroll-remove-temp", "Temp Remove", widgets.ModeCustom)
			scroll.Add(temp)
			scroll.Remove(temp.ID())
			for _, child := range scroll.Children() {
				if child != nil && child.ID() == temp.ID() {
					return "", fmt.Errorf("child %q still present", temp.ID())
				}
			}
			return temp.ID(), nil
		},
		"Children": func() (string, error) { return fmt.Sprintf("%d", len(scroll.Children())), nil },
	}
}

func (c *demoController) documentInvokers() map[string]methodInvoker {
	return map[string]methodInvoker{
		"PrimaryWindow": func() (string, error) {
			if c.doc.PrimaryWindow() != c.window {
				return "", fmt.Errorf("primary window mismatch")
			}
			return c.doc.PrimaryWindow().ID, nil
		},
		"Window": func() (string, error) {
			aux := c.doc.Window("aux")
			if aux == nil {
				return "", fmt.Errorf("aux window is nil")
			}
			return aux.ID, nil
		},
		"FindWidget": func() (string, error) {
			widget := c.doc.FindWidget("main", "saveBtn")
			if widget == nil {
				return "", fmt.Errorf("saveBtn not found")
			}
			return widget.ID(), nil
		},
		"SetData": func() (string, error) {
			temp := newDemoStore()
			temp.Set("demo.windowTitle", "Document.SetData")
			temp.Set("demo.reportSummary", "document temp report")
			temp.Set("demo.reportPath", "output\\document-temp.txt")
			temp.Set("demo.paletteName", "Temp Palette")
			temp.Set("demo.showVerticalScroll", false)
			c.doc.SetData(temp)
			defer c.doc.SetData(c.store)
			if c.window.Meta.Title != "Document.SetData" {
				return "", fmt.Errorf("title=%q", c.window.Meta.Title)
			}
			return c.window.Meta.Title, nil
		},
		"NewApps": func() (string, error) {
			hosted, err := c.doc.NewApps(core.Options{
				ClassName:      "WinUIJSONFullDemoTest",
				Title:          "JSON Test",
				Width:          800,
				Height:         600,
				Style:          core.DefaultWindowStyle,
				ExStyle:        core.DefaultWindowExStyle,
				Cursor:         core.CursorArrow,
				Background:     core.RGB(255, 255, 255),
				DoubleBuffered: true,
				RenderMode:     core.RenderModeAuto,
			})
			if err != nil {
				return "", err
			}
			if len(hosted) != len(c.doc.Windows) {
				return "", fmt.Errorf("hosted=%d windows=%d", len(hosted), len(c.doc.Windows))
			}
			return fmt.Sprintf("hosted=%d", len(hosted)), nil
		},
	}
}

func (c *demoController) windowInvokers() map[string]methodInvoker {
	return map[string]methodInvoker{
		"ApplyOptions": func() (string, error) {
			opts := core.Options{Title: "before", Width: 640, Height: 480}
			c.window.ApplyOptions(&opts)
			if opts.Title != c.window.Meta.Title {
				return "", fmt.Errorf("title=%q want %q", opts.Title, c.window.Meta.Title)
			}
			return fmt.Sprintf("%s %dx%d", opts.Title, opts.Width, opts.Height), nil
		},
		"Attach": func() (string, error) {
			aux := c.doc.Window("aux")
			if aux == nil {
				return "", fmt.Errorf("aux window not found")
			}
			scratch := widgets.NewScene(nil)
			scratch.Resize(widgets.Rect{W: 480, H: 320})
			if err := aux.Attach(scratch); err != nil {
				return "", err
			}
			defer aux.Detach()
			if aux.Scene() != scratch {
				return "", fmt.Errorf("scene did not attach")
			}
			return aux.ID, nil
		},
		"Detach": func() (string, error) {
			aux := c.doc.Window("aux")
			if aux == nil {
				return "", fmt.Errorf("aux window not found")
			}
			scratch := widgets.NewScene(nil)
			scratch.Resize(widgets.Rect{W: 480, H: 320})
			if err := aux.Attach(scratch); err != nil {
				return "", err
			}
			if err := aux.Detach(); err != nil {
				return "", err
			}
			if aux.Scene() != nil {
				return "", fmt.Errorf("scene still attached")
			}
			return "detached", nil
		},
		"SetData": func() (string, error) {
			temp := newDemoStore()
			temp.Set("demo.windowTitle", "Window.SetData")
			c.window.SetData(temp)
			defer c.window.SetData(c.store)
			if c.window.Data() != temp {
				return "", fmt.Errorf("window data did not update")
			}
			return c.window.Meta.Title, nil
		},
		"Data": func() (string, error) {
			return fmt.Sprintf("%T", c.window.Data()), nil
		},
		"Scene": func() (string, error) {
			scene := c.window.Scene()
			if scene == nil {
				return "nil", nil
			}
			return "attached", nil
		},
		"App": func() (string, error) {
			if c.window.App() == nil {
				return "nil", nil
			}
			return "attached", nil
		},
		"FindWidget": func() (string, error) {
			widget := c.window.FindWidget("saveBtn")
			if widget == nil {
				return "", fmt.Errorf("saveBtn not found")
			}
			return widget.ID(), nil
		},
		"RefreshBindings": func() (string, error) {
			temp := newDemoStore()
			temp.Set("demo.windowTitle", "Window.RefreshBindings")
			c.window.SetData(temp)
			defer c.window.SetData(c.store)
			c.window.RefreshBindings("demo.windowTitle")
			if c.window.Meta.Title != "Window.RefreshBindings" {
				return "", fmt.Errorf("title=%q", c.window.Meta.Title)
			}
			return c.window.Meta.Title, nil
		},
	}
}

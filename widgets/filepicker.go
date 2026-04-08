//go:build windows

package widgets

import (
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/dialogs"
)

// FilePicker 表示由只读路径框和触发按钮组成的文件选择控件。
type FilePicker struct {
	Panel

	field  *EditBox
	button *Button

	options           dialogs.Options
	paths             []string
	separator         string
	onChange          func([]string)
	customButtonText  bool
	customPlaceholder bool
}

// NewFilePicker 创建一个新的文件选择控件。
func NewFilePicker(id string, mode ControlMode) *FilePicker {
	picker := &FilePicker{
		Panel:     *NewPanel(id),
		separator: "; ",
		options: dialogs.Options{
			Mode: dialogs.DialogOpen,
		},
	}
	picker.SetLayout(RowLayout{
		Gap:        8,
		CrossAlign: AlignStretch,
	})

	picker.field = NewEditBox(id+"-field", mode)
	picker.field.SetMultiline(false)
	picker.field.SetReadOnly(true)
	picker.field.SetLayoutData(FlexLayoutData{Grow: 1, Align: AlignStretch})
	SetPreferredSize(picker.field, core.Size{Width: 220, Height: 36})

	picker.button = NewButton(id+"-button", "", mode)
	picker.button.SetLayoutData(FlexLayoutData{Align: AlignStretch})
	picker.button.SetOnClick(func() {
		picker.showDialog()
	})

	picker.Panel.AddAll(picker.field, picker.button)
	picker.applyDefaults()
	return picker
}

// SetVisible 更新控件可见性，并同步内部子控件。
func (p *FilePicker) SetVisible(visible bool) {
	p.Panel.SetVisible(visible)
	if p.field != nil {
		p.field.SetVisible(visible)
	}
	if p.button != nil {
		p.button.SetVisible(visible)
	}
}

// SetEnabled 更新控件可用状态，并同步内部子控件。
func (p *FilePicker) SetEnabled(enabled bool) {
	p.Panel.SetEnabled(enabled)
	if p.field != nil {
		p.field.SetEnabled(enabled)
	}
	if p.button != nil {
		p.button.SetEnabled(enabled)
	}
}

// SetFieldStyle 设置只读路径框的样式。
func (p *FilePicker) SetFieldStyle(style EditStyle) {
	if p == nil || p.field == nil {
		return
	}
	p.field.SetStyle(style)
}

// SetButtonStyle 设置触发按钮的样式。
func (p *FilePicker) SetButtonStyle(style ButtonStyle) {
	if p == nil || p.button == nil {
		return
	}
	p.button.SetStyle(style)
}

// SetDialogOptions 更新文件对话框配置。
func (p *FilePicker) SetDialogOptions(opts dialogs.Options) {
	if p == nil {
		return
	}
	p.options = opts
	if p.options.Mode == "" {
		p.options.Mode = dialogs.DialogOpen
	}
	p.applyDefaults()
}

// DialogOptions 返回当前文件对话框配置。
func (p *FilePicker) DialogOptions() dialogs.Options {
	if p == nil {
		return dialogs.Options{}
	}
	return p.options
}

// SetSeparator 设置多选路径显示时的分隔符。
func (p *FilePicker) SetSeparator(separator string) {
	if p == nil {
		return
	}
	separator = strings.TrimSpace(separator)
	if separator == "" {
		separator = "; "
	}
	p.separator = separator
	p.updateFieldText()
}

// SetButtonText 设置按钮文本。
func (p *FilePicker) SetButtonText(text string) {
	if p == nil || p.button == nil {
		return
	}
	p.customButtonText = true
	text = strings.TrimSpace(text)
	if text == "" {
		text = defaultButtonText(p.options.Mode)
	}
	p.button.SetText(text)
	SetPreferredSize(p.button, core.Size{Width: pickerButtonWidth(text), Height: 36})
}

// SetPlaceholder 设置空值时的占位文本。
func (p *FilePicker) SetPlaceholder(text string) {
	if p == nil || p.field == nil {
		return
	}
	p.customPlaceholder = true
	p.field.SetPlaceholder(text)
}

// SetPaths 直接更新当前已选路径。
func (p *FilePicker) SetPaths(paths []string) {
	p.setPaths(paths, false)
}

// Paths 返回当前已选路径列表的副本。
func (p *FilePicker) Paths() []string {
	if p == nil || len(p.paths) == 0 {
		return nil
	}
	out := make([]string, len(p.paths))
	copy(out, p.paths)
	return out
}

// SetOnChange 注册路径变更回调。
func (p *FilePicker) SetOnChange(fn func([]string)) {
	if p == nil {
		return
	}
	p.onChange = fn
}

func (p *FilePicker) showDialog() {
	if p == nil {
		return
	}
	var app *core.App
	if scene := p.scene(); scene != nil {
		app = scene.app
	}
	result, err := dialogs.ShowFileDialog(app, p.options)
	if err != nil || result.Canceled() {
		return
	}
	p.setPaths(result.Paths, true)
}

func (p *FilePicker) setPaths(paths []string, notify bool) {
	if p == nil {
		return
	}
	cleaned := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			cleaned = append(cleaned, path)
		}
	}
	p.paths = cleaned
	if len(cleaned) > 0 {
		p.options.InitialPath = cleaned[0]
	}
	p.updateFieldText()
	if notify && p.onChange != nil {
		out := make([]string, len(cleaned))
		copy(out, cleaned)
		p.onChange(out)
	}
}

func (p *FilePicker) updateFieldText() {
	if p == nil || p.field == nil {
		return
	}
	if len(p.paths) == 0 {
		p.field.SetText("")
		return
	}
	p.field.SetText(strings.Join(p.paths, p.separator))
}

func (p *FilePicker) applyDefaults() {
	if p == nil {
		return
	}
	if p.options.Mode == "" {
		p.options.Mode = dialogs.DialogOpen
	}
	if p.field != nil && !p.customPlaceholder {
		p.field.SetPlaceholder(defaultPlaceholder(p.options.Mode))
	}
	if p.button != nil && !p.customButtonText {
		text := defaultButtonText(p.options.Mode)
		p.button.SetText(text)
		SetPreferredSize(p.button, core.Size{Width: pickerButtonWidth(text), Height: 36})
	}
}

func defaultButtonText(mode dialogs.DialogMode) string {
	switch mode {
	case dialogs.DialogSave:
		return "Save As"
	case dialogs.DialogFolder:
		return "Choose Folder"
	default:
		return "Browse"
	}
}

func defaultPlaceholder(mode dialogs.DialogMode) string {
	switch mode {
	case dialogs.DialogSave:
		return "Choose a save path"
	case dialogs.DialogFolder:
		return "No folder selected"
	default:
		return "No file selected"
	}
}

func pickerButtonWidth(text string) int32 {
	length := int32(len([]rune(strings.TrimSpace(text))))
	if length < 6 {
		length = 6
	}
	width := 24 + length*8
	if width < 92 {
		width = 92
	}
	return width
}

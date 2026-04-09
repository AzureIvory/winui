//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/sysapi"
	"github.com/AzureIvory/winui/widgets"
)

type documentSpec struct {
	Wins []windowSpec `json:"wins"`
}

type windowSpec struct {
	ID    string          `json:"id"`
	Title json.RawMessage `json:"title"`
	Icon  string          `json:"icon"`
	W     json.RawMessage `json:"w"`
	H     json.RawMessage `json:"h"`
	MinW  json.RawMessage `json:"minW"`
	MinH  json.RawMessage `json:"minH"`
	BG    string          `json:"bg"`
	Root  *nodeSpec       `json:"root"`
}

type nodeSpec struct {
	Type        string          `json:"type"`
	ID          string          `json:"id"`
	Input       string          `json:"input"`
	Text        json.RawMessage `json:"text"`
	Value       json.RawMessage `json:"value"`
	Placeholder json.RawMessage `json:"placeholder"`
	Visible     json.RawMessage `json:"visible"`
	Enabled     json.RawMessage `json:"enabled"`
	Checked     json.RawMessage `json:"checked"`
	Items       json.RawMessage `json:"items"`
	Sel         json.RawMessage `json:"sel"`
	Layout      json.RawMessage `json:"layout"`
	Frame       *frameSpec      `json:"frame"`
	Style       json.RawMessage `json:"style"`
	Children    []nodeSpec      `json:"children"`

	Group    string          `json:"group"`
	Src      string          `json:"src"`
	Fit      string          `json:"fit"`
	Autoplay json.RawMessage `json:"autoplay"`
	Icon     string          `json:"icon"`
	IconPos  string          `json:"iconPos"`

	Dialog      string          `json:"dialog"`
	Multiple    json.RawMessage `json:"multiple"`
	Accept      json.RawMessage `json:"accept"`
	Filters     json.RawMessage `json:"filters"`
	ButtonText  json.RawMessage `json:"buttonText"`
	DialogTitle json.RawMessage `json:"dialogTitle"`
	DefaultExt  string          `json:"defaultExt"`
	ValueSep    string          `json:"valueSep"`

	OnClick    string `json:"onClick"`
	OnChange   string `json:"onChange"`
	OnSubmit   string `json:"onSubmit"`
	OnActivate string `json:"onActivate"`
}

type frameSpec struct {
	X json.RawMessage `json:"x"`
	Y json.RawMessage `json:"y"`
	R json.RawMessage `json:"r"`
	B json.RawMessage `json:"b"`
	W json.RawMessage `json:"w"`
	H json.RawMessage `json:"h"`
}

type builder struct {
	opts LoadOptions
}

// LoadDocumentFile loads a JSON UI document from disk.
func LoadDocumentFile(path string, opts LoadOptions) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if opts.AssetsDir == "" {
		opts.AssetsDir = filepath.Dir(path)
	}
	return LoadDocumentString(string(data), opts)
}

// LoadDocumentString builds a Document from JSON text.
func LoadDocumentString(text string, opts LoadOptions) (*Document, error) {
	var spec documentSpec
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if err := decoder.Decode(&spec); err != nil {
		return nil, err
	}
	if len(spec.Wins) == 0 {
		return nil, fmt.Errorf("wins must contain at least one window")
	}

	doc := &Document{
		Windows: make([]*Window, 0, len(spec.Wins)),
		index:   map[string]*Window{},
	}
	builder := &builder{opts: opts}
	for _, item := range spec.Wins {
		if item.ID == "" {
			return nil, fmt.Errorf("window id is required")
		}
		if _, exists := doc.index[item.ID]; exists {
			return nil, fmt.Errorf("duplicate window id %q", item.ID)
		}
		if item.Root == nil {
			return nil, fmt.Errorf("window %q is missing root", item.ID)
		}
		window, err := builder.buildWindow(item)
		if err != nil {
			return nil, err
		}
		doc.Windows = append(doc.Windows, window)
		doc.index[window.ID] = window
	}
	doc.SetData(opts.Data)
	return doc, nil
}

// LoadIntoScene loads a single-window JSON UI document and attaches it to a scene.
func LoadIntoScene(scene *widgets.Scene, text string, opts LoadOptions) (*Window, error) {
	doc, err := LoadDocumentString(text, opts)
	if err != nil {
		return nil, err
	}
	window := doc.PrimaryWindow()
	if window == nil {
		return nil, fmt.Errorf("document does not contain a window")
	}
	if err := window.Attach(scene); err != nil {
		return nil, err
	}
	return window, nil
}

// LoadFileIntoScene loads a single-window JSON UI document file and attaches it to a scene.
func LoadFileIntoScene(scene *widgets.Scene, path string, opts LoadOptions) (*Window, error) {
	doc, err := LoadDocumentFile(path, opts)
	if err != nil {
		return nil, err
	}
	window := doc.PrimaryWindow()
	if window == nil {
		return nil, fmt.Errorf("document does not contain a window")
	}
	if err := window.Attach(scene); err != nil {
		return nil, err
	}
	return window, nil
}

func (b *builder) buildWindow(spec windowSpec) (*Window, error) {
	window := &Window{
		ID:    spec.ID,
		theme: b.opts.Theme,
		Meta: WindowMeta{
			ID: spec.ID,
		},
	}
	if spec.Icon != "" {
		icon, err := b.loadICO(spec.Icon)
		if err != nil {
			return nil, fmt.Errorf("window %q icon: %w", spec.ID, err)
		}
		window.Meta.Icon = icon
		window.Meta.IconPath = spec.Icon
	}
	if spec.BG != "" {
		color, ok, err := parseColorValue(spec.BG)
		if err != nil {
			return nil, err
		}
		if ok {
			window.Meta.Background = color
		}
	}
	if source, err := parseStringSource(spec.Title); err != nil {
		return nil, err
	} else if source.Has {
		window.Meta.Title = resolveStringSource(source, b.opts.Data)
		if source.Binding != "" {
			b.addBinding(window, []string{source.Binding}, func(ctx *bindingContext) {
				window.setWindowTitle(resolveStringSource(source, ctx.data))
			})
		}
	}
	if source, err := parseIntSource(spec.W); err != nil {
		return nil, err
	} else if source.Has {
		window.Meta.Width = resolveIntSource(source, b.opts.Data)
	}
	if source, err := parseIntSource(spec.H); err != nil {
		return nil, err
	} else if source.Has {
		window.Meta.Height = resolveIntSource(source, b.opts.Data)
	}
	if source, err := parseIntSource(spec.MinW); err != nil {
		return nil, err
	} else if source.Has {
		window.Meta.MinWidth = resolveIntSource(source, b.opts.Data)
	}
	if source, err := parseIntSource(spec.MinH); err != nil {
		return nil, err
	} else if source.Has {
		window.Meta.MinHeight = resolveIntSource(source, b.opts.Data)
	}

	root, err := b.buildNode(window, *spec.Root, "")
	if err != nil {
		return nil, err
	}
	window.Root = root
	return window, nil
}

func (b *builder) buildNode(window *Window, spec nodeSpec, parentLayout string) (widgets.Widget, error) {
	if spec.Type == "" {
		return nil, fmt.Errorf("widget type is required")
	}

	var (
		widget widgets.Widget
		err    error
	)
	switch spec.Type {
	case "panel":
		widget, err = b.buildPanel(window, spec)
	case "label":
		widget, err = b.buildLabel(window, spec)
	case "button":
		widget, err = b.buildButton(window, spec)
	case "input":
		widget, err = b.buildInput(window, spec)
	case "textarea":
		widget, err = b.buildTextArea(window, spec)
	case "progress":
		widget, err = b.buildProgress(window, spec)
	case "checkbox":
		widget, err = b.buildCheckBox(window, spec)
	case "radio":
		widget, err = b.buildRadio(window, spec)
	case "select":
		widget, err = b.buildSelect(window, spec)
	case "listbox":
		widget, err = b.buildListBox(window, spec)
	case "file":
		widget, err = b.buildFilePicker(window, spec)
	case "image":
		widget, err = b.buildImage(window, spec)
	case "animimg":
		widget, err = b.buildAnimatedImage(window, spec)
	default:
		return nil, fmt.Errorf("unsupported widget type %q", spec.Type)
	}
	if err != nil {
		return nil, err
	}
	if err := b.applyLayoutData(window, widget, spec, parentLayout); err != nil {
		return nil, err
	}
	return widget, nil
}

func (b *builder) buildPanel(window *Window, spec nodeSpec) (widgets.Widget, error) {
	panel := widgets.NewPanel(nodeID(spec))
	layout, layoutKind, err := buildLayout(window, spec.Layout)
	if err != nil {
		return nil, err
	}
	panel.SetLayout(layout)
	if style, err := parsePanelStyle(spec.Style); err != nil {
		return nil, err
	} else {
		panel.SetStyle(style)
	}
	b.applyCommonState(window, panel, spec)
	for _, child := range spec.Children {
		built, err := b.buildNode(window, child, layoutKind)
		if err != nil {
			return nil, err
		}
		panel.Add(built)
	}
	return panel, nil
}

func (b *builder) buildLabel(window *Window, spec nodeSpec) (widgets.Widget, error) {
	textSource, err := parseStringSource(spec.Text)
	if err != nil {
		return nil, err
	}
	label := widgets.NewLabel(nodeID(spec), resolveStringSource(textSource, b.opts.Data))
	style, err := parseTextStyle(spec.Style, false)
	if err != nil {
		return nil, err
	}
	label.SetStyle(style)
	widgets.SetPreferredSize(label, core.Size{Width: 180, Height: 28})
	b.applyCommonState(window, label, spec)
	if textSource.Binding != "" {
		b.addBinding(window, []string{textSource.Binding}, func(ctx *bindingContext) {
			label.SetText(resolveStringSource(textSource, ctx.data))
		})
	}
	return label, nil
}

func (b *builder) buildButton(window *Window, spec nodeSpec) (widgets.Widget, error) {
	textSource, err := parseStringSource(spec.Text)
	if err != nil {
		return nil, err
	}
	button := widgets.NewButton(nodeID(spec), resolveStringSource(textSource, b.opts.Data), b.opts.DefaultMode)
	style, err := parseButtonStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	button.SetStyle(style)
	widgets.SetPreferredSize(button, core.Size{Width: 120, Height: 36})
	if spec.Icon != "" {
		icon, err := b.loadICO(spec.Icon)
		if err != nil {
			return nil, err
		}
		button.SetIcon(icon)
	}
	switch strings.ToLower(strings.TrimSpace(spec.IconPos)) {
	case "", "auto":
	case "left":
		button.SetKind(widgets.BtnLeft)
	case "top":
		button.SetKind(widgets.BtnTop)
	default:
		return nil, fmt.Errorf("unsupported iconPos %q", spec.IconPos)
	}
	b.applyCommonState(window, button, spec)
	if textSource.Binding != "" {
		b.addBinding(window, []string{textSource.Binding}, func(ctx *bindingContext) {
			button.SetText(resolveStringSource(textSource, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnClick); actionName != "" {
		button.SetOnClick(func() {
			ctx := b.baseActionContext(actionName, button)
			ctx.Value = button.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	return button, nil
}

func (b *builder) buildInput(window *Window, spec nodeSpec) (widgets.Widget, error) {
	inputType := strings.ToLower(strings.TrimSpace(spec.Input))
	if inputType == "" {
		inputType = "text"
	}
	switch inputType {
	case "text", "password":
	default:
		return nil, fmt.Errorf("unsupported input kind %q", spec.Input)
	}

	valueSource, err := parseStringSource(spec.Value)
	if err != nil {
		return nil, err
	}
	placeholderSource, err := parseStringSource(spec.Placeholder)
	if err != nil {
		return nil, err
	}
	edit := widgets.NewEditBox(nodeID(spec), b.opts.DefaultMode)
	edit.SetMultiline(false)
	if inputType == "password" {
		edit.SetPassword(true)
	}
	style, err := parseEditStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	edit.SetStyle(style)
	edit.SetText(resolveStringSource(valueSource, b.opts.Data))
	edit.SetPlaceholder(resolveStringSource(placeholderSource, b.opts.Data))
	widgets.SetPreferredSize(edit, core.Size{Width: 220, Height: 36})
	b.applyCommonState(window, edit, spec)
	if valueSource.Binding != "" {
		b.addBinding(window, []string{valueSource.Binding}, func(ctx *bindingContext) {
			edit.SetText(resolveStringSource(valueSource, ctx.data))
		})
	}
	if placeholderSource.Binding != "" {
		b.addBinding(window, []string{placeholderSource.Binding}, func(ctx *bindingContext) {
			edit.SetPlaceholder(resolveStringSource(placeholderSource, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		edit.SetOnChange(func(_ string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName := strings.TrimSpace(spec.OnSubmit); actionName != "" {
		edit.SetOnSubmit(func(_ string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	return edit, nil
}

func (b *builder) buildTextArea(window *Window, spec nodeSpec) (widgets.Widget, error) {
	valueSource, err := parseStringSource(spec.Value)
	if err != nil {
		return nil, err
	}
	placeholderSource, err := parseStringSource(spec.Placeholder)
	if err != nil {
		return nil, err
	}
	edit := widgets.NewEditBox(nodeID(spec), b.opts.DefaultMode)
	edit.SetMultiline(true)
	edit.SetAcceptReturn(true)
	edit.SetVerticalScroll(true)
	style, err := parseEditStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	edit.SetStyle(style)
	edit.SetText(resolveStringSource(valueSource, b.opts.Data))
	edit.SetPlaceholder(resolveStringSource(placeholderSource, b.opts.Data))
	widgets.SetPreferredSize(edit, core.Size{Width: 320, Height: 96})
	b.applyCommonState(window, edit, spec)
	if valueSource.Binding != "" {
		b.addBinding(window, []string{valueSource.Binding}, func(ctx *bindingContext) {
			edit.SetText(resolveStringSource(valueSource, ctx.data))
		})
	}
	if placeholderSource.Binding != "" {
		b.addBinding(window, []string{placeholderSource.Binding}, func(ctx *bindingContext) {
			edit.SetPlaceholder(resolveStringSource(placeholderSource, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		edit.SetOnChange(func(_ string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName := strings.TrimSpace(spec.OnSubmit); actionName != "" {
		edit.SetOnSubmit(func(_ string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	return edit, nil
}

func (b *builder) buildProgress(window *Window, spec nodeSpec) (widgets.Widget, error) {
	valueSource, err := parseIntSource(spec.Value)
	if err != nil {
		return nil, err
	}
	progress := widgets.NewProgressBar(nodeID(spec))
	style, err := parseProgressStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	progress.SetStyle(style)
	progress.SetValue(resolveIntSource(valueSource, b.opts.Data))
	widgets.SetPreferredSize(progress, core.Size{Width: 220, Height: 24})
	b.applyCommonState(window, progress, spec)
	if valueSource.Binding != "" {
		b.addBinding(window, []string{valueSource.Binding}, func(ctx *bindingContext) {
			progress.SetValue(resolveIntSource(valueSource, ctx.data))
		})
	}
	return progress, nil
}

func (b *builder) buildCheckBox(window *Window, spec nodeSpec) (widgets.Widget, error) {
	textSource, err := parseStringSource(spec.Text)
	if err != nil {
		return nil, err
	}
	checkedSource, err := parseBoolSource(spec.Checked)
	if err != nil {
		return nil, err
	}
	check := widgets.NewCheckBox(nodeID(spec), resolveStringSource(textSource, b.opts.Data), b.opts.DefaultMode)
	style, err := parseChoiceStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	check.SetStyle(style)
	check.SetChecked(resolveBoolSource(checkedSource, b.opts.Data))
	widgets.SetPreferredSize(check, core.Size{Width: 160, Height: 28})
	b.applyCommonState(window, check, spec)
	if textSource.Binding != "" {
		b.addBinding(window, []string{textSource.Binding}, func(ctx *bindingContext) {
			check.SetText(resolveStringSource(textSource, ctx.data))
		})
	}
	if checkedSource.Binding != "" {
		b.addBinding(window, []string{checkedSource.Binding}, func(ctx *bindingContext) {
			check.SetChecked(resolveBoolSource(checkedSource, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		check.SetOnChange(func(checked bool) {
			ctx := b.baseActionContext(actionName, check)
			ctx.Checked = checked
			ctx.Value = check.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	return check, nil
}

func (b *builder) buildRadio(window *Window, spec nodeSpec) (widgets.Widget, error) {
	textSource, err := parseStringSource(spec.Text)
	if err != nil {
		return nil, err
	}
	checkedSource, err := parseBoolSource(spec.Checked)
	if err != nil {
		return nil, err
	}
	radio := widgets.NewRadioButton(nodeID(spec), resolveStringSource(textSource, b.opts.Data), b.opts.DefaultMode)
	if spec.Group != "" {
		radio.SetGroup(spec.Group)
	}
	style, err := parseChoiceStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	radio.SetStyle(style)
	radio.SetChecked(resolveBoolSource(checkedSource, b.opts.Data))
	widgets.SetPreferredSize(radio, core.Size{Width: 160, Height: 28})
	b.applyCommonState(window, radio, spec)
	if textSource.Binding != "" {
		b.addBinding(window, []string{textSource.Binding}, func(ctx *bindingContext) {
			radio.SetText(resolveStringSource(textSource, ctx.data))
		})
	}
	if checkedSource.Binding != "" {
		b.addBinding(window, []string{checkedSource.Binding}, func(ctx *bindingContext) {
			radio.SetChecked(resolveBoolSource(checkedSource, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		radio.SetOnChange(func(checked bool) {
			ctx := b.baseActionContext(actionName, radio)
			ctx.Checked = checked
			ctx.Value = radio.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	return radio, nil
}

func (b *builder) buildSelect(window *Window, spec nodeSpec) (widgets.Widget, error) {
	itemsSource, selectedLiteral, err := b.parseItemsAndSelection(spec)
	if err != nil {
		return nil, err
	}
	selSource, err := parseSelectionSource(spec.Sel)
	if err != nil {
		return nil, err
	}
	placeholderSource, err := parseStringSource(spec.Placeholder)
	if err != nil {
		return nil, err
	}
	combo := widgets.NewComboBox(nodeID(spec), b.opts.DefaultMode)
	style, err := parseComboStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	combo.SetStyle(style)
	items := resolveItemsSource(itemsSource, b.opts.Data)
	combo.SetItems(items)
	combo.SetPlaceholder(resolveStringSource(placeholderSource, b.opts.Data))
	combo.SetSelected(resolveSelectionSource(selSource, items, selectedLiteral, b.opts.Data))
	widgets.SetPreferredSize(combo, core.Size{Width: 220, Height: 36})
	b.applyCommonState(window, combo, spec)
	if itemsSource.Binding != "" {
		b.addBinding(window, []string{itemsSource.Binding}, func(ctx *bindingContext) {
			items := resolveItemsSource(itemsSource, ctx.data)
			combo.SetItems(items)
			combo.SetSelected(resolveSelectionSource(selSource, items, selectedLiteral, ctx.data))
		})
	}
	if placeholderSource.Binding != "" {
		b.addBinding(window, []string{placeholderSource.Binding}, func(ctx *bindingContext) {
			combo.SetPlaceholder(resolveStringSource(placeholderSource, ctx.data))
		})
	}
	if selSource.Binding != "" {
		b.addBinding(window, []string{selSource.Binding}, func(ctx *bindingContext) {
			combo.SetSelected(resolveSelectionSource(selSource, combo.Items(), selectedLiteral, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		combo.SetOnChange(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, combo)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	return combo, nil
}

func (b *builder) buildListBox(window *Window, spec nodeSpec) (widgets.Widget, error) {
	itemsSource, selectedLiteral, err := b.parseItemsAndSelection(spec)
	if err != nil {
		return nil, err
	}
	selSource, err := parseSelectionSource(spec.Sel)
	if err != nil {
		return nil, err
	}
	list := widgets.NewListBox(nodeID(spec))
	style, err := parseListStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	list.SetStyle(style)
	items := resolveItemsSource(itemsSource, b.opts.Data)
	list.SetItems(items)
	list.SetSelected(resolveSelectionSource(selSource, items, selectedLiteral, b.opts.Data))
	widgets.SetPreferredSize(list, core.Size{Width: 220, Height: 160})
	b.applyCommonState(window, list, spec)
	if itemsSource.Binding != "" {
		b.addBinding(window, []string{itemsSource.Binding}, func(ctx *bindingContext) {
			items := resolveItemsSource(itemsSource, ctx.data)
			list.SetItems(items)
			list.SetSelected(resolveSelectionSource(selSource, items, selectedLiteral, ctx.data))
		})
	}
	if selSource.Binding != "" {
		b.addBinding(window, []string{selSource.Binding}, func(ctx *bindingContext) {
			list.SetSelected(resolveSelectionSource(selSource, list.Items(), selectedLiteral, ctx.data))
		})
	}
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		list.SetOnChange(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, list)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName := strings.TrimSpace(spec.OnActivate); actionName != "" {
		list.SetOnActivate(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, list)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	return list, nil
}

func (b *builder) buildFilePicker(window *Window, spec nodeSpec) (widgets.Widget, error) {
	picker := widgets.NewFilePicker(nodeID(spec), b.opts.DefaultMode)
	editStyle, err := parseEditStyle(spec.Style)
	if err != nil {
		return nil, err
	}
	picker.SetFieldStyle(editStyle)
	if styles, err := decodeStyleMap(spec.Style); err != nil {
		return nil, err
	} else if styles != nil && len(styles["btn"]) > 0 {
		buttonStyle, err := parseButtonStyle(styles["btn"])
		if err != nil {
			return nil, err
		}
		picker.SetButtonStyle(buttonStyle)
	}
	options, err := b.fileDialogOptions(spec)
	if err != nil {
		return nil, err
	}
	picker.SetDialogOptions(options)
	if spec.ValueSep != "" {
		picker.SetSeparator(spec.ValueSep)
	}
	if placeholderSource, err := parseStringSource(spec.Placeholder); err != nil {
		return nil, err
	} else if placeholderSource.Has {
		picker.SetPlaceholder(resolveStringSource(placeholderSource, b.opts.Data))
		if placeholderSource.Binding != "" {
			b.addBinding(window, []string{placeholderSource.Binding}, func(ctx *bindingContext) {
				picker.SetPlaceholder(resolveStringSource(placeholderSource, ctx.data))
			})
		}
	}
	if buttonTextSource, err := parseStringSource(spec.ButtonText); err != nil {
		return nil, err
	} else if buttonTextSource.Has {
		picker.SetButtonText(resolveStringSource(buttonTextSource, b.opts.Data))
		if buttonTextSource.Binding != "" {
			b.addBinding(window, []string{buttonTextSource.Binding}, func(ctx *bindingContext) {
				picker.SetButtonText(resolveStringSource(buttonTextSource, ctx.data))
			})
		}
	}
	if valueSource, err := parseStringSource(spec.Value); err != nil {
		return nil, err
	} else if valueSource.Has {
		initial := resolveStringSource(valueSource, b.opts.Data)
		if initial != "" {
			picker.SetPaths([]string{initial})
		}
		if valueSource.Binding != "" {
			b.addBinding(window, []string{valueSource.Binding}, func(ctx *bindingContext) {
				value := resolveStringSource(valueSource, ctx.data)
				if value == "" {
					picker.SetPaths(nil)
					return
				}
				picker.SetPaths([]string{value})
			})
		}
	}
	widgets.SetPreferredSize(picker, core.Size{Width: 360, Height: 36})
	b.applyCommonState(window, picker, spec)
	if actionName := strings.TrimSpace(spec.OnChange); actionName != "" {
		picker.SetOnChange(func(paths []string) {
			ctx := b.baseActionContext(actionName, picker)
			ctx.Paths = append([]string(nil), paths...)
			if len(paths) > 0 {
				ctx.Value = paths[0]
			}
			b.dispatchAction(actionName, ctx)
		})
	}
	return picker, nil
}

func (b *builder) buildImage(window *Window, spec nodeSpec) (widgets.Widget, error) {
	imageWidget := widgets.NewImage(nodeID(spec))
	if fit := strings.ToLower(strings.TrimSpace(spec.Fit)); fit != "" {
		switch fit {
		case "contain":
			imageWidget.SetScaleMode(widgets.ImageScaleContain)
		case "fill", "cover":
			imageWidget.SetScaleMode(widgets.ImageScaleStretch)
		default:
			return nil, fmt.Errorf("unsupported fit %q", spec.Fit)
		}
	}
	if spec.Src != "" {
		data, err := os.ReadFile(b.resolveAssetPath(spec.Src))
		if err != nil {
			return nil, err
		}
		if err := imageWidget.LoadBytes(data); err != nil {
			return nil, err
		}
	}
	widgets.SetPreferredSize(imageWidget, core.Size{Width: 160, Height: 120})
	b.applyCommonState(window, imageWidget, spec)
	return imageWidget, nil
}

func (b *builder) buildAnimatedImage(window *Window, spec nodeSpec) (widgets.Widget, error) {
	imageWidget := widgets.NewAnimatedImage(nodeID(spec))
	if fit := strings.ToLower(strings.TrimSpace(spec.Fit)); fit != "" {
		switch fit {
		case "contain":
			imageWidget.SetScaleMode(widgets.ImageScaleContain)
		case "fill", "cover":
			imageWidget.SetScaleMode(widgets.ImageScaleStretch)
		default:
			return nil, fmt.Errorf("unsupported fit %q", spec.Fit)
		}
	}
	if spec.Src != "" {
		data, err := os.ReadFile(b.resolveAssetPath(spec.Src))
		if err != nil {
			return nil, err
		}
		if err := imageWidget.LoadGIF(data); err != nil {
			return nil, err
		}
	}
	autoplaySource, err := parseBoolSource(spec.Autoplay)
	if err != nil {
		return nil, err
	}
	if autoplaySource.Has {
		imageWidget.SetPlaying(resolveBoolSource(autoplaySource, b.opts.Data))
	}
	widgets.SetPreferredSize(imageWidget, core.Size{Width: 64, Height: 64})
	b.applyCommonState(window, imageWidget, spec)
	if autoplaySource.Binding != "" {
		b.addBinding(window, []string{autoplaySource.Binding}, func(ctx *bindingContext) {
			imageWidget.SetPlaying(resolveBoolSource(autoplaySource, ctx.data))
		})
	}
	return imageWidget, nil
}

func (b *builder) applyCommonState(window *Window, widget widgets.Widget, spec nodeSpec) {
	visibleSource, _ := parseBoolSource(spec.Visible)
	if visibleSource.Has {
		widget.SetVisible(resolveBoolSource(visibleSource, b.opts.Data))
	}
	enabledSource, _ := parseBoolSource(spec.Enabled)
	if enabledSource.Has {
		widget.SetEnabled(resolveBoolSource(enabledSource, b.opts.Data))
	}
	if visibleSource.Binding != "" {
		b.addBinding(window, []string{visibleSource.Binding}, func(ctx *bindingContext) {
			widget.SetVisible(resolveBoolSource(visibleSource, ctx.data))
		})
	}
	if enabledSource.Binding != "" {
		b.addBinding(window, []string{enabledSource.Binding}, func(ctx *bindingContext) {
			widget.SetEnabled(resolveBoolSource(enabledSource, ctx.data))
		})
	}
}

func (b *builder) applyLayoutData(window *Window, widget widgets.Widget, spec nodeSpec, parentLayout string) error {
	if widget == nil || spec.Frame == nil {
		return nil
	}
	if parentLayout != "abs" {
		return fmt.Errorf("frame is only supported inside abs layout")
	}
	data, err := b.buildAbsoluteLayoutData(window, *spec.Frame)
	if err != nil {
		return err
	}
	widget.SetLayoutData(data)
	b.registerFrameBindings(window, widget, data)
	return nil
}

func (b *builder) buildAbsoluteLayoutData(window *Window, frame frameSpec) (absoluteLayoutData, error) {
	data := absoluteLayoutData{window: window}
	var err error
	if data.frame.X, err = parseExprSource(frame.X); err != nil {
		return data, err
	}
	if data.frame.Y, err = parseExprSource(frame.Y); err != nil {
		return data, err
	}
	if data.frame.R, err = parseExprSource(frame.R); err != nil {
		return data, err
	}
	if data.frame.B, err = parseExprSource(frame.B); err != nil {
		return data, err
	}
	if data.frame.W, err = parseExprSource(frame.W); err != nil {
		return data, err
	}
	if data.frame.H, err = parseExprSource(frame.H); err != nil {
		return data, err
	}
	data.frame.X = resolveExprSource(data.frame.X, b.opts.Data)
	data.frame.Y = resolveExprSource(data.frame.Y, b.opts.Data)
	data.frame.R = resolveExprSource(data.frame.R, b.opts.Data)
	data.frame.B = resolveExprSource(data.frame.B, b.opts.Data)
	data.frame.W = resolveExprSource(data.frame.W, b.opts.Data)
	data.frame.H = resolveExprSource(data.frame.H, b.opts.Data)
	return data, nil
}

func resolveExprSource(source exprSource, data DataSource) exprSource {
	if !source.Has || source.Binding == "" {
		return source
	}
	raw, ok := dataLookup(data, source.Binding)
	if ok {
		if expr, ok := bindingExprValue(raw); ok {
			source.Literal = expr
			return source
		}
	}
	source.Literal = source.Default
	return source
}

func (b *builder) registerFrameBindings(window *Window, widget widgets.Widget, data absoluteLayoutData) {
	paths := []string{}
	for _, source := range []exprSource{
		data.frame.X,
		data.frame.Y,
		data.frame.R,
		data.frame.B,
		data.frame.W,
		data.frame.H,
	} {
		if source.Binding != "" {
			paths = append(paths, source.Binding)
		}
	}
	if len(paths) == 0 {
		return
	}
	b.addBinding(window, paths, func(ctx *bindingContext) {
		updated := data
		updated.frame.X = resolveExprSource(updated.frame.X, ctx.data)
		updated.frame.Y = resolveExprSource(updated.frame.Y, ctx.data)
		updated.frame.R = resolveExprSource(updated.frame.R, ctx.data)
		updated.frame.B = resolveExprSource(updated.frame.B, ctx.data)
		updated.frame.W = resolveExprSource(updated.frame.W, ctx.data)
		updated.frame.H = resolveExprSource(updated.frame.H, ctx.data)
		widget.SetLayoutData(updated)
	})
}

func (b *builder) parseItemsAndSelection(spec nodeSpec) (itemsSource, any, error) {
	itemsSource, err := parseItemsSource(spec.Items)
	if err != nil {
		return itemsSource, nil, err
	}
	selectedLiteral, err := b.selectedLiteral(spec.Items)
	if err != nil {
		return itemsSource, nil, err
	}
	return itemsSource, selectedLiteral, nil
}

func (b *builder) selectedLiteral(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if _, ok, err := parseBindingSpec(raw); err != nil {
		return nil, err
	} else if ok {
		return nil, nil
	}
	var items []itemLiteral
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, nil
	}
	for index, item := range items {
		if item.Selected {
			if item.Value != "" {
				return item.Value, nil
			}
			if item.Text != "" {
				return item.Text, nil
			}
			return index, nil
		}
	}
	return nil, nil
}

func (b *builder) addBinding(window *Window, paths []string, apply func(*bindingContext)) {
	if window == nil || apply == nil {
		return
	}
	normalized := make([]string, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = normalizeBindingPath(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, path)
	}
	window.bindings = append(window.bindings, windowBinding{
		paths: normalized,
		apply: apply,
	})
}

func resolveStringSource(source stringSource, data DataSource) string {
	if !source.Has {
		return ""
	}
	if source.Binding == "" {
		return source.Literal
	}
	if raw, ok := dataLookup(data, source.Binding); ok {
		if value, ok := bindingStringValue(raw); ok {
			return value
		}
	}
	return source.Default
}

func resolveBoolSource(source boolSource, data DataSource) bool {
	if !source.Has {
		return true
	}
	if source.Binding == "" {
		return source.Literal
	}
	if raw, ok := dataLookup(data, source.Binding); ok {
		if value, ok := bindingBoolValue(raw); ok {
			return value
		}
	}
	return source.Default
}

func resolveIntSource(source intSource, data DataSource) int32 {
	if !source.Has {
		return 0
	}
	if source.Binding == "" {
		return source.Literal
	}
	if raw, ok := dataLookup(data, source.Binding); ok {
		if value, ok := bindingInt32Value(raw); ok {
			return value
		}
	}
	return source.Default
}

func resolveItemsSource(source itemsSource, data DataSource) []widgets.ListItem {
	if !source.Has {
		return nil
	}
	if source.Binding == "" {
		return cloneListItems(source.Literal)
	}
	if raw, ok := dataLookup(data, source.Binding); ok {
		if value, ok := bindingListItemsValue(raw); ok {
			return value
		}
	}
	return cloneListItems(source.Default)
}

func resolveSelectionSource(source selectionSource, items []widgets.ListItem, selectedLiteral any, data DataSource) int {
	if source.Has {
		if source.Binding == "" {
			if index, ok := bindingSelectionIndex(items, source.Literal); ok {
				return index
			}
		} else if raw, ok := dataLookup(data, source.Binding); ok {
			if index, ok := bindingSelectionIndex(items, raw); ok {
				return index
			}
		} else if source.HasSelected {
			if index, ok := bindingSelectionIndex(items, source.Default); ok {
				return index
			}
		}
	}
	if selectedLiteral != nil {
		if index, ok := bindingSelectionIndex(items, selectedLiteral); ok {
			return index
		}
	}
	return -1
}

func dataLookup(data DataSource, path string) (any, bool) {
	if data == nil {
		return nil, false
	}
	return data.Get(path)
}

func nodeID(spec nodeSpec) string {
	if strings.TrimSpace(spec.ID) != "" {
		return spec.ID
	}
	return ""
}

func (b *builder) resolveAssetPath(path string) string {
	resolved := strings.TrimSpace(path)
	if resolved == "" {
		return ""
	}
	if filepath.IsAbs(resolved) || b.opts.AssetsDir == "" {
		return resolved
	}
	return filepath.Join(b.opts.AssetsDir, resolved)
}

func (b *builder) loadICO(src string) (*core.Icon, error) {
	if strings.ToLower(filepath.Ext(src)) != ".ico" {
		return nil, fmt.Errorf("icon %q must be a local .ico file", src)
	}
	data, err := os.ReadFile(b.resolveAssetPath(src))
	if err != nil {
		return nil, err
	}
	return core.LoadIconFromICO(data, 32)
}

func (b *builder) baseActionContext(name string, widget widgets.Widget) ActionContext {
	ctx := ActionContext{
		Name:   name,
		Widget: widget,
		Index:  -1,
	}
	if widget != nil {
		ctx.ID = widget.ID()
	}
	return ctx
}

func (b *builder) dispatchAction(name string, ctx ActionContext) {
	if handler := b.opts.ActionHandlers[name]; handler != nil {
		handler(ctx)
		return
	}
	if action := b.opts.Actions[name]; action != nil {
		action()
	}
}

func (b *builder) fileDialogOptions(spec nodeSpec) (sysapi.Options, error) {
	options := sysapi.Options{}
	switch strings.ToLower(strings.TrimSpace(spec.Dialog)) {
	case "", "open":
		options.Mode = sysapi.DialogOpen
	case "save":
		options.Mode = sysapi.DialogSave
	case "folder":
		options.Mode = sysapi.DialogFolder
	default:
		return options, fmt.Errorf("unsupported dialog mode %q", spec.Dialog)
	}
	if titleSource, err := parseStringSource(spec.DialogTitle); err != nil {
		return options, err
	} else if titleSource.Has {
		options.Title = resolveStringSource(titleSource, b.opts.Data)
	}
	if buttonTextSource, err := parseStringSource(spec.ButtonText); err != nil {
		return options, err
	} else if buttonTextSource.Has {
		options.ButtonLabel = resolveStringSource(buttonTextSource, b.opts.Data)
	}
	options.DefaultExtension = strings.TrimSpace(spec.DefaultExt)
	if multipleSource, err := parseBoolSource(spec.Multiple); err != nil {
		return options, err
	} else if multipleSource.Has {
		options.MultiSelect = resolveBoolSource(multipleSource, b.opts.Data)
	}
	filters, err := parseFileFilters(spec.Filters, spec.Accept)
	if err != nil {
		return options, err
	}
	options.Filters = filters
	return options, nil
}

type filterLiteral struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

func parseFileFilters(filtersRaw json.RawMessage, acceptRaw json.RawMessage) ([]sysapi.FileFilter, error) {
	if len(filtersRaw) > 0 {
		var filters []filterLiteral
		if err := json.Unmarshal(filtersRaw, &filters); err == nil {
			out := make([]sysapi.FileFilter, 0, len(filters))
			for _, filter := range filters {
				if filter.Name == "" || filter.Pattern == "" {
					continue
				}
				out = append(out, sysapi.FileFilter{Name: filter.Name, Pattern: filter.Pattern})
			}
			return out, nil
		}

		text, err := decodeStringLiteral(filtersRaw)
		if err != nil {
			return nil, err
		}
		parts := strings.Split(text, ",")
		out := make([]sysapi.FileFilter, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			chunks := strings.SplitN(part, "=", 2)
			if len(chunks) != 2 {
				return nil, fmt.Errorf("invalid filter %q", part)
			}
			out = append(out, sysapi.FileFilter{
				Name:    strings.TrimSpace(chunks[0]),
				Pattern: strings.TrimSpace(chunks[1]),
			})
		}
		return out, nil
	}

	if len(acceptRaw) == 0 {
		return nil, nil
	}
	var acceptList []string
	if err := json.Unmarshal(acceptRaw, &acceptList); err == nil {
		if len(acceptList) == 0 {
			return nil, nil
		}
		return []sysapi.FileFilter{{
			Name:    "Accepted Files",
			Pattern: strings.Join(acceptList, ";"),
		}}, nil
	}

	text, err := decodeStringLiteral(acceptRaw)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	return []sysapi.FileFilter{{
		Name:    "Accepted Files",
		Pattern: strings.ReplaceAll(text, ",", ";"),
	}}, nil
}

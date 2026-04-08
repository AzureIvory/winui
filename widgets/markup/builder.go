//go:build windows

package markup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// LoadOptions 描述�?HTML/CSS DSL 构建控件树时使用的可选参数�?
type LoadOptions struct {
	// Actions ���ڰ� onclick/onchange/onsubmit �ȶ������󶨵��޲λص���
	Actions map[string]func()
	// ActionHandlers ���ڰ󶨴������ĵĶ����ص�������ʱ������ Actions��
	ActionHandlers map[string]func(ActionContext)
	// AssetsDir ָ�������Դ�ļ������� img[src]���Ĳ���Ŀ¼��
	AssetsDir string
	// DefaultMode ָ��������ؼ�Ĭ��ʹ�õĺ��ģʽ��
	DefaultMode widgets.ControlMode
	// Theme ���ĵ����ص� Scene ʱ��Ӧ�õ����⡣
	Theme *widgets.Theme
}

// LoadHTMLFile �?.ui.html 文件加载控件树，并自动尝试读取同�?.ui.css 文件�?
func LoadHTMLFile(path string, opts LoadOptions) (widgets.Widget, error) {
	doc, err := LoadDocumentFile(path, opts)
	if err != nil {
		return nil, err
	}
	return doc.Root, nil
}

// LoadHTMLString �?HTML 文本�?CSS 文本构建控件树�?
func LoadHTMLString(htmlText string, cssText string, opts LoadOptions) (widgets.Widget, error) {
	doc, err := LoadDocumentString(htmlText, cssText, opts)
	if err != nil {
		return nil, err
	}
	return doc.Root, nil
}

type uiBuilder struct {
	opts       LoadOptions
	autoIDSeed int
}

func (b *uiBuilder) buildDocument(root *node) (*Document, error) {
	if root == nil || root.Kind != nodeElement {
		return nil, newParseError("builder", position{}, "document", "invalid root element")
	}
	doc := &Document{}
	uiRoot := root
	if root.Tag == "window" {
		body, meta, err := b.extractWindowRoot(root)
		if err != nil {
			return nil, err
		}
		doc.Meta = meta
		uiRoot = body
	}
	widget, err := b.buildNode(uiRoot)
	if err != nil {
		return nil, err
	}
	doc.Root = widget
	return doc, nil
}

func (b *uiBuilder) extractWindowRoot(window *node) (*node, WindowMeta, error) {
	meta, err := b.parseWindowMeta(window)
	if err != nil {
		return nil, meta, err
	}
	var body *node
	for _, child := range window.Children {
		if child == nil {
			continue
		}
		switch child.Kind {
		case nodeText:
			if strings.TrimSpace(child.Text) != "" {
				return nil, meta, newParseError("builder", child.Pos, window.inlineContext(), "text is not allowed directly inside <window>")
			}
		case nodeElement:
			if child.Tag != "body" {
				return nil, meta, newParseError("builder", child.Pos, child.inlineContext(), "<window> only accepts one <body> child")
			}
			if body != nil {
				return nil, meta, newParseError("builder", child.Pos, window.inlineContext(), "<window> only accepts one <body> child")
			}
			body = child
		}
	}
	if body == nil {
		return nil, meta, newParseError("builder", window.Pos, window.inlineContext(), "<window> requires one <body> child")
	}
	return body, meta, nil
}

func (b *uiBuilder) parseWindowMeta(window *node) (WindowMeta, error) {
	meta := WindowMeta{}
	meta.Title = strings.TrimSpace(window.attr("title"))
	iconPath := strings.TrimSpace(window.attr("icon"))
	if iconPath != "" {
		icon, err := b.loadICO(iconPath, window.Pos, window.inlineContext(), "window icon")
		if err != nil {
			return meta, err
		}
		meta.IconPath = iconPath
		meta.Icon = icon
	}
	if value := strings.TrimSpace(window.attr("min-width")); value != "" {
		length, ok, err := parseLengthValue(value)
		if err != nil || !ok {
			return meta, newParseError("builder", window.Pos, window.inlineContext(), "invalid min-width %q", value)
		}
		meta.MinWidth = max32(0, length)
	}
	if value := strings.TrimSpace(window.attr("min-height")); value != "" {
		length, ok, err := parseLengthValue(value)
		if err != nil || !ok {
			return meta, newParseError("builder", window.Pos, window.inlineContext(), "invalid min-height %q", value)
		}
		meta.MinHeight = max32(0, length)
	}
	return meta, nil
}

func (b *uiBuilder) buildNode(n *node) (widgets.Widget, error) {
	if n == nil || n.Kind != nodeElement {
		return nil, newParseError("builder", position{}, "node", "invalid node")
	}
	switch n.Tag {
	case "body", "div", "section", "panel", "row", "column", "form", "scroll":
		return b.buildContainer(n)
	case "label", "span":
		return b.buildLabel(n)
	case "button":
		return b.buildButton(n)
	case "input":
		return b.buildInput(n)
	case "textarea":
		return b.buildTextArea(n)
	case "img":
		return b.buildImage(n)
	case "animated-img":
		return b.buildAnimatedImage(n)
	case "progress":
		return b.buildProgress(n)
	case "checkbox":
		return b.buildCheckBox(n)
	case "radio":
		return b.buildRadio(n)
	case "select":
		return b.buildSelect(n)
	case "listbox":
		return b.buildListBox(n)
	case "option":
		return nil, newParseError("builder", n.Pos, n.inlineContext(), "option can only appear inside select/listbox")
	default:
		return nil, newParseError("builder", n.Pos, n.inlineContext(), "unsupported tag <%s>", n.Tag)
	}
}

func (b *uiBuilder) buildContainer(n *node) (widgets.Widget, error) {
	id := b.nodeID(n)
	layoutKind, layout, err := b.layoutForNode(n)
	if err != nil {
		return nil, err
	}
	useScroll, allowX, allowY, err := b.scrollConfig(n)
	if err != nil {
		return nil, err
	}
	if useScroll || n.Tag == "scroll" {
		scroll := widgets.NewScrollView(id)
		scroll.SetStyle(b.panelStyle(n))
		scroll.SetHorizontalScroll(allowX)
		scroll.SetVerticalScroll(allowY)
		if err := b.applyCommonWidgetState(scroll, n); err != nil {
			return nil, err
		}
		content, err := b.buildScrollContent(n, id, layoutKind, layout)
		if err != nil {
			return nil, err
		}
		if err := b.bindOnClickIfAny(content, n); err != nil {
			return nil, err
		}
		scroll.SetContent(content)
		defaultSize := core.Size{Width: 320, Height: 180}
		if content != nil {
			defaultSize = widgets.MeasureNatural(content)
			if defaultSize.Width <= 0 {
				defaultSize.Width = 320
			}
			if defaultSize.Height <= 0 {
				defaultSize.Height = 180
			}
		}
		if err := b.applyPreferredSize(scroll, n, defaultSize); err != nil {
			return nil, err
		}
		return scroll, nil
	}

	panel := widgets.NewPanel(id)
	panel.SetLayout(layout)
	panel.SetStyle(b.panelStyle(n))
	if err := b.applyCommonWidgetState(panel, n); err != nil {
		return nil, err
	}
	if err := b.bindOnClickIfAny(panel, n); err != nil {
		return nil, err
	}
	if err := b.addContainerChildren(panel, n, layoutKind); err != nil {
		return nil, err
	}
	defaultSize := widgets.MeasureNatural(panel)
	if defaultSize.Width <= 0 && defaultSize.Height <= 0 {
		defaultSize = core.Size{Width: 240, Height: 120}
	}
	if err := b.applyPreferredSize(panel, n, defaultSize); err != nil {
		return nil, err
	}
	return panel, nil
}

func (b *uiBuilder) buildScrollContent(n *node, outerID string, layoutKind string, layout widgets.Layout) (widgets.Widget, error) {
	childElements := n.elementChildren()
	textChildren := b.nonWhitespaceTextChildren(n)
	if n.Tag == "scroll" && len(childElements) == 1 && len(textChildren) == 0 {
		return b.buildNode(childElements[0])
	}
	contentID := outerID + "-content"
	panel := widgets.NewPanel(contentID)
	if layout == nil {
		layout = widgets.ColumnLayout{}
		layoutKind = "column"
	}
	panel.SetLayout(layout)
	if err := b.addContainerChildren(panel, n, layoutKind); err != nil {
		return nil, err
	}
	defaultSize := widgets.MeasureNatural(panel)
	if defaultSize.Width <= 0 {
		defaultSize.Width = 240
	}
	if defaultSize.Height <= 0 {
		defaultSize.Height = 120
	}
	if err := b.applyPreferredSize(panel, n, defaultSize); err != nil {
		return nil, err
	}
	return panel, nil
}

func (b *uiBuilder) addContainerChildren(panel *widgets.Panel, n *node, parentLayout string) error {
	if panel == nil || n == nil {
		return nil
	}
	textIndex := 0
	for _, child := range n.Children {
		switch child.Kind {
		case nodeText:
			text := normalizeInlineText(child.Text)
			if text == "" {
				continue
			}
			textIndex++
			labelNode := &node{
				Kind:  nodeElement,
				Tag:   "label",
				Attrs: map[string]string{"id": fmt.Sprintf("%s-text-%d", panel.ID(), textIndex)},
				Children: []*node{{
					Kind: nodeText,
					Text: text,
					Pos:  child.Pos,
				}},
				Pos: child.Pos,
			}
			widget, err := b.buildLabel(labelNode)
			if err != nil {
				return err
			}
			if err := b.applyLayoutData(widget, labelNode, parentLayout); err != nil {
				return err
			}
			panel.Add(widget)
		case nodeElement:
			if child.Tag == "option" {
				return newParseError("builder", child.Pos, child.inlineContext(), "option can only appear inside select/listbox")
			}
			widget, err := b.buildNode(child)
			if err != nil {
				return err
			}
			if err := b.applyLayoutData(widget, child, parentLayout); err != nil {
				return err
			}
			panel.Add(widget)
		}
	}
	return nil
}

func (b *uiBuilder) buildLabel(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	label := widgets.NewLabel(b.nodeID(n), normalizeInlineText(n.textContent()))
	label.SetStyle(b.textStyle(n, false))
	if err := b.applyCommonWidgetState(label, n); err != nil {
		return nil, err
	}
	defaultSize := measureTextBox(label.Text, b.textStyle(n, false).Font, 12, 24)
	if err := b.applyPreferredSize(label, n, defaultSize); err != nil {
		return nil, err
	}
	return label, nil
}

func (b *uiBuilder) buildButton(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	button := widgets.NewButton(b.nodeID(n), normalizeInlineText(n.textContent()), b.mode())
	button.SetStyle(b.buttonStyle(n))
	if err := b.applyCommonWidgetState(button, n); err != nil {
		return nil, err
	}
	if iconPath := strings.TrimSpace(n.attr("icon")); iconPath != "" {
		icon, err := b.loadICO(iconPath, n.Pos, n.inlineContext(), "button icon")
		if err != nil {
			return nil, err
		}
		button.SetIcon(icon)
	}
	if position := strings.ToLower(strings.TrimSpace(n.attr("icon-position"))); position != "" {
		switch position {
		case "auto":
			button.SetKind(widgets.BtnAuto)
		case "left":
			button.SetKind(widgets.BtnLeft)
		case "top":
			button.SetKind(widgets.BtnTop)
		default:
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "invalid icon-position %q", position)
		}
	}
	if actionName, err := b.actionName(n, "onclick"); err != nil {
		return nil, err
	} else if actionName != "" {
		button.SetOnClick(func() {
			ctx := b.baseActionContext(actionName, button)
			ctx.Value = button.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := measureTextBox(button.Text, b.buttonStyle(n).Font, 28, 36)
	if defaultSize.Width < 80 {
		defaultSize.Width = 80
	}
	if err := b.applyPreferredSize(button, n, defaultSize); err != nil {
		return nil, err
	}
	return button, nil
}

func (b *uiBuilder) buildInput(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	inputType := strings.ToLower(strings.TrimSpace(n.attr("type")))
	if inputType == "" {
		inputType = "text"
	}
	switch inputType {
	case "text", "password", "search", "file":
	default:
		return nil, newParseError("builder", n.Pos, n.inlineContext(), "unsupported input type %q", inputType)
	}
	if inputType == "file" {
		return b.buildFileInput(n)
	}
	edit := widgets.NewEditBox(b.nodeID(n), b.mode())
	edit.SetMultiline(false)
	if inputType == "password" {
		edit.SetPassword(true)
	}
	edit.SetStyle(b.editStyle(n))
	edit.SetPlaceholder(n.attr("placeholder"))
	edit.SetReadOnly(b.boolAttrOrStyle(n, "readonly", "readonly"))
	edit.SetText(n.attr("value"))
	if err := b.applyCommonWidgetState(edit, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		edit.SetOnChange(func(string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName, err := b.actionName(n, "onsubmit"); err != nil {
		return nil, err
	} else if actionName != "" {
		edit.SetOnSubmit(func(string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := core.Size{Width: 220, Height: 36}
	if err := b.applyPreferredSize(edit, n, defaultSize); err != nil {
		return nil, err
	}
	return edit, nil
}

func (b *uiBuilder) buildTextArea(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	edit := widgets.NewEditBox(b.nodeID(n), b.mode())
	edit.SetMultiline(true)
	wordWrap := b.boolStyleDefault(n, "word-wrap", true)
	horizontalScroll, verticalScroll := b.textAreaScrollFlags(n)
	if horizontalScroll && strings.TrimSpace(n.Styles["word-wrap"]) == "" {
		wordWrap = false
	}
	edit.SetWordWrap(wordWrap)
	edit.SetVerticalScroll(verticalScroll)
	edit.SetHorizontalScroll(horizontalScroll)
	edit.SetAcceptReturn(true)
	edit.SetStyle(b.editStyle(n))
	edit.SetPlaceholder(n.attr("placeholder"))
	edit.SetReadOnly(b.boolAttrOrStyle(n, "readonly", "readonly"))
	if multiline, ok, _ := parseBoolValue(n.Styles["multiline"]); ok {
		edit.SetMultiline(multiline)
	}
	value := n.attr("value")
	if value == "" {
		value = normalizeBlockText(n.textContent())
	}
	edit.SetText(value)
	if err := b.applyCommonWidgetState(edit, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		edit.SetOnChange(func(string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName, err := b.actionName(n, "onsubmit"); err != nil {
		return nil, err
	} else if actionName != "" {
		edit.SetOnSubmit(func(string) {
			ctx := b.baseActionContext(actionName, edit)
			ctx.Value = edit.TextValue()
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := core.Size{Width: 320, Height: 96}
	if err := b.applyPreferredSize(edit, n, defaultSize); err != nil {
		return nil, err
	}
	return edit, nil
}

func (b *uiBuilder) buildImage(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	img := widgets.NewImage(b.nodeID(n))
	if fit, ok, err := parseObjectFitValue(n.Styles["object-fit"]); err != nil {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		switch fit {
		case "contain":
			img.SetScaleMode(widgets.ImageScaleContain)
		case "fill", "cover":
			img.SetScaleMode(widgets.ImageScaleStretch)
		}
	}
	if src := strings.TrimSpace(n.attr("src")); src != "" {
		path := b.resolveAssetPath(src)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "load image %q failed: %v", src, err)
		}
		if err := img.LoadBytes(data); err != nil {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "decode image %q failed: %v", src, err)
		}
	}
	if err := b.applyCommonWidgetState(img, n); err != nil {
		return nil, err
	}
	defaultSize := core.Size{Width: 160, Height: 120}
	if err := b.applyPreferredSize(img, n, defaultSize); err != nil {
		return nil, err
	}
	return img, nil
}

func (b *uiBuilder) buildProgress(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	progress := widgets.NewProgressBar(b.nodeID(n))
	progress.SetStyle(b.progressStyle(n))
	if value, ok, err := parseIntegerValue(n.attr("value")); err != nil {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		progress.SetValue(int32(value))
	}
	if err := b.applyCommonWidgetState(progress, n); err != nil {
		return nil, err
	}
	defaultSize := core.Size{Width: 220, Height: 24}
	if err := b.applyPreferredSize(progress, n, defaultSize); err != nil {
		return nil, err
	}
	return progress, nil
}

func (b *uiBuilder) buildCheckBox(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	check := widgets.NewCheckBox(b.nodeID(n), normalizeInlineText(n.textContent()), b.mode())
	check.SetStyle(b.choiceStyle(n))
	check.SetChecked(b.boolAttrOrStyle(n, "checked", "checked"))
	if err := b.applyCommonWidgetState(check, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		check.SetOnChange(func(checked bool) {
			ctx := b.baseActionContext(actionName, check)
			ctx.Checked = checked
			ctx.Value = check.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := measureTextBox(check.Text, widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 16}, 28, 28)
	if defaultSize.Width < 120 {
		defaultSize.Width = 120
	}
	if err := b.applyPreferredSize(check, n, defaultSize); err != nil {
		return nil, err
	}
	return check, nil
}

func (b *uiBuilder) buildRadio(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	radio := widgets.NewRadioButton(b.nodeID(n), normalizeInlineText(n.textContent()), b.mode())
	radio.SetStyle(b.choiceStyle(n))
	if group := strings.TrimSpace(n.attr("name")); group != "" {
		radio.SetGroup(group)
	}
	radio.SetChecked(b.boolAttrOrStyle(n, "checked", "checked"))
	if err := b.applyCommonWidgetState(radio, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		radio.SetOnChange(func(checked bool) {
			ctx := b.baseActionContext(actionName, radio)
			ctx.Checked = checked
			ctx.Value = radio.Text
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := measureTextBox(radio.Text, widgets.FontSpec{Face: "Microsoft YaHei UI", SizeDP: 16}, 28, 28)
	if defaultSize.Width < 120 {
		defaultSize.Width = 120
	}
	if err := b.applyPreferredSize(radio, n, defaultSize); err != nil {
		return nil, err
	}
	return radio, nil
}

func (b *uiBuilder) buildSelect(n *node) (widgets.Widget, error) {
	combo := widgets.NewComboBox(b.nodeID(n), b.mode())
	combo.SetStyle(b.comboStyle(n))
	combo.SetPlaceholder(n.attr("placeholder"))
	items := make([]widgets.ListItem, 0)
	selected := -1
	selectedValue := strings.TrimSpace(n.attr("value"))
	for _, child := range n.Children {
		if child.Kind == nodeText {
			if strings.TrimSpace(child.Text) == "" {
				continue
			}
			return nil, newParseError("builder", child.Pos, n.inlineContext(), "text is not allowed directly inside select")
		}
		if child.Tag != "option" {
			return nil, newParseError("builder", child.Pos, child.inlineContext(), "select only accepts option children")
		}
		itemText := normalizeInlineText(child.textContent())
		itemValue := child.attr("value")
		if itemValue == "" {
			itemValue = itemText
		}
		items = append(items, widgets.ListItem{Value: itemValue, Text: itemText})
		if selected == -1 {
			if selectedValue != "" && itemValue == selectedValue {
				selected = len(items) - 1
			}
			if selectedValue == "" && child.hasAttr("selected") {
				selected = len(items) - 1
			}
		}
	}
	combo.SetItems(items)
	if selected >= 0 {
		combo.SetSelected(selected)
	}
	if err := b.applyCommonWidgetState(combo, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		combo.SetOnChange(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, combo)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := core.Size{Width: 220, Height: 36}
	if err := b.applyPreferredSize(combo, n, defaultSize); err != nil {
		return nil, err
	}
	return combo, nil
}

func (b *uiBuilder) buildListBox(n *node) (widgets.Widget, error) {
	list := widgets.NewListBox(b.nodeID(n))
	list.SetStyle(b.listStyle(n))
	items := make([]widgets.ListItem, 0)
	selected := -1
	selectedValue := strings.TrimSpace(n.attr("value"))
	for _, child := range n.Children {
		if child.Kind == nodeText {
			if strings.TrimSpace(child.Text) == "" {
				continue
			}
			return nil, newParseError("builder", child.Pos, n.inlineContext(), "text is not allowed directly inside listbox")
		}
		if child.Tag != "option" {
			return nil, newParseError("builder", child.Pos, child.inlineContext(), "listbox only accepts option children")
		}
		itemText := normalizeInlineText(child.textContent())
		itemValue := child.attr("value")
		if itemValue == "" {
			itemValue = itemText
		}
		items = append(items, widgets.ListItem{Value: itemValue, Text: itemText})
		if selected == -1 {
			if selectedValue != "" && itemValue == selectedValue {
				selected = len(items) - 1
			}
			if selectedValue == "" && child.hasAttr("selected") {
				selected = len(items) - 1
			}
		}
	}
	list.SetItems(items)
	if selected >= 0 {
		list.SetSelected(selected)
	}
	if err := b.applyCommonWidgetState(list, n); err != nil {
		return nil, err
	}
	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		list.SetOnChange(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, list)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	if actionName, err := b.actionName(n, "onactivate"); err != nil {
		return nil, err
	} else if actionName != "" {
		list.SetOnActivate(func(index int, item widgets.ListItem) {
			ctx := b.baseActionContext(actionName, list)
			ctx.Index = index
			ctx.Item = item
			ctx.Value = item.Value
			b.dispatchAction(actionName, ctx)
		})
	}
	defaultSize := core.Size{Width: 220, Height: 160}
	if err := b.applyPreferredSize(list, n, defaultSize); err != nil {
		return nil, err
	}
	return list, nil
}

func (b *uiBuilder) buildAnimatedImage(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	animated := widgets.NewAnimatedImage(b.nodeID(n))
	if fit, ok, err := parseObjectFitValue(n.Styles["object-fit"]); err != nil {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		switch fit {
		case "contain":
			animated.SetScaleMode(widgets.ImageScaleContain)
		case "fill", "cover":
			animated.SetScaleMode(widgets.ImageScaleStretch)
		}
	}
	autoplay := true
	if raw := strings.TrimSpace(n.attr("autoplay")); raw != "" {
		parsed, ok, err := parseBoolValue(raw)
		if err != nil || !ok {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "invalid autoplay value %q", raw)
		}
		autoplay = parsed
	}
	if src := strings.TrimSpace(n.attr("src")); src != "" {
		if strings.ToLower(filepath.Ext(src)) != ".gif" {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "animated-img src %q must be a local .gif file", src)
		}
		path := b.resolveAssetPath(src)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "load image %q failed: %v", src, err)
		}
		if err := animated.LoadGIF(data); err != nil {
			return nil, newParseError("builder", n.Pos, n.inlineContext(), "decode gif %q failed: %v", src, err)
		}
	}
	animated.SetPlaying(autoplay)
	if err := b.applyCommonWidgetState(animated, n); err != nil {
		return nil, err
	}
	defaultSize := core.Size{Width: 64, Height: 64}
	if err := b.applyPreferredSize(animated, n, defaultSize); err != nil {
		return nil, err
	}
	return animated, nil
}

func (b *uiBuilder) mode() widgets.ControlMode {
	return b.opts.DefaultMode
}

func (b *uiBuilder) nodeID(n *node) string {
	if n == nil {
		b.autoIDSeed++
		return fmt.Sprintf("widget-%d", b.autoIDSeed)
	}
	if id := strings.TrimSpace(n.attr("id")); id != "" {
		return id
	}
	b.autoIDSeed++
	return fmt.Sprintf("%s-%d", n.Tag, b.autoIDSeed)
}

func (b *uiBuilder) ensureNoElementChildren(n *node) error {
	for _, child := range n.Children {
		if child.Kind == nodeElement {
			return newParseError("builder", child.Pos, n.inlineContext(), "<%s> does not accept child elements", n.Tag)
		}
	}
	return nil
}

func (b *uiBuilder) actionName(n *node, attr string) (string, error) {
	name := strings.TrimSpace(n.attr(attr))
	if name == "" {
		return "", nil
	}
	if b.opts.ActionHandlers[name] != nil || b.opts.Actions[name] != nil {
		return name, nil
	}
	return "", newParseError("builder", n.Pos, n.inlineContext(), "action %q referenced by %s was not provided", name, attr)
}

func (b *uiBuilder) dispatchAction(name string, ctx ActionContext) {
	if name == "" {
		return
	}
	if ctx.Name == "" {
		ctx.Name = name
	}
	if ctx.Index == 0 && ctx.Item.Value == "" && ctx.Item.Text == "" {
		ctx.Index = -1
	}
	if ctx.Widget != nil && ctx.ID == "" {
		ctx.ID = ctx.Widget.ID()
	}
	if handler := b.opts.ActionHandlers[name]; handler != nil {
		handler(ctx)
		return
	}
	if action := b.opts.Actions[name]; action != nil {
		action()
	}
}

func (b *uiBuilder) baseActionContext(name string, widget widgets.Widget) ActionContext {
	ctx := ActionContext{Name: name, Widget: widget, Index: -1}
	if widget != nil {
		ctx.ID = widget.ID()
	}
	return ctx
}

func (b *uiBuilder) resolveAssetPath(path string) string {
	resolved := strings.TrimSpace(path)
	if resolved == "" {
		return ""
	}
	if filepath.IsAbs(resolved) || b.opts.AssetsDir == "" {
		return resolved
	}
	return filepath.Join(b.opts.AssetsDir, resolved)
}

func (b *uiBuilder) loadICO(src string, pos position, context string, usage string) (*core.Icon, error) {
	if strings.ToLower(filepath.Ext(src)) != ".ico" {
		return nil, newParseError("builder", pos, context, "%s %q must be a local .ico file", usage, src)
	}
	path := b.resolveAssetPath(src)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, newParseError("builder", pos, context, "load icon %q failed: %v", src, err)
	}
	icon, err := core.LoadIconFromICO(data, 32)
	if err != nil {
		return nil, newParseError("builder", pos, context, "decode icon %q failed: %v", src, err)
	}
	return icon, nil
}

func (b *uiBuilder) bindOnClickIfAny(widget widgets.Widget, n *node) error {
	actionName, err := b.actionName(n, "onclick")
	if err != nil || actionName == "" {
		return err
	}
	switch typed := widget.(type) {
	case *widgets.Panel:
		typed.SetOnClick(func() {
			ctx := b.baseActionContext(actionName, typed)
			b.dispatchAction(actionName, ctx)
		})
		return nil
	case *widgets.Button:
		typed.SetOnClick(func() {
			ctx := b.baseActionContext(actionName, typed)
			ctx.Value = typed.Text
			b.dispatchAction(actionName, ctx)
		})
		return nil
	default:
		return newParseError("builder", n.Pos, n.inlineContext(), "onclick is not supported by <%s>", n.Tag)
	}
}

func (b *uiBuilder) applyCommonWidgetState(widget widgets.Widget, n *node) error {
	if widget == nil || n == nil {
		return nil
	}
	visible := true
	if display, ok, err := parseDisplayValue(n.Styles["display"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && display == "none" {
		visible = false
	}
	if parsed, ok, err := parseBoolValue(n.Styles["visible"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		visible = parsed
	}
	enabled := true
	if parsed, ok, err := parseBoolValue(n.Styles["enabled"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		enabled = parsed
	}
	widget.SetVisible(visible)
	widget.SetEnabled(enabled)
	return nil
}

func (b *uiBuilder) applyPreferredSize(widget widgets.Widget, n *node, defaultSize core.Size) error {
	width := defaultSize.Width
	height := defaultSize.Height
	if parsed, ok, err := parseLengthValue(n.Styles["width"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		width = parsed
	}
	if parsed, ok, err := parseLengthValue(n.Styles["height"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		height = parsed
	}
	if minWidth, ok, err := parseLengthValue(n.Styles["min-width"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && width < minWidth {
		width = minWidth
	}
	if minHeight, ok, err := parseLengthValue(n.Styles["min-height"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && height < minHeight {
		height = minHeight
	}
	if maxWidth, ok, err := parseLengthValue(n.Styles["max-width"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && maxWidth > 0 && (width == 0 || width > maxWidth) {
		width = maxWidth
	}
	if maxHeight, ok, err := parseLengthValue(n.Styles["max-height"]); err != nil {
		return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && maxHeight > 0 && (height == 0 || height > maxHeight) {
		height = maxHeight
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	if width == 0 && height == 0 {
		return nil
	}
	widget.SetBounds(widgets.Rect{W: width, H: height})
	widgets.SetPreferredSize(widget, core.Size{Width: width, Height: height})
	return nil
}

func (b *uiBuilder) layoutForNode(n *node) (string, widgets.Layout, error) {
	display := b.defaultDisplay(n)
	if parsed, ok, err := parseDisplayValue(n.Styles["display"]); err != nil {
		return "", nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok && parsed != "none" {
		display = parsed
	}
	padding, err := b.paddingForNode(n)
	if err != nil {
		return "", nil, err
	}
	gap, err := b.lengthStyle(n, "gap")
	if err != nil {
		return "", nil, err
	}
	rowGap, err := b.lengthStyle(n, "row-gap")
	if err != nil {
		return "", nil, err
	}
	columnGap, err := b.lengthStyle(n, "column-gap")
	if err != nil {
		return "", nil, err
	}
	columns, err := b.intStyle(n, "layout-columns")
	if err != nil {
		return "", nil, err
	}
	switch display {
	case "absolute":
		return "absolute", widgets.AbsoluteLayout{}, nil
	case "row":
		return "row", widgets.RowLayout{Gap: gap, Padding: padding}, nil
	case "grid":
		if columns <= 0 {
			columns = 2
		}
		return "grid", widgets.GridLayout{Columns: columns, Gap: gap, RowGap: rowGap, ColumnGap: columnGap, Padding: padding}, nil
	case "form":
		return "form", widgets.FormLayout{Padding: padding, RowGap: rowGap, ColumnGap: columnGap}, nil
	case "column", "none":
		fallthrough
	default:
		return "column", widgets.ColumnLayout{Gap: gap, Padding: padding}, nil
	}
}

func (b *uiBuilder) defaultDisplay(n *node) string {
	switch n.Tag {
	case "row":
		return "row"
	case "form":
		return "form"
	case "body", "div", "section", "panel", "column", "scroll":
		return "column"
	default:
		return "column"
	}
}

func (b *uiBuilder) textAreaScrollFlags(n *node) (bool, bool) {
	horizontal := false
	vertical := true
	if overflow, ok, _ := parseOverflowValue(n.Styles["overflow"]); ok {
		switch overflow {
		case "hidden", "visible":
			horizontal = false
			vertical = false
		case "auto", "scroll":
			vertical = true
		}
	}
	if overflowX, ok, _ := parseOverflowValue(n.Styles["overflow-x"]); ok {
		horizontal = overflowX == "auto" || overflowX == "scroll"
	}
	if overflowY, ok, _ := parseOverflowValue(n.Styles["overflow-y"]); ok {
		vertical = overflowY == "auto" || overflowY == "scroll"
	}
	return horizontal, vertical
}

func (b *uiBuilder) scrollConfig(n *node) (bool, bool, bool, error) {
	if n.Tag == "scroll" {
		allowX := false
		allowY := true
		if overflowX, ok, err := parseOverflowValue(n.Styles["overflow-x"]); err != nil {
			return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			allowX = overflowX == "auto" || overflowX == "scroll"
		}
		if overflowY, ok, err := parseOverflowValue(n.Styles["overflow-y"]); err != nil {
			return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			allowY = overflowY == "auto" || overflowY == "scroll"
		}
		if overflow, ok, err := parseOverflowValue(n.Styles["overflow"]); err != nil {
			return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			switch overflow {
			case "hidden", "visible":
				allowX = false
				allowY = false
			case "auto", "scroll":
				allowY = true
			}
		}
		return true, allowX, allowY, nil
	}
	allowX := false
	allowY := false
	used := false
	if overflow, ok, err := parseOverflowValue(n.Styles["overflow"]); err != nil {
		return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		used = overflow != "visible"
		allowX = overflow == "auto" || overflow == "scroll"
		allowY = overflow == "auto" || overflow == "scroll"
	}
	if overflowX, ok, err := parseOverflowValue(n.Styles["overflow-x"]); err != nil {
		return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		used = used || overflowX != "visible"
		allowX = overflowX == "auto" || overflowX == "scroll"
	}
	if overflowY, ok, err := parseOverflowValue(n.Styles["overflow-y"]); err != nil {
		return false, false, false, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		used = used || overflowY != "visible"
		allowY = overflowY == "auto" || overflowY == "scroll"
	}
	return used, allowX, allowY, nil
}

func (b *uiBuilder) paddingForNode(n *node) (widgets.Insets, error) {
	padding, _, err := parseInsetsValue(n.Styles["padding"])
	if err != nil {
		return widgets.Insets{}, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	}
	if top, ok, err := parseLengthValue(n.Styles["padding-top"]); err != nil {
		return widgets.Insets{}, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		padding.Top = top
	}
	if right, ok, err := parseLengthValue(n.Styles["padding-right"]); err != nil {
		return widgets.Insets{}, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		padding.Right = right
	}
	if bottom, ok, err := parseLengthValue(n.Styles["padding-bottom"]); err != nil {
		return widgets.Insets{}, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		padding.Bottom = bottom
	}
	if left, ok, err := parseLengthValue(n.Styles["padding-left"]); err != nil {
		return widgets.Insets{}, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		padding.Left = left
	}
	return padding, nil
}

func (b *uiBuilder) lengthStyle(n *node, key string) (int32, error) {
	if value, ok, err := parseLengthValue(n.Styles[key]); err != nil {
		return 0, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		return value, nil
	}
	return 0, nil
}

func (b *uiBuilder) intStyle(n *node, key string) (int, error) {
	if value, ok, err := parseIntegerValue(n.Styles[key]); err != nil {
		return 0, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	} else if ok {
		return value, nil
	}
	return 0, nil
}

func (b *uiBuilder) textStyle(n *node, multiline bool) widgets.TextStyle {
	style := widgets.TextStyle{Format: parseTextFormat(b.alignmentStyle(n), multiline)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.Color = color
	}
	style.Font = b.fontSpec(n)
	return style
}

func (b *uiBuilder) buttonStyle(n *node) widgets.ButtonStyle {
	style := widgets.ButtonStyle{TextAlign: b.alignmentStyle(n), Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if downText, ok, _ := parseColorValue(n.Styles["down-text-color"]); ok {
		style.DownText = downText
	}
	if disabledText, ok, _ := parseColorValue(n.Styles["disabled-text-color"]); ok {
		style.DisabledText = disabledText
	}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if hover, ok, _ := parseColorValue(n.Styles["hover-background"]); ok {
		style.Hover = hover
	}
	if pressed, ok, _ := parseColorValue(n.Styles["pressed-background"]); ok {
		style.Pressed = pressed
	}
	if disabled, ok, _ := parseColorValue(n.Styles["disabled-background"]); ok {
		style.Disabled = disabled
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.Border = border
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if iconSize, ok, _ := parseLengthValue(n.Styles["icon-size"]); ok {
		style.IconSizeDP = iconSize
	}
	if inset, ok, _ := parseLengthValue(n.Styles["text-inset"]); ok {
		style.TextInsetDP = inset
	}
	if gap, ok, _ := parseLengthValue(n.Styles["gap"]); ok {
		style.GapDP = gap
	}
	if padding, ok, _ := parseLengthValue(n.Styles["padding"]); ok {
		style.PadDP = padding
	}
	return style
}

func (b *uiBuilder) progressStyle(n *node) widgets.ProgressStyle {
	style := widgets.ProgressStyle{Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if track, ok, _ := parseColorValue(n.Styles["track-color"]); ok {
		style.TrackColor = track
	}
	if fill, ok, _ := parseColorValue(n.Styles["fill-color"]); ok {
		style.FillColor = fill
	}
	if bubble, ok, _ := parseColorValue(n.Styles["bubble-color"]); ok {
		style.BubbleColor = bubble
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if showPercent, ok, _ := parseBoolValue(n.Styles["show-percent"]); ok {
		style.ShowPercent = showPercent
	}
	return style
}

func (b *uiBuilder) choiceStyle(n *node) widgets.ChoiceStyle {
	style := widgets.ChoiceStyle{Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if disabledText, ok, _ := parseColorValue(n.Styles["disabled-text-color"]); ok {
		style.DisabledText = disabledText
	}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.BorderColor = border
	}
	if hoverBorder, ok, _ := parseColorValue(n.Styles["hover-border-color"]); ok {
		style.HoverBorder = hoverBorder
	}
	if focusBorder, ok, _ := parseColorValue(n.Styles["focus-border-color"]); ok {
		style.FocusBorder = focusBorder
	}
	if indicator, ok, _ := parseColorValue(n.Styles["indicator-color"]); ok {
		style.IndicatorColor = indicator
	}
	if checkColor, ok, _ := parseColorValue(n.Styles["check-color"]); ok {
		style.CheckColor = checkColor
	}
	if indicatorStyle, ok, _ := parseChoiceIndicatorStyle(n.Styles["indicator-style"]); ok {
		style.IndicatorStyle = indicatorStyle
	}
	if hoverBg, ok, _ := parseColorValue(n.Styles["hover-background"]); ok {
		style.HoverBackground = hoverBg
	}
	if disabledBg, ok, _ := parseColorValue(n.Styles["disabled-background"]); ok {
		style.DisabledBg = disabledBg
	}
	if disabledBorder, ok, _ := parseColorValue(n.Styles["disabled-border-color"]); ok {
		style.DisabledBorder = disabledBorder
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if indicatorSize, ok, _ := parseLengthValue(n.Styles["indicator-size"]); ok {
		style.IndicatorSizeDP = indicatorSize
	}
	if indicatorGap, ok, _ := parseLengthValue(n.Styles["indicator-gap"]); ok {
		style.IndicatorGapDP = indicatorGap
	}
	return style
}

func (b *uiBuilder) comboStyle(n *node) widgets.ComboStyle {
	style := widgets.ComboStyle{Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if placeholder, ok, _ := parseColorValue(n.Styles["placeholder-color"]); ok {
		style.PlaceholderColor = placeholder
	}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.BorderColor = border
	}
	if hoverBorder, ok, _ := parseColorValue(n.Styles["hover-border-color"]); ok {
		style.HoverBorder = hoverBorder
	}
	if focusBorder, ok, _ := parseColorValue(n.Styles["focus-border-color"]); ok {
		style.FocusBorder = focusBorder
	}
	if arrow, ok, _ := parseColorValue(n.Styles["arrow-color"]); ok {
		style.ArrowColor = arrow
	}
	if popupBg, ok, _ := parseColorValue(n.Styles["popup-background"]); ok {
		style.PopupBackground = popupBg
	}
	if itemHover, ok, _ := parseColorValue(n.Styles["item-hover-color"]); ok {
		style.ItemHoverColor = itemHover
	}
	if itemSelected, ok, _ := parseColorValue(n.Styles["item-selected-color"]); ok {
		style.ItemSelectedColor = itemSelected
	}
	if itemText, ok, _ := parseColorValue(n.Styles["item-text-color"]); ok {
		style.ItemTextColor = itemText
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if padding, ok, _ := parseLengthValue(n.Styles["padding"]); ok {
		style.PaddingDP = padding
	}
	if itemHeight, ok, _ := parseLengthValue(n.Styles["item-height"]); ok {
		style.ItemHeightDP = itemHeight
	}
	if maxVisibleItems, ok, _ := parseIntegerValue(n.Styles["max-visible-items"]); ok {
		style.MaxVisibleItems = int32(maxVisibleItems)
	}
	return style
}

func (b *uiBuilder) listStyle(n *node) widgets.ListStyle {
	style := widgets.ListStyle{Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if disabledText, ok, _ := parseColorValue(n.Styles["disabled-text-color"]); ok {
		style.DisabledText = disabledText
	}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.BorderColor = border
	}
	if hoverBorder, ok, _ := parseColorValue(n.Styles["hover-border-color"]); ok {
		style.HoverBorder = hoverBorder
	}
	if focusBorder, ok, _ := parseColorValue(n.Styles["focus-border-color"]); ok {
		style.FocusBorder = focusBorder
	}
	if hover, ok, _ := parseColorValue(n.Styles["item-hover-color"]); ok {
		style.ItemHoverColor = hover
	}
	if selected, ok, _ := parseColorValue(n.Styles["item-selected-color"]); ok {
		style.ItemSelectedColor = selected
	}
	if selectedText, ok, _ := parseColorValue(n.Styles["item-text-color"]); ok {
		style.ItemTextColor = selectedText
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if padding, ok, _ := parseLengthValue(n.Styles["padding"]); ok {
		style.PaddingDP = padding
	}
	if itemHeight, ok, _ := parseLengthValue(n.Styles["item-height"]); ok {
		style.ItemHeightDP = itemHeight
	}
	return style
}

func (b *uiBuilder) editStyle(n *node) widgets.EditStyle {
	style := widgets.EditStyle{TextAlign: b.alignmentStyle(n), Font: b.fontSpec(n)}
	if color, ok, _ := parseColorValue(n.Styles["color"]); ok {
		style.TextColor = color
	}
	if placeholder, ok, _ := parseColorValue(n.Styles["placeholder-color"]); ok {
		style.PlaceholderColor = placeholder
	}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.BorderColor = border
	}
	if hoverBorder, ok, _ := parseColorValue(n.Styles["hover-border-color"]); ok {
		style.HoverBorder = hoverBorder
	}
	if focusBorder, ok, _ := parseColorValue(n.Styles["focus-border-color"]); ok {
		style.FocusBorder = focusBorder
	}
	if disabledText, ok, _ := parseColorValue(n.Styles["disabled-text-color"]); ok {
		style.DisabledText = disabledText
	}
	if disabledBg, ok, _ := parseColorValue(n.Styles["disabled-background"]); ok {
		style.DisabledBg = disabledBg
	}
	if caret, ok, _ := parseColorValue(n.Styles["caret-color"]); ok {
		style.CaretColor = caret
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	if padding, ok, _ := parseLengthValue(n.Styles["padding"]); ok {
		style.PaddingDP = padding
	}
	return style
}

func (b *uiBuilder) panelStyle(n *node) widgets.PanelStyle {
	style := widgets.PanelStyle{}
	if bg, ok, _ := parseColorValue(n.Styles["background"]); ok {
		style.Background = bg
	}
	if border, ok, _ := parseColorValue(n.Styles["border-color"]); ok {
		style.BorderColor = border
	}
	if width, ok, _ := parseLengthValue(n.Styles["border-width"]); ok {
		style.BorderWidth = width
	}
	if radius, ok, _ := parseLengthValue(n.Styles["border-radius"]); ok {
		style.CornerRadius = radius
	}
	return style
}

func (b *uiBuilder) alignmentStyle(n *node) widgets.Alignment {
	if align, ok, _ := parseAlignmentValue(n.Styles["text-align"]); ok {
		return align
	}
	return widgets.AlignDefault
}

func (b *uiBuilder) fontSpec(n *node) widgets.FontSpec {
	var font widgets.FontSpec
	if face := strings.TrimSpace(n.Styles["font-family"]); face != "" {
		font.Face = face
	}
	if size, ok, _ := parseLengthValue(n.Styles["font-size"]); ok {
		font.SizeDP = size
	}
	if weight, ok, _ := parseFontWeightValue(n.Styles["font-weight"]); ok {
		font.Weight = weight
	}
	return font
}

func (b *uiBuilder) boolAttrOrStyle(n *node, attrName string, styleName string) bool {
	if styleValue := strings.TrimSpace(n.Styles[styleName]); styleValue != "" {
		if parsed, ok, _ := parseBoolValue(styleValue); ok {
			return parsed
		}
	}
	if n.hasAttr(attrName) {
		if value := strings.TrimSpace(n.attr(attrName)); value != "" && value != "true" {
			if parsed, ok, _ := parseBoolValue(value); ok {
				return parsed
			}
		}
		return true
	}
	return false
}

func (b *uiBuilder) boolStyleDefault(n *node, key string, fallback bool) bool {
	if parsed, ok, _ := parseBoolValue(n.Styles[key]); ok {
		return parsed
	}
	return fallback
}

func (b *uiBuilder) applyLayoutData(widget widgets.Widget, n *node, parentLayout string) error {
	if widget == nil || n == nil {
		return nil
	}
	if parentLayout == "absolute" {
		var data widgets.AbsoluteLayoutData
		if left, ok, err := parseLengthValue(n.Styles["left"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Left = left
			data.HasLeft = true
		} else if x, ok, err := parseLengthValue(n.Styles["x"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Left = x
			data.HasLeft = true
		}
		if top, ok, err := parseLengthValue(n.Styles["top"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Top = top
			data.HasTop = true
		} else if y, ok, err := parseLengthValue(n.Styles["y"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Top = y
			data.HasTop = true
		}
		if right, ok, err := parseLengthValue(n.Styles["right"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Right = right
			data.HasRight = true
		}
		if bottom, ok, err := parseLengthValue(n.Styles["bottom"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Bottom = bottom
			data.HasBottom = true
		}
		if width, ok, err := parseLengthValue(n.Styles["width"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Width = width
			data.HasWidth = true
		}
		if height, ok, err := parseLengthValue(n.Styles["height"]); err != nil {
			return newParseError("builder", n.Pos, n.inlineContext(), err.Error())
		} else if ok {
			data.Height = height
			data.HasHeight = true
		}
		widget.SetLayoutData(data)
		return nil
	}
	grow, _, _ := parseIntegerValue(n.Styles["flex-grow"])
	align, _, _ := parseAlignmentValue(n.Styles["align-self"])
	columnSpan, _, _ := parseIntegerValue(n.Styles["column-span"])
	rowSpan, _, _ := parseIntegerValue(n.Styles["row-span"])
	if align == widgets.AlignDefault {
		align = widgets.AlignStretch
	}
	var data any
	switch parentLayout {
	case "row", "column":
		data = widgets.FlexLayoutData{Grow: int32(grow), Align: align}
	case "form":
		data = widgets.FormLayoutData{Grow: int32(grow), Align: align}
	case "grid":
		grid := widgets.GridLayoutData{HorizontalAlign: align, VerticalAlign: align}
		if columnSpan > 0 {
			grid.ColumnSpan = columnSpan
		}
		if rowSpan > 0 {
			grid.RowSpan = rowSpan
		}
		data = grid
	default:
		return nil
	}
	widget.SetLayoutData(data)
	return nil
}

func (b *uiBuilder) nonWhitespaceTextChildren(n *node) []*node {
	children := make([]*node, 0)
	for _, child := range n.Children {
		if child.Kind == nodeText && strings.TrimSpace(child.Text) != "" {
			children = append(children, child)
		}
	}
	return children
}

func normalizeInlineText(text string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(text), " "))
}

func normalizeBlockText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func measureTextBox(text string, font widgets.FontSpec, horizontalPad int32, height int32) core.Size {
	text = normalizeInlineText(text)
	if text == "" {
		text = "M"
	}
	size := font.SizeDP
	if size <= 0 {
		size = 16
	}
	width := int32(len([]rune(text)))*max32(6, size*3/5) + horizontalPad
	if width < horizontalPad+24 {
		width = horizontalPad + 24
	}
	if height <= 0 {
		height = size + 12
	}
	return core.Size{Width: width, Height: height}
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

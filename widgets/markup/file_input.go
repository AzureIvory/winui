//go:build windows

package markup

import (
	"fmt"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/dialogs"
	"github.com/AzureIvory/winui/widgets"
)

func (b *uiBuilder) buildFileInput(n *node) (widgets.Widget, error) {
	if err := b.ensureNoElementChildren(n); err != nil {
		return nil, err
	}
	mode, err := parseDialogModeAttr(n.attr("dialog"))
	if err != nil {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	}
	multiple := b.boolAttrOrStyle(n, "multiple", "multiple")
	if mode != dialogs.DialogOpen && multiple {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), "multiple is only supported when dialog=open")
	}

	filters, err := parseDialogFilters(n.attr("filters"), n.attr("accept"))
	if err != nil {
		return nil, newParseError("builder", n.Pos, n.inlineContext(), err.Error())
	}

	picker := widgets.NewFilePicker(b.nodeID(n), b.mode())
	picker.SetFieldStyle(b.editStyle(n))

	if buttonNode := b.prefixedStyleNode(n, "button-"); buttonNode != nil {
		picker.SetButtonStyle(b.buttonStyle(buttonNode))
	}
	if pickerNode := b.prefixedStyleNode(n, "picker-"); pickerNode != nil {
		picker.SetStyle(b.panelStyle(pickerNode))
	}

	initialPath := strings.TrimSpace(n.attr("initial-path"))
	if initialPath == "" {
		initialPath = strings.TrimSpace(n.attr("value"))
	}
	picker.SetDialogOptions(dialogs.Options{
		Mode:             mode,
		Title:            strings.TrimSpace(n.attr("dialog-title")),
		InitialPath:      initialPath,
		DefaultExtension: strings.TrimSpace(n.attr("default-extension")),
		ButtonLabel:      strings.TrimSpace(n.attr("dialog-button-text")),
		Filters:          filters,
		MultiSelect:      multiple,
	})
	if separator := strings.TrimSpace(n.attr("value-separator")); separator != "" {
		picker.SetSeparator(separator)
	}
	if placeholder := strings.TrimSpace(n.attr("placeholder")); placeholder != "" {
		picker.SetPlaceholder(placeholder)
	}
	if buttonText := strings.TrimSpace(n.attr("button-text")); buttonText != "" {
		picker.SetButtonText(buttonText)
	}
	if err := b.applyCommonWidgetState(picker, n); err != nil {
		return nil, err
	}

	initialPaths := parseInitialDialogPaths(n.attr("value"), multiple)
	if len(initialPaths) > 0 {
		picker.SetPaths(initialPaths)
	}

	if actionName, err := b.actionName(n, "onchange"); err != nil {
		return nil, err
	} else if actionName != "" {
		separator := strings.TrimSpace(n.attr("value-separator"))
		if separator == "" {
			separator = "; "
		}
		picker.SetOnChange(func(paths []string) {
			ctx := b.baseActionContext(actionName, picker)
			ctx.Paths = append(ctx.Paths, paths...)
			ctx.Value = strings.Join(paths, separator)
			b.dispatchAction(actionName, ctx)
		})
	}

	defaultSize := core.Size{Width: 320, Height: 36}
	if err := b.applyPreferredSize(picker, n, defaultSize); err != nil {
		return nil, err
	}
	return picker, nil
}

func (b *uiBuilder) prefixedStyleNode(n *node, prefix string) *node {
	if n == nil {
		return nil
	}
	styles := make(map[string]string)
	for key, value := range n.Styles {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		trimmed := strings.TrimSpace(strings.TrimPrefix(key, prefix))
		if trimmed != "" {
			styles[trimmed] = value
		}
	}
	if len(styles) == 0 {
		return nil
	}
	return &node{
		Kind:   n.Kind,
		Tag:    n.Tag,
		Attrs:  n.Attrs,
		Styles: styles,
		Pos:    n.Pos,
	}
}

func parseDialogModeAttr(value string) (dialogs.DialogMode, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" || text == string(dialogs.DialogOpen) {
		return dialogs.DialogOpen, nil
	}
	switch dialogs.DialogMode(text) {
	case dialogs.DialogSave:
		return dialogs.DialogSave, nil
	case dialogs.DialogFolder:
		return dialogs.DialogFolder, nil
	default:
		return dialogs.DialogOpen, fmt.Errorf("invalid dialog value %q", value)
	}
}

func parseDialogFilters(raw string, accept string) ([]dialogs.FileFilter, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		items := strings.Split(raw, ",")
		filters := make([]dialogs.FileFilter, 0, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			parts := strings.SplitN(item, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid filters value %q", raw)
			}
			name := strings.TrimSpace(parts[0])
			pattern := strings.TrimSpace(parts[1])
			if name == "" || pattern == "" {
				return nil, fmt.Errorf("invalid filters value %q", raw)
			}
			filters = append(filters, dialogs.FileFilter{Name: name, Pattern: pattern})
		}
		if len(filters) > 0 {
			return filters, nil
		}
	}

	accept = strings.TrimSpace(accept)
	if accept == "" {
		return nil, nil
	}
	parts := strings.Split(accept, ",")
	patterns := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		switch {
		case strings.HasPrefix(part, "."):
			patterns = append(patterns, "*"+part)
		case strings.ContainsAny(part, "*?"):
			patterns = append(patterns, part)
		case strings.Contains(part, "/"):
			continue
		default:
			patterns = append(patterns, "*."+strings.TrimPrefix(part, "."))
		}
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	return []dialogs.FileFilter{{Name: "Accepted Files", Pattern: strings.Join(patterns, ";")}}, nil
}

func parseInitialDialogPaths(raw string, multiple bool) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if !multiple {
		return []string{raw}
	}
	normalized := strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(raw)
	fields := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '\n' || r == ';'
	})
	paths := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			paths = append(paths, field)
		}
	}
	if len(paths) == 0 {
		return []string{raw}
	}
	return paths
}

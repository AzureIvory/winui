//go:build windows

package markup

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type itemBindingOptions struct {
	TextField     string
	ValueField    string
	DisabledField string
}

func (b *uiBuilder) addBinding(paths []string, apply func(*bindingContext)) {
	if apply == nil {
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
	if len(normalized) == 0 {
		return
	}
	b.bindings = append(b.bindings, documentBinding{
		paths: normalized,
		apply: apply,
	})
}

func (b *uiBuilder) bindingPath(n *node, attr string) (string, bool, error) {
	if n == nil || !n.hasAttr(attr) {
		return "", false, nil
	}
	path := normalizeBindingPath(n.attr(attr))
	if path == "" {
		return "", false, newParseError("builder", n.Pos, n.inlineContext(), "%s requires a non-empty value", attr)
	}
	return path, true, nil
}

func (b *uiBuilder) bindingPathAny(n *node, attrs ...string) (string, bool, error) {
	for _, attr := range attrs {
		path, ok, err := b.bindingPath(n, attr)
		if err != nil {
			return "", false, err
		}
		if ok {
			return path, true, nil
		}
	}
	return "", false, nil
}

func (b *uiBuilder) registerStringBinding(n *node, attr string, fallback string, apply func(string)) error {
	path, ok, err := b.bindingPath(n, attr)
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		text := fallback
		if value, present := ctx.Lookup(path); present {
			if converted, ok := bindingStringValue(value); ok {
				text = converted
			}
		}
		apply(text)
	})
	return nil
}

func (b *uiBuilder) registerBoolBinding(n *node, attr string, fallback bool, apply func(bool)) error {
	path, ok, err := b.bindingPath(n, attr)
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		value := fallback
		if raw, present := ctx.Lookup(path); present {
			if converted, ok := bindingBoolValue(raw); ok {
				value = converted
			}
		}
		apply(value)
	})
	return nil
}

func (b *uiBuilder) registerValueBinding(n *node, attr string, fallback string, apply func(string)) error {
	return b.registerStringBinding(n, attr, fallback, apply)
}

func (b *uiBuilder) registerCheckedBinding(n *node, fallback bool, apply func(bool)) error {
	return b.registerBoolBinding(n, "bind-checked", fallback, apply)
}

func (b *uiBuilder) registerProgressBinding(n *node, fallback int32, apply func(int32)) error {
	path, ok, err := b.bindingPath(n, "bind-value")
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		value := fallback
		if raw, present := ctx.Lookup(path); present {
			if converted, ok := bindingInt32Value(raw); ok {
				value = converted
			}
		}
		apply(value)
	})
	return nil
}

func (b *uiBuilder) registerItemsBinding(
	n *node,
	fallback []widgets.ListItem,
	options itemBindingOptions,
	apply func([]widgets.ListItem),
) error {
	path, ok, err := b.bindingPath(n, "bind-items")
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		items := copyListItems(fallback)
		if raw, present := ctx.Lookup(path); present {
			if converted, ok := bindingListItemsValue(raw, options); ok {
				items = converted
			}
		}
		apply(items)
	})
	return nil
}

func (b *uiBuilder) registerSelectionBinding(
	n *node,
	fallbackIndex int,
	fallbackValue string,
	items func() []widgets.ListItem,
	apply func(index int, hasSelection bool),
) error {
	path, ok, err := b.bindingPath(n, "bind-selected")
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		index := fallbackIndex
		hasSelection := fallbackIndex >= 0
		if raw, present := ctx.Lookup(path); present {
			if raw == nil {
				index = -1
				hasSelection = false
			} else if resolved, ok := bindingSelectionIndex(items(), raw); ok {
				index = resolved
				hasSelection = true
			} else {
				index = -1
				hasSelection = false
			}
		} else if !hasSelection && fallbackValue != "" {
			if resolved, ok := bindingSelectionIndex(items(), fallbackValue); ok {
				index = resolved
				hasSelection = true
			}
		}
		apply(index, hasSelection)
	})
	return nil
}

func (b *uiBuilder) registerWindowTitleBinding(window *node, fallback string) error {
	path, ok, err := b.bindingPath(window, "bind-title")
	if err != nil || !ok {
		return err
	}
	b.addBinding([]string{path}, func(ctx *bindingContext) {
		title := fallback
		if raw, present := ctx.Lookup(path); present {
			if converted, ok := bindingStringValue(raw); ok {
				title = converted
			}
		}
		ctx.document.setWindowTitle(title)
	})
	return nil
}

func (b *uiBuilder) registerCommonBindings(widget widgets.Widget, n *node, visible bool, enabled bool) error {
	if err := b.registerBoolBinding(n, "bind-visible", visible, widget.SetVisible); err != nil {
		return err
	}
	if err := b.registerBoolBinding(n, "bind-enabled", enabled, widget.SetEnabled); err != nil {
		return err
	}
	return nil
}

func (b *uiBuilder) registerPreferredSizeBindings(widget widgets.Widget, n *node) error {
	widthPath, hasWidth, err := b.bindingPath(n, "bind-width")
	if err != nil {
		return err
	}
	heightPath, hasHeight, err := b.bindingPath(n, "bind-height")
	if err != nil {
		return err
	}
	if !hasWidth && !hasHeight {
		return nil
	}

	base := widget.Bounds()
	paths := make([]string, 0, 2)
	if hasWidth {
		paths = append(paths, widthPath)
	}
	if hasHeight {
		paths = append(paths, heightPath)
	}
	b.addBinding(paths, func(ctx *bindingContext) {
		size := core.Size{Width: base.W, Height: base.H}
		if hasWidth {
			if raw, present := ctx.Lookup(widthPath); present {
				if converted, ok := bindingLengthValue(raw); ok {
					size.Width = converted
				}
			}
		}
		if hasHeight {
			if raw, present := ctx.Lookup(heightPath); present {
				if converted, ok := bindingLengthValue(raw); ok {
					size.Height = converted
				}
			}
		}
		applyBoundPreferredSize(widget, size)
	})
	return nil
}

func (b *uiBuilder) registerLayoutBindings(widget widgets.Widget, n *node, parentLayout string) error {
	if widget == nil || n == nil || parentLayout != "absolute" {
		return nil
	}

	base, _ := widget.LayoutData().(widgets.AbsoluteLayoutData)
	leftPath, hasLeft, err := b.bindingPathAny(n, "bind-left", "bind-x")
	if err != nil {
		return err
	}
	topPath, hasTop, err := b.bindingPathAny(n, "bind-top", "bind-y")
	if err != nil {
		return err
	}
	rightPath, hasRight, err := b.bindingPath(n, "bind-right")
	if err != nil {
		return err
	}
	bottomPath, hasBottom, err := b.bindingPath(n, "bind-bottom")
	if err != nil {
		return err
	}
	widthPath, hasWidth, err := b.bindingPath(n, "bind-width")
	if err != nil {
		return err
	}
	heightPath, hasHeight, err := b.bindingPath(n, "bind-height")
	if err != nil {
		return err
	}
	if !hasLeft && !hasTop && !hasRight && !hasBottom && !hasWidth && !hasHeight {
		return nil
	}

	paths := make([]string, 0, 6)
	for _, path := range []string{leftPath, topPath, rightPath, bottomPath, widthPath, heightPath} {
		if path != "" {
			paths = append(paths, path)
		}
	}
	b.addBinding(paths, func(ctx *bindingContext) {
		data := base
		if hasLeft {
			data.Left, data.HasLeft = bindingAbsoluteLength(ctx, leftPath, base.Left, base.HasLeft)
		}
		if hasTop {
			data.Top, data.HasTop = bindingAbsoluteLength(ctx, topPath, base.Top, base.HasTop)
		}
		if hasRight {
			data.Right, data.HasRight = bindingAbsoluteLength(ctx, rightPath, base.Right, base.HasRight)
		}
		if hasBottom {
			data.Bottom, data.HasBottom = bindingAbsoluteLength(ctx, bottomPath, base.Bottom, base.HasBottom)
		}
		if hasWidth {
			data.Width, data.HasWidth = bindingAbsoluteLength(ctx, widthPath, base.Width, base.HasWidth)
		}
		if hasHeight {
			data.Height, data.HasHeight = bindingAbsoluteLength(ctx, heightPath, base.Height, base.HasHeight)
		}
		widget.SetLayoutData(data)
	})
	return nil
}

func bindingAbsoluteLength(ctx *bindingContext, path string, fallback int32, hasFallback bool) (int32, bool) {
	if ctx == nil || path == "" {
		return fallback, hasFallback
	}
	raw, present := ctx.Lookup(path)
	if !present {
		return fallback, hasFallback
	}
	if converted, ok := bindingLengthValue(raw); ok {
		return converted, true
	}
	return fallback, hasFallback
}

func bindingStringValue(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", true
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case fmt.Stringer:
		return typed.String(), true
	default:
		return fmt.Sprint(value), true
	}
}

func bindingBoolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, ok, err := parseBoolValue(typed)
		return parsed, ok && err == nil
	case int:
		return typed != 0, true
	case int8:
		return typed != 0, true
	case int16:
		return typed != 0, true
	case int32:
		return typed != 0, true
	case int64:
		return typed != 0, true
	case uint:
		return typed != 0, true
	case uint8:
		return typed != 0, true
	case uint16:
		return typed != 0, true
	case uint32:
		return typed != 0, true
	case uint64:
		return typed != 0, true
	default:
		return false, false
	}
}

func bindingInt32Value(value any) (int32, bool) {
	switch typed := value.(type) {
	case int:
		return int32(typed), true
	case int8:
		return int32(typed), true
	case int16:
		return int32(typed), true
	case int32:
		return typed, true
	case int64:
		return int32(typed), true
	case uint:
		return int32(typed), true
	case uint8:
		return int32(typed), true
	case uint16:
		return int32(typed), true
	case uint32:
		return int32(typed), true
	case uint64:
		return int32(typed), true
	case float32:
		return int32(typed), true
	case float64:
		return int32(typed), true
	case string:
		if parsed, ok, err := parseIntegerValue(typed); err == nil && ok {
			return int32(parsed), true
		}
		if parsed, ok, err := parseLengthValue(typed); err == nil && ok {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func bindingLengthValue(value any) (int32, bool) {
	return bindingInt32Value(value)
}

func bindingSelectionIndex(items []widgets.ListItem, value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return bindingSelectionIndexFromNumber(items, typed)
	case int8:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case int16:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case int32:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case int64:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case uint:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case uint8:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case uint16:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case uint32:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case uint64:
		return bindingSelectionIndexFromNumber(items, int(typed))
	case widgets.ListItem:
		if index, ok := bindingSelectionIndex(items, typed.Value); ok {
			return index, true
		}
		return bindingSelectionIndex(items, typed.Text)
	default:
		text, ok := bindingStringValue(value)
		if !ok {
			return -1, false
		}
		for index, item := range items {
			if item.Value == text || item.Text == text {
				return index, true
			}
		}
		return -1, false
	}
}

func bindingSelectionIndexFromNumber(items []widgets.ListItem, index int) (int, bool) {
	if index < 0 || index >= len(items) {
		return -1, false
	}
	return index, true
}

func bindingListItemsValue(value any, options itemBindingOptions) ([]widgets.ListItem, bool) {
	if value == nil {
		return []widgets.ListItem{}, true
	}
	if items, ok := value.([]widgets.ListItem); ok {
		return copyListItems(items), true
	}

	current := reflect.ValueOf(value)
	for current.IsValid() && (current.Kind() == reflect.Interface || current.Kind() == reflect.Pointer) {
		if current.IsNil() {
			return []widgets.ListItem{}, true
		}
		current = current.Elem()
	}
	if !current.IsValid() {
		return []widgets.ListItem{}, true
	}
	if current.Kind() != reflect.Slice && current.Kind() != reflect.Array {
		return nil, false
	}

	items := make([]widgets.ListItem, 0, current.Len())
	for index := 0; index < current.Len(); index++ {
		item, ok := bindingListItemValue(current.Index(index).Interface(), options)
		if !ok {
			return nil, false
		}
		items = append(items, item)
	}
	return items, true
}

func bindingListItemValue(value any, options itemBindingOptions) (widgets.ListItem, bool) {
	switch typed := value.(type) {
	case widgets.ListItem:
		if typed.Text == "" {
			typed.Text = typed.Value
		}
		return typed, true
	case string:
		return widgets.ListItem{Value: typed, Text: typed}, true
	case []byte:
		text := string(typed)
		return widgets.ListItem{Value: text, Text: text}, true
	}

	item := widgets.ListItem{}
	if text, ok := bindingItemFieldString(value, options.TextField, "text", "label", "name", "title"); ok {
		item.Text = text
	}
	if itemValue, ok := bindingItemFieldString(value, options.ValueField, "value", "id", "key", "text", "name"); ok {
		item.Value = itemValue
	}
	if disabled, ok := bindingItemFieldBool(value, options.DisabledField, "disabled"); ok {
		item.Disabled = disabled
	}
	if item.Value == "" && item.Text == "" {
		text, ok := bindingStringValue(value)
		if !ok {
			return widgets.ListItem{}, false
		}
		item.Value = text
		item.Text = text
	}
	if item.Text == "" {
		item.Text = item.Value
	}
	if item.Value == "" {
		item.Value = item.Text
	}
	return item, true
}

func bindingItemFieldString(value any, explicit string, fallbacks ...string) (string, bool) {
	if explicit != "" {
		raw, ok := lookupStateValue(value, explicit)
		if !ok {
			return "", false
		}
		return bindingStringValue(raw)
	}
	for _, candidate := range fallbacks {
		raw, ok := lookupStateValue(value, candidate)
		if !ok {
			continue
		}
		if text, ok := bindingStringValue(raw); ok {
			return text, true
		}
	}
	return "", false
}

func bindingItemFieldBool(value any, explicit string, fallbacks ...string) (bool, bool) {
	if explicit != "" {
		raw, ok := lookupStateValue(value, explicit)
		if !ok {
			return false, false
		}
		return bindingBoolValue(raw)
	}
	for _, candidate := range fallbacks {
		raw, ok := lookupStateValue(value, candidate)
		if !ok {
			continue
		}
		if flag, ok := bindingBoolValue(raw); ok {
			return flag, true
		}
	}
	return false, false
}

func applyBoundPreferredSize(widget widgets.Widget, size core.Size) {
	if widget == nil {
		return
	}
	if size.Width < 0 {
		size.Width = 0
	}
	if size.Height < 0 {
		size.Height = 0
	}
	bounds := widget.Bounds()
	bounds.W = size.Width
	bounds.H = size.Height
	widget.SetBounds(bounds)
	widgets.SetPreferredSize(widget, size)
}

func copyListItems(items []widgets.ListItem) []widgets.ListItem {
	if len(items) == 0 {
		return []widgets.ListItem{}
	}
	out := make([]widgets.ListItem, 0, len(items))
	for _, item := range items {
		if item.Text == "" {
			item.Text = item.Value
		}
		out = append(out, item)
	}
	return out
}

func bindingOptionsForNode(n *node) itemBindingOptions {
	if n == nil {
		return itemBindingOptions{}
	}
	return itemBindingOptions{
		TextField:     strings.TrimSpace(n.attr("item-text-field")),
		ValueField:    strings.TrimSpace(n.attr("item-value-field")),
		DisabledField: strings.TrimSpace(n.attr("item-disabled-field")),
	}
}

func preferredSelectionValue(selected int, items []widgets.ListItem) string {
	if selected < 0 || selected >= len(items) {
		return ""
	}
	return items[selected].Value
}

func bindingIndexPath(path string) (int, bool) {
	index, err := strconv.Atoi(path)
	if err != nil {
		return 0, false
	}
	return index, true
}

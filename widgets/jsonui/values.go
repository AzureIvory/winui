//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type bindingSpec struct {
	Bind    string          `json:"bind"`
	Default json.RawMessage `json:"default"`
}

type stringSource struct {
	Has     bool
	Literal string
	Binding string
	Default string
}

type boolSource struct {
	Has        bool
	Literal    bool
	Binding    string
	Default    bool
	HasDefault bool
}

type intSource struct {
	Has     bool
	Literal int32
	Binding string
	Default int32
}

type exprSource struct {
	Has     bool
	Literal ScalarExpr
	Binding string
	Default ScalarExpr
}

type itemsSource struct {
	Has     bool
	Literal []widgets.ListItem
	Binding string
	Default []widgets.ListItem
}

type selectionSource struct {
	Has         bool
	Literal     any
	Binding     string
	Default     any
	HasSelected bool
}

func parseBindingSpec(raw json.RawMessage) (bindingSpec, bool, error) {
	if len(raw) == 0 {
		return bindingSpec{}, false, nil
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return bindingSpec{}, false, nil
	}
	_, ok := probe["bind"]
	if !ok {
		return bindingSpec{}, false, nil
	}
	var spec bindingSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return bindingSpec{}, false, err
	}
	spec.Bind = normalizeBindingPath(spec.Bind)
	if spec.Bind == "" {
		return bindingSpec{}, false, fmt.Errorf("binding path is empty")
	}
	return spec, true, nil
}

func parseStringSource(raw json.RawMessage) (stringSource, error) {
	if len(raw) == 0 {
		return stringSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return stringSource{}, err
	} else if ok {
		defaultValue := ""
		if len(binding.Default) > 0 {
			value, err := decodeStringLiteral(binding.Default)
			if err != nil {
				return stringSource{}, err
			}
			defaultValue = value
		}
		return stringSource{
			Has:     true,
			Binding: binding.Bind,
			Default: defaultValue,
		}, nil
	}
	value, err := decodeStringLiteral(raw)
	if err != nil {
		return stringSource{}, err
	}
	return stringSource{Has: true, Literal: value}, nil
}

func parseBoolSource(raw json.RawMessage) (boolSource, error) {
	if len(raw) == 0 {
		return boolSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return boolSource{}, err
	} else if ok {
		defaultValue := false
		hasDefault := false
		if len(binding.Default) > 0 {
			value, err := decodeBoolLiteral(binding.Default)
			if err != nil {
				return boolSource{}, err
			}
			defaultValue = value
			hasDefault = true
		}
		return boolSource{
			Has:        true,
			Binding:    binding.Bind,
			Default:    defaultValue,
			HasDefault: hasDefault,
		}, nil
	}
	value, err := decodeBoolLiteral(raw)
	if err != nil {
		return boolSource{}, err
	}
	return boolSource{Has: true, Literal: value}, nil
}

func parseIntSource(raw json.RawMessage) (intSource, error) {
	if len(raw) == 0 {
		return intSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return intSource{}, err
	} else if ok {
		var defaultValue int32
		if len(binding.Default) > 0 {
			value, err := decodeInt32Literal(binding.Default)
			if err != nil {
				return intSource{}, err
			}
			defaultValue = value
		}
		return intSource{
			Has:     true,
			Binding: binding.Bind,
			Default: defaultValue,
		}, nil
	}
	value, err := decodeInt32Literal(raw)
	if err != nil {
		return intSource{}, err
	}
	return intSource{Has: true, Literal: value}, nil
}

func parseExprSource(raw json.RawMessage) (exprSource, error) {
	if len(raw) == 0 {
		return exprSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return exprSource{}, err
	} else if ok {
		defaultExpr := ScalarExpr{}
		if len(binding.Default) > 0 {
			decoded, err := decodeScalarLiteral(binding.Default)
			if err != nil {
				return exprSource{}, err
			}
			expr, err := ParseScalarExpr(decoded)
			if err != nil {
				return exprSource{}, err
			}
			defaultExpr = expr
		}
		return exprSource{
			Has:     true,
			Binding: binding.Bind,
			Default: defaultExpr,
		}, nil
	}
	decoded, err := decodeScalarLiteral(raw)
	if err != nil {
		return exprSource{}, err
	}
	expr, err := ParseScalarExpr(decoded)
	if err != nil {
		return exprSource{}, err
	}
	return exprSource{Has: true, Literal: expr}, nil
}

func parseItemsSource(raw json.RawMessage) (itemsSource, error) {
	if len(raw) == 0 {
		return itemsSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return itemsSource{}, err
	} else if ok {
		var defaults []widgets.ListItem
		if len(binding.Default) > 0 {
			items, err := decodeListItemsLiteral(binding.Default)
			if err != nil {
				return itemsSource{}, err
			}
			defaults = items
		}
		return itemsSource{
			Has:     true,
			Binding: binding.Bind,
			Default: defaults,
		}, nil
	}
	items, err := decodeListItemsLiteral(raw)
	if err != nil {
		return itemsSource{}, err
	}
	return itemsSource{Has: true, Literal: items}, nil
}

func parseSelectionSource(raw json.RawMessage) (selectionSource, error) {
	if len(raw) == 0 {
		return selectionSource{}, nil
	}
	if binding, ok, err := parseBindingSpec(raw); err != nil {
		return selectionSource{}, err
	} else if ok {
		var defaultValue any
		if len(binding.Default) > 0 {
			value, err := decodeScalarLiteral(binding.Default)
			if err != nil {
				return selectionSource{}, err
			}
			defaultValue = value
		}
		return selectionSource{
			Has:         true,
			Binding:     binding.Bind,
			Default:     defaultValue,
			HasSelected: defaultValue != nil,
		}, nil
	}
	value, err := decodeScalarLiteral(raw)
	if err != nil {
		return selectionSource{}, err
	}
	return selectionSource{
		Has:         true,
		Literal:     value,
		HasSelected: value != nil,
	}, nil
}

func decodeScalarLiteral(raw json.RawMessage) (any, error) {
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return nil, err
		}
		return int32(number), nil
	default:
		return typed, nil
	}
}

func decodeStringLiteral(raw json.RawMessage) (string, error) {
	value, err := decodeScalarLiteral(raw)
	if err != nil {
		return "", err
	}
	text, ok := bindingStringValue(value)
	if !ok {
		return "", fmt.Errorf("expected string literal")
	}
	return text, nil
}

func decodeBoolLiteral(raw json.RawMessage) (bool, error) {
	value, err := decodeScalarLiteral(raw)
	if err != nil {
		return false, err
	}
	parsed, ok := bindingBoolValue(value)
	if !ok {
		return false, fmt.Errorf("expected bool literal")
	}
	return parsed, nil
}

func decodeInt32Literal(raw json.RawMessage) (int32, error) {
	value, err := decodeScalarLiteral(raw)
	if err != nil {
		return 0, err
	}
	parsed, ok := bindingInt32Value(value)
	if !ok {
		return 0, fmt.Errorf("expected integer literal")
	}
	return parsed, nil
}

type itemLiteral struct {
	Value    string `json:"value"`
	Text     string `json:"text"`
	Disabled bool   `json:"disabled"`
	Selected bool   `json:"selected"`
}

func decodeListItemsLiteral(raw json.RawMessage) ([]widgets.ListItem, error) {
	var items []itemLiteral
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	out := make([]widgets.ListItem, 0, len(items))
	for _, item := range items {
		if item.Value == "" {
			item.Value = item.Text
		}
		out = append(out, widgets.ListItem{
			Value:    item.Value,
			Text:     item.Text,
			Disabled: item.Disabled,
		})
	}
	return out, nil
}

func cloneListItems(items []widgets.ListItem) []widgets.ListItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]widgets.ListItem, len(items))
	copy(out, items)
	return out
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
		return parseBoolValue(typed)
	case int, int8, int16, int32, int64:
		return fmt.Sprint(typed) != "0", true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(typed) != "0", true
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
	case float64:
		return int32(typed), true
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int32(number), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return int32(parsed), true
	default:
		return 0, false
	}
}

func bindingListItemsValue(value any) ([]widgets.ListItem, bool) {
	switch typed := value.(type) {
	case []widgets.ListItem:
		return cloneListItems(typed), true
	case []any:
		out := make([]widgets.ListItem, 0, len(typed))
		for _, item := range typed {
			switch entry := item.(type) {
			case map[string]any:
				text, _ := bindingStringValue(entry["text"])
				val, _ := bindingStringValue(entry["value"])
				if val == "" {
					val = text
				}
				disabled, _ := bindingBoolValue(entry["disabled"])
				out = append(out, widgets.ListItem{
					Value:    val,
					Text:     text,
					Disabled: disabled,
				})
			case string:
				out = append(out, widgets.ListItem{Value: entry, Text: entry})
			default:
				text, _ := bindingStringValue(entry)
				out = append(out, widgets.ListItem{Value: text, Text: text})
			}
		}
		return out, true
	default:
		return nil, false
	}
}

func bindingSelectionIndex(items []widgets.ListItem, value any) (int, bool) {
	switch typed := value.(type) {
	case nil:
		return -1, true
	case int:
		return selectionIndexFromInt(items, typed)
	case int32:
		return selectionIndexFromInt(items, int(typed))
	case int64:
		return selectionIndexFromInt(items, int(typed))
	case float64:
		return selectionIndexFromInt(items, int(typed))
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return -1, false
		}
		return selectionIndexFromInt(items, int(number))
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

func selectionIndexFromInt(items []widgets.ListItem, index int) (int, bool) {
	if index < 0 {
		return -1, true
	}
	if index >= len(items) {
		return -1, false
	}
	return index, true
}

func parseBoolValue(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return true, true
	case "false", "0", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func parseColorValue(value string) (core.Color, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return 0, false, nil
	}
	if strings.HasPrefix(text, "#") {
		hex := strings.TrimPrefix(text, "#")
		switch len(hex) {
		case 3:
			r, err := strconv.ParseUint(strings.Repeat(string(hex[0]), 2), 16, 8)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color %q", value)
			}
			g, err := strconv.ParseUint(strings.Repeat(string(hex[1]), 2), 16, 8)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color %q", value)
			}
			b, err := strconv.ParseUint(strings.Repeat(string(hex[2]), 2), 16, 8)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color %q", value)
			}
			return core.RGB(byte(r), byte(g), byte(b)), true, nil
		case 6:
			parsed, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color %q", value)
			}
			return core.RGB(byte(parsed>>16), byte(parsed>>8), byte(parsed)), true, nil
		default:
			return 0, false, fmt.Errorf("invalid color %q", value)
		}
	}
	named := map[string]core.Color{
		"black":       core.RGB(0, 0, 0),
		"white":       core.RGB(255, 255, 255),
		"red":         core.RGB(255, 0, 0),
		"green":       core.RGB(0, 128, 0),
		"blue":        core.RGB(0, 0, 255),
		"gray":        core.RGB(128, 128, 128),
		"grey":        core.RGB(128, 128, 128),
		"transparent": 0,
	}
	color, ok := named[text]
	if !ok {
		return 0, false, fmt.Errorf("invalid color %q", value)
	}
	return color, true, nil
}

func parseAlignmentValue(value string) (widgets.Alignment, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return widgets.AlignDefault, false, nil
	case "left", "start", "top":
		return widgets.AlignStart, true, nil
	case "center", "middle":
		return widgets.AlignCenter, true, nil
	case "right", "end", "bottom":
		return widgets.AlignEnd, true, nil
	case "stretch":
		return widgets.AlignStretch, true, nil
	default:
		return widgets.AlignDefault, false, fmt.Errorf("invalid alignment %q", value)
	}
}

func parseFontWeightValue(value string) (int32, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return 0, false, nil
	case "normal":
		return 400, true, nil
	case "bold":
		return 700, true, nil
	default:
		number, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false, fmt.Errorf("invalid font weight %q", value)
		}
		return int32(number), true, nil
	}
}

func parseIndicatorStyleValue(value string) (widgets.ChoiceIndicatorStyle, bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return widgets.ChoiceIndicatorAuto, false, nil
	case "auto":
		return widgets.ChoiceIndicatorAuto, true, nil
	case "dot":
		return widgets.ChoiceIndicatorDot, true, nil
	case "check":
		return widgets.ChoiceIndicatorCheck, true, nil
	default:
		return widgets.ChoiceIndicatorAuto, false, fmt.Errorf("invalid indicator style %q", value)
	}
}

func parseTextFormat(align widgets.Alignment, multiline bool) uint32 {
	format := uint32(0)
	switch align {
	case widgets.AlignCenter:
		format |= core.DTCenter
	case widgets.AlignEnd:
		format |= core.DTRight
	}
	if !multiline {
		format |= core.DTVCenter | core.DTSingleLine | core.DTEndEllipsis
	}
	return format
}

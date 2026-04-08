//go:build windows

package markup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

func parseBoolValue(value string) (bool, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return false, false, nil
	}
	switch text {
	case "true", "1", "yes", "on":
		return true, true, nil
	case "false", "0", "no", "off":
		return false, true, nil
	default:
		return false, false, fmt.Errorf("invalid bool value %q", value)
	}
}

func parseIntegerValue(value string) (int, bool, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(text)
	if err != nil {
		return 0, false, fmt.Errorf("invalid integer value %q", value)
	}
	return parsed, true, nil
}

func parseLengthValue(value string) (int32, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" || text == "auto" {
		return 0, false, nil
	}
	text = strings.TrimSuffix(text, "px")
	parsed, err := strconv.Atoi(text)
	if err != nil {
		return 0, false, fmt.Errorf("invalid length value %q", value)
	}
	return int32(parsed), true, nil
}

func parseInsetsValue(value string) (widgets.Insets, bool, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return widgets.Insets{}, false, nil
	}
	parts := strings.Fields(text)
	values := make([]int32, 0, len(parts))
	for _, part := range parts {
		parsed, ok, err := parseLengthValue(part)
		if err != nil {
			return widgets.Insets{}, false, err
		}
		if !ok {
			return widgets.Insets{}, false, fmt.Errorf("invalid insets value %q", value)
		}
		values = append(values, parsed)
	}
	switch len(values) {
	case 1:
		return widgets.UniformInsets(values[0]), true, nil
	case 2:
		return widgets.Insets{Top: values[0], Right: values[1], Bottom: values[0], Left: values[1]}, true, nil
	case 3:
		return widgets.Insets{Top: values[0], Right: values[1], Bottom: values[2], Left: values[1]}, true, nil
	case 4:
		return widgets.Insets{Top: values[0], Right: values[1], Bottom: values[2], Left: values[3]}, true, nil
	default:
		return widgets.Insets{}, false, fmt.Errorf("invalid insets value %q", value)
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
				return 0, false, fmt.Errorf("invalid color value %q", value)
			}
			g, err := strconv.ParseUint(strings.Repeat(string(hex[1]), 2), 16, 8)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color value %q", value)
			}
			b, err := strconv.ParseUint(strings.Repeat(string(hex[2]), 2), 16, 8)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color value %q", value)
			}
			return core.RGB(byte(r), byte(g), byte(b)), true, nil
		case 6:
			parsed, err := strconv.ParseUint(hex, 16, 32)
			if err != nil {
				return 0, false, fmt.Errorf("invalid color value %q", value)
			}
			return core.RGB(byte(parsed>>16), byte(parsed>>8), byte(parsed)), true, nil
		default:
			return 0, false, fmt.Errorf("invalid color value %q", value)
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
	if color, ok := named[text]; ok {
		return color, true, nil
	}
	return 0, false, fmt.Errorf("invalid color value %q", value)
}

func parseDisplayValue(value string) (string, bool, error) {
	return parseKeyword(value, "display", "none", "row", "column", "grid", "form", "absolute")
}

func parseOverflowValue(value string) (string, bool, error) {
	return parseKeyword(value, "overflow", "visible", "hidden", "auto", "scroll")
}

func parseObjectFitValue(value string) (string, bool, error) {
	return parseKeyword(value, "object-fit", "contain", "cover", "fill")
}

func parseAlignmentValue(value string) (widgets.Alignment, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return widgets.AlignDefault, false, nil
	}
	switch text {
	case "left", "start", "top":
		return widgets.AlignStart, true, nil
	case "center", "middle":
		return widgets.AlignCenter, true, nil
	case "right", "end", "bottom":
		return widgets.AlignEnd, true, nil
	case "stretch":
		return widgets.AlignStretch, true, nil
	default:
		return widgets.AlignDefault, false, fmt.Errorf("invalid alignment value %q", value)
	}
}

func parseFontWeightValue(value string) (int32, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return 0, false, nil
	}
	switch text {
	case "normal":
		return 400, true, nil
	case "bold":
		return 700, true, nil
	default:
		parsed, ok, err := parseIntegerValue(text)
		if err != nil || !ok {
			return 0, false, fmt.Errorf("invalid font-weight value %q", value)
		}
		return int32(parsed), true, nil
	}
}

func parseChoiceIndicatorStyle(value string) (widgets.ChoiceIndicatorStyle, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return widgets.ChoiceIndicatorAuto, false, nil
	}
	switch text {
	case "auto":
		return widgets.ChoiceIndicatorAuto, true, nil
	case "dot":
		return widgets.ChoiceIndicatorDot, true, nil
	case "check":
		return widgets.ChoiceIndicatorCheck, true, nil
	default:
		return widgets.ChoiceIndicatorAuto, false, fmt.Errorf("invalid indicator-style value %q", value)
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

func parseKeyword(value string, name string, allowed ...string) (string, bool, error) {
	text := strings.TrimSpace(strings.ToLower(value))
	if text == "" {
		return "", false, nil
	}
	for _, item := range allowed {
		if text == item {
			return text, true, nil
		}
	}
	return "", false, fmt.Errorf("invalid %s value %q", name, value)
}

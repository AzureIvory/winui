//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"

	"github.com/AzureIvory/winui/widgets"
)

type styleMap map[string]json.RawMessage

func decodeStyleMap(raw json.RawMessage) (styleMap, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var styles styleMap
	if err := json.Unmarshal(raw, &styles); err != nil {
		return nil, err
	}
	return styles, nil
}

func parseFontSpec(styles styleMap) (widgets.FontSpec, error) {
	var font widgets.FontSpec
	if styles == nil {
		return font, nil
	}
	if raw := styles["font"]; len(raw) > 0 {
		value, err := decodeStringLiteral(raw)
		if err != nil {
			return font, err
		}
		font.Face = value
	}
	if raw := styles["size"]; len(raw) > 0 {
		value, err := decodeInt32Literal(raw)
		if err != nil {
			return font, err
		}
		font.SizeDP = value
	}
	if raw := styles["weight"]; len(raw) > 0 {
		text, err := decodeStringLiteral(raw)
		if err == nil {
			if value, ok, err := parseFontWeightValue(text); err != nil {
				return font, err
			} else if ok {
				font.Weight = value
			}
		} else {
			value, err := decodeInt32Literal(raw)
			if err != nil {
				return font, err
			}
			font.Weight = value
		}
	}
	return font, nil
}

func parseTextStyle(raw json.RawMessage, multiline bool) (widgets.TextStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.TextStyle{}, err
	}
	style := widgets.TextStyle{}
	if styles == nil {
		return style, nil
	}
	if alignRaw := styles["align"]; len(alignRaw) > 0 {
		alignText, err := decodeStringLiteral(alignRaw)
		if err != nil {
			return style, err
		}
		align, ok, err := parseAlignmentValue(alignText)
		if err != nil {
			return style, err
		}
		if ok {
			style.Format = parseTextFormat(align, multiline)
		}
	}
	if style.Format == 0 {
		style.Format = parseTextFormat(widgets.AlignDefault, multiline)
	}
	if colorRaw := styles["fg"]; len(colorRaw) > 0 {
		text, err := decodeStringLiteral(colorRaw)
		if err != nil {
			return style, err
		}
		color, ok, err := parseColorValue(text)
		if err != nil {
			return style, err
		}
		if ok {
			style.Color = color
		}
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	return style, nil
}

func parseButtonStyle(raw json.RawMessage) (widgets.ButtonStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.ButtonStyle{}, err
	}
	style := widgets.ButtonStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	if alignRaw := styles["align"]; len(alignRaw) > 0 {
		alignText, err := decodeStringLiteral(alignRaw)
		if err != nil {
			return style, err
		}
		align, ok, err := parseAlignmentValue(alignText)
		if err != nil {
			return style, err
		}
		if ok {
			style.TextAlign = align
		}
	}
	if err := assignColor(styles, "fg", &style.TextColor); err != nil {
		return style, err
	}
	if err := assignColor(styles, "downFg", &style.DownText); err != nil {
		return style, err
	}
	if err := assignColor(styles, "disabledFg", &style.DisabledText); err != nil {
		return style, err
	}
	if err := assignColor(styles, "bg", &style.Background); err != nil {
		return style, err
	}
	if err := assignColor(styles, "hoverBg", &style.Hover); err != nil {
		return style, err
	}
	if err := assignColor(styles, "pressedBg", &style.Pressed); err != nil {
		return style, err
	}
	if err := assignColor(styles, "disabledBg", &style.Disabled); err != nil {
		return style, err
	}
	if err := assignColor(styles, "border", &style.Border); err != nil {
		return style, err
	}
	assignInt(styles, "radius", &style.CornerRadius)
	assignInt(styles, "iconSize", &style.IconSizeDP)
	assignInt(styles, "textInset", &style.TextInsetDP)
	assignInt(styles, "gap", &style.GapDP)
	assignInt(styles, "pad", &style.PadDP)
	return style, nil
}

func parseProgressStyle(raw json.RawMessage) (widgets.ProgressStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.ProgressStyle{}, err
	}
	style := widgets.ProgressStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	if err := assignColor(styles, "fg", &style.TextColor); err != nil {
		return style, err
	}
	if err := assignColor(styles, "track", &style.TrackColor); err != nil {
		return style, err
	}
	if err := assignColor(styles, "fill", &style.FillColor); err != nil {
		return style, err
	}
	if err := assignColor(styles, "bubble", &style.BubbleColor); err != nil {
		return style, err
	}
	assignInt(styles, "radius", &style.CornerRadius)
	assignBool(styles, "showPct", &style.ShowPercent)
	return style, nil
}

func parseChoiceStyle(raw json.RawMessage) (widgets.ChoiceStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.ChoiceStyle{}, err
	}
	style := widgets.ChoiceStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	assignColor(styles, "fg", &style.TextColor)
	assignColor(styles, "disabledFg", &style.DisabledText)
	assignColor(styles, "bg", &style.Background)
	assignColor(styles, "border", &style.BorderColor)
	assignColor(styles, "hoverBorder", &style.HoverBorder)
	assignColor(styles, "focusBorder", &style.FocusBorder)
	assignColor(styles, "indicator", &style.IndicatorColor)
	assignColor(styles, "check", &style.CheckColor)
	assignColor(styles, "hoverBg", &style.HoverBackground)
	assignColor(styles, "disabledBg", &style.DisabledBg)
	assignColor(styles, "disabledBorder", &style.DisabledBorder)
	assignInt(styles, "radius", &style.CornerRadius)
	assignInt(styles, "indicatorSize", &style.IndicatorSizeDP)
	assignInt(styles, "indicatorGap", &style.IndicatorGapDP)
	if raw := styles["indicatorStyle"]; len(raw) > 0 {
		value, err := decodeStringLiteral(raw)
		if err != nil {
			return style, err
		}
		indicatorStyle, ok, err := parseIndicatorStyleValue(value)
		if err != nil {
			return style, err
		}
		if ok {
			style.IndicatorStyle = indicatorStyle
		}
	}
	return style, nil
}

func parseComboStyle(raw json.RawMessage) (widgets.ComboStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.ComboStyle{}, err
	}
	style := widgets.ComboStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	assignColor(styles, "fg", &style.TextColor)
	assignColor(styles, "ph", &style.PlaceholderColor)
	assignColor(styles, "bg", &style.Background)
	assignColor(styles, "border", &style.BorderColor)
	assignColor(styles, "hoverBorder", &style.HoverBorder)
	assignColor(styles, "focusBorder", &style.FocusBorder)
	assignColor(styles, "arrow", &style.ArrowColor)
	assignColor(styles, "popupBg", &style.PopupBackground)
	assignColor(styles, "itemHoverBg", &style.ItemHoverColor)
	assignColor(styles, "itemSelectedBg", &style.ItemSelectedColor)
	assignColor(styles, "itemFg", &style.ItemTextColor)
	assignInt(styles, "radius", &style.CornerRadius)
	assignInt(styles, "pad", &style.PaddingDP)
	assignInt(styles, "itemH", &style.ItemHeightDP)
	assignInt(styles, "maxItems", &style.MaxVisibleItems)
	return style, nil
}

func parseListStyle(raw json.RawMessage) (widgets.ListStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.ListStyle{}, err
	}
	style := widgets.ListStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	assignColor(styles, "fg", &style.TextColor)
	assignColor(styles, "disabledFg", &style.DisabledText)
	assignColor(styles, "bg", &style.Background)
	assignColor(styles, "border", &style.BorderColor)
	assignColor(styles, "hoverBorder", &style.HoverBorder)
	assignColor(styles, "focusBorder", &style.FocusBorder)
	assignColor(styles, "itemHoverBg", &style.ItemHoverColor)
	assignColor(styles, "itemSelectedBg", &style.ItemSelectedColor)
	assignColor(styles, "itemFg", &style.ItemTextColor)
	assignInt(styles, "radius", &style.CornerRadius)
	assignInt(styles, "pad", &style.PaddingDP)
	assignInt(styles, "itemH", &style.ItemHeightDP)
	return style, nil
}

func parseEditStyle(raw json.RawMessage) (widgets.EditStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.EditStyle{}, err
	}
	style := widgets.EditStyle{}
	if styles == nil {
		return style, nil
	}
	font, err := parseFontSpec(styles)
	if err != nil {
		return style, err
	}
	style.Font = font
	if alignRaw := styles["align"]; len(alignRaw) > 0 {
		alignText, err := decodeStringLiteral(alignRaw)
		if err != nil {
			return style, err
		}
		align, ok, err := parseAlignmentValue(alignText)
		if err != nil {
			return style, err
		}
		if ok {
			style.TextAlign = align
		}
	}
	assignColor(styles, "fg", &style.TextColor)
	assignColor(styles, "ph", &style.PlaceholderColor)
	assignColor(styles, "bg", &style.Background)
	assignColor(styles, "border", &style.BorderColor)
	assignColor(styles, "hoverBorder", &style.HoverBorder)
	assignColor(styles, "focusBorder", &style.FocusBorder)
	assignColor(styles, "disabledFg", &style.DisabledText)
	assignColor(styles, "disabledBg", &style.DisabledBg)
	assignColor(styles, "caret", &style.CaretColor)
	assignInt(styles, "radius", &style.CornerRadius)
	assignInt(styles, "pad", &style.PaddingDP)
	return style, nil
}

func parsePanelStyle(raw json.RawMessage) (widgets.PanelStyle, error) {
	styles, err := decodeStyleMap(raw)
	if err != nil {
		return widgets.PanelStyle{}, err
	}
	style := widgets.PanelStyle{}
	if styles == nil {
		return style, nil
	}
	assignColor(styles, "bg", &style.Background)
	assignColor(styles, "border", &style.BorderColor)
	assignInt(styles, "borderW", &style.BorderWidth)
	assignInt(styles, "radius", &style.CornerRadius)
	return style, nil
}

func assignColor(styles styleMap, key string, target *widgets.Color) error {
	if styles == nil || target == nil {
		return nil
	}
	raw := styles[key]
	if len(raw) == 0 {
		return nil
	}
	text, err := decodeStringLiteral(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	value, ok, err := parseColorValue(text)
	if err != nil {
		return err
	}
	if ok {
		*target = value
	}
	return nil
}

func assignInt(styles styleMap, key string, target *int32) error {
	if styles == nil || target == nil {
		return nil
	}
	raw := styles[key]
	if len(raw) == 0 {
		return nil
	}
	value, err := decodeInt32Literal(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	*target = value
	return nil
}

func assignBool(styles styleMap, key string, target *bool) error {
	if styles == nil || target == nil {
		return nil
	}
	raw := styles[key]
	if len(raw) == 0 {
		return nil
	}
	value, err := decodeBoolLiteral(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	*target = value
	return nil
}

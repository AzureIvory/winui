//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AzureIvory/winui/widgets"
)

type demoLang struct {
	locales map[string]map[string]any
}

func loadDemoLang(path string) (*demoLang, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	locales := map[string]map[string]any{}
	for _, code := range []string{"en", "zh"} {
		entry, ok := raw[code]
		if !ok {
			continue
		}
		table, ok := entry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("locale %q is not an object", code)
		}
		locales[code] = table
	}
	if len(locales) == 0 {
		return nil, fmt.Errorf("lang file does not contain en/zh locales")
	}
	if locales["en"] == nil {
		locales["en"] = locales["zh"]
	}
	return &demoLang{locales: locales}, nil
}

func normalizeDemoLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "zh", "zh-cn", "zh_hans":
		return "zh"
	default:
		return "en"
	}
}

func (l *demoLang) text(locale string, path string, fallback string, args ...any) string {
	value, ok := l.stringValue(locale, path)
	if !ok || strings.TrimSpace(value) == "" {
		value = fallback
	}
	if len(args) == 0 {
		return value
	}
	if strings.Contains(value, "%") {
		return fmt.Sprintf(value, args...)
	}
	return value
}

func (l *demoLang) stringValue(locale string, path string) (string, bool) {
	raw, ok := l.value(locale, path)
	if !ok {
		return "", false
	}
	text, ok := raw.(string)
	return text, ok
}

func (l *demoLang) listItems(locale string, path string, fallback []widgets.ListItem) []widgets.ListItem {
	raw, ok := l.value(locale, path)
	if !ok {
		return append([]widgets.ListItem(nil), fallback...)
	}
	nodes, ok := raw.([]any)
	if !ok {
		return append([]widgets.ListItem(nil), fallback...)
	}
	items := make([]widgets.ListItem, 0, len(nodes))
	for _, node := range nodes {
		entry, ok := node.(map[string]any)
		if !ok {
			continue
		}
		item := widgets.ListItem{}
		if value, ok := entry["value"].(string); ok {
			item.Value = value
		}
		if text, ok := entry["text"].(string); ok {
			item.Text = text
		}
		if disabled, ok := entry["disabled"].(bool); ok {
			item.Disabled = disabled
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return append([]widgets.ListItem(nil), fallback...)
	}
	return items
}

func (l *demoLang) value(locale string, path string) (any, bool) {
	if l == nil {
		return nil, false
	}
	canonical := normalizeDemoLocale(locale)
	visited := map[string]struct{}{}
	for _, code := range []string{canonical, "en"} {
		if _, ok := visited[code]; ok {
			continue
		}
		visited[code] = struct{}{}
		root := l.locales[code]
		if root == nil {
			continue
		}
		if value, ok := lookupLangPath(root, path); ok {
			return value, true
		}
	}
	return nil, false
}

func lookupLangPath(root map[string]any, path string) (any, bool) {
	current := any(root)
	for _, part := range strings.Split(path, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, false
		}
		table, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := table[part]
		if !ok {
			return nil, false
		}
		current = next
	}
	return current, true
}

func normalizeGoMode(mode widgets.ControlMode) widgets.ControlMode {
	if mode == widgets.ModeNative {
		return widgets.ModeNative
	}
	return widgets.ModeCustom
}

func goModeLabel(lang *demoLang, locale string, mode widgets.ControlMode) string {
	mode = normalizeGoMode(mode)
	if mode == widgets.ModeNative {
		if lang != nil {
			return lang.text(locale, "mode.native", "Native controls")
		}
		return "Native controls"
	}
	if lang != nil {
		return lang.text(locale, "mode.custom", "Custom draw")
	}
	return "Custom draw"
}

func goModeButtonText(lang *demoLang, locale string, mode widgets.ControlMode) string {
	label := goModeLabel(lang, locale, mode)
	tpl := "Mode: %s"
	if lang != nil {
		tpl = lang.text(locale, "i18n.header.toggleControlModeBtn", tpl)
	}
	if strings.Contains(tpl, "%") {
		return fmt.Sprintf(tpl, label)
	}
	return tpl
}

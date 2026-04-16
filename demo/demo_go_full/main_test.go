//go:build windows

package main

import (
	"testing"

	"github.com/AzureIvory/winui/widgets"
)

func TestNormalizeGoMode(t *testing.T) {
	if got := normalizeGoMode(widgets.ModeNative); got != widgets.ModeNative {
		t.Fatalf("normalizeGoMode(ModeNative) = %v, want %v", got, widgets.ModeNative)
	}
	if got := normalizeGoMode(widgets.ControlMode(255)); got != widgets.ModeCustom {
		t.Fatalf("normalizeGoMode(invalid) = %v, want %v", got, widgets.ModeCustom)
	}
}

func TestGoModeButtonTextUsesLocaleAndFallback(t *testing.T) {
	lang := &demoLang{
		locales: map[string]map[string]any{
			"en": {
				"mode": map[string]any{
					"custom": "Custom draw",
					"native": "Native controls",
				},
				"i18n": map[string]any{
					"header": map[string]any{
						"toggleControlModeBtn": "Mode: %s",
					},
				},
			},
		},
	}

	if got := goModeLabel(lang, "en", widgets.ModeNative); got != "Native controls" {
		t.Fatalf("goModeLabel(native) = %q, want %q", got, "Native controls")
	}
	if got := goModeButtonText(lang, "en", widgets.ModeCustom); got != "Mode: Custom draw" {
		t.Fatalf("goModeButtonText(custom) = %q, want %q", got, "Mode: Custom draw")
	}
	if got := goModeButtonText(nil, "en", widgets.ModeCustom); got != "Mode: Custom draw" {
		t.Fatalf("goModeButtonText(nil lang) = %q, want %q", got, "Mode: Custom draw")
	}
}

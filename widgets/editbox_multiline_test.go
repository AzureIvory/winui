//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestEditBoxSingleLineBehaviorRegression(t *testing.T) {
	edit := NewEditBox("single", ModeCustom)
	submits := 0
	edit.SetOnSubmit(func(string) {
		submits++
	})

	edit.SetText("hello\nworld")
	if got := edit.TextValue(); got != "hello world" {
		t.Fatalf("expected single-line SetText to collapse newlines, got %q", got)
	}

	edit.OnEvent(Event{Type: EventFocus})
	edit.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyReturn}})
	if submits != 1 {
		t.Fatalf("expected Enter to submit in single-line mode, got %d", submits)
	}
}

func TestEditBoxMultilineEnterAndCtrlEnter(t *testing.T) {
	edit := NewEditBox("multi", ModeCustom)
	edit.SetMultiline(true)
	submits := 0
	edit.SetOnSubmit(func(string) {
		submits++
	})

	edit.OnEvent(Event{Type: EventFocus})
	edit.OnEvent(Event{Type: EventChar, Rune: 'A'})
	edit.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyReturn}})
	edit.OnEvent(Event{Type: EventChar, Rune: 'B'})

	if got := edit.TextValue(); got != "A\nB" {
		t.Fatalf("expected multiline Enter to insert newline, got %q", got)
	}
	if submits != 0 {
		t.Fatalf("expected plain Enter not to submit in multiline mode, got %d", submits)
	}
	if edit.LineCount() != 2 {
		t.Fatalf("expected 2 logical lines, got %d", edit.LineCount())
	}

	edit.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyReturn, Flags: editKeyFlagCtrl}})
	if submits != 1 {
		t.Fatalf("expected Ctrl+Enter to submit in multiline mode, got %d", submits)
	}
}

func TestEditBoxMultilineSubmitWhenAcceptReturnDisabled(t *testing.T) {
	edit := NewEditBox("multi-submit", ModeCustom)
	edit.SetMultiline(true)
	edit.SetAcceptReturn(false)
	edit.SetText("hello")
	submits := 0
	edit.SetOnSubmit(func(string) {
		submits++
	})

	edit.OnEvent(Event{Type: EventFocus})
	edit.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyReturn}})

	if got := edit.TextValue(); got != "hello" {
		t.Fatalf("expected text unchanged when AcceptReturn is false, got %q", got)
	}
	if submits != 1 {
		t.Fatalf("expected Enter to submit when AcceptReturn is false, got %d", submits)
	}
}

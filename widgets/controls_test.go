//go:build windows

package widgets

import (
	"github.com/yourname/winui/core"
	"testing"
)

// TestCheckBoxToggle 测试复选框切换。
func TestCheckBoxToggle(t *testing.T) {
	check := NewCheckBox("check", "demo")

	changed := false
	check.SetOnChange(func(checked bool) {
		changed = checked
	})

	check.OnEvent(Event{Type: EventClick})
	if !check.IsChecked() {
		t.Fatalf("expected checkbox checked after click")
	}
	if !changed {
		t.Fatalf("expected checkbox change callback")
	}
}

// TestRadioButtonGroup 测试单选按钮分组。
func TestRadioButtonGroup(t *testing.T) {
	panel := NewPanel("root")
	left := NewRadioButton("left", "Left")
	right := NewRadioButton("right", "Right")
	left.SetGroup("mode")
	right.SetGroup("mode")

	panel.Add(left)
	panel.Add(right)

	left.SetChecked(true)
	right.OnEvent(Event{Type: EventClick})
	if !right.IsChecked() {
		t.Fatalf("expected clicked radio selected")
	}
	if left.IsChecked() {
		t.Fatalf("expected peer radio cleared")
	}
}

// TestComboBoxSelectByClick 测试组合框点击选择。
func TestComboBoxSelectByClick(t *testing.T) {
	combo := NewComboBox("combo")
	combo.SetBounds(Rect{X: 0, Y: 0, W: 200, H: 40})
	combo.SetItems([]ListItem{
		{Value: "a", Text: "Alpha"},
		{Value: "b", Text: "Beta"},
	})

	combo.OnEvent(Event{Type: EventClick, Point: core.Point{X: 10, Y: 10}})
	if !combo.open {
		t.Fatalf("expected combo open after base click")
	}

	point := core.Point{X: 10, Y: combo.popupRect().Y + combo.dp(12)}
	combo.OnEvent(Event{Type: EventClick, Point: point})
	if combo.SelectedIndex() != 0 {
		t.Fatalf("expected first item selected, got %d", combo.SelectedIndex())
	}
	if combo.open {
		t.Fatalf("expected combo close after selection")
	}
}

// TestEditBoxTyping 测试编辑框输入。
func TestEditBoxTyping(t *testing.T) {
	edit := NewEditBox("edit")
	edit.OnEvent(Event{Type: EventFocus})
	edit.OnEvent(Event{Type: EventChar, Rune: 'A'})
	edit.OnEvent(Event{Type: EventChar, Rune: 'B'})
	edit.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyBack}})

	if edit.TextValue() != "A" {
		t.Fatalf("expected text A, got %q", edit.TextValue())
	}
}

// TestListBoxKeyboardSelect 测试列表框键盘选择。
func TestListBoxKeyboardSelect(t *testing.T) {
	list := NewListBox("list")
	list.SetItems([]ListItem{
		{Value: "1", Text: "One"},
		{Value: "2", Text: "Two"},
	})
	list.OnEvent(Event{Type: EventFocus})
	list.OnEvent(Event{Type: EventKeyDown, Key: core.KeyEvent{Key: core.KeyDown}})

	if list.SelectedIndex() != 0 {
		t.Fatalf("expected keyboard to select first item, got %d", list.SelectedIndex())
	}
}

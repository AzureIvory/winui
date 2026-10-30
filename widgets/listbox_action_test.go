//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

type listActionEvents struct {
	actions, changes, activations, checks int
	index                                 int
	item                                  ListItem
}

func newListActionTest(t *testing.T) (*Scene, *Panel, *ListBox, *listActionEvents) {
	t.Helper()
	scene := NewScene(nil)
	parent := NewPanel("parent")
	list := NewListBox("phrases")
	list.SetBounds(Rect{X: 20, Y: 30, W: 320, H: 120})
	list.SetStyle(ListStyle{ItemHeightDP: 40, PaddingDP: 4})
	list.SetItems([]ListItem{{Value: "first", Text: "第一条"}, {Value: "second", Text: "第二条"}})
	list.SetShowCheck(true)
	events := &listActionEvents{}
	list.SetItemAction(&core.Image{}, func(index int, item ListItem) {
		events.actions++
		events.index, events.item = index, item
	})
	list.SetOnChange(func(int, ListItem) { events.changes++ })
	list.SetOnActivate(func(int, ListItem) { events.activations++ })
	list.SetOnCheckChange(func(int, ListItem) { events.checks++ })
	parent.Add(list)
	scene.Root().Add(parent)
	return scene, parent, list, events
}

func listActionMouse(scene *Scene, point core.Point, down bool) {
	event := core.MouseEvent{Point: point, Button: core.MouseButtonLeft}
	if down {
		scene.DispatchMouseMove(event)
		scene.DispatchMouseDown(event)
	} else {
		scene.DispatchMouseUp(event)
	}
}

func TestListBoxItemActionMouseIsIndependentOfSelectionActivationAndChecks(t *testing.T) {
	scene, _, list, events := newListActionTest(t)
	list.SetSelected(1)
	for i := 0; i < 2; i++ {
		listActionMouse(scene, core.Point{X: 310, Y: 54}, true)
		listActionMouse(scene, core.Point{X: 310, Y: 54}, false)
	}
	if events.actions != 2 || events.index != 0 || events.item.Value != "first" ||
		list.SelectedIndex() != 1 || events.changes != 0 || events.activations != 0 || events.checks != 0 {
		t.Fatalf("action mixed with row behavior: events=%+v selected=%d", events, list.SelectedIndex())
	}
	// The independent checkbox occupies the slot immediately to the action's left.
	listActionMouse(scene, core.Point{X: 285, Y: 54}, true)
	listActionMouse(scene, core.Point{X: 285, Y: 54}, false)
	if !list.IsItemChecked(0) || events.checks != 1 || events.actions != 2 || list.SelectedIndex() != 1 {
		t.Fatalf("checkbox collided with item action: events=%+v", events)
	}
	for i := 0; i < 2; i++ {
		listActionMouse(scene, core.Point{X: 60, Y: 54}, true)
		listActionMouse(scene, core.Point{X: 60, Y: 54}, false)
	}
	if events.actions != 2 || events.activations != 1 || events.changes != 1 {
		t.Fatalf("body double-click lost ordinary activation: %+v", events)
	}
}

func TestListBoxItemActionCanceledGesturesCannotBecomeRowClicks(t *testing.T) {
	for _, name := range []string{
		"different-row", "body-release", "leave-and-return", "items-reordered", "items-refreshed",
		"scroll", "hidden-and-shown", "disabled-and-enabled", "ancestor-hidden-and-shown",
		"ancestor-disabled-and-enabled", "bounds-moved-back", "ancestor-bounds-moved-back",
		"action-hidden-and-restored", "action-disabled-and-enabled", "check-column-changed", "detached-and-reattached",
	} {
		t.Run(name, func(t *testing.T) {
			scene, parent, list, events := newListActionTest(t)
			list.SetSelected(1)
			if name == "scroll" {
				list.SetBounds(Rect{X: 20, Y: 30, W: 320, H: 50})
			}
			point := core.Point{X: 310, Y: 54}
			listActionMouse(scene, point, true)
			switch name {
			case "different-row":
				point.Y += 40
			case "body-release":
				point.X = 60
			case "leave-and-return":
				scene.DispatchMouseMove(core.MouseEvent{Point: core.Point{X: 5, Y: 5}})
				scene.DispatchMouseMove(core.MouseEvent{Point: point})
			case "items-reordered":
				items := list.Items()
				items[0], items[1] = items[1], items[0]
				list.SetItems(items)
			case "items-refreshed":
				list.SetItems(list.Items())
			case "scroll":
				scene.DispatchMouseWheel(core.MouseEvent{Point: point, Delta: -120})
			case "hidden-and-shown":
				list.SetVisible(false)
				list.SetVisible(true)
			case "disabled-and-enabled":
				list.SetEnabled(false)
				list.SetEnabled(true)
			case "ancestor-hidden-and-shown":
				parent.SetVisible(false)
				parent.SetVisible(true)
			case "ancestor-disabled-and-enabled":
				parent.SetEnabled(false)
				parent.SetEnabled(true)
			case "bounds-moved-back":
				bounds := list.Bounds()
				list.SetBounds(offsetRect(bounds, 0, 1))
				list.SetBounds(bounds)
			case "ancestor-bounds-moved-back":
				bounds := parent.Bounds()
				parent.SetBounds(offsetRect(bounds, 0, 1))
				parent.SetBounds(bounds)
			case "action-hidden-and-restored":
				callback := list.OnItemAction
				list.SetItemAction(nil, callback)
				list.SetItemAction(&core.Image{}, callback)
			case "action-disabled-and-enabled":
				list.SetItemActionEnabled(false)
				list.SetItemActionEnabled(true)
			case "check-column-changed":
				list.SetShowCheck(false)
			case "detached-and-reattached":
				parent.Remove(list.ID())
				parent.Add(list)
			}
			listActionMouse(scene, point, false)
			if events.actions != 0 || events.changes != 0 || events.activations != 0 || events.checks != 0 || list.SelectedIndex() != 1 {
				t.Fatalf("canceled action triggered row behavior: events=%+v selected=%d", events, list.SelectedIndex())
			}
		})
	}
}

func TestListBoxItemActionUsesSceneClippingAndCancelsAncestorScroll(t *testing.T) {
	scene := NewScene(nil)
	scroll := NewScrollView("viewport")
	scroll.SetBounds(Rect{X: 20, Y: 30, W: 320, H: 65})
	content := scrollTestContent("content", 320, 200)
	list := NewListBox("phrases")
	list.SetBounds(Rect{X: 0, Y: 0, W: 320, H: 160})
	scrollTestPlace(list, 0, 0, 320, 160)
	list.SetStyle(ListStyle{ItemHeightDP: 40, PaddingDP: 4})
	list.SetItems([]ListItem{{Value: "first"}, {Value: "second"}, {Value: "third"}})
	actions := 0
	list.SetItemAction(&core.Image{}, func(int, ListItem) { actions++ })
	content.Add(list)
	scroll.SetContent(content)
	scene.Root().Add(scroll)
	point := core.Point{X: 310, Y: 54}
	listActionMouse(scene, point, true)
	scroll.ScrollTo(0, 20)
	scroll.ScrollTo(0, 0)
	listActionMouse(scene, point, false)
	if actions != 0 {
		t.Fatal("ancestor fast scroll revived an action gesture")
	}
	edge := core.Point{X: 310, Y: 84}
	listActionMouse(scene, edge, true)
	// This point remains inside the second list row, below the ancestor clip.
	clipped := core.Point{X: 310, Y: 105}
	listActionMouse(scene, clipped, false)
	if actions != 0 {
		t.Fatal("action fired outside the ancestor clip")
	}
	listActionMouse(scene, point, true)
	listActionMouse(scene, point, false)
	if actions != 1 {
		t.Fatal("visible action stopped working after clipping cancellation")
	}
}

func TestListBoxItemActionUnrelatedWidgetMutationDoesNotCancelCapture(t *testing.T) {
	scene, _, _, events := newListActionTest(t)
	other := NewPanel("unrelated")
	scene.Root().Add(other)
	point := core.Point{X: 310, Y: 54}
	listActionMouse(scene, point, true)
	other.SetBounds(Rect{X: 500, Y: 500, W: 20, H: 20})
	other.SetVisible(false)
	other.SetEnabled(false)
	listActionMouse(scene, point, false)
	if events.actions != 1 {
		t.Fatal("unrelated widget mutation canceled another subtree's action")
	}
}

func TestListBoxItemActionCallbackCanReplaceItemsAndRegistration(t *testing.T) {
	scene, _, list, events := newListActionTest(t)
	list.SetItemAction(&core.Image{}, func(index int, item ListItem) {
		events.actions++
		if index != 0 || item.Value != "first" {
			t.Errorf("callback received a changed identity: %d %+v", index, item)
		}
		list.SetItems([]ListItem{{Value: "replacement"}})
		list.SetItemAction(nil, nil)
	})
	listActionMouse(scene, core.Point{X: 310, Y: 54}, true)
	listActionMouse(scene, core.Point{X: 310, Y: 54}, false)
	if events.actions != 1 || list.Items()[0].Value != "replacement" {
		t.Fatal("reentrant action callback lost or repeated its mutation")
	}
}

func TestListBoxItemActionLayoutKeepsTextChecksAndScrollbarSeparate(t *testing.T) {
	_, _, list, _ := newListActionTest(t)
	for _, height := range []int32{120, 50} {
		list.SetBounds(Rect{X: 20, Y: 30, W: 320, H: height})
		style := list.resolveStyle(nil)
		row := list.rowRect(0, &PaintCtx{}, style)
		text, check, action := list.rowContentRects(row, style)
		if action.W != 32 || text.X+text.W > check.X || check.X+check.W > action.X {
			t.Errorf("row content overlaps: text=%+v check=%+v action=%+v", text, check, action)
		}
		if height == 50 && action.X+action.W >= list.Bounds().X+list.Bounds().W-8 {
			t.Error("action occupies the scrollbar track")
		}
		list.SetItemActionEnabled(false)
		_, _, disabledAction := list.rowContentRects(row, style)
		if disabledAction != action {
			t.Error("disabling action removed its reserved slot")
		}
		list.SetItemActionEnabled(true)
	}
}

func TestListBoxItemActionDisabledOrMissingCannotFire(t *testing.T) {
	for _, name := range []string{"action-disabled", "item-disabled", "image-missing", "callback-missing"} {
		t.Run(name, func(t *testing.T) {
			scene, _, list, events := newListActionTest(t)
			switch name {
			case "action-disabled":
				list.SetItemActionEnabled(false)
			case "item-disabled":
				items := list.Items()
				items[0].Disabled = true
				list.SetItems(items)
			case "image-missing":
				list.SetItemAction(nil, list.OnItemAction)
			case "callback-missing":
				list.SetItemAction(&core.Image{}, nil)
			}
			listActionMouse(scene, core.Point{X: 310, Y: 54}, true)
			listActionMouse(scene, core.Point{X: 310, Y: 54}, false)
			if events.actions != 0 {
				t.Fatal("unavailable item action fired")
			}
			if (name == "action-disabled" || name == "item-disabled") && (events.changes != 0 || events.checks != 0) {
				t.Fatal("disabled action fell through to row selection/check")
			}
		})
	}
}

func TestListBoxItemActionCtrlEnterFiresOnceAndPreservesEnter(t *testing.T) {
	scene, _, list, events := newListActionTest(t)
	list.SetSelected(1)
	scene.FocusWidget(list)
	if !scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Ctrl: true}) || events.actions != 1 || events.index != 1 || events.item.Value != "second" {
		t.Fatalf("Ctrl+Enter did not invoke selected action: %+v", events)
	}
	for _, key := range []core.KeyEvent{
		{Key: core.KeyReturn, Ctrl: true, Repeat: true},
		{Key: core.KeyReturn, Ctrl: true, Flags: 1 << 30},
		{Key: core.KeyReturn, Ctrl: true, Alt: true},
		{Key: core.KeyReturn, Ctrl: true, Shift: true},
		{Key: core.KeyReturn, Ctrl: true, Flags: 1 << 29},
		{Key: core.KeyReturn, Ctrl: true, Composing: true},
	} {
		scene.DispatchKeyDown(key)
	}
	if events.actions != 1 || events.changes != 0 || events.activations != 0 || events.checks != 0 {
		t.Fatalf("repeat/modifiers/IME leaked into an action: %+v", events)
	}
	if list.handleKeyBeforeTranslate(core.KeyEvent{Key: core.KeyReturn, Ctrl: true, Composing: true}) {
		t.Fatal("action pretranslation consumed an IME key")
	}
	scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn})
	if events.actions != 1 || events.activations != 1 || events.changes != 1 {
		t.Fatalf("ordinary Enter lost existing activation: %+v", events)
	}
	list.SetItemActionEnabled(false)
	if scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Ctrl: true}) || events.actions != 1 || events.activations != 1 {
		t.Fatal("disabled Ctrl+Enter fired or fell through to activation")
	}
}

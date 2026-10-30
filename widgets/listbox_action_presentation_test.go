//go:build windows

package widgets

import (
	"testing"

	"github.com/AzureIvory/winui/core"
)

func TestListBoxItemActionPresentationKeepsStableSeparateGeometry(t *testing.T) {
	_, _, list, _ := newListActionTest(t)
	presentation := ItemActionPresentation{Text: "Send"}
	provider := func(int, ListItem) ItemActionPresentation { return presentation }
	for _, height := range []int32{120, 50} {
		list.SetBounds(Rect{X: 20, Y: 30, W: 320, H: height})
		style := list.resolveStyle(nil)
		row := list.rowRect(0, &PaintCtx{}, style)
		list.SetItemActionPresentation(nil)
		_, _, legacy := list.rowContentRects(row, style)
		if legacy.W != 32 {
			t.Fatalf("legacy action width=%d want=32", legacy.W)
		}
		var previous Rect
		for _, state := range []ActionVisualState{
			ActionVisualDefault, ActionVisualWorking, ActionVisualSuccess, ActionVisualError,
		} {
			presentation.State = state
			presentation.Text += " longer caption"
			presentation.Disabled = state != ActionVisualDefault
			list.SetItemActionPresentation(provider)
			text, check, action := list.rowContentRects(row, style)
			if action.W != 76 || text.X+text.W > check.X || check.X+check.W > action.X {
				t.Fatalf("presentation columns overlap: text=%+v check=%+v action=%+v", text, check, action)
			}
			if previous != (Rect{}) && previous != action {
				t.Fatal("caption/state/availability changed the reserved width")
			}
			if height == 50 && action.X+action.W >= list.Bounds().X+list.Bounds().W-8 {
				t.Fatal("presentation action overlaps the scrollbar")
			}
			previous = action
		}
		list.SetItemActionPresentation(nil)
		_, _, restored := list.rowContentRects(row, style)
		if restored != legacy {
			t.Fatalf("nil provider did not restore legacy geometry: %+v want=%+v", restored, legacy)
		}
	}
	list.SetItemActionPresentation(provider)
	list.SetItemAction(nil, nil)
	_, _, action := list.rowContentRects(list.Bounds(), list.resolveStyle(nil))
	if !action.Empty() {
		t.Fatal("presentation provider registered an action without image/callback")
	}
}

func TestListBoxItemActionPresentationAvailabilityFollowsCurrentValue(t *testing.T) {
	scene, _, list, events := newListActionTest(t)
	list.SetItemActionPresentation(func(_ int, item ListItem) ItemActionPresentation {
		return ItemActionPresentation{
			Text: "Working", State: ActionVisualWorking, Disabled: item.Value == "first",
		}
	})
	list.SetSelected(1)
	point := core.Point{X: 300, Y: 54}
	for range 2 {
		listActionMouse(scene, point, true)
		listActionMouse(scene, point, false)
	}
	if events.actions != 0 || events.changes != 0 || events.checks != 0 ||
		events.activations != 0 || list.SelectedIndex() != 1 {
		t.Fatalf("disabled row action leaked into row behavior: %+v selected=%d", events, list.SelectedIndex())
	}
	scene.FocusWidget(list)
	list.SetSelected(0)
	if scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Ctrl: true}) || events.actions != 0 {
		t.Fatal("per-row Disabled failed to prevent Ctrl+Enter")
	}
	list.SetSelected(1)
	if !scene.DispatchKeyDown(core.KeyEvent{Key: core.KeyReturn, Ctrl: true}) ||
		events.actions != 1 || events.item.Value != "second" {
		t.Fatalf("Working implicitly disabled an eligible row: %+v", events)
	}
	items := list.Items()
	items[0], items[1] = items[1], items[0]
	list.SetItems(items)
	listActionMouse(scene, point, true)
	listActionMouse(scene, point, false)
	if events.actions != 2 || events.index != 0 || events.item.Value != "second" {
		t.Fatalf("presentation retained the previous row identity after reorder: %+v", events)
	}
	point.Y += 40
	listActionMouse(scene, point, true)
	listActionMouse(scene, point, false)
	if events.actions != 2 || events.changes != 0 || events.activations != 0 || events.checks != 0 {
		t.Fatalf("reordered disabled value became actionable: %+v", events)
	}
}

func TestListBoxItemActionPresentationRefreshPreservesEligibleGesture(t *testing.T) {
	for _, replaceProvider := range []bool{false, true} {
		scene, _, list, events := newListActionTest(t)
		presentation := ItemActionPresentation{Text: "Send"}
		provider := func(int, ListItem) ItemActionPresentation { return presentation }
		list.SetItemActionPresentation(provider)
		point := core.Point{X: 310, Y: 54}
		listActionMouse(scene, point, true)
		presentation.Text = "Done"
		presentation.State = ActionVisualSuccess
		if replaceProvider {
			provider = func(int, ListItem) ItemActionPresentation { return presentation }
		}
		list.SetItemActionPresentation(provider)
		listActionMouse(scene, point, false)
		if events.actions != 1 || events.changes != 0 || events.checks != 0 || events.activations != 0 {
			t.Fatalf("pure presentation refresh swallowed or duplicated an eligible action: %+v", events)
		}
	}
}

func TestListBoxItemActionPresentationCancelsChangedGeometryOrAvailability(t *testing.T) {
	for _, name := range []string{"provider-added", "provider-removed", "disabled-and-enabled", "disabled-before-release"} {
		t.Run(name, func(t *testing.T) {
			scene, _, list, events := newListActionTest(t)
			presentation := ItemActionPresentation{Text: "Send"}
			provider := func(int, ListItem) ItemActionPresentation { return presentation }
			if name != "provider-added" {
				list.SetItemActionPresentation(provider)
			}
			list.SetSelected(1)
			point := core.Point{X: 310, Y: 54}
			listActionMouse(scene, point, true)
			switch name {
			case "provider-added":
				list.SetItemActionPresentation(provider)
			case "provider-removed":
				list.SetItemActionPresentation(nil)
			case "disabled-and-enabled":
				presentation.Disabled = true
				list.SetItemActionPresentation(provider)
				presentation.Disabled = false
				list.SetItemActionPresentation(provider)
			case "disabled-before-release":
				presentation.Disabled = true
			}
			listActionMouse(scene, point, false)
			if events.actions != 0 || events.changes != 0 || events.checks != 0 ||
				events.activations != 0 || list.SelectedIndex() != 1 {
				t.Fatalf("presentation update revived a pointer gesture: %+v selected=%d", events, list.SelectedIndex())
			}
		})
	}
}

func TestActionButtonStyleKeepsFeedbackReadableWhenUnavailable(t *testing.T) {
	seen := make(map[core.Color]bool)
	for _, state := range []ActionVisualState{
		ActionVisualDefault, ActionVisualWorking, ActionVisualSuccess, ActionVisualError,
	} {
		style := ActionButtonStyle(state)
		if seen[style.Background] {
			t.Fatalf("visual state %d shares another state's background", state)
		}
		seen[style.Background] = true
		if state != ActionVisualDefault &&
			(style.Disabled != style.Background || style.DisabledText != style.TextColor) {
			t.Fatalf("disabled feedback state %d lost its status appearance", state)
		}
		button := NewButton("shared", "Send", ModeCustom)
		button.SetStyle(style)
		resolved := button.resolveStyle(nil)
		if resolved.Background != style.Background || resolved.DisabledText != style.DisabledText {
			t.Fatal("shared action style is not preserved by Button's merge path")
		}
	}
	if ActionButtonStyle(ActionVisualState(255)) != ActionButtonStyle(ActionVisualDefault) {
		t.Fatal("unknown visual state did not fall back to the default appearance")
	}
}

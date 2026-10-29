//go:build windows

package widgets

import (
	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

// App 返回场景绑定的底层应用实例。
func (s *Scene) App() *core.App {
	if s == nil {
		return nil
	}
	return s.app
}

// SceneOf 返回控件当前所属的场景；未附着时返回 nil。
func SceneOf(widget Widget) *Scene {
	node := asWidgetNode(widget)
	if node == nil {
		return nil
	}
	return node.scene()
}

// FocusWidget focuses an enabled, visible, focusable widget in this scene.
// Hidden/disabled ancestors and widgets from another scene are ignored.
// Ancestor ScrollViews reveal the target before focus transfer. Nil clears widget
// focus and returns native keyboard focus to the app host. Scheduling follows the UI thread.
// Focusing an editor transfers keyboard focus to its native RichEdit control.
func (s *Scene) FocusWidget(widget Widget) {
	if s == nil {
		return
	}
	s.runOnUI(func() {
		s.focusWidget(widget, true)
	})
}

// FocusNext selects the next eligible widget in depth-first child order, wrapping
// within the topmost visible, enabled Modal, or the scene if no modal is open.
// Ancestor ScrollViews reveal the selected target before focus transfer.
// Reverse traverses in the opposite direction. It returns false if no eligible
// widget exists or the caller is not on the scene's UI thread. Call from a key
// preview to opt into Tab traversal; this method installs no default key binding.
func (s *Scene) FocusNext(reverse bool) bool {
	if s == nil || s.root == nil || (s.app != nil && !s.app.IsUIThread()) {
		return false
	}
	scope := Widget(s.root)
	if modal := topmostVisibleWidget(s.root, func(widget Widget) bool {
		_, ok := widget.(*Modal)
		return ok && SceneOf(widget) == s
	}); modal != nil {
		scope = modal
	}
	var candidates []Widget
	Walk(scope, func(widget Widget) bool {
		if s.canFocusWidget(widget) {
			candidates = append(candidates, widget)
		}
		return true
	})
	count := len(candidates)
	if count == 0 {
		return false
	}
	next := 0
	if reverse {
		next = count - 1
	}
	for index, widget := range candidates {
		if widget == s.focus {
			next = (index + 1) % count
			if reverse {
				next = (index + count - 1) % count
			}
			break
		}
	}
	return s.focusWidget(candidates[next], true)
}

func (s *Scene) canFocusWidget(widget Widget) bool {
	if widget == nil || SceneOf(widget) != s {
		return false
	}
	focusable, ok := widget.(focusableWidget)
	if (!ok || !focusable.acceptsFocus()) && s.nativeHandleFor(widget) == 0 {
		return false
	}
	path := s.pathTo(widget)
	if len(path) == 0 || path[0] != s.root {
		return false
	}
	for _, ancestor := range path {
		if SceneOf(ancestor) != s || !ancestor.Visible() || !ancestor.Enabled() {
			return false
		}
	}
	return true
}

func (s *Scene) focusWidget(widget Widget, reveal bool) bool {
	if widget != nil && !s.canFocusWidget(widget) {
		return false
	}
	if widget != nil && reveal {
		s.revealWidget(widget)
	}
	s.setFocus(widget)
	if handle := s.nativeHandleFor(widget); handle != 0 {
		setNativeFocus(handle)
	} else if s.app != nil {
		setNativeFocus(s.app.Handle())
	}
	return s.focus == widget
}

// revealWidget works from inner to outer viewport, revealing only the portion
// exposed by each inner clip. Every offset still belongs to its ScrollView.
func (s *Scene) revealWidget(widget Widget) {
	path := s.pathTo(widget)
	for _, ancestor := range path[:len(path)-1] {
		if scroll, ok := ancestor.(*ScrollView); ok {
			scroll.layoutContent()
		}
	}
	visible := widget.Bounds()
	for index := len(path) - 2; index >= 0 && !visible.Empty(); index-- {
		ancestor := path[index]
		if scroll, ok := ancestor.(*ScrollView); ok {
			oldX, oldY := scroll.ScrollOffset()
			scroll.revealRect(visible)
			visible.X += oldX - scroll.offsetX
			visible.Y += oldY - scroll.offsetY
		}
		if clipper, ok := ancestor.(clipBoundsWidget); ok {
			visible = intersectRect(visible, clipper.clipBounds())
		}
	}
}

// SetKeyPreview sets the scene-wide shortcut handler. True consumes the key
// before it reaches a widget or produces a character. IME composition bypasses
// shortcuts. Use BindScene to include keys from native child controls.
func (s *Scene) SetKeyPreview(fn func(core.KeyEvent) bool) {
	if s != nil {
		s.runOnUI(func() { s.keyPreview = fn })
	}
}

func (s *Scene) previewKeyDown(ev core.KeyEvent) bool {
	if s == nil || s.keyComposing(ev) {
		return false
	}
	return s.keyPreview != nil && s.keyPreview(ev)
}

func (s *Scene) keyComposing(ev core.KeyEvent) bool {
	edit := s.nativeEditForKey(ev)
	return ev.Composing || (edit != nil && edit.imeComposing)
}

func (s *Scene) nativeEditForKey(ev core.KeyEvent) *EditBox {
	edit, _ := s.keyTarget(ev).(*EditBox)
	return edit
}

func (s *Scene) keyTarget(ev core.KeyEvent) Widget {
	if ev.Handle == 0 || (s.app != nil && ev.Handle == s.app.Handle()) {
		return s.focus
	}
	s.nativeMu.RLock()
	target, _ := s.nativeTargets[uintptr(ev.Handle)].(Widget)
	s.nativeMu.RUnlock()
	return target
}

// The native registration table is the single source of widget/handle identity.
func (s *Scene) nativeHandleFor(widget Widget) windows.Handle {
	s.nativeMu.RLock()
	defer s.nativeMu.RUnlock()
	for handle, handler := range s.nativeTargets {
		if target, ok := handler.(Widget); ok && target == widget {
			return windows.Handle(handle)
		}
	}
	return 0
}

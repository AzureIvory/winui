//go:build windows

package markup

import (
	"strings"

	"github.com/AzureIvory/winui/widgets"
)

type documentBinding struct {
	paths []string
	apply func(*bindingContext)
}

type bindingContext struct {
	document *Document
	scene    *widgets.Scene
	state    any
}

func (c *bindingContext) Lookup(path string) (any, bool) {
	if c == nil {
		return nil, false
	}
	return lookupStateValue(c.state, path)
}

func (b documentBinding) matches(paths []string) bool {
	if len(b.paths) == 0 || len(paths) == 0 {
		return true
	}
	for _, changed := range paths {
		changed = normalizeBindingPath(changed)
		if changed == "" {
			return true
		}
		for _, path := range b.paths {
			if bindingPathsOverlap(path, changed) {
				return true
			}
		}
	}
	return false
}

func bindingPathsOverlap(left string, right string) bool {
	left = normalizeBindingPath(left)
	right = normalizeBindingPath(right)
	if left == "" || right == "" {
		return true
	}
	return left == right ||
		strings.HasPrefix(left, right+".") ||
		strings.HasPrefix(right, left+".")
}

func normalizeBindingPath(path string) string {
	return strings.Trim(strings.TrimSpace(path), ".")
}

func splitBindingPath(path string) []string {
	path = normalizeBindingPath(path)
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (d *Document) setStateInternal(state *State) {
	if d == nil {
		return
	}

	d.mu.Lock()
	if d.unsubscribe != nil {
		d.unsubscribe()
		d.unsubscribe = nil
	}
	d.state = state
	if state != nil {
		d.unsubscribe = state.subscribe(func(change stateChange) {
			d.scheduleBindingRefresh(change.paths)
		})
	}
	d.mu.Unlock()
}

func (d *Document) currentScene() *widgets.Scene {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.scene
}

func (d *Document) bindingSnapshot() ([]documentBinding, *widgets.Scene, any) {
	if d == nil {
		return nil, nil, nil
	}

	d.mu.Lock()
	bindings := append([]documentBinding(nil), d.bindings...)
	scene := d.scene
	state := d.state
	d.mu.Unlock()

	var snapshot any
	if state != nil {
		snapshot = state.snapshot()
	}
	return bindings, scene, snapshot
}

func (d *Document) scheduleBindingRefresh(paths []string) {
	if d == nil {
		return
	}

	scene := d.currentScene()
	if scene != nil {
		if app := scene.App(); app != nil && !app.IsUIThread() {
			copied := append([]string(nil), paths...)
			_ = app.Post(func() {
				d.RefreshBindings(copied...)
			})
			return
		}
	}

	d.RefreshBindings(paths...)
}

// SetState connects the document to a binding state store and refreshes all
// declared bindings immediately.
func (d *Document) SetState(state *State) {
	d.setStateInternal(state)
	d.scheduleBindingRefresh(nil)
}

// State returns the state store currently bound to the document.
func (d *Document) State() *State {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// RefreshBindings reapplies declarative bindings using the current state
// snapshot. When paths are provided, unrelated bindings are skipped.
func (d *Document) RefreshBindings(paths ...string) {
	bindings, scene, snapshot := d.bindingSnapshot()
	if len(bindings) == 0 {
		return
	}

	ctx := &bindingContext{
		document: d,
		scene:    scene,
		state:    snapshot,
	}
	for _, binding := range bindings {
		if binding.apply == nil || !binding.matches(paths) {
			continue
		}
		binding.apply(ctx)
	}
}

func (d *Document) setWindowTitle(title string) {
	if d == nil {
		return
	}

	d.Meta.Title = title
	scene := d.currentScene()
	if scene == nil {
		return
	}
	if app := scene.App(); app != nil {
		app.SetTitle(title)
	}
}

//go:build windows

package jsonui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// ActionContext describes runtime information exposed to JSON UI action handlers.
type ActionContext struct {
	Name    string
	Window  *Window
	Widget  widgets.Widget
	ID      string
	Value   string
	Paths   []string
	Checked bool
	Index   int
	Item    widgets.ListItem
}

// WindowMeta stores window-level metadata declared in JSON.
type WindowMeta struct {
	ID          string
	Title       string
	Image       *core.Image
	ImagePath   string
	ImageSizeDP int32
	Width       int32
	Height      int32
	MinWidth    int32
	MinHeight   int32
	Background  core.Color
}

// LoadOptions controls how a JSON UI document is built.
type LoadOptions struct {
	ActionHandlers map[string]func(ActionContext)
	Actions        map[string]func()
	AssetsDir      string
	DefaultMode    widgets.ControlMode
	Data           DataSource
	Theme          *widgets.Theme
	ImageSizeDP    int32
}

// Window is one built top-level window from a JSON UI document.
type Window struct {
	ID   string
	Root widgets.Widget
	Meta WindowMeta

	theme       *widgets.Theme
	data        DataSource
	bindings    []windowBinding
	reloaders   []resourceReloader
	unsubscribe func()
	widgetIndex map[string]widgets.Widget

	mu    sync.Mutex
	scene *widgets.Scene
}

// ApplyOptions copies window metadata into a core.Options value.
func (w *Window) ApplyOptions(opts *core.Options) {
	if w == nil || opts == nil {
		return
	}
	if w.Meta.Title != "" {
		opts.Title = w.Meta.Title
	}
	if w.Meta.Image != nil {
		opts.WindowImage = w.Meta.Image
	}
	if w.Meta.ImageSizeDP > 0 {
		opts.WindowImageSizeDP = w.Meta.ImageSizeDP
	}
	if w.Meta.Width > 0 {
		opts.Width = w.Meta.Width
	}
	if w.Meta.Height > 0 {
		opts.Height = w.Meta.Height
	}
	if w.Meta.MinWidth > 0 {
		opts.MinWidth = w.Meta.MinWidth
	}
	if w.Meta.MinHeight > 0 {
		opts.MinHeight = w.Meta.MinHeight
	}
	if w.Meta.Background != 0 {
		opts.Background = w.Meta.Background
	}
}

// Attach mounts the window root into a Scene.
func (w *Window) Attach(scene *widgets.Scene) error {
	if w == nil {
		return errors.New("window is nil")
	}
	if scene == nil {
		return errors.New("scene is nil")
	}
	if current := w.Scene(); current != nil {
		if current == scene {
			return nil
		}
		return fmt.Errorf("window %q is already attached", w.ID)
	}

	if w.theme != nil {
		scene.SetTheme(w.theme)
	}

	root := scene.Root()
	if root == nil {
		return errors.New("scene root panel is nil")
	}

	w.mu.Lock()
	w.scene = scene
	w.mu.Unlock()

	if w.Root != nil {
		root.Add(w.Root)
		w.Root.SetBounds(root.Bounds())
	}
	if err := w.ReloadResources(ReloadReasonAttach); err != nil {
		_ = w.Detach()
		return err
	}
	w.RefreshBindings()
	return nil
}

// Detach unmounts the window root from its current scene.
func (w *Window) Detach() error {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	scene := w.scene
	w.scene = nil
	w.mu.Unlock()

	if scene == nil {
		return nil
	}
	root := scene.Root()
	if root == nil || w.Root == nil {
		return nil
	}
	root.Remove(w.Root.ID())
	return nil
}

// SetData connects the window to a host data source.
func (w *Window) SetData(data DataSource) {
	if w == nil {
		return
	}

	w.mu.Lock()
	if w.unsubscribe != nil {
		w.unsubscribe()
		w.unsubscribe = nil
	}
	w.data = data
	if data != nil {
		w.unsubscribe = data.Subscribe(func(change Change) {
			w.scheduleBindingRefresh(change.Paths)
		})
	}
	w.mu.Unlock()

	w.scheduleBindingRefresh(nil)
}

// Data returns the current host data source.
func (w *Window) Data() DataSource {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.data
}

// Scene returns the scene the window is currently attached to.
func (w *Window) Scene() *widgets.Scene {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.scene
}

// App returns the bound app when the window is attached.
func (w *Window) App() *core.App {
	scene := w.Scene()
	if scene == nil {
		return nil
	}
	return scene.App()
}

// FindWidget looks up a widget by ID within the window.
func (w *Window) FindWidget(id string) widgets.Widget {
	if w == nil || id == "" {
		return nil
	}
	if w.widgetIndex != nil {
		return w.widgetIndex[id]
	}
	return widgets.FindByID(w.Root, id)
}

func (w *Window) registerWidget(widget widgets.Widget) error {
	if w == nil || widget == nil {
		return nil
	}
	id := widget.ID()
	if id == "" {
		return nil
	}
	if w.widgetIndex == nil {
		w.widgetIndex = map[string]widgets.Widget{}
	}
	if existing := w.widgetIndex[id]; existing != nil {
		return fmt.Errorf("duplicate widget id %q in window %q", id, w.ID)
	}
	w.widgetIndex[id] = widget
	return nil
}

func (w *Window) addResourceReloader(reload func(resourceReloadContext) error) {
	if w == nil || reload == nil {
		return
	}
	w.reloaders = append(w.reloaders, resourceReloader{reload: reload})
}

// RefreshBindings reapplies all matching bindings against the current data source.
func (w *Window) RefreshBindings(paths ...string) {
	if w == nil {
		return
	}

	w.mu.Lock()
	bindings := append([]windowBinding(nil), w.bindings...)
	scene := w.scene
	data := w.data
	w.mu.Unlock()

	if len(bindings) == 0 {
		return
	}

	ctx := &bindingContext{
		window: w,
		scene:  scene,
		data:   data,
	}
	for _, binding := range bindings {
		if binding.apply == nil || !binding.matches(paths) {
			continue
		}
		binding.apply(ctx)
	}
}

func (w *Window) scheduleBindingRefresh(paths []string) {
	if w == nil {
		return
	}

	w.mu.Lock()
	scene := w.scene
	w.mu.Unlock()

	if scene != nil {
		if app := scene.App(); app != nil && !app.IsUIThread() {
			copied := append([]string(nil), paths...)
			_ = app.Post(func() {
				w.RefreshBindings(copied...)
			})
			return
		}
	}
	w.RefreshBindings(paths...)
}

func (w *Window) setWindowTitle(title string) {
	if w == nil {
		return
	}
	w.Meta.Title = title

	w.mu.Lock()
	scene := w.scene
	w.mu.Unlock()

	if scene != nil {
		if app := scene.App(); app != nil {
			app.SetTitle(title)
		}
	}
}

func (w *Window) sceneClientSize() core.Size {
	w.mu.Lock()
	scene := w.scene
	w.mu.Unlock()
	if scene != nil && scene.App() != nil {
		return scene.App().ClientSize()
	}
	return core.Size{Width: w.Meta.Width, Height: w.Meta.Height}
}

// Document is the built result of a JSON UI document.
type Document struct {
	Windows []*Window
	index   map[string]*Window
}

// PrimaryWindow returns the first declared window.
func (d *Document) PrimaryWindow() *Window {
	if d == nil || len(d.Windows) == 0 {
		return nil
	}
	return d.Windows[0]
}

// Window looks up a built window by id.
func (d *Document) Window(id string) *Window {
	if d == nil || id == "" {
		return nil
	}
	return d.index[id]
}

// FindWidget looks up a widget by window id and widget id.
func (d *Document) FindWidget(windowID string, widgetID string) widgets.Widget {
	window := d.Window(windowID)
	if window == nil {
		return nil
	}
	return window.FindWidget(widgetID)
}

// SetData reconnects every built window to the provided host data source.
func (d *Document) SetData(data DataSource) {
	if d == nil {
		return
	}
	for _, window := range d.Windows {
		window.SetData(data)
	}
}

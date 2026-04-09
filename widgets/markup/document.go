//go:build windows

package markup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// ActionContext describes runtime information exposed to markup action handlers.
type ActionContext struct {
	// Name is the action name declared in markup, such as onclick or onchange.
	Name string
	// Widget is the widget instance that triggered the action.
	Widget widgets.Widget
	// ID is the triggering widget ID.
	ID string
	// Value stores the current value exposed by the widget.
	Value string
	// Paths contains full file paths returned by file-input style widgets.
	Paths []string
	// Checked reports the current checkbox/radio state.
	Checked bool
	// Index reports the selected index for select/listbox widgets, or -1.
	Index int
	// Item reports the current selected item for select/listbox widgets.
	Item widgets.ListItem
}

// WindowMeta describes window-level metadata extracted from markup.
type WindowMeta struct {
	// Title is the window title.
	Title string
	// Icon is the resolved window icon.
	Icon *core.Icon
	// IconPath stores the original icon attribute value.
	IconPath string
	// MinWidth is the minimum client width.
	MinWidth int32
	// MinHeight is the minimum client height.
	MinHeight int32
}

// Document is the built result of the markup HTML/CSS DSL.
type Document struct {
	// Root is the root widget tree built from the document body.
	Root widgets.Widget
	// Meta contains parsed window-level metadata.
	Meta WindowMeta

	theme *widgets.Theme
	mu    sync.Mutex

	scene       *widgets.Scene
	state       *State
	bindings    []documentBinding
	unsubscribe func()
}

// ApplyWindowMeta copies document window metadata into core.Options.
func (d *Document) ApplyWindowMeta(opts *core.Options) {
	if d == nil || opts == nil {
		return
	}
	if d.Meta.Title != "" {
		opts.Title = d.Meta.Title
	}
	if d.Meta.Icon != nil {
		opts.Icon = d.Meta.Icon
	}
	if d.Meta.MinWidth > 0 {
		opts.MinWidth = d.Meta.MinWidth
	}
	if d.Meta.MinHeight > 0 {
		opts.MinHeight = d.Meta.MinHeight
	}
}

// Attach mounts the document root into a scene.
func (d *Document) Attach(scene *widgets.Scene) error {
	if scene == nil {
		return errors.New("scene is nil")
	}
	if d == nil {
		return errors.New("document is nil")
	}
	if d.theme != nil {
		scene.SetTheme(d.theme)
	}
	if d.Root == nil {
		d.mu.Lock()
		d.scene = scene
		d.mu.Unlock()
		d.RefreshBindings()
		return nil
	}
	root := scene.Root()
	if root == nil {
		return errors.New("scene root panel is nil")
	}

	d.mu.Lock()
	d.scene = scene
	d.mu.Unlock()

	root.Add(d.Root)
	d.Root.SetBounds(root.Bounds())
	d.RefreshBindings()
	return nil
}

// LoadDocumentFile loads a document from a .ui.html file and optionally reads
// a sibling .ui.css file with the same base name.
func LoadDocumentFile(path string, opts LoadOptions) (*Document, error) {
	htmlBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cssPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".css"
	cssBytes, err := os.ReadFile(cssPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if opts.AssetsDir == "" {
		opts.AssetsDir = filepath.Dir(path)
	}
	return LoadDocumentString(string(htmlBytes), string(cssBytes), opts)
}

// LoadDocumentString builds a document from HTML and CSS strings.
func LoadDocumentString(htmlText string, cssText string, opts LoadOptions) (*Document, error) {
	root, err := parseHTMLDocument(htmlText)
	if err != nil {
		return nil, err
	}
	rules, err := parseCSS(cssText)
	if err != nil {
		return nil, err
	}
	if err := applyCSS(root, rules); err != nil {
		return nil, err
	}
	builder := &uiBuilder{opts: opts}
	doc, err := builder.buildDocument(root)
	if err != nil {
		return nil, err
	}
	doc.theme = opts.Theme
	doc.SetState(opts.State)
	return doc, nil
}

// LoadIntoScene builds a document and attaches it to the provided scene.
func LoadIntoScene(scene *widgets.Scene, htmlText string, cssText string, opts LoadOptions) (*Document, error) {
	doc, err := LoadDocumentString(htmlText, cssText, opts)
	if err != nil {
		return nil, err
	}
	if err := doc.Attach(scene); err != nil {
		return nil, err
	}
	return doc, nil
}

// LoadFileIntoScene loads a document file and attaches it to the provided scene.
func LoadFileIntoScene(scene *widgets.Scene, path string, opts LoadOptions) (*Document, error) {
	doc, err := LoadDocumentFile(path, opts)
	if err != nil {
		return nil, err
	}
	if err := doc.Attach(scene); err != nil {
		return nil, err
	}
	return doc, nil
}

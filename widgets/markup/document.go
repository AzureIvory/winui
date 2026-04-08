//go:build windows

package markup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

// ActionContext 描述 HTML 动作触发时可用的上下文信息。
type ActionContext struct {
	// Name 是动作名，对应 HTML 中的 onclick、onchange、onsubmit、onactivate。
	Name string
	// Widget 是触发动作的控件实例。
	Widget widgets.Widget
	// ID 是触发控件的 ID。
	ID string
	// Value 是触发时可读取到的当前值。
	Value string
	// Paths 保存 file input 等控件返回的完整路径列表。
	Paths []string
	// Checked 是 checkbox/radio 的勾选状态。
	Checked bool
	// Index 是 select/listbox 的当前索引，未命中时为 -1。
	Index int
	// Item 是 select/listbox 的当前项，未命中时为零值。
	Item widgets.ListItem
}

// WindowMeta 描述窗口级元数据。
type WindowMeta struct {
	// Title 是窗口标题。
	Title string
	// Icon 是窗口图标对象，当前仅支持本地 .ico。
	Icon *core.Icon
	// IconPath 是 icon 属性的原始路径。
	IconPath string
	// MinWidth 是窗口客户区最小宽度。
	MinWidth int32
	// MinHeight 是窗口客户区最小高度。
	MinHeight int32
}

// Document 表示 HTML/CSS DSL 构建后的文档结果。
type Document struct {
	// Root 是文档构建得到的控件树根节点。
	Root widgets.Widget
	// Meta 是文档解析得到的窗口级元数据。
	Meta WindowMeta

	theme *widgets.Theme
}

// ApplyWindowMeta 把文档中的窗口元数据写入 core.Options。
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

// Attach 将文档挂载到 Scene。
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
		return nil
	}
	root := scene.Root()
	if root == nil {
		return errors.New("scene root panel is nil")
	}
	root.Add(d.Root)
	d.Root.SetBounds(root.Bounds())
	return nil
}

// LoadDocumentFile 从 .ui.html 文件加载文档，并自动尝试读取同名 .ui.css 文件。
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

// LoadDocumentString 从 HTML 文本和 CSS 文本构建文档。
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
	return doc, nil
}

// LoadIntoScene 直接把 HTML/CSS 构建结果挂载到 Scene。
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

// LoadFileIntoScene 直接把 .ui.html 文件构建结果挂载到 Scene。
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

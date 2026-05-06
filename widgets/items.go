//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ListItem 描述列表类控件中的一个选项。
type ListItem struct {
	// Value 保存选项值。
	Value string
	// Text 保存主显示文本；为空时回退到 Value。
	Text string
	// Subtitle 保存富样式下的副标题文本。
	Subtitle string
	// Image 保存富样式下的左侧图片资源。
	Image *core.Image
	// ImagePath 保存 JSON 字面量声明的图片路径。
	ImagePath string
	// Disabled 表示该项是否禁用。
	Disabled bool
}

// displayText 返回列表项最合适的显示文本。
func (i ListItem) displayText() string {
	if i.Text != "" {
		return i.Text
	}
	return i.Value
}

func (i ListItem) hasRichContent() bool {
	return i.Subtitle != "" || i.Image != nil || i.ImagePath != ""
}

// cloneItems 返回列表项切片的浅拷贝。
func cloneItems(items []ListItem) []ListItem {
	cloned := make([]ListItem, 0, len(items))
	for _, item := range items {
		if item.Text == "" {
			item.Text = item.Value
		}
		cloned = append(cloned, item)
	}
	return cloned
}

//go:build windows

package widgets

// ListItem 描述列表类控件中的一个选项。
type ListItem struct {
	// Value 保存选项值。
	Value string
	// Text 保存显示文本；为空时回退到 Value。
	Text string
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

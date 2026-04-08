//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

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

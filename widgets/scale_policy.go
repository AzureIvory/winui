//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ScaleMode describes how a metric should respond to DPI.
type ScaleMode uint8

const (
	// ScaleInherit uses the nearest explicit policy or the library default.
	ScaleInherit ScaleMode = iota
	// ScaleDP scales the value using the current scene DPI.
	ScaleDP
	// ScalePX keeps the value in design pixels.
	ScalePX
)

// ScalePolicy configures widget-level DPI behavior.
type ScalePolicy struct {
	Mode    ScaleMode
	Layout  ScaleMode
	Font    ScaleMode
	Image   ScaleMode
	Padding ScaleMode
	Gap     ScaleMode
	Radius  ScaleMode
}

type scaleSlot uint8

const (
	scaleSlotLayout scaleSlot = iota + 1
	scaleSlotFont
	scaleSlotImage
	scaleSlotPadding
	scaleSlotGap
	scaleSlotRadius
)

type scalePolicyHolder interface {
	scalePolicy() ScalePolicy
	setScalePolicy(ScalePolicy)
}

// SetScalePolicy updates a widget's DPI scaling policy.
func SetScalePolicy(widget Widget, policy ScalePolicy) {
	if widget == nil {
		return
	}
	holder, ok := widget.(scalePolicyHolder)
	if !ok {
		return
	}
	if holder.scalePolicy() == policy {
		return
	}
	holder.setScalePolicy(policy)
	scene := SceneOf(widget)
	if scene != nil {
		scene.resetFonts()
	}

	if panel, ok := widget.(*Panel); ok {
		panel.applyLayout()
	}
	if parent, ok := widgetParent(widget).(*Panel); ok {
		parent.applyLayout()
		parent.invalidate(parent)
		return
	}
	if scene != nil {
		scene.Invalidate(widget)
	}
}

// ScalePolicyOf returns the policy explicitly assigned to a widget.
func ScalePolicyOf(widget Widget) ScalePolicy {
	holder, ok := widget.(scalePolicyHolder)
	if !ok {
		return ScalePolicy{}
	}
	return holder.scalePolicy()
}

// LayoutScaleModeOf returns the effective layout scale mode for a widget.
func LayoutScaleModeOf(widget Widget) ScaleMode {
	return effectiveScaleModeForWidget(widget, scaleSlotLayout)
}

// ScaleLayoutValue resolves a layout metric using a widget's effective policy.
func ScaleLayoutValue(widget Widget, value int32) int32 {
	return scaleValueForWidget(widget, scaleSlotLayout, value)
}

func effectiveScaleModeForWidget(widget Widget, slot scaleSlot) ScaleMode {
	node := asWidgetNode(widget)
	if node == nil {
		return ScaleDP
	}
	if holder, ok := node.(scalePolicyHolder); ok {
		policy := holder.scalePolicy()
		if mode := scaleModeForSlot(policy, slot); mode != ScaleInherit {
			return mode
		}
		if policy.Mode != ScaleInherit {
			return policy.Mode
		}
	}
	for current := asWidgetNode(node.parent()); current != nil; {
		if holder, ok := current.(scalePolicyHolder); ok {
			policy := holder.scalePolicy()
			if slot != scaleSlotLayout {
				if mode := scaleModeForSlot(policy, slot); mode != ScaleInherit {
					return mode
				}
			}
			if policy.Mode != ScaleInherit {
				return policy.Mode
			}
		}
		parent := current.parent()
		if parent == nil {
			break
		}
		current = asWidgetNode(parent)
	}
	return ScaleDP
}

func scaleModeForSlot(policy ScalePolicy, slot scaleSlot) ScaleMode {
	switch slot {
	case scaleSlotLayout:
		return policy.Layout
	case scaleSlotFont:
		return policy.Font
	case scaleSlotImage:
		return policy.Image
	case scaleSlotPadding:
		return policy.Padding
	case scaleSlotGap:
		return policy.Gap
	case scaleSlotRadius:
		return policy.Radius
	default:
		return ScaleInherit
	}
}

func scaleValueForWidget(widget Widget, slot scaleSlot, value int32) int32 {
	if value == 0 {
		return 0
	}
	if effectiveScaleModeForWidget(widget, slot) == ScalePX {
		return value
	}
	scene := SceneOf(widget)
	if scene == nil || scene.app == nil {
		return value
	}
	return scene.app.DP(value)
}

func scaleInsetsForWidgetSlot(widget Widget, slot scaleSlot, value Insets) Insets {
	return Insets{
		Top:    scaleValueForWidget(widget, slot, value.Top),
		Right:  scaleValueForWidget(widget, slot, value.Right),
		Bottom: scaleValueForWidget(widget, slot, value.Bottom),
		Left:   scaleValueForWidget(widget, slot, value.Left),
	}
}

func scaleSizeForWidgetSlot(widget Widget, slot scaleSlot, value core.Size) core.Size {
	return core.Size{
		Width:  scaleValueForWidget(widget, slot, value.Width),
		Height: scaleValueForWidget(widget, slot, value.Height),
	}
}

func scaledFontHeightForWidget(widget Widget, spec FontSpec) int32 {
	size := spec.SizeDP
	if size <= 0 {
		size = 16
	}
	return scaleValueForWidget(widget, scaleSlotFont, size)
}

func widgetParent(widget Widget) Container {
	node := asWidgetNode(widget)
	if node == nil {
		return nil
	}
	return node.parent()
}

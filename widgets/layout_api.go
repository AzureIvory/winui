//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// PreferredSize returns a widget's current preferred size in logical DP units.
func PreferredSize(widget Widget) core.Size {
	return preferredSizeOf(widget)
}

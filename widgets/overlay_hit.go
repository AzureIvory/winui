//go:build windows

package widgets

type overlayHitWidget interface {
	Widget
	overlayHitTest(x, y int32) bool
}

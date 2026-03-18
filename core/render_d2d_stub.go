//go:build windows && !cgo

package core

import "errors"

import "golang.org/x/sys/windows"

// d2dRenderer 是禁用 cgo 时的占位实现，上层会自动回退到 GDI。
type d2dRenderer struct{}

// newD2DRenderer 在无 cgo 构建中始终返回不可用错误。
func newD2DRenderer() (*d2dRenderer, error) {
	return nil, errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) Close() {}

func (r *d2dRenderer) begin(windows.Handle, Rect) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) end() error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) flush() error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) fillRect(Rect, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) fillRoundRect(Rect, int32, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) strokeRoundRect(Rect, int32, Color, int32) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) fillPolygon([]Point, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) drawText(string, Rect, *Font, Color, uint32) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) measureText(string, *Font) (Size, error) {
	return Size{}, errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) drawIcon(*Icon, Rect) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) drawBitmap(*Bitmap, Rect, byte) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

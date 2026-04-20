//go:build windows && !cgo

package core

import (
	"errors"

	"golang.org/x/sys/windows"
)

// d2dRenderer 是禁用 cgo 时的占位实现，上层会自动回退到 GDI。
type d2dRenderer struct{}

// newD2DRenderer 在无 cgo 构建中始终返回不可用错误。
func newD2DRenderer() (*d2dRenderer, error) {
	return nil, errors.New("Direct2D backend unavailable: cgo is disabled")
}

// Close 释放占位渲染器；无 cgo 构建下无需实际操作。
func (r *d2dRenderer) Close() {}

// begin 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) begin(windows.Handle, Rect) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// end 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) end() error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// fillRect 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) fillRect(Rect, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// fillRoundRect 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) fillRoundRect(Rect, int32, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// strokeRoundRect 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) strokeRoundRect(Rect, int32, Color, int32) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// fillPolygon 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) fillPolygon([]Point, Color) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// drawText 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) drawText(string, Rect, *Font, Color, uint32) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// measureText 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) measureText(string, *Font) (Size, error) {
	return Size{}, errors.New("Direct2D backend unavailable: cgo is disabled")
}

// drawIcon 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) drawIcon(*Icon, Rect) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

// drawBitmap 在无 cgo 构建中始终返回不可用错误。
func (r *d2dRenderer) drawBitmap(*Bitmap, Rect, byte) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) pushClipRect(Rect) error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

func (r *d2dRenderer) popClipRect() error {
	return errors.New("Direct2D backend unavailable: cgo is disabled")
}

//go:build windows && cgo

package core

/*
#cgo CFLAGS: -DCOBJMACROS
#cgo LDFLAGS: -ld2d1 -ldwrite -lole32 -lwindowscodecs
#include <stdint.h>
#include <stddef.h>

typedef struct WinUID2DRenderer WinUID2DRenderer;

typedef struct WinUIDPoint {
	int32_t x;
	int32_t y;
} WinUIDPoint;

WinUID2DRenderer* winui_d2d_renderer_create(char* err, size_t err_len);
void winui_d2d_renderer_destroy(WinUID2DRenderer* renderer);
int winui_d2d_renderer_begin(WinUID2DRenderer* renderer, uintptr_t hdc, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len);
int winui_d2d_renderer_end(WinUID2DRenderer* renderer, char* err, size_t err_len);
int winui_d2d_fill_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, uint32_t color, char* err, size_t err_len);
int winui_d2d_fill_round_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, float radius, uint32_t color, char* err, size_t err_len);
int winui_d2d_stroke_round_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, float radius, uint32_t color, float width, char* err, size_t err_len);
int winui_d2d_fill_polygon(WinUID2DRenderer* renderer, const WinUIDPoint* points, int32_t count, uint32_t color, char* err, size_t err_len);
int winui_d2d_draw_text(WinUID2DRenderer* renderer, const uint16_t* text, const uint16_t* font_family, float font_size, int32_t font_weight, uint32_t color, uint32_t format, uint32_t options, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len);
int winui_d2d_measure_text(WinUID2DRenderer* renderer, const uint16_t* text, const uint16_t* font_family, float font_size, int32_t font_weight, int32_t* width, int32_t* height, char* err, size_t err_len);
int winui_d2d_draw_icon(WinUID2DRenderer* renderer, uintptr_t hicon, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len);
int winui_d2d_draw_bitmap(WinUID2DRenderer* renderer, uintptr_t hbitmap, int32_t left, int32_t top, int32_t right, int32_t bottom, uint8_t alpha, char* err, size_t err_len);
int winui_d2d_push_clip_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len);
int winui_d2d_pop_clip_rect(WinUID2DRenderer* renderer, char* err, size_t err_len);
int winui_d2d_flush(WinUID2DRenderer* renderer, char* err, size_t err_len);
*/
import "C"

import (
	"errors"
	"math"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	d2dDrawTextOptionsClip            uint32 = 0x00000001
	d2dDrawTextOptionsEnableColorFont uint32 = 0x00000004
)

func d2dTextDrawOptions() uint32 {
	return d2dDrawTextOptionsClip | d2dDrawTextOptionsEnableColorFont
}

// d2dRenderer 持有 Go 侧对 Direct2D/DirectWrite/WIC 桥接对象的引用。
// d2dRenderer 持有 Go 侧对 Direct2D/DirectWrite/WIC 桥接对象的引用。
type d2dRenderer struct {
	// ptr 保存底层 C 桥接渲染器指针。
	ptr *C.WinUID2DRenderer
}

// newD2DRenderer 创建可复用的 Direct2D 渲染器。
func newD2DRenderer() (*d2dRenderer, error) {
	var errBuf [512]C.char
	ptr := C.winui_d2d_renderer_create(&errBuf[0], C.size_t(len(errBuf)))
	if ptr == nil {
		if errBuf[0] != 0 {
			return nil, errors.New(C.GoString(&errBuf[0]))
		}
		return nil, errors.New("failed to initialize Direct2D renderer")
	}
	return &d2dRenderer{ptr: ptr}, nil
}

// Close 释放底层 Direct2D 渲染器资源。
func (r *d2dRenderer) Close() {
	if r == nil || r.ptr == nil {
		return
	}
	C.winui_d2d_renderer_destroy(r.ptr)
	r.ptr = nil
}

// begin 绑定本次 WM_PAINT 使用的 HDC 和绘制范围。
func (r *d2dRenderer) begin(hdc windows.Handle, bounds Rect) error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_renderer_begin(
			r.ptr,
			C.uintptr_t(uintptr(hdc)),
			C.int32_t(bounds.X),
			C.int32_t(bounds.Y),
			C.int32_t(bounds.X+bounds.W),
			C.int32_t(bounds.Y+bounds.H),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// end 提交当前绘制批次并结束目标绑定。
func (r *d2dRenderer) end() error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_renderer_end(r.ptr, &errBuf[0], C.size_t(len(errBuf)))
	})
}

// flush 强制把尚未提交的绘制命令刷新到底层目标。
func (r *d2dRenderer) flush() error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_flush(r.ptr, &errBuf[0], C.size_t(len(errBuf)))
	})
}

// fillRect 使用 Direct2D 填充矩形。
func (r *d2dRenderer) fillRect(rect Rect, color Color) error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_fill_rect(
			r.ptr,
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			C.uint32_t(color),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// fillRoundRect 使用 Direct2D 填充圆角矩形。
func (r *d2dRenderer) fillRoundRect(rect Rect, radius int32, color Color) error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_fill_round_rect(
			r.ptr,
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			C.float(float32(radius)),
			C.uint32_t(color),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// strokeRoundRect 使用 Direct2D 绘制圆角矩形边框。
func (r *d2dRenderer) strokeRoundRect(rect Rect, radius int32, color Color, width int32) error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_stroke_round_rect(
			r.ptr,
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			C.float(float32(radius)),
			C.uint32_t(color),
			C.float(float32(width)),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// fillPolygon 使用 Direct2D 填充多边形。
func (r *d2dRenderer) fillPolygon(points []Point, color Color) error {
	if len(points) < 3 {
		return nil
	}
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_fill_polygon(
			r.ptr,
			(*C.WinUIDPoint)(unsafe.Pointer(&points[0])),
			C.int32_t(len(points)),
			C.uint32_t(color),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// drawText 使用 Direct2D 绘制文本。
func (r *d2dRenderer) drawText(text string, rect Rect, font *Font, color Color, format uint32) error {
	if text == "" {
		return nil
	}

	fontFamily, fontSize, fontWeight := resolveFontMetrics(font)
	textUTF16, err := utf16z(text)
	if err != nil {
		return err
	}
	fontUTF16, err := utf16z(fontFamily)
	if err != nil {
		return err
	}

	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_draw_text(
			r.ptr,
			(*C.uint16_t)(unsafe.Pointer(&textUTF16[0])),
			(*C.uint16_t)(unsafe.Pointer(&fontUTF16[0])),
			C.float(fontSize),
			C.int32_t(fontWeight),
			C.uint32_t(color),
			C.uint32_t(format),
			C.uint32_t(d2dTextDrawOptions()),
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// measureText 使用 Direct2D 测量文本尺寸。
func (r *d2dRenderer) measureText(text string, font *Font) (Size, error) {
	if text == "" {
		return Size{}, nil
	}

	fontFamily, fontSize, fontWeight := resolveFontMetrics(font)
	textUTF16, err := utf16z(text)
	if err != nil {
		return Size{}, err
	}
	fontUTF16, err := utf16z(fontFamily)
	if err != nil {
		return Size{}, err
	}

	var width C.int32_t
	var height C.int32_t
	err = r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_measure_text(
			r.ptr,
			(*C.uint16_t)(unsafe.Pointer(&textUTF16[0])),
			(*C.uint16_t)(unsafe.Pointer(&fontUTF16[0])),
			C.float(fontSize),
			C.int32_t(fontWeight),
			&width,
			&height,
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
	if err != nil {
		return Size{}, err
	}
	return Size{Width: int32(width), Height: int32(height)}, nil
}

// drawIcon 使用 Direct2D 绘制图标。
func (r *d2dRenderer) drawIcon(icon *Icon, rect Rect) error {
	if icon == nil || icon.handle == 0 {
		return nil
	}
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_draw_icon(
			r.ptr,
			C.uintptr_t(icon.handle),
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// drawBitmap 使用 Direct2D 绘制位图。
func (r *d2dRenderer) drawBitmap(bitmap *Bitmap, rect Rect, alpha byte) error {
	if bitmap == nil || bitmap.handle == 0 {
		return nil
	}
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_draw_bitmap(
			r.ptr,
			C.uintptr_t(bitmap.handle),
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			C.uint8_t(alpha),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

// call 统一处理 C 桥接调用和错误转换。
func (r *d2dRenderer) call(fn func(errBuf *[512]C.char) C.int) error {
	if r == nil || r.ptr == nil {
		return errors.New("Direct2D renderer is not initialized")
	}
	var errBuf [512]C.char
	if fn(&errBuf) != 0 {
		return nil
	}
	if errBuf[0] != 0 {
		return errors.New(C.GoString(&errBuf[0]))
	}
	return errors.New("Direct2D call failed")
}

// resolveFontMetrics 把 winui 的字体描述转换成 DirectWrite 需要的参数。
func resolveFontMetrics(font *Font) (string, float32, int32) {
	face := "Microsoft YaHei UI"
	if font != nil && font.face != "" {
		face = font.face
	}

	size := float32(16)
	if font != nil && font.height > 0 {
		size = float32(font.height)
	}

	weight := int32(400)
	if font != nil && font.weight > 0 {
		weight = font.weight
	}

	return face, float32(math.Max(1, float64(size))), weight
}

// utf16z 为 Win32 和 DirectWrite 调用准备以 NUL 结尾的 UTF-16 字符串。
func utf16z(value string) ([]uint16, error) {
	out, err := windows.UTF16FromString(value)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *d2dRenderer) pushClipRect(rect Rect) error {
	if rect.Empty() {
		return nil
	}
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_push_clip_rect(
			r.ptr,
			C.int32_t(rect.X),
			C.int32_t(rect.Y),
			C.int32_t(rect.X+rect.W),
			C.int32_t(rect.Y+rect.H),
			&errBuf[0],
			C.size_t(len(errBuf)),
		)
	})
}

func (r *d2dRenderer) popClipRect() error {
	return r.call(func(errBuf *[512]C.char) C.int {
		return C.winui_d2d_pop_clip_rect(r.ptr, &errBuf[0], C.size_t(len(errBuf)))
	})
}

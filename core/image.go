//go:build windows

package core

import (
	"bytes"
	"image"
	stddraw "image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"
	"unsafe"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/sys/windows"
)

var (
	procCreateBitmap       = gdi32.NewProc("CreateBitmap")
	procCreateIconIndirect = user32.NewProc("CreateIconIndirect")
	// bitmapSourceRegistry 让控件可以从原生位图反查到源图资源，
	// 这样按钮和动画帧在缩小时都能复用统一的高质量缩放与缓存。
	bitmapSourceRegistry sync.Map
)

// ImageScaleQuality 表示缓存位图时使用的缩放质量。
type ImageScaleQuality uint8

const (
	// ImageScaleLinear 适合普通 UI 图像。
	ImageScaleLinear ImageScaleQuality = iota + 1
	// ImageScaleHigh 适合缩小显示的位图。
	ImageScaleHigh
)

// ImageCacheKey 是 core.Image 内部位图缓存键。
type ImageCacheKey struct {
	Width   int32
	Height  int32
	Quality ImageScaleQuality
}

// Image 是通用图片资源。
// - source 保存原始 RGBA 图像，用于高质量缩放。
// - master 保存原始尺寸对应的 GDI 位图。
// - variants 按目标像素尺寸缓存缩放后的位图。
// - icons 按方形像素边长缓存窗口图标。
type Image struct {
	mu sync.Mutex

	source     *image.RGBA
	sourceSize Size
	master     *Bitmap

	variants map[ImageCacheKey]*Bitmap
	icons    map[int32]*Icon
}

func registerBitmapSource(bitmap *Bitmap, img *Image) {
	if bitmap == nil || bitmap.handle == 0 || img == nil {
		return
	}
	bitmapSourceRegistry.Store(bitmap.handle, img)
}

func unregisterBitmapSource(bitmap *Bitmap) {
	if bitmap == nil || bitmap.handle == 0 {
		return
	}
	bitmapSourceRegistry.Delete(bitmap.handle)
}

// ImageForBitmap 返回原生位图对应的源图资源。
// 当位图来自 core.Image 或动画帧解码结果时，调用方可以借此复用统一的缩放缓存。
func ImageForBitmap(bitmap *Bitmap) *Image {
	if bitmap == nil || bitmap.handle == 0 {
		return nil
	}
	value, ok := bitmapSourceRegistry.Load(bitmap.handle)
	if !ok {
		return nil
	}
	img, _ := value.(*Image)
	return img
}

// ChooseScaleQuality 根据源尺寸和目标尺寸决定缩放质量。
// 统一规则是：只要目标尺寸小于源尺寸，就默认使用高质量缩放。
func ChooseScaleQuality(src, dst Size) ImageScaleQuality {
	if src.Width <= 0 || src.Height <= 0 || dst.Width <= 0 || dst.Height <= 0 {
		return ImageScaleLinear
	}
	if dst.Width < src.Width || dst.Height < src.Height {
		return ImageScaleHigh
	}
	return ImageScaleLinear
}

type iconInfo struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  windows.Handle
	HbmColor windows.Handle
}

// LoadImageBytes 从编码图像数据创建 core.Image。
func LoadImageBytes(data []byte) (*Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return NewImageFromDecoded(img)
}

// LoadImageFile 从本地文件创建 core.Image。
func LoadImageFile(path string) (*Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadImageBytes(data)
}

// NewImageFromDecoded 从 Go image.Image 创建 core.Image。
func NewImageFromDecoded(src image.Image) (*Image, error) {
	rgba := imageToRGBA(src)
	if rgba == nil || rgba.Bounds().Dx() <= 0 || rgba.Bounds().Dy() <= 0 {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_DATA)
	}
	return newImageFromRGBA(rgba)
}

// newImageFromRGBA 从规范化后的 RGBA 数据创建图像资源。
func newImageFromRGBA(rgba *image.RGBA) (*Image, error) {
	if rgba == nil || rgba.Bounds().Dx() <= 0 || rgba.Bounds().Dy() <= 0 {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_DATA)
	}
	master, err := BitmapFromRGBA(rgba)
	if err != nil {
		return nil, err
	}
	img := &Image{
		source:     rgba,
		sourceSize: Size{Width: int32(rgba.Bounds().Dx()), Height: int32(rgba.Bounds().Dy())},
		master:     master,
		variants:   make(map[ImageCacheKey]*Bitmap),
		icons:      make(map[int32]*Icon),
	}
	registerBitmapSource(master, img)
	return img, nil
}

// NaturalSize 返回原始像素尺寸。
func (i *Image) NaturalSize() Size {
	if i == nil {
		return Size{}
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.sourceSize
}

// MasterBitmap 返回原始尺寸位图。
func (i *Image) MasterBitmap() *Bitmap {
	if i == nil {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.master
}

// BitmapFor 返回指定目标像素尺寸下的缓存位图。
// 这里的 width / height 是最终要绘制的目标像素尺寸。
func (i *Image) BitmapFor(width, height int32, quality ImageScaleQuality) (*Bitmap, error) {
	if i == nil {
		return nil, nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.bitmapForLocked(width, height, quality)
}

// IconFor 返回指定方形像素边长的 HICON。
func (i *Image) IconFor(size int32) (*Icon, error) {
	if i == nil {
		return nil, nil
	}
	if size <= 0 {
		return nil, wrapError("CreateIconIndirect", windows.ERROR_INVALID_PARAMETER)
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if icon := i.icons[size]; icon != nil {
		return icon, nil
	}
	if i.source == nil || i.sourceSize.Width <= 0 || i.sourceSize.Height <= 0 {
		return nil, wrapError("CreateIconIndirect", windows.ERROR_INVALID_DATA)
	}

	fitted := FitContain(i.sourceSize, size, size)
	if fitted.Width <= 0 || fitted.Height <= 0 {
		return nil, wrapError("CreateIconIndirect", windows.ERROR_INVALID_DATA)
	}

	scaled, err := i.scaledRGBALocked(fitted.Width, fitted.Height, ImageScaleHigh)
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, int(size), int(size)))
	dst := image.Rect(
		int((size-fitted.Width)/2),
		int((size-fitted.Height)/2),
		int((size-fitted.Width)/2+fitted.Width),
		int((size-fitted.Height)/2+fitted.Height),
	)
	stddraw.Draw(canvas, dst, scaled, image.Point{}, stddraw.Src)

	colorBitmap, err := BitmapFromRGBA(canvas)
	if err != nil {
		return nil, err
	}
	defer colorBitmap.Close()

	icon, err := createIconFromBitmap(colorBitmap)
	if err != nil {
		return nil, err
	}
	i.icons[size] = icon
	return icon, nil
}

// Close 释放图片资源及其缓存。
func (i *Image) Close() error {
	if i == nil {
		return nil
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	for key, bmp := range i.variants {
		if bmp != nil {
			_ = bmp.Close()
		}
		delete(i.variants, key)
	}
	for key, icon := range i.icons {
		if icon != nil {
			_ = icon.Close()
		}
		delete(i.icons, key)
	}
	if i.master != nil {
		_ = i.master.Close()
		i.master = nil
	}
	i.source = nil
	i.sourceSize = Size{}
	return nil
}

// FitContain 计算 src 放入 maxW x maxH 后的等比尺寸。
func FitContain(src Size, maxW, maxH int32) Size {
	if src.Width <= 0 || src.Height <= 0 || maxW <= 0 || maxH <= 0 {
		return Size{}
	}
	scaleX := float64(maxW) / float64(src.Width)
	scaleY := float64(maxH) / float64(src.Height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	if scale <= 0 {
		scale = 1
	}
	w := int32(float64(src.Width) * scale)
	h := int32(float64(src.Height) * scale)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{Width: w, Height: h}
}

func (i *Image) bitmapForLocked(width, height int32, quality ImageScaleQuality) (*Bitmap, error) {
	if width <= 0 || height <= 0 {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_PARAMETER)
	}
	if i.source == nil || i.master == nil {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_DATA)
	}

	quality = normalizeImageScaleQuality(quality)

	if width == i.sourceSize.Width && height == i.sourceSize.Height {
		return i.master, nil
	}

	key := ImageCacheKey{
		Width:   width,
		Height:  height,
		Quality: quality,
	}
	if cached := i.variants[key]; cached != nil {
		return cached, nil
	}

	scaled, err := i.scaledRGBALocked(width, height, quality)
	if err != nil {
		return nil, err
	}
	bmp, err := BitmapFromRGBA(scaled)
	if err != nil {
		return nil, err
	}
	i.variants[key] = bmp
	registerBitmapSource(bmp, i)
	return bmp, nil
}

func (i *Image) scaledRGBALocked(width, height int32, quality ImageScaleQuality) (*image.RGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_PARAMETER)
	}
	if i.source == nil {
		return nil, wrapError("BitmapFromRGBA", windows.ERROR_INVALID_DATA)
	}
	if width == i.sourceSize.Width && height == i.sourceSize.Height {
		clone := image.NewRGBA(i.source.Bounds())
		copy(clone.Pix, i.source.Pix)
		return clone, nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	switch normalizeImageScaleQuality(quality) {
	case ImageScaleHigh:
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), i.source, i.source.Bounds(), stddraw.Src, nil)
	default:
		xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), i.source, i.source.Bounds(), stddraw.Src, nil)
	}
	return dst, nil
}

func normalizeImageScaleQuality(q ImageScaleQuality) ImageScaleQuality {
	switch q {
	case ImageScaleHigh:
		return ImageScaleHigh
	default:
		return ImageScaleLinear
	}
}

func imageToRGBA(src image.Image) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil
	}
	if rgba, ok := src.(*image.RGBA); ok && rgba.Rect.Min.X == 0 && rgba.Rect.Min.Y == 0 {
		clone := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		copy(clone.Pix, rgba.Pix)
		return clone
	}
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	stddraw.Draw(dst, dst.Bounds(), src, bounds.Min, stddraw.Src)
	return dst
}

func createIconFromBitmap(color *Bitmap) (*Icon, error) {
	if color == nil || color.handle == 0 || color.Width <= 0 || color.Height <= 0 {
		return nil, wrapError("CreateIconIndirect", windows.ERROR_INVALID_DATA)
	}

	mask, _, maskErr := procCreateBitmap.Call(
		uintptr(color.Width),
		uintptr(color.Height),
		1,
		1,
		0,
	)
	if mask == 0 {
		return nil, wrapError("CreateBitmap(mask)", maskErr)
	}
	defer procDeleteObject.Call(mask)

	info := iconInfo{
		FIcon:    1,
		HbmMask:  windows.Handle(mask),
		HbmColor: color.handle,
	}
	h, _, err := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&info)))
	if h == 0 {
		return nil, wrapError("CreateIconIndirect", err)
	}
	return &Icon{handle: windows.Handle(h)}, nil
}

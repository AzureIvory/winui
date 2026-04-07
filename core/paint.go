//go:build windows

package core

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procBeginPaint            = user32.NewProc("BeginPaint")
	procEndPaint              = user32.NewProc("EndPaint")
	procFillRect              = user32.NewProc("FillRect")
	procDrawTextW             = user32.NewProc("DrawTextW")
	procDrawIconEx            = user32.NewProc("DrawIconEx")
	procCreateIconFromResEx   = user32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon           = user32.NewProc("DestroyIcon")
	procCreateSolidBrush      = gdi32.NewProc("CreateSolidBrush")
	procCreatePen             = gdi32.NewProc("CreatePen")
	procCreateFontIndirectW   = gdi32.NewProc("CreateFontIndirectW")
	procCreateCompatibleDC    = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC              = gdi32.NewProc("DeleteDC")
	procDeleteObject          = gdi32.NewProc("DeleteObject")
	procSelectObject          = gdi32.NewProc("SelectObject")
	procRoundRect             = gdi32.NewProc("RoundRect")
	procPolygon               = gdi32.NewProc("Polygon")
	procGetTextExtentPoint32W = gdi32.NewProc("GetTextExtentPoint32W")
	procSetBkMode             = gdi32.NewProc("SetBkMode")
	procSetTextColor          = gdi32.NewProc("SetTextColor")
	procCreateDIBSection      = gdi32.NewProc("CreateDIBSection")
	procBitBlt                = gdi32.NewProc("BitBlt")
	procGetStockObject        = gdi32.NewProc("GetStockObject")
	procAlphaBlend            = msimg32.NewProc("AlphaBlend")
	procSaveDC                = gdi32.NewProc("SaveDC")
	procRestoreDC             = gdi32.NewProc("RestoreDC")
	procIntersectClipRect     = gdi32.NewProc("IntersectClipRect")
)

const (
	// nullBrush 表示空画刷对象。
	nullBrush = 5
	// nullPen 表示空画笔对象。
	nullPen = 8
)

// paintStruct 对应 Win32 的 PAINTSTRUCT 结构。
type paintStruct struct {
	// Hdc 表示 BeginPaint 返回的绘图设备上下文。
	Hdc windows.Handle
	// FErase 表示是否需要擦除背景。
	FErase int32
	// RcPaint 表示本次需要重绘的区域。
	RcPaint winRect
	// FRestore 表示系统内部恢复标志。
	FRestore int32
	// FIncUpdate 表示系统内部增量更新标志。
	FIncUpdate int32
	// RgbReserved 保存保留字节。
	RgbReserved [32]byte
}

// logFontW 对应 Win32 的 LOGFONTW 结构。
type logFontW struct {
	// Height 指定字体高度。
	Height int32
	// Width 指定字体宽度。
	Width int32
	// Escapement 指定字形基线角度。
	Escapement int32
	// Orientation 指定字符方向角度。
	Orientation int32
	// Weight 指定字体粗细。
	Weight int32
	// Italic 表示是否斜体。
	Italic byte
	// Underline 表示是否带下划线。
	Underline byte
	// StrikeOut 表示是否带删除线。
	StrikeOut byte
	// CharSet 指定字符集。
	CharSet byte
	// OutPrecision 指定输出精度。
	OutPrecision byte
	// ClipPrecision 指定裁剪精度。
	ClipPrecision byte
	// Quality 指定字体渲染质量。
	Quality byte
	// PitchAndFamily 指定字距和字族。
	PitchAndFamily byte
	// FaceName 保存字体名称。
	FaceName [32]uint16
}

// bitmapInfoHeader 对应 Win32 的 BITMAPINFOHEADER 结构。
type bitmapInfoHeader struct {
	// Size 表示结构体大小。
	Size uint32
	// Width 表示位图宽度。
	Width int32
	// Height 表示位图高度。
	Height int32
	// Planes 表示颜色平面数。
	Planes uint16
	// BitCount 表示每像素位数。
	BitCount uint16
	// Compression 表示压缩方式。
	Compression uint32
	// SizeImage 表示像素数据大小。
	SizeImage uint32
	// XPelsPerMeter 表示水平分辨率。
	XPelsPerMeter int32
	// YPelsPerMeter 表示垂直分辨率。
	YPelsPerMeter int32
	// ClrUsed 表示使用的调色板颜色数。
	ClrUsed uint32
	// ClrImportant 表示重要颜色数。
	ClrImportant uint32
}

// bitmapInfo 对应 Win32 的 BITMAPINFO 结构。
type bitmapInfo struct {
	// Header 保存位图头信息。
	Header bitmapInfoHeader
	// Colors 保存颜色表。
	Colors [1]uint32
}

// blendFunction 对应 Win32 的 BLENDFUNCTION 结构。
type blendFunction struct {
	// BlendOp 指定混合操作。
	BlendOp byte
	// BlendFlags 保存保留标志。
	BlendFlags byte
	// SourceConstantAlpha 指定整体透明度。
	SourceConstantAlpha byte
	// AlphaFormat 指定是否使用源 Alpha。
	AlphaFormat byte
}

// Font 表示可复用的原生字体资源。
type Font struct {
	// handle 保存字体句柄。
	handle windows.Handle
	// face 保存字体名称。
	face string
	// height 保存字体高度。
	height int32
	// weight 保存字体粗细。
	weight int32
	// quality 保存字体渲染质量。
	quality FontQuality
}

// Brush 表示可复用的原生画刷资源。
type Brush struct {
	// handle 保存画刷句柄。
	handle windows.Handle
}

// Pen 表示可复用的原生画笔资源。
type Pen struct {
	// handle 保存画笔句柄。
	handle windows.Handle
}

// Icon 表示可复用的原生图标资源。
type Icon struct {
	// handle 保存图标句柄。
	handle windows.Handle
}

// Bitmap 表示可绘制的原生位图资源。
type Bitmap struct {
	// handle 保存位图句柄。
	handle windows.Handle
	// Width 保存位图宽度。
	Width int32
	// Height 保存位图高度。
	Height int32
}

// AnimatedFrame 表示动画图像中的单帧数据。
type AnimatedFrame struct {
	// Bitmap 保存帧位图。
	Bitmap *Bitmap
	// Width 保存帧宽度。
	Width int32
	// Height 保存帧高度。
	Height int32
	// Delay 保存下一帧播放前的延迟。
	Delay time.Duration
}

// Canvas 表示一次绘制过程使用的画布。
type Canvas struct {
	// hdc 保存底层绘图上下文句柄。
	hdc windows.Handle
	// bounds 保存当前画布边界。
	bounds Rect
	// d2d 保存可选的 Direct2D 渲染器。
	d2d *d2dRenderer
}

// PaintCtx 是 Canvas 的语义别名。
type PaintCtx = Canvas

// paintSession 保存一次 WM_PAINT 周期中的临时资源。
type paintSession struct {
	// app 指向所属应用实例。
	app *App
	// hwnd 保存当前绘制窗口句柄。
	hwnd windows.Handle
	// paintDC 保存 BeginPaint 返回的 DC。
	paintDC windows.Handle
	// memDC 保存双缓冲内存 DC。
	memDC windows.Handle
	// oldBitmap 保存原先选入 memDC 的位图。
	oldBitmap uintptr
	// buffer 保存双缓冲位图。
	buffer *Bitmap
	// canvas 保存对外暴露的画布包装。
	canvas *Canvas
	// ps 保存 BeginPaint 返回的 PAINTSTRUCT。
	ps paintStruct
	// direct 表示是否直接绘制到 paintDC。
	direct bool
	// d2d 保存本次绘制使用的 Direct2D 渲染器。
	d2d *d2dRenderer
}

// NewSolidBrush 创建一个新的实心画刷。
func NewSolidBrush(color Color) (*Brush, error) {
	h, _, err := procCreateSolidBrush.Call(uintptr(color))
	if h == 0 {
		return nil, wrapError("CreateSolidBrush", err)
	}
	return &Brush{handle: windows.Handle(h)}, nil
}

// Handle 返回画刷底层的原生句柄。
func (b *Brush) Handle() windows.Handle {
	if b == nil {
		return 0
	}
	return b.handle
}

// Close 释放画刷持有的资源。
func (b *Brush) Close() error {
	if b == nil || b.handle == 0 {
		return nil
	}
	h := b.handle
	b.handle = 0
	r1, _, err := procDeleteObject.Call(uintptr(h))
	if r1 == 0 {
		return wrapError("DeleteObject(brush)", err)
	}
	return nil
}

// NewPen 创建一个新的画笔。
func NewPen(width int32, color Color) (*Pen, error) {
	h, _, err := procCreatePen.Call(psSolid, uintptr(width), uintptr(color))
	if h == 0 {
		return nil, wrapError("CreatePen", err)
	}
	return &Pen{handle: windows.Handle(h)}, nil
}

// Handle 返回画笔底层的原生句柄。
func (p *Pen) Handle() windows.Handle {
	if p == nil {
		return 0
	}
	return p.handle
}

// Close 释放画笔持有的资源。
func (p *Pen) Close() error {
	if p == nil || p.handle == 0 {
		return nil
	}
	h := p.handle
	p.handle = 0
	r1, _, err := procDeleteObject.Call(uintptr(h))
	if r1 == 0 {
		return wrapError("DeleteObject(pen)", err)
	}
	return nil
}

// NewFont 创建一个新的字体。
func NewFont(face string, height int32, weight int32, quality FontQuality) (*Font, error) {
	var lf logFontW
	if weight <= 0 {
		weight = 400
	}
	lf.Height = -height
	lf.Weight = weight
	lf.CharSet = 1
	lf.Quality = byte(quality)
	copy(lf.FaceName[:], windows.StringToUTF16(face))

	h, _, err := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&lf)))
	if h == 0 {
		return nil, wrapError("CreateFontIndirectW", err)
	}
	return &Font{
		handle:  windows.Handle(h),
		face:    face,
		height:  height,
		weight:  weight,
		quality: quality,
	}, nil
}

// NewFont 创建一个新的字体。
func (a *App) NewFont(face string, heightDP int32, weight int32) (*Font, error) {
	return NewFont(face, a.DP(heightDP), weight, FontQualityClearType)
}

// Handle 返回字体底层的原生句柄。
func (f *Font) Handle() windows.Handle {
	if f == nil {
		return 0
	}
	return f.handle
}

// Close 释放字体持有的资源。
func (f *Font) Close() error {
	if f == nil || f.handle == 0 {
		return nil
	}
	h := f.handle
	f.handle = 0
	r1, _, err := procDeleteObject.Call(uintptr(h))
	if r1 == 0 {
		return wrapError("DeleteObject(font)", err)
	}
	return nil
}

// LoadIconFromICO 从 ICO 数据加载图标。
func LoadIconFromICO(data []byte, want int32) (*Icon, error) {
	if len(data) < 6 {
		return nil, wrapError("CreateIconFromResourceEx", windows.ERROR_INVALID_DATA)
	}

	count := int(*(*uint16)(unsafe.Pointer(&data[4])))
	if count <= 0 {
		return nil, wrapError("CreateIconFromResourceEx", windows.ERROR_INVALID_DATA)
	}

	// iconDirEntry 对应 ICO 文件中的目录项。
	type iconDirEntry struct {
		// W 和 H 表示图标宽高。
		W, H byte
		// ColorCount 表示调色板颜色数。
		ColorCount byte
		// Reserved 表示保留字段。
		Reserved byte
		// Planes 表示颜色平面数。
		Planes uint16
		// BitCount 表示每像素位数。
		BitCount uint16
		// BytesInRes 表示图像资源大小。
		BytesInRes uint32
		// ImageOffset 表示图像资源偏移。
		ImageOffset uint32
	}

	best := -1
	bestScore := int32(1 << 30)
	for i := 0; i < count; i++ {
		offset := 6 + i*16
		if offset+16 > len(data) {
			break
		}
		entry := (*iconDirEntry)(unsafe.Pointer(&data[offset]))
		w := int32(entry.W)
		h := int32(entry.H)
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		score := abs32(w-want) + abs32(h-want)
		if score < bestScore {
			bestScore = score
			best = i
		}
	}
	if best < 0 {
		return nil, wrapError("CreateIconFromResourceEx", windows.ERROR_INVALID_DATA)
	}

	entry := (*iconDirEntry)(unsafe.Pointer(&data[6+best*16]))
	start := int(entry.ImageOffset)
	end := start + int(entry.BytesInRes)
	if start < 0 || end > len(data) || start >= end {
		return nil, wrapError("CreateIconFromResourceEx", windows.ERROR_INVALID_DATA)
	}

	raw := data[start:end]
	h, _, err := procCreateIconFromResEx.Call(
		uintptr(unsafe.Pointer(&raw[0])),
		uintptr(len(raw)),
		1,
		0x00030000,
		uintptr(want),
		uintptr(want),
		0,
	)
	if h == 0 {
		return nil, wrapError("CreateIconFromResourceEx", err)
	}
	return &Icon{handle: windows.Handle(h)}, nil
}

// Handle 返回图标底层的原生句柄。
func (i *Icon) Handle() windows.Handle {
	if i == nil {
		return 0
	}
	return i.handle
}

// Close 释放图标持有的资源。
func (i *Icon) Close() error {
	if i == nil || i.handle == 0 {
		return nil
	}
	h := i.handle
	i.handle = 0
	r1, _, err := procDestroyIcon.Call(uintptr(h))
	if r1 == 0 {
		return wrapError("DestroyIcon", err)
	}
	return nil
}

// BitmapFromRGBA 根据 RGBA 图像创建位图。
func BitmapFromRGBA(img *image.RGBA) (*Bitmap, error) {
	w := int32(img.Bounds().Dx())
	h := int32(img.Bounds().Dy())
	if w <= 0 || h <= 0 {
		return nil, wrapError("CreateDIBSection", windows.ERROR_INVALID_DATA)
	}

	var bi bitmapInfo
	bi.Header.Size = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.Width = w
	bi.Header.Height = -h
	bi.Header.Planes = 1
	bi.Header.BitCount = 32

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return nil, wrapError("GetDC", windows.ERROR_INVALID_WINDOW_HANDLE)
	}
	defer procReleaseDC.Call(0, hdc)

	var bits unsafe.Pointer
	hbmp, _, err := procCreateDIBSection.Call(
		hdc,
		uintptr(unsafe.Pointer(&bi)),
		0,
		uintptr(unsafe.Pointer(&bits)),
		0,
		0,
	)
	if hbmp == 0 {
		return nil, wrapError("CreateDIBSection", err)
	}

	dst := unsafe.Slice((*byte)(bits), int(w*h*4))
	for i := 0; i < len(img.Pix); i += 4 {
		r := img.Pix[i]
		g := img.Pix[i+1]
		b := img.Pix[i+2]
		a := img.Pix[i+3]

		pr := byte(uint16(r) * uint16(a) / 255)
		pg := byte(uint16(g) * uint16(a) / 255)
		pb := byte(uint16(b) * uint16(a) / 255)

		dst[i] = pb
		dst[i+1] = pg
		dst[i+2] = pr
		dst[i+3] = a
	}

	return &Bitmap{
		handle: windows.Handle(hbmp),
		Width:  w,
		Height: h,
	}, nil
}

// Handle 返回位图底层的原生句柄。
func (b *Bitmap) Handle() windows.Handle {
	if b == nil {
		return 0
	}
	return b.handle
}

// Close 释放位图持有的资源。
func (b *Bitmap) Close() error {
	if b == nil || b.handle == 0 {
		return nil
	}
	h := b.handle
	b.handle = 0
	r1, _, err := procDeleteObject.Call(uintptr(h))
	if r1 == 0 {
		return wrapError("DeleteObject(bitmap)", err)
	}
	return nil
}

// DecodeGIF 将 GIF 数据解码为便于 Go 使用的结构。
func DecodeGIF(data []byte) ([]AnimatedFrame, error) {
	all, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, all.Config.Width, all.Config.Height))
	frames := make([]AnimatedFrame, 0, len(all.Image))
	for idx, frame := range all.Image {
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		rgba := image.NewRGBA(canvas.Bounds())
		copy(rgba.Pix, canvas.Pix)

		bmp, bmpErr := BitmapFromRGBA(rgba)
		if bmpErr != nil {
			for i := range frames {
				_ = frames[i].Bitmap.Close()
			}
			return nil, bmpErr
		}

		delay := 100 * time.Millisecond
		if idx < len(all.Delay) && all.Delay[idx] > 0 {
			delay = time.Duration(all.Delay[idx]) * 10 * time.Millisecond
		}
		if delay < 10*time.Millisecond {
			delay = 10 * time.Millisecond
		}

		frames = append(frames, AnimatedFrame{
			Bitmap: bmp,
			Width:  bmp.Width,
			Height: bmp.Height,
			Delay:  delay,
		})
	}
	return frames, nil
}

// Handle 返回画布底层的原生句柄。
func (c *Canvas) Handle() windows.Handle {
	if c == nil {
		return 0
	}
	return c.hdc
}

// Bounds 返回画布的绘制边界。
func (c *Canvas) Bounds() Rect {
	if c == nil {
		return Rect{}
	}
	return c.bounds
}

// PushClipRect limits subsequent drawing to rect until the returned restore function is called.
func (c *Canvas) PushClipRect(rect Rect) func() {
	if c == nil || c.hdc == 0 || rect.Empty() {
		return func() {}
	}
	saved, _, _ := procSaveDC.Call(uintptr(c.hdc))
	if int32(saved) == 0 {
		return func() {}
	}
	procIntersectClipRect.Call(
		uintptr(c.hdc),
		uintptr(rect.X),
		uintptr(rect.Y),
		uintptr(rect.X+rect.W),
		uintptr(rect.Y+rect.H),
	)
	return func() {
		procRestoreDC.Call(uintptr(c.hdc), ^uintptr(0))
	}
}

// FillRect 在当前画布上填充指定矩形。
func (c *Canvas) FillRect(rect Rect, color Color) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.fillRect(rect, color)
	}
	return c.fillRectGDI(rect, color)
}

// fillRectGDI 使用 GDI 填充矩形。
func (c *Canvas) fillRectGDI(rect Rect, color Color) error {
	brush, err := NewSolidBrush(color)
	if err != nil {
		return err
	}
	defer brush.Close()

	wr := rect.toWinRect()
	r1, _, fillErr := procFillRect.Call(
		uintptr(c.hdc),
		uintptr(unsafe.Pointer(&wr)),
		uintptr(brush.handle),
	)
	if r1 == 0 {
		return wrapError("FillRect", fillErr)
	}
	return nil
}

// Clear 清空当前画布。
func (c *Canvas) Clear(color Color) error {
	return c.FillRect(c.bounds, color)
}

// FillRoundRect 在当前画布上填充指定圆角矩形。
func (c *Canvas) FillRoundRect(rect Rect, radius int32, color Color) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.fillRoundRect(rect, radius, color)
	}
	return c.RoundRect(rect, radius, color, color)
}

// StrokeRoundRect 在当前画布上描边指定圆角矩形。
func (c *Canvas) StrokeRoundRect(rect Rect, radius int32, color Color, width int32) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.strokeRoundRect(rect, radius, color, width)
	}
	return c.strokeRoundRectGDI(rect, radius, color, width)
}

// strokeRoundRectGDI 使用 GDI 绘制圆角矩形边框。
func (c *Canvas) strokeRoundRectGDI(rect Rect, radius int32, color Color, width int32) error {
	if width <= 0 {
		width = 1
	}

	pen, err := NewPen(width, color)
	if err != nil {
		return err
	}
	defer pen.Close()

	hollowBrush, _, brushErr := procGetStockObject.Call(nullBrush)
	if hollowBrush == 0 {
		return wrapError("GetStockObject", brushErr)
	}

	oldBrush, _, _ := procSelectObject.Call(uintptr(c.hdc), hollowBrush)
	oldPen, _, _ := procSelectObject.Call(uintptr(c.hdc), uintptr(pen.handle))
	defer procSelectObject.Call(uintptr(c.hdc), oldBrush)
	defer procSelectObject.Call(uintptr(c.hdc), oldPen)

	r1, _, roundErr := procRoundRect.Call(
		uintptr(c.hdc),
		uintptr(rect.X),
		uintptr(rect.Y),
		uintptr(rect.X+rect.W),
		uintptr(rect.Y+rect.H),
		uintptr(radius),
		uintptr(radius),
	)
	if r1 == 0 {
		return wrapError("RoundRect", roundErr)
	}
	return nil
}

// FillPolygon 在当前画布上填充指定多边形。
func (c *Canvas) FillPolygon(points []Point, color Color) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.fillPolygon(points, color)
	}
	return c.fillPolygonGDI(points, color)
}

// fillPolygonGDI 使用 GDI 填充多边形。
func (c *Canvas) fillPolygonGDI(points []Point, color Color) error {
	if len(points) < 3 {
		return nil
	}

	brush, err := NewSolidBrush(color)
	if err != nil {
		return err
	}
	defer brush.Close()

	hollowPen, _, penErr := procGetStockObject.Call(nullPen)
	if hollowPen == 0 {
		return wrapError("GetStockObject", penErr)
	}

	oldBrush, _, _ := procSelectObject.Call(uintptr(c.hdc), uintptr(brush.handle))
	oldPen, _, _ := procSelectObject.Call(uintptr(c.hdc), hollowPen)
	defer procSelectObject.Call(uintptr(c.hdc), oldBrush)
	defer procSelectObject.Call(uintptr(c.hdc), oldPen)

	r1, _, drawErr := procPolygon.Call(
		uintptr(c.hdc),
		uintptr(unsafe.Pointer(&points[0])),
		uintptr(len(points)),
	)
	if r1 == 0 {
		return wrapError("Polygon", drawErr)
	}
	return nil
}

// RoundRect 在当前画布上绘制指定圆角矩形。
func (c *Canvas) RoundRect(rect Rect, radius int32, fill Color, stroke Color) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		if err := c.d2d.fillRoundRect(rect, radius, fill); err != nil {
			return err
		}
		return c.d2d.strokeRoundRect(rect, radius, stroke, 1)
	}
	return c.roundRectGDI(rect, radius, fill, stroke)
}

// roundRectGDI 使用 GDI 绘制圆角矩形。
func (c *Canvas) roundRectGDI(rect Rect, radius int32, fill Color, stroke Color) error {
	brush, err := NewSolidBrush(fill)
	if err != nil {
		return err
	}
	defer brush.Close()

	pen, err := NewPen(1, stroke)
	if err != nil {
		return err
	}
	defer pen.Close()

	oldBrush, _, _ := procSelectObject.Call(uintptr(c.hdc), uintptr(brush.handle))
	oldPen, _, _ := procSelectObject.Call(uintptr(c.hdc), uintptr(pen.handle))
	defer procSelectObject.Call(uintptr(c.hdc), oldBrush)
	defer procSelectObject.Call(uintptr(c.hdc), oldPen)

	r1, _, roundErr := procRoundRect.Call(
		uintptr(c.hdc),
		uintptr(rect.X),
		uintptr(rect.Y),
		uintptr(rect.X+rect.W),
		uintptr(rect.Y+rect.H),
		uintptr(radius),
		uintptr(radius),
	)
	if r1 == 0 {
		return wrapError("RoundRect", roundErr)
	}
	return nil
}

// DrawText 在当前画布上绘制文本。
func (c *Canvas) DrawText(text string, rect Rect, font *Font, color Color, format uint32) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.drawText(text, rect, font, color, format)
	}
	return c.drawTextGDI(text, rect, font, color, format)
}

// drawTextGDI 使用 GDI 绘制文本。
func (c *Canvas) drawTextGDI(text string, rect Rect, font *Font, color Color, format uint32) error {
	if text == "" {
		return nil
	}
	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return err
	}

	procSetBkMode.Call(uintptr(c.hdc), bkModeTransparent)
	procSetTextColor.Call(uintptr(c.hdc), uintptr(color))

	var oldFont uintptr
	if font != nil && font.handle != 0 {
		oldFont, _, _ = procSelectObject.Call(uintptr(c.hdc), uintptr(font.handle))
		defer procSelectObject.Call(uintptr(c.hdc), oldFont)
	}

	wr := rect.toWinRect()
	r1, _, drawErr := procDrawTextW.Call(
		uintptr(c.hdc),
		uintptr(unsafe.Pointer(ptr)),
		drawTextAutoLen,
		uintptr(unsafe.Pointer(&wr)),
		uintptr(format),
	)
	if r1 == 0 {
		return wrapError("DrawTextW", drawErr)
	}
	return nil
}

// MeasureText 测量渲染给定文本所需的尺寸。
func (c *Canvas) MeasureText(text string, font *Font) (Size, error) {
	if c == nil {
		return Size{}, nil
	}
	if c.d2d != nil {
		return c.d2d.measureText(text, font)
	}
	return c.measureTextGDI(text, font)
}

// measureTextGDI 使用 GDI 测量文本尺寸。
func (c *Canvas) measureTextGDI(text string, font *Font) (Size, error) {
	if text == "" {
		return Size{}, nil
	}

	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return Size{}, err
	}

	var oldFont uintptr
	if font != nil && font.handle != 0 {
		oldFont, _, _ = procSelectObject.Call(uintptr(c.hdc), uintptr(font.handle))
		defer procSelectObject.Call(uintptr(c.hdc), oldFont)
	}

	var size Size
	r1, _, measureErr := procGetTextExtentPoint32W.Call(
		uintptr(c.hdc),
		uintptr(unsafe.Pointer(ptr)),
		uintptr(len([]rune(text))),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return Size{}, wrapError("GetTextExtentPoint32W", measureErr)
	}
	return size, nil
}

// DrawIcon 在当前画布上绘制图标。
func (c *Canvas) DrawIcon(icon *Icon, rect Rect) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		if err := c.d2d.flush(); err != nil {
			return err
		}
	}
	return c.drawIconGDI(icon, rect)
}

// drawIconGDI 使用 GDI 绘制图标。
func (c *Canvas) drawIconGDI(icon *Icon, rect Rect) error {
	if icon == nil || icon.handle == 0 {
		return nil
	}
	r1, _, err := procDrawIconEx.Call(
		uintptr(c.hdc),
		uintptr(rect.X),
		uintptr(rect.Y),
		uintptr(icon.handle),
		uintptr(rect.W),
		uintptr(rect.H),
		0,
		0,
		diNormal,
	)
	if r1 == 0 {
		return wrapError("DrawIconEx", err)
	}
	return nil
}

// DrawBitmapAlpha 在当前画布上按透明度绘制位图。
func (c *Canvas) DrawBitmapAlpha(bitmap *Bitmap, rect Rect, alpha byte) error {
	if c == nil {
		return nil
	}
	if c.d2d != nil {
		return c.d2d.drawBitmap(bitmap, rect, alpha)
	}
	return c.drawBitmapAlphaGDI(bitmap, rect, alpha)
}

// drawBitmapAlphaGDI 使用 GDI 按透明度绘制位图。
func (c *Canvas) drawBitmapAlphaGDI(bitmap *Bitmap, rect Rect, alpha byte) error {
	if bitmap == nil || bitmap.handle == 0 {
		return nil
	}

	srcDC, _, err := procCreateCompatibleDC.Call(uintptr(c.hdc))
	if srcDC == 0 {
		return wrapError("CreateCompatibleDC", err)
	}
	defer procDeleteDC.Call(srcDC)

	old, _, _ := procSelectObject.Call(srcDC, uintptr(bitmap.handle))
	defer procSelectObject.Call(srcDC, old)

	blend := blendFunction{
		BlendOp:             acSrcOver,
		SourceConstantAlpha: alpha,
		AlphaFormat:         acSrcAlpha,
	}
	blendValue := *(*uint32)(unsafe.Pointer(&blend))

	r1, _, alphaErr := procAlphaBlend.Call(
		uintptr(c.hdc),
		uintptr(rect.X),
		uintptr(rect.Y),
		uintptr(rect.W),
		uintptr(rect.H),
		srcDC,
		0,
		0,
		uintptr(bitmap.Width),
		uintptr(bitmap.Height),
		uintptr(blendValue),
	)
	if r1 == 0 {
		return wrapError("AlphaBlend", alphaErr)
	}
	return nil
}

// beginPaintSession 为当前 WM_PAINT 周期创建并初始化绘制会话。
func beginPaintSession(app *App, hwnd windows.Handle, doubleBuffered bool) (*paintSession, error) {
	if renderer, backend := app.rendererForPaint(); renderer != nil && backend == RenderBackendDirect2D {
		session, err := beginD2DPaintSession(app, hwnd, doubleBuffered, renderer)
		if err == nil {
			return session, nil
		}
		app.fallbackToGDI(err)
	}
	return beginGDIPaintSession(app, hwnd, doubleBuffered)
}

// beginGDIPaintSession 创建基于 GDI 的绘制会话。
func beginGDIPaintSession(app *App, hwnd windows.Handle, doubleBuffered bool) (*paintSession, error) {
	session := &paintSession{app: app, hwnd: hwnd}
	paintDC, _, err := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&session.ps)))
	if paintDC == 0 {
		return nil, wrapError("BeginPaint", err)
	}
	session.paintDC = windows.Handle(paintDC)

	bounds, err := clientRect(hwnd)
	if err != nil {
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&session.ps)))
		return nil, err
	}

	if !doubleBuffered || bounds.Empty() {
		session.direct = true
		session.canvas = &Canvas{hdc: session.paintDC, bounds: bounds}
		return session, nil
	}

	memDC, _, dcErr := procCreateCompatibleDC.Call(uintptr(session.paintDC))
	if memDC == 0 {
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&session.ps)))
		return nil, wrapError("CreateCompatibleDC", dcErr)
	}
	session.memDC = windows.Handle(memDC)

	buffer, err := newDIB(bounds.W, bounds.H, session.paintDC)
	if err != nil {
		procDeleteDC.Call(uintptr(session.memDC))
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&session.ps)))
		return nil, err
	}

	session.buffer = buffer
	session.oldBitmap, _, _ = procSelectObject.Call(uintptr(session.memDC), uintptr(buffer.handle))
	session.canvas = &Canvas{hdc: session.memDC, bounds: bounds}
	return session, nil
}

// beginD2DPaintSession 创建基于 Direct2D 的绘制会话。
func beginD2DPaintSession(app *App, hwnd windows.Handle, doubleBuffered bool, renderer *d2dRenderer) (*paintSession, error) {
	session, err := beginGDIPaintSession(app, hwnd, doubleBuffered)
	if err != nil {
		return nil, err
	}
	session.d2d = renderer
	session.canvas.d2d = renderer
	if err := renderer.begin(session.canvas.hdc, session.canvas.bounds); err != nil {
		_ = session.closeRaw()
		return nil, err
	}
	return session, nil
}

// close 释放绘制会话持有的资源。
func (s *paintSession) close() error {
	var retErr error
	if s != nil && s.d2d != nil {
		retErr = s.d2d.end()
		if retErr != nil && s.app != nil {
			s.app.fallbackToGDI(retErr)
		}
	}
	closeErr := s.closeRaw(retErr == nil)
	if retErr != nil {
		return retErr
	}
	return closeErr
}

// closeRaw 释放绘制会话资源，并在需要时把缓冲内容复制到窗口。
func (s *paintSession) closeRaw(doBlit ...bool) error {
	if s == nil {
		return nil
	}

	var retErr error
	shouldBlit := true
	if len(doBlit) > 0 {
		shouldBlit = doBlit[0]
	}
	if shouldBlit && !s.direct && s.canvas != nil && s.buffer != nil {
		r1, _, err := procBitBlt.Call(
			uintptr(s.paintDC),
			0,
			0,
			uintptr(s.canvas.bounds.W),
			uintptr(s.canvas.bounds.H),
			uintptr(s.memDC),
			0,
			0,
			srccopy,
		)
		if r1 == 0 && retErr == nil {
			retErr = wrapError("BitBlt", err)
		}
	}

	if s.memDC != 0 {
		procSelectObject.Call(uintptr(s.memDC), s.oldBitmap)
		procDeleteDC.Call(uintptr(s.memDC))
	}
	if s.buffer != nil {
		_ = s.buffer.Close()
	}
	if s.paintDC != 0 {
		procEndPaint.Call(uintptr(s.hwnd), uintptr(unsafe.Pointer(&s.ps)))
	}
	return retErr
}

// newDIB 分配与给定参考 DC 兼容的设备无关位图。
func newDIB(width, height int32, refDC windows.Handle) (*Bitmap, error) {
	if width <= 0 || height <= 0 {
		return &Bitmap{Width: width, Height: height}, nil
	}

	var bi bitmapInfo
	bi.Header.Size = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.Width = width
	bi.Header.Height = -height
	bi.Header.Planes = 1
	bi.Header.BitCount = 32

	var bits unsafe.Pointer
	hbmp, _, err := procCreateDIBSection.Call(
		uintptr(refDC),
		uintptr(unsafe.Pointer(&bi)),
		0,
		uintptr(unsafe.Pointer(&bits)),
		0,
		0,
	)
	if hbmp == 0 {
		return nil, wrapError("CreateDIBSection", err)
	}

	return &Bitmap{
		handle: windows.Handle(hbmp),
		Width:  width,
		Height: height,
	}, nil
}

// abs32 返回有符号 32 位整数的绝对值。
func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

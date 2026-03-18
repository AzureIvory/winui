//go:build windows

package widgets

import (
	"bytes"
	"github.com/yourname/winui/core"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"time"
)

type ImageScaleMode int

const (
	ImageScaleStretch ImageScaleMode = iota + 1
	ImageScaleContain
	ImageScaleCenter
)

type Image struct {
	widgetBase
	bitmap     *core.Bitmap
	owned      bool
	opacity    byte
	scaleMode  ImageScaleMode
	sourceSize core.Size
}

// NewImage 创建一个新的图像控件。
func NewImage(id string) *Image {
	return &Image{
		widgetBase: newWidgetBase(id, "image"),
		opacity:    255,
		scaleMode:  ImageScaleContain,
	}
}

// SetBounds 更新图像的边界。
func (i *Image) SetBounds(rect Rect) {
	i.runOnUI(func() {
		i.widgetBase.setBounds(i, rect)
	})
}

// SetVisible 更新图像的可见状态。
func (i *Image) SetVisible(visible bool) {
	i.runOnUI(func() {
		i.widgetBase.setVisible(i, visible)
	})
}

// SetEnabled 更新图像的可用状态。
func (i *Image) SetEnabled(enabled bool) {
	i.runOnUI(func() {
		i.widgetBase.setEnabled(i, enabled)
	})
}

// SetScaleMode 更新图像的缩放模式。
func (i *Image) SetScaleMode(mode ImageScaleMode) {
	i.runOnUI(func() {
		if mode == 0 {
			mode = ImageScaleContain
		}
		if i.scaleMode == mode {
			return
		}
		i.scaleMode = mode
		i.invalidate(i)
	})
}

// SetOpacity 更新图像的不透明度。
func (i *Image) SetOpacity(alpha byte) {
	i.runOnUI(func() {
		if i.opacity == alpha {
			return
		}
		i.opacity = alpha
		i.invalidate(i)
	})
}

// SetBitmap 更新图像使用的位图。
func (i *Image) SetBitmap(bitmap *core.Bitmap) {
	i.runOnUI(func() {
		i.replaceBitmap(bitmap, false)
	})
}

// SetBitmapOwned 更新图像控件持有的位图。
func (i *Image) SetBitmapOwned(bitmap *core.Bitmap) {
	i.runOnUI(func() {
		i.replaceBitmap(bitmap, true)
	})
}

// LoadBytes 将字节数据加载到图像控件中。
func (i *Image) LoadBytes(data []byte) error {
	img, err := decodeImage(data)
	if err != nil {
		return err
	}
	bitmap, err := bitmapFromImage(img)
	if err != nil {
		return err
	}
	i.SetBitmapOwned(bitmap)
	return nil
}

// NaturalSize 返回已加载图像数据的原始尺寸。
func (i *Image) NaturalSize() core.Size {
	return i.sourceSize
}

// Bitmap 返回当前附加到图像控件上的位图。
func (i *Image) Bitmap() *core.Bitmap {
	return i.bitmap
}

// OnEvent 处理输入事件或生命周期事件。
func (i *Image) OnEvent(Event) bool {
	return false
}

// Paint 使用给定的绘制上下文完成绘制。
func (i *Image) Paint(ctx *PaintCtx) {
	if !i.Visible() || ctx == nil {
		return
	}
	if i.bitmap == nil {
		return
	}
	_ = ctx.canvas.DrawBitmapAlpha(i.bitmap, i.drawRect(), i.opacity)
}

// Close 释放图像控件持有的资源。
func (i *Image) Close() error {
	if i.owned && i.bitmap != nil {
		err := i.bitmap.Close()
		i.bitmap = nil
		i.owned = false
		i.sourceSize = core.Size{}
		return err
	}
	i.bitmap = nil
	i.owned = false
	i.sourceSize = core.Size{}
	return nil
}

// replaceBitmap 替换图像控件持有的位图。
func (i *Image) replaceBitmap(bitmap *core.Bitmap, owned bool) {
	if i.bitmap == bitmap && i.owned == owned {
		return
	}
	if i.owned && i.bitmap != nil && i.bitmap != bitmap {
		_ = i.bitmap.Close()
	}
	i.bitmap = bitmap
	i.owned = owned
	if bitmap != nil {
		i.sourceSize = core.Size{Width: bitmap.Width, Height: bitmap.Height}
	} else {
		i.sourceSize = core.Size{}
	}
	i.invalidate(i)
}

// drawRect 返回图像在当前画布上的绘制矩形。
func (i *Image) drawRect() Rect {
	target := i.Bounds()
	if i.bitmap == nil || target.Empty() {
		return target
	}

	srcW := i.sourceSize.Width
	srcH := i.sourceSize.Height
	if srcW <= 0 || srcH <= 0 {
		return target
	}

	switch i.scaleMode {
	case ImageScaleCenter:
		x := target.X + (target.W-srcW)/2
		y := target.Y + (target.H-srcH)/2
		return Rect{X: x, Y: y, W: srcW, H: srcH}
	case ImageScaleStretch:
		return target
	default:
		width := target.W
		height := target.H
		if width <= 0 || height <= 0 {
			return target
		}
		scaleX := float64(width) / float64(srcW)
		scaleY := float64(height) / float64(srcH)
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		if scale <= 0 {
			scale = 1
		}
		drawW := int32(float64(srcW) * scale)
		drawH := int32(float64(srcH) * scale)
		x := target.X + (target.W-drawW)/2
		y := target.Y + (target.H-drawH)/2
		return Rect{X: x, Y: y, W: drawW, H: drawH}
	}
}

type AnimatedImage struct {
	widgetBase
	frames      []core.AnimatedFrame
	ownedFrames bool
	frameIndex  int
	playing     bool
	opacity     byte
	scaleMode   ImageScaleMode
	timerID     uintptr
}

// NewAnimatedImage 创建一个新的动画图像控件。
func NewAnimatedImage(id string) *AnimatedImage {
	return &AnimatedImage{
		widgetBase: newWidgetBase(id, "animated-image"),
		playing:    true,
		opacity:    255,
		scaleMode:  ImageScaleContain,
	}
}

// SetBounds 更新动画图像的边界。
func (a *AnimatedImage) SetBounds(rect Rect) {
	a.runOnUI(func() {
		a.widgetBase.setBounds(a, rect)
	})
}

// SetVisible 更新动画图像的可见状态。
func (a *AnimatedImage) SetVisible(visible bool) {
	a.runOnUI(func() {
		a.widgetBase.setVisible(a, visible)
		a.syncTimer()
	})
}

// SetEnabled 更新动画图像的可用状态。
func (a *AnimatedImage) SetEnabled(enabled bool) {
	a.runOnUI(func() {
		a.widgetBase.setEnabled(a, enabled)
	})
}

// setScene 更新动画图像关联的场景。
func (a *AnimatedImage) setScene(scene *Scene) {
	oldScene := a.scene()
	if oldScene == scene {
		a.widgetBase.setScene(scene)
		return
	}
	if oldScene != nil && a.timerID != 0 {
		_ = oldScene.stopTimer(a.timerID)
		a.timerID = 0
	}
	a.widgetBase.setScene(scene)
	a.syncTimer()
}

// SetScaleMode 更新动画图像的缩放模式。
func (a *AnimatedImage) SetScaleMode(mode ImageScaleMode) {
	a.runOnUI(func() {
		if mode == 0 {
			mode = ImageScaleContain
		}
		if a.scaleMode == mode {
			return
		}
		a.scaleMode = mode
		a.invalidate(a)
	})
}

// SetOpacity 更新动画图像的不透明度。
func (a *AnimatedImage) SetOpacity(alpha byte) {
	a.runOnUI(func() {
		if a.opacity == alpha {
			return
		}
		a.opacity = alpha
		a.invalidate(a)
	})
}

// SetPlaying 更新动画图像的播放状态。
func (a *AnimatedImage) SetPlaying(playing bool) {
	a.runOnUI(func() {
		if a.playing == playing {
			return
		}
		a.playing = playing
		a.syncTimer()
	})
}

// LoadGIF 将 GIF 数据加载到动画图像中。
func (a *AnimatedImage) LoadGIF(data []byte) error {
	frames, err := core.DecodeGIF(data)
	if err != nil {
		return err
	}
	a.SetFramesOwned(frames)
	return nil
}

// SetFrames 更新动画图像的帧集合。
func (a *AnimatedImage) SetFrames(frames []core.AnimatedFrame) {
	a.runOnUI(func() {
		a.replaceFrames(frames, false)
	})
}

// SetFramesOwned 更新动画图像持有的帧集合。
func (a *AnimatedImage) SetFramesOwned(frames []core.AnimatedFrame) {
	a.runOnUI(func() {
		a.replaceFrames(frames, true)
	})
}

// NaturalSize 返回已加载图像数据的原始尺寸。
func (a *AnimatedImage) NaturalSize() core.Size {
	if len(a.frames) == 0 {
		return core.Size{}
	}
	return core.Size{Width: a.frames[0].Width, Height: a.frames[0].Height}
}

// CurrentFrame 返回动画图像当前显示的帧索引。
func (a *AnimatedImage) CurrentFrame() int {
	return a.frameIndex
}

// OnEvent 处理输入事件或生命周期事件。
func (a *AnimatedImage) OnEvent(evt Event) bool {
	if evt.Type != EventTimer || evt.TimerID != a.timerID || len(a.frames) == 0 {
		return false
	}
	if !a.playing || !a.Visible() {
		return false
	}
	a.frameIndex = (a.frameIndex + 1) % len(a.frames)
	a.invalidate(a)
	a.syncTimer()
	return true
}

// Paint 使用给定的绘制上下文完成绘制。
func (a *AnimatedImage) Paint(ctx *PaintCtx) {
	if !a.Visible() || ctx == nil || len(a.frames) == 0 {
		return
	}
	frame := a.frames[a.frameIndex%len(a.frames)]
	if frame.Bitmap == nil {
		return
	}
	_ = ctx.canvas.DrawBitmapAlpha(frame.Bitmap, a.drawRect(), a.opacity)
}

// Close 释放动画图像持有的资源。
func (a *AnimatedImage) Close() error {
	if a.scene() != nil && a.timerID != 0 {
		_ = a.scene().stopTimer(a.timerID)
	}
	a.timerID = 0
	if a.ownedFrames {
		for i := range a.frames {
			if a.frames[i].Bitmap != nil {
				_ = a.frames[i].Bitmap.Close()
			}
		}
	}
	a.frames = nil
	a.ownedFrames = false
	a.frameIndex = 0
	return nil
}

// replaceFrames 替换动画图像持有的帧集合。
func (a *AnimatedImage) replaceFrames(frames []core.AnimatedFrame, owned bool) {
	if a.ownedFrames {
		for i := range a.frames {
			if a.frames[i].Bitmap != nil {
				_ = a.frames[i].Bitmap.Close()
			}
		}
	}
	a.frames = append(a.frames[:0], frames...)
	a.ownedFrames = owned
	a.frameIndex = 0
	a.invalidate(a)
	a.syncTimer()
}

// syncTimer 让动画定时器与当前播放状态保持一致。
func (a *AnimatedImage) syncTimer() {
	scene := a.scene()
	if scene == nil {
		return
	}
	if !a.playing || !a.Visible() || len(a.frames) <= 1 {
		if a.timerID != 0 {
			_ = scene.stopTimer(a.timerID)
			a.timerID = 0
		}
		return
	}

	delay := a.frames[a.frameIndex].Delay
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}
	if a.timerID == 0 {
		timerID, err := scene.startTimer(a, 0, delay)
		if err == nil {
			a.timerID = timerID
		}
		return
	}
	_ = scene.updateTimer(a.timerID, delay)
}

// drawRect 返回动画图像在当前画布上的绘制矩形。
func (a *AnimatedImage) drawRect() Rect {
	target := a.Bounds()
	if len(a.frames) == 0 || target.Empty() {
		return target
	}
	frame := a.frames[a.frameIndex%len(a.frames)]
	srcW := frame.Width
	srcH := frame.Height
	if srcW <= 0 || srcH <= 0 {
		return target
	}

	switch a.scaleMode {
	case ImageScaleCenter:
		x := target.X + (target.W-srcW)/2
		y := target.Y + (target.H-srcH)/2
		return Rect{X: x, Y: y, W: srcW, H: srcH}
	case ImageScaleStretch:
		return target
	default:
		scaleX := float64(target.W) / float64(srcW)
		scaleY := float64(target.H) / float64(srcH)
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}
		if scale <= 0 {
			scale = 1
		}
		drawW := int32(float64(srcW) * scale)
		drawH := int32(float64(srcH) * scale)
		x := target.X + (target.W-drawW)/2
		y := target.Y + (target.H-drawH)/2
		return Rect{X: x, Y: y, W: drawW, H: drawH}
	}
}

// decodeImage 将图像数据解码为便于 Go 使用的结构。
func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// bitmapFromImage 将解码后的 Go 图像转换为 core 位图。
func bitmapFromImage(img image.Image) (*core.Bitmap, error) {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	return core.BitmapFromRGBA(rgba)
}

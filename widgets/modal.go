//go:build windows

package widgets

import (
	"image"

	"github.com/AzureIvory/winui/core"
)

// Modal provides a lightweight scene-level dialog container with backdrop behavior.
type Modal struct {
	Panel
	backdropColor       core.Color
	backdropOpacity     byte
	blurRadiusDP        int32
	dismissOnBackdrop   bool
	onDismiss           func()
	backdropBitmap      *core.Bitmap
	backdropBitmapColor core.Color
}

// NewModal creates a new modal container.
func NewModal(id string) *Modal {
	modal := &Modal{
		Panel:             *NewPanel(id),
		backdropColor:     core.RGB(0, 0, 0),
		backdropOpacity:   112,
		dismissOnBackdrop: true,
	}
	return modal
}

// SetBackdropColor updates the backdrop tint color.
func (m *Modal) SetBackdropColor(color core.Color) {
	m.runOnUI(func() {
		if m.backdropColor == color {
			return
		}
		m.backdropColor = color
		m.releaseBackdropBitmap()
		m.invalidate(m)
	})
}

// BackdropColor returns the backdrop tint color.
func (m *Modal) BackdropColor() core.Color {
	return m.backdropColor
}

// SetBackdropOpacity updates the backdrop alpha value.
func (m *Modal) SetBackdropOpacity(opacity byte) {
	m.runOnUI(func() {
		if m.backdropOpacity == opacity {
			return
		}
		m.backdropOpacity = opacity
		m.invalidate(m)
	})
}

// BackdropOpacity returns the backdrop alpha value.
func (m *Modal) BackdropOpacity() byte {
	return m.backdropOpacity
}

// SetBlurRadiusDP updates the requested blur radius in logical pixels.
func (m *Modal) SetBlurRadiusDP(radius int32) {
	m.runOnUI(func() {
		if radius < 0 {
			radius = 0
		}
		if m.blurRadiusDP == radius {
			return
		}
		m.blurRadiusDP = radius
		m.invalidate(m)
	})
}

// BlurRadiusDP returns the requested blur radius.
func (m *Modal) BlurRadiusDP() int32 {
	return m.blurRadiusDP
}

// SetDismissOnBackdrop updates whether backdrop clicks dismiss the modal.
func (m *Modal) SetDismissOnBackdrop(enabled bool) {
	m.runOnUI(func() {
		if m.dismissOnBackdrop == enabled {
			return
		}
		m.dismissOnBackdrop = enabled
	})
}

// DismissOnBackdrop reports whether backdrop clicks dismiss the modal.
func (m *Modal) DismissOnBackdrop() bool {
	return m.dismissOnBackdrop
}

// SetOnDismiss registers a dismissal callback.
func (m *Modal) SetOnDismiss(fn func()) {
	m.runOnUI(func() {
		m.onDismiss = fn
	})
}

// OnEvent handles backdrop dismiss clicks while preserving normal panel behavior.
func (m *Modal) OnEvent(evt Event) bool {
	if evt.Type == EventClick && evt.Source == m && m.dismissOnBackdrop {
		if m.onDismiss != nil {
			m.onDismiss()
		}
		return true
	}
	return m.Panel.OnEvent(evt)
}

// Paint draws the modal backdrop first, then its children.
func (m *Modal) Paint(ctx *PaintCtx) {
	if !m.Visible() || ctx == nil {
		return
	}
	bounds := m.Bounds()
	if !bounds.Empty() && m.backdropOpacity > 0 {
		if m.shouldPaintBlur(ctx) {
			m.paintBlurTint(ctx, bounds)
		}
		if bitmap := m.backdropFillBitmap(); bitmap != nil {
			_ = ctx.Canvas().DrawBitmapAlpha(bitmap, bounds, m.backdropOpacity)
		}
	}
	if m.Style.Background != 0 || m.Style.BorderColor != 0 {
		radius := ctx.DP(m.Style.CornerRadius)
		if m.Style.Background != 0 {
			if radius > 0 {
				_ = ctx.FillRoundRect(bounds, radius, m.Style.Background)
			} else {
				_ = ctx.FillRect(bounds, m.Style.Background)
			}
		}
		if m.Style.BorderColor != 0 {
			width := m.Style.BorderWidth
			if width <= 0 {
				width = 1
			}
			_ = ctx.StrokeRoundRect(bounds, max32(0, radius), m.Style.BorderColor, width)
		}
	}
	for _, child := range m.Children() {
		if child.Visible() {
			child.Paint(ctx)
		}
	}
}

func (m *Modal) overlayHitTest(x, y int32) bool {
	if !m.Visible() || !m.Enabled() || !m.Bounds().Contains(x, y) {
		return false
	}
	for i := len(m.children) - 1; i >= 0; i-- {
		if widgetMouseHit(m.children[i], x, y) {
			return false
		}
	}
	return true
}

// Close releases modal-owned graphical resources.
func (m *Modal) Close() error {
	m.releaseBackdropBitmap()
	return nil
}

func (m *Modal) shouldPaintBlur(ctx *PaintCtx) bool {
	if m.blurRadiusDP <= 0 || ctx == nil || ctx.scene == nil || ctx.scene.app == nil {
		return false
	}
	return ctx.scene.app.RenderBackend() == core.RenderBackendDirect2D
}

func (m *Modal) paintBlurTint(ctx *PaintCtx, bounds Rect) {
	intensity := byte(24)
	if m.blurRadiusDP >= 12 {
		intensity = 36
	}
	if bitmap := solidBitmap(core.RGB(255, 255, 255)); bitmap != nil {
		defer bitmap.Close()
		_ = ctx.Canvas().DrawBitmapAlpha(bitmap, bounds, intensity)
	}
}

func (m *Modal) backdropFillBitmap() *core.Bitmap {
	if m.backdropBitmap != nil && m.backdropBitmapColor == m.backdropColor {
		return m.backdropBitmap
	}
	m.releaseBackdropBitmap()
	m.backdropBitmap = solidBitmap(m.backdropColor)
	m.backdropBitmapColor = m.backdropColor
	return m.backdropBitmap
}

func (m *Modal) releaseBackdropBitmap() {
	if m.backdropBitmap != nil {
		_ = m.backdropBitmap.Close()
		m.backdropBitmap = nil
	}
	m.backdropBitmapColor = 0
}

func solidBitmap(color core.Color) *core.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Pix[0] = byte(color & 0xFF)
	img.Pix[1] = byte((color >> 8) & 0xFF)
	img.Pix[2] = byte((color >> 16) & 0xFF)
	img.Pix[3] = 0xFF
	bitmap, err := core.BitmapFromRGBA(img)
	if err != nil {
		return nil
	}
	return bitmap
}

//go:build windows

package widgets

var (
	procCreateRectRgn = nativeGdi32.NewProc("CreateRectRgn")
	procDeleteObject  = nativeGdi32.NewProc("DeleteObject")
	procSetWindowRgn  = nativeUser32.NewProc("SetWindowRgn")
	procRedrawWindow  = nativeUser32.NewProc("RedrawWindow")
)

// nativeHostGeometry caches only values. Windows owns every successfully
// installed region; no region handle is retained or deleted by this cache.
type nativeHostGeometry struct {
	bounds        Rect
	boundsSet     bool
	clip          Rect
	clipSet       bool
	regionApplied bool
}

// nativeHostClip uses the same ancestor clipBounds as scene painting/hit tests.
// Window regions are local to the outer HWND, including its nonclient area.
func nativeHostClip(owner Widget, bounds Rect) Rect {
	visible := bounds
	for node := asWidgetNode(owner); node != nil; {
		if !node.Visible() {
			return Rect{}
		}
		parent := node.parent()
		if parent == nil {
			break
		}
		if clipper, ok := parent.(clipBoundsWidget); ok {
			visible = intersectRect(visible, clipper.clipBounds())
			if visible.Empty() {
				return Rect{}
			}
		}
		node = asWidgetNode(parent)
	}
	if visible.Empty() {
		return Rect{}
	}
	return Rect{X: visible.X - bounds.X, Y: visible.Y - bounds.Y, W: visible.W, H: visible.H}
}

func (s *nativeControlState) syncBounds(owner Widget, bounds Rect) {
	if !s.valid() {
		return
	}
	clip := nativeHostClip(owner, bounds)
	needsRegion := clip != (Rect{W: bounds.W, H: bounds.H})
	clipChanged := !s.geometry.clipSet || s.geometry.clip != clip || s.geometry.regionApplied != needsRegion
	if clipChanged {
		if s.syncClip(bounds, clip) {
			s.geometry.clip = clip
			s.geometry.clipSet = true
		}
	}
	boundsChanged := !s.geometry.boundsSet || s.geometry.bounds != bounds
	if boundsChanged && setNativeBounds(s.handle, bounds) {
		s.geometry.bounds = bounds
		s.geometry.boundsSet = true
	}
	if clipChanged || boundsChanged {
		// Region and placement changes suppress immediate drawing until both
		// describe the new host, avoiding transient paints over adjacent controls.
		procRedrawWindow.Call(uintptr(s.handle), 0, 0, 0x0001|0x0004|0x0400) // INVALIDATE | ERASE | FRAME
	}
}

func (s *nativeControlState) syncClip(bounds, clip Rect) bool {
	full := Rect{W: bounds.W, H: bounds.H}
	if clip == full {
		if !s.geometry.regionApplied {
			return true
		}
		ok, _, _ := procSetWindowRgn.Call(uintptr(s.handle), 0, 0)
		if ok != 0 {
			s.geometry.regionApplied = false
		}
		return ok != 0
	}
	region, _, _ := procCreateRectRgn.Call(uintptr(clip.X), uintptr(clip.Y),
		uintptr(clip.X+clip.W), uintptr(clip.Y+clip.H))
	if region == 0 {
		return false
	}
	ok, _, _ := procSetWindowRgn.Call(uintptr(s.handle), region, 0)
	if ok == 0 {
		// Ownership transfers only on success. Failed installs remain ours.
		procDeleteObject.Call(region)
		return false
	}
	s.geometry.regionApplied = true
	return true
}

// syncNativeTreeGeometry also covers clip changes without any child movement.
func syncNativeTreeGeometry(root Widget) {
	Walk(root, func(widget Widget) bool {
		if native, ok := widget.(nativeAfterMove); ok {
			native.syncNativeAfterTreeMove()
		}
		return true
	})
}

//go:build windows

package widgets

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

const (
	nativeEditReplaceSel   = 0x00C2
	nativeEditGetScrollPos = 0x04DD
	nativeEditSetScrollPos = 0x04DE
)

type nativeScrollInfo struct {
	Size, Mask    uint32
	Min, Max      int32
	Page          uint32
	Pos, TrackPos int32
}

// SetTextPreservingView updates a cumulative text snapshot without changing the
// user's selection or reading position. A viewport already at the bottom follows
// new text; scrolling away pauses following until the user returns to the bottom.
// It does not report OnChange or add an undo step. SetText retains reset behavior.
// During IME composition only the latest snapshot is deferred, and user edits,
// SetText or detachment discard it rather than replaying over a committed edit.
func (e *EditBox) SetTextPreservingView(text string) {
	e.runOnUI(func() {
		if e.imeComposing {
			e.pendingTextUpdate = &text
			return
		}
		e.setTextPreservingView(e.normalizeText(text))
	})
}

func (e *EditBox) flushPendingTextUpdate() {
	pending := e.pendingTextUpdate
	e.pendingTextUpdate = nil
	base := e.imeBaseText
	e.imeBaseText = ""
	if pending == nil || !e.native.valid() || e.normalizeText(getNativeText(e.native.handle)) != base {
		return
	}
	e.setTextPreservingView(e.normalizeText(*pending))
}

func (e *EditBox) setTextPreservingView(text string) {
	old := e.Text
	if old == text {
		return
	}
	style := e.resolveStyle(&PaintCtx{scene: e.scene()})
	content := e.contentRect(style)
	before := e.textLayout(style, content.W)
	follow := e.multiline && e.scrollY >= max32(0, before.contentHeight-content.H)-1
	oldCaret := e.caret
	change := editTextChange(old, text)
	var selectionStart, selectionEnd int
	var scroll core.Point
	nativeView := e.native.valid() && (isNativeMode(e.mode) || e.Focused)
	if e.native.valid() {
		selectionStart, selectionEnd = getNativeSelection(e.native.handle)
		scroll = nativeEditScrollPosition(e)
		if nativeView {
			follow = e.multiline && nativeEditAtBottom(e)
		}
	}
	e.Text = text
	e.caret = change.mapRunePosition(oldCaret)
	e.clampScroll()
	if follow {
		e.scrollY = max32(0, e.textLayout(style, content.W).contentHeight-content.H)
	}
	if e.native.valid() {
		e.applyingText = true
		defer func() { e.applyingText = false }()
		// Only visible native surfaces need redraw suppression. Re-enabling
		// WM_SETREDRAW on a hidden child would incorrectly make it visible.
		styleIndex := int32(-16) // GWL_STYLE
		nativeStyle, _, _ := nativeUser32.NewProc("GetWindowLongW").Call(uintptr(e.native.handle), uintptr(styleIndex))
		redraw := nativeStyle&uintptr(nativeWindowVisible) != 0
		if redraw {
			sendNativeMessage(e.native.handle, 0x000B, 0, 0) // WM_SETREDRAW
		}
		start := editUTF16Length(old[:change.startByte])
		end := editUTF16Length(old[:change.oldEndByte])
		setNativeSelection(e.native.handle, start, end)
		// FALSE excludes programmatic stream chunks from the user's undo log.
		replacement := text[change.startByte:change.newEndByte]
		if e.multiline {
			replacement = strings.ReplaceAll(replacement, "\n", "\r\n")
		}
		if ptr, err := windows.UTF16PtrFromString(replacement); err == nil {
			sendNativeMessage(e.native.handle, nativeEditReplaceSel, 0, uintptr(unsafe.Pointer(ptr)))
		}
		e.syncNativeInsets()
		selectionStart = change.mapUTF16Position(selectionStart, old, text)
		selectionEnd = change.mapUTF16Position(selectionEnd, old, text)
		setNativeSelection(e.native.handle, selectionStart, selectionEnd)
		e.caret = editRunePosition(text, selectionEnd)
		if nativeView && follow {
			sendNativeMessage(e.native.handle, 0x0115, 7, 0) // WM_VSCROLL / SB_BOTTOM
			scroll.Y = nativeEditScrollPosition(e).Y
		}
		// Restore after selection; never use ScrollToCaret for a snapshot update.
		sendNativeMessage(e.native.handle, nativeEditSetScrollPos, 0, uintptr(unsafe.Pointer(&scroll)))
		if nativeView {
			actual := nativeEditScrollPosition(e)
			e.scrollX, e.scrollY = actual.X, actual.Y
		}
		if redraw {
			sendNativeMessage(e.native.handle, 0x000B, 1, 0)
			procRedrawWindow.Call(uintptr(e.native.handle), 0, 0, 0x0001|0x0004|0x0400)
		}
	}
	e.invalidate(e)
}

func nativeEditAtBottom(e *EditBox) bool {
	info, ok := nativeEditScrollInfo(e, 1)
	return !ok || info.Pos >= info.Max-max32(0, int32(info.Page)-1)-1
}

func nativeEditScrollInfo(e *EditBox, axis int32) (nativeScrollInfo, bool) {
	info := nativeScrollInfo{Mask: 0x0001 | 0x0002 | 0x0004} // SIF_RANGE | PAGE | POS
	info.Size = uint32(unsafe.Sizeof(info))
	ok, _, _ := nativeUser32.NewProc("GetScrollInfo").Call(uintptr(e.native.handle), uintptr(axis), uintptr(unsafe.Pointer(&info)))
	return info, ok != 0
}

func nativeEditScrollPosition(e *EditBox) core.Point {
	var point core.Point
	sendNativeMessage(e.native.handle, nativeEditGetScrollPos, 0, uintptr(unsafe.Pointer(&point)))
	// EM_GETSCROLLPOS documents a 16-bit limit. Scrollbar positions retain
	// their full 32-bit range for documents taller/wider than 65535 pixels.
	if info, ok := nativeEditScrollInfo(e, 0); ok {
		point.X = info.Pos
	}
	if info, ok := nativeEditScrollInfo(e, 1); ok {
		point.Y = info.Pos
	}
	return point
}

// A single replacement preserves unchanged prefixes/suffixes and never splits
// UTF-8 code points. Native RichEdit character positions count UTF-16 units.
type editTextDelta struct {
	startByte, oldEndByte, newEndByte int
	startRune, oldEndRune, newEndRune int
}

func editTextChange(old, text string) editTextDelta {
	start := 0
	for start < len(old) && start < len(text) {
		a, n := utf8.DecodeRuneInString(old[start:])
		b, m := utf8.DecodeRuneInString(text[start:])
		if a != b || n != m {
			break
		}
		start += n
	}
	oldEnd, newEnd := len(old), len(text)
	for oldEnd > start && newEnd > start {
		a, n := utf8.DecodeLastRuneInString(old[start:oldEnd])
		b, m := utf8.DecodeLastRuneInString(text[start:newEnd])
		if a != b || n != m {
			break
		}
		oldEnd -= n
		newEnd -= m
	}
	return editTextDelta{start, oldEnd, newEnd, utf8.RuneCountInString(old[:start]), utf8.RuneCountInString(old[:oldEnd]), utf8.RuneCountInString(text[:newEnd])}
}

func mapEditPosition(position, start, oldEnd, newEnd int) int {
	if position <= start {
		return position
	}
	if position >= oldEnd {
		return position + newEnd - oldEnd
	}
	return start + min(position-start, newEnd-start)
}

func (d editTextDelta) mapRunePosition(position int) int {
	return mapEditPosition(position, d.startRune, d.oldEndRune, d.newEndRune)
}

func (d editTextDelta) mapUTF16Position(position int, old, text string) int {
	return editUTF16Position(text, d.mapRunePosition(editRunePosition(old, position)))
}

func editUTF16Position(text string, position int) int {
	remaining := max(0, position)
	units := 0
	for _, r := range text {
		if remaining == 0 {
			break
		}
		units += utf16.RuneLen(r)
		remaining--
	}
	return units
}

func editUTF16Length(text string) int {
	length := 0
	for _, r := range text {
		length += utf16.RuneLen(r)
	}
	return length
}

func editRunePosition(text string, position int) int {
	count, units := 0, 0
	for _, r := range text {
		if units+utf16.RuneLen(r) > position {
			break
		}
		units += utf16.RuneLen(r)
		count++
	}
	return count
}

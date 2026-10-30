//go:build windows

package widgets

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

func editSnapshotScroll(edit *EditBox) core.Point {
	var point core.Point
	sendNativeMessage(edit.native.handle, nativeEditGetScrollPos, 0, uintptr(unsafe.Pointer(&point)))
	return point
}

func TestEditTextSnapshotCustomViewportFollowsOnlyAtBottom(t *testing.T) {
	edit := NewEditBox("snapshot", ModeCustom)
	edit.SetMultiline(true)
	edit.SetBounds(Rect{W: 240, H: 100})
	edit.SetText("first")
	changes := 0
	edit.SetOnChange(func(string) { changes++ })
	text := "first" + strings.Repeat("\nline 😀", 30)
	edit.SetTextPreservingView(text)
	style := edit.resolveStyle(nil)
	maxY := func() int32 {
		return edit.textLayout(style, edit.contentRect(style).W).contentHeight - edit.contentRect(style).H
	}
	if edit.scrollY != maxY() || edit.caret != 5 {
		t.Fatalf("custom snapshot lost tail following/caret: scroll=%d max=%d caret=%d", edit.scrollY, maxY(), edit.caret)
	}
	edit.OnEvent(Event{Type: EventMouseWheel, Delta: 240})
	before := edit.scrollY
	text += "\nnew content"
	edit.SetTextPreservingView(text)
	if edit.scrollY != before || changes != 0 {
		t.Fatalf("custom snapshot stole manual reading: before=%d after=%d changes=%d", before, edit.scrollY, changes)
	}
	edit.OnEvent(Event{Type: EventMouseWheel, Delta: -12000})
	text += "\nfollowing again"
	edit.SetTextPreservingView(text)
	if edit.scrollY != maxY() {
		t.Fatal("returning to the custom bottom did not resume following")
	}
}

func TestEditTextSnapshotNativePreservesSelectionAndManualScrolling(t *testing.T) {
	for _, mode := range []ControlMode{ModeCustom, ModeNative} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			var edit *EditBox
			app, onUI := scrollTestNativeApp(t, func(app *core.App, scene *Scene) {
				edit = NewEditBox("snapshot", mode)
				edit.SetMultiline(true)
				edit.SetBounds(Rect{X: 20, Y: 20, W: 300, H: 120})
				edit.SetText("😀标题\n" + editScrollTestText())
				scene.Root().Add(edit)
				scene.FocusWidget(edit)
			})
			onUI(func() {
				changes := 0
				edit.SetOnChange(func(string) { changes++ })
				for _, dpi := range []uintptr{96, 144, 192} {
					sendNativeMessage(app.Handle(), 0x02E0, dpi|dpi<<16, 0)
					edit.SetBounds(Rect{X: app.DP(20), Y: app.DP(20), W: app.DP(300), H: app.DP(120)})
					setNativeSelection(edit.native.handle, 2, 7)
					sendNativeMessage(edit.native.handle, 0x0115, 7, 0)
					text := edit.TextValue() + "\n新增 😀\n末尾"
					edit.SetTextPreservingView(text)
					if !nativeEditAtBottom(edit) || edit.TextValue() != text {
						t.Fatalf("DPI %d snapshot lost bottom/text: bottom=%v text=%q", dpi, nativeEditAtBottom(edit), edit.TextValue())
					}
					if start, end := getNativeSelection(edit.native.handle); start != 2 || end != 7 {
						t.Fatalf("DPI %d snapshot collapsed UTF16 selection: %d,%d", dpi, start, end)
					}
					if undo := sendNativeMessage(edit.native.handle, 0x00C6, 0, 0); undo != 0 { // EM_CANUNDO
						t.Fatal("programmatic stream created a user undo entry")
					}
					sendNativeMessage(edit.native.handle, 0x0115, 2, 0) // SB_PAGEUP
					before := editSnapshotScroll(edit)
					if nativeEditAtBottom(edit) {
						t.Fatal("fixture page-up did not leave the bottom")
					}
					for i := 0; i < 4; i++ {
						text += "\nstream chunk 😀"
						edit.SetTextPreservingView(text)
						if got := editSnapshotScroll(edit); got != before {
							t.Fatalf("DPI %d snapshot moved manual viewport: before=%+v after=%+v", dpi, before, got)
						}
					}
					sendNativeMessage(edit.native.handle, 0x0115, 7, 0)
					edit.SetTextPreservingView(text + "\nfollow resumed")
					if !nativeEditAtBottom(edit) {
						t.Fatal("returning to native bottom did not resume following")
					}
					if changes != 0 {
						t.Fatalf("programmatic snapshots emitted %d OnChange callbacks", changes)
					}
				}
			})
		})
	}
}

func TestEditTextSnapshotDoesNotShowUnfocusedCustomNativeChild(t *testing.T) {
	var edit *EditBox
	_, onUI := scrollTestNativeApp(t, func(_ *core.App, scene *Scene) {
		edit = NewEditBox("unfocused", ModeCustom)
		edit.SetMultiline(true)
		edit.SetBounds(Rect{W: 260, H: 100})
		edit.SetText("short")
		scene.Root().Add(edit)
	})
	onUI(func() {
		edit.SetTextPreservingView("short" + strings.Repeat("\nline", 30))
		edit.OnEvent(Event{Type: EventMouseWheel, Delta: 240})
		before := edit.scrollY
		edit.SetTextPreservingView(edit.Text + "\nnew")
		index := int32(-16)
		style, _, _ := nativeUser32.NewProc("GetWindowLongW").Call(uintptr(edit.native.handle), uintptr(index))
		if edit.Focused || style&uintptr(nativeWindowVisible) != 0 || edit.scrollY != before {
			t.Fatalf("snapshot showed hidden input or moved custom viewport: focused=%v nativeStyle=%x scroll=%d before=%d", edit.Focused, style, edit.scrollY, before)
		}
	})
}

func TestEditTextSnapshotChangedRangePreservesUnicodeSelectionAndHorizontalScroll(t *testing.T) {
	var edit *EditBox
	_, onUI := scrollTestNativeApp(t, func(_ *core.App, scene *Scene) {
		edit = NewEditBox("changed-range", ModeCustom)
		edit.SetMultiline(true)
		edit.SetHorizontalScroll(true)
		edit.SetWordWrap(false)
		edit.SetBounds(Rect{W: 180, H: 100})
		edit.SetText("首😀\n选中😀" + strings.Repeat("long content ", 30) + "\n" + editScrollTestText())
		scene.Root().Add(edit)
		scene.FocusWidget(edit)
	})
	onUI(func() {
		changes := 0
		edit.SetOnChange(func(string) { changes++ })
		selected := func() string {
			buffer := make([]uint16, 64)
			sendNativeMessage(edit.native.handle, 0x043E, 0, uintptr(unsafe.Pointer(&buffer[0]))) // EM_GETSELTEXT
			return windows.UTF16ToString(buffer)
		}
		start := editUTF16Length("首😀\n")
		setNativeSelection(edit.native.handle, start, start+editUTF16Length("选中😀"))
		if got := selected(); got != "选中😀" {
			t.Fatalf("fixture UTF16 selection=%q", got)
		}
		point := core.Point{X: 80, Y: 80}
		sendNativeMessage(edit.native.handle, nativeEditSetScrollPos, 0, uintptr(unsafe.Pointer(&point)))
		before := editSnapshotScroll(edit)
		if before.X == 0 {
			t.Fatal("fixture did not scroll horizontally")
		}
		text := "新增🧪\n" + edit.Text
		edit.SetTextPreservingView(text)
		if got := selected(); got != "选中😀" {
			t.Fatalf("changed range shifted selected content to %q", got)
		}
		if got := editSnapshotScroll(edit); got != before {
			t.Fatalf("changed range moved viewport: %+v -> %+v", before, got)
		}
		edit.SetReadOnly(true)
		text += "\nread-only stream tail"
		edit.SetTextPreservingView(text)
		if edit.TextValue() != text || selected() != "选中😀" || changes != 0 {
			t.Fatalf("read-only programmatic update changed selection/user callbacks or failed: text=%q selection=%q changes=%d", edit.TextValue(), selected(), changes)
		}
		edit.SetReadOnly(false)
		edit.SetText("ab")
		setNativeSelection(edit.native.handle, 1, 1)
		edit.SetTextPreservingView("😀q")
		if start, end := getNativeSelection(edit.native.handle); start != 2 || end != 2 {
			t.Fatalf("changed-region caret split a surrogate pair: %d,%d", start, end)
		}
		if edit.TextValue() != "😀q" || edit.caret != 1 {
			t.Fatalf("native text query corrupted logical caret: text=%q caret=%d", edit.Text, edit.caret)
		}
	})
}

func TestEditTextSnapshotIMEPendingCannotOverwriteUserEditsOrReset(t *testing.T) {
	var edit *EditBox
	_, onUI := scrollTestNativeApp(t, func(_ *core.App, scene *Scene) {
		edit = NewEditBox("ime-snapshot", ModeCustom)
		edit.SetMultiline(true)
		edit.SetBounds(Rect{W: 260, H: 120})
		edit.SetText("base")
		scene.Root().Add(edit)
		scene.FocusWidget(edit)
	})
	onUI(func() {
		changes := 0
		edit.SetOnChange(func(string) { changes++ })
		handle := edit.native.handle
		sendNativeMessage(handle, 0x010D, 0, 0)
		edit.SetTextPreservingView("base first")
		edit.SetTextPreservingView("base latest")
		if edit.TextValue() != "base" {
			t.Fatal("snapshot interrupted IME composition")
		}
		sendNativeMessage(handle, 0x010E, 0, 0)
		if edit.TextValue() != "base latest" || changes != 0 {
			t.Fatal("unchanged IME did not apply latest snapshot silently")
		}
		sendNativeMessage(handle, 0x010D, 0, 0)
		edit.SetTextPreservingView("stale stream")
		setNativeSelection(handle, -1, -1)
		ptr, err := windows.UTF16PtrFromString("用户编辑")
		if err != nil {
			t.Fatal(err)
		}
		sendNativeMessage(handle, nativeEditReplaceSel, 1, uintptr(unsafe.Pointer(ptr)))
		userText := edit.TextValue()
		sendNativeMessage(handle, 0x010E, 0, 0)
		if edit.TextValue() != userText || changes != 1 {
			t.Fatalf("pending snapshot overwrote user edit or lost change: text=%q changes=%d", edit.TextValue(), changes)
		}
		sendNativeMessage(handle, 0x010D, 0, 0)
		edit.SetTextPreservingView("stale after reset")
		edit.SetText("new epoch")
		sendNativeMessage(handle, 0x010E, 0, 0)
		if edit.TextValue() != "new epoch" {
			t.Fatal("ordinary SetText did not cancel pending snapshot")
		}
		sendNativeMessage(handle, 0x010D, 0, 0)
		edit.SetTextPreservingView("stale after close")
		edit.Close()
		if edit.pendingTextUpdate != nil {
			t.Fatal("destroyed editor retained pending snapshot")
		}
	})
}

func TestEditTextSnapshotKeepsLongDocumentScrollBeyond16BitRange(t *testing.T) {
	var edit *EditBox
	_, onUI := scrollTestNativeApp(t, func(_ *core.App, scene *Scene) {
		edit = NewEditBox("long-snapshot", ModeCustom)
		edit.SetMultiline(true)
		edit.SetBounds(Rect{W: 260, H: 120})
		edit.SetText(strings.Repeat("长行 😀 abc\n", 6000))
		scene.Root().Add(edit)
		scene.FocusWidget(edit)
	})
	var before, after core.Point
	onUI(func() {
		// RichEdit can initially format only a few screenfuls. Complete native
		// layout before requesting a position beyond that temporary range.
		sendNativeMessage(edit.native.handle, 0x0115, 7, 0) // WM_VSCROLL / SB_BOTTOM
		point := core.Point{Y: 70000}
		sendNativeMessage(edit.native.handle, nativeEditSetScrollPos, 0, uintptr(unsafe.Pointer(&point)))
		before = nativeEditScrollPosition(edit)
		if before.Y <= 65535 {
			return
		}
		edit.SetTextPreservingView(edit.Text + "追加末尾\n")
		after = nativeEditScrollPosition(edit)
	})
	// Fatal assertions belong to the test goroutine; Goexit on the UI thread
	// would terminate its message pump and deadlock app.Run during cleanup.
	if before.Y <= 65535 {
		t.Fatalf("fixture did not enter long document range: %+v", before)
	}
	if after != before {
		t.Fatalf("long snapshot wrapped scroll: %+v -> %+v", before, after)
	}
}

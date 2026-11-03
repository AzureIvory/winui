//go:build windows

package core

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	trayShell32                = windows.NewLazySystemDLL("shell32.dll")
	procShellNotifyIconW       = trayShell32.NewProc("Shell_NotifyIconW")
	procShellNotifyIconGetRect = trayShell32.NewProc("Shell_NotifyIconGetRect")
	procRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	procCopyIcon               = user32.NewProc("CopyIcon")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procEndMenu                = user32.NewProc("EndMenu")
)

const (
	nimAdd             = 0
	nimModify          = 1
	nimDelete          = 2
	nimSetFocus        = 3
	nimSetVersion      = 4
	nifMessage         = 1
	nifIcon            = 2
	nifTip             = 4
	nifGUID            = 0x20
	nifShowTip         = 0x80
	notifyIconVersion4 = 4
	ninSelect          = 0x400
	ninKeySelect       = 0x401
	wmTrayContextMenu  = 0x7b
)

// notifyIconDataW follows the modern shellapi.h NOTIFYICONDATAW, including
// the uTimeout/uVersion union's single UINT storage slot.
type notifyIconDataW struct {
	Size                uint32
	Window              windows.Handle
	ID, Flags, Callback uint32
	Icon                windows.Handle
	Tip                 [128]uint16
	State, StateMask    uint32
	Info                [256]uint16
	Version             uint32
	InfoTitle           [64]uint16
	InfoFlags           uint32
	GUID                windows.GUID
	BalloonIcon         windows.Handle
}

type notifyIconIdentifier struct {
	Size   uint32
	Window windows.Handle
	ID     uint32
	GUID   windows.GUID
}

// Compile-time ABI invariants on both pointer widths. These cover the whole
// structure and the alignment boundaries most likely to drift during edits.
const trayPointerSize = unsafe.Sizeof(uintptr(0))
const trayDataSize = 956 + (trayPointerSize-4)*5
const trayIconOffset = 20 + (trayPointerSize-4)*3
const trayGUIDOffset = 936 + (trayPointerSize-4)*4

var (
	_ [trayDataSize - unsafe.Sizeof(notifyIconDataW{})]byte
	_ [unsafe.Sizeof(notifyIconDataW{}) - trayDataSize]byte
	_ [trayPointerSize - unsafe.Offsetof(notifyIconDataW{}.Window)]byte
	_ [unsafe.Offsetof(notifyIconDataW{}.Window) - trayPointerSize]byte
	_ [trayIconOffset - unsafe.Offsetof(notifyIconDataW{}.Icon)]byte
	_ [unsafe.Offsetof(notifyIconDataW{}.Icon) - trayIconOffset]byte
	_ [trayGUIDOffset - unsafe.Offsetof(notifyIconDataW{}.GUID)]byte
	_ [unsafe.Offsetof(notifyIconDataW{}.GUID) - trayGUIDOffset]byte
	_ [28 + (trayPointerSize-4)*3 - unsafe.Sizeof(notifyIconIdentifier{})]byte
	_ [unsafe.Sizeof(notifyIconIdentifier{}) - (28 + (trayPointerSize-4)*3)]byte
)

func (t *TrayIcon) menuAnchor() (int32, int32, error) {
	identifier := notifyIconIdentifier{Size: uint32(unsafe.Sizeof(notifyIconIdentifier{})), Window: t.app.Handle(), ID: uint32(t.id), GUID: t.guid}
	var bounds winRect
	hr, _, _ := procShellNotifyIconGetRect.Call(uintptr(unsafe.Pointer(&identifier)), uintptr(unsafe.Pointer(&bounds)))
	if int32(hr) < 0 {
		return 0, 0, fmt.Errorf("winui/core: Shell_NotifyIconGetRect failed (HRESULT 0x%08X)", uint32(hr))
	}
	return bounds.Left + (bounds.Right-bounds.Left)/2, bounds.Bottom, nil
}

func notifyTray(operation uint32, data *notifyIconDataW) error {
	ok, _, _ := procShellNotifyIconW.Call(uintptr(operation), uintptr(unsafe.Pointer(data)))
	if ok == 0 {
		// Shell_NotifyIcon documents BOOL, not a reliable GetLastError contract.
		return fmt.Errorf("winui/core: Shell_NotifyIconW operation %d failed", operation)
	}
	return nil
}

func registerTaskbarMessage() (uint32, error) {
	name, err := windows.UTF16PtrFromString("TaskbarCreated")
	if err != nil {
		return 0, err
	}
	id, _, callErr := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(name)))
	if id == 0 {
		return 0, wrapError("RegisterWindowMessageW(TaskbarCreated)", callErr)
	}
	return uint32(id), nil
}

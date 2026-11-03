//go:build windows

package core

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
)

// TrayCommandID identifies a command within one immutable menu snapshot.
// Zero is reserved for separators and cancellation.
type TrayCommandID uint32

// TrayMenuItemKind describes the supported native menu rows.
type TrayMenuItemKind uint8

const (
	TrayMenuCommand TrayMenuItemKind = iota
	TrayMenuCheckbox
	TrayMenuSeparator
)

// TrayMenuItem is presentation data. Checkboxes never toggle themselves;
// the host updates the menu from its authoritative state through SetMenu.
type TrayMenuItem struct {
	ID       TrayCommandID
	Text     string
	Kind     TrayMenuItemKind
	Disabled bool
	Checked  bool
}

// TrayIconOptions configures one notification-area icon owned by an App.
// Callbacks execute on that App's UI thread and must not block its message pump.
// OnError reports recovery and automatic/deferred cleanup failures. During
// shutdown the App is already closed; callbacks must not assume a usable HWND.
type TrayIconOptions struct {
	Icon    *Icon
	Tooltip string
	Menu    []TrayMenuItem
	// GUID optionally supplies a stable Shell identity. Zero uses HWND + ID.
	// The host coordinates GUID uniqueness across processes.
	GUID       windows.GUID
	OnActivate func()
	OnCommand  func(TrayCommandID)
	OnError    func(error)
}

var ErrTrayClosed = errors.New("winui/core: tray icon closed")

// Process-local reservation also prevents two Apps from using the same GUID
// while a previous icon's Shell deletion remains unconfirmed.
var trayGUIDOwners sync.Map

type trayManager struct {
	closing, ownerGone bool // UI-thread only; closing never resets.
	nextID             uint32
	taskbarMessage     uint32
	icons              map[uint16]*TrayIcon
	menuOwner          *TrayIcon
}

type trayCloseResult struct{ err error }

// TrayIcon owns private CopyIcon copies and native menu resources. All mutable
// state is UI-thread confined except the published close result. No extra HWND,
// thread, goroutine or message loop is created.
type TrayIcon struct {
	app                           *App
	id                            uint16
	guid                          windows.GUID
	options                       TrayIconOptions
	tooltip                       [128]uint16
	menu                          []TrayMenuItem
	icon                          windows.Handle
	ownedIcons                    map[windows.Handle]struct{}
	ownedMenus                    map[windows.Handle]struct{}
	shellPresent, ready, retired  bool
	busy                          int
	cleaning, reporting, tracking bool
	// heldMenu belongs to a showMenu stack, including its construction and
	// foreground setup before native tracking begins. Cleanup must not free it.
	heldMenu windows.Handle
	closed   atomic.Pointer[trayCloseResult]
}

// NewTrayIcon synchronously registers an icon on the existing App UI thread.
// Call after Init or from OnCreate. Icon must remain valid throughout this call;
// the caller retains it and may close it after the call returns.
func (a *App) NewTrayIcon(options TrayIconOptions) (*TrayIcon, error) {
	tip, err := trayTooltip(options.Tooltip)
	if err != nil {
		return nil, err
	}
	menu, err := trayMenuSnapshot(options.Menu)
	if err != nil {
		return nil, err
	}
	var tray *TrayIcon
	err = a.callOnUI(func() error {
		m := &a.trays
		if m.closing {
			return ErrAppClosed
		}
		if options.Icon == nil || options.Icon.Handle() == 0 {
			return fmt.Errorf("winui/core: tray icon is required")
		}
		if m.nextID == 0xffff {
			return fmt.Errorf("winui/core: tray callback IDs exhausted")
		}
		if m.taskbarMessage == 0 {
			id, err := registerTaskbarMessage()
			if err != nil {
				return err
			}
			m.taskbarMessage = id
		}
		if m.closing || a.closed.Load() {
			return ErrAppClosed
		}
		m.nextID++ // Never reuse a published callback identity, even after rollback.
		t := &TrayIcon{app: a, id: uint16(m.nextID), guid: options.GUID, options: options,
			tooltip: tip, menu: menu, ownedIcons: make(map[windows.Handle]struct{}), ownedMenus: make(map[windows.Handle]struct{})}
		// Do not retain caller-owned icon or menu slices in callback options.
		t.options.Icon, t.options.Menu = nil, nil
		if t.guid != (windows.GUID{}) {
			if _, exists := trayGUIDOwners.LoadOrStore(t.guid, t); exists {
				return fmt.Errorf("winui/core: tray GUID already reserved")
			}
		}
		if m.icons == nil {
			m.icons = make(map[uint16]*TrayIcon)
		}
		m.icons[t.id] = t // Also owns partially initialized and retired resources.
		t.busy++
		icon, operationErr := t.copyIcon(options.Icon)
		if operationErr == nil {
			t.icon = icon
			if !t.available() {
				operationErr = ErrAppClosed
			} else {
				operationErr = t.add()
			}
		}
		t.busy--
		if operationErr != nil || !t.available() {
			if operationErr == nil {
				operationErr = ErrAppClosed
			}
			t.retire()
			return errors.Join(operationErr, t.cleanup())
		}
		t.ready = true
		tray = t
		return nil
	})
	return tray, err
}

func trayTooltip(text string) ([128]uint16, error) {
	var result [128]uint16
	if !utf8.ValidString(text) {
		return result, fmt.Errorf("winui/core: tooltip must be valid UTF-8")
	}
	units, err := windows.UTF16FromString(text)
	if err != nil {
		return result, fmt.Errorf("winui/core: tooltip: %w", err)
	}
	if len(units) > len(result) {
		return result, fmt.Errorf("winui/core: tooltip exceeds 127 UTF-16 units")
	}
	copy(result[:], units)
	return result, nil
}

func trayMenuSnapshot(items []TrayMenuItem) ([]TrayMenuItem, error) {
	if len(items) > 0xffff {
		return nil, fmt.Errorf("winui/core: tray menu too large")
	}
	result := append([]TrayMenuItem(nil), items...)
	ids := make(map[TrayCommandID]struct{}, len(items))
	for _, item := range result {
		if item.Kind == TrayMenuSeparator {
			if item.ID != 0 || item.Text != "" || item.Disabled || item.Checked {
				return nil, fmt.Errorf("winui/core: separator must contain only its kind")
			}
			continue
		}
		if item.Kind != TrayMenuCommand && item.Kind != TrayMenuCheckbox {
			return nil, fmt.Errorf("winui/core: invalid tray menu kind")
		}
		if item.ID == 0 {
			return nil, fmt.Errorf("winui/core: tray command ID must not be zero")
		}
		if _, exists := ids[item.ID]; exists {
			return nil, fmt.Errorf("winui/core: duplicate tray command ID %d", item.ID)
		}
		ids[item.ID] = struct{}{}
		if !utf8.ValidString(item.Text) || item.Text == "" {
			return nil, fmt.Errorf("winui/core: tray command text must be nonempty UTF-8")
		}
		if _, err := windows.UTF16FromString(item.Text); err != nil {
			return nil, fmt.Errorf("winui/core: tray command text: %w", err)
		}
		if item.Kind != TrayMenuCheckbox && item.Checked {
			return nil, fmt.Errorf("winui/core: only checkbox commands can be checked")
		}
	}
	return result, nil
}

func (t *TrayIcon) available() bool {
	return !t.retired && !t.app.trays.closing && !t.app.closed.Load()
}

func (t *TrayIcon) data() notifyIconDataW {
	d := notifyIconDataW{Size: uint32(unsafe.Sizeof(notifyIconDataW{})), Window: t.app.Handle(), ID: uint32(t.id)}
	if t.guid != (windows.GUID{}) {
		d.Flags = nifGUID
		d.GUID = t.guid
	}
	return d
}

func (t *TrayIcon) add() error {
	d := t.data()
	d.Flags |= nifMessage | nifIcon | nifTip | nifShowTip
	d.Callback, d.Icon, d.Tip = wmAppTray, t.icon, t.tooltip
	if err := notifyTray(nimAdd, &d); err != nil {
		return err
	}
	t.shellPresent = true
	if !t.available() {
		return ErrAppClosed
	}
	d = t.data()
	d.Version = notifyIconVersion4
	if err := notifyTray(nimSetVersion, &d); err != nil {
		return err
	}
	if !t.available() {
		return ErrAppClosed
	}
	return nil
}

func (t *TrayIcon) copyIcon(icon *Icon) (windows.Handle, error) {
	if icon == nil || icon.Handle() == 0 {
		return 0, fmt.Errorf("winui/core: tray icon is required")
	}
	handle, _, err := procCopyIcon.Call(uintptr(icon.Handle()))
	if handle == 0 {
		return 0, wrapError("CopyIcon(tray)", err)
	}
	h := windows.Handle(handle)
	t.ownedIcons[h] = struct{}{}
	return h, nil
}

func (t *TrayIcon) change(fn func() error) error {
	if t == nil || t.app == nil {
		return ErrTrayClosed
	}
	return t.app.callOnUI(func() error {
		if !t.available() {
			return ErrTrayClosed
		}
		if t.busy != 0 {
			return fmt.Errorf("winui/core: tray operation already in progress")
		}
		t.busy++
		err := fn()
		t.busy--
		if !t.available() {
			return errors.Join(err, ErrTrayClosed, t.cleanup())
		}
		return err
	})
}

// SetTooltip validates the complete UTF-16 string and commits only after a
// successful Shell update. Embedded NUL and overlength input are errors.
func (t *TrayIcon) SetTooltip(text string) error {
	tip, err := trayTooltip(text)
	if err != nil {
		return err
	}
	return t.change(func() error {
		if !t.shellPresent || !t.ready {
			return fmt.Errorf("winui/core: tray is not registered")
		}
		d := t.data()
		d.Flags |= nifTip | nifShowTip
		d.Tip = tip
		if err := notifyTray(nimModify, &d); err != nil {
			return err
		}
		if t.available() {
			t.tooltip = tip
		}
		return nil
	})
}

// SetMenu atomically replaces copied presentation data. An open native menu
// retains its original snapshot; the replacement appears on the next opening.
func (t *TrayIcon) SetMenu(items []TrayMenuItem) error {
	menu, err := trayMenuSnapshot(items)
	if err != nil {
		return err
	}
	return t.change(func() error { t.menu = menu; return nil })
}

// SetIcon copies the caller's valid icon and retains caller ownership. A failed
// replacement leaves the previous accepted image intact; failed destruction of
// any private copy remains recorded for later cleanup.
func (t *TrayIcon) SetIcon(icon *Icon) error {
	return t.change(func() error {
		if !t.shellPresent || !t.ready {
			return fmt.Errorf("winui/core: tray is not registered")
		}
		h, err := t.copyIcon(icon)
		if err != nil {
			return err
		}
		if !t.available() {
			return ErrTrayClosed
		}
		d := t.data()
		d.Flags |= nifIcon
		d.Icon = h
		if err := notifyTray(nimModify, &d); err != nil {
			return errors.Join(err, t.destroyIcon(h))
		}
		if !t.available() {
			return ErrTrayClosed
		}
		old := t.icon
		t.icon = h
		return t.destroyIcon(old)
	})
}

func (t *TrayIcon) retire() {
	t.retired, t.ready = true, false
	if t.closed.Load() == nil {
		t.closed.Store(&trayCloseResult{})
	}
}

// Close irreversibly retires callback identity and attempts native cleanup.
// It is idempotent; a later call can retry residual resources while the App is
// alive. Inside TrackPopupMenu it requests EndMenu without waiting: that stack
// owns HMENU until tracking returns, and subsequent cleanup errors go to OnError.
// After App shutdown Close only reads the final published result, never posts.
func (t *TrayIcon) Close() error {
	if t == nil || t.app == nil {
		return nil
	}
	if t.app.closed.Load() {
		if result := t.closed.Load(); result != nil {
			return result.err
		}
		return ErrAppClosed
	}
	err := t.app.callOnUI(func() error { t.retire(); return t.cleanup() })
	if errors.Is(err, ErrAppClosed) {
		if result := t.closed.Load(); result != nil {
			return result.err
		}
	}
	return err
}

func (t *TrayIcon) destroyIcon(h windows.Handle) error {
	if h == 0 {
		return nil
	}
	t.busy++
	ok, _, err := procDestroyIcon.Call(uintptr(h))
	t.busy--
	if ok == 0 {
		return wrapError("DestroyIcon(tray)", err)
	}
	delete(t.ownedIcons, h)
	return nil
}

func (t *TrayIcon) destroyMenu(h windows.Handle) error {
	t.busy++
	ok, _, err := procDestroyMenu.Call(uintptr(h))
	t.busy--
	if ok == 0 {
		return wrapError("DestroyMenu(tray)", err)
	}
	delete(t.ownedMenus, h)
	return nil
}

// cleanup never uses Post/callOnUI and is safe after the early App.closed fence.
// Resource records survive failed native destruction. A reentrant call cannot
// retry the native operation already on this thread's stack.
func (t *TrayIcon) cleanup() error {
	if t.cleaning || t.busy != 0 {
		if result := t.closed.Load(); result != nil {
			return result.err
		}
		return nil
	}
	t.cleaning = true
	var errs []error
	if t.tracking {
		if ok, _, err := procEndMenu.Call(); ok == 0 {
			errs = append(errs, wrapError("EndMenu(tray)", err))
		}
	}
	if t.shellPresent {
		if t.app.trays.ownerGone {
			errs = append(errs, fmt.Errorf("winui/core: tray Shell deletion unconfirmed before owner destruction"))
		} else {
			d := t.data()
			if err := notifyTray(nimDelete, &d); err != nil {
				errs = append(errs, err)
			} else {
				t.shellPresent = false
			}
		}
	}
	for h := range t.ownedIcons {
		if err := t.destroyIcon(h); err != nil {
			errs = append(errs, err)
		}
	}
	for h := range t.ownedMenus {
		if h != t.heldMenu {
			if err := t.destroyMenu(h); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if !t.shellPresent && t.guid != (windows.GUID{}) {
		trayGUIDOwners.CompareAndDelete(t.guid, t)
	}
	err := errors.Join(errs...)
	t.closed.Store(&trayCloseResult{err: err})
	t.cleaning = false
	return err
}

func (t *TrayIcon) report(err error) {
	if err == nil || t.options.OnError == nil || t.reporting {
		return
	}
	t.reporting = true
	defer func() { t.reporting = false }()
	t.options.OnError(err)
}

func (a *App) closeTrays() {
	a.trays.closing = true
	a.closed.Store(true)
	icons := a.traySnapshot()
	for _, t := range icons {
		t.retire()
	}
	// Set every terminal identity and cleanup result before calling user code.
	errs := make([]error, len(icons))
	for i, t := range icons {
		errs[i] = t.cleanup()
	}
	for i, t := range icons {
		t.report(errs[i])
	}
}

func (a *App) traySnapshot() []*TrayIcon {
	icons := make([]*TrayIcon, 0, len(a.trays.icons))
	for _, t := range a.trays.icons {
		icons = append(icons, t)
	}
	return icons
}

func (a *App) handleTrayMessage(message uint32, wParam, lParam uintptr) bool {
	m := &a.trays
	if m.taskbarMessage != 0 && message == m.taskbarMessage {
		icons := a.traySnapshot()
		// The broadcast invalidates all old Shell registrations, including retired
		// identities with an earlier unconfirmed NIM_DELETE. Never restore those.
		for _, t := range icons {
			t.shellPresent = false
			if t.retired && t.busy == 0 && !t.cleaning && t.guid != (windows.GUID{}) {
				trayGUIDOwners.CompareAndDelete(t.guid, t)
			}
		}
		for _, t := range icons {
			if !t.available() || t.busy != 0 {
				continue
			}
			t.busy++
			err := t.add()
			if err != nil && t.shellPresent && t.available() {
				// SETVERSION failed. Keep accepted configuration but remove this
				// unversioned registration until a later TaskbarCreated can retry.
				d := t.data()
				if deleteErr := notifyTray(nimDelete, &d); deleteErr != nil {
					err = errors.Join(err, deleteErr)
				} else {
					t.shellPresent = false
				}
			}
			t.busy--
			if !t.available() {
				err = errors.Join(err, t.cleanup())
			}
			t.ready = err == nil && t.available()
			t.report(err)
		}
		return true
	}
	if message != wmAppTray {
		return false
	}
	t := m.icons[uint16(lParam>>16)]
	if t == nil || !t.available() || !t.ready || t.busy != 0 {
		return true
	}
	switch uint16(lParam) {
	case ninSelect, ninKeySelect:
		if t.options.OnActivate != nil {
			t.options.OnActivate()
		}
	case wmTrayContextMenu:
		t.showMenu()
	}
	return true
}

func (t *TrayIcon) showMenu() {
	m := &t.app.trays
	if !t.available() || m.menuOwner != nil || len(t.menu) == 0 {
		return
	}
	m.menuOwner = t
	defer func() { m.menuOwner = nil }()
	items := t.menu // SetMenu replaces rather than mutates this snapshot.
	handle, _, callErr := procCreatePopupMenu.Call()
	if handle == 0 {
		t.report(wrapError("CreatePopupMenu(tray)", callErr))
		return
	}
	h := windows.Handle(handle)
	t.ownedMenus[h] = struct{}{}
	t.heldMenu = h
	// This stack is the only releaser, including reentry during construction or
	// SetForegroundWindow. Retired cleanup retries other residual resources.
	finish := func(err error) error {
		t.heldMenu = 0
		if t.retired {
			return errors.Join(err, t.cleanup())
		}
		err = errors.Join(err, t.destroyMenu(h))
		if t.retired {
			err = errors.Join(err, t.cleanup())
		}
		return err
	}
	var err error
	for i, item := range items {
		if !t.available() {
			err = ErrTrayClosed
			break
		}
		flags := uintptr(0)
		var text *uint16
		if item.Kind == TrayMenuSeparator {
			flags = 0x800
		} else {
			text, _ = windows.UTF16PtrFromString(item.Text) // validated snapshot
			if item.Disabled {
				flags |= 1
			}
			if item.Checked {
				flags |= 8
			}
		}
		if ok, _, callErr := procAppendMenuW.Call(handle, flags, uintptr(i+1), uintptr(unsafe.Pointer(text))); ok == 0 {
			err = wrapError("AppendMenuW(tray)", callErr)
			break
		}
	}
	if err != nil || !t.available() {
		t.report(finish(err))
		return
	}
	// WM_CONTEXTMENU is outside v4's documented coordinate-bearing events.
	// Query the actual icon rectangle for mouse and keyboard alike; signed
	// screen coordinates also preserve monitors left/above the desktop origin.
	x, y, anchorErr := t.menuAnchor()
	if anchorErr != nil {
		t.report(finish(anchorErr))
		return
	}
	if !t.available() {
		t.report(finish(nil))
		return
	}
	// Reserve this thread's active menu before SetForegroundWindow can reenter.
	if ok, _, _ := procSetForegroundWindow.Call(uintptr(t.app.Handle())); ok == 0 {
		err = fmt.Errorf("winui/core: foreground activation for tray menu declined")
	}
	var selected uintptr
	if err == nil && t.available() {
		t.tracking = true
		// Zero means cancel OR error; do not fabricate a last-error diagnosis.
		selected, _, _ = procTrackPopupMenu.Call(handle, 0x100|0x80|0x02, uintptr(x), uintptr(y), 0, uintptr(t.app.Handle()), 0)
		t.tracking = false // Only this returning stack releases its HMENU.
	}
	err = finish(err)
	if t.available() {
		if ok, _, callErr := procPostMessageW.Call(uintptr(t.app.Handle()), 0, 0, 0); ok == 0 {
			err = errors.Join(err, wrapError("PostMessageW(tray WM_NULL)", callErr))
		}
		if selected == 0 && t.available() {
			d := t.data()
			err = errors.Join(err, notifyTray(nimSetFocus, &d))
		}
	}
	t.report(err)
	// OnError may have closed this icon/App. Validate again before user command.
	if t.available() && selected > 0 && selected <= uintptr(len(items)) {
		item := items[selected-1]
		if item.Kind != TrayMenuSeparator && !item.Disabled && t.options.OnCommand != nil {
			t.options.OnCommand(item.ID)
		}
	}
}

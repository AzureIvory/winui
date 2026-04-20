//go:build windows

package sysapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/AzureIvory/winui/core"
	"golang.org/x/sys/windows"
)

var (
	dialogOle32  = windows.NewLazySystemDLL("ole32.dll")
	dialogShell  = windows.NewLazySystemDLL("shell32.dll")
	procCoInit   = dialogOle32.NewProc("CoInitializeEx")
	procCoUninit = dialogOle32.NewProc("CoUninitialize")
	procCreate   = dialogOle32.NewProc("CoCreateInstance")
	procFreeMem  = dialogOle32.NewProc("CoTaskMemFree")
	procSHCreate = dialogShell.NewProc("SHCreateItemFromParsingName")
)

const (
	coInitApartmentThreaded = 0x2
	clsctxInprocServer      = 0x1

	fosOverwritePrompt = 0x2
	fosNoChangeDir     = 0x8
	fosPickFolders     = 0x20
	fosForceFilesystem = 0x40
	fosAllowMulti      = 0x200
	fosPathMustExist   = 0x800
	fosFileMustExist   = 0x1000
	fosCreatePrompt    = 0x2000

	sigdnFileSysPath = 0x80058000

	hResultCanceled    = 0x800704C7
	hResultChangedMode = 0x80010106
)

var (
	clsidFileOpenDialog = dialogGUID{0xDC1C5A9C, 0xE88A, 0x4DDE, [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	iidFileOpenDialog   = dialogGUID{0xD57C7288, 0xD4AD, 0x4768, [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
	clsidFileSaveDialog = dialogGUID{0xC0B4E2F3, 0xBA21, 0x4773, [8]byte{0x8D, 0xBA, 0x33, 0x5E, 0xC9, 0x46, 0xEB, 0x8B}}
	iidFileSaveDialog   = dialogGUID{0x84BCCD23, 0x5FDE, 0x4CDB, [8]byte{0xAE, 0xA4, 0xAF, 0x64, 0xB8, 0x3D, 0x78, 0xAB}}
	iidShellItem        = dialogGUID{0x43826D1E, 0xE718, 0x42EE, [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE}}
)

// DialogMode 表示原生文件对话框的运行模式。
type DialogMode string

const (
	// DialogOpen 表示打开文件。
	DialogOpen DialogMode = "open"
	// DialogSave 表示保存文件。
	DialogSave DialogMode = "save"
	// DialogFolder 表示选择目录。
	DialogFolder DialogMode = "folder"
)

// FileFilter 描述原生文件对话框中的一个过滤器项。
type FileFilter struct {
	// Name 是显示给用户的过滤器名称。
	Name string
	// Pattern 是 Win32 通配符模式，例如 `*.txt;*.md`。
	Pattern string
}

// Options 描述文件对话框的可选参数。
type Options struct {
	// Mode 指定对话框模式；空值时默认为 `open`。
	Mode DialogMode
	// Title 指定窗口标题。
	Title string
	// InitialPath 指定初始目录，或包含默认文件名的路径。
	InitialPath string
	// DefaultExtension 指定保存模式下默认扩展名，例如 `txt`。
	DefaultExtension string
	// ButtonLabel 指定确认按钮文本。
	ButtonLabel string
	// Filters 指定文件过滤器列表。
	Filters []FileFilter
	// MultiSelect 指定是否允许选择多个文件，仅对 `open` 生效。
	MultiSelect bool
	// PathMustExist 指定初始目录或目标目录是否必须存在；零值按模式采用默认行为。
	PathMustExist bool
	// FileMustExist 指定打开模式下文件是否必须存在；零值按模式采用默认行为。
	FileMustExist bool
	// CreatePrompt 指定保存模式下是否在新建文件前提示。
	CreatePrompt bool
	// OverwritePrompt 指定保存模式下是否在覆盖文件前提示；零值按模式采用默认行为。
	OverwritePrompt bool
}

// Result 表示文件对话框返回的路径集合。
type Result struct {
	Paths []string
}

// Path 返回第一个结果路径；未选择时返回空字符串。
func (r Result) Path() string {
	if len(r.Paths) == 0 {
		return ""
	}
	return r.Paths[0]
}

// Canceled 报告对话框是否以“未选择任何路径”的方式结束。
func (r Result) Canceled() bool {
	return len(r.Paths) == 0
}

// ShowFileDialog 显示一个原生文件对话框。
func ShowFileDialog(app *core.App, opts Options) (Result, error) {
	opts = normalizeOptions(opts)
	if opts.Mode == DialogSave && opts.MultiSelect {
		return Result{}, fmt.Errorf("save dialog does not support MultiSelect")
	}
	if opts.Mode == DialogFolder && opts.MultiSelect {
		return Result{}, fmt.Errorf("folder dialog does not support MultiSelect")
	}
	if app == nil {
		return showFileDialogNative(0, opts)
	}
	if app.Handle() == 0 {
		return Result{}, core.ErrNotInitialized
	}
	if app.IsUIThread() {
		return showFileDialogNative(app.Handle(), opts)
	}

	type dialogResult struct {
		result Result
		err    error
	}
	ch := make(chan dialogResult, 1)
	if err := app.Post(func() {
		result, showErr := showFileDialogNative(app.Handle(), opts)
		ch <- dialogResult{result: result, err: showErr}
	}); err != nil {
		return Result{}, err
	}
	done := <-ch
	return done.result, done.err
}

// OpenFile 显示一个单选打开文件对话框，并返回首个结果。
func OpenFile(app *core.App, opts Options) (string, error) {
	opts.Mode = DialogOpen
	opts.MultiSelect = false
	result, err := ShowFileDialog(app, opts)
	return result.Path(), err
}

// OpenFiles 显示一个多选打开文件对话框。
func OpenFiles(app *core.App, opts Options) ([]string, error) {
	opts.Mode = DialogOpen
	opts.MultiSelect = true
	result, err := ShowFileDialog(app, opts)
	return result.Paths, err
}

// SaveFile 显示一个保存文件对话框，并返回最终路径。
func SaveFile(app *core.App, opts Options) (string, error) {
	opts.Mode = DialogSave
	result, err := ShowFileDialog(app, opts)
	return result.Path(), err
}

// PickFolder 显示一个目录选择对话框，并返回所选目录。
func PickFolder(app *core.App, opts Options) (string, error) {
	opts.Mode = DialogFolder
	opts.MultiSelect = false
	result, err := ShowFileDialog(app, opts)
	return result.Path(), err
}

type dialogGUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type dialogFilterSpec struct {
	Name *uint16
	Spec *uint16
}

type fileOpenDialog struct {
	lpVtbl *fileOpenDialogVtbl
}

type fileOpenDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
	GetResults          uintptr
	GetSelectedItems    uintptr
}

type fileSaveDialog struct {
	lpVtbl *fileSaveDialogVtbl
}

type fileSaveDialogVtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	Show                uintptr
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
	SetSaveAsItem       uintptr
	SetProperties       uintptr
	SetCollectedProps   uintptr
	GetProperties       uintptr
	ApplyProperties     uintptr
}

type shellItem struct {
	lpVtbl *shellItemVtbl
}

type shellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

type shellItemArray struct {
	lpVtbl *shellItemArrayVtbl
}

type shellItemArrayVtbl struct {
	QueryInterface         uintptr
	AddRef                 uintptr
	Release                uintptr
	BindToHandler          uintptr
	GetPropertyStore       uintptr
	GetPropertyDescription uintptr
	GetAttributes          uintptr
	GetCount               uintptr
	GetItemAt              uintptr
	EnumItems              uintptr
}

func normalizeOptions(opts Options) Options {
	switch opts.Mode {
	case DialogSave, DialogFolder:
	default:
		opts.Mode = DialogOpen
	}
	return opts
}

func showFileDialogNative(owner windows.Handle, opts Options) (Result, error) {
	needUninit, err := coInitialize()
	if err != nil {
		return Result{}, err
	}
	if needUninit {
		defer procCoUninit.Call()
	}

	switch opts.Mode {
	case DialogSave:
		dialog, err := createSaveDialog()
		if err != nil {
			return Result{}, err
		}
		defer dialog.release()
		if err := configureSaveDialog(dialog, opts); err != nil {
			return Result{}, err
		}
		if canceled, err := dialog.show(owner); err != nil || canceled {
			return Result{}, err
		}
		item, err := dialog.getResult()
		if err != nil {
			return Result{}, err
		}
		defer item.release()
		path, err := item.fileSystemPath()
		if err != nil {
			return Result{}, err
		}
		if path == "" {
			return Result{}, nil
		}
		return Result{Paths: []string{path}}, nil
	default:
		dialog, err := createOpenDialog()
		if err != nil {
			return Result{}, err
		}
		defer dialog.release()
		if err := configureOpenDialog(dialog, opts); err != nil {
			return Result{}, err
		}
		if canceled, err := dialog.show(owner); err != nil || canceled {
			return Result{}, err
		}
		if opts.MultiSelect {
			items, err := dialog.getResults()
			if err != nil {
				return Result{}, err
			}
			defer items.release()
			paths, err := items.fileSystemPaths()
			if err != nil {
				return Result{}, err
			}
			return Result{Paths: paths}, nil
		}
		item, err := dialog.getResult()
		if err != nil {
			return Result{}, err
		}
		defer item.release()
		path, err := item.fileSystemPath()
		if err != nil {
			return Result{}, err
		}
		if path == "" {
			return Result{}, nil
		}
		return Result{Paths: []string{path}}, nil
	}
}

func configureOpenDialog(dialog *fileOpenDialog, opts Options) error {
	if err := dialog.setFileTypes(opts.Filters); err != nil {
		return err
	}
	if err := dialog.setOptions(openDialogFlags(opts)); err != nil {
		return err
	}
	if err := dialog.setTitle(opts.Title); err != nil {
		return err
	}
	if err := dialog.setButtonLabel(opts.ButtonLabel); err != nil {
		return err
	}
	return dialog.setInitialPath(opts.InitialPath, opts.Mode == DialogFolder)
}

func configureSaveDialog(dialog *fileSaveDialog, opts Options) error {
	if err := dialog.setFileTypes(opts.Filters); err != nil {
		return err
	}
	if err := dialog.setOptions(saveDialogFlags(opts)); err != nil {
		return err
	}
	if err := dialog.setTitle(opts.Title); err != nil {
		return err
	}
	if err := dialog.setButtonLabel(opts.ButtonLabel); err != nil {
		return err
	}
	if err := dialog.setDefaultExtension(opts.DefaultExtension); err != nil {
		return err
	}
	return dialog.setInitialPath(opts.InitialPath)
}

func openDialogFlags(opts Options) uint32 {
	flags := uint32(fosNoChangeDir | fosForceFilesystem)
	if opts.Mode == DialogFolder {
		flags |= fosPickFolders
	}
	if opts.MultiSelect {
		flags |= fosAllowMulti
	}
	if opts.PathMustExist || opts.Mode != DialogSave {
		flags |= fosPathMustExist
	}
	if opts.FileMustExist || opts.Mode == DialogOpen {
		flags |= fosFileMustExist
	}
	return flags
}

func saveDialogFlags(opts Options) uint32 {
	flags := uint32(fosNoChangeDir | fosForceFilesystem | fosPathMustExist)
	if opts.OverwritePrompt || !opts.CreatePrompt {
		flags |= fosOverwritePrompt
	}
	if opts.CreatePrompt {
		flags |= fosCreatePrompt
	}
	return flags
}

func coInitialize() (bool, error) {
	hr, _, _ := procCoInit.Call(0, coInitApartmentThreaded)
	switch uint32(hr) {
	case 0, 1:
		return true, nil
	case hResultChangedMode:
		return false, nil
	default:
		if failed(hr) {
			return false, fmt.Errorf("CoInitializeEx failed: 0x%08X", uint32(hr))
		}
		return false, nil
	}
}

func createOpenDialog() (*fileOpenDialog, error) {
	var dialog *fileOpenDialog
	hr, _, _ := procCreate.Call(
		uintptr(unsafe.Pointer(&clsidFileOpenDialog)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(&iidFileOpenDialog)),
		uintptr(unsafe.Pointer(&dialog)),
	)
	if failed(hr) || dialog == nil {
		return nil, fmt.Errorf("CoCreateInstance(IFileOpenDialog) failed: 0x%08X", uint32(hr))
	}
	return dialog, nil
}

func createSaveDialog() (*fileSaveDialog, error) {
	var dialog *fileSaveDialog
	hr, _, _ := procCreate.Call(
		uintptr(unsafe.Pointer(&clsidFileSaveDialog)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(&iidFileSaveDialog)),
		uintptr(unsafe.Pointer(&dialog)),
	)
	if failed(hr) || dialog == nil {
		return nil, fmt.Errorf("CoCreateInstance(IFileSaveDialog) failed: 0x%08X", uint32(hr))
	}
	return dialog, nil
}

func (d *fileOpenDialog) release() {
	if d == nil || d.lpVtbl == nil || d.lpVtbl.Release == 0 {
		return
	}
	syscall.SyscallN(d.lpVtbl.Release, uintptr(unsafe.Pointer(d)))
}

func (d *fileSaveDialog) release() {
	if d == nil || d.lpVtbl == nil || d.lpVtbl.Release == 0 {
		return
	}
	syscall.SyscallN(d.lpVtbl.Release, uintptr(unsafe.Pointer(d)))
}

func (d *fileOpenDialog) show(owner windows.Handle) (bool, error) {
	hr, _, _ := syscall.SyscallN(d.lpVtbl.Show, uintptr(unsafe.Pointer(d)), uintptr(owner))
	switch uint32(hr) {
	case 0:
		return false, nil
	case hResultCanceled:
		return true, nil
	default:
		if failed(hr) {
			return false, fmt.Errorf("IFileOpenDialog.Show failed: 0x%08X", uint32(hr))
		}
		return false, nil
	}
}

func (d *fileSaveDialog) show(owner windows.Handle) (bool, error) {
	hr, _, _ := syscall.SyscallN(d.lpVtbl.Show, uintptr(unsafe.Pointer(d)), uintptr(owner))
	switch uint32(hr) {
	case 0:
		return false, nil
	case hResultCanceled:
		return true, nil
	default:
		if failed(hr) {
			return false, fmt.Errorf("IFileSaveDialog.Show failed: 0x%08X", uint32(hr))
		}
		return false, nil
	}
}

func (d *fileOpenDialog) setFileTypes(filters []FileFilter) error {
	return setDialogFileTypes(
		uintptr(unsafe.Pointer(d)),
		d.lpVtbl.SetFileTypes,
		d.lpVtbl.SetFileTypeIndex,
		filters,
		"IFileOpenDialog",
	)
}

func (d *fileSaveDialog) setFileTypes(filters []FileFilter) error {
	return setDialogFileTypes(
		uintptr(unsafe.Pointer(d)),
		d.lpVtbl.SetFileTypes,
		d.lpVtbl.SetFileTypeIndex,
		filters,
		"IFileSaveDialog",
	)
}

func setDialogFileTypes(self uintptr, setFileTypes uintptr, setFileTypeIndex uintptr, filters []FileFilter, label string) error {
	if self == 0 || setFileTypes == 0 {
		return fmt.Errorf("%s is nil", label)
	}
	if len(filters) == 0 {
		return nil
	}

	specs := make([]dialogFilterSpec, 0, len(filters))
	for _, filter := range filters {
		name := strings.TrimSpace(filter.Name)
		if name == "" {
			name = "Files"
		}
		pattern := strings.TrimSpace(filter.Pattern)
		if pattern == "" {
			pattern = "*.*"
		}
		namePtr, err := windows.UTF16PtrFromString(name)
		if err != nil {
			return err
		}
		specPtr, err := windows.UTF16PtrFromString(pattern)
		if err != nil {
			return err
		}
		specs = append(specs, dialogFilterSpec{Name: namePtr, Spec: specPtr})
	}

	hr, _, _ := syscall.SyscallN(setFileTypes, self, uintptr(len(specs)), uintptr(unsafe.Pointer(&specs[0])))
	if failed(hr) {
		return fmt.Errorf("%s.SetFileTypes failed: 0x%08X", label, uint32(hr))
	}
	hr, _, _ = syscall.SyscallN(setFileTypeIndex, self, 1)
	if failed(hr) {
		return fmt.Errorf("%s.SetFileTypeIndex failed: 0x%08X", label, uint32(hr))
	}
	return nil
}

func (d *fileOpenDialog) setOptions(flags uint32) error {
	return setDialogOptions(
		uintptr(unsafe.Pointer(d)),
		d.lpVtbl.GetOptions,
		d.lpVtbl.SetOptions,
		flags,
		"IFileOpenDialog",
	)
}

func (d *fileSaveDialog) setOptions(flags uint32) error {
	return setDialogOptions(
		uintptr(unsafe.Pointer(d)),
		d.lpVtbl.GetOptions,
		d.lpVtbl.SetOptions,
		flags,
		"IFileSaveDialog",
	)
}

func setDialogOptions(self uintptr, getOptions uintptr, setOptions uintptr, flags uint32, label string) error {
	if self == 0 {
		return fmt.Errorf("%s is nil", label)
	}
	var current uint32
	hr, _, _ := syscall.SyscallN(getOptions, self, uintptr(unsafe.Pointer(&current)))
	if failed(hr) {
		return fmt.Errorf("%s.GetOptions failed: 0x%08X", label, uint32(hr))
	}
	hr, _, _ = syscall.SyscallN(setOptions, self, uintptr(current|flags))
	if failed(hr) {
		return fmt.Errorf("%s.SetOptions failed: 0x%08X", label, uint32(hr))
	}
	return nil
}

func (d *fileOpenDialog) setTitle(title string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetTitle, title, "IFileOpenDialog.SetTitle")
}

func (d *fileSaveDialog) setTitle(title string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetTitle, title, "IFileSaveDialog.SetTitle")
}

func (d *fileOpenDialog) setButtonLabel(text string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetOkButtonLabel, text, "IFileOpenDialog.SetOkButtonLabel")
}

func (d *fileSaveDialog) setButtonLabel(text string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetOkButtonLabel, text, "IFileSaveDialog.SetOkButtonLabel")
}

func (d *fileOpenDialog) setFileName(name string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetFileName, name, "IFileOpenDialog.SetFileName")
}

func (d *fileSaveDialog) setFileName(name string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetFileName, name, "IFileSaveDialog.SetFileName")
}

func (d *fileSaveDialog) setDefaultExtension(ext string) error {
	return setDialogText(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetDefaultExtension, normalizeExtension(ext), "IFileSaveDialog.SetDefaultExtension")
}

func setDialogText(self uintptr, proc uintptr, text string, label string) error {
	text = strings.TrimSpace(text)
	if self == 0 || proc == 0 || text == "" {
		return nil
	}
	ptr, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return err
	}
	hr, _, _ := syscall.SyscallN(proc, self, uintptr(unsafe.Pointer(ptr)))
	if failed(hr) {
		return fmt.Errorf("%s failed: 0x%08X", label, uint32(hr))
	}
	return nil
}

func (d *fileOpenDialog) setFolder(path string) error {
	return setDialogFolder(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetFolder, path, "IFileOpenDialog")
}

func (d *fileSaveDialog) setFolder(path string) error {
	return setDialogFolder(uintptr(unsafe.Pointer(d)), d.lpVtbl.SetFolder, path, "IFileSaveDialog")
}

func setDialogFolder(self uintptr, proc uintptr, path string, label string) error {
	if self == 0 {
		return fmt.Errorf("%s is nil", label)
	}
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return nil
	}
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var folder *shellItem
	hr, _, _ := procSHCreate.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&iidShellItem)),
		uintptr(unsafe.Pointer(&folder)),
	)
	if failed(hr) || folder == nil {
		return fmt.Errorf("SHCreateItemFromParsingName failed: 0x%08X", uint32(hr))
	}
	defer folder.release()
	hr, _, _ = syscall.SyscallN(proc, self, uintptr(unsafe.Pointer(folder)))
	if failed(hr) {
		return fmt.Errorf("%s.SetFolder failed: 0x%08X", label, uint32(hr))
	}
	return nil
}

func (d *fileOpenDialog) setInitialPath(initial string, folderOnly bool) error {
	folder, name := splitInitialPath(initial, folderOnly)
	if folder != "" {
		if err := d.setFolder(folder); err != nil {
			return err
		}
	}
	if !folderOnly && name != "" {
		if err := d.setFileName(name); err != nil {
			return err
		}
	}
	return nil
}

func (d *fileSaveDialog) setInitialPath(initial string) error {
	folder, name := splitInitialPath(initial, false)
	if folder != "" {
		if err := d.setFolder(folder); err != nil {
			return err
		}
	}
	if name != "" {
		if err := d.setFileName(name); err != nil {
			return err
		}
	}
	return nil
}

func splitInitialPath(initial string, folderOnly bool) (folder string, name string) {
	initial = strings.TrimSpace(initial)
	if initial == "" {
		return "", ""
	}
	if info, err := os.Stat(initial); err == nil && info.IsDir() {
		return initial, ""
	}
	if folderOnly {
		return filepath.Dir(initial), ""
	}
	dir := filepath.Dir(initial)
	base := filepath.Base(initial)
	if base == "." || base == string(filepath.Separator) {
		return initial, ""
	}
	if dir == "." && !strings.ContainsAny(initial, `/\`) {
		return "", base
	}
	return dir, base
}

func (d *fileOpenDialog) getResult() (*shellItem, error) {
	return getDialogResult(uintptr(unsafe.Pointer(d)), d.lpVtbl.GetResult, "IFileOpenDialog.GetResult")
}

func (d *fileSaveDialog) getResult() (*shellItem, error) {
	return getDialogResult(uintptr(unsafe.Pointer(d)), d.lpVtbl.GetResult, "IFileSaveDialog.GetResult")
}

func getDialogResult(self uintptr, proc uintptr, label string) (*shellItem, error) {
	var item *shellItem
	hr, _, _ := syscall.SyscallN(proc, self, uintptr(unsafe.Pointer(&item)))
	if failed(hr) || item == nil {
		return nil, fmt.Errorf("%s failed: 0x%08X", label, uint32(hr))
	}
	return item, nil
}

func (d *fileOpenDialog) getResults() (*shellItemArray, error) {
	var items *shellItemArray
	hr, _, _ := syscall.SyscallN(d.lpVtbl.GetResults, uintptr(unsafe.Pointer(d)), uintptr(unsafe.Pointer(&items)))
	if failed(hr) || items == nil {
		return nil, fmt.Errorf("IFileOpenDialog.GetResults failed: 0x%08X", uint32(hr))
	}
	return items, nil
}

func (s *shellItem) release() {
	if s == nil || s.lpVtbl == nil || s.lpVtbl.Release == 0 {
		return
	}
	syscall.SyscallN(s.lpVtbl.Release, uintptr(unsafe.Pointer(s)))
}

func (s *shellItem) fileSystemPath() (string, error) {
	if s == nil {
		return "", fmt.Errorf("IShellItem is nil")
	}
	var rawPath *uint16
	hr, _, _ := syscall.SyscallN(
		s.lpVtbl.GetDisplayName,
		uintptr(unsafe.Pointer(s)),
		sigdnFileSysPath,
		uintptr(unsafe.Pointer(&rawPath)),
	)
	if failed(hr) {
		return "", fmt.Errorf("IShellItem.GetDisplayName failed: 0x%08X", uint32(hr))
	}
	if rawPath == nil {
		return "", nil
	}
	defer procFreeMem.Call(uintptr(unsafe.Pointer(rawPath)))
	return strings.TrimSpace(windows.UTF16PtrToString(rawPath)), nil
}

func (a *shellItemArray) release() {
	if a == nil || a.lpVtbl == nil || a.lpVtbl.Release == 0 {
		return
	}
	syscall.SyscallN(a.lpVtbl.Release, uintptr(unsafe.Pointer(a)))
}

func (a *shellItemArray) count() (uint32, error) {
	var count uint32
	hr, _, _ := syscall.SyscallN(a.lpVtbl.GetCount, uintptr(unsafe.Pointer(a)), uintptr(unsafe.Pointer(&count)))
	if failed(hr) {
		return 0, fmt.Errorf("IShellItemArray.GetCount failed: 0x%08X", uint32(hr))
	}
	return count, nil
}

func (a *shellItemArray) itemAt(index uint32) (*shellItem, error) {
	var item *shellItem
	hr, _, _ := syscall.SyscallN(
		a.lpVtbl.GetItemAt,
		uintptr(unsafe.Pointer(a)),
		uintptr(index),
		uintptr(unsafe.Pointer(&item)),
	)
	if failed(hr) || item == nil {
		return nil, fmt.Errorf("IShellItemArray.GetItemAt failed: 0x%08X", uint32(hr))
	}
	return item, nil
}

func (a *shellItemArray) fileSystemPaths() ([]string, error) {
	if a == nil {
		return nil, nil
	}
	count, err := a.count()
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, count)
	for i := uint32(0); i < count; i++ {
		item, err := a.itemAt(i)
		if err != nil {
			return nil, err
		}
		path, pathErr := item.fileSystemPath()
		item.release()
		if pathErr != nil {
			return nil, pathErr
		}
		if path != "" {
			paths = append(paths, path)
		}
	}
	return paths, nil
}

func normalizeExtension(ext string) string {
	ext = strings.TrimSpace(ext)
	ext = strings.TrimPrefix(ext, ".")
	return ext
}

func failed(hr uintptr) bool {
	return int32(uint32(hr)) < 0
}

//go:build windows

package core

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	ErrNotInitialized = errors.New("winui/core: app not initialized")
	ErrAppClosed      = errors.New("winui/core: app closed")
	ErrTimerIDZero    = errors.New("winui/core: timer id must not be zero")
)

// opError 为底层系统调用错误补充操作上下文。
type opError struct {
	// Op 表示失败的操作名称。
	Op string
	// Err 表示底层错误。
	Err error
}

// Error 返回带操作上下文的错误文本。
func (e *opError) Error() string {
	return fmt.Sprintf("winui/core: %s: %v", e.Op, e.Err)
}

// Unwrap 返回底层错误。
func (e *opError) Unwrap() error {
	return e.Err
}

// wrapError 为失败操作补充来源调用信息。
func wrapError(op string, err error) error {
	if err == nil {
		err = syscall.EINVAL
	}
	return &opError{
		Op:  op,
		Err: normalizeSyscallError(err),
	}
}

// wrapHRESULT 将 HRESULT 风格结果转换为带操作上下文的 Go 错误。
func wrapHRESULT(op string, hr uintptr) error {
	return &opError{
		Op:  op,
		Err: fmt.Errorf("HRESULT 0x%08X", uint32(hr)),
	}
}

// normalizeSyscallError 将零值系统调用错误归一化为 nil。
func normalizeSyscallError(err error) error {
	if err == nil {
		return syscall.EINVAL
	}
	if errno, ok := err.(windows.Errno); ok && errno == windows.ERROR_SUCCESS {
		return syscall.EINVAL
	}
	if errno, ok := err.(syscall.Errno); ok && errno == 0 {
		return syscall.EINVAL
	}
	return err
}

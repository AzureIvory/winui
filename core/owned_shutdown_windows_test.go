//go:build windows

package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// A subprocess prevents a failed native shutdown from leaving UI threads in the
// test host. Every HWND belongs to this hidden fixture; no desktop input is sent.
func TestOwnedWindowShutdown(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"no-popup", "ready-popup", "init-popup"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestOwnedWindowShutdownHelper$", "-test.v", "-test.timeout=6s")
			cmd.Env = append(os.Environ(), "WINUI_OWNED_SHUTDOWN="+scenario)
			output, err := cmd.CombinedOutput()
			t.Log(string(output))
			if err != nil {
				t.Fatalf("hidden owner-first shutdown failed: %v", err)
			}
		})
	}
}

func TestOwnedWindowShutdownHelper(t *testing.T) {
	scenario := os.Getenv("WINUI_OWNED_SHUTDOWN")
	if scenario == "" {
		t.Skip("subprocess helper")
	}
	var rootHandle atomic.Uintptr
	rootClosing := make(chan struct{})
	var closeOnce sync.Once
	globalWndProc = windows.NewCallback(func(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
		if message == wmClose && hwnd == rootHandle.Load() {
			fmt.Printf("root WM_CLOSE tid=%d\n", currentThreadID())
			closeOnce.Do(func() { close(rootClosing) })
		}
		return appWndProc(hwnd, message, wParam, lParam)
	})
	var popup atomic.Pointer[App]
	var rootDestroyed, popupDestroyed atomic.Bool
	root, err := NewApp(Options{StartHidden: true, OnDestroy: func(app *App) {
		fmt.Printf("root destroy tid=%d ui=%v\n", currentThreadID(), app.IsUIThread())
		rootDestroyed.Store(true)
		if !app.IsUIThread() {
			t.Error("owner destroyed outside its UI thread")
		}
		if child := popup.Load(); child != nil {
			child.Close()
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := root.Init(); err != nil {
		t.Fatal(err)
	}
	rootHandle.Store(uintptr(root.Handle()))
	owned := make(chan struct{})
	allowInit := make(chan struct{})
	childInit := make(chan error, 1)
	if scenario != "no-popup" {
		child, err := NewApp(Options{StartHidden: true, OnCreate: func(app *App) error {
			fmt.Printf("popup create tid=%d ui=%v\n", currentThreadID(), app.IsUIThread())
			if err := app.SetOwner(root.Handle()); err != nil {
				return err
			}
			close(owned)
			if scenario == "init-popup" {
				<-allowInit
			}
			return nil
		}, OnDestroy: func(app *App) {
			fmt.Printf("popup destroy tid=%d ui=%v\n", currentThreadID(), app.IsUIThread())
			popupDestroyed.Store(true)
			if !app.IsUIThread() {
				t.Error("popup destroyed outside its UI thread")
			}
		}})
		if err != nil {
			t.Fatal(err)
		}
		popup.Store(child)
		go func() { childInit <- child.Init() }()
		<-owned
		if scenario == "ready-popup" {
			if err := <-childInit; err != nil {
				t.Fatal(err)
			}
			// The production owner can synchronously arrange an owned window
			// from its own UI callback before the eventual main-window close.
			arranged := make(chan error, 1)
			if err := root.Post(func() { arranged <- child.SetScreenBounds(Rect{X: 10, Y: 10, W: 320, H: 200}) }); err != nil {
				t.Fatal(err)
			}
			if err := <-arranged; err != nil {
				t.Fatal(err)
			}
			fmt.Println("popup arranged from root UI")
		}
	}
	if ok, _, err := procPostMessageW.Call(uintptr(root.Handle()), wmClose, 0, 0); ok == 0 {
		t.Fatal(err)
	}
	<-rootClosing
	if scenario == "init-popup" {
		close(allowInit)
		if err := <-childInit; err != nil {
			t.Fatal(err)
		}
	}
	if code := root.Run(); code != 0 {
		t.Fatalf("root exit=%d", code)
	}
	fmt.Println("root Run exited")
	if !rootDestroyed.Load() {
		t.Error("owner message loop exited without destruction")
	}
	if child := popup.Load(); child != nil {
		if code := child.Run(); code != 0 {
			t.Fatalf("popup exit=%d", code)
		}
		fmt.Println("popup Run exited")
		if !popupDestroyed.Load() {
			t.Error("popup message loop exited without destruction")
		}
	}
}

//go:build windows

package core

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPostBeforeInitRejectsCallbackWithoutStartingWindow(t *testing.T) {
	app, err := NewApp(Options{StartHidden: true})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	for _, candidate := range []*App{nil, app} {
		if err := candidate.Post(func() { called = true }); !errors.Is(err, ErrNotInitialized) {
			t.Errorf("uninitialized Post returned %v", err)
		}
		if err := candidate.Post(nil); !errors.Is(err, ErrNotInitialized) {
			t.Errorf("uninitialized nil Post returned %v", err)
		}
		if candidate.Handle() != 0 || candidate.IsUIThread() {
			t.Error("uninitialized app exposed window/thread identity")
		}
	}
	app.drainPosted()
	if called || len(app.postQueue) != 0 {
		t.Fatal("rejected callback was retained or executed")
	}
}

func TestAppIdentityAndPostsAreSafeDuringWindowCreation(t *testing.T) {
	creating := make(chan struct{})
	release := make(chan struct{})
	var entered, released sync.Once
	app, err := NewApp(Options{StartHidden: true, OnResize: func(*App, Size) {
		// WM_SIZE can run inside CreateWindowExW, before Init is ready. This
		// preserves that real creation phase while background callers Post.
		entered.Do(func() { close(creating); <-release })
	}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		released.Do(func() { close(release) })
		app.Close()
		app.Run()
	})
	start := make(chan struct{})
	initResult := make(chan error, 1)
	go func() { <-start; initResult <- app.Init() }()
	var workers sync.WaitGroup
	var callbacks atomic.Int32
	const workerCount, postsPerWorker = 4, 64
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			// A fixed concurrent workload, not a readiness poll or Post retry.
			for range 256 {
				_ = app.Handle()
				if app.IsUIThread() {
					t.Error("background caller was identified as the UI thread")
				}
			}
			<-creating
			for range postsPerWorker {
				if app.Handle() == 0 || app.IsUIThread() {
					t.Error("creation callback did not publish window/thread identity")
				}
				if err := app.Post(func() {
					if !app.IsUIThread() || app.Handle() == 0 {
						t.Error("posted callback ran outside its UI window thread")
					}
					callbacks.Add(1)
				}); err != nil {
					t.Errorf("Post after creation publication failed: %v", err)
				}
			}
		}()
	}
	close(start)
	select {
	case <-creating:
	case err := <-initResult:
		t.Fatalf("Init returned before the creation barrier: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("native creation callback did not start")
	}
	workers.Wait()
	if callbacks.Load() != 0 {
		t.Error("creation callbacks ran while the UI thread was held in native creation")
	}
	released.Do(func() { close(release) })
	if err := <-initResult; err != nil {
		t.Fatal(err)
	}
	drained := make(chan struct{})
	if err := app.Post(func() { close(drained) }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-drained:
	case <-time.After(3 * time.Second):
		t.Fatal("posted callbacks did not drain")
	}
	if got := callbacks.Load(); got != workerCount*postsPerWorker {
		t.Fatalf("creation callbacks ran %d times; want %d", got, workerCount*postsPerWorker)
	}
}

func TestOnCreateCanPostInlineAndCloseWithoutWaitingForInit(t *testing.T) {
	var calls atomic.Int32
	app, err := NewApp(Options{StartHidden: true, OnCreate: func(app *App) error {
		if app.Handle() == 0 || !app.IsUIThread() {
			t.Error("OnCreate did not receive its window/thread identity")
		}
		return app.Post(func() { calls.Add(1); app.Close() })
	}})
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- app.Init() }()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("OnCreate Post/Close waited for Init and deadlocked")
	}
	if exitCode := app.Run(); exitCode != 0 || calls.Load() != 1 {
		t.Fatalf("creation Post/Close exit=%d callbacks=%d", exitCode, calls.Load())
	}
	if err := app.Post(func() { t.Error("callback executed after close") }); !errors.Is(err, ErrAppClosed) {
		t.Fatalf("closed Post returned %v", err)
	}
}

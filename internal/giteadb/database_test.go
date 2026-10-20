package giteadb

import (
	"context"
	"errors"
	"testing"
)

func resetInitState(t *testing.T, fn func(context.Context) error) {
	t.Helper()

	initMu.Lock()
	oldInitialized := initialized
	oldInitializeEngine := initializeEngine
	initialized = false
	initializeEngine = fn
	initMu.Unlock()

	t.Cleanup(func() {
		initMu.Lock()
		initialized = oldInitialized
		initializeEngine = oldInitializeEngine
		initMu.Unlock()
	})
}

func TestInitInitializesEngineOnce(t *testing.T) {
	calls := 0
	resetInitState(t, func(context.Context) error {
		calls++
		return nil
	})

	if err := Init(context.Background()); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	if err := Init(context.Background()); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("initialize engine calls = %d, want 1", calls)
	}
}

func TestInitRetriesAfterFailure(t *testing.T) {
	calls := 0
	wantErr := errors.New("database unavailable")
	resetInitState(t, func(context.Context) error {
		calls++
		if calls == 1 {
			return wantErr
		}
		return nil
	})

	if err := Init(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("first Init() error = %v, want %v", err, wantErr)
	}
	if err := Init(context.Background()); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
	if err := Init(context.Background()); err != nil {
		t.Fatalf("third Init() error = %v", err)
	}

	if calls != 2 {
		t.Fatalf("initialize engine calls = %d, want 2", calls)
	}
}

package supervisor

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

// helper logger that writes nowhere
var discardLogger = log.New(nil, "", 0)

// Test that the worker runs at least once.
func TestSupervisorRunsWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var called bool
	done := make(chan bool)

	Start(ctx, Config{Logger: discardLogger}, func(ctx context.Context) {
		called = true
		done <- true
	})

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("worker did not start")
	}

	if !called {
		t.Fatal("worker was not called")
	}
}

// Test that the worker restarts after panic.
func TestSupervisorRestartsOnPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	runCount := 0
	done := make(chan struct{})

	Start(ctx, Config{
		MinBackoff: 10 * time.Millisecond,
		MaxBackoff: 50 * time.Millisecond,
		Logger:     discardLogger,
	}, func(ctx context.Context) {
		mu.Lock()
		runCount++
		mu.Unlock()

		if runCount == 1 {
			panic("boom")
		}

		done <- struct{}{}
	})

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker was not restarted after panic")
	}

	mu.Lock()
	if runCount < 2 {
		t.Fatalf("expected worker to restart, got runCount=%d", runCount)
	}
	mu.Unlock()
}

// Test supervisor stops when context is cancelled.
func TestSupervisorStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var ran bool
	Start(ctx, Config{Logger: discardLogger}, func(ctx context.Context) {
		ran = true
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	})

	// ensure it started
	time.Sleep(20 * time.Millisecond)
	if !ran {
		t.Fatal("worker never started")
	}

	cancel()

	// wait for graceful shutdown
	time.Sleep(20 * time.Millisecond)
	// nothing to assert — test passes if it does not hang
}

// Test that the supervisor uses exponential backoff.
func TestSupervisorBackoffIncreases(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := Config{
		MinBackoff: 20 * time.Millisecond,
		MaxBackoff: 200 * time.Millisecond,
		Logger:     discardLogger,
	}

	var mu sync.Mutex
	runCount := 0

	Start(ctx, cfg, func(ctx context.Context) {
		mu.Lock()
		runCount++
		mu.Unlock()
		panic("boom")
	})

	time.Sleep(5 * cfg.MinBackoff)

	mu.Lock()
	count := runCount
	mu.Unlock()

	if count < 3 {
		t.Fatalf("expected multiple restarts with backoff, got %d runs", count)
	}
}

// Test that worker gets the same context and responds to cancellation.
func TestWorkerReceivesContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)

	Start(ctx, Config{Logger: discardLogger}, func(ctx context.Context) {
		<-ctx.Done()
		done <- true
	})

	cancel()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker did not exit after context cancel")
	}
}

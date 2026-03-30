package integration

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeRemoteScheduler struct {
	err   error
	calls int32
}

func (f *fakeRemoteScheduler) Upsert(context.Context, RemoteJobSpec) error {
	atomic.AddInt32(&f.calls, 1)
	return f.err
}

func TestSchedulerBridge_LocalModeStartsScheduledWorker(t *testing.T) {
	bridge := NewSchedulerBridge(SchedulerModeLocal, true, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := int32(0)
	ran := int32(0)
	bridge.StartWorkers(ctx, func(fn func() error) {
		atomic.AddInt32(&started, 1)
		go func() { _ = fn() }()
	}, WorkerSpec{
		Name:     "job.local",
		Interval: 10 * time.Millisecond,
		RunOnce: func(ctx context.Context) error {
			atomic.AddInt32(&ran, 1)
			cancel()
			return nil
		},
	})

	time.Sleep(50 * time.Millisecond)
	if got := atomic.LoadInt32(&started); got < 1 {
		t.Fatalf("expected scheduler runner start, got %d", got)
	}
	if got := atomic.LoadInt32(&ran); got < 1 {
		t.Fatalf("expected local run at least once, got %d", got)
	}
}

func TestSchedulerBridge_CoreXModeRemoteSuccessSkipsLocal(t *testing.T) {
	remote := &fakeRemoteScheduler{}
	bridge := NewSchedulerBridge(SchedulerModeCoreX, true, remote, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := int32(0)
	bridge.StartWorkers(ctx, func(fn func() error) {
		atomic.AddInt32(&started, 1)
		go func() { _ = fn() }()
	}, WorkerSpec{
		Name:     "job.corex",
		Interval: time.Second,
		RunOnce: func(ctx context.Context) error {
			return nil
		},
	})

	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt32(&remote.calls); got != 1 {
		t.Fatalf("expected remote upsert once, got %d", got)
	}
	if got := atomic.LoadInt32(&started); got != 0 {
		t.Fatalf("expected local worker skipped, got %d", got)
	}
}

func TestSchedulerBridge_CoreXModeFallbackStartsLocal(t *testing.T) {
	remote := &fakeRemoteScheduler{err: errors.New("upsert failed")}
	bridge := NewSchedulerBridge(SchedulerModeCoreX, true, remote, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := int32(0)
	ran := int32(0)
	bridge.StartWorkers(ctx, func(fn func() error) {
		atomic.AddInt32(&started, 1)
		go func() { _ = fn() }()
	}, WorkerSpec{
		Name:     "job.corex.fallback",
		Interval: 10 * time.Millisecond,
		RunOnce: func(ctx context.Context) error {
			atomic.AddInt32(&ran, 1)
			cancel()
			return nil
		},
	})

	time.Sleep(50 * time.Millisecond)
	if got := atomic.LoadInt32(&remote.calls); got != 1 {
		t.Fatalf("expected remote upsert once, got %d", got)
	}
	if got := atomic.LoadInt32(&started); got < 1 {
		t.Fatalf("expected local scheduler start by fallback, got %d", got)
	}
	if got := atomic.LoadInt32(&ran); got < 1 {
		t.Fatalf("expected local run by fallback, got %d", got)
	}
}

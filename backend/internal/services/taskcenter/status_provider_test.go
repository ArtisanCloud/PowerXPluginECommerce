package taskcenter

import (
	"context"
	"errors"
	"testing"
)

type fakeStatusProvider struct {
	job *JobStatus
	err error
}

func (p *fakeStatusProvider) Get(_ context.Context, _ string, _ string) (*JobStatus, error) {
	if p == nil {
		return nil, ErrTaskStatusNotAvailable
	}
	return p.job, p.err
}

func TestChainStatusProvider_FrameworkUnavailableFallbackLocal(t *testing.T) {
	local := &JobStatus{TaskID: "task-local", Status: "running"}
	chain := &ChainStatusProvider{Providers: []StatusProvider{
		&fakeStatusProvider{err: ErrTaskStatusNotAvailable},
		&fakeStatusProvider{job: local},
	}}

	job, err := chain.Get(context.Background(), "task-local", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil || job.TaskID != "task-local" {
		t.Fatalf("unexpected job: %#v", job)
	}
}

func TestChainStatusProvider_ReturnsLastNonAvailabilityError(t *testing.T) {
	expectedErr := errors.New("upstream failed")
	chain := &ChainStatusProvider{Providers: []StatusProvider{
		&fakeStatusProvider{err: expectedErr},
		&fakeStatusProvider{err: ErrTaskStatusNotAvailable},
	}}

	_, err := chain.Get(context.Background(), "task-1", "")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

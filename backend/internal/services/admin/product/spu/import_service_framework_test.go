package spu

import (
	"context"
	"errors"
	"testing"

	taskcenter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
)

type fakeTaskSubmitter struct {
	taskID string
	err    error
}

func (f *fakeTaskSubmitter) Submit(_ context.Context, _ taskcenter.TaskSubmitRequest) (*taskcenter.TaskSubmitResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &taskcenter.TaskSubmitResult{TaskID: f.taskID}, nil
}

type fakeTaskReporter struct {
	updates []taskcenter.TaskUpdateRequest
	err     error
}

func (f *fakeTaskReporter) Update(_ context.Context, req taskcenter.TaskUpdateRequest) error {
	f.updates = append(f.updates, req)
	return f.err
}

func TestImportServiceSubmitFrameworkTask(t *testing.T) {
	svc := &ImportService{taskSubmitter: &fakeTaskSubmitter{taskID: "task-101"}}
	taskID, err := svc.submitFrameworkTask(context.Background(), importJobType, "tenant-1", map[string]any{"filename": "demo.csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taskID != "task-101" {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}

func TestImportServiceSubmitFrameworkTaskReturnsError(t *testing.T) {
	svc := &ImportService{taskSubmitter: &fakeTaskSubmitter{err: errors.New("boom")}}
	_, err := svc.submitFrameworkTask(context.Background(), importJobType, "tenant-1", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestImportServiceReportFrameworkTask(t *testing.T) {
	reporter := &fakeTaskReporter{}
	svc := &ImportService{taskReporter: reporter}
	svc.reportFrameworkTask(context.Background(), "task-202", "running", "processing", "", nil)
	if len(reporter.updates) != 1 {
		t.Fatalf("unexpected updates: %d", len(reporter.updates))
	}
	if reporter.updates[0].TaskID != "task-202" || reporter.updates[0].Status != "running" {
		t.Fatalf("unexpected update payload: %#v", reporter.updates[0])
	}
}

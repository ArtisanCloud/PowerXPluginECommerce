package customer

import (
	"context"
	"errors"
	"testing"

	taskcenter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	taskbusx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/taskbus"
)

type fakeCustomerTaskSubmitter struct {
	taskID string
	err    error
}

func (f *fakeCustomerTaskSubmitter) Submit(_ context.Context, _ taskcenter.TaskSubmitRequest) (*taskcenter.TaskSubmitResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &taskcenter.TaskSubmitResult{TaskID: f.taskID}, nil
}

type fakeCustomerTaskReporter struct {
	updates []taskcenter.TaskUpdateRequest
}

func (f *fakeCustomerTaskReporter) Update(_ context.Context, req taskcenter.TaskUpdateRequest) error {
	f.updates = append(f.updates, req)
	return nil
}

func TestServiceSubmitFrameworkTask(t *testing.T) {
	svc := &Service{taskSubmitter: &fakeCustomerTaskSubmitter{taskID: "task-c-1"}}
	taskID, err := svc.submitFrameworkTask(context.Background(), importJobType, "tenant-1", map[string]any{"filename": "a.csv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if taskID != "task-c-1" {
		t.Fatalf("unexpected task id: %s", taskID)
	}
}

func TestServiceSubmitFrameworkTaskError(t *testing.T) {
	svc := &Service{taskSubmitter: &fakeCustomerTaskSubmitter{err: errors.New("boom")}}
	_, err := svc.submitFrameworkTask(context.Background(), importJobType, "tenant-1", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestServicePublishTaskEvent(t *testing.T) {
	bus := taskbusx.NewLocalClient(nil)
	received := map[string]bool{}
	topics := []string{"task.progress", "powerx.task.progress.v1", "worker.task.updated"}
	for _, topic := range topics {
		topic := topic
		_ = bus.Subscribe(topic, func(_ context.Context, evt taskbusx.Event) error {
			received[topic] = true
			return nil
		})
	}
	svc := &Service{deps: &app.Deps{TaskBus: bus}}
	svc.publishTaskEvent(context.Background(), "task-c-2", "running", "processing", "", 35)
	for _, topic := range topics {
		if !received[topic] {
			t.Fatalf("topic not published: %s", topic)
		}
	}
}

func TestServiceReportFrameworkTask(t *testing.T) {
	reporter := &fakeCustomerTaskReporter{}
	svc := &Service{taskReporter: reporter}
	svc.reportFrameworkTask(context.Background(), "task-c-3", "success", "done", "", nil)
	if len(reporter.updates) != 1 {
		t.Fatalf("unexpected updates: %d", len(reporter.updates))
	}
	if reporter.updates[0].TaskID != "task-c-3" || reporter.updates[0].Status != "success" {
		t.Fatalf("unexpected update: %#v", reporter.updates[0])
	}
}

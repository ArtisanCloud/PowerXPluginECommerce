package taskcenter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrTaskStatusNotAvailable = errors.New("task status not available")
var ErrTaskNotFound = errors.New("task not found")

type StatusProvider interface {
	Get(ctx context.Context, taskID string, tenantUUID string) (*JobStatus, error)
}

type LocalStatusProvider struct {
	Store *Store
}

func (p *LocalStatusProvider) Get(_ context.Context, taskID string, tenantUUID string) (*JobStatus, error) {
	if p == nil || p.Store == nil {
		return nil, ErrTaskStatusNotAvailable
	}
	job, ok := p.Store.Get(strings.TrimSpace(taskID))
	if !ok || job == nil {
		return nil, ErrTaskNotFound
	}
	if tenantUUID != "" {
		if metaTenant, _ := job.Metadata["tenantUuid"].(string); metaTenant != "" && metaTenant != tenantUUID {
			return nil, fmt.Errorf("%w: task %s", ErrTaskNotFound, taskID)
		}
	}
	return job, nil
}

type ChainStatusProvider struct {
	Providers []StatusProvider
}

func (p *ChainStatusProvider) Get(ctx context.Context, taskID string, tenantUUID string) (*JobStatus, error) {
	if p == nil || len(p.Providers) == 0 {
		return nil, ErrTaskStatusNotAvailable
	}
	var lastErr error
	for _, provider := range p.Providers {
		if provider == nil {
			continue
		}
		job, err := provider.Get(ctx, taskID, tenantUUID)
		if err == nil && job != nil {
			return job, nil
		}
		if errors.Is(err, ErrTaskStatusNotAvailable) {
			continue
		}
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrTaskStatusNotAvailable
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 8 * time.Second}
}

package product_sku

import "context"

// SubmitBulkTask enqueues a price/inventory/channel adjustment payload.
func (s *Service) SubmitBulkTask(ctx context.Context, payload map[string]any) error {
	if err := s.HealthProbe(ctx); err != nil {
		return err
	}
	return s.NotImplemented()
}

// GetBulkTask provides a placeholder for reading task progress.
func (s *Service) GetBulkTask(ctx context.Context, taskID string) (map[string]any, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	return nil, s.NotImplemented()
}

package product_sku

import "context"

// UpsertChannelMapping persists channel metadata for a SKU.
func (s *Service) UpsertChannelMapping(ctx context.Context, skuID string, payload map[string]any) error {
	if err := s.HealthProbe(ctx); err != nil {
		return err
	}
	return s.NotImplemented()
}

// PublishChannelMapping triggers the asynchronous publish flow.
func (s *Service) PublishChannelMapping(ctx context.Context, skuID string, payload map[string]any) error {
	if err := s.HealthProbe(ctx); err != nil {
		return err
	}
	return s.NotImplemented()
}

package product_sku

import "context"

// UpsertSerialRecord records serial or batch level metadata for compliance.
func (s *Service) UpsertSerialRecord(ctx context.Context, skuID string, payload map[string]any) error {
	if err := s.HealthProbe(ctx); err != nil {
		return err
	}
	return s.NotImplemented()
}

// ListSerialRecords fetches historical serial numbers for a SKU.
func (s *Service) ListSerialRecords(ctx context.Context, skuID string) ([]map[string]any, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	return nil, s.NotImplemented()
}

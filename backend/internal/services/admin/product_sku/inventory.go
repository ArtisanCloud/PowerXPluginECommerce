package product_sku

import "context"

// SnapshotInventory will load the latest inventory state for a SKU.
func (s *Service) SnapshotInventory(ctx context.Context, skuID string) (map[string]any, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	return nil, s.NotImplemented()
}

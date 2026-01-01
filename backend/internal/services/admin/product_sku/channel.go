package product_sku

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"gorm.io/datatypes"
)

const (
	channelStatusPending   = "pending"
	channelStatusPublished = "published"
	channelStatusFailed    = "failed"

	defaultSyncMode = "push"
)

// ListChannelMappings fetches all mappings for a SKU.
func (s *Service) ListChannelMappings(ctx context.Context, skuID string) ([]ChannelMapping, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	rows, err := s.ChannelRepo.ListBySKU(ctx, tenantID, skuID)
	if err != nil {
		return nil, err
	}
	result := make([]ChannelMapping, 0, len(rows))
	for _, row := range rows {
		result = append(result, normalizeChannelEntity(row))
	}
	return result, nil
}

// UpsertChannelMapping persists channel metadata for a SKU.
func (s *Service) UpsertChannelMapping(ctx context.Context, skuID string, payload ChannelMappingRequest) (*ChannelMapping, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	channelCode := strings.TrimSpace(payload.ChannelCode)
	if channelCode == "" {
		return nil, errors.New("channel_code is required")
	}
	channelSkuID := strings.TrimSpace(payload.ChannelSKUId)
	if channelSkuID == "" {
		return nil, errors.New("channel_sku_id is required")
	}

	entry := &productskumodel.ProductSKUChannel{
		TenantUUID:   tenantID,
		SKUId:        skuID,
		ChannelCode:  channelCode,
		ChannelSkuID: channelSkuID,
		Status:       normalizeChannelStatus(payload.Status),
		SyncMode:     normalizeSyncMode(payload.SyncMode),
		PublishTime:  payload.PublishTime,
	}
	entry.PriceOverride = marshalJSONValue(payload.PriceOverride)
	entry.MediaOverride = marshalJSONValue(payload.MediaOverride)
	entry.Metadata = marshalJSONValue(payload.Metadata)

	if err := s.ChannelRepo.UpsertMapping(ctx, entry); err != nil {
		return nil, err
	}
	updated, err := s.ChannelRepo.FindBySKUAndCode(ctx, tenantID, skuID, channelCode)
	if err != nil {
		return nil, err
	}
	mapping := normalizeChannelEntity(*updated)
	return &mapping, nil
}

// PublishChannelMapping triggers the asynchronous publish flow.
func (s *Service) PublishChannelMapping(ctx context.Context, skuID string, payload ChannelPublishRequest) (*BulkTaskResponse, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	channelCode := strings.TrimSpace(payload.ChannelCode)
	if channelCode == "" {
		return nil, errors.New("channel_code is required")
	}

	mapping, err := s.ChannelRepo.FindBySKUAndCode(ctx, tenantID, skuID, channelCode)
	if err != nil {
		return nil, err
	}
	if mapping.Status == channelStatusPublished && !payload.Force {
		return nil, fmt.Errorf("channel %s already published", mapping.ChannelCode)
	}

	scope := map[string]any{
		"sku_id":       skuID,
		"channel_code": mapping.ChannelCode,
	}
	op := map[string]any{
		"channel_code":   mapping.ChannelCode,
		"channel_sku_id": mapping.ChannelSkuID,
		"action":         "publish",
	}
	stats := &BulkTaskStats{Succeeded: 1}
	result := map[string]any{
		"message": fmt.Sprintf("channel %s publish simulated", mapping.ChannelCode),
	}

	resp, err := s.createInstantTask(ctx, tenantID, "channel_publish", scope, op, stats, result, nil)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	update := map[string]any{
		"status":          channelStatusPublished,
		"publish_time":    &now,
		"publish_task_id": resp.TaskID,
		"last_error":      "",
	}
	if err := s.ChannelRepo.DB.WithContext(ctx).
		Model(mapping).
		Where("tenant_uuid = ? AND sku_id = ? AND channel_code = ?", tenantID, skuID, mapping.ChannelCode).
		Updates(update).Error; err != nil {
		return nil, err
	}
	return resp, nil
}

func normalizeChannelStatus(in string) string {
	switch strings.ToLower(strings.TrimSpace(in)) {
	case channelStatusPublished:
		return channelStatusPublished
	case channelStatusFailed:
		return channelStatusFailed
	default:
		return channelStatusPending
	}
}

func normalizeSyncMode(in string) string {
	val := strings.ToLower(strings.TrimSpace(in))
	switch val {
	case "push", "pull", "hybrid":
		return val
	default:
		return defaultSyncMode
	}
}

func marshalJSONValue(v any) datatypes.JSON {
	if v == nil {
		return nil
	}
	body, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return datatypes.JSON(body)
}

func normalizeChannelEntity(entity productskumodel.ProductSKUChannel) ChannelMapping {
	var price map[string]any
	if len(entity.PriceOverride) > 0 {
		_ = json.Unmarshal(entity.PriceOverride, &price)
	}
	var media []map[string]any
	if len(entity.MediaOverride) > 0 {
		_ = json.Unmarshal(entity.MediaOverride, &media)
	}
	var metadata map[string]any
	if len(entity.Metadata) > 0 {
		_ = json.Unmarshal(entity.Metadata, &metadata)
	}
	return ChannelMapping{
		ID:              entity.ID,
		ChannelCode:     entity.ChannelCode,
		ChannelSKUId:    entity.ChannelSkuID,
		Status:          entity.Status,
		PublishTime:     entity.PublishTime,
		SyncMode:        entity.SyncMode,
		PriceOverride:   price,
		MediaOverride:   media,
		LastError:       entity.LastError,
		PublishTaskID:   entity.PublishTaskID,
		LastPublishedAt: entity.PublishTime,
		Metadata:        metadata,
	}
}

package spu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelVisibilityInput captures required fields to configure a specific channel.
type ChannelVisibilityInput struct {
	Channel         string         `json:"channel"`
	Availability    string         `json:"availability"`
	PublishAt       *time.Time     `json:"publishAt"`
	WithdrawAt      *time.Time     `json:"withdrawAt"`
	ContentOverride map[string]any `json:"contentOverride"`
	AuditState      string         `json:"auditState"`
	LastFeedback    map[string]any `json:"lastFeedback,omitempty"`
}

// ChannelVisibilityEntry serializes the per-channel configuration persisted for an SPU.
type ChannelVisibilityEntry struct {
	ID              string         `json:"id"`
	Channel         string         `json:"channel"`
	Availability    string         `json:"availability"`
	PublishAt       *time.Time     `json:"publishAt,omitempty"`
	WithdrawAt      *time.Time     `json:"withdrawAt,omitempty"`
	ContentOverride map[string]any `json:"contentOverride,omitempty"`
	AuditState      string         `json:"auditState"`
	LastFeedback    map[string]any `json:"lastFeedback,omitempty"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	CreatedAt       time.Time      `json:"createdAt"`
}

// ChannelService manages per-channel visibility configuration for SPUs.
type ChannelService struct {
	deps        *app.Deps
	spuRepo     *productrepo.SPURepository
	channelRepo *productrepo.ChannelRepository
}

// NewChannelService wires repositories used for channel visibility flows.
func NewChannelService(deps *app.Deps) *ChannelService {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &ChannelService{
		deps:        deps,
		spuRepo:     productrepo.NewSPURepository(deps.DB),
		channelRepo: productrepo.NewChannelRepository(deps.DB),
	}
}

// List returns all channel visibility entries bound to the provided SPU.
func (s *ChannelService) List(ctx context.Context, spuID string) ([]ChannelVisibilityEntry, error) {
	if s == nil {
		return nil, errors.New("channel service uninitialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var records []productmodel.ChannelVisibility
	if err := s.channelRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID).
		Order("channel ASC").
		Find(&records).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []ChannelVisibilityEntry{}, nil
		}
		return nil, err
	}
	return s.toEntries(records), nil
}

// Upsert validates and persists a single channel record for the SPU.
func (s *ChannelService) Upsert(ctx context.Context, spuID string, input ChannelVisibilityInput) (*ChannelVisibilityEntry, error) {
	if s == nil {
		return nil, errors.New("channel service uninitialized")
	}
	if err := s.validateInput(input); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var entry *ChannelVisibilityEntry
	err = s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		payload := productmodel.ChannelVisibility{
			TenantUUID:      tenantID,
			SPUID:           spu.ID,
			Channel:         input.Channel,
			Availability:    strings.ToLower(input.Availability),
			PublishAt:       input.PublishAt,
			WithdrawAt:      input.WithdrawAt,
			ContentOverride: encodeGenericJSON(input.ContentOverride),
			AuditState:      normalizeAuditState(input.AuditState),
			LastFeedback:    encodeGenericJSON(input.LastFeedback),
		}
		var existing productmodel.ChannelVisibility
		err := tx.Where("tenant_uuid = ? AND spu_id = ? AND channel = ?", tenantID, spu.ID, input.Channel).
			First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				payload.ID = uuidString()
				payload.CreatedAt = now
				payload.UpdatedAt = now
				if err := tx.Create(&payload).Error; err != nil {
					return err
				}
				existing = payload
			} else {
				return err
			}
		} else {
			updates := map[string]any{
				"availability":     payload.Availability,
				"publish_at":       payload.PublishAt,
				"withdraw_at":      payload.WithdrawAt,
				"content_override": payload.ContentOverride,
				"audit_state":      payload.AuditState,
				"last_feedback":    payload.LastFeedback,
				"updated_at":       now,
			}
			if err := tx.Model(&existing).
				Where("tenant_uuid = ? AND id = ?", tenantID, existing.ID).
				Updates(updates).Error; err != nil {
				return err
			}
			existing.Availability = payload.Availability
			existing.PublishAt = payload.PublishAt
			existing.WithdrawAt = payload.WithdrawAt
			existing.ContentOverride = payload.ContentOverride
			existing.AuditState = payload.AuditState
			existing.LastFeedback = payload.LastFeedback
			existing.UpdatedAt = now
		}
		if err := s.refreshSummary(tx, tenantID, spu.ID); err != nil {
			return err
		}
		entries := s.toEntries([]productmodel.ChannelVisibility{existing})
		if len(entries) > 0 {
			entry = &entries[0]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// Remove deletes or unlists a channel configuration.
func (s *ChannelService) Remove(ctx context.Context, spuID, channel string) error {
	if s == nil {
		return errors.New("channel service uninitialized")
	}
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return errors.New("channel is required")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return err
	}
	return s.spuRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var spu productmodel.SPU
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantID, spuID).First(&spu).Error; err != nil {
			return err
		}
		if err := tx.Where("tenant_uuid = ? AND spu_id = ? AND channel = ?", tenantID, spu.ID, channel).
			Delete(&productmodel.ChannelVisibility{}).Error; err != nil {
			return err
		}
		return s.refreshSummary(tx, tenantID, spu.ID)
	})
}

func (s *ChannelService) validateInput(input ChannelVisibilityInput) error {
	channel := strings.TrimSpace(input.Channel)
	if channel == "" {
		return errors.New("channel is required")
	}
	availability := strings.ToLower(strings.TrimSpace(input.Availability))
	switch availability {
	case "unlisted", "scheduled", "published", "withheld":
	default:
		return fmt.Errorf("unsupported availability: %s", input.Availability)
	}
	return nil
}

func (s *ChannelService) toEntries(records []productmodel.ChannelVisibility) []ChannelVisibilityEntry {
	if len(records) == 0 {
		return []ChannelVisibilityEntry{}
	}
	result := make([]ChannelVisibilityEntry, 0, len(records))
	for _, record := range records {
		result = append(result, ChannelVisibilityEntry{
			ID:              record.ID,
			Channel:         record.Channel,
			Availability:    record.Availability,
			PublishAt:       record.PublishAt,
			WithdrawAt:      record.WithdrawAt,
			ContentOverride: decodeGenericJSON(record.ContentOverride),
			AuditState:      record.AuditState,
			LastFeedback:    decodeGenericJSON(record.LastFeedback),
			CreatedAt:       record.CreatedAt,
			UpdatedAt:       record.UpdatedAt,
		})
	}
	return result
}

func (s *ChannelService) refreshSummary(tx *gorm.DB, tenantID, spuID string) error {
	var records []productmodel.ChannelVisibility
	if err := tx.Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID).
		Find(&records).Error; err != nil {
		return err
	}
	summary := summarizeChannelRecords(records)
	return tx.Model(&productmodel.SPU{}).
		Where("tenant_uuid = ? AND id = ?", tenantID, spuID).
		Update("channels_summary", summary).Error
}

func (s *ChannelService) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil || s == nil {
		return "", errors.New("channel service unavailable")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", ErrMissingTenant
}

func summarizeChannelRecords(records []productmodel.ChannelVisibility) datatypes.JSON {
	if len(records) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	payload := make(map[string]map[string]any, len(records))
	for _, record := range records {
		if strings.TrimSpace(record.Channel) == "" {
			continue
		}
		payload[record.Channel] = map[string]any{
			"availability": record.Availability,
		}
		if record.PublishAt != nil {
			payload[record.Channel]["publishAt"] = record.PublishAt
		}
		if record.WithdrawAt != nil {
			payload[record.Channel]["withdrawAt"] = record.WithdrawAt
		}
	}
	if len(payload) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	body, _ := json.Marshal(payload)
	return datatypes.JSON(body)
}

func encodeGenericJSON(data map[string]any) datatypes.JSON {
	if len(data) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	body, err := json.Marshal(data)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(body)
}

func decodeGenericJSON(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func normalizeAuditState(state string) string {
	state = strings.ToLower(strings.TrimSpace(state))
	switch state {
	case "approved", "rejected", "pending":
		return state
	default:
		return "pending"
	}
}

package channel_master

import (
	"context"
	"errors"
	"strings"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel_master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/sirupsen/logrus"
)

// AuditEmitter is a narrow contract for emitting structured audit events.
type AuditEmitter interface {
	EmitChannelAudit(ctx context.Context, action string, payload map[string]any) error
}

// DTOMapper converts database entities into transport-friendly payloads.
type DTOMapper interface {
	ToChannelSummary(channel *channelmodel.ChannelMaster) ChannelSummaryDTO
}

// ChannelSummaryDTO is a light-weight response used by admin APIs.
type ChannelSummaryDTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Platform    string   `json:"platform"`
	Region      string   `json:"region"`
	Status      string   `json:"status"`
	OwnerUUID   string   `json:"ownerUuid"`
	Tags        []string `json:"tags"`
	ChannelType string   `json:"channelType"`
}

// CreateChannelInput captures the minimal fields required to draft a channel.
type CreateChannelInput struct {
	Name        string   `json:"name"`
	Platform    string   `json:"platform"`
	StoreID     string   `json:"storeId"`
	Region      string   `json:"region"`
	OwnerUUID   string   `json:"ownerUuid"`
	ChannelType string   `json:"channelType"`
	Tags        []string `json:"tags"`
}

// Normalize trims whitespace and deduplicates tags.
func (in *CreateChannelInput) Normalize() {
	if in == nil {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Platform = strings.TrimSpace(in.Platform)
	in.StoreID = strings.TrimSpace(in.StoreID)
	in.Region = strings.TrimSpace(in.Region)
	in.OwnerUUID = strings.TrimSpace(in.OwnerUUID)
	in.ChannelType = strings.TrimSpace(strings.ToLower(in.ChannelType))
	if len(in.Tags) == 0 {
		return
	}
	set := make(map[string]struct{}, len(in.Tags))
	for _, tag := range in.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			set[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	in.Tags = in.Tags[:0]
	for tag := range set {
		in.Tags = append(in.Tags, tag)
	}
}

// Validate ensures required fields are present before persistence.
func (in *CreateChannelInput) Validate() error {
	if in == nil {
		return errors.New("input is required")
	}
	in.Normalize()
	var missing []string
	if in.Name == "" {
		missing = append(missing, "name")
	}
	if in.Platform == "" {
		missing = append(missing, "platform")
	}
	if in.StoreID == "" {
		missing = append(missing, "storeId")
	}
	if in.Region == "" {
		missing = append(missing, "region")
	}
	if in.OwnerUUID == "" {
		missing = append(missing, "ownerUuid")
	}
	if in.ChannelType == "" {
		missing = append(missing, "channelType")
	}
	if len(missing) > 0 {
		return errors.New("missing required fields: " + strings.Join(missing, ","))
	}
	return nil
}

// Service coordinates repositories, mappers and observability helpers for channel master flows.
type Service struct {
	deps    *app.Deps
	repo    *channelrepo.ChannelMasterRepository
	mapper  DTOMapper
	audit   AuditEmitter
	metrics *channelobs.Metrics
	logger  *logrus.Entry
}

// NewService wires required collaborators; mapper/audit emitters can be nil for now.
func NewService(deps *app.Deps, mapper DTOMapper, audit AuditEmitter) *Service {
	if deps == nil || deps.DB == nil {
		panic("channel master service requires database dependency")
	}
	if mapper == nil {
		mapper = defaultMapper{}
	}
	logger := deps.RuntimeLogger(nil, "channel-master-service", nil)
	return &Service{
		deps:    deps,
		repo:    channelrepo.NewChannelMasterRepository(deps.DB),
		mapper:  mapper,
		audit:   audit,
		metrics: channelobs.NewMetrics(logger),
		logger:  logger,
	}
}

// List returns a paginated result mapped into DTOs.
func (s *Service) List(ctx context.Context, filters channelrepo.ChannelListFilters) (*repo.Page[[]ChannelSummaryDTO], error) {
	page, err := s.repo.FindPage(ctx, filters)
	if err != nil {
		return nil, err
	}
	items := make([]ChannelSummaryDTO, 0, len(page.List))
	for _, channel := range page.List {
		items = append(items, s.mapper.ToChannelSummary(channel))
	}
	return &repo.Page[[]ChannelSummaryDTO]{
		List:      items,
		PageIndex: filters.Page,
		PageSize:  filters.PageSize,
		Total:     page.Total,
	}, nil
}

// CreateDraft persists a new channel placeholder and emits audit events.
func (s *Service) CreateDraft(ctx context.Context, input CreateChannelInput) (*ChannelSummaryDTO, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	entity := &channelmodel.ChannelMaster{
		Name:        input.Name,
		Platform:    input.Platform,
		StoreID:     input.StoreID,
		Region:      input.Region,
		OwnerUUID:   input.OwnerUUID,
		ChannelType: input.ChannelType,
		Status:      "draft",
		Tags:        input.Tags,
	}
	saved, err := s.repo.Upsert(ctx, entity)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		_ = s.audit.EmitChannelAudit(ctx, "channel.created", map[string]any{"channel_id": saved.ID, "platform": saved.Platform})
	}
	summary := s.mapper.ToChannelSummary(saved)
	return &summary, nil
}

// defaultMapper converts models to basic DTOs.
type defaultMapper struct{}

func (defaultMapper) ToChannelSummary(channel *channelmodel.ChannelMaster) ChannelSummaryDTO {
	if channel == nil {
		return ChannelSummaryDTO{}
	}
	return ChannelSummaryDTO{
		ID:          channel.ID,
		Name:        channel.Name,
		Platform:    channel.Platform,
		Region:      channel.Region,
		Status:      channel.Status,
		OwnerUUID:   channel.OwnerUUID,
		Tags:        []string(channel.Tags),
		ChannelType: channel.ChannelType,
	}
}

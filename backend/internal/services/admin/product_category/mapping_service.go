package product_category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	adminconsole "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/admin_console"
	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	mappingrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/jackc/pgconn"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CategoryMappingDTO struct {
	ID                 string         `json:"id"`
	CategoryID         string         `json:"categoryId"`
	Channel            string         `json:"channel"`
	PlatformCategoryID string         `json:"platformCategoryId"`
	Strategy           string         `json:"strategy"`
	SyncStatus         string         `json:"syncStatus"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	CreatedAt          time.Time      `json:"createdAt"`
}

type UpsertMappingRequest struct {
	Operation          string         `json:"operation,omitempty"` // upsert|delete
	Channel            string         `json:"channel"`
	PlatformCategoryID string         `json:"platformCategoryId,omitempty"`
	Strategy           string         `json:"strategy,omitempty"`
	SyncStatus         string         `json:"syncStatus,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

func (s *Service) ListMappings(ctx context.Context, categoryID string) ([]CategoryMappingDTO, error) {
	if s == nil || !s.Ready() || s.MappingRepo == nil {
		return nil, errors.New("mapping service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, errors.New("categoryId is required")
	}
	records, err := s.MappingRepo.ListByCategory(ctx, tenantID, categoryID)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryMappingDTO, 0, len(records))
	for _, rec := range records {
		out = append(out, mapMappingDTO(rec))
	}
	return out, nil
}

func (s *Service) UpsertMapping(ctx context.Context, categoryID string, req UpsertMappingRequest) (*CategoryMappingDTO, error) {
	if s == nil || !s.Ready() || s.MappingRepo == nil {
		return nil, errors.New("mapping service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryID = strings.TrimSpace(categoryID)
	req.Channel = strings.TrimSpace(req.Channel)
	req.Operation = strings.TrimSpace(strings.ToLower(req.Operation))
	if categoryID == "" {
		return nil, errors.New("categoryId is required")
	}
	if req.Channel == "" {
		return nil, ValidationErrors{}.add("channel", "channel 为必填")
	}
	if req.Operation == "delete" {
		if err := s.MappingRepo.DeleteByCategoryChannel(ctx, tenantID, categoryID, req.Channel); err != nil {
			return nil, err
		}
		_ = s.writeCategoryAudit(ctx, tenantID, "mapping_delete", categoryID, map[string]any{"channel": req.Channel})
		return nil, nil
	}

	req.PlatformCategoryID = strings.TrimSpace(req.PlatformCategoryID)
	if req.PlatformCategoryID == "" {
		return nil, ValidationErrors{}.add("platformCategoryId", "platformCategoryId 为必填")
	}
	strategy := strings.TrimSpace(strings.ToLower(req.Strategy))
	if strategy == "" {
		strategy = productcategory.MappingStrategyManual
	}
	if strategy != productcategory.MappingStrategyManual && strategy != productcategory.MappingStrategyAuto {
		return nil, ValidationErrors{}.add("strategy", "strategy 必须为 manual/auto")
	}
	syncStatus := strings.TrimSpace(strings.ToLower(req.SyncStatus))
	if syncStatus == "" {
		syncStatus = productcategory.MappingSyncPending
	}

	now := time.Now().UTC()
	metaBody, _ := json.Marshal(req.Metadata)
	entity := &productcategory.CategoryMapping{
		ID:                 utils.NewUUID(),
		TenantUUID:         tenantID,
		CategoryID:         categoryID,
		Channel:            req.Channel,
		PlatformCategoryID: req.PlatformCategoryID,
		Strategy:           strategy,
		SyncStatus:         syncStatus,
		Metadata:           datatypes.JSON(metaBody),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	var saved *productcategory.CategoryMapping
	err = s.MappingRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := mappingrepo.NewMappingRepository(tx)
		rec, err := repoTx.Upsert(ctx, tx, entity)
		if err != nil {
			if conflict := toMappingConflictError(err); conflict != nil {
				return conflict
			}
			return err
		}
		saved = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.writeCategoryAudit(ctx, tenantID, "mapping_upsert", categoryID, map[string]any{
		"channel":            req.Channel,
		"platformCategoryId": req.PlatformCategoryID,
		"strategy":           strategy,
		"syncStatus":         syncStatus,
	})
	dto := mapMappingDTO(*saved)
	return &dto, nil
}

func toMappingConflictError(err error) *ConflictError {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch strings.TrimSpace(pgErr.ConstraintName) {
		case "uk_category_mapping":
			return &ConflictError{Field: "channel", Message: "同一类目下该 channel 映射已存在"}
		default:
			return &ConflictError{Message: "数据冲突（可能违反唯一性约束）"}
		}
	}
	return nil
}

func mapMappingDTO(rec productcategory.CategoryMapping) CategoryMappingDTO {
	rec.Normalize()
	var meta map[string]any
	_ = json.Unmarshal(rec.Metadata, &meta)
	return CategoryMappingDTO{
		ID:                 rec.ID,
		CategoryID:         rec.CategoryID,
		Channel:            rec.Channel,
		PlatformCategoryID: rec.PlatformCategoryID,
		Strategy:           rec.Strategy,
		SyncStatus:         rec.SyncStatus,
		Metadata:           meta,
		UpdatedAt:          rec.UpdatedAt,
		CreatedAt:          rec.CreatedAt,
	}
}

func (s *Service) writeCategoryAudit(ctx context.Context, tenantID string, action string, categoryID string, diff map[string]any) error {
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return nil
	}
	body, _ := json.Marshal(diff)
	now := time.Now().UTC()
	categoryID = strings.TrimSpace(categoryID)
	var ref *string
	if categoryID != "" {
		ref = &categoryID
	}
	event := &adminconsole.AuditEvent{
		ID:             utils.NewUUID(),
		PluginID:       app.PluginID,
		TenantUuid:     &tenantID,
		ActorID:        "system",
		PermissionCode: "admin.product.categories.mappings",
		Action:         action,
		ResourceType:   "product_category",
		ResourceRef:    ref,
		Summary:        ptrString(fmt.Sprintf("category %s", action)),
		Diff:           datatypes.JSON(body),
		OccurredAt:     now,
		CreatedAt:      now,
	}
	return s.deps.DB.WithContext(ctx).Create(event).Error
}

func ptrString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

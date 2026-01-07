package product_spec

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_spec"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

type Service struct {
	deps      *app.Deps
	GroupRepo *repo.GroupRepository
	OptionRepo *repo.OptionRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	return &Service{
		deps:      deps,
		GroupRepo: repo.NewGroupRepository(deps.DB),
		OptionRepo: repo.NewOptionRepository(deps.DB),
	}
}

func (s *Service) Ready() bool { return s != nil && s.deps != nil && s.deps.DB != nil }

func (s *Service) HealthProbe(ctx context.Context) error {
	if !s.Ready() {
		return errors.New("product spec service not ready")
	}
	return ctx.Err()
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("request context missing")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", errors.New("tenant context missing")
}

type OptionInput struct {
	ID        string         `json:"id,omitempty"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	SortOrder int            `json:"sort_order,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	Status    string         `json:"status,omitempty"`
}

type GroupInput struct {
	ID        string        `json:"id,omitempty"`
	Code      string        `json:"code"`
	Name      string        `json:"name"`
	SortOrder int           `json:"sort_order,omitempty"`
	Required  *bool         `json:"required,omitempty"`
	Status    string        `json:"status,omitempty"`
	Options   []OptionInput `json:"options,omitempty"`
}

type ReplaceRequest struct {
	Groups []GroupInput `json:"groups"`
}

type SpecGroupDTO struct {
	ID        string                 `json:"id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	SortOrder int                    `json:"sort_order"`
	Required  bool                   `json:"required"`
	Status    string                 `json:"status"`
	Options   []SpecOptionDTO        `json:"options"`
}

type SpecOptionDTO struct {
	ID        string                 `json:"id"`
	GroupID   string                 `json:"group_id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	SortOrder int                    `json:"sort_order"`
	Meta      map[string]any         `json:"meta,omitempty"`
	Status    string                 `json:"status"`
}

func (s *Service) List(ctx context.Context, spuID string) ([]SpecGroupDTO, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	spuID = strings.TrimSpace(spuID)
	if spuID == "" {
		return nil, errors.New("spu id is required")
	}
	groups, err := s.GroupRepo.ListBySPU(ctx, tenantID, spuID)
	if err != nil {
		return nil, err
	}
	groupIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}
	options, err := s.OptionRepo.ListByGroupIDs(ctx, tenantID, groupIDs)
	if err != nil {
		return nil, err
	}
	optsByGroup := make(map[string][]productspecmodel.ProductSpecOption, len(groupIDs))
	for _, o := range options {
		optsByGroup[o.GroupID] = append(optsByGroup[o.GroupID], o)
	}
	out := make([]SpecGroupDTO, 0, len(groups))
	for _, g := range groups {
		dto := SpecGroupDTO{
			ID:        g.ID,
			Code:      g.Code,
			Name:      g.Name,
			SortOrder: g.SortOrder,
			Required:  g.Required,
			Status:    g.Status,
			Options:   []SpecOptionDTO{},
		}
		for _, o := range optsByGroup[g.ID] {
			dto.Options = append(dto.Options, SpecOptionDTO{
				ID:        o.ID,
				GroupID:   o.GroupID,
				Code:      o.Code,
				Name:      o.Name,
				SortOrder: o.SortOrder,
				Meta:      decodeMeta(o.Meta),
				Status:    o.Status,
			})
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) Replace(ctx context.Context, spuID string, req ReplaceRequest) ([]SpecGroupDTO, error) {
	if err := s.HealthProbe(ctx); err != nil {
		return nil, err
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	spuID = strings.TrimSpace(spuID)
	if spuID == "" {
		return nil, errors.New("spu id is required")
	}
	now := time.Now().UTC()

	err = s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existingGroups, err := s.GroupRepo.ListBySPU(ctx, tenantID, spuID)
		if err != nil {
			return err
		}
		existingGroupIDs := make([]string, 0, len(existingGroups))
		for _, g := range existingGroups {
			existingGroupIDs = append(existingGroupIDs, g.ID)
		}
		existingOptions, err := s.OptionRepo.ListByGroupIDs(ctx, tenantID, existingGroupIDs)
		if err != nil {
			return err
		}
		existingGroupsByID := make(map[string]productspecmodel.ProductSpecGroup, len(existingGroups))
		for _, g := range existingGroups {
			existingGroupsByID[g.ID] = g
		}
		existingOptionsByID := make(map[string]productspecmodel.ProductSpecOption, len(existingOptions))
		for _, o := range existingOptions {
			existingOptionsByID[o.ID] = o
		}

		keepGroups := make(map[string]struct{}, len(req.Groups))
		keepOptions := make(map[string]struct{})

		for _, input := range req.Groups {
			code := strings.TrimSpace(input.Code)
			name := strings.TrimSpace(input.Name)
			if code == "" || name == "" {
				return errors.New("group code/name is required")
			}
			id := strings.TrimSpace(input.ID)
			if id == "" {
				id = utils.NewUUID()
			}
			required := true
			if input.Required != nil {
				required = *input.Required
			}
			status := strings.TrimSpace(input.Status)
			if status == "" {
				status = "active"
			}
			group := productspecmodel.ProductSpecGroup{
				ID:        id,
				TenantUUID: tenantID,
				SPUID:     spuID,
				Code:      code,
				Name:      name,
				SortOrder: input.SortOrder,
				Required:  required,
				Status:    status,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if existing, ok := existingGroupsByID[id]; ok {
				group.CreatedAt = existing.CreatedAt
				if err := tx.Model(&productspecmodel.ProductSpecGroup{}).
					Where("tenant_uuid = ? AND id = ?", tenantID, id).
					Updates(map[string]any{
						"code":       group.Code,
						"name":       group.Name,
						"sort_order": group.SortOrder,
						"required":   group.Required,
						"status":     group.Status,
						"updated_at": now,
						"spu_id":     spuID,
					}).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Create(&group).Error; err != nil {
					return err
				}
			}
			keepGroups[id] = struct{}{}

			for _, opt := range input.Options {
				optCode := strings.TrimSpace(opt.Code)
				optName := strings.TrimSpace(opt.Name)
				if optCode == "" || optName == "" {
					return fmt.Errorf("option code/name is required for group %s", code)
				}
				optID := strings.TrimSpace(opt.ID)
				if optID == "" {
					optID = utils.NewUUID()
				}
				optStatus := strings.TrimSpace(opt.Status)
				if optStatus == "" {
					optStatus = "active"
				}
				metaJSON, err := encodeMeta(opt.Meta)
				if err != nil {
					return err
				}
				option := productspecmodel.ProductSpecOption{
					ID:        optID,
					TenantUUID: tenantID,
					SPUID:     spuID,
					GroupID:   id,
					Code:      optCode,
					Name:      optName,
					SortOrder: opt.SortOrder,
					Meta:      metaJSON,
					Status:    optStatus,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if existing, ok := existingOptionsByID[optID]; ok {
					option.CreatedAt = existing.CreatedAt
					if err := tx.Model(&productspecmodel.ProductSpecOption{}).
						Where("tenant_uuid = ? AND id = ?", tenantID, optID).
						Updates(map[string]any{
							"spu_id":     spuID,
							"group_id":   id,
							"code":       option.Code,
							"name":       option.Name,
							"sort_order": option.SortOrder,
							"meta":       option.Meta,
							"status":     option.Status,
							"updated_at": now,
						}).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Create(&option).Error; err != nil {
						return err
					}
				}
				keepOptions[optID] = struct{}{}
			}
		}

		// Soft delete removed options first.
		if len(existingOptions) > 0 {
			toDelete := make([]string, 0)
			for _, o := range existingOptions {
				if _, keep := keepOptions[o.ID]; keep {
					continue
				}
				toDelete = append(toDelete, o.ID)
			}
			if len(toDelete) > 0 {
				if err := tx.Where("tenant_uuid = ? AND id IN ?", tenantID, toDelete).
					Delete(&productspecmodel.ProductSpecOption{}).Error; err != nil {
					return err
				}
			}
		}
		// Soft delete removed groups.
		if len(existingGroups) > 0 {
			toDelete := make([]string, 0)
			for _, g := range existingGroups {
				if _, keep := keepGroups[g.ID]; keep {
					continue
				}
				toDelete = append(toDelete, g.ID)
			}
			if len(toDelete) > 0 {
				if err := tx.Where("tenant_uuid = ? AND id IN ?", tenantID, toDelete).
					Delete(&productspecmodel.ProductSpecGroup{}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.List(ctx, spuID)
}

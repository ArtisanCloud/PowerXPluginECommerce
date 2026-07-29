package product_category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SaleSpecOptionInput struct {
	ID        string         `json:"id,omitempty"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	SortOrder int            `json:"sortOrder,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	Status    string         `json:"status,omitempty"`
}

type SaleSpecGroupInput struct {
	ID          string                `json:"id,omitempty"`
	Code        string                `json:"code"`
	Name        string                `json:"name"`
	SortOrder   int                   `json:"sortOrder,omitempty"`
	Required    *bool                 `json:"required,omitempty"`
	AllowCustom bool                  `json:"allowCustom,omitempty"`
	Status      string                `json:"status,omitempty"`
	Options     []SaleSpecOptionInput `json:"options,omitempty"`
}

type ReplaceSaleSpecsRequest struct {
	Groups []SaleSpecGroupInput `json:"groups"`
}

type SaleSpecOptionDTO struct {
	ID        string         `json:"id"`
	GroupID   string         `json:"groupId"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	SortOrder int            `json:"sortOrder"`
	Meta      map[string]any `json:"meta,omitempty"`
	Status    string         `json:"status"`
}

type SaleSpecGroupDTO struct {
	ID          string              `json:"id"`
	CategoryID  string              `json:"categoryId"`
	Code        string              `json:"code"`
	Name        string              `json:"name"`
	SortOrder   int                 `json:"sortOrder"`
	Required    bool                `json:"required"`
	AllowCustom bool                `json:"allowCustom"`
	Status      string              `json:"status"`
	Options     []SaleSpecOptionDTO `json:"options"`
}

func (s *Service) ListSaleSpecs(ctx context.Context, categoryID string) ([]SaleSpecGroupDTO, error) {
	if s == nil || !s.Ready() {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, errors.New("category id is required")
	}
	var exists int64
	if err := s.deps.DB.WithContext(ctx).
		Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ? AND id = ? AND deleted_at IS NULL", tenantID, categoryID).
		Count(&exists).Error; err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.listSaleSpecs(ctx, tenantID, categoryID)
}

func (s *Service) ReplaceSaleSpecs(ctx context.Context, categoryID string, req ReplaceSaleSpecsRequest) ([]SaleSpecGroupDTO, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, errors.New("category id is required")
	}
	if err := validateSaleSpecInputs(req.Groups); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&productcategory.ProductCategory{}).
			Where("tenant_uuid = ? AND id = ? AND deleted_at IS NULL", tenantID, categoryID).
			Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return gorm.ErrRecordNotFound
		}

		var existingGroups []productcategory.CategorySaleSpecGroup
		if err := tx.Where("tenant_uuid = ? AND category_id = ? AND deleted_at IS NULL", tenantID, categoryID).
			Find(&existingGroups).Error; err != nil {
			return err
		}
		groupIDs := make([]string, 0, len(existingGroups))
		existingGroupsByID := make(map[string]productcategory.CategorySaleSpecGroup, len(existingGroups))
		for _, group := range existingGroups {
			groupIDs = append(groupIDs, group.ID)
			existingGroupsByID[group.ID] = group
		}

		var existingOptions []productcategory.CategorySaleSpecOption
		if len(groupIDs) > 0 {
			if err := tx.Where("tenant_uuid = ? AND group_id IN ? AND deleted_at IS NULL", tenantID, groupIDs).
				Find(&existingOptions).Error; err != nil {
				return err
			}
		}
		existingOptionsByID := make(map[string]productcategory.CategorySaleSpecOption, len(existingOptions))
		for _, option := range existingOptions {
			existingOptionsByID[option.ID] = option
		}

		keepGroups := make(map[string]struct{}, len(req.Groups))
		keepOptions := make(map[string]struct{})
		for _, input := range req.Groups {
			groupID := strings.TrimSpace(input.ID)
			if groupID == "" {
				groupID = utils.NewUUID()
			} else if _, ok := existingGroupsByID[groupID]; !ok {
				return ValidationErrors{}.add("groups", fmt.Sprintf("sale spec group id does not belong to category: %s", groupID))
			}
			required := true
			if input.Required != nil {
				required = *input.Required
			}
			status, err := normalizeSaleSpecStatus(input.Status)
			if err != nil {
				return ValidationErrors{}.add(input.Code, err.Error())
			}
			group := productcategory.CategorySaleSpecGroup{
				ID:          groupID,
				TenantUUID:  tenantID,
				CategoryID:  categoryID,
				Code:        strings.TrimSpace(input.Code),
				Name:        strings.TrimSpace(input.Name),
				SortOrder:   input.SortOrder,
				Required:    required,
				AllowCustom: input.AllowCustom,
				Status:      status,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if existing, ok := existingGroupsByID[groupID]; ok {
				group.CreatedAt = existing.CreatedAt
				if err := tx.Model(&productcategory.CategorySaleSpecGroup{}).
					Where("tenant_uuid = ? AND id = ?", tenantID, groupID).
					Updates(map[string]any{
						"category_id":  categoryID,
						"code":         group.Code,
						"name":         group.Name,
						"sort_order":   group.SortOrder,
						"required":     group.Required,
						"allow_custom": group.AllowCustom,
						"status":       group.Status,
						"updated_at":   now,
					}).Error; err != nil {
					return err
				}
			} else if err := tx.Create(&group).Error; err != nil {
				return err
			}
			keepGroups[groupID] = struct{}{}

			for _, optInput := range input.Options {
				optionID := strings.TrimSpace(optInput.ID)
				if optionID == "" {
					optionID = utils.NewUUID()
				} else if _, ok := existingOptionsByID[optionID]; !ok {
					return ValidationErrors{}.add(input.Code, fmt.Sprintf("sale spec option id does not belong to category: %s", optionID))
				}
				meta, err := encodeSaleSpecMeta(optInput.Meta)
				if err != nil {
					return err
				}
				optionStatus, err := normalizeSaleSpecStatus(optInput.Status)
				if err != nil {
					return ValidationErrors{}.add(input.Code, err.Error())
				}
				option := productcategory.CategorySaleSpecOption{
					ID:         optionID,
					TenantUUID: tenantID,
					CategoryID: categoryID,
					GroupID:    groupID,
					Code:       strings.TrimSpace(optInput.Code),
					Name:       strings.TrimSpace(optInput.Name),
					SortOrder:  optInput.SortOrder,
					Meta:       meta,
					Status:     optionStatus,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				if existing, ok := existingOptionsByID[optionID]; ok {
					option.CreatedAt = existing.CreatedAt
					if err := tx.Model(&productcategory.CategorySaleSpecOption{}).
						Where("tenant_uuid = ? AND id = ?", tenantID, optionID).
						Updates(map[string]any{
							"category_id": categoryID,
							"group_id":    groupID,
							"code":        option.Code,
							"name":        option.Name,
							"sort_order":  option.SortOrder,
							"meta":        option.Meta,
							"status":      option.Status,
							"updated_at":  now,
						}).Error; err != nil {
						return err
					}
				} else if err := tx.Create(&option).Error; err != nil {
					return err
				}
				keepOptions[optionID] = struct{}{}
			}
		}

		if len(existingOptions) > 0 {
			toDelete := make([]string, 0)
			for _, option := range existingOptions {
				if _, keep := keepOptions[option.ID]; !keep {
					toDelete = append(toDelete, option.ID)
				}
			}
			if len(toDelete) > 0 {
				if err := tx.Where("tenant_uuid = ? AND id IN ?", tenantID, toDelete).
					Delete(&productcategory.CategorySaleSpecOption{}).Error; err != nil {
					return err
				}
			}
		}
		if len(existingGroups) > 0 {
			toDelete := make([]string, 0)
			for _, group := range existingGroups {
				if _, keep := keepGroups[group.ID]; !keep {
					toDelete = append(toDelete, group.ID)
				}
			}
			if len(toDelete) > 0 {
				if err := tx.Where("tenant_uuid = ? AND id IN ?", tenantID, toDelete).
					Delete(&productcategory.CategorySaleSpecGroup{}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.listSaleSpecs(ctx, tenantID, categoryID)
}

func (s *Service) listSaleSpecs(ctx context.Context, tenantID, categoryID string) ([]SaleSpecGroupDTO, error) {
	var groups []productcategory.CategorySaleSpecGroup
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND category_id = ? AND deleted_at IS NULL", tenantID, categoryID).
		Order("sort_order ASC, code ASC, created_at ASC").
		Find(&groups).Error; err != nil {
		return nil, err
	}
	groupIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}
	optionsByGroup := make(map[string][]productcategory.CategorySaleSpecOption, len(groups))
	if len(groupIDs) > 0 {
		var options []productcategory.CategorySaleSpecOption
		if err := s.deps.DB.WithContext(ctx).
			Where("tenant_uuid = ? AND group_id IN ? AND deleted_at IS NULL", tenantID, groupIDs).
			Order("sort_order ASC, code ASC, created_at ASC").
			Find(&options).Error; err != nil {
			return nil, err
		}
		for _, option := range options {
			optionsByGroup[option.GroupID] = append(optionsByGroup[option.GroupID], option)
		}
	}
	result := make([]SaleSpecGroupDTO, 0, len(groups))
	for _, group := range groups {
		dto := SaleSpecGroupDTO{
			ID:          group.ID,
			CategoryID:  group.CategoryID,
			Code:        group.Code,
			Name:        group.Name,
			SortOrder:   group.SortOrder,
			Required:    group.Required,
			AllowCustom: group.AllowCustom,
			Status:      group.Status,
			Options:     []SaleSpecOptionDTO{},
		}
		for _, option := range optionsByGroup[group.ID] {
			dto.Options = append(dto.Options, SaleSpecOptionDTO{
				ID:        option.ID,
				GroupID:   option.GroupID,
				Code:      option.Code,
				Name:      option.Name,
				SortOrder: option.SortOrder,
				Meta:      decodeSaleSpecMeta(option.Meta),
				Status:    option.Status,
			})
		}
		result = append(result, dto)
	}
	return result, nil
}

func validateSaleSpecInputs(groups []SaleSpecGroupInput) error {
	groupCodes := map[string]struct{}{}
	for _, group := range groups {
		code := strings.TrimSpace(group.Code)
		name := strings.TrimSpace(group.Name)
		if code == "" || name == "" {
			return ValidationErrors{}.add("groups", "sale spec group code/name is required")
		}
		codeKey := strings.ToLower(code)
		if _, exists := groupCodes[codeKey]; exists {
			return ValidationErrors{}.add("groups", fmt.Sprintf("duplicate sale spec group code: %s", code))
		}
		groupCodes[codeKey] = struct{}{}
		if len(group.Options) == 0 {
			return ValidationErrors{}.add(code, "sale spec group must have at least one option")
		}
		optionCodes := map[string]struct{}{}
		for _, option := range group.Options {
			optCode := strings.TrimSpace(option.Code)
			optName := strings.TrimSpace(option.Name)
			if optCode == "" || optName == "" {
				return ValidationErrors{}.add(code, "sale spec option code/name is required")
			}
			optKey := strings.ToLower(optCode)
			if _, exists := optionCodes[optKey]; exists {
				return ValidationErrors{}.add(code, fmt.Sprintf("duplicate sale spec option code: %s", optCode))
			}
			optionCodes[optKey] = struct{}{}
		}
	}
	return nil
}

func normalizeSaleSpecStatus(raw string) (string, error) {
	status := strings.TrimSpace(strings.ToLower(raw))
	if status == "" {
		return productcategory.SaleSpecStatusActive, nil
	}
	switch status {
	case productcategory.SaleSpecStatusActive, productcategory.SaleSpecStatusDisabled:
		return status, nil
	default:
		return "", fmt.Errorf("sale spec status must be %s/%s", productcategory.SaleSpecStatusActive, productcategory.SaleSpecStatusDisabled)
	}
}

func encodeSaleSpecMeta(meta map[string]any) (datatypes.JSON, error) {
	if len(meta) == 0 {
		return datatypes.JSON([]byte("null")), nil
	}
	body, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(body), nil
}

func decodeSaleSpecMeta(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

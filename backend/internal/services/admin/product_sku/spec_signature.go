package product_sku

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	productspecmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	"gorm.io/gorm"
)

type signatureEntry struct {
	GroupCode string
	OptionCode string
	SortOrder int
}

func (s *Service) buildSpecSignature(ctx context.Context, tx *gorm.DB, tenantID, spuID string, specs []SkuSpec) (string, error) {
	if tx == nil {
		return "", errors.New("db unavailable")
	}
	if strings.TrimSpace(spuID) == "" {
		return "", errors.New("spu id is required")
	}
	// If the SPU has no spec groups configured, keep signature empty for backward compatibility.
	var groupCount int64
	if err := tx.WithContext(ctx).
		Model(&productspecmodel.ProductSpecGroup{}).
		Where("tenant_uuid = ? AND spu_id = ? AND deleted_at IS NULL", tenantID, spuID).
		Count(&groupCount).Error; err != nil {
		// 兼容旧环境/测试环境：若尚未迁移出规格表，则直接跳过 signature。
		if isMissingSpecTable(err) {
			return "", nil
		}
		return "", err
	}
	if groupCount == 0 {
		return "", nil
	}

	if len(specs) == 0 {
		// SPU has spec groups but SKU specs are empty -> treat as invalid.
		return "", errors.New("sku specs are required for spec-enabled spu")
	}

	groupIDs := make([]string, 0, len(specs))
	optionIDs := make([]string, 0, len(specs))
	seenGroups := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		gid := strings.TrimSpace(spec.SpecID)
		oid := strings.TrimSpace(spec.ValueID)
		if gid == "" || oid == "" {
			continue
		}
		if _, dup := seenGroups[gid]; dup {
			return "", fmt.Errorf("duplicate spec group %s in sku specs", gid)
		}
		seenGroups[gid] = struct{}{}
		groupIDs = append(groupIDs, gid)
		optionIDs = append(optionIDs, oid)
	}
	if len(groupIDs) == 0 {
		return "", errors.New("sku specs missing spec_id/value_id")
	}

	var groups []productspecmodel.ProductSpecGroup
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ? AND id IN ? AND deleted_at IS NULL AND status = 'active'", tenantID, spuID, groupIDs).
		Find(&groups).Error; err != nil {
		return "", err
	}
	groupByID := make(map[string]productspecmodel.ProductSpecGroup, len(groups))
	for _, g := range groups {
		groupByID[g.ID] = g
	}

	var options []productspecmodel.ProductSpecOption
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ? AND id IN ? AND deleted_at IS NULL AND status = 'active'", tenantID, spuID, optionIDs).
		Find(&options).Error; err != nil {
		return "", err
	}
	optionByID := make(map[string]productspecmodel.ProductSpecOption, len(options))
	for _, o := range options {
		optionByID[o.ID] = o
	}

	entries := make([]signatureEntry, 0, len(groupIDs))
	for _, spec := range specs {
		gid := strings.TrimSpace(spec.SpecID)
		oid := strings.TrimSpace(spec.ValueID)
		if gid == "" || oid == "" {
			continue
		}
		g, ok := groupByID[gid]
		if !ok {
			return "", fmt.Errorf("spec group %s not found for spu", gid)
		}
		o, ok := optionByID[oid]
		if !ok {
			return "", fmt.Errorf("spec option %s not found for spu", oid)
		}
		if o.GroupID != g.ID {
			return "", fmt.Errorf("spec option %s does not belong to group %s", oid, gid)
		}
		groupCode := strings.TrimSpace(g.Code)
		optionCode := strings.TrimSpace(o.Code)
		if groupCode == "" || optionCode == "" {
			return "", errors.New("spec code missing")
		}
		entries = append(entries, signatureEntry{
			GroupCode: groupCode,
			OptionCode: optionCode,
			SortOrder: g.SortOrder,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].SortOrder == entries[j].SortOrder {
			return entries[i].GroupCode < entries[j].GroupCode
		}
		return entries[i].SortOrder < entries[j].SortOrder
	})
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, e.GroupCode+"="+e.OptionCode)
	}
	return strings.Join(parts, "|"), nil
}

func isMissingSpecTable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "no such table: product_spec_groups") {
		return true
	}
	if strings.Contains(msg, "relation \"product_spec_groups\" does not exist") {
		return true
	}
	return false
}

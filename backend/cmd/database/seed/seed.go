package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	iammodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	pricingmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	productsku "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	productspec "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_spec"
	templatemodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/template"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const defaultTenantUUID = "00000000-0000-0000-0000-000000000001"
const defaultAdminRoleCode = "system.admin"

func SeedPluginData(ctx context.Context, db *gorm.DB) error {
	ctxDB := db.WithContext(ctx)
	if err := seedTemplates(ctxDB); err != nil {
		return err
	}
	permIDs, err := seedCustomerPermissions(ctxDB)
	if err != nil {
		return err
	}
	if err := seedCustomerRoleBindings(ctxDB, permIDs); err != nil {
		return err
	}
	if err := seedSampleCustomers(ctxDB); err != nil {
		return err
	}
	if err := seedSportsCatalog(ctxDB); err != nil {
		return err
	}
	return nil
}

func seedTemplates(db *gorm.DB) error {
	seedTemplates := []struct {
		Name        string
		Description string
		Content     string
	}{
		{
			Name:        "欢迎模板",
			Description: "展示如何在插件中定义第一条模板记录",
			Content:     "# 欢迎使用 PowerX Base 插件\n这是一个示例模板内容，您可以根据需要修改。",
		},
		{
			Name:        "周报模板",
			Description: "帮助团队快速整理一周的工作进展",
			Content:     "## 本周进展\n- 事项 A\n- 事项 B\n\n## 下周计划\n- 计划 A\n- 计划 B",
		},
	}
	for _, tpl := range seedTemplates {
		var existing templatemodel.Template
		err := db.Where("tenant_uuid = ? AND name = ?", defaultTenantUUID, tpl.Name).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			newTpl := templatemodel.Template{
				BaseModel:   models.BaseModel{TenantUuid: defaultTenantUUID},
				Name:        tpl.Name,
				Description: tpl.Description,
				Content:     tpl.Content,
			}
			if err := db.Create(&newTpl).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			updates := map[string]interface{}{
				"description": tpl.Description,
				"content":     tpl.Content,
			}
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedCustomerPermissions(db *gorm.DB) (map[string]uint64, error) {
	ids := make(map[string]uint64)
	if db == nil || db.Migrator() == nil || !db.Migrator().HasTable(&iammodel.Permission{}) {
		return ids, nil
	}
	customerPermissions := []iammodel.Permission{
		{Resource: "customer.read", Action: "read", Description: "查看客户列表与详情"},
		{Resource: "customer.manage", Action: "write", Description: "客户批量操作与导入"},
		{Resource: "customer.export", Action: "write", Description: "导出客户/会员名单"},
		{Resource: "customer.delete", Action: "delete", Description: "删除客户记录"},
	}

	productPermissions := []iammodel.Permission{
		// SKU 基础
		{Resource: "com.powerx.plugin.ecommerce:product.sku", Action: "read", Description: "查看 SKU 列表与详情"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku", Action: "manage", Description: "创建/编辑/删除 SKU"},
		// SKU 批量任务
		{Resource: "com.powerx.plugin.ecommerce:product.sku.bulk", Action: "read", Description: "查看 SKU 批量任务"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.bulk", Action: "manage", Description: "提交/审批/重试 SKU 批量任务"},
		// 渠道映射
		{Resource: "com.powerx.plugin.ecommerce:product.sku.channel", Action: "read", Description: "查看 SKU 渠道映射"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.channel", Action: "manage", Description: "维护 SKU 渠道映射与发布"},
		// 库存
		{Resource: "com.powerx.plugin.ecommerce:product.sku.inventory", Action: "read", Description: "查看 SKU 库存与快照"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.inventory", Action: "manage", Description: "调整 SKU 库存策略"},
		// 序列号与条码
		{Resource: "com.powerx.plugin.ecommerce:product.sku.serial", Action: "read", Description: "查看 SKU 序列号/批次"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.serial", Action: "manage", Description: "录入或导出 SKU 序列号/批次"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.barcode", Action: "read", Description: "查看 SKU 条码"},
		{Resource: "com.powerx.plugin.ecommerce:product.sku.barcode", Action: "manage", Description: "生成/校验 SKU 条码"},

		// 商品类目（004-product-categories）
		{Resource: "com.powerx.plugin.ecommerce:product.category", Action: "read", Description: "查看类目树、类目列表与审计记录"},
		{Resource: "com.powerx.plugin.ecommerce:product.category", Action: "manage", Description: "创建/编辑/迁移/启停类目"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.template", Action: "read", Description: "查看类目模板、预览与历史版本"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.template", Action: "manage", Description: "创建/编辑/发布/回滚类目模板"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.mapping", Action: "read", Description: "查看渠道类目映射"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.mapping", Action: "manage", Description: "维护渠道类目映射"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.import", Action: "read", Description: "导出类目/映射 CSV"},
		{Resource: "com.powerx.plugin.ecommerce:product.category.import", Action: "manage", Description: "导入类目/映射 CSV"},
	}

	customerPermissions = append(customerPermissions, productPermissions...)

	for _, perm := range customerPermissions {
		var existing iammodel.Permission
		err := db.Where("resource = ? AND action = ?", perm.Resource, perm.Action).
			First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := db.Create(&perm).Error; err != nil {
				return nil, err
			}
			existing = perm
		case err != nil:
			return nil, err
		default:
			if perm.Description != "" && existing.Description != perm.Description {
				if err := db.Model(&existing).
					Update("description", perm.Description).Error; err != nil {
					return nil, err
				}
			}
		}
		key := perm.Resource + ":" + perm.Action
		ids[key] = existing.ID
	}
	return ids, nil
}

func seedCustomerRoleBindings(db *gorm.DB, permIDs map[string]uint64) error {
	if len(permIDs) == 0 {
		return nil
	}
	if db == nil || db.Migrator() == nil || !db.Migrator().HasTable(&iammodel.Role{}) || !db.Migrator().HasTable(&iammodel.RolePermission{}) {
		return nil
	}
	var role iammodel.Role
	err := db.Where("tenant_uuid = ? AND code = ?", defaultTenantUUID, defaultAdminRoleCode).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	for _, permID := range permIDs {
		if permID == 0 {
			continue
		}
		rp := iammodel.RolePermission{
			RoleID:       role.ID,
			PermissionID: permID,
		}
		if err := db.Where("role_id = ? AND permission_id = ?", role.ID, permID).
			FirstOrCreate(&rp).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSampleCustomers(db *gorm.DB) error {
	if db == nil || db.Migrator() == nil || !db.Migrator().HasTable(&customermodel.Customer{}) {
		return nil
	}
	samples := []struct {
		CustomerID          string
		Name                string
		Type                string
		Email               string
		Phone               string
		Source              string
		Country             string
		Region              string
		MembershipTier      string
		MembershipTierLabel string
		AccountManager      string
		Tags                []string
		Notes               string
		Status              string
	}{
		{
			CustomerID:          "seed-retail-001",
			Name:                "示例零售客户",
			Type:                "individual",
			Email:               "seed-retail-001@demo.powerx",
			Phone:               "+8613512345670",
			Source:              "seed-data",
			Country:             "中国",
			Region:              "华北",
			MembershipTier:      "silver",
			MembershipTierLabel: "Silver",
			AccountManager:      "Demo Owner",
			Tags:                []string{"demo", "retail"},
			Notes:               "演示用途客户，避免与导入模板冲突。",
			Status:              "active",
		},
		{
			CustomerID:          "seed-enterprise-001",
			Name:                "示例企业客户",
			Type:                "enterprise",
			Email:               "seed-enterprise-001@demo.powerx",
			Phone:               "+8613612345670",
			Source:              "seed-data",
			Country:             "新加坡",
			Region:              "亚太",
			MembershipTier:      "platinum",
			MembershipTierLabel: "Platinum",
			AccountManager:      "Demo Owner",
			Tags:                []string{"demo", "enterprise"},
			Notes:               "默认企业客户示例。",
			Status:              "active",
		},
	}
	for _, sample := range samples {
		var existing customermodel.Customer
		err := db.Where("tenant_uuid = ? AND customer_id = ?", defaultTenantUUID, sample.CustomerID).
			First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			entity := customermodel.Customer{
				TenantUUID:          defaultTenantUUID,
				CustomerID:          sample.CustomerID,
				Name:                sample.Name,
				Type:                sample.Type,
				Email:               sample.Email,
				Phone:               sample.Phone,
				Source:              sample.Source,
				Country:             sample.Country,
				Region:              sample.Region,
				MembershipTier:      sample.MembershipTier,
				MembershipTierLabel: sample.MembershipTierLabel,
				AccountManager:      sample.AccountManager,
				Tags:                encodeStringSlice(sample.Tags),
				Status:              sample.Status,
				Notes:               sample.Notes,
			}
			if err := db.Create(&entity).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			updates := map[string]interface{}{
				"name":                  sample.Name,
				"type":                  sample.Type,
				"email":                 sample.Email,
				"phone":                 sample.Phone,
				"source":                sample.Source,
				"country":               sample.Country,
				"region":                sample.Region,
				"membership_tier":       sample.MembershipTier,
				"membership_tier_label": sample.MembershipTierLabel,
				"account_manager":       sample.AccountManager,
				"tags":                  encodeStringSlice(sample.Tags),
				"status":                sample.Status,
				"notes":                 sample.Notes,
			}
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type categorySeedSpec struct {
	ID          string
	Code        string
	DisplayName string
	AliasSlug   string
	ParentCode  string
	SortOrder   int
	IsFeatured  bool
	ImageURL    string
}

type spuSeedSpec struct {
	ID            string
	Code          string
	Name          string
	Type          string
	CategoryCode  string
	Status        string
	DefaultLocale string
	Responsible   string
	Tags          []string
	Description   string
}

type skuSeedSpec struct {
	ID        string
	SPUCode   string
	SKUCode   string
	Status    string
	Barcode   string
	Tags      []string
	Spec      map[string]any
	SalePrice float64
	Currency  string
	MediaID   string
	ImageURL  string
}

func seedSportsCatalog(db *gorm.DB) error {
	if db == nil || db.Migrator() == nil {
		return nil
	}
	if !db.Migrator().HasTable(&productcategory.ProductCategory{}) || !db.Migrator().HasTable(&productmodel.SPU{}) {
		return nil
	}

	if err := seedSportsCategories(db); err != nil {
		return err
	}
	if err := seedSportsSPUs(db); err != nil {
		return err
	}
	if err := seedSportsSpecs(db); err != nil {
		return err
	}
	if err := seedSportsSKUs(db); err != nil {
		return err
	}
	if err := seedSportsBasePricebookItems(db); err != nil {
		return err
	}
	return nil
}

type skuSpecJSON struct {
	SpecID    string `json:"spec_id"`
	SpecName  string `json:"spec_name,omitempty"`
	ValueID   string `json:"value_id"`
	ValueName string `json:"value_name,omitempty"`
}

func seedSportsSpecs(db *gorm.DB) error {
	if db == nil || db.Migrator() == nil {
		return nil
	}
	if !db.Migrator().HasTable(&productspec.ProductSpecGroup{}) || !db.Migrator().HasTable(&productspec.ProductSpecOption{}) {
		return nil
	}
	spuIDs, err := loadSPUIdsByCode(db)
	if err != nil {
		return err
	}

	// Mirror the SKU seed spec list so that the spec definition exists before seeding SKUs.
	specs := []skuSeedSpec{
		{SPUCode: "BALL-BASKET-001", Spec: map[string]any{"size": "7", "material": "leather"}},
		{SPUCode: "BALL-BASKET-002", Spec: map[string]any{"size": "7", "material": "leather"}},
		{SPUCode: "BALL-SOCCER-001", Spec: map[string]any{"size": "5", "surface": "training"}},
		{SPUCode: "SHOE-BASKET-001", Spec: map[string]any{"size": "42", "color": "black"}},
		{SPUCode: "SHOE-BASKET-001", Spec: map[string]any{"size": "43", "color": "white"}},
		{SPUCode: "APP-JERSEY-001", Spec: map[string]any{"size": "M", "color": "blue"}},
		{SPUCode: "APP-JERSEY-001", Spec: map[string]any{"size": "L", "color": "blue"}},
	}

	preferredOrder := map[string]int{
		"color":    10,
		"size":     20,
		"material": 30,
		"surface":  40,
	}

	type groupKey struct {
		SPUID string
		Code  string
	}
	groups := make(map[groupKey]productspec.ProductSpecGroup)
	valuesByGroup := make(map[groupKey]map[string]string)

	for _, spec := range specs {
		spuID, ok := spuIDs[spec.SPUCode]
		if !ok {
			continue
		}
		for rawKey, rawVal := range spec.Spec {
			code := strings.ToLower(strings.TrimSpace(rawKey))
			valStr := strings.TrimSpace(fmt.Sprintf("%v", rawVal))
			if code == "" || valStr == "" {
				continue
			}
			key := groupKey{SPUID: spuID, Code: code}
			if _, ok := groups[key]; !ok {
				sortOrder := 100
				if v, ok := preferredOrder[code]; ok {
					sortOrder = v
				}
				groups[key] = productspec.ProductSpecGroup{
					ID:         utils.NewUUID(),
					TenantUUID: defaultTenantUUID,
					SPUID:      spuID,
					Code:       code,
					Name:       rawKey,
					SortOrder:  sortOrder,
					Required:   true,
					Status:     "active",
				}
			}
			if _, ok := valuesByGroup[key]; !ok {
				valuesByGroup[key] = map[string]string{}
			}
			valuesByGroup[key][strings.ToLower(valStr)] = valStr
		}
	}

	now := time.Now().UTC()
	for key, candidate := range groups {
		var existing productspec.ProductSpecGroup
		err := db.Where("tenant_uuid = ? AND spu_id = ? AND code = ?", defaultTenantUUID, key.SPUID, key.Code).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			candidate.CreatedAt = now
			candidate.UpdatedAt = now
			if err := db.Create(&candidate).Error; err != nil {
				return err
			}
			existing = candidate
		case err != nil:
			return err
		default:
			updates := map[string]any{
				"name":       candidate.Name,
				"sort_order": candidate.SortOrder,
				"required":   candidate.Required,
				"status":     candidate.Status,
				"updated_at": now,
			}
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Upsert options for the group.
		values := valuesByGroup[key]
		for optCode, optName := range values {
			if strings.TrimSpace(optCode) == "" || strings.TrimSpace(optName) == "" {
				continue
			}
			var existingOpt productspec.ProductSpecOption
			err := db.Where("tenant_uuid = ? AND group_id = ? AND code = ?", defaultTenantUUID, existing.ID, optCode).First(&existingOpt).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				opt := productspec.ProductSpecOption{
					ID:         utils.NewUUID(),
					TenantUUID: defaultTenantUUID,
					SPUID:      existing.SPUID,
					GroupID:    existing.ID,
					Code:       optCode,
					Name:       optName,
					SortOrder:  0,
					Status:     "active",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				if err := db.Create(&opt).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				updates := map[string]any{
					"name":       optName,
					"status":     "active",
					"updated_at": now,
				}
				if err := db.Model(&existingOpt).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func seedSportsCategories(db *gorm.DB) error {
	specs := []categorySeedSpec{
		{
			ID:          "8fc75aca-42d3-4d03-ba79-852c87c4f180",
			Code:        "basketball",
			DisplayName: "篮球",
			AliasSlug:   "basketball",
			ParentCode:  "",
			SortOrder:   10,
			IsFeatured:  true,
		},
		{
			ID:          "d4a603a5-16b4-4aa1-9b4f-9ccecd0b63aa",
			Code:        "basketball-balls",
			DisplayName: "篮球（用球）",
			AliasSlug:   "basketball-balls",
			ParentCode:  "basketball",
			SortOrder:   10,
			IsFeatured:  true,
		},
		{
			ID:          "b46f7d0d-cc6d-49f9-95c9-88b9a7da8e92",
			Code:        "basketball-jerseys",
			DisplayName: "篮球服饰",
			AliasSlug:   "jerseys",
			ParentCode:  "basketball",
			SortOrder:   20,
		},
		{
			ID:          "d9a0b8b4-7b24-45de-9f9b-3a2a8f1e25e2",
			Code:        "football",
			DisplayName: "足球",
			AliasSlug:   "football",
			ParentCode:  "",
			SortOrder:   20,
			IsFeatured:  true,
		},
		{
			ID:          "d5f5c1b5-55f4-4d7c-9e6c-6228e15a13ed",
			Code:        "football-balls",
			DisplayName: "足球（用球）",
			AliasSlug:   "football-balls",
			ParentCode:  "football",
			SortOrder:   10,
			IsFeatured:  true,
		},
		{
			ID:          "0ebcd9b4-9fe3-4f7f-8f28-0d92b6be8e5a",
			Code:        "apparel",
			DisplayName: "运动服饰",
			AliasSlug:   "apparel",
			ParentCode:  "",
			SortOrder:   30,
		},
		{
			ID:          "c6b7e3ed-9d9d-4a58-8b86-cf6d5b0e9ea1",
			Code:        "apparel-tops",
			DisplayName: "上衣",
			AliasSlug:   "tops",
			ParentCode:  "apparel",
			SortOrder:   10,
		},
		{
			ID:          "e3dd0cbe-7e12-4c79-95f2-7e699bdfb3cc",
			Code:        "apparel-pants",
			DisplayName: "裤装",
			AliasSlug:   "pants",
			ParentCode:  "apparel",
			SortOrder:   20,
		},
		{
			ID:          "0d5b9b1e-2d07-4b5a-a3f9-9d9d4c5f30ec",
			Code:        "shoes",
			DisplayName: "运动鞋",
			AliasSlug:   "shoes",
			ParentCode:  "",
			SortOrder:   40,
		},
		{
			ID:          "1f7c19a9-5e6a-4ee9-8f37-16aabca19a39",
			Code:        "shoes-basketball",
			DisplayName: "篮球鞋",
			AliasSlug:   "shoes-basketball",
			ParentCode:  "shoes",
			SortOrder:   10,
			IsFeatured:  true,
		},
		{
			ID:          "9c8d1a0e-14b3-4d5e-9a88-6f5de6cf0c1a",
			Code:        "shoes-football",
			DisplayName: "足球鞋",
			AliasSlug:   "shoes-football",
			ParentCode:  "shoes",
			SortOrder:   20,
			IsFeatured:  true,
		},
		{
			ID:          "1f9f5e88-0a2b-4e0d-96a1-2d2c3f4a5b6c",
			Code:        "magazines",
			DisplayName: "体育内容订阅",
			AliasSlug:   "magazines",
			ParentCode:  "",
			SortOrder:   50,
		},
	}

	byCode := map[string]*productcategory.ProductCategory{}
	// Preload existing ones.
	var existing []productcategory.ProductCategory
	if err := db.Where("tenant_uuid = ?", defaultTenantUUID).Find(&existing).Error; err != nil {
		return err
	}
	for i := range existing {
		c := existing[i]
		byCode[c.Code] = &c
	}

	for _, spec := range specs {
		var parent *productcategory.ProductCategory
		var parentID *string
		if spec.ParentCode != "" {
			p, ok := byCode[spec.ParentCode]
			if !ok {
				return fmt.Errorf("missing parent category %s for %s", spec.ParentCode, spec.Code)
			}
			parent = p
			parentID = &p.ID
		}
		level := 0
		parentPath := ""
		if parent != nil {
			level = parent.Level + 1
			parentPath = parent.Path
		}
		id := spec.ID
		if id == "" {
			return fmt.Errorf("category %s missing id", spec.Code)
		}
		path := fmt.Sprintf("/%s/", id)
		if parentPath != "" {
			path = parentPath + id + "/"
		}

		var category productcategory.ProductCategory
		err := db.Where("tenant_uuid = ? AND code = ?", defaultTenantUUID, spec.Code).First(&category).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			category = productcategory.ProductCategory{
				ID:          id,
				TenantUUID:  defaultTenantUUID,
				ParentID:    parentID,
				Code:        spec.Code,
				DisplayName: spec.DisplayName,
				AliasSlug:   spec.AliasSlug,
				Path:        path,
				Level:       level,
				SortOrder:   spec.SortOrder,
				Status:      productcategory.CategoryStatusEnabled,
				IsFeatured:  spec.IsFeatured,
				ImageURL:    spec.ImageURL,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			}
			category.Normalize()
			if err := db.Create(&category).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			// Use existing ID for path (cannot change primary key).
			id = category.ID
			path = fmt.Sprintf("/%s/", id)
			if parentPath != "" {
				path = parentPath + id + "/"
			}
			updates := map[string]any{
				"parent_id":    parentID,
				"display_name": spec.DisplayName,
				"alias_slug":   spec.AliasSlug,
				"path":         path,
				"level":        level,
				"sort_order":   spec.SortOrder,
				"status":       productcategory.CategoryStatusEnabled,
				"is_featured":  spec.IsFeatured,
				"image_url":    spec.ImageURL,
				"updated_at":   time.Now().UTC(),
			}
			if err := db.Model(&category).Updates(updates).Error; err != nil {
				return err
			}
		}
		// Update cache for children.
		c := category
		byCode[spec.Code] = &c
	}
	return nil
}

func seedSportsSPUs(db *gorm.DB) error {
	if !db.Migrator().HasTable(&productmodel.SPUVersion{}) || !db.Migrator().HasTable(&productmodel.SPULocale{}) {
		return nil
	}

	catByCode, err := loadCategoriesByCode(db)
	if err != nil {
		return err
	}

	specs := []spuSeedSpec{
		{
			ID:            "9d9d4c5f-30ec-45f2-bef5-34b37d01f452",
			Code:          "BALL-BASKET-001",
			Name:          "NBA 室内外训练篮球 7号",
			Type:          "one_time",
			CategoryCode:  "basketball-balls",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"basketball", "ball", "training"},
			Description:   "耐磨防滑，适合室内外训练与比赛。",
		},
		{
			ID:            "7d9e3f1b-2b1a-4c22-8b5d-2e8c1dfaa6b0",
			Code:          "BALL-SOCCER-001",
			Name:          "FIFA 训练足球 5号",
			Type:          "one_time",
			CategoryCode:  "football-balls",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"football", "ball", "training"},
			Description:   "标准 5 号训练球，脚感舒适耐用。",
		},
		{
			ID:            "2a2c2c2c-9f7e-4f0d-8c3e-0b9a4d7e1c20",
			Code:          "SHOE-BASKET-001",
			Name:          "篮球鞋 Pro Jump 高帮",
			Type:          "one_time",
			CategoryCode:  "shoes-basketball",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"basketball", "shoes", "high-top"},
			Description:   "高帮支撑，缓震回弹，适合强对抗。",
		},
		{
			ID:            "3b3c3c3c-6e12-4a9a-9f2a-1b2c3d4e5f60",
			Code:          "SHOE-SOCCER-001",
			Name:          "足球鞋 Speed Cleats 碎钉",
			Type:          "one_time",
			CategoryCode:  "shoes-football",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"football", "shoes", "cleats"},
			Description:   "轻量贴合，抓地稳定，适合人草场地。",
		},
		{
			ID:            "4c4d4d4d-7e12-4c79-95f2-7e699bdfb3dd",
			Code:          "APP-JERSEY-001",
			Name:          "篮球背心 速干球衣",
			Type:          "one_time",
			CategoryCode:  "basketball-jerseys",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"basketball", "apparel", "quick-dry"},
			Description:   "速干透气，训练与比赛皆可。",
		},
		{
			ID:            "5d5e5e5e-7a12-4c79-95f2-7e699bdfb3ee",
			Code:          "APP-TRACK-001",
			Name:          "运动长裤 训练款",
			Type:          "one_time",
			CategoryCode:  "apparel-pants",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "ops-01",
			Tags:          []string{"apparel", "training", "pants"},
			Description:   "弹力面料，日常训练/跑步皆适用。",
		},
		{
			ID:            "6e6f6f6f-1234-4f0d-8c3e-0b9a4d7e1c21",
			Code:          "SUB-MAG-SPORTS-001",
			Name:          "体育月刊",
			Type:          "subscription",
			CategoryCode:  "magazines",
			Status:        "published",
			DefaultLocale: "zh-CN",
			Responsible:   "editor-01",
			Tags:          []string{"subscription", "content"},
			Description:   "每月一期：篮球/足球/装备评测/训练方法。",
		},
	}

	for _, spec := range specs {
		cat, ok := catByCode[spec.CategoryCode]
		if !ok {
			return fmt.Errorf("missing category %s for spu %s", spec.CategoryCode, spec.Code)
		}
		if err := upsertSPUWithVersionAndLocale(db, spec, cat); err != nil {
			return err
		}
	}

	// Subscription plans for 体育月刊.
	if db.Migrator().HasTable(&productmodel.SubscriptionPlan{}) {
		monthlyID := "9c1c0b9a-1c20-4d7e-8c3e-0b9a4d7e1c22"
		yearlyID := "a1b2c3d4-5f60-4a9a-9f2a-1b2c3d4e5f61"
		subSpuID := specs[len(specs)-1].ID
		plans := []productmodel.SubscriptionPlan{
			{
				ID:           monthlyID,
				TenantUUID:   defaultTenantUUID,
				SPUID:        subSpuID,
				PlanCode:     "monthly",
				Name:         "月订阅",
				BillingCycle: "monthly",
				BillingValue: 0,
				Price:        29.9,
				Currency:     "CNY",
				TrialDays:    7,
				AutoRenew:    true,
				CancelPolicy: "anytime",
				EffectScope:  "new_only",
				Status:       "active",
				Metadata:     datatypes.JSON([]byte(`{}`)),
			},
			{
				ID:           yearlyID,
				TenantUUID:   defaultTenantUUID,
				SPUID:        subSpuID,
				PlanCode:     "yearly",
				Name:         "年订阅",
				BillingCycle: "yearly",
				BillingValue: 0,
				Price:        299.0,
				Currency:     "CNY",
				TrialDays:    14,
				AutoRenew:    true,
				CancelPolicy: "anytime",
				EffectScope:  "new_only",
				Status:       "active",
				Metadata:     datatypes.JSON([]byte(`{}`)),
			},
		}
		for _, plan := range plans {
			var existingPlan productmodel.SubscriptionPlan
			err := db.Where("tenant_uuid = ? AND spu_id = ? AND plan_code = ?", defaultTenantUUID, plan.SPUID, plan.PlanCode).
				First(&existingPlan).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				now := time.Now().UTC()
				plan.CreatedAt = now
				plan.UpdatedAt = now
				if err := db.Create(&plan).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				updates := map[string]any{
					"name":          plan.Name,
					"billing_cycle": plan.BillingCycle,
					"billing_value": plan.BillingValue,
					"price":         plan.Price,
					"currency":      plan.Currency,
					"trial_days":    plan.TrialDays,
					"auto_renew":    plan.AutoRenew,
					"cancel_policy": plan.CancelPolicy,
					"effect_scope":  plan.EffectScope,
					"status":        plan.Status,
					"metadata":      plan.Metadata,
					"updated_at":    time.Now().UTC(),
				}
				if err := db.Model(&existingPlan).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func seedSportsSKUs(db *gorm.DB) error {
	if !db.Migrator().HasTable(&productsku.ProductSKU{}) {
		return nil
	}
	hasMedia := db.Migrator().HasTable(&productsku.ProductSKUMedia{})
	hasSpecTables := db.Migrator().HasTable(&productspec.ProductSpecGroup{}) && db.Migrator().HasTable(&productspec.ProductSpecOption{})
	hasSkuAttributes := db.Migrator().HasTable(&productsku.ProductSKUAttribute{})
	spuIDs, err := loadSPUIdsByCode(db)
	if err != nil {
		return err
	}
	specs := []skuSeedSpec{
		{
			ID:        "8a8a8a8a-1111-4f0d-8c3e-0b9a4d7e1c30",
			SPUCode:   "BALL-BASKET-001",
			SKUCode:   "BALL-BASKET-001-STD",
			Status:    "published",
			Barcode:   "6900000000001",
			Tags:      []string{"basketball", "ball"},
			Spec:      map[string]any{"size": "7", "material": "composite"},
			SalePrice: 199,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-1111-4f0d-8c3e-0b9a4d7e1c30",
		},
		{
			ID:        "8a8a8a8a-2222-4f0d-8c3e-0b9a4d7e1c31",
			SPUCode:   "BALL-BASKET-001",
			SKUCode:   "BALL-BASKET-001-PRO",
			Status:    "published",
			Barcode:   "6900000000002",
			Tags:      []string{"basketball", "ball", "premium"},
			Spec:      map[string]any{"size": "7", "material": "leather"},
			SalePrice: 399,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-2222-4f0d-8c3e-0b9a4d7e1c31",
		},
		{
			ID:        "8a8a8a8a-3333-4f0d-8c3e-0b9a4d7e1c32",
			SPUCode:   "BALL-SOCCER-001",
			SKUCode:   "BALL-SOCCER-001-STD",
			Status:    "published",
			Barcode:   "6900000000003",
			Tags:      []string{"football", "ball"},
			Spec:      map[string]any{"size": "5", "surface": "training"},
			SalePrice: 169,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-3333-4f0d-8c3e-0b9a4d7e1c32",
		},
		{
			ID:        "8a8a8a8a-4444-4f0d-8c3e-0b9a4d7e1c33",
			SPUCode:   "SHOE-BASKET-001",
			SKUCode:   "SHOE-BASKET-001-42",
			Status:    "published",
			Barcode:   "6900000000004",
			Tags:      []string{"basketball", "shoes"},
			Spec:      map[string]any{"size": "42", "color": "black"},
			SalePrice: 699,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-4444-4f0d-8c3e-0b9a4d7e1c33",
		},
		{
			ID:        "8a8a8a8a-5555-4f0d-8c3e-0b9a4d7e1c34",
			SPUCode:   "SHOE-BASKET-001",
			SKUCode:   "SHOE-BASKET-001-43",
			Status:    "published",
			Barcode:   "6900000000005",
			Tags:      []string{"basketball", "shoes"},
			Spec:      map[string]any{"size": "43", "color": "white"},
			SalePrice: 699,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-5555-4f0d-8c3e-0b9a4d7e1c34",
		},
		{
			ID:        "8a8a8a8a-6666-4f0d-8c3e-0b9a4d7e1c35",
			SPUCode:   "APP-JERSEY-001",
			SKUCode:   "APP-JERSEY-001-M",
			Status:    "published",
			Barcode:   "6900000000006",
			Tags:      []string{"basketball", "apparel"},
			Spec:      map[string]any{"size": "M", "color": "blue"},
			SalePrice: 299,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-6666-4f0d-8c3e-0b9a4d7e1c35",
		},
		{
			ID:        "8a8a8a8a-7777-4f0d-8c3e-0b9a4d7e1c36",
			SPUCode:   "APP-JERSEY-001",
			SKUCode:   "APP-JERSEY-001-L",
			Status:    "published",
			Barcode:   "6900000000007",
			Tags:      []string{"basketball", "apparel"},
			Spec:      map[string]any{"size": "L", "color": "blue"},
			SalePrice: 299,
			Currency:  "CNY",
			MediaID:   "9a8a8a8a-7777-4f0d-8c3e-0b9a4d7e1c36",
		},
	}
	now := time.Now().UTC()
	for _, spec := range specs {
		spuID, ok := spuIDs[spec.SPUCode]
		if !ok {
			return fmt.Errorf("missing spu %s for sku %s", spec.SPUCode, spec.SKUCode)
		}
		specBytes, _ := json.Marshal(spec.Spec)
		specSignature := ""
		if hasSpecTables {
			var groups []productspec.ProductSpecGroup
			if err := db.Where("tenant_uuid = ? AND spu_id = ? AND deleted_at IS NULL AND status = 'active'", defaultTenantUUID, spuID).
				Order("sort_order ASC, code ASC").
				Find(&groups).Error; err != nil {
				return err
			}
			groupByCode := make(map[string]productspec.ProductSpecGroup, len(groups))
			groupIDs := make([]string, 0, len(groups))
			for _, g := range groups {
				groupByCode[strings.ToLower(strings.TrimSpace(g.Code))] = g
				groupIDs = append(groupIDs, g.ID)
			}
			var options []productspec.ProductSpecOption
			if len(groupIDs) > 0 {
				if err := db.Where("tenant_uuid = ? AND spu_id = ? AND group_id IN ? AND deleted_at IS NULL AND status = 'active'",
					defaultTenantUUID, spuID, groupIDs).
					Find(&options).Error; err != nil {
					return err
				}
			}
			optionByGroupAndCode := make(map[string]productspec.ProductSpecOption, len(options))
			for _, o := range options {
				key := o.GroupID + "::" + strings.ToLower(strings.TrimSpace(o.Code))
				optionByGroupAndCode[key] = o
			}

			specPairs := make([]skuSpecJSON, 0, len(spec.Spec))
			signParts := make([]string, 0, len(spec.Spec))
			type signPart struct {
				SortOrder int
				GroupCode string
				Text      string
			}
			sorted := make([]signPart, 0, len(spec.Spec))
			for rawKey, rawVal := range spec.Spec {
				gc := strings.ToLower(strings.TrimSpace(rawKey))
				valStr := strings.TrimSpace(fmt.Sprintf("%v", rawVal))
				oc := strings.ToLower(valStr)
				if gc == "" || oc == "" {
					continue
				}
				g, ok := groupByCode[gc]
				if !ok {
					continue
				}
				o, ok := optionByGroupAndCode[g.ID+"::"+oc]
				if !ok {
					continue
				}
				specPairs = append(specPairs, skuSpecJSON{
					SpecID:    g.ID,
					SpecName:  g.Name,
					ValueID:   o.ID,
					ValueName: o.Name,
				})
				sorted = append(sorted, signPart{
					SortOrder: g.SortOrder,
					GroupCode: g.Code,
					Text:      strings.TrimSpace(g.Code) + "=" + strings.TrimSpace(o.Code),
				})
			}
			sort.Slice(sorted, func(i, j int) bool {
				if sorted[i].SortOrder == sorted[j].SortOrder {
					return sorted[i].GroupCode < sorted[j].GroupCode
				}
				return sorted[i].SortOrder < sorted[j].SortOrder
			})
			for _, p := range sorted {
				signParts = append(signParts, p.Text)
			}
			specSignature = strings.Join(signParts, "|")
			if len(specPairs) > 0 {
				specBytes, _ = json.Marshal(specPairs)
			}
		}

		defaultValuesBytes, _ := json.Marshal(map[string]any{
			"sale_price": spec.SalePrice,
			"currency":   strings.TrimSpace(spec.Currency),
		})
		var sku productsku.ProductSKU
		err := db.Where("tenant_uuid = ? AND sku_code = ?", defaultTenantUUID, spec.SKUCode).First(&sku).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			sku = productsku.ProductSKU{
				ID:            spec.ID,
				TenantUUID:    defaultTenantUUID,
				SPUID:         spuID,
				SKUCode:       spec.SKUCode,
				Barcode:       spec.Barcode,
				Status:        spec.Status,
				SpecValues:    datatypes.JSON(specBytes),
				SpecSignature: strings.TrimSpace(specSignature),
				DefaultValues: datatypes.JSON(defaultValuesBytes),
				Tags:          pq.StringArray(spec.Tags),
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := db.Create(&sku).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			updates := map[string]any{
				"spu_id":         spuID,
				"barcode":        spec.Barcode,
				"status":         spec.Status,
				"spec_values":    datatypes.JSON(specBytes),
				"spec_signature": strings.TrimSpace(specSignature),
				"default_values": datatypes.JSON(defaultValuesBytes),
				"tags":           pq.StringArray(spec.Tags),
				"updated_at":     now,
			}
			if err := db.Model(&sku).Updates(updates).Error; err != nil {
				return err
			}
		}

		if hasSkuAttributes && hasSpecTables {
			// Keep product_sku_attributes in sync for filtering.
			var pairs []skuSpecJSON
			_ = json.Unmarshal(specBytes, &pairs)
			for idx, p := range pairs {
				if strings.TrimSpace(p.SpecID) == "" || strings.TrimSpace(p.ValueID) == "" {
					continue
				}
				var existingAttr productsku.ProductSKUAttribute
				err := db.Where("tenant_uuid = ? AND sku_id = ? AND spec_id = ? AND spec_value_id = ?",
					defaultTenantUUID, spec.ID, p.SpecID, p.ValueID).
					First(&existingAttr).Error
				switch {
				case errors.Is(err, gorm.ErrRecordNotFound):
					attr := productsku.ProductSKUAttribute{
						ID:           utils.NewUUID(),
						TenantUUID:   defaultTenantUUID,
						SKUId:        spec.ID,
						SpecID:       p.SpecID,
						SpecValueID:  p.ValueID,
						SpecName:     p.SpecName,
						ValueName:    p.ValueName,
						DisplayOrder: idx,
						CreatedAt:    now,
						UpdatedAt:    now,
					}
					if err := db.Create(&attr).Error; err != nil {
						return err
					}
				case err != nil:
					return err
				default:
					updates := map[string]any{
						"spec_name":     p.SpecName,
						"value_name":    p.ValueName,
						"display_order": idx,
						"updated_at":    now,
					}
					if err := db.Model(&existingAttr).Updates(updates).Error; err != nil {
						return err
					}
				}
			}
		}

		if hasMedia {
			url := strings.TrimSpace(spec.ImageURL)
			if url == "" {
				url = fmt.Sprintf("https://picsum.photos/seed/%s/600/750", spec.SKUCode)
			}
			var media productsku.ProductSKUMedia
			err := db.Where("tenant_uuid = ? AND sku_id = ? AND is_primary = ?", defaultTenantUUID, spec.ID, true).
				First(&media).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				media = productsku.ProductSKUMedia{
					ID:         spec.MediaID,
					TenantUUID: defaultTenantUUID,
					SKUId:      spec.ID,
					MediaType:  "image",
					URL:        url,
					IsPrimary:  true,
					SortOrder:  0,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				if err := db.Create(&media).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			default:
				updates := map[string]any{
					"url":        url,
					"media_type": "image",
					"sort_order": 0,
					"updated_at": now,
				}
				if err := db.Model(&media).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func seedSportsBasePricebookItems(db *gorm.DB) error {
	if db == nil || db.Migrator() == nil {
		return nil
	}
	if !db.Migrator().HasTable(&productsku.ProductSKU{}) {
		return nil
	}
	if !db.Migrator().HasTable(&pricingmodel.Pricebook{}) ||
		!db.Migrator().HasTable(&pricingmodel.PricebookVersion{}) ||
		!db.Migrator().HasTable(&pricingmodel.PricebookItem{}) {
		return nil
	}

	var skus []productsku.ProductSKU
	if err := db.Where("tenant_uuid = ?", defaultTenantUUID).Order("created_at asc").Find(&skus).Error; err != nil {
		return err
	}
	if len(skus) == 0 {
		return nil
	}

	tenantCtx := authx.ContextWithTenantUUID(context.Background(), defaultTenantUUID)
	deps := &app.Deps{DB: db}
	pbSvc := pricingsvc.NewPricebookService(deps)
	verSvc := pricingsvc.NewVersionService(deps)
	itemSvc := pricingsvc.NewItemService(deps)

	currencyHint := "CNY"
	if cur := strings.TrimSpace(extractSKUCurrency(skus[0])); cur != "" {
		currencyHint = cur
	}

	pb, err := pbSvc.EnsureBasePricebook(tenantCtx, currencyHint, "seed")
	if err != nil {
		return err
	}

	draft, err := verSvc.CreateDraft(tenantCtx, pricingsvc.CreateVersionInput{
		PricebookID: pb.ID,
		Actor:       "seed",
	})
	if err != nil {
		return err
	}

	toMinor := func(price float64) *int64 {
		if price <= 0 {
			return nil
		}
		minor := int64(math.Round(price * 100))
		if minor < 0 {
			return nil
		}
		return &minor
	}

	buildSpecDisplay := func(raw datatypes.JSON) string {
		if len(raw) == 0 {
			return ""
		}
		var pairs []skuSpecJSON
		if err := json.Unmarshal(raw, &pairs); err != nil {
			return ""
		}
		parts := make([]string, 0, len(pairs))
		for _, p := range pairs {
			k := strings.TrimSpace(p.SpecName)
			v := strings.TrimSpace(p.ValueName)
			if k == "" || v == "" {
				continue
			}
			parts = append(parts, k+"="+v)
		}
		return strings.Join(parts, " | ")
	}

	falsePtr := func() *bool {
		v := false
		return &v
	}

	items := make([]pricingsvc.ItemInput, 0, len(skus))
	for _, sku := range skus {
		salePrice := extractSKUSalePrice(sku)
		if salePrice <= 0 {
			continue
		}
		basePrice := salePrice * 1.1
		meta := map[string]any{
			"sku_code":        sku.SKUCode,
			"spec_signature":  strings.TrimSpace(sku.SpecSignature),
			"spec_display":    buildSpecDisplay(sku.SpecValues),
			"seed_sale_price": salePrice,
		}
		items = append(items, pricingsvc.ItemInput{
			SKUID:           sku.ID,
			BaseAmountMinor: toMinor(basePrice),
			SaleAmountMinor: toMinor(salePrice),
			TaxIncluded:     falsePtr(),
			Meta:            meta,
		})
	}

	const batchSize = 200
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		if _, err := itemSvc.UpsertItems(tenantCtx, pricingsvc.UpsertItemsInput{
			PricebookID: pb.ID,
			VersionID:   draft.ID,
			Items:       items[i:end],
			Actor:       "seed",
		}); err != nil {
			return err
		}
	}

	note := "seed base pricebook items"
	if _, err := verSvc.Publish(tenantCtx, pricingsvc.PublishVersionInput{
		PricebookID: pb.ID,
		VersionID:   draft.ID,
		Note:        &note,
		Actor:       "seed",
	}); err != nil {
		return err
	}
	return nil
}

func extractSKUSalePrice(sku productsku.ProductSKU) float64 {
	if len(sku.DefaultValues) == 0 {
		return 0
	}
	var m map[string]any
	if err := json.Unmarshal(sku.DefaultValues, &m); err != nil {
		return 0
	}
	v, ok := m["sale_price"]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	case string:
		f, _ := json.Number(strings.TrimSpace(n)).Float64()
		return f
	default:
		f, _ := json.Number(strings.TrimSpace(fmt.Sprintf("%v", v))).Float64()
		return f
	}
}

func extractSKUCurrency(sku productsku.ProductSKU) string {
	if len(sku.DefaultValues) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(sku.DefaultValues, &m); err != nil {
		return ""
	}
	if v, ok := m["currency"]; ok && v != nil {
		return strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", v)))
	}
	return ""
}

func loadCategoriesByCode(db *gorm.DB) (map[string]productcategory.ProductCategory, error) {
	out := make(map[string]productcategory.ProductCategory)
	var cats []productcategory.ProductCategory
	if err := db.Where("tenant_uuid = ?", defaultTenantUUID).Find(&cats).Error; err != nil {
		return nil, err
	}
	for _, cat := range cats {
		out[cat.Code] = cat
	}
	return out, nil
}

func loadSPUIdsByCode(db *gorm.DB) (map[string]string, error) {
	out := make(map[string]string)
	var spus []productmodel.SPU
	if err := db.Where("tenant_uuid = ?", defaultTenantUUID).Find(&spus).Error; err != nil {
		return nil, err
	}
	for _, spu := range spus {
		out[spu.Code] = spu.ID
	}
	return out, nil
}

func upsertSPUWithVersionAndLocale(db *gorm.DB, spec spuSeedSpec, cat productcategory.ProductCategory) error {
	now := time.Now().UTC()
	var spu productmodel.SPU
	err := db.Where("tenant_uuid = ? AND code = ?", defaultTenantUUID, spec.Code).First(&spu).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		spu = productmodel.SPU{
			ID:              spec.ID,
			TenantUUID:      defaultTenantUUID,
			Code:            spec.Code,
			Name:            spec.Name,
			Type:            spec.Type,
			CategoryID:      cat.ID,
			CategoryPath:    cat.Path,
			DefaultLocale:   spec.DefaultLocale,
			Status:          spec.Status,
			Tags:            pq.StringArray(spec.Tags),
			ResponsibleUser: spec.Responsible,
			ChannelsSummary: datatypes.JSON([]byte(`{}`)),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := db.Create(&spu).Error; err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		updates := map[string]any{
			"name":             spec.Name,
			"type":             spec.Type,
			"category_id":      cat.ID,
			"category_path":    cat.Path,
			"default_locale":   spec.DefaultLocale,
			"status":           spec.Status,
			"tags":             pq.StringArray(spec.Tags),
			"responsible_user": spec.Responsible,
			"updated_at":       now,
		}
		if err := db.Model(&spu).Updates(updates).Error; err != nil {
			return err
		}
	}

	// Version + current_version_id
	payload := map[string]any{
		"version": "seed",
		"input": map[string]any{
			"code":            spec.Code,
			"name":            spec.Name,
			"type":            spec.Type,
			"categoryId":      cat.ID,
			"categoryPath":    cat.Path,
			"defaultLocale":   spec.DefaultLocale,
			"tags":            spec.Tags,
			"responsibleUser": spec.Responsible,
			"locales": []map[string]any{
				{"locale": "zh-CN", "title": spec.Name, "description": spec.Description},
			},
			"attributes": map[string]any{},
		},
		"skus": []any{},
	}
	body, _ := json.Marshal(payload)
	var version productmodel.SPUVersion
	err = db.Where("tenant_uuid = ? AND spu_id = ? AND version_number = ?", defaultTenantUUID, spu.ID, 1).First(&version).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		version = productmodel.SPUVersion{
			ID:            spu.ID, // keep uuid shape
			TenantUUID:    defaultTenantUUID,
			SPUID:         spu.ID,
			VersionNumber: 1,
			Status:        spec.Status,
			Payload:       datatypes.JSON(body),
			SubmittedBy:   spec.Responsible,
			ApprovedAt:    &now,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := db.Create(&version).Error; err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		updates := map[string]any{
			"status":      spec.Status,
			"payload":     datatypes.JSON(body),
			"approved_at": now,
			"updated_at":  now,
		}
		if err := db.Model(&version).Updates(updates).Error; err != nil {
			return err
		}
	}
	if spu.CurrentVersionID == nil || *spu.CurrentVersionID != version.ID {
		if err := db.Model(&spu).Updates(map[string]any{"current_version_id": version.ID}).Error; err != nil {
			return err
		}
	}

	// Locale (zh-CN)
	var locale productmodel.SPULocale
	err = db.Where("tenant_uuid = ? AND spu_id = ? AND locale = ?", defaultTenantUUID, spu.ID, "zh-CN").First(&locale).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		locale = productmodel.SPULocale{
			ID:          spu.ID,
			TenantUUID:  defaultTenantUUID,
			SPUID:       spu.ID,
			Locale:      "zh-CN",
			Title:       spec.Name,
			Description: spec.Description,
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Create(&locale).Error; err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		updates := map[string]any{
			"title":       spec.Name,
			"description": spec.Description,
			"updated_at":  now,
		}
		if err := db.Model(&locale).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func encodeStringSlice(values []string) datatypes.JSON {
	if len(values) == 0 {
		return datatypes.JSON([]byte("[]"))
	}
	data, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(data)
}

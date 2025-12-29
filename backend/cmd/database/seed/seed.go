package seed

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	iammodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	templatemodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/template"
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

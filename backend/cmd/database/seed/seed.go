package seed

import (
	"context"
	"errors"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	iammodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	templatemodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/template"
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
		ids[perm.Resource] = existing.ID
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

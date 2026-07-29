package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	identitymodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	pricingsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/pricing"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

type basePricebookUpsertOptions struct {
	Currency         string
	DryRun           bool
	TenantUUIDs      string
	DiscoverFromIAM  bool
	DiscoverFromData bool
}

func upsertBasePricebooks(ctx context.Context, db *gorm.DB, opt basePricebookUpsertOptions) error {
	if db == nil {
		return errors.New("db is nil")
	}
	ctx = ensureContext(ctx)

	currency := strings.TrimSpace(opt.Currency)
	if currency == "" {
		currency = "USD"
	}

	tenantSet := map[string]struct{}{}
	addTenants(tenantSet, strings.Split(opt.TenantUUIDs, ",")...)

	if opt.DiscoverFromIAM {
		_ = discoverTenantUUIDsFromIAM(db, tenantSet)
	}
	if opt.DiscoverFromData {
		_ = discoverTenantUUIDsFromTenantColumns(db, tenantSet)
	}

	tenants := make([]string, 0, len(tenantSet))
	for t := range tenantSet {
		tenants = append(tenants, t)
	}
	sort.Strings(tenants)
	if len(tenants) == 0 {
		return errors.New("no tenant UUIDs found (use -tenant-uuids or enable discovery sources)")
	}

	deps := &app.Deps{Ctx: ctx, DB: db}
	pbSvc := pricingsvc.NewPricebookService(deps)

	var okCount, skipCount, errCount int
	for _, tenantUUID := range tenants {
		tenantCtx := authx.ContextWithTenantUUID(ctx, tenantUUID)
		if opt.DryRun {
			action, err := planBaseUpsert(tenantCtx, db, currency)
			if err != nil {
				errCount++
				log.Printf("[pricing-base-upsert] tenant=%s error=%v", tenantUUID, err)
				continue
			}
			if action == "" {
				skipCount++
				log.Printf("[pricing-base-upsert] tenant=%s ok (no-op)", tenantUUID)
				continue
			}
			okCount++
			log.Printf("[pricing-base-upsert] tenant=%s %s", tenantUUID, action)
			continue
		}

		if _, err := pbSvc.EnsureBasePricebook(tenantCtx, currency, "system"); err != nil {
			errCount++
			log.Printf("[pricing-base-upsert] tenant=%s error=%v", tenantUUID, err)
			continue
		}
		okCount++
		log.Printf("[pricing-base-upsert] tenant=%s ok", tenantUUID)
	}
	log.Printf("[pricing-base-upsert] done tenants=%d ok=%d no-op=%d errors=%d dry_run=%v",
		len(tenants), okCount, skipCount, errCount, opt.DryRun,
	)
	if errCount > 0 {
		return fmt.Errorf("completed with %d errors", errCount)
	}
	return nil
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func addTenants(set map[string]struct{}, raw ...string) {
	for _, v := range raw {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		set[v] = struct{}{}
	}
}

func discoverTenantUUIDsFromIAM(db *gorm.DB, set map[string]struct{}) error {
	if db == nil || db.Migrator() == nil {
		return errors.New("db migrator not available")
	}
	if !db.Migrator().HasTable(&identitymodel.Tenant{}) {
		return nil
	}
	var tenants []identitymodel.Tenant
	if err := db.Model(&identitymodel.Tenant{}).Where("deleted_at IS NULL").Find(&tenants).Error; err != nil {
		return err
	}
	for _, t := range tenants {
		addTenants(set, t.Key)
	}
	return nil
}

func discoverTenantUUIDsFromTenantColumns(db *gorm.DB, set map[string]struct{}) error {
	if db == nil {
		return errors.New("db is nil")
	}
	type source struct {
		table string
		col   string
	}
	// 仅挑选最可能覆盖租户的表；不存在则跳过。
	sources := []source{
		{table: pricingModel.Pricebook{}.TableName(), col: "tenant_uuid"},
		{table: pricingModel.PricebookVersion{}.TableName(), col: "tenant_uuid"},
		{table: pricingModel.PricebookItem{}.TableName(), col: "tenant_uuid"},
	}
	for _, s := range sources {
		var rows []string
		q := fmt.Sprintf(`SELECT DISTINCT %s FROM %s WHERE %s IS NOT NULL AND TRIM(%s) <> ''`, s.col, s.table, s.col, s.col)
		if err := db.Raw(q).Scan(&rows).Error; err != nil {
			continue
		}
		addTenants(set, rows...)
	}
	return nil
}

func planBaseUpsert(ctx context.Context, db *gorm.DB, currency string) (string, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return "", err
	}
	var pb pricingModel.Pricebook
	res := db.WithContext(ctx).
		Where("tenant_uuid = ? AND code = ?", tenantUUID, pricingsvc.BasePricebookCode).
		First(&pb)
	switch {
	case res.Error == nil:
		needs := []string{}
		if pb.Status != "active" {
			needs = append(needs, "activate")
		}
		var activeCount int64
		if err := db.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
			Where("tenant_uuid = ? AND pricebook_id = ? AND state = ?", tenantUUID, pb.ID, "active").
			Count(&activeCount).Error; err != nil {
			return "", err
		}
		if activeCount == 0 {
			needs = append(needs, "ensure_active_version")
		}
		if len(needs) == 0 {
			return "", nil
		}
		return "would " + strings.Join(needs, ","), nil
	case errors.Is(res.Error, gorm.ErrRecordNotFound):
		return fmt.Sprintf("would create base (currency=%s)", currency), nil
	default:
		return "", res.Error
	}
}

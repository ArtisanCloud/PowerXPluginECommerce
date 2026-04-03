package migrate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/cmd/database/migrate/migrations"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	adminconsoleModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/admin_console"
	iammodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	marketplaceModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/marketplace"
	operationsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/operations"
	runtimeOpsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/runtime_ops"
	securityModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/security"
	templateModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/template"
	toolgrantModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/tool_grant"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

var (
	corePluginTables = []interface{}{
		&models.PluginCredential{},
		&models.PluginTenantExt{},
		&templateModel.Template{},
	}

	marketplaceTables = []interface{}{
		&marketplaceModel.Listing{},
		&marketplaceModel.ListingAsset{},
		&marketplaceModel.ListingVersion{},
		&marketplaceModel.ChecklistRun{},
		&marketplaceModel.ChecklistItem{},
		&marketplaceModel.PricingPlan{},
		&marketplaceModel.PlanTier{},
		&marketplaceModel.License{},
		&marketplaceModel.LicenseEvent{},
		&marketplaceModel.TaxTransaction{},
	}

	runtimeOpsTables = []interface{}{
		&runtimeOpsModel.MCPSession{},
		&runtimeOpsModel.RuntimeAuditEvent{},
		&runtimeOpsModel.QuotaLedger{},
		&runtimeOpsModel.MarketplaceOverage{},
	}

	operationsTables = []interface{}{
		&operationsModel.SupportChannel{},
		&operationsModel.SupportTicket{},
		&operationsModel.SupportTicketEvent{},
		&operationsModel.ReadinessChecklistItem{},
		&operationsModel.SLAProfile{},
		&operationsModel.SLAAdjustment{},
		&operationsModel.Incident{},
		&operationsModel.IncidentTimelineEntry{},
		&operationsModel.IncidentChecklistItem{},
	}

	securityTables = []interface{}{
		&securityModel.BaselineChecklist{},
		&securityModel.AuditReport{},
	}

	toolGrantTables = []interface{}{
		&toolgrantModel.Revocation{},
		&toolgrantModel.UsageEvent{},
	}

	adminConsoleTables = []interface{}{
		&adminconsoleModel.AuditEvent{},
		&adminconsoleModel.ConfigChange{},
		&adminconsoleModel.JobRun{},
	}
)

var businessTables = func() []interface{} {
	tables := append([]interface{}{}, corePluginTables...)
	tables = append(tables, migrations.CustomerOpsCustomerTables...)
	tables = append(tables, migrations.CustomerAddressTables...)
	tables = append(tables, migrations.IntegrationTables...)
	tables = append(tables, migrations.ProductSPUTables...)
	tables = append(tables, migrations.ProductCategoryTables...)
	tables = append(tables, migrations.ChannelMasterTables...)
	tables = append(tables, migrations.PricingPricebookTables...)
	tables = append(tables, migrations.CouponTables...)
	tables = append(tables, migrations.OrderTables...)
	tables = append(tables, migrations.CartTables...)
	tables = append(tables, migrations.PaymentTables...)
	tables = append(tables, migrations.MembershipEntitlementTables...)
	tables = append(tables, migrations.FulfillmentLogisticsTables...)
	tables = append(tables, migrations.AfterSalesRMATables...)
	tables = append(tables, migrations.SubscriptionReconciliationTables...)
	tables = append(tables, marketplaceTables...)
	tables = append(tables, runtimeOpsTables...)
	tables = append(tables, operationsTables...)
	tables = append(tables, securityTables...)
	tables = append(tables, toolGrantTables...)
	tables = append(tables, adminConsoleTables...)
	tables = append(tables, migrations.ProductSkuTables...) // 002-product-sku-management
	return tables
}()

var iamTables = []interface{}{
	&iammodel.Tenant{},
	&iammodel.User{},
	&iammodel.Member{},
	&iammodel.Role{},
	&iammodel.Permission{},
	&iammodel.Department{},
	&iammodel.MemberRole{},
	&iammodel.RolePermission{},
	&iammodel.RefreshToken{},
}

// MigratePluginModels 只做 AutoMigrate（最小实现）
func MigratePluginModels(ctx context.Context, db *gorm.DB, includeIAM bool) error {
	if db == nil {
		return nil
	}
	tables := append([]interface{}{}, businessTables...)
	if isSQLite(db) {
		tables = filterSQLiteIncompatibleTables(tables)
	}
	if includeIAM {
		tables = append(tables, iamTables...)
	}
	if len(tables) == 0 {
		return nil
	}
	if err := safeAutoMigrate(ctx, db, tables); err != nil {
		return err
	}
	if err := ensureChannelMasterUniqueIndex(ctx, db); err != nil {
		return err
	}
	if err := ensureChannelMetricUniqueIndex(ctx, db); err != nil {
		return err
	}
	if err := ensurePaymentProviderUniqueIndex(ctx, db); err != nil {
		return err
	}
	if err := ensureCustomerIdentityUniqueIndex(ctx, db); err != nil {
		return err
	}
	if err := ensureChannelRLSPolicies(ctx, db); err != nil {
		return err
	}
	if err := ensurePricingRLSPolicies(ctx, db); err != nil {
		return err
	}
	if err := ensureCouponIndexes(ctx, db); err != nil {
		return err
	}
	return ensureOrderRLSPolicies(ctx, db)
}

func safeAutoMigrate(ctx context.Context, db *gorm.DB, tables []interface{}) error {
	for _, tbl := range tables {
		if err := migrateWithTolerance(ctx, db, tbl); err != nil {
			return fmt.Errorf("auto migrate table %T failed: %w", tbl, err)
		}
	}
	return nil
}

func migrateWithTolerance(ctx context.Context, db *gorm.DB, table interface{}) error {
	const maxRetries = 5
	for attempts := 0; attempts < maxRetries; attempts++ {
		if err := db.WithContext(ctx).AutoMigrate(table); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == duplicateObjectCode {
				log.Printf("[migrate] duplicate constraint for %T, skipping: %s", table, pgErr.ConstraintName)
				return nil
			}
			// 部分包装错误解析不到 pgErr，但文本包含 already exists/constraint，直接跳过以保证幂等
			if strings.Contains(err.Error(), "already exists") && strings.Contains(err.Error(), "constraint") {
				log.Printf("[migrate] duplicate constraint (fallback) for %T, skipping: %v", table, err)
				return nil
			}
			handled, handleErr := tryHandleAutoMigrateError(ctx, db, table, err)
			if !handled {
				// 兜底：即便未被处理但仍是 42710，也不让迁移失败
				if errors.As(err, &pgErr) && pgErr.Code == duplicateObjectCode {
					log.Printf("[migrate] duplicate constraint (post-handle) for %T, skipping: %v", table, err)
					return nil
				}
				return err
			}
			if handleErr != nil {
				return handleErr
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("auto migrate retries exceeded for %T", table)
}

func tryHandleAutoMigrateError(ctx context.Context, db *gorm.DB, table interface{}, migrateErr error) (bool, error) {
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return false, nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(migrateErr, &pgErr) {
		return false, nil
	}
	if pgErr.Code != duplicateObjectCode {
		return false, nil
	}
	if strings.TrimSpace(pgErr.ConstraintName) == "" {
		log.Printf("[migrate] duplicate object reported without constraint name, cannot auto fix: %v", migrateErr)
		return false, nil
	}
	tableName, err := resolveTableName(db, table)
	if err != nil {
		return true, err
	}
	if err := dropConstraintIfExists(ctx, db, tableName, pgErr.ConstraintName); err != nil {
		return true, err
	}
	log.Printf("[migrate] dropped duplicate constraint %q on %s, retrying migrate", pgErr.ConstraintName, tableName)
	if retryErr := db.WithContext(ctx).AutoMigrate(table); retryErr != nil {
		return true, retryErr
	}
	return true, nil
}

const duplicateObjectCode = "42710"

func dropConstraintIfExists(ctx context.Context, db *gorm.DB, tableName, constraintName string) error {
	clean := sanitizeConstraintName(constraintName)
	query := fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s`, tableName, quoteIdentifier(clean))
	return db.WithContext(ctx).Exec(query).Error
}

func ensureChannelMasterUniqueIndex(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	const idxName = "idx_channel_master_tenant_store"
	tableName := models.S(models.TableChannelMasters)
	dropSQL := fmt.Sprintf(`DROP INDEX IF EXISTS %s`, quoteIdentifier(idxName))
	if err := db.WithContext(ctx).Exec(dropSQL).Error; err != nil {
		return fmt.Errorf("drop index %s failed: %w", idxName, err)
	}
	createSQL := fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, store_id)`, quoteIdentifier(idxName), tableName)
	if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create unique index %s failed: %w", idxName, err)
	}
	return nil
}

func ensureChannelMetricUniqueIndex(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	tableName := models.S(models.TableChannelMetrics)
	tenantCol := quoteIdentifier("tenant_uuid")
	channelCol := quoteIdentifier("channel_id")
	windowCol := quoteIdentifier("window")
	if err := dedupeChannelMetricScope(ctx, db, tableName); err != nil {
		return fmt.Errorf("dedupe channel metrics failed: %w", err)
	}
	const idxName = "uniq_channel_metric_scope"
	createSQL := fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(%s, %s, %s)`,
		quoteIdentifier(idxName), tableName, tenantCol, channelCol, windowCol)
	if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create unique index %s failed: %w", idxName, err)
	}
	return nil
}

func ensurePaymentProviderUniqueIndex(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	tableName := models.S(models.TablePaymentProviders)
	if err := dedupePaymentProviderSelector(ctx, db, tableName); err != nil {
		return fmt.Errorf("dedupe payment providers failed: %w", err)
	}
	const idxName = "uniq_payment_provider_selector"
	createSQL := fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, provider_type, mch_id, app_id) WHERE deleted_at IS NULL`,
		quoteIdentifier(idxName), tableName)
	if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create unique index %s failed: %w", idxName, err)
	}
	return nil
}

func ensureCustomerIdentityUniqueIndex(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	tableName := models.S(models.TableCustomerIdentity)
	if err := dedupeCustomerIdentityScope(ctx, db, tableName); err != nil {
		return fmt.Errorf("dedupe customer identities failed: %w", err)
	}
	const oldIdxName = "uniq_customer_identity_subject"
	dropSQL := fmt.Sprintf(`DROP INDEX IF EXISTS %s`, quoteIdentifier(oldIdxName))
	if err := db.WithContext(ctx).Exec(dropSQL).Error; err != nil {
		return fmt.Errorf("drop index %s failed: %w", oldIdxName, err)
	}
	const idxName = "uniq_customer_identity_scope"
	createSQL := fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, provider, app_id, subject) WHERE deleted_at IS NULL`,
		quoteIdentifier(idxName), tableName)
	if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
		return fmt.Errorf("create unique index %s failed: %w", idxName, err)
	}
	return nil
}

func ensureCouponIndexes(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	stmts := []string{
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, code) WHERE deleted_at IS NULL`,
			quoteIdentifier("uk_coupon_template_tenant_code"), models.S(models.TableCouponTemplates)),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, coupon_code) WHERE deleted_at IS NULL`,
			quoteIdentifier("uk_coupon_asset_code"), models.S(models.TableCouponAssets)),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, action, idempotency_key)`,
			quoteIdentifier("uk_coupon_usage_idempotency"), models.S(models.TableCouponUsageLogs)),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, order_id) WHERE deleted_at IS NULL`,
			quoteIdentifier("uk_order_coupon_snapshot"), models.S(models.TableOrderCouponSnapshots)),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, user_id, status, valid_to) WHERE deleted_at IS NULL`,
			quoteIdentifier("idx_coupon_asset_user_status"), models.S(models.TableCouponAssets)),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, reserved_order_id) WHERE deleted_at IS NULL`,
			quoteIdentifier("idx_coupon_asset_reserved_order"), models.S(models.TableCouponAssets)),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, order_id, created_at)`,
			quoteIdentifier("idx_coupon_usage_order_created"), models.S(models.TableCouponUsageLogs)),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(tenant_uuid, priced_at) WHERE deleted_at IS NULL`,
			quoteIdentifier("idx_order_coupon_snapshot_priced"), models.S(models.TableOrderCouponSnapshots)),
	}
	for _, stmt := range stmts {
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func dedupePaymentProviderSelector(ctx context.Context, db *gorm.DB, tableName string) error {
	tenantCol := quoteIdentifier("tenant_uuid")
	typeCol := quoteIdentifier("provider_type")
	mchCol := quoteIdentifier("mch_id")
	appCol := quoteIdentifier("app_id")
	updatedCol := quoteIdentifier("updated_at")
	createdCol := quoteIdentifier("created_at")
	idCol := quoteIdentifier("id")
	deletedCol := quoteIdentifier("deleted_at")
	query := fmt.Sprintf(`
UPDATE %s AS p
SET %s = NOW()
FROM (
	SELECT ctid
	FROM (
		SELECT ctid,
			ROW_NUMBER() OVER (
				PARTITION BY %s, %s, %s, %s
				ORDER BY %s DESC NULLS LAST, %s DESC NULLS LAST, %s DESC
			) AS rn
		FROM %s
		WHERE %s IS NULL
	) ranked
	WHERE ranked.rn > 1
) dup
WHERE p.ctid = dup.ctid`, tableName, deletedCol, tenantCol, typeCol, mchCol, appCol, updatedCol, createdCol, idCol, tableName, deletedCol)
	return db.WithContext(ctx).Exec(query).Error
}

func dedupeCustomerIdentityScope(ctx context.Context, db *gorm.DB, tableName string) error {
	tenantCol := quoteIdentifier("tenant_uuid")
	providerCol := quoteIdentifier("provider")
	appCol := quoteIdentifier("app_id")
	subjectCol := quoteIdentifier("subject")
	updatedCol := quoteIdentifier("updated_at")
	createdCol := quoteIdentifier("created_at")
	idCol := quoteIdentifier("id")
	deletedCol := quoteIdentifier("deleted_at")
	query := fmt.Sprintf(`
UPDATE %s AS c
SET %s = NOW()
FROM (
	SELECT ctid
	FROM (
		SELECT ctid,
			ROW_NUMBER() OVER (
				PARTITION BY %s, %s, %s, %s
				ORDER BY %s DESC NULLS LAST, %s DESC NULLS LAST, %s DESC
			) AS rn
		FROM %s
		WHERE %s IS NULL
	) ranked
	WHERE ranked.rn > 1
) dup
WHERE c.ctid = dup.ctid`, tableName, deletedCol, tenantCol, providerCol, appCol, subjectCol, updatedCol, createdCol, idCol, tableName, deletedCol)
	return db.WithContext(ctx).Exec(query).Error
}

func dedupeChannelMetricScope(ctx context.Context, db *gorm.DB, tableName string) error {
	tenantCol := quoteIdentifier("tenant_uuid")
	channelCol := quoteIdentifier("channel_id")
	windowCol := quoteIdentifier("window")
	updatedCol := quoteIdentifier("updated_at")
	createdCol := quoteIdentifier("created_at")
	idCol := quoteIdentifier("id")
	query := fmt.Sprintf(`
DELETE FROM %s AS cm
USING (
	SELECT ctid
	FROM (
		SELECT ctid,
			ROW_NUMBER() OVER (
				PARTITION BY %s, %s, %s
				ORDER BY %s DESC NULLS LAST, %s DESC NULLS LAST, %s DESC
			) AS rn
		FROM %s
	) ranked
	WHERE ranked.rn > 1
) dup
WHERE cm.ctid = dup.ctid`, tableName, tenantCol, channelCol, windowCol, updatedCol, createdCol, idCol, tableName)
	return db.WithContext(ctx).Exec(query).Error
}

func resolveTableName(db *gorm.DB, table interface{}) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(table); err != nil {
		return "", err
	}
	if stmt.Schema == nil || stmt.Schema.Table == "" {
		return "", fmt.Errorf("failed to resolve table name for %T", table)
	}
	return stmt.Schema.Table, nil
}

func quoteIdentifier(name string) string {
	escaped := strings.ReplaceAll(name, "\"", "\"\"")
	return fmt.Sprintf(`"%s"`, escaped)
}

// sanitizeConstraintName strips surrounding quotes and schema prefixes that may appear in pg error messages.
func sanitizeConstraintName(name string) string {
	trimmed := strings.Trim(name, `"`)
	// drop all quotes that may be embedded (fk_"public"_foo)
	trimmed = strings.ReplaceAll(trimmed, `"`, "")
	// remove schema prefix like public_ or public__
	trimmed = strings.TrimPrefix(trimmed, `public_`)
	trimmed = strings.ReplaceAll(trimmed, `public__`, ``)
	return trimmed
}

func isSQLite(db *gorm.DB) bool {
	if db == nil || db.Dialector == nil {
		return false
	}
	return strings.EqualFold(db.Dialector.Name(), "sqlite")
}

func filterSQLiteIncompatibleTables(tables []interface{}) []interface{} {
	filtered := make([]interface{}, 0, len(tables))
	skipped := 0
	for _, tbl := range tables {
		if !isSQLiteSafeTable(tbl) {
			skipped++
			continue
		}
		filtered = append(filtered, tbl)
	}
	if skipped > 0 {
		log.Printf("[migrate] sqlite 环境仅迁移 IAM + 插件核心表，跳过 %d 张业务表", skipped)
	}
	return filtered
}

func isSQLiteSafeTable(tbl interface{}) bool {
	switch tbl.(type) {
	case *models.PluginCredential,
		*models.PluginTenantExt,
		*templateModel.Template:
		return true
	default:
		return false
	}
}

func ensureChannelRLSPolicies(ctx context.Context, db *gorm.DB) error {
	if db == nil || db.Dialector == nil || !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	type policy struct {
		table string
		name  string
	}
	targets := []policy{
		{models.S(models.TableChannelMasters), "channel_master_tenant_rls"},
		{models.S(models.TableChannelTaskLinks), "channel_task_link_tenant_rls"},
		{models.S(models.TableChannelNotes), "channel_note_tenant_rls"},
		{models.S(models.TableChannelSyncHistory), "channel_sync_history_tenant_rls"},
		{models.S(models.TableChannelCredentials), "channel_credential_tenant_rls"},
		{models.S(models.TableChannelAlerts), "channel_alert_tenant_rls"},
		{models.S(models.TableChannelConfigs), "channel_config_tenant_rls"},
	}
	for _, p := range targets {
		if err := db.WithContext(ctx).Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", p.table)).Error; err != nil {
			return err
		}
		exists, err := rlsPolicyExists(ctx, db, p.table, p.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		using := "tenant_uuid::text = current_setting('app.tenant_uuid', true)"
		stmt := fmt.Sprintf(`CREATE POLICY %s ON %s USING (%s) WITH CHECK (%s)`, p.name, p.table, using, using)
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func rlsPolicyExists(ctx context.Context, db *gorm.DB, tableName, policyName string) (bool, error) {
	schema := models.Schema()
	cleanTable := tableName
	if schema != "" {
		prefix := fmt.Sprintf(`"%s".`, schema)
		cleanTable = strings.TrimPrefix(cleanTable, prefix)
	}
	query := `SELECT COUNT(*) FROM pg_policies WHERE schemaname = current_schema() AND tablename = ? AND policyname = ?`
	args := []any{cleanTable, policyName}
	if schema != "" {
		query = `SELECT COUNT(*) FROM pg_policies WHERE schemaname = ? AND tablename = ? AND policyname = ?`
		args = []any{schema, cleanTable, policyName}
	}
	var count int64
	if err := db.WithContext(ctx).Raw(query, args...).Scan(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func ResetDatabase(ctx context.Context, db *gorm.DB, cfg *config.DatabaseConfig) error {
	if strings.EqualFold(db.Dialector.Name(), "sqlite") || strings.TrimSpace(cfg.Schema) == "" {
		tables := append([]interface{}{}, businessTables...)
		tables = append(tables, iamTables...)
		return db.WithContext(ctx).Migrator().DropTable(tables...)
	}

	// 如果你用 GORM，可以直接 drop 所有表
	// 或者先获取表名，再循环 drop
	// 这里举例简单版本：
	err := db.Exec("DROP SCHEMA " + cfg.Schema + " CASCADE; CREATE SCHEMA " + cfg.Schema + ";").Error
	if err != nil {
		return err
	}
	return nil
}

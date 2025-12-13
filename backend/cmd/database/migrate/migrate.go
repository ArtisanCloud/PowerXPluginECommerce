package migrate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	adminconsoleModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/admin_console"
	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	iammodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	marketplaceModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/marketplace"
	operationsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/operations"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	runtimeOpsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/runtime_ops"
	securityModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/security"
	templateModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/template"
	toolgrantModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/tool_grant"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

var businessTables = []interface{}{
	&models.PluginCredential{},
	&models.PluginTenantExt{},
	&customermodel.Customer{},
	&templateModel.Template{},
	&productmodel.SPU{},
	&productmodel.SPUVersion{},
	&productmodel.SPULocale{},
	&productmodel.ChannelVisibility{},
	&productmodel.SubscriptionPlan{},
	&productmodel.SPUImportTask{},
	&productmodel.SPUExportTask{},
	&productmodel.SPUApprovalRecord{},
	&productmodel.SPUAuditLog{},
	&channelmodel.ChannelMaster{},
	&channelmodel.ChannelTaskLink{},
	&channelmodel.ChannelNote{},
	&channelmodel.ChannelSyncHistory{},
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
	&runtimeOpsModel.MCPSession{},
	&runtimeOpsModel.RuntimeAuditEvent{},
	&runtimeOpsModel.QuotaLedger{},
	&runtimeOpsModel.MarketplaceOverage{},
	&operationsModel.SupportChannel{},
	&operationsModel.SupportTicket{},
	&operationsModel.SupportTicketEvent{},
	&operationsModel.ReadinessChecklistItem{},
	&operationsModel.SLAProfile{},
	&operationsModel.SLAAdjustment{},
	&operationsModel.Incident{},
	&operationsModel.IncidentTimelineEntry{},
	&operationsModel.IncidentChecklistItem{},
	&securityModel.BaselineChecklist{},
	&securityModel.AuditReport{},
	&toolgrantModel.Revocation{},
	&toolgrantModel.UsageEvent{},
	&adminconsoleModel.AuditEvent{},
	&adminconsoleModel.ConfigChange{},
	&adminconsoleModel.JobRun{},
}

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
	return ensureChannelRLSPolicies(ctx, db)
}

func safeAutoMigrate(ctx context.Context, db *gorm.DB, tables []interface{}) error {
	for _, tbl := range tables {
		if err := migrateWithTolerance(ctx, db, tbl); err != nil {
			return err
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
	}
	for _, p := range targets {
		if err := db.WithContext(ctx).Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", p.table)).Error; err != nil {
			return err
		}
		using := "tenant_uuid::text = current_setting('app.tenant_uuid', true)"
		stmt := fmt.Sprintf(`CREATE POLICY IF NOT EXISTS %s ON %s USING (%s) WITH CHECK (%s)`, p.name, p.table, using, using)
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
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

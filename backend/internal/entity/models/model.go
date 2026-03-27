package models

// backend/internal/entity/models/model.go

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;comment:自增ID" json:"id"`
	TenantUuid string         `gorm:"type:uuid;not null;index;comment:租户UUID" json:"tenant_uuid"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;comment:创建时间"          json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime;comment:更新时间"          json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:软删除时间"                  json:"deleted_at,omitempty"`
}

type BaseNoTenantModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

const (
	TablePluginTenantExt                 = "plugin_tenant_ext"
	TableCustomer                        = "customers"
	TableCustomerAccount                 = "customer_accounts"
	TableCustomerIdentity                = "customer_identities"
	TableTemplate                        = "template"
	TablePluginCredentials               = "plugin_credentials"
	TablePrivacyDataClassifications      = "privacy_data_classifications"
	TablePrivacyConsentTokens            = "privacy_consent_tokens"
	TablePrivacyLifecycleEvents          = "privacy_lifecycle_events"
	TableSecurityBaselineChecklists      = "security_baseline_checklists"
	TableSecurityAuditReports            = "security_audit_reports"
	TableSecurityVulnerabilityAdvisory   = "security_vulnerability_advisories"
	TableSecurityAdvisoryDistributions   = "security_advisory_distributions"
	TableToolGrantRevocations            = "tool_grant_revocations"
	TableToolGrantUsageEvents            = "tool_grant_usage_events"
	TableIntegrationWebhookSubscriptions = "integration_webhook_subscriptions"
	TableIntegrationWebhookAttempts      = "integration_webhook_attempts"
	TableIntegrationSecrets              = "integration_secrets"
	TableMarketplaceListings             = "marketplace_listings"
	TableMarketplaceListingAssets        = "marketplace_listing_assets"
	TableMarketplaceListingVersions      = "marketplace_listing_versions"
	TableMarketplacePricingPlans         = "marketplace_pricing_plans"
	TableMarketplacePlanTiers            = "marketplace_plan_tiers"
	TableMarketplaceChecklistRuns        = "marketplace_checklist_runs"
	TableMarketplaceChecklistItems       = "marketplace_checklist_items"
	TableMarketplaceLicenses             = "marketplace_licenses"
	TableMarketplaceLicenseEvents        = "marketplace_license_events"
	TableMarketplaceTaxTransactions      = "marketplace_tax_transactions"
	TableMarketplaceUsageEnvelopes       = "marketplace_usage_envelopes"
	TableMarketplaceUsageAggregates      = "marketplace_usage_aggregates"
	TableMarketplaceRevenueReports       = "marketplace_revenue_share_reports"
	TableMarketplaceNotifications        = "marketplace_notifications"
	TableChannelMasters                  = "channel_masters"
	TableChannelCredentials              = "channel_credentials"
	TableChannelMetrics                  = "channel_metrics"
	TableChannelAlerts                   = "channel_alerts"
	TableChannelTaskLinks                = "channel_task_links"
	TableChannelNotes                    = "channel_notes"
	TableChannelConfigs                  = "channel_configs"
	TableChannelSyncHistory              = "channel_sync_history"
	TableOperationsSupportChannels       = "operations_support_channels"
	TableOperationsSupportTickets        = "operations_support_tickets"
	TableOperationsIncidents             = "operations_incidents"
	TableOperationsIncidentUpdates       = "operations_incident_updates"
	TableOperationsIncidentChecklist     = "operations_incident_checklist"
	TableOperationsSupportTicketEvents   = "operations_support_ticket_events"
	TableOperationsReadinessItems        = "operations_readiness_checklist_items"
	TableOperationsSLAScores             = "operations_sla_profiles"
	TableOperationsSLAAdjustments        = "operations_sla_adjustments"
	TableProductSpus                     = "product_spus"
	TableProductSpuVersions              = "product_spu_versions"
	TableProductSpuLocales               = "product_spu_locales"
	TableProductSpuChannels              = "product_spu_channels"
	TableProductSpuSubscriptionPlans     = "product_spu_subscription_plans"
	TableProductSpuSubscriptionPlanLinks = "product_spu_subscription_plan_benefits"
	TableProductSpuImportTasks           = "product_spu_import_tasks"
	TableProductSpuExportTasks           = "product_spu_export_tasks"
	TableProductSpuApprovals             = "product_spu_approvals"
	TableProductSpuAuditLogs             = "product_spu_audit_logs"
	TableProductSpecGroups               = "product_spec_groups"
	TableProductSpecOptions              = "product_spec_options"
	TableProductSkus                     = "product_skus"
	TableProductSkuAttributes            = "product_sku_attributes"
	TableProductSkuChannels              = "product_sku_channels"
	TableProductSkuInventories           = "product_sku_inventories"
	TableProductSkuMedia                 = "product_sku_media"
	TableProductSkuBulkTasks             = "product_sku_bulk_tasks"
	TableProductSkuBulkTaskItems         = "product_sku_bulk_task_items"
	TableProductSkuSerials               = "product_sku_serials"
	TableProductSkuAuditLogs             = "product_sku_audit_logs"
	TableProductCategories               = "product_categories"
	TableProductCategoryLocales          = "product_category_locales"
	TableProductCategoryTemplates        = "product_category_templates"
	TableProductCategoryTemplateFields   = "product_category_template_fields"
	TableProductCategoryTemplateVersions = "product_category_template_versions"
	TableProductCategoryMappings         = "product_category_mappings"
	TableProductCategoryPermissions      = "product_category_permissions"
	TableAdminConsoleAuditEvents         = "admin_console_audit_events"
	TableAdminConsoleConfigChanges       = "admin_console_config_changes"
	TableAdminConsoleJobRuns             = "admin_console_job_runs"
	TableIAMTenants                      = "iam_tenants"
	TableIAMUsers                        = "iam_users"
	TableIAMMembers                      = "iam_members"
	TableIAMRoles                        = "iam_roles"
	TableIAMPermissions                  = "iam_permissions"
	TableIAMDepartments                  = "iam_departments"
	TableIAMMemberRoles                  = "iam_member_roles"
	TableIAMRolePermissions              = "iam_role_permissions"
	TableIAMRefreshTokens                = "iam_refresh_tokens"

	// Pricing (Pricebook)
	TablePricebooks         = "pricebooks"
	TablePricebookVersions  = "pricebook_versions"
	TablePricebookScopes    = "pricebook_scopes"
	TablePricebookItems     = "pricebook_items"
	TablePricebookAuditLogs = "pricebook_audit_logs"

	// Customer Address Book (Checkout MVP)
	TableCustomerAddresses = "customer_addresses"

	// Order (Checkout MVP)
	TableOrders              = "orders"
	TableOrderItems          = "order_items"
	TableOrderEvents         = "order_events"
	TableOrderBenefitReviews = "order_benefit_reviews"

	// Cart (Checkout MVP)
	TableCarts = "carts"

	// Membership & Entitlements
	TableMembershipTiers       = "membership_tiers"
	TableMembershipBenefits    = "membership_benefits"
	TableMembershipAssignments = "membership_assignments"
	TableEntitlements          = "entitlements"
	TableTokenAccounts         = "token_accounts"
	TableTokenTransactions     = "token_transactions"

	// Payments
	TablePaymentProviders           = "payment_providers"
	TablePaymentProviderConfigs     = "payment_provider_configs"
	TablePaymentTransactions        = "payment_transactions"
	TablePaymentLogs                = "payment_logs"
	TablePaymentReconciliations     = "payment_reconciliations"
	TablePaymentRiskEvents          = "payment_risk_events"
	TablePaymentSplitRules          = "payment_split_rules"
	TablePaymentSplitResults        = "payment_split_results"
	TablePaymentReconciliationItems = "payment_reconciliation_items"
	TablePaymentRefunds             = "payment_refunds"
	TablePaymentManualReviews       = "payment_manual_reviews"
	TablePaymentManualReviewLogs    = "payment_manual_review_logs"

	// Fulfillment & Logistics
	TableLogisticsCarriers              = "logistics_carriers"
	TableLogisticsCarrierServices       = "logistics_carrier_services"
	TableLogisticsRateTemplates         = "logistics_rate_templates"
	TableLogisticsRateZones             = "logistics_rate_zones"
	TableLogisticsRoutingRules          = "logistics_routing_rules"
	TableLogisticsRiskRules             = "logistics_risk_rules"
	TableLogisticsBlacklists            = "logistics_blacklists"
	TableLogisticsRiskHits              = "logistics_risk_hits"
	TableLogisticsWaybills              = "logistics_waybills"
	TableLogisticsTrackingEvents        = "logistics_tracking_events"
	TableLogisticsTrackingSyncJobs      = "logistics_tracking_sync_jobs"
	TableLogisticsRedeliveryTasks       = "logistics_redelivery_tasks"
	TableLogisticsETAPolicies           = "logistics_eta_policies"
	TableLogisticsETARecords            = "logistics_eta_records"
	TableLogisticsLabelPrintTasks       = "logistics_label_print_tasks"
	TableLogisticsBillingCases          = "logistics_billing_cases"
	TableLogisticsNotificationTemplates = "logistics_notification_templates"
	TableLogisticsNotificationRecords   = "logistics_notification_records"
	TableFulfillmentTasks               = "fulfillment_tasks"
	TableFulfillmentTaskLogs            = "fulfillment_task_logs"
	TableFulfillmentWaves               = "fulfillment_waves"
	TableFulfillmentWaveTaskLinks       = "fulfillment_wave_task_links"
	TableFulfillmentExceptions          = "fulfillment_exceptions"
	TableFulfillmentWaveStrategies      = "fulfillment_wave_strategies"
	TableReverseWaybills                = "reverse_waybills"
	TableReverseWaybillTrackingEvents   = "reverse_waybill_tracking_events"
	TableReverseWaybillWarehouseResults = "reverse_waybill_warehouse_results"
	TableReverseInspectionRules         = "reverse_inspection_rules"
	TableReverseWaybillInspections      = "reverse_waybill_inspections"
)

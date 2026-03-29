package logistics

import (
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts logistics admin route group.
func RegisterRoutes(router *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if router == nil {
		return nil
	}
	rg := router.Group("/logistics", httpmw.EnsureTenant())
	if deps == nil {
		return rg
	}
	handler := NewHandler(
		logisticssvc.NewCarrierService(deps),
		logisticssvc.NewRateTemplateService(deps),
		logisticssvc.NewRateQuoteService(deps),
		logisticssvc.NewWaybillService(deps),
		logisticssvc.NewETAService(deps),
		logisticssvc.NewRoutingService(deps),
		logisticssvc.NewRedeliveryService(deps),
		logisticssvc.NewRiskService(deps),
		logisticssvc.NewBillingService(deps),
		logisticssvc.NewBillingCaseService(deps),
		logisticssvc.NewNotificationService(deps),
		logisticssvc.NewSLAService(deps),
		logisticssvc.NewLabelPrintService(deps),
		logisticssvc.NewTrackingSyncJobService(deps),
		logisticssvc.NewTrackingSyncSchedulerService(deps),
		logisticssvc.NewGatewayMetricsService(deps),
		logisticssvc.NewGatewayCostService(deps),
		logisticssvc.NewGatewayRecoveryService(deps),
		logisticssvc.NewExceptionOrchestrationService(deps),
		logisticssvc.NewAddressValidationService(deps),
		logisticssvc.NewRoutingOptimizerService(deps),
		logisticssvc.NewSettlementService(deps),
		logisticssvc.NewReconciliationService(deps),
		logisticssvc.NewControlTowerService(deps),
		logisticssvc.NewAllocationService(deps),
		logisticssvc.NewLastmileRecoveryService(deps),
		logisticssvc.NewCrossborderService(deps),
		logisticssvc.NewCustomsRuleService(deps),
		logisticssvc.NewKPIDashboardService(deps),
		logisticssvc.NewSLOGuardService(deps),
		logisticssvc.NewCapacityForecastService(deps),
		logisticssvc.NewWebhookService(deps),
	)

	rg.GET("/carriers", handler.ListCarriers)
	rg.POST("/carriers", handler.UpsertCarrier)
	rg.PATCH("/carriers/:id", handler.DisableCarrier)
	rg.POST("/carriers/:id/test", handler.TestCarrier)

	rg.GET("/templates", handler.ListTemplates)
	rg.POST("/templates", handler.UpsertTemplate)
	rg.PATCH("/templates/:id", handler.UpsertTemplate)
	rg.POST("/templates/:id/publish", handler.PublishTemplate)
	rg.POST("/templates/:id/quote", handler.QuoteTemplate)

	rg.GET("/waybills", handler.ListWaybills)
	rg.GET("/routing/rules", handler.ListRoutingRules)
	rg.POST("/routing/rules", handler.UpsertRoutingRule)
	rg.PATCH("/routing/rules/:id", handler.UpsertRoutingRule)
	rg.DELETE("/routing/rules/:id", handler.DeleteRoutingRule)
	rg.POST("/routing/preview", handler.PreviewRouting)
	rg.GET("/redelivery/tasks", handler.ListRedeliveryTasks)
	rg.POST("/redelivery/tasks/initiate", handler.InitiateRedeliveryTask)
	rg.POST("/redelivery/tasks/:id/address", handler.UpdateRedeliveryAddress)
	rg.POST("/redelivery/tasks/:id/redispatch", handler.RedispatchRedeliveryTask)
	rg.POST("/redelivery/tasks/:id/close", handler.CloseRedeliveryTask)
	rg.GET("/risk/rules", handler.ListRiskRules)
	rg.POST("/risk/rules", handler.UpsertRiskRule)
	rg.PATCH("/risk/rules/:id", handler.UpsertRiskRule)
	rg.GET("/risk/blacklist", handler.ListRiskBlacklist)
	rg.POST("/risk/blacklist", handler.UpsertRiskBlacklist)
	rg.PATCH("/risk/blacklist/:id", handler.UpsertRiskBlacklist)
	rg.GET("/risk/hits", handler.ListRiskHits)
	rg.POST("/risk/evaluate", handler.EvaluateRisk)
	rg.POST("/risk/hits/:id/release", handler.ReleaseRiskHit)
	rg.GET("/eta", handler.ListWaybillETA)
	rg.GET("/eta/:waybill_id", handler.GetWaybillETA)
	rg.POST("/waybills", handler.CreateWaybill)
	rg.GET("/waybills/:id", handler.GetWaybill)
	rg.POST("/waybills/:id/track", handler.AppendTracking)
	rg.POST("/waybills/:id/sync-track", handler.SyncWaybillTracking)
	rg.POST("/waybills/:id/cancel", handler.CancelWaybill)
	rg.PATCH("/waybills/:id/cost", handler.UpdateWaybillCost)
	rg.GET("/billing/summary", handler.BillingSummary)
	rg.GET("/billing/export", handler.ExportBilling)
	rg.GET("/billing/cases", handler.ListBillingCases)
	rg.POST("/billing/cases", handler.CreateBillingCase)
	rg.PATCH("/billing/cases/:id/transition", handler.TransitionBillingCase)
	rg.GET("/notifications/templates", handler.ListNotificationTemplates)
	rg.POST("/notifications/templates", handler.UpsertNotificationTemplate)
	rg.GET("/notifications/records", handler.ListNotificationRecords)
	rg.POST("/notifications/send", handler.SendNotification)
	rg.POST("/notifications/records/:id/retry", handler.RetryNotification)
	rg.GET("/sla/dashboard", handler.SLADashboard)
	rg.GET("/labels/prints", handler.ListLabelPrintTasks)
	rg.POST("/labels/prints", handler.BatchPrintLabels)
	rg.POST("/labels/prints/retry", handler.RetryLabelPrint)
	rg.GET("/tracking-sync/jobs", handler.ListTrackingSyncJobs)
	rg.POST("/tracking-sync/jobs", handler.CreateTrackingSyncJob)
	rg.POST("/tracking-sync/jobs/:id/cancel", handler.CancelTrackingSyncJob)
	rg.POST("/tracking-sync/jobs/:id/retry", handler.RetryTrackingSyncJob)
	rg.GET("/tracking-sync/schedules", handler.ListTrackingSyncSchedules)
	rg.POST("/tracking-sync/schedules", handler.UpsertTrackingSyncSchedule)
	rg.PATCH("/tracking-sync/schedules/:id", handler.UpsertTrackingSyncSchedule)
	rg.POST("/tracking-sync/schedules/:id/toggle", handler.ToggleTrackingSyncSchedule)
	rg.POST("/tracking-sync/schedules/:id/trigger", handler.TriggerTrackingSyncSchedule)
	rg.POST("/tracking-sync/schedules/run-due", handler.RunDueTrackingSyncSchedules)
	rg.GET("/gateway/health", handler.GatewayHealth)
	rg.GET("/gateway/costs", handler.GatewayCosts)
	rg.GET("/gateway/cost-alerts", handler.GatewayCostAlerts)
	rg.GET("/gateway/failures", handler.ListGatewayFailures)
	rg.POST("/gateway/failures/ingest", handler.IngestGatewayFailures)
	rg.POST("/gateway/failures/:id/compensate", handler.CompensateGatewayFailure)
	rg.GET("/exceptions/orchestration/rules", handler.ListExceptionOrchestrationRules)
	rg.POST("/exceptions/orchestration/rules", handler.UpsertExceptionOrchestrationRule)
	rg.PATCH("/exceptions/orchestration/rules/:id", handler.UpsertExceptionOrchestrationRule)
	rg.GET("/exceptions/orchestration/runs", handler.ListExceptionOrchestrationRuns)
	rg.POST("/exceptions/orchestration/execute", handler.ExecuteExceptionOrchestration)
	rg.POST("/address-validation/check", handler.CheckAddressValidation)
	rg.GET("/address-validation/records", handler.ListAddressValidationRecords)
	rg.GET("/routing/optimizer/strategy", handler.GetRoutingOptimizerStrategy)
	rg.POST("/routing/optimizer/strategy", handler.UpsertRoutingOptimizerStrategy)
	rg.POST("/routing/optimizer/simulate", handler.SimulateRoutingOptimizer)
	rg.GET("/settlement/batches", handler.ListSettlementBatches)
	rg.POST("/settlement/batches", handler.CreateSettlementBatch)
	rg.GET("/settlement/diffs", handler.ListSettlementDiffs)
	rg.POST("/settlement/diffs/:id/handle", handler.HandleSettlementDiff)
	rg.POST("/settlement/batches/:id/confirm", handler.ConfirmSettlementBatch)
	rg.GET("/reconciliation/batches", handler.ListReconciliationBatches)
	rg.POST("/reconciliation/batches", handler.CreateReconciliationBatch)
	rg.GET("/reconciliation/records", handler.ListReconciliationRecords)
	rg.GET("/reconciliation/cases", handler.ListReconciliationCases)
	rg.POST("/reconciliation/cases/:id/handle", handler.HandleReconciliationCase)
	rg.GET("/control-tower/overview", handler.GetControlTowerOverview)
	rg.GET("/control-tower/drilldown", handler.GetControlTowerDrilldown)
	rg.GET("/control-tower/subscriptions", handler.ListControlTowerSubscriptions)
	rg.POST("/control-tower/subscriptions", handler.UpsertControlTowerSubscription)
	rg.PATCH("/control-tower/subscriptions/:id", handler.UpsertControlTowerSubscription)
	rg.GET("/allocation/plans", handler.ListCapacityPlans)
	rg.POST("/allocation/plans", handler.UpsertCapacityPlan)
	rg.PATCH("/allocation/plans/:id", handler.UpsertCapacityPlan)
	rg.POST("/allocation/allocate", handler.AllocateCarrier)
	rg.POST("/allocation/override", handler.OverrideAllocation)
	rg.GET("/allocation/forecasts", handler.ListCapacityForecasts)
	rg.POST("/allocation/forecasts/generate", handler.GenerateCapacityForecast)
	rg.POST("/allocation/forecasts/:id/apply", handler.ApplyCapacityForecast)
	rg.GET("/lastmile-recovery/rules", handler.ListLastmileRecoveryRules)
	rg.POST("/lastmile-recovery/rules", handler.UpsertLastmileRecoveryRule)
	rg.PATCH("/lastmile-recovery/rules/:id", handler.UpsertLastmileRecoveryRule)
	rg.POST("/lastmile-recovery/execute", handler.ExecuteLastmileRecovery)
	rg.GET("/lastmile-recovery/runs", handler.ListLastmileRecoveryRuns)
	rg.POST("/lastmile-recovery/runs/:id/takeover", handler.TakeoverLastmileRecovery)
	rg.GET("/crossborder/documents", handler.ListCrossborderDocuments)
	rg.POST("/crossborder/documents", handler.UpsertCrossborderDocument)
	rg.PATCH("/crossborder/documents/:id", handler.UpsertCrossborderDocument)
	rg.POST("/crossborder/tax-quote", handler.QuoteCrossborderTax)
	rg.GET("/crossborder/tracking-maps", handler.ListCrossborderTrackingMaps)
	rg.POST("/crossborder/tracking-maps", handler.UpsertCrossborderTrackingMap)
	rg.PATCH("/crossborder/tracking-maps/:id", handler.UpsertCrossborderTrackingMap)
	rg.POST("/crossborder/tracking-maps/normalize", handler.NormalizeCrossborderTracking)
	rg.GET("/customs/rule-packs", handler.ListCustomsRulePacks)
	rg.POST("/customs/rule-packs", handler.UpsertCustomsRulePack)
	rg.PATCH("/customs/rule-packs/:id", handler.UpsertCustomsRulePack)
	rg.GET("/customs/rule-packs/:id/versions", handler.ListCustomsRuleVersions)
	rg.POST("/customs/rule-packs/:id/versions", handler.PublishCustomsRuleVersion)
	rg.POST("/customs/precheck", handler.CustomsPrecheck)
	rg.GET("/kpi-dashboard/overview", handler.GetKPIDashboardOverview)
	rg.GET("/kpi-dashboard/trends", handler.GetKPIDashboardTrends)
	rg.GET("/kpi-dashboard/drilldown", handler.GetKPIDashboardDrilldown)
	rg.GET("/kpi-dashboard/export", handler.ExportKPIDashboard)
	rg.GET("/slo-guard/policies", handler.ListSLOGuardPolicies)
	rg.POST("/slo-guard/policies", handler.UpsertSLOGuardPolicy)
	rg.PATCH("/slo-guard/policies/:id", handler.UpsertSLOGuardPolicy)
	rg.GET("/slo-guard/status", handler.GetSLOGuardStatus)
	rg.POST("/slo-guard/evaluate", handler.EvaluateSLOGuard)
	rg.POST("/slo-guard/policies/:id/release", handler.ReleaseSLOGuardPolicy)

	rg.POST("/webhook", handler.HandleWebhook)
	return rg
}

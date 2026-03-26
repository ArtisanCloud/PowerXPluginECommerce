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
		logisticssvc.NewBillingService(deps),
		logisticssvc.NewBillingCaseService(deps),
		logisticssvc.NewNotificationService(deps),
		logisticssvc.NewSLAService(deps),
		logisticssvc.NewLabelPrintService(deps),
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
	rg.GET("/eta", handler.ListWaybillETA)
	rg.GET("/eta/:waybill_id", handler.GetWaybillETA)
	rg.POST("/waybills", handler.CreateWaybill)
	rg.GET("/waybills/:id", handler.GetWaybill)
	rg.POST("/waybills/:id/track", handler.AppendTracking)
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

	rg.POST("/webhook", handler.HandleWebhook)
	return rg
}

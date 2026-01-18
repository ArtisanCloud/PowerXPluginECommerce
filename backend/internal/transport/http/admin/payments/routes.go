package payments

import (
	paymentsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers admin payment routes.
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil {
		return
	}
	providerSvc := paymentsvc.NewProviderService(deps)
	transactionSvc := paymentsvc.NewTransactionService(deps)
	refundSvc := paymentsvc.NewRefundService(deps)
	riskSvc := paymentsvc.NewRiskEventService(deps)
	splitRuleSvc := paymentsvc.NewSplitRuleService(deps)
	splitResultSvc := paymentsvc.NewSplitResultService(deps)
	manualReviewSvc := paymentsvc.NewManualReviewService(deps)
	reconciliationSvc := paymentsvc.NewReconciliationService(deps)

	providers := NewProviderHandler(providerSvc)
	transactions := NewTransactionHandler(transactionSvc)
	refunds := NewRefundHandler(refundSvc)
	riskEvents := NewRiskEventHandler(riskSvc)
	splitRules := NewSplitRuleHandler(splitRuleSvc)
	splitResults := NewSplitResultHandler(splitResultSvc)
	manualReviews := NewManualReviewHandler(manualReviewSvc)
	reconciliations := NewReconciliationHandler(reconciliationSvc)

	group := rg.Group("/payments")
	{
		group.GET("/providers", providers.List)
		group.POST("/providers", providers.Create)
		group.PATCH("/providers/:id", providers.Update)
		group.POST("/providers/:id/test", providers.Test)

		group.GET("/transactions", transactions.List)
		group.GET("/transactions/:id", transactions.Get)
		group.POST("/transactions/:id/refund", refunds.Create)

		group.GET("/risk-events", riskEvents.List)
		group.GET("/split-rules", splitRules.List)
		group.GET("/split-results", splitResults.List)
		group.GET("/reconciliations", reconciliations.List)
		group.POST("/reconciliations", reconciliations.Create)
		group.GET("/reconciliations/:id/items", reconciliations.ListItems)
		group.POST("/reconciliations/:id/items/:itemId/resolve", reconciliations.ResolveItem)
		group.GET("/manual-payments", manualReviews.List)
		group.GET("/manual-payments/logs", manualReviews.ListLogs)
		group.POST("/manual-payments", manualReviews.Create)
		group.POST("/manual-payments/:id/approve", manualReviews.Approve)
		group.POST("/manual-payments/:id/reject", manualReviews.Reject)
	}
}

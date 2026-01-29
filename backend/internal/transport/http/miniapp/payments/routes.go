package payments

import (
	paymentssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires mini-app payment endpoints for protected and open groups.
func RegisterRoutes(openGroup, protectedGroup *gin.RouterGroup, deps *app.Deps) {
	if openGroup == nil && protectedGroup == nil {
		return
	}
	transactionSvc := paymentssvc.NewTransactionService(deps)
	transactions := NewTransactionHandler(transactionSvc)
	callbacks := NewProviderCallbackHandler(transactionSvc)

	if protectedGroup != nil {
		group := protectedGroup.Group("/payments")
		{
			group.POST("/transactions", transactions.Create)
			group.GET("/transactions/:id", transactions.GetStatus)
		}
	}
	if openGroup != nil {
		group := openGroup.Group("/payments")
		{
			group.POST("/providers/id/:id/callback", callbacks.Callback)
			group.POST("/providers/:type/:mchId/:appId/callback", callbacks.CallbackBySelector)
		}
	}
}

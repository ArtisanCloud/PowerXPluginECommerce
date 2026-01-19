package payments

import (
	paymentssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) {
	if rg == nil {
		return
	}
	transactionSvc := paymentssvc.NewTransactionService(deps)
	transactions := NewTransactionHandler(transactionSvc)
	callbacks := NewProviderCallbackHandler(transactionSvc)

	group := rg.Group("/payments")
	group.Use(httpmw.EnsureTenant())
	{
		group.POST("/transactions", transactions.Create)
		group.GET("/transactions/:id", transactions.GetStatus)
		group.POST("/providers/:id/callback", callbacks.Callback)
	}
}

package auth

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 挂载 mini-app auth endpoints。
func RegisterRoutes(rg *gin.RouterGroup, deps *app.Deps) *gin.RouterGroup {
	if rg == nil {
		return nil
	}
	handler := NewHandler(deps)
	group := rg.Group("/auth")
	group.POST("/register", handler.Register)
	group.POST("/login", handler.Login)
	group.POST("/wechat/login", handler.WechatLogin)
	group.POST("/wechat/phone", handler.WechatPhoneNumber)
	group.POST("/wechat/decrypt", handler.WechatDecryptData)
	group.POST("/wechat/check-encrypted", handler.WechatCheckEncryptedData)
	group.POST("/wechat/paid-unionid", handler.WechatPaidUnionID)
	return group
}

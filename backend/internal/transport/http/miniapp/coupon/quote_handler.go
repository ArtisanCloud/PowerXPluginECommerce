package coupon

import "github.com/gin-gonic/gin"

// Quote provides a phase-1 placeholder for coupon quote endpoint.
func Quote(c *gin.Context) {
	c.JSON(501, gin.H{
		"code":    "NOT_IMPLEMENTED",
		"message": "coupon quote is not implemented yet",
	})
}

package miniapp

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesAddsMiniAppEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	root := engine.Group("/api/v1")

	RegisterRoutes(root, nil)

	var (
		foundProducts bool
		foundTags     bool
		foundSkus     bool
		foundAuth     bool
	)
	for _, route := range engine.Routes() {
		if route.Method == "GET" && route.Path == "/api/v1/mini-app/products" {
			foundProducts = true
		}
		if route.Method == "GET" && route.Path == "/api/v1/mini-app/products/tags" {
			foundTags = true
		}
		if route.Method == "GET" && route.Path == "/api/v1/mini-app/products/:id/skus" {
			foundSkus = true
		}
		if route.Method == "POST" && route.Path == "/api/v1/mini-app/auth/login" {
			foundAuth = true
		}
	}
	if !foundProducts || !foundTags || !foundSkus || !foundAuth {
		t.Fatalf("expected mini-app product routes registered, got %+v", engine.Routes())
	}
}

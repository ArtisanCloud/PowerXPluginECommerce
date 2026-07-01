package skills

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSkillsListAndSchemaRoutes(t *testing.T) {
	t.Setenv("PLUGIN_SKILLS_DIR", filepath.Join("..", "..", "..", "..", "..", "..", "skills"))
	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterAPIRoutes(router.Group("/api/v1"), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugin/skills", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "ecommerce.template.basic")

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugin/skills/ecommerce.template.basic/schema", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "input_schema")
}

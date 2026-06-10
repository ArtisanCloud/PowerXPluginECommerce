package runtime_ops

import (
	"net/http"
	"strings"
	"sync"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	runtimelogging "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/runtime/logging"
	"github.com/gin-gonic/gin"
)

type LoggingPolicyHandler struct{}

var (
	loggingPolicyMu       sync.RWMutex
	loggingDefaultPolicy  = runtimelogging.ResolveWithHostDefaults(runtimelogging.DefaultPolicy())
	loggingPolicyByTenant = map[string]runtimelogging.Policy{}
)

func NewLoggingPolicyHandler() *LoggingPolicyHandler {
	return &LoggingPolicyHandler{}
}

type loggingPolicyRequest struct {
	TenantUUID           string   `json:"tenant_uuid"`
	PolicyVersion        string   `json:"policy_version"`
	Mode                 string   `json:"mode"`
	Sinks                []string `json:"sinks"`
	Format               string   `json:"format"`
	Level                string   `json:"level"`
	AuthorizedExtraSinks []string `json:"authorized_extra_sinks"`
	Retry                struct {
		Enabled     bool `json:"enabled"`
		MaxAttempts int  `json:"max_attempts"`
		BackoffMS   int  `json:"backoff_ms"`
	} `json:"retry"`
}

func policyKey(tenantUUID string) string {
	key := strings.TrimSpace(tenantUUID)
	if key == "" {
		return "__global__"
	}
	return key
}

func currentLoggingPolicyForTenant(tenantUUID string) runtimelogging.Policy {
	loggingPolicyMu.RLock()
	defer loggingPolicyMu.RUnlock()
	if policy, ok := loggingPolicyByTenant[policyKey(tenantUUID)]; ok {
		return policy
	}
	return loggingDefaultPolicy
}

func setLoggingPolicyForTenant(tenantUUID string, policy runtimelogging.Policy) {
	loggingPolicyMu.Lock()
	defer loggingPolicyMu.Unlock()
	loggingPolicyByTenant[policyKey(tenantUUID)] = policy
}

func loggingPolicySuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func (h *LoggingPolicyHandler) Get(c *gin.Context) {
	tenantUUID, mismatch := resolvePolicyTenant(c, "")
	if mismatch {
		contracts.ResponseError(c, http.StatusForbidden, contracts.ErrCodeTenantMismatch, "tenant mismatch")
		return
	}
	loggingPolicySuccess(c, currentLoggingPolicyForTenant(tenantUUID))
}

func (h *LoggingPolicyHandler) Put(c *gin.Context) {
	var req loggingPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload")
		return
	}
	tenantUUID, mismatch := resolvePolicyTenant(c, req.TenantUUID)
	if mismatch {
		contracts.ResponseError(c, http.StatusForbidden, contracts.ErrCodeTenantMismatch, "tenant mismatch")
		return
	}

	p := runtimelogging.Policy{
		PolicyVersion: req.PolicyVersion,
		Mode:          runtimelogging.PolicyMode(req.Mode),
		Format:        req.Format,
		Level:         req.Level,
		Retry: runtimelogging.RetryPolicy{
			Enabled:     req.Retry.Enabled,
			MaxAttempts: req.Retry.MaxAttempts,
			BackoffMS:   req.Retry.BackoffMS,
		},
	}
	for _, sink := range req.Sinks {
		p.Sinks = append(p.Sinks, runtimelogging.SinkType(sink))
	}
	for _, sink := range req.AuthorizedExtraSinks {
		p.AuthorizedExtraSinks = append(p.AuthorizedExtraSinks, runtimelogging.SinkType(sink))
	}

	p = runtimelogging.ResolveWithHostDefaults(p)
	if err := runtimelogging.ValidatePolicy(p); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}

	setLoggingPolicyForTenant(tenantUUID, p)
	loggingPolicySuccess(c, p)
}

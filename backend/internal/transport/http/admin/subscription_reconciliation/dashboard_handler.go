package subscription_reconciliation

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	SubscriptionReconciliationSvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/subscription_reconciliation"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Dashboard(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.Dashboard(c.Request.Context(), tenantUUID, parseDashboardQuery(c))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	contracts.ResponseSuccess(c, resp)
}

func (h *Handler) ExportDashboard(c *gin.Context) {
	if h == nil || h.service == nil {
		contracts.ResponseServiceUnavailable(c, "subscription reconciliation service unavailable", nil)
		return
	}
	tenantUUID, _ := httpmw.TenantUUIDFromContext(c)
	resp, err := h.service.Dashboard(c.Request.Context(), tenantUUID, parseDashboardQuery(c))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	content, err := buildDashboardCSV(resp)
	if err != nil {
		contracts.ResponseInternalError(c, err)
		return
	}
	filename := "subscription-reconciliation-dashboard-" + time.Now().UTC().Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.String(http.StatusOK, content)
}

func parseDashboardQuery(c *gin.Context) SubscriptionReconciliationSvc.DashboardQuery {
	return SubscriptionReconciliationSvc.DashboardQuery{
		From:          strings.TrimSpace(c.Query("from")),
		To:            strings.TrimSpace(c.Query("to")),
		Channel:       strings.TrimSpace(c.Query("channel")),
		Plan:          strings.TrimSpace(c.Query("plan")),
		Region:        strings.TrimSpace(c.Query("region")),
		FailureReason: strings.TrimSpace(c.Query("failureReason")),
	}
}

func buildDashboardCSV(data map[string]any) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"metric", "value"}); err != nil {
		return "", err
	}
	rows := [][2]string{
		{"deltaRate", toCSVString(data["deltaRate"])},
		{"recoveryRate", toCSVString(data["recoveryRate"])},
		{"avgHandleHours", toCSVString(data["avgHandleHours"])},
		{"pendingTasks", toCSVString(data["pendingTasks"])},
	}
	for _, row := range rows {
		if err := w.Write([]string{row[0], row[1]}); err != nil {
			return "", err
		}
	}
	if byType, ok := data["byDeltaType"].([]SubscriptionReconciliationSvc.DashboardDeltaTypeCount); ok {
		for _, item := range byType {
			if err := w.Write([]string{"byDeltaType." + item.Type, toCSVString(item.Count)}); err != nil {
				return "", err
			}
		}
	} else if raw, ok := data["byDeltaType"].([]any); ok {
		for _, item := range raw {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if err := w.Write([]string{"byDeltaType." + toCSVString(obj["type"]), toCSVString(obj["count"])}); err != nil {
				return "", err
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func toCSVString(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

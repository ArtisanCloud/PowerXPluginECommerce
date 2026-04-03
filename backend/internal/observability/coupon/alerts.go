package coupon

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	AlertRedeemFailureBacklog = "coupon.redeem_failure_backlog"
	AlertTimeoutUnreleased    = "coupon.timeout_unreleased"
)

// AlertEvaluator 负责根据观测值触发告警事件。
type AlertEvaluator struct {
	logger *logrus.Entry
}

func NewAlertEvaluator(logger *logrus.Entry) *AlertEvaluator {
	if logger == nil {
		logger = logrus.New().WithField("component", "coupon-alert")
	}
	return &AlertEvaluator{logger: logger}
}

// Evaluate 根据当前 backlog 指标判断是否触发告警。
// redeemFailureBacklog: 核销失败积压数量
// timeoutUnreleased: 超时后仍未释放数量
func (e *AlertEvaluator) Evaluate(ctx context.Context, tenantUUID string, redeemFailureBacklog, timeoutUnreleased int64, at time.Time) {
	if e == nil || e.logger == nil {
		return
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	if redeemFailureBacklog > 0 {
		e.emit(ctx, AlertRedeemFailureBacklog, tenantUUID, map[string]any{
			"backlog": redeemFailureBacklog,
			"at":      at.Format(time.RFC3339),
		})
	}
	if timeoutUnreleased > 0 {
		e.emit(ctx, AlertTimeoutUnreleased, tenantUUID, map[string]any{
			"pending": timeoutUnreleased,
			"at":      at.Format(time.RFC3339),
		})
	}
}

func (e *AlertEvaluator) emit(ctx context.Context, event, tenantUUID string, payload map[string]any) {
	fields := logrus.Fields{
		"event":       event,
		"tenant_uuid": tenantUUID,
	}
	for k, v := range payload {
		fields[k] = v
	}
	if ctx != nil {
		if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
			fields["request_id"] = requestID
		}
	}
	e.logger.WithFields(fields).Warn("coupon alert triggered")
}

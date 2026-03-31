package after_sales

import (
	"time"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
)

type CaseQuery struct {
	Status   string
	CaseType string
	OrderID  string
	Keyword  string
	Page     int
	PageSize int
}

type CaseSummaryDTO struct {
	ID                   string     `json:"id"`
	CaseNo               string     `json:"caseNo"`
	OrderID              string     `json:"orderId"`
	OrderItemID          string     `json:"orderItemId"`
	CustomerID           string     `json:"customerId"`
	CaseType             string     `json:"caseType"`
	Status               string     `json:"status"`
	ReasonCode           string     `json:"reasonCode"`
	ReasonDetail         string     `json:"reasonDetail"`
	RequestedQty         int        `json:"requestedQty"`
	RequestedAmountMinor int64      `json:"requestedAmountMinor"`
	Currency             string     `json:"currency"`
	ReverseWaybillNo     string     `json:"reverseWaybillNo,omitempty"`
	ReverseReceiveStatus string     `json:"reverseReceiveStatus,omitempty"`
	ReverseLinkedAt      *time.Time `json:"reverseLinkedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	ClosedAt             *time.Time `json:"closedAt,omitempty"`
}

type TimelineEventDTO struct {
	Action       string    `json:"action"`
	FromStatus   string    `json:"fromStatus,omitempty"`
	ToStatus     string    `json:"toStatus"`
	OperatorType string    `json:"operatorType"`
	OperatorID   string    `json:"operatorId,omitempty"`
	Note         string    `json:"note,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CaseDetailDTO struct {
	Case     CaseSummaryDTO     `json:"case"`
	Timeline []TimelineEventDTO `json:"timeline"`
}

type CaseListDTO struct {
	Items []CaseSummaryDTO `json:"items"`
	Meta  struct {
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"pageSize"`
	} `json:"meta"`
}

type CaseActionRequest struct {
	ReasonCode string `json:"reasonCode,omitempty"`
	Note       string `json:"note,omitempty"`
}

type DashboardSnapshot struct {
	PendingCount    int64 `json:"pendingCount"`
	ProcessingCount int64 `json:"processingCount"`
	CompletedCount  int64 `json:"completedCount"`
	RejectedCount   int64 `json:"rejectedCount"`
}

type CoreFieldsUpdateRequest struct {
	CaseType             string
	ReasonCode           string
	ReasonDetail         string
	RequestedQty         int
	RequestedAmountMinor int64
	Currency             string
}

func toSummaryDTO(row AfterSalesModel.AfterSaleCase) CaseSummaryDTO {
	return CaseSummaryDTO{
		ID:                   row.ID,
		CaseNo:               row.CaseNo,
		OrderID:              row.OrderID,
		OrderItemID:          row.OrderItemID,
		CustomerID:           row.CustomerID,
		CaseType:             row.CaseType,
		Status:               row.Status,
		ReasonCode:           row.ReasonCode,
		ReasonDetail:         row.ReasonDetail,
		RequestedQty:         row.RequestedQty,
		RequestedAmountMinor: row.RequestedAmountMinor,
		Currency:             row.Currency,
		CreatedAt:            row.CreatedAt,
		ClosedAt:             row.ClosedAt,
	}
}

func toTimelineDTO(rows []AfterSalesModel.AfterSaleTimeline) []TimelineEventDTO {
	items := make([]TimelineEventDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, TimelineEventDTO{
			Action:       row.Action,
			FromStatus:   row.FromStatus,
			ToStatus:     row.ToStatus,
			OperatorType: row.OperatorType,
			OperatorID:   row.OperatorID,
			Note:         row.Note,
			CreatedAt:    row.CreatedAt,
		})
	}
	return items
}

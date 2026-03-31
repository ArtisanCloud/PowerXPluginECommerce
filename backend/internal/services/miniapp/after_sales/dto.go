package after_sales

import (
	"time"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
)

type EvidenceInput struct {
	EvidenceType string `json:"evidenceType"`
	ContentRef   string `json:"contentRef"`
	Description  string `json:"description,omitempty"`
}

type CreateCaseRequest struct {
	OrderID              string          `json:"orderId"`
	OrderItemID          string          `json:"orderItemId"`
	CaseType             string          `json:"caseType"`
	ReasonCode           string          `json:"reasonCode"`
	ReasonDetail         string          `json:"reasonDetail,omitempty"`
	RequestedQty         int             `json:"requestedQty,omitempty"`
	RequestedAmountMinor int64           `json:"requestedAmountMinor,omitempty"`
	Currency             string          `json:"currency,omitempty"`
	Evidences            []EvidenceInput `json:"evidences,omitempty"`
}

type CaseSummaryDTO struct {
	ID                   string    `json:"id"`
	CaseNo               string    `json:"caseNo"`
	OrderID              string    `json:"orderId"`
	OrderItemID          string    `json:"orderItemId"`
	CaseType             string    `json:"caseType"`
	Status               string    `json:"status"`
	RequestedAmountMinor int64     `json:"requestedAmountMinor"`
	Currency             string    `json:"currency"`
	CreatedAt            time.Time `json:"createdAt"`
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

type CaseListMeta struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type CaseListDTO struct {
	Items []CaseSummaryDTO `json:"items"`
	Meta  CaseListMeta     `json:"meta"`
}

func toSummaryDTO(row AfterSalesModel.AfterSaleCase) CaseSummaryDTO {
	return CaseSummaryDTO{
		ID:                   row.ID,
		CaseNo:               row.CaseNo,
		OrderID:              row.OrderID,
		OrderItemID:          row.OrderItemID,
		CaseType:             row.CaseType,
		Status:               row.Status,
		RequestedAmountMinor: row.RequestedAmountMinor,
		Currency:             row.Currency,
		CreatedAt:            row.CreatedAt,
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

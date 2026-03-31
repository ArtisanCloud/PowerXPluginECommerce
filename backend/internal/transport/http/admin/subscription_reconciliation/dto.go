package subscription_reconciliation

import "strings"

type createBatchRequest struct {
	BillingCycle string `json:"billingCycle" binding:"required"`
	RunType      string `json:"runType"`
}

func (r createBatchRequest) normalize() createBatchRequest {
	r.BillingCycle = strings.TrimSpace(r.BillingCycle)
	r.RunType = strings.TrimSpace(strings.ToLower(r.RunType))
	if r.RunType == "" {
		r.RunType = "daily"
	}
	return r
}

type createDeltaTaskRequest struct {
	Assignee string `json:"assignee" binding:"required"`
	SLALevel string `json:"slaLevel" binding:"required"`
	Note     string `json:"note"`
}

func (r createDeltaTaskRequest) normalize() createDeltaTaskRequest {
	r.Assignee = strings.TrimSpace(r.Assignee)
	r.SLALevel = strings.TrimSpace(strings.ToLower(r.SLALevel))
	r.Note = strings.TrimSpace(r.Note)
	return r
}

type closeTaskRequest struct {
	Resolution     string `json:"resolution" binding:"required"`
	ResolutionNote string `json:"resolutionNote"`
}

func (r closeTaskRequest) normalize() closeTaskRequest {
	r.Resolution = strings.TrimSpace(strings.ToLower(r.Resolution))
	r.ResolutionNote = strings.TrimSpace(r.ResolutionNote)
	return r
}

type runGovernanceRequest struct {
	BillingCycle string `json:"billingCycle" binding:"required"`
	DryRun       bool   `json:"dryRun"`
}

func (r runGovernanceRequest) normalize() runGovernanceRequest {
	r.BillingCycle = strings.TrimSpace(r.BillingCycle)
	return r
}

type adjustDeltaRequest struct {
	ExpectedAmountMinor int64  `json:"expectedAmountMinor"`
	ActualAmountMinor   int64  `json:"actualAmountMinor"`
	ReasonCode          string `json:"reasonCode"`
	Note                string `json:"note"`
}

func (r adjustDeltaRequest) normalize() adjustDeltaRequest {
	r.ReasonCode = strings.TrimSpace(r.ReasonCode)
	r.Note = strings.TrimSpace(r.Note)
	return r
}

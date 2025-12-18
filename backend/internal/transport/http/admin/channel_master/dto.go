package channel_master

import (
	"strings"
	"time"

	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
)

// ChannelListQuery exposes optional filters for GET /channels.
type ChannelListQuery struct {
	Keyword   string   `form:"keyword"`
	Platform  string   `form:"platform"`
	Status    []string `form:"status[]"`
	Owner     string   `form:"owner"`
	Region    string   `form:"region"`
	Tags      []string `form:"tags[]"`
	MinHealth int      `form:"minHealth"`
	MinGMV    float64  `form:"minGmv"`
	Page      int      `form:"page"`
	PageSize  int      `form:"pageSize"`
}

// ChannelUpsertRequest mirrors the OpenAPI create/update payload.
type ChannelUpsertRequest struct {
	Name         string         `json:"name"`
	Platform     string         `json:"platform"`
	StoreID      string         `json:"storeId"`
	Region       string         `json:"region"`
	OwnerUUID    string         `json:"ownerUuid"`
	ApproverUUID string         `json:"approverUuid"`
	ChannelType  string         `json:"channelType"`
	Domain       string         `json:"domain"`
	Tags         []string       `json:"tags"`
	Contact      ContactPayload `json:"contact"`
}

// ContactPayload carries contact info from HTTP.
type ContactPayload struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// ToServiceInput converts HTTP requests to service DTOs.
func (req ChannelUpsertRequest) ToServiceInput() channelservice.CreateChannelInput {
	return channelservice.CreateChannelInput{
		Name:         req.Name,
		Platform:     req.Platform,
		StoreID:      req.StoreID,
		Region:       req.Region,
		OwnerUUID:    req.OwnerUUID,
		ApproverUUID: req.ApproverUUID,
		ChannelType:  req.ChannelType,
		Domain:       req.Domain,
		Tags:         req.Tags,
		Contact: channelservice.ChannelContactInput{
			Name:  req.Contact.Name,
			Phone: req.Contact.Phone,
			Email: req.Contact.Email,
		},
	}
}

// SubmitChannelRequest wraps submit payload.
type SubmitChannelRequest struct {
	Note               string `json:"note"`
	OfflineEvidenceURL string `json:"offlineEvidenceUrl"`
}

// ToServiceInput converts submit payload.
func (req SubmitChannelRequest) ToServiceInput() channelservice.SubmitChannelInput {
	return channelservice.SubmitChannelInput{
		Note:               req.Note,
		OfflineEvidenceURL: req.OfflineEvidenceURL,
	}
}

// ApprovalRequest wraps approve/reject payload.
type ApprovalRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func (req ApprovalRequest) ToServiceInput() channelservice.ApprovalDecisionInput {
	return channelservice.ApprovalDecisionInput{
		Decision: req.Decision,
		Reason:   req.Reason,
	}
}

// CredentialDTO returns sanitized credential info.
type CredentialDTO struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	Scope           []string   `json:"scope"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	LastRefreshedAt *time.Time `json:"last_refreshed_at,omitempty"`
	LastTestedAt    *time.Time `json:"last_tested_at,omitempty"`
	TestResult      any        `json:"test_result,omitempty"`
	AttachmentURL   string     `json:"attachment_url,omitempty"`
}

// CredentialUpsertRequest aligns with OpenAPI payload.
type CredentialUpsertRequest struct {
	Type          string         `json:"type" binding:"required"`
	Payload       map[string]any `json:"payload" binding:"required"`
	Scope         []string       `json:"scope"`
	ExpiresAt     *time.Time     `json:"expires_at"`
	Metadata      map[string]any `json:"metadata"`
	AttachmentURL string         `json:"attachment_url"`
}

// CredentialTestRequest captures manual test outcomes.
type CredentialTestRequest struct {
	Type      string         `json:"type" binding:"required"`
	Succeeded bool           `json:"succeeded"`
	Result    map[string]any `json:"result"`
}

// StrategyPayload mirrors strategy subsection from HTTP layer.
type StrategyPayload struct {
	PricebookID         string     `json:"pricebookId"`
	InventoryStrategyID string     `json:"inventoryStrategyId"`
	LogisticsStrategyID string     `json:"logisticsStrategyId"`
	CSSLAID             string     `json:"csSlaId"`
	FeeRate             *float64   `json:"feeRate"`
	SettlementCycle     string     `json:"settlementCycle"`
	PaymentTerms        string     `json:"paymentTerms"`
	Notes               string     `json:"notes"`
	EffectiveAt         *time.Time `json:"effectiveAt"`
}

// StrategyTeamPayload captures team assignments.
type StrategyTeamPayload struct {
	OwnerUUID    string   `json:"ownerUuid"`
	ApproverUUID string   `json:"approverUuid"`
	Operators    []string `json:"operators"`
}

// StrategyUpsertRequest wraps strategy/team payloads.
type StrategyUpsertRequest struct {
	Strategy StrategyPayload     `json:"strategy"`
	Team     StrategyTeamPayload `json:"team"`
}

// ToServiceInput converts payload into service DTO.
func (req StrategyUpsertRequest) ToServiceInput() channelservice.StrategyUpsertInput {
	return channelservice.StrategyUpsertInput{
		Strategy: channelservice.ChannelStrategyInput{
			PricebookID:         req.Strategy.PricebookID,
			InventoryStrategyID: req.Strategy.InventoryStrategyID,
			LogisticsStrategyID: req.Strategy.LogisticsStrategyID,
			CSSLAID:             req.Strategy.CSSLAID,
			FeeRate:             req.Strategy.FeeRate,
			SettlementCycle:     req.Strategy.SettlementCycle,
			PaymentTerms:        req.Strategy.PaymentTerms,
			Notes:               req.Strategy.Notes,
			EffectiveAt:         req.Strategy.EffectiveAt,
		},
		Team: channelservice.ChannelTeamInput{
			OwnerUUID:    req.Team.OwnerUUID,
			ApproverUUID: req.Team.ApproverUUID,
			Operators:    req.Team.Operators,
		},
	}
}

// StrategyDTO returned to clients for config snapshots.
type StrategyDTO struct {
	PricebookID         string  `json:"pricebookId,omitempty"`
	InventoryStrategyID string  `json:"inventoryStrategyId,omitempty"`
	LogisticsStrategyID string  `json:"logisticsStrategyId,omitempty"`
	CSSLAID             string  `json:"csSlaId,omitempty"`
	FeeRate             float64 `json:"feeRate"`
	SettlementCycle     string  `json:"settlementCycle,omitempty"`
	PaymentTerms        string  `json:"paymentTerms,omitempty"`
	Notes               string  `json:"notes,omitempty"`
	EffectiveAt         *string `json:"effectiveAt,omitempty"`
}

// TeamDTO exposes owner/approver/operators.
type TeamDTO struct {
	OwnerUUID    string   `json:"ownerUuid"`
	ApproverUUID string   `json:"approverUuid,omitempty"`
	Operators    []string `json:"operators,omitempty"`
}

// StrategyResponse returned by strategy endpoints.
type StrategyResponse struct {
	Strategy StrategyDTO `json:"strategy"`
	Team     TeamDTO     `json:"team"`
}

// OwnerListQuery describes search filters for owner dropdown.
type OwnerListQuery struct {
	Keyword string `form:"keyword"`
	Limit   int    `form:"limit"`
}

func mapStrategySnapshot(snapshot *channelservice.ChannelStrategySnapshot) StrategyResponse {
	if snapshot == nil {
		return StrategyResponse{}
	}
	var effective *string
	if snapshot.Config.EffectiveAt != nil {
		ts := snapshot.Config.EffectiveAt.UTC().Format(time.RFC3339)
		effective = &ts
	}
	strategy := StrategyDTO{
		PricebookID:         snapshot.Config.PricebookID,
		InventoryStrategyID: snapshot.Config.InventoryStrategyID,
		LogisticsStrategyID: snapshot.Config.LogisticsStrategyID,
		CSSLAID:             snapshot.Config.CSSLAID,
		FeeRate:             snapshot.Config.FeeRate,
		SettlementCycle:     snapshot.Config.SettlementCycle,
		PaymentTerms:        snapshot.Config.PaymentTerms,
		Notes:               snapshot.Config.Notes,
		EffectiveAt:         effective,
	}
	team := TeamDTO{
		OwnerUUID: snapshot.Team.OwnerUUID,
		Operators: snapshot.Team.Operators,
	}
	if trimmed := strings.TrimSpace(snapshot.Team.ApproverUUID); trimmed != "" {
		team.ApproverUUID = trimmed
	}
	return StrategyResponse{
		Strategy: strategy,
		Team:     team,
	}
}

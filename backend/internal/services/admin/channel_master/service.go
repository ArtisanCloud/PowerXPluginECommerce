package channel_master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditEmitter is a narrow contract for emitting structured audit events.
type AuditEmitter interface {
	EmitChannelAudit(ctx context.Context, action string, payload map[string]any) error
}

// DTOMapper converts database entities into transport-friendly payloads.
type DTOMapper interface {
	ToChannelSummary(channel *channelmodel.ChannelMaster) ChannelSummaryDTO
}

// ChannelSummaryDTO is a light-weight response used by admin APIs.
type ChannelSummaryDTO struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	StoreID         string                 `json:"storeId"`
	Platform        string                 `json:"platform"`
	Region          string                 `json:"region"`
	Status          string                 `json:"status"`
	OwnerUUID       string                 `json:"ownerUuid"`
	Tags            []string               `json:"tags"`
	ChannelType     string                 `json:"channelType"`
	ApprovalHistory []ApprovalHistoryEntry `json:"approvalHistory,omitempty"`
}

// ApprovalHistoryEntry records submission/decision lifecycle events.
type ApprovalHistoryEntry struct {
	Event    string `json:"event"`
	Actor    string `json:"actor"`
	Note     string `json:"note,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Decision string `json:"decision,omitempty"`
	At       string `json:"at"`
}

const (
	StatusDraft         = "draft"
	StatusPendingReview = "pending_review"
	StatusRejected      = "rejected"
	StatusUnauthorized  = "unauthorized"
	StatusAuthorized    = "authorized"
	StatusDisabled      = "disabled"
)

const (
	ChannelTypePlatformOAuth  = "platform_oauth"
	ChannelTypePlatformManual = "platform_manual"
	ChannelTypeOffline        = "offline"
)

const (
	metadataKeyLastSubmissionAt   = "last_submission_at"
	metadataKeySubmissionNote     = "submission_note"
	metadataKeyOfflineEvidenceURL = "offline_evidence_url"
	metadataKeyLastApprovalReason = "last_approval_reason"
	metadataKeyApprovalHistory    = "approval_history"
)

var (
	// ErrInvalidStatusTransition indicates the requested action does not align with the current state.
	ErrInvalidStatusTransition = errors.New("invalid channel status transition")
	// ErrInvalidChannelType indicates channel_type not supported by current flow.
	ErrInvalidChannelType = errors.New("invalid channel type")
	// ErrDuplicateStore indicates the store ID already exists within the tenant scope.
	ErrDuplicateStore = errors.New("channel store already exists for tenant")
)

var validChannelTypes = map[string]struct{}{
	ChannelTypePlatformOAuth:  {},
	ChannelTypePlatformManual: {},
	ChannelTypeOffline:        {},
}

// CreateChannelInput captures the minimal fields required to draft a channel.
type CreateChannelInput struct {
	Name         string              `json:"name"`
	Platform     string              `json:"platform"`
	StoreID      string              `json:"storeId"`
	Region       string              `json:"region"`
	OwnerUUID    string              `json:"ownerUuid"`
	ApproverUUID string              `json:"approverUuid"`
	ChannelType  string              `json:"channelType"`
	Domain       string              `json:"domain"`
	Tags         []string            `json:"tags"`
	Contact      ChannelContactInput `json:"contact"`
}

// ChannelContactInput captures contact metadata.
type ChannelContactInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// SubmitChannelInput carries submission metadata such as remark/offline attachments.
type SubmitChannelInput struct {
	Note               string `json:"note"`
	OfflineEvidenceURL string `json:"offlineEvidenceUrl"`
}

// ApprovalDecisionInput encapsulates approve/reject payloads.
type ApprovalDecisionInput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// Normalize trims whitespace and deduplicates tags.
func (in *CreateChannelInput) Normalize() {
	if in == nil {
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Platform = strings.TrimSpace(in.Platform)
	in.StoreID = strings.TrimSpace(in.StoreID)
	in.Region = strings.TrimSpace(in.Region)
	in.OwnerUUID = strings.TrimSpace(in.OwnerUUID)
	in.ApproverUUID = strings.TrimSpace(in.ApproverUUID)
	in.ChannelType = strings.TrimSpace(strings.ToLower(in.ChannelType))
	in.Domain = strings.TrimSpace(in.Domain)
	in.Contact.Normalize()
	if len(in.Tags) == 0 {
		return
	}
	set := make(map[string]struct{}, len(in.Tags))
	for _, tag := range in.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			set[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	in.Tags = in.Tags[:0]
	for tag := range set {
		in.Tags = append(in.Tags, tag)
	}
	sort.Strings(in.Tags)
}

// Validate ensures required fields are present before persistence.
func (in *CreateChannelInput) Validate() error {
	if in == nil {
		return errors.New("input is required")
	}
	in.Normalize()
	var missing []string
	if in.Name == "" {
		missing = append(missing, "name")
	}
	if in.Platform == "" {
		missing = append(missing, "platform")
	}
	if in.StoreID == "" {
		missing = append(missing, "storeId")
	}
	if in.Region == "" {
		missing = append(missing, "region")
	}
	if in.OwnerUUID == "" {
		missing = append(missing, "ownerUuid")
	}
	if in.ChannelType == "" {
		missing = append(missing, "channelType")
	}
	if contactErrs := in.Contact.validateMissing(); len(contactErrs) > 0 {
		missing = append(missing, contactErrs...)
	}
	if len(missing) > 0 {
		return errors.New("missing required fields: " + strings.Join(missing, ","))
	}
	if err := validateChannelType(in.ChannelType); err != nil {
		return err
	}
	return nil
}

func (c *ChannelContactInput) Normalize() {
	if c == nil {
		return
	}
	c.Name = strings.TrimSpace(c.Name)
	c.Phone = strings.TrimSpace(c.Phone)
	c.Email = strings.TrimSpace(c.Email)
}

func (c *ChannelContactInput) validateMissing() []string {
	if c == nil {
		return []string{"contact.name", "contact.phone", "contact.email"}
	}
	var missing []string
	if c.Name == "" {
		missing = append(missing, "contact.name")
	}
	if c.Phone == "" {
		missing = append(missing, "contact.phone")
	}
	if c.Email == "" {
		missing = append(missing, "contact.email")
	}
	return missing
}

// Normalize trims optional fields for submit payload.
func (in *SubmitChannelInput) Normalize() {
	if in == nil {
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	in.OfflineEvidenceURL = strings.TrimSpace(in.OfflineEvidenceURL)
}

// Validate ensures offline channels provide proofs.
func (in *SubmitChannelInput) Validate(channelType string) error {
	if in == nil {
		return nil
	}
	if channelType == ChannelTypeOffline && in.OfflineEvidenceURL == "" {
		return errors.New("offline channels require offlineEvidenceUrl during submission")
	}
	return nil
}

// Normalize trims decision payload.
func (in *ApprovalDecisionInput) Normalize() {
	if in == nil {
		return
	}
	in.Decision = strings.TrimSpace(strings.ToLower(in.Decision))
	in.Reason = strings.TrimSpace(in.Reason)
}

// Validate ensures decision payload correctness.
func (in *ApprovalDecisionInput) Validate() error {
	if in == nil {
		return errors.New("decision payload required")
	}
	in.Normalize()
	if in.Decision != "approve" && in.Decision != "reject" {
		return errors.New("decision must be approve or reject")
	}
	if in.Decision == "reject" && in.Reason == "" {
		return errors.New("rejection reason required")
	}
	return nil
}

// Service coordinates repositories, mappers and observability helpers for channel master flows.
type Service struct {
	deps    *app.Deps
	repo    *channelrepo.ChannelMasterRepository
	mapper  DTOMapper
	audit   AuditEmitter
	metrics *channelobs.Metrics
	logger  *logrus.Entry
}

// NewService wires required collaborators; mapper/audit emitters can be nil for now.
func NewService(deps *app.Deps, mapper DTOMapper, audit AuditEmitter, metrics *channelobs.Metrics) *Service {
	if deps == nil || deps.DB == nil {
		panic("channel master service requires database dependency")
	}
	if mapper == nil {
		mapper = defaultMapper{}
	}
	logger := deps.RuntimeLogger(context.TODO(), "channel-master-service", nil)
	if metrics == nil {
		metrics = channelobs.NewMetrics(logger)
	}
	return &Service{
		deps:    deps,
		repo:    channelrepo.NewChannelMasterRepository(deps.DB),
		mapper:  mapper,
		audit:   audit,
		metrics: metrics,
		logger:  logger,
	}
}

// List returns a paginated result mapped into DTOs.
func (s *Service) List(ctx context.Context, filters channelrepo.ChannelListFilters) (*repo.Page[[]ChannelSummaryDTO], error) {
	page, err := s.repo.FindPage(ctx, filters)
	if err != nil {
		return nil, err
	}
	items := make([]ChannelSummaryDTO, 0, len(page.List))
	for _, channel := range page.List {
		items = append(items, s.mapper.ToChannelSummary(channel))
	}
	return &repo.Page[[]ChannelSummaryDTO]{
		List:      items,
		PageIndex: filters.Page,
		PageSize:  filters.PageSize,
		Total:     page.Total,
	}, nil
}

// CreateDraft persists a new channel placeholder and emits audit events.
func (s *Service) CreateDraft(ctx context.Context, input CreateChannelInput) (*ChannelSummaryDTO, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if exists, err := s.repo.FindByStoreID(ctx, input.StoreID); err != nil {
		return nil, err
	} else if exists != nil {
		return nil, ErrDuplicateStore
	}
	actor := actorFromContext(ctx)
	entity := &channelmodel.ChannelMaster{
		ID:        utils.NewUUID(),
		Status:    StatusDraft,
		CreatedBy: actor,
		UpdatedBy: actor,
	}
	applyUpsertFields(entity, input)
	saved, err := s.repo.Upsert(ctx, entity)
	if err != nil {
		return nil, err
	}
	s.emitAudit(ctx, "channel.created", saved.ID, map[string]any{
		"platform":     saved.Platform,
		"owner_uuid":   saved.OwnerUUID,
		"channel_type": saved.ChannelType,
	})
	summary := s.mapper.ToChannelSummary(saved)
	return &summary, nil
}

// UpdateDraft mutates draft/rejected channels before submission.
func (s *Service) UpdateDraft(ctx context.Context, channelID string, input CreateChannelInput) (*ChannelSummaryDTO, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	channel, err := s.repo.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if !canEditChannel(channel.Status) {
		return nil, ErrInvalidStatusTransition
	}
	if channel.StoreID != input.StoreID {
		if exists, err := s.repo.FindByStoreID(ctx, input.StoreID); err != nil {
			return nil, err
		} else if exists != nil && exists.ID != channel.ID {
			return nil, ErrDuplicateStore
		}
	}
	applyUpsertFields(channel, input)
	channel.UpdatedBy = actorFromContext(ctx)
	saved, err := s.repo.Save(ctx, channel)
	if err != nil {
		return nil, err
	}
	s.emitAudit(ctx, "channel.updated", saved.ID, map[string]any{
		"status": saved.Status,
	})
	summary := s.mapper.ToChannelSummary(saved)
	return &summary, nil
}

// Submit transitions a channel into pending_review for approval.
func (s *Service) Submit(ctx context.Context, channelID string, input SubmitChannelInput) (*ChannelSummaryDTO, error) {
	input.Normalize()
	channel, err := s.repo.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if !canSubmitChannel(channel.Status) {
		return nil, ErrInvalidStatusTransition
	}
	if err := input.Validate(channel.ChannelType); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	channel.Status = StatusPendingReview
	channel.UpdatedBy = actorFromContext(ctx)
	channel.Metadata = setMetadataValue(channel.Metadata, metadataKeyLastSubmissionAt, now.Format(time.RFC3339Nano))
	note := strings.TrimSpace(input.Note)
	if note != "" {
		channel.Metadata = setMetadataValue(channel.Metadata, metadataKeySubmissionNote, note)
	} else {
		channel.Metadata = setMetadataValue(channel.Metadata, metadataKeySubmissionNote, nil)
	}
	if input.OfflineEvidenceURL != "" {
		channel.Metadata = setMetadataValue(channel.Metadata, metadataKeyOfflineEvidenceURL, input.OfflineEvidenceURL)
	} else {
		channel.Metadata = setMetadataValue(channel.Metadata, metadataKeyOfflineEvidenceURL, nil)
	}
	entry := ApprovalHistoryEntry{
		Event: "submitted",
		Actor: actorFromContext(ctx),
		At:    now.Format(time.RFC3339Nano),
	}
	if note != "" {
		entry.Note = note
	}
	channel.Metadata = appendApprovalHistory(channel.Metadata, entry)

	saved, err := s.repo.Save(ctx, channel)
	if err != nil {
		return nil, err
	}
	s.emitAudit(ctx, "channel.submitted", saved.ID, map[string]any{
		"channel_type": saved.ChannelType,
	})
	summary := s.mapper.ToChannelSummary(saved)
	return &summary, nil
}

// ProcessApproval handles approve/reject actions.
func (s *Service) ProcessApproval(ctx context.Context, channelID string, decision ApprovalDecisionInput) (*ChannelSummaryDTO, error) {
	if err := decision.Validate(); err != nil {
		return nil, err
	}
	channel, err := s.repo.FindByID(ctx, channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if strings.TrimSpace(channel.Status) != StatusPendingReview {
		return nil, ErrInvalidStatusTransition
	}
	submittedAt, hasSubmittedAt := metadataTime(channel.Metadata, metadataKeyLastSubmissionAt)
	channel.UpdatedBy = actorFromContext(ctx)
	action := "channel.rejected"
	if decision.Decision == "approve" {
		channel.Status = StatusUnauthorized
		action = "channel.approved"
	} else {
		channel.Status = StatusRejected
	}
	reason := strings.TrimSpace(decision.Reason)
	if reason != "" {
		channel.Metadata = setMetadataValue(channel.Metadata, metadataKeyLastApprovalReason, reason)
	}
	channel.Metadata = setMetadataValue(channel.Metadata, metadataKeySubmissionNote, nil)
	decisionEntry := ApprovalHistoryEntry{
		Event:    decision.Decision,
		Decision: decision.Decision,
		Actor:    actorFromContext(ctx),
		At:       time.Now().UTC().Format(time.RFC3339Nano),
	}
	if reason != "" {
		decisionEntry.Reason = reason
	}
	channel.Metadata = appendApprovalHistory(channel.Metadata, decisionEntry)
	saved, err := s.repo.Save(ctx, channel)
	if err != nil {
		return nil, err
	}
	if decision.Decision == "approve" && hasSubmittedAt {
		s.metrics.ObserveApprovalLead(time.Since(submittedAt))
	}
	s.emitAudit(ctx, action, saved.ID, map[string]any{
		"decision": decision.Decision,
		"reason":   decision.Reason,
	})
	summary := s.mapper.ToChannelSummary(saved)
	return &summary, nil
}

// defaultMapper converts models to basic DTOs.
type defaultMapper struct{}

func (defaultMapper) ToChannelSummary(channel *channelmodel.ChannelMaster) ChannelSummaryDTO {
	if channel == nil {
		return ChannelSummaryDTO{}
	}
	return ChannelSummaryDTO{
		ID:              channel.ID,
		Name:            channel.Name,
		StoreID:         channel.StoreID,
		Platform:        channel.Platform,
		Region:          channel.Region,
		Status:          channel.Status,
		OwnerUUID:       channel.OwnerUUID,
		Tags:            []string(channel.Tags),
		ChannelType:     channel.ChannelType,
		ApprovalHistory: metadataApprovalHistory(channel.Metadata),
	}
}

func applyUpsertFields(channel *channelmodel.ChannelMaster, input CreateChannelInput) {
	if channel == nil {
		return
	}
	channel.Name = input.Name
	channel.Platform = input.Platform
	channel.StoreID = input.StoreID
	channel.Region = input.Region
	channel.OwnerUUID = input.OwnerUUID
	channel.ChannelType = input.ChannelType
	channel.Domain = input.Domain
	channel.Tags = pq.StringArray(input.Tags)
	channel.ContactName = input.Contact.Name
	channel.ContactPhone = input.Contact.Phone
	channel.ContactEmail = input.Contact.Email
	assignApprover(channel, input.ApproverUUID)
}

func assignApprover(channel *channelmodel.ChannelMaster, approver string) {
	if channel == nil {
		return
	}
	trimmed := strings.TrimSpace(approver)
	if trimmed == "" {
		channel.ApproverUUID = nil
		return
	}
	channel.ApproverUUID = ptrString(trimmed)
}

func canEditChannel(status string) bool {
	switch strings.TrimSpace(status) {
	case StatusDraft, StatusRejected:
		return true
	default:
		return false
	}
}

func canSubmitChannel(status string) bool {
	switch strings.TrimSpace(status) {
	case StatusDraft, StatusRejected:
		return true
	default:
		return false
	}
}

func validateChannelType(channelType string) error {
	if _, ok := validChannelTypes[strings.TrimSpace(channelType)]; !ok {
		return ErrInvalidChannelType
	}
	return nil
}

func appendApprovalHistory(meta datatypes.JSON, entry ApprovalHistoryEntry) datatypes.JSON {
	if strings.TrimSpace(entry.At) == "" {
		entry.At = time.Now().UTC().Format(time.RFC3339Nano)
	}
	history := metadataApprovalHistory(meta)
	history = append(history, entry)
	const maxEntries = 50
	if len(history) > maxEntries {
		history = history[len(history)-maxEntries:]
	}
	return setMetadataValue(meta, metadataKeyApprovalHistory, history)
}

func metadataApprovalHistory(meta datatypes.JSON) []ApprovalHistoryEntry {
	data := metadataAsMap(meta)
	raw, ok := data[metadataKeyApprovalHistory]
	if !ok {
		return nil
	}
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var history []ApprovalHistoryEntry
	if err := json.Unmarshal(bytes, &history); err != nil {
		return nil
	}
	return history
}

func setMetadataValue(meta datatypes.JSON, key string, value any) datatypes.JSON {
	data := metadataAsMap(meta)
	if value == nil {
		delete(data, key)
	} else {
		data[key] = value
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return meta
	}
	return datatypes.JSON(raw)
}

func metadataAsMap(meta datatypes.JSON) map[string]any {
	var data map[string]any
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &data)
	}
	if data == nil {
		data = make(map[string]any)
	}
	return data
}

func metadataTime(meta datatypes.JSON, key string) (time.Time, bool) {
	data := metadataAsMap(meta)
	raw, ok := data[key]
	if !ok {
		return time.Time{}, false
	}
	switch v := raw.(type) {
	case string:
		if ts, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}

func ptrString(val string) *string {
	v := val
	return &v
}

func (s *Service) emitAudit(ctx context.Context, action, channelID string, payload map[string]any) {
	if s.audit == nil {
		return
	}
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["channel_id"] = channelID
	if err := s.audit.EmitChannelAudit(ctx, action, payload); err != nil && s.logger != nil {
		s.logger.WithError(err).Warn("failed to emit channel audit")
	}
}

func actorFromContext(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}
	if tc, ok := authx.TenantContextFromContext(ctx); ok {
		if tc.UserID > 0 {
			return fmt.Sprintf("user:%d", tc.UserID)
		}
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && tid != "" {
		return "tenant:" + tid
	}
	return "system"
}

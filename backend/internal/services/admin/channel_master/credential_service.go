package channel_master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/security/encryption"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CredentialService manages channel credential flows.
type CredentialService struct {
	deps         *app.Deps
	repo         *channelrepo.ChannelCredentialRepository
	alertRepo    *channelrepo.ChannelAlertRepository
	alertEmitter *channelobs.AlertEmitter
	logger       *logrus.Entry
	keyProvider  encryption.KeyProvider
	metrics      *channelobs.Metrics
}

// CredentialDTO trims sensitive fields for transport.
type CredentialDTO struct {
	ID              string         `json:"id"`
	Type            string         `json:"type"`
	Status          string         `json:"status"`
	Scope           []string       `json:"scope"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	LastRefreshedAt *time.Time     `json:"last_refreshed_at,omitempty"`
	LastTestedAt    *time.Time     `json:"last_tested_at,omitempty"`
	TestResult      map[string]any `json:"test_result,omitempty"`
	AttachmentURL   string         `json:"attachment_url,omitempty"`
}

// CredentialUpsertInput captures payload to encrypt.
type CredentialUpsertInput struct {
	Type          string
	Payload       map[string]any
	Scope         []string
	ExpiresAt     *time.Time
	Metadata      map[string]any
	AttachmentURL string
}

// CredentialTestInput stores manual test results.
type CredentialTestInput struct {
	Type      string
	Result    map[string]any
	Succeeded bool
}

// NewCredentialService constructs service.
func NewCredentialService(deps *app.Deps, alertEmitter *channelobs.AlertEmitter, metrics *channelobs.Metrics) *CredentialService {
	if deps == nil || deps.DB == nil {
		panic("credential service requires database dependency")
	}
	logger := deps.RuntimeLogger(context.TODO(), "channel-credential-service", nil)
	if metrics == nil {
		metrics = channelobs.NewMetrics(logger)
	}
	svc := &CredentialService{
		deps:         deps,
		repo:         channelrepo.NewChannelCredentialRepository(deps.DB),
		alertRepo:    channelrepo.NewChannelAlertRepository(deps.DB),
		alertEmitter: alertEmitter,
		logger:       logger,
		metrics:      metrics,
	}
	svc.keyProvider = encryption.NewStaticKeyProvider(svc.secretKey())
	return svc
}

// List returns sanitized credentials.
func (s *CredentialService) List(ctx context.Context, channelID string) ([]CredentialDTO, error) {
	creds, err := s.repo.ListByChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	s.observeCredentialCoverage(creds)
	output := make([]CredentialDTO, 0, len(creds))
	for _, cred := range creds {
		output = append(output, mapCredentialDTO(cred))
	}
	return output, nil
}

// Upsert encrypts payload and stores credential.
func (s *CredentialService) Upsert(ctx context.Context, channelID string, input CredentialUpsertInput) (*CredentialDTO, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, errors.New("channelID is required")
	}
	if err := validateCredentialType(input.Type); err != nil {
		return nil, err
	}
	payloadBytes, err := json.Marshal(input.Payload)
	if err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}
	aad := []byte(fmt.Sprintf("channel:%s|type:%s", channelID, input.Type))
	if s.keyProvider == nil {
		s.keyProvider = encryption.NewStaticKeyProvider(s.secretKey())
	}
	env, err := encryption.EncryptEnvelope(ctx, s.keyProvider, payloadBytes, aad)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	credential := &channelmodel.ChannelCredential{
		ChannelID:        channelID,
		Type:             strings.TrimSpace(strings.ToLower(input.Type)),
		Scope:            pq.StringArray(normalizeScope(input.Scope)),
		SecretCiphertext: env.Ciphertext,
		SecretNonce:      env.CipherNonce,
		DekCiphertext:    env.DekCiphertext,
		DekNonce:         env.DekNonce,
		KeyVersion:       env.KeyVersion,
		Algorithm:        env.Algorithm,
		ExpiresAt:        input.ExpiresAt,
		Metadata:         jsonMap(input.Metadata),
		LastRefreshedAt:  &now,
		LastRotatedAt:    &env.CreatedAt,
		Status:           deriveCredentialStatus(input.ExpiresAt),
		AttachmentURL:    strings.TrimSpace(input.AttachmentURL),
		CreatedBy:        actorFromContext(ctx),
		UpdatedBy:        actorFromContext(ctx),
	}
	saved, err := s.repo.Upsert(ctx, credential)
	if err != nil {
		return nil, err
	}
	s.handleStatusAlert(ctx, saved)
	s.observeCredentialCoverage([]*channelmodel.ChannelCredential{saved})
	dto := mapCredentialDTO(saved)
	return &dto, nil
}

// RecordTestResult updates test outcome.
func (s *CredentialService) RecordTestResult(ctx context.Context, channelID string, input CredentialTestInput) (*CredentialDTO, error) {
	if strings.TrimSpace(channelID) == "" {
		return nil, errors.New("channelID required")
	}
	if err := validateCredentialType(input.Type); err != nil {
		return nil, err
	}
	creds, err := s.repo.ListByChannel(ctx, channelID)
	if err != nil {
		return nil, err
	}
	var target *channelmodel.ChannelCredential
	for _, cred := range creds {
		if cred.Type == strings.TrimSpace(strings.ToLower(input.Type)) {
			target = cred
			break
		}
	}
	if target == nil {
		return nil, gorm.ErrRecordNotFound
	}
	now := time.Now().UTC()
	target.LastTestedAt = &now
	target.TestResult = jsonMap(input.Result)
	if input.Succeeded {
		target.Status = deriveCredentialStatus(target.ExpiresAt)
	} else {
		target.Status = "test_failed"
	}
	target.UpdatedBy = actorFromContext(ctx)
	if _, err := s.repo.Upsert(ctx, target); err != nil {
		return nil, err
	}
	s.handleStatusAlert(ctx, target)
	s.handleTestAlert(ctx, channelID, input)
	s.observeCredentialCoverage([]*channelmodel.ChannelCredential{target})
	dto := mapCredentialDTO(target)
	return &dto, nil
}

func (s *CredentialService) secretKey() string {
	if s.deps == nil || s.deps.Config == nil || strings.TrimSpace(s.deps.Config.Server.SecretKey) == "" {
		return "dev-only-change-me"
	}
	return s.deps.Config.Server.SecretKey
}

func mapCredentialDTO(entity *channelmodel.ChannelCredential) CredentialDTO {
	if entity == nil {
		return CredentialDTO{}
	}
	return CredentialDTO{
		ID:              entity.ID,
		Type:            entity.Type,
		Status:          entity.Status,
		Scope:           []string(entity.Scope),
		ExpiresAt:       entity.ExpiresAt,
		LastRefreshedAt: entity.LastRefreshedAt,
		LastTestedAt:    entity.LastTestedAt,
		TestResult:      jsonToMap(entity.TestResult),
		AttachmentURL:   entity.AttachmentURL,
	}
}

func (s *CredentialService) observeCredentialCoverage(creds []*channelmodel.ChannelCredential) {
	if s == nil || s.metrics == nil {
		return
	}
	for _, cred := range creds {
		if cred == nil {
			continue
		}
		s.metrics.ObserveCredentialCoverage(isCredentialHealthyStatus(cred.Status))
	}
}

func isCredentialHealthyStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "valid", "expiring":
		return true
	default:
		return false
	}
}

func deriveCredentialStatus(expiresAt *time.Time) string {
	if expiresAt == nil {
		return "valid"
	}
	now := time.Now()
	if now.After(*expiresAt) {
		return "expired"
	}
	if expiresAt.Sub(now) <= 7*24*time.Hour {
		return "expiring"
	}
	return "valid"
}

func validateCredentialType(val string) error {
	switch strings.TrimSpace(strings.ToLower(val)) {
	case "oauth", "api_key", "offline":
		return nil
	default:
		return fmt.Errorf("unsupported credential type: %s", val)
	}
}

func normalizeScope(scope []string) []string {
	var cleaned []string
	for _, s := range scope {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func jsonMap(m map[string]any) datatypes.JSON {
	if len(m) == 0 {
		return nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return datatypes.JSON(raw)
}

func jsonToMap(data datatypes.JSON) map[string]any {
	if len(data) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func (s *CredentialService) handleStatusAlert(ctx context.Context, cred *channelmodel.ChannelCredential) {
	if s == nil || cred == nil {
		return
	}
	expiryStatus := deriveCredentialStatus(cred.ExpiresAt)
	meta := map[string]any{
		"credential_id": cred.ID,
	}
	if cred.ExpiresAt != nil {
		meta["expires_at"] = cred.ExpiresAt.Format(time.RFC3339)
	}
	switch expiryStatus {
	case "expired":
		s.raiseAlert(ctx, cred.ChannelID, channelmodel.AlertTypeCredentialExpiring, "critical", "渠道凭证已过期", "该渠道凭证已失效，请尽快重新授权。", meta)
	case "expiring":
		s.raiseAlert(ctx, cred.ChannelID, channelmodel.AlertTypeCredentialExpiring, "warning", "渠道凭证即将到期", "渠道凭证将在 7 天内到期，请安排续期。", meta)
	default:
		s.resolveAlert(ctx, cred.ChannelID, channelmodel.AlertTypeCredentialExpiring)
	}
}

func (s *CredentialService) handleTestAlert(ctx context.Context, channelID string, input CredentialTestInput) {
	if s == nil {
		return
	}
	if input.Succeeded {
		s.resolveAlert(ctx, channelID, channelmodel.AlertTypeCredentialTestFailed)
		return
	}
	meta := map[string]any{
		"type":   input.Type,
		"result": input.Result,
	}
	s.raiseAlert(ctx, channelID, channelmodel.AlertTypeCredentialTestFailed, "error", "渠道凭证巡检失败", "最近一次凭证巡检失败，请排查凭证有效性。", meta)
}

func (s *CredentialService) raiseAlert(ctx context.Context, channelID, alertType, severity, title, description string, metadata map[string]any) {
	if s.alertRepo == nil {
		return
	}
	alert := &channelmodel.ChannelAlert{
		ChannelID:   channelID,
		Type:        alertType,
		Severity:    severity,
		Title:       title,
		Description: description,
		Metadata:    jsonMap(metadata),
		TriggeredAt: time.Now().UTC(),
		Status:      "open",
	}
	if _, err := s.alertRepo.UpsertStatus(ctx, alert); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"channel_id": channelID,
			"type":       alertType,
		}).Warn("failed to persist channel alert")
		return
	}
	if s.alertEmitter != nil {
		payload := map[string]any{
			"channel_id": channelID,
			"type":       alertType,
			"severity":   severity,
		}
		if alert.Metadata != nil {
			payload["metadata"] = jsonToMap(alert.Metadata)
		}
		s.alertEmitter.Emit(ctx, fmt.Sprintf("channel.%s", alertType), payload)
	}
}

func (s *CredentialService) resolveAlert(ctx context.Context, channelID, alertType string) {
	if s.alertRepo == nil {
		return
	}
	if err := s.alertRepo.Resolve(ctx, channelID, alertType); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"channel_id": channelID,
			"type":       alertType,
		}).Warn("failed to resolve channel alert")
	}
}

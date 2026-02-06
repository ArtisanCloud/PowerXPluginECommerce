package membership

import (
	"context"
	"errors"
	"strings"
	"time"

	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	membershipRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/membership"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

type Service struct {
	deps        *app.Deps
	tierRepo    *membershipRepo.TierRepository
	benefitRepo *membershipRepo.BenefitRepository
	entRepo     *membershipRepo.EntitlementRepository
	tokenRepo   *membershipRepo.TokenAccountRepository
	tokenTxRepo *membershipRepo.TokenTransactionRepository
}

func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("membership admin service requires initialized DB dependency")
	}
	return &Service{
		deps:        deps,
		tierRepo:    membershipRepo.NewTierRepository(deps.DB),
		benefitRepo: membershipRepo.NewBenefitRepository(deps.DB),
		entRepo:     membershipRepo.NewEntitlementRepository(deps.DB),
		tokenRepo:   membershipRepo.NewTokenAccountRepository(deps.DB),
		tokenTxRepo: membershipRepo.NewTokenTransactionRepository(deps.DB),
	}
}

func (s *Service) ListTiers(ctx context.Context, tenantUUID string) ([]membershipModel.MembershipTier, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant uuid required")
	}
	var tiers []membershipModel.MembershipTier
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&tiers).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return tiers, nil
}

func (s *Service) ListBenefits(ctx context.Context, tenantUUID string) ([]membershipModel.MembershipBenefit, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant uuid required")
	}
	var benefits []membershipModel.MembershipBenefit
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&benefits).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return benefits, nil
}

func (s *Service) GrantEntitlement(ctx context.Context, tenantUUID string, req GrantEntitlementRequest) (*membershipModel.Entitlement, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" || strings.TrimSpace(req.CustomerID) == "" {
		return nil, errors.New("tenant_uuid/customer_id required")
	}
	serviceCode := strings.TrimSpace(req.ServiceCode)
	if serviceCode == "" {
		return nil, errors.New("service_code required")
	}
	if req.Quantity == 0 {
		return nil, errors.New("quantity required")
	}
	stackPolicy := strings.TrimSpace(req.StackPolicy)
	if stackPolicy == "" {
		stackPolicy = "stack"
	}
	sourceID := strings.TrimSpace(req.SourceID)
	if sourceID == "" {
		sourceID = utils.NewUUID()
	}
	ent := &membershipModel.Entitlement{
		ID:          utils.NewUUID(),
		TenantUUID:  tenantUUID,
		CustomerID:  strings.TrimSpace(req.CustomerID),
		ServiceCode: serviceCode,
		Quantity:    req.Quantity,
		ValidFrom:   req.ValidFrom,
		ValidTo:     req.ValidTo,
		StackPolicy: stackPolicy,
		SourceType:  "manual",
		SourceID:    sourceID,
	}
	if err := s.deps.DB.WithContext(ctx).Create(ent).Error; err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_entitlement_grant",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "entitlement",
		TargetID:   ent.ID,
		Result:     "created",
		Reason:     strings.TrimSpace(req.Reason),
		Metadata: map[string]any{
			"customer_id": ent.CustomerID,
			"service_code": ent.ServiceCode,
			"quantity":     ent.Quantity,
			"source_id":    ent.SourceID,
		},
		EmittedAt: time.Now().UTC(),
	})
	return ent, nil
}

func (s *Service) RevokeEntitlement(ctx context.Context, tenantUUID string, req RevokeEntitlementRequest) error {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" || strings.TrimSpace(req.EntitlementID) == "" {
		return errors.New("tenant_uuid/entitlement_id required")
	}
	var ent membershipModel.Entitlement
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(req.EntitlementID)).
		First(&ent).Error; err != nil {
		return err
	}
	if err := s.deps.DB.WithContext(ctx).Delete(&ent).Error; err != nil {
		return err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_entitlement_revoke",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "entitlement",
		TargetID:   ent.ID,
		Result:     "revoked",
		Reason:     strings.TrimSpace(req.Reason),
		Metadata: map[string]any{
			"customer_id": ent.CustomerID,
			"service_code": ent.ServiceCode,
			"quantity":     ent.Quantity,
		},
		EmittedAt: time.Now().UTC(),
	})
	return nil
}

func (s *Service) AdjustToken(ctx context.Context, tenantUUID string, req AdjustTokenRequest) (*membershipModel.TokenAccount, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" || strings.TrimSpace(req.CustomerID) == "" {
		return nil, errors.New("tenant_uuid/customer_id required")
	}
	tokenCode := strings.TrimSpace(req.TokenCode)
	if tokenCode == "" {
		return nil, errors.New("token_code required")
	}
	if req.Delta == 0 {
		return nil, errors.New("delta required")
	}
	var account membershipModel.TokenAccount
	err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND token_code = ?", tenantUUID, strings.TrimSpace(req.CustomerID), tokenCode).
		First(&account).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	now := time.Now().UTC()
	if err == gorm.ErrRecordNotFound {
		if req.Delta < 0 {
			return nil, errors.New("balance cannot be negative")
		}
		account = membershipModel.TokenAccount{
			ID:         utils.NewUUID(),
			TenantUUID: tenantUUID,
			CustomerID: strings.TrimSpace(req.CustomerID),
			TokenCode:  tokenCode,
			Balance:    req.Delta,
		}
		if err := s.deps.DB.WithContext(ctx).Create(&account).Error; err != nil {
			return nil, err
		}
	} else {
		next := account.Balance + req.Delta
		if next < 0 {
			return nil, errors.New("balance cannot be negative")
		}
		if err := s.deps.DB.WithContext(ctx).
			Model(&membershipModel.TokenAccount{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, account.ID).
			Updates(map[string]any{
				"balance":    next,
				"updated_at": now,
			}).Error; err != nil {
			return nil, err
		}
		account.Balance = next
	}
	tx := membershipModel.TokenTransaction{
		ID:         utils.NewUUID(),
		TenantUUID: tenantUUID,
		CustomerID: strings.TrimSpace(req.CustomerID),
		TokenCode:  tokenCode,
		Delta:      req.Delta,
		SourceType: "manual",
		SourceID:   strings.TrimSpace(req.SourceID),
	}
	if tx.SourceID == "" {
		tx.SourceID = utils.NewUUID()
	}
	if err := s.deps.DB.WithContext(ctx).Create(&tx).Error; err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_token_adjust",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "token_account",
		TargetID:   account.ID,
		Result:     "adjusted",
		Reason:     strings.TrimSpace(req.Reason),
		Metadata: map[string]any{
			"customer_id": account.CustomerID,
			"token_code":  account.TokenCode,
			"delta":       req.Delta,
			"balance":     account.Balance,
			"txn_id":      tx.ID,
		},
		EmittedAt: now,
	})
	return &account, nil
}

func (s *Service) emitAudit(ctx context.Context, tenantUUID string, evt paymentslogger.AuditEvent) {
	if s == nil || s.deps == nil {
		return
	}
	entry := s.deps.RuntimeLogger(ctx, "admin_membership", nil)
	if entry == nil {
		return
	}
	logger := paymentslogger.NewLogger(entry)
	logger.EmitAudit(evt)
}

package membership

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
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

func (s *Service) CreateTier(ctx context.Context, tenantUUID string, req CreateTierRequest) (*membershipModel.MembershipTier, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant_uuid required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, errors.New("code required")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}

	tier := &membershipModel.MembershipTier{
		ID:         utils.NewUUID(),
		TenantUUID: tenantUUID,
		Name:       name,
		Code:       code,
		Status:     status,
		Rules:      req.Rules,
	}
	if len(tier.Rules) == 0 {
		tier.Rules = []byte("{}")
	}
	if err := s.deps.DB.WithContext(ctx).Create(tier).Error; err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_membership_tier_create",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "membership_tier",
		TargetID:   tier.ID,
		Result:     "created",
		Metadata: map[string]any{
			"name":   tier.Name,
			"code":   tier.Code,
			"status": tier.Status,
		},
		EmittedAt: time.Now().UTC(),
	})
	return tier, nil
}

func (s *Service) CreateBenefit(ctx context.Context, tenantUUID string, req CreateBenefitRequest) (*membershipModel.MembershipBenefit, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant_uuid required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name required")
	}
	benefitType := strings.TrimSpace(req.Type)
	if benefitType == "" {
		benefitType = "single"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	benefit := &membershipModel.MembershipBenefit{
		ID:         utils.NewUUID(),
		TenantUUID: tenantUUID,
		Name:       name,
		Type:       benefitType,
		Status:     status,
		Items:      req.Items,
	}
	if len(benefit.Items) == 0 {
		benefit.Items = []byte("[]")
	}
	if err := s.deps.DB.WithContext(ctx).Create(benefit).Error; err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_membership_benefit_create",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "membership_benefit",
		TargetID:   benefit.ID,
		Result:     "created",
		Metadata: map[string]any{
			"name":   benefit.Name,
			"type":   benefit.Type,
			"status": benefit.Status,
		},
		EmittedAt: time.Now().UTC(),
	})
	return benefit, nil
}

func (s *Service) UpdateTierStatus(ctx context.Context, tenantUUID string, tierID string, req UpdateTierStatusRequest) (*membershipModel.MembershipTier, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	tierID = strings.TrimSpace(tierID)
	if tenantUUID == "" || tierID == "" {
		return nil, errors.New("tenant_uuid/tier_id required")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return nil, errors.New("status required")
	}
	var tier membershipModel.MembershipTier
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, tierID).
		First(&tier).Error; err != nil {
		return nil, err
	}
	tier.Status = status
	if err := s.deps.DB.WithContext(ctx).Save(&tier).Error; err != nil {
		return nil, err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_membership_tier_status_update",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "membership_tier",
		TargetID:   tier.ID,
		Result:     "updated",
		Metadata: map[string]any{
			"name":   tier.Name,
			"code":   tier.Code,
			"status": tier.Status,
		},
		EmittedAt: time.Now().UTC(),
	})
	return &tier, nil
}

func (s *Service) DeleteTier(ctx context.Context, tenantUUID string, tierID string, actorID string) error {
	tenantUUID = strings.TrimSpace(tenantUUID)
	tierID = strings.TrimSpace(tierID)
	if tenantUUID == "" || tierID == "" {
		return errors.New("tenant_uuid/tier_id required")
	}
	var tier membershipModel.MembershipTier
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, tierID).
		First(&tier).Error; err != nil {
		return err
	}
	if err := s.deps.DB.WithContext(ctx).Delete(&tier).Error; err != nil {
		return err
	}
	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_membership_tier_delete",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(actorID),
		TargetType: "membership_tier",
		TargetID:   tier.ID,
		Result:     "deleted",
		Metadata: map[string]any{
			"name": tier.Name,
			"code": tier.Code,
		},
		EmittedAt: time.Now().UTC(),
	})
	return nil
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
			"customer_id":  ent.CustomerID,
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
			"customer_id":  ent.CustomerID,
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

func (s *Service) RedeemPointsBenefit(ctx context.Context, tenantUUID string, req RedeemPointsBenefitRequest) (*RedeemPointsBenefitResult, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID := strings.TrimSpace(req.CustomerID)
	benefitID := strings.TrimSpace(req.BenefitID)
	sourceID := strings.TrimSpace(req.SourceID)
	if tenantUUID == "" || customerID == "" || benefitID == "" {
		return nil, errors.New("tenant_uuid/customer_id/benefit_id required")
	}
	if req.PointsCost <= 0 {
		return nil, errors.New("points_cost must be greater than 0")
	}
	if sourceID == "" {
		sourceID = utils.NewUUID()
	}
	if existed, err := s.getRedeemResultBySourceID(ctx, tenantUUID, customerID, benefitID, sourceID); err == nil && existed != nil {
		return existed, nil
	}

	var result RedeemPointsBenefitResult
	now := time.Now().UTC()
	err := s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var benefit membershipModel.MembershipBenefit
		if err := tx.Where("tenant_uuid = ? AND id = ?", tenantUUID, benefitID).First(&benefit).Error; err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(benefit.Status)) != "active" {
			return errors.New("benefit is inactive")
		}

		var account membershipModel.TokenAccount
		if err := tx.Where("tenant_uuid = ? AND customer_id = ? AND token_code = ?", tenantUUID, customerID, "points").First(&account).Error; err != nil {
			return err
		}
		if account.Balance < req.PointsCost {
			return errors.New("insufficient points balance")
		}
		nextBalance := account.Balance - req.PointsCost
		if err := tx.Model(&membershipModel.TokenAccount{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, account.ID).
			Updates(map[string]any{
				"balance":    nextBalance,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		tokenTx := membershipModel.TokenTransaction{
			ID:         utils.NewUUID(),
			TenantUUID: tenantUUID,
			CustomerID: customerID,
			TokenCode:  "points",
			Delta:      -req.PointsCost,
			SourceType: "points_mall_redeem",
			SourceID:   sourceID,
		}
		if err := tx.Create(&tokenTx).Error; err != nil {
			return err
		}

		items := parseBenefitItemsForRedeem(benefit)
		grantedServices := make([]string, 0, len(items))
		for _, item := range items {
			serviceCode := strings.TrimSpace(item.ServiceCode)
			if serviceCode == "" {
				continue
			}
			quantity := item.Quantity
			if quantity == 0 {
				quantity = 1
			}
			stackPolicy := strings.TrimSpace(item.StackPolicy)
			if stackPolicy == "" {
				stackPolicy = "stack"
			}
			source := strings.TrimSpace(sourceID + ":" + benefitID + ":" + utils.NewUUID() + ":" + strings.TrimSpace(item.ServiceCode))
			ent := membershipModel.Entitlement{
				ID:          utils.NewUUID(),
				TenantUUID:  tenantUUID,
				CustomerID:  customerID,
				ServiceCode: serviceCode,
				Quantity:    quantity,
				StackPolicy: stackPolicy,
				SourceType:  "points_mall_redeem",
				SourceID:    source,
				ValidFrom:   &now,
			}
			if item.ValidDays > 0 {
				validTo := now.AddDate(0, 0, item.ValidDays)
				ent.ValidTo = &validTo
			}
			if err := tx.Create(&ent).Error; err != nil {
				return err
			}
			grantedServices = append(grantedServices, serviceCode)
		}

		result = RedeemPointsBenefitResult{
			BenefitID:       benefitID,
			CustomerID:      customerID,
			TransactionID:   tokenTx.ID,
			Balance:         nextBalance,
			GrantedServices: grantedServices,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.emitAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
		Action:     "manual_points_mall_redeem",
		TenantID:   tenantUUID,
		ActorID:    strings.TrimSpace(req.ActorID),
		TargetType: "membership_benefit",
		TargetID:   benefitID,
		Result:     "redeemed",
		Reason:     strings.TrimSpace(req.Reason),
		Metadata: map[string]any{
			"customer_id":      customerID,
			"benefit_id":       benefitID,
			"points_cost":      req.PointsCost,
			"transaction_id":   result.TransactionID,
			"balance":          result.Balance,
			"granted_services": result.GrantedServices,
			"redeem_source_id": sourceID,
		},
		EmittedAt: now,
	})
	return &result, nil
}

func (s *Service) getRedeemResultBySourceID(ctx context.Context, tenantUUID, customerID, benefitID, sourceID string) (*RedeemPointsBenefitResult, error) {
	var tx membershipModel.TokenTransaction
	err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND token_code = ? AND source_type = ? AND source_id = ?",
			tenantUUID, customerID, "points", "points_mall_redeem", sourceID).
		First(&tx).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	var account membershipModel.TokenAccount
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND token_code = ?", tenantUUID, customerID, "points").
		First(&account).Error; err != nil {
		return nil, err
	}
	var ents []membershipModel.Entitlement
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND source_type = ? AND source_id LIKE ?",
			tenantUUID, customerID, "points_mall_redeem", sourceID+":"+"%").
		Order("created_at ASC").
		Find(&ents).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	services := make([]string, 0, len(ents))
	for _, ent := range ents {
		services = append(services, strings.TrimSpace(ent.ServiceCode))
	}
	return &RedeemPointsBenefitResult{
		BenefitID:       benefitID,
		CustomerID:      customerID,
		TransactionID:   tx.ID,
		Balance:         account.Balance,
		GrantedServices: services,
	}, nil
}

type redeemBenefitItem struct {
	ServiceCode string `json:"service_code"`
	Quantity    int64  `json:"quantity"`
	ValidDays   int    `json:"valid_days"`
	StackPolicy string `json:"stack_policy"`
}

func parseBenefitItemsForRedeem(benefit membershipModel.MembershipBenefit) []redeemBenefitItem {
	fallbackCode := strings.TrimSpace(benefit.Name)
	if fallbackCode == "" {
		fallbackCode = strings.TrimSpace(benefit.ID)
	}
	if len(benefit.Items) == 0 {
		if fallbackCode == "" {
			return nil
		}
		return []redeemBenefitItem{{ServiceCode: fallbackCode, Quantity: 1}}
	}

	var arr []redeemBenefitItem
	if err := json.Unmarshal(benefit.Items, &arr); err == nil {
		out := make([]redeemBenefitItem, 0, len(arr))
		for _, item := range arr {
			if strings.TrimSpace(item.ServiceCode) == "" {
				item.ServiceCode = fallbackCode
			}
			if item.Quantity == 0 {
				item.Quantity = 1
			}
			if strings.TrimSpace(item.StackPolicy) == "" {
				item.StackPolicy = "stack"
			}
			if strings.TrimSpace(item.ServiceCode) != "" {
				out = append(out, item)
			}
		}
		return out
	}

	var obj map[string]any
	if err := json.Unmarshal(benefit.Items, &obj); err != nil {
		if fallbackCode == "" {
			return nil
		}
		return []redeemBenefitItem{{ServiceCode: fallbackCode, Quantity: 1, StackPolicy: "stack"}}
	}
	item := redeemBenefitItem{
		ServiceCode: strings.TrimSpace(toString(obj["service_code"])),
		Quantity:    toInt64(obj["quantity"]),
		ValidDays:   int(toInt64(obj["valid_days"])),
		StackPolicy: strings.TrimSpace(toString(obj["stack_policy"])),
	}
	if item.ServiceCode == "" {
		item.ServiceCode = fallbackCode
	}
	if item.Quantity == 0 {
		item.Quantity = 1
	}
	if item.StackPolicy == "" {
		item.StackPolicy = "stack"
	}
	if item.ServiceCode == "" {
		return nil
	}
	return []redeemBenefitItem{item}
}

func toString(value any) string {
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case int32:
		return int64(v)
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return 0
		}
		out, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return out
	default:
		return 0
	}
}

func (s *Service) ListTokenTransactions(ctx context.Context, tenantUUID string, req ListTokenTransactionsRequest) ([]membershipModel.TokenTransaction, int64, error) {
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID := strings.TrimSpace(req.CustomerID)
	tokenCode := strings.TrimSpace(req.TokenCode)
	sourceType := strings.TrimSpace(req.SourceType)
	sourceID := strings.TrimSpace(req.SourceID)
	if tenantUUID == "" {
		return nil, 0, errors.New("tenant_uuid required")
	}
	if customerID == "" && tokenCode == "" {
		return nil, 0, errors.New("customer_id or token_code required")
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	query := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID)
	if customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if tokenCode != "" {
		query = query.Where("token_code = ?", tokenCode)
	}
	if sourceType != "" {
		parts := strings.Split(sourceType, ",")
		values := make([]string, 0, len(parts))
		for _, part := range parts {
			p := strings.TrimSpace(part)
			if p != "" {
				values = append(values, p)
			}
		}
		if len(values) == 1 {
			query = query.Where("source_type = ?", values[0])
		} else if len(values) > 1 {
			query = query.Where("source_type IN ?", values)
		}
	}
	if sourceID != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(sourceID)
		query = query.Where("source_id LIKE ? ESCAPE '\\'", "%"+escaped+"%")
	}
	if req.CreatedFrom != nil {
		query = query.Where("created_at >= ?", req.CreatedFrom.UTC())
	}
	if req.CreatedTo != nil {
		query = query.Where("created_at <= ?", req.CreatedTo.UTC())
	}

	var total int64
	if err := query.Model(&membershipModel.TokenTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []membershipModel.TokenTransaction
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	return items, total, nil
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

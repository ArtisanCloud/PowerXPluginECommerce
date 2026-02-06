package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	paymentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	membershipModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/membership"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	paymentslogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	metadataTokenPlanIDsKey = "tokenPlanIds"
)

type planBinding struct {
	MembershipTierID string
	BenefitIDs       []string
	TokenCode        string
	TokenAmount      int64
	TokenExpireDays  int
	TokenRollover    bool
}

type benefitItem struct {
	ServiceCode string `json:"service_code"`
	Quantity    int64  `json:"quantity"`
	ValidDays   int    `json:"valid_days"`
	StackPolicy string `json:"stack_policy"`
}

type entitlementGrant struct {
	EntitlementID string
	ServiceCode   string
	Quantity      int64
	ValidFrom     time.Time
	ValidTo       *time.Time
	StackPolicy   string
	BenefitID     string
	SourceID      string
}

type tokenGrant struct {
	TransactionID string
	AccountID     string
	TokenCode     string
	Amount        int64
	SourceID      string
	PlanID        string
}

func (s *TransactionService) applySubscriptionEntitlements(ctx context.Context, db *gorm.DB, tenantUUID string, txn *paymentModel.PaymentTransaction) error {
	if s == nil || db == nil || txn == nil {
		return ErrInvalidArgument
	}
	if strings.TrimSpace(txn.OrderID) == "" {
		return nil
	}
	var order ordermodel.Order
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, txn.OrderID).
		First(&order).Error; err != nil {
		return err
	}
	var items []ordermodel.OrderItem
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, order.ID).
		Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	idemKey := resolveTransactionIdemKey(txn)
	if strings.TrimSpace(idemKey) == "" {
		idemKey = strings.TrimSpace(txn.OrderID)
	}
	paidAt := time.Now().UTC()
	if txn.CompletedAt != nil {
		paidAt = txn.CompletedAt.UTC()
	}

	for _, item := range items {
		plan, err := findSubscriptionPlanBySKU(ctx, db, tenantUUID, item.SKUID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return err
		}
		binding := parsePlanBinding(plan)
		if len(binding.BenefitIDs) == 0 {
			if ids, err := loadPlanBenefitIDs(ctx, db, tenantUUID, plan.ID); err == nil {
				binding.BenefitIDs = ids
			} else {
				return err
			}
		}
		if binding.MembershipTierID == "" && len(binding.BenefitIDs) == 0 && strings.TrimSpace(binding.TokenCode) == "" {
			logger.WithFields(logger.Fields{
				"tenant_uuid": tenantUUID,
				"order_id":    order.ID,
				"plan_id":     plan.ID,
			}).Warn("subscription plan missing membership/benefit/token bindings")
			continue
		}
		validTo := computePlanValidTo(paidAt, plan)
		baseMetadata := map[string]any{
			"order_id":        strings.TrimSpace(order.ID),
			"order_no":        strings.TrimSpace(order.OrderNo),
			"transaction_id":  txn.ID,
			"subscription_id": strings.TrimSpace(plan.ID),
			"plan_code":       strings.TrimSpace(plan.PlanCode),
			"sku_id":          strings.TrimSpace(item.SKUID),
			"source_id":       strings.TrimSpace(idemKey),
		}
		if binding.MembershipTierID != "" {
			assignmentID, result, err := upsertMembershipAssignment(ctx, db, tenantUUID, order.CustomerID, binding.MembershipTierID, paidAt, validTo, idemKey)
			if err != nil {
				return err
			}
			if assignmentID != "" && result != "" {
				metadata := cloneMetadata(baseMetadata)
				metadata["tier_id"] = binding.MembershipTierID
				s.emitSubscriptionAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
					Action:     "membership_assignment_grant",
					TenantID:   tenantUUID,
					ActorID:    strings.TrimSpace(order.CustomerID),
					TargetType: "membership_assignment",
					TargetID:   assignmentID,
					Result:     result,
					Reason:     "subscription_paid",
					Metadata:   metadata,
					EmittedAt:  time.Now().UTC(),
				})
			}
		}
		if len(binding.BenefitIDs) > 0 {
			grants, err := grantBenefits(ctx, db, tenantUUID, order.CustomerID, binding.BenefitIDs, paidAt, idemKey)
			if err != nil {
				return err
			}
			for _, grant := range grants {
				metadata := cloneMetadata(baseMetadata)
				metadata["benefit_id"] = grant.BenefitID
				metadata["service_code"] = grant.ServiceCode
				metadata["stack_policy"] = grant.StackPolicy
				metadata["quantity"] = grant.Quantity
				if grant.ValidTo != nil {
					metadata["valid_to"] = grant.ValidTo.UTC().Format(time.RFC3339)
				}
				s.emitSubscriptionAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
					Action:     "entitlement_grant",
					TenantID:   tenantUUID,
					ActorID:    strings.TrimSpace(order.CustomerID),
					TargetType: "entitlement",
					TargetID:   grant.EntitlementID,
					Result:     "created",
					Reason:     "subscription_paid",
					Metadata:   metadata,
					EmittedAt:  time.Now().UTC(),
				})
			}
		}
		if strings.TrimSpace(binding.TokenCode) != "" && binding.TokenAmount > 0 {
			grant, err := grantTokens(ctx, db, tenantUUID, order.CustomerID, binding.TokenCode, binding.TokenAmount, idemKey, plan.ID)
			if err != nil {
				return err
			}
			if grant != nil {
				metadata := cloneMetadata(baseMetadata)
				metadata["token_code"] = grant.TokenCode
				metadata["token_amount"] = grant.Amount
				metadata["token_account_id"] = grant.AccountID
				s.emitSubscriptionAudit(ctx, tenantUUID, paymentslogger.AuditEvent{
					Action:     "token_grant",
					TenantID:   tenantUUID,
					ActorID:    strings.TrimSpace(order.CustomerID),
					TargetType: "token_transaction",
					TargetID:   grant.TransactionID,
					Result:     "created",
					Reason:     "subscription_paid",
					Metadata:   metadata,
					EmittedAt:  time.Now().UTC(),
				})
			}
		}
	}
	return nil
}

func resolveTransactionIdemKey(txn *paymentModel.PaymentTransaction) string {
	if txn == nil {
		return ""
	}
	if len(txn.Metadata) > 0 {
		payload := map[string]any{}
		_ = json.Unmarshal(txn.Metadata, &payload)
		if v := pickString(payload, "wechat_transaction_id", "transaction_id"); v != "" {
			return v
		}
	}
	if strings.TrimSpace(txn.TransactionNo) != "" {
		return strings.TrimSpace(txn.TransactionNo)
	}
	return strings.TrimSpace(txn.OrderID)
}

func findSubscriptionPlanBySKU(ctx context.Context, db *gorm.DB, tenantUUID, skuID string) (*productmodel.SubscriptionPlan, error) {
	if strings.TrimSpace(skuID) == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var plan productmodel.SubscriptionPlan
	err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND status = ? AND (metadata->>'skuId' = ? OR metadata->>'sku_id' = ? OR metadata->>'skuID' = ?)", tenantUUID, "active", skuID, skuID, skuID).
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func parsePlanBinding(plan *productmodel.SubscriptionPlan) planBinding {
	binding := planBinding{}
	if plan == nil || len(plan.Metadata) == 0 {
		return binding
	}
	payload := map[string]any{}
	if err := json.Unmarshal(plan.Metadata, &payload); err != nil {
		return binding
	}
	binding.MembershipTierID = pickString(payload, "membershipTierId", "membership_tier_id", "membershipTierID")
	binding.TokenCode = pickString(payload, "tokenCode", "token_code")
	binding.TokenAmount = pickInt64(payload, "tokenAmount", "token_amount")
	binding.TokenExpireDays = int(pickInt64(payload, "tokenExpireDays", "token_expire_days"))
	binding.TokenRollover = pickBool(payload, "tokenRollover", "token_rollover")
	binding.BenefitIDs = pickStringArray(payload, "benefitIds", "benefit_ids")
	return binding
}

func loadPlanBenefitIDs(ctx context.Context, db *gorm.DB, tenantUUID, planID string) ([]string, error) {
	if db == nil || strings.TrimSpace(planID) == "" {
		return []string{}, nil
	}
	var rows []productmodel.SubscriptionPlanBenefit
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND plan_id = ?", tenantUUID, planID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		id := strings.TrimSpace(row.BenefitID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func computePlanValidTo(start time.Time, plan *productmodel.SubscriptionPlan) *time.Time {
	if plan == nil {
		return nil
	}
	cycle := strings.ToLower(strings.TrimSpace(plan.BillingCycle))
	var end time.Time
	switch cycle {
	case "monthly":
		end = start.AddDate(0, 1, 0)
	case "quarterly":
		end = start.AddDate(0, 3, 0)
	case "yearly":
		end = start.AddDate(1, 0, 0)
	case "custom":
		if plan.BillingValue > 0 {
			end = start.AddDate(0, 0, plan.BillingValue)
		}
	default:
		return nil
	}
	if plan.TrialDays > 0 {
		end = end.AddDate(0, 0, plan.TrialDays)
	}
	return &end
}

func upsertMembershipAssignment(ctx context.Context, db *gorm.DB, tenantUUID, customerID, tierID string, validFrom time.Time, validTo *time.Time, sourceID string) (string, string, error) {
	if strings.TrimSpace(tierID) == "" || strings.TrimSpace(customerID) == "" {
		return "", "", nil
	}
	var existing membershipModel.MembershipAssignment
	err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND tier_id = ?", tenantUUID, customerID, tierID).
		First(&existing).Error
	switch {
	case err == nil:
		updates := map[string]any{
			"status":     "active",
			"valid_from": validFrom,
			"updated_at": time.Now().UTC(),
		}
		if validTo != nil {
			updates["valid_to"] = validTo
		}
		if err := db.WithContext(ctx).
			Model(&membershipModel.MembershipAssignment{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, existing.ID).
			Updates(updates).Error; err != nil {
			return "", "", err
		}
		return existing.ID, "updated", nil
	case err == gorm.ErrRecordNotFound:
		assignment := membershipModel.MembershipAssignment{
			ID:         utils.NewUUID(),
			TenantUUID: tenantUUID,
			CustomerID: customerID,
			TierID:     tierID,
			Status:     "active",
			ValidFrom:  &validFrom,
			ValidTo:    validTo,
			SourceType: "subscription",
			SourceID:   strings.TrimSpace(sourceID),
		}
		if err := db.WithContext(ctx).Create(&assignment).Error; err != nil {
			return "", "", err
		}
		return assignment.ID, "created", nil
	default:
		return "", "", err
	}
}

func grantBenefits(ctx context.Context, db *gorm.DB, tenantUUID, customerID string, benefitIDs []string, validFrom time.Time, sourceID string) ([]entitlementGrant, error) {
	var grants []entitlementGrant
	if len(benefitIDs) == 0 {
		return grants, nil
	}
	var benefits []membershipModel.MembershipBenefit
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND id IN ?", tenantUUID, benefitIDs).
		Find(&benefits).Error; err != nil {
		return grants, err
	}
	for _, benefit := range benefits {
		items := parseBenefitItems(benefit)
		for _, item := range items {
			if strings.TrimSpace(item.ServiceCode) == "" {
				continue
			}
			entSourceID := strings.TrimSpace(fmt.Sprintf("%s:%s", sourceID, benefit.ID))
			grant, err := grantEntitlement(ctx, db, tenantUUID, customerID, item, validFrom, entSourceID, benefit.ID)
			if err != nil {
				return grants, err
			}
			if grant != nil {
				grants = append(grants, *grant)
			}
		}
	}
	return grants, nil
}

func parseBenefitItems(benefit membershipModel.MembershipBenefit) []benefitItem {
	fallbackCode := strings.TrimSpace(benefit.Name)
	if fallbackCode == "" {
		fallbackCode = strings.TrimSpace(benefit.ID)
	}
	if len(benefit.Items) == 0 {
		if fallbackCode == "" {
			return nil
		}
		return []benefitItem{{ServiceCode: fallbackCode, Quantity: 1}}
	}
	var items []benefitItem
	if err := json.Unmarshal(benefit.Items, &items); err == nil {
		normalized := make([]benefitItem, 0, len(items))
		for _, it := range items {
			if strings.TrimSpace(it.ServiceCode) == "" {
				it.ServiceCode = fallbackCode
			}
			if it.Quantity == 0 {
				it.Quantity = 1
			}
			if strings.TrimSpace(it.ServiceCode) != "" {
				normalized = append(normalized, it)
			}
		}
		return normalized
	}
	// fallback for single item
	payload := map[string]any{}
	if err := json.Unmarshal(benefit.Items, &payload); err != nil {
		if fallbackCode == "" {
			return nil
		}
		return []benefitItem{{ServiceCode: fallbackCode, Quantity: 1}}
	}
	item := benefitItem{
		ServiceCode: pickString(payload, "service_code", "serviceCode"),
		Quantity:    pickInt64(payload, "quantity"),
		ValidDays:   int(pickInt64(payload, "valid_days", "validDays")),
		StackPolicy: pickString(payload, "stack_policy", "stackPolicy"),
	}
	if strings.TrimSpace(item.ServiceCode) == "" {
		item.ServiceCode = fallbackCode
	}
	if item.Quantity == 0 {
		item.Quantity = 1
	}
	if strings.TrimSpace(item.ServiceCode) == "" {
		return nil
	}
	return []benefitItem{item}
}

func grantEntitlement(ctx context.Context, db *gorm.DB, tenantUUID, customerID string, item benefitItem, validFrom time.Time, sourceID, benefitID string) (*entitlementGrant, error) {
	stackPolicy := strings.TrimSpace(item.StackPolicy)
	if stackPolicy == "" {
		stackPolicy = "stack"
	}
	qty := item.Quantity
	if qty == 0 {
		return nil, nil
	}
	var validTo *time.Time
	if item.ValidDays > 0 {
		end := validFrom.AddDate(0, 0, item.ValidDays)
		validTo = &end
	}
	var existing membershipModel.Entitlement
	err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND source_type = ? AND source_id = ? AND service_code = ?",
			tenantUUID, customerID, "subscription", sourceID, item.ServiceCode).
		First(&existing).Error
	if err == nil {
		return nil, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	ent := membershipModel.Entitlement{
		ID:          utils.NewUUID(),
		TenantUUID:  tenantUUID,
		CustomerID:  customerID,
		ServiceCode: strings.TrimSpace(item.ServiceCode),
		Quantity:    qty,
		ValidFrom:   &validFrom,
		ValidTo:     validTo,
		StackPolicy: stackPolicy,
		SourceType:  "subscription",
		SourceID:    strings.TrimSpace(sourceID),
	}
	if err := db.WithContext(ctx).Create(&ent).Error; err != nil {
		return nil, err
	}
	return &entitlementGrant{
		EntitlementID: ent.ID,
		ServiceCode:   ent.ServiceCode,
		Quantity:      ent.Quantity,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
		StackPolicy:   ent.StackPolicy,
		BenefitID:     strings.TrimSpace(benefitID),
		SourceID:      ent.SourceID,
	}, nil
}

func grantTokens(ctx context.Context, db *gorm.DB, tenantUUID, customerID, tokenCode string, amount int64, sourceID, planID string) (*tokenGrant, error) {
	if strings.TrimSpace(tokenCode) == "" || amount <= 0 {
		return nil, nil
	}
	var existing membershipModel.TokenTransaction
	err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND source_type = ? AND source_id = ? AND token_code = ?",
			tenantUUID, "subscription", sourceID, tokenCode).
		First(&existing).Error
	if err == nil {
		return nil, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	account := membershipModel.TokenAccount{}
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND token_code = ?", tenantUUID, customerID, tokenCode).
		First(&account).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		account = membershipModel.TokenAccount{
			ID:         utils.NewUUID(),
			TenantUUID: tenantUUID,
			CustomerID: customerID,
			TokenCode:  strings.TrimSpace(tokenCode),
			Balance:    amount,
		}
		if err := db.WithContext(ctx).Create(&account).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.WithContext(ctx).
			Model(&membershipModel.TokenAccount{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, account.ID).
			Updates(map[string]any{
				"balance":    gorm.Expr("balance + ?", amount),
				"updated_at": time.Now().UTC(),
			}).Error; err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(planID) != "" {
		if err := appendTokenPlanLink(ctx, db, tenantUUID, account.ID, planID); err != nil {
			return nil, err
		}
	}
	transactionID := utils.NewUUID()
	transaction := membershipModel.TokenTransaction{
		ID:         transactionID,
		TenantUUID: tenantUUID,
		CustomerID: customerID,
		TokenCode:  strings.TrimSpace(tokenCode),
		Delta:      amount,
		SourceType: "subscription",
		SourceID:   strings.TrimSpace(sourceID),
	}
	if err := db.WithContext(ctx).Create(&transaction).Error; err != nil {
		return nil, err
	}
	return &tokenGrant{
		TransactionID: transactionID,
		AccountID:     account.ID,
		TokenCode:     transaction.TokenCode,
		Amount:        amount,
		SourceID:      strings.TrimSpace(sourceID),
		PlanID:        strings.TrimSpace(planID),
	}, nil
}

func appendTokenPlanLink(ctx context.Context, db *gorm.DB, tenantUUID, accountID, planID string) error {
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(planID) == "" {
		return nil
	}
	var account membershipModel.TokenAccount
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, accountID).
		First(&account).Error; err != nil {
		return err
	}
	meta := map[string]any{}
	if len(account.Metadata) > 0 {
		_ = json.Unmarshal(account.Metadata, &meta)
	}
	var ids []string
	if raw, ok := meta[metadataTokenPlanIDsKey]; ok {
		switch t := raw.(type) {
		case []any:
			for _, it := range t {
				ids = append(ids, strings.TrimSpace(fmt.Sprint(it)))
			}
		case []string:
			ids = append(ids, t...)
		case string:
			if strings.TrimSpace(t) != "" {
				ids = append(ids, strings.Split(t, ",")...)
			}
		}
	}
	planID = strings.TrimSpace(planID)
	for _, id := range ids {
		if strings.TrimSpace(id) == planID {
			return nil
		}
	}
	ids = append(ids, planID)
	meta[metadataTokenPlanIDsKey] = ids
	buf, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).
		Model(&membershipModel.TokenAccount{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, account.ID).
		Updates(map[string]any{
			"metadata":   datatypes.JSON(buf),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (s *TransactionService) emitSubscriptionAudit(ctx context.Context, tenantUUID string, evt paymentslogger.AuditEvent) {
	if s == nil {
		return
	}
	logger := s.logger(ctx)
	if logger == nil {
		return
	}
	if strings.TrimSpace(evt.TenantID) == "" {
		evt.TenantID = strings.TrimSpace(tenantUUID)
	}
	logger.EmitAudit(evt)
}

func cloneMetadata(source map[string]any) map[string]any {
	if len(source) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func pickInt64(payload map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case float64:
				return int64(t)
			case int:
				return int64(t)
			case int64:
				return t
			case string:
				if t == "" {
					continue
				}
				if parsed, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
					return parsed
				}
			}
		}
	}
	return 0
}

func pickBool(payload map[string]any, keys ...string) bool {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case bool:
				return t
			case string:
				val := strings.ToLower(strings.TrimSpace(t))
				return val == "true" || val == "1" || val == "yes"
			}
		}
	}
	return false
}

func pickStringArray(payload map[string]any, keys ...string) []string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case []string:
				return t
			case []any:
				out := make([]string, 0, len(t))
				for _, item := range t {
					out = append(out, strings.TrimSpace(fmt.Sprint(item)))
				}
				return out
			case string:
				if strings.TrimSpace(t) == "" {
					continue
				}
				parts := strings.Split(t, ",")
				out := make([]string, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						out = append(out, p)
					}
				}
				return out
			}
		}
	}
	return nil
}

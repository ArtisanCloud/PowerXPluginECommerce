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
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
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
		if binding.MembershipTierID == "" && len(binding.BenefitIDs) == 0 && strings.TrimSpace(binding.TokenCode) == "" {
			logger.WithFields(logger.Fields{
				"tenant_uuid": tenantUUID,
				"order_id":    order.ID,
				"plan_id":     plan.ID,
			}).Warn("subscription plan missing membership/benefit/token bindings")
			continue
		}
		validTo := computePlanValidTo(paidAt, plan)
		if binding.MembershipTierID != "" {
			if err := upsertMembershipAssignment(ctx, db, tenantUUID, order.CustomerID, binding.MembershipTierID, paidAt, validTo, idemKey); err != nil {
				return err
			}
		}
		if len(binding.BenefitIDs) > 0 {
			if err := grantBenefits(ctx, db, tenantUUID, order.CustomerID, binding.BenefitIDs, paidAt, idemKey); err != nil {
				return err
			}
		}
		if strings.TrimSpace(binding.TokenCode) != "" && binding.TokenAmount > 0 {
			if err := grantTokens(ctx, db, tenantUUID, order.CustomerID, binding.TokenCode, binding.TokenAmount, idemKey); err != nil {
				return err
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

func upsertMembershipAssignment(ctx context.Context, db *gorm.DB, tenantUUID, customerID, tierID string, validFrom time.Time, validTo *time.Time, sourceID string) error {
	if strings.TrimSpace(tierID) == "" || strings.TrimSpace(customerID) == "" {
		return nil
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
		return db.WithContext(ctx).
			Model(&membershipModel.MembershipAssignment{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, existing.ID).
			Updates(updates).Error
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
		return db.WithContext(ctx).Create(&assignment).Error
	default:
		return err
	}
}

func grantBenefits(ctx context.Context, db *gorm.DB, tenantUUID, customerID string, benefitIDs []string, validFrom time.Time, sourceID string) error {
	if len(benefitIDs) == 0 {
		return nil
	}
	var benefits []membershipModel.MembershipBenefit
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND id IN ?", tenantUUID, benefitIDs).
		Find(&benefits).Error; err != nil {
		return err
	}
	for _, benefit := range benefits {
		items := parseBenefitItems(benefit)
		for _, item := range items {
			if strings.TrimSpace(item.ServiceCode) == "" {
				continue
			}
			entSourceID := strings.TrimSpace(fmt.Sprintf("%s:%s", sourceID, benefit.ID))
			if err := grantEntitlement(ctx, db, tenantUUID, customerID, item, validFrom, entSourceID); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseBenefitItems(benefit membershipModel.MembershipBenefit) []benefitItem {
	if len(benefit.Items) == 0 {
		return nil
	}
	var items []benefitItem
	if err := json.Unmarshal(benefit.Items, &items); err == nil {
		return items
	}
	// fallback for single item
	payload := map[string]any{}
	if err := json.Unmarshal(benefit.Items, &payload); err != nil {
		return nil
	}
	item := benefitItem{
		ServiceCode: pickString(payload, "service_code", "serviceCode"),
		Quantity:    pickInt64(payload, "quantity"),
		ValidDays:   int(pickInt64(payload, "valid_days", "validDays")),
		StackPolicy: pickString(payload, "stack_policy", "stackPolicy"),
	}
	return []benefitItem{item}
}

func grantEntitlement(ctx context.Context, db *gorm.DB, tenantUUID, customerID string, item benefitItem, validFrom time.Time, sourceID string) error {
	stackPolicy := strings.TrimSpace(item.StackPolicy)
	if stackPolicy == "" {
		stackPolicy = "stack"
	}
	qty := item.Quantity
	if qty == 0 {
		return nil
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
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
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
	return db.WithContext(ctx).Create(&ent).Error
}

func grantTokens(ctx context.Context, db *gorm.DB, tenantUUID, customerID, tokenCode string, amount int64, sourceID string) error {
	if strings.TrimSpace(tokenCode) == "" || amount <= 0 {
		return nil
	}
	var existing membershipModel.TokenTransaction
	err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND source_type = ? AND source_id = ? AND token_code = ?",
			tenantUUID, "subscription", sourceID, tokenCode).
		First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	account := membershipModel.TokenAccount{}
	if err := db.WithContext(ctx).
		Where("tenant_uuid = ? AND customer_id = ? AND token_code = ?", tenantUUID, customerID, tokenCode).
		First(&account).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		account = membershipModel.TokenAccount{
			ID:         utils.NewUUID(),
			TenantUUID: tenantUUID,
			CustomerID: customerID,
			TokenCode:  strings.TrimSpace(tokenCode),
			Balance:    amount,
		}
		if err := db.WithContext(ctx).Create(&account).Error; err != nil {
			return err
		}
	} else {
		if err := db.WithContext(ctx).
			Model(&membershipModel.TokenAccount{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, account.ID).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}
	}
	transaction := membershipModel.TokenTransaction{
		ID:         utils.NewUUID(),
		TenantUUID: tenantUUID,
		CustomerID: customerID,
		TokenCode:  strings.TrimSpace(tokenCode),
		Delta:      amount,
		SourceType: "subscription",
		SourceID:   strings.TrimSpace(sourceID),
	}
	return db.WithContext(ctx).Create(&transaction).Error
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

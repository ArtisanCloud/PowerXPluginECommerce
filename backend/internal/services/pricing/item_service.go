package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	pricingRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ItemService struct {
	*Service
	audit *AuditService
}

func NewItemService(deps *app.Deps) *ItemService {
	svc := NewService(deps)
	return &ItemService{Service: svc, audit: &AuditService{Service: svc}}
}

type ItemInput struct {
	SKUID           string
	BaseAmountMinor *int64
	SaleAmountMinor *int64
	MsrpAmountMinor *int64
	CostAmountMinor *int64
	MinAmountMinor  *int64
	MaxAmountMinor  *int64
	TaxIncluded     *bool
	Meta            map[string]any
}

type UpsertItemsInput struct {
	PricebookID string
	VersionID   string
	Items       []ItemInput
	Actor       string
}

type UpsertItemsResult struct {
	Upserted int
	Skipped  []string
}

func (s *ItemService) UpsertItems(ctx context.Context, in UpsertItemsInput) (*UpsertItemsResult, error) {
	if !s.Ready() {
		return nil, E(CodeServiceUnavailable, ErrServiceUnavailable)
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, E(CodeTenantMissing, err)
	}
	pricebookID := strings.TrimSpace(in.PricebookID)
	versionID := strings.TrimSpace(in.VersionID)
	if pricebookID == "" || versionID == "" {
		return nil, E(CodeInvalidArgument, errors.New("pricebook_id/version_id are required"))
	}
	if len(in.Items) == 0 {
		return &UpsertItemsResult{Upserted: 0}, nil
	}

	tx, err := s.PricebookItemRepo.BeginTenantTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	verRepo := pricingRepo.NewPricebookVersionRepository(tx)
	var v pricingModel.PricebookVersion
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ? AND pricebook_id = ?", tenantUUID, versionID, pricebookID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, E(CodeVersionNotFound, err)
		}
		return nil, err
	}
	_ = verRepo // reserved for later query helpers

	if v.State != "draft" {
		return nil, E(CodeVersionNotEditable, errors.New("only draft version can be edited"))
	}

	now := time.Now().UTC()
	seen := map[string]struct{}{}
	var skipped []string
	var rows []*pricingModel.PricebookItem

	for _, item := range in.Items {
		skuID := strings.TrimSpace(item.SKUID)
		if skuID == "" {
			continue
		}
		if _, ok := seen[skuID]; ok {
			skipped = append(skipped, skuID)
			continue
		}
		seen[skuID] = struct{}{}

		if err := validateItem(item); err != nil {
			return nil, E(CodeInvalidArgument, err)
		}

		metaJSON := datatypes.JSON(nil)
		if item.Meta != nil {
			if buf, marshalErr := json.Marshal(item.Meta); marshalErr == nil {
				metaJSON = datatypes.JSON(buf)
			}
		}
		taxIncluded := false
		if item.TaxIncluded != nil {
			taxIncluded = *item.TaxIncluded
		}

		rows = append(rows, &pricingModel.PricebookItem{
			ID:          uuid.NewString(),
			TenantUUID:  tenantUUID,
			PricebookID: pricebookID,
			VersionID:   versionID,
			SKUID:       skuID,
			BaseAmount:  item.BaseAmountMinor,
			SaleAmount:  item.SaleAmountMinor,
			MsrpAmount:  item.MsrpAmountMinor,
			CostAmount:  item.CostAmountMinor,
			MinAmount:   item.MinAmountMinor,
			MaxAmount:   item.MaxAmountMinor,
			TaxIncluded: taxIncluded,
			Meta:        metaJSON,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	if len(rows) == 0 {
		return &UpsertItemsResult{Upserted: 0, Skipped: skipped}, nil
	}

	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_uuid"},
			{Name: "version_id"},
			{Name: "sku_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"pricebook_id",
			"base_amount_minor",
			"sale_amount_minor",
			"msrp_amount_minor",
			"cost_amount_minor",
			"min_amount_minor",
			"max_amount_minor",
			"tax_included",
			"meta",
			"updated_at",
		}),
	}).Create(&rows).Error; err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, tx, "pricebook_version", versionID, "items_upsert", in.Actor, map[string]any{
		"pricebook_id": pricebookID,
		"version_id":   versionID,
		"upserted":     len(rows),
		"skipped":      len(skipped),
	})

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &UpsertItemsResult{Upserted: len(rows), Skipped: skipped}, nil
}

func validateItem(in ItemInput) error {
	if strings.TrimSpace(in.SKUID) == "" {
		return errors.New("sku_id is required")
	}
	if err := validateNonNegative("base_amount_minor", in.BaseAmountMinor); err != nil {
		return err
	}
	if err := validateNonNegative("sale_amount_minor", in.SaleAmountMinor); err != nil {
		return err
	}
	if err := validateNonNegative("msrp_amount_minor", in.MsrpAmountMinor); err != nil {
		return err
	}
	if err := validateNonNegative("cost_amount_minor", in.CostAmountMinor); err != nil {
		return err
	}
	if err := validateNonNegative("min_amount_minor", in.MinAmountMinor); err != nil {
		return err
	}
	if err := validateNonNegative("max_amount_minor", in.MaxAmountMinor); err != nil {
		return err
	}
	if in.MinAmountMinor != nil && in.MaxAmountMinor != nil {
		if *in.MinAmountMinor > *in.MaxAmountMinor {
			return errors.New("min_amount_minor must be <= max_amount_minor")
		}
	}
	return nil
}

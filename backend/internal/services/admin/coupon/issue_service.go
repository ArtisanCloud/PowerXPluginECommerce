package coupon

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrCouponIssueServiceUnavailable = errors.New("coupon issue service unavailable")
)

type IssueInput struct {
	TenantUUID      string         `json:"tenant_uuid"`
	TemplateID      string         `json:"template_id"`
	UserIDs         []string       `json:"user_ids"`
	QuantityPerUser int            `json:"quantity_per_user"`
	ValidFrom       *time.Time     `json:"valid_from,omitempty"`
	ValidTo         *time.Time     `json:"valid_to,omitempty"`
	Meta            map[string]any `json:"meta,omitempty"`
	RequestID       string         `json:"request_id,omitempty"`
	Operator        string         `json:"operator,omitempty"`
}

type IssueResult struct {
	Issued   int      `json:"issued"`
	AssetIDs []string `json:"asset_ids"`
}

type IssueService struct {
	deps      *app.Deps
	assetRepo *couponrepo.AssetRepository
}

func NewIssueService(deps *app.Deps) *IssueService {
	if deps == nil || deps.DB == nil {
		return &IssueService{deps: deps}
	}
	return &IssueService{
		deps:      deps,
		assetRepo: couponrepo.NewAssetRepository(deps.DB),
	}
}

func (s *IssueService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.assetRepo != nil
}

func (s *IssueService) Issue(ctx context.Context, in IssueInput) (*IssueResult, error) {
	if !s.Ready() {
		return nil, ErrCouponIssueServiceUnavailable
	}
	in.TenantUUID = strings.TrimSpace(in.TenantUUID)
	in.TemplateID = strings.TrimSpace(in.TemplateID)
	if in.TenantUUID == "" || in.TemplateID == "" {
		return nil, errors.New("tenant uuid and template id are required")
	}
	if in.QuantityPerUser <= 0 {
		in.QuantityPerUser = 1
	}
	userIDs := normalizeUserIDs(in.UserIDs)
	if len(userIDs) == 0 {
		return nil, errors.New("user ids are required")
	}
	metaJSON := datatypes.JSON("{}")
	if len(in.Meta) > 0 {
		raw, err := datatypes.JSONMap(in.Meta).MarshalJSON()
		if err != nil {
			return nil, err
		}
		metaJSON = datatypes.JSON(raw)
	}
	now := time.Now().UTC()
	var template couponmodel.CouponTemplate
	if err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", in.TenantUUID, in.TemplateID).
		First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponTemplateNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(strings.ToLower(template.Status)) != "active" {
		return nil, errors.New("template is not active")
	}
	validFrom := template.ValidFrom
	validTo := template.ValidTo
	if in.ValidFrom != nil {
		validFrom = in.ValidFrom.UTC()
	}
	if in.ValidTo != nil {
		validTo = in.ValidTo.UTC()
	}
	if validFrom.After(validTo) {
		return nil, errors.New("valid_from must be before or equal to valid_to")
	}
	res := &IssueResult{AssetIDs: make([]string, 0, len(userIDs)*in.QuantityPerUser)}
	err := s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userID := range userIDs {
			for i := 0; i < in.QuantityPerUser; i++ {
				assetID := uuid.NewString()
				code := buildCouponCode(template.Code)
				row := &couponmodel.CouponAsset{
					ID:         assetID,
					TenantUUID: in.TenantUUID,
					TemplateID: in.TemplateID,
					UserID:     userID,
					CouponCode: code,
					Status:     AssetStatusAvailable,
					ValidFrom:  &validFrom,
					ValidTo:    &validTo,
					Meta:       metaJSON,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				if err := tx.WithContext(ctx).Create(row).Error; err != nil {
					return err
				}
				res.AssetIDs = append(res.AssetIDs, assetID)
				res.Issued++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func normalizeUserIDs(inputs []string) []string {
	seen := make(map[string]struct{}, len(inputs))
	out := make([]string, 0, len(inputs))
	for _, raw := range inputs {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func buildCouponCode(templateCode string) string {
	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))
	if templateCode == "" {
		templateCode = "CPN"
	}
	return fmt.Sprintf("%s-%s-%d", templateCode, strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)
}

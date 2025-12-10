package customer

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	customerrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Service 提供客户域的占位实现，依赖宿主 DB。
type Service struct {
	deps *app.Deps
	repo *customerrepo.Repository
}

// NewService 创建客户服务。
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("customer service requires initialized DB dependency")
	}
	return &Service{
		deps: deps,
		repo: customerrepo.NewRepository(deps.DB),
	}
}

// ListFilters 描述查询条件。
type ListFilters struct {
	Keyword   string
	Tier      string
	Source    string
	Region    string
	RiskLevel string
	Type      string
	Tags      []string
	Page      int
	PageSize  int
	Sort      string
}

// CustomerListResult 列表响应。
type CustomerListResult struct {
	Data []Customer       `json:"data"`
	Meta CustomerListMeta `json:"meta"`
}

// CustomerListMeta 附带分页信息。
type CustomerListMeta struct {
	Total       int     `json:"total"`
	SavedViewID *string `json:"savedViewId,omitempty"`
}

// Customer 兼容前端所需字段。
type Customer struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Type                string              `json:"type"`
	Email               string              `json:"email,omitempty"`
	Phone               string              `json:"phone,omitempty"`
	Country             string              `json:"country,omitempty"`
	Region              string              `json:"region,omitempty"`
	MembershipTier      string              `json:"membershipTier,omitempty"`
	MembershipTierLabel string              `json:"membershipTierLabel,omitempty"`
	GrowthValue         int                 `json:"growthValue,omitempty"`
	Points              int                 `json:"points,omitempty"`
	LastOrderAmount     float64             `json:"lastOrderAmount,omitempty"`
	LastOrderAt         string              `json:"lastOrderAt,omitempty"`
	Status              string              `json:"status,omitempty"`
	RiskLevel           string              `json:"riskLevel,omitempty"`
	Source              string              `json:"source,omitempty"`
	AccountManager      string              `json:"accountManager,omitempty"`
	Tags                []string            `json:"tags,omitempty"`
	CreatedAt           string              `json:"createdAt,omitempty"`
	UpdatedAt           string              `json:"updatedAt,omitempty"`
	Notes               string              `json:"notes,omitempty"`
	MaskedFields        []string            `json:"maskedFields,omitempty"`
	MembershipSnapshot  *MembershipSnapshot `json:"membershipSnapshot,omitempty"`
	Metadata            map[string]any      `json:"metadata,omitempty"`
}

// MembershipSnapshot 快照。
type MembershipSnapshot struct {
	CustomerID        string              `json:"customerId"`
	Tier              string              `json:"tier"`
	GrowthValue       int                 `json:"growthValue"`
	Points            int                 `json:"points"`
	RetentionStatus   string              `json:"retentionStatus"`
	Benefits          []MembershipBenefit `json:"benefits,omitempty"`
	LastBenefitUsedAt *string             `json:"lastBenefitUsedAt,omitempty"`
}

// MembershipBenefit 权益状态。
type MembershipBenefit struct {
	Name string `json:"name"`
	Used bool   `json:"used"`
}

// ListCustomers 返回过滤后的客户列表。
func (s *Service) ListCustomers(ctx context.Context, filters ListFilters) (*CustomerListResult, error) {
	page, pageSize := normalizePagination(filters.Page, filters.PageSize)
	sanitized := filters
	sanitized.Tags = normalizeTags(filters.Tags)
	sanitized.Page = page
	sanitized.PageSize = pageSize

	repoFilters := customerrepo.ListQueryOptions{
		Keyword:   sanitized.Keyword,
		Tier:      sanitized.Tier,
		Source:    sanitized.Source,
		Region:    sanitized.Region,
		RiskLevel: sanitized.RiskLevel,
		Type:      sanitized.Type,
		Tags:      sanitized.Tags,
		Sort:      sanitized.Sort,
		Page:      sanitized.Page,
		PageSize:  sanitized.PageSize,
	}

	pageResult, err := s.repo.List(ctx, repoFilters)
	if err != nil {
		return nil, err
	}
	customers := make([]Customer, 0, len(pageResult.List))
	for _, entity := range pageResult.List {
		if entity == nil {
			continue
		}
		customers = append(customers, convertEntityToCustomer(entity))
	}
	return &CustomerListResult{
		Data: customers,
		Meta: CustomerListMeta{Total: int(pageResult.Total)},
	}, nil
}

func convertEntityToCustomer(entity *customermodel.Customer) Customer {
	if entity == nil {
		return Customer{}
	}
	result := Customer{
		ID:                  entity.CustomerID,
		Name:                entity.Name,
		Type:                entity.Type,
		Email:               entity.Email,
		Phone:               entity.Phone,
		Country:             entity.Country,
		Region:              entity.Region,
		MembershipTier:      entity.MembershipTier,
		MembershipTierLabel: entity.MembershipTierLabel,
		GrowthValue:         entity.GrowthValue,
		Points:              entity.Points,
		LastOrderAmount:     entity.LastOrderAmount,
		LastOrderAt:         formatTimePtr(entity.LastOrderAt),
		Status:              entity.Status,
		RiskLevel:           entity.RiskLevel,
		Source:              entity.Source,
		AccountManager:      entity.AccountManager,
		Tags:                decodeStringSlice(entity.Tags),
		MaskedFields:        decodeStringSlice(entity.MaskedFields),
		MembershipSnapshot:  decodeMembershipSnapshot(entity.MembershipSnapshot),
		Metadata:            decodeMetadata(entity.Metadata),
		Notes:               entity.Notes,
		CreatedAt:           entity.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           entity.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return result
}

func copyCustomerToEntity(entity *customermodel.Customer, c Customer) {
	entity.CustomerID = c.ID
	if entity.CustomerID == "" {
		entity.CustomerID = uuid.NewString()
	}
	entity.Name = c.Name
	entity.Type = c.Type
	entity.Email = c.Email
	entity.Phone = c.Phone
	entity.Country = c.Country
	entity.Region = c.Region
	entity.MembershipTier = c.MembershipTier
	entity.MembershipTierLabel = c.MembershipTierLabel
	entity.GrowthValue = c.GrowthValue
	entity.Points = c.Points
	entity.LastOrderAmount = c.LastOrderAmount
	entity.LastOrderAt = parseTimePtr(c.LastOrderAt)
	entity.Status = c.Status
	entity.RiskLevel = c.RiskLevel
	entity.Source = c.Source
	entity.AccountManager = c.AccountManager
	entity.Tags = encodeStringSlice(c.Tags)
	entity.MaskedFields = encodeStringSlice(c.MaskedFields)
	entity.MembershipSnapshot = encodeMembershipSnapshot(c.MembershipSnapshot)
	entity.Metadata = encodeMetadata(c.Metadata)
	entity.Notes = c.Notes
	if created := parseTime(c.CreatedAt); !created.IsZero() {
		entity.CreatedAt = created
	}
	if updated := parseTime(c.UpdatedAt); !updated.IsZero() {
		entity.UpdatedAt = updated
	}
}

func encodeStringSlice(values []string) datatypes.JSON {
	if len(values) == 0 {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeStringSlice(data datatypes.JSON) []string {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func encodeMembershipSnapshot(snapshot *MembershipSnapshot) datatypes.JSON {
	if snapshot == nil {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeMembershipSnapshot(data datatypes.JSON) *MembershipSnapshot {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var snapshot MembershipSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil
	}
	return &snapshot
}

func encodeMetadata(meta map[string]any) datatypes.JSON {
	if len(meta) == 0 {
		return datatypes.JSON("null")
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(payload)
}

func decodeMetadata(data datatypes.JSON) map[string]any {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil
	}
	return meta
}

func parseTimePtr(value string) *time.Time {
	t := parseTime(value)
	if t.IsZero() {
		return nil
	}
	return &t
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Time{}
}

func normalizeTags(tags []string) []string {
	var out []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if strings.Contains(tag, ",") {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				p := strings.TrimSpace(part)
				if p != "" {
					out = append(out, p)
				}
			}
			continue
		}
		out = append(out, tag)
	}
	return out
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

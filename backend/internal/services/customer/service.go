package customer

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// Service 提供客户域的占位实现，后续可替换为真实 CRM/DB。
type Service struct {
	deps *app.Deps
}

// NewService 创建客户服务。
func NewService(deps *app.Deps) *Service {
	return &Service{deps: deps}
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
func (s *Service) ListCustomers(ctx context.Context, tenantUUID string, filters ListFilters) (*CustomerListResult, error) {
	data := mockCustomers(tenantUUID)
	filtered := applyFilters(data, filters)

	sortKey := strings.TrimSpace(filters.Sort)
	if sortKey == "" {
		sortKey = "-createdAt"
	}
	filtered = sortCustomers(filtered, sortKey)

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	total := len(filtered)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return &CustomerListResult{
		Data: filtered[start:end],
		Meta: CustomerListMeta{Total: total},
	}, nil
}

func applyFilters(customers []Customer, filters ListFilters) []Customer {
	if len(customers) == 0 {
		return customers
	}
	keyword := strings.ToLower(strings.TrimSpace(filters.Keyword))
	wantsTags := normalizeTags(filters.Tags)

	var out []Customer
	for _, c := range customers {
		if keyword != "" && !matchesKeyword(c, keyword) {
			continue
		}
		if filters.Tier != "" && !strings.EqualFold(c.MembershipTier, filters.Tier) {
			continue
		}
		if filters.Source != "" && !strings.EqualFold(c.Source, filters.Source) {
			continue
		}
		if filters.Region != "" && !strings.Contains(strings.ToLower(c.Region), strings.ToLower(filters.Region)) {
			continue
		}
		if filters.RiskLevel != "" && !strings.EqualFold(c.RiskLevel, filters.RiskLevel) {
			continue
		}
		if filters.Type != "" && !strings.EqualFold(c.Type, filters.Type) {
			continue
		}
		if len(wantsTags) > 0 && !hasAllTags(c.Tags, wantsTags) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func matchesKeyword(c Customer, keyword string) bool {
	fields := []string{c.ID, c.Name, c.Email, c.Phone, c.AccountManager}
	for _, f := range fields {
		if f == "" {
			continue
		}
		if strings.Contains(strings.ToLower(f), keyword) {
			return true
		}
	}
	return false
}

func hasAllTags(tags []string, wants []string) bool {
	if len(wants) == 0 {
		return true
	}
	if len(tags) == 0 {
		return false
	}
	lookup := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		lookup[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}
	for _, want := range wants {
		if _, ok := lookup[strings.ToLower(want)]; !ok {
			return false
		}
	}
	return true
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

func sortCustomers(customers []Customer, key string) []Customer {
	if len(customers) <= 1 {
		return customers
	}
	descending := false
	field := strings.TrimSpace(key)
	if strings.HasPrefix(field, "-") {
		descending = true
		field = strings.TrimPrefix(field, "-")
	}
	sort.SliceStable(customers, func(i, j int) bool {
		less := compareByField(customers[i], customers[j], field)
		if descending {
			return !less
		}
		return less
	})
	return customers
}

func compareByField(a, b Customer, field string) bool {
	switch field {
	case "lastOrderAt":
		return parseTime(a.LastOrderAt).Before(parseTime(b.LastOrderAt))
	case "name":
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "tier", "membershipTier":
		return strings.ToLower(a.MembershipTier) < strings.ToLower(b.MembershipTier)
	case "status":
		return strings.ToLower(a.Status) < strings.ToLower(b.Status)
	default:
		return parseTime(a.CreatedAt).Before(parseTime(b.CreatedAt))
	}
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

func mockCustomers(tenantUUID string) []Customer {
	now := time.Now()
	tenDaysAgo := now.AddDate(0, 0, -10)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	sixtyDaysAgo := now.AddDate(0, 0, -60)

	base := []Customer{
		{
			ID:                  "cust-1001",
			Name:                "张三",
			Type:                "individual",
			Email:               "zhangsan@example.com",
			Phone:               "+86-188-0000-1111",
			Country:             "CN",
			Region:              "华东/上海",
			MembershipTier:      "gold",
			MembershipTierLabel: "金卡",
			GrowthValue:         180,
			Points:              5200,
			LastOrderAmount:     1289.5,
			LastOrderAt:         thirtyDaysAgo.Format(time.RFC3339),
			Status:              "active",
			RiskLevel:           "low",
			Source:              "website",
			AccountManager:      "ops-li",
			Tags:                []string{"VIP", "重点客户"},
			CreatedAt:           now.AddDate(-1, 0, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			MaskedFields:        []string{"email"},
			MembershipSnapshot: &MembershipSnapshot{
				CustomerID:      "cust-1001",
				Tier:            "gold",
				GrowthValue:     180,
				Points:          5200,
				RetentionStatus: "safe",
				Benefits: []MembershipBenefit{
					{Name: "生日礼券", Used: true},
					{Name: "专属客服", Used: false},
				},
			},
			Metadata: map[string]any{"notes": "近90天GMV 20w+"},
		},
		{
			ID:                  "cust-1002",
			Name:                "李雷",
			Type:                "individual",
			Email:               "lilei@example.com",
			Phone:               "+86-139-0000-2222",
			Country:             "CN",
			Region:              "华北/北京",
			MembershipTier:      "silver",
			MembershipTierLabel: "银卡",
			GrowthValue:         80,
			Points:              1800,
			LastOrderAmount:     640,
			LastOrderAt:         tenDaysAgo.Format(time.RFC3339),
			Status:              "active",
			RiskLevel:           "medium",
			Source:              "referral",
			AccountManager:      "ops-wang",
			Tags:                []string{"高潜力", "自营渠道"},
			CreatedAt:           now.AddDate(-2, 0, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			MembershipSnapshot: &MembershipSnapshot{
				CustomerID:      "cust-1002",
				Tier:            "silver",
				GrowthValue:     80,
				Points:          1800,
				RetentionStatus: "warning",
			},
		},
		{
			ID:                  "cust-1003",
			Name:                "ACME 批发",
			Type:                "enterprise",
			Email:               "ops@acme.example.com",
			Phone:               "+86-21-5099-5566",
			Country:             "CN",
			Region:              "华东/苏州",
			MembershipTier:      "platinum",
			MembershipTierLabel: "白金",
			GrowthValue:         420,
			Points:              9800,
			LastOrderAmount:     53200,
			LastOrderAt:         sixtyDaysAgo.Format(time.RFC3339),
			Status:              "active",
			RiskLevel:           "low",
			Source:              "offline",
			AccountManager:      "b2b-queen",
			Tags:                []string{"B2B", "合同客户"},
			CreatedAt:           now.AddDate(-3, 0, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			MembershipSnapshot: &MembershipSnapshot{
				CustomerID:      "cust-1003",
				Tier:            "platinum",
				GrowthValue:     420,
				Points:          9800,
				RetentionStatus: "safe",
			},
		},
		{
			ID:                  "cust-1004",
			Name:                "王五",
			Type:                "individual",
			Email:               "wangwu@example.com",
			Phone:               "+86-137-1111-3333",
			Country:             "CN",
			Region:              "华南/深圳",
			MembershipTier:      "gold",
			MembershipTierLabel: "金卡",
			GrowthValue:         130,
			Points:              3200,
			LastOrderAmount:     299,
			LastOrderAt:         tenDaysAgo.AddDate(0, 0, -2).Format(time.RFC3339),
			Status:              "inactive",
			RiskLevel:           "high",
			Source:              "miniapp",
			AccountManager:      "ops-li",
			Tags:                []string{"风险关注"},
			CreatedAt:           now.AddDate(-1, -2, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			MaskedFields:        []string{"phone"},
			MembershipSnapshot: &MembershipSnapshot{
				CustomerID:      "cust-1004",
				Tier:            "gold",
				GrowthValue:     130,
				Points:          3200,
				RetentionStatus: "downgrade",
			},
		},
		{
			ID:                  "cust-1005",
			Name:                "华东旗舰店",
			Type:                "enterprise",
			Email:               "buyer@flagship.example.com",
			Phone:               "+86-576-2200-8888",
			Country:             "CN",
			Region:              "华东/宁波",
			MembershipTier:      "platinum",
			MembershipTierLabel: "白金",
			GrowthValue:         600,
			Points:              15800,
			LastOrderAmount:     98000,
			LastOrderAt:         now.AddDate(0, 0, -5).Format(time.RFC3339),
			Status:              "active",
			RiskLevel:           "low",
			Source:              "offline",
			AccountManager:      "b2b-queen",
			Tags:                []string{"分销", "重点客户"},
			CreatedAt:           now.AddDate(-4, 0, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
			MembershipSnapshot: &MembershipSnapshot{
				CustomerID:      "cust-1005",
				Tier:            "platinum",
				GrowthValue:     600,
				Points:          15800,
				RetentionStatus: "safe",
			},
		},
		{
			ID:                  "cust-1006",
			Name:                "测试账号",
			Type:                "individual",
			Email:               "demo@example.com",
			Phone:               "+86-130-0000-1000",
			Country:             "CN",
			Region:              "华中/武汉",
			MembershipTier:      "silver",
			MembershipTierLabel: "银卡",
			GrowthValue:         45,
			Points:              400,
			LastOrderAmount:     120,
			LastOrderAt:         now.AddDate(0, 0, -1).Format(time.RFC3339),
			Status:              "blocked",
			RiskLevel:           "high",
			Source:              "website",
			AccountManager:      "ops-li",
			Tags:                []string{"测试", "黑名单"},
			CreatedAt:           now.AddDate(-1, -6, 0).Format(time.RFC3339),
			UpdatedAt:           now.Format(time.RFC3339),
		},
	}

	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID != "" {
		prefix := tenantUUID
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		for i := range base {
			base[i].ID = prefix + "-" + base[i].ID
			if base[i].MembershipSnapshot != nil {
				base[i].MembershipSnapshot.CustomerID = base[i].ID
			}
		}
	}

	return base
}

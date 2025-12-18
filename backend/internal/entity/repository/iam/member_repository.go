package iam

import (
	"context"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/iam"
	"gorm.io/gorm"
)

// MemberRepository exposes helpers to query IAM members for selection lists.
type MemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository builds a repository backed by gorm DB.
func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// MemberSearchResult is a projection for directory dropdowns.
type MemberSearchResult struct {
	ID          uint64
	Username    string
	DisplayName string
	Email       string
}

// SearchActiveMembers returns active IAM members within a tenant filtered by keyword.
func (r *MemberRepository) SearchActiveMembers(
	ctx context.Context,
	tenantUUID string,
	keyword string,
	limit int,
) ([]MemberSearchResult, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := r.db.WithContext(ctx).
		Table(models.S(models.TableIAMMembers)+" AS m").
		Select("m.id", "m.username", "COALESCE(m.display_name, '') AS display_name", "COALESCE(u.email, '') AS email").
		Joins("LEFT JOIN "+models.S(models.TableIAMUsers)+" AS u ON u.id = m.user_id").
		Where("m.tenant_uuid = ?", tenantUUID).
		Where("m.status = ?", iam.StatusActive)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		wildcard := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			r.db.
				Where("LOWER(m.username) LIKE ?", wildcard).
				Or("LOWER(COALESCE(m.display_name, '')) LIKE ?", wildcard).
				Or("LOWER(COALESCE(u.email, '')) LIKE ?", wildcard),
		)
	}
	var members []MemberSearchResult
	if err := query.Order("m.display_name ASC, m.username ASC").Limit(limit).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

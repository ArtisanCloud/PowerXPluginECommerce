package channel_master

import (
	"context"
	"errors"
	"strings"

	iamrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/iam"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// OwnerDirectory provides lookup helpers for IAM members to serve as channel owners.
type OwnerDirectory struct {
	repo *iamrepo.MemberRepository
}

// OwnerOption describes minimal IAM member info required by the UI dropdown.
type OwnerOption struct {
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// NewOwnerDirectory returns an owner directory using shared dependencies.
func NewOwnerDirectory(deps *app.Deps) *OwnerDirectory {
	if deps == nil || deps.DB == nil {
		return nil
	}
	return &OwnerDirectory{repo: iamrepo.NewMemberRepository(deps.DB)}
}

// Search returns owner candidates matching the keyword with a sane limit.
func (d *OwnerDirectory) Search(ctx context.Context, tenantUUID, keyword string, limit int) ([]OwnerOption, error) {
	if d == nil || d.repo == nil {
		return nil, errors.New("owner directory unavailable")
	}
	if strings.TrimSpace(tenantUUID) == "" {
		return nil, errors.New("tenant uuid is required")
	}
	results, err := d.repo.SearchActiveMembers(ctx, tenantUUID, keyword, limit)
	if err != nil {
		return nil, err
	}
	opts := make([]OwnerOption, 0, len(results))
	for _, member := range results {
		opts = append(opts, OwnerOption{
			ID:          member.ID,
			Username:    member.Username,
			DisplayName: member.DisplayName,
			Email:       member.Email,
		})
	}
	return opts, nil
}

package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type RoutingScoreProfileRepository struct {
	*Repository[LogisticsModel.RoutingScoreProfile]
}

func NewRoutingScoreProfileRepository(db *gorm.DB) *RoutingScoreProfileRepository {
	return &RoutingScoreProfileRepository{Repository: NewRepository[LogisticsModel.RoutingScoreProfile](db)}
}

func (r *RoutingScoreProfileRepository) GetByName(ctx context.Context, name string) (*LogisticsModel.RoutingScoreProfile, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RoutingScoreProfile
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND name = ?", tenantUUID, strings.TrimSpace(name)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RoutingScoreProfileRepository) FirstEnabled(ctx context.Context) (*LogisticsModel.RoutingScoreProfile, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RoutingScoreProfile
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND enabled = ?", tenantUUID, true).
		Order("updated_at DESC, created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RoutingScoreProfileRepository) Save(ctx context.Context, row *LogisticsModel.RoutingScoreProfile) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

type RoutingScoreSimulationRepository struct {
	*Repository[LogisticsModel.RoutingScoreSimulation]
}

func NewRoutingScoreSimulationRepository(db *gorm.DB) *RoutingScoreSimulationRepository {
	return &RoutingScoreSimulationRepository{Repository: NewRepository[LogisticsModel.RoutingScoreSimulation](db)}
}

func (r *RoutingScoreSimulationRepository) Create(ctx context.Context, row *LogisticsModel.RoutingScoreSimulation) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(row).Error
}

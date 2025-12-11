package spu

import (
	"context"
	"errors"
	"strings"

	productrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/domain/repository/product"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// ErrMissingTenant indicates the request context did not carry tenant metadata.
var ErrMissingTenant = errors.New("tenant context missing")

// Service orchestrates tenant-scoped SPU lifecycle operations.
type Service struct {
	deps        *app.Deps
	spuRepo     *productrepo.SPURepository
	versionRepo *productrepo.VersionRepository
	channelRepo *productrepo.ChannelRepository
	planRepo    *productrepo.SubscriptionPlanRepository
	locales     *LocaleService
}

// NewService wires repositories and helpers using the shared dependency bundle.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		panic("spu service requires initialized DB dependency")
	}
	return &Service{
		deps:        deps,
		spuRepo:     productrepo.NewSPURepository(deps.DB),
		versionRepo: productrepo.NewVersionRepository(deps.DB),
		channelRepo: productrepo.NewChannelRepository(deps.DB),
		planRepo:    productrepo.NewSubscriptionPlanRepository(deps.DB),
		locales:     NewLocaleService([]string{"zh-CN"}),
	}
}

// UpsertSPURequest captures fields required to create or update an SPU draft.
type UpsertSPURequest struct {
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	CategoryID      string          `json:"categoryId"`
	CategoryPath    string          `json:"categoryPath"`
	BrandID         string          `json:"brandId"`
	DefaultLocale   string          `json:"defaultLocale"`
	Tags            []string        `json:"tags"`
	ResponsibleUser string          `json:"responsibleUser"`
	Locales         []LocaleContent `json:"locales"`
}

// Normalize trims whitespace and deduplicates tags for consistent processing.
func (r *UpsertSPURequest) Normalize() {
	r.Code = strings.TrimSpace(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	r.Type = strings.TrimSpace(strings.ToLower(r.Type))
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.CategoryPath = strings.TrimSpace(r.CategoryPath)
	r.BrandID = strings.TrimSpace(r.BrandID)
	r.DefaultLocale = strings.TrimSpace(r.DefaultLocale)
	r.ResponsibleUser = strings.TrimSpace(r.ResponsibleUser)
	tagSet := map[string]struct{}{}
	for _, tag := range r.Tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			tagSet[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	r.Tags = make([]string, 0, len(tagSet))
	for tag := range tagSet {
		r.Tags = append(r.Tags, tag)
	}
}

// ValidateUpsertRequest enforces field level constraints before persistence.
func (s *Service) ValidateUpsertRequest(ctx context.Context, req UpsertSPURequest) error {
	if s == nil {
		return errors.New("spu service not initialized")
	}
	req.Normalize()
	req.Locales = s.locales.Normalize(req.Locales)
	var errs ValidationErrors
	if req.Code == "" {
		errs = errs.add("code", "code is required")
	} else if len(req.Code) > 120 {
		errs = errs.add("code", "code cannot exceed 120 characters")
	}
	if req.Name == "" {
		errs = errs.add("name", "name is required")
	}
	supportedTypes := map[string]struct{}{"one_time": {}, "subscription": {}, "bundle": {}}
	if req.Type == "" {
		errs = errs.add("type", "type is required")
	} else if _, ok := supportedTypes[req.Type]; !ok {
		errs = errs.add("type", "unsupported type")
	}
	if req.CategoryID == "" {
		errs = errs.add("categoryId", "category id is required")
	}
	if req.CategoryPath == "" {
		errs = errs.add("categoryPath", "category path is required")
	}
	if req.DefaultLocale == "" {
		errs = errs.add("defaultLocale", "default locale is required")
	}
	localeErrs := s.locales.Validate(req.DefaultLocale, req.Locales)
	if !localeErrs.empty() {
		errs = append(errs, localeErrs...)
	}
	if errs.empty() {
		if _, err := s.tenantFromContext(ctx); err != nil {
			return err
		}
		return nil
	}
	return errs
}

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = s.deps.Ctx
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", ErrMissingTenant
}

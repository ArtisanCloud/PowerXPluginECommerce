package product_category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	catrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TemplateFieldInput struct {
	FieldKey        string         `json:"fieldKey"`
	Group           string         `json:"group,omitempty"`
	FieldType       string         `json:"fieldType"`
	Required        bool           `json:"required"`
	SortOrder       int            `json:"sortOrder"`
	ValidationRules map[string]any `json:"validationRules,omitempty"`
	DefaultValue    any            `json:"defaultValue,omitempty"`
	I18nLabel       map[string]any `json:"i18nLabel,omitempty"`
	I18nHelp        map[string]any `json:"i18nHelp,omitempty"`
	Inheritable     bool           `json:"inheritable"`
}

type TemplateCreateRequest struct {
	Name            string               `json:"name"`
	Status          string               `json:"status"`
	ApplicableLevel *int                 `json:"applicableLevel,omitempty"`
	Notes           string               `json:"notes,omitempty"`
	Fields          []TemplateFieldInput `json:"fields"`
}

type TemplateUpdateRequest struct {
	Name            *string               `json:"name"`
	Status          *string               `json:"status"`
	ApplicableLevel *int                  `json:"applicableLevel,omitempty"`
	Notes           *string               `json:"notes,omitempty"`
	Fields          *[]TemplateFieldInput `json:"fields,omitempty"`
}

type TemplatePublishRequest struct {
	PublishedBy string `json:"publishedBy,omitempty"`
}

type TemplateRollbackRequest struct {
	TargetVersionID string `json:"targetVersionId"`
	RollbackBy      string `json:"rollbackBy,omitempty"`
}

type TemplateListFilters struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type TemplateListResult struct {
	Items    []CategoryTemplateDetail `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
}

type CategoryTemplateDetail struct {
	ID                        string                `json:"id"`
	Name                      string                `json:"name"`
	Status                    string                `json:"status"`
	ApplicableLevel           *int                  `json:"applicableLevel,omitempty"`
	Notes                     string                `json:"notes,omitempty"`
	CurrentPublishedVersionID *string               `json:"currentPublishedVersionId,omitempty"`
	CurrentPublishedVersionNo int                   `json:"currentPublishedVersionNo"`
	LastPublishedAt           *time.Time            `json:"lastPublishedAt,omitempty"`
	Fields                    []TemplateFieldOutput `json:"fields,omitempty"`
	UpdatedAt                 time.Time             `json:"updatedAt"`
	CreatedAt                 time.Time             `json:"createdAt"`
}

type TemplateFieldOutput struct {
	ID              string         `json:"id"`
	FieldKey        string         `json:"fieldKey"`
	Group           string         `json:"group,omitempty"`
	FieldType       string         `json:"fieldType"`
	Required        bool           `json:"required"`
	SortOrder       int            `json:"sortOrder"`
	ValidationRules map[string]any `json:"validationRules,omitempty"`
	DefaultValue    any            `json:"defaultValue,omitempty"`
	I18nLabel       map[string]any `json:"i18nLabel,omitempty"`
	I18nHelp        map[string]any `json:"i18nHelp,omitempty"`
	Inheritable     bool           `json:"inheritable"`
}

type TemplateVersionSummary struct {
	ID            string     `json:"id"`
	TemplateID    string     `json:"templateId"`
	VersionNumber int        `json:"versionNumber"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	PublishedBy   string     `json:"publishedBy,omitempty"`
	RollbackFrom  *string    `json:"rollbackFrom,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type EffectiveTemplate struct {
	TemplateID    string                 `json:"templateId"`
	VersionID     string                 `json:"versionId"`
	VersionNumber int                    `json:"versionNumber"`
	PublishedAt   *time.Time             `json:"publishedAt,omitempty"`
	Template      CategoryTemplateDetail `json:"template"`
	Fields        []TemplateFieldOutput  `json:"fields"`
}

type templateSnapshot struct {
	Template CategoryTemplateDetail `json:"template"`
	Fields   []TemplateFieldOutput  `json:"fields"`
}

func (s *Service) ListTemplates(ctx context.Context, filters TemplateListFilters) (*TemplateListResult, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	records, total, err := s.TemplateRepo.ListTemplates(ctx, tenantID, catrepo.TemplateListFilters{
		Keyword:  filters.Keyword,
		Status:   filters.Status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]CategoryTemplateDetail, 0, len(records))
	for _, rec := range records {
		items = append(items, mapTemplateDetail(rec, nil))
	}
	return &TemplateListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) GetTemplate(ctx context.Context, id string) (*CategoryTemplateDetail, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	template, err := s.TemplateRepo.GetTemplate(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	fields, err := s.TemplateRepo.ListFields(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	outFields := make([]TemplateFieldOutput, 0, len(fields))
	for _, f := range fields {
		outFields = append(outFields, mapTemplateFieldOutput(f))
	}
	detail := mapTemplateDetail(*template, outFields)
	return &detail, nil
}

func (s *Service) CreateTemplate(ctx context.Context, req TemplateCreateRequest) (*CategoryTemplateDetail, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(strings.ToLower(req.Status))
	if req.Status == "" {
		req.Status = productcategory.TemplateStatusEnabled
	}
	vErrs := ValidationErrors{}
	if req.Name == "" {
		vErrs = vErrs.add("name", "name 为必填")
	}
	if req.Status != productcategory.TemplateStatusEnabled && req.Status != productcategory.TemplateStatusDisabled {
		vErrs = vErrs.add("status", "status 必须为 enabled/disabled")
	}
	fieldKeys := map[string]struct{}{}
	for idx, field := range req.Fields {
		k := strings.TrimSpace(field.FieldKey)
		if k == "" {
			vErrs = vErrs.add(fmt.Sprintf("fields[%d].fieldKey", idx), "fieldKey 为必填")
			continue
		}
		if _, ok := fieldKeys[strings.ToLower(k)]; ok {
			vErrs = vErrs.add("fields", fmt.Sprintf("fieldKey 重复：%s", k))
		}
		fieldKeys[strings.ToLower(k)] = struct{}{}
		if strings.TrimSpace(field.FieldType) == "" {
			vErrs = vErrs.add(fmt.Sprintf("fields[%d].fieldType", idx), "fieldType 为必填")
		}
	}
	if !vErrs.empty() {
		return nil, vErrs
	}

	var created *CategoryTemplateDetail
	now := time.Now().UTC()
	err = s.TemplateRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewTemplateRepository(tx)
		id := utils.NewUUID()
		tpl := &productcategory.CategoryTemplate{
			ID:              id,
			TenantUUID:      tenantID,
			Name:            req.Name,
			Status:          req.Status,
			ApplicableLevel: req.ApplicableLevel,
			Notes:           strings.TrimSpace(req.Notes),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := tx.Create(tpl).Error; err != nil {
			return err
		}
		fields := make([]productcategory.CategoryTemplateField, 0, len(req.Fields))
		for _, input := range req.Fields {
			fields = append(fields, buildTemplateField(tenantID, id, input))
		}
		if err := repoTx.ReplaceFields(ctx, tx, tenantID, id, fields); err != nil {
			return err
		}
		outFields := make([]TemplateFieldOutput, 0, len(fields))
		for _, f := range fields {
			outFields = append(outFields, mapTemplateFieldOutput(f))
		}
		detail := mapTemplateDetail(*tpl, outFields)
		created = &detail
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, id string, req TemplateUpdateRequest) (*CategoryTemplateDetail, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	var updated *CategoryTemplateDetail
	err = s.TemplateRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewTemplateRepository(tx)
		existing, err := repoTx.GetTemplate(ctx, tenantID, id)
		if err != nil {
			return err
		}
		updates := map[string]any{"updated_at": time.Now().UTC()}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return ValidationErrors{}.add("name", "name 不能为空")
			}
			updates["name"] = name
		}
		if req.Status != nil {
			status := strings.TrimSpace(strings.ToLower(*req.Status))
			if status != productcategory.TemplateStatusEnabled && status != productcategory.TemplateStatusDisabled {
				return ValidationErrors{}.add("status", "status 必须为 enabled/disabled")
			}
			updates["status"] = status
		}
		if req.ApplicableLevel != nil {
			updates["applicable_level"] = req.ApplicableLevel
		}
		if req.Notes != nil {
			updates["notes"] = strings.TrimSpace(*req.Notes)
		}
		if err := tx.Model(&productcategory.CategoryTemplate{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(updates).Error; err != nil {
			return err
		}
		if req.Fields != nil {
			fieldKeys := map[string]struct{}{}
			for idx, field := range *req.Fields {
				k := strings.TrimSpace(field.FieldKey)
				if k == "" {
					return ValidationErrors{}.add(fmt.Sprintf("fields[%d].fieldKey", idx), "fieldKey 为必填")
				}
				if _, ok := fieldKeys[strings.ToLower(k)]; ok {
					return ValidationErrors{}.add("fields", fmt.Sprintf("fieldKey 重复：%s", k))
				}
				fieldKeys[strings.ToLower(k)] = struct{}{}
				if strings.TrimSpace(field.FieldType) == "" {
					return ValidationErrors{}.add(fmt.Sprintf("fields[%d].fieldType", idx), "fieldType 为必填")
				}
			}
			fields := make([]productcategory.CategoryTemplateField, 0, len(*req.Fields))
			for _, input := range *req.Fields {
				fields = append(fields, buildTemplateField(tenantID, id, input))
			}
			if err := repoTx.ReplaceFields(ctx, tx, tenantID, id, fields); err != nil {
				return err
			}
		}
		// reload
		tpl, err := repoTx.GetTemplate(ctx, tenantID, id)
		if err != nil {
			return err
		}
		fields, err := repoTx.ListFields(ctx, tenantID, id)
		if err != nil {
			return err
		}
		outFields := make([]TemplateFieldOutput, 0, len(fields))
		for _, f := range fields {
			outFields = append(outFields, mapTemplateFieldOutput(f))
		}
		detail := mapTemplateDetail(*tpl, outFields)
		updated = &detail
		_ = existing
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) PublishTemplate(ctx context.Context, id string, req TemplatePublishRequest) (*TemplateVersionSummary, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	publisher := strings.TrimSpace(req.PublishedBy)
	if publisher == "" {
		publisher = "system"
	}
	now := time.Now().UTC()

	var summary *TemplateVersionSummary
	err = s.TemplateRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewTemplateRepository(tx)
		tpl, err := repoTx.GetTemplate(ctx, tenantID, id)
		if err != nil {
			return err
		}
		fields, err := repoTx.ListFields(ctx, tenantID, id)
		if err != nil {
			return err
		}
		outFields := make([]TemplateFieldOutput, 0, len(fields))
		for _, f := range fields {
			outFields = append(outFields, mapTemplateFieldOutput(f))
		}
		detail := mapTemplateDetail(*tpl, outFields)
		snap := templateSnapshot{Template: detail, Fields: outFields}
		body, _ := json.Marshal(snap)
		maxNo, err := repoTx.MaxVersionNumber(ctx, tenantID, id)
		if err != nil {
			return err
		}
		version := &productcategory.CategoryTemplateVersion{
			ID:            utils.NewUUID(),
			TenantUUID:    tenantID,
			TemplateID:    id,
			VersionNumber: maxNo + 1,
			PublishedAt:   &now,
			PublishedBy:   publisher,
			Snapshot:      datatypes.JSON(body),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		if err := tx.Model(&productcategory.CategoryTemplate{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(map[string]any{
				"current_published_version_id": version.ID,
				"current_published_version_no": version.VersionNumber,
				"last_published_at":            now,
				"updated_at":                   now,
			}).Error; err != nil {
			return err
		}
		summary = &TemplateVersionSummary{
			ID:            version.ID,
			TemplateID:    version.TemplateID,
			VersionNumber: version.VersionNumber,
			PublishedAt:   version.PublishedAt,
			PublishedBy:   version.PublishedBy,
			RollbackFrom:  version.RollbackFrom,
			CreatedAt:     version.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *Service) RollbackTemplate(ctx context.Context, id string, req TemplateRollbackRequest) (*TemplateVersionSummary, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	req.TargetVersionID = strings.TrimSpace(req.TargetVersionID)
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req.TargetVersionID == "" {
		return nil, ValidationErrors{}.add("targetVersionId", "targetVersionId 为必填")
	}
	actor := strings.TrimSpace(req.RollbackBy)
	if actor == "" {
		actor = "system"
	}
	now := time.Now().UTC()

	var summary *TemplateVersionSummary
	err = s.TemplateRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewTemplateRepository(tx)
		target, err := repoTx.GetVersion(ctx, tenantID, req.TargetVersionID)
		if err != nil {
			return err
		}
		if strings.TrimSpace(target.TemplateID) != id {
			return ValidationErrors{}.add("targetVersionId", "目标版本不属于该模板")
		}
		maxNo, err := repoTx.MaxVersionNumber(ctx, tenantID, id)
		if err != nil {
			return err
		}
		version := &productcategory.CategoryTemplateVersion{
			ID:            utils.NewUUID(),
			TenantUUID:    tenantID,
			TemplateID:    id,
			VersionNumber: maxNo + 1,
			PublishedAt:   &now,
			PublishedBy:   actor,
			RollbackFrom:  &target.ID,
			Snapshot:      target.Snapshot,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		if err := tx.Model(&productcategory.CategoryTemplate{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(map[string]any{
				"current_published_version_id": version.ID,
				"current_published_version_no": version.VersionNumber,
				"last_published_at":            now,
				"updated_at":                   now,
			}).Error; err != nil {
			return err
		}
		summary = &TemplateVersionSummary{
			ID:            version.ID,
			TemplateID:    version.TemplateID,
			VersionNumber: version.VersionNumber,
			PublishedAt:   version.PublishedAt,
			PublishedBy:   version.PublishedBy,
			RollbackFrom:  version.RollbackFrom,
			CreatedAt:     version.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func (s *Service) ListTemplateVersions(ctx context.Context, templateID string, limit int) ([]TemplateVersionSummary, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return nil, errors.New("templateId is required")
	}
	versions, err := s.TemplateRepo.ListVersions(ctx, tenantID, templateID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]TemplateVersionSummary, 0, len(versions))
	for _, v := range versions {
		out = append(out, TemplateVersionSummary{
			ID:            v.ID,
			TemplateID:    v.TemplateID,
			VersionNumber: v.VersionNumber,
			PublishedAt:   v.PublishedAt,
			PublishedBy:   v.PublishedBy,
			RollbackFrom:  v.RollbackFrom,
			CreatedAt:     v.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) GetEffectiveTemplateByCategory(ctx context.Context, categoryID string) (*EffectiveTemplate, error) {
	if s == nil || !s.Ready() || s.TemplateRepo == nil || s.CategoryRepo == nil {
		return nil, errors.New("template service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, errors.New("categoryId is required")
	}
	category, err := s.CategoryRepo.GetByID(ctx, tenantID, categoryID, false)
	if err != nil {
		return nil, err
	}
	ids := parseCategoryPathIDs(category.Path)
	if len(ids) == 0 {
		ids = []string{category.ID}
	}
	type row struct {
		ID         string
		TemplateID *string
	}
	var rows []row
	if err := s.deps.DB.WithContext(ctx).
		Table(productcategory.ProductCategory{}.TableName()).
		Select("id, template_id").
		Where("tenant_uuid = ? AND id IN ?", tenantID, ids).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	m := map[string]*string{}
	for _, r := range rows {
		m[strings.TrimSpace(r.ID)] = r.TemplateID
	}
	for i := len(ids) - 1; i >= 0; i-- {
		id := strings.TrimSpace(ids[i])
		tid := m[id]
		if tid == nil || strings.TrimSpace(*tid) == "" {
			continue
		}
		version, err := s.TemplateRepo.LatestPublishedVersion(ctx, tenantID, strings.TrimSpace(*tid))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		var snap templateSnapshot
		if len(version.Snapshot) > 0 {
			_ = json.Unmarshal(version.Snapshot, &snap)
		}
		if snap.Template.ID == "" {
			// fallback: use current template detail
			detail, err := s.GetTemplate(ctx, strings.TrimSpace(*tid))
			if err == nil && detail != nil {
				snap.Template = *detail
				snap.Fields = detail.Fields
			}
		}
		return &EffectiveTemplate{
			TemplateID:    strings.TrimSpace(*tid),
			VersionID:     version.ID,
			VersionNumber: version.VersionNumber,
			PublishedAt:   version.PublishedAt,
			Template:      snap.Template,
			Fields:        snap.Fields,
		}, nil
	}
	return nil, nil
}

func (s *Service) ValidateAttributesForCategory(ctx context.Context, categoryID string, attributes map[string]any) (ValidationErrors, error) {
	effective, err := s.GetEffectiveTemplateByCategory(ctx, categoryID)
	if err != nil {
		if isMissingTableError(err) {
			return nil, nil
		}
		return nil, err
	}
	if effective == nil {
		return nil, nil
	}
	return validateAttributes(effective.Fields, attributes), nil
}

func (s *Service) ValidateAttributesPreview(fields []TemplateFieldOutput, attributes map[string]any) ValidationErrors {
	return validateAttributes(fields, attributes)
}

func validateAttributes(fields []TemplateFieldOutput, attributes map[string]any) ValidationErrors {
	if attributes == nil {
		attributes = map[string]any{}
	}
	errs := ValidationErrors{}
	for _, field := range fields {
		key := strings.TrimSpace(field.FieldKey)
		if key == "" {
			continue
		}
		val, ok := attributes[key]
		if field.Required {
			if !ok || val == nil || (isString(val) && strings.TrimSpace(fmt.Sprintf("%v", val)) == "") {
				errs = errs.add(key, "字段为必填")
				continue
			}
		}
		if !ok || val == nil {
			continue
		}
		if !matchFieldType(field.FieldType, val) {
			errs = errs.add(key, fmt.Sprintf("字段类型不匹配，期望 %s", field.FieldType))
			continue
		}
		if options := enumOptions(field.ValidationRules); len(options) > 0 {
			if s, ok := val.(string); ok {
				if !containsString(options, s) {
					errs = errs.add(key, "字段值不在允许范围内")
				}
			}
		}
	}
	return errs
}

func buildTemplateField(tenantID string, templateID string, input TemplateFieldInput) productcategory.CategoryTemplateField {
	tenantID = strings.TrimSpace(tenantID)
	templateID = strings.TrimSpace(templateID)
	fieldKey := strings.TrimSpace(input.FieldKey)
	group := strings.TrimSpace(input.Group)
	fieldType := strings.TrimSpace(strings.ToLower(input.FieldType))
	if fieldType == "" {
		fieldType = "string"
	}
	rules, _ := json.Marshal(input.ValidationRules)
	def, _ := json.Marshal(input.DefaultValue)
	label, _ := json.Marshal(input.I18nLabel)
	help, _ := json.Marshal(input.I18nHelp)
	return productcategory.CategoryTemplateField{
		ID:              utils.NewUUID(),
		TenantUUID:      tenantID,
		TemplateID:      templateID,
		FieldKey:        fieldKey,
		Group:           group,
		FieldType:       fieldType,
		Required:        input.Required,
		SortOrder:       input.SortOrder,
		ValidationRules: datatypes.JSON(rules),
		DefaultValue:    datatypes.JSON(def),
		I18nLabel:       datatypes.JSON(label),
		I18nHelp:        datatypes.JSON(help),
		Inheritable:     input.Inheritable,
	}
}

func mapTemplateDetail(tpl productcategory.CategoryTemplate, fields []TemplateFieldOutput) CategoryTemplateDetail {
	return CategoryTemplateDetail{
		ID:                        tpl.ID,
		Name:                      tpl.Name,
		Status:                    tpl.Status,
		ApplicableLevel:           tpl.ApplicableLevel,
		Notes:                     tpl.Notes,
		CurrentPublishedVersionID: tpl.CurrentPublishedVersionID,
		CurrentPublishedVersionNo: tpl.CurrentPublishedVersionNo,
		LastPublishedAt:           tpl.LastPublishedAt,
		Fields:                    fields,
		UpdatedAt:                 tpl.UpdatedAt,
		CreatedAt:                 tpl.CreatedAt,
	}
}

func mapTemplateFieldOutput(field productcategory.CategoryTemplateField) TemplateFieldOutput {
	var rules map[string]any
	_ = json.Unmarshal(field.ValidationRules, &rules)
	var def any
	_ = json.Unmarshal(field.DefaultValue, &def)
	var label map[string]any
	_ = json.Unmarshal(field.I18nLabel, &label)
	var help map[string]any
	_ = json.Unmarshal(field.I18nHelp, &help)
	return TemplateFieldOutput{
		ID:              field.ID,
		FieldKey:        field.FieldKey,
		Group:           field.Group,
		FieldType:       field.FieldType,
		Required:        field.Required,
		SortOrder:       field.SortOrder,
		ValidationRules: rules,
		DefaultValue:    def,
		I18nLabel:       label,
		I18nHelp:        help,
		Inheritable:     field.Inheritable,
	}
}

func parseCategoryPathIDs(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}

func isString(v any) bool {
	_, ok := v.(string)
	return ok
}

func matchFieldType(fieldType string, value any) bool {
	switch strings.ToLower(strings.TrimSpace(fieldType)) {
	case "string":
		_, ok := value.(string)
		return ok
	case "number", "int", "float":
		switch value.(type) {
		case int, int32, int64, float32, float64, json.Number:
			return true
		default:
			return false
		}
	case "boolean", "bool":
		_, ok := value.(bool)
		return ok
	case "json", "object":
		switch value.(type) {
		case map[string]any, []any:
			return true
		default:
			return false
		}
	case "enum":
		_, ok := value.(string)
		return ok
	case "date", "datetime":
		_, ok := value.(string)
		return ok
	default:
		return true
	}
}

func enumOptions(rules map[string]any) []string {
	if rules == nil {
		return nil
	}
	for _, key := range []string{"options", "enum", "values"} {
		if raw, ok := rules[key]; ok {
			switch v := raw.(type) {
			case []string:
				return v
			case []any:
				out := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
						out = append(out, strings.TrimSpace(s))
					}
				}
				return out
			}
		}
	}
	return nil
}

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, v := range values {
		if strings.EqualFold(strings.TrimSpace(v), target) {
			return true
		}
	}
	return false
}

func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such table"):
		return true
	case strings.Contains(msg, "does not exist") && strings.Contains(msg, "relation"):
		return true
	case strings.Contains(msg, "relation") && strings.Contains(msg, "does not exist"):
		return true
	default:
		return false
	}
}

package customer

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/google/uuid"
)

var (
	// ErrCustomerNotFound signals the requested record does not exist.
	ErrCustomerNotFound = errors.New("customer not found")
	// ErrCustomerConflict is returned when unique fields collide.
	ErrCustomerConflict = errors.New("customer conflict")
)

// ValidationError captures a single field level violation.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors aggregates validation failures into a single error.
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	parts := make([]string, len(v))
	for i, err := range v {
		if err.Field != "" {
			parts[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
			continue
		}
		parts[i] = err.Message
	}
	return strings.Join(parts, "; ")
}

func (v ValidationErrors) add(field, message string) ValidationErrors {
	return append(v, ValidationError{Field: field, Message: message})
}

func (v ValidationErrors) empty() bool {
	return len(v) == 0
}

// CreateCustomerInput enumerates fields required to create a customer.
type CreateCustomerInput struct {
	Name           string
	Type           string
	Email          string
	Phone          string
	Source         string
	Country        string
	Region         string
	MembershipTier string
	AccountManager string
	Tags           []string
	Notes          string
}

// UpdateCustomerInput enumerates fields that can be updated.
type UpdateCustomerInput struct {
	Name           *string
	Type           *string
	Email          *string
	Phone          *string
	Source         *string
	Country        *string
	Region         *string
	MembershipTier *string
	AccountManager *string
	Status         *string
	Tags           *[]string
	Notes          *string
}

// CreateCustomer persists a new customer record for the tenant.
func (s *Service) CreateCustomer(ctx context.Context, input CreateCustomerInput) (*Customer, error) {
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}
	tenantUUID := tenantFromContext(ctx)
	cleanTags := normalizeTags(input.Tags)
	if field, err := s.repo.DetectConflict(ctx, input.Email, input.Phone, ""); err != nil {
		return nil, err
	} else if field != "" {
		return nil, fmt.Errorf("%w:%s", ErrCustomerConflict, field)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	customer := Customer{
		ID:             uuid.NewString(),
		Name:           strings.TrimSpace(input.Name),
		Type:           strings.TrimSpace(input.Type),
		Email:          strings.TrimSpace(input.Email),
		Phone:          strings.TrimSpace(input.Phone),
		Country:        strings.TrimSpace(input.Country),
		Region:         strings.TrimSpace(input.Region),
		Source:         strings.TrimSpace(input.Source),
		MembershipTier: strings.TrimSpace(input.MembershipTier),
		AccountManager: strings.TrimSpace(input.AccountManager),
		Tags:           cleanTags,
		Notes:          strings.TrimSpace(input.Notes),
		Status:         "active",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if customer.Notes != "" {
		customer.Metadata = map[string]any{"notes": customer.Notes}
	}
	entity := &customermodel.Customer{}
	copyCustomerToEntity(entity, customer)
	created, err := s.repo.CreateCustomer(ctx, entity)
	if err != nil {
		return nil, err
	}
	customer.CreatedAt = created.CreatedAt.UTC().Format(time.RFC3339)
	customer.UpdatedAt = created.UpdatedAt.UTC().Format(time.RFC3339)
	s.emitChange(ctx, tenantUUID, "created", customer, "")
	return &customer, nil
}

// UpdateCustomer mutates an existing record using the provided patch.
func (s *Service) UpdateCustomer(ctx context.Context, id string, input UpdateCustomerInput) (*Customer, error) {
	if err := validateUpdateInput(input); err != nil {
		return nil, err
	}
	tenantUUID := tenantFromContext(ctx)
	entity, err := s.repo.FindByCustomerID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrCustomerNotFound
	}
	if field, err := s.repo.DetectConflict(ctx, deref(input.Email), deref(input.Phone), id); err != nil {
		return nil, err
	} else if field != "" {
		return nil, fmt.Errorf("%w:%s", ErrCustomerConflict, field)
	}
	current := convertEntityToCustomer(entity)
	patched := applyPatch(current, input)
	patched.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	copyCustomerToEntity(entity, patched)
	saved, err := s.repo.SaveCustomer(ctx, entity)
	if err != nil {
		return nil, err
	}
	result := convertEntityToCustomer(saved)
	s.emitChange(ctx, tenantUUID, "updated", result, "")
	return &result, nil
}

// DeleteCustomer removes a record for given tenant.
func (s *Service) DeleteCustomer(ctx context.Context, id, reason string) (*Customer, error) {
	reason = strings.TrimSpace(reason)
	var errs ValidationErrors
	if reason == "" {
		errs = errs.add("reason", "删除原因必填")
	}
	if !errs.empty() {
		return nil, errs
	}
	tenantUUID := tenantFromContext(ctx)
	entity, err := s.repo.FindByCustomerID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrCustomerNotFound
	}
	if err := s.repo.DeleteByCustomerID(ctx, id); err != nil {
		return nil, err
	}
	customer := convertEntityToCustomer(entity)
	s.emitChange(ctx, tenantUUID, "deleted", customer, reason)
	return &customer, nil
}

func validateCreateInput(input CreateCustomerInput) error {
	var errs ValidationErrors
	if strings.TrimSpace(input.Name) == "" {
		errs = errs.add("name", "客户姓名不能为空")
	}
	if strings.TrimSpace(input.Type) == "" {
		errs = errs.add("type", "客户类型不能为空")
	} else if !isAllowedType(input.Type) {
		errs = errs.add("type", "客户类型无效")
	}
	if strings.TrimSpace(input.Phone) == "" {
		errs = errs.add("phone", "联系方式必填")
	} else if !isValidPhone(input.Phone) {
		errs = errs.add("phone", "手机号格式无效")
	}
	if strings.TrimSpace(input.Email) != "" && !isValidEmail(input.Email) {
		errs = errs.add("email", "邮箱格式无效")
	}
	if strings.TrimSpace(input.MembershipTier) == "" {
		errs = errs.add("membershipTier", "会员等级不能为空")
	}
	if len(input.Tags) > 0 {
		if msg := validateTags(input.Tags); msg != "" {
			errs = errs.add("tags", msg)
		}
	}
	if utf8.RuneCountInString(strings.TrimSpace(input.Notes)) > 500 {
		errs = errs.add("notes", "备注长度不能超过 500 字")
	}
	if errs.empty() {
		return nil
	}
	return errs
}

func validateUpdateInput(input UpdateCustomerInput) error {
	if input.Name == nil && input.Type == nil && input.Email == nil && input.Phone == nil &&
		input.Source == nil && input.Country == nil && input.Region == nil &&
		input.MembershipTier == nil && input.AccountManager == nil && input.Status == nil &&
		input.Tags == nil && input.Notes == nil {
		return ValidationErrors{{Field: "payload", Message: "至少提交一个字段"}}
	}
	var errs ValidationErrors
	if input.Type != nil && !isAllowedType(*input.Type) {
		errs = errs.add("type", "客户类型无效")
	}
	if input.Phone != nil && *input.Phone != "" && !isValidPhone(*input.Phone) {
		errs = errs.add("phone", "手机号格式无效")
	}
	if input.Email != nil && *input.Email != "" && !isValidEmail(*input.Email) {
		errs = errs.add("email", "邮箱格式无效")
	}
	if input.Tags != nil {
		if msg := validateTags(*input.Tags); msg != "" {
			errs = errs.add("tags", msg)
		}
	}
	if input.Notes != nil && utf8.RuneCountInString(strings.TrimSpace(*input.Notes)) > 500 {
		errs = errs.add("notes", "备注长度不能超过 500 字")
	}
	if errs.empty() {
		return nil
	}
	return errs
}

func isAllowedType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "individual", "enterprise":
		return true
	}
	return false
}

var phonePattern = regexp.MustCompile(`^[+0-9()\s-]{6,}$`)

func isValidPhone(value string) bool {
	return phonePattern.MatchString(strings.TrimSpace(value))
}

func isValidEmail(value string) bool {
	_, err := mail.ParseAddress(strings.TrimSpace(value))
	return err == nil
}

func validateTags(tags []string) string {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			return "标签内容不能为空"
		}
		if utf8.RuneCountInString(trimmed) > 20 {
			return "单个标签最长 20 个字符"
		}
	}
	return ""
}

func applyPatch(current Customer, patch UpdateCustomerInput) Customer {
	if patch.Name != nil {
		current.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.Type != nil {
		current.Type = strings.TrimSpace(*patch.Type)
	}
	if patch.Email != nil {
		current.Email = strings.TrimSpace(*patch.Email)
	}
	if patch.Phone != nil {
		current.Phone = strings.TrimSpace(*patch.Phone)
	}
	if patch.Source != nil {
		current.Source = strings.TrimSpace(*patch.Source)
	}
	if patch.Country != nil {
		current.Country = strings.TrimSpace(*patch.Country)
	}
	if patch.Region != nil {
		current.Region = strings.TrimSpace(*patch.Region)
	}
	if patch.MembershipTier != nil {
		current.MembershipTier = strings.TrimSpace(*patch.MembershipTier)
	}
	if patch.AccountManager != nil {
		current.AccountManager = strings.TrimSpace(*patch.AccountManager)
	}
	if patch.Status != nil {
		current.Status = strings.TrimSpace(*patch.Status)
	}
	if patch.Tags != nil {
		current.Tags = normalizeTags(*patch.Tags)
	}
	if patch.Notes != nil {
		current.Notes = strings.TrimSpace(*patch.Notes)
		if current.Metadata == nil {
			current.Metadata = map[string]any{}
		}
		if current.Notes == "" {
			delete(current.Metadata, "notes")
		} else {
			current.Metadata["notes"] = current.Notes
		}
	}
	return current
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func tenantFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if tenant, ok := authx.TenantUUIDFromContext(ctx); ok {
		return strings.TrimSpace(tenant)
	}
	return ""
}

func (s *Service) emitChange(ctx context.Context, tenant, action string, customer Customer, reason string) {
	if s == nil || s.deps == nil {
		return
	}
	extra := logger.Fields{
		"tenant":      strings.TrimSpace(tenant),
		"action":      action,
		"customer_id": customer.ID,
	}
	if reason != "" {
		extra["reason"] = reason
	}
	entry := s.deps.RuntimeLogger(ctx, "customer.service", extra)
	entry.Infof("CustomerChanged: %s", action)
}

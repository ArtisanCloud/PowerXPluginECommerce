package payments

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrProviderServiceUnavailable = errors.New("payment provider service unavailable")
	ErrProviderNotFound           = errors.New("payment provider not found")
	ErrProviderNameRequired       = errors.New("provider name is required")
	ErrProviderSelectorDuplicate  = errors.New("payment provider selector already exists")
)

type ProviderService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentProviderRepository
}

func NewProviderService(deps *app.Deps) *ProviderService {
	if deps == nil || deps.DB == nil {
		return &ProviderService{deps: deps}
	}
	return &ProviderService{deps: deps, repo: paymentrepo.NewPaymentProviderRepository(deps.DB)}
}

func (s *ProviderService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *ProviderService) ListProviders(ctx context.Context, tenantUUID, adminID string, filter ProviderListFilter) ([]ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	query := s.deps.DB.WithContext(ctx).Model(&models.PaymentProvider{}).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := strings.TrimSpace(filter.ProviderType); v != "" {
		query = query.Where("provider_type = ?", v)
	}
	if v := strings.TrimSpace(filter.Keyword); v != "" {
		query = query.Where("name ILIKE ?", "%"+v+"%")
	}
	var rows []models.PaymentProvider
	if err := query.Order("updated_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ProviderDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProviderDTO{
			ID:              row.ID,
			Name:            row.Name,
			ProviderType:    row.ProviderType,
			Status:          row.Status,
			IsDefault:       row.IsDefault,
			FeeRate:         row.FeeRate,
			SettlementCycle: row.SettlementCycle,
			Currency:        row.Currency,
			AppID:           row.AppID,
			MchID:           row.MchID,
			SerialNo:        row.SerialNo,
			NotifyURL:       row.NotifyURL,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (s *ProviderService) CreateProvider(ctx context.Context, tenantUUID, adminID string, req CreateProviderRequest) (*ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrProviderNameRequired
	}
	credRaw := datatypes.JSON([]byte(`{}`))
	riskRaw, _ := json.Marshal(req.RiskPolicy)
	appID, mchID, serialNo, notifyURL := extractProviderIdentifiers(req.Credentials)
	row := &models.PaymentProvider{
		BaseModel:       models.BaseModel{TenantUuid: strings.TrimSpace(tenantUUID)},
		Name:            strings.TrimSpace(req.Name),
		ProviderType:    strings.TrimSpace(req.Type),
		AppID:           appID,
		MchID:           mchID,
		SerialNo:        serialNo,
		NotifyURL:       notifyURL,
		Status:          strings.TrimSpace(req.Status),
		IsDefault:       req.IsDefault,
		FeeRate:         req.FeeRate,
		SettlementCycle: strings.TrimSpace(req.SettlementCycle),
		Currency:        strings.TrimSpace(req.Currency),
		Credentials:     credRaw,
		RiskPolicy:      datatypes.JSON(riskRaw),
	}
	var created *models.PaymentProvider
	err := s.repo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		if err := s.ensureProviderSelectorUnique(ctx, tx, tenantUUID, row.ProviderType, row.MchID, row.AppID, 0); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
		if req.Credentials != nil {
			if strings.EqualFold(row.ProviderType, "wechat") {
				if err := validateWechatCredentialInput(req.Credentials); err != nil {
					return err
				}
			}
			encrypted, err := EncryptProviderCredentials(
				s.deps.Config,
				tenantUUID,
				row.ID,
				stripProviderPublicCredentials(req.Credentials),
			)
			if err != nil {
				return err
			}
			if err := tx.WithContext(ctx).
				Model(&models.PaymentProvider{}).
				Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).
				Update("credentials", encrypted).Error; err != nil {
				return err
			}
			row.Credentials = encrypted
		}
		if row.IsDefault {
			if err := s.clearOtherDefaults(ctx, tx, tenantUUID, row.ProviderType, row.ID); err != nil {
				return err
			}
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ProviderDTO{
		ID:              created.ID,
		Name:            created.Name,
		ProviderType:    created.ProviderType,
		Status:          created.Status,
		IsDefault:       created.IsDefault,
		FeeRate:         created.FeeRate,
		SettlementCycle: created.SettlementCycle,
		Currency:        created.Currency,
		AppID:           created.AppID,
		MchID:           created.MchID,
		SerialNo:        created.SerialNo,
		NotifyURL:       created.NotifyURL,
		UpdatedAt:       created.UpdatedAt,
	}, nil
}

func (s *ProviderService) UpdateProvider(ctx context.Context, tenantUUID, adminID string, id uint64, req UpdateProviderRequest) (*ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if req.Credentials != nil {
		logger.WithFields(logger.Fields{
			"component":  "payments.providers",
			"providerId": id,
			"stage":      "incoming",
		}).Info("provider credentials received", credentialSummary(req.Credentials))
	}
	var updated *models.PaymentProvider
	err := s.repo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var row models.PaymentProvider
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProviderNotFound
			}
			return err
		}
		if req.Status != nil {
			row.Status = strings.TrimSpace(*req.Status)
		}
		if req.IsDefault != nil {
			row.IsDefault = *req.IsDefault
		}
		if req.FeeRate != nil {
			row.FeeRate = *req.FeeRate
		}
		if req.SettlementCycle != nil {
			row.SettlementCycle = strings.TrimSpace(*req.SettlementCycle)
		}
		if req.Currency != nil {
			row.Currency = strings.TrimSpace(*req.Currency)
		}
		if req.Credentials != nil {
			merged, err := s.mergeProviderCredentials(ctx, tenantUUID, row.ID, row.Credentials, req.Credentials)
			if err != nil {
				return err
			}
			if strings.EqualFold(row.ProviderType, "wechat") {
				if err := validateWechatCredentialInput(merged); err != nil {
					return err
				}
			}
			logger.WithFields(logger.Fields{
				"component":  "payments.providers",
				"providerId": row.ID,
				"stage":      "merged",
			}).Info("provider credentials merged", credentialSummary(merged))
			appID, mchID, serialNo, notifyURL := extractProviderIdentifiers(merged)
			if appID != "" {
				row.AppID = appID
			}
			if mchID != "" {
				row.MchID = mchID
			}
			if serialNo != "" {
				row.SerialNo = serialNo
			}
			if notifyURL != "" {
				row.NotifyURL = notifyURL
			}
			if err := s.ensureProviderSelectorUnique(ctx, tx, tenantUUID, row.ProviderType, row.MchID, row.AppID, row.ID); err != nil {
				return err
			}
			encrypted, err := EncryptProviderCredentials(
				s.deps.Config,
				tenantUUID,
				row.ID,
				stripProviderPublicCredentials(merged),
			)
			if err != nil {
				return err
			}
			row.Credentials = encrypted
		}
		if req.RiskPolicy != nil {
			riskRaw, _ := json.Marshal(req.RiskPolicy)
			row.RiskPolicy = datatypes.JSON(riskRaw)
		}
		if err := tx.WithContext(ctx).Save(&row).Error; err != nil {
			return err
		}
		if row.IsDefault {
			if err := s.clearOtherDefaults(ctx, tx, tenantUUID, row.ProviderType, row.ID); err != nil {
				return err
			}
		}
		updated = &row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ProviderDTO{
		ID:              updated.ID,
		Name:            updated.Name,
		ProviderType:    updated.ProviderType,
		Status:          updated.Status,
		IsDefault:       updated.IsDefault,
		FeeRate:         updated.FeeRate,
		SettlementCycle: updated.SettlementCycle,
		Currency:        updated.Currency,
		AppID:           updated.AppID,
		MchID:           updated.MchID,
		SerialNo:        updated.SerialNo,
		NotifyURL:       updated.NotifyURL,
		UpdatedAt:       updated.UpdatedAt,
	}, nil
}

func (s *ProviderService) ensureProviderSelectorUnique(ctx context.Context, tx *gorm.DB, tenantUUID, providerType, mchID, appID string, excludeID uint64) error {
	if tx == nil {
		return nil
	}
	if strings.TrimSpace(tenantUUID) == "" {
		return nil
	}
	if strings.TrimSpace(providerType) == "" || strings.TrimSpace(mchID) == "" || strings.TrimSpace(appID) == "" {
		return nil
	}
	query := tx.WithContext(ctx).
		Model(&models.PaymentProvider{}).
		Where("tenant_uuid = ? AND provider_type = ? AND mch_id = ? AND app_id = ? AND deleted_at IS NULL",
			tenantUUID, providerType, mchID, appID)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrProviderSelectorDuplicate
	}
	return nil
}

func extractProviderIdentifiers(creds map[string]interface{}) (string, string, string, string) {
	if creds == nil {
		return "", "", "", ""
	}
	appID := pickString(creds, "appId", "appid", "app_id")
	mchID := pickString(creds, "mchId", "mchid", "mch_id")
	serialNo := pickString(creds, "serialNo", "serial_no", "serial")
	notifyURL := pickString(creds, "notifyUrl", "notify_url", "notifyURL")
	return appID, mchID, serialNo, notifyURL
}

func pickString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			}
		}
	}
	return ""
}

func pickBool(payload map[string]interface{}, keys ...string) (bool, bool) {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case bool:
				return t, true
			case string:
				if strings.EqualFold(strings.TrimSpace(t), "true") {
					return true, true
				}
				if strings.EqualFold(strings.TrimSpace(t), "false") {
					return false, true
				}
			case float64:
				return t != 0, true
			case int:
				return t != 0, true
			default:
				if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
					if strings.EqualFold(s, "true") {
						return true, true
					}
					if strings.EqualFold(s, "false") {
						return false, true
					}
				}
			}
		}
	}
	return false, false
}

func normalizeSerial(serial string) string {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return ""
	}
	serial = strings.ToUpper(serial)
	serial = strings.ReplaceAll(serial, ":", "")
	serial = strings.TrimLeft(serial, "0")
	return serial
}

func validateWechatCredentialInput(creds map[string]interface{}) error {
	privateKey := pickString(creds, "privateKeyPem", "private_key_pem")
	cert := pickString(creds, "certPem", "cert_pem")
	apiV3Key := strings.TrimSpace(pickString(creds, "apiV3Key", "api_v3_key", "apiv3key"))
	if strings.Contains(privateKey, "BEGIN CERTIFICATE") {
		return errors.New("商户私钥内容疑似为证书，请上传 apiclient_key.pem")
	}
	if strings.Contains(cert, "BEGIN PRIVATE KEY") {
		return errors.New("商户证书内容疑似为私钥，请上传 apiclient_cert.pem")
	}
	if apiV3Key != "" && len(apiV3Key) != 32 {
		return errors.New("API v3 Key 长度应为 32 位")
	}
	return nil
}

func credentialSummary(creds map[string]interface{}) logger.Fields {
	privateKey := pickString(creds, "privateKeyPem", "private_key_pem")
	cert := pickString(creds, "certPem", "cert_pem")
	return logger.Fields{
		"privateKeyHeader": pemHeader(privateKey),
		"privateKeyLen":    len(strings.TrimSpace(privateKey)),
		"certHeader":       pemHeader(cert),
		"certLen":          len(strings.TrimSpace(cert)),
	}
}

func pemHeader(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}

func stripProviderPublicCredentials(creds map[string]interface{}) map[string]interface{} {
	if creds == nil {
		return nil
	}
	out := map[string]interface{}{}
	for key, value := range creds {
		switch strings.ToLower(key) {
		case "appid", "app_id", "mchid", "mch_id", "serialno", "serial_no", "notifyurl", "notify_url",
			"privatekeypath", "private_key_path", "certpath", "cert_path":
			continue
		default:
			out[key] = value
		}
	}
	return out
}

func (s *ProviderService) mergeProviderCredentials(ctx context.Context, tenantUUID string, providerID uint64, existingRaw datatypes.JSON, incoming map[string]interface{}) (map[string]interface{}, error) {
	if incoming == nil {
		return nil, nil
	}
	merged := map[string]interface{}{}
	for key, value := range incoming {
		merged[key] = value
	}
	existing, err := DecryptProviderCredentials(s.deps.Config, tenantUUID, providerID, existingRaw)
	if err != nil {
		return nil, err
	}
	for key, value := range existing {
		if _, ok := merged[key]; ok {
			continue
		}
		merged[key] = value
	}
	if shouldPreservePrivateKey(merged) {
		if preserved := pickCredential(existing, "privateKeyPem", "private_key_pem"); preserved != "" {
			merged["privateKeyPem"] = preserved
		}
	}
	return merged, nil
}

func shouldPreservePrivateKey(merged map[string]interface{}) bool {
	if merged == nil {
		return false
	}
	if v, ok := merged["privateKeyPem"]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return false
		}
	}
	if v, ok := merged["private_key_pem"]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return false
		}
	}
	return true
}

func pickCredential(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			default:
				if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

func (s *ProviderService) TestProvider(ctx context.Context, tenantUUID, adminID string, id uint64) error {
	if !s.Ready() {
		return ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return errors.New("admin id is required")
	}
	var row models.PaymentProvider
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProviderNotFound
		}
		return err
	}
	return nil
}

func (s *ProviderService) TestMiniApp(ctx context.Context, tenantUUID, adminID string, id uint64, appID, appSecret string) error {
	if !s.Ready() {
		return ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return errors.New("admin id is required")
	}
	var row models.PaymentProvider
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProviderNotFound
		}
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(row.ProviderType), "wechat") {
		return errors.New("仅支持微信渠道检测")
	}
	creds, err := DecryptProviderCredentials(s.deps.Config, tenantUUID, row.ID, row.Credentials)
	if err != nil {
		return err
	}
	appID = strings.TrimSpace(appID)
	if appID == "" {
		appID = strings.TrimSpace(row.AppID)
	}
	if appID == "" {
		appID = pickString(creds, "appId", "appid", "app_id")
	}
	appSecret = strings.TrimSpace(appSecret)
	if appSecret == "" || appSecret == "*" {
		appSecret = pickString(creds, "appSecret", "app_secret", "secret")
	}
	if appID == "" || appSecret == "" {
		return errors.New("AppID 或 AppSecret 未配置")
	}
	httpDebug, hasHttpDebug := pickBool(creds, "httpDebug", "http_debug")
	debug, hasDebug := pickBool(creds, "debug")
	if s.deps.Config != nil && s.deps.Config.Server != nil && s.deps.Config.Server.DevMode {
		if !hasHttpDebug {
			httpDebug = true
		}
		if !hasDebug {
			debug = true
		}
	}
	app, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:     appID,
		Secret:    appSecret,
		HttpDebug: httpDebug,
		Debug:     debug,
		Log: miniProgram.Log{
			Level:  "debug",
			ENV:    "develop",
			Stdout: true,
		},
	})
	if err != nil {
		return err
	}
	tokenResp, err := app.GetAccessToken().GetToken(ctx, true)
	if err != nil {
		return err
	}
	if tokenResp != nil && tokenResp.IsError() {
		return errors.New(tokenResp.Error())
	}
	if tokenResp == nil || strings.TrimSpace(tokenResp.AccessToken) == "" {
		return errors.New("获取 access_token 失败")
	}
	return nil
}

func (s *ProviderService) TestWechatCertSerial(ctx context.Context, tenantUUID, adminID string, id uint64, serialNo, certPem string) error {
	if !s.Ready() {
		return ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return errors.New("admin id is required")
	}
	var row models.PaymentProvider
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProviderNotFound
		}
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(row.ProviderType), "wechat") {
		return errors.New("仅支持微信渠道检测")
	}
	creds, err := DecryptProviderCredentials(s.deps.Config, tenantUUID, row.ID, row.Credentials)
	if err != nil {
		return err
	}
	serialNo = strings.TrimSpace(serialNo)
	if serialNo == "" {
		serialNo = strings.TrimSpace(row.SerialNo)
	}
	if serialNo == "" {
		serialNo = pickString(creds, "serialNo", "serial_no", "serial")
	}
	certPem = strings.TrimSpace(certPem)
	if certPem == "" || certPem == "*" {
		certPem = pickString(creds, "certPem", "cert_pem")
	}
	if serialNo == "" || certPem == "" {
		return errors.New("证书序列号或证书内容未配置")
	}
	block, _ := pem.Decode([]byte(certPem))
	if block == nil {
		return errors.New("证书内容解析失败")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	expected := normalizeSerial(serialNo)
	actual := normalizeSerial(fmt.Sprintf("%X", cert.SerialNumber))
	if expected == "" || actual == "" {
		return errors.New("证书序列号解析失败")
	}
	if expected != actual {
		return fmt.Errorf("证书序列号不匹配，实际为 %s", actual)
	}
	return nil
}

func (s *ProviderService) GetProviderDetail(ctx context.Context, tenantUUID, adminID string, id uint64) (*ProviderDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var row models.PaymentProvider
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProviderNotFound
		}
		return nil, err
	}
	creds, err := DecryptProviderCredentials(s.deps.Config, tenantUUID, row.ID, row.Credentials)
	if err != nil {
		return nil, err
	}
	creds = maskProviderCredentials(creds)
	var risk map[string]interface{}
	if len(row.RiskPolicy) > 0 {
		_ = json.Unmarshal(row.RiskPolicy, &risk)
	}
	return &ProviderDetailDTO{
		ID:              row.ID,
		Name:            row.Name,
		ProviderType:    row.ProviderType,
		Status:          row.Status,
		IsDefault:       row.IsDefault,
		FeeRate:         row.FeeRate,
		SettlementCycle: row.SettlementCycle,
		Currency:        row.Currency,
		AppID:           row.AppID,
		MchID:           row.MchID,
		SerialNo:        row.SerialNo,
		NotifyURL:       row.NotifyURL,
		Credentials:     creds,
		RiskPolicy:      risk,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func maskProviderCredentials(creds map[string]interface{}) map[string]interface{} {
	if creds == nil {
		return nil
	}
	masked := make(map[string]interface{}, len(creds))
	for k, v := range creds {
		masked[k] = v
	}
	for key := range masked {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "apiv3key", "api_v3_key", "privatekeypem", "private_key_pem", "certpem", "cert_pem",
			"appsecret", "app_secret", "secret":
			masked[key] = "*"
		}
	}
	return masked
}

func (s *ProviderService) clearOtherDefaults(ctx context.Context, tx *gorm.DB, tenantUUID, providerType string, currentID uint64) error {
	if tx == nil {
		return errors.New("db transaction required")
	}
	return tx.WithContext(ctx).
		Model(&models.PaymentProvider{}).
		Where("tenant_uuid = ? AND provider_type = ? AND id <> ?", tenantUUID, providerType, currentID).
		Update("is_default", false).Error
}

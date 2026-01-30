package payments

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/payment"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	pxmodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	adminpayments "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/payments"
)

type wechatCredentials struct {
	AppID          string
	MchID          string
	SerialNo       string
	ApiV3Key       string
	PrivateKeyPem  string
	PrivateKeyPath string
	CertPem        string
	CertPath       string
	NotifyURL      string
}

type wechatAppCacheEntry struct {
	app         *payment.Payment
	cleanup     func()
	fingerprint string
	updatedAt   time.Time
}

var wechatAppCache = struct {
	mu    sync.Mutex
	items map[string]*wechatAppCacheEntry
}{
	items: map[string]*wechatAppCacheEntry{},
}

const wechatAppCacheTTL = time.Hour

func parseWechatCredentials(provider *pxmodels.PaymentProvider, payload map[string]any) (wechatCredentials, error) {
	appID := pickCredential(payload, "appId", "appid", "app_id")
	if appID == "" && provider != nil {
		appID = strings.TrimSpace(provider.AppID)
	}
	mchID := pickCredential(payload, "mchId", "mchid", "mch_id")
	if mchID == "" && provider != nil {
		mchID = strings.TrimSpace(provider.MchID)
	}
	serialNo := pickCredential(payload, "serialNo", "serial_no")
	if serialNo == "" && provider != nil {
		serialNo = strings.TrimSpace(provider.SerialNo)
	}
	notifyURL := pickCredential(payload, "notifyUrl", "notify_url", "notifyURL")
	if notifyURL == "" && provider != nil {
		notifyURL = strings.TrimSpace(provider.NotifyURL)
	}
	creds := wechatCredentials{
		AppID:          appID,
		MchID:          mchID,
		SerialNo:       serialNo,
		ApiV3Key:       pickCredential(payload, "apiV3Key", "api_v3_key", "apiv3key"),
		PrivateKeyPem:  pickCredential(payload, "privateKeyPem", "private_key_pem", "private_key"),
		PrivateKeyPath: pickCredential(payload, "privateKeyPath", "private_key_path"),
		CertPem:        pickCredential(payload, "certPem", "cert_pem"),
		CertPath:       pickCredential(payload, "certPath", "cert_path"),
		NotifyURL:      notifyURL,
	}
	if creds.AppID == "" || creds.MchID == "" || creds.SerialNo == "" || creds.ApiV3Key == "" {
		return wechatCredentials{}, ErrWechatCredentialsMissingRequiredFields
	}
	if creds.PrivateKeyPem == "" && creds.PrivateKeyPath == "" {
		return wechatCredentials{}, ErrWechatCredentialsMissingPrivateKey
	}
	if creds.NotifyURL == "" {
		return wechatCredentials{}, ErrWechatCredentialsMissingNotifyURL
	}
	return creds, nil
}

func getWechatPaymentAppCachedForProvider(cfg *config.Config, provider *pxmodels.PaymentProvider) (*payment.Payment, func(), error) {
	if provider == nil {
		return nil, nil, ErrProviderUnavailable
	}
	key := fmt.Sprintf("%s:%d", strings.TrimSpace(provider.TenantUuid), provider.ID)
	fingerprint := wechatProviderFingerprint(provider)
	now := time.Now()
	httpDebug := cfg != nil && cfg.Logging != nil && cfg.Logging.WechatHTTPDebug
	wechatAppCache.mu.Lock()
	if entry := wechatAppCache.items[key]; entry != nil && entry.app != nil && entry.fingerprint == fingerprint {
		if now.Sub(entry.updatedAt) < wechatAppCacheTTL {
			app := entry.app
			wechatAppCache.mu.Unlock()
			return app, func() {}, nil
		}
	}
	wechatAppCache.mu.Unlock()

	credsRaw, err := adminpayments.DecryptProviderCredentials(cfg, provider.TenantUuid, provider.ID, provider.Credentials)
	if err != nil {
		return nil, nil, err
	}
	creds, err := parseWechatCredentials(provider, credsRaw)
	if err != nil {
		return nil, nil, err
	}
	app, cleanup, err := buildWechatPaymentApp(creds, httpDebug)
	if err != nil {
		return nil, nil, err
	}

	wechatAppCache.mu.Lock()
	if entry := wechatAppCache.items[key]; entry != nil && entry.cleanup != nil {
		entry.cleanup()
	}
	wechatAppCache.items[key] = &wechatAppCacheEntry{
		app:         app,
		cleanup:     cleanup,
		fingerprint: fingerprint,
		updatedAt:   now,
	}
	wechatAppCache.mu.Unlock()
	return app, func() {}, nil
}

func buildWechatPaymentApp(creds wechatCredentials, httpDebug bool) (*payment.Payment, func(), error) {
	keyPath, keyCleanup, err := ensurePemFile(creds.PrivateKeyPath, creds.PrivateKeyPem, "wxpay-key-*.pem")
	if err != nil {
		return nil, nil, err
	}
	certPath, certCleanup, err := ensurePemFile(creds.CertPath, creds.CertPem, "wxpay-cert-*.pem")
	if err != nil {
		keyCleanup()
		return nil, nil, err
	}
	cfg := &payment.UserConfig{
		AppID:       creds.AppID,
		MchID:       creds.MchID,
		MchApiV3Key: creds.ApiV3Key,
		CertPath:    certPath,
		KeyPath:     keyPath,
		SerialNo:    creds.SerialNo,
		NotifyURL:   creds.NotifyURL,
		HttpDebug:   httpDebug,
	}
	app, err := payment.NewPayment(cfg)
	if err != nil {
		keyCleanup()
		certCleanup()
		return nil, nil, err
	}
	cleanup := func() {
		keyCleanup()
		certCleanup()
	}
	return app, cleanup, nil
}

func wechatProviderFingerprint(provider *pxmodels.PaymentProvider) string {
	sum := sha256.New()
	parts := []string{
		strings.TrimSpace(provider.AppID),
		strings.TrimSpace(provider.MchID),
		strings.TrimSpace(provider.SerialNo),
		strings.TrimSpace(provider.NotifyURL),
		string(provider.Credentials),
	}
	for _, part := range parts {
		if part == "" {
			part = "-"
		}
		sum.Write([]byte(part))
		sum.Write([]byte{0})
	}
	return hex.EncodeToString(sum.Sum(nil))
}

func ensurePemFile(pathValue, pemValue, prefix string) (string, func(), error) {
	if strings.TrimSpace(pathValue) != "" {
		return strings.TrimSpace(pathValue), func() {}, nil
	}
	if strings.TrimSpace(pemValue) == "" {
		return "", func() {}, nil
	}
	tmp, err := os.CreateTemp(os.TempDir(), prefix)
	if err != nil {
		return "", func() {}, err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", func() {}, err
	}
	normalized := strings.ReplaceAll(pemValue, "\r\n", "\n")
	if _, err := tmp.WriteString(normalized); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", func() {}, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", func() {}, err
	}
	path := filepath.Clean(tmp.Name())
	cleanup := func() { _ = os.Remove(path) }
	return path, cleanup, nil
}

func pickCredential(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			default:
				s := strings.TrimSpace(fmt.Sprintf("%v", v))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

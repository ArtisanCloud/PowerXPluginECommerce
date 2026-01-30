package payments

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/crypto"
	"gorm.io/datatypes"
)

const providerCredentialsVersion = 1

type encryptedCredentials struct {
	Encrypted  bool   `json:"encrypted"`
	Version    int    `json:"v"`
	Ciphertext string `json:"ct"`
	Nonce      string `json:"iv"`
}

// EncryptProviderCredentials encrypts provider credentials using tenant+provider AAD.
func EncryptProviderCredentials(cfg *config.Config, tenantUUID string, providerID uint64, credentials map[string]interface{}) (datatypes.JSON, error) {
	secret, err := providerSecret(cfg)
	if err != nil {
		return nil, err
	}
	plain, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}
	key := crypto.DeriveKey32(secret)
	aad := []byte(fmt.Sprintf("tenant:%s|provider:%d", strings.TrimSpace(tenantUUID), providerID))
	ciphertext, nonce, err := crypto.EncryptAESGCM(key, plain, aad)
	if err != nil {
		return nil, err
	}
	payload := encryptedCredentials{
		Encrypted:  true,
		Version:    providerCredentialsVersion,
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(buf), nil
}

// DecryptProviderCredentials decrypts stored provider credentials.
func DecryptProviderCredentials(cfg *config.Config, tenantUUID string, providerID uint64, raw datatypes.JSON) (map[string]interface{}, error) {
	if len(raw) == 0 {
		return map[string]interface{}{}, nil
	}
	var encrypted encryptedCredentials
	if err := json.Unmarshal(raw, &encrypted); err == nil && encrypted.Encrypted {
		secret, err := providerSecret(cfg)
		if err != nil {
			return nil, err
		}
		ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encrypted.Ciphertext))
		if err != nil {
			return nil, err
		}
		nonce, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encrypted.Nonce))
		if err != nil {
			return nil, err
		}
		key := crypto.DeriveKey32(secret)
		aad := []byte(fmt.Sprintf("tenant:%s|provider:%d", strings.TrimSpace(tenantUUID), providerID))
		plain, err := crypto.DecryptAESGCM(key, ciphertext, nonce, aad)
		if err != nil {
			return nil, err
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(plain, &payload); err != nil {
			return nil, err
		}
		return payload, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func providerSecret(cfg *config.Config) (string, error) {
	if cfg == nil || cfg.Server == nil {
		return "", errors.New("server config not ready")
	}
	secret := strings.TrimSpace(cfg.Server.SecretKey)
	if secret != "" {
		return secret, nil
	}
	if cfg.IsProduction() {
		return "", errors.New("server.secret_key not configured")
	}
	logger.Warn("server.secret_key is empty; using DEV-ONLY fallback key. Do NOT use in production.")
	return "dev-only-change-me", nil
}

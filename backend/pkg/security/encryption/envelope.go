package encryption

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/crypto"
)

// Envelope stores ciphertext plus encrypted DEK metadata.
type Envelope struct {
	Ciphertext    []byte
	CipherNonce   []byte
	DekCiphertext []byte
	DekNonce      []byte
	Algorithm     string
	KeyVersion    string
	CreatedAt     time.Time
}

// KeyProvider supplies root key material to wrap/unwrap DEK.
type KeyProvider interface {
	Material(ctx context.Context) ([]byte, error)
	Version() string
}

// StaticKeyProvider derives key material from a static secret (e.g. config).
type StaticKeyProvider struct {
	secret  string
	version string
}

// NewStaticKeyProvider builds a provider backed by static secret material.
func NewStaticKeyProvider(secret string) *StaticKeyProvider {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		trimmed = "dev-only-change-me"
	}
	return &StaticKeyProvider{
		secret:  trimmed,
		version: "static-v1",
	}
}

// Material returns the derived 32-byte key.
func (p *StaticKeyProvider) Material(ctx context.Context) ([]byte, error) {
	if p == nil || strings.TrimSpace(p.secret) == "" {
		return nil, errors.New("static key provider not configured")
	}
	return crypto.DeriveKey32(p.secret), nil
}

// Version reports the logical key version (used for rotation metadata).
func (p *StaticKeyProvider) Version() string {
	if p == nil || p.version == "" {
		return "static-v1"
	}
	return p.version
}

// EncryptEnvelope performs envelope encryption and returns persisted fields.
func EncryptEnvelope(ctx context.Context, provider KeyProvider, plaintext []byte, aad []byte) (*Envelope, error) {
	if provider == nil {
		return nil, errors.New("encryption key provider missing")
	}
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext cannot be empty")
	}
	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("generate DEK: %w", err)
	}
	payloadCipher, payloadNonce, err := crypto.EncryptAESGCM(dek, plaintext, aad)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}
	rootKey, err := provider.Material(ctx)
	if err != nil {
		return nil, err
	}
	dekCipher, dekNonce, err := crypto.EncryptAESGCM(rootKey, dek, aad)
	if err != nil {
		return nil, fmt.Errorf("wrap DEK: %w", err)
	}
	return &Envelope{
		Ciphertext:    payloadCipher,
		CipherNonce:   payloadNonce,
		DekCiphertext: dekCipher,
		DekNonce:      dekNonce,
		Algorithm:     "AES-GCM",
		KeyVersion:    provider.Version(),
		CreatedAt:     time.Now().UTC(),
	}, nil
}

// DecryptEnvelope unwraps the stored DEK then decrypts payload.
func DecryptEnvelope(ctx context.Context, provider KeyProvider, envelope *Envelope, aad []byte) ([]byte, error) {
	if provider == nil {
		return nil, errors.New("encryption key provider missing")
	}
	if envelope == nil {
		return nil, errors.New("envelope cannot be nil")
	}
	rootKey, err := provider.Material(ctx)
	if err != nil {
		return nil, err
	}
	dek, err := crypto.DecryptAESGCM(rootKey, envelope.DekCiphertext, envelope.DekNonce, aad)
	if err != nil {
		return nil, fmt.Errorf("unwrap DEK: %w", err)
	}
	plaintext, err := crypto.DecryptAESGCM(dek, envelope.Ciphertext, envelope.CipherNonce, aad)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}
	return plaintext, nil
}

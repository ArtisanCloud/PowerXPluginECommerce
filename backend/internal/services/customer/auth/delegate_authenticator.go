package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DelegateAuthenticator 通过宿主接口校验 token。
type DelegateAuthenticator struct {
	endpoint     string
	serviceToken string
	httpClient   *http.Client
	cacheTTL     time.Duration
	mu           sync.RWMutex
	cache        map[string]cacheEntry
}

type cacheEntry struct {
	ctx       *Context
	expiresAt time.Time
}

// DelegateOptions 自定义代理行为。
type DelegateOptions struct {
	HTTPClient   *http.Client
	Endpoint     string
	ServiceToken string
	CacheTTL     time.Duration
}

// NewDelegateAuthenticator 构造宿主鉴权器。
func NewDelegateAuthenticator(opts DelegateOptions) (*DelegateAuthenticator, error) {
	endpoint := strings.TrimSpace(opts.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("customer auth: delegate endpoint required")
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	cacheTTL := opts.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 30 * time.Second
	}
	return &DelegateAuthenticator{
		endpoint:     endpoint,
		serviceToken: strings.TrimSpace(opts.ServiceToken),
		httpClient:   client,
		cacheTTL:     cacheTTL,
		cache:        make(map[string]cacheEntry),
	}, nil
}

// Authenticate 调用宿主校验。
func (d *DelegateAuthenticator) Authenticate(ctx context.Context, token string) (*Context, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrTokenRequired
	}
	if cached := d.lookupCache(token); cached != nil {
		return cached, nil
	}
	payload := map[string]string{"token": token}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.serviceToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.serviceToken))
		req.Header.Set("X-PowerX-Service-Token", d.serviceToken)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, ErrUnauthorized
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnauthorized
	}
	var envelope struct {
		Data delegatePayload `json:"data"`
		delegatePayload
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	payloadData := envelope.Data
	if payloadData.TenantUUID == "" && envelope.delegatePayload.TenantUUID != "" {
		payloadData = envelope.delegatePayload
	}
	if payloadData.TenantUUID == "" || payloadData.CustomerID == "" {
		return nil, ErrUnauthorized
	}
	result := &Context{
		TenantUUID: payloadData.TenantUUID,
		CustomerID: payloadData.CustomerID,
		Roles:      payloadData.Roles,
		ExpiresAt:  time.Unix(payloadData.ExpiresAt, 0),
		IssuedAt:   time.Unix(payloadData.IssuedAt, 0),
		RawToken:   token,
	}
	d.storeCache(token, result)
	return result, nil
}

type delegatePayload struct {
	TenantUUID string   `json:"tenant_uuid"`
	CustomerID string   `json:"customer_uuid"`
	Roles      []string `json:"roles"`
	ExpiresAt  int64    `json:"exp"`
	IssuedAt   int64    `json:"iat"`
}

func (d *DelegateAuthenticator) lookupCache(token string) *Context {
	if d.cacheTTL <= 0 {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if entry, ok := d.cache[token]; ok {
		if time.Now().Before(entry.expiresAt) {
			return entry.ctx
		}
	}
	return nil
}

func (d *DelegateAuthenticator) storeCache(token string, ctx *Context) {
	if d.cacheTTL <= 0 || ctx == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache[token] = cacheEntry{
		ctx:       ctx,
		expiresAt: time.Now().Add(d.cacheTTL),
	}
}

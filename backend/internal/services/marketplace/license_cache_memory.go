package marketplace

import (
	"context"
	"fmt"
	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
	"strings"
	"sync"
	"time"

	dbm "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/marketplace"
	"github.com/sirupsen/logrus"
)

// MemoryLicenseCache stores license snapshots in-memory for standalone mode.
type MemoryLicenseCache struct {
	mu      sync.RWMutex
	entries map[string]memoryLicenseCacheEntry
	logger  *logrus.Entry
}

type memoryLicenseCacheEntry struct {
	license   *dbm.License
	expiresAt time.Time
}

func NewMemoryLicenseCache(logger *logrus.Entry) *MemoryLicenseCache {
	if logger == nil {
		logger = pxlogger.WithField("component", "marketplace.license_cache.memory")
	}
	return &MemoryLicenseCache{
		entries: map[string]memoryLicenseCacheEntry{},
		logger:  logger,
	}
}

func (c *MemoryLicenseCache) Get(_ context.Context, tenantID, listingID string) (*dbm.License, bool) {
	if c == nil {
		return nil, false
	}
	key := c.cacheKey(tenantID, listingID)
	if key == "" {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || entry.license == nil {
		return nil, false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, false
	}
	return cloneLicense(entry.license), true
}

func (c *MemoryLicenseCache) Set(_ context.Context, tenantID, listingID string, license *dbm.License, ttl time.Duration) {
	if c == nil || license == nil {
		return
	}
	key := c.cacheKey(tenantID, listingID)
	if key == "" {
		return
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	c.mu.Lock()
	c.entries[key] = memoryLicenseCacheEntry{
		license:   cloneLicense(license),
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

func (c *MemoryLicenseCache) Delete(_ context.Context, tenantID, listingID string) {
	if c == nil {
		return
	}
	key := c.cacheKey(tenantID, listingID)
	if key == "" {
		return
	}
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

func (c *MemoryLicenseCache) cacheKey(tenantID, listingID string) string {
	tenantID = strings.TrimSpace(tenantID)
	listingID = strings.TrimSpace(listingID)
	if tenantID == "" || listingID == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", tenantID, listingID)
}

func cloneLicense(in *dbm.License) *dbm.License {
	if in == nil {
		return nil
	}
	out := *in
	if in.OfflineUntil != nil {
		v := *in.OfflineUntil
		out.OfflineUntil = &v
	}
	if in.LastValidatedAt != nil {
		v := *in.LastValidatedAt
		out.LastValidatedAt = &v
	}
	if in.RenewalToken != nil {
		v := *in.RenewalToken
		out.RenewalToken = &v
	}
	if in.IssuedBy != nil {
		v := *in.IssuedBy
		out.IssuedBy = &v
	}
	if in.Metadata != nil {
		copied := map[string]interface{}{}
		for k, v := range in.Metadata {
			copied[k] = v
		}
		out.Metadata = copied
	}
	return &out
}

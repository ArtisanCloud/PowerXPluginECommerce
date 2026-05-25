package capability

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	capcontract "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts/capability"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	deps  *app.Deps
	store *memoryStore
}

type memoryStore struct {
	mu          sync.RWMutex
	registry    map[string]map[string]any
	exposure    map[string]map[string]any
	plansByID   map[string]map[string]any
	plansByCap  map[string][]string
	planCounter int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		registry:   map[string]map[string]any{},
		exposure:   map[string]map[string]any{},
		plansByID:  map[string]map[string]any{},
		plansByCap: map[string][]string{},
	}
}

var sharedStore = newMemoryStore()

func NewHandler(deps *app.Deps) *Handler {
	return &Handler{deps: deps, store: sharedStore}
}

func (h *Handler) ListCatalog(c *gin.Context) {
	source := strings.TrimSpace(c.Query("source"))
	normalizedSource := strings.ToLower(source)
	switch normalizedSource {
	case "", "all", "any":
		normalizedSource = "all"
	case "platform":
		normalizedSource = "corex"
	}

	if normalizedSource == "corex" {
		items, err := h.fetchCoreXCatalog(c.Request.Context(), "corex")
		if err != nil {
			contracts.ResponseErrorWithDetails(
				c,
				http.StatusBadGateway,
				contracts.ErrCodeInternalError,
				"failed to fetch corex capabilities from gateway",
				gin.H{
					"source": "corex",
					"error":  err.Error(),
				},
			)
			return
		}
		// 兼容部分网关未实现 source=corex 过滤：为空时再拉一次全量并做 corex 特征筛选。
		if len(items) == 0 {
			if allItems, fallbackErr := h.fetchCoreXCatalog(c.Request.Context(), ""); fallbackErr == nil && len(allItems) > 0 {
				if filtered := filterCoreXCandidates(allItems); len(filtered) > 0 {
					items = filtered
				} else {
					items = allItems
				}
			}
		}
		contracts.ResponseSuccess(c, items)
		return
	}

	localItems, localErr := loadLocalCatalogItems()
	if localErr != nil {
		contracts.ResponseErrorWithDetails(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, "failed to load capability catalog", gin.H{"error": localErr.Error(), "dir": resolveCapabilityDir()})
		return
	}
	if normalizedSource == "plugin" {
		c.Header("X-Capability-Source", "plugin")
		contracts.ResponseSuccess(c, localItems)
		return
	}

	// source=all: 尝试合并 corex + plugin；若 corex 不可达则优先返回本地清单，避免页面空白。
	corexItems, corexErr := h.fetchCoreXCatalog(c.Request.Context(), "corex")
	if corexErr != nil {
		c.Header("X-Capability-Source", "all")
		c.Header("X-Capability-Corex-Warning", strings.TrimSpace(corexErr.Error()))
		contracts.ResponseSuccess(c, localItems)
		return
	}
	if len(corexItems) == 0 {
		if allItems, fallbackErr := h.fetchCoreXCatalog(c.Request.Context(), ""); fallbackErr == nil && len(allItems) > 0 {
			if filtered := filterCoreXCandidates(allItems); len(filtered) > 0 {
				corexItems = filtered
			} else {
				corexItems = allItems
			}
		}
	}
	merged := mergeCatalogItems(corexItems, localItems)
	c.Header("X-Capability-Source", "all")
	contracts.ResponseSuccess(c, merged)
}

func loadLocalCatalogItems() ([]gin.H, error) {
	dir := resolveCapabilityDir()
	catalog, err := capcontract.LoadCatalog(dir)
	if err != nil {
		return nil, err
	}

	records := catalog.List()
	items := make([]gin.H, 0, len(records))
	for _, rec := range records {
		id := strings.TrimSpace(rec.Descriptor.ID)
		if id == "" {
			continue
		}
		relDescriptor := relPathFromPluginRoot(rec.Source)
		checksum := fileChecksum(rec.Source)
		metadata := rec.Descriptor.Metadata
		protocols := mapFromAny(metadata["protocols"])
		tags := stringSliceAny(metadata["tags"])
		if len(tags) == 0 {
			tags = []string{"corex", "plugin"}
		}
		mode := "sync"
		if asyncMode := strings.TrimSpace(fmt.Sprint(metadata["async_mode"])); asyncMode != "" {
			mode = asyncMode
		}
		items = append(items, gin.H{
			"id":         id,
			"version":    nonEmpty(rec.Descriptor.Version, "1.0.0"),
			"descriptor": relDescriptor,
			"module":     deriveModule(id),
			"kind":       strings.ToLower(string(rec.Descriptor.Type)),
			"tags":       tags,
			"checksum":   checksum,
			"execution": gin.H{
				"mode": mode,
			},
			"protocols": protocols,
		})
	}
	return items, nil
}

func (h *Handler) ListSources(c *gin.Context) {
	contracts.ResponseSuccess(c, gin.H{
		"default": "all",
		"aliases": gin.H{
			"all":      "all",
			"any":      "all",
			"platform": "corex",
		},
		"sources": []gin.H{
			{"id": "all", "label": "all", "description": "查询全部来源（不传 source 或 source=all）"},
			{"id": "corex", "label": "corex", "description": "PowerX 底座能力"},
			{"id": "plugin", "label": "plugin", "description": "插件/租户注册能力"},
		},
	})
}

func (h *Handler) GetRegisterTemplate(c *gin.Context) {
	contracts.ResponseSuccess(c, gin.H{
		"namespace":           "com.powerx.plugins.ecommerce",
		"sensitivity_options": []string{"low", "normal", "high"},
		"async_modes":         []string{"sync", "async"},
		"tag_suggestions":     []string{"catalog", "corex", "template"},
		"field_hints": map[string]string{
			"scenario":    "说明租户范围、前置条件与调用场景",
			"description": "描述能力边界、幂等性与异常处理",
		},
		"schema_placeholders": map[string]string{
			"input":  "{\n  \"tenant_uuid\": \"\"\n}",
			"output": "{\n  \"ok\": true\n}",
		},
		"protocol_samples": map[string]string{
			"rest":     "{\n  \"method\": \"POST\",\n  \"path\": \"/api/v1/resources\"\n}",
			"grpc":     "{\n  \"service\": \"powerx.capability.v1.CapabilityService\",\n  \"method\": \"Invoke\"\n}",
			"workflow": "{\n  \"template\": \"contracts/exposure/workflow/template-create.json\"\n}",
		},
		"identifier_example": "com.powerx.plugins.ecommerce.template.create",
	})
}

func (h *Handler) ValidateDraft(c *gin.Context) {
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	capabilityID := buildCapabilityID(payload)
	errs := make([]gin.H, 0)
	if capabilityID == "" {
		errs = append(errs, gin.H{"field": "capability_id", "message": "capability id cannot be empty", "suggestion": "fill namespace/resource/action"})
	}
	if strings.TrimSpace(fmt.Sprint(payload["scenario"])) == "" {
		errs = append(errs, gin.H{"field": "scenario", "message": "scenario cannot be empty"})
	}
	result := gin.H{
		"capability_id": capabilityID,
		"valid":         len(errs) == 0,
		"errors":        errs,
	}
	if len(errs) > 0 {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "capability validation failed", result)
		return
	}
	contracts.ResponseSuccess(c, result)
}

func (h *Handler) Submit(c *gin.Context) {
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	capabilityID := buildCapabilityID(payload)
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capability_id cannot be empty")
		return
	}
	now := time.Now().Format(time.RFC3339)
	status := "submitted"
	if asBool(payload["draft"]) {
		status = "draft"
	}
	record := deepCopyMap(payload)
	record["capability_id"] = capabilityID
	record["status"] = status
	record["created_at"] = now
	record["updated_at"] = now

	h.store.mu.Lock()
	h.store.registry[capabilityID] = record
	h.store.mu.Unlock()

	contracts.ResponseSuccess(c, record)
}

func (h *Handler) GetExposureTemplate(c *gin.Context) {
	contracts.ResponseSuccess(c, gin.H{
		"channel_types":   []string{"rest", "grpc", "workflow", "agent_stream", "tool"},
		"auth_strategies": []string{"powerx_session", "jwt", "sts"},
		"default_rate":    gin.H{"requests_per_minute": 120, "burst": 60, "concurrency": 30},
	})
}

func (h *Handler) GetExposurePackage(c *gin.Context) {
	capabilityID := strings.TrimSpace(c.Param("capabilityID"))
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capabilityID is required")
		return
	}
	h.store.mu.RLock()
	pkg := deepCopyMap(h.store.exposure[capabilityID])
	h.store.mu.RUnlock()
	contracts.ResponseSuccess(c, gin.H{"package": pkg})
}

func (h *Handler) UpsertExposurePackage(c *gin.Context) {
	capabilityID := strings.TrimSpace(c.Param("capabilityID"))
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capabilityID is required")
		return
	}
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	now := time.Now().Format(time.RFC3339)
	record := deepCopyMap(payload)
	record["capability_id"] = capabilityID
	record["sync_status"] = "synced"
	record["updated_at"] = now
	record["updated_by"] = "root"
	if _, ok := record["tenants"]; !ok {
		record["tenants"] = []any{}
	}
	h.store.mu.Lock()
	h.store.exposure[capabilityID] = record
	h.store.mu.Unlock()
	contracts.ResponseSuccess(c, record)
}

func (h *Handler) ListQuotas(c *gin.Context) {
	capabilityID := strings.TrimSpace(c.Param("capabilityID"))
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capabilityID is required")
		return
	}
	h.store.mu.RLock()
	pkg := h.store.exposure[capabilityID]
	h.store.mu.RUnlock()
	if pkg == nil {
		contracts.ResponseSuccess(c, gin.H{"quotas": []any{}})
		return
	}
	quotas, _ := pkg["tenants"].([]any)
	if quotas == nil {
		quotas = []any{}
	}
	contracts.ResponseSuccess(c, gin.H{"quotas": quotas})
}

func (h *Handler) UpsertQuota(c *gin.Context) {
	capabilityID := strings.TrimSpace(c.Param("capabilityID"))
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capabilityID is required")
		return
	}
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	tenantID := strings.TrimSpace(fmt.Sprint(payload["tenant_id"]))
	if tenantID == "" {
		contracts.ResponseBadRequest(c, "tenant_id is required")
		return
	}

	h.store.mu.Lock()
	record := deepCopyMap(h.store.exposure[capabilityID])
	if record == nil {
		record = map[string]any{
			"capability_id": capabilityID,
			"channels":      []any{},
			"auth":          map[string]any{"strategy": "powerx_session", "scopes": []any{}},
			"rate_limit":    map[string]any{"requests_per_minute": 120, "burst": 60, "concurrency": 30},
			"docs_version":  "1.0.0",
			"sdk_version":   "1.0.0",
			"sync_status":   "synced",
		}
	}
	tenants := asMapSlice(record["tenants"])
	next := make([]map[string]any, 0, len(tenants)+1)
	updated := false
	now := time.Now().Format(time.RFC3339)
	for _, item := range tenants {
		if strings.TrimSpace(fmt.Sprint(item["tenant_id"])) == tenantID {
			merge := deepCopyMap(item)
			for k, v := range payload {
				merge[k] = v
			}
			merge["updated_at"] = now
			next = append(next, merge)
			updated = true
			continue
		}
		next = append(next, item)
	}
	if !updated {
		insert := deepCopyMap(payload)
		insert["updated_at"] = now
		next = append(next, insert)
	}

	record["tenants"] = mapSliceToAny(next)
	record["updated_at"] = now
	record["updated_by"] = "root"
	h.store.exposure[capabilityID] = record
	h.store.mu.Unlock()

	contracts.ResponseSuccess(c, record)
}

func (h *Handler) GetLifecycleTemplate(c *gin.Context) {
	contracts.ResponseSuccess(c, gin.H{
		"change_types":    []string{"minor", "major", "deprecate", "sunset"},
		"status_options":  []string{"draft", "pending", "approved", "paused", "completed"},
		"channel_options": []string{"email", "slack", "webhook", "notice-center"},
	})
}

func (h *Handler) ListLifecyclePlans(c *gin.Context) {
	capabilityID := strings.TrimSpace(c.Query("capability_id"))
	h.store.mu.RLock()
	plans := make([]map[string]any, 0)
	if capabilityID != "" {
		for _, pid := range h.store.plansByCap[capabilityID] {
			if plan := deepCopyMap(h.store.plansByID[pid]); plan != nil {
				plans = append(plans, plan)
			}
		}
	} else {
		for _, plan := range h.store.plansByID {
			if cp := deepCopyMap(plan); cp != nil {
				plans = append(plans, cp)
			}
		}
	}
	h.store.mu.RUnlock()
	sort.Slice(plans, func(i, j int) bool {
		left := fmt.Sprint(plans[i]["updated_at"])
		right := fmt.Sprint(plans[j]["updated_at"])
		return left > right
	})
	contracts.ResponseSuccess(c, gin.H{"plans": plans})
}

func (h *Handler) CreateLifecyclePlan(c *gin.Context) {
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	capabilityID := strings.TrimSpace(fmt.Sprint(payload["capability_id"]))
	if capabilityID == "" {
		contracts.ResponseBadRequest(c, "capability_id is required")
		return
	}
	now := time.Now().Format(time.RFC3339)

	h.store.mu.Lock()
	h.store.planCounter++
	planID := fmt.Sprintf("plan-%d", h.store.planCounter)
	record := deepCopyMap(payload)
	record["id"] = planID
	record["status"] = nonEmpty(fmt.Sprint(payload["status"]), "pending")
	record["created_at"] = now
	record["updated_at"] = now
	record["created_by"] = "root"
	h.store.plansByID[planID] = record
	h.store.plansByCap[capabilityID] = append(h.store.plansByCap[capabilityID], planID)
	h.store.mu.Unlock()

	contracts.ResponseSuccess(c, record)
}

func (h *Handler) UpdateLifecycleStatus(c *gin.Context) {
	planID := strings.TrimSpace(c.Param("planID"))
	if planID == "" {
		contracts.ResponseBadRequest(c, "planID is required")
		return
	}
	payload := map[string]any{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		contracts.ResponseBadRequest(c, "invalid payload: "+err.Error())
		return
	}
	status := strings.TrimSpace(fmt.Sprint(payload["status"]))
	if status == "" {
		contracts.ResponseBadRequest(c, "status is required")
		return
	}
	h.store.mu.Lock()
	record := deepCopyMap(h.store.plansByID[planID])
	if record == nil {
		h.store.mu.Unlock()
		contracts.ResponseNotFound(c, "lifecycle plan not found")
		return
	}
	record["status"] = status
	if notes := strings.TrimSpace(fmt.Sprint(payload["notes"])); notes != "" {
		record["notes"] = notes
	}
	record["updated_at"] = time.Now().Format(time.RFC3339)
	h.store.plansByID[planID] = record
	h.store.mu.Unlock()
	contracts.ResponseSuccess(c, record)
}

func resolveCapabilityDir() string {
	candidates := []string{
		filepath.Join("contracts", "capabilities"),
		filepath.Join("..", "contracts", "capabilities"),
	}
	if raw := strings.TrimSpace(os.Getenv("CONFIG_PATH")); raw != "" {
		p := raw
		if stat, err := os.Stat(raw); err == nil && !stat.IsDir() {
			p = filepath.Dir(raw)
		}
		candidates = append(candidates,
			filepath.Join(p, "..", "contracts", "capabilities"),
			filepath.Join(p, "..", "..", "contracts", "capabilities"),
		)
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return filepath.Join("contracts", "capabilities")
}

func relPathFromPluginRoot(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.ToSlash(path)
	if idx := strings.Index(clean, "/contracts/"); idx > 0 {
		return clean[idx+1:]
	}
	if strings.HasPrefix(clean, "contracts/") {
		return clean
	}
	if strings.HasPrefix(clean, "../contracts/") {
		return strings.TrimPrefix(clean, "../")
	}
	return clean
}

func fileChecksum(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

func deriveModule(id string) string {
	parts := strings.Split(strings.TrimSpace(id), ".")
	if len(parts) < 2 {
		return "core"
	}
	if len(parts) >= 6 {
		return parts[4]
	}
	return parts[len(parts)-2]
}

func nonEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func buildCapabilityID(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if v := strings.TrimSpace(fmt.Sprint(payload["capability_id"])); v != "" {
		return v
	}
	ns := strings.TrimSpace(fmt.Sprint(payload["namespace"]))
	resource := strings.Trim(strings.TrimSpace(fmt.Sprint(payload["resource"])), " ./")
	action := strings.Trim(strings.TrimSpace(fmt.Sprint(payload["action"])), " ./")
	if ns == "" || resource == "" || action == "" {
		return ""
	}
	resource = strings.ReplaceAll(resource, "/", ".")
	resource = strings.ReplaceAll(resource, " ", "_")
	action = strings.ReplaceAll(action, "/", ".")
	action = strings.ReplaceAll(action, " ", "_")
	return fmt.Sprintf("%s.%s.%s", ns, resource, action)
}

func mapFromAny(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	if out, ok := v.(map[string]any); ok {
		return out
	}
	if out, ok := v.(map[any]any); ok {
		ret := make(map[string]any, len(out))
		for k, val := range out {
			ret[fmt.Sprint(k)] = val
		}
		return ret
	}
	return map[string]any{}
}

func stringSliceAny(v any) []string {
	if v == nil {
		return nil
	}
	switch vv := v.(type) {
	case []string:
		out := make([]string, 0, len(vv))
		for _, item := range vv {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(vv))
		for _, item := range vv {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
			return []string{s}
		}
		return nil
	}
}

func deepCopyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = deepCopyAny(v)
	}
	return out
}

func deepCopyAny(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		return deepCopyMap(vv)
	case []any:
		out := make([]any, 0, len(vv))
		for _, item := range vv {
			out = append(out, deepCopyAny(item))
		}
		return out
	default:
		return vv
	}
}

func asMapSlice(v any) []map[string]any {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		if typed, ok2 := v.([]map[string]any); ok2 {
			return typed
		}
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func mapSliceToAny(in []map[string]any) []any {
	out := make([]any, 0, len(in))
	for _, item := range in {
		out = append(out, item)
	}
	return out
}

func asBool(v any) bool {
	s := strings.TrimSpace(strings.ToLower(fmt.Sprint(v)))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func (h *Handler) fetchCoreXCatalog(ctx context.Context, source string) ([]any, error) {
	base := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	if base == "" {
		return nil, fmt.Errorf("gateway base not configured")
	}
	base = ensureAPIV1(base)
	u, err := url.Parse(strings.TrimRight(base, "/") + "/admin/capabilities")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if source != "" {
		q.Set("source", source)
	}
	u.RawQuery = q.Encode()

	scheme := resolveCoreXGatewayAuthScheme()
	token := ""
	if h != nil && h.deps != nil {
		if stsToken, err := h.deps.PowerXAccessToken(ctx); err == nil {
			token = strings.TrimSpace(stsToken)
		}
	}
	if token == "" {
		token = firstNonEmptyEnv("POWERX_AUTH_TOKEN")
	}
	apiKey := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY"))
	items, statusCode, rawBody, err := fetchCoreXCatalogWithAuth(ctx, u.String(), scheme, token, apiKey)
	if err == nil {
		return items, nil
	}
	// 兼容网关鉴权差异：主鉴权 401 时自动切换另一种鉴权重试一次。
	if statusCode == http.StatusUnauthorized {
		if scheme == "apikey" && token != "" {
			if retryItems, _, _, retryErr := fetchCoreXCatalogWithAuth(ctx, u.String(), "bearer", token, apiKey); retryErr == nil {
				return retryItems, nil
			}
		}
		if scheme == "bearer" && apiKey != "" {
			if retryItems, _, _, retryErr := fetchCoreXCatalogWithAuth(ctx, u.String(), "apikey", token, apiKey); retryErr == nil {
				return retryItems, nil
			}
		}
	}
	return nil, fmt.Errorf("gateway status=%d body=%s", statusCode, strings.TrimSpace(rawBody))
}

func fetchCoreXCatalogWithAuth(ctx context.Context, endpoint, scheme, token, apiKey string) ([]any, int, string, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, "", err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	req.Header.Set("Accept", "application/json")
	if scheme == "apikey" {
		if apiKey != "" {
			req.Header.Set("Authorization", "ApiKey "+apiKey)
			req.Header.Set("X-API-Key", apiKey)
		}
	} else if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-PowerX-Service-Token", token)
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	rawBody := strings.TrimSpace(string(raw))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, resp.StatusCode, rawBody, fmt.Errorf("gateway status=%d body=%s", resp.StatusCode, rawBody)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, resp.StatusCode, rawBody, err
	}
	if data, ok := payload["data"].([]any); ok {
		return data, resp.StatusCode, rawBody, nil
	}
	if nested, ok := payload["data"].(map[string]any); ok {
		if items, ok := nested["items"].([]any); ok {
			return items, resp.StatusCode, rawBody, nil
		}
	}
	return []any{}, resp.StatusCode, rawBody, nil
}

func ensureAPIV1(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "/api/v1") {
		return base
	}
	if strings.HasSuffix(lower, "/api") {
		return base + "/v1"
	}
	if strings.Contains(lower, "/api/v1/") {
		return base
	}
	return base + "/api/v1"
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val
		}
	}
	return ""
}

func normalizeCoreXAuthScheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "apikey", "api_key", "api-key":
		return "apikey"
	case "bearer":
		return "bearer"
	default:
		return ""
	}
}

func resolveCoreXGatewayAuthScheme() string {
	explicit := normalizeCoreXAuthScheme(os.Getenv("PX_GATEWAY_AUTH_SCHEME"))
	if explicit != "" {
		return explicit
	}
	if strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY")) != "" {
		return "apikey"
	}
	return "bearer"
}

func filterCoreXCandidates(items []any) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := strings.ToLower(normalizeCatalogItemID(m))
		module := strings.ToLower(strings.TrimSpace(fmt.Sprint(m["module"])))
		source := strings.ToLower(strings.TrimSpace(fmt.Sprint(m["source"])))
		if strings.Contains(id, "corex") || strings.Contains(module, "corex") || source == "corex" {
			out = append(out, item)
			continue
		}
		if tags, ok := m["tags"].([]any); ok {
			for _, tag := range tags {
				if strings.EqualFold(strings.TrimSpace(fmt.Sprint(tag)), "corex") {
					out = append(out, item)
					break
				}
			}
		}
	}
	return out
}

func mergeCatalogItems(corexItems []any, localItems []gin.H) []any {
	seen := make(map[string]struct{})
	out := make([]any, 0, len(corexItems)+len(localItems))
	for _, item := range corexItems {
		id := ""
		if m, ok := item.(map[string]any); ok {
			id = normalizeCatalogItemID(m)
		}
		if id != "" {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
		}
		out = append(out, item)
	}
	for _, item := range localItems {
		id := normalizeCatalogItemID(item)
		if id != "" {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
		}
		out = append(out, item)
	}
	return out
}

func normalizeCatalogItemID(item map[string]any) string {
	if item == nil {
		return ""
	}
	candidates := []string{
		strings.TrimSpace(fmt.Sprint(item["id"])),
		strings.TrimSpace(fmt.Sprint(item["capability_id"])),
		strings.TrimSpace(fmt.Sprint(item["capabilityId"])),
		strings.TrimSpace(fmt.Sprint(item["name"])),
	}
	for _, candidate := range candidates {
		if candidate != "" && candidate != "<nil>" {
			return candidate
		}
	}
	return ""
}

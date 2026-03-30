package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PolicyOrchestrationService struct {
	flowRepo    *LogisticsRepo.PolicyOrchestrationFlowRepository
	versionRepo *LogisticsRepo.PolicyOrchestrationVersionRepository
}

type PolicyOrchestrationFlowQuery struct {
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type UpsertPolicyOrchestrationFlowRequest struct {
	ID                string         `json:"id,omitempty"`
	Name              string         `json:"name"`
	Priority          int            `json:"priority,omitempty"`
	Status            string         `json:"status,omitempty"`
	FlowDefinition    map[string]any `json:"flow_definition,omitempty"`
	ConflictRelations []string       `json:"conflict_relations,omitempty"`
	GrayReleaseConfig map[string]any `json:"gray_release_config,omitempty"`
	Description       string         `json:"description,omitempty"`
	OperatorID        string         `json:"operator_id,omitempty"`
}

type PolicyOrchestrationVersionQuery struct {
	FlowID string `json:"flow_id"`
	Limit  int    `json:"limit,omitempty"`
}

type PolicyOrchestrationConflictPreviewRequest struct {
	FlowID            string         `json:"flow_id,omitempty"`
	Name              string         `json:"name,omitempty"`
	ConflictRelations []string       `json:"conflict_relations,omitempty"`
	FlowDefinition    map[string]any `json:"flow_definition,omitempty"`
	GrayReleaseConfig map[string]any `json:"gray_release_config,omitempty"`
}

type PolicyOrchestrationConflictItem struct {
	FlowID        string   `json:"flow_id"`
	FlowName      string   `json:"flow_name"`
	ConflictKeys  []string `json:"conflict_keys"`
	ConflictCount int      `json:"conflict_count"`
	Reason        string   `json:"reason"`
}

type PolicyOrchestrationConflictPreviewResult struct {
	HasConflict bool                              `json:"has_conflict"`
	Items       []PolicyOrchestrationConflictItem `json:"items"`
}

type PublishPolicyOrchestrationRequest struct {
	FlowID          string         `json:"flow_id"`
	RequestKey      string         `json:"request_key,omitempty"`
	Force           bool           `json:"force,omitempty"`
	ChangeSummary   string         `json:"change_summary,omitempty"`
	GrayReleasePlan map[string]any `json:"gray_release_plan,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
}

type RollbackPolicyOrchestrationRequest struct {
	FlowID          string `json:"flow_id"`
	TargetVersionID string `json:"target_version_id"`
	Reason          string `json:"reason,omitempty"`
	OperatorID      string `json:"operator_id,omitempty"`
}

type PolicyOrchestrationPublishResult struct {
	Flow            *LogisticsModel.PolicyOrchestrationFlow    `json:"flow"`
	Version         *LogisticsModel.PolicyOrchestrationVersion `json:"version"`
	Idempotency     string                                     `json:"idempotency_status"`
	ConflictPreview PolicyOrchestrationConflictPreviewResult   `json:"conflict_preview"`
}

func NewPolicyOrchestrationService(deps *app.Deps) *PolicyOrchestrationService {
	if deps == nil || deps.DB == nil {
		return &PolicyOrchestrationService{}
	}
	return &PolicyOrchestrationService{
		flowRepo:    LogisticsRepo.NewPolicyOrchestrationFlowRepository(deps.DB),
		versionRepo: LogisticsRepo.NewPolicyOrchestrationVersionRepository(deps.DB),
	}
}

func (s *PolicyOrchestrationService) ListFlows(ctx context.Context, tenantUUID string, query PolicyOrchestrationFlowQuery) ([]LogisticsModel.PolicyOrchestrationFlow, error) {
	if s == nil || s.flowRepo == nil {
		return nil, errors.New("policy orchestration service unavailable")
	}
	return s.flowRepo.List(withTenantContext(ctx, tenantUUID), LogisticsRepo.PolicyOrchestrationFlowFilter{
		Status: normalizePolicyFlowStatus(query.Status),
		Limit:  query.Limit,
	})
}

func (s *PolicyOrchestrationService) UpsertFlow(ctx context.Context, tenantUUID string, req UpsertPolicyOrchestrationFlowRequest) (*LogisticsModel.PolicyOrchestrationFlow, error) {
	if s == nil || s.flowRepo == nil {
		return nil, errors.New("policy orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}
	status := normalizePolicyFlowStatus(req.Status)
	if status == "" {
		status = "draft"
	}
	defJSON, err := jsonBytes(req.FlowDefinition, []byte("{}"))
	if err != nil {
		return nil, err
	}
	conflictKeys := cleanConflictKeys(req.ConflictRelations)
	conflictJSON, err := jsonBytes(conflictKeys, []byte("[]"))
	if err != nil {
		return nil, err
	}
	grayJSON, err := jsonBytes(req.GrayReleaseConfig, []byte("{}"))
	if err != nil {
		return nil, err
	}

	row := &LogisticsModel.PolicyOrchestrationFlow{ID: strings.TrimSpace(req.ID)}
	if row.ID == "" {
		existing, getErr := s.flowRepo.GetByName(ctx, name)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		} else {
			row.ID = utils.NewUUID()
		}
	} else {
		existing, getErr := s.flowRepo.GetByID(ctx, row.ID)
		if getErr == nil && existing != nil {
			row = existing
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
		if row.ID == "" {
			row.ID = strings.TrimSpace(req.ID)
		}
	}

	row.Name = name
	row.Priority = priority
	row.Status = status
	row.FlowDefinition = datatypes.JSON(defJSON)
	row.ConflictRelations = datatypes.JSON(conflictJSON)
	row.GrayReleaseConfig = datatypes.JSON(grayJSON)
	row.Description = strings.TrimSpace(req.Description)
	if op := strings.TrimSpace(req.OperatorID); op != "" {
		if row.CreatedBy == "" {
			row.CreatedBy = op
		}
		row.UpdatedBy = op
	}

	if err := s.flowRepo.Save(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *PolicyOrchestrationService) ListVersions(ctx context.Context, tenantUUID string, query PolicyOrchestrationVersionQuery) ([]LogisticsModel.PolicyOrchestrationVersion, error) {
	if s == nil || s.versionRepo == nil {
		return nil, errors.New("policy orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(query.FlowID) == "" {
		return nil, errors.New("flow_id is required")
	}
	return s.versionRepo.List(ctx, LogisticsRepo.PolicyOrchestrationVersionFilter{
		FlowID: strings.TrimSpace(query.FlowID),
		Limit:  query.Limit,
	})
}

func (s *PolicyOrchestrationService) PreviewConflicts(ctx context.Context, tenantUUID string, req PolicyOrchestrationConflictPreviewRequest) (PolicyOrchestrationConflictPreviewResult, error) {
	if s == nil || s.flowRepo == nil {
		return PolicyOrchestrationConflictPreviewResult{}, errors.New("policy orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	targetID := strings.TrimSpace(req.FlowID)
	targetName := strings.TrimSpace(req.Name)
	targetKeys := cleanConflictKeys(req.ConflictRelations)
	if len(targetKeys) == 0 && req.FlowDefinition != nil {
		targetKeys = cleanConflictKeys(extractConflictKeysFromDefinition(req.FlowDefinition))
	}
	if len(targetKeys) == 0 && targetID != "" {
		row, err := s.flowRepo.GetByID(ctx, targetID)
		if err != nil {
			return PolicyOrchestrationConflictPreviewResult{}, err
		}
		targetKeys = cleanConflictKeys(policyJSONStringSliceFromRaw(row.ConflictRelations))
		targetName = row.Name
	}
	if len(targetKeys) == 0 {
		return PolicyOrchestrationConflictPreviewResult{HasConflict: false, Items: []PolicyOrchestrationConflictItem{}}, nil
	}

	allRows, err := s.flowRepo.List(ctx, LogisticsRepo.PolicyOrchestrationFlowFilter{Status: "active", Limit: 500})
	if err != nil {
		return PolicyOrchestrationConflictPreviewResult{}, err
	}

	items := make([]PolicyOrchestrationConflictItem, 0)
	for _, item := range allRows {
		if targetID != "" && item.ID == targetID {
			continue
		}
		if targetID == "" && targetName != "" && strings.EqualFold(strings.TrimSpace(item.Name), targetName) {
			continue
		}
		currentKeys := cleanConflictKeys(policyJSONStringSliceFromRaw(item.ConflictRelations))
		overlaps := intersectConflictKeys(targetKeys, currentKeys)
		if len(overlaps) == 0 {
			continue
		}
		items = append(items, PolicyOrchestrationConflictItem{
			FlowID:        item.ID,
			FlowName:      item.Name,
			ConflictKeys:  overlaps,
			ConflictCount: len(overlaps),
			Reason:        fmt.Sprintf("与已发布流程 %s 存在冲突键重叠", item.Name),
		})
	}
	return PolicyOrchestrationConflictPreviewResult{
		HasConflict: len(items) > 0,
		Items:       items,
	}, nil
}

func (s *PolicyOrchestrationService) Publish(ctx context.Context, tenantUUID string, req PublishPolicyOrchestrationRequest) (*PolicyOrchestrationPublishResult, error) {
	if s == nil || s.flowRepo == nil || s.versionRepo == nil {
		return nil, errors.New("policy orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	flowID := strings.TrimSpace(req.FlowID)
	if flowID == "" {
		return nil, errors.New("flow_id is required")
	}
	flow, err := s.flowRepo.GetByID(ctx, flowID)
	if err != nil {
		return nil, err
	}
	conflictPreview, err := s.PreviewConflicts(ctx, tenantUUID, PolicyOrchestrationConflictPreviewRequest{FlowID: flowID})
	if err != nil {
		return nil, err
	}
	if conflictPreview.HasConflict && !req.Force {
		return nil, errors.New("conflict detected, set force=true to publish anyway")
	}

	requestKey := strings.TrimSpace(req.RequestKey)
	if requestKey == "" {
		requestKey = fmt.Sprintf("publish#%s#%s", flowID, strings.TrimSpace(req.OperatorID))
	}
	if existing, getErr := s.versionRepo.GetByRequestKey(ctx, flowID, requestKey); getErr == nil && existing != nil {
		return &PolicyOrchestrationPublishResult{
			Flow:            flow,
			Version:         existing,
			Idempotency:     "replayed",
			ConflictPreview: conflictPreview,
		}, nil
	} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
		return nil, getErr
	}

	result := &PolicyOrchestrationPublishResult{Idempotency: "created", ConflictPreview: conflictPreview}
	err = s.flowRepo.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockedFlow, lockErr := s.flowRepo.GetByIDForUpdate(ctx, tx, flowID)
		if lockErr != nil {
			return lockErr
		}
		if existing, getErr := s.versionRepo.GetByRequestKey(ctx, flowID, requestKey); getErr == nil && existing != nil {
			result.Flow = lockedFlow
			result.Version = existing
			result.Idempotency = "replayed"
			return nil
		} else if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return getErr
		}

		nextNo, nextErr := s.versionRepo.NextVersionNo(ctx, tx, flowID)
		if nextErr != nil {
			return nextErr
		}
		grayJSON, grayErr := jsonBytes(req.GrayReleasePlan, []byte("{}"))
		if grayErr != nil {
			return grayErr
		}
		conflictJSON, conflictErr := jsonBytes(conflictPreview.Items, []byte("[]"))
		if conflictErr != nil {
			return conflictErr
		}
		snapshotJSON, snapshotErr := jsonBytes(map[string]any{
			"flow_definition":     policyJSONMapFromRaw(lockedFlow.FlowDefinition),
			"conflict_relations":  policyJSONStringSliceFromRaw(lockedFlow.ConflictRelations),
			"gray_release_config": policyJSONMapFromRaw(lockedFlow.GrayReleaseConfig),
			"priority":            lockedFlow.Priority,
			"description":         lockedFlow.Description,
		}, []byte("{}"))
		if snapshotErr != nil {
			return snapshotErr
		}
		now := time.Now().UTC()
		version := &LogisticsModel.PolicyOrchestrationVersion{
			ID:              utils.NewUUID(),
			TenantUUID:      lockedFlow.TenantUUID,
			FlowID:          flowID,
			VersionNo:       nextNo,
			RequestKey:      requestKey,
			Status:          "published",
			ChangeSummary:   strings.TrimSpace(req.ChangeSummary),
			Snapshot:        datatypes.JSON(snapshotJSON),
			ConflictReport:  datatypes.JSON(conflictJSON),
			GrayReleasePlan: datatypes.JSON(grayJSON),
			PublishedAt:     &now,
			CreatedBy:       strings.TrimSpace(req.OperatorID),
			UpdatedBy:       strings.TrimSpace(req.OperatorID),
		}
		if err := tx.WithContext(ctx).Save(version).Error; err != nil {
			return err
		}

		lockedFlow.PublishedVersionID = version.ID
		lockedFlow.Status = "active"
		if op := strings.TrimSpace(req.OperatorID); op != "" {
			lockedFlow.UpdatedBy = op
		}
		if err := tx.WithContext(ctx).Save(lockedFlow).Error; err != nil {
			return err
		}
		result.Flow = lockedFlow
		result.Version = version
		return nil
	})
	if err != nil {
		if strings.TrimSpace(requestKey) != "" {
			if existing, getErr := s.versionRepo.GetByRequestKey(ctx, flowID, requestKey); getErr == nil && existing != nil {
				flow, _ := s.flowRepo.GetByID(ctx, flowID)
				return &PolicyOrchestrationPublishResult{
					Flow:            flow,
					Version:         existing,
					Idempotency:     "replayed",
					ConflictPreview: conflictPreview,
				}, nil
			}
		}
		return nil, err
	}
	return result, nil
}

func (s *PolicyOrchestrationService) Rollback(ctx context.Context, tenantUUID string, req RollbackPolicyOrchestrationRequest) (*PolicyOrchestrationPublishResult, error) {
	if s == nil || s.flowRepo == nil || s.versionRepo == nil {
		return nil, errors.New("policy orchestration service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	flowID := strings.TrimSpace(req.FlowID)
	targetVersionID := strings.TrimSpace(req.TargetVersionID)
	if flowID == "" || targetVersionID == "" {
		return nil, errors.New("flow_id and target_version_id are required")
	}
	targetVersion, err := s.versionRepo.GetByID(ctx, targetVersionID)
	if err != nil {
		return nil, err
	}
	if targetVersion.FlowID != flowID {
		return nil, errors.New("target version does not belong to flow")
	}
	result := &PolicyOrchestrationPublishResult{Idempotency: "created"}
	err = s.flowRepo.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		flow, lockErr := s.flowRepo.GetByIDForUpdate(ctx, tx, flowID)
		if lockErr != nil {
			return lockErr
		}
		snapshot := policyJSONMapFromRaw(targetVersion.Snapshot)
		defJSON, defErr := jsonBytes(snapshot["flow_definition"], []byte("{}"))
		if defErr != nil {
			return defErr
		}
		conflictJSON, conflictErr := jsonBytes(snapshot["conflict_relations"], []byte("[]"))
		if conflictErr != nil {
			return conflictErr
		}
		grayJSON, grayErr := jsonBytes(snapshot["gray_release_config"], []byte("{}"))
		if grayErr != nil {
			return grayErr
		}
		flow.FlowDefinition = datatypes.JSON(defJSON)
		flow.ConflictRelations = datatypes.JSON(conflictJSON)
		flow.GrayReleaseConfig = datatypes.JSON(grayJSON)
		if priority, ok := snapshot["priority"].(float64); ok && int(priority) > 0 {
			flow.Priority = int(priority)
		}
		if desc, ok := snapshot["description"].(string); ok {
			flow.Description = desc
		}
		flow.PublishedVersionID = targetVersion.ID
		flow.Status = "active"
		if op := strings.TrimSpace(req.OperatorID); op != "" {
			flow.UpdatedBy = op
		}
		if err := tx.WithContext(ctx).Save(flow).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		nextNo, nextErr := s.versionRepo.NextVersionNo(ctx, tx, flowID)
		if nextErr != nil {
			return nextErr
		}
		newVersion := &LogisticsModel.PolicyOrchestrationVersion{
			ID:               utils.NewUUID(),
			TenantUUID:       flow.TenantUUID,
			FlowID:           flowID,
			VersionNo:        nextNo,
			RequestKey:       fmt.Sprintf("rollback#%s#%s", flowID, targetVersion.ID),
			Status:           "rolled_back",
			ChangeSummary:    strings.TrimSpace(req.Reason),
			Snapshot:         targetVersion.Snapshot,
			ConflictReport:   targetVersion.ConflictReport,
			GrayReleasePlan:  targetVersion.GrayReleasePlan,
			RolledBackFromID: targetVersion.ID,
			PublishedAt:      &now,
			CreatedBy:        strings.TrimSpace(req.OperatorID),
			UpdatedBy:        strings.TrimSpace(req.OperatorID),
		}
		if err := tx.WithContext(ctx).Save(newVersion).Error; err != nil {
			return err
		}

		result.Flow = flow
		result.Version = newVersion
		result.ConflictPreview = PolicyOrchestrationConflictPreviewResult{HasConflict: false, Items: []PolicyOrchestrationConflictItem{}}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func normalizePolicyFlowStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "draft", "active", "archived":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func cleanConflictKeys(raw []string) []string {
	if len(raw) == 0 {
		return []string{}
	}
	set := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		set[key] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func extractConflictKeysFromDefinition(flow map[string]any) []string {
	if flow == nil {
		return []string{}
	}
	raw, ok := flow["conflict_keys"]
	if !ok || raw == nil {
		return []string{}
	}
	switch vv := raw.(type) {
	case []any:
		keys := make([]string, 0, len(vv))
		for _, item := range vv {
			keys = append(keys, fmt.Sprintf("%v", item))
		}
		return keys
	case []string:
		return vv
	default:
		return []string{}
	}
}

func intersectConflictKeys(a, b []string) []string {
	if len(a) == 0 || len(b) == 0 {
		return []string{}
	}
	lookup := make(map[string]struct{}, len(a))
	for _, item := range a {
		lookup[item] = struct{}{}
	}
	out := make([]string, 0)
	for _, item := range b {
		if _, ok := lookup[item]; ok {
			out = append(out, item)
		}
	}
	out = cleanConflictKeys(out)
	return out
}

func policyJSONMapFromRaw(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}
	}
	out := make(map[string]any)
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func policyJSONStringSliceFromRaw(raw datatypes.JSON) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return cleanConflictKeys(out)
}

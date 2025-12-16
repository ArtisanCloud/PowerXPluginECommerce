package channel_master

import (
	"net/http"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/channel_master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DetailHandler aggregates channel detail payloads.
type DetailHandler struct {
	channelRepo  *channelrepo.ChannelMasterRepository
	alertRepo    *channelrepo.ChannelAlertRepository
	metrics      *channelservice.MetricsService
	taskService  *channelservice.TaskNoteService
	syncService  *channelservice.SyncHistoryService
	credentialSv *channelservice.CredentialService
	strategySvc  *channelservice.StrategyService
}

// ChannelDetailResponse describes the payload returned to UI.
type ChannelDetailResponse struct {
	ID           string                           `json:"id"`
	Name         string                           `json:"name"`
	Platform     string                           `json:"platform"`
	Region       string                           `json:"region"`
	Status       string                           `json:"status"`
	StoreID      string                           `json:"storeId"`
	ChannelType  string                           `json:"channelType"`
	OwnerUUID    string                           `json:"ownerUuid"`
	ApproverUUID *string                          `json:"approverUuid,omitempty"`
	Contact      ContactPayload                   `json:"contact"`
	Tags         []string                         `json:"tags"`
	Health       channelservice.HealthComputation `json:"health"`
	Metrics      []MetricDTO                      `json:"metrics"`
	Alerts       []AlertDTO                       `json:"alerts"`
	Tasks        []TaskLinkDTO                    `json:"tasks"`
	Notes        []ChannelNoteDTO                 `json:"notes"`
	SyncHistory  []SyncHistoryDTO                 `json:"syncHistory"`
	Strategy     StrategyDTO                      `json:"strategy"`
	Team         TeamDTO                          `json:"team"`
}

// MetricDTO surfaces KPI snapshots.
type MetricDTO struct {
	Window            string  `json:"window"`
	GMV               float64 `json:"gmv"`
	Orders            int64   `json:"orders"`
	GMVGrowthRate     float64 `json:"gmvGrowthRate"`
	InventoryCoverage float64 `json:"inventoryCoverage"`
	ErrorRate         float64 `json:"errorRate"`
	SyncSuccessRate   float64 `json:"syncSuccessRate"`
	HealthScore       int     `json:"healthScore"`
	SourceTimestamp   *string `json:"sourceTimestamp,omitempty"`
}

// AlertDTO for UI timeline.
type AlertDTO struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	TriggeredAt string  `json:"triggeredAt"`
	ResolvedAt  *string `json:"resolvedAt,omitempty"`
}

// TaskLinkDTO describes remediation tasks.
type TaskLinkDTO struct {
	ID         string `json:"id"`
	TaskID     string `json:"taskId"`
	TaskSource string `json:"taskSource"`
	Status     string `json:"status"`
	Note       string `json:"note,omitempty"`
	LinkedBy   string `json:"linkedBy"`
	LinkedAt   string `json:"linkedAt"`
}

// ChannelNoteDTO exposes operator notes.
type ChannelNoteDTO struct {
	ID         string `json:"id"`
	AuthorUUID string `json:"authorUuid"`
	Visibility string `json:"visibility"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
}

// SyncHistoryDTO summarized sync history row.
type SyncHistoryDTO struct {
	ID          string `json:"id"`
	TriggerType string `json:"triggerType"`
	TriggeredBy string `json:"triggeredBy"`
	Result      string `json:"result"`
	DurationMs  int64  `json:"durationMs"`
	CreatedAt   string `json:"createdAt"`
}

// NewDetailHandler builds handler.
func NewDetailHandler(deps *app.Deps, metrics *channelservice.MetricsService, taskSvc *channelservice.TaskNoteService, syncSvc *channelservice.SyncHistoryService, credSvc *channelservice.CredentialService, strategySvc *channelservice.StrategyService) *DetailHandler {
	if deps == nil || deps.DB == nil {
		return &DetailHandler{}
	}
	return &DetailHandler{
		channelRepo:  channelrepo.NewChannelMasterRepository(deps.DB),
		alertRepo:    channelrepo.NewChannelAlertRepository(deps.DB),
		metrics:      metrics,
		taskService:  taskSvc,
		syncService:  syncSvc,
		credentialSv: credSvc,
		strategySvc:  strategySvc,
	}
}

// Get returns aggregated channel detail.
func (h *DetailHandler) Get(c *gin.Context) {
	if h == nil || h.channelRepo == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "channel detail unavailable")
		return
	}
	channelID := c.Param("channelId")
	channel, err := h.channelRepo.FindByID(c.Request.Context(), channelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			contracts.ResponseNotFound(c, "channel not found")
			return
		}
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	resp := ChannelDetailResponse{
		ID:           channel.ID,
		Name:         channel.Name,
		Platform:     channel.Platform,
		Region:       channel.Region,
		Status:       channel.Status,
		StoreID:      channel.StoreID,
		OwnerUUID:    channel.OwnerUUID,
		ApproverUUID: channel.ApproverUUID,
		ChannelType:  channel.ChannelType,
		Tags:         []string(channel.Tags),
		Contact: ContactPayload{
			Name:  channel.ContactName,
			Phone: channel.ContactPhone,
			Email: channel.ContactEmail,
		},
		Team: TeamDTO{
			OwnerUUID: channel.OwnerUUID,
		},
	}
	if channel.ApproverUUID != nil {
		resp.Team.ApproverUUID = *channel.ApproverUUID
	}

	if h.metrics != nil {
		if metrics, err := h.metrics.LoadMetrics(c.Request.Context(), channelID); err == nil {
			resp.Metrics = mapMetricDTO(metrics)
			resp.Health = h.metrics.ComputeHealthScore(c.Request.Context(), channelID, metrics)
		}
	}
	if len(resp.Health.Labels) == 0 && resp.Health.Score == 0 && channel.HealthScore != nil {
		resp.Health = channelservice.HealthComputation{Score: *channel.HealthScore}
	}
	if h.alertRepo != nil {
		if alerts, err := h.alertRepo.ListByChannel(c.Request.Context(), channelID, 20); err == nil {
			resp.Alerts = mapAlertDTO(alerts)
		}
	}
	if h.taskService != nil {
		if tasks, err := h.taskService.ListTasks(c.Request.Context(), channelID); err == nil {
			resp.Tasks = mapTaskDTO(tasks)
		}
		if notes, err := h.taskService.ListNotes(c.Request.Context(), channelID, 10); err == nil {
			resp.Notes = mapNoteDTO(notes)
		}
	}
	if h.syncService != nil {
		if history, err := h.syncService.ListRecent(c.Request.Context(), channelID, 10); err == nil {
			resp.SyncHistory = mapSyncDTO(history)
		}
	}
	if h.strategySvc != nil {
		if snapshot, err := h.strategySvc.GetStrategy(c.Request.Context(), channelID); err == nil {
			dto := mapStrategySnapshot(snapshot)
			resp.Strategy = dto.Strategy
			resp.Team = dto.Team
		}
	}
	contracts.ResponseSuccess(c, resp)
}

func mapMetricDTO(metrics []*channelmodel.ChannelMetric) []MetricDTO {
	out := make([]MetricDTO, 0, len(metrics))
	for _, metric := range metrics {
		if metric == nil {
			continue
		}
		dto := MetricDTO{
			Window:            metric.Window,
			GMV:               metric.GMV,
			Orders:            metric.Orders,
			GMVGrowthRate:     metric.GMVGrowthRate,
			InventoryCoverage: metric.InventoryCoverage,
			ErrorRate:         metric.ErrorRate,
			SyncSuccessRate:   metric.SyncSuccessRate,
			HealthScore:       metric.HealthScore,
		}
		if metric.SourceTimestamp != nil {
			ts := metric.SourceTimestamp.UTC().Format(time.RFC3339)
			dto.SourceTimestamp = &ts
		}
		out = append(out, dto)
	}
	return out
}

func mapAlertDTO(alerts []*channelmodel.ChannelAlert) []AlertDTO {
	out := make([]AlertDTO, 0, len(alerts))
	for _, alert := range alerts {
		if alert == nil {
			continue
		}
		dto := AlertDTO{
			ID:          alert.ID,
			Type:        alert.Type,
			Severity:    alert.Severity,
			Title:       alert.Title,
			Description: alert.Description,
			Status:      alert.Status,
			TriggeredAt: alert.TriggeredAt.UTC().Format(time.RFC3339),
		}
		if alert.ResolvedAt != nil {
			ts := alert.ResolvedAt.UTC().Format(time.RFC3339)
			dto.ResolvedAt = &ts
		}
		out = append(out, dto)
	}
	return out
}

func mapTaskDTO(tasks []*channelmodel.ChannelTaskLink) []TaskLinkDTO {
	out := make([]TaskLinkDTO, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		out = append(out, TaskLinkDTO{
			ID:         task.ID,
			TaskID:     task.TaskID,
			TaskSource: task.TaskSource,
			Status:     task.Status,
			Note:       task.Note,
			LinkedBy:   task.LinkedBy,
			LinkedAt:   task.LinkedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func mapNoteDTO(notes []*channelmodel.ChannelNote) []ChannelNoteDTO {
	out := make([]ChannelNoteDTO, 0, len(notes))
	for _, note := range notes {
		if note == nil {
			continue
		}
		out = append(out, ChannelNoteDTO{
			ID:         note.ID,
			AuthorUUID: note.AuthorUUID,
			Visibility: note.Visibility,
			Body:       note.Body,
			CreatedAt:  note.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func mapSyncDTO(history []*channelmodel.ChannelSyncHistory) []SyncHistoryDTO {
	out := make([]SyncHistoryDTO, 0, len(history))
	for _, h := range history {
		if h == nil {
			continue
		}
		out = append(out, SyncHistoryDTO{
			ID:          h.ID,
			TriggerType: h.TriggerType,
			TriggeredBy: h.TriggeredBy,
			Result:      h.Result,
			DurationMs:  h.DurationMs,
			CreatedAt:   h.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

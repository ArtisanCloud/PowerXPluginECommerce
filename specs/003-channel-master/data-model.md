# Data Model — Channel Master Data & Authorization

## Entity: ChannelMaster
- **Table**: `channel_masters`
- **Fields**:
  - `id` (UUID, PK)
  - `tenant_uuid` (UUID, NOT NULL, indexed, participates in all unique constraints)
  - `platform` (enum: `tmall`, `jd`, `douyin`, `offline`, etc., NOT NULL)
  - `channel_type` (enum: `platform_oauth`, `platform_manual`, `offline`), 决定授权流程（OAuth、手动凭证或线下凭证）
  - `store_id` (string, NOT NULL, UNIQUE with `tenant_uuid`)
  - `name` (string, NOT NULL)
  - `domain` (string, optional, URL validation)
  - `region` (enum/ISO region code, NOT NULL)
  - `status` (enum: `draft`, `pending_review`, `rejected`, `unauthorized`, `authorized`, `disabled`)
  - `tags` (jsonb array, stores derived labels such as `data_gap`, `alerting`)
  - `owner_uuid` (UUID referencing PowerX user, NOT NULL)
  - `approver_uuid` (UUID, nullable)
  - `contact_name` / `contact_phone` / `contact_email` (nullable strings with format validation)
  - `health_score` (int 0-100, nullable snapshot)
  - `last_sync_at` (timestamp)
  - `sync_status` (enum: `ok`, `warning`, `failed`)
  - `created_by`, `updated_by`, `created_at`, `updated_at`, `deleted_at`
- **Relationships**:
  - One-to-many with `ChannelCredential`, `ChannelConfig`, `ChannelMetric`, `ChannelAlert`, `ChannelAuditLog`.
- **State transitions**:
  - `draft → pending_review → (rejected | unauthorized)` based on approval; 
  - `unauthorized → authorized` upon valid credential; 
  - `authorized → disabled` when decommissioned; 
  - Tag-based labels (`data_gap`, `alerting`) independent of `status`.
- **Indexes**: `(tenant_uuid, platform, store_id)` unique; `GIN(tags)` for label search; `status`, `region`, `owner_uuid` indexes for filters.

## Entity: ChannelCredential
- **Table**: `channel_credentials`
- **Fields**:
  - `id` (UUID, PK)
  - `tenant_uuid` (UUID, NOT NULL)
  - `channel_id` (UUID FK → ChannelMaster)
  - `type` (enum: `oauth`, `api_key`, `offline`)
  - `ciphertext` (bytea) + `dek_ciphertext` (bytea) + `algorithm` (string, default `AES-GCM`)
  - `scope` (string array / jsonb, describes permission range)
  - `expires_at` (timestamp)
  - `last_refreshed_at` (timestamp)
  - `status` (enum: `valid`, `expiring`, `expired`, `revoked`, `test_failed`)
  - `test_result` (jsonb: last connection test logs)
  - `created_at`, `updated_at`, `created_by`
- **Validation**: `expires_at` > now when status = `valid`; `scope` cannot be empty; `tenant_uuid` matches parent channel.

## Entity: ChannelConfig
- **Table**: `channel_configs`
- **Fields**:
  - `id` (UUID, PK)
  - `tenant_uuid`
  - `channel_id`
  - `pricebook_id`, `inventory_strategy_id`, `logistics_strategy_id`, `cs_sla_id` (UUID references)
  - `fee_rate` (numeric, %)
  - `settlement_cycle` (enum: `dd`, `net7`, `net30`, etc.)
  - `payment_terms` (string)
  - `notes` (text)
  - `effective_at`, `deprecated_at`
  - audit fields
- **Relationships**: one-to-one per channel (latest active row) but keep history via `effective_at`.

## Entity: ChannelMetric
- **Table**: `channel_metrics`
- **Fields**:
  - `id` (UUID PK)
  - `tenant_uuid`
  - `channel_id`
  - `window` (enum: `d1`, `d7`, `d30`)
  - Metrics: `gmv`, `orders`, `gmv_growth_rate`, `inventory_coverage`, `error_rate`, `sync_success_rate`
  - `health_score` (int)
  - `source_timestamp` (timestamp)
  - `source_job_id` (string, ETL job reference)
  - audit fields
- **Use**: read-only from ETL job; combined with ChannelMaster.health_score.

## Entity: ChannelAlert
- **Table**: `channel_alerts`
- **Fields**:
  - `id` (UUID)
  - `tenant_uuid`
  - `channel_id`
  - `type` (enum: `credential_expiring`, `credential_failed`, `data_gap`, `sync_delayed`, `kpi_threshold`)
  - `severity` (enum: `info`, `warning`, `critical`)
  - `title`, `description`
  - `triggered_at`, `resolved_at`
  - `status` (enum: `open`, `acknowledged`, `resolved`)
  - `assignee_uuid`
  - `task_id` (optional link to task center)
  - audit fields
- **Relationships**: Many alerts per channel; used by notification pipeline.

## Entity: ChannelTaskLink
- **Table**: `channel_task_links`
- **Fields**:
  - `id` (UUID)
  - `tenant_uuid`
  - `channel_id`
  - `task_id` (string, references external task center ID)
  - `task_source` (enum: `task_center`, `manual`)
  - `status` (enum: `open`, `in_progress`, `done`)
  - `note` (text)
  - `linked_by`
  - `linked_at`, `resolved_at`
  - audit fields
- **Usage**: persists FR-009 的任务关联；提供历史与审计。

## Entity: ChannelNote
- **Table**: `channel_notes`
- **Fields**:
  - `id` (UUID)
  - `tenant_uuid`
  - `channel_id`
  - `author_uuid`
  - `visibility` (enum: `team`, `admin`)
  - `body` (text/markdown)
  - `created_at`, `updated_at`
- **Usage**: 渠道运营备注，详情页展示与导出。

## Entity: ChannelSyncHistory
- **Table**: `channel_sync_history`
- **Fields**:
  - `id` (UUID)
  - `tenant_uuid`
  - `channel_id`
  - `trigger_type` (enum: `manual`, `scheduled`, `retry`)
  - `triggered_by` (UUID，系统触发则为空)
  - `duration_ms` (integer)
  - `result` (enum: `success`, `partial`, `failed`)
  - `payload_summary` (jsonb)
  - `created_at`
- **Usage**: 后端与 UI 展示最近 N 次同步记录，支持 SLA 统计与诊断。

## Entity: ChannelAuditLog
- **Table**: `channel_audit_logs`
- **Fields**:
  - `id` (UUID)
  - `tenant_uuid`
  - `channel_id` (nullable for global operations)
  - `action` (enum covering CRUD, approval, authorization, config, alert handling)
  - `actor_uuid`
  - `actor_role`
  - `context` (jsonb payload)
  - `created_at`
- **Usage**: writes from service for each state mutation + surfaced via API/export.

## Derived Structures
- **Health Score Calculation**: 
  - Hard gates: credential status, `sync_success_rate`, `error_rate` thresholds → immediate `Poor`.
  - Weighted metrics for remaining attributes stored alongside ChannelMetric row; service recalculates on demand.
- **Tags**: Derived labels persisted in `ChannelMaster.tags` to display `数据中断`, `告警中`, `巡检中` without增加状态枚举。

## Validation & Policies
- Multi-column unique constraints: `(tenant_uuid, platform, store_id)` on `channel_masters`; `(tenant_uuid, channel_id, type)` on `channel_credentials`.
- Foreign keys cascade delete disabled; soft delete through `deleted_at`.
- All tables register in `backend/cmd/database/migrate/migrate.go` and include `TableName()` returning `models.S(TableChannelMasters)` style constants.

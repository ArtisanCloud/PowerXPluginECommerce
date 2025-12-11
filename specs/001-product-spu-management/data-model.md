# Data Model

## SPU (product_spus)
- **id**: string (UUID) – 主键。
- **tenant_uuid**: string – 多租户隔离字段，所有查询必须过滤。
- **code**: string – 全局唯一 SPU 编码，可供导入/渠道识别。
- **name**: string – 默认语言标题。
- **type**: enum {one_time, subscription, bundle}。
- **category_id / category_path**: string – 关联类目。
- **brand_id**: string – 品牌。
- **default_locale**: string – 默认语言。
- **status**: enum {draft, reviewing, published, offboarded}。
- **current_version_id**: string – 指向 `product_spu_versions` 中的发布版本。
- **tags**: string[] – 自定义标签。
- **responsible_user**: string – 负责人。
- **channels_summary**: JSON – 每渠道状态/上架时间摘要，供列表展示。
- **created_at / updated_at / published_at**: timestamptz。

## SPUVersion (product_spu_versions)
- **id**: string (UUID)。
- **spu_id**: FK → SPU。
- **version_number**: int – 递增版本号。
- **status**: enum {draft, reviewing, rejected, scheduled, published, rolled_back}。
- **payload**: JSONB – 完整 SPU 表单快照（基础信息、属性、媒体、多语言文案、订阅计划、渠道配置）。
- **diff_summary**: JSONB – 字段差异摘要，供版本对比。
- **submitted_by / submitted_at**: string / timestamptz。
- **approved_by / approved_at**: string / timestamptz（最后一级）。
- **rollback_source_version_id**: string? – 若来自回滚则记录来源。
- **audit_log_id**: string – 对应审计记录。

## SPULocale (product_spu_locales)
- **id**: string。
- **spu_id**: FK → SPU。
- **locale**: string – 语言标签（zh-CN、en-US 等）。
- **title / subtitle**: string。
- **description**: rich text。
- **attributes**: JSONB – 多语言属性值。
- **status**: enum {active, missing}
- **updated_at**: timestamptz。

## ChannelVisibility (product_spu_channels)
- **id**: string。
- **spu_id**: FK → SPU。
- **channel**: enum（自定义列表，如 official_store, marketplace_x）。
- **availability**: enum {unlisted, scheduled, published, withheld}。
- **publish_at / withdraw_at**: timestamptz。
- **content_override**: JSONB – 针对该渠道的标题/价格/媒体差异。
- **audit_state**: enum {pending, approved, rejected}。
- **last_feedback**: JSON – 渠道审核回执。

## SubscriptionPlan (product_spu_subscription_plans)
- **id**: string。
- **spu_id**: FK → SPU。
- **plan_code**: string – 唯一标识。
- **name**: string。
- **billing_cycle**: enum {monthly, quarterly, yearly, custom}
- **billing_value**: int – 自定义天数或多周期标记。
- **price**: decimal(18,4)。
- **trial_days**: int?。
- **auto_renew**: boolean。
- **cancel_policy**: enum {anytime, notice_required, locked_in}。
- **effect_scope**: enum {new_only, new_and_existing} – 记录最新变更作用范围。
- **status**: enum {active, archived}。
- **metadata**: JSONB – 代扣、渠道策略等。

## SKU (product_skus) 关联字段（引用）
- **id**: string。
- **spu_id**: FK。
- **code / barcode**: string。
- **attributes**: JSON。
- **pricing_id**: string。
- **inventory_id**: string。

## ImportExportTask (product_spu_import_tasks / product_spu_export_tasks)
- **task_id**: string – 任务中心 ID。
- **tenant_uuid**: string。
- **type**: enum {import, export}。
- **template**: string – 采用的导入模板 ID。
- **success_rows / failed_rows**: int。
- **failed_report_url**: string – 失败行报告（短期签名）。
- **status**: enum {queued, running, success, failed}。
- **initiator**: string。
- **created_at / completed_at**: timestamptz。

## ApprovalRecord (product_spu_approvals)
- **id**: string。
- **spu_id / version_id**: FK。
- **chain_order**: int – 顺序（运营=1、品控=2、法务=3）。
- **role**: enum {ops, qc, legal}。
- **status**: enum {pending, approved, rejected, escalated}。
- **comment**: text。
- **sla_due_at**: timestamptz。
- **acted_by / acted_at**: string / timestamptz。

## AuditLog (product_spu_audit_logs)
- **id**: string。
- **spu_id**: FK。
- **event_type**: enum {draft_saved, submitted, approved, rejected, published, withdrawn, import, export, channel_push}。
- **payload**: JSONB – 字段变更/渠道反馈。
- **operator**: string。
- **created_at**: timestamptz。

### Relationships
- `SPU` 1↔N `SPUVersion`（当前版本指向 latest published）。
- `SPU` 1↔N `SPULocale`、`ChannelVisibility`、`SubscriptionPlan`、`SKU`。
- `SPUVersion` 1↔N `ApprovalRecord`（记录审批链）。
- `SPU` 1↔N `AuditLog`、`ImportExportTask`。
- `ChannelVisibility` 与 `SubscriptionPlan` 数据在版本 payload 中备份；发布时同步到主表。

### State Transitions
- **SPU.status**: `draft` → `reviewing` → `published` → `offboarded`，或从 `reviewing` → `draft`（被驳回）。批量下架时 `published` → `offboarded`。
- **SPUVersion.status**: `draft` → `reviewing` → (`published` 或 `rejected`)；回滚会新增 `draft` 并维系来源。
- **ChannelVisibility.availability**: `unlisted` → `scheduled` → `published`；若渠道驳回则 `published` → `withheld`。
- **ImportExportTask.status**: `queued` → `running` → (`success` or `failed`)；失败提供报告。
- **ApprovalRecord.status**: `pending` → (`approved` or `rejected`)，SLA 触发可 `escalated` 并重新指派。

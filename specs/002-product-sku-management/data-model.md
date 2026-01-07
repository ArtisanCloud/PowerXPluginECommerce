# Data Model

## Entity: ProductSKU
- **Identifiers**: `id (uuid)`, `tenant_uuid`, `sku_code` (unique per tenant), `spu_id`  
- **Attributes**: `spec_values[]`, `spec_signature`（基于 groupCode=optionCode 的稳定组合签名，用于唯一约束与前端匹配）、`barcode`, `status (draft/ready/online/offline)`, `lifecycle_phase`, `price_refs`（关联价目表 ID、阶梯价 JSON）、`min_order_qty`, `weight`, `dimensions`, `logistics (hs_code, package_type)`、`media_refs[]`, `tags[]`。  
- **Relationships**: belongs to `SPU`; has many `SkuChannels`, `SkuInventories`, `SkuMedia`, `SkuAuditLogs`。  
- **Validation Rules**: `sku_code`、`barcode` 每租户唯一；规格组合必须覆盖 SPU 的各必填规格；状态变迁需经过审批/推送（draft→ready→online/offline）。

## Entity: ProductSpecGroup
- **Identifiers**: `id (uuid)`, `tenant_uuid`, `spu_id`, `code`（SPU 内唯一）  
- **Attributes**: `name`, `sort_order`, `required`, `status (active/inactive)`  
- **Relationships**: has many `ProductSpecOption`；用于 SKU 生成器与前端规格选择渲染。  
- **Validation**: `(tenant_uuid, spu_id, code)` 唯一；`required=true` 的维度必须在每个 SKU 中出现一次。

## Entity: ProductSpecOption
- **Identifiers**: `id (uuid)`, `tenant_uuid`, `spu_id`, `group_id`, `code`（Group 内唯一）  
- **Attributes**: `name`, `sort_order`, `meta (jsonb)`, `status (active/inactive)`  
- **Relationships**: belongs to `ProductSpecGroup`；用于 SKU 组合与前端选项禁用态计算。  
- **Validation**: `(tenant_uuid, group_id, code)` 唯一；禁用项不参与新组合生成（历史 SKU 可保留）。

## Entity: ProductSkuAttribute
- **Identifiers**: `id`, `sku_id`, `spec_id`, `spec_value_id`.  
- **Attributes**: `display_order`.  
- **Relationships**: 每个 SKU 多个 attribute，约束组合唯一。  
- **Validation**: `spec_id` 绑定 `ProductSpecGroup.id`，`spec_value_id` 绑定 `ProductSpecOption.id`；同一 SKU 不可重复同一 spec。

## Entity: ProductSkuChannel
- **Identifiers**: `id`, `sku_id`, `channel_code`, `tenant_uuid`.  
- **Attributes**: `channel_sku_id`, `status (pending/published/failed/offline)`, `publish_time`, `sync_mode (inventory/push/pull)`, `last_error`, `task_id`, `price_override`, `media_override`.  
- **Relationships**: belongs to SKU；与渠道适配任务关联。  
- **Validation**: `channel_code + channel_sku_id` 唯一；失败需记录 `last_error`；状态切换需记录审计。

## Entity: ProductSkuInventory
- **Identifiers**: `sku_id`, `warehouse_id`, `tenant_uuid`.  
- **Attributes**: `available_qty`, `locked_qty`, `in_transit_qty`, `safety_stock`, `last_synced_at`, `alert_threshold`.  
- **Rules**: `available` 不得为负；`last_synced_at` 用于 SLA 监控；超过阈值触发预警记录。

## Entity: BulkTask
- **Identifiers**: `task_id (uuid)`, `tenant_uuid`.  
- **Attributes**: `task_type (price_adjustment/inventory_adjustment/channel_publish/import/export)`, `payload`（JSON diff）、`affected_count`, `status (pending/approved/running/succeeded/failed/cancelled)`, `approval_required (bool)`, `approval_state`, `submitted_by`, `approved_by`, `error_report_url`, `audit_log_id`.  
- **Relationships**: 链接多个 SKU（通过关联表），与渠道/导入模块复用。  
- **Rules**: 根据规则引擎设置 `approval_required`；状态机必须遵循 pending→(approved/rejected)→running→(succeeded/failed)。

## Entity: SkuMedia
- **Identifiers**: `id`, `sku_id`.  
- **Attributes**: `media_type (image/video)`, `url`, `is_primary`, `channel_override`, `sort_order`.  
- **Rules**: 每 SKU 至少一张主图；channel 特定媒体与主图独立。

## Entity: AuditLog
- **Identifiers**: `id`, `tenant_uuid`.  
- **Attributes**: `actor_id`, `action (generate_sku/bulk_update/channel_publish/etc)`, `target_type`, `target_id`, `payload_snapshot`, `result`, `created_at`.  
- **Rules**: 必须可追踪到任务/渠道操作；用于审批与合规。

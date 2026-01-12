# Phase 1 Data Model — SKU Inventory Stock（数据模型草案）

> 目标：支撑 `spec.md` 的 Phase 1（SKU 库存可写闭环 + 快照读取 + 上架前置校验的库存部分）。所有表需满足宪章：`tenant_uuid` + RLS；表名常量集中在 `backend/internal/entity/models/model.go`；模型放置于 `backend/internal/entity/models/product_sku`。

## 1. 实体与关系（MVP）

```text
SKU 1 ── N ProductSKUInventory（按 warehouse_id 维度）
SKU 1 ── N ProductSKUAuditLog（记录 inventory.adjust 等审计事件）
```

## 2. 字段字典（对齐现有模型）

### 2.1 ProductSKUInventory（`product_sku_inventories`）

现有模型：`backend/internal/entity/models/product_sku/inventory.go`。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| sku_id | uuid | NOT NULL, index | SKU 主键 |
| warehouse_id | varchar(64) | NOT NULL | 仓库维度；MVP 固定 `default` |
| available_qty | bigint | NOT NULL, default 0 | 可用库存（MVP 的可售判断口径） |
| locked_qty | bigint | NOT NULL, default 0 | 锁定库存（MVP 允许为 0） |
| in_transit_qty | bigint | NOT NULL, default 0 | 在途库存（MVP 允许为 0） |
| safety_stock | bigint | NOT NULL, default 0 | 安全库存（MVP 不参与可售计算） |
| alert_threshold | bigint | NOT NULL, default 0 | 告警阈值（MVP 不强制使用） |
| last_synced_at | timestamptz | NULL | 同步时间（MVP 可空） |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

唯一性建议（为实现“每 SKU 每仓一条 current 记录”）：
- 建议使用唯一约束：`unique(tenant_uuid, sku_id, warehouse_id)`（现有 gorm tag 仅是 index，需要在实现时升级为 uniqueIndex 或迁移补约束）。

### 2.2 ProductSKUAuditLog（`product_sku_audit_logs`）

表名常量已存在：`models.TableProductSkuAuditLogs`，但当前缺少具体模型声明。MVP 建议补齐模型以记录库存调整审计事件。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| sku_id | uuid | NOT NULL, index | SKU 主键 |
| warehouse_id | varchar(64) | NOT NULL | 仓库维度（default） |
| action | text | NOT NULL | `inventory.adjust` |
| actor | text | NULL | 操作者（用户标识/账号） |
| delta | bigint | NOT NULL | 变更量（可正可负） |
| before_available_qty | bigint | NOT NULL | 调整前可用库存 |
| after_available_qty | bigint | NOT NULL | 调整后可用库存 |
| request_id | text | NULL | 请求追踪 |
| created_at | timestamptz | NOT NULL | 记录时间 |

索引建议：
- `idx_product_sku_audit_logs_lookup (tenant_uuid, sku_id, created_at desc)`

## 3. 可售与校验依赖

- 可售口径（MVP）：`available_qty > 0`（仅 default 仓）。
- 校验点：SPU 发布、渠道上架/同步发布。


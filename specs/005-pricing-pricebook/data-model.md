# Phase 1 Data Model — Pricebook（数据模型草案）

> 目标：支撑 `spec.md` 的 Phase 1（主数据 + 版本发布 + 查价）。所有表需满足宪章：`tenant_uuid` + RLS；表名常量集中在 `backend/internal/entity/models/model.go`；模型放置于 `backend/internal/entity/models/pricing`（建议目录名 `pricing`）。

## 1. 实体与关系

```text
Pricebook 1 ── N PricebookVersion 1 ── N PricebookItem
   │
   └── N PricebookScope
```

- **Pricebook**：价目表主档（币种、类型、状态、当前生效版本指针）。
- **PricebookVersion**：可发布的版本（生效/失效时间、发布信息、状态）。
- **PricebookItem**：版本下 SKU 价格条目。
- **PricebookScope**：价目表适用范围维度（一期：channel/customer_group/supplier）。
- **PricebookAuditLog**：关键审计事件（一期：create/update/publish/archive）。

## 2. 字段字典（建议口径）

### 2.1 Pricebook（`pricebooks`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| code | text | NOT NULL, unique(tenant_uuid,code) | 业务唯一键（如 `base`） |
| name | text | NOT NULL | 名称 |
| type | text | NOT NULL | 一期枚举：`sales` / `purchase`（默认 `sales`） |
| currency | text | NOT NULL | ISO 4217（大写） |
| description | text | NULL | 描述 |
| status | text | NOT NULL | `active` / `archived` |
| current_version_id | uuid | NULL | 当前发布版本指针（可为空） |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

索引（建议）：
- `idx_pricebooks_tenant_type_currency_status (tenant_uuid, type, currency, status)`

### 2.2 PricebookVersion（`pricebook_versions`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| pricebook_id | uuid | NOT NULL, index | 归属 pricebook |
| version | int | NOT NULL, unique(tenant_uuid,pricebook_id,version) | 递增版本号 |
| state | text | NOT NULL | `draft` / `active` / `archived` / `expired` |
| effective_at | timestamptz | NOT NULL | 生效时间 |
| expires_at | timestamptz | NULL | 失效时间（空=永久） |
| published_at | timestamptz | NULL | 发布时间 |
| published_by | text | NULL | 发布人（user id） |
| note | text | NULL | 发布说明 |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

状态约束（来自 Clarifications）：
- 发布新版本时，自动终止旧版本有效期，保证同一 pricebook 任一时刻最多一个版本可命中。

索引（建议）：
- `idx_pricebook_versions_tenant_pricebook_state (tenant_uuid, pricebook_id, state)`
- `idx_pricebook_versions_effective (tenant_uuid, effective_at, expires_at)`

### 2.3 PricebookScope（`pricebook_scopes`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| pricebook_id | uuid | NOT NULL, index | 归属 pricebook |
| dimension | text | NOT NULL | 一期枚举：`channel` / `customer_group` / `supplier` |
| dimension_id | uuid | NOT NULL | 对应维度主数据 id |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

唯一约束：
- `unique(tenant_uuid, pricebook_id, dimension, dimension_id)`

索引（建议）：
- `idx_pricebook_scopes_lookup (tenant_uuid, dimension, dimension_id)`

### 2.4 PricebookItem（`pricebook_items`）

金额字段统一使用 **minor units**（`bigint`），避免 float 误差。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| pricebook_id | uuid | NOT NULL, index | 归属 pricebook |
| version_id | uuid | NOT NULL, index | 归属版本 |
| sku_id | uuid | NOT NULL, index | SKU 主键 |
| base_amount_minor | bigint | NULL | 基准价 |
| sale_amount_minor | bigint | NULL | 销售价/渠道价（一期成交价优先来源） |
| msrp_amount_minor | bigint | NULL | 建议零售价 |
| cost_amount_minor | bigint | NULL | 成本价 |
| min_amount_minor | bigint | NULL | 最低价 |
| max_amount_minor | bigint | NULL | 最高价 |
| tax_included | boolean | NOT NULL | 含税标记（一期可默认 false） |
| meta | jsonb | NULL | 扩展字段 |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

唯一约束：
- `unique(tenant_uuid, version_id, sku_id)`

索引（建议）：
- `idx_pricebook_items_tenant_version_sku (tenant_uuid, version_id, sku_id)`
- `idx_pricebook_items_tenant_pricebook_sku (tenant_uuid, pricebook_id, sku_id)`

### 2.5 PricebookAuditLog（`pricebook_audit_logs`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| resource_type | text | NOT NULL | `pricebook` / `pricebook_version` / `pricebook_item` |
| resource_id | uuid | NOT NULL, index | 资源 id |
| action | text | NOT NULL | `create` / `update` / `publish` / `archive` |
| actor | text | NULL | 操作人 |
| payload | jsonb | NULL | 摘要信息（条目数、变更字段等） |
| created_at | timestamptz | NOT NULL | 记录时间 |

## 3. 查价算法落库依赖点

为满足确定性选择（具体度优先 → 显式优先级 → 最近发布优先），数据侧可预留：
- `pricebooks.priority`（int，默认 0）用于显式优先级；若一期不实现，可在后续迁移补字段。
- `pricebook_versions.published_at` 作为“最近发布”排序依据。

## 4. RLS 与租户隔离（实现要点）

- 所有表必须包含 `tenant_uuid` 并建立索引；写入必须从租户上下文获得 `tenant_uuid`。
- Repository 使用 `BeginTenantTx/WithTenantTx` 并 `SET LOCAL app.tenant_uuid`（按宪章约束）。


# Phase 1 Data Model — 自营下单（Mini-app + Admin）MVP（数据模型草案）

> 目标：支撑 `specs/007-order-mini-app/spec.md` 的自营下单闭环（创建订单/库存锁定/后台查询与取消/审计）。所有新表需满足宪章：`tenant_uuid` + RLS；表名常量集中在 `backend/internal/entity/models/model.go`；模型放置于 `backend/internal/entity/models/order`。

## 1. 实体与关系（MVP）

```text
Customer 1 ── N Order
Order    1 ── N OrderItem
Order    1 ── N OrderEvent
SKU      1 ── N OrderItem

（库存锁定不新增 reservation 表，复用 ProductSKUInventory.locked_qty）
```

## 2. 字段字典（建议）

### 2.1 Order（`orders`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| order_no | varchar(64) | NOT NULL, unique(tenant_uuid, order_no) | 人可读订单号（租户内唯一） |
| customer_id | uuid | NOT NULL, index | 客户标识 |
| channel | varchar(64) | NOT NULL, index | 渠道标识（用于可售/定价口径） |
| status | varchar(32) | NOT NULL, index | `draft / pending_payment / paid / cancelled` |
| currency | varchar(8) | NOT NULL | 币种（如 `CNY`） |
| subtotal_amount | bigint | NOT NULL, default 0 | 小计金额（分） |
| total_amount | bigint | NOT NULL, default 0 | 总金额（分） |
| price_snapshot | jsonb | NULL | 价格/金额快照（提交时价） |
| sellability_snapshot | jsonb | NULL | 可售判断快照（用于追溯） |
| created_by_type | varchar(16) | NOT NULL | `customer/admin/system` |
| created_by | varchar(128) | NULL | 操作者标识（customer_id 或 admin 用户标识等） |
| created_at/updated_at/deleted_at | timestamptz | - | 软删除与审计 |

索引建议：
- `idx_orders_lookup (tenant_uuid, customer_id, created_at desc)`
- `idx_orders_admin_list (tenant_uuid, status, created_at desc)`

### 2.2 OrderItem（`order_items`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| order_id | uuid | NOT NULL, index | 订单主键 |
| sku_id | uuid | NOT NULL, index | SKU 主键 |
| qty | bigint | NOT NULL | 数量（整数件） |
| unit_price | bigint | NOT NULL | 单价（分） |
| line_amount | bigint | NOT NULL | 行金额（分） |
| price_source | varchar(64) | NULL | 价格来源标识（MVP 可空） |
| created_at | timestamptz | NOT NULL | 创建时间 |

唯一性建议：
- `unique(tenant_uuid, order_id, sku_id)`（同一订单同一 SKU 不重复行；若允许同 SKU 多行则取消该约束）

### 2.3 OrderEvent（`order_events`）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uuid | PK | 主键 |
| tenant_uuid | uuid | NOT NULL, index | 租户 UUID |
| order_id | uuid | NOT NULL, index | 订单主键 |
| event_type | varchar(64) | NOT NULL, index | `order.created / order.cancelled / ...` |
| operator_type | varchar(16) | NOT NULL | `customer/admin/system` |
| operator | varchar(128) | NULL | 操作者标识 |
| payload | jsonb | NULL | 事件上下文（例如取消原因、请求ID、库存变更摘要） |
| created_at | timestamptz | NOT NULL | 事件发生时间 |

索引建议：
- `idx_order_events_list (tenant_uuid, order_id, created_at desc)`

## 3. 幂等记录（复用既有基础设施）

创建订单幂等建议复用既有表：`integration_idempotency_records`（模型：`backend/internal/entity/models/integration/idempotency_record.go`）。

建议字段使用方式：
- `key`: 客户端提供的幂等键（建议从 `Idempotency-Key` header 读取）
- `tenant_uuid`: 从租户上下文注入
- `scope`: 建议使用 `order`
- `operation`: 建议使用 `create`
- `payload_hash`: 对请求体做 hash，避免“同 key 不同 payload”带来的语义歧义
- `response_data`: 存储创建订单的响应快照（order_id/order_no/status/amounts）

## 4. 状态机与约束（MVP）

- 状态集：`draft`（可选，仅后台草稿）、`pending_payment`、`paid`（可选：模拟）、`cancelled`
- 可取消：仅 `pending_payment`
- 取消副作用：释放库存锁定（`locked_qty -= qty`，下限为 0）并写 `order.cancelled` 事件

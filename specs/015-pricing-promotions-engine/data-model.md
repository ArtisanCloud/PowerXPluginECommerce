# Phase 1 Data Model — 促销规则引擎（订单自动促销）

## 1. 实体关系

```text
PromotionCampaign 1 ── N PromotionAuditLog
PromotionCampaign N ── N OrderPromotionSnapshot (via applied_promotions json)
Order 1 ── 1 OrderPromotionSnapshot
```

## 2. 实体定义

### 2.1 PromotionCampaign（促销规则）

- **Purpose**: 运营配置的自动促销规则，参与订单结算试算。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `code` (string, unique within tenant)
  - `name` (string)
  - `description` (string, nullable)
  - `promotion_type` (`amount_off` | `percent_off`)
  - `condition_rule` (json)
  - `scope_rule` (json)
  - `action_rule` (json)
  - `stacking_rule` (json)
  - `valid_from`, `valid_to` (timestamp with timezone)
  - `status` (`draft` | `active` | `paused` | `expired`)
  - `created_by`, `updated_by` (string)
  - `created_at`, `updated_at`, `deleted_at` (timestamp)
- **Validation Rules**:
  - `code` 在同一 `tenant_uuid` 下唯一，软删除记录不参与冲突。
  - `valid_from <= valid_to`，且结算判定包含 `valid_to` 边界。
  - `promotion_type=amount_off` 时 `action_rule.discount_amount_minor > 0`。
  - `promotion_type=percent_off` 时 `0 < action_rule.discount_percent_bps <= 10000`。
  - `scope_rule.scope_type=sku` 时 `sku_ids` 不为空。
  - `stacking_rule.priority` 必须存在，数字越小越先计算。

### 2.2 PromotionConditionRule（促销条件 JSON）

- **Purpose**: 判断订单是否达到促销门槛。
- **Fields**:
  - `min_order_amount_minor` (int64, optional, default 0)
- **Validation Rules**:
  - 金额必须 `>= 0`。
  - 金额币种由订单 `currency` 决定，MVP 不做跨币种换算。

### 2.3 PromotionScopeRule（促销范围 JSON）

- **Purpose**: 判断促销适用的 SKU 与渠道范围。
- **Fields**:
  - `scope_type` (`all` | `sku`)
  - `sku_ids` ([]string)
  - `channels` ([]string)
- **Validation Rules**:
  - `scope_type=all` 时 `sku_ids` 可为空。
  - `scope_type=sku` 时至少包含一个 SKU。
  - `channels` 为空表示全渠道。

### 2.4 PromotionActionRule（促销动作 JSON）

- **Purpose**: 定义促销命中后的优惠金额计算方式。
- **Fields**:
  - `discount_amount_minor` (int64, for `amount_off`)
  - `discount_percent_bps` (int, for `percent_off`, 8500 means 85%)
  - `max_discount_minor` (int64, optional, for `percent_off`)
- **Validation Rules**:
  - 满减金额不得大于可优惠订单金额；计算结果不足时按订单可优惠金额封顶。
  - 折扣金额按订单促销前金额计算，并受 `max_discount_minor` 限制。
  - 任何促销计算结果不得使订单金额小于 0。

### 2.5 PromotionStackingRule（促销叠加策略 JSON）

- **Purpose**: 控制多促销之间、促销与优惠券之间的叠加/互斥。
- **Fields**:
  - `priority` (int)
  - `stackable` (bool)
  - `stackable_with_coupon` (bool)
  - `exclusion_group` (string, nullable)
- **Validation Rules**:
  - 同一互斥组多条促销命中时，只应用优惠金额最大的一条。
  - `stackable=false` 的促销命中后，后续同层促销不再叠加。
  - `stackable_with_coupon=false` 时，后续优惠券必须被拒绝并返回 `promotion_excludes_coupon`。

### 2.6 OrderPromotionSnapshot（订单促销快照）

- **Purpose**: 固化订单创建时的促销计算结果，支撑客服解释、售后退款和财务对账。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `order_id` (uuid/string)
  - `currency` (string)
  - `base_total_minor` (int64)
  - `promotion_discount_minor` (int64)
  - `after_promotion_total_minor` (int64)
  - `applied_promotions` (json)
  - `rejected_promotions` (json)
  - `line_allocations` (json)
  - `priced_at` (timestamp with timezone)
  - `created_at`, `updated_at` (timestamp)
- **Validation Rules**:
  - `tenant_uuid + order_id` 唯一。
  - `after_promotion_total_minor = base_total_minor - promotion_discount_minor`。
  - `after_promotion_total_minor >= 0`。
  - `sum(line_allocations.discount_minor) = promotion_discount_minor`。
  - 快照创建后不随促销规则编辑、停用或删除重新计算覆盖。

### 2.7 PromotionAuditLog（促销审计记录）

- **Purpose**: 记录促销配置变更、状态变更、复制、订单命中和跳过原因。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `promotion_id` (uuid)
  - `order_id` (uuid/string, nullable)
  - `action` (`create` | `update` | `activate` | `pause` | `clone` | `quote_applied` | `quote_rejected`)
  - `action_reason` (string, nullable)
  - `request_id` (string)
  - `created_by` (string)
  - `payload` (json)
  - `created_at` (timestamp)
- **Validation Rules**:
  - 促销配置和状态变更必须写审计。
  - 订单命中可按订单级聚合写入，避免单次 quote 为每个拒绝原因产生过多行。

## 3. 状态流转

### 3.1 PromotionCampaign

- `draft -> active`（规则完整且有效期合法）
- `draft -> paused`（允许保存后暂停，不参与结算）
- `active -> paused`（运营手动停用）
- `paused -> active`（重新启用，必须重新校验规则）
- `draft|active|paused -> expired`（系统或查询侧按时间识别过期）
- `active|paused|expired -> draft` 不允许

### 3.2 状态参与结算规则

- 只有 `status=active` 且 `valid_from <= submitted_at <= valid_to` 的促销参与订单计算。
- `draft`、`paused`、`expired` 必须返回明确拒绝原因或不进入候选集。

## 4. 快照 JSON 建议

### 4.1 applied_promotions

```json
[
  {
    "promotion_id": "promo-1",
    "code": "ORDER1000",
    "name": "满100减10",
    "promotion_type": "amount_off",
    "discount_minor": 1000,
    "priority": 100,
    "exclusion_group": "order_discount",
    "stackable_with_coupon": true
  }
]
```

### 4.2 rejected_promotions

```json
[
  {
    "promotion_id": "promo-2",
    "code": "SKU_ONLY",
    "reason": "scope_mismatch"
  }
]
```

### 4.3 line_allocations

```json
[
  {
    "line_id": "line-1",
    "sku_id": "sku-1",
    "base_amount_minor": 10000,
    "promotion_discount_minor": 1000,
    "after_promotion_amount_minor": 9000
  }
]
```

## 5. 索引建议

- **Unique**:
  - `promotion_campaigns (tenant_uuid, code)` where `deleted_at is null`
  - `order_promotion_snapshots (tenant_uuid, order_id)`
- **Indexes**:
  - `promotion_campaigns (tenant_uuid, status, valid_from, valid_to)`
  - `promotion_campaigns (tenant_uuid, promotion_type, status)`
  - `promotion_audit_logs (tenant_uuid, promotion_id, created_at)`
  - `promotion_audit_logs (tenant_uuid, order_id, created_at)`
  - `order_promotion_snapshots (tenant_uuid, priced_at)`

## 6. RLS 与迁移约束

- 所有表必须部署在 `powerx_plugin_base` schema。
- 所有表必须包含 `tenant_uuid` 并启用 RLS。
- Go 模型放在 `backend/internal/entity/models/promotion`。
- 表名常量集中注册到 `backend/internal/entity/models/model.go`，`TableName()` 使用 `models.S(...)`。
- AutoMigrate 注册到 `backend/cmd/database/migrate/migrate.go`。

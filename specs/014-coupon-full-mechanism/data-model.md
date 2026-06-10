# Phase 1 Data Model — 完整优惠券机制

## 1. 实体关系

```text
CouponTemplate 1 ── N CouponAsset 1 ── N CouponUsageLog
                              └── N OrderCouponSnapshot (via order_id)
```

## 2. 实体定义

### 2.1 CouponTemplate（券模板）
- **Purpose**: 定义优惠规则与叠加策略。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `code` (string, unique within tenant)
  - `name` (string)
  - `coupon_type` (`amount` | `percent`)
  - `threshold_rule` (json)
  - `scope_rule` (json: channel/category/spu/sku/customer_group)
  - `stacking_rule` (json: stackable, exclusion_group, priority, level)
  - `refund_rule` (json: `no_refund`/`full_refund`/`proportional_refund`)
  - `valid_from`, `valid_to` (timestamp)
  - `status` (`draft` | `active` | `inactive`)
- **Validation Rules**:
  - `valid_from <= valid_to`
  - `priority` 在模板作用域内必须可比较且确定

### 2.2 CouponAsset（用户券资产）
- **Purpose**: 可被订单消费的券实例。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `template_id` (uuid)
  - `user_id` (string/uuid)
  - `coupon_code` (string)
  - `status` (`available` | `reserved` | `redeemed` | `refunded` | `expired`)
  - `reserved_order_id` (uuid, nullable)
  - `reserved_at`, `redeemed_at`, `refunded_at`, `expired_at` (timestamp, nullable)
  - `meta` (json)
- **Validation Rules**:
  - `reserved_order_id` 仅在 `reserved` 状态非空
  - 同一时刻一张券只允许一个 `reserved_order_id`

### 2.3 CouponUsageLog（券动作流水）
- **Purpose**: 记录券生命周期动作，供审计与对账。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `asset_id` (uuid)
  - `order_id` (uuid, nullable)
  - `action` (`reserve` | `release` | `redeem` | `refund`)
  - `action_reason` (string)
  - `idempotency_key` (string)
  - `request_id` (string)
  - `created_by` (string)
  - `created_at` (timestamp)
- **Validation Rules**:
  - 同一 `tenant_uuid + action + idempotency_key` 必须唯一

### 2.4 OrderCouponSnapshot（订单优惠快照）
- **Purpose**: 固化订单结算时优惠明细。
- **Core Fields**:
  - `id` (uuid)
  - `tenant_uuid` (uuid)
  - `order_id` (uuid)
  - `currency` (string)
  - `base_total_minor`, `discount_total_minor`, `payable_total_minor` (int64)
  - `line_allocations` (json: 按订单行分摊结果)
  - `applied_coupons` (json)
  - `rejected_coupons` (json)
  - `priced_at` (timestamp)
- **Validation Rules**:
  - `payable_total_minor >= 0`
  - `sum(line_allocations.discount_minor) = discount_total_minor`

## 3. 状态流转

### 3.1 CouponAsset
- `available -> reserved`（下单提交成功）
- `reserved -> available`（支付失败/订单关闭/超时取消）
- `reserved -> redeemed`（支付成功）
- `redeemed -> refunded`（退款且模板允许返券）
- `available|reserved -> expired`（到期任务）

### 3.2 非法流转（必须拒绝）
- `redeemed -> available`
- `expired -> reserved`
- `refunded -> redeemed`

## 4. 一致性与索引建议
- **Uniq**:
  - `coupon_templates (tenant_uuid, code)`
  - `coupon_assets (tenant_uuid, coupon_code)`
  - `coupon_usage_logs (tenant_uuid, action, idempotency_key)`
  - `order_coupon_snapshots (tenant_uuid, order_id)`
- **Indexes**:
  - `coupon_assets (tenant_uuid, user_id, status, valid_to)`
  - `coupon_assets (tenant_uuid, reserved_order_id)`
  - `coupon_usage_logs (tenant_uuid, order_id, created_at)`
  - `order_coupon_snapshots (tenant_uuid, priced_at)`

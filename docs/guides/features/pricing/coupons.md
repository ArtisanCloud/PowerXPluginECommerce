# 优惠券机制（对齐实现版）

## 1. 功能范围

本功能覆盖以下链路：

- 交易侧：`/api/v1/v1/coupons/quote` 优惠试算。
- 管理侧：模板管理、批量发券、资产查询、流水查询。
- 订单支付联动：预占、核销、释放。

## 2. 业务流程

1. 运营创建模板并激活（`active`）。
2. 运营按模板向用户发券，生成 `coupon_assets`。
3. 下单前调用试算接口，返回可用券、拒绝原因、券后金额。
4. 订单创建后预占券（`available -> reserved`）。
5. 支付成功触发核销（`reserved -> redeemed`）。
6. 支付失败/取消/超时触发释放（`reserved -> available`）。

## 3. 管理端 API

### 3.1 模板管理

- `GET /api/v1/admin/coupons/templates`
- `POST /api/v1/admin/coupons/templates`
- `PATCH /api/v1/admin/coupons/templates/:id`

### 3.2 发券

- `POST /api/v1/admin/coupons/issues`

请求示例：

```json
{
  "template_id": "tpl-1",
  "user_ids": ["u-1", "u-2"],
  "quantity_per_user": 2
}
```

### 3.3 资产与流水查询

- `GET /api/v1/admin/coupons/assets`
- `GET /api/v1/admin/coupons/usage-logs`

常用查询参数：`couponCode`、`orderId`、`userId`、`status`。

## 4. 状态机

- 资产状态：`available`、`reserved`、`redeemed`、`refunded`、`expired`
- 合法迁移：
  - `available -> reserved`
  - `reserved -> redeemed`
  - `reserved -> available`
  - `redeemed -> refunded`

## 5. 排障指引

### 5.1 支付成功但券未核销

检查：

- 支付回调是否已落库并进入 `paid` 分支。
- `coupon_assets` 是否存在 `reserved_order_id = 订单ID` 且 `status = reserved`。
- `coupon_usage_logs` 是否有 `action=redeem` 记录。

### 5.2 取消订单后券未释放

检查：

- 订单取消链路是否执行释放服务。
- 资产是否仍被其他订单占用。
- `coupon_usage_logs` 是否有 `action=release` 记录。

### 5.3 重复回调导致重复核销

检查：

- `coupon_usage_logs` 唯一键 `(tenant_uuid, action, idempotency_key)` 是否生效。
- 幂等键是否按 `action:orderId:assetId` 生成。

## 6. 观测与告警

- 指标：发放、预占、核销、释放、返券总量与失败总量。
- 告警：
  - 核销失败积压（`coupon.redeem_failure_backlog`）
  - 超时未释放（`coupon.timeout_unreleased`）

## 7. 验证命令

```bash
cd backend
mkdir -p ../tmp/gocache ../tmp/gomodcache
GOCACHE=$PWD/../tmp/gocache GOMODCACHE=$PWD/../tmp/gomodcache go test ./internal/services/admin/coupon ./internal/transport/http/admin/coupon
```

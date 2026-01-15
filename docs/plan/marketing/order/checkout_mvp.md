# 自营下单 MVP（Admin 代客下单 + Mini-app 下单 + 后台订单管理）

> 目标：在“one_time（一次性商品）”范围内，交付最小可用的下单闭环：可售校验 → 创建订单 → 锁库存 → 后台可查询/管理。订阅（会籍/权益）与第三方渠道订单同步先不纳入本 MVP。

## 1. 范围与非目标

### 1.1 范围（MVP）
- **Mini-app 下单**：客户在小程序对 `sellable=true` 的 SKU 下单。
- **Mini-app 购物车（混合模式）**：本地购物车 + 服务端购物车 + 登录后自动同步（多端合并），用于支持“加入购物车 → 结算下单”路径。
- **Web-admin 代客下单**：运营/客服在后台为客户创建订单（同样走 sellability/价格快照/库存锁定）。
- **订单管理**：后台可查看订单列表/详情、取消订单、查看操作日志。

### 1.2 非目标（后续）
- 订阅下单产生会籍/权益（membership entitlement）。
- 支付真实对接（可先做“模拟支付成功”或仅停留在 `pending_payment`）。
- 复杂履约（拆单、多仓、运费模板、发票税务）、优惠叠加、售后/退款联动。
- 第三方渠道订单同步（属于 `docs/plan/marketing/order/README.md` 的渠道订单主题）。
- 购物车“库存预占/保留”（cart 不锁库存，库存只在创建订单时锁定）。
- 购物车跨账号共享（仅同一 customer 的多端同步）。

## 2. 统一口径：可下单门槛

下单入口必须复用同一门槛口径，避免前后端分散判断：

- 以 `GET /api/v1/mini-app/products/{spuId}/sellability?channel=...` 的规则作为唯一可售判断来源（one_time）。
- 下单时服务端必须**二次校验** sellability（防止前端缓存/篡改）。
- 库存口径：`availableQty = SUM(available_qty - locked_qty)`（MVP 直接写 `locked_qty`）。

## 3. 数据模型（建议）

### 3.1 核心表
- `orders`：订单主表（tenant_uuid、customer_id、channel、status、currency、amounts、snapshot、created_at...）
- `order_items`：订单明细（order_id、sku_id、qty、unit_price、line_amount、price_source...）
- `order_events`：订单事件/审计（order_id、event_type、operator、payload、created_at）
- `carts`：购物车（tenant_uuid、customer_id、items(jsonb)、updated_at...；同一客户仅一条记录）

### 3.2 可选表（后续）
- `order_payments`、`order_fulfillments`、`order_refunds`、`inventory_reservations`
- `cart_items`（若需要按条目追踪更新时间/多活动规则，可拆表）

## 4. 状态机（MVP）

- `draft`（仅 admin 可见/可编辑）
- `pending_payment`（已创建待支付）
- `paid`（可选：模拟）
- `cancelled`

> 先不引入发货/签收/售后状态，避免把履约链路一次性做全。

## 5. API（建议）

### 5.1 Mini-app
- `POST /api/v1/mini-app/orders`
  - body：`{ channel, items:[{ skuId, qty }], locale? }`
  - response：`{ orderId, status, amounts, createdAt }`
- `GET /api/v1/mini-app/orders`
- `GET /api/v1/mini-app/orders/{id}`

#### 5.1.1 购物车（混合模式）

目标：在“本地购物车”与“服务端购物车”之间提供最小同步机制，支持多端登录后的合并与一致展示。

- `GET /api/v1/mini-app/cart`
  - response：`{ items:[{ skuId, qty }], updatedAt }`
- `POST /api/v1/mini-app/cart/sync`
  - body：`{ items:[{ skuId, qty }], strategy?: "max" | "overwrite" }`
  - response：`{ items:[{ skuId, qty }], updatedAt }`

同步策略（默认 `max`）：
- `max`：同一 SKU 取 `qty = max(localQty, serverQty)`，避免多端重复同步导致数量翻倍。
- `overwrite`：直接以本地 items 覆盖服务端（适用于用户明确“以本机为准”的场景）。

约束：
- cart 不锁库存；库存与价格的权威校验只在创建订单时发生（避免“加购即锁库存”带来的复杂补偿）。

### 5.2 Admin（代客下单 + 管理）
- `POST /api/v1/admin/orders`（代客下单）
  - body：`{ customerId, channel, items:[{ skuId, qty }], note? }`
- `GET /api/v1/admin/orders`
- `GET /api/v1/admin/orders/{id}`
- `POST /api/v1/admin/orders/{id}/cancel`

## 6. 库存锁定（MVP 策略）

MVP 先用强一致、低复杂度方案：
- 创建订单时在同一事务内：
  1) 再算一遍可用库存 `available_qty - locked_qty`
  2) 满足则 `locked_qty += qty`
  3) 写入订单与明细
- 取消订单时 `locked_qty -= qty`（下限为 0）

后续再升级为 reservation 表与异步补偿，支持并发更高、可追溯锁定记录。

## 7. 与渠道订单 PRD 的关系

- 本 MVP 属于“自营订单（first-party order）”，服务于小程序/后台下单。
- `docs/plan/marketing/order/README.md` 的“渠道订单”偏第三方平台同步与运营看板，可在本 MVP 完成后并行推进，并复用订单域的列表/详情/审计能力。

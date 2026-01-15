# Quickstart — 自营下单（Mini-app + Admin）MVP

目标：用最少步骤验证“可售校验 → 创建订单 → 锁库存 → 后台查询/取消/审计”的闭环。

本 Quickstart 面向本地开发与 CI 自测；不覆盖真实支付、复杂履约、优惠叠加、售后/退款与第三方渠道订单同步。

## 1) 启动后端（本地）

- `make dev`（或 `go run ./backend/cmd/plugin`）

## 2) 预置数据

- 准备一个 `sellable=true` 且库存充足的 SKU（default 仓 `available_qty - locked_qty >= qty`）
- 准备一个 customer（用于小程序用户态请求）

## 3) 小程序创建订单（本期新增）

- `POST /api/v1/mini-app/orders`
  - header：`Idempotency-Key: <uuid>`
  - body：`{"channel":"default","items":[{"skuId":"<sku-uuid>","qty":1}]}`

预期：
- 返回 `orderId/orderNo/status/amounts/createdAt`
- 对应 SKU 的 `locked_qty` 增加
- 使用同一 `Idempotency-Key` 重试，返回同一订单结果且 `locked_qty` 不再重复增加

## 3.5) 小程序购物车同步（混合模式，本期新增）

- `GET /api/v1/mini-app/cart`
- `POST /api/v1/mini-app/cart/sync`
  - body：`{"items":[{"skuId":"<sku-uuid>","qty":2}],"strategy":"max"}`

预期：
- 未登录调用返回 401；登录后可读写
- sync 返回合并后的 `items` 与 `updatedAt`
- 购物车不锁库存（不会改动 `locked_qty`）

## 4) 小程序查询订单

- `GET /api/v1/mini-app/orders`
- `GET /api/v1/mini-app/orders/{id}`

预期：仅可查询到本人订单，且金额/状态与创建时一致。

## 5) 后台代客下单（本期新增）

- `POST /api/v1/admin/orders`
  - header：`Idempotency-Key: <uuid>`
  - body：`{"customerId":"<customer-uuid>","channel":"default","items":[{"skuId":"<sku-uuid>","qty":1}]}`

预期：创建订单成功、库存锁定成功、幂等行为与小程序一致。

## 6) 后台订单管理（本期新增）

- `GET /api/v1/admin/orders`
- `GET /api/v1/admin/orders/{id}`
- `POST /api/v1/admin/orders/{id}/cancel`

预期：
- 取消仅允许 `pending_payment`，成功后状态变为 `cancelled`
- 取消成功后释放库存锁定（`locked_qty` 下降，且不应为负）
- 订单事件（创建/取消）可在详情中被查询到

## 7) RBAC（本期新增）

建议最小权限集合：
- 创建订单：`order:create`
- 读取订单：`order:read`
- 取消订单：`order:cancel`

## 8) 本地验证记录（建议执行）

- 后端：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./...`
- 后端 lint：`cd backend && GOLANGCI_LINT_CACHE=$PWD/.cache/golangci-lint ./bin/golangci-lint run --timeout 5m`

### 已执行记录

- 2026-01-13：`cd backend && go test ./...`
- 2026-01-13：`cd web-admin && npm run test:ci`

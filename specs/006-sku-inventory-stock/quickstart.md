# Quickstart — SKU Inventory Stock（MVP）

目标：用最少步骤验证“库存可写闭环 + 上架前置校验（库存部分）”。

本 Quickstart 面向本地开发与 CI 自测；不覆盖 WMS/OMS 同步、仓库管理、调拨、盘点等完整库存域能力。

## 1) 启动后端（本地）

- `make dev`（或 `go run ./backend/cmd/plugin`）

## 2) 读取库存快照（现有接口）

- `GET /api/v1/admin/product/skus/{skuId}/inventory`

预期：返回 `warehouses` 列表与 `summary` 汇总；MVP 以 `warehouse_id=default` 为主。

## 3) 调整可用库存（本期新增）

- `POST /api/v1/admin/product/skus/{skuId}/inventory/adjust`
  - body: `{"delta": 10}`

预期：返回调整后的库存快照；`summary.available_qty` 增加 10；不允许调整后为负数。

## 4) 发布/上架校验（本期新增）

在以下动作触发库存校验（MVP）：
- SPU 发布
- 渠道上架/同步发布

预期：
- 若 SPU 下所有 SKU 在 default 仓 `available_qty == 0`：操作被阻止并返回“缺少可售库存”
- 若至少 1 个 SKU `available_qty > 0`：允许继续

## 5) RBAC（本期新增）

- 读取库存快照：`product.sku.inventory:read`
- 调整库存：`product.sku.inventory:manage`

## 6) 本地验证记录（已执行）

- 后端：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./...`
- 后端 lint：`cd backend && GOLANGCI_LINT_CACHE=$PWD/.cache/golangci-lint ./bin/golangci-lint run --timeout 5m`
- web-admin：`cd web-admin && npm run test:unit`

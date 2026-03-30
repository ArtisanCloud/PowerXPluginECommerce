# SKU 库存（Stock）MVP 测试指南

## 背景与目标

本指南覆盖“SKU 可用库存（default 仓）可写闭环 + 上架/发布库存门禁（库存部分）”。

完成标准：
- 运营能为 SKU 写入/调整可用库存（delta）。
- SPU 发布与 SKU 渠道上架在无可售库存时会被阻止。

## 依赖 & 权限

### RBAC
- 读取库存快照：`com.powerx.plugins.ecommerce:product.sku.inventory:read`
- 调整库存：`com.powerx.plugins.ecommerce:product.sku.inventory:manage`
- 触发发布/上架：对应现有 SPU/SKU/channel 权限（按你当前环境配置为准）

### 可售口径（当前实现）
- 仅默认仓：`warehouse_id = "default"`
- 可售判断：`available_qty > 0`

## 测试步骤（UI）

1) 打开 `SKU 详情页`
- 路径：商品 → SKU → 进入某个 SKU 详情（`/product/skus/:id`）
- 预期：出现“库存”卡片，展示快照与汇总（Total）。

2) 写入库存（delta 调整）
- 在库存卡片输入 `delta`（如 `10`）点击“应用”
- 预期：`default` 仓 `available_qty` 增加，Total 同步变化。

3) 负库存保护
- 输入一个会导致结果为负的 `delta`（如当前为 10，输入 `-20`）
- 预期：操作失败并提示错误；库存保持不变。

## 测试步骤（API）

1) 读取库存快照
- `GET /api/v1/admin/product/skus/{skuId}/inventory`

2) 调整可用库存（增量）
- `POST /api/v1/admin/product/skus/{skuId}/inventory/adjust`
- body：`{"delta": 10}`

预期：
- `delta` 必须为非 0 整数
- 调整后不可为负

## 上架/发布门禁验证（库存部分）

### A) SPU 发布（Publish）门禁
1) 准备：保证该 SPU 下所有 SKU 在 default 仓 `available_qty == 0`（或根本没有 inventory 行）
2) 执行：发起 SPU Publish
3) 预期：发布被阻止，返回“缺少可售库存/需要库存”等语义的错误信息

然后：
1) 任选一个 SKU，执行 `inventory/adjust delta=+1`
2) 再次 Publish
3) 预期：允许发布继续（进入发布流程/返回成功）

### B) SKU 渠道上架（PublishChannelMapping）门禁
1) 准备：该 SKU default 仓 `available_qty == 0`
2) 执行：SKU 渠道上架/同步发布
3) 预期：上架被阻止（库存不足）

## 回归测试提示

- 后端：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./...`
- web-admin：`cd web-admin && npm run test:unit`

## 常见故障与排障

- `403`：检查是否缺少 `product.sku.inventory:read/manage` 或上架/发布相关权限。
- `400` 且提示 negative/required：检查 delta 是否为 0，或是否会导致 `available_qty` 变为负数。
- 发布一直失败：确认“至少 1 个 SKU default 仓 available_qty > 0”已满足；并确认该 SPU 下确实存在 SKU。


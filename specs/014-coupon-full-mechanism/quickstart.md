# Quickstart — 完整优惠券机制（结算复算 + 支付核销）

## 1. 目标

验证“模板配置、发券、试算、预占、核销、释放、返券、审计”是否形成完整闭环。

## 2. 前置准备

- 准备 1 个测试租户、1 个测试用户、2 个可售 SKU。
- 准备 1 条固定减券模板（订单级）与 1 条商品级折扣券模板。
- 确认订单支付超时策略已配置（用于验证自动释放）。

## 3. 核心流程

1. 创建并启用券模板。
2. 向测试用户发放 2 张券（商品级 + 订单级）。
3. 调用结算试算接口，确认：
- 计算顺序为“先商品级后订单级”；
- 返回已应用券与不可用券原因。
4. 提交订单并携带券，确认券资产状态变为 `reserved`。
5. 触发支付成功，确认状态变为 `redeemed`，并生成核销流水。
6. 触发支付失败/订单关闭场景，确认 `reserved` 券被释放回 `available`。
7. 触发支付退款，确认返券行为符合模板策略（默认不返券；模板开启返券时 `redeemed -> refunded` 并生成 `refund` 流水）。
8. 查询订单优惠快照，确认订单行分摊金额与总优惠一致。

## 4. 验收检查点

- 金额一致：`base_total + shipping + tax - discount_total = payable_total`。
- 分摊一致：订单行优惠分摊和等于订单总优惠。
- 幂等一致：重复支付回调不产生重复核销。
- 状态一致：券资产状态与订单支付状态不冲突。
- 审计可追溯：按订单号可查询完整动作链路。

## 5. 回归记录（2026-04-04）

### 5.1 执行命令

```bash
cd backend
mkdir -p ../tmp/gocache ../tmp/gomodcache
GOCACHE=$PWD/../tmp/gocache GOMODCACHE=$PWD/../tmp/gomodcache go test ./internal/services/admin/coupon ./internal/transport/http/admin/coupon ./internal/transport/http/admin/...

cd ../web-admin
npm run -s build
```

### 5.2 结果摘要

- `admin/coupon` 服务与处理器测试通过。
- 管理端全量 transport 编译/测试通过。
- 前端构建通过（存在仓库既有 chunk/cycle warning，不阻断构建）。

### 5.3 风险备注

- miniapp 路由当前实际路径为 `/api/v1/v1/coupons/quote`，契约与路由矩阵已按实现记录。
- 建议在后续 API 整理阶段统一 miniapp 子路径前缀，避免 `/v1/v1` 认知负担。

## 6. 收尾回归记录（2026-06-09）

### 6.1 执行命令

```bash
cd backend
GOCACHE=$PWD/../tmp/gocache GOMODCACHE=$PWD/../tmp/gomodcache go test ./internal/services/admin/coupon ./internal/transport/http/admin/coupon ./internal/transport/http/miniapp/coupon ./internal/services/admin/payment ./internal/services/admin/payments ./internal/services/admin/order ./tests/performance

cd ../web-admin
npm run -s build
```

### 6.2 结果摘要

- 优惠券域服务、管理端/交易端处理器、订单/支付接入与性能测试通过。
- 支付退款创建时已接入优惠券返券策略：默认不返券，模板开启返券时写入 `refund` 流水并更新资产为 `refunded`。
- 前端构建通过。

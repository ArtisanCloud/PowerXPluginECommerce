# Quickstart — 促销规则引擎（订单自动促销）

## 1. 目标

验证“促销规则配置、订单自动试算、促销与优惠券叠加/互斥、订单促销快照、审计追踪”是否形成一期闭环。

## 2. 前置准备

- 准备 1 个测试租户。
- 准备 1 个测试渠道，例如 `miniapp`。
- 准备 2 个可售 SKU，并确保 pricebook 可返回基础价。
- 如需验证与优惠券互斥，准备 1 张可用于同一订单的测试优惠券。

## 3. 核心流程

1. 在 `pricing/promotions.vue` 创建满减促销：
   - `promotion_type=amount_off`
   - `min_order_amount_minor=10000`
   - `discount_amount_minor=1000`
   - `scope_type=all`
   - `stackable_with_coupon=true`
2. 启用促销，确认列表状态变为 `active`。
3. 调用订单 quote 或促销试算接口，使用满足门槛的订单：
   - 确认返回 `base_total_minor=10000`
   - 确认返回 `promotion_discount_minor=1000`
   - 确认返回 `after_promotion_total_minor=9000`
   - 确认 `applied_promotions` 包含该促销。
4. 使用不满足门槛的订单重新试算：
   - 确认促销未应用。
   - 确认 `rejected_promotions.reason=threshold_not_met`。
5. 创建同一互斥组内的两条促销，构造同时命中的订单：
   - 确认只应用优惠金额最大的促销。
6. 创建 `stackable_with_coupon=false` 的促销并携带可用优惠券下单：
   - 确认促销保留。
   - 确认优惠券被拒绝，原因是 `promotion_excludes_coupon`。
7. 创建订单并支付前查询订单详情：
   - 确认写入 `order_promotion_snapshots`。
   - 确认订单行分摊金额之和等于促销优惠总额。
8. 修改或停用原促销后再次查询历史订单：
   - 确认历史订单促销金额、命中促销和分摊信息不变。

## 4. 验收检查点

- 金额一致：`base_total_minor - promotion_discount_minor = after_promotion_total_minor`。
- 顺序一致：基础价先算，促销后算，优惠券最后算。
- 互斥一致：同一互斥组只应用优惠金额最大的一条。
- 边界一致：订单提交时间等于促销结束时间时仍可命中。
- 分摊一致：订单行促销分摊和等于订单总促销优惠。
- 快照一致：促销后续变更不影响历史订单快照。
- 租户一致：不同租户之间促销规则、快照、审计不可见。

## 5. 建议验证命令

```bash
cd backend
go test ./internal/services/admin/promotion ./internal/transport/http/admin/promotion ./internal/transport/http/miniapp/promotion ./internal/services/admin/order

cd ../web-admin
npm run -s lint
npm run -s build
```

## 6. 回归风险

- 订单未命中促销时，既有 pricebook、coupon、payment、after-sales 流程不得发生金额偏差。
- 交易侧若先通过独立 `/v1/promotions/quote` 验证，最终仍必须接入订单 quote/下单链路，避免试算与落单金额分裂。
- `stackable_with_coupon=false` 必须传递到 coupon 计算层，否则优惠券可能错误核销。

## 7. 实施回归记录（2026-06-10）

### 7.1 执行命令

```bash
cd backend
GOCACHE=$PWD/../tmp/gocache GOMODCACHE=$PWD/../tmp/gomodcache go test ./internal/services/admin/promotion ./internal/transport/http/admin/promotion ./internal/transport/http/miniapp/promotion ./internal/services/admin/order ./cmd/database/migrate ./internal/router ./internal/transport/http ./internal/transport/http/admin ./internal/transport/http/miniapp

cd ../web-admin
npm run -s build
```

### 7.2 结果摘要

- 促销服务、管理端/交易端处理器、订单接入、迁移注册与路由注册测试通过。
- 前端 Nuxt build 通过。
- 已实现管理端促销规则配置、交易侧促销试算、订单创建促销接入、促销快照写入、与优惠券排斥原因 `promotion_excludes_coupon`。

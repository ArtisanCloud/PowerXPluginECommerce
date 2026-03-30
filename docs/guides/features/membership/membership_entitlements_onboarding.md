# 订阅会籍/权益/代币发放验证指南

本指南对应 `specs/010-membership-entitlements`，用于验证“订阅支付成功后自动发放会籍、权益与代币”的端到端链路。

## 背景与目标

- 规格来源：`specs/010-membership-entitlements/spec.md`
- 目标：
  - 支付成功后自动发放会籍/权益/代币
  - 同一交易重复回调不重复发放（幂等）
  - mini-app 与管理端都能查询到结果

## 前置条件

- 已完成数据库迁移（含 `011_membership_entitlements`）：`make migrate`
- 已配置订阅计划 metadata（至少包含）：
  - `membershipTierId`
  - `benefitIds`
  - `tokenCode`
  - `tokenAmount`
- 支付回调可正常到达插件
- 租户上下文可用（宿主鉴权上下文或 query `tenant_uuid`）

## 关键接口

- mini-app 查询：
  - `GET /api/v1/mini-app/membership/profile`
  - `GET /api/v1/mini-app/membership/entitlements`
  - `GET /api/v1/mini-app/membership/tokens`
  - `GET /api/v1/mini-app/membership/benefits`
- admin 会籍管理：
  - `GET /api/v1/admin/membership/tiers`
  - `GET /api/v1/admin/membership/benefits`
  - `POST /api/v1/admin/membership/entitlements/grant`
  - `POST /api/v1/admin/membership/entitlements/revoke`
  - `POST /api/v1/admin/membership/tokens/adjust`
- admin 客户维度查询：
  - `GET /api/v1/admin/customers/{id}/entitlements`
  - `GET /api/v1/admin/customers/{id}/tokens`

## 测试步骤

### 1) 创建订阅订单并完成支付

1. 选择已绑定会籍/权益/代币规则的订阅计划创建订单。
2. 完成支付，确保支付回调进入插件。

验证点：
- 订单进入支付成功状态。
- 日志无发放异常（计划绑定缺失、参数非法等）。

### 2) 校验 mini-app 客户视角

1. 调用 `/api/v1/mini-app/membership/profile`。
2. 调用 `/api/v1/mini-app/membership/entitlements`。
3. 调用 `/api/v1/mini-app/membership/tokens`。

验证点：
- 会籍等级已生效（`profile` 可见 tier 信息）。
- 权益实例可见（数量、有效期、服务编码符合预期）。
- 代币余额增加，且币种与订阅计划一致。

### 3) 校验 admin 运营视角

1. 在客户详情页「权益与代币」页签查看数据。
2. 或直接调 `/api/v1/admin/customers/{id}/entitlements` 与 `/api/v1/admin/customers/{id}/tokens`。

验证点：
- 与 mini-app 看到的数据一致。
- 能执行手动发放/回收/调整并落审计。

### 4) 幂等回归（重点）

1. 对同一交易重复推送支付成功回调 2~3 次。
2. 重新查询客户权益与代币。

验证点：
- 不出现重复权益实例。
- 代币账本不重复累加。

## 预期结果

- 支付成功后 60 秒内，客户可查询到会籍/权益/代币。
- 重复回调不产生重复发放。
- 管理端与 mini-app 查询结果一致。

## 常见问题

- `tenant context missing`：
  - 检查是否携带宿主租户上下文，或在直连时补 `tenant_uuid`。
- 支付成功但无发放：
  - 检查订阅计划 metadata 绑定字段是否完整。
- 数据不一致：
  - 先校验发放流水（sourceType/sourceId），再对比客户接口返回。

## 回归建议

- 后端：优先覆盖支付回调编排与幂等路径。
- 前端：覆盖 mini-app 个人中心/会员页、admin 客户详情「权益与代币」页签。
- 参考：`specs/010-membership-entitlements/quickstart.md`、`specs/010-membership-entitlements/tasks.md`。

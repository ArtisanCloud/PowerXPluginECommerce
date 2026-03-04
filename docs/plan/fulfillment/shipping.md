# 履约与物流 PRD（可开发版）

> 覆盖页面：`web-admin/app/pages/shipping/carriers.vue`、`shipping/templates.vue`、`shipping/waybills.vue`。该模块负责承运商、运费模板、运单、轨迹、异常处理，并支持第三方物流适配。

## 1. 背景与目标
现有页面为占位，需要实现可跑通的发货与轨迹流程，并预留第三方承运商接口。

**目标**
1. 管理承运商及服务：区域覆盖、服务类型、API 凭证。
2. 配置运费模板与规则，供订单/渠道计算运费。
3. 生成运单、获取轨迹、处理异常。
4. 支持第三方物流适配器。

## 2. 角色
| 角色 | 诉求 |
| --- | --- |
| 供应链/物流 | 配置承运商、监控运力、处理异常 |
| 渠道运营 | 选择运费模板、确认渠道策略 |
| 客服/售后 | 查询物流状态、处理投诉 |
| 财务 | 结算运费、对账 |

## 3. 信息架构
1. **承运商管理**
   - 列表字段：名称、类型、服务区域、服务类型、状态、联系人、费率、时效。
   - 详情：服务描述、账期、API 凭证、支持渠道、SLA。
2. **运费模板**
   - 模板列表：名称、渠道/店铺、计费方式（重量/件数/体积/区域）、币种、状态、更新时间。
   - 详情：区域分区、首重/续重、偏远费、免邮条件。
3. **运单管理**
   - 列表：运单号、订单/调拨、承运商、状态（创建/揽收/运输中/签收/异常）。
   - 详情：轨迹节点、异常、签收信息、面单。

## 4. 核心流程
1) 订单支付成功 → 选择承运商服务
2) 生成运单 → 打印面单 → 发货
3) 同步轨迹 → 更新订单状态
4) 异常处理 → 赔付/重发

## 5. 第三方承运商适配
### 5.1 适配器接口
- `CreateShipment`：生成运单
- `GetTracking`：获取轨迹
- `CancelShipment`：取消运单
- `GetLabel`：获取面单

### 5.2 配置项
- app_id / key / secret / sandbox / 回调地址 / 服务编码映射

### 5.3 支持策略
- 内置适配器：顺丰/京东物流/菜鸟（预留）
- 自定义适配器：Webhook/SDK

## 6. 数据 & API
### 6.1 数据表
- `logistics_carriers`
- `logistics_carrier_services`
- `logistics_rate_templates`
- `logistics_rate_zones`
- `logistics_waybills`
- `logistics_tracking_events`
- `logistics_exceptions`

### 6.2 API
- `GET/POST /api/v1/admin/logistics/carriers`
- `PATCH /api/v1/admin/logistics/carriers/{id}`
- `POST /api/v1/admin/logistics/carriers/{id}/test`
- `GET/POST /api/v1/admin/logistics/templates`
- `POST /api/v1/admin/logistics/templates/{id}/publish`
- `POST /api/v1/admin/logistics/waybills`
- `GET /api/v1/admin/logistics/waybills/{id}`
- `POST /api/v1/admin/logistics/waybills/{id}/track`

## 7. 权限 & 审计
- 权限：`logistics.carriers`、`logistics.templates`、`logistics.waybills`
- 审计：承运商配置、模板变更、运单异常处理、结算调整记录

## 8. KPI
- 运单准时率 ≥ 95%
- 异常处理时效 < 24h
- 运费结算准确率 ≥ 99%

## 9. 风险
- 承运商 API 不稳定 → 降级/多承运商路由
- 运费模板复杂 → 版本 + 校验

## 10. Backlog
- 承运商自动选择（成本/时效）
- 物流 IoT 数据接入
- 与 TMS 深度集成


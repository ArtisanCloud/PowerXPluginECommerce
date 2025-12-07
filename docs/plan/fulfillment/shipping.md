# 履约与物流 PRD

> 覆盖页面：`web-admin/app/pages/shipping/carriers.vue`、`shipping/templates.vue`、`shipping/waybills.vue`，以及订单/售后/渠道中引用的物流组件。该模块负责承运商、运费模板、运单、轨迹、物流任务、异常处理。

## 1. 背景与目标
- 现有页面仅列出承运商、模板、运单示例，需要完整的物流管理：承运商配置、运费计算、运单生成、轨迹跟踪、异常处理，与订单/调拨/售后联动。

**目标**
1. 管理承运商及服务：区域覆盖、服务类型（快递、重货、冷链）、权限、结算。
2. 配置运费模板与规则，供渠道/订单计算运费、出价。
3. 生成运单、跟踪轨迹、处理异常（延误、丢失、签收失败），记录任务。

## 2. 角色
| 角色 | 诉求 |
| --- | --- |
| 供应链/物流 | 配置承运商、监控运力、处理异常 |
| 渠道运营 | 选择运费模板、确认渠道运费策略 |
| 客服/售后 | 查询物流状态、处理投诉 |
| 财务 | 结算运费、对账 |

## 3. 信息架构
1. **承运商管理 (`carriers.vue`)**
   - 列表字段：承运商名称、类型、服务区域、服务类型、状态、联系人、费率、平均时效、满意度。
   - 详情：服务描述、费用、账期、API 凭证、支持渠道、SLA、异常记录。
   - **第三方物流平台对接**：对主流 3PL（顺丰、菜鸟、京东物流、FedEx、DHL 等）封装连接器，配置认证信息、API 版本、订阅能力，支持扩展自定义承运商。
2. **运费模板 (`templates.vue`)**
   - 模板列表：名称、渠道/店铺、计费方式（重量、件数、区域）、币种、状态、更新时间。
   - 详情：区域分区、首重/续重、偏远费、免邮条件、渠道覆盖、实验定价。
3. **运单管理 (`waybills.vue`)**
   - 列表：运单号、订单/调拨、承运商、渠道、状态（创建/揽收/运输中/签收/异常）、预计/实际时效。
   - 详情：物流节点、轨迹、异常、签名、照片、费用、结算状态。
4. **任务 & 异常**
   - 异常类型（延迟、丢失、损坏、退回）；任务分配给供应链/客服；记录处理步骤。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 承运商配置 | 新增/编辑承运商，设置服务、费率、SLA、API、权限；对主流 3PL 预置连接器 |
| 运费模板 | 定义模板、版本管理、实验、渠道绑定、自动选择策略 |
| 运单生成 | 根据订单/调拨生成运单、打印面单、推送到承运商 API |
| 轨迹同步 | 获取承运商轨迹，展示节点、异常、预估到达 |
| 异常处理 | 延迟/丢失/破损告警，任务分派、赔付管理 |
| 结算 & 对账 | 运费结算、对账、费用分析 |
| 可视化 | 物流热力图、在途监控、承运商 KPI |

## 5. 流程
1. **承运商接入**：配置→测试→启用→绑定渠道/仓库。
2. **运费计算**：订单/渠道根据模板计算运费，显示/收取。
3. **运单生成**：订单发货→选择承运商→生成运单→推送→打印面单。
4. **轨迹 & 异常**：同步轨迹→更新状态→异常告警→任务→处理→反馈。
5. **结算**：定期导入承运商账单→对账→确认费用→付款。

## 6. 数据 & API
- 表：`logistics_carriers`、`logistics_carrier_services`、`logistics_rate_templates`、`logistics_zones`、`logistics_waybills`、`logistics_events`、`logistics_alerts`、`logistics_settlements`。
- API：
  - `GET/POST /api/logistics/carriers`、`PATCH /api/logistics/carriers/{id}`、`POST /api/logistics/carriers/{id}/test`（内置对主流 3PL 的标准适配器；支持自定义 Webhook/SDK）。
  - `GET/POST /api/logistics/templates`、`POST /api/logistics/templates/{id}/publish`。
  - `POST /api/logistics/waybills`、`GET /api/logistics/waybills/{id}`、`POST /api/logistics/waybills/{id}/track`。
  - `POST /api/logistics/alerts/{id}/resolve`。

## 7. 权限 & 审计
- 权限：`logistics.carriers`、`logistics.templates`、`logistics.waybills`、`logistics.alerts`、`logistics.settlement`。
- 审计：承运商配置、模板变更、运单异常处理、结算调整记录。

## 8. KPI
| 指标 | 目标 |
| --- | --- |
| 运单准时率 | ≥ 95% |
| 异常处理时效 | < 24h |
| 运费结算准确率 | ≥ 99% |
| 承运商满意度 | ≥ 4.5/5 |

## 9. 风险
- 承运商 API 不稳定；需容错与多渠道路由。
- 运费模板复杂，易配置错误；需版本与测试。
- 轨迹数据缺失影响客户体验；需多数据源。

## 10. Backlog
- 自动选择最佳承运商（基于成本/时效）。
- 物流 IoT 数据接入（温度、位置）。
- 逆向物流（退货、换货）流程。
- 与第三方 TMS 深度集成。

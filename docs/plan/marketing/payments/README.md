# 渠道支付单 / 交易流水 PRD

> 覆盖页面：`web-admin/app/pages/market/payment.vue`、`web-admin/app/pages/payments/providers.vue`、以及结算/财务相关模块。该模块聚焦支付渠道配置、支付单（流水）、对账、分账、风控。

## 1. 背景与目标
- 多渠道、多支付方式（第三方支付、钱包、礼品卡、分期）需要统一管理与对账。当前页面仅展示静态卡片，缺乏支付渠道配置、交易明细、对账、风险控制。

**目标**
1. 配置支付渠道/Provider：不同渠道、币种、费率、结算周期、凭证、风控策略。
2. 聚合支付单/流水：展示每笔交易状态（待支付、支付中、成功、失败、退款）、金额、渠道、关联订单、手续费、分账。
3. 提供对账、异常处理、风控能力，与财务结算协作。

## 2. 角色
| 角色 | 诉求 |
| --- | --- |
| 财务/结算 | 对账、结算、分账、费用核算 |
| 渠道运营 | 管理渠道支付配置、监控支付成功率 |
| 客服 | 查询交易状态、协助退款 |
| 风控 | 监控异常支付、拒付、欺诈 |

## 3. 信息架构
1. **支付渠道配置** (`payments/providers.vue`)
   - 渠道/Provider 列表：名称、类型（支付宝/微信/Stripe/银行）、状态、费率、结算周期、币种、更新时间。
   - 配置详情：API Key、证书、回调 URL、风控规则、手续费、账期、关联店铺。
2. **支付单列表** (`market/payment.vue`)
   - 字段：支付单号、订单号、渠道、支付方式、用户、金额（原价/折扣/实付）、手续费、状态、创建/完成时间、关联礼品卡/优惠券、分账信息。
   - 筛选：渠道、状态、时间、金额区间、支付方式、异常类型。
   - 操作：查看、重发通知、手动对账、退款、冻结、备注。
3. **支付详情**
   - 基本信息、订单/客户、渠道响应、回调日志、风险评分、**分账规则/结果（参与主体、比例、金额、状态）**、退款记录/退款单（与售后单关联）、审计日志。
4. **对账面板**
   - 与第三方文件/接口对比，显示差异、可一键处理（补记、调整、调查）。
5. **风控 & 告警**
   - 实时监控支付成功率、失败原因、拒付、欺诈；配置阈值与自动动作（暂停、降级、人工复核）。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 渠道配置 | 管理支付渠道、API 凭证、费率、结算周期、风控策略、权限 |
| 支付单聚合 | 汇总自营、第三方渠道交易；与订单关联 |
| 对账 | 自动/手动对账，处理差异、导出报表 |
| 风控 | 规则引擎、黑名单、IP/设备监控、限额、人工复核 |
| 退款/退货 | 发起退款、部分退款、退回礼品卡、记录审批 |
| 分账 | 配置分账规则（多方分成）、查看分账结果 |
| 通知 | 支付回调通知、失败告警、对账提醒 |

## 5. 流程
1. **支付渠道接入**：配置 provider、凭证、费率、回调 → 测试 → 上线。
2. **支付处理**：订单发起支付 → 渠道返回状态 → 更新支付单 → 通知订单/履约。
3. **对账**：日/周对账 → 识别差异 → 调整、补记、调查。
4. **退款**：客服/系统发起 → 渠道处理 → 更新支付单、通知用户、记录审计。

## 6. 数据 & API
- 表：`payment_providers`、`payment_provider_configs`、`payment_transactions`、`payment_logs`、`payment_reconciliations`、`payment_risk_events`、`payment_split_rules`。
- API：
  - `GET/POST /v1/admin/payments/providers`、`PATCH /v1/admin/payments/providers/{id}`、`POST /v1/admin/payments/providers/{id}/test`。
  - `GET /v1/admin/payments/transactions`、`GET /v1/admin/payments/transactions/{id}`。
  - `POST /v1/admin/payments/transactions/{id}/refund`、`POST /v1/admin/payments/transactions/{id}/reconcile`。
  - Webhook：`POST /v1/agent/payments/providers/{id}/callback`。

## 7. 权限 & 审计
- 权限：`payments.providers.read/manage`、`payments.transactions.read/manage`、`payments.reconcile`、`payments.refund`、`payments.risk`。
- 审计：渠道配置、支付单状态、对账调整、退款等操作记录。

## 8. KPI
| 指标 | 目标 |
| --- | --- |
| 支付成功率 | ≥ 98% |
| 对账差异处理时效 | < 24h |
| 退款处理时效 | < 48h |
| 风控命中率 | ≥ 95% |

## 9. 风险
- 渠道 API 变更、凭证过期导致支付中断；需监控和测试。
- 支付异常/欺诈；需风控规则与人工复核。
- 对账数据量大，需要高效任务与存储。

## 10. Backlog
- 与财务系统集成（ERP、总账）。
- Payment Orchestration（多渠道自动切换）。
- 高风险国家/地区策略。
- 资金池/多币种管理。

# 对账与报表 PRD

> 关联支付流水（`market/payment.vue`）、结算单（`settlements.md`）、财务系统。负责定期对账、差异处理、财务报表、导出、审计。

## 1. 目标
- 确保站内交易与支付渠道、银行账务一致；提供自动对账、差异处理、财务报表和审计日志。

## 2. 信息架构
- **对账看板**：显示日/周/月对账进度、差异金额、待处理项。
- **对账批次列表**：批次号、范围（渠道、时间）、来源（支付渠道、银行）、状态（待对账/处理中/完成）、差异数、负责人。
- **差异详情**：订单号、渠道流水、系统流水、差异原因、处理状态。
- **报表中心**：
  - 财务报表（收入、费用、税费、利润）。
  - 渠道/供应商对账单。
  - 导出格式（CSV、Excel、PDF）。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 自动对账 | 与支付渠道、银行导入对账文件/接口，自动匹配订单/支付单，标记差异 |
| 差异处理 | 提供处理动作（补记、调整、重试、调查），记录原因与负责人 |
| 手工对账 | 手动上传文件、匹配；支持多币种、汇率 |
| 报表生成 | 生成渠道、供应商、财务报表；支持计划任务、订阅 |
| 数据锁定 | 对账完成后锁定数据，防止篡改 |
| 审计 | 记录对账操作、差异处理、报表生成日志 |

## 4. 数据 & API
- 表：`reconciliation_batches`、`reconciliation_items`、`reconciliation_logs`、`finance_reports`。
- API：`POST /api/reconciliation/batches`、`POST /api/reconciliation/batches/{id}/import`、`POST /api/reconciliation/items/{id}/resolve`、`GET /api/reports`。

## 5. 权限 & 审计
- 权限：`reconciliation.read`、`reconciliation.manage`、`reconciliation.resolve`、`finance.report`。
- 审计：对账结果、处理操作、报表下载。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 对账完成时效 | 日对账 ≤ 24h |
| 差异处理时效 | < 48h |
| 对账差异率 | < 1% |

## 7. 风险
- 数据源格式不一致；需标准化、映射。
- 差异处理权责不清；需流程与权限。
- 报表数据敏感；需权限与脱敏。

## 8. Backlog
- AI 辅助差异原因分析。
- BI 对接（Looker/Tableau）。
- 自动推送报表到 ERP/总账。

# 渠道配置 & Connector 框架 PRD

> 描述渠道通用配置、第三方平台适配器（Connector）以及扩展机制。为接入新渠道、管理 API 配置、同步策略提供标准化框架。

## 1. 目标
- 通过 Connector 框架封装主流渠道（Shopify、Amazon、eBay、京东、天猫、拼多多、小红书等）的 API 差异，提供统一接口；同时允许新增自定义渠道。
- 在渠道配置中集中管理定价策略、库存同步、物流、客服、促销、风控等参数。

## 2. 信息架构
- **渠道配置页**：在渠道详情中展示多个配置模块：
  - 定价策略：引用 pricebook、折扣规则、促销同步、税率。
  - 库存策略：库存来源（全局、仓库、门店）、锁定策略、同步频率、阈值。
  - 物流/履约：默认承运商、运费模板、发货仓、逆向策略。
  - 客服/售后：客服联系方式、SLA、退货地址。
  - 促销/内容：可同步的促销类型、模板。
  - 风控 & 审批：订单/上架风控阈值、审批流。
  - Connector 配置：平台类型、API 版本、回调、应用 ID/Secret、签名方式、限流规则。
- **Connector Registry**：列出可用的 Connector、版本、支持的能力；支持启用/禁用、升级。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| Connector 管理 | 预置主流渠道 Connector，支持安装、升级、禁用；提供自定义 Connector SDK（Node/Go）。
| 配置向导 | 引导用户配置定价、库存、物流策略，保存为模板、复用。
| 能力检测 | 测试 API 连接、权限、限流，展示报告。
| 扩展挂钩 | 提供 Hook（beforePublish、afterOrderSync）供自定义逻辑。
| 安全 | 凭证加密、权限隔离、审计日志。

## 4. 数据 & API
- 表：`channel_connectors`、`channel_connector_versions`、`channel_connector_logs`、`channel_configs`、`channel_hooks`。
- API：`GET/POST /api/channels/connectors`、`POST /api/channels/connectors/{id}/install`、`POST /api/channels/connectors/{id}/test`、`POST /api/channels/{id}/config`。

## 5. 权限 & 审计
- 权限：`channel.connector.read`、`channel.connector.manage`、`channel.config`。
- 审计：Connector 安装/升级/禁用、配置变更、Hook 操作。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 新渠道接入时间 | ≤ 2 周 |
| Connector 可用率 | ≥ 99% |
| 配置错误率 | < 1% |

## 7. Backlog
- Connector 市场，允许第三方贡献。
- 自动生成配置文档。
- 自适应限流与重试策略。

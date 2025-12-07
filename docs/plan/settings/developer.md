# 开发者 & API 设置 PRD

> 涵盖 API Key、Webhooks、Connector 配置、SDK 文档。为第三方集成/开发人员提供统一的开发者控制台。

## 1. 目标
- 管理 API Key、密钥、权限、速率；配置 Webhooks、回调 URL；提供 SDK/文档入口，与渠道 Connector、履约接口联动。

## 2. 信息架构
- API Key 管理：Key 列表、权限、环境（Dev/Prod）、创建人、状态、使用日志。
- Webhook & Callback：事件订阅、URL、签名、重试、日志。
- SDK/文档：链接到 API 文档、示例、版本。
- Connector 设置：与 `channels/connectors.md`、`fulfillment/shipping.md` 的 Hook 联动。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| API Key | 生成/禁用 Key、设置权限、到期、环境、审计 |
| Webhooks | 订阅事件、配置 URL、签名、重试策略、测试工具 |
| API 限流 | 设置限流、重试、告警 |
| SDK & 文档 | 展示 SDK、文档、版本、更新日志 |
| 日志 | 列出 API 调用、错误、告警 |

## 4. 数据 & API
- 表：`api_keys`、`api_key_permissions`、`api_webhooks`、`api_webhook_logs`、`api_call_logs`。
- API：`POST /api/settings/api-keys`、`POST /api/settings/api-keys/{id}/revoke`、`POST /api/settings/webhooks`、`POST /api/settings/webhooks/{id}/test`、`GET /api/settings/api-logs`。

## 5. 权限 & 审计
- 权限：`settings.api.manage`、`settings.api.view`。
- 审计：Key 创建/撤销、Webhook 变更、限流调整。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| Key 发布时长 | < 5 分钟 |
| Webhook 成功率 | ≥ 98% |
| API 错误响应 | < 1% |

## 7. Backlog
- 自动生成 Key 使用报告。
- 与开发者 Portal 集成。
- AI 生成代码示例。

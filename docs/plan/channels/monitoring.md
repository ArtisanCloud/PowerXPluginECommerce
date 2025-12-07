# 渠道健康度与监控 PRD

> 针对渠道看板（现计划新增页面，如 `channels/health.vue`），聚焦渠道 KPI、告警、任务协同。提供 GMV、订单、异常、授权状态、库存覆盖等维度的实时监控。

## 1. 目标
- 打造渠道健康度仪表板，提供多维 KPI、异常告警、任务跟踪，帮助渠道运营及时响应问题。

## 2. 信息架构
- **渠道健康度看板**：每个渠道显示健康度评分（0-100）、关键指标（GMV、订单、转化、库存覆盖、授权状态、错误率）。
- **告警列表**：授权失效、SKU 映射失败、库存不足、价格同步失败、API 错误、订单异常。
- **任务面板**：由告警生成任务，分配负责人、设置 SLA、跟踪进度。
- **日志**：渠道同步日志、API 调用、错误详情。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| KPI 指标库 | 定义渠道 KPI（GMV、订单、库存覆盖、库存周转、促销执行、错误率） |
| 健康度模型 | 基于 KPI 权重计算健康度评分；支持自定义权重 |
| 告警引擎 | 阈值、趋势、异常检测；支持多种渠道事件 |
| 任务协同 | 告警生成任务，指派负责人、SLA、评论、附件 |
| 日志 & 追踪 | 汇总 API 请求、同步记录、错误日志 |
| 通知 & 订阅 | 通过 Slack/邮件/短信订阅渠道健康报告 |

## 4. 数据 & API
- 表：`channel_health_metrics`、`channel_alerts`、`channel_alert_logs`、`channel_health_tasks`、`channel_sync_logs`。
- API：`GET /api/channels/health`、`GET /api/channels/{id}/alerts`、`POST /api/channels/alerts/{id}/ack`、`POST /api/channels/alerts/{id}/task`。

## 5. 权限 & 审计
- 权限：`channel.health.read`、`channel.alerts.manage`、`channel.health.task`。
- 审计：告警处理、任务更新、健康度配置调整。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 告警响应时间 | < 30 分钟 |
| 健康度准确度 | ≥ 95%（与实际数据核对） |
| 任务按时完成 | ≥ 90% |

## 7. Backlog
- AI 异常检测、预测健康度趋势。
- 渠道对比视图。
- 与外部运维工具（PagerDuty）集成。

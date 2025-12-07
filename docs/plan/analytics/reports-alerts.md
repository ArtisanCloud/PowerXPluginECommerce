# 报表导出 & 告警 PRD

> 涉及 `reports/**` 页面的导出、订阅、告警能力，提供财务/运营/管理层需要的定期报告与实时报警。

## 1. 目标
- 统一报表导出、订阅、告警机制：支持多格式（CSV、Excel、PDF）、定时任务、实时告警、多渠道通知。

## 2. 信息架构
- 报表中心：可用报表列表（销售、商品、渠道、财务、库存、售后），描述、更新频率、可见范围。
- 订阅管理：订阅对象、频率、格式、收件人、筛选条件。
- 告警规则：指标、阈值、触发条件、通知渠道、处理任务。
- 历史日志：生成记录、下载、发送结果、告警处理。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 报表模板 | 定义报表结构、字段、过滤器、数据源 |
| 导出 | 手动导出（即时/后台任务）、多格式、压缩、权限控制 |
| 订阅 | 定时/事件触发，发送邮件/Slack/Webhook |
| 告警 | 指标阈值、异常检测、智能告警，生成任务 |
| 历史/日志 | 保存报表生成与告警日志，便于审计 |
| API | `GET /api/reports`、`POST /api/reports/export`、`POST /api/reports/subscriptions`、`POST /api/reports/alerts` |

## 4. 数据 & API
- 表：`analytics_reports`、`analytics_report_runs`、`analytics_subscriptions`、`analytics_alert_rules`、`analytics_alert_logs`。

## 5. 权限 & 审计
- 权限：`analytics.reports.read`、`analytics.reports.manage`、`analytics.alerts.manage`。
- 审计：导出、订阅、告警配置、处理记录。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 报表生成时长 | < 5 分钟 |
| 告警响应时间 | < 30 分钟 |
| 订阅送达率 | ≥ 98% |

## 7. Backlog
- 自然语言生成报告摘要。
- 与外部 BI/数据湖集成。
- AI 告警优先级推荐。

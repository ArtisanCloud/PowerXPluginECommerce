# 审计与风控策略 PRD

> 描述系统级审计、日志、风控策略配置，与 `admin_console_audit_events`、安全模块对齐。

## 1. 目标
- 暴露审计/风控配置：审计级别、日志导出、告警渠道、风控规则（如登录、权限、交易风险），与安全模块协同。

## 2. 信息架构
- 审计设置：审计级别（关键操作、全部）、保留天数、导出、加密、Webhook。
- 风控规则：
  - 登录/认证：失败次数、IP 白名单、MFA。
  - 权限：高权限操作审批。
  - 交易/资金：支付/退款阈值、风控策略、告警。
- 通知：告警渠道（邮件、Slack、Webhook）。
- 日志查看：审计日志列表、过滤、导出。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 审计策略 | 设置审计范围、保留、导出计划、加密；调用 scripts 导出 |
| 风控规则 | 配置规则（条件+动作），如“连续失败登录 → 阻断”，与交易风控对接 |
| 告警 | 配置渠道、模板、严重级别 |
| API | `GET/POST /api/settings/audit`、`GET/POST /api/settings/risk-rules`、`POST /api/settings/audit/export` |

## 4. 数据 & API
- 表：`audit_settings`、`risk_rules`、`audit_exports`、`risk_logs`。

## 5. 权限 & 审计
- 权限：`settings.audit.manage`、`settings.risk.manage`。
- 审计：任何审计/风控配置变更都需记录。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 审计覆盖率 | 100% |
| 风控响应时间 | < 10 分钟 |

## 7. Backlog
- AI 风控策略推荐。
- SIEM/安全平台集成。

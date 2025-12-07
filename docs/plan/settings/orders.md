# 订单与流程配置 PRD

> 覆盖 `web-admin/app/pages/settings/orders.vue` 以及订单/售后流程控制。负责订单状态机、SLA、通知、审批、自动化规则。

## 1. 目标
- 提供管理员可配置的订单/售后流程，包括状态流转、SLA、自动动作、通知模板、例外处理，与渠道/履约/售后联动。

## 2. 信息架构
- 流程图：可视化订单状态（待付款→待发货→履约→完成→售后）和可配置的转移规则。
- 配置项：
  - 状态：名称、显示标签、是否可见、允许操作（取消、发货、退款）。
  - 转移规则：触发条件（事件、时间、审批）、动作（状态变更、通知、锁定库存）。
  - SLA：每个状态的时限、提醒、升级策略。
  - 通知模板：邮件/短信/站内信/Webhook。
  - 自动化规则：如“订单 x 天未发货自动提醒”、“售后超期自动升级”。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 状态机编辑 | 拖拽式编辑状态、转移、条件，支持保存版本、预览 |
| SLA 配置 | 设置时限、提醒方式、升级流程 |
| 通知中心 | 配置通知模板与渠道，支持变量 |
| 自动化 | 触发器（事件/时间） + 动作（通知、状态变更、任务）|
| 审批流程 | 某些状态需要审批（财务/运营），可定义审批人、SLA |
| 日志 | 记录配置变更、启用/禁用、版本 |

## 4. 数据 & API
- 表：`order_workflows`、`order_states`、`order_transitions`、`order_sla`、`order_notifications`、`order_workflow_versions`。
- API：`GET/POST /api/settings/orders/workflows`、`POST /api/settings/orders/workflows/{id}/publish`、`GET /api/settings/orders/workflows/{id}`。

## 5. 权限 & 审计
- 权限：`settings.orders.manage`、`settings.orders.view`、`settings.orders.approval`。
- 审计：配置变更、版本、审批记录。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| SLA 达成率 | ≥ 95% |
| 配置变更审核周期 | ≤ 2 天 |

## 7. Backlog
- 多流程（不同渠道/区域）支持。
- 与自动化旅程联动（SLA 触发旅程）。
- 模板库（B2B/B2C/海外流程）。

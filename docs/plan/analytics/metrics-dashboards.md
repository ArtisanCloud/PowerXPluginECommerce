# 指标与看板 PRD

> 覆盖 `reports/sales.vue`、`reports/products.vue` 的升级需求，以及计划中的自定义 KPI 看板。负责指标定义、看板配置、权限、订阅。

## 1. 目标
- 提供统一的指标库和看板管理，支持自助配置（拖拽组件）、保存视图、权限控制、订阅通知。

## 2. 信息架构
- 指标库：指标、维度、描述、来源、刷新频率、权限、负责人。
- 看板：组件（KPI 卡、折线、柱状、饼图、表格）、布局、筛选、分享。
- 权限：看板共享给个人/团队、公开/私有。
- 订阅：定时发送邮件/Slack、阈值告警。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 指标管理 | 定义指标（日 GMV、订单、转化率、客单、库存周转等），支持版本、审批、备注 |
| 看板编辑器 | 拖拽组件、设置数据源、筛选、联动；保存为模板或个人视图 |
| 共享与权限 | 指定可见范围、角色权限、只读/编辑、复制 |
| 订阅/告警 | 定时发送看板或特定指标；支持阈值告警、异常检测 |
| 历史/快照 | 记录看板变更、保存快照、回滚 |
| API | `GET /api/analytics/metrics`、`GET /api/analytics/dashboards` |

## 4. 数据 & API
- 表：`analytics_metrics`、`analytics_dashboards`、`analytics_dashboard_widgets`、`analytics_subscriptions`。
- API：`POST /api/analytics/dashboards`、`PATCH /api/analytics/dashboards/{id}`、`POST /api/analytics/dashboards/{id}/share`、`POST /api/analytics/subscriptions`。

## 5. 权限 & 审计
- 权限：`analytics.metrics.manage`、`analytics.dashboards.manage`、`analytics.dashboards.view`。
- 审计：指标变更、看板修改、共享、订阅。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 指标准确率 | ≥ 99.5% |
| 看板加载时间 | < 2 秒 |
| 订阅送达率 | ≥ 98% |

## 7. Backlog
- 自然语言搜索看板。
- AI KPI 推荐。
- 多看板联动、钻取。

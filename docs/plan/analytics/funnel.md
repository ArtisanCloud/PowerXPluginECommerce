# 漏斗与路径分析 PRD

> 面向 `reports/conversion.vue`、`customer/analytics/funnel.vue` 的升级，以及计划中的路径分析工具。

## 1. 目标
- 支持自定义漏斗阶段、过滤条件、分群对比、路径探索、告警。

## 2. 信息架构
- 漏斗配置：名称、阶段（事件/页面）、过滤（渠道、活动、客户分群、设备）、时间窗口、目标。
- 漏斗结果：转化率、人数、对比（环比、同比、分群）、流失原因。
- 路径分析：展示 Top Paths、路径分支、平均停留、瓶颈节点。
- 告警：转化率异常、漏斗段异常。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 自定义漏斗 | 可视化配置阶段、拖拽排序、保存模板 |
| 多漏斗对比 | 按渠道、分群、实验版本对比 |
| 路径探索 | 钻取流失路径、Top paths，提供筛选、导出 |
| 告警 | 漏斗转化异常告警，推送通知 |
| API | `POST /api/analytics/funnels`、`GET /api/analytics/funnels/{id}` |

## 4. 数据 & API
- 表：`analytics_funnels`、`analytics_funnel_results`、`analytics_paths`、`analytics_funnel_alerts`。

## 5. 权限 & 审计
- 权限：`analytics.funnel.manage`、`analytics.funnel.view`。
- 审计：漏斗配置、告警处理。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 漏斗计算延迟 | < 5 分钟 |
| 路径分析命中率 | ≥ 95% |

## 7. Backlog
- AI 自动识别瓶颈。
- 自然语言查询漏斗。
- 漏斗与旅程联动（自动触发动作）。

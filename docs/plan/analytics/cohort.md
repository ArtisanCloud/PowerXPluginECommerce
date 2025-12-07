# 分群 & Cohort PRD

> 对应 `customer/analytics/segmentation.vue`、`customer/analytics/cohort.vue`，负责分群管理、留存、LTV、预测。

## 1. 目标
- 提供动态分群、留存(Cohort)分析、LTV、分群对比，支持旅程/营销引用。

## 2. 信息架构
- 分群列表：名称、类型（条件/上传/旅程）、人数、更新时间、来源。
- 分群条件：事件、属性、标签、行为、旅程状态；支持 AND/OR。
- Cohort 分析：按 cohort（日期、渠道、活动）展示留存曲线、ARPU、LTV。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 分群构建 | 可视化条件、支持实时/离线、保存模板 |
| 数据同步 | 分群可导出、同步到旅程、营销、广告渠道 |
| 留存分析 | 按 cohort 查看留存、ARPU、LTV，支持对比 |
| 预测 | LTV 预测、流失预测 |
| API | `POST /api/analytics/segments`、`GET /api/analytics/segments/{id}`、`GET /api/analytics/cohorts` |

## 4. 数据 & API
- 表：`analytics_segments`、`analytics_segment_members`、`analytics_cohorts`、`analytics_ltv_predictions`。

## 5. 权限 & 审计
- 权限：`analytics.segment.manage`、`analytics.cohort.view`。
- 审计：分群变更、导出、同步记录。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 分群生成延迟 | < 10 分钟 |
| 留存计算准确率 | ≥ 98% |

## 7. Backlog
- 与旅程/营销自动联动。
- AI 自动分群建议。
- 与广告平台同步（Facebook、Google）。

# 归因分析 PRD

> 针对渠道/活动/旅程的归因分析，衡量多个触点对转化/GMV 的贡献。

## 1. 目标
- 提供多模型归因工具（首次、末次、线性、时间衰减、算法），帮助营销/渠道评估投入产出。

## 2. 信息架构
- 归因模型设置：选择维度（渠道、活动、旅程、促销、内容）、模型类型、时间窗口、目标指标。
- 归因结果：展示每个触点的贡献度、转化、GMV、ROI。
- 对比 & 报表：不同模型对比、分群筛选、导出。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 模型管理 | 配置/保存归因模型、调整权重、模板 |
| 数据收集 | 聚合多渠道触点事件（广告、邮件、旅程、站内行为）|
| 计算引擎 | 根据模型计算贡献度，支持批处理/实时 |
| 可视化 | Sankey 图、柱状图展示贡献度 |
| API | `POST /api/analytics/attribution`、`GET /api/analytics/attribution/{id}` |

## 4. 数据 & API
- 表：`attribution_models`、`attribution_results`、`attribution_logs`。

## 5. 权限 & 审计
- 权限：`analytics.attribution.manage`、`analytics.attribution.view`。
- 审计：模型变更、结果导出。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 归因计算时效 | < 30 min |
| 数据覆盖率 | ≥ 95% |

## 7. Backlog
- AI/算法归因（Shapley value）。
- 与广告平台的 API 对接。
- 归因结果自动反馈到预算分配。

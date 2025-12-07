# 上架发布与版本管理 PRD

> 面向 `web-admin/app/pages/channels/publishing.vue` 及活动/商品中触发的渠道发布流程。负责商品选择、内容生成、定价/库存策略应用、发布计划、日志、回滚。

## 1. 目标
- 提供统一的上架工作流：选择商品/SKU → 配置渠道参数 → 生成内容 → 校验 → 提交发布 → 监控日志 → 回滚/重新发布。

## 2. 信息架构
- 发布计划列表：渠道、计划名称、批次、SKU 数量、状态、创建人、上线/下线时间、结果。
- 发布向导：
  1. 选择渠道/店铺；
  2. 选择商品/SKU（支持按分组、标签、活动筛选）；
  3. 应用定价策略（引用 pricebook/规则）、库存策略、物流模板、内容模板；
  4. 预览＋校验；
  5. 设置上线/下线时间、灰度策略；
  6. 审批提交；
  7. 发布执行，查看日志；
  8. 成功/失败、回滚、重试。
- 日志与回滚：每个计划的 API 请求/响应、错误、成功项、失败项、已回滚记录。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 批次/计划管理 | 多个发布计划，支持复制、模板、定时 |
| 灰度 & 版本 | 分批上线、A/B、版本记录、回滚 |
| 校验 | 校验 SKU 映射、属性、图片、库存、价格、合规；提示修复 |
| 审批流 | 发布前审批（渠道运营→商品→法务） |
| 执行调度 | 将计划转换为任务队列，调用 Connector API，记录日志 |
| 回滚 & 重试 | 从历史版本回滚；对失败项重试 |
| 通知 | 发布状态推送到 Slack/邮件、仪表板 |

## 4. 数据 & API
- 表：`channel_publish_plans`、`channel_publish_items`、`channel_publish_logs`、`channel_publish_versions`。
- API：`POST /api/channels/publish`、`GET /api/channels/publish/{id}`、`POST /api/channels/publish/{id}/retry`、`POST /api/channels/publish/{id}/rollback`。

## 5. 权限 & 审计
- 权限：`channel.publish.read`、`channel.publish.manage`、`channel.publish.approve`。
- 审计：计划创建、审批、执行、回滚记录，关联操作人。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 发布成功率 | ≥ 97% |
| 失败修复时效 | < 2 小时 |
| 发布周期 | 批次执行延迟 < 15 分钟 |

## 7. Backlog
- 发布排程（按渠道 API 限额控制）。
- 自动内容生成（AI 文案/图）。
- 多版本 diff 查看。

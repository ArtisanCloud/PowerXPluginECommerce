# 履约与物流模块规划总览（可开发版）

> 适用范围：`web-admin/app/pages/shipping/**`、订单/售后/渠道中的发货、运单、运费配置组件。该模块负责承运商管理、运费模板、运单生成、轨迹监控、异常处理、履约任务协同，并预留第三方物流/ TMS 适配层。

## 1. 目标与范围
- **目标**：在“自有仓发货”基础上先跑通履约闭环，同时为顺丰/京东物流等第三方对接预留适配接口。
- **范围**：承运商、运费模板、运单与轨迹、履约任务/波次、逆向物流。
- **不包含**（本阶段）：深度 TMS、复杂波次优化、AI 承运商推荐。

## 2. 当前页面覆盖
| 模块 | 前端路径 | 状态摘要 |
| --- | --- | --- |
| 承运商管理 | `shipping/carriers.vue` | 列表/表单雏形，缺乏服务配置、API 对接、KPI |
| 运费模板 | `shipping/templates.vue` | 基本表单，需支持区域分区、渠道策略、版本 |
| 运单管理 | `shipping/waybills.vue` | 展示占位数据，需完整生命周期与轨迹 |
| 履约任务/波次 | 暂无页面 | 需与仓库、调拨联动 |
| 逆向物流 | 售后模块 | 需统一管理退货运单、换货、补发 |

## 3. 核心概念
- **承运商（Carrier）**：自有配送或第三方物流供应商。
- **承运商服务（Service）**：承运商的不同服务类型（快递/重货/同城/冷链）。
- **运费模板（Rate Template）**：计价规则（重量/件数/体积/区域）。
- **运单（Waybill）**：订单发货的物流凭证与轨迹载体。
- **履约任务（Fulfillment Task）**：拣货/包装/交接的执行单元。
- **逆向运单（Reverse Waybill）**：退货/换货/补发相关物流。

## 4. 自有仓履约流程（必须先跑通）
1) 订单支付成功 → 生成拣货任务
2) 拣货完成 → 包装/称重 → 选择承运商服务
3) 生成运单 → 打印面单 → 发货出库
4) 轨迹同步 → 签收/异常处理

## 5. 第三方承运商适配层（必须预留）
### 5.1 适配器接口（示意）
- `CreateShipment(order, address, items) -> waybill_no`
- `GetTracking(waybill_no) -> events[]`
- `CancelShipment(waybill_no)`
- `GetLabel(waybill_no) -> label_url`

### 5.2 适配策略
- **内置适配器**：顺丰、京东物流、菜鸟、DHL 等。
- **自定义适配器**：通过 webhook/SDK 接入，走统一接口。
- **配置项**：app_id、key、secret、sandbox、回调地址、服务编码映射。

## 6. 数据模型（建议表）
- `logistics_carriers`（承运商）
- `logistics_carrier_services`（承运商服务）
- `logistics_rate_templates`（运费模板）
- `logistics_rate_zones`（区域分区）
- `logistics_waybills`（运单）
- `logistics_tracking_events`（轨迹）
- `logistics_exceptions`（异常）
- `fulfillment_tasks` / `fulfillment_waves`（履约任务/波次）
- `reverse_waybills`（逆向运单）

## 7. API 规划（最小可用）
### 7.1 承运商
- `GET /api/v1/admin/logistics/carriers`
- `POST /api/v1/admin/logistics/carriers`
- `PATCH /api/v1/admin/logistics/carriers/{id}`
- `POST /api/v1/admin/logistics/carriers/{id}/test`

### 7.2 运费模板
- `GET /api/v1/admin/logistics/templates`
- `POST /api/v1/admin/logistics/templates`
- `PATCH /api/v1/admin/logistics/templates/{id}`
- `POST /api/v1/admin/logistics/templates/{id}/publish`

### 7.3 运单与轨迹
- `POST /api/v1/admin/logistics/waybills`
- `GET /api/v1/admin/logistics/waybills/{id}`
- `POST /api/v1/admin/logistics/waybills/{id}/track`
- `POST /api/v1/admin/logistics/waybills/{id}/cancel`

### 7.4 履约任务
- `GET /api/v1/admin/fulfillment/tasks`
- `POST /api/v1/admin/fulfillment/tasks`
- `PATCH /api/v1/admin/fulfillment/tasks/{id}/complete`

### 7.5 逆向物流
- `POST /api/v1/admin/reverse/waybills`
- `GET /api/v1/admin/reverse/waybills/{id}`
- `POST /api/v1/admin/reverse/waybills/{id}/track`

## 8. 与库存/订单边界
- **订单创建**：触发库存锁定
- **发货出库**：扣减库存
- **退货入库**：走逆向运单 + 库存回写
- **异常处理**：触发库存/财务补偿流程

## 9. 迭代路线
- Phase 1：承运商 + 运费模板 + 运单生成/轨迹（自有仓可跑）
- Phase 2：履约任务/波次 + 异常处理 + 自动通知
- Phase 3：第三方承运商适配器 + TMS 深度集成

## 10. 输出物
- PRD（shipping、tasks、reverse-logistics）
- API Contract
- 数据库迁移
- 管理端 UI/交互

## 11. 文档索引
- `index.md`
- `shipping.md`
- `reverse-logistics.md`
- `tasks.md`
- `events.md`
- `integrations.md`
- `schema.md`
- `api-contract.md`
- `roadmap.md`

## 12. Framework 对齐执行约束（PowerXPlugin）

### 12.1 统一原则
- 任务状态与进度统一走 framework task + WS Bus，不新增平行自研机制。
- 前端禁止轮询任务进度，统一由 WS topic 事件驱动。
- standalone 与 host 必须保持同一接口语义（状态、topic、payload 字段）。

### 12.2 客户导入链路（已对齐）
- 提交入口：`POST /api/v1/admin/customers/import`
- 结果页入口：`/customer/import-tasks/:taskId`
- 任务进度 topic：
  - `task.progress`
  - `powerx.task.progress.v1`
  - `worker.task.updated`
- 冲突报告下载：`GET /api/v1/admin/customers/import/conflicts/:taskId`

### 12.3 任务状态读取策略
- 首选 framework 状态读取；
- 本地仅作为短期兼容 fallback（迁移阶段），最终目标是 framework-only。

### 12.4 验收标准
- 导入提交后直接进入结果页，不停留弹窗内等待。
- 结果页进度条可由 WS 实时推进并显示终态。
- 冲突任务可下载冲突明细并可复盘失败原因。
- 启动日志可解释当前运行模式与鉴权决策（2x2x2）。

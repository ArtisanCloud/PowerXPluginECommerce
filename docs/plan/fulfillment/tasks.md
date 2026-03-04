# 履约任务与波次 PRD（可开发版）

> 与库存调拨、仓库、订单联动。该模块负责拣货、包装、波次、派工、执行追踪。

## 1. 目标
- 将订单/调拨转化为可执行任务，支持分配、执行、异常处理。

## 2. 流程
1) 订单支付成功 → 生成拣货任务
2) 拣货完成 → 包装/称重 → 交接承运商
3) 异常上报 → 处理/补偿

## 3. 数据 & API
- 表：`fulfillment_waves`、`fulfillment_tasks`、`fulfillment_task_logs`、`fulfillment_exceptions`
- API：
  - `POST /api/v1/admin/fulfillment/waves`
  - `POST /api/v1/admin/fulfillment/tasks/{id}/complete`
  - `POST /api/v1/admin/fulfillment/exceptions`

## 4. KPI
- 任务按时完成率 ≥ 95%
- 异常处理时效 < 2 小时


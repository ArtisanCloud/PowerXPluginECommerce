# 逆向物流 PRD（可开发版）

> 与 `docs/plan/customer/returns-portal.md`（售后）和支付退款联动，负责退货、换货、补发的物流执行：逆向运单、轨迹、入库、报损、退款结算。

## 1. 目标
1. 为退货/换货/补发生成逆向运单并跟踪轨迹。
2. 支持退货入库质检与库存回写。
3. 联动退款/补发/补偿流程。

## 2. 流程
1) 售后审批通过 → 生成逆向运单
2) 客户寄回/上门取件 → 轨迹同步
3) 到仓质检 → 入库/报损
4) 触发退款/补偿 → 结束

## 3. 数据 & API
- 表：`reverse_waybills`、`reverse_tasks`、`reverse_logs`、`reverse_compensations`
- API：
  - `POST /api/v1/admin/reverse/waybills`
  - `GET /api/v1/admin/reverse/waybills/{id}`
  - `POST /api/v1/admin/reverse/waybills/{id}/track`
  - `POST /api/v1/admin/reverse/tasks/{id}/complete`

## 4. 权限 & KPI
- 权限：`reverse.read`、`reverse.manage`、`reverse.compensation`
- KPI：退货处理周期、逆向异常率、可再售率

## 5. 与售后 RMA 联动（012 收口）
- 联动入口：`POST /api/v1/admin/after-sales/cases/{id}/reverse-logistics`
- 联动约束：
  - 仅 `return_refund` / `exchange` 可绑定逆向物流。
  - 逆向运单必须与售后单 `order_id` 一致，否则返回冲突错误。
- 回写字段（售后摘要）：
  - `reverseWaybillNo`
  - `reverseReceiveStatus`（`pending|received`）
  - `reverseLinkedAt`
- 审计与观测：
  - 事件 `after_sales.reverse_logistics.linked`
  - 轨迹动作 `link_reverse_logistics`

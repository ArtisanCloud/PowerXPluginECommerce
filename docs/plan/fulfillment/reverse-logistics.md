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


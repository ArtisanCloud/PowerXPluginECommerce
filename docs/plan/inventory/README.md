# 库存与仓储模块规划总览（可开发版）

> 适用范围：`web-admin/app/pages/inventory/**`、`shipping/**`、以及商品、渠道、订单、售后模块中涉及库存/仓储的组件。库存与仓储模块涵盖仓库主数据、库存视图、锁定/释放/扣减、补货、盘点、调拨，并为后续第三方 WMS/物流集成预留扩展位。

## 1. 目标与范围
- **目标**：保证“可卖/可履约”的最小闭环，并形成可扩展的数据模型。
- **范围**：仓库主数据、库存状态、库存流转、调拨、盘点、补货与预警。
- **不包含**（本阶段）：复杂波次、WMS 深度对接、预测模型。

## 2. 当前页面覆盖
| 模块 | 前端路径 | 状态摘要 |
| --- | --- | --- |
| 仓库管理 | `inventory/warehouses.vue` | 列表 + 基本信息，缺乏仓库拓扑、权限、容量、KPI |
| 库存总览 | `inventory/stock.vue` | 表格展示库存、锁定、在途，需接入实时数据、告警 |
| 补货与安全库存 | `inventory/replenishment.vue` | 补货建议示例，需接入阈值、任务与审批 |
| 盘点 | `inventory/stocktake.vue` | 盘点计划 UI，需任务、差异分析 |
| 调拨 | `inventory/transfers.vue` | 调拨单列表，需流程、审批、物流、异常处理 |

## 3. 核心概念与状态口径（必须统一）
### 3.1 库存维度
- **维度**：仓库（warehouse）× SKU × 批次/库位（可选）× 状态
- **最小维度**：仓库 + SKU

### 3.2 状态定义
- **available（可用）**：可售库存。
- **locked（锁定）**：订单已下但未支付或未确认发货，暂不可售。
- **on_hand（在库）**：仓库实际物理库存。
- **in_transit（在途）**：调拨/补货运输中。
- **damaged（残次）**：不可售。
- **return_pending（退货待检）**：退货入库前。

> 口径建议：
- `on_hand = available + locked + damaged + return_pending`
- `sellable = available`（前台可售口径）

## 4. 库存流转时序（必须可实现）
### 4.1 订单链路
1) **下单**：`available -> locked`（锁定）
2) **支付成功**：保持锁定（直到拣货/发货）或直接扣减（无仓库场景）
3) **发货出库**：`locked -> on_hand` 扣减（出库）
4) **取消/支付失败**：`locked -> available`（释放）

### 4.2 售后链路（退货）
1) 申请退货：不动库存
2) 退货到仓：`return_pending` 入库
3) 质检通过：`return_pending -> available`
4) 质检失败：`return_pending -> damaged`

### 4.3 调拨链路
1) 发起调拨：源仓 `available -> locked`
2) 出库发运：源仓 `locked -> on_hand` 扣减；目标仓 `in_transit + qty`
3) 到货入库：目标仓 `in_transit -> on_hand`，再转换到 `available`

## 5. 数据模型（建议表）
### 5.1 基础表
- `inventory_warehouses`：仓库主数据（名称、地址、负责人、状态、类型）
- `inventory_warehouse_zones`（可选）：库区/库位
- `inventory_stock`：仓库 SKU 库存（on_hand / available / locked / in_transit / damaged / return_pending）
- `inventory_lots`（可选）：批次/序列号

### 5.2 交易表
- `inventory_transactions`：库存流水（type、delta、source_type、source_id）
- `inventory_locks`：锁定记录（order_id、sku_id、qty、expires_at）
- `inventory_transfers`：调拨单头
- `inventory_transfer_items`：调拨明细
- `inventory_stocktakes`：盘点单
- `inventory_stocktake_items`：盘点明细

### 5.3 预警/补货
- `inventory_replenishment_rules`：安全库存阈值
- `inventory_replenishment_tasks`：补货任务/审批

## 6. API 规划（最小可用）
### 6.1 仓库
- `GET /api/v1/admin/inventory/warehouses`
- `POST /api/v1/admin/inventory/warehouses`
- `PATCH /api/v1/admin/inventory/warehouses/{id}`

### 6.2 库存
- `GET /api/v1/admin/inventory/stock?warehouseId=&skuId=`
- `POST /api/v1/admin/inventory/stock/lock`（订单锁定）
- `POST /api/v1/admin/inventory/stock/release`（订单取消/失败释放）
- `POST /api/v1/admin/inventory/stock/deduct`（发货扣减）
- `POST /api/v1/admin/inventory/stock/adjust`（盘点差异）

### 6.3 调拨
- `GET /api/v1/admin/inventory/transfers`
- `POST /api/v1/admin/inventory/transfers`
- `PATCH /api/v1/admin/inventory/transfers/{id}/approve`
- `PATCH /api/v1/admin/inventory/transfers/{id}/ship`
- `PATCH /api/v1/admin/inventory/transfers/{id}/receive`

### 6.4 盘点
- `GET /api/v1/admin/inventory/stocktakes`
- `POST /api/v1/admin/inventory/stocktakes`
- `PATCH /api/v1/admin/inventory/stocktakes/{id}/submit`
- `PATCH /api/v1/admin/inventory/stocktakes/{id}/confirm`

## 7. 与订单/支付/履约的边界
- **订单创建**：调用库存锁定接口
- **支付失败/取消**：释放锁定
- **发货出库**：扣减库存
- **退货入库**：进入 `return_pending`，质检后转可用/残次

## 8. 扩展点（为第三方 WMS/物流预留）
- 事件：`inventory.locked/released/deducted/adjusted/transferred`（用于同步外部系统）
- 外部适配器：`wms_provider`（仅定义接口，后续实现）
- 可选回调：`POST /api/v1/admin/inventory/webhook`（外部库存变更回写）

## 9. 迭代路线
- **Phase 1**：仓库 + 库存视图 + 锁定/释放/扣减 + 调拨（最小闭环）
- **Phase 2**：盘点 + 补货/预警 + 在途/退货库存
- **Phase 3**：WMS 对接、预测补货、波次任务

## 10. 输出物
- PRD（仓库、库存、调拨、盘点）
- API Contract（库存锁定/释放/扣减）
- 数据库迁移脚本
- 前端页面联动（库存视图/调拨/盘点）


# 路由矩阵（SKU Inventory Stock）

目标：让“库存可写闭环 + 上架前置校验（库存部分）”在既有路由体系下可落地，并保持与宪章一致的安全/租户策略。

## 1. 路由与用途（MVP）

| 路径 | 用途 | 认证/权限 | 备注 |
|------|------|----------|------|
| `/api/v1/admin/product/skus/{skuId}/inventory` | 读取 SKU 库存快照 | JWT + RBAC（admin） | 现有只读接口 |
| `/api/v1/admin/product/skus/{skuId}/inventory/adjust` | 增量调整 default 仓可用库存 | JWT + RBAC（admin） | 本期新增（delta，整数） |

## 2. 中间件一致性要求（强约束）

以上管理端路由必须使用与现有 admin API 一致的中间件链（鉴权、RBAC、tenant context 注入、request_id/trace），避免出现不同路径不同安全策略。


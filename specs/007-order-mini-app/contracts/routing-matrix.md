# 路由矩阵（自营下单 Mini-app + Admin）

目标：让“自营下单闭环（创建订单/库存锁定/后台查询与取消/审计）”在既有路由体系下可落地，并保持与宪章一致的安全/租户策略。

## 1. 路由与用途（MVP）

| 路径 | 用途 | 认证/权限 | 备注 |
|------|------|----------|------|
| `/api/v1/mini-app/orders` | 小程序创建订单 | Customer Token + tenant | 幂等键必需；整单原子；写库存锁定 |
| `/api/v1/mini-app/orders` | 小程序订单列表 | Customer Token + tenant | 仅返回本人订单 |
| `/api/v1/mini-app/orders/{id}` | 小程序订单详情 | Customer Token + tenant | 仅返回本人订单 |
| `/api/v1/admin/orders` | 后台代客下单 | JWT + RBAC + tenant | 幂等键必需；写库存锁定 |
| `/api/v1/admin/orders` | 后台订单列表 | JWT + RBAC + tenant | 支持条件筛选与分页 |
| `/api/v1/admin/orders/{id}` | 后台订单详情 | JWT + RBAC + tenant | 含明细与审计事件 |
| `/api/v1/admin/orders/{id}/cancel` | 后台取消订单 | JWT + RBAC + tenant | 仅 `pending_payment` 可取消；释放库存锁定 |

## 2. 中间件一致性要求（强约束）

- 小程序路由必须使用与现有 miniapp 体系一致的中间件链（tenant 注入 + customer token 鉴权）。
- 后台路由必须使用与现有 admin API 一致的中间件链（鉴权、RBAC、tenant context 注入、request_id/trace）。
- 所有写操作必须在租户事务中执行（`WithTenantTx/BeginTenantTx`），依赖 RLS 强制隔离。

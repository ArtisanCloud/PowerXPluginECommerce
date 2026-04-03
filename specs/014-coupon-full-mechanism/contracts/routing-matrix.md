# 路由矩阵（Coupon Full Mechanism）

| Path | Purpose | Auth | Notes |
|---|---|---|---|
| `/api/v1/admin/coupons/templates**` | 券模板管理 | JWT + RBAC(admin) + tenant | 管理端配置入口 |
| `/api/v1/admin/coupons/issue` | 发券 | JWT + RBAC(admin) + tenant | 运营发放入口 |
| `/api/v1/admin/coupons/assets` | 券资产查询 | JWT + RBAC(admin) + tenant | 客服/运营排障 |
| `/api/v1/admin/coupons/usages` | 券流水查询 | JWT + RBAC(admin) + tenant | 审计与对账 |
| `/v1/coupons/quote` | 订单优惠试算 | JWT + RBAC + tenant | 交易主路径 |

## 中间件一致性

- 交易路径 `/v1/coupons/quote` 必须与订单创建路径使用同一 tenant/JWT/RBAC 中间件链。
- 管理路径必须统一纳入 `/api/v1/admin/rbac` 的权限声明。

## 事件联动约束

- 支付成功回调触发核销时，必须复用既有支付回调鉴权与幂等链路。
- 订单关闭事件触发释放时，必须复用订单状态机合法性校验。

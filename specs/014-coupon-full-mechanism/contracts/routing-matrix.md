# 路由矩阵（Coupon Full Mechanism）

## 1. 管理端路由

| Method | Path | Handler | Service | Tenant | RBAC Key |
|---|---|---|---|---|---|
| `GET` | `/api/v1/admin/coupons/templates` | `admin/coupon.TemplateHandler.List` | `coupon.TemplateService.List` | `EnsureTenant` | `GET:/api/v1/admin/coupons/templates` |
| `POST` | `/api/v1/admin/coupons/templates` | `admin/coupon.TemplateHandler.Create` | `coupon.TemplateService.Create` | `EnsureTenant` | `POST:/api/v1/admin/coupons/templates` |
| `PATCH` | `/api/v1/admin/coupons/templates/:id` | `admin/coupon.TemplateHandler.Update` | `coupon.TemplateService.Update` | `EnsureTenant` | `PATCH:/api/v1/admin/coupons/templates/:id` |
| `POST` | `/api/v1/admin/coupons/issues` | `admin/coupon.IssueHandler.Issue` | `coupon.IssueService.Issue` | `EnsureTenant` | `POST:/api/v1/admin/coupons/issues` |
| `GET` | `/api/v1/admin/coupons/assets` | `admin/coupon.QueryHandler.ListAssets` | `coupon.QueryService.ListAssets` | `EnsureTenant` | `GET:/api/v1/admin/coupons/assets` |
| `GET` | `/api/v1/admin/coupons/usage-logs` | `admin/coupon.QueryHandler.ListUsageLogs` | `coupon.QueryService.ListUsageLogs` | `EnsureTenant` | `GET:/api/v1/admin/coupons/usage-logs` |

## 2. 交易端路由

| Method | Path | Handler | Service | Tenant | Auth |
|---|---|---|---|---|---|
| `POST` | `/api/v1/v1/coupons/quote` | `miniapp/coupon.Handler.Quote` | `coupon.QuoteService.Quote` | `EnsureTenant` | miniapp JWT/customer context |

> 说明：miniapp 路由组挂载在 `/api/v1` 下，子路由声明为 `/v1/coupons/quote`，最终路径为 `/api/v1/v1/coupons/quote`。

## 3. 中间件链校验

- 管理端券路由通过 `admin.RegisterAPIRoutes -> admincoupon.RegisterRoutes` 注册，组级中间件为 `EnsureTenant`。
- 管理端鉴权/权限由外层 admin 中间件链统一处理，券模块内部只依赖 tenant + admin user context。
- miniapp 试算路由通过 miniapp router 注册，使用 miniapp 侧鉴权链 + `EnsureTenant`。

## 4. RBAC 一致性校验

- 券管理 RBAC 定义在 `backend/internal/transport/http/admin/coupon/rbac.go`。
- `backend/internal/transport/http/registry.go` 已合并 `admincoupon.RBACEntries(...)`，确保 `/admin/rbac` 可见。
- 关键资源动作：
  - `com.powerx.plugins.ecommerce:pricing.coupon.template`：`read/manage`
  - `com.powerx.plugins.ecommerce:pricing.coupon.issue`：`manage`
  - `com.powerx.plugins.ecommerce:pricing.coupon.asset`：`read`
  - `com.powerx.plugins.ecommerce:pricing.coupon.usage`：`read`

## 5. 状态机与事件联动约束

- 支付成功回调：`pending_payment -> paid`，同步触发券核销（`reserved -> redeemed`）。
- 支付失败/关闭：订单关闭后触发券释放（`reserved -> available`）。
- 订单取消与超时关闭链路都走同一释放服务，保证行为一致。

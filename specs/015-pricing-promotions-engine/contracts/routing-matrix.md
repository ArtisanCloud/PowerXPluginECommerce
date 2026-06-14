# 路由矩阵（Pricing Promotions Engine）

## 1. 管理端路由

| Method | Path | Handler | Service | Tenant | RBAC Key |
|---|---|---|---|---|---|
| `GET` | `/api/v1/admin/promotions` | `admin/promotion.CampaignHandler.List` | `promotion.CampaignService.List` | `EnsureTenant` | `GET:/api/v1/admin/promotions` |
| `POST` | `/api/v1/admin/promotions` | `admin/promotion.CampaignHandler.Create` | `promotion.CampaignService.Create` | `EnsureTenant` | `POST:/api/v1/admin/promotions` |
| `PATCH` | `/api/v1/admin/promotions/:id` | `admin/promotion.CampaignHandler.Update` | `promotion.CampaignService.Update` | `EnsureTenant` | `PATCH:/api/v1/admin/promotions/:id` |
| `POST` | `/api/v1/admin/promotions/:id/activate` | `admin/promotion.CampaignHandler.Activate` | `promotion.CampaignService.Activate` | `EnsureTenant` | `POST:/api/v1/admin/promotions/:id/activate` |
| `POST` | `/api/v1/admin/promotions/:id/pause` | `admin/promotion.CampaignHandler.Pause` | `promotion.CampaignService.Pause` | `EnsureTenant` | `POST:/api/v1/admin/promotions/:id/pause` |
| `POST` | `/api/v1/admin/promotions/:id/clone` | `admin/promotion.CampaignHandler.Clone` | `promotion.CampaignService.Clone` | `EnsureTenant` | `POST:/api/v1/admin/promotions/:id/clone` |
| `GET` | `/api/v1/admin/promotions/:id/audit-logs` | `admin/promotion.AuditHandler.List` | `promotion.AuditLogService.List` | `EnsureTenant` | `GET:/api/v1/admin/promotions/:id/audit-logs` |

## 2. 交易端路由

| Method | Path | Handler | Service | Tenant | Auth |
|---|---|---|---|---|---|
| `POST` | `/api/v1/v1/promotions/quote` | `miniapp/promotion.Handler.Quote` | `promotion.QuoteService.Quote` | `EnsureTenant` | miniapp JWT/customer context |

> 说明：优先将促销计算接入既有订单 quote/下单链路。若按现有 miniapp 路由挂载模式提供独立调试接口，子路由 `/v1/promotions/quote` 挂在 `/api/v1` 下时，最终路径为 `/api/v1/v1/promotions/quote`；宿主反代后的业务语义仍是 `/v1/promotions/quote`。

## 3. 中间件链校验

- 管理端促销路由通过 `admin.RegisterAPIRoutes -> adminpromotion.RegisterRoutes` 注册，组级中间件包含 `EnsureTenant`。
- 管理端鉴权/权限由外层 admin 中间件链统一处理，促销模块内部依赖 tenant + admin user context。
- 交易端独立试算路由通过 miniapp router 注册，使用 miniapp 鉴权链 + `EnsureTenant`。
- 订单最终下单链路必须在服务层调用 `promotion.QuoteService` 并写入 `OrderPromotionSnapshot`。

## 4. RBAC 一致性校验

- 促销管理 RBAC 定义在 `backend/internal/transport/http/admin/promotion/rbac.go`。
- `backend/internal/transport/http/registry.go` 需要合并 `adminpromotion.RBACEntries(...)`，确保 `/admin/rbac` 可见。
- 关键资源动作：
  - `com.powerx.plugins.ecommerce:pricing.promotion`：`read/manage/publish`
  - `com.powerx.plugins.ecommerce:pricing.promotion.audit`：`read`
  - `com.powerx.plugins.ecommerce:pricing.promotion.metrics`：`read`（MVP 可预留，页面不要求完整指标看板）

## 5. 订单与优惠券联动约束

- 订单 quote：基础价完成后调用促销 quote，使用促销后金额进入 coupon quote。
- 订单提交：重新在服务端计算促销，不能信任前端试算金额。
- 促销快照：订单创建成功时写入 `order_promotion_snapshots`。
- 优惠券互斥：若促销结果 `coupon_stacking_allowed=false`，coupon 层必须拒绝券并返回 `promotion_excludes_coupon`。

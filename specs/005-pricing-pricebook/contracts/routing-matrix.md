# 路由矩阵（Pricing Query）

目标：满足宪章 **Host Contract First** 的路径约束，并避免 `/v1` 主路径与 `/api/v1` 兼容别名在鉴权/租户/权限上的行为差异。

## 1. 路由与用途

| 路径 | 用途 | 认证/权限 | 备注 |
|------|------|----------|------|
| `/api/v1/admin/pricing/**` | 管理端价目表维护（CRUD/版本/条目） | JWT + RBAC（管理端） | 与现有 admin 注册方式保持一致 |
| `/v1/pricing/query` | **业务主路径**：基础查价 | JWT + RBAC + tenant | 宿主反代：`/_p/<plugin-id>/api/* -> backend /v1/**` |
| `/api/v1/pricing/query` | **兼容别名**：基础查价 | JWT + RBAC + tenant | 仅用于存量联调/直连；行为必须与主路径一致 |

## 2. 中间件一致性要求（强约束）

`/v1/pricing/query` 与 `/api/v1/pricing/query` 必须使用**同一套**中间件链（顺序也一致）：

1. `RequestTrace`（request_id 追踪）
2. `JWTAuth`（鉴权）
3. `RBAC`（权限校验）
4. `EnsureTenant`（租户上下文注入/校验）

> 说明：若主路径与兼容别名任意一个缺少 RBAC 或 tenant 校验，会导致“同一接口不同路径不同安全策略”的高风险缺陷。

## 3. 注册策略（实现侧应对齐）

- `/v1` 业务组：由 `backend/internal/router/router.go` 额外挂载（专门满足宪章业务路径），并复用与管理端相同的 JWT/RBAC 组件。
- `/api/v1` 兼容别名：通过现有 `apiPrefix`（默认 `/api/v1`）下的 `Registry.RegisterAPIRoutes(gApi)` 注册。
- 两条路径最终应复用同一 handler（同一函数/同一 service），避免出现实现分叉。

## 4. 验收（冒烟测试）

需要至少一个路由级测试证明：

- `/v1/pricing/query` 与 `/api/v1/pricing/query` 均存在；
- 两者绑定的 handler 是同一个（或等效同一执行路径）。


# Tasks: 自营下单（Mini-app 下单 + Admin 代客下单）MVP

**Input**: Design documents from `/specs/007-order-mini-app/`  
**Prerequisites**: `specs/007-order-mini-app/plan.md`, `specs/007-order-mini-app/spec.md`, `specs/007-order-mini-app/research.md`, `specs/007-order-mini-app/data-model.md`, `specs/007-order-mini-app/contracts/`, `specs/007-order-mini-app/quickstart.md`

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 复核 miniapp 路由鉴权结构（open vs protected）并确认订单路由将注册到 protected 组（`backend/internal/transport/http/miniapp/router.go`）
- [x] T002 复核可售校验与库存口径：确认 sellability 使用 `SUM(available_qty - locked_qty)` 且可复用到下单校验（`backend/internal/services/miniapp/sellability/service.go`）
- [x] T003 复核幂等基础设施（Postgres 后备存储、TTL 配置）并确认创建订单将使用 `Idempotency-Key` header（`backend/internal/services/integration/factory.go`，`backend/internal/entity/repository/integration/idempotency_repository.go`，`backend/internal/config/integration.go`）

---

## Phase 2: Foundational (Blocking Prerequisites)

- [ ] T004 新增订单域表名常量并集中管理（`backend/internal/entity/models/model.go`）
- [ ] T005 [P] 新增订单域模型：Order / OrderItem / OrderEvent（`backend/internal/entity/models/order/order.go`，`backend/internal/entity/models/order/order_item.go`，`backend/internal/entity/models/order/order_event.go`）
- [ ] T006 新增订单域仓储：创建订单（含明细/事件同事务）、按 customer/admin 查询列表、按 id 查询详情（`backend/internal/entity/repository/order/order_repository.go`，`backend/internal/entity/repository/order/order_item_repository.go`，`backend/internal/entity/repository/order/order_event_repository.go`）
- [ ] T007 新增订单域数据库迁移注册（AutoMigrate + RLS/索引按宪章要求），并加入迁移列表（`backend/cmd/database/migrate/migrations/007_order.go`，`backend/cmd/database/migrate/migrate.go`）
- [ ] T008 扩展库存仓储以支持“锁定/解锁”最小操作（在同一租户事务内、行锁/for update、更新 `locked_qty`），供订单服务复用（`backend/internal/entity/repository/product_sku/inventory_repository.go`）
- [ ] T009 新增订单域错误语义与 DTO（请求/响应/列表分页）以便 miniapp/admin handler 复用（`backend/internal/services/admin/order/types.go`，`backend/internal/services/miniapp/order/types.go`）
- [ ] T010 新增订单 RBAC 资源声明（默认 RBAC 资源列表）与路由级权限映射入口（`backend/internal/transport/http/admin/rbac_loader.go`，`backend/internal/transport/http/admin/order/rbac.go`）

**Checkpoint**: Phase 2 完成后，可开始按用户故事并行实现接口与业务逻辑。

---

## Phase 3: User Story 1 - 小程序下单（一次性商品）(Priority: P1) 🎯 MVP

**Goal**: 小程序用户态创建订单（幂等 + 整单原子 + 锁库存），并可查询订单列表/详情（仅本人订单）。

**Independent Test**: 使用 `Idempotency-Key` 调用 `POST /api/v1/mini-app/orders` 创建订单；验证库存 `locked_qty` 增加且重复提交不重复锁定；再调用列表/详情仅返回本人订单。

### Implementation for User Story 1

- [ ] T011 [P] [US1] 在 sellability 服务中增加“按 SKU 列表评估”的复用入口（输出 price + availableQty + reasons），并让现有按 SPU 的 Evaluate 复用该能力（`backend/internal/services/miniapp/sellability/service.go`，`backend/internal/services/miniapp/sellability/service_test.go`）
- [ ] T012 [US1] 实现 miniapp 订单创建 Service：校验入参、幂等 claim、调用 sellability 做二次可售校验、校验 `availableQty >= qty`、计算金额快照、写订单/明细/事件、锁库存、保存幂等响应（`backend/internal/services/miniapp/order/service.go`）
- [ ] T013 [US1] 实现 miniapp 订单查询 Service：列表分页与详情读取，强制 customer_id 过滤（`backend/internal/services/miniapp/order/query.go`）
- [ ] T014 [P] [US1] 增加 miniapp 订单 handler 与路由（POST/GET list/GET detail），并挂载到 protected 组（`backend/internal/transport/http/miniapp/order/handler.go`，`backend/internal/transport/http/miniapp/order/routes.go`，`backend/internal/transport/http/miniapp/router.go`）
- [ ] T015 [P] [US1] 增加 service 单测：幂等重复提交、库存不足整单失败无副作用、不可售失败无副作用（`backend/internal/services/miniapp/order/service_test.go`）

---

## Phase 4: User Story 2 - Web-admin 代客下单（运营/客服）(Priority: P2)

**Goal**: 后台为指定客户创建订单（复用 US1 的校验/幂等/锁库存/金额快照），返回订单摘要。

**Independent Test**: 调用 `POST /api/v1/admin/orders`（含 `Idempotency-Key`）创建订单，验证订单归属 customer_id、库存锁定成功且幂等重试不重复锁定。

### Implementation for User Story 2

- [ ] T016 [US2] 实现 admin 代客下单 Service：复用通用创建逻辑，但以 customerId 为目标客户，记录 created_by 为 admin（`backend/internal/services/admin/order/service.go`）
- [ ] T017 [P] [US2] 增加 admin 订单 handler 与路由（POST /admin/orders），并注册到 admin API（`backend/internal/transport/http/admin/order/handler.go`，`backend/internal/transport/http/admin/order/routes.go`，`backend/internal/transport/http/admin/routes.go`）
- [ ] T018 [P] [US2] 补齐 route-level RBAC 映射（create/read/cancel）并更新默认 RBAC 资源列表（`backend/internal/transport/http/admin/order/rbac.go`，`backend/internal/transport/http/admin/rbac_loader.go`）
- [ ] T019 [P] [US2] 增加 service 单测：幂等、库存不足、不可售、customer 不存在（`backend/internal/services/admin/order/service_test.go`）

---

## Phase 5: User Story 3 - 后台订单管理（列表/详情/取消/日志）(Priority: P3)

**Goal**: 后台可查询订单列表/详情、取消订单（仅 `pending_payment`），并查看订单事件日志用于追溯。

**Independent Test**: 后台创建订单后可在列表检索到并查看详情；取消后状态变更为 `cancelled` 且库存解锁；详情可看到创建/取消事件。

### Implementation for User Story 3

- [ ] T020 [US3] 实现 admin 订单列表与详情查询接口（GET /admin/orders，GET /admin/orders/{id}），支持分页与基础筛选（`backend/internal/transport/http/admin/order/handler.go`）
- [ ] T021 [US3] 实现取消订单 Service：仅后台、仅 `pending_payment` 可取消；同事务内更新订单状态、写取消事件、释放库存锁定（`backend/internal/services/admin/order/cancel.go`）
- [ ] T022 [US3] 实现取消订单接口（POST /admin/orders/{id}/cancel）并将错误语义对齐（`backend/internal/transport/http/admin/order/handler.go`）
- [ ] T023 [P] [US3] 增加 repository/service 单测：取消状态机约束、库存解锁下限为 0、事件写入（`backend/internal/services/admin/order/cancel_test.go`，`backend/internal/entity/repository/order/order_repository_test.go`）
- [ ] T024 [P] [US3] web-admin 最小订单管理页：订单列表、订单详情、取消按钮与错误提示（`web-admin/app/pages/orders/index.vue`，`web-admin/app/pages/orders/[id].vue`，`web-admin/app/composables/api/useOrder.ts`）

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T025 [P] 统一错误码与可解释错误消息（参数错误/不可售/库存不足/幂等冲突/禁止取消/权限不足），并确保前端可直接展示（`backend/internal/services/**/order/*.go`，`backend/internal/transport/http/**/order/*.go`）
- [ ] T026 [P] 结构化日志与审计字段补齐（包含 request_id/tenant_uuid/operator），确保问题可追踪（`backend/internal/entity/models/order/order_event.go`，`backend/internal/services/**/order/*.go`）
- [ ] T027 运行 quickstart 验证清单并更新“已执行记录”（`specs/007-order-mini-app/quickstart.md`）

---

## Dependencies & Execution Order

- Phase 1（T001-T003）→ Phase 2（T004-T010）为所有 User Story 的阻塞前置
- US1 依赖：T004-T010（尤其是模型/迁移、库存锁定仓储、sellability 复用入口）
- US2 依赖：T004-T010 + US1 的通用创建逻辑可复用（T012）
- US3 依赖：T004-T010 + admin 路由/权限框架已接入（T017-T018）

## Parallel Opportunities

- 可并行（标记为 `[P]` 的任务）：T005/T006、T011/T014/T015、T017/T018/T019、T023/T024、T025/T026/T027

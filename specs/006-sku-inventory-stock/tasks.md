# Tasks: 库存（Stock）MVP：SKU 可用库存闭环与上架前置校验

**Input**: Design documents from `/specs/006-sku-inventory-stock/`  
**Prerequisites**: `specs/006-sku-inventory-stock/plan.md`, `specs/006-sku-inventory-stock/spec.md`, `specs/006-sku-inventory-stock/research.md`, `specs/006-sku-inventory-stock/data-model.md`, `specs/006-sku-inventory-stock/contracts/`, `specs/006-sku-inventory-stock/quickstart.md`

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 确认库存表与模型注册现状，并补齐/验证唯一约束 `unique(tenant_uuid, sku_id, warehouse_id)`（`backend/internal/entity/models/product_sku/inventory.go`，`backend/cmd/database/migrate/migrations/002_product_sku.go`）
- [x] T002 确认 SKU 详情页路由/文件位置（`web-admin/app/pages/product/skus/[id].vue`）并定位可插入“库存卡片”的组件/页面（`web-admin/app/components/product/sku/InventorySnapshotCard.vue`）

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T003 定义库存调整请求/响应 DTO（`backend/internal/services/admin/product_sku/types.go`）
- [x] T004 [P] 扩展库存仓储：所有读写通过 `BeginTenantTx/WithTenantTx` 注入 `app.tenant_uuid`，并提供按（tenant, sku, warehouse）读取与 adjust（delta）能力（`backend/internal/entity/repository/product_sku/inventory_repository.go`）
- [x] T005 [P] 新增库存审计模型（ProductSKUAuditLog）并注册表名（`backend/internal/entity/models/product_sku/`，`backend/cmd/database/migrate/migrations/002_product_sku.go`）
- [x] T006 [P] 新增库存审计仓储（`backend/internal/entity/repository/product_sku/audit_log_repository.go`）
- [x] T007 在 SKU Service 中实现库存增量调整用例：在同一 `WithTenantTx` 内完成“读取当前值→校验→更新库存→写审计→提交”，default 仓、delta 整数、结果不可为负、返回快照（`backend/internal/services/admin/product_sku/inventory.go`）
- [x] T008 在库存调整用例中写入审计记录（与库存更新同一事务）（`backend/internal/services/admin/product_sku/inventory.go`，`backend/internal/entity/repository/product_sku/audit_log_repository.go`）
- [x] T009 统一错误语义：sku 不存在/权限不足/并发冲突/结果为负（`backend/internal/transport/http/admin/product_sku/handler.go`）

**Checkpoint**: Foundational 完成后，可独立完成 US1/US2/US3 的实现与验收。

---

## Phase 3: User Story 1 - 运营可为 SKU 写入可用库存 (Priority: P1) 🎯 MVP

**Goal**: 为 SKU 提供 default 仓的可用库存增量调整（delta）能力，并可读取快照验证。

**Independent Test**: 调用库存调整接口（delta=+10/ -7）后，再调用快照接口，验证 `available_qty` 变化；尝试导致负数时应被拒绝。

### Implementation for User Story 1

- [x] T010 [P] [US1] 增加库存调整路由与 handler，并补齐对应 RBAC 映射（`backend/internal/transport/http/admin/product_sku/inventory_handler.go`，`backend/internal/transport/http/admin/product_sku/routes.go`，`backend/internal/transport/http/admin/product/rbac.go`）
- [x] T011 [US1] 实现 `POST /api/v1/admin/product/skus/{skuId}/inventory/adjust`：解析 delta、调用 service、返回快照（`backend/internal/transport/http/admin/product_sku/inventory_handler.go`）
- [x] T012 [US1] 补齐/调整路由注册（`backend/internal/transport/http/admin/product_sku/routes.go`）
- [x] T013 [P] [US1] 增加 service 单测：delta 正/负、禁止负结果、tenant 缺失（`backend/internal/services/admin/product_sku/inventory_test.go`）
- [x] T014 [P] [US1] 增加 repository 单测：upsert/adjust 与唯一性（`backend/internal/entity/repository/product_sku/inventory_repository_test.go`）

---

## Phase 4: User Story 2 - web-admin 提供 SKU 库存维护入口 (Priority: P2)

**Goal**: 运营在 SKU 详情页查看库存快照并调整 default 仓可用库存（delta，整数）。

**Independent Test**: 打开 SKU 详情页可看到库存快照；输入 delta 保存后快照更新；错误时有提示且页面不崩溃。

### Implementation for User Story 2

- [ ] T015 [P] [US2] 扩展 SKU API：新增库存调整方法（delta）并对齐类型（`web-admin/app/composables/api/useSku.ts`，`web-admin/app/types/product/sku.ts`）
- [ ] T016 [US2] 扩展 SKU store：新增 `adjustInventory` action（调用 API → 刷新快照）（`web-admin/app/stores/productSku.ts`）
- [ ] T017 [US2] 扩展库存卡片：在 `InventorySnapshotCard` 增加 delta 输入与保存按钮，并复用现有快照刷新（`web-admin/app/components/product/sku/InventorySnapshotCard.vue`）
- [ ] T018 [US2] 校验 SKU 详情页 inventory tab 已挂载 `InventorySnapshotCard` 且 props 传递正确（`web-admin/app/pages/product/skus/[id].vue`）
- [ ] T019 [P] [US2] 组件测试：快照渲染、保存成功刷新、保存失败提示（`web-admin/tests/component/inventory-snapshot-card.spec.ts`）

---

## Phase 5: User Story 3 - 上架/发布的库存前置校验 (Priority: P3)

**Goal**: 在 SPU 发布 与 渠道上架/同步发布 两处增加库存校验：至少 1 个 SKU 在 default 仓 `available_qty > 0`；失败仅阻止并提示原因。

**Independent Test**: 构造 SPU 下所有 SKU 库存为 0 时发布/上架失败；将其中一个 SKU `available_qty` 调整为正数后，发布/上架允许继续。

### Implementation for User Story 3

- [ ] T020 [US3] 在 SPU 发布流程中插入库存校验（定位发布 service：`backend/internal/services/admin/product/spu/`）
- [ ] T021 [US3] 在渠道上架/同步发布流程中插入库存校验（定位渠道发布/同步 service：`backend/internal/services/admin/product_sku/channel.go` 或渠道发布相关 service）
- [ ] T022 [P] [US3] 增加库存校验 service 方法：输入 tenant + spu_id（或 sku_ids），输出是否存在可售 SKU（`backend/internal/services/admin/product_sku/` 或 `backend/internal/services/admin/product/spu/`）
- [ ] T023 [P] [US3] 增加集成测试：覆盖发布与渠道上架两个触发点（`backend/tests/` 下新增 `inventory_gate_publish_test.go` 等）

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T024 [P] 文档回写：在 `docs/plan/inventory/stock.md` 或相关 guide 补充“default 仓 + delta 调整 + 上架校验”落地说明
- [ ] T025 运行 quickstart 验证清单并记录结果（`specs/006-sku-inventory-stock/quickstart.md`）
- [ ] T026 [P] 确认 RBAC scope/manifest 输出（`/api/v1/admin/rbac`）包含库存相关权限点（对应库存调整与读取）

---

## Dependencies & Execution Order

- Phase 1（T001-T002）→ Phase 2（T003-T009）为所有 User Story 的阻塞前置
- US1 依赖：T003-T009
- US2 依赖：T003-T009，且依赖 US1 的接口可用（T010-T012）
- US3 依赖：T003-T009，且依赖 US1 的库存数据可写以便验收

## Parallel Opportunities

- 任务可并行（标记为 `[P]` 的任务）：T004/T005/T006、T010/T013/T014、T015/T019、T022/T023、T024/T026

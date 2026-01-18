# Tasks: 营销支付（后台管理）

**Input**: Design documents from `/specs/008-order-payments/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 未要求新增测试任务。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 对齐支付模块文档与索引路径（如需）在 `docs/plan/marketing/payments/README.md`
- [x] T002 [P] 确认权限点清单与命名规范在 `docs/plan/marketing/payments/README.md`
- [x] T003 [P] 确认小程序支付文档与现状一致在 `docs/plan/marketing/payments/miniapp.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

- [x] T004 设计支付相关领域模型与表结构草案在 `specs/008-order-payments/data-model.md`
- [x] T005 统一支付状态机与退款/风控状态枚举在 `specs/008-order-payments/data-model.md`
- [x] T006 统一对外 API 合同与字段命名在 `specs/008-order-payments/contracts/openapi.yaml`
- [x] T007 [P] 规划支付回调处理与幂等策略说明在 `specs/008-order-payments/research.md`
- [x] T008 在 `backend/internal/entity/models/model.go` 新增支付相关表名常量并绑定 `TableName()` 返回值
- [x] T009 在 `backend/cmd/database/migrate/migrate.go` 注册支付相关模型的 AutoMigrate
- [x] T010 在支付相关仓储中内嵌 `*repository.BaseRepository[T]` 并通过 `BeginTenantTx/WithTenantTx` 注入租户上下文

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 后台支付管理总览 (Priority: P1) 🎯 MVP

**Goal**: 完成支付渠道配置与支付单列表/详情的后台管理闭环。

**Independent Test**: 通过后台页面完成渠道配置与支付单查询/详情查看。

### Implementation for User Story 1

- [x] T011 [P] [US1] 新增支付渠道模型定义在 `backend/internal/entity/models/payment_provider.go`
- [x] T012 [P] [US1] 新增支付单模型定义在 `backend/internal/entity/models/payment_transaction.go`
- [x] T013 [P] [US1] 新增退款模型定义在 `backend/internal/entity/models/payment_refund.go`
- [x] T014 [P] [US1] 新增风险事件模型定义在 `backend/internal/entity/models/payment_risk_event.go`
- [x] T015 [P] [US1] 新增分账规则模型定义在 `backend/internal/entity/models/payment_split_rule.go`
- [x] T016 [P] [US1] 新增分账结果模型定义在 `backend/internal/entity/models/payment_split_result.go`
- [x] T017 [P] [US1] 新增支付渠道仓储在 `backend/internal/entity/repository/payment_provider_repository.go`
- [x] T018 [P] [US1] 新增支付单仓储在 `backend/internal/entity/repository/payment_transaction_repository.go`
- [x] T019 [P] [US1] 新增退款仓储在 `backend/internal/entity/repository/payment_refund_repository.go`
- [x] T020 [P] [US1] 新增风险事件仓储在 `backend/internal/entity/repository/payment_risk_event_repository.go`
- [x] T021 [P] [US1] 新增分账规则仓储在 `backend/internal/entity/repository/payment_split_rule_repository.go`
- [x] T022 [P] [US1] 新增分账结果仓储在 `backend/internal/entity/repository/payment_split_result_repository.go`
- [x] T023 [US1] 新增支付渠道服务在 `backend/internal/services/admin/payments/providers_service.go`
- [x] T024 [US1] 新增支付单服务在 `backend/internal/services/admin/payments/transactions_service.go`
- [x] T025 [US1] 新增退款服务在 `backend/internal/services/admin/payments/refunds_service.go`
- [x] T026 [US1] 新增风险事件服务在 `backend/internal/services/admin/payments/risk_events_service.go`
- [x] T027 [US1] 新增分账规则服务在 `backend/internal/services/admin/payments/split_rules_service.go`
- [x] T028 [US1] 新增分账结果服务在 `backend/internal/services/admin/payments/split_results_service.go`
- [x] T029 [US1] 新增支付渠道 API 在 `backend/internal/transport/http/admin/payments/providers_handler.go`
- [x] T030 [US1] 新增支付单 API 在 `backend/internal/transport/http/admin/payments/transactions_handler.go`
- [x] T031 [US1] 新增退款 API 在 `backend/internal/transport/http/admin/payments/refunds_handler.go`
- [x] T032 [US1] 新增风险事件 API 在 `backend/internal/transport/http/admin/payments/risk_events_handler.go`
- [x] T033 [US1] 新增分账规则 API 在 `backend/internal/transport/http/admin/payments/split_rules_handler.go`
- [x] T034 [US1] 新增分账结果 API 在 `backend/internal/transport/http/admin/payments/split_results_handler.go`
- [x] T035 [US1] 增加 RBAC 权限映射在 `backend/internal/transport/http/admin/payments/rbac.go`
- [x] T036 [US1] 更新后台支付渠道页面在 `web-admin/app/pages/payments/providers.vue`
- [x] T037 [US1] 更新后台支付单页面在 `web-admin/app/pages/market/payment.vue`
- [x] T038 [US1] 在支付详情增加退款入口与记录展示在 `web-admin/app/pages/market/payment.vue`
- [x] T039 [US1] 在支付详情增加风险事件查看入口在 `web-admin/app/pages/market/payment.vue`
- [x] T040 [US1] 在支付详情增加分账结果查看入口在 `web-admin/app/pages/market/payment.vue`

**Checkpoint**: User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - 对账与异常处理 (Priority: P2)

**Goal**: 支持日/周对账流程与差异处理记录。

**Independent Test**: 能生成对账批次并记录差异处理结果。

### Implementation for User Story 2

- [x] T041 [P] [US2] 新增对账批次模型在 `backend/internal/entity/models/payment_reconciliation.go`
- [x] T042 [P] [US2] 新增对账差异项模型在 `backend/internal/entity/models/payment_reconciliation_item.go`
- [x] T043 [P] [US2] 新增对账仓储在 `backend/internal/entity/repository/payment_reconciliation_repository.go`
- [x] T044 [US2] 新增对账服务在 `backend/internal/services/admin/payments/reconciliation_service.go`
- [x] T045 [US2] 新增对账 API 在 `backend/internal/transport/http/admin/payments/reconciliation_handler.go`
- [x] T046 [US2] 在支付详情增加对账处理入口在 `web-admin/app/pages/market/payment.vue`

**Checkpoint**: User Stories 1 AND 2 should both work independently

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T054 [P] 补充支付审计与事件记录在 `backend/internal/observability/payments/`
- [x] T055 校验支付状态与订单状态一致性策略说明在 `docs/plan/marketing/payments/miniapp.md`
- [x] T056 运行 quickstart 检查项与自检记录在 `specs/008-order-payments/quickstart.md`
- [x] T057 验证支付列表与状态查询性能目标（记录 p95 结果与样本量）在 `specs/008-order-payments/quickstart.md`

---

## Phase 7: User Story 4 - 后台手动收款与审核 (Priority: P1)

**Goal**: 支持后台补录线下/转账支付记录并完成审核；审核通过后更新订单状态为 `paid`，审核拒绝保持订单不变，并记录审计链路。

**Independent Test**: 对 `pending_payment` 订单创建手动收款记录 → 另一管理员审核通过 → 订单状态变更为 `paid`，事件记录完整；拒绝审核时状态不变。

### Implementation for User Story 4

- [x] T058 [US4] 新增手动收款审核模型（含提交人/审核人/状态/原因/凭证信息）并加入迁移与表名常量（`backend/internal/entity/models/`，`backend/cmd/database/migrate/migrations/`，`backend/internal/entity/models/model.go`）
- [x] T059 [US4] 新增手动收款仓储与状态机约束（防重复审核/回滚）（`backend/internal/entity/repository/payment_manual_review_repository.go`）
- [x] T060 [US4] 实现手动收款 Service：创建申请、审核通过/拒绝、同事务更新订单状态并写入订单事件与支付审计（`backend/internal/services/admin/payments/manual_reviews_service.go`）
- [x] T061 [US4] 增加 admin API：创建手动收款、审核通过/拒绝、查询记录（`backend/internal/transport/http/admin/payments/manual_reviews_handler.go`，`backend/internal/transport/http/admin/payments/routes.go`）
- [x] T062 [US4] 更新 OpenAPI 合同补充手动收款接口（`specs/008-order-payments/contracts/openapi.yaml`）
- [x] T063 [US4] web-admin：订单详情增加“手动收款”入口与审核信息展示（`web-admin/app/pages/orders/[id].vue`）
- [x] T064 [US4] 增加服务层单测：审核通过更新订单状态、拒绝不变、重复审核禁止（`backend/internal/services/admin/payments/manual_reviews_service_test.go`）

---

## Phase 8: User Story 5 - 后台代客下单体验对齐 (Priority: P1)

**Goal**: 后台“新建订单”流程对齐小程序下单逻辑：先选 SPU + 规格，再落到 SKU；支持多 SKU 下单；客户选择使用搜索 + 最近优先，不提供分页控件。

**Independent Test**: 在后台新建订单中，选客户 → 搜索 SPU → 选择规格组合 → 生成/选择对应 SKU → 添加多行 SKU 并成功创建订单；客户下拉在无关键词时按最近订单排序展示，输入关键词后按匹配结果展示。

### Implementation for User Story 5

- [ ] T065 [US5] 未开未实现：web-admin 新建订单表单支持多 SKU 行（增删行、每行 SKU + 数量），提交时映射为 `items[]`（`web-admin/app/pages/market/orders.vue`）
- [ ] T066 [US5] 未开未实现：新增 SPU 搜索与选择（typeahead），选中后加载规格组；仅在有规格时加载 SKU 列表（`web-admin/app/pages/market/orders.vue`，`web-admin/app/composables/api/useSpu.ts`，`web-admin/app/composables/api/useProductSpec.ts`，`web-admin/app/composables/api/useSku.ts`）
- [ ] T067 [US5] 未开未实现：规格选择组件（必选规格校验），根据规格组合自动定位唯一 SKU 并展示规格摘要；不允许手动选择 SKU（`web-admin/app/pages/market/orders.vue`）
- [ ] T068 [US5] 未开未实现：客户下拉改为“搜索优先 + 最近排序”，无关键词时按 `lastOrderAt` 降序取 TopN；有关键词时按匹配结果返回；不显示分页控件（`web-admin/app/pages/market/orders.vue`，`backend/internal/entity/repository/customer/customer_repo.go`）
- [ ] T069 [US5] 未开未实现：更新文案与空态提示，强调“先选 SPU/规格 → 自动定位 SKU”，补充多 SKU 下单校验与“无规格不可下单”提示（`web-admin/app/pages/market/orders.vue`，`web-admin/app/i18n/*.json`）
- [ ] T070 [US5] 未开未实现：渠道下拉改为从渠道主数据接口拉取，保持与渠道管理一致（`web-admin/app/pages/market/orders.vue`，`web-admin/app/composables/useChannels.ts`）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2)
- **User Story 3 (P3)**: Can start after Foundational (Phase 2)

### Within Each User Story

- Models before services
- Services before endpoints
- Core implementation before integration

### Parallel Opportunities

- T002, T003 can run in parallel
- T011-T022 can run in parallel
- T041-T043 can run in parallel
- T047-T048 can run in parallel

---

## Parallel Example: User Story 1

```bash
Task: "新增支付渠道模型定义 in backend/internal/entity/models/payment_provider.go"
Task: "新增支付单模型定义 in backend/internal/entity/models/payment_transaction.go"
Task: "新增支付渠道仓储 in backend/internal/entity/repository/payment_provider_repository.go"
Task: "新增支付单仓储 in backend/internal/entity/repository/payment_transaction_repository.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo

### Parallel Team Strategy

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3

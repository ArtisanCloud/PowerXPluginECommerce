# Tasks: 售后 RMA 与退货门户

**Input**: Design documents from `/specs/012-after-sales-rma/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 未要求新增“先测后码”测试任务；在 Polish 阶段执行回归与门禁验证。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立售后模块文档与目录骨架

- [X] T001 创建售后模块目录骨架在 `backend/internal/{entity/models,entity/repository,services,transport/http}/{admin,miniapp}/after_sales/`
- [X] T002 [P] 创建售后页面与 API composable 骨架在 `web-admin/app/pages/{customer/returns-portal.vue,market/after-sales.vue}` 与 `web-admin/app/composables/api/useAfterSales.ts`
- [X] T003 [P] 初始化本特性合同与 quickstart 索引在 `specs/012-after-sales-rma/{contracts/openapi.yaml,quickstart.md}`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 多租户、模型注册、RBAC、状态机约束等公共能力（阻塞所有故事）

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 新增售后域表名常量并绑定 `TableName()` 在 `backend/internal/entity/models/model.go`
- [X] T005 [P] 新增售后主模型 `AfterSaleCase` 在 `backend/internal/entity/models/after_sales/case.go`
- [X] T006 [P] 新增售后轨迹/凭证/决策模型在 `backend/internal/entity/models/after_sales/{timeline.go,evidence.go,decision.go}`
- [X] T007 [P] 新增逆向关联模型 `ReturnLogisticsLink` 在 `backend/internal/entity/models/after_sales/reverse_logistics_link.go`
- [X] T008 在 `backend/cmd/database/migrate/migrate.go` 注册售后域模型迁移
- [X] T009 [P] 实现售后仓储基础封装（含租户事务）在 `backend/internal/entity/repository/after_sales/{case_repository.go,timeline_repository.go,evidence_repository.go,decision_repository.go,reverse_logistics_link_repository.go}`
- [X] T010 [P] 定义售后状态机与校验器在 `backend/internal/services/admin/after_sales/state_machine.go`
- [X] T011 [P] 新增售后 RBAC 资源与权限映射在 `backend/internal/transport/http/admin/after_sales/rbac.go`
- [X] T012 挂载 admin/mini-app 售后路由入口在 `backend/internal/transport/http/admin/routes.go` 与 `backend/internal/transport/http/miniapp/router.go`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 客户提交售后申请并跟踪进度 (Priority: P1) 🎯 MVP

**Goal**: 客户可提交仅退款/退货退款/换货申请，并查询列表与详情进度。

**Independent Test**: 客户完成“提交申请 -> 查看列表 -> 查看详情时间线”闭环，且重复进行中申请被拒绝。

### Implementation for User Story 1

- [X] T013 [US1] 实现 mini-app 申请创建服务（窗口校验、防重、初始轨迹）在 `backend/internal/services/miniapp/after_sales/case_service.go`
- [X] T014 [US1] 实现 mini-app 售后查询服务（列表/详情/时间线）在 `backend/internal/services/miniapp/after_sales/query_service.go`
- [X] T015 [US1] 实现 mini-app 售后 Handler 在 `backend/internal/transport/http/miniapp/after_sales/handler.go`
- [X] T016 [US1] 注册 mini-app 售后路由在 `backend/internal/transport/http/miniapp/after_sales/routes.go`
- [X] T017 [US1] 实现客户侧售后 API composable 在 `web-admin/app/composables/api/useAfterSales.ts`
- [X] T018 [US1] 实现客户侧售后入口页面（申请+列表+详情）在 `web-admin/app/pages/customer/returns-portal.vue`
- [X] T019 [US1] 补充客户侧售后文案键值在 `web-admin/app/i18n/{zh-CN.json,en-US.json}`
- [X] T020 [US1] 对齐 US1 接口合同（请求/响应字段）在 `specs/012-after-sales-rma/contracts/openapi.yaml`

**Checkpoint**: User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - 运营审核与标准化流转 (Priority: P1)

**Goal**: 后台完成售后单受理、审核通过/拒绝、完结/关闭，并沉淀审计轨迹。

**Independent Test**: 运营可对售后单执行完整状态流转，非法跳转被拦截且轨迹完整可查。

### Implementation for User Story 2

- [X] T021 [US2] 实现 admin 售后工作台服务（检索、受理、审核、完结、关闭）在 `backend/internal/services/admin/after_sales/case_service.go`
- [X] T022 [US2] 实现审核决策与轨迹写入服务在 `backend/internal/services/admin/after_sales/decision_service.go`
- [X] T023 [US2] 实现 admin 售后 Handler 在 `backend/internal/transport/http/admin/after_sales/handler.go`
- [X] T024 [US2] 注册 admin 售后路由在 `backend/internal/transport/http/admin/after_sales/routes.go`
- [X] T025 [US2] 实现 admin 售后看板统计服务在 `backend/internal/services/admin/after_sales/dashboard_service.go`
- [X] T026 [US2] 实现 admin 售后看板接口在 `backend/internal/transport/http/admin/after_sales/dashboard_handler.go`
- [X] T027 [US2] 实现运营侧售后页面（列表/详情/流转动作）在 `web-admin/app/pages/market/after-sales.vue`
- [X] T028 [US2] 补充运营侧售后文案键值在 `web-admin/app/i18n/{zh-CN.json,en-US.json}`
- [X] T029 [US2] 对齐 US2 接口合同（admin actions + dashboard）在 `specs/012-after-sales-rma/contracts/openapi.yaml`
- [X] T030 [US2] 明确终态字段冻结策略并在服务层实现字段只读约束在 `backend/internal/services/admin/after_sales/case_service.go`

**Checkpoint**: User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - 订单/支付/履约最小联动 (Priority: P2)

**Goal**: 完成售后与订单/支付/逆向物流的最小一致性联动与防重。

**Independent Test**: 审核通过后可看到订单侧售后标记，退款重复申请受限，退货类可绑定逆向物流并回写状态。

### Implementation for User Story 3

- [X] T031 [US3] 实现订单售后标记同步服务在 `backend/internal/services/admin/after_sales/order_sync_service.go`
- [X] T032 [US3] 实现退款防重联动逻辑在 `backend/internal/services/admin/after_sales/payment_guard_service.go`
- [X] T033 [US3] 实现逆向物流关联服务在 `backend/internal/services/admin/after_sales/reverse_link_service.go`
- [X] T034 [US3] 实现逆向关联 admin 接口在 `backend/internal/transport/http/admin/after_sales/reverse_logistics_handler.go`
- [X] T035 [US3] 在订单详情页增加售后进度展示区块在 `web-admin/app/pages/market/orders/[id].vue`
- [X] T036 [US3] 对齐 US3 接口合同（reverse-logistics link + 联动返回）在 `specs/012-after-sales-rma/contracts/openapi.yaml`
- [X] T037 [US3] 固化换货首版边界（禁止自动补发/库存自动占用）在 `backend/internal/services/admin/after_sales/{case_service.go,order_sync_service.go}`

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 跨故事收口、质量门禁与验收记录

- [ ] T038 [P] 补充售后领域审计/观测埋点在 `backend/internal/observability/after_sales/events.go`
- [ ] T039 实现统一失败原因分类与错误码映射在 `backend/internal/contracts/{error_codes.go,error_mapping.go}` 与 `backend/internal/transport/http/admin/after_sales/handler.go`
- [ ] T040 同步 OpenAPI 错误响应示例（校验失败/状态冲突/权限不足/数据不存在）在 `specs/012-after-sales-rma/contracts/openapi.yaml`
- [ ] T041 同步 docs/plan 规划文档（reverse-logistics 与 returns-portal）在 `docs/plan/fulfillment/reverse-logistics.md` 与 `docs/plan/customer/returns-portal.md`
- [ ] T042 执行后端回归并记录结果在 `specs/012-after-sales-rma/quickstart.md`（命令：`make test`）
- [ ] T043 执行前端回归并记录结果在 `specs/012-after-sales-rma/quickstart.md`（命令：`make test-admin-ci`）
- [ ] T044 执行端到端冒烟并记录结果在 `specs/012-after-sales-rma/quickstart.md`（按 US1→US2→US3 顺序）
- [ ] T045 记录 SC 指标验收结果（5 秒可见性、非法流转拦截率、处理时长样本）在 `specs/012-after-sales-rma/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - 可通过种子数据独立验证审核流转
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - 依赖 US2 的审核结果触发跨域联动

### Within Each User Story

- Models/Repositories before Services
- Services before Endpoints
- Endpoints before UI integration
- Contract alignment before final checkpoint

### Parallel Opportunities

- T002 与 T003 可并行
- T005/T006/T007/T009/T010/T011 可并行
- T017 与 T018 可并行
- T025 与 T027 可并行
- T031/T032/T033 可并行
- T038 与 T041 可并行

---

## Parallel Example: User Story 1

```bash
Task: "实现 mini-app 申请创建服务 in backend/internal/services/miniapp/after_sales/case_service.go"
Task: "实现 mini-app 查询服务 in backend/internal/services/miniapp/after_sales/query_service.go"
Task: "实现客户侧售后页面 in web-admin/app/pages/customer/returns-portal.vue"
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
   - Developer A: User Story 1（mini-app + customer portal）
   - Developer B: User Story 2（admin case flow + dashboard）
   - Developer C: User Story 3（order/payment/reverse linkage）

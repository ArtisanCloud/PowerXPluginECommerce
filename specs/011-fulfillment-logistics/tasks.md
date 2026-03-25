# Tasks: 履约与物流全模块闭环

**Input**: Design documents from `/specs/011-fulfillment-logistics/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 本任务清单包含后端回归、前端 lint/build、RBAC/越权回归、NFR 验证与全链路冒烟任务（非严格 TDD 先测后码）。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立履约模块文档与代码目录骨架，确保后续分阶段开发可并行。

- [X] T001 创建履约模块目录骨架（`backend/internal/{entity/models,entity/repository,services,transport/http/admin,observability}/{logistics,fulfillment,reverse}/`）
- [X] T002 [P] 创建 web-admin 页面与 API 目录骨架（`web-admin/app/pages/shipping/`、`web-admin/app/composables/api/`）
- [X] T003 [P] 在 `specs/011-fulfillment-logistics/contracts/fulfillment.openapi.yaml` 增补统一错误响应与幂等字段说明（保持与当前合同一致）

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 多租户、权限、审计、事件与公共模型能力（阻塞所有故事）

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 定义物流域模型常量与表名映射（`backend/internal/entity/models/model.go` + `backend/internal/entity/models/logistics/*.go`）
- [ ] T005 [P] 实现物流域 Repository 基类与 tenant 事务接入（`backend/internal/entity/repository/logistics/*.go`）
- [ ] T006 [P] 实现履约/逆向域 Repository 基类与 tenant 事务接入（`backend/internal/entity/repository/fulfillment/*.go`、`backend/internal/entity/repository/reverse/*.go`）
- [ ] T007 注册迁移入口并补充迁移说明（`backend/cmd/database/migrate/migrate.go`、`specs/011-fulfillment-logistics/quickstart.md`）
- [ ] T008 [P] 定义 RBAC scope 与 admin 路由挂载（`backend/internal/transport/http/admin/*/rbac.go`、`backend/internal/transport/http/admin/*/routes.go`）
- [ ] T009 [P] 实现履约域审计/事件基础能力（`backend/internal/observability/{logistics,fulfillment,reverse}/*.go`）

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 完成标准发货履约 (Priority: P1) 🎯 MVP

**Goal**: 打通承运商、运费模板、运单创建与轨迹同步的正向履约闭环。

**Independent Test**: 可独立完成“承运商配置→模板发布→创建运单→轨迹更新到签收”的端到端流程。

### Implementation for User Story 1

- [ ] T010 [P] [US1] 实现 Carrier/CarrierService 模型与仓储（`backend/internal/entity/models/logistics/carrier*.go`、`backend/internal/entity/repository/logistics/carrier*_repository.go`）
- [ ] T011 [P] [US1] 实现 RateTemplate/RateZone 模型与仓储（`backend/internal/entity/models/logistics/rate_template*.go`、`backend/internal/entity/repository/logistics/rate_template*_repository.go`）
- [ ] T012 [P] [US1] 实现 Waybill/TrackingEvent 模型与仓储（`backend/internal/entity/models/logistics/waybill*.go`、`backend/internal/entity/repository/logistics/waybill*_repository.go`）
- [ ] T013 [US1] 实现承运商管理 Service（增改停用/连通性测试）（`backend/internal/services/admin/logistics/carrier_service.go`）
- [ ] T014 [US1] 实现运费模板 Service（草稿/发布/只读约束）（`backend/internal/services/admin/logistics/rate_template_service.go`）
- [ ] T015 [US1] 实现运单与轨迹 Service（订单触发主入口、人工补录、状态源优先级）（`backend/internal/services/admin/logistics/waybill_service.go`）
- [ ] T016 [US1] 实现 webhook 幂等处理（事件ID+运单号）（`backend/internal/services/admin/logistics/webhook_service.go`）
- [ ] T017 [US1] 实现物流 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [ ] T018 [US1] 对齐 shipping 页面 API 调用（承运商/模板/运单）并打通列表与详情（`web-admin/app/pages/shipping/{carriers.vue,templates.vue,waybills.vue}`、`web-admin/app/composables/api/useLogistics.ts`）
- [ ] T019 [US1] 增加 US1 回归测试（仓储+service 单测）（`backend/internal/services/admin/logistics/*_test.go`、`backend/internal/entity/repository/logistics/*_test.go`）

**Checkpoint**: User Story 1 should be fully functional and independently testable

---

## Phase 4: User Story 2 - 组织仓内履约任务执行 (Priority: P2)

**Goal**: 订单到履约任务执行闭环，支持任务推进与异常上报。

**Independent Test**: 可独立完成“生成任务→推进状态→提交异常→查看日志”的闭环流程。

### Implementation for User Story 2

- [ ] T020 [P] [US2] 实现 FulfillmentTask/FulfillmentTaskLog 模型与仓储（`backend/internal/entity/models/fulfillment/*.go`、`backend/internal/entity/repository/fulfillment/*_repository.go`）
- [ ] T021 [P] [US2] 实现 FulfillmentException 模型与仓储（`backend/internal/entity/models/fulfillment/exception.go`、`backend/internal/entity/repository/fulfillment/exception_repository.go`）
- [ ] T022 [US2] 实现任务编排与状态推进 Service（`backend/internal/services/admin/fulfillment/task_service.go`）
- [ ] T023 [US2] 实现异常处理与 24h 自动升级 Service（`backend/internal/services/admin/fulfillment/exception_service.go`）
- [ ] T024 [US2] 实现履约任务 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/fulfillment/{handler.go,routes.go,dto.go}`）
- [ ] T025 [US2] 在 web-admin 对齐任务看板与异常入口（`web-admin/app/pages/shipping/tasks.vue`）
- [ ] T026 [US2] 增加 US2 回归测试（任务状态机、异常升级）（`backend/internal/services/admin/fulfillment/*_test.go`）

**Checkpoint**: User Stories 1 and 2 both work independently

---

## Phase 5: User Story 3 - 处理逆向物流与补偿 (Priority: P3)

**Goal**: 打通逆向运单创建、回仓记录、补偿信息输出。

**Independent Test**: 可独立完成“售后审批通过→逆向运单→回仓结论→补偿衔接信息输出”。

### Implementation for User Story 3

- [ ] T027 [P] [US3] 实现 ReverseWaybill 与相关日志模型/仓储（`backend/internal/entity/models/reverse/*.go`、`backend/internal/entity/repository/reverse/*_repository.go`）
- [ ] T028 [US3] 实现逆向物流 Service（创建、轨迹、回仓结论、补偿衔接）（`backend/internal/services/admin/reverse/waybill_service.go`）
- [ ] T029 [US3] 实现逆向 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/reverse/{handler.go,routes.go,dto.go}`）
- [ ] T030 [US3] 对齐 web-admin 逆向入口与详情展示（`web-admin/app/pages/shipping/reverse-waybills.vue`）
- [ ] T031 [US3] 增加 US3 回归测试（逆向状态流转、回仓结论约束）（`backend/internal/services/admin/reverse/*_test.go`）

**Checkpoint**: User Stories 1~3 are independently functional

---

## Phase 6: User Story 4 - 管理第三方承运商接入 (Priority: P3)

**Goal**: 统一第三方承运商适配边界，不改业务入口语义。

**Independent Test**: 可独立完成“配置第三方→连通性验证→创建运单/拉取轨迹”的能力验证。

### Implementation for User Story 4

- [ ] T032 [P] [US4] 实现第三方适配器接口与默认实现骨架（`backend/internal/services/admin/logistics/integrations/{adapter.go,self_adapter.go}`）
- [ ] T033 [P] [US4] 实现 provider 配置与服务编码映射存取（`backend/internal/services/admin/logistics/integrations/config_service.go`）
- [ ] T034 [US4] 将 waybill/track 流程接入适配器分发（`backend/internal/services/admin/logistics/waybill_service.go`）
- [ ] T035 [US4] 扩展承运商管理页面支持 provider 配置（`web-admin/app/pages/shipping/carriers.vue`）
- [ ] T036 [US4] 增加 US4 回归测试（连通性失败、回调幂等、provider 切换）（`backend/internal/services/admin/logistics/integrations/*_test.go`）

**Checkpoint**: All user stories should now be independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 跨故事收口与发布前验证

- [ ] T037 [P] 更新履约文档与索引（`docs/plan/fulfillment/*.md` 与 `specs/011-fulfillment-logistics/quickstart.md`）
- [ ] T038 执行后端回归（`go test ./...`，至少覆盖 logistics/fulfillment/reverse 相关包）
- [ ] T039 执行前端 lint 门禁（`cd web-admin && npm run lint -- --max-warnings=0`）
- [ ] T040 执行前端构建回归（`cd web-admin && npm run build`）
- [ ] T041 执行 RBAC/越权回归（覆盖允许/拒绝/跨租户访问审计）（`backend/internal/transport/http/admin/**/*_test.go`、`backend/internal/services/admin/**/*_test.go`）
- [ ] T042 执行 NFR 验证（轨迹 60 秒可见性、迁移幂等/回滚、结构化日志字段检查）并记录结果（`specs/011-fulfillment-logistics/quickstart.md`、`backend/internal/observability/**/*_test.go`）
- [ ] T043 执行 quickstart 全链路冒烟并记录结果（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: Depend on Foundational completion
  - 建议按 P1 → P2 → P3(P3.1/P3.2)推进
- **Polish (Phase 7)**: Depends on all target user stories completion

### User Story Dependencies

- **US1 (P1)**: 可在 Foundational 后立即开始，是 MVP 主链路
- **US2 (P2)**: 依赖 US1 的订单到运单基础语义，但可独立验收任务/异常能力
- **US3 (P3)**: 依赖售后上下文，可独立于 US2 实现
- **US4 (P3)**: 依赖 US1 的运单/轨迹主流程，作为适配扩展层

### Within Each User Story

- 模型/仓储优先于 Service
- Service 优先于 Handler/UI
- 核心实现完成后再做故事内回归

### Parallel Opportunities

- Phase 2 中标记 `[P]` 的任务可并行
- US1 的模型任务（T010/T011/T012）可并行
- US2 的模型任务（T020/T021）可并行
- US4 的适配器与配置任务（T032/T033）可并行

---

## Parallel Example: User Story 1

```bash
# 并行实现 US1 模型与仓储
Task: "T010 [US1] Carrier/CarrierService 模型与仓储"
Task: "T011 [US1] RateTemplate/RateZone 模型与仓储"
Task: "T012 [US1] Waybill/TrackingEvent 模型与仓储"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1 与 Phase 2
2. 完成 Phase 3 (US1)
3. 立即执行 T019 + T039/T040/T043
4. 通过后可先交付 M1

### Incremental Delivery

1. M1: US1（正向履约闭环）
2. M2: US2（任务与异常）
3. M3-A: US3（逆向闭环）
4. M3-B: US4（三方适配）
5. 最后执行 Phase 7 收口

### Parallel Team Strategy

- 基础阶段完成后：
  - 开发 A：US1/US4（物流主链 + 适配）
  - 开发 B：US2（仓内任务异常）
  - 开发 C：US3（逆向履约）

---

## Notes

- `[P]` 任务需确保不同文件且无未完成依赖
- 每个用户故事可独立测试与验收
- 合并前必须通过 T038/T039/T040/T041/T042/T043

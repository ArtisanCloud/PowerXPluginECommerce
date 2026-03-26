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

- [X] T004 定义物流域模型常量与表名映射（`backend/internal/entity/models/model.go` + `backend/internal/entity/models/logistics/*.go`）
- [X] T005 [P] 实现物流域 Repository 基类与 tenant 事务接入（`backend/internal/entity/repository/logistics/*.go`）
- [X] T006 [P] 实现履约/逆向域 Repository 基类与 tenant 事务接入（`backend/internal/entity/repository/fulfillment/*.go`、`backend/internal/entity/repository/reverse/*.go`）
- [X] T007 注册迁移入口并补充迁移说明（`backend/cmd/database/migrate/migrate.go`、`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T008 [P] 定义 RBAC scope 与 admin 路由挂载（`backend/internal/transport/http/admin/*/rbac.go`、`backend/internal/transport/http/admin/*/routes.go`）
- [X] T009 [P] 实现履约域审计/事件基础能力（`backend/internal/observability/{logistics,fulfillment,reverse}/*.go`）

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 完成标准发货履约 (Priority: P1) 🎯 MVP

**Goal**: 打通承运商、运费模板、运单创建与轨迹同步的正向履约闭环。

**Independent Test**: 可独立完成“承运商配置→模板发布→创建运单→轨迹更新到签收”的端到端流程。

### Implementation for User Story 1

- [X] T010 [P] [US1] 实现 Carrier/CarrierService 模型与仓储（`backend/internal/entity/models/logistics/carrier*.go`、`backend/internal/entity/repository/logistics/carrier*_repository.go`）
- [X] T011 [P] [US1] 实现 RateTemplate/RateZone 模型与仓储（`backend/internal/entity/models/logistics/rate_template*.go`、`backend/internal/entity/repository/logistics/rate_template*_repository.go`）
- [X] T012 [P] [US1] 实现 Waybill/TrackingEvent 模型与仓储（`backend/internal/entity/models/logistics/waybill*.go`、`backend/internal/entity/repository/logistics/waybill*_repository.go`）
- [X] T013 [US1] 实现承运商管理 Service（增改停用/连通性测试）（`backend/internal/services/admin/logistics/carrier_service.go`）
- [X] T014 [US1] 实现运费模板 Service（草稿/发布/只读约束）（`backend/internal/services/admin/logistics/rate_template_service.go`）
- [X] T015 [US1] 实现运单与轨迹 Service（订单触发主入口、人工补录、状态源优先级）（`backend/internal/services/admin/logistics/waybill_service.go`）
- [X] T016 [US1] 实现 webhook 幂等处理（事件ID+运单号）（`backend/internal/services/admin/logistics/webhook_service.go`）
- [X] T017 [US1] 实现物流 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T018 [US1] 对齐 shipping 页面 API 调用（承运商/模板/运单）并打通列表与详情（`web-admin/app/pages/shipping/{carriers.vue,templates.vue,waybills.vue}`、`web-admin/app/composables/api/useLogistics.ts`）
- [X] T019 [US1] 增加 US1 回归测试（仓储+service 单测）（`backend/internal/services/admin/logistics/*_test.go`、`backend/internal/entity/repository/logistics/*_test.go`）

**Checkpoint**: User Story 1 should be fully functional and independently testable

---

## Phase 4: User Story 2 - 组织仓内履约任务执行 (Priority: P2)

**Goal**: 订单到履约任务执行闭环，支持任务推进与异常上报。

**Independent Test**: 可独立完成“生成任务→推进状态→提交异常→查看日志”的闭环流程。

### Implementation for User Story 2

- [X] T020 [P] [US2] 实现 FulfillmentTask/FulfillmentTaskLog 模型与仓储（`backend/internal/entity/models/fulfillment/*.go`、`backend/internal/entity/repository/fulfillment/*_repository.go`）
- [X] T021 [P] [US2] 实现 FulfillmentException 模型与仓储（`backend/internal/entity/models/fulfillment/exception.go`、`backend/internal/entity/repository/fulfillment/exception_repository.go`）
- [X] T022 [US2] 实现任务编排与状态推进 Service（`backend/internal/services/admin/fulfillment/task_service.go`）
- [X] T023 [US2] 实现异常处理与 24h 自动升级 Service（`backend/internal/services/admin/fulfillment/exception_service.go`）
- [X] T024 [US2] 实现履约任务 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/fulfillment/{handler.go,routes.go,dto.go}`）
- [X] T025 [US2] 在 web-admin 对齐任务看板与异常入口（`web-admin/app/pages/shipping/tasks.vue`）
- [X] T026 [US2] 增加 US2 回归测试（任务状态机、异常升级）（`backend/internal/services/admin/fulfillment/*_test.go`）

**Checkpoint**: User Stories 1 and 2 both work independently

---

## Phase 5: User Story 3 - 处理逆向物流与补偿 (Priority: P3)

**Goal**: 打通逆向运单创建、回仓记录、补偿信息输出。

**Independent Test**: 可独立完成“售后审批通过→逆向运单→回仓结论→补偿衔接信息输出”。

### Implementation for User Story 3

- [X] T027 [P] [US3] 实现 ReverseWaybill 与相关日志模型/仓储（`backend/internal/entity/models/reverse/*.go`、`backend/internal/entity/repository/reverse/*_repository.go`）
- [X] T028 [US3] 实现逆向物流 Service（创建、轨迹、回仓结论、补偿衔接）（`backend/internal/services/admin/reverse/waybill_service.go`）
- [X] T029 [US3] 实现逆向 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/reverse/{handler.go,routes.go,dto.go}`）
- [X] T030 [US3] 对齐 web-admin 逆向入口与详情展示（`web-admin/app/pages/shipping/reverse-waybills.vue`）
- [X] T031 [US3] 增加 US3 回归测试（逆向状态流转、回仓结论约束）（`backend/internal/services/admin/reverse/*_test.go`）

**Checkpoint**: User Stories 1~3 are independently functional

---

## Phase 6: User Story 4 - 管理第三方承运商接入 (Priority: P3)

**Goal**: 统一第三方承运商适配边界，不改业务入口语义。

**Independent Test**: 可独立完成“配置第三方→连通性验证→创建运单/拉取轨迹”的能力验证。

### Implementation for User Story 4

- [X] T032 [P] [US4] 实现第三方适配器接口与默认实现骨架（`backend/internal/services/admin/logistics/integrations/{adapter.go,self_adapter.go}`）
- [X] T033 [P] [US4] 实现 provider 配置与服务编码映射存取（`backend/internal/services/admin/logistics/integrations/config_service.go`）
- [X] T034 [US4] 将 waybill/track 流程接入适配器分发（`backend/internal/services/admin/logistics/waybill_service.go`）
- [X] T035 [US4] 扩展承运商管理页面支持 provider 配置（`web-admin/app/pages/shipping/carriers.vue`）
- [X] T036 [US4] 增加 US4 回归测试（连通性失败、回调幂等、provider 切换）（`backend/internal/services/admin/logistics/integrations/*_test.go`）

**Checkpoint**: All user stories should now be independently functional

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 跨故事收口与发布前验证

- [X] T037 [P] 更新履约文档与索引（`docs/plan/fulfillment/*.md` 与 `specs/011-fulfillment-logistics/quickstart.md`）
- [X] T038 执行后端回归（`go test ./...`，至少覆盖 logistics/fulfillment/reverse 相关包）
- [X] T039 执行前端 lint 门禁（`cd web-admin && npm run lint -- --max-warnings=0`）
- [X] T040 执行前端构建回归（`cd web-admin && npm run build`）
- [X] T041 执行 RBAC/越权回归（覆盖允许/拒绝/跨租户访问审计）（`backend/internal/transport/http/admin/**/*_test.go`、`backend/internal/services/admin/**/*_test.go`）
- [X] T042 执行 NFR 验证（轨迹 60 秒可见性、迁移幂等/回滚、结构化日志字段检查）并记录结果（`specs/011-fulfillment-logistics/quickstart.md`、`backend/internal/observability/**/*_test.go`）
- [X] T043 执行 quickstart 全链路冒烟并记录结果（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Phase 8: Iteration-2 - User Story 5/6/7（履约增强）

**Purpose**: 在既有履约闭环基础上补齐高频运营能力（部分发货、多包裹、波次批量、成本对账）。

### User Story 5 - 一单多包裹与部分发货（P1）

- [X] T044 [P] [US5] 扩展运单模型支持订单多包裹标识与包裹序号（`backend/internal/entity/models/logistics/waybill*.go`）
- [X] T045 [US5] 实现订单履约聚合状态计算（部分发货/全部发货）（`backend/internal/services/admin/logistics/waybill_service.go`）
- [X] T046 [US5] 扩展运单创建接口支持包裹级明细（`backend/internal/transport/http/admin/logistics/{dto.go,handler.go}`）
- [X] T047 [US5] 在运单页面展示同订单多包裹分组与聚合状态（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T048 [US5] 增加 US5 回归测试（部分发货、重复补发幂等、状态聚合一致性）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 6 - 波次拣货与批量履约（P2）

- [X] T049 [P] [US6] 新增波次模型与仓储（`backend/internal/entity/models/fulfillment/wave*.go`、`backend/internal/entity/repository/fulfillment/wave*_repository.go`）
- [X] T050 [US6] 实现波次创建/挂接/批量推进服务（`backend/internal/services/admin/fulfillment/wave_service.go`）
- [X] T051 [US6] 实现波次 admin HTTP Handler 与路由（`backend/internal/transport/http/admin/fulfillment/{wave_handler.go,routes.go,dto.go}`）
- [X] T052 [US6] 新增波次管理页面（创建、批量执行、失败回执）（`web-admin/app/pages/shipping/waves.vue`）
- [X] T053 [US6] 增加 US6 回归测试（批量部分失败、重分配、审计字段）（`backend/internal/services/admin/fulfillment/*_test.go`）

### User Story 7 - 履约成本与承运商对账（P3）

- [X] T054 [P] [US7] 扩展运单成本字段与对账投影仓储（`backend/internal/entity/models/logistics/waybill*.go`、`backend/internal/entity/repository/logistics/billing*_repository.go`）
- [X] T055 [US7] 实现成本回填与差异计算服务（`backend/internal/services/admin/logistics/billing_service.go`）
- [X] T056 [US7] 实现承运商账单查询与导出接口（`backend/internal/transport/http/admin/logistics/{billing_handler.go,routes.go,dto.go}`）
- [X] T057 [US7] 新增对账页面（聚合统计、差异明细、导出入口）（`web-admin/app/pages/shipping/billing.vue`）
- [X] T058 [US7] 增加 US7 回归测试（差异计算准确性、导出参数校验、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-2 Polish

- [X] T059 [P] 更新 M4 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T060 执行 M4 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment`）
- [X] T061 执行 M4 前端构建与页面回归（`cd web-admin && npm run build`）

---

## Phase 9: Iteration-3 - 履约运营增强（Backlog）

**Purpose**: 在 M4 基础上补齐仓内效率、服务商运营与成本闭环能力。

### User Story 8 - 批量面单打印与补打（P1）

- [X] T062 [P] [US8] 增加面单打印任务模型与仓储（`backend/internal/entity/models/logistics/label_print*.go`、`backend/internal/entity/repository/logistics/label_print*_repository.go`）
- [X] T063 [US8] 实现批量打印/补打服务（失败重试队列）（`backend/internal/services/admin/logistics/label_print_service.go`）
- [X] T064 [US8] 实现面单打印 admin 接口（`backend/internal/transport/http/admin/logistics/{label_print_handler.go,routes.go,dto.go}`）
- [X] T065 [US8] 新增面单打印页面（批量选择、状态回执、补打入口）（`web-admin/app/pages/shipping/labels.vue`）
- [X] T066 [US8] 增加 US8 回归测试（批量失败重试、幂等补打）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 9 - 承运商 SLA 看板（P2）

- [X] T067 [P] [US9] 增加 SLA 指标投影与聚合仓储（`backend/internal/entity/repository/logistics/sla*_repository.go`）
- [X] T068 [US9] 实现 SLA 统计服务（揽收时效/签收时效/异常率）（`backend/internal/services/admin/logistics/sla_service.go`）
- [X] T069 [US9] 实现 SLA 看板接口（`backend/internal/transport/http/admin/logistics/{sla_handler.go,routes.go,dto.go}`）
- [X] T070 [US9] 新增 SLA 看板页面（`web-admin/app/pages/shipping/sla.vue`）
- [X] T071 [US9] 增加 US9 回归测试（窗口聚合准确性、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 10 - 运费模板模拟器（P2）

- [ ] T072 [P] [US10] 扩展运费规则解析与试算模型（`backend/internal/entity/models/logistics/rate_quote*.go`）
- [ ] T073 [US10] 实现运费试算服务（地址/重量/件数输入）（`backend/internal/services/admin/logistics/rate_quote_service.go`）
- [ ] T074 [US10] 实现运费模拟接口（`backend/internal/transport/http/admin/logistics/{rate_quote_handler.go,routes.go,dto.go}`）
- [ ] T075 [US10] 在模板页面增加“模拟试算”交互（`web-admin/app/pages/shipping/templates.vue`）
- [ ] T076 [US10] 增加 US10 回归测试（规则命中顺序、边界输入）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 11 - 逆向质检规则引擎（P3）

- [ ] T077 [P] [US11] 增加逆向质检规则模型与仓储（`backend/internal/entity/models/reverse/inspection_rule*.go`、`backend/internal/entity/repository/reverse/inspection*_repository.go`）
- [ ] T078 [US11] 实现质检判定服务（可二销/报损/维修）（`backend/internal/services/admin/reverse/inspection_service.go`）
- [ ] T079 [US11] 实现质检规则与判定接口（`backend/internal/transport/http/admin/reverse/{inspection_handler.go,routes.go,dto.go}`）
- [ ] T080 [US11] 在逆向页面展示质检结果与建议（`web-admin/app/pages/shipping/reverse-waybills.vue`）
- [ ] T081 [US11] 增加 US11 回归测试（规则优先级、重复判定幂等）（`backend/internal/services/admin/reverse/*_test.go`）

### User Story 12 - 波次智能分组策略（P2）

- [ ] T082 [P] [US12] 增加波次策略配置模型与仓储（`backend/internal/entity/models/fulfillment/wave_strategy*.go`、`backend/internal/entity/repository/fulfillment/wave_strategy*_repository.go`）
- [ ] T083 [US12] 实现智能分组服务（仓/承运商/时段/优先级）（`backend/internal/services/admin/fulfillment/wave_strategy_service.go`）
- [ ] T084 [US12] 实现策略管理与预览接口（`backend/internal/transport/http/admin/fulfillment/{wave_strategy_handler.go,routes.go,dto.go}`）
- [ ] T085 [US12] 在波次页面增加“智能分组”入口（`web-admin/app/pages/shipping/waves.vue`）
- [ ] T086 [US12] 增加 US12 回归测试（分组稳定性、部分异常隔离）（`backend/internal/services/admin/fulfillment/*_test.go`）

### User Story 13 - 对账异常工单闭环（P3）

- [ ] T087 [P] [US13] 增加对账异常工单模型与仓储（`backend/internal/entity/models/logistics/billing_case*.go`、`backend/internal/entity/repository/logistics/billing_case*_repository.go`）
- [ ] T088 [US13] 实现异常工单状态机服务（确认/申诉/核销）（`backend/internal/services/admin/logistics/billing_case_service.go`）
- [ ] T089 [US13] 实现异常工单接口（`backend/internal/transport/http/admin/logistics/{billing_case_handler.go,routes.go,dto.go}`）
- [ ] T090 [US13] 在对账页面增加异常工单流转视图（`web-admin/app/pages/shipping/billing.vue`）
- [ ] T091 [US13] 增加 US13 回归测试（状态流转约束、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 14 - 物流通知中心（P3）

- [ ] T092 [P] [US14] 增加物流通知模板与发送记录模型（`backend/internal/entity/models/logistics/notification*.go`）
- [ ] T093 [US14] 实现通知编排服务（发货/派送/签收/异常）（`backend/internal/services/admin/logistics/notification_service.go`）
- [ ] T094 [US14] 实现通知管理接口（模板管理、发送历史）（`backend/internal/transport/http/admin/logistics/{notification_handler.go,routes.go,dto.go}`）
- [ ] T095 [US14] 新增通知中心页面（`web-admin/app/pages/shipping/notifications.vue`）
- [ ] T096 [US14] 增加 US14 回归测试（模板渲染、重试与幂等）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-3 Polish

- [ ] T097 [P] 更新 M5 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [ ] T098 执行 M5 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`）
- [ ] T099 执行 M5 前端构建与页面回归（`cd web-admin && npm run build`）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: Depend on Foundational completion
  - 建议按 P1 → P2 → P3(P3.1/P3.2)推进
- **Polish (Phase 7)**: Depends on M1~M3 completion
- **Iteration-2 (Phase 8)**: Depends on Phase 7 completion

### User Story Dependencies

- **US1 (P1)**: 可在 Foundational 后立即开始，是 MVP 主链路
- **US2 (P2)**: 依赖 US1 的订单到运单基础语义，但可独立验收任务/异常能力
- **US3 (P3)**: 依赖售后上下文，可独立于 US2 实现
- **US4 (P3)**: 依赖 US1 的运单/轨迹主流程，作为适配扩展层
- **US5 (P1)**: 依赖 US1 的运单主流程，扩展为多包裹与部分发货
- **US6 (P2)**: 依赖 US2 的任务模型，扩展波次批量执行
- **US7 (P3)**: 依赖 US1/US4 的运单与承运商配置，补齐成本对账

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

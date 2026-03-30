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

- [X] T072 [P] [US10] 扩展运费规则解析与试算模型（`backend/internal/entity/models/logistics/rate_quote*.go`）
- [X] T073 [US10] 实现运费试算服务（地址/重量/件数输入）（`backend/internal/services/admin/logistics/rate_quote_service.go`）
- [X] T074 [US10] 实现运费模拟接口（`backend/internal/transport/http/admin/logistics/{rate_quote_handler.go,routes.go,dto.go}`）
- [X] T075 [US10] 在模板页面增加“模拟试算”交互（`web-admin/app/pages/shipping/templates.vue`）
- [X] T076 [US10] 增加 US10 回归测试（规则命中顺序、边界输入）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 11 - 逆向质检规则引擎（P3）

- [X] T077 [P] [US11] 增加逆向质检规则模型与仓储（`backend/internal/entity/models/reverse/inspection_rule*.go`、`backend/internal/entity/repository/reverse/inspection*_repository.go`）
- [X] T078 [US11] 实现质检判定服务（可二销/报损/维修）（`backend/internal/services/admin/reverse/inspection_service.go`）
- [X] T079 [US11] 实现质检规则与判定接口（`backend/internal/transport/http/admin/reverse/{inspection_handler.go,routes.go,dto.go}`）
- [X] T080 [US11] 在逆向页面展示质检结果与建议（`web-admin/app/pages/shipping/reverse-waybills.vue`）
- [X] T081 [US11] 增加 US11 回归测试（规则优先级、重复判定幂等）（`backend/internal/services/admin/reverse/*_test.go`）

### User Story 12 - 波次智能分组策略（P2）

- [X] T082 [P] [US12] 增加波次策略配置模型与仓储（`backend/internal/entity/models/fulfillment/wave_strategy*.go`、`backend/internal/entity/repository/fulfillment/wave_strategy*_repository.go`）
- [X] T083 [US12] 实现智能分组服务（仓/承运商/时段/优先级）（`backend/internal/services/admin/fulfillment/wave_strategy_service.go`）
- [X] T084 [US12] 实现策略管理与预览接口（`backend/internal/transport/http/admin/fulfillment/{wave_strategy_handler.go,routes.go,dto.go}`）
- [X] T085 [US12] 在波次页面增加“智能分组”入口（`web-admin/app/pages/shipping/waves.vue`）
- [X] T086 [US12] 增加 US12 回归测试（分组稳定性、部分异常隔离）（`backend/internal/services/admin/fulfillment/*_test.go`）

### User Story 13 - 对账异常工单闭环（P3）

- [X] T087 [P] [US13] 增加对账异常工单模型与仓储（`backend/internal/entity/models/logistics/billing_case*.go`、`backend/internal/entity/repository/logistics/billing_case*_repository.go`）
- [X] T088 [US13] 实现异常工单状态机服务（确认/申诉/核销）（`backend/internal/services/admin/logistics/billing_case_service.go`）
- [X] T089 [US13] 实现异常工单接口（`backend/internal/transport/http/admin/logistics/{billing_case_handler.go,routes.go,dto.go}`）
- [X] T090 [US13] 在对账页面增加异常工单流转视图（`web-admin/app/pages/shipping/billing.vue`）
- [X] T091 [US13] 增加 US13 回归测试（状态流转约束、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 14 - 物流通知中心（P3）

- [X] T092 [P] [US14] 增加物流通知模板与发送记录模型（`backend/internal/entity/models/logistics/notification*.go`）
- [X] T093 [US14] 实现通知编排服务（发货/派送/签收/异常）（`backend/internal/services/admin/logistics/notification_service.go`）
- [X] T094 [US14] 实现通知管理接口（模板管理、发送历史）（`backend/internal/transport/http/admin/logistics/{notification_handler.go,routes.go,dto.go}`）
- [X] T095 [US14] 新增通知中心页面（`web-admin/app/pages/shipping/notifications.vue`）
- [X] T096 [US14] 增加 US14 回归测试（模板渲染、重试与幂等）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-3 Polish

- [X] T097 [P] 更新 M5 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T098 执行 M5 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`）
- [X] T099 执行 M5 前端构建与页面回归（`cd web-admin && npm run build`）

---

## Phase 10: Iteration-4 - 履约体验与风控增强（Backlog）

**Purpose**: 在 M5 基础上补齐履约承诺、仓配路由、二次派送与物流风控能力，形成“效率 + 体验 + 风险”闭环。

### User Story 15 - 物流承诺时效与预计达（P2）

- [X] T100 [P] [US15] 增加承诺时效/预计达模型与仓储（`backend/internal/entity/models/logistics/eta*.go`、`backend/internal/entity/repository/logistics/eta*_repository.go`）
- [X] T101 [US15] 实现 ETA 计算服务（揽收时效/派送时效/预计达时间）（`backend/internal/services/admin/logistics/eta_service.go`）
- [X] T102 [US15] 实现 ETA 查询接口（`backend/internal/transport/http/admin/logistics/{eta_handler.go,routes.go,dto.go}`）
- [X] T103 [US15] 在运单页面展示“承诺达/预计达”字段（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T104 [US15] 增加 US15 回归测试（时效窗口、跨时区与租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 16 - 仓配路由规则（P2）

- [X] T105 [P] [US16] 增加仓配路由规则模型与仓储（`backend/internal/entity/models/logistics/routing_rule*.go`、`backend/internal/entity/repository/logistics/routing_rule*_repository.go`）
- [X] T106 [US16] 实现仓配路由决策服务（仓/承运商择优、兜底策略）（`backend/internal/services/admin/logistics/routing_service.go`）
- [X] T107 [US16] 实现路由规则管理与预览接口（`backend/internal/transport/http/admin/logistics/{routing_handler.go,routes.go,dto.go}`）
- [X] T108 [US16] 在承运商/运单页增加“路由预览”入口（`web-admin/app/pages/shipping/{carriers.vue,waybills.vue}`）
- [X] T109 [US16] 增加 US16 回归测试（规则优先级、兜底与异常隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 17 - 妥投失败二次派送闭环（P3）

- [X] T110 [P] [US17] 增加二次派送任务模型与仓储（`backend/internal/entity/models/logistics/redelivery*.go`、`backend/internal/entity/repository/logistics/redelivery*_repository.go`）
- [X] T111 [US17] 实现二次派送状态机服务（发起/改址/重派/关闭）（`backend/internal/services/admin/logistics/redelivery_service.go`）
- [X] T112 [US17] 实现二次派送接口（`backend/internal/transport/http/admin/logistics/{redelivery_handler.go,routes.go,dto.go}`）
- [X] T113 [US17] 在运单页增加“失败重派”流程视图（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T114 [US17] 增加 US17 回归测试（状态流转约束、幂等重派）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 18 - 物流风控与黑名单（P3）

- [X] T115 [P] [US18] 增加物流风控规则与黑名单模型（`backend/internal/entity/models/logistics/risk_rule*.go`、`backend/internal/entity/models/logistics/blacklist*.go`）
- [X] T116 [US18] 实现风控评估服务（高风险地址/收件人识别、拦截建议）（`backend/internal/services/admin/logistics/risk_service.go`）
- [X] T117 [US18] 实现风控管理接口（规则、命中记录、人工放行）（`backend/internal/transport/http/admin/logistics/{risk_handler.go,routes.go,dto.go}`）
- [X] T118 [US18] 新增风控页面（规则管理与命中处置）（`web-admin/app/pages/shipping/risk-control.vue`）
- [X] T119 [US18] 增加 US18 回归测试（误拦截豁免、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-4 Polish

- [X] T120 [P] 更新 M6 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T121 执行 M6 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`）
- [X] T122 执行 M6 前端构建与页面回归（`cd web-admin && npm run build`）

---

## Phase 11: Iteration-5 - 承运商真实网关适配层（Backlog）

**Purpose**: 在既有履约主链路上补齐 provider pull 轨迹回流与网关鉴权对齐，形成“后台代理网关 + 页面手动同步”的可调试闭环。

### User Story 19 - 网关轨迹拉取与手动同步（P2）

- [X] T123 [US19] 扩展物流适配器抽象，增加 provider pull 轨迹拉取能力（`backend/internal/services/admin/logistics/integrations/{adapter.go,self_adapter.go}`）
- [X] T124 [US19] 实现 GatewayAdapter（鉴权 scheme 对齐、重试与回退）并默认接管第三方承运商（`backend/internal/services/admin/logistics/integrations/gateway_adapter.go`）
- [X] T125 [US19] 新增运单手动同步轨迹服务与管理端接口（`backend/internal/services/admin/logistics/waybill_service.go`、`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,rbac.go}`）
- [X] T126 [US19] 运单页面接入“同步轨迹”动作与结果提示（`web-admin/app/pages/shipping/waybills.vue`、`web-admin/app/composables/api/useLogistics.ts`）
- [X] T127 [US19] 增加网关适配与轨迹同步回归测试（`backend/internal/services/admin/logistics/integrations/gateway_adapter_test.go`、`backend/internal/services/admin/logistics/waybill_provider_sync_test.go`、`backend/internal/transport/http/admin/logistics/rbac_test.go`）

---

## Phase 12: Iteration-6 - 网关稳定性与运营可观测（Backlog）

**Purpose**: 在已打通的网关适配链路上补齐可观测、失败重试与运营面板，避免“能用但不可运维”。

### User Story 20 - 轨迹同步作业化（P2）

- [X] T128 [US20] 增加轨迹同步作业模型与仓储（`backend/internal/entity/models/logistics/tracking_sync_job*.go`、`backend/internal/entity/repository/logistics/tracking_sync_job*_repository.go`）
- [X] T129 [US20] 实现批量同步作业服务（按承运商/状态筛选运单，分批触发 provider pull）（`backend/internal/services/admin/logistics/tracking_sync_job_service.go`）
- [X] T130 [US20] 新增作业管理接口（创建、查询、取消、重试）（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T131 [US20] 运单页与运营页接入“批量同步任务”入口（`web-admin/app/pages/shipping/{waybills.vue,sla.vue}`）
- [X] T132 [US20] 增加 US20 回归测试（分批幂等、失败重试、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 21 - 网关健康与告警看板（P2）

- [X] T133 [P] [US21] 增加网关请求指标聚合（成功率、P95、4xx/5xx 分布）（`backend/internal/services/admin/logistics/gateway_metrics_service.go`）
- [X] T134 [US21] 实现网关健康接口与 SLA 卡片数据接口（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T135 [US21] 在 SLA 页面新增“网关健康”卡片与告警列表（`web-admin/app/pages/shipping/sla.vue`）
- [X] T136 [US21] 增加 US21 回归测试（窗口聚合正确性、告警阈值触发）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-6 Polish

- [X] T137 [P] 更新 M7 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T138 执行 M7 后端回归（`go test ./internal/services/admin/logistics ./internal/transport/http/admin/logistics ./internal/entity/repository/logistics -count=1`）
- [X] T139 执行 M7 前端构建与页面回归（`make build-admin`）

---

## Phase 13: Iteration-7 - 调度化与成本治理（Backlog）

**Purpose**: 从“可手动运营”升级到“可持续运行”，补齐调度、补偿、成本与配额治理能力。

### User Story 22 - 轨迹同步调度化（P2）

- [X] T140 [US22] 增加同步调度策略模型与仓储（cron、启停、并发上限、租户配额）（`backend/internal/entity/models/logistics/tracking_sync_schedule*.go`、`backend/internal/entity/repository/logistics/tracking_sync_schedule*_repository.go`）
- [X] T141 [US22] 实现调度执行器（按策略触发同步作业、去重窗口、防并发重入）（`backend/internal/services/admin/logistics/tracking_sync_scheduler_service.go`）
- [X] T142 [US22] 实现调度管理接口（创建/更新/启停/立即执行）（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T143 [US22] 在运营页增加“同步计划”管理区（`web-admin/app/pages/shipping/sla.vue`）
- [X] T144 [US22] 增加 US22 回归测试（调度去重、并发上限、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 23 - 网关失败分类与补偿策略（P2）

- [X] T145 [P] [US23] 扩展失败分类模型（4xx/5xx/timeout/auth/contract）与补偿记录（`backend/internal/entity/models/logistics/gateway_failure*.go`）
- [X] T146 [US23] 实现补偿策略服务（指数退避、熔断、降级到手工队列）（`backend/internal/services/admin/logistics/gateway_recovery_service.go`）
- [X] T147 [US23] 实现失败事件查询与补偿触发接口（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T148 [US23] 在运单页增加“失败补偿”操作入口与状态提示（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T149 [US23] 增加 US23 回归测试（错误分类准确性、补偿幂等、熔断恢复）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 24 - 网关成本与调用配额（P3）

- [X] T150 [P] [US24] 增加网关调用账本模型（调用次数、计费单价、成本、配额消耗）（`backend/internal/entity/models/logistics/gateway_usage*.go`）
- [X] T151 [US24] 实现成本与配额聚合服务（按租户/承运商/窗口统计）（`backend/internal/services/admin/logistics/gateway_cost_service.go`）
- [X] T152 [US24] 实现成本/配额看板接口与超额告警接口（`backend/internal/transport/http/admin/logistics/{handler.go,routes.go,dto.go}`）
- [X] T153 [US24] 在 SLA 页新增“成本与配额”卡片（`web-admin/app/pages/shipping/sla.vue`）
- [X] T154 [US24] 增加 US24 回归测试（聚合准确性、阈值告警、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-7 Polish

- [X] T155 [P] 更新 M8 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T156 执行 M8 后端回归（`go test ./internal/services/admin/logistics ./internal/transport/http/admin/logistics ./internal/entity/repository/logistics -count=1`）
- [X] T157 执行 M8 前端构建与页面回归（`make build-admin`）
- [X] T158 执行批量同步压测与记录（100/500 运单批次，统计成功率与 P95）（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Phase 14: Iteration-8 - 履约自动化与财务闭环（Backlog）

**Purpose**: 在 M8 基础上补齐“仓配执行 + 异常编排 + 智能校验 + 联合路由 + 财务核对”，把履约从“可运营”升级为“可规模化自治”。

### User Story 25 - 仓配一体联动（P1）

- [X] T159 [P] [US25] 增加仓配执行实体（出库单、拣货明细、装箱单、履约波次关联）（`backend/internal/entity/models/fulfillment/{outbound*.go,pick*.go,pack*.go}`）
- [X] T160 [US25] 实现仓配联动服务（库存预占→拣货→装箱→出库→运单回写）（`backend/internal/services/admin/fulfillment/warehouse_bridge_service.go`）
- [X] T161 [US25] 实现仓配联动接口（出库任务创建、执行回执、异常回滚）（`backend/internal/transport/http/admin/fulfillment/{warehouse_handler.go,routes.go,dto.go}`）
- [X] T162 [US25] 在发货执行页新增仓配联动面板（`web-admin/app/pages/shipping/tasks.vue`）
- [X] T163 [US25] 增加 US25 回归测试（预占幂等、出库状态一致性、跨租户隔离）（`backend/internal/services/admin/fulfillment/*_test.go`）

### User Story 26 - 物流异常自动编排中心（P2）

- [X] T164 [P] [US26] 增加异常编排规则与执行记录模型（`backend/internal/entity/models/logistics/exception_orchestration*.go`）
- [X] T165 [US26] 实现异常编排服务（延误/拒收/丢件自动建单、自动补偿、SLA 升级）（`backend/internal/services/admin/logistics/exception_orchestration_service.go`）
- [X] T166 [US26] 实现异常编排接口（规则管理、执行重放、手工接管）（`backend/internal/transport/http/admin/logistics/{exception_orchestration_handler.go,routes.go,dto.go}`）
- [X] T167 [US26] 在运单页新增“异常自动化”入口与状态流（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T168 [US26] 增加 US26 回归测试（规则命中准确性、补偿幂等、人工接管优先级）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 27 - 地址智能与可达性校验（P2）

- [X] T169 [P] [US27] 增加地址规范化与可达性缓存模型（`backend/internal/entity/models/logistics/address_validation*.go`）
- [X] T170 [US27] 实现地址智能服务（标准化、风险地址拦截、改址建议）（`backend/internal/services/admin/logistics/address_validation_service.go`）
- [X] T171 [US27] 实现地址校验接口（下单前校验、改址推荐、人工确认）（`backend/internal/transport/http/admin/logistics/{address_validation_handler.go,routes.go,dto.go}`）
- [X] T172 [US27] 在运单创建流程接入地址智能提示（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T173 [US27] 增加 US27 回归测试（命中准确性、误拦截放行、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 28 - 多目标路由引擎 2.0（P2）

- [X] T174 [P] [US28] 扩展路由打分模型（时效/成本/配额/风险权重）（`backend/internal/entity/models/logistics/routing_score*.go`）
- [X] T175 [US28] 实现联合路由服务（多目标打分、兜底策略、实时降级）（`backend/internal/services/admin/logistics/routing_optimizer_service.go`）
- [X] T176 [US28] 实现路由优化接口（策略配置、仿真、命中解释）（`backend/internal/transport/http/admin/logistics/{routing_optimizer_handler.go,routes.go,dto.go}`）
- [X] T177 [US28] 在承运商/运单页新增“联合路由仿真”卡片（`web-admin/app/pages/shipping/{carriers.vue,waybills.vue}`）
- [X] T178 [US28] 增加 US28 回归测试（权重稳定性、降级正确性、异常输入隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 29 - 结算财务闭环（P3）

- [X] T179 [P] [US29] 增加承运商结算批次与差异归因模型（`backend/internal/entity/models/logistics/settlement*.go`）
- [X] T180 [US29] 实现自动核对服务（账单拉取、差异归因、自动建议动作）（`backend/internal/services/admin/logistics/settlement_service.go`）
- [X] T181 [US29] 实现结算闭环接口（批次创建、差异处理、结算确认）（`backend/internal/transport/http/admin/logistics/{settlement_handler.go,routes.go,dto.go}`）
- [X] T182 [US29] 在对账页新增“结算批次”与“归因建议”区块（`web-admin/app/pages/shipping/billing.vue`）
- [X] T183 [US29] 增加 US29 回归测试（归因准确性、状态流转约束、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-8 Polish

- [X] T184 [P] 更新 M9 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T185 执行 M9 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`）
- [X] T186 执行 M9 前端构建与页面回归（`make build-admin`）
- [X] T187 执行仓配-路由-结算全链路冒烟并记录（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Phase 12: User Story 30 - 履约控制塔（Priority: P1）

**Goal**: 建立履约运营总览，统一呈现在途、异常、SLA、成本与告警状态。

**Independent Test**: 可独立完成“筛选仓/承运商/区域→查看核心指标→钻取异常单据→触发处置动作”流程。

### Implementation for User Story 30

- [X] T188 [P] [US30] 增加控制塔聚合快照模型与仓储（`backend/internal/entity/models/logistics/control_tower*.go`、`backend/internal/entity/repository/logistics/control_tower*_repository.go`）
- [X] T189 [US30] 实现控制塔聚合服务（在途/异常/SLA/成本统一聚合与维度钻取）（`backend/internal/services/admin/logistics/control_tower_service.go`）
- [X] T190 [US30] 实现控制塔接口（总览、钻取、告警订阅）（`backend/internal/transport/http/admin/logistics/{control_tower_handler.go,routes.go,dto.go}`）
- [X] T191 [US30] 新增控制塔页面（大盘卡片、趋势图、异常钻取抽屉）（`web-admin/app/pages/shipping/control-tower.vue`）
- [X] T192 [US30] 增加 US30 回归测试（聚合准确性、筛选一致性、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

---

## Phase 13: User Story 31 - 承运商智能分单与容量预留（Priority: P1）

**Goal**: 在发货前按 SLA/成本/容量自动分单，并支持大促前容量预留与超配保护。

**Independent Test**: 可独立完成“配置容量计划→自动分单→容量扣减→超限降级”流程。

### Implementation for User Story 31

- [X] T193 [P] [US31] 增加容量计划与分单决策模型（`backend/internal/entity/models/logistics/capacity_plan*.go`、`backend/internal/entity/models/logistics/allocation_decision*.go`）
- [X] T194 [US31] 实现容量预留与智能分单服务（配额扣减、超限兜底、人工覆盖）（`backend/internal/services/admin/logistics/allocation_service.go`）
- [X] T195 [US31] 实现分单与容量管理接口（计划管理、自动分配、手工改派）（`backend/internal/transport/http/admin/logistics/{allocation_handler.go,routes.go,dto.go}`）
- [X] T196 [US31] 在运单/承运商页新增“智能分单与容量”视图（`web-admin/app/pages/shipping/{waybills.vue,carriers.vue}`）
- [X] T197 [US31] 增加 US31 回归测试（容量扣减幂等、超限保护、规则优先级）（`backend/internal/services/admin/logistics/*_test.go`）

---

## Phase 14: User Story 32 - 末端异常自愈中心（Priority: P2）

**Goal**: 针对延误/拒收/丢件等末端异常，自动触发改派、补发、退款等补偿动作并闭环。

**Independent Test**: 可独立完成“异常触发→自动动作执行→人工接管→闭环归档”流程。

### Implementation for User Story 32

- [X] T198 [P] [US32] 增加末端异常动作编排模型（`backend/internal/entity/models/logistics/lastmile_recovery*.go`）
- [X] T199 [US32] 实现自愈编排服务（策略命中、动作编排、重试与人工接管）（`backend/internal/services/admin/logistics/lastmile_recovery_service.go`）
- [X] T200 [US32] 实现自愈中心接口（规则、执行记录、手工介入）（`backend/internal/transport/http/admin/logistics/{lastmile_recovery_handler.go,routes.go,dto.go}`）
- [X] T201 [US32] 在运单页新增“末端异常自愈”面板（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T202 [US32] 增加 US32 回归测试（动作幂等、重试策略、接管优先级）（`backend/internal/services/admin/logistics/*_test.go`）

---

## Phase 15: User Story 33 - 跨境履约扩展（Priority: P2）

**Goal**: 打通跨境履约关键链路，覆盖清关资料校验、税费预估与国际轨迹标准化。

**Independent Test**: 可独立完成“跨境单创建→资料校验→税费预估→轨迹同步”流程。

### Implementation for User Story 33

- [X] T203 [P] [US33] 增加跨境资料与税费模型（`backend/internal/entity/models/logistics/crossborder*.go`）
- [X] T204 [US33] 实现跨境履约服务（资料校验、税费估算、标准化状态映射）（`backend/internal/services/admin/logistics/crossborder_service.go`）
- [X] T205 [US33] 实现跨境接口（资料管理、预估查询、轨迹映射）（`backend/internal/transport/http/admin/logistics/{crossborder_handler.go,routes.go,dto.go}`）
- [X] T206 [US33] 在运单页新增“跨境履约”区块（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T207 [US33] 增加 US33 回归测试（校验规则、税费边界、状态映射准确性）（`backend/internal/services/admin/logistics/*_test.go`）

---

## Iteration-9 Polish

- [X] T208 [P] 更新 M10 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T209 执行 M10 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`）
- [X] T210 执行 M10 前端构建与页面回归（`make build-admin`）

---

## Phase 16: Iteration-10 - 运营分析与清关治理（Backlog）

**Purpose**: 在 M10 基础上补齐“经营分析 + 财务自动核对 + 跨境清关规则”，将履约从执行闭环升级到治理闭环。

### User Story 34 - 物流履约 KPI 大屏（P2）

- [X] T211 [P] [US34] 增加 KPI 聚合快照与趋势模型（`backend/internal/entity/models/logistics/kpi_snapshot*.go`）
- [X] T212 [US34] 实现 KPI 聚合服务（租户/仓库/承运商维度、同比环比、异常钻取）（`backend/internal/services/admin/logistics/kpi_dashboard_service.go`）
- [X] T213 [US34] 实现 KPI 大屏接口（总览、趋势、钻取、导出）（`backend/internal/transport/http/admin/logistics/{kpi_dashboard_handler.go,routes.go,dto.go}`）
- [X] T214 [US34] 新增 KPI 大屏页面（筛选面板、趋势图、异常明细抽屉）（`web-admin/app/pages/shipping/kpi-dashboard.vue`）
- [X] T215 [US34] 增加 US34 回归测试（聚合准确性、筛选一致性、导出口径正确性）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 35 - 承运商结算自动对账（P2）

- [X] T216 [P] [US35] 增加自动对账批次与三方匹配模型（账单/流水/发票）（`backend/internal/entity/models/logistics/reconciliation*.go`）
- [X] T217 [US35] 实现自动对账服务（匹配规则、差异归因、建议动作）（`backend/internal/services/admin/logistics/reconciliation_service.go`）
- [X] T218 [US35] 实现自动对账接口（批次执行、差异列表、工单流转）（`backend/internal/transport/http/admin/logistics/{reconciliation_handler.go,routes.go,dto.go}`）
- [X] T219 [US35] 在对账页新增“自动对账”区块（批次执行、异常工单、处理动作）（`web-admin/app/pages/shipping/billing.vue`）
- [X] T220 [US35] 增加 US35 回归测试（匹配准确性、状态约束、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 36 - 跨境清关规则中心（P3）

- [X] T221 [P] [US36] 增加清关规则包与版本模型（国家规则、命中策略、启停状态）（`backend/internal/entity/models/logistics/customs_rule*.go`）
- [X] T222 [US36] 实现清关规则服务（规则编排、风险预判、建议动作生成）（`backend/internal/services/admin/logistics/customs_rule_service.go`）
- [X] T223 [US36] 实现清关规则接口（规则管理、预检执行、命中解释）（`backend/internal/transport/http/admin/logistics/{customs_rule_handler.go,routes.go,dto.go}`）
- [X] T224 [US36] 在运单页新增“清关预检”面板（国家规则选择、风险预判结果、人工确认）（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T225 [US36] 增加 US36 回归测试（规则版本切换、命中解释完整性、误拦截放行）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-10 Polish

- [X] T226 [P] 更新 M11 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T227 执行 M11 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`）
- [X] T228 执行 M11 前端构建与页面回归（`make build-admin`）
- [X] T229 执行 M11 关键链路冒烟（US34/US35/US36 核心服务用例）（`go test ./internal/services/admin/logistics -run 'TestKPIDashboardService_OverviewTrendDrilldown|TestReconciliationService_MatchingAndCaseFlow|TestCustomsRuleService_VersionSwitching' -count=1`）
- [X] T230 归档 M11 执行记录与验收结论（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Phase 17: Iteration-11 - 履约稳定性治理（Backlog）

**Purpose**: 在 M11 基础上补齐履约稳定性治理能力，覆盖 SLO 监控、阈值预警与自动限流，降低高峰期接口抖动导致的履约失败。

### User Story 37 - 履约 SLO 守卫与自动限流（P2）

- [X] T231 [P] [US37] 增加 SLO 指标与限流策略模型（窗口、阈值、动作、租户作用域）（`backend/internal/entity/models/logistics/slo_guard*.go`）
- [X] T232 [US37] 实现 SLO 守卫服务（指标聚合、阈值判定、自动限流/恢复）（`backend/internal/services/admin/logistics/slo_guard_service.go`）
- [X] T233 [US37] 实现 SLO 守卫接口（策略管理、状态查询、手动解除限流）（`backend/internal/transport/http/admin/logistics/{slo_guard_handler.go,routes.go,dto.go}`）
- [X] T234 [US37] 新增 SLO 守卫页面区块（策略配置、实时状态、告警与处置）（`web-admin/app/pages/shipping/sla.vue`）
- [X] T235 [US37] 增加 US37 回归测试（阈值命中准确性、限流恢复、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

---

## Phase 18: Iteration-12 - 预测与协同履约（Backlog）

**Purpose**: 在 M11 稳定性治理基础上，补齐“预测 + 根因 + 跨仓协同 + 质量复盘”，将履约从被动响应升级到主动优化。

### User Story 38 - 履约容量预测与动态配额（P2）

- [X] T236 [P] [US38] 增加容量预测与动态配额模型（预测窗口、目标容量、调整策略、置信区间）（`backend/internal/entity/models/logistics/capacity_forecast*.go`）
- [X] T237 [US38] 实现容量预测服务（历史样本聚合、峰值预警、配额建议生成）（`backend/internal/services/admin/logistics/capacity_forecast_service.go`）
- [X] T238 [US38] 实现容量预测接口（预测查询、建议确认、配额下发）（`backend/internal/transport/http/admin/logistics/{capacity_forecast_handler.go,routes.go,dto.go}`）
- [X] T239 [US38] 新增容量预测页面（趋势图、建议面板、一键下发）（`web-admin/app/pages/shipping/capacity-forecast.vue`）
- [X] T240 [US38] 增加 US38 回归测试（预测口径一致性、建议幂等、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 39 - 轨迹异常根因分析中心（P2）

- [X] T241 [P] [US39] 增加轨迹异常根因模型（异常类型、责任归因、修复动作、证据链）（`backend/internal/entity/models/logistics/tracking_root_cause*.go`）
- [X] T242 [US39] 实现根因分析服务（异常聚类、归因规则、处置建议）（`backend/internal/services/admin/logistics/tracking_root_cause_service.go`）
- [X] T243 [US39] 实现根因分析接口（异常聚合、明细钻取、动作回写）（`backend/internal/transport/http/admin/logistics/{tracking_root_cause_handler.go,routes.go,dto.go}`）
- [X] T244 [US39] 在 SLA 页新增“根因分析”区块（异常分布、责任占比、建议动作）（`web-admin/app/pages/shipping/sla.vue`）
- [X] T245 [US39] 增加 US39 回归测试（归因准确性、动作闭环、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 40 - 跨仓协同与调拨履约（P2）

- [X] T246 [P] [US40] 增加跨仓协同与调拨模型（源仓/目标仓、调拨成本、时效影响）（`backend/internal/entity/models/logistics/interwarehouse_allocation*.go`）
- [X] T247 [US40] 实现跨仓协同服务（缺货检测、调拨候选评分、履约路径重算）（`backend/internal/services/admin/logistics/interwarehouse_allocation_service.go`）
- [X] T248 [US40] 实现跨仓协同接口（候选查询、调拨确认、重算结果）（`backend/internal/transport/http/admin/logistics/{interwarehouse_allocation_handler.go,routes.go,dto.go}`）
- [X] T249 [US40] 在运单页新增“跨仓协同”面板（候选仓对比、成本/时效影响展示）（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T250 [US40] 增加 US40 回归测试（候选排序稳定性、调拨约束、幂等防重）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 41 - 履约质量审计与复盘报告（P3）

- [X] T251 [P] [US41] 增加履约质量审计与复盘模型（报告周期、指标快照、结论与行动项）（`backend/internal/entity/models/logistics/quality_audit_report*.go`）
- [X] T252 [US41] 实现复盘报告服务（指标汇总、异常摘要、行动建议生成）（`backend/internal/services/admin/logistics/quality_audit_report_service.go`）
- [X] T253 [US41] 实现复盘报告接口（报告生成、详情查看、导出）（`backend/internal/transport/http/admin/logistics/{quality_audit_report_handler.go,routes.go,dto.go}`）
- [X] T254 [US41] 新增复盘报告页面（周期筛选、结论摘要、行动项追踪）（`web-admin/app/pages/shipping/quality-reports.vue`）
- [X] T255 [US41] 增加 US41 回归测试（报告口径正确性、导出稳定性、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-12 Polish

- [X] T256 [P] 更新 M12 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [X] T257 执行 M12 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`）
- [X] T258 执行 M12 前端构建与页面回归（`make build-admin`）
- [X] T259 执行 M12 关键链路冒烟（US38/US39/US40 核心服务用例）（`go test ./internal/services/admin/logistics -run 'TestCapacityForecastService_|TestTrackingRootCauseService_|TestInterwarehouseAllocationService_' -count=1`）
- [X] T260 归档 M12 执行记录与验收结论（`specs/011-fulfillment-logistics/quickstart.md`）

---

## Phase 19: Iteration-13 - 决策编排与运营韧性（Backlog）

**Purpose**: 在 M12 预测与协同能力基础上，补齐“仿真 + 编排 + 画像 + 合规 + 资金风控 + 运维自治”，形成履约治理中台。

### User Story 42 - 履约仿真沙盘（What-if）（P2）

- [X] T261 [P] [US42] 增加履约仿真模型（场景参数、策略组合、仿真结果快照）（`backend/internal/entity/models/logistics/fulfillment_sandbox*.go`）
- [X] T262 [US42] 实现履约仿真服务（时效/成本/异常率预测、策略对比）（`backend/internal/services/admin/logistics/fulfillment_sandbox_service.go`）
- [X] T263 [US42] 实现履约仿真接口（场景保存、仿真执行、结果对比）（`backend/internal/transport/http/admin/logistics/{fulfillment_sandbox_handler.go,routes.go,dto.go}`）
- [X] T264 [US42] 新增履约仿真页面（参数面板、对比图、推荐策略）（`web-admin/app/pages/shipping/sandbox.vue`）
- [X] T265 [US42] 增加 US42 回归测试（仿真可重复性、策略差异可解释性、租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 43 - 履约策略编排中心（P2）

- [X] T266 [P] [US43] 增加策略编排模型（规则流、优先级、冲突关系、版本）（`backend/internal/entity/models/logistics/policy_orchestration*.go`）
- [X] T267 [US43] 实现策略编排服务（规则编排、冲突检测、灰度发布）（`backend/internal/services/admin/logistics/policy_orchestration_service.go`）
- [X] T268 [US43] 实现策略编排接口（流程管理、冲突预检、发布回滚）（`backend/internal/transport/http/admin/logistics/{policy_orchestration_handler.go,routes.go,dto.go}`）
- [X] T269 [US43] 新增策略编排页面（流程画布、冲突提示、发布面板）（`web-admin/app/pages/shipping/policy-orchestration.vue`）
- [X] T270 [US43] 增加 US43 回归测试（冲突检测准确性、版本回滚、并发发布保护）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 44 - 承运商服务画像与淘汰机制（P2）

- [X] T271 [P] [US44] 增加承运商画像模型（稳定性得分、成本得分、服务评级、淘汰状态）（`backend/internal/entity/models/logistics/carrier_profile*.go`）
- [X] T272 [US44] 实现承运商画像服务（多维评分、趋势分析、淘汰建议）（`backend/internal/services/admin/logistics/carrier_profile_service.go`）
- [X] T273 [US44] 实现承运商画像接口（画像查询、评级确认、淘汰/恢复操作）（`backend/internal/transport/http/admin/logistics/{carrier_profile_handler.go,routes.go,dto.go}`）
- [X] T274 [US44] 在承运商页新增“服务画像”区块（评分雷达图、趋势、建议动作）（`web-admin/app/pages/shipping/carriers.vue`）
- [X] T275 [US44] 增加 US44 回归测试（评分稳定性、淘汰约束、跨租户隔离）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 45 - 跨境合规知识库联动（P3）

- [X] T276 [P] [US45] 增加跨境合规知识库模型（政策版本、国家规则映射、生效窗口）（`backend/internal/entity/models/logistics/compliance_kb*.go`）
- [X] T277 [US45] 实现合规联动服务（政策差异比对、规则更新建议、灰度生效）（`backend/internal/services/admin/logistics/compliance_kb_service.go`）
- [X] T278 [US45] 实现合规联动接口（政策同步、差异查看、规则发布）（`backend/internal/transport/http/admin/logistics/{compliance_kb_handler.go,routes.go,dto.go}`）
- [X] T279 [US45] 在运单页跨境区块新增“合规版本”视图（版本对比、生效状态、影响提示）（`web-admin/app/pages/shipping/waybills.vue`）
- [X] T280 [US45] 增加 US45 回归测试（版本差异准确性、灰度范围控制、回滚安全）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 46 - 履约资金风险联动（P3）

- [X] T281 [P] [US46] 增加履约资金风险模型（赔付风险、拒付关联、止损动作、阈值）（`backend/internal/entity/models/logistics/fulfillment_finance_risk*.go`）
- [X] T282 [US46] 实现资金风险服务（风险评分、止损建议、处置闭环）（`backend/internal/services/admin/logistics/fulfillment_finance_risk_service.go`）
- [X] T283 [US46] 实现资金风险接口（风险查询、动作执行、处置审计）（`backend/internal/transport/http/admin/logistics/{fulfillment_finance_risk_handler.go,routes.go,dto.go}`）
- [X] T284 [US46] 在对账页新增“资金风险联动”区块（风险榜单、止损动作、处置记录）（`web-admin/app/pages/shipping/billing.vue`）
- [X] T285 [US46] 增加 US46 回归测试（风险评分边界、动作幂等、审计完整性）（`backend/internal/services/admin/logistics/*_test.go`）

### User Story 47 - 履约自动化运维中心（P2）

- [X] T286 [P] [US47] 增加运维自动化模型（重试策略、熔断策略、抑制规则、升级链路）（`backend/internal/entity/models/logistics/ops_automation*.go`）
- [X] T287 [US47] 实现运维自动化服务（告警抑制、自动恢复、值班升级）（`backend/internal/services/admin/logistics/ops_automation_service.go`）
- [X] T288 [US47] 实现运维自动化接口（规则管理、执行记录、手动接管）（`backend/internal/transport/http/admin/logistics/{ops_automation_handler.go,routes.go,dto.go}`）
- [X] T289 [US47] 在 SLA 页新增“运维中心”区块（告警态势、抑制命中、升级轨迹）（`web-admin/app/pages/shipping/sla.vue`）
- [X] T290 [US47] 增加 US47 回归测试（告警抑制准确性、自动恢复稳定性、接管权限控制）（`backend/internal/services/admin/logistics/*_test.go`）

### Iteration-13 Polish

- [ ] T291 [P] 更新 M13 quickstart 与执行记录模板（`specs/011-fulfillment-logistics/quickstart.md`）
- [ ] T292 执行 M13 后端回归（`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`）
- [ ] T293 执行 M13 前端构建与页面回归（`make build-admin`）
- [ ] T294 执行 M13 关键链路冒烟（US42/US43/US44 核心服务用例）（`go test ./internal/services/admin/logistics -run 'TestFulfillmentSandboxService_|TestPolicyOrchestrationService_|TestCarrierProfileService_' -count=1`）
- [ ] T295 归档 M13 执行记录与验收结论（`specs/011-fulfillment-logistics/quickstart.md`）

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
- **US34 (P2)**: 依赖 US30 的控制塔聚合能力，扩展为多维 KPI 分析与导出
- **US35 (P2)**: 依赖 US29 的结算闭环能力，扩展账单/流水/发票三方自动核对
- **US36 (P3)**: 依赖 US33 的跨境履约能力，扩展为清关规则中心与预检风控
- **US37 (P2)**: 依赖 US34 的指标聚合与 US31/US32 的异常处置能力，扩展为稳定性守卫与自动限流
- **US38 (P2)**: 依赖 US34/US37 的指标与稳定性能力，扩展为预测驱动的容量治理
- **US39 (P2)**: 依赖 US31/US37 的异常与守卫能力，扩展为根因分析与处置建议
- **US40 (P2)**: 依赖 US25/US30 的路由与仓配能力，扩展跨仓协同调拨
- **US41 (P3)**: 依赖 US34-US40 的数据沉淀，扩展为质量审计与复盘报告
- **US42 (P2)**: 依赖 US38-US41 的数据基座，扩展履约仿真与策略评估
- **US43 (P2)**: 依赖 US37/US42 的策略与阈值能力，扩展统一编排中心
- **US44 (P2)**: 依赖 US34/US39 的指标与异常数据，扩展承运商画像与淘汰机制
- **US45 (P3)**: 依赖 US36 的清关规则能力，扩展跨境合规知识联动
- **US46 (P3)**: 依赖 US35/US41 的对账与质量数据，扩展资金风险联动
- **US47 (P2)**: 依赖 US37/US39 的稳定性与异常能力，扩展自动化运维中心

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

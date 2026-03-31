# Tasks: 订阅对账与续费治理

**Input**: Design documents from `/specs/013-subscription-reconciliation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/openapi.yaml, quickstart.md

**Tests**: 本特性在 spec.md 中明确包含独立测试与验收场景，需为每个用户故事补充测试任务。

**Organization**: 任务按用户故事分组，保证每个故事可独立实现与独立验证。

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立特性路由、错误码、观测常量与任务执行基线。

- [x] T001 校验任务编号与需求覆盖基线在 `specs/013-subscription-reconciliation/tasks.md`
- [x] T002 在 `backend/internal/transport/http/admin/routes.go` 注册订阅对账路由分组 `/subscription-reconciliation`
- [x] T003 [P] 在 `backend/internal/contracts/errors.go` 增加对账治理域错误码常量
- [x] T004 [P] 在 `backend/internal/observability/subscription_reconciliation/events.go` 定义领域事件名常量

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 完成所有用户故事共享的数据模型、仓储、服务骨架与 DTO。

**⚠️ CRITICAL**: 本阶段完成前不得开始任一用户故事实现。

- [x] T005 新增对账批次与差异表迁移在 `backend/cmd/database/migrations/20260331190000_create_subscription_reconciliation_tables.sql`
- [x] T006 [P] 新增差异任务与治理执行日志表迁移在 `backend/cmd/database/migrations/20260331191000_create_subscription_reconciliation_task_tables.sql`
- [x] T007 [P] 定义 `ReconciliationBatch/ReconciliationDelta` 模型在 `backend/internal/entity/models/subscription_reconciliation/reconciliation_models.go`
- [x] T008 [P] 定义 `DeltaTask/RenewalGovernancePolicy/RenewalExecutionLog` 模型在 `backend/internal/entity/models/subscription_reconciliation/governance_models.go`
- [x] T009 实现基础仓储接口与 tenant 过滤在 `backend/internal/entity/repository/subscription_reconciliation/repository.go`
- [x] T010 [P] 实现 GORM 仓储（批次+差异）在 `backend/internal/entity/repository/subscription_reconciliation/reconciliation_repository_gorm.go`
- [x] T011 [P] 实现 GORM 仓储（任务+治理）在 `backend/internal/entity/repository/subscription_reconciliation/governance_repository_gorm.go`
- [x] T012 定义通用请求/响应 DTO 与分页结构在 `backend/internal/transport/http/admin/subscription_reconciliation/dto.go`
- [x] T013 [P] 创建 service 骨架与依赖装配在 `backend/internal/services/admin/subscription_reconciliation/service.go`
- [x] T014 [P] 创建 handler 骨架与参数校验入口在 `backend/internal/transport/http/admin/subscription_reconciliation/handler.go`

**Checkpoint**: 共享基础设施完成，可进入用户故事开发。

---

## Phase 3: User Story 1 - 每日订阅对账闭环 (Priority: P1) 🎯 MVP

**Goal**: 按净应收口径生成日对账结果、输出差异并可转处置任务，支持手工纠偏并审计。

**Independent Test**: 使用正常/漏扣/重复扣费/金额不一致样本执行日对账，验证汇总、差异明细、任务创建、手工纠偏与幂等均符合预期。

### Tests for User Story 1

- [x] T015 [P] [US1] 新增对账批次创建与幂等服务测试在 `backend/internal/services/admin/subscription_reconciliation/reconciliation_service_test.go`
- [x] T016 [P] [US1] 新增差异类型判定与净应收口径测试在 `backend/internal/services/admin/subscription_reconciliation/reconciliation_delta_classifier_test.go`
- [x] T017 [P] [US1] 新增差异转任务唯一性测试在 `backend/internal/services/admin/subscription_reconciliation/delta_task_service_test.go`
- [x] T018 [P] [US1] 新增手工纠偏与审计日志测试在 `backend/internal/services/admin/subscription_reconciliation/manual_adjustment_service_test.go`
- [x] T019 [P] [US1] 新增 API 合同测试（批次创建/批次查询/差异查询/转任务/纠偏）在 `backend/internal/transport/http/admin/subscription_reconciliation/handler_contract_test.go`

### Implementation for User Story 1

- [x] T020 [US1] 实现日对账编排（应收/实收/差异汇总）在 `backend/internal/services/admin/subscription_reconciliation/reconciliation_service.go`
- [x] T021 [US1] 实现标准五分类差异生成与风险分级在 `backend/internal/services/admin/subscription_reconciliation/delta_classifier.go`
- [x] T022 [US1] 实现差异任务创建与未关闭任务去重在 `backend/internal/services/admin/subscription_reconciliation/delta_task_service.go`
- [x] T023 [US1] 实现批次与差异查询接口在 `backend/internal/transport/http/admin/subscription_reconciliation/query_handler.go`
- [x] T024 [US1] 实现差异转任务与任务关闭接口在 `backend/internal/transport/http/admin/subscription_reconciliation/task_handler.go`
- [x] T025 [US1] 实现手工纠偏接口与字段变更审计在 `backend/internal/transport/http/admin/subscription_reconciliation/manual_adjustment_handler.go`
- [x] T026 [US1] 记录 `reconciliation.generated` 与 `delta.task.created/closed` 与 `delta.adjusted` 事件在 `backend/internal/observability/subscription_reconciliation/publisher.go`

**Checkpoint**: US1 可独立上线验证（MVP）。

---

## Phase 4: User Story 2 - 自动续费失败治理 (Priority: P2)

**Goal**: 对续费失败订阅执行递增重试、通知与升级处置，输出完整执行记录。

**Independent Test**: 构造续费失败样本执行治理任务，验证 1h/24h/72h/7d 窗口调度、结果记录、连续失败升级流程。

### Tests for User Story 2

- [x] T027 [P] [US2] 新增递增窗口调度与重试执行测试在 `backend/internal/services/admin/subscription_reconciliation/renewal_governance_service_test.go`
- [x] T028 [P] [US2] 新增连续失败升级阈值测试在 `backend/internal/services/admin/subscription_reconciliation/renewal_escalation_test.go`
- [x] T029 [P] [US2] 新增治理执行 API 合同测试在 `backend/internal/transport/http/admin/subscription_reconciliation/governance_handler_contract_test.go`

### Implementation for User Story 2

- [x] T030 [US2] 实现治理策略加载与生效区间校验在 `backend/internal/services/admin/subscription_reconciliation/governance_policy_service.go`
- [x] T031 [US2] 实现重试窗口执行器（1h/24h/72h/7d）在 `backend/internal/services/admin/subscription_reconciliation/renewal_retry_executor.go`
- [x] T032 [US2] 实现通知与升级处置编排在 `backend/internal/services/admin/subscription_reconciliation/renewal_orchestrator.go`
- [x] T033 [US2] 实现治理执行接口 `/governance/run` 在 `backend/internal/transport/http/admin/subscription_reconciliation/governance_handler.go`
- [x] T034 [US2] 记录 `renewal.retry.executed` 与 `renewal.escalated` 审计事件在 `backend/internal/observability/subscription_reconciliation/publisher.go`

**Checkpoint**: US2 可独立验证恢复率与升级处置链路。

---

## Phase 5: User Story 3 - 差异责任归因与看板 (Priority: P3)

**Goal**: 提供按渠道/套餐/地区/失败原因聚合的运营看板与导出能力。

**Independent Test**: 跑完一轮对账与治理后，验证看板指标与归因分布可筛选、可追溯、可导出。

### Tests for User Story 3

- [ ] T035 [P] [US3] 新增看板聚合服务测试在 `backend/internal/services/admin/subscription_reconciliation/dashboard_service_test.go`
- [ ] T036 [P] [US3] 新增看板与导出 API 合同测试在 `backend/internal/transport/http/admin/subscription_reconciliation/dashboard_handler_contract_test.go`
- [ ] T037 [P] [US3] 新增运营看板页面单测在 `web-admin/tests/unit/subscription-reconciliation-dashboard.spec.ts`

### Implementation for User Story 3

- [ ] T038 [US3] 实现看板指标聚合（差异率/恢复率/SLA/积压）在 `backend/internal/services/admin/subscription_reconciliation/dashboard_service.go`
- [ ] T039 [US3] 实现多维筛选与归因分布查询在 `backend/internal/services/admin/subscription_reconciliation/analytics_query_service.go`
- [ ] T040 [US3] 实现看板接口 `/dashboard` 与导出接口在 `backend/internal/transport/http/admin/subscription_reconciliation/dashboard_handler.go`
- [ ] T041 [US3] 新增运营看板页在 `web-admin/app/pages/finance/subscription-reconciliation.vue`
- [ ] T042 [US3] 新增看板状态管理与查询组合式在 `web-admin/app/stores/subscription-reconciliation.ts`
- [ ] T043 [US3] 增加渠道/套餐/地区/失败原因维度筛选控件在 `web-admin/app/pages/finance/subscription-reconciliation.vue`

**Checkpoint**: US3 可独立提供经营分析与审计导出。

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 跨故事质量提升、性能验收、业务指标闭环与回归。

- [ ] T044 [P] 补充错误码与用户文案映射在 `backend/internal/contracts/errors.go` 与 `web-admin/app/constants/error-codes.ts`
- [ ] T045 [P] 补充 quickstart 回归说明与示例请求在 `specs/013-subscription-reconciliation/quickstart.md`
- [ ] T046 建立对账性能基准测试（10k 账单 30 分钟内）在 `backend/internal/services/admin/subscription_reconciliation/reconciliation_performance_test.go`
- [ ] T047 [P] 建立看板查询性能测试（p95 < 1s）在 `backend/internal/services/admin/subscription_reconciliation/dashboard_performance_test.go`
- [ ] T048 [P] 增加 SC-002/SC-003 运营指标采集与口径说明在 `backend/internal/observability/subscription_reconciliation/kpi_metrics.go`
- [ ] T049 执行后端与前端回归并记录失败用例清单在 `specs/013-subscription-reconciliation/tasks.md`
- [ ] T050 执行 `make test` 与 `make test-admin-ci` 并记录最终通过结果在 `specs/013-subscription-reconciliation/tasks.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → 可立即开始
- Phase 2 → 依赖 Phase 1，且阻塞全部用户故事
- Phase 3/4/5 → 均依赖 Phase 2 完成
- Phase 6 → 依赖目标用户故事完成

### User Story Dependencies

- US1 (P1): 仅依赖 Phase 2，可先作为 MVP
- US2 (P2): 依赖 Phase 2，可与 US3 并行，但建议在 US1 稳定后推进
- US3 (P3): 依赖 Phase 2，可与 US2 并行

### Within Each User Story

- 先完成测试任务（TDD）
- 再完成模型/服务
- 再完成 handler/API 与前端接入
- 最后做故事级独立验收

## Parallel Opportunities

- Phase 1: T003 与 T004 可并行
- Phase 2: T007/T008、T010/T011、T013/T014 可并行
- US1: T015/T016/T017/T018/T019 可并行；实现阶段可先并行 T021 与 T023
- US2: T027/T028/T029 可并行；T031 与 T032 可并行
- US3: T035/T036/T037 可并行；T041 与 T042 可并行
- Phase 6: T046/T047/T048 可并行

## Parallel Example: User Story 1

```bash
Task: "T015 [US1] 对账批次创建与幂等服务测试"
Task: "T016 [US1] 差异类型判定与净应收口径测试"
Task: "T017 [US1] 差异转任务唯一性测试"
Task: "T018 [US1] 手工纠偏与审计日志测试"
Task: "T019 [US1] API 合同测试（批次/差异/任务/纠偏）"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1 + Phase 2
2. 完成 Phase 3（US1）
3. 依据 US1 Independent Test 做独立验收
4. 通过后即可做首版演示/灰度

### Incremental Delivery

1. 先交付 US1（日对账闭环 + 纠偏审计）
2. 再交付 US2（续费治理提升恢复率）
3. 最后交付 US3（看板与归因分析）

### Parallel Team Strategy

1. 1 人处理 Phase 1/2 主线
2. 基础完成后：A 负责 US1，B 负责 US2，C 负责 US3
3. 在 Phase 6 汇总联调、性能验收与回归

## Notes

- 所有任务均使用 `- [ ] Txxx ...` 格式，且用户故事任务包含 `[USx]` 标签。
- `[P]` 仅用于可并行任务（不同文件/无前置依赖冲突）。
- 每个用户故事都定义了可独立执行的测试与验收标准。

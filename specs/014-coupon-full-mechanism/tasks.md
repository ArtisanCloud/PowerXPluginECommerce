# Tasks: 完整优惠券机制（订单复算与支付核销）

**Input**: Design documents from `/specs/014-coupon-full-mechanism/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 本特性规格明确了独立验收标准与幂等/并发要求，包含测试任务。  
**Organization**: 任务按用户故事分组，保证可独立实现与验证。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件且无前置依赖冲突）
- **[Story]**: 对应用户故事（US1/US2/US3）
- 每条任务都包含明确文件路径

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立优惠券子域基础目录、路由入口与配置骨架

- [X] T001 创建优惠券子域目录骨架 `backend/internal/entity/models/coupon/`、`backend/internal/entity/repository/coupon/`、`backend/internal/services/admin/coupon/`、`backend/internal/transport/http/admin/coupon/`
- [X] T002 新增优惠券表名常量到 `backend/internal/entity/models/model.go`
- [X] T003 [P] 在 `backend/cmd/database/migrate/migrate.go` 注册优惠券模型迁移入口
- [X] T004 [P] 在 `backend/internal/transport/http/admin/routes.go` 挂载优惠券管理路由分组
- [X] T005 [P] 在 `backend/internal/transport/http/routes.go` 挂载交易侧 `/v1/coupons/quote` 路由入口

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 完成所有用户故事共享的核心数据结构、状态机与错误语义  
**⚠️ CRITICAL**: 完成前不得开始任何用户故事实现

- [X] T006 创建模型 `backend/internal/entity/models/coupon/template.go`（CouponTemplate）
- [X] T007 [P] 创建模型 `backend/internal/entity/models/coupon/asset.go`（CouponAsset）
- [X] T008 [P] 创建模型 `backend/internal/entity/models/coupon/usage_log.go`（CouponUsageLog）
- [X] T009 [P] 创建模型 `backend/internal/entity/models/coupon/order_snapshot.go`（OrderCouponSnapshot）
- [X] T010 编写迁移 `backend/cmd/database/migrate/versions/20260403120000_create_coupon_tables.go`（唯一约束与索引：模板编码、券码、幂等键、订单快照）
- [X] T011 创建仓储 `backend/internal/entity/repository/coupon/template_repository.go`
- [X] T012 [P] 创建仓储 `backend/internal/entity/repository/coupon/asset_repository.go`
- [X] T013 [P] 创建仓储 `backend/internal/entity/repository/coupon/usage_log_repository.go`
- [X] T014 [P] 创建仓储 `backend/internal/entity/repository/coupon/order_snapshot_repository.go`
- [X] T015 创建领域错误码与原因枚举 `backend/internal/services/admin/coupon/errors.go`（门槛不满足、范围不匹配、已过期、已占用、不可叠加）
- [X] T016 创建状态机与幂等守卫 `backend/internal/services/admin/coupon/state_machine.go`
- [X] T017 创建金额分摊器 `backend/internal/services/admin/coupon/allocation.go`（订单行占比分摊）
- [X] T018 创建叠加排序器 `backend/internal/services/admin/coupon/stacking.go`（先商品级后订单级、同层优先级）

**Checkpoint**: 基础设施完成，可进入用户故事实现

---

## Phase 3: User Story 1 - 结算复算与最优用券 (Priority: P1) 🎯 MVP

**Goal**: 下单提交时返回服务端权威券后金额与可解释明细  
**Independent Test**: 携带可用/不可用券提交订单，确认复算金额、应用券明细、拒绝原因及订单快照一致

### Tests for User Story 1

- [X] T019 [P] [US1] 新增试算契约测试 `backend/internal/transport/http/miniapp/coupon/quote_handler_contract_test.go`
- [X] T020 [P] [US1] 新增规则计算单测 `backend/internal/services/admin/coupon/quote_service_test.go`（门槛、范围、叠加顺序、过期边界）
- [X] T021 [P] [US1] 新增金额分摊单测 `backend/internal/services/admin/coupon/allocation_test.go`

### Implementation for User Story 1

- [X] T022 [US1] 实现试算服务 `backend/internal/services/admin/coupon/quote_service.go`
- [X] T023 [US1] 实现订单提交预占服务 `backend/internal/services/admin/coupon/reservation_service.go`
- [X] T024 [US1] 实现交易接口处理器 `backend/internal/transport/http/miniapp/coupon/quote_handler.go`
- [X] T025 [US1] 在下单链路接入优惠复算与预占 `backend/internal/services/admin/order/service.go`
- [X] T026 [US1] 在订单落库时写入优惠快照 `backend/internal/services/admin/order/service.go`
- [X] T027 [US1] 扩展订单 DTO 输出优惠信息 `backend/internal/services/admin/order/types.go`
- [X] T028 [US1] 更新订单查询聚合返回优惠快照 `backend/internal/services/admin/order/query.go`
- [X] T029 [US1] 增加下单链路审计事件 `backend/internal/services/admin/order/service.go`

**Checkpoint**: US1 可独立上线（MVP）

---

## Phase 4: User Story 2 - 支付核销与失败释放 (Priority: P2)

**Goal**: 支付成功核销、失败/取消释放，保障券资产与资金状态一致  
**Independent Test**: 覆盖支付成功、失败关闭、重复回调三路径，验证状态机正确且幂等

### Tests for User Story 2

- [X] T030 [P] [US2] 新增核销服务单测 `backend/internal/services/admin/coupon/redeem_service_test.go`
- [X] T031 [P] [US2] 新增释放服务单测 `backend/internal/services/admin/coupon/release_service_test.go`
- [X] T032 [P] [US2] 新增回调幂等集成测试 `backend/internal/services/admin/payment/coupon_idempotency_test.go`

### Implementation for User Story 2

- [X] T033 [US2] 实现核销服务 `backend/internal/services/admin/coupon/redeem_service.go`
- [X] T034 [US2] 实现释放服务 `backend/internal/services/admin/coupon/release_service.go`
- [X] T035 [US2] 在支付成功回调接入核销 `backend/internal/services/admin/payment/service.go`
- [X] T036 [US2] 在订单取消/关闭链路接入释放 `backend/internal/services/admin/order/cancel.go`
- [X] T037 [US2] 在超时关闭任务中接入释放 `backend/internal/jobs/order_timeout_release_job.go`
- [X] T038 [US2] 新增券动作流水写入与幂等键校验 `backend/internal/services/admin/coupon/usage_log_service.go`

**Checkpoint**: US2 可独立验证，不依赖 US3

---

## Phase 5: User Story 3 - 运营配置与审计追踪 (Priority: P3)

**Goal**: 提供模板管理、发券、资产与流水查询，支持运营和客服排障  
**Independent Test**: 创建模板并发券后完成一次订单支付，管理端可查到完整生命周期

### Tests for User Story 3

- [X] T039 [P] [US3] 新增模板管理处理器测试 `backend/internal/transport/http/admin/coupon/template_handler_test.go`
- [X] T040 [P] [US3] 新增发券服务测试 `backend/internal/services/admin/coupon/issue_service_test.go`
- [X] T041 [P] [US3] 新增资产/流水查询测试 `backend/internal/services/admin/coupon/query_service_test.go`

### Implementation for User Story 3

- [X] T042 [US3] 实现模板管理服务 `backend/internal/services/admin/coupon/template_service.go`
- [X] T043 [US3] 实现发券服务 `backend/internal/services/admin/coupon/issue_service.go`
- [X] T044 [US3] 实现资产与流水查询服务 `backend/internal/services/admin/coupon/query_service.go`
- [X] T045 [US3] 实现管理端处理器 `backend/internal/transport/http/admin/coupon/template_handler.go`
- [X] T046 [P] [US3] 实现管理端处理器 `backend/internal/transport/http/admin/coupon/issue_handler.go`
- [X] T047 [P] [US3] 实现管理端处理器 `backend/internal/transport/http/admin/coupon/query_handler.go`
- [X] T048 [US3] 注册 RBAC 资源与动作 `backend/internal/transport/http/admin/coupon/rbac.go`
- [X] T049 [US3] 管理端 API 封装 `web-admin/app/composables/api/useCoupons.ts`
- [X] T050 [US3] 管理端页面：模板列表与编辑 `web-admin/app/pages/pricing/coupons.vue`
- [X] T051 [US3] 管理端页面：资产与流水查询 `web-admin/app/pages/pricing/coupon-usages.vue`

**Checkpoint**: 全量用户故事独立可测

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 横切质量、文档与验收收口

- [X] T052 [P] 补充 OpenAPI 与实现一致性校验 `specs/014-coupon-full-mechanism/contracts/coupons.openapi.yaml`
- [X] T053 [P] 补充路由矩阵与中间件链校验 `specs/014-coupon-full-mechanism/contracts/routing-matrix.md`
- [X] T054 增加可观测指标埋点（发放/预占/核销/释放/返券）`backend/internal/observability/coupon/metrics.go`
- [X] T055 增加告警规则（核销失败堆积、超时未释放）`backend/internal/observability/coupon/alerts.go`
- [X] T056 [P] 更新功能文档与排障手册 `docs/guides/features/pricing/coupons.md`
- [X] T057 执行 quickstart 场景回归并记录结果 `specs/014-coupon-full-mechanism/quickstart.md`
- [X] T058 性能压测脚本：结算试算接口基准 `backend/tests/performance/coupon_quote_benchmark_test.go`
- [X] T059 结算性能门禁校验（P95 <= 500ms）`backend/tests/performance/coupon_quote_sla_test.go`
- [X] T060 [P] 客服链路定位查询耗时测试（订单号维度）`backend/internal/services/admin/coupon/query_latency_test.go`
- [X] T061 [P] 客服链路定位查询耗时测试（券码维度）`backend/internal/services/admin/coupon/query_by_code_latency_test.go`
- [X] T062 索引命中与查询计划校验（usage/snapshot）`backend/internal/entity/repository/coupon/query_index_validation_test.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2 → Phase 3/4/5 → Phase 6
- Phase 2 完成前，所有用户故事任务不得启动
- 用户故事建议顺序：US1（MVP）→ US2 → US3

### User Story Dependencies

- **US1 (P1)**: 仅依赖 Foundational，可单独交付
- **US2 (P2)**: 依赖 US1 的预占与快照基础
- **US3 (P3)**: 可与 US2 并行后段推进，但最终依赖 US1/US2 的真实流水数据才能完整验收

### Within Each User Story

- 测试任务先写并执行失败验证（红灯）
- 模型/服务 → 处理器/路由 → 集成到订单/支付链路
- 每个故事完成后按独立验收标准验证

---

## Parallel Opportunities

- Phase 1: T003/T004/T005 可并行
- Phase 2: T007/T008/T009、T012/T013/T014 可并行
- US1: T019/T020/T021 可并行
- US2: T030/T031/T032 可并行
- US3: T039/T040/T041 与 T046/T047 可并行
- Polish: T052/T053/T056 可并行
- 质量门禁: T060/T061 可并行

---

## Parallel Example: User Story 1

```bash
Task: "T019 [US1] backend/internal/transport/http/miniapp/coupon/quote_handler_contract_test.go"
Task: "T020 [US1] backend/internal/services/admin/coupon/quote_service_test.go"
Task: "T021 [US1] backend/internal/services/admin/coupon/allocation_test.go"
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. 完成 Phase 1 + Phase 2
2. 完成 US1（Phase 3）
3. 执行独立验收：券后金额、拒绝原因、订单快照一致
4. 可先行灰度发布 MVP

### Incremental Delivery

1. US1 上线：完成下单复算与预占
2. US2 上线：完成支付核销与失败释放
3. US3 上线：补齐运营配置与审计查询
4. Phase 6 收口：可观测、文档、回归

# Tasks: 促销规则引擎（订单自动促销）

**Input**: Design documents from `/specs/015-pricing-promotions-engine/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 本特性规格明确独立验收标准、金额准确率、租户隔离、结算性能与回归要求，包含测试任务。  
**Organization**: 任务按用户故事分组，保证可独立实现与验证。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件且无前置依赖冲突）
- **[Story]**: 对应用户故事（US1/US2/US3）
- 每条任务都包含明确文件路径

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立促销子域目录、路由入口、模型注册与前端入口骨架

- [X] T001 创建促销子域目录骨架 `backend/internal/entity/models/promotion/`、`backend/internal/entity/repository/promotion/`、`backend/internal/services/admin/promotion/`、`backend/internal/transport/http/admin/promotion/`、`backend/internal/transport/http/miniapp/promotion/`
- [X] T002 新增促销表名常量到 `backend/internal/entity/models/model.go`
- [X] T003 [P] 在 `backend/cmd/database/migrate/migrate.go` 注册促销模型迁移入口
- [X] T004 [P] 在 `backend/internal/transport/http/admin/routes.go` 挂载促销管理路由分组
- [X] T005 [P] 在 `backend/internal/transport/http/routes.go` 挂载交易侧 `/v1/promotions/quote` 调试路由入口
- [X] T006 [P] 创建管理端 API composable 骨架 `web-admin/app/composables/api/usePromotions.ts`
- [X] T007 [P] 创建管理端页面入口骨架 `web-admin/app/pages/pricing/promotions.vue`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 完成所有用户故事共享的数据结构、规则语义、错误码、金额计算基础  
**CRITICAL**: 完成前不得开始任何用户故事实现

- [X] T008 创建模型 `backend/internal/entity/models/promotion/campaign.go`（PromotionCampaign）
- [X] T009 [P] 创建模型 `backend/internal/entity/models/promotion/order_snapshot.go`（OrderPromotionSnapshot）
- [X] T010 [P] 创建模型 `backend/internal/entity/models/promotion/audit_log.go`（PromotionAuditLog）
- [X] T011 编写迁移 `backend/cmd/database/migrate/versions/20260610150000_create_promotion_tables.go`（唯一约束、索引、RLS 相关表结构）
- [X] T012 创建仓储 `backend/internal/entity/repository/promotion/campaign_repository.go`
- [X] T013 [P] 创建仓储 `backend/internal/entity/repository/promotion/order_snapshot_repository.go`
- [X] T014 [P] 创建仓储 `backend/internal/entity/repository/promotion/audit_log_repository.go`
- [X] T015 创建领域错误码与拒绝原因枚举 `backend/internal/services/admin/promotion/errors.go`（not_started、expired、paused、threshold_not_met、scope_mismatch、exclusive_conflict、not_stackable、invalid_rule、promotion_excludes_coupon）
- [X] T016 创建规则 DTO 与服务输入输出类型 `backend/internal/services/admin/promotion/types.go`
- [X] T017 创建规则校验器 `backend/internal/services/admin/promotion/validation.go`（有效期、类型、范围、动作、叠加规则）
- [X] T018 创建金额分摊器 `backend/internal/services/admin/promotion/allocation.go`（订单行促销前金额占比分摊）
- [X] T019 创建互斥与叠加排序器 `backend/internal/services/admin/promotion/stacking.go`（互斥组最大优惠、priority ASC、stackable）
- [X] T020 [P] 创建促销审计服务骨架 `backend/internal/services/admin/promotion/audit_log_service.go`

**Checkpoint**: 基础设施完成，可进入用户故事实现

---

## Phase 3: User Story 1 - 配置并启用基础促销规则 (Priority: P1) MVP

**Goal**: 定价经理能创建、启用、停用、复制并查询订单级满减/折扣促销规则。  
**Independent Test**: 创建一条订单满减促销，启用后在促销列表中按状态、类型和关键词能查询到，并能查看范围、门槛、优惠动作和叠加配置。

### Tests for User Story 1

- [X] T021 [P] [US1] 新增促销规则校验单测 `backend/internal/services/admin/promotion/validation_test.go`
- [X] T022 [P] [US1] 新增管理服务单测 `backend/internal/services/admin/promotion/campaign_service_test.go`（创建、编码唯一、启用、停用、复制）
- [X] T023 [P] [US1] 新增管理端处理器契约测试 `backend/internal/transport/http/admin/promotion/campaign_handler_test.go`
- [X] T024 [P] [US1] 新增前端 API composable 测试 `web-admin/tests/unit/promotions-api.spec.ts`

### Implementation for User Story 1

- [X] T025 [US1] 实现促销规则管理服务 `backend/internal/services/admin/promotion/campaign_service.go`
- [X] T026 [US1] 实现促销管理 DTO `backend/internal/transport/http/admin/promotion/dto.go`
- [X] T027 [US1] 实现促销管理处理器 `backend/internal/transport/http/admin/promotion/campaign_handler.go`
- [X] T028 [US1] 实现促销审计查询处理器 `backend/internal/transport/http/admin/promotion/audit_handler.go`
- [X] T029 [US1] 注册管理端促销路由 `backend/internal/transport/http/admin/promotion/routes.go`
- [X] T030 [US1] 注册促销 RBAC 资源与动作 `backend/internal/transport/http/admin/promotion/rbac.go`
- [X] T031 [US1] 合并促销 RBAC 到注册表 `backend/internal/transport/http/registry.go`
- [X] T032 [US1] 完成管理端 API 封装 `web-admin/app/composables/api/usePromotions.ts`
- [X] T033 [US1] 实现促销列表、筛选、分页与状态展示 `web-admin/app/pages/pricing/promotions.vue`
- [X] T034 [US1] 实现新建/编辑促销表单与前端校验 `web-admin/app/pages/pricing/promotions.vue`
- [X] T035 [US1] 实现启用、停用、复制与审计日志入口 `web-admin/app/pages/pricing/promotions.vue`
- [X] T036 [US1] 补充促销菜单与权限文案 `plugin.yaml`
- [X] T037 [P] [US1] 补充促销页面 i18n 文案 `web-admin/i18n/zh-CN/promotions.json`
- [X] T038 [P] [US1] 补充促销页面 i18n 文案 `web-admin/i18n/en/promotions.json`

**Checkpoint**: US1 可独立上线，完成规则配置 MVP

---

## Phase 4: User Story 2 - 订单结算自动命中促销 (Priority: P2)

**Goal**: 订单结算自动匹配 active 促销并返回促销前金额、促销优惠、促销后金额、命中促销与未命中原因。  
**Independent Test**: 准备满足门槛的订单草稿和一条 active 满减促销，执行结算试算后返回正确促销金额与命中明细。

### Tests for User Story 2

- [X] T039 [P] [US2] 新增促销 quote 计算单测 `backend/internal/services/admin/promotion/quote_service_test.go`（门槛、范围、渠道、时间边界）
- [X] T040 [P] [US2] 新增互斥和叠加单测 `backend/internal/services/admin/promotion/stacking_test.go`（同组最大优惠、不可叠加）
- [X] T041 [P] [US2] 新增金额分摊单测 `backend/internal/services/admin/promotion/allocation_test.go`
- [X] T042 [P] [US2] 新增交易端试算契约测试 `backend/internal/transport/http/miniapp/promotion/quote_handler_contract_test.go`
- [X] T043 [P] [US2] 新增订单促销集成测试 `backend/internal/services/admin/order/promotion_quote_test.go`
- [X] T044 [P] [US2] 新增促销与优惠券互斥测试 `backend/internal/services/admin/order/promotion_coupon_stacking_test.go`

### Implementation for User Story 2

- [X] T045 [US2] 实现促销 quote 服务 `backend/internal/services/admin/promotion/quote_service.go`
- [X] T046 [US2] 实现促销候选查询方法 `backend/internal/entity/repository/promotion/campaign_repository.go`
- [X] T047 [US2] 实现交易端试算 DTO `backend/internal/transport/http/miniapp/promotion/dto.go`
- [X] T048 [US2] 实现交易端试算处理器 `backend/internal/transport/http/miniapp/promotion/quote_handler.go`
- [X] T049 [US2] 注册交易端促销试算路由 `backend/internal/transport/http/miniapp/promotion/routes.go`
- [X] T050 [US2] 在订单 quote/创建链路接入促销计算 `backend/internal/services/admin/order/service.go`
- [X] T051 [US2] 扩展订单服务类型输出促销金额与明细 `backend/internal/services/admin/order/types.go`
- [X] T052 [US2] 将促销后金额传递给优惠券计算 `backend/internal/services/admin/order/coupon_mapper.go`
- [X] T053 [US2] 在优惠券互斥路径返回 `promotion_excludes_coupon` `backend/internal/services/admin/coupon/quote_service.go`
- [X] T054 [US2] 保证未命中促销订单金额不变并补回归路径 `backend/internal/services/admin/order/service.go`

**Checkpoint**: US2 可独立验证，促销进入交易金额链路

---

## Phase 5: User Story 3 - 订单促销快照与客服追踪 (Priority: P3)

**Goal**: 订单创建时保存促销快照，客服/财务可按订单查看命中促销、未命中原因和订单行分摊。  
**Independent Test**: 创建并支付一笔命中促销的订单后，按订单号查询订单详情，应看到促销优惠汇总、命中促销、未命中原因和分摊信息；促销后续变更不影响历史展示。

### Tests for User Story 3

- [X] T055 [P] [US3] 新增促销快照服务测试 `backend/internal/services/admin/promotion/snapshot_service_test.go`
- [X] T056 [P] [US3] 新增订单详情促销展示测试 `backend/internal/services/admin/order/promotion_snapshot_query_test.go`
- [X] T057 [P] [US3] 新增历史一致性测试 `backend/internal/services/admin/order/promotion_snapshot_immutability_test.go`
- [X] T058 [P] [US3] 新增审计日志查询测试 `backend/internal/services/admin/promotion/audit_log_service_test.go`

### Implementation for User Story 3

- [X] T059 [US3] 实现促销快照服务 `backend/internal/services/admin/promotion/snapshot_service.go`
- [X] T060 [US3] 在订单创建成功时写入促销快照 `backend/internal/services/admin/order/service.go`
- [X] T061 [US3] 在订单查询聚合中加载促销快照 `backend/internal/services/admin/order/query.go`
- [X] T062 [US3] 扩展订单详情 DTO 输出促销快照 `backend/internal/services/admin/order/types.go`
- [X] T063 [US3] 实现促销审计日志查询服务 `backend/internal/services/admin/promotion/audit_log_service.go`
- [X] T064 [US3] 为 quote 命中和拒绝写入聚合审计 `backend/internal/services/admin/promotion/quote_service.go`
- [X] T065 [US3] 在管理端订单详情展示促销优惠与分摊信息 `web-admin/app/pages/market/orders/[id].vue`
- [X] T066 [US3] 在促销页面提供按规则查看审计日志的抽屉 `web-admin/app/pages/pricing/promotions.vue`

**Checkpoint**: 全量用户故事独立可测，历史订单促销解释闭环完成

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 横切质量、文档、可观测、性能与最终验收收口

- [X] T067 [P] 补充促销 OpenAPI 与实现一致性校验 `specs/015-pricing-promotions-engine/contracts/promotions.openapi.yaml`
- [X] T068 [P] 补充促销路由矩阵与中间件链校验 `specs/015-pricing-promotions-engine/contracts/routing-matrix.md`
- [X] T069 增加促销可观测指标（命中次数、优惠金额、拒绝原因、quote 延迟）`backend/internal/observability/promotion/metrics.go`
- [X] T070 [P] 增加促销告警规则（quote 错误率、异常拒绝原因突增）`backend/internal/observability/promotion/alerts.go`
- [X] T071 [P] 更新促销功能使用文档 `docs/guides/features/pricing/promotions.md`
- [X] T072 执行 quickstart 场景回归并记录结果 `specs/015-pricing-promotions-engine/quickstart.md`
- [X] T073 增加促销 quote 性能基准 `backend/tests/performance/promotion_quote_benchmark_test.go`
- [X] T074 增加促销 quote SLA 测试（P95 <= 500ms）`backend/tests/performance/promotion_quote_sla_test.go`
- [X] T075 [P] 增加客服按订单号定位促销来源耗时测试 `backend/internal/services/admin/promotion/query_latency_test.go`
- [X] T076 [P] 增加促销索引命中与查询计划校验 `backend/internal/entity/repository/promotion/query_index_validation_test.go`
- [X] T077 执行后端目标测试并修复失败 `backend/`
- [X] T078 执行前端 lint/build 并修复失败 `web-admin/`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2 → Phase 3/4/5 → Phase 6
- Phase 2 完成前，所有用户故事任务不得启动
- 用户故事建议顺序：US1（规则配置 MVP）→ US2（交易自动促销）→ US3（快照追踪）

### User Story Dependencies

- **US1 (P1)**: 仅依赖 Foundational，可单独交付管理端规则配置 MVP
- **US2 (P2)**: 依赖 US1 的 active 促销规则数据；可在服务层使用测试数据独立验证 quote 计算
- **US3 (P3)**: 依赖 US2 的 quote 结果写入订单；可独立验证快照不可变与客服查询

### Within Each User Story

- 测试任务先写并执行失败验证（红灯）
- 模型/仓储 → 服务 → 处理器/路由 → 前端页面或订单集成
- 每个故事完成后按独立验收标准验证

---

## Parallel Opportunities

- Phase 1: T003/T004/T005/T006/T007 可并行
- Phase 2: T009/T010、T013/T014、T020 可并行
- US1: T021/T022/T023/T024 可并行；T037/T038 可并行
- US2: T039/T040/T041/T042/T043/T044 可并行
- US3: T055/T056/T057/T058 可并行
- Polish: T067/T068/T070/T071/T075/T076 可并行

---

## Parallel Example: User Story 1

```bash
Task: "T021 [US1] backend/internal/services/admin/promotion/validation_test.go"
Task: "T022 [US1] backend/internal/services/admin/promotion/campaign_service_test.go"
Task: "T023 [US1] backend/internal/transport/http/admin/promotion/campaign_handler_test.go"
Task: "T024 [US1] web-admin/tests/promotions-api.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "T039 [US2] backend/internal/services/admin/promotion/quote_service_test.go"
Task: "T040 [US2] backend/internal/services/admin/promotion/stacking_test.go"
Task: "T041 [US2] backend/internal/services/admin/promotion/allocation_test.go"
Task: "T042 [US2] backend/internal/transport/http/miniapp/promotion/quote_handler_contract_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T055 [US3] backend/internal/services/admin/promotion/snapshot_service_test.go"
Task: "T056 [US3] backend/internal/services/admin/order/promotion_snapshot_query_test.go"
Task: "T057 [US3] backend/internal/services/admin/order/promotion_snapshot_immutability_test.go"
Task: "T058 [US3] backend/internal/services/admin/promotion/audit_log_service_test.go"
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. 完成 Phase 1 + Phase 2
2. 完成 US1（Phase 3）
3. 执行独立验收：创建满减促销、启用、列表筛选、复制、停用、审计可查
4. 停止并确认规则配置 MVP 可演示

### Incremental Delivery

1. US1 上线：促销规则可配置、可启停
2. US2 上线：订单 quote/下单自动应用促销，与优惠券互斥对齐
3. US3 上线：订单促销快照、审计和客服解释闭环
4. Phase 6 收口：可观测、文档、性能与 quickstart 回归

### Suggested Validation Commands

```bash
cd backend
go test ./internal/services/admin/promotion ./internal/transport/http/admin/promotion ./internal/transport/http/miniapp/promotion ./internal/services/admin/order

cd ../web-admin
npm run -s lint
npm run -s build
```

## Notes

- [P] tasks = 不同文件且无未完成前置依赖，可并行处理
- [US1]/[US2]/[US3] 与 `spec.md` 的用户故事一一对应
- 交易侧独立 `/v1/promotions/quote` 用于调试和契约验证，最终金额仍必须在订单 quote/下单服务端复算
- 任何订单金额写入都不能信任前端试算结果

# Tasks: 商品（SPU）管理

**Input**: Design documents from `/specs/002-product-spu-management/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立 backend/product 与 web-admin/product 的基础目录与管线，以承载后续实现。

- [x] T001 初始化 product 模块目录与 Go 包（backend/internal/domain/{models,repository}/product, backend/internal/services/admin/product/spu, backend/internal/transport/http/admin/product/spu）。
- [x] T002 建立渠道任务与观测目录骨架（backend/internal/jobs/channels/、backend/internal/observability/product/）并写入 TODO 注释说明目标。
- [x] T003 创建 web-admin/product 目录结构（web-admin/app/pages/product/spus/**、app/components/product、app/composables、app/stores/product）并导出空组件骨架。

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 搭建数据库、模型、仓储、接口与前端 API 客户端，作为所有用户故事的统一底座。

- [x] T004 增加 product_spus、product_spu_versions、product_spu_locales 的迁移脚本（backend/cmd/database/migrate/migrations/*_create_spu_tables.go）。
- [x] T005 增加 product_spu_channels、product_spu_subscription_plans、product_spu_approvals、product_spu_audit_logs、product_spu_import_tasks、product_spu_export_tasks 的迁移脚本（同目录）。
- [x] T006 实现数据模型结构体及 TableName 常量（backend/internal/domain/models/product/{spu.go,version.go,locale.go,channel.go,subscription_plan.go,approval.go,audit_log.go}）。
- [x] T007 实现仓储层并内嵌 BaseRepository，统一 Tenant 过滤（backend/internal/domain/repository/product/{spu_repository.go,version_repository.go,channel_repository.go,subscription_plan_repository.go}）。
- [x] T008 定义 SPU 服务接口、DTO 与通用校验逻辑（backend/internal/services/admin/product/spu/service.go）。
- [x] T009 注册 HTTP 路由与基础 Handler 入口（backend/internal/transport/http/admin/product/spu/router.go）。
- [x] T010 实现多语言字段验证与默认语言复制工具（backend/internal/services/admin/product/spu/locale_service.go），确保 FR-002 后端校验。
- [x] T011 [P] 创建 SPU API client 及 Pinia store skeleton（web-admin/app/composables/useSpuApi.ts、app/stores/product/spu.ts）。

**Checkpoint**: 完成后可独立实现各用户故事。

---

## Phase 3: User Story 1 - 全流程创建并发布 SPU (Priority: P1) 🎯 MVP

**Goal**: 让商品运营可从列表创建/编辑 SPU，提交审核并发布到渠道，形成首个可上线的流程。

**Independent Test**: 通过 web-admin 完成一次「创建→保存草稿→提交→审批通过→发布」流程，验证渠道任务与状态更新。

### Tests for User Story 1

- [x] T012 [P] [US1] 编写创建→发布生命周期的 service 单测（backend/internal/services/admin/product/spu/service_create_test.go）。
- [x] T013 [P] [US1] 编写端到端测试脚本覆盖创建与发布（web-admin/tests/e2e/product-spu-create.spec.ts）。

### Implementation for User Story 1

- [x] T014 [US1] 实现创建/草稿/发布核心服务方法并写入审计与渠道任务钩子（backend/internal/services/admin/product/spu/service.go）。
- [x] T015 [US1] 实现 SPU 列表、创建、更新、提交、发布的 HTTP Handler（backend/internal/transport/http/admin/product/spu/handler.go）。
- [x] T016 [US1] 实现渠道发布任务触发器与回执记录（backend/internal/jobs/channels/publisher.go）。
- [x] T017 [P] [US1] 实现 Pinia actions（列表、创建、提交、发布）和 API 绑定（web-admin/app/stores/product/spu.ts）。
- [x] T018 [P] [US1] 实现 SPU 列表页（过滤器、概览卡、批量入口）（web-admin/app/pages/product/spus/index.vue）。
- [x] T019 [P] [US1] 实现多步骤创建向导与订阅计划配置（web-admin/app/pages/product/spus/create.vue 及 components/product/SpuWizardStep*.vue）。
- [x] T020 [US1] 实现 SPU↔SKU 关联服务/Handler（backend/internal/services/admin/product/spu/sku_link_service.go 与 transport/http/admin/product/spu/handler_skus.go），支持批量创建/复制并同步库存与定价模块。
- [x] T021 [P] [US1] 实现 SKU 关联管理 UI（web-admin/app/pages/product/spus/components/SpuSkuLinker.vue）以批量创建、复制与可视化继承关系。
- [x] T022 [P] [US1] 实现多语言表单组件与必填校验（web-admin/app/components/product/SpuLocaleForm.vue），支持默认语言一键复制与缺失字段阻断提交。

**Checkpoint**: US1 可独立测试并演示，覆盖多语言、SKU 关联与渠道发布。

---

## Phase 4: User Story 2 - 审��与版本控制 (Priority: P2)

**Goal**: 让审批角色可查看差异、审批/驳回，并允许运营回滚历史版本。

**Independent Test**: 对已发布 SPU 进行编辑→提交→审批→回滚，验证版本差异与审计记录完整。

### Tests for User Story 2

- [x] T023 [P] [US2] 编写版本 diff/回滚服务单测（backend/internal/services/admin/product/spu/version_service_test.go）。
- [x] T024 [P] [US2] 编写审批 SLA 提醒逻辑测试（backend/internal/services/admin/product/spu/approval_notifier_test.go）。

### Implementation for User Story 2

- [x] T025 [US2] 实现版本快照、差异与回滚服务（backend/internal/services/admin/product/spu/version_service.go）。
- [x] T026 [US2] 实现版本/审批相关 Handler（列表、差异、审批、驳回、回滚）（backend/internal/transport/http/admin/product/spu/handler_versions.go）。
- [x] T027 [US2] 实现审批 SLA 提醒/升级任务（backend/internal/jobs/channels/approval_sla_worker.go）。
- [x] T028 [P] [US2] 实现版本时间线与 diff 组件（web-admin/app/pages/product/spus/components/VersionDiff.vue）。
- [x] T029 [US2] 在详情页添加审批面板与意见输入（web-admin/app/pages/product/spus/edit/[id].vue）。
- [x] T030 [US2] 扩展审计时间轴组件以展示版本与审批事件（web-admin/app/components/product/SpuAuditTimeline.vue）。

**Checkpoint**: US1 + US2 可并行验证且互不阻塞。

---

## Phase 5: User Story 3 - 渠道发布与批量运营 (Priority: P3)

**Goal**: 提供渠道可见性配置、定时上下架、批量导入导出及失败报告，支撑大规模运营。

**Independent Test**: 完成一次渠道配置+定时上架，并执行一次导入/导出任务，验证任务中心与失败报告。

### Tests for User Story 3

- [x] T031 [P] [US3] 编写导入部分成功/失败报告的 service 测试（backend/internal/services/admin/product/spu/import_service_test.go）。
- [ ] T032 [P] [US3] 编写渠道配置+批量导入端到端测试（web-admin/tests/e2e/product-spu-bulk.spec.ts）。

### Implementation for User Story 3

- [x] T033 [US3] 实现渠道可见性服务与 Handler（backend/internal/services/admin/product/spu/channel_service.go、transport/http/admin/product/spu/channel_handler.go）。
- [x] T034 [US3] 实现订阅计划 effectScope 选项及更新接口（backend/internal/transport/http/admin/product/spu/handler_plans.go）。
- [x] T035 [US3] 实现导入/导出 orchestrator，写入任务中心与审计（backend/internal/services/admin/product/spu/import_service.go）。
- [x] T036 [US3] 实现异步 worker 处理导入导出与渠道同步回执（backend/internal/jobs/channels/spu_sync_worker.go）。
- [x] T037 [P] [US3] 实现前端渠道配置 Tab（web-admin/app/pages/product/spus/components/ChannelVisibilityForm.vue）。
- [x] T038 [US3] 实现批量导入/导出对话框与失败报告查看（web-admin/app/components/product/SpuBulkDialog.vue）。
- [x] T039 [US3] 在导入流水中实现 SPU 编码/SKU 关联重复校验与错误提示（backend/internal/services/admin/product/spu/import_validator.go），确保 Edge Case 处理。
- [x] T043 [US3] 实现 SPU 下架服务与 Handler（backend/internal/services/admin/product/spu/service.go + transport/http/.../handler_withdraw.go），支持多渠道/定时下架并写入审计与渠道任务。
- [x] T044 [P] [US3] 在 Pinia store 与详情页添加“下架”动作（web-admin/app/stores/product/spu.ts、app/pages/product/spus/edit/[id].vue），包含渠道选择、定时与理由输入。
- [x] T045 [US3] 补充服务单测与前端 E2E（backend/internal/services/admin/product/spu/service_withdraw_test.go、web-admin/tests/e2e/product-spu-withdraw.spec.ts），覆盖“发布→下架→重新发布”路径。
- [x] T046 [US3] 实现软删除 API（backend/internal/services/admin/product/spu/service.go、transport/http/.../handler_delete.go）与审计、权限校验，限定仅草稿/下架状态可删除。
- [x] T047 [P] [US3] 在详情页增加“删除”弹窗及 Pinia action（web-admin/app/pages/product/spus/edit/[id].vue、app/stores/product/spu.ts），支持删除理由并在成功后跳转列表。
- [x] T048 [US3] 设计硬删除清理脚本与运维指南（backend/cmd/tools/spu_hard_delete.go 或 docs/plan/addenda.md），确保级联清理与审计留痕。
- [x] T049 [US3] 编写软删除单测与前端 E2E（backend/internal/services/admin/product/spu/service_delete_test.go、web-admin/tests/e2e/product-spu-delete.spec.ts），验证删除条件及提示信息。

**Checkpoint**: 三个用户故事均可独立演示。

---

## Phase 6: Polish & Cross-Cutting

**Purpose**: 收尾文档、可观测性、验证 quickstart，并量化成功指标。

- [x] T040 [P] 更新 quickstart 与 API 合同以记录最新流程（specs/002-product-spu-management/{quickstart.md,contracts/spu-api.md}）。
- [x] T041 增加 KPI 指标与日志（backend/internal/observability/product/spu_metrics.go），覆盖 SC-001~SC-004（上线 Lead Time、导入成功率、渠道同步率、审批 SLA）。
- [x] T042 执行 quickstart 自检并将记录写入 reports/spu-management-validation.md，确认成功标准与边界场景。

---

## Dependencies & Execution Order

1. **Phase 1 → Phase 2**：完成目录与骨架后，方可编写迁移/模型/仓储。
2. **Phase 2 → US1/US2/US3**：基础数据结构、路由、store 与多语言工具完成后，三个用户故事可并行；建议优先 US1（MVP）。
3. **US1 → US2**：US2 依赖 US1 的版本/审计字段，但通过独立版本服务实现，待 T014-T020 完成即可启动 US2。
4. **US1 → US3**：渠道与批量操作依赖 SPU 主数据、SKU 关联与多语言校验，故需在 US1 完成后开展。
5. **Polish**：在目标用户故事完成后执行，以确保 KPI 与文档更新。

## Parallel Execution Examples

- **US1**：在 T014-T016 完成后，可并行执行 T017（Pinia store）、T018（列表页）、T019（向导）与 T021-T022（SKU/多语言 UI）。
- **US2**：T025 完成后，可并行进行 T028（前端 diff 组件）与 T027（SLA worker）。
- **US3**：T035-T036（后端任务）与 T037-T038（前端渠道 & 批量 UI）可分别由不同开发者推进，同时 T039 负责导入校验。

## Implementation Strategy

1. **MVP（US1）**：完成 Phase1-2 与 T012-T022，即可交付从创建到发布的最小闭环，覆盖多语言与 SKU 关联。
2. **增量扩展**：在 MVP 稳定后，分别追加 US2（版本/审批）与 US3（渠道+批量），每个阶段都具备独立验收标准。
3. **并行团队分工**：
   - 小组 A：后端服务与迁移（T004-T020, T025-T036, T039）。
   - 小组 B：前端页面/交互（T011, T017-T022, T028-T029, T037-T038）。
   - 小组 C：测试与可观测性（T012-T013, T023-T024, T031-T032, T040-T042）。
4. **验证节奏**：每完成一阶段即跑 quickstart + 相关测试（例如完 US1 后运行 T012-T013/T019-T022 验证），确保故事独立可演示并记录 KPI。

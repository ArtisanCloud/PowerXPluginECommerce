# Tasks: Channel Master Data & Authorization

**Input**: Design documents from `/specs/003-channel-master/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/channel.yaml`, `quickstart.md`

**Tests**: 每个用户故事都包含至少一项测试任务，以满足宪章的可测试性交付要求。

**Organization**: 任务按用户故事分组，确保每个故事可独立实现与验收。

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 为后端 channel 模块与 web-admin 页面建立基础骨架。

- [ ] T001 Scaffold channel_master 目录结构（`backend/internal/domain/channel_master/`, `backend/internal/services/admin/channel_master/`, `backend/internal/transport/http/admin/channel_master/`, `backend/internal/observability/channel_master/`, `backend/internal/jobs/channel_master/`）并放置 doc.go 以便编译。
- [ ] T002 初始化 Nuxt 渠道路由骨架：创建 `web-admin/app/pages/channels/index.vue`, `[id].vue`, `approval.vue` 空壳以及 `web-admin/app/app.config.ts` 临时菜单。

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 所有用户故事前置的模型、迁移、Service/DTO 骨架与前端 API 基建。

- [ ] T003 定义 `ChannelMaster` GORM 模型（含 `channel_type`）与常量 (`backend/internal/domain/channel_master/models/channel_master.go`, `backend/internal/domain/models/model.go`) 并补齐 JSON/GORM 标签。
- [ ] T004 创建 `channel_masters` 迁移 + RLS 策略 (`backend/cmd/database/migrate/migrations/20251213_channel_masters.go`) 并注册至 `migrate.go`。
- [ ] T005 定义 `ChannelSyncHistory` 模型 + 迁移 (`backend/internal/domain/channel_master/models/channel_sync_history.go`, `…/migrations/20251213_channel_sync_history.go`) 以记录手动/自动同步历史。
- [ ] T006 定义 `ChannelTaskLink` 与 `ChannelNote` 模型/迁移 (`backend/internal/domain/channel_master/models/channel_task_link.go`, `channel_note.go`) 支撑任务关联与备注。
- [ ] T007 实现 `ChannelMasterRepository` (`backend/internal/domain/channel_master/repository/channel_master_repository.go`) 含 `BeginTenantTx`、联合唯一校验、筛选器与审计读辅助。
- [ ] T008 建立 `ChannelMasterService` 基础接口 (`backend/internal/services/admin/channel_master/service.go`) 注入 repository/audit emitter，并暴露 Draft/Approval stub。
- [ ] T009 添加观测与审计骨架 (`backend/internal/observability/channel_master/{audit_emitter,metrics}.go`) 输出结构化日志、审批 SLA、凭证覆盖率、同步成功率指标。
- [ ] T010 注册 `/v1/channels/**` HTTP 路由与 middleware (`backend/internal/transport/http/admin/channel_master/router.go`) 并创建 `handler.go`/`dto.go` 占位，预留 tasks/notes/sync endpoints。
- [ ] T011 将 `contracts/channel.yaml` 纳入 Spectral 校验流程，并在 `backend/internal/transport/http/admin/channel_master/dto.go` 生成 DTO。
- [ ] T012 建立前端 API client & store 基础 (`web-admin/app/composables/useChannels.ts`, `web-admin/app/stores/channels.ts`, `web-admin/app/types/channels.ts`) 以 `$fetch` + `runtimeConfig.public.apiBaseUrl` 调用 `/v1/channels`。

**Checkpoint**: RLS/模型/契约/日志指标就位，可开始 P1 故事。

---

## Phase 3: User Story 1 - 渠道主数据入驻 (Priority: P1) 🎯 MVP

**Goal**: 渠道运营可创建/编辑渠道、提交审批并支持线下渠道类型。
**Independent Test**: 新建渠道→提交→审批通过→列表展示“未授权”，线下渠道可直接上传线下凭证信息。

### Tests for User Story 1

- [ ] T013 [P] [US1] 编写 `backend/internal/services/admin/channel_master/service_test.go`，覆盖 CreateDraft/Update/Submit/Approve、`channel_type` 验证与审计写入。
- [ ] T014 [P] [US1] 添加 `web-admin/tests/component/ChannelForm.spec.ts` 测试必填校验、线下渠道切换、提交事件。

### Implementation for User Story 1

- [ ] T015 [US1] 实现 Service CRUD/审批逻辑 (`backend/internal/services/admin/channel_master/service.go`)，处理 `channel_type`、租户唯一校验与线下渠道流程，并写入 `channel_audit_logs`。
- [ ] T016 [US1] 扩展 Repository 查询 (`backend/internal/domain/channel_master/repository/channel_master_repository.go`) 提供平台/状态/负责人/区域/GMV/健康度/标签筛选 + 分页。
- [ ] T017 [US1] 实现 HTTP handler (`backend/internal/transport/http/admin/channel_master/handler.go`)：`POST /channels`, `PATCH /channels/{id}`, `POST /channels/{id}/submit`, `POST /channels/{id}/approval`，DTO 支持线下字段。
- [ ] T018 [P] [US1] 构建 `web-admin/app/components/channels/ChannelForm.vue` 包含 UForm、平台/区域/类型联动、线下渠道文件上传占位。
- [ ] T019 [US1] 完成渠道列表页 (`web-admin/app/pages/channels/index.vue`)：表格、筛选器、健康度/标签渲染、创建/编辑抽屉。
- [ ] T020 [US1] 扩展 Pinia store/composable (`web-admin/app/stores/channels.ts`, `app/composables/useChannels.ts`) 以支持分页、创建、编辑、提交、审批动作。
- [ ] T021 [US1] 构建审批工作台 (`web-admin/app/pages/channels/approval.vue`) 展示候审渠道、审批历史、审计日志。

**Checkpoint**: 渠道创建→审批（含线下渠道）可独立演示，形成 MVP。

---

## Phase 4: User Story 2 - 授权与凭证生命周期 (Priority: P1)

**Goal**: 管理 OAuth/手动/线下凭证，巡检有效期并告警。
**Independent Test**: 对渠道完成授权→保存凭证→巡检检测→触发到期提醒。

### Tests for User Story 2

- [ ] T022 [P] [US2] 在 `backend/internal/services/admin/channel_master/credential_service_test.go` 覆盖 envelope encryption、OAuth/线下凭证刷新与过期检测。
- [ ] T023 [P] [US2] 添加 `web-admin/tests/component/ChannelCredentialDrawer.spec.ts` 测试凭证类型切换、scope 校验、上传行为。

### Implementation for User Story 2

- [ ] T024 [US2] 定义 `ChannelCredential` 模型/迁移 (`backend/internal/domain/channel_master/models/channel_credential.go`, `…/migrations/20251213_channel_credentials.go`) 含 scope/status/test_result。
- [ ] T025 [US2] 扩展 `pkg/security/encryption` 与 Service (`backend/internal/services/admin/channel_master/credential_service.go`) 支持 `UpsertCredential`/`TestCredential`、线下凭证附件以及 STS 解密。
- [ ] T026 [US2] 实现 `/channels/{id}/credentials` 与 `/credentials/test` handler (`backend/internal/transport/http/admin/channel_master/credentials_handler.go`) 覆盖 OAuth 回调、线下上传。
- [ ] T027 [US2] 创建凭证巡检 job (`backend/internal/jobs/channel_master/credential_checker.go`) 并在 `cmd/plugin/main.go` 注册，向 `ChannelAlert` 写入即将到期记录。
- [ ] T028 [US2] 定义 `ChannelAlert` 模型/迁移 (`backend/internal/domain/channel_master/models/channel_alert.go`, `…/migrations/20251213_channel_alerts.go`) 支持凭证/同步/KPI 告警。
- [ ] T029 [US2] 在 `backend/internal/observability/channel_master/alert_emitter.go` 完成告警写入 + 通知中心推送，更新 service 调用。
- [ ] T030 [P] [US2] 构建 `web-admin/app/components/channels/ChannelCredentialDrawer.vue`，实现 OAuth/手动/线下凭证表单与文件上传。
- [ ] T031 [US2] 更新渠道详情页 (`web-admin/app/pages/channels/[id].vue`) 展示凭证状态、到期提醒、巡检结果。

**Checkpoint**: 渠道可完成授权，凭证巡检与告警链路生效。

---

## Phase 5: User Story 3 - KPI、健康度、同步历史与告警 (Priority: P2)

**Goal**: 在详情页展示 KPI、健康度评分、告警、同步历史、任务关联与备注。
**Independent Test**: 为单个渠道加载 KPI 面板→模拟异常→生成告警→查看任务/备注→查看同步历史。

### Tests for User Story 3

- [ ] T032 [P] [US3] `backend/internal/services/admin/channel_master/health_service_test.go` 验证闸门式阈值、健康度加权、告警生成。
- [ ] T033 [P] [US3] `web-admin/tests/component/ChannelHealthCard.spec.ts` 检测指标/状态颜色/加载态。
- [ ] T034 [P] [US3] `backend/internal/services/admin/channel_master/task_note_service_test.go` 覆盖任务关联/备注 CRUD 与审计。

### Implementation for User Story 3

- [ ] T035 [US3] 定义 `ChannelMetric` 模型/迁移 (`backend/internal/domain/channel_master/models/channel_metric.go`, `…/migrations/20251213_channel_metrics.go`) 支持 d1/d7/d30。
- [ ] T036 [US3] 在 service 新增 `LoadMetrics` + `ComputeHealthScore` (`backend/internal/services/admin/channel_master/metrics_service.go`) 聚合 KPI/健康度/同步状态。
- [ ] T037 [US3] 实现 `/channels/{id}` 详情补充字段、`/channels/{id}/alerts` GET/PATCH、`/channels/{id}/sync` 触发等 handler (`backend/internal/transport/http/admin/channel_master/detail_handler.go`, `alerts_handler.go`, `sync_handler.go`)；返回 KPI、sync 历史摘要。
- [ ] T038 [US3] 创建 KPI/健康度刷新 job (`backend/internal/jobs/channel_master/metric_refresh.go`) 读取任务中心视图并写入 `channel_metrics`。
- [ ] T039 [US3] 实现 `ChannelTaskLinkRepository`/service (`backend/internal/domain/channel_master/repository/channel_task_link_repository.go`, `services/.../task_note_service.go`) 提供任务关联 CRUD + 审计。
- [ ] T040 [US3] 实现 `ChannelNoteRepository`/service (`backend/internal/domain/channel_master/repository/channel_note_repository.go`) 提供备注增删查。
- [ ] T041 [US3] 暴露 `/channels/{id}/tasks` 与 `/channels/{id}/notes` API (`backend/internal/transport/http/admin/channel_master/task_note_handler.go`) 支持关联/解除/备注。
- [ ] T042 [US3] 实现 `ChannelSyncHistoryRepository` 与 service (`backend/internal/services/admin/channel_master/sync_history_service.go`) 记录手动/自动同步结果、耗时、触发人。
- [ ] T043 [US3] 在 detail handler 暴露 `/channels/{id}/sync-history` GET，并将手动触发写入 `channel_sync_history`。
- [ ] T044 [US3] 构建健康度/KPI 组件 (`web-admin/app/components/channels/ChannelHealthCard.vue`, `ChannelKpiTrend.vue`) 并接入数据。
- [ ] T045 [US3] 构建告警与任务/备注组件 (`web-admin/app/components/channels/ChannelAlertTimeline.vue`, `ChannelTaskPanel.vue`, `ChannelNotePanel.vue`) 提供关联/指派/备注能力。
- [ ] T046 [US3] 构建同步历史组件 (`web-admin/app/components/channels/ChannelSyncHistory.vue`) 及 `[id].vue` 集成，展示最近 N 条记录与手动触发入口。
- [ ] T047 [US3] 扩展 store/composable (`web-admin/app/stores/channels.ts`, `app/composables/useChannels.ts`) 以轮询 KPI、告警、任务/备注、同步历史。

**Checkpoint**: KPI/健康度、告警、任务关联、备注、同步历史链路可独立演示。

---

## Phase 6: User Story 4 - 策略配置与团队权限 (Priority: P3)

**Goal**: 渠道经理配置策略/团队/审批人并受 RBAC 限制。
**Independent Test**: 在渠道详情中更新策略/团队→仅授权角色可见→审计可检索。

### Tests for User Story 4

- [ ] T048 [P] [US4] `backend/internal/services/admin/channel_master/strategy_service_test.go` 覆盖策略写入、团队权限、审计断言。

### Implementation for User Story 4

- [ ] T049 [US4] 定义 `ChannelConfig` 模型/迁移 (`backend/internal/domain/channel_master/models/channel_config.go`, `…/migrations/20251213_channel_configs.go`) 记录 pricebook/inventory/logistics/cs/财务字段。
- [ ] T050 [US4] 实现策略/团队 service (`backend/internal/services/admin/channel_master/strategy_service.go`) 校验引用、fee_rate、审批人、负责人及 audit。
- [ ] T051 [US4] 暴露策略/权限 API (`backend/internal/transport/http/admin/channel_master/strategy_handler.go`) 或扩展 `/v1/channels/{id}` PATCH，返回最新 config。
- [ ] T052 [US4] 更新 RBAC/manifest (`backend/internal/transport/http/middleware/permission.go`, `plugin.yaml`) 限定策略与团队指派权限。
- [ ] T053 [P] [US4] 构建策略编辑组件 (`web-admin/app/components/channels/ChannelStrategyForm.vue`) 并在详情页 Tab 呈现。
- [ ] T054 [US4] 构建团队/审批配置 UI (`web-admin/app/components/channels/ChannelTeamSection.vue`) 并在 `[id].vue` 集成，结合 RBAC 限制。

**Checkpoint**: 策略/团队配置可独立演示且受权限控制。

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 文档、配置、性能、指标与交付验证。

- [ ] T055 更新 `specs/003-channel-master/quickstart.md` 与 `docs/` 相关章节，记录授权/KPI/任务关联/同步历史操作说明。
- [ ] T056 更新 `plugin.yaml`（菜单、RBAC scope、version）、`web-admin/app/i18n/` 多语言文案确保渠道导航可见。
- [ ] T057 为 SC-001~SC-005 添加仪表与监控：在 `backend/internal/observability/channel_master/metrics.go` 暴露审批耗时、凭证覆盖率、同步成功率、KPI 加载直方图，并编写 `docs/observability/channel_master.md` 使用说明。
- [ ] T058 在 web-admin 添加 KPI 加载性能探针（`web-admin/app/plugins/perf.client.ts` 或 composable）并记录到浏览器指标/日志，验证 ≤3s；同时在 `reports/` 记录手动测试结果。
- [ ] T059 依照 `quickstart.md` 执行 `make test`, `make frontend-build`, `npm run test`，并在 `reports/003-channel-master-validation.md` 写入结果与成功指标汇总。

---

## Dependencies & Execution Order

- Phase 1 → Phase 2：需先完成目录与路由骨架，再落地模型/契约。
- Phase 2 → Phases 3-6：所有故事依赖共享模型、RLS、Service skeleton、观测指标与 API 契约。
- US1/US2 (P1) 可并行；US3 依赖 ChannelAlert/TaskLink/SyncHistory（在 Phase2/Phase5 中构建）；US4 依赖 ChannelMaster/credential 已完成即可启动。
- Phase 7 在目标故事完成后执行。

### Story Completion Order
1. **US1 (P1)** — 渠道入驻 + 审批 + 线下渠道。
2. **US2 (P1)** — 授权与凭证巡检。
3. **US3 (P2)** — KPI、健康度、任务/备注、同步历史、告警。
4. **US4 (P3)** — 策略配置与团队权限。

### Parallel Opportunities
- Phase 2 中模型/迁移（T003-T006）与服务/观测（T007-T011）可多⼈并行。
- US1/US2 前后端任务（T013-T031）互不冲突，可双线推进。
- US3 的 KPI/同步任务（T035-T047）与 US4 的策略任务（T048-T054）可在前述故事完成后并发。
- Polish 阶段 T057-T059 中仪表 vs 测试执行可并行。

---

## Parallel Example: User Story 1

```
# Backend / frontend 并行：
Task T015: Service CRUD/审批逻辑（含 channel_type 与线下流程）
Task T016: Repository filters/pagination
Task T018: ChannelForm（支持线下凭证上传）
Task T019: 渠道路由列表 UI
```

---

## Implementation Strategy

### MVP First (User Story 1)
1. 完成 Phase 1-2。
2. 实现 US1 (T013-T021) 并通过测试，交付渠道入驻 MVP。
3. 演示/验证后再进入 US2-4。

### Incremental Delivery
- Iteration 1：US1（入驻 + 审批）。
- Iteration 2：US2（授权 + 凭证巡检 + 告警）。
- Iteration 3：US3（KPI/健康度 + 任务/备注 + 同步历史）。
- Iteration 4：US4（策略/团队）。
每次迭代结束执行 Phase 7 检查，确保指标/文档齐备。

### Parallel Team Strategy
- Phase 1-2 全员共建。
- Phase 3 起按专长拆分：Dev A 负责后端（T015-T043），Dev B 着手前端（T018-T047），Dev C 负责策略/RBAC（T048-T054），Dev D 负责观测/文档/测试（T055-T059）。
- 各故事完成即合并，保持可独立验收。

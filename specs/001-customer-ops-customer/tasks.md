---

description: "Task list for implementing Customer List & Membership Views"
---

# Tasks: Customer List & Membership Views

**Input**: Design documents from `/specs/001-customer-ops-customer/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/customers-api.md, quickstart.md

**Tests**: 覆盖列表筛选、会员提醒与导入/导出任务的核心路径；每个用户故事至少包含一项验证任务。

**Organization**: 依用户故事拆分阶段，确保每一阶段可独立交付和测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行处理（不同文件且无直接依赖）。
- **[Story]**: 关联的用户故事标签（US1/US2/US3）。
- 描述需包含明确文件路径。

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 确认宿主运行配置与文档入口，避免后续实现受阻。

- [X] T001 在 `specs/001-customer-ops-customer/plan.md`（或 quickstart）记录如何使用脚手架已提供的 `runtimeConfig.public.apiBaseUrl/tenantUuid/audit` 配置，仅做复核和文档说明，不改动 `web-admin/nuxt.config.ts`。
- [X] T002 [P] 在 `web-admin/README.md` 增补本特性所需的开发前提（API Base、STS、任务中心入口）与运行命令。

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 搭好跨用户故事复用的类型、API Service、Store 与批量动作通用逻辑。

- [X] T003 [P] 在 `web-admin/app/types/customer.ts` 定义 `Customer`、`SavedView`、`BulkTask`、`MembershipSnapshot` 等接口与遮罩字段标识。
- [X] T004 [P] 新增 `web-admin/app/composables/api/services/customerService.ts`，封装 `/api/customers`、`/customers/members`、`/customers/bulk-actions`、`/customers/import`、`/customers/export`、`/customers/bulk-remind`、`/jobs/{id}` 请求方法。
- [X] T005 在 `web-admin/app/stores/customer/index.ts` 建立 Pinia store 初始结构（filters、savedViews、selection、bulkTasks、membershipStats）与空 actions。
- [X] T006 [P] 创建 `web-admin/app/composables/useCustomerBulkActions.ts`，集中处理批量动作参数校验、任务中心轮询、审计上下文拼装。
- [X] T007 [P] 新建 `web-admin/app/composables/useCustomerMetrics.ts`，记录筛选交互耗时、导出任务耗时、保级提醒成功率与错误率指标，对应 SC-001~SC-004。
- [X] T008 [P] 扩展 `web-admin/app/plugins/metrics.client.ts` 及相关配置，注入 `useCustomerMetrics` 事件上报（含过滤点击计数、任务完成时间、提醒结果统计）。

---

## Phase 3: User Story 1 - Operate Customer Directory (Priority: P1) 🎯 MVP

**Goal**: 让客服在 `/customer` 页面中多维筛选、保存视图、查看详情抽屉，并执行批量标签/负责人/禁用等动作且可追踪。

**Independent Test**: 通过 UI 完成“筛选→保存视图→批量指派负责人→查看详情抽屉”流程，并在任务中心看到批量任务记录。

### Implementation for User Story 1

- [ ] T009 [US1] 在 `web-admin/app/stores/customer/index.ts` 补齐列表查询、分页、行选择、保存/切换视图与敏感字段遮罩的 actions。
- [ ] T010 [P] [US1] 新建 `web-admin/app/components/customer/CustomerFilterBar.vue`，实现关键词+高级筛选（地区、风险、等级、标签等）与保存/加载视图操作。
- [ ] T011 [P] [US1] 新建 `web-admin/app/components/customer/CustomerTable.vue`，使用 Nuxt UI `UTable` 呈现可配置列、复选框、空状态、权限遮罩。
- [ ] T012 [US1] 新建 `web-admin/app/components/customer/CustomerDetailDrawer.vue`，含概览、订单、售后、权益、审计等 Tab，并支持快捷动作（发券、备注）。
- [ ] T013 [P] [US1] 创建 `web-admin/app/components/customer/CustomerBulkActions.vue`，调用 `useCustomerBulkActions` 触发批量标签/负责人/禁用，并在 UI 中展示任务状态。
- [ ] T014 [US1] 在 `web-admin/app/pages/customer/index.vue` 组合 FilterBar、Table、BulkActions、DetailDrawer，处理路由 Query 同步、空态提示、通知反馈，并接入 `useCustomerMetrics` 统计筛选点击与响应耗时。
- [ ] T015 [US1] 在 `web-admin/app/pages/customer/index.vue` 与相关组件中实现 FR-009 错误提示/重试链路（含批量操作/列表失败的 toast、retry CTA、权限不足提示）。
- [ ] T016 [P] [US1] 更新 `web-admin/i18n/en/menus.json` 与 `web-admin/i18n/zh-CN/menus.json`，补充筛选项、批量操作、抽屉 Tab、错误提示相关文案。
- [ ] T017 [US1] 编写 `web-admin/tests/unit/customer-list.spec.ts` 与 `web-admin/tests/e2e/customer-directory.cy.ts`，覆盖筛选保存、批量负责人、详情抽屉遮罩，以及错误/重试场景与 KPI 事件触发。

---

## Phase 4: User Story 2 - 会员视角洞察与保级 (Priority: P2)

**Goal**: 在 `/customer/members` 提供等级/成长值筛选、指标卡、即将降级分群与批量保级提醒。

**Independent Test**: 设置“金卡 + 成长值<50”过滤，查看指标卡刷新，并对子集发起批量提醒，确认任务中心与审计日志记录。

### Implementation for User Story 2

- [ ] T018 [US2] 新建 `web-admin/app/composables/useMembershipInsights.ts`，封装 `/customers/members` 查询、指标聚合与保级状态计算逻辑。
- [ ] T019 [US2] 扩展 `web-admin/app/stores/customer/index.ts`，加入 `membershipSnapshots`、指标卡统计与 `bulkReminder` 相关 actions。
- [ ] T020 [P] [US2] 新建 `web-admin/app/components/customer/MembershipCards.vue` 展示总人数、活跃、即将降级、平均成长值等 KPI。
- [ ] T021 [P] [US2] 新建 `web-admin/app/components/customer/MembershipFilterBar.vue`，支持等级、成长值区间、积分、权益状态、保级状态筛选。
- [ ] T022 [US2] 在 `web-admin/app/pages/customer/members.vue` 整合过滤器、指标卡、表格与分群 Tag，接入 `useCustomerMetrics` 统计保级分群、筛选耗时，并提供“即将降级”快捷筛选与导出入口。
- [ ] T023 [US2] 新建 `web-admin/app/components/customer/MembershipReminderDrawer.vue`，配置渠道/模板、记录提醒成功率/失败率（SC-003/SC-004），并调用 `useCustomerBulkActions` 的 `bulk-remind` 能力。
- [ ] T024 [P] [US2] 创建 `web-admin/tests/e2e/membership-view.cy.ts`（含 KPI 事件）与必要的 `web-admin/tests/unit/membership-store.spec.ts`，验证筛选刷新、提醒任务触达、结果统计与审计记录。

---

## Phase 5: User Story 3 - 可管控的导入/导出与审计 (Priority: P3)

**Goal**: 支持模板下载、导入校验、按筛选条件导出，并在任务中心/审计中可追踪且受权限控制。

**Independent Test**: 无权限用户尝试导出被拒；有权限用户下载模板→上传导入→在任务中心查看进度→导出结果并通过签名链接下载。

### Implementation for User Story 3

- [ ] T025 [US3] 扩展 `web-admin/app/stores/customer/index.ts`，实现导入/导出/任务轮询 actions，并复用 `BulkTask` 状态在页面展示。
- [ ] T026 [P] [US3] 新建 `web-admin/app/components/customer/CustomerImportDialog.vue`，提供模板下载、文件上传、校验反馈与任务链接。
- [ ] T027 [P] [US3] 新建 `web-admin/app/components/customer/CustomerExportDialog.vue`，支持字段选择、筛选摘要展示及任务完成后的下载提示。
- [ ] T028 [US3] 在 `web-admin/app/pages/customer/index.vue` 接入导入/导出入口、权限判断（`customer.manage`/`customer.export`）、任务中心提醒，并向 `useCustomerMetrics` 上报导入/导出任务耗时。
- [ ] T029 [US3] 在 `web-admin/app/pages/customer/members.vue` 复用导入/导出入口，区分会员视角默认字段与筛选摘要，并上报 KPI。
- [ ] T030 [US3] 在导入/导出流程与任务轮询中实现 FR-009 错误提示/重试（含文件校验错误报告、签名链接失效后的提示与再触发），并计入错误率统计。
- [ ] T031 [US3] 更新 `web-admin/app/composables/api/_client.ts`，为导入/导出/批量操作请求统一注入 `X-Audit-Action`、`X-Audit-Resource` 头、捕获错误信息并暴露给 UI。
- [ ] T032 [P] [US3] 编写 `web-admin/tests/unit/customer-import-export.spec.ts` 与 `web-admin/tests/e2e/customer-import-export.cy.ts`，模拟权限缺失/成功/错误重试场景、下载链接签名校验与 KPI 统计。

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 文档、运行指南与收尾工作，确保易用与可运维。

- [ ] T033 [P] 在 `docs/plan/customer/readme.md` 记录新的客户目录/会员视角能力、指标埋点与依赖组件，便于后续团队查阅。
- [ ] T034 [P] 扩充 `specs/001-customer-ops-customer/quickstart.md`，加入 KPI/错误提示验证步骤（筛选≤3 次点击、导出耗时、提醒成功率、错误率观察）及任务中心检查指引。
- [ ] T035 [P] 更新 `docs/plan/customer/customer.md`，补充与任务中心、审计、批量操作、错误处理/指标监控的关系及上线注意事项。

---

## Dependencies & Execution Order

1. **Setup → Foundational**：必须先完成运行配置与文档（Phase 1），再完成类型、API、Store、批量动作基础设施（Phase 2）。
2. **User Story 顺序**：在 Phase 2 完成后，可按优先级交付：US1 (目录) → US2 (会员) → US3 (导入/导出)。若人力充足，US2/US3 可并行但需依赖 US1 中 store/table 的成熟度。
3. **Polish**：Phase 6 依赖所有计划交付的用户故事完成后再开展文档与 Runbook 更新。

## Parallel Execution Opportunities

- Phase 1: T001、T002 触达不同文件，可并行。
- Phase 2: T003/T004/T006 可与 T005 并行，`useCustomerMetrics`（T007/T008）也可在类型约束明确后同步推进。
- US1: 组件开发（T010-T013）可并行，最后在 T014 汇总；i18n (T016) 与测试 (T017) 可在页面联调后独立推进。
- US2: T020 与 T021 可并行制作 UI，提醒抽屉 T023 可在 store (T019) 完成后同步；测试 T024 可等待主流程完成再独立运行。
- US3: 导入/导出对话框（T026、T027）可在 store 扩展 (T025) 的接口契约确定后同步推进；测试 T032 与页面集成 (T028、T029) 并行。

## Implementation Strategy

- **MVP**: 完成 Phase 1~3（US1）即可 Demo：客服能筛选客户、保存视图、触发批量任务并查看详情。
- **Incremental**: 在 MVP 稳定后，交付 US2 以支持会员洞察，再交付 US3 以补齐导入/导出合规链路，最后执行 Polish。
- **Quality Gates**: 每个用户故事结束前，确保对应测试任务（T017、T024、T032）通过，并按 quickstart 指南验证任务中心与审计记录。

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

- [X] T009 [US1] 在 `web-admin/app/stores/customer/index.ts` 补齐列表查询、分页、行选择、保存/切换视图与敏感字段遮罩的 actions。
- [X] T010 [P] [US1] 新建 `web-admin/app/components/customer/CustomerFilterBar.vue`，实现关键词+高级筛选（地区、风险、等级、标签等）与保存/加载视图操作。
- [X] T011 [P] [US1] 新建 `web-admin/app/components/customer/CustomerTable.vue`，使用 Nuxt UI `UTable` 呈现可配置列、复选框、空状态、权限遮罩。
- [X] T012 [US1] 新建 `web-admin/app/components/customer/CustomerDetailDrawer.vue`，含概览、订单、售后、权益、审计等 Tab，并支持快捷动作（发券、备注）。
- [X] T013 [P] [US1] 创建 `web-admin/app/components/customer/CustomerBulkActions.vue`，调用 `useCustomerBulkActions` 触发批量标签/负责人/禁用，并在 UI 中展示任务状态。
- [X] T014 [US1] 在 `web-admin/app/pages/customer/index.vue` 组合 FilterBar、Table、BulkActions、DetailDrawer，处理路由 Query 同步、空态提示、通知反馈，并接入 `useCustomerMetrics` 统计筛选点击与响应耗时。
- [X] T015 [US1] 在 `web-admin/app/pages/customer/index.vue` 与相关组件中实现 FR-009 错误提示/重试链路（含批量操作/列表失败的 toast、retry CTA、权限不足提示）。
- [X] T016 [P] [US1] 更新 `web-admin/i18n/en/menus.json` 与 `web-admin/i18n/zh-CN/menus.json`，补充筛选项、批量操作、抽屉 Tab、错误提示相关文案。
- [X] T017 [US1] 编写 `web-admin/tests/unit/customer-list.spec.ts` 与 `web-admin/tests/e2e/customer-directory.cy.ts`，覆盖筛选保存、批量负责人、详情抽屉遮罩，以及错误/重试场景与 KPI 事件触发。

---

## Phase 4: User Story 2 - 会员视角洞察与保级 (Priority: P2)

**Goal**: 在 `/customer/members` 提供等级/成长值筛选、指标卡、即将降级分群与批量保级提醒。

**Independent Test**: 设置“金卡 + 成长值<50”过滤，查看指标卡刷新，并对子集发起批量提醒，确认任务中心与审计日志记录。

### Implementation for User Story 2

- [X] T018 [US2] 新建 `web-admin/app/composables/useMembershipInsights.ts`，封装 `/customers/members` 查询、指标聚合与保级状态计算逻辑。
- [X] T019 [US2] 扩展 `web-admin/app/stores/customer/index.ts`，加入 `membershipSnapshots`、指标卡统计与 `bulkReminder` 相关 actions。
- [X] T020 [P] [US2] 新建 `web-admin/app/components/customer/MembershipCards.vue` 展示总人数、活跃、即将降级、平均成长值等 KPI。
- [X] T021 [P] [US2] 新建 `web-admin/app/components/customer/MembershipFilterBar.vue`，支持等级、成长值区间、积分、权益状态、保级状态筛选。
- [X] T022 [US2] 在 `web-admin/app/pages/customer/members.vue` 整合过滤器、指标卡、表格与分群 Tag，接入 `useCustomerMetrics` 统计保级分群、筛选耗时，并提供“即将降级”快捷筛选与导出入口。
- [X] T023 [US2] 新建 `web-admin/app/components/customer/MembershipReminderDrawer.vue`，配置渠道/模板、记录提醒成功率/失败率（SC-003/SC-004），并调用 `useCustomerBulkActions` 的 `bulk-remind` 能力。
- [X] T024 [P] [US2] 创建 `web-admin/tests/e2e/membership-view.cy.ts`（含 KPI 事件）与必要的 `web-admin/tests/unit/membership-store.spec.ts`，验证筛选刷新、提醒任务触达、结果统计与审计记录。

---

## Phase 5: User Story 3 - 可管控的导入/导出与审计 (Priority: P3)

**Goal**: 支持模板下载、导入校验、按筛选条件导出，并在任务中心/审计中可追踪且受权限控制。

**Independent Test**: 无权限用户尝试导出被拒；有权限用户下载模板→上传导入→在任务中心查看进度→导出结果并通过签名链接下载。

### Implementation for User Story 3

- [X] T025 [US3] 扩展 `web-admin/app/stores/customer/index.ts`，实现导入/导出/任务轮询 actions，并复用 `BulkTask` 状态在页面展示。
- [X] T026 [P] [US3] 新建 `web-admin/app/components/customer/CustomerImportDialog.vue`，提供模板下载、文件上传、校验反馈与任务链接。
- [X] T027 [P] [US3] 新建 `web-admin/app/components/customer/CustomerExportDialog.vue`，支持字段选择、筛选摘要展示及任务完成后的下载提示。
- [X] T028 [US3] 在 `web-admin/app/pages/customer/index.vue` 接入导入/导出入口、权限判断（`customer.manage`/`customer.export`）、任务中心提醒，并向 `useCustomerMetrics` 上报导入/导出任务耗时。
- [X] T029 [US3] 在 `web-admin/app/pages/customer/members.vue` 复用导入/导出入口，区分会员视角默认字段与筛选摘要，并上报 KPI。
- [X] T030 [US3] 在导入/导出流程与任务轮询中实现 FR-009 错误提示/重试（含文件校验错误报告、签名链接失效后的提示与再触发），并计入错误率统计。
- [X] T031 [US3] 更新 `web-admin/app/composables/api/_client.ts`，统一处理批量任务请求错误并暴露给 UI（不再新增自定义头）。
- [X] T032 [P] [US3] 编写 `web-admin/tests/unit/customer-import-export.spec.ts` 与 `web-admin/tests/e2e/customer-import-export.cy.ts`，模拟权限缺失/成功/错误重试场景、下载链接签名校验与 KPI 统计。

---

## Phase 6: User Story 4 - 客户 CRUD 接入 (Priority: P1+)

**Goal**: 当宿主 CRM 开放写入权限时，提供端到端的客户创建/编辑/删除能力，并符合权限、审计与数据一致性要求。

**Independent Test**: 在 `/customer` 页面发起创建→列表出现→抽屉内编辑→在更多操作中删除，验证 `customer.manage/customer.delete` 权限以及任务中心事件都按 `docs/plan/customer/customer.md` 第 8 节记录。

### Implementation for User Story 4

- [X] T036 [US4] 在 `backend/internal/transport/http/customer/routes.go` 与 `backend/internal/transport/http/customer/rbac.go` 注册 `POST|PATCH|DELETE /api/admin/customers` 路由，绑定 `handler.CreateCustomer/UpdateCustomer/DeleteCustomer`，并配置 `customer.manage`、`customer.delete` 权限映射及审计动作常量。
- [X] T037 [P] [US4] 在 `backend/internal/transport/http/customer/handler.go`（及 `dto.go` 如有）实现 CRUD Handler：包含请求体校验（手机/邮箱/标签长度、会员等级枚举）与 403/409/422 错误映射，并复用现有上下文记录日志。
- [X] T038 [US4] 扩展 `backend/internal/services/customer/service.go`、`internal/services/customer/types.go` 与 CRM Client（REST/gRPC adapter），提供 `Create/Update/Delete` 方法、写操作后的 `CustomerChanged` 事件发布与缓存刷新，并处理宿主 409/400/500 错误。
- [X] T039 [P] [US4] 在 `backend/internal/transport/http/customer/handler_test.go` 与 `backend/internal/services/customer/service_test.go` 编写单元测试，覆盖 201/204 成功、403 权限拒绝、409 并发冲突与宿主错误透传。
- [X] T040 [US4] 更新 `web-admin/app/composables/api/services/customerService.ts`，补充 `createCustomer`, `updateCustomer`, `deleteCustomer` 方法，支持删除原因参数与宿主 409/400 错误重试提示。
- [X] T041 [US4] 扩展 `web-admin/app/stores/customer/index.ts`，新增对应 actions 与 loading/error 状态；创建/编辑成功需刷新 `list`、`savedViews`、`membershipSnapshots`，删除后清理当前 selection，并在 `bulkTasks` 中记录 `CustomerChanged` 事件。
- [X] T042 [P] [US4] 在 `web-admin/app/components/customer/CreateCustomerModal.vue` 与 `EditCustomerModal.vue` 实现 `UForm` + schema 校验、初始值、提交禁用/Toast，并支持会员等级、来源渠道、标签等字段映射。
- [X] T043 [US4] 在 `web-admin/app/components/customer/CustomerDetailDrawer.vue`、`CustomerTable.vue` 与 `web-admin/app/pages/customer/index.vue` 启用“创建/编辑/删除”入口：列表顶部按钮、行级操作、抽屉快捷编辑，成功后调用 store action 并刷新数据。
- [X] T044 [P] [US4] 新增 `web-admin/app/components/customer/CustomerDeleteConfirm.vue`（或同等弹窗），要求输入客户名称/删除原因，调用 `store.deleteCustomer` 并在失败时展示宿主返回的保护信息；同步在抽屉/表格中引用。
- [X] T045 [P] [US4] 在 `web-admin/app/pages/customer/index.vue` 与 `web-admin/app/pages/customer/members.vue` 加入宿主写入可用性检测（来自 featureFlag/health API），当不可用时禁用创建/编辑按钮并显示 Banner，提示“客户写入能力暂不可用”。
- [X] T046 [US4] 编写 `web-admin/tests/unit/customer-crud-store.spec.ts` 与 `web-admin/tests/e2e/customer-crud.cy.ts`，覆盖创建→刷新列表、编辑 diff 提交、删除需输入原因、宿主 409/400 提示、写入禁用 Banner 以及 KPI 事件触发。
- [X] T047 [US4] 新增 `backend/internal/entity/models/customer/customer.go` 与 `backend/cmd/database/migrate/migrate.go` 迁移项：定义 `customers` 表字段、JSON 列、租户索引，并确保 `make migrate` 会创建该表。
- [X] T048 [US4] 改造 `backend/internal/services/customer/service.go` / `mutations.go`：优先使用 GORM CRUD（含 mock seed、JSON 编解码、冲突检测），保留内存 fallback，并更新 `service_test.go` 使用 sqlite 验证。

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 文档、运行指南与收尾工作，确保易用与可运维。

- [X] T033 [P] 在 `docs/plan/customer/readme.md` 记录新的客户目录/会员视角能力、指标埋点与依赖组件，便于后续团队查阅。
- [X] T034 [P] 扩充 `specs/001-customer-ops-customer/quickstart.md`，加入 KPI/错误提示验证步骤（筛选≤3 次点击、导出耗时、提醒成功率、错误率观察）及任务中心检查指引。
- [X] T035 [P] 更新 `docs/plan/customer/customer.md`，补充与任务中心、审计、批量操作、错误处理/指标监控的关系及上线注意事项。

---

## Dependencies & Execution Order

1. **Setup → Foundational**：必须先完成运行配置与文档（Phase 1），再完成类型、API、Store、批量动作基础设施（Phase 2）。
2. **User Story 顺序**：在 Phase 2 完成后，可按优先级交付：US1 (目录) → US2 (会员) → US3 (导入/导出) → US4 (CRUD)。若人力充足，US2/US3 可并行但需依赖 US1 中 store/table 的成熟度；US4 需等宿主开放写入权限后再执行。
3. **Polish**：Phase 7 依赖所有计划交付的用户故事完成后再开展文档与 Runbook 更新。

## Parallel Execution Opportunities

- Phase 1: T001、T002 触达不同文件，可并行。
- Phase 2: T003/T004/T006 可与 T005 并行，`useCustomerMetrics`（T007/T008）也可在类型约束明确后同步推进。
- US1: 组件开发（T010-T013）可并行，最后在 T014 汇总；i18n (T016) 与测试 (T017) 可在页面联调后独立推进。
- US2: T020 与 T021 可并行制作 UI，提醒抽屉 T023 可在 store (T019) 完成后同步；测试 T024 可等待主流程完成再独立运行。
- US3: 导入/导出对话框（T026、T027）可在 store 扩展 (T025) 的接口契约确定后同步推进；测试 T032 与页面集成 (T028、T029) 并行。

## Implementation Strategy

- **MVP**: 完成 Phase 1~3（US1）即可 Demo：客服能筛选客户、保存视图、触发批量任务并查看详情。
- **Incremental**: 在 MVP 稳定后，交付 US2 以支持会员洞察，再交付 US3 以补齐导入/导出合规链路，宿主开放写入后推进 US4（CRUD），最后执行 Polish。
- **Quality Gates**: 每个用户故事结束前，确保对应测试任务（T017、T024、T032、T046）通过，并按 quickstart 指南验证任务中心与审计记录。

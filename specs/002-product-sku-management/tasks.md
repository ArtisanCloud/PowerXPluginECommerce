# Tasks: SKU 与变体管理能力增强

**Input**: Design documents from `/specs/002-product-sku-management/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 规范未要求先写测试，本任务集以内建验收步骤为准；若后续需 TDD，可在相关阶段附加用例。

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 准备本地/CI 环境、配置宿主依赖，确保多租户与前后端共用基础一致。

- [X] T001 同步并校验 `backend/etc/config.yaml` 与 `plugin.yaml`，确保新增 API 占位及菜单配置齐备
- [X] T002 [P] 初始化 SKU 相关 Nuxt 运行配置（`web-admin/nuxt.config.ts` 与 `app.config.ts`）以暴露 `/_p/<plugin-id>/api/v1/products/skus` 基础 URL
- [X] T003 [P] 准备前端类型/常量定义文件（`web-admin/app/types/product/sku.ts`）供所有页面复用

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 夯实数据模型、仓储、服务骨架与前端状态管理，完成前置审批/任务框架。

- [X] T004 创建 SKU/渠道/库存/批量任务迁移脚本（`backend/cmd/database/migrate/migrations/002_product_sku.go`）并注册到 `migrate.go`
- [X] T005 [P] 定义领域模型结构体与表常量（`backend/internal/entity/models/product_sku/*.go`）含 gorm/json 标签
- [X] T006 [P] 搭建 SKU、渠道、库存、批量任务仓储实现（`backend/internal/entity/repository/product_sku/repository.go`）附多租户查询封装
- [X] T007 [P] 建立服务层骨架与接口（`backend/internal/services/admin/product_sku/service.go`），含生成、批量、渠道、库存子模块占位
- [X] T008 配置 HTTP 路由与中间件（`backend/internal/transport/http/admin/product_sku/router.go`）映射 contracts 中所有新端点
- [X] T009 构建 Pinia store 占位（`web-admin/app/stores/productSku.ts`）并声明状态片段/动作接口
- [X] T010 [P] 创建公用组件壳与布局（`web-admin/app/components/product/sku/index.ts`、`web-admin/app/layouts/sku-dashboard.vue`）便于后续故事引用
- [X] T011 对 `backend/internal/{domain,repository,services}/product_sku` 目录进行命名调整或文档化豁免，确保遵守 lower_snake_case 宪章要求
- [X] T012 扩展序列号/批次追踪基础（`backend/internal/entity/models/product_sku/serial_batch.go`、`repository_serial.go`、`services/admin/product_sku/serial_service.go`），并串联审计日志钩子
- [X] T055 [P] 实现 SKU 列表查询 API（`backend/internal/services/admin/product_sku/service.go`、`.../handler.go`）含分页/筛选与 RBAC，对接 `GET /api/v1/admin/product/skus`

**Checkpoint**: 完成上述任务后方可进入各用户故事，确保服务/前端具备统一依赖。

---

## Phase 3: User Story 1 - 运营批量生成 SKU (Priority: P1) 🎯 MVP

**Goal**: 运营可在 SPU 编辑页通过生成器选择规格组合、设置默认字段并一次性创建 SKU。
**Independent Test**: 给定含多规格 SPU，运行生成器勾选组合并提交后，`product_skus` 表生成对应记录且前端列表可查看。

### Implementation for User Story 1

- [X] T013 [US1] 实现笛卡尔组合/默认值计算器（`backend/internal/services/admin/product_sku/generator.go`）含规格合法性校验
- [X] T014 [US1] 扩展仓储写入批量创建/去重逻辑（`backend/internal/entity/repository/product_sku/repository.go`）供生成器调用
- [X] T015 [US1] 实现生成器 API Handler（`backend/internal/transport/http/admin/product_sku/generator_handler.go`）对接 `/products/spus/{spuId}/skus/generate`
- [X] T016 [US1] 新增 SKU 创建/复制接口（`backend/internal/transport/http/admin/product_sku/skus_handler.go`）可写入默认条码/重量
- [X] T017 [P] [US1] 构建前端生成器核心组件（`web-admin/app/components/product/sku/GeneratorPanel.vue`）展示规格树与组合勾选
- [X] T018 [P] [US1] 实现默认字段表单与模板规则（`web-admin/app/components/product/sku/DefaultValueForm.vue`）支持条码前缀/重量/尺寸
- [X] T019 [US1] 将生成器入口接入 SPU 编辑页（`web-admin/app/pages/product/spus/edit/[id].vue`）含权限校验与交互反馈
- [X] T020 [US1] 更新 Pinia store 与 API 客户端（`web-admin/app/stores/productSku.ts`、`web-admin/app/services/productSku.ts`）以保存生成结果并刷新 SKU 列表
- [X] T021 [US1] 在 SKU 列表/矩阵页展示新建结果与冲突提示（`web-admin/app/pages/product/skus/index.vue`、`web-admin/app/components/product/sku/MatrixGrid.vue`）
- [X] T022 [US1] 处理 SPU 无规格/单规格场景（`backend/internal/services/admin/product_sku/generator.go`）确保跳过矩阵、仍可生成单一 SKU
- [X] T023 [US1] 在前端生成器中提供单规格 UI/提示（`web-admin/app/components/product/sku/GeneratorPanel.vue`）并保证冲突提示一致
- [X] T058 [US1] 将 SKU 列表/矩阵页面从 mock 改为真实 API（`web-admin/app/pages/product/skus/index.vue`、`web-admin/app/stores/productSku.ts`），保存后自动刷新。
- [X] T059 [US1] 在 SPU 编辑页渲染关联 SKU 摘要（`web-admin/app/pages/product/spus/edit/[id].vue`），并在“保存关联”成功后刷新列表供运营确认。
- [X] T060 [US1] 调整 SKU Upsert 去重策略（`backend/internal/services/admin/product_sku/upsert.go`、`spu/sku_link_service.go`），确保同一 SPU 内以 SKU 编码唯一为准，payload 仅作用于版本而非最终数据，同时完善插入 `product_sku_attributes` 的同步与测试。
- [X] T061 [US1] 更新 SPU 编辑页摘要与相关 API（`web-admin/app/pages/product/spus/edit/[id].vue`、`app/composables/api/useSku.ts`），新增基于 `GET /api/v1/admin/product/skus?spuId=` 的后端接口及前端调用，使“关联 SKU”展示只读取 `product_skus` 正式数据，payload 不再驱动 UI。
- [X] T062 [US1] 创建 mini-app 路由骨架（`backend/internal/transport/http/mini-app/router.go`）并在 `registry.go` 注册 `/api/v1/mini-app` 前缀，复用统一的中间件。
- [X] T063 [US1] 实现 mini-app 商品/SKU Handler（`backend/internal/transport/http/mini-app/product_handler.go`），复用现有 service 输出裁剪后的 DTO。
- [X] T064 [US1] 更新 `plugin.yaml`、contracts 与任务文档，补充 mini-app API 的 OpenAPI 描述及基础测试。

**Checkpoint**: 完成后可演示“创建 50 个 SKU 并写入列表”全流程。

---

## Phase 4: User Story 2 - 批量调整价格与库存 (Priority: P2)

**Goal**: 运营/定价可批量修改价格、库存或起订量，含审批策略、预览与任务日志。
**Independent Test**: 选中 ≥50 个 SKU，提交 +5% 价格任务；审批触发、任务完成后能在列表看到更新并查询日志。

### Implementation for User Story 2

- [X] T024 [US2] 增补批量任务/关联表字段（`backend/internal/entity/models/product_sku/bulk_task.go`、`repository.go`）并覆盖索引
- [X] T025 [US2] 实现批量调整服务含审批判定（`backend/internal/services/admin/product_sku/bulk_adjustment_service.go`）根据操作类型/阈值写入 `approval_required`
- [X] T026 [US2] 编排后台任务执行与重试机制（`backend/internal/services/admin/product_sku/bulk_executor.go`）记录审计/错误
- [X] T027 [US2] 实现 `/products/skus/bulk-tasks` 提交与状态查询 Handler（`backend/internal/transport/http/admin/product_sku/bulk_tasks_handler.go`）
- [X] T028 [US2] 支持 Excel/CSV 导入解析与字段校验（`backend/internal/services/admin/product_sku/import_export.go`）及错误报告
- [X] T029 [US2] 新增导出服务/任务（`backend/internal/services/admin/product_sku/import_export.go`）按筛选条件生成文件
- [X] T030 [US2] 暴露 `/products/skus/export` Handler（`backend/internal/transport/http/admin/product_sku/import_handler.go`）并接入异步任务
- [X] T031 [P] [US2] 搭建前端批量操作选择器与预览（`web-admin/app/components/product/sku/BulkAdjustModal.vue`）呈现新旧值
- [X] T032 [P] [US2] 实现导入模板上传/错误下载 UI（`web-admin/app/components/product/sku/BulkImportUploader.vue`）
- [X] T033 [US2] 添加导出入口与下载反馈（`web-admin/app/pages/product/skus/index.vue`、`web-admin/app/components/product/sku/ToolbarActions.vue`）
- [X] T034 [US2] 在 SKU 列表/矩阵页面接入批量操作入口与任务轮询（`web-admin/app/pages/product/skus/index.vue`、`MatrixGrid.vue`）
- [X] T035 [US2] 开发审批面板/通知（`web-admin/app/components/product/sku/ApprovalDrawer.vue`）供审批人查看阈值、通过或驳回
- [X] T036 [US2] 构建任务日志/失败重试界面（`web-admin/app/pages/product/skus/tasks.vue`）读取 `/bulk-tasks/{id}`

**Checkpoint**: 批量任务、导入/导出、审批与日志流程均可独立演示。

---

## Phase 5: User Story 3 - 渠道映射与库存可视 (Priority: P3)

**Goal**: 渠道运营配置渠道 SKU、上架状态与同步策略，实时查看库存与失败日志，并可管理条码/序列号。
**Independent Test**: 针对单 SKU 配置渠道映射并发布；若远端报错，详情页可展示失败记录，库存页签显示最新库存时间且可录入序列号/打印条码。

### Implementation for User Story 3

- [X] T037 [US3] 扩展渠道映射仓储与模型（`backend/internal/entity/models/product_sku/channel_mapping.go`、`repository.go`）记录状态/错误
- [X] T038 [US3] 实现渠道映射服务逻辑（`backend/internal/services/admin/product_sku/channel_service.go`）含幂等任务 ID 与错误回推
- [X] T039 [US3] 提供 `/products/skus/{id}/channels` 与 `/publish` Handler（`backend/internal/transport/http/admin/product_sku/channel_handler.go`）
- [X] T040 [US3] 对接库存服务订阅/回放接口（`backend/internal/services/admin/product_sku/inventory_sync.go`）并维护 `last_synced_at`
- [X] T041 [US3] 暴露 `/products/skus/{id}/inventory` 查询 Handler（`backend/internal/transport/http/admin/product_sku/inventory_handler.go`）含 SLA 超时提醒
- [X] T042 [P] [US3] 构建前端渠道映射页签组件（`web-admin/app/components/product/sku/ChannelMappingTab.vue`）支持新增/编辑/状态展示
- [X] T043 [US3] 实现渠道发布按钮与任务日志展示（`web-admin/app/pages/product/skus/[id].vue`）含错误提醒、重试
- [X] T044 [US3] 构建库存与预警视图（`web-admin/app/pages/product/inventory.vue`）显示可售/锁定/在途与更新时间，并对超 SLA 情况高亮
- [X] T045 [US3] 实现条码/编码唯一性校验与标签生成服务（`backend/internal/services/admin/product_sku/barcode_service.go`）及审计记录
- [X] T046 [US3] 暴露条码生成/打印 API（`backend/internal/transport/http/admin/product_sku/barcode_handler.go`）提供批量校验
- [X] T047 [P] [US3] 构建条码管理面板（`web-admin/app/components/product/sku/BarcodePanel.vue`）调用新 API 并输出打印模板
- [X] T048 [US3] 实现序列号/批次管理 API（`backend/internal/transport/http/admin/product_sku/serial_handler.go`）覆盖创建、查询、审计
- [X] T049 [P] [US3] 在 SKU 详情新增序列号/批次管理组件（`web-admin/app/components/product/sku/SerialBatchSection.vue`）支持录入/查询

**Checkpoint**: 渠道映射/发布、库存可视、条码与序列号管理均可单独演示。

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 文档、可观测性、性能与交付收尾。

- [X] T050 [P] 补充审计/日志字段与结构化打点（`backend/internal/observability/product_sku/logger.go`）
- [X] T051 完善 quickstart/README（`specs/002-product-sku-management/quickstart.md`、`docs/plan/product/README.md`）同步新流程
- [X] T052 [P] 添加端到端冒烟脚本（`web-admin/tests/e2e/product-sku.spec.ts`）覆盖三大故事 happy path（含导出、渠道发布）
- [X] T053 核对 `plugin.yaml`、菜单与 RBAC 权限（`plugin.yaml`、`backend/etc/rbac.yaml`）确保上线前一致
- [X] T054 [P] 验证 `make test`, `make test-admin`, `make integration-smoke` 并整理发布 checklist（`reports/sku-release.md`）
- [X] T055 [P] 新增规格维度/取值一等数据表（`backend/internal/entity/models/product_spec/*`、`backend/cmd/database/migrate/migrations/002_product_sku.go`）
- [X] T056 [P] 为 SKU 增加 `spec_signature` 并写入唯一约束（`backend/internal/entity/models/product_sku/sku.go`、`backend/internal/services/admin/product_sku/spec_signature.go`）
- [X] T057 [P] 提供管理端规格编辑 API（`GET/PUT /api/v1/admin/product/spus/{id}/spec-groups`，代码见 `backend/internal/transport/http/admin/product_spec/*`）
- [X] T058 [P] 提供 mini-app 规格选择聚合接口（`GET /api/v1/mini-app/products/{id}/detail`，代码见 `backend/internal/transport/http/miniapp/product/*`）
- [X] T059 [P] 更新合同与调用指南（`specs/002-product-sku-management/contracts/skus.openapi.yaml`、`docs/guides/features/product/miniapp_open_api.md`）

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → 用户故事阶段 → Phase 6；任何用户故事必须在 Phase 2 完成后方可开始。
- 用户故事之间按优先级 P1→P2→P3；若资源允许，可在各自依赖满足后并行推进（各 story 内 [P] 任务可同步执行）。
- 关键依赖链：迁移/仓储 (T004-T006-T012) → 服务骨架 (T007) → 对应 Handler/前端实现；US3 的条码/序列功能依赖 T012。

## Parallel Execution Examples

- Phase 2 中 T005/T006/T010 可与 T007 并行，T011/T012 完成后整个域命名与序列支撑就绪。
- US1 阶段 T017 与 T018 可在 T013~T016 进行时并行，T022/T023 处理边界可独立实现。
- US2 阶段 T031-T033 可并行，T029/T030 完成后统一打通导出功能。
- US3 阶段 T042/T047/T049 可平行：一人负责渠道 UI，一人处理条码，另一人实现序列号界面。

## Implementation Strategy

1. **MVP**：完成 US1（SKU 生成）即可对外 demo，满足“10 分钟生成 50 个 SKU”目标。
2. **迭代**：在 MVP 稳定后开启 US2 批量任务（含审批、导入/导出），再实现 US3 渠道/库存/条码/序列闭环。
3. **验证**：每个用户故事结束后运行相关 API/前端回归，最终在 Phase 6 执行全链路冒烟并更新文档。

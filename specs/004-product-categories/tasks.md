# Tasks: 商品类目与类目模板管理

**Input**: Design documents from `specs/004-product-categories/`  
**Prerequisites**: `specs/004-product-categories/plan.md`, `specs/004-product-categories/spec.md`, `specs/004-product-categories/data-model.md`, `specs/004-product-categories/contracts/openapi.yaml`, `specs/004-product-categories/research.md`, `specs/004-product-categories/quickstart.md`

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件/目录，且不依赖未完成任务）
- **[Story]**: 仅用户故事阶段必须包含：`[US1]` / `[US2]` / `[US3]`
- 每条任务描述必须包含明确文件路径

---

## Phase 1: Setup（共享准备）

- [X] T001 对齐设计文档引用与目录：确认 `specs/004-product-categories/` 下 `spec.md`/`plan.md`/`research.md`/`data-model.md`/`quickstart.md`/`contracts/openapi.yaml` 均存在且一致
- [X] T002 校验本分支可执行命令：按 `specs/004-product-categories/quickstart.md` 逐条运行并记录结果（仅记录，不修 unrelated failures）

执行记录（2026-01-01）：

- 后端迁移：`make migrate`（首次因 Go 缓存权限失败，改用 workspace 缓存并以提升权限重跑后成功）
- 后端单测：`make test`（需写入 Go toolchain 缓存，提升权限后通过）
- 后端 lint：`make lint` 通过
- 前端构建：`make build-admin` 通过（仅有 Nuxt chunk/circular import 警告）
- 长驻命令未执行：`make dev`、`cd web-admin && npm run dev`（需手动在本地终端跑以便交互调试）

---

## Phase 2: Foundational（阻塞前置）

**⚠️ CRITICAL**：未完成本阶段前，不开始任何 US1/US2/US3 的业务实现。

- [X] T003 新增表名常量：在 `backend/internal/entity/models/model.go` 增加 `TableProductCategories`、`TableProductCategoryLocales`、`TableProductCategoryTemplates`、`TableProductCategoryTemplateFields`、`TableProductCategoryTemplateVersions`、`TableProductCategoryMappings`、`TableProductCategoryPermissions`
- [X] T004 新增迁移注册：创建 `backend/cmd/database/migrate/migrations/004_product_categories.go` 并声明 `ProductCategoryTables`（包含本功能所有持久化模型）
- [X] T005 将迁移纳入业务表集合：在 `backend/cmd/database/migrate/migrate.go` 的 `businessTables` 追加 `migrations.ProductCategoryTables`
- [X] T006 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category.go`（含 ProductCategory + TableName）
- [X] T007 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_locale.go`
- [X] T008 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_template.go`
- [X] T009 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_template_field.go`
- [X] T010 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_template_version.go`
- [X] T011 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_mapping.go`
- [X] T012 [P] 新增领域模型骨架：创建 `backend/internal/entity/models/product_category/category_permission.go`
- [X] T013 [P] 新增仓储层骨架：创建 `backend/internal/entity/repository/product_category/category_repository.go`（内嵌 `*repository.BaseRepository[productcategory.ProductCategory]` 并提供 `NewCategoryRepository`）
- [X] T014 [P] 新增仓储层骨架：创建 `backend/internal/entity/repository/product_category/template_repository.go`
- [X] T015 [P] 新增仓储层骨架：创建 `backend/internal/entity/repository/product_category/mapping_repository.go`
- [X] T016 [P] 新增管理端 service 骨架：创建 `backend/internal/services/admin/product_category/service.go`（聚合 deps、校验、BeginTenantTx/WithTenantTx 使用范式）
- [X] T017 [P] 新增 HTTP 路由骨架：创建 `backend/internal/transport/http/admin/product_category/routes.go`（仅注册路由，不写具体 handler 逻辑）
- [X] T018 [P] 新增 HTTP handler 骨架：创建 `backend/internal/transport/http/admin/product_category/handler.go`（统一错误响应风格）
- [X] T019 [P] 新增 miniapp 路由骨架：创建 `backend/internal/transport/http/miniapp/category/routes.go`（注册 `GET /mini-app/categories/tree`）
- [X] T020 [P] 新增 miniapp handler 骨架：创建 `backend/internal/transport/http/miniapp/category/handler.go`
- [X] T021 将管理端类目路由接入总路由：在 `backend/internal/transport/http/admin/routes.go` 接入 `admin/product_category/routes.go`
- [X] T022 将 miniapp 类目路由接入：在 `backend/internal/transport/http/miniapp/router.go` 增加 category 路由注册
- [X] T023 RBAC 权限声明：在 `backend/internal/transport/http/admin/product/rbac.go` 增加类目/模板/映射/import 权限资源与动作（对应 `spec.md` FR-014）

**Checkpoint**：能完成迁移启动（`make migrate`）且后端能编译通过（`go test ./...` 不要求全绿，但 category 相关包必须通过）。

---

## Phase 3: User Story 1（P1）维护类目树并用于前台展示 🎯 MVP

**Goal**：后台可维护类目树（CRUD/启停/排序/迁移），miniapp 可获取“可展示类目树”（停用类目及其子树不返回）。

**Independent Test**：在后端插入一棵类目树后，调用 `GET /mini-app/categories/tree` 仅返回启用子树；管理端 API 可创建/编辑/迁移且不会形成环。

- [X] T024 [US1] 定义类目状态与基础字段规则：在 `backend/internal/entity/models/product_category/category.go` 补齐字段与约束（停用子树不返回、唯一性约束字段）
- [X] T025 [US1] 实现类目仓储：在 `backend/internal/entity/repository/product_category/category_repository.go` 增加 tree 查询、按条件查询、创建/更新、迁移校验辅助方法
- [X] T026 [US1] 实现类目服务：在 `backend/internal/services/admin/product_category/category_service.go` 增加 Create/Update/List/Tree/Move/SetStatus/DeleteGuard 逻辑（DeleteGuard 对应 FR-006）
- [X] T027 [US1] 实现管理端类目 handler：在 `backend/internal/transport/http/admin/product_category/category_handler.go` 实现 `GET /admin/product/categories/tree`、`GET /admin/product/categories`、`POST /admin/product/categories`、`PATCH /admin/product/categories/{id}`、`POST /admin/product/categories/{id}/move`、`PATCH /admin/product/categories/{id}/status`
- [X] T028 [US1] 实现 miniapp 可展示类目树：在 `backend/internal/transport/http/miniapp/category/handler.go` 实现 `GET /mini-app/categories/tree`（明确：停用类目及其子树不返回）
- [X] T029 [US1] 类目删除保护：在 `backend/internal/services/admin/product_category/category_service.go` 实现“存在子类目或被商品引用则禁止删除”（引用校验可先基于 `product_spus.category_id/category_path`）
- [X] T030 [US1] 管理端类目页面落地：替换占位实现 `web-admin/app/pages/product/categories.vue`（树 + 搜索/过滤 + 详情编辑 + 启停 + 拖拽排序/迁移可先做基础版本）
- [X] T031 [US1] miniapp 商品列表按类目筛选：在 `backend/internal/transport/http/miniapp/product/handler.go` 支持 `categoryId` 或 `categoryPathPrefix` 查询参数（与现有 `ListProducts` 兼容）

**Checkpoint**：US1 前后端联调通过（后台能维护类目树，miniapp 能拿到可展示树并据此筛选商品）。

---

## Phase 4: User Story 2（P2）配置类目模板并驱动商品录入校验

**Goal**：模板可管理与发布/回滚；商品录入选择类目后可拿到“生效模板”（最近祖先已发布模板；子类目绑定模板则覆盖）并在新建/编辑时校验；发布后提供影响范围预览+批量重检（不自动改存量商品状态）。

**Independent Test**：为父类目发布模板后，子类目未绑定模板时会继承；子类目绑定模板后覆盖；商品录入提交缺失必填字段会得到可定位字段错误。

- [X] T032 [US2] 完善模板实体与版本实体字段：在 `backend/internal/entity/models/product_category/category_template.go` 与 `backend/internal/entity/models/product_category/category_template_version.go` 补齐模板/版本字段与状态
- [X] T033 [US2] 完善模板字段实体：在 `backend/internal/entity/models/product_category/category_template_field.go` 补齐字段定义（required/type/rules/default/i18n）
- [X] T034 [US2] 实现模板仓储：在 `backend/internal/entity/repository/product_category/template_repository.go` 增加模板草稿读写、发布版本写入、历史版本查询
- [X] T035 [US2] 实现模板服务：在 `backend/internal/services/admin/product_category/template_service.go` 实现 CRUD、publish、rollback、getEffectiveTemplateByCategory（继承/覆盖规则按 spec Clarifications）
- [X] T036 [US2] 管理端模板 handler：在 `backend/internal/transport/http/admin/product_category/template_handler.go` 实现 `GET/POST/PATCH /admin/product/category-templates` 与 `POST /admin/product/category-templates/{id}/publish`、`POST /admin/product/category-templates/{id}/rollback`
- [X] T037 [US2] 模板预览/模拟填报：在 `backend/internal/transport/http/admin/product_category/template_preview_handler.go` 增加预览/模拟校验 endpoint（对应 FR-011）
- [X] T038 [US2] 影响范围预览：在 `backend/internal/services/admin/product_category/template_impact_service.go` 提供“受影响类目/商品数量”查询（对应 FR-011A）
- [X] T039 [US2] 批量重检任务入口：在 `backend/internal/transport/http/admin/product_category/template_impact_handler.go` 提供“触发批量重检” endpoint（仅触发与记录，不自动改存量商品状态）
- [X] T040 [US2] 商品录入校验接入：在 `backend/internal/services/admin/product/spu/service.go` 增加“按类目取生效模板并校验请求字段”的调用（仅新建/编辑校验，符合 Clarifications）
- [X] T041 [US2] 管理端模板页面：实现 `web-admin/app/pages/product/category-templates/index.vue`（列表）与 `web-admin/app/pages/product/category-templates/[id].vue`（编辑/发布/回滚/预览）
- [X] T042 [US2] SPU 录入类目选择联动模板：更新 `web-admin/app/pages/product/spus/create.vue`（或对应 SPU 编辑页）在选类目后加载模板字段并渲染校验提示

**Checkpoint**：US2 后台模板发布/回滚可用；SPU 新建/编辑按模板校验；影响范围预览与批量重检触发可用。

---

## Phase 5: User Story 3（P3）管理渠道类目映射与批量处理（CSV）

**Goal**：类目映射可维护（增删改查）、可批量 CSV 导入导出、可审计追踪变更。

**Independent Test**：对某类目创建映射后可查询与编辑；导入重复映射会被拒绝并返回可定位错误；导出得到 CSV。

- [X] T043 [US3] 完善映射实体字段：在 `backend/internal/entity/models/product_category/category_mapping.go` 补齐字段（channel/platform_category_id/strategy/sync_status/metadata）
- [X] T044 [US3] 实现映射仓储：在 `backend/internal/entity/repository/product_category/mapping_repository.go` 增加按类目查询、upsert、删除、唯一性冲突检测
- [X] T045 [US3] 实现映射服务：在 `backend/internal/services/admin/product_category/mapping_service.go` 实现 CRUD 与冲突错误返回（可定位到行/字段）
- [X] T046 [US3] 管理端映射 handler：在 `backend/internal/transport/http/admin/product_category/mapping_handler.go` 实现 `GET/POST /admin/product/categories/{id}/mappings`
- [X] T047 [US3] CSV 导入导出（映射）：在 `backend/internal/services/admin/product_category/mapping_import_export.go` 实现 CSV 解析/生成与错误报告
- [X] T048 [US3] 管理端导入导出 endpoint：在 `backend/internal/transport/http/admin/product_category/import_export_handler.go` 实现 `POST /admin/product/categories/import`、`POST /admin/product/categories/export`（按 Clarifications：CSV）
- [X] T049 [US3] 审计查询 endpoint：在 `backend/internal/transport/http/admin/product_category/audit_handler.go` 实现 `GET /admin/product/categories/{id}/audit`
- [X] T050 [US3] 管理端 UI：在 `web-admin/app/pages/product/categories.vue` 增加“渠道映射”tab（读写映射、导入导出、审计列表入口）

**Checkpoint**：US3 映射 CRUD/导入导出/审计可用且可联调。

---

## Phase 6: Polish & Cross-Cutting

- [X] T051 [P] 文档更新：在 `docs/plan/product/categories.md` 追加“已实现范围/未实现范围”与接口对齐说明（保持 PRD 与落地一致）
- [X] T052 统一 RBAC manifest 输出：检查并修正 `backend/internal/transport/http/admin/product/rbac.go` 与插件 RBAC 暴露是否覆盖类目资源（与 FR-014 对齐）
- [X] T053 运行后端质量门禁：执行 `make lint` 与 `make test`（记录结果，入口：`Makefile` / `make-files/`）
- [X] T054 运行前端质量门禁：执行 `make build-admin`（记录结果，入口：`make-files/test.mk`）
- [ ] T055 端到端冒烟：执行 `make dev` 后用最小数据走通 US1 的“创建类目→前台树查询→按类目筛选商品”（入口：`make-files/dev.mk`）

执行记录（2026-01-02）：
- `make lint` ✅
- `make test` ✅
- `make build-admin` ✅（存在 Nuxt circular-chunk 警告）
- `make dev` ✅（可完成迁移并启动 HTTP；完整 E2E 需准备租户/鉴权与最小数据）

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → Phase 3（US1 MVP）为最小可交付链路。
- Phase 4（US2）依赖 Phase 2（模型/仓储骨架）与 Phase 3（类目基础能力，尤其模板绑定点）。
- Phase 5（US3）依赖 Phase 2（模型/仓储骨架）与 Phase 3（类目存在）。

## Parallel Opportunities

- Phase 2 中标注 `[P]` 的模型/仓储/路由骨架可并行推进（不同文件）。
- Phase 4 与 Phase 5 在 Phase 2 完成后可由不同同学并行推进（模板 vs 映射），但都需要复用 Phase 3 的类目基础能力。

## Implementation Strategy（建议）

1. 先交付 US1（让 miniapp 可展示/筛选）作为 MVP。
2. 再交付 US2（模板+校验）提升录入质量与一致性。
3. 最后交付 US3（渠道映射+CSV）支撑多渠道上架与批量化运营。

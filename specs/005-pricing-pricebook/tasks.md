# Tasks: 定价中心—价目表（Pricebook）与基础查价

**Input**: Design documents from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/005-pricing-pricebook/`  
**Prerequisites**: `plan.md`、`spec.md`（用户故事）、`research.md`、`data-model.md`、`contracts/`、`quickstart.md`  
**Contracts**: `specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml`

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件/无前置依赖）
- **[Story]**: 用户故事标签（`[US1]`、`[US2]`、`[US3]`）
- 任务描述必须包含明确文件路径

---

## Phase 1: Setup（共享准备）

- [x] T001 对齐合同与命名：核对 `specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml` 与 `specs/005-pricing-pricebook/spec.md`（字段/路径/状态机/回退语义一致）
- [x] T002 [P] 固化路由矩阵与中间件一致性：新增 `specs/005-pricing-pricebook/contracts/routing-matrix.md`（说明 `/v1/pricing/query` 主路径、`/api/v1/pricing/query` 兼容别名，以及两者必须使用相同的 JWT/RBAC/tenant 中间件链）
- [x] T003 [P] 建立定价域目录骨架：创建 `backend/internal/entity/models/pricing/`、`backend/internal/entity/repository/pricing/`、`backend/internal/services/pricing/`、`backend/internal/transport/http/admin/pricing/`、`backend/internal/transport/http/pricing/`

---

## Phase 2: Foundational（阻塞性基础设施）

**⚠️ 说明**：本阶段完成前不要开始任何用户故事的业务功能实现。

- [ ] T004 新增表名常量：在 `backend/internal/entity/models/model.go` 增加 `TablePricebooks`、`TablePricebookVersions`、`TablePricebookScopes`、`TablePricebookItems`、`TablePricebookAuditLogs`
- [ ] T005 [P] 定义 GORM 模型：新增 `backend/internal/entity/models/pricing/pricebook.go`
- [ ] T006 [P] 定义 GORM 模型：新增 `backend/internal/entity/models/pricing/pricebook_version.go`
- [ ] T007 [P] 定义 GORM 模型：新增 `backend/internal/entity/models/pricing/pricebook_scope.go`
- [ ] T008 [P] 定义 GORM 模型：新增 `backend/internal/entity/models/pricing/pricebook_item.go`
- [ ] T009 [P] 定义 GORM 模型：新增 `backend/internal/entity/models/pricing/pricebook_audit_log.go`
- [ ] T010 注册迁移表集合：新增 `backend/cmd/database/migrate/migrations/005_pricing_pricebook.go`（导出 `PricingPricebookTables`）
- [ ] T011 将迁移纳入业务表清单：更新 `backend/cmd/database/migrate/migrate.go` 将 `migrations.PricingPricebookTables` 追加到 `businessTables`
- [ ] T012 [P] 新增定价域 RLS 策略：新增 `backend/cmd/database/migrate/rls_pricing.go` 并实现 `ensurePricingRLSPolicies(ctx, db)`（覆盖 `pricebooks/pricebook_versions/pricebook_scopes/pricebook_items/pricebook_audit_logs`）
- [ ] T013 在迁移流程中启用定价域 RLS：更新 `backend/cmd/database/migrate/migrate.go` 在 `MigratePluginModels` 流程末尾调用 `ensurePricingRLSPolicies(ctx, db)`
- [ ] T014 [P] 定义 repository：新增 `backend/internal/entity/repository/pricing/pricebook_repository.go`（内嵌 `*repository.BaseRepository[T]` + `NewXXXRepository`）
- [ ] T015 [P] 定义 repository：新增 `backend/internal/entity/repository/pricing/pricebook_version_repository.go`
- [ ] T016 [P] 定义 repository：新增 `backend/internal/entity/repository/pricing/pricebook_scope_repository.go`
- [ ] T017 [P] 定义 repository：新增 `backend/internal/entity/repository/pricing/pricebook_item_repository.go`
- [ ] T018 [P] 定义 repository：新增 `backend/internal/entity/repository/pricing/pricebook_audit_log_repository.go`
- [ ] T019 定义定价服务骨架：新增 `backend/internal/services/pricing/service.go`（注入 repos、tenantFromContext、HealthProbe/Ready 口径）
- [ ] T020 [P] 定义 DTO（管理端）：新增 `backend/internal/transport/http/admin/pricing/dto.go`（Pricebook/Version/Item/Scope 请求响应结构）
- [ ] T021 [P] 定义错误码与错误响应：新增 `backend/internal/transport/http/admin/pricing/errors.go`（与 OpenAPI `ErrorResponse` 对齐）
- [ ] T022 定义 RBACEntries：新增 `backend/internal/transport/http/admin/pricing/rbac.go`（资源 `pricing:pricebook` + actions `read/manage/publish`）
- [ ] T023 注册管理端路由：新增 `backend/internal/transport/http/admin/pricing/routes.go` 并更新 `backend/internal/transport/http/admin/routes.go` 挂载到 `/api/v1/admin/pricing`
- [ ] T024 注册 RBAC 汇总：更新 `backend/internal/transport/http/registry.go` 合并 `adminpricing.RBACEntries(r.apiPrefix())`
- [ ] T025 定义查价路由（可复用注册）：新增 `backend/internal/transport/http/pricing/routes.go`（对接任意 `*gin.RouterGroup`，并在组内注册 `/pricing/query`）
- [ ] T026 在 Router 中挂载 `/v1` 业务组：更新 `backend/internal/router/router.go` 新增 `/v1` group，并使用与 `gApi` 一致的 `RequestTrace + JWTAuth + RBAC` 中间件链
- [ ] T027 将查价主路径注册到 `/v1`：更新 `backend/internal/router/router.go` 在 `/v1` group 上调用 `pricing.RegisterRoutes(...)`（生成 `/v1/pricing/query`）
- [ ] T028 将查价兼容别名注册到 `apiPrefix`：更新 `backend/internal/transport/http/registry.go` 在 `RegisterAPIRoutes(gApi)` 中调用 `pricing.RegisterRoutes(gApi, deps)`（生成 `/api/v1/pricing/query`）
- [ ] T029 [P] 路由一致性冒烟：新增 `backend/internal/router/router_pricing_test.go` 验证 `/v1/pricing/query` 与 `/api/v1/pricing/query` 均存在且使用相同 handler

**Checkpoint**：完成后，代码结构/迁移/RBAC/路由入口齐备，可以进入用户故事开发。

---

## Phase 3: User Story 1 — 创建并发布价目表版本（P1）🎯 MVP

**Goal**：实现“创建价目表→生成草稿版本→发布→生效→下线”的管理端闭环，并满足“发布新版本自动终止旧版本有效期”的澄清。

**Independent Test**：通过管理端接口完成：创建价目表、写入少量条目、发布版本（立即生效），再发布新版本并验证旧版本被自动终止；最后下线版本并验证不再命中。

- [ ] T030 [US1] 实现管理端列表/创建：新增 `backend/internal/transport/http/admin/pricing/pricebooks_handler.go`（GET/POST `/admin/pricing/pricebooks`）
- [ ] T031 [US1] 实现管理端更新：扩展 `backend/internal/transport/http/admin/pricing/pricebooks_handler.go`（PATCH `/admin/pricing/pricebooks/{pricebookId}`）
- [ ] T032 [US1] 实现版本创建：新增 `backend/internal/transport/http/admin/pricing/versions_handler.go`（POST `/admin/pricing/pricebooks/{id}/versions`）
- [ ] T033 [US1] 实现版本发布：扩展 `backend/internal/transport/http/admin/pricing/versions_handler.go`（POST `/publish`）
- [ ] T034 [US1] 实现版本下线：扩展 `backend/internal/transport/http/admin/pricing/versions_handler.go`（POST `/archive`）
- [ ] T035 [US1] 服务层：实现 `backend/internal/services/pricing/pricebook_service.go`（Create/List/Update 价目表 + 创建 v1 draft）
- [ ] T036 [US1] 服务层：实现 `backend/internal/services/pricing/version_service.go`（CreateVersion/PublishVersion/ArchiveVersion，含发布幂等）
- [ ] T037 [P] [US1] 并发发布锁定支持：扩展 `backend/internal/entity/repository/pricing/pricebook_repository.go` 与 `backend/internal/entity/repository/pricing/pricebook_version_repository.go` 增加“按 pricebook 行锁/active 版本锁”的查询方法（用于 Publish 事务）
- [ ] T038 [US1] 发布幂等与并发策略：在 `backend/internal/services/pricing/version_service.go` 落实“发布新版本自动终止旧版本有效期”（FR-004A）并使用锁/冲突检测保证同一时刻最多一个 active 版本
- [ ] T039 [US1] 审计：实现 `backend/internal/services/pricing/audit_service.go`（记录 create/update/publish/archive 摘要）
- [ ] T040 [US1] 校验：在 `backend/internal/services/pricing/version_service.go` 增加有效期校验（expires_at > effective_at；不合法返回错误码）
- [ ] T041 [US1] 最小单测：新增 `backend/internal/services/pricing/version_service_test.go` 覆盖“发布新版本终止旧版本有效期”与“并发 publish 不产生双 active”的确定性行为

**Checkpoint**：US1 可独立演示与验收（不依赖 US2/US3 完整实现，但需最小条目能力见 US2/T045）。

---

## Phase 4: User Story 2 — 维护适用范围与价格条目（P2）

**Goal**：在草稿版本中维护 scope 与 SKU 价格条目，并保证“发布前不影响对外查价”。

**Independent Test**：创建草稿版本→配置 scope→批量 upsert 10 个 SKU 条目→发布→再查价（US3）可命中；同时验证非 draft 版本写入被拒绝。

- [ ] T042 [US2] Scope DTO 与解析：完善 `backend/internal/transport/http/admin/pricing/dto.go`（channel_ids/customer_group_ids/supplier_ids）
- [ ] T043 [US2] 管理端 scope 更新：扩展 `backend/internal/transport/http/admin/pricing/pricebooks_handler.go` 支持 scopes 写入
- [ ] T044 [US2] 服务层 scope upsert：新增 `backend/internal/services/pricing/scope_service.go`（无记录=全量适用；有记录=必须命中）
- [ ] T045 [US2] 管理端条目 upsert：新增 `backend/internal/transport/http/admin/pricing/items_handler.go`（PUT `/versions/{versionId}/items`）
- [ ] T046 [US2] 服务层条目 upsert：新增 `backend/internal/services/pricing/item_service.go`（仅 draft 可写；unique(tenant,version,sku) 覆盖更新）
- [ ] T047 [US2] 条目校验：在 `backend/internal/services/pricing/item_service.go` 增加金额字段一致性校验（负数/最小大于最大等返回错误码）
- [ ] T048 [US2] 字段优先级定义：在 `backend/internal/services/pricing/item_service.go` 或 `pricing` 公共包内固化“成交价字段优先级”为 `sale > base > msrp`，并在查价返回中输出 `source_field`（对齐 spec FR-007A）
- [ ] T049 [US2] 最小单测：新增 `backend/internal/services/pricing/item_service_test.go` 覆盖“非 draft 禁止写入”与关键校验

**Checkpoint**：US2 完成后，价目表数据可完整配置，为 US3 查价提供稳定输入。

---

## Phase 5: User Story 3 — 统一基础查价能力（P3）

**Goal**：实现 `/v1/pricing/query` 的确定性查价（并保留 `/api/v1/pricing/query` 兼容别名）：范围匹配、版本有效期、具体度优先、多候选选择规则、回退策略（C）、无价 vs 错误语义（C）。

**Independent Test**：用同一租户构造：Base pricebook + scoped pricebook；验证命中 scoped；当 scoped 命中但缺条目时回退 Base；当范围不命中/币种不匹配时返回无价（带原因码）；参数缺失返回错误（带错误码）。

- [ ] T050 [US3] 查价 DTO：新增 `backend/internal/transport/http/pricing/dto.go`（PricingQueryRequest/Response、reason codes、source_field）
- [ ] T051 [US3] 查价 handler：新增 `backend/internal/transport/http/pricing/query_handler.go`（POST `/pricing/query`，参数校验与错误语义分流）
- [ ] T052 [US3] 查价 service：新增 `backend/internal/services/pricing/query_service.go`（实现选择规则与回退/无价语义）
- [ ] T053 [US3] 候选筛选（scope）：在 `backend/internal/services/pricing/query_service.go` 实现“无 scope=全量适用；有 scope=必须命中”的维度判断
- [ ] T054 [US3] 版本有效期判断：在 `backend/internal/services/pricing/query_service.go` 实现 `effective_at <= as_of < expires_at`（expires_at 为空视为永久）
- [ ] T055 [US3] 多候选选择规则：在 `backend/internal/services/pricing/query_service.go` 实现“具体度优先 → 显式优先级（如实现）→ 最近发布优先”
- [ ] T056 [US3] 回退策略（C）：在 `backend/internal/services/pricing/query_service.go` 仅在“已命中版本但缺条目”时回退 Base；其余不回退
- [ ] T057 [US3] 无价 vs 错误语义（C）：在 `backend/internal/services/pricing/query_service.go` 与 `backend/internal/transport/http/pricing/query_handler.go` 落实“无价原因码/错误码”的返回口径，并在有价时返回 `source_field`
- [ ] T058 [US3] Repository 查询优化：在 `backend/internal/entity/repository/pricing/price_query_repository.go` 封装查价所需查询（候选版本/条目命中），并确保按 tenant_uuid 过滤
- [ ] T059 [US3] 最小单测：新增 `backend/internal/services/pricing/query_service_test.go` 覆盖“多候选选择/回退/无价原因码/错误语义”

**Checkpoint**：US3 完成后，可对接订单/前台/渠道同步的统一查价能力。

---

## Phase 6: Polish & Cross-Cutting（收尾与质量）

- [ ] T060 [P] 文档同步：更新 `docs/plan/pricing/pricebooks.md`（如接口/字段有偏差，保持 PRD 与合同一致）
- [ ] T061 完成 quickstart 验证：按 `specs/005-pricing-pricebook/quickstart.md` 跑通 migrate + dev + 查价冒烟
- [ ] T062 代码格式化：运行 `gofmt`（触及文件）并修复格式问题
- [ ] T063 基础测试：运行 `go test ./...`（至少覆盖新增 pricing 域单测）
- [ ] T064 [P] OpenAPI 校验：对齐 `specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml` 与最终实现（必要时补充字段/响应示例）

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 → Phase 2（基础设施齐备）
- Phase 2 完成后：
  - **US1 可以启动**（但 US1 发布验收需要最小条目能力，建议并行推进 US2/T045-T048）
  - **US2 可以启动**
  - **US3 可以启动**（但强依赖 US2 的数据写入能力）
- Phase 6 依赖 US1/US2/US3 中你选择交付的范围完成

### User Story Dependencies

- **US1 (P1)**：依赖 Phase 2；发布逻辑建议在 US2 条目 upsert 打通后做端到端验收
- **US2 (P2)**：依赖 Phase 2；为 US3 提供数据输入
- **US3 (P3)**：依赖 Phase 2 + US2；回退/无价语义按 Clarifications 固化

---

## Parallel Opportunities

### Foundational 可并行任务

- T005–T009（模型文件）、T014–T018（仓储文件）、T020–T022（DTO/错误/RBAC）可由多人并行完成（注意避免同文件冲突）

### US1/US2 并行建议

- 管理端 handler（T030–T034、T045）与服务层实现（T035–T040、T044、T046–T048）可并行推进，但需在联调前统一 DTO/错误码（T020–T021）

### US3 并行建议

- T050（DTO）与 T058（查询 repository）可并行；T051（handler）依赖 DTO；T052–T057（service）依赖 repository 与 Clarifications

---

## Suggested MVP Scope

最小可交付建议：**US1 + US2（最小条目能力）+ US3（基础查价）**  
原因：仅有 US1（发布）但无条目写入与查价能力，无法形成可验证闭环；US2+US3 使“配置→发布→查价”可验收。

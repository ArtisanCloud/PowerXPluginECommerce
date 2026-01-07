# Implementation Plan: 商品类目与类目模板管理

**Branch**: `004-product-categories` | **Date**: 2026-01-01 | **Spec**: `specs/004-product-categories/spec.md`  
**Input**: Feature specification from `specs/004-product-categories/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

实现商品类目树（用于前台展示）、类目模板（驱动商品录入校验）与渠道类目映射（含 CSV 批量导入导出），并满足已澄清的关键业务规则：停用类目及其子树不出现在可展示类目树、模板继承为“最近祖先已发布模板”、模板发布仅影响新建/编辑校验并提供影响范围预览+批量重检、类目唯一性（code 全局唯一/同父 displayName 唯一/alias 全局唯一）、导入导出格式为 CSV。

## Technical Context

**Language/Version**: Go 1.24（backend），Node 20 + TypeScript 5.9 + Nuxt 4（web-admin）  
**Primary Dependencies**: Gin、GORM（postgres driver）、PowerX plugin framework；Nuxt 4、@nuxt/ui 3.3.x、Pinia、Nuxt i18n  
**Storage**: PostgreSQL（schema: `powerx_plugin_base`）  
**Testing**: `go test ./...`（`make test`），golangci-lint（`make lint`）；web-admin `npm run build` / `make build-admin`  
**Target Platform**: Linux server（插件后端）+ Node server bundle（Nuxt Nitro node-server）  
**Project Type**: Web application（backend + web-admin）  
**Performance Goals**: 类目树查询在 5,000 节点规模下 P95 ≤ 500ms（见 `spec.md` SC-002）  
**Constraints**: 多租户隔离（Tenant UUID + RLS）；管理端 RBAC；导入导出为 CSV  
**Scale/Scope**: 类目树多层级；模板支持版本化发布/回滚；渠道映射支持批量导入导出与审计

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

结论：本计划不触犯宪章 gate（无未解决澄清点；遵循多租户/RLS、Service-centric、RBAC entries、可观测与测试要求）。

- [x] Host Contract First：所有对外 API 归入 `/v1/**`，管理端位于 `/v1/admin/**`；miniapp 位于 `/v1/mini-app/**`（由宿主反代到 `/_p/<plugin-id>/api/v1`）。
- [x] Tenant Isolation：所有新增持久化模型包含 `tenant_uuid`（UUID 字符串），Repo 使用 `BeginTenantTx/WithTenantTx` 注入租户上下文并依赖 RLS。
- [x] Service-Centric：HTTP handler 仅做校验/鉴权/调用 service；业务编排在 `backend/internal/services/**`；Repo 封装数据访问。
- [x] Observable & Testable：关键操作写审计事件；提供 service 单测与迁移冒烟；lint/build 通过 CI。
- [x] Unified Plugin RBAC：新增域的 `RBACEntries` 统一暴露权限三元组并由中间件处理。

## Project Structure

### Documentation (this feature)

```text
specs/004-product-categories/
├── plan.md                        # This file
├── research.md                    # Phase 0 output
├── data-model.md                  # Phase 1 output
├── quickstart.md                  # Phase 1 output
├── contracts/
│   └── openapi.yaml               # Phase 1 output
└── tasks.md                       # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/database/migrate/migrations/          # 迁移注册（AutoMigrate tables）
├── internal/entity/models/                   # 新增 product_category* 等模型
├── internal/entity/repository/               # 新增 product_category* repo（BaseRepository）
├── internal/services/admin/product_category/ # 管理端类目/模板/映射/导入导出/审计 service
├── internal/transport/http/admin/            # 管理端路由与 handler（薄）
└── internal/transport/http/miniapp/          # miniapp 只读接口（类目树、筛选）

web-admin/
├── app/pages/product/categories.vue          # 类目管理页（替换占位）
├── app/pages/product/category-templates/**   # 模板管理页
└── app/pages/product/spus/**                 # SPU 录入类目选择与模板联动
```

**Structure Decision**: 采用现有插件的“backend + web-admin”双项目结构；后端按宪章的 services/repository/models/transport 分层新增 `product_category` 子域；miniapp 仅暴露读取能力以支持前台展示与筛选。

## Phase 0: Outline & Research (Output: `research.md`)

- 明确并固化关键业务规则（已在澄清阶段完成，写入 `spec.md` Clarifications）。
- 决策“类目树数据表达”方案与查询性能策略（见 `research.md`）。

## Phase 1: Design & Contracts (Outputs: `data-model.md`, `contracts/openapi.yaml`, `quickstart.md`)

- 输出数据模型（实体、关系、唯一性、状态、审计事件点）。
- 输出 API 合同（管理端 + miniapp；不包含实现细节，但可被前端联调与测试引用）。
- 输出 Quickstart（本地迁移、运行、验证命令与最小数据准备方式）。

## Phase 2: Planning Handoff

`/speckit.tasks` 将基于 P1/P2/P3 用户故事拆分任务，并以“可独立交付/可独立验收”为切片原则推进。

## Complexity Tracking

无（未发现需要破例的宪章违背项）。

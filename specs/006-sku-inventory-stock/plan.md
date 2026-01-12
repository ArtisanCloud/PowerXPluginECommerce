# Implementation Plan: 库存（Stock）MVP：SKU 可用库存闭环与上架前置校验

**Branch**: `006-sku-inventory-stock` | **Date**: 2026-01-11 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/006-sku-inventory-stock/spec.md`  
**Input**: Feature specification from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/006-sku-inventory-stock/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

交付“SKU 维度库存可写闭环”以支撑商品上架：在 default 仓为 SKU 提供**增量调整（delta，整数件数）**的可用库存维护能力；web-admin 在 SKU 详情页提供库存快照展示与调整入口；在 **SPU 发布** 与 **渠道上架/同步发布** 两个关键动作上增加库存前置校验（至少存在 1 个 SKU 的 `available_qty > 0`），校验失败仅阻止并返回可解释原因，不自动改商品状态。

## Technical Context

**Language/Version**: Go 1.24 + Node 20（Nuxt 4）  
**Primary Dependencies**: Gin、GORM（postgres driver）、PowerX plugin framework（RBAC/tenant context/STS）；Nuxt 4 + TypeScript 5.9 + Nuxt UI 3.3.x  
**Storage**: PostgreSQL（schema：`powerx_plugin_base`，RLS 强制，所有表带 `tenant_uuid`）  
**Testing**: `go test ./...`；`npm run test:ci`（必要时补后端 service/repo 单测与前端组件测试）  
**Target Platform**: Linux server（插件后端服务）+ Web（web-admin）  
**Project Type**: Web application（`backend/` + `web-admin/`）  
**Performance Goals**: 库存快照查询 P95 ≤ 200ms；库存调整接口 P95 ≤ 200ms（单行 upsert + 审计写入）  
**Constraints**: 租户隔离（tenant_uuid + RLS）；库存调整不可导致 `available_qty` 为负；接口需返回可解释错误；“仅阻止”策略不自动下架  
**Scale/Scope**: 每租户 SKU 数量可达 10^5；库存表需支持分页/按 SKU 查询；MVP 仅覆盖 default 仓与可用库存闭环

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

基于 `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/.specify/memory/constitution.md`：

- Host Contract First：通过（管理端库存接口置于 `/api/v1/admin/**`，不新增不一致路径）  
- Tenant Isolation & Zero Trust：通过（从请求上下文取 tenant_uuid；数据层依赖 RLS；禁止使用 tenant_id）  
- Service-Centric Architecture：通过（Handler 薄；库存与校验编排落在 Service；Repo 负责数据访问）  
- Observable & Testable Delivery：通过（库存调整写审计；补充单测/冒烟步骤；关键失败可追踪）  
- Unified Plugin RBAC：通过（按 inventory/sku 的资源动作声明 RBACEntries，并通过 `/api/v1/admin/rbac` 输出）

## Project Structure

### Documentation (this feature)

```text
specs/006-sku-inventory-stock/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
```text
backend/
├── cmd/
│   ├── plugin/                      # 入口
│   └── database/migrate/migrations/ # 迁移注册（AutoMigrate）
├── internal/
│   ├── entity/
│   │   ├── models/                  # gorm models（表名常量在 models/model.go）
│   │   └── repository/              # BaseRepository[T] 封装的数据访问
│   ├── services/                    # 业务编排（SKU 库存/发布校验）
│   └── transport/http/              # gin handlers & routes（admin/public/miniapp）
└── etc/                             # runtime config

web-admin/
├── app/composables/api/              # apiGet/apiPost 封装（Nuxt ruleset）
└── app/pages/product/skus/           # SKU 详情页库存入口（MVP）
```

**Structure Decision**: 沿用既有商品 SKU 子域实现库存闭环：`internal/entity/models/product_sku`（现有 `ProductSKUInventory`）+ `internal/entity/repository/product_sku`（补库存 upsert/adjust）+ `internal/services/admin/product_sku`（补 AdjustInventory + 校验与审计事件）+ `internal/transport/http/admin/product_sku`（补写接口路由）；前端在 SKU 详情页新增库存卡片并使用统一 Nuxt API client composable。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

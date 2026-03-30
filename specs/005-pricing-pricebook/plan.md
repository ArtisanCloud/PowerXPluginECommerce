# Implementation Plan: 定价中心—价目表（Pricebook）与基础查价

**Branch**: `005-pricing-pricebook` | **Date**: 2026-01-07 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugins.ecommerce/specs/005-pricing-pricebook/spec.md`  
**Input**: Feature specification from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugins.ecommerce/specs/005-pricing-pricebook/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

交付“价目表主数据 + 版本发布 + 统一查价”最小闭环：管理端可创建价目表与版本、维护 SKU 价格条目并发布；对外通过查价接口按上下文（SKU/币种/范围维度/时间点）返回确定性、可解释的基础价结果，并遵循已澄清的回退与失败语义（见 spec Clarifications）。对外业务 API 需满足宪章的 `/v1/**` 合同（必要时提供 `/api/v1/**` 兼容别名以避免存量联调断裂）。

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: Gin、GORM（postgres driver）、PowerX plugin framework（RBAC/tenant context/STS）  
**Storage**: PostgreSQL（schema：`powerx_plugin_base`，RLS 强制，所有表带 `tenant_uuid`）  
**Testing**: `go test ./...`（必要时补 service/repository 单测 + 迁移冒烟）  
**Target Platform**: Linux server（插件后端服务）  
**Project Type**: Web application（`backend/` + `web-admin/`，本期以 `backend/` 为主）  
**Performance Goals**: 查价接口 P95 ≤ 200ms（命中索引，允许后续引入缓存）  
**Constraints**: `/v1/**` 业务合同（管理端 `/api/v1/admin/**`）、租户隔离（tenant_uuid + RLS）、确定性选价（可解释）、发布幂等与并发安全  
**Scale/Scope**: 支持“每租户”多价目表与多版本；条目规模预期可到 10^5 SKU 量级（需索引与分页）

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

基于 `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugins.ecommerce/.specify/memory/constitution.md`：

- Host Contract First：通过（API 前缀使用 `/api/v1`，管理端路由置于 `/api/v1/admin/**`）  
- Tenant Isolation & Zero Trust：通过（全表 `tenant_uuid` + RLS；请求验签与 tenant context 复用现有中间件）  
- Service-Centric Architecture：通过（Handler 薄；业务编排在 `internal/services/**`；Repo 封装数据访问）  
- Observable & Testable Delivery：通过（关键发布/下线/查价可审计；补充单测与必要日志/指标钩子）  
- Unified Plugin RBAC：通过（新增 pricing 域 RBACEntries，合入 `/api/v1/admin/rbac` 输出）

## Project Structure

### Documentation (this feature)

```text
specs/005-pricing-pricebook/
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
│   ├── services/                    # 业务编排（定价域建议 internal/services/pricing）
│   └── transport/http/              # gin handlers & routes（admin/public/miniapp）
└── etc/                             # runtime config

web-admin/
└── app/pages/pricing/               # UI（本期不作为交付前置，但接口需对齐）
```

**Structure Decision**: 本期为后端定价域新增 `pricing` 子域，沿用既有分层：`internal/entity/models/pricing`、`internal/entity/repository/pricing`、`internal/services/pricing`、`internal/transport/http/admin/pricing`；公共查价以 `/v1/pricing/query` 为主（必要时保留 `/api/v1/pricing/query` 兼容别名）。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

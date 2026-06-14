# Implementation Plan: 促销规则引擎（订单自动促销）

**Branch**: `015-pricing-promotions-engine` | **Date**: 2026-06-10 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/015-pricing-promotions-engine/spec.md`
**Input**: Feature specification from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/015-pricing-promotions-engine/spec.md`

## Summary

本特性在现有 pricebook 基础定价之后、coupon 用户券之前新增订单自动促销层，交付促销规则配置、订单结算自动匹配、促销与优惠券叠加/互斥判断、订单促销快照与审计追踪。MVP 聚焦订单级满减与折扣，不扩展到完整营销活动、预算审批、渠道投放或买赠/组合包。

## Technical Context

**Language/Version**: Go 1.24（backend）, TypeScript 5.9 + Nuxt 4（web-admin, Node 20）  
**Primary Dependencies**: Gin, GORM (Postgres), PowerX plugin framework, Nuxt UI 3.3.x, Pinia  
**Storage**: PostgreSQL (`powerx_plugin_base` schema), no required Redis dependency for MVP  
**Testing**: `go test ./...`, targeted backend service/handler tests, admin FE lint/build (`npm run lint`, `npm run build`)  
**Target Platform**: Linux plugin service + web-admin browser clients  
**Project Type**: Plugin monorepo（backend + web-admin）  
**Performance Goals**: 95% 订单结算请求在 500ms 内返回促销明细；促销命中准确率 99.9%  
**Constraints**: 多租户隔离（tenant_uuid + RLS）；金额以服务端订单提交时刻为准；促销结束时间含边界；历史订单快照不可被后续促销变更覆盖  
**Scale/Scope**: 一期覆盖订单级 `amount_off`/`percent_off`，全场/SKU/渠道范围，金额门槛，优先级、互斥组、促销与优惠券叠加策略

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Host Contract First：管理端接口挂载 `/api/v1/admin/promotions`，交易试算遵循 `/v1/promotions/quote` 或接入既有订单 quote 链路；前端通过插件反代前缀与 runtimeConfig 适配。  
- [x] Tenant Isolation & Zero Trust：新增促销规则、审计、订单快照实体均包含 `tenant_uuid`；仓储读写经租户事务/RLS；禁止引入 `tenant_id`。  
- [x] Service-Centric：促销规则管理、quote 计算、快照写入与审计集中在 `internal/services/admin/promotion`；Handler 只做校验、鉴权入口与序列化。  
- [x] Observable & Testable：设计包含促销命中/拒绝原因、审计日志、金额分摊和订单快照；要求服务单测、处理器契约测试、订单集成回归。  
- [x] Minimal Footprint：MVP 不引入新基础设施；复用现有 Gin/GORM、pricing、coupon、order 目录模式与 Nuxt admin 页面模式。  
- [x] Unified Plugin RBAC：新增 RBAC entry 由 promotion admin transport 暴露并合并到 `/api/v1/admin/rbac`/manifest。

结论：当前设计无宪章阻塞项，可进入 Phase 0/1 设计。

## Project Structure

### Documentation (this feature)

```text
specs/015-pricing-promotions-engine/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── promotions.openapi.yaml
│   └── routing-matrix.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/database/migrate/
├── internal/
│   ├── entity/
│   │   ├── models/
│   │   │   └── promotion/
│   │   └── repository/
│   │       └── promotion/
│   ├── observability/
│   │   └── promotion/
│   ├── services/
│   │   ├── admin/
│   │   │   ├── order/
│   │   │   └── promotion/
│   │   └── admin/coupon/
│   └── transport/http/
│       ├── admin/
│       │   └── promotion/
│       └── miniapp/
│           └── promotion/
└── tests/

web-admin/
├── app/
│   ├── composables/api/
│   │   └── usePromotions.ts
│   ├── pages/pricing/
│   │   └── promotions.vue
│   ├── stores/
│   └── types/
└── tests/
```

**Structure Decision**: 采用现有插件分层，不新增跨层目录。促销模型与仓储放在 `internal/entity/{models,repository}/promotion`，管理端规则编排和交易 quote 计算放 `internal/services/admin/promotion`，订单创建接入点在 `internal/services/admin/order` 调用促销服务并写快照；管理端接口落在 `internal/transport/http/admin/promotion`，可选交易调试接口落在 `internal/transport/http/miniapp/promotion`。前端主入口为 `web-admin/app/pages/pricing/promotions.vue`，`market/promotions.vue` 仅作为后续营销聚合入口。

## Phase 0 — Research Output

详见 [research.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/015-pricing-promotions-engine/research.md)。

## Phase 1 — Design Output

- 数据模型：详见 [data-model.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/015-pricing-promotions-engine/data-model.md)
- 接口契约：详见 `contracts/promotions.openapi.yaml` 与 `contracts/routing-matrix.md`
- 验证路径：详见 [quickstart.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/015-pricing-promotions-engine/quickstart.md)

## Constitution Check (Post-Design)

- [x] 路由仍遵循宿主反代约定（管理端 `/api/v1/admin/*`，交易端 `/v1/*` 或订单统一 quote 入口）
- [x] 数据模型全量租户隔离，并可映射到 RLS 与 `BeginTenantTx`
- [x] 促销计算、快照、审计在 Service 层集中，Handler 保持薄层
- [x] 契约已定义金额、叠加、拒绝原因、快照与 RBAC 语义，便于测试和可观测落地
- [x] MVP 未引入预算、审批、渠道同步、买赠等额外复杂度

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

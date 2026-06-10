# Implementation Plan: 完整优惠券机制（订单复算与支付核销）

**Branch**: `014-coupon-full-mechanism` | **Date**: 2026-04-03 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/014-coupon-full-mechanism/spec.md`
**Input**: Feature specification from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/014-coupon-full-mechanism/spec.md`

## Summary

本特性在现有 pricebook 基础定价之上新增优惠券权益层，建立“订单提交复算 → 券资产预占 → 支付成功核销 → 失败释放 → 退款按策略返券”的完整闭环。设计重点是交易一致性（金额与状态）、并发安全（同券不可重复占用）、事件幂等（重复回调不重复动作）与可追溯审计（订单快照 + 券流水）。

## Technical Context

**Language/Version**: Go 1.24（backend）, TypeScript + Nuxt 4（web-admin）  
**Primary Dependencies**: Gin, GORM (Postgres), PowerX plugin framework, Nuxt UI 3.3.x, Pinia  
**Storage**: PostgreSQL (`powerx_plugin_base` schema), optional Redis for idempotency/locks  
**Testing**: `go test ./...`, admin FE lint/test (`npm run lint`, `npm run test`)  
**Target Platform**: Linux server + web-admin browser clients  
**Project Type**: Plugin monorepo（backend + web-admin）  
**Performance Goals**: 券结算请求 P95 <= 500ms；支付核销链路最终一致延迟 <= 60s  
**Constraints**: 多租户隔离（tenant_uuid + RLS）；支付回调幂等；订单关闭自动释放  
**Scale/Scope**: 目标 10^6 级订单规模下可稳定处理券生命周期，支持运营查询与审计回溯

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Host Contract First：对外交易路径使用 `/v1/**`，管理端使用 `/api/v1/admin/**`，遵循现有插件反代合同。  
- [x] Tenant Isolation & Zero Trust：新实体全量 `tenant_uuid`，事务中注入租户上下文，禁止 `tenant_id` 回流。  
- [x] Service-Centric：优惠规则、占用/核销编排落在 `internal/services`，Handler 保持薄层。  
- [x] Observable & Testable：要求新增审计事件、核心指标与幂等/并发测试覆盖。  
- [x] Minimal Footprint：不引入与目标不匹配的新基础设施，优先复用现有订单/支付链路与仓储模式。

结论：当前设计无宪章阻塞项，可进入 Phase 0/1 设计。

## Project Structure

### Documentation (this feature)

```text
specs/014-coupon-full-mechanism/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── coupons.openapi.yaml
│   └── routing-matrix.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── entity/
│   │   ├── models/
│   │   └── repository/
│   ├── services/
│   │   ├── admin/order/
│   │   ├── admin/payment/
│   │   └── pricing/
│   └── transport/http/
│       ├── admin/
│       └── miniapp/
│           └── coupon/
└── cmd/database/migrate/

web-admin/
├── app/pages/
├── app/composables/api/
├── app/stores/
└── tests/
```

**Structure Decision**: 采用现有插件分层，不新增跨层目录。优惠券领域模型与仓储放 `internal/entity/{models,repository}`，交易编排放 `internal/services/admin/order` 与支付/退款联动服务中；交易侧 `/v1/coupons/quote` 处理器统一落在 `internal/transport/http/miniapp/coupon`，管理侧接口落在 `internal/transport/http/admin/coupon`；前端在 `web-admin` 复用现有 API composable 与页面模式。

## Phase 0 — Research Output

详见 [research.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/014-coupon-full-mechanism/research.md)。

## Phase 1 — Design Output

- 数据模型：详见 [data-model.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/014-coupon-full-mechanism/data-model.md)
- 接口契约：详见 `contracts/coupons.openapi.yaml` 与 `contracts/routing-matrix.md`
- 验证路径：详见 [quickstart.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/014-coupon-full-mechanism/quickstart.md)

## Constitution Check (Post-Design)

- [x] 路由仍遵循宿主反代约定（管理端 `/api/v1/admin/*`，交易端 `/v1/*`）
- [x] 数据模型全量租户隔离并可映射到 RLS
- [x] 规则与状态机在 Service 层集中，不在 Handler 内散落
- [x] 契约已定义幂等与错误语义，便于测试和可观测落地

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

# Implementation Plan: 自营下单（Mini-app 下单 + Admin 代客下单）MVP

**Branch**: `007-order-mini-app` | **Date**: 2026-01-12 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/007-order-mini-app/spec.md`  
**Input**: Feature specification from `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/007-order-mini-app/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

在 `one_time（一次性商品）` 范围内交付最小可用下单闭环：小程序与后台代客下单统一复用“可售校验”口径，创建订单时执行服务端二次校验并在同一租户事务内完成**整单原子**的库存锁定（写入 `product_sku_inventories.locked_qty`），同时记录订单事件审计；后台提供订单查询与取消（仅 `pending_payment` 可取消）能力；创建订单支持幂等键（同一幂等键重复提交返回同一订单结果且不得重复锁库存）；订单生成并对外展示租户内唯一的人可读订单号；订单价格以“提交时服务端计算的当前价”为准并保存价格/金额快照。

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24 + Node 20（Nuxt 4）  
**Primary Dependencies**: Gin、GORM（postgres driver）、PowerX plugin framework（RBAC/tenant context/STS）；Nuxt 4 + TypeScript 5.9 + Nuxt UI 3.3.x  
**Storage**: PostgreSQL（schema：`powerx_plugin_base`，RLS 强制，所有表带 `tenant_uuid`）  
**Testing**: `go test ./...`（优先补 service/repository 单测；必要时补 router 冒烟）  
**Target Platform**: Linux server（插件后端服务）+ Web（web-admin）  
**Project Type**: Web application（`backend/` + `web-admin/`；本期以 `backend/` 为主，web-admin 仅订单管理页面与代客下单入口）  
**Performance Goals**: 创建订单接口 P95 ≤ 2s（含可售与库存校验、写入订单/明细/审计）；订单列表/详情查询 P95 ≤ 200ms（命中索引）  
**Constraints**: 多租户隔离（tenant_uuid + RLS）；创建订单幂等（防重放/重试）；整单原子库存锁定（不留残余锁）；取消仅后台且仅 `pending_payment`；订单号租户内唯一；返回可解释错误  
**Scale/Scope**: 每租户订单量可达 10^6（需分页与索引）；并发下单需避免超卖；MVP 不含支付、履约、优惠与售后

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

基于 `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/.specify/memory/constitution.md`：

- Host Contract First：通过（管理端 `/api/v1/admin/**`；小程序 `/api/v1/mini-app/**`；遵循宿主反代前缀，不新增宿主私有耦合）  
- Tenant Isolation & Zero Trust：通过（中间件注入 tenant_uuid；Repo 通过 `WithTenantTx/BeginTenantTx` 执行并依赖 RLS；鉴权使用既有 JWT/HMAC 与 customer token 中间件）  
- Service-Centric Architecture：通过（Handler 薄；下单编排在 Service；Repo 封装订单写入与库存行锁）  
- Observable & Testable Delivery：通过（订单事件审计 + 关键失败结构化日志；补 service/repo 单测与迁移冒烟）  
- Unified Plugin RBAC：通过（新增 order 域 RBACEntries：create/read/cancel；合入 `/api/v1/admin/rbac` 输出）

## Project Structure

### Documentation (this feature)

```text
specs/007-order-mini-app/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
backend/
├── cmd/
│   ├── plugin/                      # 入口
│   └── database/migrate/migrations/ # 迁移注册（AutoMigrate）
├── internal/
│   ├── entity/
│   │   ├── models/
│   │   │   └── order/               # 订单域模型（orders/order_items/order_events）
│   │   └── repository/
│   │       └── order/               # 订单域仓储（写入、查询、行锁）
│   ├── services/
│   │   ├── admin/order/             # 后台：代客下单、查询、取消
│   │   └── miniapp/order/           # 小程序：下单、查询
│   └── transport/http/
│       ├── admin/order/             # /api/v1/admin/orders...
│       └── miniapp/order/           # /api/v1/mini-app/orders...
└── etc/                             # runtime config

web-admin/
└── app/pages/orders/                # 订单列表/详情/取消（MVP）
```

**Structure Decision**: 新增 `order` 子域并严格沿用分层：`models/order` + `repository/order` + `services/{admin,miniapp}/order` + `transport/http/{admin,miniapp}/order`。库存锁定复用既有 `product_sku` 库存行（`product_sku_inventories.locked_qty`），可售校验复用既有 miniapp sellability 规则。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

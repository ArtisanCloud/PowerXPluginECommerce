# Implementation Plan: 订阅对账与续费治理

**Branch**: `013-subscription-reconciliation` | **Date**: 2026-03-31 | **Spec**: /specs/013-subscription-reconciliation/spec.md
**Input**: Feature specification from `/specs/013-subscription-reconciliation/spec.md`

## Summary

实现“日对账 + 差异处置 + 续费失败治理”的最小业务闭环：按净应收口径生成对账结果、输出标准化差异并转任务，执行递增重试（1h/24h/72h/7d）与分层 SLA（高 24h / 中 48h / 低 72h），并提供运营看板与审计导出。

## Technical Context

**Language/Version**: Backend Go 1.24；Frontend TypeScript 4.x + Nuxt 4  
**Primary Dependencies**: Gin, GORM(Postgres), PowerX plugin framework, taskcenter；Nuxt UI 3.3.x + Pinia  
**Storage**: PostgreSQL（schema `powerx_plugin_base`），可选 Redis 仅用于任务去重缓存  
**Testing**: `make test`、`make test-admin-ci`、`go test ./internal/services/admin/subscription_reconciliation -run Test`  
**Target Platform**: PowerX Plugin Backend + web-admin  
**Project Type**: multi（backend + web-admin）  
**Performance Goals**: 日对账任务 10k 账单在 30 分钟内完成；运营查询接口 p95 < 1s（50 条分页）  
**Constraints**: 严格多租户隔离、对账任务幂等、口径统一（净应收）、错误码标准化、可审计追溯  
**Scale/Scope**: 单租户日账单 10k~100k，差异率常态 <5%，人工处置任务并发 100+

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Host Contract First: PASS  
  - 仅规划 `/api/v1/admin/**` 与 `/api/v1/agent/**` 受控接口，遵循插件反代合同。
- Tenant Isolation & Zero Trust: PASS  
  - 新增实体全部带 `tenant_uuid`，Repo 走 tenant tx，错误映射不泄漏跨租户信息。
- Service-Centric Architecture: PASS  
  - Handler 仅做参数/鉴权/序列化；对账、差异分类、续费治理编排全部在 Service。
- Observable & Testable Delivery: PASS  
  - 要求结构化观测事件、审计轨迹、单测与回归命令可复现。
- Minimal Footprint & Versioned Releases: PASS  
  - 在现有 membership/payments/taskcenter 基础上扩展，不引入新基础设施。

## Project Structure

### Documentation (this feature)

```text
specs/013-subscription-reconciliation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── openapi.yaml
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── internal/entity/models/subscription_reconciliation/
├── internal/entity/repository/subscription_reconciliation/
├── internal/services/admin/subscription_reconciliation/
├── internal/transport/http/admin/subscription_reconciliation/
├── internal/observability/subscription_reconciliation/
└── internal/contracts/

web-admin/
└── app/pages/finance/subscription-reconciliation.vue
```

**Structure Decision**: 采用插件既有多层架构，后端新增独立子域 `subscription_reconciliation`，前端新增运营看板页，避免侵入订单/支付核心路径。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

## Post-Design Constitution Check (Phase 1)

- Host Contract First: PASS
  - `contracts/openapi.yaml` 仅使用 `/api/v1/admin/subscription-reconciliation/**` 路由并保持插件边界内。
- Tenant Isolation & Zero Trust: PASS
  - 数据模型中所有核心实体均包含 `tenant_uuid`，并明确按租户幂等与去重键。
- Service-Centric Architecture: PASS
  - 对账、差异分流、续费治理的业务编排均归属 service 层，符合 handler 轻量化约束。
- Observable & Testable Delivery: PASS
  - quickstart 给出可执行回归命令，research/data-model 明确审计事件与验证口径。
- Minimal Footprint & Versioned Releases: PASS
  - 方案复用现有 taskcenter/支付链路，不引入新基础设施，仅增量扩展子域。

# Implementation Plan: 订阅会籍权益代币发放

**Branch**: `010-membership-entitlements` | **Date**: 2026-01-30 | **Spec**: /specs/010-membership-entitlements/spec.md
**Input**: Feature specification from `/specs/010-membership-entitlements/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

订阅支付成功后自动生成会籍、发放权益与代币余额，支持组合型权益、默认可叠加、幂等防重复，并提供客户查询视图。

## Technical Context

**Language/Version**: Go 1.24; TypeScript 5.9 + Nuxt 4 (web-admin) + Uni-app (mini-app)  
**Primary Dependencies**: Gin, GORM (Postgres), PowerX plugin framework; Nuxt UI 3.3.x  
**Storage**: PostgreSQL (`powerx_plugin_base` schema)  
**Testing**: `go test ./...`, `make test`, `npm run test` (admin)  
**Target Platform**: PowerX 插件后端 + web-admin + mini-app  
**Project Type**: multi (backend + web-admin + mini-app)  
**Performance Goals**: 查询权益/余额接口 p95 ≤ 1s（50 条以内）  
**Constraints**: 幂等处理、RLS 多租户隔离、回调可重放  
**Scale/Scope**: 订阅订单 10k+/日规模

## Constitution Check

- Host Contract First：遵循 `/api/v1` 路由与插件反代前缀。
- Tenant Isolation & Zero Trust：所有新模型携带 `tenant_uuid`，repo 使用 `BeginTenantTx`。
- Service-Centric Architecture：Handler 薄、业务逻辑下沉 `internal/services`。
- Observable & Testable：回调日志含 request_id/tenant_uuid，补齐基础测试。
- Minimal Footprint：只引入必要的模型/接口与配置。

## Project Structure

### Documentation (this feature)

```text
specs/010-membership-entitlements/
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
├── internal/
│   ├── entity/models/
│   ├── entity/repository/
│   ├── services/
│   └── transport/http/
web-admin/
└── app/
mini-app/
└── src/
```

**Structure Decision**: 该功能涉及后端支付回调、会籍/权益/代币数据模型与服务，以及 web-admin/mini-app 查询展示。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | N/A | N/A |

---

## Phase 0: Outline & Research

### Research Tasks

- 决定订阅权益与代币的最小落地实体与字段范围。
- 明确幂等键优先级在系统内的统一使用位置。
- 明确权益叠加策略与有效期计算规则。

### Output

- `research.md`

---

## Phase 1: Design & Contracts

### Data Model Design

- 会员：`membership_tiers`、`membership_assignments`
- 权益：`membership_benefits`（支持 bundle）与 `entitlements`
- 代币：`token_accounts`、`token_transactions`
- 关联字段：`source_type/source_id`、`transaction_id/out_trade_no/order_id`

### Contracts

- 订阅发放（内部触发）
- 客户权益查询
- 客户代币余额查询

### Output

- `data-model.md`
- `/contracts/*`
- `quickstart.md`

---

## Phase 1: Update Agent Context

- 运行 `.specify/scripts/bash/update-agent-context.sh codex`

---

## Phase 2: Planning

- 产出任务拆解与时间线（由 `/speckit.tasks` 生成）

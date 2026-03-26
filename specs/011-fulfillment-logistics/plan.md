# Implementation Plan: 履约与物流全模块闭环

**Branch**: `011-fulfillment-logistics` | **Date**: 2026-03-25 | **Spec**: [/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/011-fulfillment-logistics/spec.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/011-fulfillment-logistics/spec.md)
**Input**: Feature specification from `/specs/011-fulfillment-logistics/spec.md`

## Summary

围绕“承运商、运费模板、运单与轨迹、履约任务、逆向物流、第三方适配”构建电商履约闭环，采用 M1/M2/M3 分阶段交付策略：
- M1：正向履约闭环（承运商+模板+运单+轨迹）
- M2：履约任务与异常升级
- M3：第三方承运商适配与逆向增强
- M4：履约增强（部分发货与多包裹、波次批量执行、履约成本对账）

技术路径采用现有 PowerXPlugin 分层架构（HTTP Handler → Service → Repository），保持多租户 RLS、RBAC、审计与事件一致性，避免引入平行机制。

## Technical Context

**Language/Version**: Go 1.24（backend）、TypeScript 4.x + Nuxt 4（web-admin）  
**Primary Dependencies**: Gin、GORM(Postgres)、PowerX plugin framework、Nuxt UI 3.3.x、Pinia  
**Storage**: PostgreSQL（schema `powerx_plugin_base`），可选 Redis 用于异步任务/事件缓存  
**Testing**: `go test ./...`（后端单测/集成）、`npm run lint -- --max-warnings=0` + `npm run build`（前端质量与构建校验）  
**Target Platform**: Linux server + Node 20（插件 backend + web-admin）  
**Project Type**: Web application（backend + web-admin）  
**Performance Goals**: 轨迹状态更新 1 分钟内可见；95% 订单可完成运单创建并进入可追踪状态  
**Constraints**: 多租户隔离（tenant_uuid + RLS）；幂等键为“事件ID+运单号”；异常 24h 自动升级；保持宿主/standalone 语义一致  
**Scale/Scope**: 面向单插件多租户运营场景，覆盖履约主链路 + 仓内任务 + 逆向 + 三方适配 + 履约增强对账

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

- Host Contract First: PASS（统一管理端口径，沿用 `/api/v1/admin/**`）
- Tenant Isolation & Zero Trust: PASS（统一 tenant_uuid；RLS + JWT/HMAC；无 tenant_id 新增）
- Service-Centric Architecture: PASS（Handler 薄化，Service 编排，Repo 封装）
- Observable & Testable Delivery: PASS（审计、事件、日志、构建/测试门禁）
- Minimal Footprint & Versioned Releases: PASS（复用现有栈，不新增平行系统）
- Unified Plugin RBAC: PASS（RBAC 通过 manifest/rbac 输出，不在 Handler 硬编码）

### Post-Design Gate

- 设计产物（research/data-model/contracts/quickstart）均遵循上述原则，无新增违例。

## Project Structure

### Documentation (this feature)

```text
specs/011-fulfillment-logistics/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── fulfillment.openapi.yaml
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/
├── internal/
│   ├── entity/
│   │   ├── models/
│   │   └── repository/
│   ├── services/
│   ├── transport/http/
│   └── observability/
└── tests/

web-admin/
├── app/pages/
├── app/components/
├── app/composables/
└── app/types/
```

**Structure Decision**: 采用现有 backend + web-admin 双层结构，不新增子工程；按域（logistics/fulfillment/reverse）落在既有目录分层中实现。

## Phase 0: Research

1. 决策履约模块的阶段边界（M1/M2/M3）与每阶段验收口径。  
2. 决策运单状态源优先级（承运商回调优先，人工补录）的一致性策略。  
3. 决策幂等联合键（事件ID + 运单号）的全链路落点。  
4. 决策异常升级 SLA（24h）与提醒触发方式。  
5. 决策第三方适配边界（统一能力接口 + 可替换配置）。

## Phase 1: Design & Contracts

1. 产出实体模型：Carrier、CarrierService、RateTemplate、Waybill、TrackingEvent、FulfillmentTask、ReverseWaybill。  
2. 产出 OpenAPI 合同：承运商、模板、运单、轨迹、任务、逆向、Webhook。  
3. 产出 quickstart：按 M1→M2→M3 验证关键流程。  
4. 更新 agent context（`update-agent-context.sh codex`）。

## Phase 2: Implementation Planning Input

1. 按 M1/M2/M3 将任务拆分为可独立验收的 stories。  
2. 明确每个 story 的测试入口（API、页面、事件、异常场景）。  
3. 生成后续 `/speckit.tasks` 所需输入（按优先级与依赖排序）。

## Phase 3: Iteration-2 Planning (M4)

1. 设计一单多包裹与部分发货模型扩展（订单聚合状态与包裹明细关联）。  
2. 设计波次（Wave）实体、批量推进接口与部分失败回执协议。  
3. 设计履约成本字段与承运商账单聚合/导出接口。  
4. 规划 M4 独立验收路径并补充 quickstart。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |

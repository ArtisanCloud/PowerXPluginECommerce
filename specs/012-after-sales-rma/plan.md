# Implementation Plan: 售后 RMA 与退货门户

**Branch**: `012-after-sales-rma` | **Date**: 2026-03-31 | **Spec**: [/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/012-after-sales-rma/spec.md](/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/012-after-sales-rma/spec.md)
**Input**: Feature specification from `/specs/012-after-sales-rma/spec.md`

## Summary

围绕“客户发起售后 + 运营审核流转 + 订单/支付/履约最小联动”交付 RMA MVP：
- 用户侧：提交仅退款/退货退款/换货申请，查看进度与时间线。
- 管理侧：受理、审核通过/拒绝、完结/关闭，形成审计轨迹。
- 联动侧：同步订单售后标记、退货逆向物流关联、退款防重。

本期遵循“主模块 `reverse-logistics` + 最小 `returns-portal` 入口”策略，不扩展自动补发与复杂结算。

## Technical Context

**Language/Version**: Go 1.24（backend）、TypeScript 4.x + Nuxt 4（web-admin）  
**Primary Dependencies**: Gin、GORM(Postgres)、PowerX plugin framework、Nuxt UI 3.3.x、Pinia  
**Storage**: PostgreSQL（schema `powerx_plugin_base`）  
**Testing**: `go test ./...`、`make lint`、`make test-admin-ci`  
**Target Platform**: Linux server + Node 20（插件 backend + web-admin）
**Project Type**: Web application（backend + web-admin）  
**Performance Goals**: 95% 售后申请在提交后 5 秒内可查询；运营标准处理路径平均操作时长 ≤ 3 分钟/单  
**Constraints**: 多租户隔离（tenant_uuid + RLS）；状态机禁止非法流转；同一订单明细进行中售后防重；换货首版不触发自动补发/库存自动占用  
**Scale/Scope**: 面向单插件多租户运营场景，覆盖售后申请、审核、逆向关联与最小订单支付履约联动

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Gate

- Host Contract First: PASS（统一管理端与 mini-app 路由在 `/api/v1/**`，不引入平行协议）
- Tenant Isolation & Zero Trust: PASS（全链路 tenant_uuid；不新增 tenant_id；鉴权沿用中间件）
- Service-Centric Architecture: PASS（Handler 薄化，业务编排在 Service，Repo 负责持久化）
- Observable & Testable Delivery: PASS（状态变更审计 + 关键链路测试门禁）
- Minimal Footprint & Versioned Releases: PASS（复用现有订单/支付/履约能力，不新增独立子系统）
- Unified Plugin RBAC: PASS（通过现有 RBACEntries 机制暴露资源动作，不在 Handler 硬编码）

### Post-Design Gate

- 设计产物（research/data-model/contracts/quickstart）均保持与宪章一致，无新增违例。

## Project Structure

### Documentation (this feature)

```text
specs/012-after-sales-rma/
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
├── internal/
│   ├── entity/models/
│   ├── entity/repository/
│   ├── services/
│   ├── transport/http/
│   └── observability/
└── cmd/database/migrate/

web-admin/
├── app/pages/
├── app/components/
├── app/composables/
└── app/types/
```

**Structure Decision**: 采用既有 backend + web-admin 双层结构，不新增子工程。售后能力以 `after_sales/reverse` 领域落在现有分层目录中实现。

## Phase 0: Research

1. 固化售后时间窗口分层规则（仅退款/退货退款/换货）。
2. 固化重复申请策略（进行中禁止、终态可再申请）。
3. 决策售后状态机与终态冻结规则。
4. 决策换货首版边界（不自动补发、不自动库存占用）。
5. 决策订单/支付/履约联动最小集合与防重策略。

## Phase 1: Design & Contracts

1. 产出数据模型：AfterSaleCase、AfterSaleTimeline、AfterSaleEvidence、AfterSaleDecision、ReturnLogisticsLink。  
2. 产出 OpenAPI 合同：mini-app 申请与查询、admin 审核与流转、逆向关联、统计看板。  
3. 产出 quickstart：按“客户申请 → 运营审核 → 联动回写”最小闭环验证。  
4. 更新 agent context（`update-agent-context.sh codex`）。

## Phase 2: Implementation Planning Input

1. 按 P1/P2 拆分 stories：客户入口、运营流转、联动与审计。  
2. 为每个 story 指定独立测试入口（HTTP API、页面流、状态机边界）。  
3. 生成 `/speckit.tasks` 所需任务骨架（模型、仓储、服务、接口、前端、测试、迁移、RBAC）。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |

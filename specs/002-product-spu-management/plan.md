# Implementation Plan: 商品（SPU）管理

**Branch**: `002-product-spu-management` | **Date**: 2025-12-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-product-spu-management/spec.md`

**Note**: 本计划遵循 `.specify/templates/commands/plan.md` 所定义的流程。

## Summary

为商品中心交付端到端的 SPU 管理体验：商品运营可在 web-admin 中通过创建向导、版本管理与渠道配置完成草稿→审核→发布的闭环，同时后台 Go 服务提供 SPU/版本/渠道/订阅计划/批量任务等 API，具备多租户审计、批量导入导出与渠道推送钩子。实现方式包括：扩展 `backend` 的产品域模型、仓储、Service、HTTP handlers 及渠道/任务中心集成，并在 `web-admin` 中新增 `product/spus` 列表、编辑页、创建向导、批量工具与审批面板，确保 KPI（上线≤2 天、导入成功率≥99%、渠道同步≥98%、审计 100% 覆盖）可量化验证。

## Technical Context

**Language/Version**: Go 1.24（backend）、TypeScript 5.9 + Nuxt 4（web-admin）  
**Primary Dependencies**: Gin HTTP、GORM（Postgres driver）、PowerX plugin framework、JWT/STS、多租户任务中心、Nuxt UI 3.3.x、Pinia、@vueuse/core、Nuxt i18n  
**Storage**: PostgreSQL (`powerx_plugin_base` schema) — 表：`product_spus`、`product_spu_versions`、`product_spu_channels`、`product_spu_subscription_plans`、`product_spu_audit_logs`、任务记录表等  
**Testing**: `make test`（Go 单元+集成）、`make integration-smoke`（渠道/任务中心）、`npm run test` + `npm run lint -- --max-warnings=0`（Nuxt + Vitest + Playwright）  
**Target Platform**: 插件后端服务（宿主 `/_p/<plugin-id>/api/v1`）与 web-admin（宿主 `/_p/<plugin-id>/admin/*` 反代）  
**Project Type**: Monorepo（backend Go service + Nuxt admin 前端）  
**Performance Goals**: 列表筛选/分页首屏 < 2s；详情加载 & 版本 diff < 1.5s；批量导入 10k 行校验在 5 分钟内完成并生成报告；渠道推送任务在 1 小时内反馈状态  
**Constraints**: 必须遵循多租户零信任（tenant_uuid、STS、RLS）、所有审批/批量动作写入 `admin_console_audit_events`；SPU 发布触发异步渠道任务需可幂等；订阅计划影响范围需记录审批；前端遵循 Nuxt UI 组件规范  
**Scale/Scope**: 管理 ≥ 20k SPU、每个 SPU 可关联 50+ SKU、订阅计划 ≤ 10 个/商品；导入批量一次 ≤ 10k 行；审批链 3 层并行 30+ 审批人

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **Host Contract First**：所有 API 均暴露在 `/api/v1/products/spus/**` 与 `/api/v1/jobs/**`，前端通过 `runtimeConfig.public.apiBaseUrl` 访问，渠道任务遵循宿主契约。
- ✅ **Tenant Isolation & Zero Trust**：所有表添加 `tenant_uuid` 并在 Repo `BeginTenantTx` 下执行；API 验证 JWT/STS；导出/导入与渠道推送均写入审计与任务中心。
- ✅ **Service-Centric Architecture**：控制器只做鉴权 + DTO 校验；业务编排集中在 `internal/services/admin/product/spu`；Repo 内嵌 `BaseRepository`；HTTP/gRPC（若延伸）共用 Service。
- ✅ **Observable & Testable Delivery**：新增审计 emitter、结构化日志、任务状态指标；后端 `make test` 覆盖 service/repo，前端提供列表/导入/审批 E2E；导入/渠道任务暴露指标。
- ✅ **Minimal Footprint & Versioned Releases**：沿现有 Go + Nuxt 栈，无新语言；输出文档/合同/quickstart，准备 `make release` 打包。

## Project Structure

### Documentation (this feature)

```text
specs/002-product-spu-management/
├── plan.md          # /speckit.plan 产物
├── research.md      # Phase 0 研究结论
├── data-model.md    # Phase 1 数据模型
├── quickstart.md    # Phase 1 环境/验证指南
├── contracts/       # Phase 1 API 契约
└── tasks.md         # 待 /speckit.tasks 生成
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── plugin/main.go
│   └── database/ (migrate/seed)
├── internal/
│   ├── transport/http/admin/product/spu/
│   ├── services/admin/product/spu/
│   ├── entity/models/product/
│   ├── entity/repository/product/
│   ├── jobs/channel/product/
│   └── observability/product/
├── pkg/ (shared infra)
├── etc/ (configs)
└── mocks/tests/

web-admin/
├── app/pages/product/spus/
│   ├── index.vue
│   ├── create.vue
│   ├── edit/[id].vue
│   └── components/{Wizard,ChannelVisibility,VersionDiff}.vue
├── app/components/product/
├── app/composables/useSpu*
├── app/stores/product/spu.ts
├── app/types/product.ts
└── tests/
    ├── unit/product/
    └── e2e/product-spu.spec.ts
```

**Structure Decision**: 本特性同时修改 `backend` 与 `web-admin`，均沿既有 PowerX 插件分层：后端在 `internal/{transport,services,domain}` 建立 `product/spu` 模块；前端在 `app/pages/product/spus/**`、组件、composables、Pinia store 与测试目录内扩展，无需新增项目或额外仓库。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| *(None)* | — | — |

## Offboarding Strategy

1. **流程目标**：为已发布 SPU 提供即时/定时下架能力，保持渠道与审计一致，并能在需要时重新走发布流程。
2. **服务能力**：
   - 在 `backend/internal/services/admin/product/spu/service.go` 新增 `Withdraw` 逻辑，参数包含 `channels[]`（默认全渠道）、`withdrawAt`、`reason`，逐个更新 `product_spu_channels`，当所有渠道下架后将 `product_spus.status` 置为 `offboarded`。
   - 触发渠道任务中心（沿用 publisher worker）并记录任务 ID + 下架理由到 `product_spu_audit_logs`，供时间轴展示。
3. **前端体验**：
   - 在 `web-admin/app/stores/product/spu.ts` 增加 `withdraw` action，与新的 API `POST /api/v1/admin/product/spus/{id}/withdraw` 对接。
   - 在 `spus/edit/[id].vue` 添加“下架”按钮与对话框，支持多渠道勾选+定时下架，操作完成后刷新详情并在 `SpuAuditTimeline` 中显示。
   - 在渠道任务看板增加下架任务状态，以便运营跟踪执行结果。
4. **测试与文档**：在 `tasks.md` 中新增对应任务，要求补充 service 单测 / E2E；在 `spec.md` 的“下架策略”章节描述状态流转，quickstart 更新“重新上架/下架”步骤。

## Deletion Strategy

1. **软删除（Soft Delete）**
   - 适用状态：`draft`、`offboarded`。对于 `published` 或 `reviewing`，必须下架/驳回后才能删除。
   - 实现：在 `product_spus` 上设置 `deleted_at` 并隐藏于列表 API；同时保留所有版本、审批、渠道和审计记录。
   - 审计：`POST /api/v1/admin/product/spus/{id}/delete`（待实现）需记录删除原因、操作者、时间，并发通知给责任人。
2. **硬删除（Hard Delete）**
   - 仅平台管理员通过后台工具触发，前提是 SPU 已软删除 ≥ 30 天且无活跃引用（订单、套餐、营销活动等）。
   - 执行时级联清理 `product_spu_versions`、`product_spu_channels`、`product_spu_subscription_plans`、`product_spu_audit_logs` 等表；写入审计与变更日志。运维流程记录于 `docs/plan/addenda/spu-hard-delete.md`，并提供 `backend/cmd/tools/spu_hard_delete.go` 辅助生成 SQL。
3. **UI/权限**
   - 详情页增加“删除”入口，仅对草稿/下架状态显示，提交删除前需再次确认理由。
   - 权限项复用 `product.spu.manage`，硬删除需额外的后台运维权限。
4. **任务规划**
   - 后端：新增 Delete Service + Handler、软删校验、审计写入、硬删脚本。
   - 前端：新增删除按钮/确认弹窗、Pinia action、成功后的跳转体验。
   - 测试：单测覆盖软删条件，E2E 覆盖删除 UI；文档更新 spec/quickstart 说明.

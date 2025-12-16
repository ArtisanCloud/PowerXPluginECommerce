# Implementation Plan: Channel Master Data & Authorization

**Branch**: `003-channel-master` | **Date**: 2025-12-13 | **Spec**: `specs/003-channel-master/spec.md`
**Input**: Feature specification from `/specs/003-channel-master/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

实现渠道主数据入驻、审批、授权与凭证管理、KPI/健康度可视化、策略配置，以及任务关联/备注与同步历史审计的端到端流程。后端延续 PowerX 插件 Go 栈，新增 channel 域模型（含 ChannelTaskLink、ChannelNote、ChannelSyncHistory）、服务、HTTP/gRPC handler、巡检 job 与审计/告警事件；所有表挂 `tenant_uuid` 并启用 RLS，凭证使用 envelope encryption 储存，并通过 Prometheus/审计指标满足 SLA。前端基于 Nuxt 4 + Nuxt UI 构建渠道列表、详情、授权向导与任务/同步面板，复用 `runtimeConfig.public.apiBaseUrl` 兼容宿主反代，并提供健康度卡片、KPI 趋势、同步历史与告警处理入口。

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Backend Go 1.24；Frontend TypeScript 5.9 + Nuxt 4（Node 20）  
**Primary Dependencies**: Gin、GORM（Postgres driver）、PowerX plugin framework、Redis client、Nuxt UI 3.3.x、Pinia、Nuxt i18n、@vueuse/core  
**Storage**: PostgreSQL（schema `powerx_plugin_base`）存放 channel_* 表；Redis 用于凭证巡检与 KPI 缓存  
**Testing**: `go test ./backend/...`、`make integration-smoke`（connector/授权流程）、`npm run lint -- --max-warnings=0`、`npm run test` + component 测试  
**Target Platform**: Linux 宿主容器 + 宿主反代 `/_p/<plugin-id>/api|admin`  
**Project Type**: Web application（backend + web-admin）  
**Performance Goals**: 渠道详情 KPI 加载 ≤ 3s；凭证告警延迟 ≤ 1h；同步成功率 ≥ 99%；列表筛选 p95 ≤ 1s  
**Constraints**: 多租户 RLS、JWT/ST S 鉴权、凭证 envelope encryption、只允许 `/v1/**` API、日志需带 `tenant_uuid`/`request_id`、Nuxt 遵循 host baseURL、禁止长效凭据；Prometheus/报表需输出审批/凭证覆盖率/同步成功率/加载时延指标以验证 SC-001~SC-005  
**Scale/Scope**: 200~500 渠道/租户，10+ KPI 指标/渠道，告警保留 90 天，凭证巡检每 12h 跑批

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **Host Contract First**：API 均位于 `/v1/channels/**` 并在 `plugin.yaml` manifest 注册；所有外部访问通过 STS 取得短期凭证，禁止硬编码宿主 URL。  
- ✅ **Tenant Isolation & Zero Trust**：`ChannelMaster` 及相关表包含 `tenant_uuid` 并启用 RLS；Handler 校验 JWT 签名并注入 Tenant 背景；凭证密钥仅由 STS/密钥管理器解密。  
- ✅ **Service-Centric Architecture**：新增 Service 层（`internal/services/admin/channel_master`）负责业务与健康度算法，HTTP/gRPC handler 只做鉴权/DTO 映射；Repository 封装读写。  
- ✅ **Observable & Testable Delivery**：实现结构化日志（含 `request_id`、`tenant_uuid`）、Prometheus 指标、审计事件、告警流水；覆盖 service/repo 单测与集成冒烟。  
- ✅ **Minimal Footprint & Versioned Releases**：沿用既有 Go/Nuxt 栈，无额外运行时；发布执行 `make release && make package`，同步 `plugin.yaml` 版本、`web-admin/.output`.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
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
│   ├── plugin/
│   └── database/
├── internal/
│   ├── entity/models/channel_master/
│   ├── entity/repository/channel_master/
│   ├── services/admin/channel_master/
│   ├── transport/http/admin/channel_master/
│   ├── transport/grpc/admin/channel_master/    # 预留
│   ├── jobs/channel/master/                    # 凭证巡检、KPI 刷新调度
│   └── observability/channel/master/
├── pkg/security/encryption/                    # envelope helper（可复用）
└── tests/integration/channel_master/

web-admin/
├── app/
│   ├── pages/channels/
│   │   ├── index.vue
│   │   ├── [id].vue
│   │   └── approval.vue
│   ├── components/channels/
│   ├── stores/channels.ts
│   ├── composables/useChannels.ts
│   └── types/channels.ts
├── server/api/channels/*.ts                    # 若需 Nuxt server 代理
└── tests/
    ├── unit/channels.spec.ts
    └── component/channel-panel.spec.ts
```

**Structure Decision**: 采取标准「backend + web-admin」结构：后端在 `internal/entity/models|repository/channel_master` 存放模型与仓储，服务/transport/job/observability 位于既定层级；前端在 `web-admin/app/pages/channels` 及相关组件、store、composable 中实现 UI，与 tests/unit+component 对应。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | 架构遵循宪章与既有栈 | N/A |

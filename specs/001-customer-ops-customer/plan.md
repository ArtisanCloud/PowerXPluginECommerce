# Implementation Plan: Customer List & Membership Views

**Branch**: `001-customer-ops-customer` | **Date**: 2025-12-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-customer-ops-customer/spec.md`

**Note**: This plan follows `.specify/templates/commands/plan.md` execution flow.

## Summary

升级 web-admin 客户列表与会员视角：支持多维筛选、批量标签/负责人/发券操作、任务中心反馈、详情抽屉信息整合，并强化会员视角的筛选与保级提醒流程。实现方式基于现有 CRM/任务中心 REST API，新增 Nuxt 页面逻辑、Pinia store 与 composables，保持所有批量操作与导出均写入审计并通过任务中心可追踪；同时追加 `useCustomerMetrics` 埋点与错误提示/重试体验，保障 KPI（筛选≤3 次点击、导出≤10 分钟、提醒成功率≥98%、错误率<1%）可量化验证。

## Technical Context

**Language/Version**: TypeScript 5.x + Nuxt 4（Node 20 运行时）  
**Primary Dependencies**: Nuxt UI 3.3.x、Pinia、@vueuse/core、@nuxtuse/asyncData、内部任务中心与 CRM REST API  
**Storage**: N/A（前端仅消费既有 API；后端由 CRM 服务管理数据）  
**Testing**: `npm run lint`, `npm run test`（Vitest 单测）以及 Cypress/Playwright E2E  
**Target Platform**: web-admin（宿主 `/_p/<plugin-id>/admin/*` 反代 Nuxt 输出）  
**Project Type**: Web 前端单仓（Nuxt 应用）  
**Performance Goals**: 列表筛选/分页响应 < 2s；详情抽屉初次渲染 < 1.5s；导出 5k 条客户数据任务 ≤ 10 分钟完成  
**Constraints**: 需遵循多租户/零信任约束，所有请求携带 tenant_uuid；敏感字段遮罩并需权限；批量任务/导出必须写入 `admin_console_audit_events` 并展示在任务中心  
**Scale/Scope**: 目标支撑 ≥50k 客户分页浏览、单次批量操作≤5k 记录、会员视角覆盖全部等级/渠道

## Runtime Config & Tenant Context

- **API Base**：`web-admin/nuxt.config.ts` 已将 `runtimeConfig.public.apiBaseUrl` 指向宿主代理（`/_p/<plugin-id>/api/v1`）或 standalone `http://localhost:8078/api/v1`。页面/composable 统一通过 `const apiBase = useRuntimeConfig().public.apiBaseUrl;` 获取，不再手动拼接。
- **Tenant UUID**：宿主脚手架在登录后写入 `tenant_uuid` Cookie 并将 tid/tenant_uuid 编码进 access token，`useAuth`/`getTenantUuid()` 已封装读取逻辑。后续客户域 API 只需依赖 API 客户端读取 cookie / token 上下文或显式携带 query `tenant_uuid` 即可。
- **审计记录**：批量操作/导入导出结果依旧由宿主任务中心写入审计，但本插件不再附加额外自定义头部，仅依赖宿主已有上下文。

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **Host Contract First**：前端只通过 `runtimeConfig.public.apiBaseUrl` 调用 `/api/customers/**`、`/api/jobs/**`，保持宿主反代契约。
- ✅ **Tenant Isolation & Zero Trust**：遵循宿主 JWT/STS；敏感字段强制 Mask，导出使用签名链接。
- ✅ **Service-Centric Architecture**：仅扩展 web-admin 视图与 state，不新增后端 handler/service，沿用既有 CRM/任务中心。
- ✅ **Observable & Testable Delivery**：批量操作与导出触发任务中心并写入审计；前端新增单测/E2E 用例。
- ✅ **Minimal Footprint**：局限在 `web-admin/app/pages/customer/**`、组件、Pinia store、i18n，无新增子项目或语言栈。

## Project Structure

### Documentation (this feature)

```text
specs/001-customer-ops-customer/
├── plan.md          # 本文件
├── research.md      # Phase 0 输出
├── data-model.md    # Phase 1 输出
├── quickstart.md    # Phase 1 输出
├── contracts/       # Phase 1 输出（API 契约）
└── tasks.md         # Phase 2 (/speckit.tasks) 生成
```

### Source Code (repository root)

```text
web-admin/
├── app/
│   ├── pages/customer/index.vue
│   ├── pages/customer/members.vue
│   ├── components/customer/
│   │   ├── CustomerDetailDrawer.vue
│   │   ├── CustomerBulkActions.vue
│   │   └── MembershipCards.vue
│   ├── components/Modals/{Create,Edit,View}CustomerModal.vue
│   ├── composables/
│   │   ├── useCustomerFilters.ts
│   │   ├── useCustomerBulkActions.ts
│   │   ├── useMembershipInsights.ts
│   │   └── useCustomerMetrics.ts
│   ├── stores/customer/index.ts
│   └── types/customer.ts
└── tests/
    ├── unit/customer-list.spec.ts
    └── e2e/customer-members.cy.ts

backend/ (参考现有 API，无需代码变更)
```

**Structure Decision**: 此特性完全落在 `web-admin` Nuxt 应用中，更新页面/组件/store/composables；后端代码仅作为 API 合同参考，无新增目录。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| *(None)* | — | — |

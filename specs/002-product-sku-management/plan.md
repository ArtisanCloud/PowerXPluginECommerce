# Implementation Plan: SKU 与变体管理能力增强

**Branch**: `002-product-sku-management` | **Date**: 2025-12-18 | **Spec**: specs/002-product-sku-management/spec.md
**Input**: Feature specification from `/specs/002-product-sku-management/spec.md`

## Summary

围绕 SKU 生成、批量调整与渠道映射建立统一的后台+前端交互：后端提供生成器、批量任务、审批流、渠道映射及库存同步 API，并保障多租户审计；同时补齐“规格维度/取值（SpecGroup/SpecOption）”一等数据与 `spec_signature` 唯一约束，并提供 mini-app 的 `GET /mini-app/products/{id}/detail` 聚合接口供前端做规格选择/禁用态与 SKU 匹配；前端 web-admin 在 SKU 列表、矩阵与详情中提供 SKU 生成器、批量操作、渠道配置、库存可视与条码管理体验，确保 50 个 SKU 10 分钟内配置完成、批量任务 ≥98% 成功且库存同步 ≤5 分钟。

## Technical Context

**Language/Version**: Backend Go 1.24（PowerX 插件栈），Frontend TypeScript 5.9 + Nuxt 4（Node 20）  
**Primary Dependencies**: Gin HTTP、GORM(Postgres driver)、PowerX plugin SDK、Redis 客户端、Nuxt UI 3.3.x、Pinia、Nuxt i18n、@nuxt/icon  
**Storage**: PostgreSQL（schema `powerx_plugin_base`，表含 `product_skus`、`product_sku_channels`、`product_sku_inventory` 等），可选 Redis 做库存缓存与任务锁  
**Testing**: `make test`（Go 单测+集成）、`make test-admin`（Nuxt vitest/Playwright），批量任务模拟需覆盖审批/失败分支  
**Target Platform**: 插件后端（Linux 宿主上运行）+ Web Admin（宿主反代 `/_p/<plugin-id>/admin`）  
**Project Type**: Web（后端 API + Nuxt 管理端）  
**Performance Goals**: SKU 生成器 50 条组合 10 分钟内可提交；批量任务 ≥98% 成功率；库存同步 ≤5 分钟；导入 1k 行 1 分钟内返回校验  
**Constraints**: 多租户隔离（tenant_uuid + RLS）、STS 鉴权、`/v1` API 合同、批量任务需审批策略、条码唯一性 <0.5% 冲突、库存 SLA 5 分钟  
**Scale/Scope**: 典型单租户 1k~5k SKU、数十渠道；批量任务一次影响 ≤5k SKU；导入/导出 CSV 10k 行以内。

## Constitution Check

1. **Host Contract First**：所有新 API 放置 `/v1/products/skus/**`，并在 `plugin.yaml` manifest 更新，确保与宿主反代兼容 —— *PASS*。  
2. **Tenant Isolation & Zero Trust**：模型包含 `tenant_uuid`、Handlers 通过 JWT/STSes 验签，Repo 在 `BeginTenantTx` 内执行 —— *PASS*。  
3. **Service-Centric Architecture**：新增服务/仓储放置 `backend/internal/services/admin/productsku` 与对应 transport/repository，Handlers 仅调用 Service —— *PASS*。  
4. **Observable & Testable**：批量任务、审批、渠道推送需写审计、结构化日志并提供任务状态查询；计划配套单测/集成测 —— *PASS*。  
5. **Minimal Footprint & Versioned Releases**：沿用现有 Go/Nuxt 栈，无新增语言，交付更新 `plugin.yaml` & 文档 —— *PASS*。  
6. **Operational Constraints**：Postgres schema 统一 `powerx_plugin_base`，配置保存在 `backend/etc/`，前端满足 Nuxt UI 约束 —— *PASS*。  
**Post-Design Re-check**：Phase 1 数据模型与合同维持相同分层，无新增风险 —— *PASS*。

> 注：规格定义相关的新接口位于 `/v1/admin/product/spus/{id}/spec-groups`，mini-app 聚合接口位于 `/v1/mini-app/products/{id}/detail`；均已补齐合同与调用指南。

## Project Structure

### Documentation (this feature)

```text
specs/002-product-sku-management/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── skus.openapi.yaml
└── tasks.md  # 由 /speckit.tasks 生成
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── plugin/               # 启动入口，加载 HTTP/gRPC
├── internal/
│   ├── transport/http/admin/productsku/   # 新增 SKU HTTP handler
│   ├── services/admin/productsku/         # 业务聚合（生成、批量、渠道）
│   ├── entity/models/productsku/          # SKU、渠道、任务、媒体模型
│   └── entity/repository/productsku/      # SKU & 任务仓储
├── pkg/                                  # 共用组件（审批、任务队列）
└── tests/
    ├── integration/
    └── services/

web-admin/
├── app/pages/product/skus/               # 列表/详情/生成器页面
├── app/pages/product/spus/edit/         # SPU 编辑页中的 SKU 区块
├── app/pages/product/inventory.vue      # 库存视图联动
├── app/components/product/sku/          # SKU 矩阵、批量操作弹窗
├── app/stores/productSku.ts             # Pinia store for SKU state
└── tests/
    ├── unit/
    └── e2e/
```

**Structure Decision**: 继续沿用现有 backend + web-admin 双应用结构；新增目录全部落在 `productsku` 子域下，前端页面/组件置于 `web-admin/app/pages/product/skus/**` 与内部组件目录，满足宪章的 handler→service→repository 分层。

## Complexity Tracking

（无额外豁免需求）

## Open Follow-ups

- “关联 SKU”弹窗当前仍将 payload 数据直接渲染到 SPU 页面摘要，且 `listSpuSkus` API 仅返回版本 payload。需要新增真实的 `GET /admin/product/skus?spuId=` 查询并更新摘要组件只读 `product_skus` 数据源（新任务 T061）。

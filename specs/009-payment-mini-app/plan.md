# Implementation Plan: 小程序支付

**Branch**: `009-payment-mini-app` | **Date**: 2026-01-19 | **Spec**: `/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/009-payment-mini-app/spec.md`  
**Input**: Feature specification from `/specs/009-payment-mini-app/spec.md`

## Summary

交付小程序支付闭环：下单确认 → 拉起支付 → 结果确认 → 结果页展示与重试；回调确认优先、前端短时确认兜底；支付成功后订单状态同步更新并可追溯。

## Technical Context

**Language/Version**: Go 1.24（backend）、Node 20 + TypeScript 5.9 + Nuxt 4（mini-app）  
**Primary Dependencies**: Gin、GORM、PowerX plugin framework、PowerWechat、Uni-app  
**Storage**: PostgreSQL（`powerx_plugin_base` schema）  
**Testing**: `go test ./...`（支付服务与回调幂等）  
**Target Platform**: Linux server + 小程序  
**Project Type**: backend + mini-app  
**Performance Goals**: 95% 用户在 10 秒内获得最终支付结果  
**Constraints**: 多租户 RLS、Host Contract（`/_p/<plugin-id>/api/v1`）、回调幂等与签名验真  
**Scale/Scope**: 小程序支付路径与回调闭环，不包含后台支付管理

## Project Structure

### Documentation (this feature)

```text
specs/009-payment-mini-app/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── internal/services/agent/payments/
├── internal/transport/http/agent/payments/
└── internal/entity/{models,repository}/

mini-app/
├── src/pages/order/
└── src/services/
```

## Phase 0: Research

输出：`/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/009-payment-mini-app/research.md`

## Phase 1: Design & Contracts

输出：`data-model.md`、`contracts/openapi.yaml`、`quickstart.md`

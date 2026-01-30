# Repository Guidelines

## Project Structure & Module Organization
- Root Makefile aggregates tasks under `make-files/`; targets assume backend at `backend` and frontend at `web-admin`.
- `backend/`: Go plugin service (`cmd/plugin` entrypoint), migrations/seeds in `cmd/database`, domain logic under `internal/`, shared configs in `etc/`, and Go module files (`go.mod`, `go.sum`).
- `web-admin/`: Nuxt-based admin UI (built with `nuxi`, outputs to `.output/`); tests and lint configs sit beside usual `app/`, `components/`, and `pages/`.
- Release artifacts land in `dist/` or `target/` via Make targets; plugin metadata stays in `plugin.yaml`.

## Build, Test, and Development Commands
- `make dev`: run DB migrations then start backend in dev mode (binds `:8086`, debug logging).
- `make run` / `make migrate` / `make seed`: backend dev server or data setup scripts.
- `make frontend-build` (hosted baseURL) or `make frontend-build-standalone`: build admin UI for the proxied or standalone path.
- `make build`: Go binary build into `backend/bin`; `make package`/`make release` pack dist/release zips with version from `plugin.yaml`.
- Quick starts: `go run ./backend/cmd/plugin` for backend; `npm run dev` inside `web-admin` for UI preview.

## Coding Style & Naming Conventions
- Go: keep `go fmt` clean (`make fmt`), prefer standard library patterns, and run `golangci-lint` via `make lint`. Use descriptive package/filename names aligned with domain (`internal/<domain>`). Tests mirror packages with `_test.go`.
- Frontend: follow existing Nuxt conventions; run `npm run lint -- --max-warnings=0` or `make lint-admin`. Use kebab-case for Vue file names and consistent composable naming (`useX`).
- Keep plugin metadata consistent: update `plugin.yaml` version when shipping.

## Testing Guidelines
- Backend: `make test` (`go test ./...`) for unit coverage; `make test-coverage` to emit `backend/coverage.html`. Integration hooks live under `internal/services/integration` with targeted runs like `make integration-smoke`.
- Frontend: `make test-admin` (`npm run test`) for UI checks. Add snapshot or component tests near source files.
- Prefer “Arrange-Act-Assert” naming inside tests and meaningful test names that include intent and condition.

## Commit & Pull Request Guidelines
- Use Conventional Commits with scopes (examples in history: `feat(migrate): ...`, `feat(project): ...`). Include module scope (`backend`, `web-admin`, `docs`) when helpful.
- PRs should describe intent, main changes, and rollout notes; link issues/Tickets. Call out migrations or config changes explicitly.
- For UI changes, attach before/after screenshots or short Loom/GIF. Mention test commands executed (`make test`, `make lint-admin`) and any outstanding risks.


<!-- MANUAL ADDITIONS START -->
Always respond in Chinese-simplified
<!-- MANUAL ADDITIONS END -->

## Active Technologies
- Backend Go 1.24; Frontend Node 20 + TypeScript 5.9 + Nuxt 4 + Gin, GORM (postgres driver), PowerX plugin framework, JWT、grpc/protobuf、Redis 客户端；前端 @nuxt/ui 3.3.x、Pinia、i18n、Nuxt Icon (003-channels-subscription)
- PostgreSQL（schema: powerx_plugin_base），可选 Redis 做列表/任务缓存 (003-channels-subscription)
- TypeScript 5.x + Nuxt 4（Node 20 运行时） + Nuxt UI 3.3.x、Pinia、@vueuse/core、@nuxtuse/asyncData、内部任务中心与 CRM REST API (001-customer-ops-customer)
- N/A（前端仅消费既有 API；后端由 CRM 服务管理数据） (001-customer-ops-customer)
- Go 1.24（backend）、TypeScript 5.9 + Nuxt 4（web-admin） + Gin HTTP、GORM（Postgres driver）、PowerX plugin framework、JWT/STS、多租户任务中心、Nuxt UI 3.3.x、Pinia、@vueuse/core、Nuxt i18n (002-product-spu-management)
- PostgreSQL (`powerx_plugin_base` schema) — 表：`product_spus`、`product_spu_versions`、`product_spu_channels`、`product_spu_subscription_plans`、`product_spu_audit_logs`、任务记录表等 (002-product-spu-management)
- Backend Go 1.24；Frontend TypeScript 5.9 + Nuxt 4（Node 20） + Gin、GORM（Postgres driver）、PowerX plugin framework、Redis client、Nuxt UI 3.3.x、Pinia、Nuxt i18n、@vueuse/core (003-channel-master)
- PostgreSQL（schema `powerx_plugin_base`）存放 channel_* 表；Redis 用于凭证巡检与 KPI 缓存 (003-channel-master)
- Backend Go 1.24（PowerX 插件栈），Frontend TypeScript 5.9 + Nuxt 4（Node 20） + Gin HTTP、GORM(Postgres driver)、PowerX plugin SDK、Redis 客户端、Nuxt UI 3.3.x、Pinia、Nuxt i18n、@nuxt/icon (002-product-sku-management)
- PostgreSQL（schema `powerx_plugin_base`，表含 `product_skus`、`product_sku_channels`、`product_sku_inventory` 等），可选 Redis 做库存缓存与任务锁 (002-product-sku-management)
- Go 1.24（backend），Node 20 + TypeScript 5.9 + Nuxt 4（web-admin） + Gin、GORM（postgres driver）、PowerX plugin framework；Nuxt 4、@nuxt/ui 3.3.x、Pinia、Nuxt i18n (004-product-categories)
- PostgreSQL（schema: `powerx_plugin_base`） (004-product-categories)
- Go 1.24 + Gin、GORM（postgres driver）、PowerX plugin framework（RBAC/tenant context/STS） (005-pricing-pricebook)
- PostgreSQL（schema：`powerx_plugin_base`，RLS 强制，所有表带 `tenant_uuid`） (005-pricing-pricebook)
- Go 1.24 + Node 20（Nuxt 4） + Gin、GORM（postgres driver）、PowerX plugin framework（RBAC/tenant context/STS）；Nuxt 4 + TypeScript 5.9 + Nuxt UI 3.3.x (006-sku-inventory-stock)
- Go 1.24（backend）、Node 20 + TypeScript 5.9 + Nuxt 4（web-admin / mini-app） + Gin、GORM、PowerX plugin framework、Nuxt UI 3.3.x、Pinia、Uni-app (008-order-payments)
- PostgreSQL（`powerx_plugin_base` schema） (008-order-payments)
- Go 1.24; TypeScript 5.9 + Nuxt 4 (web-admin) + Uni-app (mini-app) + Gin, GORM (Postgres), PowerX plugin framework; Nuxt UI 3.3.x (010-membership-entitlements)

## Recent Changes
- 003-channels-subscription: Added Backend Go 1.24; Frontend Node 20 + TypeScript 5.9 + Nuxt 4 + Gin, GORM (postgres driver), PowerX plugin framework, JWT、grpc/protobuf、Redis 客户端；前端 @nuxt/ui 3.3.x、Pinia、i18n、Nuxt Icon

# Repository Guidelines

## 项目结构与模块组织

- 说明：本目录 `mini-app/` 是 `com.powerx.plugin.ecommerce` 插件仓库的一部分（仓库根在上一级目录）。
- `backend/`：Go 插件后端（入口：`backend/cmd/plugin`），HTTP 前缀默认 `/api/v1`；mini-app 相关接口在 `backend/internal/transport/http/miniapp/`。
- `web-admin/`：Nuxt 4 管理端（面向运营/管理员），构建产物在 `web-admin/.output/`，菜单与路由由 `plugin.yaml` 声明。
- `mini-app/`：`uni-app + Vue3` 客户端（面向 C 端/小程序/H5）。当前仍是脚手架页面，实际业务联调需要对接后端 mini-app API。
- `make-files/`：统一构建/测试/发布脚本；插件元数据在 `plugin.yaml`。

## 构建、测试与开发命令

- `make dev`：执行迁移后启动后端开发服务。
- `make run` / `make migrate` / `make seed`：分别用于启动后端、跑迁移、写入种子数据。
- `make build`：编译 Go 二进制到 `backend/bin/`；`make package` / `make release`：根据 `plugin.yaml` 版本打包发布。
- `make frontend-build` / `make frontend-build-standalone`：构建管理端（代理路径/独立路径）。
- 快速启动：`go run ./backend/cmd/plugin`（本地默认 `:8078`，可通过 `POWERX_BIND_ADDR` 覆盖）；`(cd web-admin && npm run dev)`。
- 小程序本地开发：在 `mini-app/` 下执行 `npm run dev:mp-weixin`（或 `npm run dev:h5`）；类型检查：`npm run type-check`。
- 样式体系：mini-app 使用 Tailwind（配置：`tailwind.config.cjs`，入口：`src/styles/tailwind.scss`，通过 `src/uni.scss` 全局引入）。

## 编码风格与命名规范

- Go：保持 `go fmt`（可用 `make fmt`），优先标准库习惯；静态检查使用 `golangci-lint`（`make lint`）。
- 前端：Nuxt 与 uni-app 都遵循各自约定；Vue 文件用 kebab-case（如 `product-list.vue`），组合式函数用 `useX`；管理端 Lint：`(cd web-admin && npm run lint -- --max-warnings=0)` 或 `make lint-admin`。

## 测试规范

- 后端：`make test`（`go test ./...`），覆盖率：`make test-coverage` 生成 `backend/coverage.html`；集成测试在 `backend/internal/services/integration`。
- 前端：`make test-admin`（`npm run test`）；新增测试尽量就近放置，并使用清晰的 Arrange-Act-Assert 命名。

## mini-app ↔ 后端联动（必须了解）

- API Base：插件内部基路径为 ` /api/v1`（注意：与标准模板常见的 `/v1` 不同），mini-app 前缀为 `/api/v1/mini-app`。
- 对接模式：
  - **宿主模式（推荐）**：通过 PowerX 访问 `/_p/<plugin-id>/api/v1/mini-app/...`；宿主会注入签名上下文（通常是 `Authorization: Bearer ...` 或 `X-PowerX-CTX`/`X-PowerX-CTX-SIG`），插件可从中解析租户信息。
  - **Skeleton/本地直连**：直接访问 `http://localhost:8078/api/v1/mini-app/...`；必须显式传 `X-Tenant-UUID: <uuid>`（或 query `tenant_uuid=<uuid>`），否则 401。开发期可用 `POWERX_DEV_MODE=1`，线上禁用。
- 响应结构：统一 JSON Envelope（通常形如 `{"code":0,"message":"ok","data":...}`），mini-app 也遵循同一规范。
- 登录与鉴权：
  - `POST /auth/register`、`POST /auth/login` 返回 `token/expiresAt/customerId/tenantUuid`（`customer_auth.mode=delegate` 时注册可能不可用）。
  - 访问受保护接口携带 `Authorization: Bearer <token>` 或 `X-Customer-Token: <token>`。
- 已实现的受保护接口：`GET /categories/tree`、`GET /products`、`GET /products/:id/skus`。

## 参考规范（建议先读）

- API/上下文头规范：`PowerXPlugin/docs/standards/powerx-plugin/developer/api_guidelines.md`
- 架构与反代路径：`PowerXPlugin/docs/standards/powerx-plugin/overview/architecture.md`
- 宿主模式 vs Skeleton 快速跑通：`PowerXPlugin/docs/standards/powerx-plugin/overview/quick_start.md`

## 提交与 PR 规范

- 提交信息：使用 Conventional Commits（如 `feat(migrate): ...`、`fix(web-admin): ...`），范围可选 `backend` / `web-admin` / `docs`。
- PR：写清意图与影响面，关联 Issue/工单；涉及迁移/配置需显式说明；UI 变更请附截图/GIF，并列出已执行命令（如 `make test`、`make lint-admin`）。

## 安全与配置提示（可选）

- 依赖版本：后端 Go 1.24；前端 Node 20 + TypeScript 5.9 + Nuxt 4。
- 数据库默认使用 PostgreSQL（schema 常见为 `powerx_plugin_base`）；如启用 Redis，请避免将敏感凭证写入仓库。

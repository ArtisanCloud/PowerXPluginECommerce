# Quickstart — Pricing Pricebook（Phase 1）

## Prerequisites
- Go 1.24、PostgreSQL ≥ 13。
- 已配置插件后端环境变量（如 `POWERX_BIND_ADDR`、`POWERX_DB_DSN`、`POWERX_DEV_MODE` 等）并能启动服务。

## Backend Development Workflow
1. **运行迁移**
   ```bash
   make migrate
   ```
2. **启动开发服务**
   ```bash
   make dev
   ```
3. **验证管理端接口（示例）**
   - 价目表列表：`GET /api/v1/admin/pricing/pricebooks`
   - 创建价目表（会自动生成 v1 draft）：`POST /api/v1/admin/pricing/pricebooks`
   - 创建新版本：`POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions`
   - 批量写入条目：`PUT /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items`
   - 发布版本：`POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/publish`
4. **验证查价接口（示例）**
   - 查价（主路径）：`POST /v1/pricing/query`
   - 查价（兼容别名）：`POST /api/v1/pricing/query`

## Contract Reference
- OpenAPI：`/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml`

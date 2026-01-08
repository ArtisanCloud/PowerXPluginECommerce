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
3. **本地无鉴权冒烟（可选）**
   - 本地开发时可设置 `POWERX_AUTH_OPTIONAL=1` 方便直接 curl 冒烟（生产环境勿用）
4. **验证管理端接口（示例）**
   - 价目表列表：`GET /api/v1/admin/pricing/pricebooks`
   - 创建价目表（会自动生成 v1 draft）：`POST /api/v1/admin/pricing/pricebooks`
   - 创建新版本：`POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions`
   - 批量写入条目：`PUT /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items`
   - 发布版本：`POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/publish`
5. **验证查价接口（示例）**
   - 查价（主路径）：`POST /v1/pricing/query`
   - 查价（兼容别名）：`POST /api/v1/pricing/query`

## Smoke Examples (curl)
> 以下以本地默认端口为例，实际请按你的 `POWERX_BIND_ADDR` 调整。

### Create Pricebook
```bash
curl -sS -X POST 'http://localhost:8086/api/v1/admin/pricing/pricebooks' \
  -H 'Content-Type: application/json' \
  -d '{"code":"base","name":"Base Pricebook","type":"sales","currency":"CNY"}'
```

### Upsert Items (draft only)
```bash
curl -sS -X PUT 'http://localhost:8086/api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items' \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"sku_id":"{skuId}","base_amount_minor":19900}]}'
```

### Publish Version
```bash
curl -sS -X POST 'http://localhost:8086/api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/publish' \
  -H 'Content-Type: application/json' \
  -d '{"note":"publish v1"}'
```

### Pricing Query
```bash
curl -sS -X POST 'http://localhost:8086/v1/pricing/query' \
  -H 'Content-Type: application/json' \
  -d '{"sku_id":"{skuId}","currency":"CNY","as_of":"2026-01-01T00:00:00Z"}'
```

## Contract Reference
- OpenAPI：`/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugin.ecommerce/specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml`

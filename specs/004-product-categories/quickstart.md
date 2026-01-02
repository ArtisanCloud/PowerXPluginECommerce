# Quickstart: 商品类目与类目模板管理（Phase 1）

## 运行后端（含迁移）

- `make dev`

或分步：

- `make migrate`
- `make run`

## 运行 web-admin

- `cd web-admin && npm install`
- `npm run dev`

## 验证（建议顺序）

- 后端单测：`make test`
- 后端 lint：`make lint`
- 前端构建：`make build-admin`

## 联调入口（路径约定）

- 管理端 API 前缀：`/_p/com.powerx.plugin.ecommerce/api/v1/admin/**`
- miniapp API 前缀：`/_p/com.powerx.plugin.ecommerce/api/v1/mini-app/**`


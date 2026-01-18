# Quickstart: 营销支付与小程序支付

## Prerequisites
- Backend: Go 1.24
- Web admin: Node 20 + npm
- Mini app: Node 20 + npm
- Database: Postgres (schema `powerx_plugin_base`)

## Backend (支付管理与回调)
1. 启动依赖数据库并完成迁移。
2. 启动开发服务：`make dev` 或 `go run ./backend/cmd/plugin`。
3. 验证管理端 API 路径通过反代前缀 `/api/v1` 可访问。

## Web Admin (支付管理页面)
1. 进入 `web-admin/`。
2. 安装依赖并启动：`npm install`、`npm run dev`。
3. 打开后台支付页面进行联调。

## Web Admin (代客下单体验对齐)
1. 打开“订单管理”页面，点击“新建订单”。
2. 搜索客户（无关键词时应显示最近下单 TopN）。
3. 搜索并选择 SPU，选择规格组合，确认定位唯一 SKU。
4. 增加多行 SKU 并提交订单。

## Mini App (支付流程)
1. 进入 `mini-app/`。
2. 安装依赖并启动本地构建/预览脚本（按项目现有脚本）。
3. 验证支付发起、结果页与订单列表入口。

## Smoke Checks
- 支付渠道列表可加载与筛选
- 支付单列表与详情可查看
- 退款与对账流程可记录结果
- 小程序支付失败可重试并回到订单列表
- 后台新建订单支持 SPU + 规格定位 SKU 与多 SKU 提交
- 客户下拉无关键词时按最近下单排序，输入关键词按匹配结果搜索

## Self-check Log
| Date | Item | Result | Notes |
| --- | --- | --- | --- |
| 2026-01-16 | Smoke checks | 未执行 | 待本地环境联调 |
| 2026-01-16 | Reconciliation flow | 未执行 | 待后端迁移与接口联调 |

## Performance Record
| Date | Endpoint | p95 (ms) | Samples | Notes |
| --- | --- | --- | --- | --- |
| 2026-01-16 | GET /admin/payments/transactions | 未执行 | 0 | 待压测 |
| 2026-01-16 | GET /admin/payments/transactions/{id} | 未执行 | 0 | 待压测 |

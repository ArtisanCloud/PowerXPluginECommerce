lear# 订阅型商品上架 & 下单准备用例

该用例覆盖「创建 SPU → 审核发布 → 生成 SKU → 渠道/库存配置 → 对外可售」的完整链路，用于验证订阅型商品在系统内可被前端展示并为后续下单提供数据基础。

## 背景与目标

- 参考文档：`specs/002-product-spu-management/*`、`specs/002-product-sku-management/*`
- 目标：确保运营可在 Web Admin 完成订阅型商品的上架，且相关 API (`/api/v1/admin/product/spus`, `/api/v1/admin/product/skus`) 能返回可消费的数据。
- 成功标准：发布后的 SPU 状态为 `published`，至少存在一条 `product_skus` 记录，渠道/库存配置可查询，前端 SKU 列表能显示基础信息。

## 依赖与权限

| 类型          | 需求                                                                               |
| ------------- | ---------------------------------------------------------------------------------- |
| RBAC Scopes   | `product.spu:create`, `product.spu:publish`, `product.sku:create`, `product.sku:read`, `channel.product:manage` |
| 环境配置      | 已执行 `make migrate && make seed`；`backend/etc/config.yaml` 指向可用的 Postgres/Redis |
| 账号/租户     | 默认本地租户 `00000000-0000-0000-0000-000000000001`，使用 `make dev` 创建的管理员账号 |

> 如果使用独立宿主，请同步脚手架里的 RBAC 策略（详见 `.specify/memory/rulesets/plugin_rbac*.yaml`），确保以上 scope 已授予登录用户。

## 数据准备

1. 启动服务：`make dev`（依次执行迁移 → backend dev server）。
2. 登录 Web Admin (`npm run dev` → http://127.0.0.1:3032) 并切换到电商插件命名空间。
3. 若是首次体验，可运行 `make seed` 写入示例品牌、类目、渠道等基础数据。

## 测试步骤

### 1. 创建订阅型 SPU

1. 进入「产品 → 商品 (SPU) → 新建」，填写：
   - 类型：`subscription`
   - 编码：`SUB-PLAN-001`
   - 类目 / 品牌：选择任意可用项
   - 其他字段保持默认或根据业务填写
2. 点击「保存草稿」后在列表中定位到该 SPU。
3. **验证**：
   - 数据库：`SELECT status FROM product_spus WHERE code='SUB-PLAN-001';` 应为 `draft`
   - API：`GET /api/v1/admin/product/spus?keyword=SUB-PLAN-001`

### 2. 提交审核并发布

1. 在 SPU 列表选择刚创建的条目 → 点击「查看」进入详情。
2. 点击「提交审核」按钮，等待状态变为 `reviewing`。
3. 点击「发布」并选择至少一个渠道（如 `official`）；发布成功后状态为 `published`。
4. **验证**：
   - `product_spus` 表 `status`= `published`，`current_version_id` 非空。
   - 审批记录：`product_spu_versions` 对应记录 `status=published`，`approved_at` 有值。

### 3. 生成 SKU

1. 在 SPU 详情页点击「关联 SKU」。弹窗的两个标签页分别为：
   - **SKU 生成器**：批量组合规格字段。
   - **关联 SKU 列表**：逐条管理/调整 SKU。
2. 若已配置规格：在「SKU 生成器」页选择组合并点击「写入 SKU」，结果会同步到「关联 SKU 列表」页以便继续编辑。
3. 若无规格或需人工微调：切换到「关联 SKU 列表」页点击「新增 SKU」，至少填写：
   - `SKU 编码`（必填，例如 `SUB-PLAN-001-A`）
   - `所属 SPU`（自动带入）
   - 可选：条码、规格描述、起订量
4. 保存后关闭弹窗。
5. **验证**：
   - 后端日志应输出 `POST /api/v1/admin/product/skus → 201`
   - `product_skus` 表出现对应记录 (`status=online|ready`)
   - 前端 `/product/skus` 列表出现该条目，SKU 编码/状态显示正常

### 4. 配置渠道与库存

1. 在 SPU 详情页点击「渠道可见性」，为目标渠道设置发布时间 / 可见范围。
2. 如需同步库存，进入 SKU 列表 → 选择 SKU → 打开「库存面板」或 API `POST /api/v1/admin/product/skus/{id}/inventory` 更新数量。
3. **验证**：
   - 渠道记录写入 `product_spu_channels`，发布任务可在后台日志中看到。
   - 库存信息 `product_sku_inventories` 更新，前端 SKU 列表的库存/状态随之变化。

### 5. 客户端/下单准备

1. 调用 `/api/v1/admin/product/spus/{id}` 或面向渠道的公开 API（若有代理）检查商品详情。
2. 若订单服务需要 SKU 详情，可使用 `/api/v1/admin/product/skus?spuId={id}` 获取 SKU 列表。
3. 迷你应用/小程序场景：调用 `/api/v1/mini-app/products?keyword=SUB-PLAN-001&tenant_uuid=<uuid>`、`/api/v1/mini-app/products/{id}/skus?tenant_uuid=<uuid>` 对齐移动端可见的数据（迷你端 API 不再要求管理员 JWT）。
4. **验证**：各接口返回的 `status`、`channels`、`priceRefs`、`code` 等字段与配置一致，mini-app 响应中至少包含 `id/code/status`，确保可作为下单输入。

## 回归测试提示

- **后端集成**：在仓库根目录执行 `go test ./backend/tests -run TestSubscriptionProductOnboardingFlow`（`make test` / `make ci-all` 会自动执行）。若你已 `cd backend`，对应命令是 `go test ./tests -run TestSubscriptionProductOnboardingFlow`。该用例支持两种方式加载测试数据库：
  - 通过环境变量：`TEST_DATABASE_URL=postgres://... go test ./tests -run TestSubscriptionProductOnboardingFlow`
  - 通过独立配置文件：`CONFIG_PATH=../etc/config.test.yaml go test ./tests -run TestSubscriptionProductOnboardingFlow`
  上述命令会先清空 `powerx_plugin_test` 数据库，再执行 SPU 发布 + SKU 写入 + 列表校验。
  1. `cd backend`
  2. 如需指定数据库（例如本地 Postgres），设置 `TEST_DATABASE_URL=postgres://user:pass@localhost:5432/powerx_plugin_test?sslmode=disable`
  3. 运行 `go test ./tests -run TestSubscriptionProductOnboardingFlow`
- **前端冒烟**：`web-admin/tests/e2e/product-spu-create.spec.ts` 验证创建/提交流程，可在未来扩展 SKU、渠道等 UI 流程。
- **手动 Checklist**：执行以上 5 个步骤并截取 SKU 列表、渠道可见性、库存面板截图存档。若任何步骤失败，请记录请求 ID（`x-request-id`）以便排障。

## 常见故障 & 排障

| 现象                                            | 排查建议                                                                                   |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------ |
| 发布按钮不可用                                  | 确认当前版本 `status=reviewing`，审批链是否已配置（`product_spu_approvals`）。             |
| SKU 创建 403 / 501                              | 检查 RBAC（`product.sku:create`）和 `/backend/internal/transport/http/admin/product/rbac.go` 配置。 |
| SKU 列表显示空白（编码/状态缺失）                | 确认前端 `web-admin/app/pages/product/skus/index.vue` 的 `formatSkuRow` 是否收到合法字段。 |
| 渠道发布无响应                                   | 查看后台日志是否出现 `channel.product.publish` 任务，必要时手动调用 `/channels/publish` API。 |
| 前端 API 404 / 401                               | 检查 `tenant_uuid`（query）与 `Authorization` 是否正确；必要时重新登录刷新 token。 |

> 若需扩展自动化测试，请在本指南末尾记录新增脚本位置与命令，保持文档与代码一致。

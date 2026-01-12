# 常规商品上架 & 库存准备用例

该用例覆盖一次性购买商品（`type=one_time`）从草稿、审核到 SKU/库存同步的端到端链路，用于确保最常见的商品上线路径可被运营复用并可支撑后续下单。

## 背景与目标

- 参考文档：`specs/002-product-spu-management/*`、`specs/002-product-sku-management/*`
- 目标：能在 Web Admin 中创建标准商品、发布到官方/经销渠道，并完成 SKU 录入与状态校验。
- 成功标准：SPU 状态 `published`，至少 1 条线上 SKU 可搜索，`product_spu_channels` / `product_skus` / `product_sku_attributes` 等表写入正常。

## 依赖与权限

| 类型        | 需求                                                                                           |
| ----------- | ---------------------------------------------------------------------------------------------- |
| RBAC Scopes | `product.spu:create`、`product.spu:publish`、`product.sku:create`、`product.sku:read`、`channel.product:manage` |
| 环境配置    | `make migrate && make seed` 已执行；`backend/etc/config.yaml` 连接有效的 Postgres/Redis；前端 `npm run dev` |
| 账号/租户   | 默认租户 `00000000-0000-0000-0000-000000000001`，使用 `make dev` 自动创建的管理员账号          |

> 在 skeleton/standalone 模式下，需要同步 `.specify/memory/rulesets/plugin_rbac*.yaml` 中约定的 RBAC 策略，确保拥有以上 scope。

## 数据准备

1. `make dev` 启动后台（包含迁移、种子、开发服务）。
2. `cd web-admin && npm run dev` 启动管理端，访问 http://127.0.0.1:3032。
3. 若首次体验，可执行 `make seed` 写入基础类目、品牌、渠道配置。

## 测试步骤

### 1. 创建标准 SPU

1. Web Admin →「产品 → 商品 (SPU) → 新建」。
2. 填写：
   - 商品类型：`一次性商品 / one_time`
   - 编码：`STD-PRODUCT-001`
   - 名称：`常规商品`
   - 类目：选择任意可用类目
   - 责任人、描述等字段按业务填写
3. 保存草稿。
4. **验证**：`SELECT status FROM product_spus WHERE code='STD-PRODUCT-001';` 应为 `draft`。

### 2. 审核并发布

1. 在列表中打开该 SPU → 点击「提交审核」，状态变为 `reviewing`。
2. 在同一详情页点击「发布」，选择 `official`、`reseller` 等渠道。
3. **验证**：
   - `product_spus.status = 'published'`
   - `product_spu_channels` 存在渠道记录，包含 `official`、`reseller`

### 3. 录入 SKU

关联弹窗拆成两个独立表单：上方“批量生成器”负责自动产出 SKU，下方“手动关联”负责逐条维护。根据实际需要任选其一或组合使用。

**批量生成（可选）**

1. 在「SKU 生成器」面板里选择规格、默认字段。
2. 点击「写入 SKU」，生成的组合会同步到下方的手动列表，后续可再微调。

**手动关联 SKU**

1. 在下方「关联 SKU」表格点击「新增 SKU」：
   - `SKU 编码`：`STD-PRODUCT-001-A`
   - `SKU 名称`、`价格 (含税)`、`币种`、`库存引用` 等字段（编码/名称/价格/状态必填）
   - 起订量 = `min_order_qty`（如 `1`），扩展属性可填写 JSON。
2. 如需更多 SKU（例如 `STD-PRODUCT-001-B`），再新增一行或使用「克隆来源」复制已有配置。

> 若要一次写入多条 SKU，需保证每条的规格组合不同（如“容量：单件 / 十件装”），否则后台的去重逻辑会跳过重复组合。

3. 点击「保存关联」。
4. **验证**：
   - 后端日志出现 `POST /api/v1/admin/product/skus → 201`
   - `SELECT sku_code, status FROM product_skus WHERE spu_id='{SPU_ID}'` 返回两条数据

### 4. 库存/渠道可见性确认

1. 在 SPU 详情页 →「渠道可见性」确认 `official` 渠道已经启用。
2. 若需要库存：在 SKU 列表点击目标 SKU → 打开「库存管理」输入安全库存。
3. **验证**：
   - `product_sku_inventory`（如启用）记录写入
   - SKU 列表显示最新库存值，状态与步骤 3 中一致

### 5. 客户端/接口校验

1. 调用 `/api/v1/admin/product/spus?keyword=STD-PRODUCT-001` 确认列表包含该商品。
2. 调用 `/api/v1/admin/product/skus?spuId={id}&status=online` 仅返回线上 SKU。
3. 调用 `/api/v1/mini-app/products?keyword=STD-PRODUCT-001` 验证迷你端可读到同一 SPU（仅需携带 `X-Tenant-UUID`，无需管理员 JWT）。
4. 调用 `/api/v1/mini-app/products/{id}/skus`（mini-app 仅返回 `published` SKU）确认返回列表与正式 `product_skus` 一致。
5. **验证**：四个接口返回的 `status`、`inventoryRef`、`minOrderQty`、`code` 等字段应彼此一致；mini-app 响应字段精简但需包含 `id/code/status/updatedAt`。

### 5.1 上线购买前的“可售性”校验（建议）

为实现“能上线开始做购买”的最短闭环，建议在客户端统一使用后端的“可售性聚合”结果（价格+库存+状态+渠道窗口），而不是在多个页面散落判断：

- **拟新增**：`GET /api/v1/mini-app/products/{spuId}/sellability?channel=xxx&locale=zh-CN`
- 前端策略二选一：不可售 SKU **隐藏** 或 **置灰禁用下单**（并用 `reasons[]` 做可解释提示）
- 详见：`docs/guides/features/product/sellability_purchase_mvp.md`

## 回归测试提示

- **后端集成**：在仓库根运行 `go test ./backend/tests -run TestStandardProductOnboardingFlow`；若已经 `cd backend`，请改用 `go test ./tests -run TestStandardProductOnboardingFlow`。命令流程：
  1. `cd backend`
  2. 任选其一：  
     - `TEST_DATABASE_URL=postgres://user:pass@localhost:5432/powerx_plugin_test?sslmode=disable go test ./tests -run TestStandardProductOnboardingFlow`  
     - 或 `CONFIG_PATH=../etc/config.test.yaml go test ./tests -run TestStandardProductOnboardingFlow`
  3. 测试会自动清空 `powerx_plugin_test`（或配置文件中指定的数据库）后再写入测试数据
- **自动清库**：所有测试使用 `tests/testutil.NewIsolatedDB`，每次运行前会 DROP 当前 schema 下的表，确保独立性。
- **前端配合**：可扩展 Playwright 用例，在创建 SPU 后验证 SKU 弹窗交互（待补充到 `web-admin/tests/e2e`）。

## 常见故障 & 排障

| 问题                                            | 解决建议                                                                 |
| ----------------------------------------------- | ------------------------------------------------------------------------ |
| 发布时报「仍在审核中」                          | 确认 `Submit` 流程成功执行，`product_spu_versions.status` 是否为 `reviewing`。 |
| SKU 保存 403                                    | 检查 RBAC scopes，确认 `backend/internal/transport/http/admin/product/rbac.go` 中映射正确。 |
| 列表没有显示 SKU 编码                           | 可能是前端 `useSku` 组合未更新，检查 API 响应是否包含 `sku_code` 字段。 |
| API 返回旧数据                                  | 如果本地使用 Postgres，确保未开启查询缓存；必要时清除 Redis 或重启 `make dev`。 |

> 如需扩展为“渠道限售”、“套装 SKU”等复杂场景，请在此文档末尾追加子章节或引用新的 feature guide。

# Mini-App（小程序）开放接口调用指南（类目树 / 商品列表 / SKU 列表）

本指南用于指导小程序以“只读/游客态”方式访问电商插件的开放接口，典型场景：**按类目展示商品列表**、**进入商品详情后加载 SKU 列表**，以及（可选）**一次性获取规格维度/取值与 SKU 映射用于规格选择**。

> 说明：mini-app 接口不继承管理端的 JWT/RBAC（无需管理员 JWT），但**必须携带租户上下文**（`X-Tenant-UUID` 或 `tenant_uuid`）。

## 基本约定

### Base URL

- 默认：`/api/v1`
- 若宿主配置了自定义 `APIPrefix`，则以宿主为准（例如 `/api/v2`）。

### 必须的 Header / Query

| 名称 | 位置 | 是否必填 | 说明 |
| --- | --- | --- | --- |
| `X-Tenant-UUID` | Header | 是 | 租户 UUID；本地默认 `00000000-0000-0000-0000-000000000001` |
| `tenant_uuid` | Query | 否 | Header 不方便时可用（同等生效） |

### 鉴权（可选）

开放接口本身不要求 customer token；若你需要后续的用户态能力（例如收藏、下单等扩展），可使用以下登录接口获取 `customer token` 并以 `Authorization: Bearer <token>` 发送：

- `POST /api/v1/mini-app/auth/register`（仅 local 模式启用）
- `POST /api/v1/mini-app/auth/login`

## 响应结构

所有接口统一返回 `APIResponse`：

- 成功：`{ "success": true, "data": ..., "timestamp": "...", "requestId": "..." }`
- 失败：`{ "success": false, "error": { "code": "...", "message": "..." }, "timestamp": "...", "requestId": "..." }`

排障时请记录 `requestId`。

## 1) 获取可展示类目树

**接口**

- `GET /api/v1/mini-app/categories/tree`

**说明**

- 仅返回可展示类目（已停用的类目及其子树不应出现）。
- 返回为树形结构数组：`data.items`。

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/categories/tree' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 2) 按类目展示商品（SPU）列表

**接口**

- `GET /api/v1/mini-app/products`

**查询参数**

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `categoryId` | string | 精确筛选某类目下的 SPU |
| `categoryPathPrefix` | string | 按类目路径前缀筛选（用于“当前类目 + 子类目”） |
| `tags` | string | 按标签过滤（逗号分隔，例如 `bestsellers,imported`；与类目筛选可叠加） |
| `tag` | string | 单个标签过滤（兼容参数；等价于 `tags=<tag>`） |
| `keyword` / `q` | string | 模糊搜索（编码/名称等） |
| `type` | string | `one_time` / `subscription` |
| `sort` | string | `comprehensive` / `updatedAt` / `price` / `sales` |
| `order` | string | `asc` / `desc`（默认 `desc`） |
| `minPrice` | number | 可选：最低“展示价”（见下方口径） |
| `maxPrice` | number | 可选：最高“展示价”（见下方口径） |
| `inStock` | boolean | 可选：是否有库存（非订阅）/是否有可用计划（订阅） |
| `hasPlans` | boolean | 可选：是否存在 `active` 订阅计划（仅对 `type=subscription` 有意义） |
| `page` | number | 页码（从 1 开始） |
| `pageSize` | number | 每页数量 |

**注意**

- mini-app 只允许读取 `published` 商品；`status` 传入非 `published` 将返回 400。
- `tags/tag` 支持与 `categoryId/categoryPathPrefix` 同时使用；标签匹配为“任一命中”（商品 tags 与筛选 tags 有交集即通过）。
- 电商化字段：返回中会包含 `coverUrl`（封面图，来自 SKU 主图）、`priceLabel/minPrice/maxPrice/currency`（价格信息）、`skuCount`（可选项数量：一次性商品为 SKU 数；订阅商品为 active 计划数）。
- `sort`/`order` 在后端生效：排序在数据库层完成（随后分页），避免“只在单页内排序”的问题。
- `sort=sales` 目前无销量数据源，后端会返回 400（避免前端误以为已按销量排序）。

**价格来源约定（当前实现）**

- `type=subscription`：来自 `product_spu_subscription_plans` 中 `status=active` 的最小价格。
- `type=one_time`：来自 `product_skus.default_values` JSON（优先 `sale_price`，其次 `price`，再其次 `salePrice/list_price`）。
  - 示例写法：`default_values={"sale_price":199,"currency":"CNY"}`。

**价格筛选/排序口径（当前实现）**

- `sort=price`、`minPrice/maxPrice` 使用同一口径：按 SPU 下已发布 SKU 的 `MIN(展示价)`（订阅商品则按订阅计划 `MIN(price)`）。
- `inStock=true/false`：
  - `type=subscription`：以是否存在 `active` 订阅计划作为可售判断（等价于 `hasPlans=true/false`）。
  - `type=one_time`：基于 `product_sku_inventories` 聚合可用库存（`available_qty - locked_qty`）判断。

**示例：按类目路径前缀筛选**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products?categoryPathPrefix=/root/child/&page=1&pageSize=20' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

**示例：类目 + 标签联合筛选**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products?categoryPathPrefix=/root/child/&tags=bestsellers,imported&page=1&pageSize=20' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

**示例：按价格排序（从低到高）**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products?sort=price&order=asc&page=1&pageSize=20' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

**示例：价格区间 + 有库存**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products?minPrice=99&maxPrice=199&inStock=true&page=1&pageSize=20' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 2.1) 获取商品可用标签（Tag）列表

用于前端渲染“标签筛选器”（如 Bestsellers / New Arrival / Flash Sale 等）。

**接口**

- `GET /api/v1/mini-app/products/tags`

**查询参数**

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `categoryId` | string | 可选：只统计某类目下的标签 |
| `categoryPathPrefix` | string | 可选：只统计某类目路径前缀下的标签 |
| `limit` | number | 返回数量上限（默认 50，最大 100） |

**返回**

- `data.items`：`[{ "tag": "bestsellers", "count": 12 }, ...]`
  - `count` 表示包含该标签的已发布商品数量

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/tags?limit=20' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 3) 获取商品（SPU）详情

**接口**

- `GET /api/v1/mini-app/products/{id}`

**说明**

- 仅能访问 `published` 的 SPU；未发布将按 404 处理。
- 返回字段精简，包含 `categoryId/categoryPath/tags`，便于前端展示与联动。
- 电商化字段：`coverUrl/subtitle/description/priceLabel/skuCount` 便于直接渲染“商品详情页”。

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/{spuId}' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 3.1) 获取商品详情（含规格与 SKU 映射，用于规格选择）

当商品存在多规格（如颜色/尺码）并需要在小程序侧做“选项禁用态 + 匹配唯一 SKU”时，推荐调用此接口一次取齐所需信息，避免前端再拼装。

**接口**

- `GET /api/v1/mini-app/products/{id}/detail`

**返回结构要点**

- `data.spu`：与 `GET /products/{id}` 类似的 SPU 详情（含封面与价格聚合字段）。
- `data.spec.groups[]`：规格维度与取值（按 `sortOrder` 排序；每个 group 带 `code/name/required/options`）。
- `data.skus[]`：SKU 列表（仅 `published`），每条包含：
  - `id`：下单/加购应使用的 SKU ID
  - `specSignature`：稳定签名（如 `color=blue|size=m`）
  - `spec`：`{ [groupCode]: optionCode }`，用于前端快速匹配与禁用态计算

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/{spuId}/detail' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 4) 获取商品下的 SKU 列表

**接口**

- `GET /api/v1/mini-app/products/{id}/skus`

**查询参数**

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | number | 页码 |
| `pageSize` | number | 每页数量 |

**注意**

- mini-app 只允许读取 `published` SKU；`status` 传入非 `published` 将返回 400。
- 电商化字段：每个 SKU 会携带 `imageUrl`（主图）以及可选的 `price/currency`（来自 `default_values` 的价格约定）。

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/{spuId}/skus?page=1&pageSize=50' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 5) 获取订阅计划（Subscription Plans）

**接口**

- `GET /api/v1/mini-app/products/{id}/plans`

**说明**

- 仅返回 `status=active` 的订阅计划。
- 用于订阅型商品在小程序详情页展示“月订阅/年订阅”等选项。

**示例**

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/{spuId}/plans' \
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

## 常见问题

### 401：tenant context missing

- 确认携带 `X-Tenant-UUID`（或 `tenant_uuid`）且为合法 UUID。
- 若在 PowerX 宿主内调用，可确认宿主是否注入了 tenant 上下文。

### 400：only published products are accessible via mini-app

- mini-app 只读接口仅面向已发布数据；请先在管理端完成发布流程。

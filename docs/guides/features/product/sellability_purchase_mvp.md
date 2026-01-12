# “可上线购买”最短闭环：渠道可见性 + 可售性聚合（价格/库存/状态）

目标：以**最短闭环**实现“能开始做购买”，优先打通硬门槛（是否可售/可下单），再逐步补齐合规与内容质量；避免一开始就把渠道/内容/素材一次性做全导致上线延期。

## 优先级建议（按最短闭环）

### 1) 价格 + 库存（硬门槛）

- **Mini-app 侧**：只展示能命中对外价 + 有可售库存的 SKU。
  - 两种策略二选一（建议先选一种并保持一致）：
    - **隐藏策略**：不可售 SKU 不展示（列表/规格选项都不出现）。
    - **置灰策略**：展示但不可下单（禁用按钮 + 可解释原因提示）。
- **后端**：提供“聚合可售性”结果，返回 `sellable` + `reasons[]`（以及价格与可用库存），避免前端散落判断逻辑。

### 2) SKU/渠道可见性（决定“能不能卖给这个渠道/这个时间”）

- 基础门槛：`SPU published` + `SKU online/published`
- 渠道关键：`product_spu_channels` 的可售窗口/审核态
- **先做最小字段集**即可跑通：
  - `enabled`
  - `availability`（可售窗口策略：always / scheduled 等）
  - `publishAt`
  - `withdrawAt`
  - `status`（审核/发布态：draft/reviewing/approved/rejected/published…以实际枚举为准）

### 3) 交易合规（决定“能不能下单”）

- 运费/配送可达：至少 1 个默认模板 + 可达区域
- 退换政策/发票税务：允许为空，但在下单时**强校验**或提供默认值（不要阻塞可售性闭环）

### 4) 内容/素材/类目/法规（上架质量与平台风控）

- 建议先做“软校验 + 降级”：
  - 缺失则降级展示/不可上架
  - 不建议在“能下单”闭环前做成硬阻断

## 统一口径：什么是“可售性（Sellability）”

**可售性**是面向前端/mini-app 的聚合判断结果，用于回答：

- 这个 SKU 在**指定渠道**、**指定时间**、**指定区域/语言（可选）**下，是否允许下单？
- 如果不允许，原因是什么（可解释、可展示给用户）？

### 推荐返回结构（面向前端）

> 具体字段以最终合同为准；此处为建议的最小闭环字段集。

- `sellable`: boolean
- `reasons`: string[]（空数组表示可售）
- `price`: `{ amount: number, currency: string }`（命中对外价）
- `availableQty`: number（可下单可用库存，通常为 `available_qty - locked_qty`）

### 推荐原因码（reasons[]）

| code | 含义（面向用户可解释） |
| --- | --- |
| `SPU_NOT_PUBLISHED` | 商品未发布 |
| `SKU_NOT_ONLINE` | SKU 未上架/不可售 |
| `NO_PUBLIC_PRICE` | 未配置对外售价 |
| `OUT_OF_STOCK` | 库存不足 |
| `CHANNEL_DISABLED` | 该渠道未启用售卖 |
| `NOT_IN_AVAILABILITY_WINDOW` | 不在渠道可售时间窗口内 |
| `CHANNEL_STATUS_BLOCKED` | 渠道审核/发布状态不允许售卖 |
| `UNKNOWN` | 未知错误（建议用于兜底并记录日志） |

## 推荐接口形态（MVP）

为避免前端在列表、详情、规格选择、下单按钮处重复判断，建议提供一个“聚合可售性”接口（或在列表接口中可选附带同等信息）。

- **已提供**：`GET /api/v1/mini-app/products/{spuId}/sellability?channel=xxx&locale=zh-CN`
  - 返回：该 SPU 下可售 SKU 列表（或所有 SKU 的 sellable 判定）
  - 优点：mini-app 可一次拿到“可下单 SKU 集合 + 不可售原因”，用于规格选项禁用态与按钮禁用

> 若你的交互是“先选规格再下单”，也可在后续扩展为 `GET /api/v1/mini-app/skus/{skuId}/sellability?...` 以减少一次性返回数据量。

可选增强（列表置灰/过滤）：

- `GET /api/v1/mini-app/products?includeSellability=1&channel=xxx`：返回每条 SPU 的可售性摘要（`sellability` 字段）。

## Mini-app 落地建议（最小闭环）

- 列表页：
  - 只展示 `sellable=true` 的 SKU/商品，或展示但禁用下单（两种策略二选一）
  - 将 `reasons[]` 映射为短提示（如“缺货/未上架/暂不可售”）
- 详情页：
  - 规格选择时：基于 `sellability` 禁用不可选组合，避免选中后才提示失败
- 下单按钮：以 `sellable` 为唯一门槛；点击不可下单时弹出 reasons 解释

## 验收步骤（建议）

1) 准备数据（管理端或 DB）：
- 同一 SPU 下至少 2 个 SKU：一个“有价+有库存”，一个“无价或缺货”
- 为该 SPU 配置 `product_spu_channels`：`channel=official`、`availability=published`、`audit_state=approved`

2) 调用接口：

```bash
curl -sS 'http://127.0.0.1:8086/api/v1/mini-app/products/{spuId}/sellability?channel=official&locale=zh-CN' \\
  -H 'X-Tenant-UUID: 00000000-0000-0000-0000-000000000001'
```

3) 预期：
- 可售 SKU：`sellable=true` 且 `reasons=[]`，并返回 `price` 与 `availableQty>0`
- 不可售 SKU：`sellable=false` 且 `reasons` 至少包含 `NO_PUBLIC_PRICE` 或 `OUT_OF_STOCK`

## 相关文档

- mini-app 开放接口：`docs/guides/features/product/miniapp_open_api.md`
- 常规上架流程：`docs/guides/features/product/standard_onboarding.md`

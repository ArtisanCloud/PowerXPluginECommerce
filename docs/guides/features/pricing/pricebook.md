# Tutorial — 价格手册（Pricebook）（Phase 1）

> 本文由 `specs/005-pricing-pricebook/quickstart.md` 迁移而来，作为可执行的验收/冒烟指导。

## 1. 建品到出价：推荐操作顺序（UI 为主）

这份文档的目标不是“接口说明”，而是让你按顺序把一件商品从 0 建出来，并最终在小程序看到价格。

> 关键依赖：**价格手册条目是“给 SKU 定价”**，因此必须先有 SKU；而 SKU 又依赖 SPU 的规格组合定义。

### 1.1 先做什么后做什么（Checklist）

1. （可选）准备类目：`商品中心 > 类目`
2. 创建 SPU：`商品中心 > 商品（SPU） > 新建`
3. 配置规格（spec，颜色/尺码等）：在 SPU 编辑页配置（见 1.2）
4. 生成/关联 SKU：在 SPU 编辑页的 **关联 SKU / 批量生成**
5. 发布 SPU（使其对外可见）：在 SPU 编辑页点击 **发布**
6. 准备价格手册（Pricebook）：
   - 使用系统 `base`（seed 会自动写入并发布一个版本），或
   - 新建自定义价格手册：`定价中心 > 价格手册`
7. 在某个 `draft` 版本写入条目（给 SKU 定价）：`定价中心 > 价格手册 > 编辑条目`
8. 发布（Publish）价格手册版本（变为 `active`）：`定价中心 > 价格手册 > 发布`
9. 验证：
   - 管理端 SKU/SPU 页面能展示价格（来自价格手册聚合/兜底）
   - 小程序 `/api/v1/mini-app/products` 出现 `priceLabel`

### 1.2 规格（spec）到底在哪里配置？

当前系统里存在两类“看起来像规格/属性”的东西，容易混淆：

- **SPU 维度的变体规格（本功能需要的 spec）**：用于生成 SKU 组合（颜色/尺码……）
  - 数据表：`product_spec_groups / product_spec_options`（见 4.1）
  - 后端接口：`GET/PUT /api/v1/admin/product/spus/:id/spec-groups`
  - 推荐 UI 入口：**SPU 编辑页** 的「规格定义」面板（与“关联 SKU / 批量生成”同一处）
- **属性和规格菜单里的“属性/规格”**：目前是属性体系（attributes）的占位页面，未对接 `product_spec_*`
  - 现状：可能显示 `No data`，不影响 SKU 变体与价格手册链路
  - 建议：该菜单应更名/隐藏，避免用户误以为在这里配置变体规格

### 1.3 依赖关系（用一句话说明“为什么要按这个顺序”）

- 没有 SKU → 不能给价格手册写条目（`pricebook_items.sku_id` 为空无意义）
- 没有 spec → SKU 无法批量生成或组合不稳定（`spec_signature` 无法保证去重）
- 价格手册版本不 publish → `draft` 不参与查价，小程序/订单仍会显示无价或仅命中 `base` 兜底
- SKU 状态不对外可见（例如一直是 `draft`）→ 小程序聚合会过滤掉，最终 `priceLabel=¥--`

### 1.4 Prerequisites

- Go 1.24、PostgreSQL ≥ 13
- 已配置并启动后端（`make dev`），且能登录 `web-admin`

## 2. 价格手册（Pricebook）规则（把 draft/publish 讲清楚）

这里的 **Draft** 指的是“价目表版本（Pricebook Version）”的状态，不是价目表主档（Pricebook）的状态：

### 2.1 主档状态 vs 版本状态（一定要区分）

- **Pricebook 主档（pricebooks.status）**
  - `active`：主档可用；只有主档为 `active` 才可能参与对外查价候选。
  - `archived`：主档停用；即使存在 `active` 版本，也不会被查价使用。
- **Pricebook Version（pricebook_versions.state）**
  - `draft`：草稿版本，可编辑（可批量 upsert 条目），**不会参与对外查价**。
  - `active`：已发布/生效版本，会参与查价候选版本选择（按有效期命中）。
  - `archived / expired`：已下线/已过期版本，不再参与对外查价。

因此：

- **是否一定要 publish？**  
  如果你希望“这个价目表的价格”被查价/订单/渠道同步等消费者使用（即对外生效），就需要把某个版本从 `draft` **publish 成 `active`**。
- **商品是否一定要 publish 才能“有价格”？**  
  不一定。系统会确保每个租户都有一个系统托管的 `base` 基础价目表（`code=base`）且带有 `active` 版本，用于兜底；但你自定义的价目表（例如某渠道价）只有在版本 `active` 后才会被命中。

### 2.2 可编辑性规则（你在业务上“能不能改”）

- 只有 `draft` 版本允许写入/更新条目（`PUT .../items`）；非 `draft` 会被拒绝。
- `publish` 只能对 `draft` 版本执行；发布后变为 `active`。
- `archive` 会将版本置为 `archived`，并写入 `expires_at`（如果原先没有失效时间或还未到期，会用“现在”作为失效时间）。

### 2.3 查价命中规则（哪些会被用来算价）

查价只会从满足以下条件的“候选版本”里选一个最佳命中：

- `pricebooks.status = active`
- `pricebook_versions.state = active`
- `effective_at <= as_of < expires_at`（`expires_at` 为空表示永久）
- 币种必须一致（`pricebooks.currency = request.currency`）

多候选选择规则（确定性）：

1. **具体度优先**：scope 维度越多越具体（例如“channel+customer_group” > “channel” > “无 scope 全量适用”）。
2. **最近发布优先**：`published_at` 越新越优先。
3. **版本号兜底**：版本号更大优先（再兜底用 id 排序保证稳定）。

Scope 匹配规则：

- 价目表**没有任何 scope**：视为“全量适用”，可命中。
- 价目表**配置了 scope**：请求中必须提供该 scope 涉及的所有维度（channel/customer_group/supplier），且值必须在允许集合里；缺任一维度或不在集合内则不命中。

回退（Fallback）规则（统一口径）：

- **仅当**“已命中一个非 base 的价目表版本，但该版本缺少该 SKU 条目”时，才允许回退到 **同币种** `base` 价目表条目。
- 范围不命中/币种不匹配/无候选版本/缺价格字段等场景 **不回退**，直接返回无价（带原因码）。

### 2.4 前端状态展示建议（用于 UI 接入时对齐后端）

当 UI 接入后端数据时，建议按以下规则映射展示状态（示例）：

- 版本 `state=draft` → “草稿”
- 版本 `state=active` 且 `effective_at > now` → “待生效”
- 版本 `state=active` 且 `effective_at <= now < expires_at(或为空)` → “生效中”
- 版本 `state in (archived, expired)` 或 `expires_at <= now` → “已失效”

## 3. UI 操作指引（按顺序可走通）

> 说明（2026-01）：`web-admin` 的 `价格手册` 页面已接入后端 API，可通过 UI 完成“创建 → 写入条目 → 发布 → 验证”。本文保留 API 方式用于自动化/联调。

### 3.1 创建/编辑 SPU（并发布）

- SPU 列表：`/product/spus`
- 新建 SPU：`/product/spus/create`
- 编辑 SPU：`/product/spus/edit/{spuId}`
- 发布 SPU：在 SPU 编辑页点击 **发布**

发布后你应该看到：

- SPU `status=published`
- 其下关联的 SKU（若已生成）应为对外可见状态（例如 `online`），否则小程序不会统计/展示价格

### 3.2 配置 spec（变体规格）与生成 SKU

入口：SPU 编辑页 `关联 SKU` 面板。

1. 打开：`/product/spus/edit/{spuId}`
2. 点击 **关联 SKU**
3. 切到 **批量生成**
4. 选择规格值组合 → 生成 SKU（系统会预览 `SKU Code` 并避免生成已存在组合）

验证：

- SKU 列表：`/product/skus`（应能看到 `SKU Code / SPU 名称 / 规格组合`）

### 3.3 价格手册（Pricebook）维护与发布

路径：`/pricing/pricebooks`（对应 `web-admin/app/pages/pricing/pricebooks.vue`）

#### 3.3.1 创建价格手册（Pricebook）

1. 进入 `价格手册`：`/pricing/pricebooks`
2. 点击 **新增价格手册**
3. 填写 `Code / 名称 / 类型(sales/purchase) / 币种`
4. 提交后会生成一个 v1 的 `draft` 版本（后端写入 `current_version_id`）

> 注意：系统托管的 `base` 价目表（`code=base`）会显示“系统”标识，且在 UI 中禁用编辑/删除/创建版本。

#### 3.3.2 写入条目（仅 draft 版本）

1. 在列表中点击某个价格手册的 **查看**
2. 切换到 **版本与发布**
3. 选择一个 `state=draft` 的版本（例如 `v1 · draft`）
4. 点击 **编辑条目**
5. 按 SKU 填写价格字段（单位为“最小货币单位”，例如 19900 表示 199.00）
6. 点击 **保存**

> 说明：保存使用 upsert（同 SKU 会更新；不会删除旧条目）。若需要“删除条目”，需补充专用删除接口/交互（当前版本未提供）。

#### 3.3.3 发布版本（Publish）

1. 在 **版本与发布** 选择一个 `draft` 版本
2. 点击 **发布**
3. 可选填写 note、生效/失效时间
4. 发布成功后该版本变为 `active` 并参与对外查价

### 3.4 验证（管理端 + 小程序）

管理端验证：

- SKU 列表：`/product/skus`（应展示 `salePrice/currency`，优先来自 `base` 价格手册的 current version 条目）
- SPU 详情的关联 SKU：应展示每个 SKU 的价格（最终以价格手册命中价为准）

小程序验证：

- SPU 列表：`GET /api/v1/mini-app/products`（one_time 商品应不再是 `priceLabel:"¥--"`）
- SKU 列表：`GET /api/v1/mini-app/products/:id/skus`（只返回对外可见状态的 SKU）

## 4. 数据库表关系（SPU / Spec / SKU / Pricebook）

> 目标：把「SPU / spec / SKU / 价格手册」在数据库里的关系讲清楚，便于你对齐 UI、API、seed、以及小程序最终展示价格的链路。

### 4.1 表与主外键（核心链路）

以下均为租户表：每张表都有 `tenant_uuid`（并通过 RLS/tenant context 约束可见性）。

- `product_spus`（SPU 主档）
  - 主键：`id`
  - 关键字段：`code / name / type / status / current_version_id`
  - 关系：
    - `current_version_id -> product_spu_versions.id`（指向“当前版本快照”，用于审批/发布流的读取）
- `product_spu_versions`（SPU 版本快照）
  - 主键：`id`
  - 外键：`spu_id -> product_spus.id`
  - 关键字段：`version_number / status / payload(jsonb)`
  - 说明：`payload` 里包含“关联 SKU”的业务结构（例如 `payload.skus[]`），用于 SPU 维度的编辑/回滚/审批。
- `product_spec_groups`（规格组/规格维度）
  - 主键：`id`
  - 外键：`spu_id -> product_spus.id`
  - 关键字段：`code / name / required / sort_order`
  - 示例：颜色、尺码、容量……
- `product_spec_options`（规格值/可选项）
  - 主键：`id`
  - 外键：`group_id -> product_spec_groups.id`，同时也保存 `spu_id`
  - 关键字段：`code / name / sort_order / meta(jsonb)`
  - 示例：红/蓝、S/M/L……
- `product_skus`（SKU 实体表：最终用于下单/渠道/小程序查询的“可售卖单元”）
  - 主键：`id`
  - 外键：`spu_id -> product_spus.id`
  - 关键字段：`sku_code / status / spec_values(jsonb) / spec_signature / default_values(jsonb)`
  - 说明：
    - `spec_values` 保存的是“规格组合”（数组），每个元素形如 `{spec_id, value_id, ...}`；其中 `spec_id` 指向 `product_spec_groups.id`，`value_id` 指向 `product_spec_options.id`。
    - `spec_signature` 是规格组合的稳定签名（用于去重，同一 SPU 下同一组合只允许一个 SKU）。

价格手册相关表：

- `pricebooks`（价格手册主档）
  - 主键：`id`
  - 关键字段：`code / name / type / currency / status / current_version_id`
  - 关系：
    - `current_version_id -> pricebook_versions.id`（指向“当前生效版本”）
- `pricebook_versions`（价格手册版本）
  - 主键：`id`
  - 外键：`pricebook_id -> pricebooks.id`
  - 关键字段：`version(int) / state(draft|active|archived|expired) / effective_at / expires_at / published_at`
- `pricebook_items`（价格条目：对某个 SKU 的定价）
  - 主键：`id`
  - 外键：`version_id -> pricebook_versions.id`，`pricebook_id -> pricebooks.id`，`sku_id -> product_skus.id`
  - 关键字段：`*_amount_minor`（如 `base_amount_minor / sale_amount_minor / tax_included`）
  - 唯一约束：`(tenant_uuid, version_id, sku_id)`（同一版本同一 SKU 只有一条价格）
- `pricebook_scopes`（价格手册适用范围：scope）
  - 外键：`pricebook_id -> pricebooks.id`
  - 关键字段：`dimension(channel/customer_group/supplier) + dimension_id`

### 4.2 一句话解释“它们怎么串起来”

- **SPU 定义可选规格（spec）**：`product_spec_groups/options` 挂在 SPU 上。
- **SKU 是规格组合实例**：`product_skus` 用 `spec_values` 保存“我选了哪些 spec option”。
- **价格手册给 SKU 定价**：`pricebook_items.sku_id -> product_skus.id`。
- **小程序展示价格聚合**：通常按 SPU 聚合其“可见 SKU”的价格（SKU 必须是对外可见状态，例如 `online/ready/published`），并优先读取 `base` 价目表或命中价目表的 `active` 版本条目。

### 4.3 为什么 UI 要显示 SKU Code / 规格组合，而不是 UUID

数据库里用 UUID 做主键是为了稳定引用与分布式写入；但业务操作应该围绕：

- SPU：`code + name`
- SKU：`sku_code + 规格组合（由 spec group/option 的 name/code 解析出来）`
- Pricebook：`code + name + 当前版本号(vN) + state`

因此 UI（以及文档中的 tutorial 步骤）都会以这些“可读业务字段”为主，UUID 仅用于系统内部引用与排障。

## 5. 依赖 & 权限（RBAC）

管理端（`/api/v1/admin/pricing/**`）需要租户上下文与 RBAC 权限，核心资源为：

- `com.powerx.plugin.ecommerce:pricing.pricebook`
  - `read`：列表
  - `manage`：创建/更新/删除、创建版本、写入条目
  - `publish`：发布/下线版本

查价接口（`/v1/pricing/query` 或 `/api/v1/pricing/query`）是租户域接口：需要租户上下文，但不属于管理端 RBAC 资源。

## 6. Backend Development Workflow

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

## 7. API 操作指引（可从头到尾跑通）

> 以下以本地默认端口为例，实际请按你的 `POWERX_BIND_ADDR` 调整（本仓库默认 `:8091`）。

建议先准备一个基址变量，避免文档里端口/前缀不一致：

```bash
BASE_URL=${BASE_URL:-http://localhost:8091}
```

> 注意：管理端接口返回使用统一 Envelope（`contracts.ResponseXXX`），真实数据在 `.data` 字段里。

### Create Pricebook

```bash
curl -sS -X POST "$BASE_URL/api/v1/admin/pricing/pricebooks" \
  -H 'Content-Type: application/json' \
  -d '{"code":"tmall_standard","name":"天猫标准价目表","type":"sales","currency":"CNY"}'
```

获取后续调用需要的 `pricebookId` / `versionId`（创建价目表会自动生成 v1 draft，并写入 `current_version_id`）：

```bash
resp=$(curl -sS -X POST "$BASE_URL/api/v1/admin/pricing/pricebooks" \
  -H 'Content-Type: application/json' \
  -d '{"code":"tmall_standard","name":"天猫标准价目表","type":"sales","currency":"CNY"}')

pricebookId=$(echo "$resp" | jq -r '.data.id')
versionId=$(echo "$resp" | jq -r '.data.current_version_id')

echo "pricebookId=$pricebookId"
echo "versionId=$versionId"
```

> 说明：每个租户会自动拥有一个系统级基础价目表 `code=base`，用于兜底定价。  
> `base` 不允许通过 Create API 创建/归档/删除（系统托管）。

### Upsert Base Pricebook (batch)

> 如需对所有已存在租户一次性补齐/修复 `base`，可使用脚本（可重复执行，幂等）。

```bash
# 预演（不写库）
cd backend && go run ./cmd/database pricing-base-upsert -dry-run

# 执行（默认 currency=USD，可覆盖）
cd backend && go run ./cmd/database pricing-base-upsert -currency=CNY
```

### Upsert Items (draft only)

```bash
curl -sS -X PUT "$BASE_URL/api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items" \
  -H 'Content-Type: application/json' \
  -d '{"items":[{"sku_id":"{skuId}","base_amount_minor":19900}]}'
```

### Create Version (optional)

> 如果不想发布“创建价目表时自动生成的 v1 draft”，可以先创建一个新版本（draft），此接口返回的 `data.id` 就是新的 `versionId`。

```bash
resp=$(curl -sS -X POST "$BASE_URL/api/v1/admin/pricing/pricebooks/{pricebookId}/versions" \
  -H 'Content-Type: application/json' \
  -d '{}')

versionId=$(echo "$resp" | jq -r '.data.id')
echo "new versionId=$versionId"
```

### Publish Version

```bash
curl -sS -X POST "$BASE_URL/api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/publish" \
  -H 'Content-Type: application/json' \
  -d '{"note":"publish v1"}'
```

### List Versions / List Items（用于验证）

```bash
# 版本列表
curl -sS "$BASE_URL/api/v1/admin/pricing/pricebooks/{pricebookId}/versions"

# 条目列表（默认 page_size=200，可调大到 1000）
curl -sS "$BASE_URL/api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items?page=1&page_size=200"
```

### Pricing Query

```bash
curl -sS -X POST "$BASE_URL/v1/pricing/query" \
  -H 'Content-Type: application/json' \
  -d '{"sku_id":"{skuId}","currency":"CNY","as_of":"2026-01-01T00:00:00Z"}'
```

## Contract Reference

- OpenAPI：`specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml`

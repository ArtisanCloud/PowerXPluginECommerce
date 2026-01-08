# 价目表（价格手册）PRD

> 覆盖页面：`web-admin/app/pages/pricing/pricebooks.vue` 以及在 SPU/SKU、渠道上架、供应商管理中引用 pricebook 的表单/组件。Pricebook 是定价中心的核心主数据，负责将具体价格与 SKU、供应商、渠道/门店、客户组绑定。

## 1. 背景与目标
- 不同渠道、供应商、门店、客户需要不同价格策略；必须通过可版本化的价目表统一管理，确保前后端（订单、渠道、结算）价格一致。
- 现有页面展示基础列表，需要补足多维绑定、版本管理、审批、导入导出、优先级策略。
- **与其他定价模块关系**：价目表是“静态基础价”；阶梯价/客户价在计算时对基础价做条件化调整；促销/优惠券/礼品卡属于营销层，通常叠加在定价结果之后。文档中各模块定位应明确，避免功能重叠。

**目标**
1. 为每个 SKU/商品提供一个或多个价格手册，支持按照渠道/门店、供应商、客户组、币种划分。
2. 支持版本管理（草稿/生效/过期）、审批、导入导出及批量更新。
3. 与上架、采购、结算等模块联动，保证价格变更影响可追踪、可回滚。

## 2. 角色与价值
| 角色 | 价值 |
| --- | --- |
| 定价经理 | 统一维护各渠道/客户价目，控制审批与版本、避免价格冲突 |
| 渠道/门店运营 | 获取适用的价目策略，推送到门店或第三方渠道 |
| 采购/供应链 | 利用采购价目表评估供应商报价、生成采购单 |
| 销售/客服 | 查询客户专属价、解释价格差异 |
| 财务/结算 | 拿到生效价目，核对账单、发票、分账 |

## 3. 信息架构
1. **价目表列表**
   - 列字段：Name、类型（销售/采购/渠道/门店/供应商）、币种、适用范围（SKU、类目、客户组、渠道）、版本、状态、有效期、负责人。
   - 筛选：类型、渠道、客户组、供应商、状态、更新人、创建时间。
   - 批量操作：导出、复制、新建版本、发布/下线。
   - **命名示例**：
     - `Base Pricebook / 基础价格手册`：所有 SKU 默认价（RRP、成本、最低价），系统初始化即自动创建且不可删除；
     - `华东区域销售价`：覆盖华东门店/仓的渠道价；
     - `JD-旗舰店价`、`TMall-旗舰店价`：特定线上渠道，继承基础价，再在部分 SKU 上覆盖；
     - `门店-上海静安店`：单店价；
     - `双十一-活动价`：短期促销价目，供特定营销活动使用。
2. **价目表详情**
   - 基本信息：名称、描述、用途、币种、税价策略、审批状态。
   - 适用范围：绑定 SKU/类目/商品集；可选客户组、渠道/门店、供应商。
   - 价格明细：表格（SKU、基准价、渠道价、采购价、最低/最高价、税率、折扣、有效时间段）。
   - 版本/变更记录：版本号、创建人、变更 diff、审批记录。
   - 依赖：关联的规则/合同/促销，以便冲突检查。
   - **继承/派生（可选能力）**：
     - 支持指定“父价目表”（如基础价），派生价目表在查询时先读取父价，再按差异表覆盖；
     - 适用于“基础价 + 渠道差异”“全国价 + 区域调整”场景，减少重复条目；
     - 在 UI 中展示继承链，并支持在派生版中查看/覆盖父级条目。
3. **价格条目编辑**
   - 支持单行编辑、批量编辑（勾选多行或导入 Excel）。
   - 允许指定优先级（当 SKU 同时落入多个 pricebook 时的决策：客户专属价 > 渠道价 > 基准价）。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 系统基础价目表 | 安装/seed 阶段自动生成 `Base Pricebook`（包含所有 SKU 的基准价），不可删除，仅允许版本化编辑 |
| 新建/复制价目表 | 可从模板、历史版本复制；初始化适用范围与默认价格 |
| 版本管理 | 草稿 → 审批 → 生效；支持回滚、定时生效、手动下线 |
| SKU 关联 | 选择单个 SKU、SKU 列表、按类目/属性批量关联；可导入 CSV 关联多条 |
| 多维适用 | 支持设置供应商、门店、渠道、客户组、地区、币种；生成 pricebook key（用于 API 查询） |
| 导入/导出 | Excel/CSV 模板；导入可更新现有条目或添加新条目 |
| 审批 & 特批 | 价目更新需走审批流；超出阈值（如低于成本）自动触发特批 |
| 冲突检测 | 检测同 SKU 在同维度是否存在多个价目冲突，提供提示或阻止发布 |
| API 查询 | 一期以“统一查价”接口为主：`POST /v1/pricing/query`（主路径）与 `POST /api/v1/pricing/query`（兼容别名）；管理端配置接口见 11.5 |
| 审计 | 记录每次变更、审批、发布，支持 diff 查看 |

## 5. 关键流程
1. **创建/发布价目表**
   1. 选择模板 → 填写基础信息 → 选择维度（SKU、渠道、供应商等）。
   2. 录入价格（手动或导入）。
   3. 保存草稿；提交审批 → 审批通过后设置生效时间 → 发布。
2. **价格更新**
   1. 在价目表详情中新建版本或批量导入更新。
   2. 系统对比旧版差异，产出变更摘要 → 发起审批。
   3. 生效后通知关联模块（上架、订单、结算）。
3. **冲突处理**
   1. 发布前系统检查 SKU 是否同时存在多个生效价目冲突。
   2. 如果冲突，根据优先级（客户价 > 渠道价 > 基准价）；如仍冲突，要求处理。
4. **供应商/门店价目**
   1. 可为供应商定义采购价目，供采购模块使用。
   2. 可按门店/区域定义零售价，供门店 POS/渠道同步。

## 6. 数据 & API
- **表**：`pricebooks`、`pricebook_versions`、`pricebook_scopes`（渠道/客户组/供应商）、`pricebook_items`、`pricebook_audit_logs`。
- **API（一期实现，以 OpenAPI 为准）**：
  - 管理端：`/api/v1/admin/pricing/pricebooks/**`（创建/更新主档、创建版本、发布/下线、条目 upsert）
  - 业务查价：`POST /v1/pricing/query`（主路径）+ `POST /api/v1/pricing/query`（兼容别名）

---

## 11. 开发设计（可落地规格 / Phase 1）
> 本章节补齐“可以直接开干”的技术规格：**表结构、状态机、接口契约、查询匹配规则、错误码**。实现以“Pricebook 主数据 + 多币种 + 版本发布 + 查询 API”形成后端闭环为目标；审批流/规则引擎/冲突模拟/导入导出异步任务放到后续迭代。

### 11.1 路由命名与版本前缀
本插件后端 API 通常挂在统一前缀（例如 `"/api/v1"`），管理端路由在其下使用 `"/admin"` 分组。

- **管理端 Pricebook API（建议）**：`/api/v1/admin/pricing/pricebooks/**`
- **对外查询 API（建议）**：`/v1/pricing/query`（主路径）与 `/api/v1/pricing/query`（兼容别名；鉴权方式按现有 tenant/jwt 中间件对齐）

> 备注：本 PRD 的旧路径 `GET /api/pricing/pricebooks` 等仅作为“概念名称”。实际落地建议统一到 `"/api/v1"` 与 `"/admin"` 规范，以便 RBAC 自动汇总。

### 11.2 数据模型与表结构（PostgreSQL / schema: `powerx_plugin_base`）
#### 11.2.1 金额与币种（一期统一口径）
- **金额存储**：使用 `BIGINT` 的 `amount_minor`（最小货币单位，如分），避免浮点误差。
- **币种**：`currency` 使用 ISO 4217（如 `CNY`、`USD`），统一大写。
- **前端展示**：接口同时返回 `amount_minor`（整数）与 `amount`（字符串，按币种 exponent 格式化），便于 UI 直接展示。

#### 11.2.2 表：`pricebooks`
用于描述“价目表主档”与其当前生效版本指针。

字段（建议）：
- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `code` text NOT NULL（业务唯一键，建议 `base` / `jd_flagship` 等），unique(`tenant_uuid`,`code`)
- `name` text NOT NULL
- `type` text NOT NULL（一期枚举：`sales`/`purchase`；默认 `sales`）
- `currency` text NOT NULL（ISO 4217）
- `description` text NULL
- `status` text NOT NULL（`active`/`archived`）
- `current_version_id` uuid NULL（指向 `pricebook_versions.id`）
- `created_at`,`updated_at`,`deleted_at`

索引建议：
- `idx_pricebooks_tenant_type_currency_status`(`tenant_uuid`,`type`,`currency`,`status`)

#### 11.2.3 表：`pricebook_versions`
用于版本化与发布（草稿→生效/过期），并承载有效期。

字段（建议）：
- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `pricebook_id` uuid NOT NULL, index
- `version` int NOT NULL（从 1 递增），unique(`tenant_uuid`,`pricebook_id`,`version`)
- `state` text NOT NULL（`draft`/`active`/`archived`/`expired`）
- `effective_at` timestamptz NOT NULL（默认“立即生效”）
- `expires_at` timestamptz NULL（空表示不自动失效）
- `published_at` timestamptz NULL
- `published_by` text NULL（用户 id）
- `note` text NULL（发布说明）
- `created_at`,`updated_at`,`deleted_at`

索引建议：
- `idx_pricebook_versions_tenant_pricebook_state`(`tenant_uuid`,`pricebook_id`,`state`)
- `idx_pricebook_versions_effective`(`tenant_uuid`,`effective_at`,`expires_at`)

#### 11.2.4 表：`pricebook_scopes`
用于表达“适用范围维度”。一期建议仅落地三类维度：`channel`、`customer_group`、`supplier`（门店/地区等后续扩展）。

字段（建议）：
- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `pricebook_id` uuid NOT NULL, index
- `dimension` text NOT NULL（枚举：`channel`/`customer_group`/`supplier`）
- `dimension_id` uuid NOT NULL（对应维度主数据的 id）
- unique(`tenant_uuid`,`pricebook_id`,`dimension`,`dimension_id`)
- `created_at`,`updated_at`,`deleted_at`

索引建议：
- `idx_pricebook_scopes_lookup`(`tenant_uuid`,`dimension`,`dimension_id`)

#### 11.2.5 表：`pricebook_items`
用于存放版本下的 SKU 价格明细。

字段（建议）：
- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `pricebook_id` uuid NOT NULL, index
- `version_id` uuid NOT NULL, index
- `sku_id` uuid NOT NULL, index
- `base_amount_minor` bigint NULL（基准价）
- `sale_amount_minor` bigint NULL（销售价/渠道价；一期最终输出优先使用它）
- `msrp_amount_minor` bigint NULL（建议零售价）
- `cost_amount_minor` bigint NULL（成本价）
- `min_amount_minor` bigint NULL（最低价）
- `max_amount_minor` bigint NULL（最高价）
- `tax_included` bool NOT NULL default false（税价策略：含税/未税；一期可先固定 false）
- `meta` jsonb NULL（扩展字段）
- unique(`tenant_uuid`,`version_id`,`sku_id`)
- `created_at`,`updated_at`,`deleted_at`

索引建议：
- `idx_pricebook_items_tenant_version_sku`(`tenant_uuid`,`version_id`,`sku_id`)
- `idx_pricebook_items_tenant_pricebook_sku`(`tenant_uuid`,`pricebook_id`,`sku_id`)

#### 11.2.6 表：`pricebook_audit_logs`（一期最小化）
一期仅要求“可追溯发布与关键变更”；细粒度 diff 可后续增强。

字段（建议）：
- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `resource_type` text NOT NULL（`pricebook`/`pricebook_version`/`pricebook_item`）
- `resource_id` uuid NOT NULL, index
- `action` text NOT NULL（`create`/`update`/`publish`/`archive`）
- `actor` text NULL
- `payload` jsonb NULL（可存摘要：变更字段、条目数等）
- `created_at`

### 11.3 状态机（一期约束）
#### 11.3.1 Pricebook（`pricebooks.status`）
- `active`：可用于查询与发布版本（若存在 active 版本）
- `archived`：不可用于查询；不可再发布新版本（允许查看历史）

#### 11.3.2 Version（`pricebook_versions.state`）
- `draft`：可编辑条目；不可被 `pricing/query` 使用
- `active`：满足 `effective_at <= as_of < expires_at(如有)` 时可用于查询
- `expired`：到期自动转换（按查询逻辑判定；可异步任务落状态）
- `archived`：手动下线/撤回

强约束（一期必须明确）：
- **同一 pricebook 任一时刻仅允许 1 个 `active` 版本在有效期内命中**。发布新版本需要自动下线旧版本（设为 `archived` 或填充 `expires_at`）。
- 发布为**幂等**：重复 publish 同一版本不应产生新版本或重复状态抖动。

### 11.4 查询匹配规则（`POST /pricing/query` 的核心算法）
输入上下文（一期）：
- `sku_id`（必填）
- `currency`（必填）
- `channel_id`/`customer_group_id`/`supplier_id`（可选）
- `as_of`（可选，默认当前时间）

匹配规则（一期定义为“维度 AND，值 OR”）：
1. 过滤候选：`pricebooks.status=active` 且 `pricebooks.currency = currency` 且 `type=sales`（如需采购价另开查询或加参数）。
2. 过滤版本：只看候选 pricebook 的 `pricebook_versions.state=active` 且满足有效期（`effective_at <= as_of` 且（`expires_at IS NULL OR as_of < expires_at`））。
3. scope 适配（对每个维度独立判断）：
   - 如果某 pricebook 在某个 `dimension` 下 **没有任何 scope 记录**，表示该维度“全量适用”；
   - 如果存在 scope 记录，则请求里必须带该维度 id，且命中 scope 集合之一。
4. 选择最优 pricebook/version：按“更具体”优先（命中的受限维度越多越优先），再按 `priority`（若一期不加字段则跳过），最后按 `published_at` 新者优先。
5. item 命中：在选定 `version_id` 下查 `pricebook_items` 的 `sku_id`。
6. 回退策略（一期建议固定如下，避免价格空洞）：
   - 若已命中某 pricebook/version，但该版本下无该 SKU 条目，则允许回退到同币种 `code=base` 的 Base Pricebook；
   - 其余场景（范围不命中、币种不匹配、无候选版本等）**不回退**，直接返回“无价”。
   - 一期返回语义：`200` 且 `priced=false`，并给出 `trace.no_price_reason`（而非 `404`）。
7. 输出字段优先级：
   - `sale_amount_minor`（若非空）→ 否则 `base_amount_minor` → 否则 `msrp_amount_minor`；三者都空则视为未配置。

### 11.5 管理端 API 契约（Phase 1 最小集）
> 说明：以下以 `basePath=/api/v1` 为例。返回结构统一使用 `{ "data": ..., "meta": ... }` 或直接返回 DTO，按现有项目惯例二选一；一期建议保持简单，直接返回 DTO。

#### 11.5.1 列表：`GET /api/v1/admin/pricing/pricebooks`
Query：
- `keyword`（按 name/code 模糊）
- `type`（`sales`/`purchase`）
- `currency`
- `status`（`active`/`archived`）
- `page`,`page_size`

返回（一期实现）：
- `200`：`{ items: Pricebook[], meta: { page, page_size, total } }`
- OpenAPI 以 `specs/005-pricing-pricebook/contracts/pricing-pricebooks.openapi.yaml` 为准

#### 11.5.2 新建：`POST /api/v1/admin/pricing/pricebooks`
- 创建 pricebook 主档，并自动创建 `v1` 草稿版本（`state=draft`），并写入 `current_version_id`

#### 11.5.3 更新：`PATCH /api/v1/admin/pricing/pricebooks/{pricebookId}`
- 支持更新 `name/description/status/scopes`
- scopes 为 replace 语义：传入 scopes 则整体替换；不传 scopes 则不变；传入空数组视为“删除 scopes=全量适用”

#### 11.5.4 版本：`POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions`
- 创建新草稿版本（`state=draft`）

#### 11.5.5 条目 upsert：`PUT /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/items`
- 仅允许 `draft` 版本写入；否则返回冲突（`409 VERSION_NOT_EDITABLE`）
- unique(`tenant_uuid`,`version_id`,`sku_id`) 覆盖更新

#### 11.5.6 发布/下线
- `POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/publish`
  - 发布后成为可命中的 `active` 版本，并自动终止旧 active 版本有效期（FR-004A）
- `POST /api/v1/admin/pricing/pricebooks/{pricebookId}/versions/{versionId}/archive`
  - 下线版本，使其不再可命中

#### 11.5.7 统一查价（业务侧）
- `POST /v1/pricing/query`（主路径）
- `POST /api/v1/pricing/query`（兼容别名）
- 返回 `priced=true/false`，并在无价时通过 `trace.no_price_reason` 说明原因；在回退时通过 `trace.fallback` 说明回退类型

Response（示例字段）：
- `items[]`: `{ id, code, name, type, currency, status, current_version_id, updated_at }`
- `page`,`page_size`,`total`

#### 11.5.2 创建：`POST /api/v1/admin/pricing/pricebooks`
Request：
- `code`,`name`,`type`,`currency`,`description`
- `scopes`: `{ channel_ids?:[], customer_group_ids?:[], supplier_ids?:[] }`

行为：
- 创建 pricebook 后自动创建 `version=1` 的 `draft` 版本；`current_version_id` 仍为空，直到 publish。

#### 11.5.3 更新：`PATCH /api/v1/admin/pricing/pricebooks/{id}`
允许更新：
- `name`,`description`,`status`（仅允许 `active→archived`）
- `scopes`（一期允许编辑；若已存在 active 版本，需明确是否立即影响查询：一期建议**允许立即生效**，但要写审计）

#### 11.5.4 版本：`POST /api/v1/admin/pricing/pricebooks/{id}/versions`
创建新版本（draft），从当前 active 版本复制条目（若有），或从 base 复制（可选）。
Request：
- `copy_from_version_id?`（可选）

Response：
- `{ id, pricebook_id, version, state }`

#### 11.5.5 条目批量 upsert：`PUT /api/v1/admin/pricing/pricebooks/{id}/versions/{versionId}/items`
Request：
- `items[]`: `{ sku_id, base_amount_minor?, sale_amount_minor?, msrp_amount_minor?, cost_amount_minor?, min_amount_minor?, max_amount_minor?, tax_included?, meta? }`

约束：
- 仅 `draft` 版本可写入；
- 同一 `sku_id` 在该版本内唯一，重复视为覆盖更新。

#### 11.5.6 发布：`POST /api/v1/admin/pricing/pricebooks/{id}/versions/{versionId}/publish`
Request：
- `effective_at?`（默认 now）
- `expires_at?`
- `note?`

行为（事务内）：
- 将该版本置为 `active`，填 `published_at/published_by`；
- 将同 pricebook 其他 “有效期可命中”的 active 版本置为 `archived` 或设置 `expires_at=effective_at`；
- 更新 `pricebooks.current_version_id = versionId`。

#### 11.5.7 下线：`POST /api/v1/admin/pricing/pricebooks/{id}/versions/{versionId}/archive`
将版本置为 `archived`；若其为 `current_version_id`，则清空或回退到上一 active 版本（一期建议清空，并在 query 走 base 回退）。

### 11.6 查询 API 契约（Phase 1）
#### `POST /v1/pricing/query`（主路径）与 `POST /api/v1/pricing/query`（兼容别名）
Request：
```json
{
  "sku_id": "uuid",
  "currency": "CNY",
  "channel_id": "uuid",
  "customer_group_id": "uuid",
  "supplier_id": "uuid",
  "as_of": "2026-01-07T00:00:00Z"
}
```

Response（有价示例）：
```json
{
  "currency": "CNY",
  "priced": true,
  "source_field": "base",
  "price": { "amount_minor": 19900, "amount": "199.00" },
  "matched": {
    "pricebook_id": "uuid",
    "pricebook_code": "base",
    "version_id": "uuid",
    "version": 3,
    "sku_id": "uuid"
  },
  "fields": {
    "sale_amount_minor": 19900,
    "base_amount_minor": 21900,
    "msrp_amount_minor": 29900,
    "min_amount_minor": 18900,
    "max_amount_minor": 25900
  },
  "trace": {
    "fallback": "none"
  }
}
```

Response（无价示例）：
```json
{
  "currency": "CNY",
  "priced": false,
  "trace": { "fallback": "none", "no_price_reason": "SCOPE_MISMATCH" }
}
```

### 11.7 错误码与错误结构（建议）
统一错误结构（示例）：
```json
{ "error": { "code": "PRICEBOOK_NOT_FOUND", "message": "pricebook not found" } }
```

一期实现错误码（以 OpenAPI 与后端实现为准）：
- `INVALID_ARGUMENT`：缺少必填字段或格式错误
- `PRICEBOOK_NOT_FOUND`：主档不存在
- `VERSION_NOT_FOUND`：版本不存在
- `VERSION_NOT_EDITABLE`：非 draft 版本不可写（如 items upsert / publish）
- `PUBLISH_CONFLICT`：发布冲突（并发 publish 或有效期重叠）

### 11.8 权限（RBAC）落地口径（一期最小）
资源建议：`pricing:pricebook`
- `read`：查看列表/详情/条目
- `manage`：创建/编辑/维护条目/创建版本
- `publish`：发布/下线
- `import`：导入（若一期不做可暂不声明）

### 11.9 与后续模块的兼容点（提前留口）
- 规则引擎（Phase 2）接入点：`pricing/query` 返回结构保留 `trace/components` 扩展位，便于后续追加规则命中信息。
- 审批流（Phase 3）接入点：publish 前置校验可扩展为“审批通过才可 publish”；当前接口先保留 `note` 与 `published_by`。
- 继承/派生（可选）：若未来加入 parent pricebook，可在 `pricebooks` 增加 `parent_id`，查询算法在 item 未命中时按继承链回退（一期不实现，但表结构预留不影响）。

## 7. 权限与审计
- 权限项：`pricebook.read`、`pricebook.manage`、`pricebook.publish`、`pricebook.import`、`pricebook.approval`。
- 所有变更、审批、发布记录 `admin_console_audit_events`，并在价目表页面展示最近操作。
- 发布价目表需双重确认（防止误发布影响全局）。

## 8. KPI
| 指标 | 目标 |
| --- | --- |
| 价目表审批周期 | ≤ 2 个工作日 |
| 价格冲突率 | < 1%（发布阶段即发现并解决） |
| API 响应时延 | < 200ms（缓存） |
| 导入成功率 | ≥ 99% |

## 9. 依赖与风险
- 依赖 SKU/渠道/客户/供应商主数据；需及时同步。
- Pricebook 改动影响广泛，需要通知机制与回滚能力。
- 多币种需考虑汇率更新、舍入规则。
- 导入导出涉及大量数据，需异步任务与错误报告。

## 10. Backlog
- Pricebook 模板库：按行业/渠道提供预置模板。
- 自动价格建议：结合成本、竞争对手、库存生成建议价。
- 渠道实时推送：价目更新自动推送第三方渠道 API。
- 多层价目合并：在查询时动态合并多个价目（如基准价 + 渠道折扣）。

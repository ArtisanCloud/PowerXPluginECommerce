# 促销规则 PRD

> 覆盖页面：`web-admin/app/pages/market/promotions.vue` 与定价中心联动的促销配置组件。促销规则用于配置限时折扣、满减、买赠、捆绑、会员价激励等，与定价/订单/营销系统互通。
> 执行落地以 `web-admin/app/pages/pricing/promotions.vue` 为主入口；`market/promotions.vue` 后续作为营销运营侧聚合入口，引用定价中心的促销规则与效果数据。

## 1. 背景与目标
- 用户需要在电商插件中快速配置多种促销方案，并与定价、优惠券、积分联动。
- 现有页面仅展示卡片 UI，需要明确促销规则模型、审批、冲突管理、渠道同步。

**目标**
1. 构建促销规则库：支持多类型（折扣、满减、买赠、组合包、限时价）。
2. 提供条件+动作模型，与 pricebook/定价规则共享数据（SKU、客户、渠道）。
3. 实现执行状态跟踪、限额控制、冲突检测、渠道推送。

## 2. 角色
| 角色 | 价值 |
| --- | --- |
| 营销/增长 | 设计促销活动、监控效果 |
| 定价经理 | 审核促销是否符合价格策略 |
| 渠道运营 | 推送到站点/门店/第三方渠道 |
| 财务 | 评估成本、控制预算 |

## 3. 信息架构
1. **促销列表**：名称、类型、范围、状态、时间、负责人、限额、渠道。
2. **促销详情/编辑**
   - 基本信息：名称、描述、海报、时间、渠道、活动标签。
   - 条件：SKU/类目、客户组、会员等级、订单金额/数量、渠道、设备、来源、优惠叠加规则。
   - 动作：折扣（金额/比例）、满减、买赠（赠品 SKU）、捆绑价、免费礼品、积分奖励、免邮。
   - 限制：总预算、每用户限购、库存锁定、审批状态。
   - 派发：可关联优惠券、积分、任务。
   - 渠道同步：哪些站点/门店/渠道生效，是否需要审核。
3. **监控面板**：实时统计曝光、下单、使用量、ROI、库存影响。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 促销模板 | 常见促销类型模板（满减、买赠、组合包）|
| 条件/动作配置 | 拖拽/表单构建器，支持多个条件及动作组合 |
| 叠加规则 | 与价目表、定价规则、优惠券/积分叠加策略（优先级、是否可叠加） |
| 库存联动 | 促销开始时锁定库存，或实时减库存；与 `inventory` 模块交互 |
| 渠道同步 | 推送促销数据到站点/第三方渠道，提供 API/导出 |
| 审批 & 限额 | 预算超限或高折扣需审批；提供限额监控 |
| AB 测试 | 可选：同一促销支持多版本对比 |
| 实时监控 | 活动指标看板、告警（预算用尽、库存不足）|

> 对齐说明：优惠券“订单结算复算 + 支付成功核销 + 释放/返券”执行细则，统一以 `docs/plan/pricing/coupons.md` 第 9 章为准，促销与券在叠加/排他语义上保持一致。

## 5. 流程
1. 创建促销 → 选择模板 → 配置条件/动作 → 设置时间/渠道 → 审批 → 发布。
2. 发布后推送到站点/渠道，并与订单引擎共享规则。
3. 运行期间监控指标，必要时暂停/修改。

## 6. 数据 & API
- **表**：`promotions`、`promotion_conditions`、`promotion_actions`、`promotion_channels`、`promotion_metrics`、`promotion_audit_logs`。
- **API**：
  - `GET /api/promotions`、`POST /api/promotions`、`PATCH /api/promotions/{id}`、`POST /api/promotions/{id}/publish`、`POST /api/promotions/{id}/pause`。
  - `GET /api/promotions/{id}/metrics`、`POST /api/promotions/{id}/sync`（推送渠道）。

## 7. 权限
- `promotion.read`、`promotion.manage`、`promotion.publish`、`promotion.metrics`、`promotion.approval`。
- 审计：活动配置、发布、暂停、变更均记录。

## 8. KPI
| 指标 | 目标 |
| --- | --- |
| 活动审批周期 | ≤ 2 天 |
| 渠道同步成功率 | ≥ 99% |
| ROI 监测覆盖率 | 100% 活动有报表 |
| 异常告警响应 | < 30 min |

## 9. 风险
- 促销可能与价目表/优惠券冲突，需统一优先级与排他策略。
- 预算/库存未锁定可能导致超卖或亏损。
- 渠道不同，投放规则差异大，需要适配。

## 10. Backlog
- 自动化营销编排（与客户旅程结合）。
- 与 AI/预测联动，推荐促销策略。
- 促销资产管理（素材、落地页）。
- 多渠道 AB 测试与归因。

---

## 11. 促销规则引擎实施计划（015-pricing-promotions-engine）

> 本章节用于承接 015 功能开发，目标是把促销从“静态卡片/规划”推进到可配置、可试算、可下单落快照的最小闭环。若本章节与前文 PRD 粒度不一致，以本章节作为一期执行依据。

### 11.1 当前定位

促销规则属于“营销权益层”，叠加在基础定价结果之后、优惠券之前：

```text
pricebook 基础价 -> promotion 自动促销 -> coupon 用户券 -> payable 应付金额
```

一期只解决“订单结算自动促销”的核心问题，不做完整营销活动运营台。完整活动生命周期、预算审批、素材、渠道投放、复盘归因后续由 `docs/plan/marketing/campaigns.md` 承接。

### 11.2 一期 MVP 边界

#### 必做

- 管理端促销规则列表、创建、编辑、启用、停用。
- 促销类型支持：
  - `amount_off`：订单级满减。
  - `percent_off`：订单级折扣。
- 适用范围支持：
  - 全场。
  - 指定 SKU。
  - 指定渠道。
- 条件支持：
  - 订单金额门槛。
  - 生效/失效时间。
- 动作支持：
  - 固定金额减免。
  - 比例折扣，支持最大优惠上限。
- 叠加支持：
  - 是否可与优惠券叠加。
  - 促销优先级。
  - 同互斥组只命中一个。
- 订单接入：
  - quote/下单时自动匹配活动。
  - 写入订单促销快照。
  - 返回促销明细、拒绝原因、券前券后金额。

#### 暂不做

- 买赠、组合包、免邮、积分奖励。
- 预算消耗与活动 ROI 看板。
- 审批流。
- 渠道同步。
- A/B 实验。
- 活动素材和落地页。
- 用户限购/总量限额的强一致扣减。
- 商品详情页、购物车页的提前活动提示可作为 015 的增强项继续做，但不阻塞一期交易闭环；确认订单、提交订单、订单详情必须纳入一期。

### 11.3 页面与菜单

#### 页面路由

- 主页面：`web-admin/app/pages/pricing/promotions.vue`
- API composable：`web-admin/app/composables/api/usePromotions.ts`
- 后续营销聚合页：`web-admin/app/pages/market/promotions.vue`，只展示运营活动入口与效果摘要，不直接承担规则配置。
- 小程序确认订单页：`mini-app/src/pages/order/confirm.vue`
- 小程序订单详情页：`mini-app/src/pages/order/detail.vue`
- 小程序订单 API：`mini-app/src/services/miniapp-order.ts`

#### 菜单建议

- 菜单分组：定价中心 / Pricing
- 菜单项：促销规则
- 路由：`/pricing/promotions`
- 权限：
  - `com.powerx.plugins.ecommerce:pricing.promotion:read`
  - `com.powerx.plugins.ecommerce:pricing.promotion:manage`
  - `com.powerx.plugins.ecommerce:pricing.promotion:publish`
  - `com.powerx.plugins.ecommerce:pricing.promotion:metrics`

### 11.4 管理端页面操作设计

#### 11.4.1 促销列表

列表顶部筛选区：

| 控件 | 字段 | 说明 |
| --- | --- | --- |
| 关键词输入 | `keyword` | 搜索编码/名称 |
| 类型选择 | `promotion_type` | 全部、满减、折扣 |
| 状态选择 | `status` | 全部、draft、active、paused、expired |
| 渠道选择 | `channel` | 全部、miniapp、web、admin、第三方渠道 |
| 时间范围 | `valid_from`/`valid_to` | 按活动有效期筛选 |

列表字段：

| 列 | 说明 |
| --- | --- |
| 编码 | 促销业务编码，租户内唯一 |
| 名称 | 促销名称 |
| 类型 | 满减/折扣 |
| 状态 | draft/active/paused/expired |
| 适用范围 | 全场、SKU 数量、渠道 |
| 门槛 | 最低订单金额 |
| 优惠 | 减免金额或折扣比例 |
| 优先级 | 数字越小越先计算 |
| 与券叠加 | 是/否 |
| 有效期 | 开始/结束时间 |
| 操作 | 编辑、启用、停用、复制、查看日志 |

列表操作：

- `新建促销`：打开新建抽屉/弹窗。
- `查询`：按筛选条件刷新。
- `重置`：清空筛选条件。
- `启用`：仅 `draft`/`paused` 可启用。
- `停用`：仅 `active` 可停用。
- `复制`：复制基础信息与规则，状态重置为 `draft`。

#### 11.4.2 新建/编辑促销表单

表单分区一：基础信息

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `code` | 文本 | 是 | 租户内唯一；编辑时不可修改 |
| `name` | 文本 | 是 | 促销名称 |
| `description` | 文本域 | 否 | 内部说明 |
| `promotion_type` | 下拉 | 是 | `amount_off` / `percent_off` |
| `status` | 下拉 | 是 | 默认 `draft` |

表单分区二：有效期与渠道

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `valid_from` | datetime-local | 是 | 生效开始 |
| `valid_to` | datetime-local | 是 | 生效结束 |
| `channels` | 多选 | 否 | 留空表示全渠道 |

表单分区三：适用范围

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `scope_type` | 单选 | 是 | `all` / `sku` |
| `sku_ids` | SKU 搜索多选 | 条件必填 | `scope_type=sku` 时必填 |

表单分区四：条件

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `min_order_amount_yuan` | 数字 | 否 | 最低订单金额，0 表示无门槛 |

表单分区五：优惠动作

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `discount_amount_yuan` | 数字 | 条件必填 | 满减金额；`amount_off` 使用 |
| `discount_percent` | 数字 | 条件必填 | 折扣比例；如 85 表示 85 折 |
| `max_discount_yuan` | 数字 | 否 | 折扣最大优惠上限 |

表单分区六：叠加规则

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `priority` | 数字 | 是 | 默认 100；越小越先执行 |
| `stackable` | 开关 | 是 | 是否允许与其他促销叠加 |
| `stackable_with_coupon` | 开关 | 是 | 是否允许与优惠券叠加 |
| `exclusion_group` | 文本 | 否 | 同组只命中一个促销 |

表单按钮：

- `保存草稿`：保存为 `draft`。
- `保存并启用`：保存后状态为 `active`。
- `取消`：关闭表单，不提交。

#### 11.4.3 状态与交互约束

- `active` 状态下允许编辑名称、描述、结束时间、叠加规则；不允许修改类型与核心优惠动作。
- `expired` 状态不可启用，只允许复制。
- `valid_from > valid_to` 时前端阻止提交，后端也必须校验。
- `amount_off` 的优惠金额必须大于 0。
- `percent_off` 的折扣比例必须大于 0 且小于等于 100。
- `scope_type=sku` 时必须至少选择一个 SKU。

### 11.5 后端数据模型

#### 11.5.1 表：`promotion_campaigns`

用于存放促销主档与规则 JSON。字段建议：

- `id` uuid PK
- `tenant_uuid` uuid NOT NULL, index
- `code` varchar(64) NOT NULL，unique(`tenant_uuid`,`code`)
- `name` varchar(128) NOT NULL
- `description` text NULL
- `promotion_type` varchar(32) NOT NULL，枚举：`amount_off`、`percent_off`
- `condition_rule` jsonb NOT NULL，门槛/渠道/时间之外的条件
- `scope_rule` jsonb NOT NULL，适用 SKU/类目/客户范围
- `action_rule` jsonb NOT NULL，优惠动作
- `stacking_rule` jsonb NOT NULL，优先级、互斥、是否与券叠加
- `valid_from` timestamptz NOT NULL
- `valid_to` timestamptz NOT NULL
- `status` varchar(16) NOT NULL，枚举：`draft`、`active`、`paused`、`expired`
- `created_by` varchar(128)
- `updated_by` varchar(128)
- `created_at`、`updated_at`、`deleted_at`

索引建议：

- unique(`tenant_uuid`,`code`) where `deleted_at is null`
- `idx_promotion_campaign_status_time`(`tenant_uuid`,`status`,`valid_from`,`valid_to`)
- `idx_promotion_campaign_type`(`tenant_uuid`,`promotion_type`,`status`)

#### 11.5.2 表：`promotion_audit_logs`

记录创建、更新、启用、停用、命中、跳过等动作。字段建议：

- `id` uuid PK
- `tenant_uuid` uuid NOT NULL
- `promotion_id` uuid NOT NULL
- `order_id` uuid NULL
- `action` varchar(32) NOT NULL
- `action_reason` varchar(128)
- `request_id` varchar(128)
- `created_by` varchar(128)
- `payload` jsonb
- `created_at`

#### 11.5.3 表：`order_promotion_snapshots`

记录订单当次促销计算结果，避免活动后续变更影响历史订单。

- `id` uuid PK
- `tenant_uuid` uuid NOT NULL
- `order_id` uuid NOT NULL，unique(`tenant_uuid`,`order_id`)
- `currency` varchar(8) NOT NULL
- `base_total_minor` bigint NOT NULL
- `promotion_discount_minor` bigint NOT NULL
- `after_promotion_total_minor` bigint NOT NULL
- `priced_at` timestamptz NOT NULL
- `applied_promotions` jsonb NOT NULL
- `rejected_promotions` jsonb NOT NULL
- `line_allocations` jsonb NOT NULL
- `created_at`、`updated_at`

### 11.6 规则 JSON 口径

#### `condition_rule`

```json
{
  "min_order_amount_minor": 10000
}
```

#### `scope_rule`

```json
{
  "scope_type": "sku",
  "sku_ids": ["sku-1", "sku-2"],
  "channels": ["miniapp"]
}
```

`scope_type=all` 时 `sku_ids` 为空或省略。

#### `action_rule`

满减：

```json
{
  "discount_amount_minor": 1000
}
```

折扣：

```json
{
  "discount_percent_bps": 8500,
  "max_discount_minor": 3000
}
```

#### `stacking_rule`

```json
{
  "priority": 100,
  "stackable": true,
  "stackable_with_coupon": true,
  "exclusion_group": "order_discount"
}
```

### 11.7 后端服务设计

建议目录：

```text
backend/internal/entity/models/promotion/
backend/internal/entity/repository/promotion/
backend/internal/services/admin/promotion/
backend/internal/transport/http/admin/promotion/
```

核心服务：

| 服务 | 职责 |
| --- | --- |
| `TemplateService` / `CampaignService` | 创建、编辑、启用、停用、复制促销规则 |
| `QuoteService` | 根据订单草稿匹配促销并计算优惠 |
| `SnapshotService` | 下单时写入 `order_promotion_snapshots` |
| `AuditLogService` | 写促销配置变更与订单命中日志 |

计算输入：

```go
type PromotionQuoteInput struct {
    TenantUUID string
    UserID     string
    Channel    string
    Currency   string
    Items      []PromotionQuoteItem
    Now        time.Time
}
```

计算输出：

```go
type PromotionQuoteResult struct {
    Currency                    string
    BaseTotalMinor              int64
    PromotionDiscountMinor      int64
    AfterPromotionTotalMinor    int64
    AppliedPromotions           []AppliedPromotion
    RejectedPromotions          []RejectedPromotion
    LineAllocations             []LineAllocation
    PricedAt                    time.Time
    CouponStackingAllowed       bool
}
```

### 11.8 订单结算接入

#### 计算顺序

1. `pricebook` 输出基础价。
2. `promotion` 自动匹配并计算优惠。
3. `coupon` 在促销后的金额基础上继续计算。
4. 写入订单价格快照、促销快照、优惠券快照。

#### 与优惠券叠加

- 如果命中的促销 `stackable_with_coupon=false`：
  - 优惠券 quote 仍可返回可用性信息；
  - 下单最终计算时必须拒绝优惠券，原因建议为 `promotion_excludes_coupon`。
- 如果多个促销命中：
  - 按 `priority ASC` 计算；
  - 同一 `exclusion_group` 只取优惠金额最大的一个；
  - `stackable=false` 的促销命中后，后续同层促销不再继续叠加。

#### 订单快照

订单 `price_snapshot` 建议包含：

```json
{
  "base_total_minor": 10000,
  "promotion": {
    "discount_total_minor": 1000,
    "applied_promotions": []
  },
  "coupon": {
    "discount_total_minor": 500,
    "applied_coupons": []
  },
  "payable_total_minor": 8500
}
```

### 11.9 管理端 API

统一挂载在 `/api/v1/admin/promotions`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/admin/promotions` | 列表 |
| `POST` | `/api/v1/admin/promotions` | 创建 |
| `PATCH` | `/api/v1/admin/promotions/:id` | 更新 |
| `POST` | `/api/v1/admin/promotions/:id/activate` | 启用 |
| `POST` | `/api/v1/admin/promotions/:id/pause` | 停用 |
| `POST` | `/api/v1/admin/promotions/:id/clone` | 复制 |
| `GET` | `/api/v1/admin/promotions/:id/audit-logs` | 审计日志 |

列表查询参数：

- `keyword`
- `promotion_type`
- `status`
- `channel`
- `page`
- `page_size`

### 11.10 交易侧 API

如果订单 quote 已有统一入口，优先接入现有订单 quote/下单接口，不新增独立交易入口。若需要独立调试接口，一期可提供：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/v1/promotions/quote` | 促销试算 |

请求示例：

```json
{
  "channel": "miniapp",
  "currency": "CNY",
  "items": [
    {
      "line_id": "line-1",
      "sku_id": "sku-1",
      "qty": 2,
      "unit_price_minor": 5000
    }
  ]
}
```

响应示例：

```json
{
  "currency": "CNY",
  "base_total_minor": 10000,
  "promotion_discount_minor": 1000,
  "after_promotion_total_minor": 9000,
  "applied_promotions": [
    {
      "promotion_id": "promo-1",
      "code": "ORDER1000",
      "name": "满100减10",
      "discount_minor": 1000
    }
  ],
  "rejected_promotions": []
}
```

### 11.11 用户侧使用链路

促销不是用户手动领取的券，而是服务端自动命中的订单促销。后台页面负责配置规则，小程序页面负责让用户在购买链路里看到和使用促销。

#### 11.11.1 确认订单页

页面：`mini-app/src/pages/order/confirm.vue`

进入确认订单页后，前端根据当前订单草稿调用促销试算接口：

- `channel`：使用订单草稿中的渠道，默认 `miniapp`。
- `currency`：使用商品币种，默认 `CNY`。
- `items`：传 SKU、数量、单价分值。

页面必须展示：

| 区域 | 内容 |
| --- | --- |
| 金额汇总 | 商品金额、运费、自动促销优惠、合计 |
| 活动命中 | 已享活动名称，支持多活动用顿号拼接 |
| 异常状态 | 促销试算加载中、未匹配活动或试算失败提示 |
| 底部提交栏 | 使用促销后的应付金额，并展示已优惠金额 |

注意事项：

- 前端试算只用于展示，不能作为最终金额写入。
- 提交订单时，后端必须重新计算促销，最终金额以后端下单结果为准。
- 如果前端试算失败，不阻塞提交订单；后端下单仍会最终复算。

#### 11.11.2 提交订单服务端复算

后端小程序订单创建服务必须在落单前重新执行促销 quote：

1. 根据订单行计算基础商品金额。
2. 使用 `channel`、`currency`、SKU、数量、单价调用促销 quote。
3. 使用 `after_promotion_total_minor` 更新订单应付金额。
4. 写入 `order_promotion_snapshots`。
5. 返回订单详情时输出促销摘要。

该步骤是促销真正“被使用”的位置。任何来自前端的优惠金额都只能作为展示参考，不能被信任。

#### 11.11.3 订单详情页

页面：`mini-app/src/pages/order/detail.vue`

订单详情必须读取订单促销快照，而不是重新试算当前促销规则。页面展示：

| 区域 | 内容 |
| --- | --- |
| 金额明细 | 商品总额、运费、自动促销优惠、实付金额 |
| 已享活动 | 订单创建时命中的促销名称 |
| 历史一致性 | 促销后续停用、修改，不影响此处展示 |

#### 11.11.4 后续增强入口

以下仍属于 015 促销 feature 的增强范围，但不作为一期交易闭环必需项：

- 商品详情页展示“可参与活动”。
- 购物车页展示活动提示和预计优惠。
- 购物车未满足门槛时提示“再买 X 元可享优惠”。
- 管理端促销试算工具，用于运营验证规则。
- 更完整的小程序端 E2E 自动化用例。

### 11.12 错误码与拒绝原因

促销 quote 不应因为单个活动不满足而整体失败；应返回拒绝原因。

| 原因 | 说明 |
| --- | --- |
| `not_started` | 活动未开始 |
| `expired` | 活动已过期 |
| `paused` | 活动未启用 |
| `threshold_not_met` | 门槛不满足 |
| `scope_mismatch` | SKU/渠道/用户范围不匹配 |
| `exclusive_conflict` | 互斥组冲突 |
| `not_stackable` | 不可继续叠加 |
| `invalid_rule` | 规则配置无效 |
| `promotion_excludes_coupon` | 命中促销后排斥优惠券 |

### 11.13 测试与验收

后端测试：

- 规则计算：
  - 满减门槛满足/不满足。
  - 折扣比例与最大优惠上限。
  - SKU 范围匹配/不匹配。
  - 渠道匹配/不匹配。
  - 互斥组只命中一个。
  - 不可叠加规则阻止后续促销。
- 订单集成：
  - 下单时自动促销写入快照。
  - 促销后再用券，金额一致。
  - `stackable_with_coupon=false` 时优惠券被拒绝。
- 管理端处理器：
  - 创建、编辑、启用、停用、复制。
  - 租户隔离。
  - RBAC 映射。

前端验证：

- 列表筛选、分页、状态 badge 正确。
- 新建/编辑表单校验正确。
- 启用/停用后列表状态刷新。
- SKU 搜索多选可用。
- 空态、加载态、错误提示完整。
- 小程序确认订单页能显示自动促销、已享活动、促销后合计。
- 小程序提交订单后，后端最终金额与促销快照一致。
- 小程序订单详情页读取历史促销快照，不重新计算当前活动。

验收命令建议：

```bash
cd backend
go test ./internal/services/admin/promotion ./internal/transport/http/admin/promotion ./internal/transport/http/miniapp/promotion ./internal/services/admin/order ./internal/services/miniapp/order

cd ../web-admin
npm run -s build

cd ../mini-app
npm run build:h5
```

### 11.14 分阶段实施

#### Phase 1：基础模型与管理端 CRUD

- 新增促销模型、迁移、仓储。
- 管理端 API：列表、创建、更新、启用、停用。
- 管理端页面：`pricing/promotions.vue`。

#### Phase 2：促销 quote 引擎

- 实现订单级满减与折扣。
- 实现范围、门槛、时间、渠道、优先级、互斥组。
- 增加规则单测。

#### Phase 3：订单接入

- 订单 quote/下单接入促销计算。
- 写 `order_promotion_snapshots`。
- 与 coupon 顺序和互斥规则对齐。
- 小程序确认订单页接入促销试算展示。
- 小程序订单详情页展示订单促销快照。

#### Phase 4：审计与可观测

- 增加 `promotion_audit_logs`。
- 增加命中量、优惠金额、拒绝原因指标。
- 补功能文档和 quickstart。

#### Phase 5：扩展能力

- 买赠、组合包、免邮。
- 预算与限额。
- 审批流。
- 渠道同步。
- 与 `marketing/campaigns.md` 的活动控制台联动。
- 商品详情页、购物车页活动提示和凑单提醒。

### 11.15 与其他文档对齐

- 与 `pricebooks.md` 对齐：促销不改写基础价，只在基础价之后计算优惠。
- 与 `coupons.md` 对齐：促销与优惠券共享叠加/排他语义；优惠券完整生命周期仍由 coupon 模块负责。
- 与 `marketing/campaigns.md` 对齐：促销规则是活动可绑定的营销资产；活动生命周期、预算、素材和复盘不在本 feature 一期范围内。
- 与 `inventory/stock.md` 对齐：一期不做促销库存锁定；买赠/组合包上线时再接库存可售与锁定能力。

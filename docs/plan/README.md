# 电商插件系统模块计划总览

## 1. 文档目的与产出物
- 统一梳理电商插件涉及的 6 大域、30+ 子模块，明确“已有页面/组件 + 计划动作”。
- 为后续各模块在 `docs/plan/<module>` 下落地详细方案提供索引（当前已产出 `customer` 章节）。
- 对齐 Nuxt Admin（`web-admin/app/pages/**`）与 Go 后端（`backend/internal/**`）的实现优先级，确保前后端节奏、数据模型与权限策略匹配。

## 2. 参考页面与当前覆盖
| 业务域 | 主要前端路径 | 状态速览 |
| --- | --- | --- |
| 客户运营 | `web-admin/app/pages/customer/**` | 已实现会员、积分、联盟、退换货、分析等主流程；部分自助门户待与前台站点打通 |
| 商品中心 | `web-admin/app/pages/product/**` | 覆盖 SPU/SKU、类目、属性、规格、媒体、供应商；需补充多语言描述 & 批量导入 |
| 定价中心 | `web-admin/app/pages/pricing/**` | 价目表/规则/合同页面已就绪，尚缺阶梯价 UI 与审批流逻辑 |
| 营销与促销 | `web-admin/app/pages/market/**`、`customer/membership/points/mall.vue` | 促销、营销触达、渠道分析已有雏形；优惠券、礼品卡、自动化旅程待补充 |
| 订单 & 交易 | `web-admin/app/pages/market/orders.vue`、`market/payment.vue`、`payments/providers.vue` | 订单列表/支付渠道可用，支付单、交易流水与风控看板待实现 |
| 售后与纠纷 | `web-admin/app/pages/customer/returns/**` | 售后流程配置、SLA、通知模板 UI 已有，需联通仓储与退款管道 |
| 库存与履约 | `web-admin/app/pages/inventory/**`、`shipping/**` | 仓库、库存、补货、调拨、物流承运商/运费/运单页面已搭建，待打通波次与履约任务 |
| 渠道与上架 | `web-admin/app/pages/channels/**` | 建立渠道列表、上架计划、SKU 映射；缺口在多店铺授权与渠道实时状态同步 |
| 财税与结算 | `web-admin/app/pages/settlements.vue`、`tax.vue`、`payments/providers.vue` | 结算、分账、税率页面半成品，仍需票据与对账视图 |
| 数据与分析 | `web-admin/app/pages/reports/**`、`customer/analytics/**`、`market/analytics.vue` | 销售/商品/漏斗报表基本可演示，待连接真实指标与导出订阅 |
| 审计与系统设置 | `web-admin/app/pages/settings/**`、`components/dev-console/AuditHistoryTable.vue` | 设置项具备雏形，`/audit` 路由与操作日志列表待上线 |

> 注：更多迁移/依赖细节可参考 `tmp/ecommerce-web-admin-migration-plan.md` 与 `docs/plan/customer/README.md`。

## 3. 模块规划详情
以下按用户提供的模块列表展开，分为「现状」与「下一步」。

### 3.1 客户运营域
- **客户**（`customer/index.vue`、`customer/members.vue`）
  - 现状：支持客户列表、详情与会员视图，含标签与会员等级筛选。
  - 下一步：接入真实 CRM API、补齐批量导入/导出、在详情中嵌入 360 度互动历史。
- **会员等级**（`customer/membership/tiers.vue`、`membership/benefits.vue`）
  - 现状：可配置等级升降规则与权益包。
  - 下一步：联动营销券包、对接成长值计算器与等级保级报表。
- **积分与成长值**（`customer/membership/points/index.vue`、`growth-value/**`、`components/Modals/BatchAdjust*`）
  - 现状：积分规则、流水、批量调整模态和“积分商城”（`points/mall.vue`）已就绪。
  - 下一步：落地积分过期任务、风控审批、兑换后的履约（发券/发货）回写。
- **推荐/分销**（`customer/affiliates/index.vue`）
  - 现状：显示联盟招募、KPI 卡片与提现进度假数据。
  - 下一步：新增层级分佣、结算周期、风控拦截配置，并与 `settlements.vue` 的佣金结算打通。
- **客户自助退换货门户**（`customer/returns/management`、`returns/process`）
  - 现状：后台可配置售后流程、SLA、通知模板。
  - 下一步：面向 C 端的门户需由 Nuxt 前台或 H5 容器承载，计划提供统一 OAuth 入口与售后单状态同步。

### 3.2 商品中心
- **商品（SPU）**（`product/spus/index.vue`、`spus/create.vue`、`spus/edit/**`）
  - 现状：CRUD 表单、草稿/发布状态、销售渠道配置基本完备。
  - 下一步：接入批量导入与多语言描述，用于跨境场景。
- **SKU & 变体**（`product/skus/index.vue`、`product/inventory.vue`）
  - 现状：SKU 表格、库存属性展示可用。
  - 下一步：建设“变体组合器”与批量价格/库存同步能力。
- **类目**（`product/categories.vue`、`product/category-templates/**`）
  - 现状：树形类目管理与模板继承示例。
  - 下一步：支持渠道专属类目映射，与 `channels/sku-mapping.vue` 联动。
- **品牌**（`product/brands.vue`）
  - 现状：基础品牌信息填写。
  - 下一步：扩展品牌授权、资质文件与上下架审核状态。
- **属性和规格**（`product/attributes/**`、`product/specifications/**`）
  - 现状：属性组、属性字典、规格模板 UI。
  - 下一步：引入属性依赖/继承、规格校验（重量、尺寸）、并与 SKU 生成器绑定。
- **媒体与素材**（`product/media.vue`）
  - 现状：资源卡片展示和占位上传交互。
  - 下一步：结合对象存储/媒体中心 API，增加批量上传与内容标签。
- **供应商**（`product/suppliers.vue`）
  - 现状：供应商列表 + 合同概览。
  - 下一步：关联采购订单、供货能力与风控评级，支持多供应商报价。

### 3.3 定价中心
- **价目表**（`pricing/pricebooks.vue`）
  - 现状：价目表列表、条目行编辑 UI 已实现。
  - 下一步：增加版本管理、批量导入和渠道可见性控制。
- **阶梯价 / 客户专属价**（`pricing/rules.vue`）
  - 现状：规则列表与条件编辑器雏形。
  - 下一步：补齐阶梯区间、客户分群绑定、冲突检测。
- **合同价 / 协议价**（`pricing/contracts.vue`）
  - 现状：合同档案页面，展示条款与客户绑定信息。
  - 下一步：集成审批流与自动续约提示，关联 `settlements` 的对账条款。

### 3.4 营销、促销与增长
- **促销规则**（`market/promotions.vue`）
  - 现状：卡片式规则展示、基本过滤器。
  - 下一步：接入多触发条件（满减、买赠、组合包）与实时排他性检测。
- **优惠券**（`customer/membership/points/mall.vue`、`points/index.vue`）
  - 现状：积分兑换券 UI、券模板字段已定义。
  - 下一步：拆分独立 `market/coupons.vue` 页面，完成发券批次、核销统计与渠道限制。
- **礼品卡**（暂无独立页面，仅在 `points/mall.vue` 以礼品形式展示）
  - 现状：仅静态示例。
  - 下一步：新建 `market/giftcards.vue`，支持虚拟卡池管理、充值/消费流水、批量发放。
- **营销增长（自动化/增长实验）**（`market/marketing.vue`、`market/analytics.vue`）
  - 现状：营销任务列表、KPI 卡片。
  - 下一步：加入旅程编排（触发-动作-渠道）、A/B 实验与归因分析，与 `customer/analytics` 同步指标。
- **营销互动**（`market/marketing.vue`、`customer/analytics/segmentation.vue`）
  - 现状：可按标签推送活动草稿。
  - 下一步：打通消息渠道（邮件、短信、Webhook），实现互动记录回写客户档案。

### 3.5 订单、支付与交易
- **订单管理**（`market/orders.vue`）
  - 现状：订单列表、筛选、侧边详情面板。
  - 下一步：实现订单状态流引擎、拣货/发货指令、异常拦截（风控标签）。
- **支付单 / 交易流水**（`market/payment.vue`、`payments/providers.vue`）
  - 现状：支付渠道配置页面可维护 Provider，交易流水缺失。
  - 下一步：新增 `payments/transactions.vue`，同步第三方支付状态、手续费、对账差异。
- **营销互动订单关联**
  - 下一步：在订单详情绑定活动 ID、优惠券、联盟渠道，实现漏斗归因。

### 3.6 售后与纠纷
- **售后流程**（`customer/returns/process/index.vue`、`returns/config.vue`）
  - 现状：可配置售后类型、节点、通知模板。
  - 下一步：与仓储收货、退款 API 互通，支持 SLA 定时器与升级机制。
- **纠纷管理**（`customer/returns/notifications`）
  - 现状：通知模板示例。
  - 下一步：新增纠纷看板、仲裁记录、证据上传入口。

### 3.7 库存与仓储
- **仓库 / 库存总览**（`inventory/warehouses.vue`、`inventory/stock.vue`）
  - 现状：仓库列表、库存表格。
  - 下一步：加入地理信息、波次任务、SKU 库存状态（可售/在途/残次）。
- **补货与安全库存**（`inventory/replenishment.vue`）
  - 现状：补货建议列表样例。
  - 下一步：接入预测模型、生成采购单草稿、联动供应商。
- **盘点**（`inventory/stocktake.vue`）
  - 现状：盘点计划 UI。
  - 下一步：移动端扫码、盘盈盘亏差异分析。
- **库存调拨**（`inventory/transfers.vue`）
  - 现状：调拨单列表。
  - 下一步：多仓审批、物流追踪、与 `shipping/waybills.vue` 的串联。

### 3.8 履约与物流
- **物流承运商**（`shipping/carriers.vue`）
  - 现状：承运商列表 + 服务区域设定。
  - 下一步：扩展服务 SLA、计费方式、API 凭证管理。
- **运费模板**（`shipping/templates.vue`）
  - 现状：模板配置表单。
  - 下一步：支持分区定价、阶梯计价及渠道匹配。
- **运单与轨迹**（`shipping/waybills.vue`）
  - 现状：静态运单表格。
  - 下一步：接入轨迹 API，提供订阅推送与异常提醒。

### 3.9 渠道与上架
- **店铺 / 渠道管理**（`channels/index.vue`、`market/channels.vue`）
  - 现状：渠道概览、授权状态卡片。
  - 下一步：多店铺授权流程、Webhook 状态同步、渠道容量指标。
- **上架发布** (`channels/publishing.vue`)
  - 现状：上架计划、任务表。
  - 下一步：支持批量推送、版本比对、错误回执展示。
- **SKU 映射** (`channels/sku-mapping.vue`)
  - 现状：映射表格。
  - 下一步：自动匹配建议、差异监控，与 `product/categories` 的映射规则联动。

### 3.10 财税与结算
- **支付渠道**（`payments/providers.vue`）
  - 现状：配置渠道、凭证、结算周期。
  - 下一步：增加风控策略、监控阈值、切换策略。
- **结算与分账**（`settlements.vue`）
  - 现状：结算周期、分账方卡片。
  - 下一步：结算单明细、对账状态、站内通知。
- **税率与税则**（`tax.vue`）
  - 现状：税率配置占位。
  - 下一步：多地区税类、合规校验、与订单/发票联动。

### 3.11 数据与增长分析
- **销售报表**（`reports/sales.vue`）
  - 现状：指标卡 + 图表。
  - 下一步：接入实时数据与导出、订阅。
- **商品报表**（`reports/products.vue`）
  - 现状：商品表现表格。
  - 下一步：加上库存周转、渠道表现维度。
- **转化漏斗**（`reports/conversion.vue`、`customer/analytics/funnel.vue`）
  - 现状：静态漏斗图。
  - 下一步：串联营销活动数据，支持自定义漏斗阶段。
- **增长分析**（`market/analytics.vue`、`customer/analytics/segmentation.vue`）
  - 现状：分群与渠道分析示例。
  - 下一步：引入实验对照、目标设定与自动警报。

### 3.12 审计日志与系统设置
- **审计日志**（`web-admin/app/components/dev-console/AuditHistoryTable.vue`，计划中的 `/audit` 页面）
  - 现状：组件 & store (`web-admin/app/stores/dev-console/audit.ts`) 已有，但未挂接页面。
  - 下一步：新建 `web-admin/app/pages/audit/index.vue`，渲染历史、导出、过滤；后台依赖 `backend/internal/transport/http/admin/console/audit_handler.go`。
- **订单与流程配置**（`settings/orders.vue`）
  - 现状：订单 SLA、流程配置项。
  - 下一步：让其驱动 `market/orders` 状态机，并记录改动审计。
- **货币与小数精度**（`settings/currency.vue`）
  - 现状：可设置货币、精度、汇率占位。
  - 下一步：与价目表、结算模块共享配置，支持多币种优先级。
- **权限与角色**（`settings/roles.vue`）
  - 现状：角色列表、权限树、操作日志表格（`auditColumns`）。
  - 下一步：对接 RBAC API、支持自定义资源与审批。

## 4. 迭代建议与协调机制
1. **阶段化推进**：按照“客户与商品 → 定价/促销 → 交易/履约 → 财税/分析”的顺序分四个迭代，每个迭代都包含最少可上线能力 + 数据校验。
2. **跨模块依赖梳理**：建立周会，匹配 `backend/internal/services/**` 与 Nuxt 页面，确保 API 合约冻结后再推进 UI。
3. **指标与审计前置**：所有新增模块在设计时同步定义关键指标（GMV、库存周转、券发券核销率）与审计需求，避免后补。
4. **文档沉淀**：本 README 提供索引；每个二级模块完成详细方案后，追加到 `docs/plan/<module>/README.md`，保持结构一致。

## 5. 下一步动作
- [ ] 按本总览拆解各模块 owner，维护共识版 Roadmap。
- [ ] 在 `web-admin` 中新增缺失的页面骨架（例如 `/audit`、`market/giftcards.vue`、`payments/transactions.vue`）。
- [ ] 后端为价目表、订单、结算、审计提供正式 API Schema，更新 API Contract 文档。
- [ ] 当模块详细方案 ready 后，在 README 顶部列表中添加链接，便于导航。

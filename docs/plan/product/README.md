# 商品中心模块规划总览

> 适用范围：`web-admin/app/pages/product/**`、`inventory/**`（与库存联动部分）及相关组件/模态。目标是梳理商品域（SPU/SKU、类目、属性、规格、媒体、供应商）现状与下一步路线，作为后续各子模块 PRD 的索引。

## 1. 文档目的
- 明确商品中心的范围、角色、能力矩阵，为研发/产品/运营提供统一蓝图。
- 对齐 Nuxt Admin 中已有页面与后台（`backend/internal/**`）需要暴露的 API。
- 拟定迭代计划与跨模块依赖（库存、定价、渠道上架、营销）。

## 2. 当前页面覆盖
| 模块 | 前端路径 | 状态摘要 |
| --- | --- | --- |
| 商品总览 | `product/index.vue` | KPI 卡、快捷入口，需接入真实指标 |
| SPU 管理 | `product/spus/**` | CRUD 表单与草稿/发布流程示例，缺少批量导入、多语描述 |
| SKU/库存 | `product/skus/index.vue`、`product/inventory.vue`、`product/skus/[id].vue` | SKU 生成器、批量操作、渠道映射、库存、条码/序列号等核心功能已贯通；后续持续扩充报表与多租户指标 |
| 类目 & 模板 | `product/categories.vue`、`product/category-templates/**` | 树形类目、属性模板、销售规格模板、渠道映射与审计入口 |
| 属性 & 规格 | `product/attributes/**`、`product/specifications/**` | 属性组、规格模板 UI 完整，需要规则校验、依赖配置 |
| 品牌 | `product/brands.vue` | 基础信息表单，需要资质审核、上下架规则 |
| 媒体 | `product/media.vue` | 资源卡片、上传占位，需串接对象存储与媒资标签 |
| 供应商 | `product/suppliers.vue` | 列表 + 合同概览，需扩展资质、供货能力、评分 |

## 3. 角色与价值
| 角色 | 核心目标 |
| --- | --- |
| 商品运营 | 快速搭建商品目录、管理多渠道、多语言内容 |
| 渠道运营 | 根据渠道/市场要求调整属性、价格、类目映射 |
| 供应链/采购 | 维护供应商、采购价格、供货能力、合同 |
| 品控/媒资 | 管理图片/视频、审核合规、输出统一素材 |
| 开发/数据 | 对接 API、保障数据同步、做 BI 分析 |

## 4. 功能规划概览
1. **SPU 管理**
   - 核心字段：基础信息、属性集、销售渠道、生命周期、上下架策略。
   - 支持：草稿/审批、版本管理、批量导入导出（CSV/Excel）、多语言。
   - 与渠道联动：推送到 `channels/publishing`、SKU 映射。
2. **SKU 管理与变体生成**
   - 维度：规格组合、条码、重量尺寸、库存上限、价格引用、序列号/批次属性。
   - 已上线 SKU 生成器（笛卡尔组合 + 默认值模板）与 Pinia store 同步，支持局部编辑、复制、冲突提示。
   - 批量任务（价格/库存/导入导出）复用 `product_sku_bulk_tasks`，含审批策略、错误报告与任务列表 UI。
   - 与库存/条码/渠道联动：实时展示库存（含 SLA 告警）、条码唯一性校验、标签打印、序列号录入，以及渠道映射与发布日志。
3. **类目与模板**
   - 类目树支持多层结构、权限控制、市场/国家维度。
   - 类目模板定义强制字段、校验规则、上下游映射（如 Amazon/B2B 渠道）。
   - 品类销售规格模板定义同品类 SPU 可复用的销售规格维度和值域，例如尺寸、体型、工艺；SPU 通过显式同步后裁剪为实际启用规格。
   - 提供类目属性继承、批量迁移工具。
4. **属性 & 规格**
   - 属性模板解决“该品类商品需要填写哪些描述/合规字段”；销售规格解决“SKU 由哪些可售变体维度组成”。
   - SKU 生成器只读取 SPU 已启用销售规格；品类模板只作为可复用来源，不直接生成 SKU。
   - 属性可配置是否对客户可见、是否参与搜索/筛选。
5. **媒体中心**
   - 管理产品图片、视频、3D 模型、文档；支持标记、版本、合规审核。
   - 提供裁剪、转码、渠道适配策略。
6. **供应商管理**
   - 供应商档案、资质、合同、报价、供货 SKU、交期表现。
   - 关系：与 SPU/SKU 绑定供货渠道，与采购/库存模块共享信息。
7. **跨模块联动**
   - 与定价中心：价目表引用 SKU，支持多币种价格同步。
   - 与库存/仓储：库存属性、批次、序列号。
   - 与渠道/上架：商品映射、内容适配、上架状态。
   - 与营销：商品标签、适用促销、Bundling。

## 4.1 品类、SPU、SKU、销售规格的业务口径

这套模型的目标不是让运营在每个商品里重复造规格，而是先在品类里沉淀常用规格，再让具体商品按需使用。

以前的做法是：每建一个 SPU，都要在这个 SPU 里重新写一遍规格。例如 10 个“毛绒娃娃”商品，都要各自维护“尺寸、体型、工艺”。时间久了容易出现命名不一致：

```text
尺寸 / 大小 / 高度
10cm / 10厘米 / 10 CM
海星体 / 海星 / starfish
```

现在的做法是：

```text
品类负责定义“这类商品通常有哪些规格”
SPU 负责决定“这个商品实际卖哪些规格”
SKU 负责落成“每一个能卖、能计价、能库存的具体组合”
```

例如：

```text
品类：毛绒玩具
  销售规格模板：
    尺寸：10cm、20cm、30cm
    体型：普通体、海星体、骨架体
    工艺：印花、绣花

SPU：努努娃娃
  从品类同步后，实际启用：
    尺寸：10cm、20cm
    体型：海星体、骨架体

SKU：
  10cm / 海星体
  10cm / 骨架体
  20cm / 海星体
  20cm / 骨架体
```

后台操作路径：

```text
类目管理 -> 选择品类 -> 销售规格 -> 维护标准规格
SPU 编辑 -> 规格定义 -> 从品类同步 -> 删除不用的规格/规格值 -> 保存 -> 生成 SKU
```

## 5. 迭代分层建议
1. **基础数据层**：SPU/SKU CRUD、属性/类目/品牌配置、媒体上传。
2. **协作 & 批量层**：多语言、批量导入、版本管理、审批、供应商。
3. **智能 & 多渠道层**：渠道映射、媒体适配、模板市场、推荐/AI 质检。
4. **扩展层**：商品包（组合/套装）、BOM、多品牌/区域差异化。

## 6. 依赖与接口（概述）
- **后端服务**：`backend/internal/services/product`（待补）、`internal/services/inventory`、`internal/services/pricing`。
- **关键 API**（示例）：
  - `GET/POST /api/products/spus`、`/spus/{id}`、`/spus/{id}/publish`
  - `GET/POST /api/products/skus`、`/skus/bulk`
  - `GET/POST /api/products/categories`、`/attributes`、`/specifications`
  - `POST /api/products/media/upload`、`/suppliers`
- 需要统一的导入/导出任务 API（复用 `jobs` 服务）。

## 7. 与其他模块的对齐
- **定价中心**：价目表/合同价引用 SKU，需要联合测试。
- **库存与履约**：库存维度、批次号、锁定逻辑；商品下架时通知库存冻结。
- **渠道上架**：商品发布状态、SKU 映射、内容同步。
- **营销增长**：商品标签、可售渠道、促销适配。

## 8. 后续工作
- 为每个子模块撰写独立 PRD：`spu.md`、`sku.md`、`categories.md`、`attributes.md`、`media.md`、`suppliers.md` 等。
- 对齐 API Schema，补充 `backend/internal` 目录下的服务设计文档。
- 评估批量工具（导入导出、模板复制）的技术方案。
- 明确与 PowerX 底座（模板、Bridge、Dev Console）复用的组件。

## 9. SKU 变体管理迭代（002-product-sku-management）
- **交付范围**：完成 US1（SKU 生成）、US2（批量调价/库存 + 导入导出 + 审批 + 任务日志）、US3（渠道映射、库存可视、条码与序列管理）以及 Phase 6 可观测性/文档工作。  
- **后端落点**：`backend/internal/services/admin/product_sku/**`、`transport/http/admin/product_sku/**`、`entity/models/product_sku/**`、`observability/product_sku/logger.go`，所有接口都在 `/v1/products/skus/**`。  
- **前端落点**：`web-admin/app/pages/product/skus/**`、`components/product/sku/**`、`stores/productSku.ts`，并提供 `web-admin/tests/e2e/product-sku.spec.ts` 冒烟用例。  
- **运维与文档**：Quickstart、Plan README、`reports/sku-release.md` checklist 统一描述 SKU 生成→批量任务→渠道/库存→条码/序列闭环，配合 `make test` / `make test-admin` / `make integration-smoke` 进行回归。

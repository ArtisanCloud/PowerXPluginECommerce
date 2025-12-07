# 商品类目 & 模板 PRD

> 适用页面：`web-admin/app/pages/product/categories.vue`、`product/category-templates/**`、`product/spus/create.vue` 的类目选择区块，以及后台 `backend/internal/services/product/category` 相关 API。

## 1. 背景与目标
- 类目承担商品信息的结构化载体，是 SPU/SKU 校验、搜索、渠道映射的基础。当前仅有静态树和示例模板，缺乏多层管理、继承规则、渠道映射能力。
- 类目模板需定义强制字段、属性继承、与渠道及第三方平台（如 Amazon）映射，保障商品在多市场场景下的一致性。

**目标**
1. 提供完整的类目 CRUD、拖拽排序、批量迁移、权限控制。
2. 让模板可配置字段/属性/校验规则，自动作用于 SPU/SKU 录入，支持多语言。
3. 完成与渠道、搜索、SEO 的映射数据模型，对接 API 和批量工具。

## 2. 用户与角色
| 角色 | 诉求 |
| --- | --- |
| 商品运营 | 调整类目结构、配置模板、快速定位商品 |
| 渠道运营 | 映射外部渠道类目、同步上架策略 |
| 品控/法务 | 强制类目字段、检查合规项 |
| 技术/数据 | 通过 API 获取类目树，用于规则引擎、BI、搜索 |

## 3. 信息架构
1. **类目树管理 (`categories.vue`)**
   - 功能区：树形导航、搜索过滤（名称、ID、标签、渠道）、状态（启用/停用）。
   - 详情面板：名称、编码、父节点、级别、别名、多语言、渠道映射、SEO 字段、模板绑定、展示顺序。
   - 操作：新增子类目、拖拽排序/调整层级、批量导入、批量停用、权限设置。
2. **类目模板 (`category-templates/**`)**
   - 模板列表：名称、适用层级、关联渠道、启用状态、最后修改人。
   - 模板编辑：字段分组（基础/内容/合规）、字段类型、是否必填、校验规则、默认值、多语言、是否继承。
   - 属性继承：与属性中心联动，选择可复用的属性集。
3. **渠道映射**
   - 映射关系：平台、平台类目 ID、映射策略（自动/手动）、同步状态。
   - 多渠道 tab：B2C、B2B、跨境平台。
4. **模板下发与校验**
   - 在 SPU 创建/编辑时，根据所选类目自动加载模板字段。
   - 提供预览面板和“模拟填报”工具，验证模板配置。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 类目 CRUD | 多层级创建、编辑、排序、启停、批量迁移 |
| 模板配置 | 字段/属性/校验规则可视化配置，支持多语言、继承 |
| 渠道映射 | 多渠道类目映射、同步状态追踪、批量导入导出 |
| 权限控制 | 按角色或组织设置可管理的类目范围 |
| 模板版本 | 模板变更记录、发布/回滚、影响范围预览 |
| SEO & 展示 | 配置 URL 别名、元信息、前台展示顺序、推荐位 |
| 校验/预览 | 在模板内模拟 SPU 表单，校验字段规则 |
| 数据导入导出 | 支持 Excel/JSON 导入类目树、模板；导出用于外部系统 |

## 5. 关键流程
1. **创建类目**
   1. 在树上选择父节点 → 点“新增类目”，填写名称、编码、语言、模板、渠道映射。
   2. 保存后自动继承父级默认模板，可手动覆盖。
2. **模板配置**
   1. 在模板列表新建或复制模板 → 定义字段、属性、校验。
   2. 选择适用的类目层级或具体节点 → 发布模板版本。
3. **批量迁移/调整层级**
   1. 选定多个类目 → 选择“迁移” → 指定新的父节点。
   2. 系统校验是否会引发模板冲突 → 生成任务 → 更新成功后通知相关 SPU/SKU 负责人。
4. **渠道映射**
   1. 在类目详情的渠道 tab 中添加平台 → 选择平台类目 ID、映射策略。
   2. 可批量导入 mapping 表格；同步状态写入日志。
5. **模板回滚**
   1. 模板编辑后产生新版本 → 在发布前预览受影响类目。
   2. 如发现问题，可回滚到历史版本，并通知 SPU 重新校验。

## 6. 数据 & API
- **表**：`product_categories`、`product_category_locales`、`product_category_templates`、`product_category_template_fields`、`product_category_template_versions`、`product_category_mappings`、`product_category_permissions`。
- **API（示例）**：
  - `GET /api/products/categories/tree`、`POST /api/products/categories`、`PATCH /api/products/categories/{id}`、`POST /api/products/categories/{id}/move`、`PATCH /api/products/categories/{id}/status`。
  - `POST /api/products/categories/import`、`POST /api/products/categories/export`。
  - `GET /api/products/category-templates`、`POST /api/products/category-templates`、`POST /api/products/category-templates/{id}/publish`、`POST /api/products/category-templates/{id}/rollback`。
  - `POST /api/products/categories/{id}/mappings`、`GET /api/products/categories/{id}/mappings`.
  - `GET /api/products/categories/{id}/audit`.

## 7. 权限与审计
- 权限：`product.category.read`、`product.category.manage`、`product.category.template`、`product.category.mapping`、`product.category.import`。
- 审计：记录类目 CRUD、模板编辑/发布、渠道映射、批量任务与导入导出，写入 `admin_console_audit_events`，并在类目详情展示最近操作与操作者。

## 8. KPI / 度量
| 指标 | 目标 |
| --- | --- |
| 类目调整审批周期 | ≤ 2 个工作日 |
| 模板发布错误率 | ≤ 1%（错误需提供回滚） |
| 渠道映射完整率 | ≥ 95%（重点渠道） |
| SPU 录入校验通过率 | ≥ 98% |
| 导入任务成功率 | ≥ 99% |

## 9. 依赖与风险
- 依赖属性/规格模块：模板中引用的属性需保持一致，避免孤立字段。
- 与 SPU/SKU 紧耦合：类目调整需触发商品重新校验，可能导致大量回归测试。
- 渠道映射需要实时同步外部平台类目，需考虑缓存与更新策略。
- 多语言/SEO 字段需与 CMS/前台共享，避免冲突。
- 批量迁移操作风险高，需要预览与回滚机制。

## 10. Backlog
- 类目差异模板：按渠道/国家派生子模板。
- 自动建议：根据历史商品表现推荐类目/模板字段。
- 类目 A/B 实验：不同前台展示顺序及 SEO 策略。
- 与外部 PIM 对接：支持 webhook 与实时同步。
- 可视化影响分析：展示模板变更影响到的商品/渠道数量。

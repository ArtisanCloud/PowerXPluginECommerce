# 属性 & 规格管理 PRD

> 适用页面：`web-admin/app/pages/product/attributes/**`、`product/specifications/**`、SPU/SKU 表单中的属性区块，以及 `backend/internal/services/product/attribute`、`.../specification` API。

## 1. 背景与目标
- 属性与规格是商品数据的基础元件，决定 SPU/SKU 信息结构、SKU 生成、搜索筛选、渠道映射等能力。当前界面仅有基本属性表，缺少属性组管理、规格模板、依赖规则、校验与批量工具。

**目标**
1. 建立属性中心，支持属性组/属性/规格值的层级管理，多语言及数据类型校验。
2. 支持规格模板与 SKU 生成器的联动，提供可视化配置与校验。
3. 提供属性继承、依赖关系、渠道/前台可见性配置，支撑多渠道、多语言场景。

## 2. 用户与角色
| 角色 | 诉求 |
| --- | --- |
| 商品运营 | 定义属性组、配置属性模板、控制表单字段 |
| 渠道运营 | 调整渠道可见属性、映射外部平台字段 |
| 品控/法务 | 校验属性/规格是否符合合规要求 |
| 技术/数据 | 通过 API 获取标准属性 schema，用于 SKU 生成、搜索、BI |

## 3. 信息架构
1. **属性组管理 (`attributes/index.vue`)**
   - 列字段：属性组名称、适用类目/模板、属性数量、状态、更新时间、负责人。
   - 操作：创建/编辑属性组、绑定类目、配置默认属性。
2. **属性详情 (`attributes/[id].vue`)**
   - 基础信息：名称、多语言、数据类型（文本、数字、枚举、布尔、日期、媒体）、默认值、是否必填、是否唯一。
   - 可见性：前台可见、搜索筛选、SKU 可继承、渠道覆盖。
   - 依赖关系：依赖其他属性值（如颜色→色板）、条件校验。
   - 值域管理：枚举值维护、排序、禁用。
   - 多语言：属性名称和值分别配置翻译。
3. **规格模板 (`specifications/**`)**
   - 模板列表：名称、适用类目、规格维度、绑定属性、启用状态。
   - 模板编辑：选择属性→生成规格维度→设置 SKU 生成规则、默认组合、可选/必选。
4. **属性继承与模板**
   - 类目模板中引用属性组，可配置继承策略（只读/可覆盖）。
   - 在 SPU/SKU 表单中按模板自动加载属性。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 属性组管理 | 分组维护属性，绑定类目/模板，支持复制、启停 |
| 属性 CRUD | 多语言、数据类型、校验、默认值、值域、多渠道可见性 |
| 规格模板 | 配置规格维度、与 SKU 生成器联动，支持可视化矩阵 |
| 依赖与规则 | 属性间依赖、联动、条件显示/校验 |
| 多语言/多渠道 | 属性名称/值多语言，控制前台/渠道可见性、映射字段 |
| 属性继承 | 类目模板继承属性组，允许覆盖/扩展 |
| 导入导出 | 支持属性/枚举的批量导入导出（Excel/JSON） |
| 审计与版本 | 记录属性/模板变更，支持版本回滚/预览 |
| 搜索/标签 | 属性标签、搜索功能，快速定位属性 |

## 5. 关键流程
1. **创建属性**
   1. 选择属性组 → 点击“新建属性”，填写名称、类型、校验规则。
   2. 配置值域（如枚举）、多语言、可见性与渠道映射。
   3. 保存后可同步到关联模板，触发校验或通知。
2. **规格模板配置**
   1. 在规格模板列表创建新模板 → 选择适用类目/属性组。
   2. 设定规格维度（如颜色、尺码）、默认组合、可选/必选。
   3. 发布模板后，SPU/SKU 可使用该模板生成 SKU。
3. **属性依赖**
   1. 在属性详情设置依赖规则（如选择“颜色”为红时，暴露“色板素材”属性）。
   2. 在 SPU 填写时，表单动态显示依赖字段。
4. **批量导入属性值**
   1. 下载属性/枚举导入模板 → 填写 → 上传。
   2. 系统校验重复/冲突 → 生成任务 → 更新成功后记录日志。

## 6. 数据 & API
- **表**：`product_attribute_groups`、`product_attributes`、`product_attribute_locales`、`product_attribute_values`、`product_attribute_value_locales`、`product_attribute_rules`、`product_spec_templates`、`product_spec_template_versions`。
- **API（示例）**：
  - `GET /api/products/attribute-groups`、`POST /api/products/attribute-groups`、`PATCH /api/products/attribute-groups/{id}`。
  - `GET /api/products/attributes`、`POST /api/products/attributes`、`PATCH /api/products/attributes/{id}`、`PATCH /api/products/attributes/{id}/status`。
  - `POST /api/products/attributes/import`、`POST /api/products/attributes/export`。
  - `GET /api/products/spec-templates`、`POST /api/products/spec-templates`、`POST /api/products/spec-templates/{id}/publish`、`POST /api/products/spec-templates/{id}/rollback`。
  - `GET /api/products/attributes/{id}/audit`。

## 7. 权限与审计
- 权限：`product.attribute.read`、`product.attribute.manage`、`product.attribute.template`、`product.attribute.import`。
- 审计：属性/规格模板创建、修改、发布、回滚，以及导入导出、依赖规则更新等操作写入 `admin_console_audit_events`，并在详情页显示操作记录。

## 8. KPI / 度量
| 指标 | 目标 |
| --- | --- |
| 属性配置准确率 | ≥ 99%（SPU/SKU 校验通过率） |
| 模板发布失误率 | ≤ 1%，必须支持回滚 |
| 属性导入任务成功率 | ≥ 99% |
| SKU 生成器错误率 | ≤ 1% |
| 属性/模板审批周期 | ≤ 2 个工作日 |

## 9. 依赖与风险
- 与类目模板高度耦合：属性变更需触发类目/商品重新校验。
- 多语言、多渠道映射导致维护成本高，需要工具支持批量更新和校验。
- 复杂依赖和规则易导致表单逻辑膨胀，需要良好 UX 和可视化。
- SKU 生成器依赖规格模板，需保证幂等，防止生成重复 SKU。
- 批量导入需严密校验，避免污染主数据。

## 10. Backlog
- 属性/规格推荐：基于历史商品自动推荐常用属性。
- 属性影响分析：变更前预估影响的商品数量与渠道。
- API Schema 自动生成：用于前端表单或第三方系统。
- 属性值翻译集成：接入翻译服务批量生成多语言。
- 条件模板：不同渠道/国家套用不同属性规则。

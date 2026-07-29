# Data Model: 商品类目与类目模板管理（Phase 1）

> 目标：定义可被实现与测试的实体、关键字段、关系与约束（保持租户隔离与 RBAC/审计要求）。

## Entities

### 1) Product Category（商品类目）

- **Identity**：`id`（UUID），`tenant_uuid`（UUID string）
- **Hierarchy**：
  - `parent_id`（可空）
  - `path`（用于表达层级路径，便于树查询/筛选；值规则在实现阶段固定）
  - `level`（层级深度）
  - `sort_order`（同级排序）
- **Business fields**：
  - `code`（租户内全局唯一）
  - `display_name`（同一 `parent_id` 下唯一）
  - `alias_slug`（租户内全局唯一，用于 SEO/路由）
  - `status`（enabled/disabled）
- **Template binding**：`template_id`（可空；绑定后表示该类目显式模板）
- **SEO & display**：`seo_title`/`seo_description`/`seo_keywords`（可选）、`is_featured`（可选）、`image_url`（可选）
- **Timestamps**：`created_at`/`updated_at`/`deleted_at`（软删除建议用于审计与恢复）

**Constraints**

- 唯一性（租户维度）：
  - `code` 全局唯一
  - `alias_slug` 全局唯一
  - (`parent_id`, `display_name`) 唯一
- 层级安全：
  - 迁移/修改父节点必须阻止形成环
- 前台展示：
  - “可展示类目树”仅返回 status=enabled 且其祖先均为 enabled 的节点（停用节点及其子树不返回）

### 2) Category Locale（类目多语言信息）

- **Identity**：`id`（UUID）或复合主键（实现阶段决定）
- **Fields**：`tenant_uuid`，`category_id`，`locale`，`name`，`description`
- **Constraints**：
  - (`category_id`, `locale`) 唯一
  - 读取时缺失语言按“默认语言回退”策略返回

### 3) Category Template（类目模板）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Fields**：`name`，`status`（enabled/disabled），`applicable_level`（可选），`notes`（可选）
- **Versioning**：模板发布后产生不可变版本快照（见 Template Version）

### 4) Template Field（模板字段）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`template_id`
- **Fields**：
  - `field_key`（模板内唯一）
  - `group`（基础/内容/合规等）
  - `field_type`（string/number/enum/date/boolean/json 等；实现阶段枚举化）
  - `required`（bool）
  - `validation_rules`（结构化规则）
  - `default_value`（可选）
  - `i18n_label`/`i18n_help`（可选）
  - `inheritable`（bool；本期模板继承为“整模板继承”，字段级继承可后置）

### 5) Template Version（模板版本）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`template_id`
- **Fields**：
  - `version_number`（递增）
  - `published_at`，`published_by`
  - `snapshot`（模板与字段的不可变快照）
- **Behavior**：
  - 子类目未绑定模板时，继承最近祖先的“已发布模板版本”
  - 回滚创建新版本或将当前版本指向历史版本（实现阶段固定）

### 6) Category Sale Spec Group（品类销售规格组）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`category_id`
- **Fields**：
  - `code`（同一类目内唯一，例如 `color`/`size`/`body_type`）
  - `name`（展示名称）
  - `sort_order`
  - `required`
  - `allow_custom`（是否允许 SPU 在同步后扩展该维度）
  - `status`（active/disabled）
- **Behavior**：
  - 作为同品类 SPU 的销售规格模板来源。
  - SPU 需显式同步后生成自己的 `ProductSpecGroup`，系统不静默覆盖 SPU 已启用规格。

### 7) Category Sale Spec Option（品类销售规格值）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`category_id`，`group_id`
- **Fields**：
  - `code`（同一规格组内唯一，例如 `red`/`10cm`/`embroidered`）
  - `name`
  - `sort_order`
  - `meta`（可选结构化扩展）
  - `status`（active/disabled）
- **Behavior**：
  - 作为 SPU 可选择的规格值模板；同步到 SPU 后成为 SKU 组合的候选值。

### 8) Category Mapping（渠道类目映射）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`category_id`
- **Fields**：
  - `channel`（渠道/平台标识）
  - `platform_category_id`
  - `strategy`（manual/auto）
  - `sync_status`（pending/synced/failed 等）
  - `metadata`（可选）
- **Constraints**：
  - (`category_id`, `channel`) 唯一（若需要同类目多映射，则在实现阶段调整唯一性）
  - CSV 导入必须检测冲突并输出可定位错误

### 9) Category Permission Scope（类目权限范围）

- **Identity**：`id`（UUID），`tenant_uuid`
- **Relationship**：`principal_type`（role/department/user 等），`principal_id`，`category_id`
- **Fields**：`action`（read/manage/template/mapping/import）
- **Behavior**：用于限定可管理类目集合（具体 RBAC 映射在实现阶段落地）

## Auditing

- 对下列动作写审计事件：类目 CRUD、排序/迁移、启停、模板编辑/发布/回滚、映射增删改、CSV 导入导出、批量重检任务触发与结果。

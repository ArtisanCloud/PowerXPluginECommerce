# 商品类目（Categories）使用指导（类目树 / 模板 / 渠道映射）

本指南覆盖商品类目相关能力的手工验证与运营使用路径：**类目树维护（US1）**、**类目模板配置并驱动 SPU 录入校验（US2）**、**渠道类目映射与 CSV 批量处理（US3）**。

## 背景与目标

- 参考规格：`specs/004-product-categories/spec.md`、`specs/004-product-categories/contracts/openapi.yaml`
- 目标：在 Web Admin 中可维护类目树并配置模板/渠道映射；miniapp 可读取“可展示类目树”；SPU 新建/编辑按类目模板校验 `attributes`。

## 依赖与权限

| 类型 | 需求 |
| --- | --- |
| RBAC（路由级） | 需要具备类目相关资源权限（例如 `com.powerx.plugins.ecommerce:product.category:manage`、`com.powerx.plugins.ecommerce:product.category.template:manage`、`com.powerx.plugins.ecommerce:product.category.mapping:manage`、`com.powerx.plugins.ecommerce:product.category.import:manage`） |
| 环境 | 已完成迁移：`make migrate`；或直接 `make dev`（会自动 migrate 后启动后端） |
| 账号/租户 | 需要可用租户与可登录账号；所有请求需带租户上下文（宿主鉴权上下文或 query `tenant_uuid`，如默认租户 `00000000-0000-0000-0000-000000000001`） |

> 提示：管理端 API 需要管理员 JWT；miniapp API 需要“客户（customer）鉴权 token”（详见下文的 miniapp 验证步骤）。

## 数据准备

1. 启动后端：`make dev`（默认绑定 `:8086`）。
2. 启动管理端：`cd web-admin && npm run dev`，访问 http://127.0.0.1:3032。
3. 登录 Web Admin（确保具备上述 RBAC 权限）。

## 测试步骤

### 1) 维护类目树（US1）

**UI 路径**：Web Admin →「产品 → 类目（Categories）」。

1. 在左侧树选择父节点（根节点可选空/默认）→ 点击「新增」创建类目。
2. 填写并保存（最小建议字段：名称/编码/状态）。
3. 在详情面板执行：
   - 编辑名称/编码并保存；
   - 启停类目（停用后该类目及其子树不应出现在 miniapp“可展示树”中）；
   - 拖拽或迁移到新的父节点（系统应阻止形成环）。
4. 如需查看该类目下的商品：点击「查看商品」跳转到 SPU 列表，并自动带上类目过滤（默认包含子树）。
5. **删除保护**：若类目存在子类目或被商品引用，删除应被禁止（建议先迁移/停用）。

**接口参考**
- `GET /api/v1/admin/product/categories/tree`
- `GET /api/v1/admin/product/categories`
- `POST /api/v1/admin/product/categories`
- `PATCH /api/v1/admin/product/categories/{id}`
- `POST /api/v1/admin/product/categories/{id}/move`
- `PATCH /api/v1/admin/product/categories/{id}/status`
- `DELETE /api/v1/admin/product/categories/{id}`

### 1.1) 类目与 SPU 列表联动（管理端）

当你从「类目（Categories）」页面点击「查看商品」，会跳转到 `SPU 列表` 并带上查询参数：

- `categoryPathPrefix`：类目路径前缀（用于筛选“当前类目 + 子类目”下的 SPU）
- `categoryId`：类目 ID（精确筛选；若同时传入则优先生效）

**接口参考**
- `GET /api/v1/admin/product/spus?categoryPathPrefix={pathPrefix}`
- `GET /api/v1/admin/product/spus?categoryId={categoryId}`

> 关系说明：类目与 **SPU** 关联（SPU 上维护 `categoryId/categoryPath`）；SKU 作为 SPU 的规格/变体，一般不单独挂类目，默认继承 SPU 的类目。

### 2) miniapp 获取可展示类目树（US1）

miniapp 的“开放只读接口”不要求管理员 JWT/RBAC，也不要求 customer token，但**必须携带租户上下文**（宿主鉴权上下文或 `tenant_uuid`）。

1. 直接调用：
   - `GET /api/v1/mini-app/categories/tree`
2. **验证**：
   - 返回的树中不包含已停用类目及其子树；
   - 树结构与后台最新层级/排序一致。

如需了解完整的 mini-app 接口（含商品列表、SKU 列表、可选登录），参见：`docs/guides/features/product/miniapp_open_api.md`。

### 3) 配置类目模板并驱动 SPU 录入校验（US2）

#### 3.1 创建与发布模板

**UI 路径**：Web Admin →「产品 → 类目模板（Category Templates）」。

1. 点击「新建模板」。
2. 添加字段（示例）：
   - `fieldKey=material`，`fieldType=string`，`required=true`
   - `fieldKey=originCountry`，`fieldType=string`，`required=false`
3. 发布模板版本（publish）；如误发布，使用回滚（rollback）。

#### 3.2 生效模板继承/覆盖验证

1. 给父类目发布模板。
2. 子类目不绑定模板时：其“生效模板”应继承最近祖先已发布模板。
3. 子类目绑定并发布自己的模板时：应完全覆盖祖先模板（不做字段合并）。

**接口参考**
- 模板列表/创建/更新：`GET|POST|PATCH /api/v1/admin/product/category-templates`
- 发布/回滚：`POST /api/v1/admin/product/category-templates/{id}/publish`、`POST /api/v1/admin/product/category-templates/{id}/rollback`
- 生效模板：`GET /api/v1/admin/product/category-templates/effective?categoryId={id}`
- 预览/模拟：`GET /api/v1/admin/product/category-templates/{id}/preview`、`POST /api/v1/admin/product/category-templates/{id}/simulate`
- 影响范围/批量重检触发：`GET /api/v1/admin/product/category-templates/{id}/impact`、`POST /api/v1/admin/product/category-templates/{id}/impact/recheck`

#### 3.3 SPU 录入校验（attributes）

SPU 新建/编辑请求支持 `attributes`，并按类目生效模板校验：

- `POST /api/v1/admin/product/spus`（或对应编辑接口）
- Body 示例（缺失必填会返回可定位字段错误）：

```json
{
  "categoryId": "<category-id>",
  "attributes": {
    "material": "cotton"
  }
}
```

### 4) 维护渠道类目映射 + CSV 批量（US3）

**UI 路径**：Web Admin →「产品 → 类目（Categories）」→ 选择类目 →「渠道映射」Tab。

1. 新增/编辑映射：填写 `channel` 与 `platformCategoryId`（可选：`strategy`、`syncStatus`、`metadata`）。
2. 删除映射：在列表中删除，或通过 API 传 `operation=delete`。
3. CSV 导入：选择 CSV 文件上传。
4. CSV 导出：下载当前类目的映射 CSV。

**接口参考**
- 查询/写入：`GET /api/v1/admin/product/categories/{id}/mappings`、`POST /api/v1/admin/product/categories/{id}/mappings`
- 导入（multipart）：`POST /api/v1/admin/product/categories/import?kind=mappings&categoryId={id}`
- 导出（返回 CSV）：`POST /api/v1/admin/product/categories/export`

**CSV 格式**
- 导入至少需要 header：`channel,platformCategoryId`（其余可选：`categoryId,strategy,syncStatus,metadata`）
- 导出 header：`categoryId,channel,platformCategoryId,strategy,syncStatus,metadata`

### 5) 审计查询（US3）

**UI 路径**：Web Admin →「产品 → 类目（Categories）」→「审计」Tab。

**接口参考**
- `GET /api/v1/admin/product/categories/{id}/audit?limit=50`

## 回归测试提示

- 后端：`make lint`、`make test`
- 前端：`make build-admin`

## 常见故障与排障

| 现象 | 排查建议 |
| --- | --- |
| 401（未登录/无 token） | 管理端补齐 `Authorization: Bearer <admin_jwt>`；miniapp 先走 `/mini-app/auth/login` 获取 customer token |
| 403（权限不足） | 检查 RBAC 资源权限是否包含类目/模板/映射/导入导出（见“依赖与权限”表） |
| 409/400（唯一性冲突） | 常见于类目 code/alias/slug 或映射唯一键冲突；根据返回字段定位并调整后重试 |
| CSV 导入失败 | 检查 header 是否包含 `channel,platformCategoryId`；`metadata` 必须是合法 JSON 字符串或空 |

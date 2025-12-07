# 权限与角色 PRD

> 覆盖 `web-admin/app/pages/settings/roles.vue` 以及权限配置组件。负责 RBAC、权限模板、权限范围、审批、审计。

## 1. 目标
- 实现细粒度、可配置的权限与角色管理，支持多租户、审批、审计，保证安全与合规。

## 2. 信息架构
- 角色列表：角色名称、描述、成员数、创建人、更新时间、状态。
- 权限树：按模块（商品、定价、渠道、库存、营销、财务）划分，支持读/写/审批等操作级别。
- 角色详情：分配权限、成员、审批流程、审计历史。
- 模板库：预置角色（管理员、运营、财务、仓库等）。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 角色管理 | 新增/编辑/复制角色，启用/停用 |
| 权限配置 | 选择模块权限、多级操作（View/Edit/Approve/Export）|
| 成员管理 | 添加/移除用户、分配角色、有效期、双重认证 |
| 审批 | 角色赋权需审批，保留记录 |
| 审计日志 | 记录所有权限变更、赋权操作，与 `admin_console_audit_events` 对齐 |
| 模板 & 导入 | 预置模板、导出/导入角色配置 |
| API Key 权限 | 与开发者设置联动，为 API Key/Connector 分配权限 |

## 4. 数据 & API
- 表：`roles`、`role_permissions`、`role_members`、`permission_templates`、`role_audit_logs`。
- API：`GET/POST /api/settings/roles`、`PATCH /api/settings/roles/{id}`、`POST /api/settings/roles/{id}/members`。

## 5. 权限 & 审计
- 权限：`settings.roles.manage`、`settings.roles.view`。
- 审计：角色变更、权限赋予、成员调整。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 权限变更审批周期 | ≤ 2 天 |
| 审计覆盖率 | 100% |

## 7. Backlog
- 权限可视化（按功能、流程）。
- 动态权限（基于审批、任务）。
- 与第三方 IAM 对接（SSO、SCIM）。

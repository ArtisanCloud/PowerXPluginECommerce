# 客户列表 & 会员视角页面 PRD

> 作用范围：`web-admin/app/pages/customer/index.vue`、`customer/members.vue` 以及其引用的 `CreateCustomerModal.vue`、`EditCustomerModal.vue`、`ViewCustomerModal.vue` 等组件。

## 1. 目标与假设
- 为客服、运营提供可操作的客户表格视图，覆盖创建、查看、编辑、删除等基础动作，并支持批量处理。
- 让会员运营可以以“等级 / 成长值”视角筛选客户，关注保级与权益使用。
- 保障所有关键动作（导入、导出、批量操作）均可审计，敏感字段遵循权限控制。

> **2025-12-08 更新**：当前插件落地版本仅实现“列表/详情/批量导入导出/会员筛选/提醒”等**读取与批量任务**能力。由于宿主 CRM 仍负责主数据创建，`POST /api/admin/customers`、`PATCH /api/admin/customers/{id}` 等 CRUD API 暂未开放，界面上的 “添加/编辑客户” 也处于占位状态。后续若开放宿主写入权限，再逐步补齐创建/更新/删除流程。

**基础假设**
1. 客户实体已在 CRM 服务中沉淀，提供 REST API；列表数据允许分页搜索。
2. 会员等级/积分来源于 `customer/membership/**` 模块，对接同一成长值计算服务。
3. 操作日志通过 `admin_console_audit_events` 记录；权限由 `/settings/roles.vue` 派生。

## 2. 页面结构
| 页面 | 描述 | 入口 |
| --- | --- | --- |
| 客户管理（列表） | 展示全部客户，支持筛选、批量操作、详情抽屉 | `/customer` | 
| 会员视角 | 以会员等级、成长值、积分为维度的客户表格 | `/customer/members` |

### 2.1 客户列表
- **顶部操作区**：标题、全局操作按钮（创建、导入、导出、批量发券、分配负责人）。导入按钮唤起 `CustomerImportDialog`，导出按钮唤起 `CustomerExportDialog`。
- **筛选器**：
  - 基础条件：关键词（姓名/手机号/邮箱）、来源渠道、会员等级、客户类型（个人/企业）、标签、注册日期范围。
  - 高级筛选抽屉：最近订单时间、所在地区、风险等级、合规状态。
  - 支持保存筛选组合为“视图”，在侧边栏快速切换。
- **表格字段**：客户 ID、姓名、联系方式（Mask）、会员等级、客户类型、来源渠道、注册日期、最近订单金额/时间、状态、标签、负责人。
- **批量操作**：
  - 批量加/减标签。
  - 批量发券/发权益。
  - 批量指派负责人。
  - 批量导出（触发后台任务）。
  - 批量禁用账号（需审批）。
- **行级操作**：查看（打开详情抽屉）、编辑（打开表单）、删除/禁用（带确认）、更多操作（重置密码、发送短信）。
- **详情抽屉内容**：基本资料、会员信息、最近订单、售后记录、积分/成长值变化、互动时间线、备注、内部任务。

### 2.2 会员视角
- **筛选**：会员等级、成长值区间、积分区间、权益包使用情况、保级状态（安全/即将降级/已降级）。
- **指标卡**：顶部显示会员总数、活跃会员、即将降级人数、平均成长值等。
- **表格字段**：会员 ID、昵称、会员等级、成长值、积分、权益包（显示可用/已用）、加入日期、最近一次权益使用、状态。
- **操作**：
  - 查看：跳转到客户详情抽屉并聚焦“会员信息” Tab。
  - 编辑：打开会员权益配置 / 成长值调整弹窗。
  - 批量动作：发送保级提醒、批量发放成长任务、导入/导出保级名单。
  - 指标埋点：筛选耗时（`customer_membership_fetch`）、提醒成功率（`customer_reminder_task` + `recordReminderResult`）。

## 3. 交互与流程
1. **创建客户（待宿主开放）**
   - 目前仅保留 `CreateCustomerModal` 设计草稿，未接入 API。实际创建仍需在宿主 PowerX IAM/CRM 中执行，再由插件同步列表。
- 当宿主提供 `POST /api/admin/customers` 写入能力后，以上 3 步流程才会启用，并写入审计日志（依赖宿主上下文）。
2. **查看详情**
   1. 在表格点击“查看” → 打开 `ViewCustomerModal`/抽屉并加载详细数据。
   2. 抽屉包含 Tab（概览、订单、售后、权益、备注、审计日志）。
   3. 支持在抽屉内直接发券/添加备注，完成后刷新局部数据。
3. **会员保级提醒**
   1. 会员视角筛选“即将降级” → 勾选若干客户 → 点击“批量提醒”。
   2. 弹窗选择渠道（短信/邮件/站内信）与模板 → 确认后调用 `POST /api/customers/bulk-remind`。
   3. 任务以异步方式执行，结果写入通知中心 + 审计。
4. **批量导出**
   1. 点击导出按钮或批量勾选后导出 → 选择字段/文件格式、确认筛选摘要。
   2. 创建导出任务，生成任务 ID 并显示在“任务中心/通知”，同时记录 KPI。
   3. 导出完成后提供下载链接（需要权限校验），并在 `AuditHistory` 中可追溯；如签名链接失效可在任务卡片里重触发。

## 4. 数据模型与 API
| 功能 | API (示例) | 状态 |
| --- | --- | --- |
| 查询客户列表 | `GET /api/admin/customers?keyword=&tier=` | ✅ 已实现（当前版本主力能力） |
| 获取客户详情 | `GET /api/admin/customers/{id}` | ✅ 抽屉侧重展示 360 信息（静态 Mock / 待真实数据） |
| 批量操作 | `POST /api/admin/customers/bulk-actions` | ✅ 已接入：标签增删、指派负责人、批量禁用（写宿主任务服务） |
| 会员列表 | `GET /api/admin/customers/members?tier=&growthRange=` | ✅ 已实现，用于会员视角指标与表格 |
| 会员操作 | `POST /api/admin/customers/{id}/membership` | 🚧 仅保级提醒/批量任务；直接调整成长值的 API 待宿主实现 |
| 导入任务 | `POST /api/admin/customers/import` | ✅ 已实现，走任务中心链路 |
| 导出任务 | `POST /api/admin/customers/export` | ✅ 已实现，字段配置 + 审计 |
| **创建客户** | `POST /api/admin/customers` | ⏳ **未开放**。需宿主 CRM 允许插件写入后再接入（详见「Backlog」） |
| **更新/删除客户** | `PATCH /api/admin/customers/{id}` / `DELETE ...` | ⏳ **未开放**。当前仅保留 UI 设计稿 |

**数据字段**：
```
Customer {
  id: string,
  name: string,
  type: "individual" | "enterprise",
  email: string,
  phone: string,
  country: string,
  membershipTier: string,
  growthValue: number,
  points: number,
  lastOrderAmount: number,
  lastOrderAt: string,
  status: "active" | "inactive" | "blocked",
  source: "website" | "miniapp" | "offline" | ...,
  tags: string[],
  accountManager: string,
  createdAt: string,
  updatedAt: string
}
```

## 5. 权限、审计与任务中心
- 权限项：
  - `customer.read`：查看列表与详情。
  - `customer.manage`：创建/编辑/批量操作。
  - `customer.export`：导出与导入审批。
  - `customer.sensitive.read`：查看完整联系方式。
- 审计：
- 所有 `POST/PATCH/DELETE` 请求需写入宿主审计日志，插件仅依赖宿主提供的上下文，无需额外自定义头。
  - 导入/导出、批量操作记录数据范围（筛选条件、数量）并写入任务中心；前端会在 `bulkTasks` store 中保存任务状态，供 UI 显示和轮询。
  - 详情抽屉展示最近 5 条审计记录，跳转 `/audit` 查看全部。
- 任务中心：
  - 导入/导出、保级提醒、批量操作都会返回 `taskId`，页面自动调用 `/jobs/{id}` 轮询直到成功/失败。
  - 任务卡片需包含动作名称、筛选摘要、审计 ID、下载链接（如有）。

## 6. 验收标准
- [ ] 客户列表支持多条件筛选、保存视图、批量选择、列配置。
- [ ] 详情抽屉可展示多 Tab，并支持操作快捷入口。
- [ ] 会员视角可按等级/成长值筛选，提供保级提醒、批量发放功能。
- [ ] 导入/导出全链路可测，附带任务状态、审计记录。
- [ ] 权限/审计/敏感信息 Mask 达到安全要求。

## 7. Backlog / 下一步
- 视图共享：允许用户将自定义视图分享给团队。
- 智能分群：结合 `customer/analytics/segmentation.vue` 输出动态人群，并在列表中作为筛选项。
- 任务协作：在详情抽屉嵌入任务评论，与 `operations` 模块联动。
- 客服工单：在时间线中引用客服系统（如 Zendesk）的工单摘要，支持跳转。
- **宿主 CRUD 能力对接**：待 PowerX IAM/CRM 开放 `POST/PATCH/DELETE /api/admin/customers` 后，恢复前文“创建/编辑/删除”流程，并把 `CreateCustomerModal`、`CustomerDetailDrawer` 中的编辑入口与真实 API 接通。

## 8. 客户 CRUD 实现方案（规划）

以下方案用于宿主开放写入权限后的落地执行，包含后端 API、前端交互、权限/审计与数据一致性要求：

### 8.1 后端 API 设计

1. **路由**（`backend/internal/transport/http/customer/routes.go`）
   - `POST /api/admin/customers` → `handler.CreateCustomer`
   - `PATCH /api/admin/customers/:id` → `handler.UpdateCustomer`
   - `DELETE /api/admin/customers/:id` → `handler.DeleteCustomer`
   - 对应 RBAC：`customer.manage`（创建/更新），`customer.delete`（删除，可与 manage 合并）。
2. **Handler 逻辑**
   - 入参校验：使用 `binding` + `validate.Struct`，约束邮箱/手机号格式、会员等级枚举、标签长度 ≤ 20 等。
   - 业务处理：调用 `internal/services/customer.Service`，再由 service 与宿主 CRM gRPC/REST 交互；无权限或宿主失败时透传错误。
  - 审计：后端 handler 成功写入宿主后，通过统一日志记录 `CustomerChanged` 行为，便于排查。
3. **Service 层**
   - 新增 `Create`, `Update`, `Delete` 方法，封装宿主 API 调用；
   - 所有写操作需在成功后刷新/合并本地缓存（如有）并触发事件 `CustomerChanged`，供列表刷新。
4. **错误场景**
   - 并发写入：宿主返回 `409` 时，前端显示“记录已被更新，请刷新后重试”；
   - 删除保护：若客户存在待处理订单，宿主返回 `400`，前端展示原因。

### 8.2 前端交互

1. **创建**
   - `CreateCustomerModal` 改为 `UForm` + schema 校验，提交调用 `useCustomerService().createCustomer`；成功后 toast + 关闭 modal + 触发 `store.fetchCustomers({ page: 1 })`。
2. **编辑**
   - 在 `CustomerDetailDrawer`、`CustomerTable` 行操作中启用“编辑”按钮，复用 `EditCustomerModal`；modal 内加载详情数据，diff 提交 `PATCH`。
3. **删除/禁用**
   - `More` 菜单新增“删除”项，需二次确认输入客户名称；调用 `DELETE` 成功后刷新列表并清理 selection。
4. **表单字段映射**
   - 字段与 API 对齐：`name, type, phone, email, source, region, membershipTier, tags, accountManager, notes`。
   - 复杂字段（地址/生日）若宿主暂不支持，可先透传备注。

### 8.3 状态管理

- `useCustomerStore` 增加 `createCustomer`, `updateCustomer`, `deleteCustomer` action，内部调用 `customerService`，统一处理 loading/error/toast。
- 保存视图时若包含新字段（例如 region）同样更新 store 的默认 filters。

### 8.4 审计 & 权限

- `customer.manage` 允许 `POST/PATCH`，`customer.delete` 或 `customer.manage_advanced` 控制删除。
- 删除操作必须要求用户填写原因并一并提交，供宿主审计使用（无需额外自定义 Header）。

### 8.5 数据同步与回退

- 写操作成功后触发后台事件或消息（若宿主支持），同步刷新 `SavedView` 缓存与 `selection`。
- 若宿主 API 不可用，前端禁用创建/编辑按钮并在顶部 Banner 标示“客户写入能力暂不可用”。

### 8.6 测试计划

- **单元测试**：`internal/transport/http/customer/handler_test.go` 覆盖 201/400/403/500；service mock 宿主接口。
- **E2E**：Cypress 脚本 `customer-crud.cy.ts`，流程：创建 → 列表出现 → 编辑字段 → 删除并验证消失。
- **回归**：确保导入/导出仍可用（CRUD 不应影响任务链路）。

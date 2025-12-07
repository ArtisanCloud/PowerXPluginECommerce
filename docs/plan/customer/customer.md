# 客户列表 & 会员视角页面 PRD

> 作用范围：`web-admin/app/pages/customer/index.vue`、`customer/members.vue` 以及其引用的 `CreateCustomerModal.vue`、`EditCustomerModal.vue`、`ViewCustomerModal.vue` 等组件。

## 1. 目标与假设
- 为客服、运营提供可操作的客户表格视图，覆盖创建、查看、编辑、删除等基础动作，并支持批量处理。
- 让会员运营可以以“等级 / 成长值”视角筛选客户，关注保级与权益使用。
- 保障所有关键动作（导入、导出、批量操作）均可审计，敏感字段遵循权限控制。

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
- **顶部操作区**：标题、全局操作按钮（创建、导入、导出、批量发券、分配负责人）。
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
  - 批量动作：发送保级提醒、批量发放成长任务、导出保级名单。

## 3. 交互与流程
1. **创建客户**
   1. 点击“添加客户” → 弹出 `CreateCustomerModal` → 输入基础信息（姓名、类型、联系人、来源）。
   2. 可选：配置初始标签、会员等级、负责人。
   3. 保存后调用 `POST /api/customers`，成功刷新表格 + 右上角 Toast 提示，并写入审计。
2. **查看详情**
   1. 在表格点击“查看” → 打开 `ViewCustomerModal`/抽屉并加载详细数据。
   2. 抽屉包含 Tab（概览、订单、售后、权益、备注、审计日志）。
   3. 支持在抽屉内直接发券/添加备注，完成后刷新局部数据。
3. **会员保级提醒**
   1. 会员视角筛选“即将降级” → 勾选若干客户 → 点击“批量提醒”。
   2. 弹窗选择渠道（短信/邮件/站内信）与模板 → 确认后调用 `POST /api/customers/bulk-remind`。
   3. 任务以异步方式执行，结果写入通知中心 + 审计。
4. **批量导出**
   1. 点击导出按钮或批量勾选后导出 → 选择字段/文件格式。
   2. 创建导出任务，生成任务 ID 并显示在“任务中心/通知”。
   3. 导出完成后提供下载链接（需要权限校验），并在 `AuditHistory` 中可追溯。

## 4. 数据模型与 API
| 功能 | API (示例) | 说明 |
| --- | --- | --- |
| 查询客户列表 | `GET /api/customers?keyword=&tier=&channel=&page=` | 返回分页数据 + 汇总指标；支持排序（注册时间、最近订单） |
| 获取客户详情 | `GET /api/customers/{id}` | 包含基本信息、订单摘要、售后、积分、标签、运营任务 |
| 创建客户 | `POST /api/customers` | 传入基本信息 + 标签；成功返回客户 ID |
| 更新客户 | `PATCH /api/customers/{id}` | 支持局部更新（标签、等级、负责人、备注） |
| 批量操作 | `POST /api/customers/bulk-actions` | body: { action: "add-tags" | "assign-owner" | ... , ids: [], payload: {} } |
| 会员列表 | `GET /api/customers/members?tier=&growthRange=` | 返回会员视角字段，附带保级状态 |
| 会员操作 | `POST /api/customers/{id}/membership` | 调整等级/成长值/权益 |
| 导入任务 | `POST /api/customers/import` | 上传文件后返回任务 ID，任务状态查询 `GET /api/jobs/{id}` |
| 导出任务 | `POST /api/customers/export` | 根据筛选条件导出，需审计上下文 |

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

## 5. 权限与审计
- 权限项：
  - `customer.read`：查看列表与详情。
  - `customer.manage`：创建/编辑/批量操作。
  - `customer.export`：导出与导入审批。
  - `customer.sensitive.read`：查看完整联系方式。
- 审计：
  - 所有 `POST/PATCH/DELETE` 请求需带上 `X-Audit-Action`、`X-Audit-Resource`。
  - 导入/导出、批量操作记录数据范围（筛选条件、数量）。
  - 详情抽屉展示最近 5 条审计记录，跳转 `/audit` 查看全部。

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

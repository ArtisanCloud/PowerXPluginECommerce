# Feature Specification: Customer List & Membership Views

**Feature Branch**: `001-customer-ops-customer`  
**Created**: 2025-12-07  
**Status**: Draft  
**Input**: User description: "Customer list & membership views spec based on docs/plan/customer/customer.md"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Operate Customer Directory (Priority: P1)
客服或运营需要在客户列表中快速定位目标客户、执行单条或批量操作，并查看详情抽屉以完成日常支持。

**Why this priority**: 列表是所有后续动作的入口，无法精确检索与操作会直接阻塞客服、合规、营销等部门。

**Independent Test**: 通过手动创建过滤器、执行批量标签/负责人指派、打开详情抽屉验证信息完整性，即可验证该故事独立交付价值。

**Acceptance Scenarios**:

1. **Given** 用户拥有 `customer.read` 权限，**When** 使用关键词+高级筛选（地区、风险、注册时间）检索，**Then** 列表仅返回满足条件的客户并可保存为自定义视图。
2. **Given** 用户勾选多条客户记录，**When** 选择“批量指派负责人”并确认，**Then** 系统创建后台任务并在完成后显示审计记录与成功/失败反馈。
3. **Given** 用户点击某客户行的“查看”，**When** 抽屉打开，**Then** 展示概览、订单、售后、权益、备注、审计等多个 Tab，且联系方式遵循权限遮罩。

---

### User Story 2 - 会员视角洞察与保级 (Priority: P2)
会员运营需要从等级/成长值维度审视客户、识别即将降级人群，并触发互动（提醒、发放成长任务）。

**Why this priority**: 会员生命周期管理直接影响复购与 GMV，缺失会员视角会导致权益策略无法执行。

**Independent Test**: 指定过滤条件（等级、成长值区间），下发保级提醒并验证任务/日志即可完成独立测试。

**Acceptance Scenarios**:

1. **Given** 会员运营打开会员视角页面，**When** 设置“等级=金卡 & 成长值<50”筛选，**Then** 表格与顶部指标同步刷新并显示保级状态统计。
2. **Given** 用户选中“即将降级”分组，**When** 点击“批量提醒”并选择短信模板，**Then** 系统生成异步任务，通知中心与审计日志记录下发结果。

---

### User Story 3 - 可管控的导入/导出与审计 (Priority: P3)
管理员需要安全地批量导入客户、导出筛选结果，且每一次操作都有审计与权限控制。

**Why this priority**: 批量数据处理频繁且风险高，若缺乏流程与审计会触发合规问题。

**Independent Test**: 上传导入模板、触发导出并查看任务列表与审计记录即可验证功能端到端完成。

**Acceptance Scenarios**:

1. **Given** 用户具有 `customer.manage` 权限，**When** 下载模板并导入客户数据，**Then** 系统校验字段→生成导入任务→在任务中心显示进度→完成后刷新列表并记录审计。
2. **Given** 用户无 `customer.export` 权限，**When** 点击“导出”，**Then** 系统阻止操作并提示缺乏权限；有权限的用户发起导出后可在通知中心获取下载链接且链接带签名有效期。

---

### Edge Cases

- 无任何客户匹配筛选条件时，需要展示空状态提示并允许保存视图（不会报错或保留旧数据）。
- 导入文件格式错误或字段缺失时，系统应阻断操作并生成错误报告供下载，而不是部分写入。
- 当用户缺失 `customer.sensitive.read` 权限时，即使在详情抽屉也必须保持联系方式遮罩；尝试导出时需自动剔除敏感字段。
- 会员视角中某些客户没有成长值/积分数据时，需要以 `—` 显示并在指标卡中不计入平均值，防止数据显示异常。

### Scope Adjustment (2025-12-08)

- 宿主 PowerX CRM 暂未向插件开放 `POST/PATCH/DELETE /api/admin/customers` 等写入接口，因此 **当前迭代仅上线“列表/详情/批量任务/会员视角/导入导出”能力**。
- `CreateCustomerModal`、详情抽屉内的“编辑”入口仅保留 UI 占位，不会触发真实 API；所有新增/编辑仍需在宿主管理台完成并通过同步列表体现。
- 一旦宿主开放写入权限，本规范内的 CRUD 要求（FR-001~FR-004 的“创建/编辑/删除”部分）将恢复为强制交付内容，并补充 API 鉴权+审计覆盖。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 客户列表 MUST 支持组合筛选（关键词、来源、等级、类型、标签、日期、地区、风险），并允许保存/命名/切换视图。
- **FR-002**: 列表 MUST 提供可配置列与分页，联系方式字段在缺少 `customer.sensitive.read` 权限时必须遮罩。
- **FR-003**: 系统 MUST 提供详情抽屉，包含基本信息、订单概览、售后记录、积分/成长值流水、互动时间线、备注、最近 5 条审计记录，并允许发券、添加备注等快捷操作。
- **FR-004**: 列表 MUST 支持批量动作：批量标签增删、批量发券/发权益、批量指派负责人、批量导出、批量禁用账号（需审批），所有任务需异步执行并在任务中心反馈。
- **FR-005**: 系统 MUST 提供导入/导出流程：模板下载→上传校验→任务创建→结果通知；导出结果需基于当前筛选条件并生成短期有效的安全下载链接。
- **FR-006**: 会员视角 MUST 提供等级/成长值/积分/权益使用情况筛选，顶部指标卡展示总数、活跃、即将降级、平均成长值等指标。
- **FR-007**: 会员视角 MUST 允许批量动作（保级提醒、发放成长任务、导出名单），并在任务/审计日志中记录渠道、数量、结果。
- **FR-008**: 系统 MUST 依据权限控制导入、导出、批量操作、敏感字段访问，并在 `admin_console_audit_events` 中写入操作类型、范围、操作者、时间、筛选条件/文件名等上下文。
- **FR-009**: 当 API 返回错误（网络或业务）时，界面 MUST 提示失败原因并允许重试，不得 silently fail。

### Key Entities

- **Customer**: 包含 ID、基础信息、会员等级、成长值/积分、最近订单摘要、来源渠道、标签、负责人、状态、敏感联系方式。
- **SavedView**: 保存的筛选组合及列配置，记录创建者、共享范围、默认排序。
- **BulkTask**: 导入、导出、批量发券/指派任务，包含任务类型、关联筛选条件、状态、结果文件、审计 ID。
- **MembershipSnapshot**: 会员视角展示所需的等级/成长值/权益统计，包含保级状态、权益使用情况。

### Assumptions

1. CRM 目前仅对插件暴露“查询 / 批量任务” REST API；主数据创建仍由宿主控制。未来若开放写入，再恢复插件内的创建/更新流程。
2. 审计与任务中心能力已存在，可直接写入审计事件并显示任务状态。
3. 批量操作限制（数量/频率）由后端统一控制，前端仅展示反馈。

## Implementation Plan – Customer CRUD (Future Scope)

为确保一旦宿主开放写权限即可快速上线，制定如下实施方案：

1. **Backend Tasks**
   - 路由：在 `customer/routes.go` 注册 `POST/PATCH/DELETE /api/admin/customers`，受 `httpmw.EnsureTenant + RBAC` 保护。
   - Handler：新增 `CreateCustomer`, `UpdateCustomer`, `DeleteCustomer`，对接 `internal/services/customer.Service`，并在失败时返回结构化错误。
   - Service：实现 `Create/Update/Delete` 调用，封装宿主 CRM 客户写入 API；支持乐观锁（etag/version）并在冲突时返回 409。
   - Tests：`handler_test.go` & `service_test.go` 覆盖成功/校验失败/无权限/宿主错误场景。

2. **Frontend Tasks**
   - `useCustomerService` 新增 `createCustomer`, `updateCustomer`, `deleteCustomer`，统一处理创建/更新/删除所需的错误态。
   - `useCustomerStore` 添加对应 action，并在成功后刷新列表或更新本地 `list`/`selection`。
   - `CreateCustomerModal`、`EditCustomerModal` 升级为基于 `UForm` 的校验表单；`CustomerDetailDrawer` 开启编辑入口；行操作菜单新增“删除”，带确认对话框与原因输入。
   - E2E：新增 `tests/e2e/customer-crud.cy.ts`，覆盖创建→编辑→删除流程。

3. **Permissions & Audit**
   - 新增权限位：`customer.manage`（create/update）、`customer.delete`（可与 manage 合并）；`customer.sensitive.read` 继续控制敏感字段。
   - 所有写操作默认要求用户提供原因（删除/禁用），并写入 `admin_console_audit_events`。

4. **Rollout Checklist**
   - 与宿主团队确认写入 API SLA、速率限制、审计需求。
   - 开启功能旗标（如 `NUXT_PUBLIC_ENABLE_CUSTOMER_CRUD`），方便逐租户灰度。
   - 更新文档、演示脚本与客服培训材料。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 90% 的客服在 3 次点击内即可通过筛选或保存视图找到目标客户（以可用性测试记录为准）。
- **SC-002**: 批量导出 5,000 条客户数据的任务在 10 分钟内完成，且 100% 操作可在任务中心与审计日志中追溯。
- **SC-003**: 会员运营可在 5 分钟内完成“筛选即将降级会员 + 批量发送提醒”流程，且提醒任务成功率≥98%。
- **SC-004**: 导入/批量操作相关的合规/权限错误率低于 1%，所有无权限操作均提供明确提示并被审计。

# Feature Specification: Channel Master Data & Authorization

**Feature Branch**: `003-channel-master`  
**Created**: 2025-12-13  
**Status**: Draft  
**Input**: User description: "Channel master data & authorization based on docs/plan/channels/master-data.md"

## Clarifications

### Session 2025-12-13

- Q: 渠道健康度评分算法采用何种策略？ → A: 闸门式阈值优先（关键指标不达标即判定“差”，其余指标再加权）
- Q: 渠道主状态是否需要扩展到“数据中断”等运行状态？ → A: 主状态保持草稿/待审核/驳回/未授权/已授权/已停用六种，数据中断等以标签展示
- Q: 渠道记录是否需要跨租户共享？ → A: 每条渠道绑定单租户，不同租户可记录同一店铺 ID

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - 渠道主数据入驻 (Priority: P1)

渠道运营在渠道列表中创建或编辑店铺，填写平台、店铺 ID、区域、域名、联系人、负责人等主数据，并提交审批后入驻。

**Why this priority**: 没有经过审批的准确主数据，就无法触发授权、配置或 KPI 计算，渠道模块也无法管理跨平台资产。

**Independent Test**: 通过新建一个渠道、完成审批并在列表中看到准确状态即可单独验收。

**Acceptance Scenarios**:

1. **Given** 运营在渠道列表点击“新增渠道”，**When** 填写必填字段并提交，**Then** 渠道进入“待审核”状态并写入审计。
2. **Given** 审批者通过渠道申请，**When** 运营刷新列表，**Then** 渠道状态更新为“未授权”并显示负责人、区域等主数据信息。

---

### User Story 2 - 授权与凭证生命周期 (Priority: P1)

授权管理员对已入驻渠道执行平台授权、上传或刷新凭证、查看有效期与权限范围，并在凭证临期时收到告警。

**Why this priority**: 凭证有效性是多渠道同步的前提，失效将导致订单、库存、价格任务全部阻断。

**Independent Test**: 只需在某渠道完成授权表单→凭证保存→自动检测→触发到期提醒，即可衡量该流程。

**Acceptance Scenarios**:

1. **Given** 渠道状态为“未授权”，**When** 管理员完成平台授权或上传凭证，**Then** 渠道状态更新为“已授权”并展示授权范围与到期时间。
2. **Given** 凭证即将过期，**When** 系统检测到有效期 < 7 天，**Then** 自动发送告警并在渠道详情标注需刷新。

---

### User Story 3 - KPI、健康度与告警 (Priority: P2)

渠道运营在渠道详情页查看 GMV、订单量、增长率、库存覆盖、错误率等 KPI，看到健康度评分、趋势图及告警，并可点击调出任务或备注。

**Why this priority**: 提供统一指标和健康度可以让运营快速发现异常、安排任务，直接影响多渠道 GMV。

**Independent Test**: 为单个渠道加载 KPI 面板并模拟异常，确认指标、评分与告警展示准确即可独立测试。

**Acceptance Scenarios**:

1. **Given** 渠道拥有最新的 GMV/订单/错误率数据，**When** 打开详情页，**Then** 显示指标卡、趋势、健康度评分以及最近同步时间。
2. **Given** 指标触发阈值（如错误率 > 5%），**When** 系统计算健康度，**Then** 生成告警项并允许运营查看处理记录。

---

### User Story 4 - 策略配置与团队权限 (Priority: P3)

渠道经理可为渠道关联定价、库存、物流、客服策略，并配置负责小组、审批人及可操作角色，同时记录操作日志。

**Why this priority**: 策略与团队信息保障跨模块一致性，确保后续上架、定价、客服动作有人负责且有据可查。

**Independent Test**: 仅通过在单个渠道中更新策略与团队，再验证权限与审计即可确认该能力。

**Acceptance Scenarios**:

1. **Given** 渠道已授权，**When** 经理配置 pricebook、库存策略、客服 SLA，**Then** 设置被保存并对相关模块可见。
2. **Given** 新的团队成员被指派操作权限，**When** 他尝试进入渠道详情，**Then** 仅能访问授权范围内的操作，且审计记录显示角色与时间。

### Edge Cases

- 渠道需要支持线下或半手动渠道：无平台授权但仍需创建记录并允许上传线下凭证。
- 同一平台存在多个店铺，店铺 ID 重复时需阻止并提示关联关系。
- 凭证或授权失败导致 KPI 暂停更新时，健康度应显示“数据中断”并产生提醒。
- 数据同步长时间未执行（>24 小时）需在列表展示警示状态。

### Assumptions

- Connector 框架已能与主要平台交互，渠道模块只需调度现有任务中心。
- 手动录入渠道默认由审批人审核真实性；审核 SLA 由运营团队定义，不在本需求限定。
- KPI 计算依赖已有数据仓或任务中心的指标汇总，本次仅负责读取与展示。

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: 渠道列表 MUST 展示平台、主状态（草稿 / 待审核 / 驳回 / 未授权 / 已授权 / 已停用）、负责人、区域、GMV、订单、异常数量、最近同步时间，并支持平台 / 状态 / 负责人 / 区域 / GMV / 健康度筛选；诸如“数据中断”“告警”等运行状态需以标签或标记展示。
- **FR-002**: 系统 MUST 提供新增、编辑、停用渠道表单，校验必填字段与租户维度下的唯一性（允许不同租户录入相同店铺 ID），并在提交后触发渠道审批流程。
- **FR-003**: 渠道详情 MUST 显示完整主数据（店铺 ID、域名、联系人、客服信息等）并允许修改时写入审计。
- **FR-004**: 授权流程 MUST 支持平台 OAuth/跳转、回调保存凭证、手动录入（含文件上传）、权限范围记录以及授权测试结果；对线下/半手动渠道需允许直接上传线下凭证并跳过 OAuth。
- **FR-005**: 系统 MUST 加密存储凭证、记录有效期、显示刷新时间、允许手动刷新，并在凭证将到期、被撤销或刷新失败时发出告警。
- **FR-006**: 渠道页面 MUST 允许配置定价策略、库存策略、物流策略、客服 SLA、手续费、账期、结算方式等，并与已有 pricebook/库存模块引用保持一致。
- **FR-007**: KPI 区域 MUST 展示 GMV、订单、GMV 环比、库存覆盖率、错误率等指标，提供趋势图、健康度评分以及 KPI 来源时间戳。
- **FR-008**: 系统 MUST 维持渠道健康度评估逻辑（例如凭证状态、同步成功率、错误率），采用“闸门式阈值”：若关键指标（凭证失效、同步成功率 < 95%、错误率 > 5% 等）任一不达标则健康度直接标记为“差”，否则再按次要指标加权评分，并生成可操作告警列表。
- **FR-009**: 渠道详情 MUST 显示关联任务、异常、告警、备注，并允许运营添加/解除与任务中心任务的关联、追加结构化备注，所有操作写入审计。
- **FR-010**: 团队与权限模块 MUST 支持分配负责人、操作角色、审批人，并限制用户操作范围，同时提供 audit trail。
- **FR-011**: 通知机制 MUST 对授权即将过期、API 错误、健康度下降、库存不足等事件触发提醒（站内与可选的外部通知）。
- **FR-012**: 系统 MUST 记录渠道同步历史（包含手动/自动触发时间、触发人、耗时、结果、来源），并在详情页展示最近 N 次记录，同时允许运营手动触发一次同步或刷新。
- **FR-013**: 所有渠道的新增、修改、授权、配置、告警处理操作 MUST 写入 `channel_audit_logs`，可被审计与导出。

### Key Entities *(include if feature involves data)*

- **ChannelMaster**: 描述渠道/店铺主数据（名称、平台、店铺 ID、域名、区域、负责人、联系人、状态、审批信息）并与单一租户/工作区绑定，允许不同租户保存同店铺 ID 的独立配置。
- **ChannelCredential**: 储存授权凭证（类型、密文、权限范围、有效期、刷新 Token、最近检测结果、测试日志）。
- **ChannelConfig**: 关联策略（pricebook 引用、库存策略 ID、物流策略 ID、客服 SLA、手续费、账期、结算方式）及渠道参数。
- **ChannelMetric**: KPI 汇总（GMV、订单、增长率、库存覆盖、错误率、健康度评分、时间戳、来源）。
- **ChannelAlert**: 告警及通知记录（类型、触发条件、状态、处理人、处理时间、关联任务）。
- **ChannelTaskLink**: 渠道与任务中心任务的关联记录（channel_id、task_id、task_source、关联状态、关联人、备注、created_at、resolved_at）。
- **ChannelNote**: 渠道运营备注（channel_id、note_id、作者、正文、可见性、created_at），支撑运营沟通与审计。
- **ChannelSyncHistory**: 渠道同步历史记录（channel_id、trigger_type、triggered_by、duration_ms、result、payload_summary、created_at）。
- **ChannelAuditLog**: 审计事件（动作、操作人、时间、详情、来源模块），覆盖 CRUD、授权、配置、告警处理。

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 渠道入驻（创建→审批通过）平均耗时 ≤ 10 个自然日。
- **SC-002**: 有效凭证覆盖率 ≥ 99%，且即将到期的告警在 1 小时内送达负责人。
- **SC-003**: 渠道数据同步成功率 ≥ 99%，失败任务在 2 小时内可查到原因并可重试。
- **SC-004**: 90% 的渠道在详情页可展示最近 30 天 GMV/订单/健康度趋势，指标加载时间 ≤ 3 秒。
- **SC-005**: 90% 的授权、配置、告警处理操作在审计日志中 1 分钟内可检索到，确保合规性。

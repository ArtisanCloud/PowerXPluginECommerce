# Data Model

## Customer (view model)
- **id**: string (UUID) – 唯一标识，来自 CRM。
- **name**: string – 客户姓名。
- **type**: enum {individual, enterprise} – 客户类型。
- **email / phone**: string – 联系方式；在无 `customer.sensitive.read` 权限时前端必须遮罩。
- **country / region**: string – 地理信息。
- **membershipTier**: string – 当前会员等级。
- **growthValue / points**: number – 成长值与积分。
- **lastOrderAmount / lastOrderAt**: number / ISO8601 string – 最近订单概要。
- **status**: enum {active, inactive, blocked}。
- **source**: enum {website, miniapp, offline, referral...}。
- **tags**: string[] – 运营标签。
- **accountManager**: string – 负责人。
- **createdAt / updatedAt**: ISO8601 string。
- **auditTrail**: AuditEntry[] – 最近 5 条审计（只在详情抽屉展示）。

## SavedView
- **id**: string – 本地生成 UUID。
- **name**: string – 视图名称。
- **filters**: JSON – 序列化的筛选条件。
- **columns**: string[] – 显示列。
- **owner**: string – 用户 ID。
- **shared**: boolean – 是否共享。
- **createdAt**: ISO8601 string。

## BulkTask
- **taskId**: string – 任务中心 ID。
- **type**: enum {import, export, bulkAssignOwner, bulkTag, bulkReminder, bulkDisable}。
- **scope**: object – 筛选条件摘要或客户 ID 列表。
- **status**: enum {queued, running, success, failed}。
- **createdAt / completedAt**: ISO8601 string。
- **auditId**: string – 对应审计事件。

## MembershipSnapshot
- **customerId**: string。
- **tier**: string。
- **growthValue**: number。
- **points**: number。
- **benefits**: {name: string, used: boolean}[]。
- **retentionStatus**: enum {safe, warning, downgrade}。
- **lastBenefitUsedAt**: ISO8601 string。

### Relationships
- `Customer` ↔ `MembershipSnapshot`: 1:1（仅在会员视角需要）。
- `Customer` ↔ `BulkTask`: M:N（批量操作记录任务；通过 scope/customerId 列表关联）。
- `SavedView` 独立存在，仅与当前用户关联。

### State Transitions
- Customer.status：active ↔ inactive（批量禁用/启用）；blocked 由风控处理。
- BulkTask.status：queued → running → success/failed（任务中心负责）。
- MembershipSnapshot.retentionStatus：safe ↔ warning ↔ downgrade（由成长值计算服务输出，前端只展示）。

# Data Model: 订阅对账与续费治理

## Entities

### 1) ReconciliationBatch（对账批次）

- Purpose: 表示一次租户账期对账执行及汇总结果。
- Key Fields:
  - id
  - tenant_uuid
  - billing_cycle (e.g. `2026-03-31`)
  - run_type (`daily`/`rerun`)
  - expected_amount_minor（净应收）
  - actual_amount_minor（实收）
  - delta_amount_minor（差异）
  - delta_count
  - status (`running`/`completed`/`failed`)
  - started_at / finished_at
  - created_by
- Validation Rules:
  - `(tenant_uuid, billing_cycle, run_type)` 唯一（幂等）
  - completed 时必须具备 finished_at

### 2) ReconciliationDelta（差异记录）

- Purpose: 记录单条订阅对账异常，作为处置入口。
- Key Fields:
  - id
  - tenant_uuid
  - batch_id
  - subscription_ref
  - bill_ref
  - payment_ref
  - delta_type (`missing_payment`/`duplicate_payment`/`amount_mismatch`/`status_mismatch`/`data_missing`)
  - risk_level (`high`/`medium`/`low`)
  - expected_amount_minor
  - actual_amount_minor
  - delta_amount_minor
  - reason_code
  - status (`open`/`processing`/`resolved`/`ignored`)
  - delta_fingerprint（去重指纹）
  - detected_at
- Validation Rules:
  - `delta_type` 必须属于标准五分类
  - `delta_fingerprint` 在同租户同账期内应稳定

### 3) DeltaTask（差异处置任务）

- Purpose: 将差异记录转化为运营执行单。
- Key Fields:
  - id
  - tenant_uuid
  - delta_id
  - assignee
  - priority
  - sla_level (`high`/`medium`/`low`)
  - sla_deadline
  - status (`pending`/`in_progress`/`closed`)
  - resolution (`fixed`/`escalated`/`false_positive`)
  - resolution_note
  - closed_at
- Validation Rules:
  - 同一 `delta_fingerprint` 仅允许一个未关闭任务
  - `sla_level=high|medium|low` 对应 deadline 24h/48h/72h

### 4) RenewalGovernancePolicy（续费治理策略）

- Purpose: 定义重试与升级规则。
- Key Fields:
  - id
  - tenant_uuid
  - enabled
  - retry_windows (`1h`,`24h`,`72h`,`7d`)
  - escalation_threshold（连续失败阈值）
  - notify_channels
  - version
  - effective_from / effective_to
- Validation Rules:
  - 默认包含四段递增窗口
  - 生效区间不可重叠

### 5) RenewalExecutionLog（续费治理执行记录）

- Purpose: 记录每次重试/通知/升级动作，用于审计与看板。
- Key Fields:
  - id
  - tenant_uuid
  - subscription_ref
  - action_type (`retry`/`notify`/`escalate`)
  - attempt_no
  - scheduled_at / executed_at
  - result (`success`/`failed`/`skipped`)
  - failure_reason
  - operator_type (`system`/`manual`)
  - operator_id
- Validation Rules:
  - retry 的 `attempt_no` 必须递增
  - result=failed 时建议填写 failure_reason

## Relationships

- ReconciliationBatch 1:N ReconciliationDelta
- ReconciliationDelta 1:N DeltaTask（仅 1 个 active）
- RenewalGovernancePolicy 1:N RenewalExecutionLog（按版本追踪）
- ReconciliationDelta 与 RenewalExecutionLog 可通过 `subscription_ref` 关联用于归因看板

## State Transitions

### ReconciliationBatch
- `running -> completed|failed`

### ReconciliationDelta
- `open -> processing -> resolved`
- `open|processing -> ignored`

### DeltaTask
- `pending -> in_progress -> closed`

## Derived Metrics

- 差异率 = `delta_count / 对账账单总数`
- 恢复率 = `重试成功订阅数 / 续费失败订阅数`
- SLA 达成率 = `在 SLA 截止前关闭任务数 / 已关闭任务总数`

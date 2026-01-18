# Data Model: 营销支付与小程序支付

## Entities

### PaymentProvider（支付渠道）
- Fields: id, tenant_uuid, name, type, status, fee_rate, settlement_cycle, currency, credentials, risk_policy, created_at, updated_at
- Notes: 需支持启用/停用与多币种配置。

### PaymentTransaction（支付单/交易）
- Fields: id, tenant_uuid, transaction_no, order_id, order_no, provider_id, pay_method, amount_total, amount_currency, fee_amount, status, created_at, completed_at, failure_reason, risk_flag, metadata
- Relationships: belongs to PaymentProvider; references Order
- State transitions:
  - pending_payment -> paying -> paid
  - pending_payment -> canceled
  - paying -> failed
  - paid -> refunded (full/partial)

### PaymentReconciliation（对账批次）
- Fields: id, tenant_uuid, period_type (daily/weekly), period_start, period_end, diff_count, diff_total_amount, status, processed_by, processed_at
- Relationships: has many PaymentReconciliationItem

### PaymentReconciliationItem（对账差异项）
- Fields: id, tenant_uuid, reconciliation_id, transaction_id, diff_type, diff_amount, resolution, resolved_by, resolved_at
- Relationships: belongs to PaymentReconciliation; references PaymentTransaction

### PaymentRefund（退款记录）
- Fields: id, tenant_uuid, transaction_id, refund_no, refund_amount, refund_currency, status, reason, created_at, completed_at
- Relationships: belongs to PaymentTransaction
- State transitions:
  - requested -> processing -> success
  - requested -> failed

### PaymentRiskEvent（风险事件）
- Fields: id, tenant_uuid, transaction_id, risk_type, risk_score, action, created_at, resolved_at, resolution_note
- Relationships: belongs to PaymentTransaction

### PaymentSplitRule（分账规则）
- Fields: id, tenant_uuid, name, participants, ratio, status, created_at, updated_at

### PaymentSplitResult（分账结果）
- Fields: id, tenant_uuid, transaction_id, rule_id, participant, amount, status, created_at
- Relationships: belongs to PaymentTransaction; references PaymentSplitRule

### PaymentManualReview（手动收款审核）
- Fields: id, tenant_uuid, order_id, order_no, transaction_id, provider_id, pay_method, amount_minor, currency, status, submitted_by, submitted_at, reviewed_by, reviewed_at, review_reason, proof_no, note, created_at, updated_at
- Relationships: references Order; optionally references PaymentTransaction

## Relationships Summary
- PaymentProvider 1..* PaymentTransaction
- PaymentTransaction 1..* PaymentRefund
- PaymentTransaction 1..* PaymentRiskEvent
- PaymentTransaction 1..* PaymentSplitResult
- PaymentReconciliation 1..* PaymentReconciliationItem
- Order 1..* PaymentManualReview

## Enumerations
- PaymentTransaction.status: pending_payment, paying, paid, failed, canceled, timeout, refunded
- PaymentRefund.status: requested, processing, success, failed
- PaymentRiskEvent.action: block, review, allow
- PaymentReconciliation.status: pending, processing, completed, failed
- PaymentManualReview.status: pending_review, approved, rejected

## Validation Rules
- amount_total > 0
- refund_amount <= amount_total
- fee_rate >= 0
- status values constrained to defined transitions
- tenant_uuid is required on all entities

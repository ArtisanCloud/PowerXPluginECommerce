# Data Model: 小程序支付

## Entities

### PaymentTransaction（支付单/交易）
- Fields: id, tenant_uuid, transaction_no, order_id, order_no, provider_id, pay_method, amount_total, amount_currency, fee_amount, status, created_at, updated_at, completed_at, failure_reason, client, idempotency_key, metadata
- Relationships: references Order
- State transitions:
  - pending_payment -> paying -> paid
  - pending_payment -> canceled
  - paying -> failed

### PaymentRiskEvent（风险事件）
- Fields: id, tenant_uuid, transaction_id, risk_type, risk_score, action, created_at, resolved_at

### CustomerIdentity（第三方账号绑定）
- Fields: id, tenant_uuid, customer_id, provider, subject, union_id, metadata, created_at, updated_at
- Notes: provider=wechat/dingtalk/alipay；subject 存 openid 等第三方主体 ID；union_id 可空

## Enumerations
- PaymentTransaction.status: pending_payment, paying, paid, failed, canceled, timeout
- PaymentRiskEvent.action: block, review, allow

## Validation Rules
- amount_total > 0
- tenant_uuid is required on all entities
- order_id and order_no required
- idempotency_key required for create

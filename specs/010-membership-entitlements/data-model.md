# Data Model: 订阅会籍权益代币发放

## Entities

### MembershipTier
- id
- name
- status
- rules (可选)

### MembershipAssignment
- id
- tenant_uuid
- customer_id
- tier_id
- status (active/expired/canceled)
- valid_from
- valid_to
- source_type (subscription)
- source_id (order_id/transaction_id)

### Benefit
- id
- name
- type (single/bundle)
- items[]
  - service_code
  - quantity (int, -1 表示无限)
  - valid_days (int, 0/空表示永久)
  - stack_policy (stack/replace/max)

### Entitlement
- id
- tenant_uuid
- customer_id
- service_code
- quantity
- valid_from
- valid_to
- stack_policy
- source_type (subscription_plan/transaction/order)
- source_id

### TokenAccount
- id
- tenant_uuid
- customer_id
- token_code
- balance
- updated_at

### TokenTransaction
- id
- tenant_uuid
- customer_id
- token_code
- delta
- source_type (subscription/transaction/order)
- source_id
- created_at

## Relationships
- SubscriptionPlan(metadata) → MembershipTier/Benefit/TokenCode
- MembershipAssignment ↔ Customer (1:N)
- Benefit(bundle) → Entitlement (N)
- TokenAccount ↔ TokenTransaction (1:N)

## Validation Rules
- quantity >= -1, -1 表示无限
- token_amount > 0 时必须有 token_code
- valid_to >= valid_from
- 幂等键优先级：transaction_id → out_trade_no → order_id

## State Transitions
- MembershipAssignment: active → expired/canceled
- Entitlement: valid → expired

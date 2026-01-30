# Research Notes: 订阅会籍权益代币发放

## Decision 1: 最小落地实体
**Decision**: 增加最小实体 `entitlements`、`token_accounts`、`token_transactions`，复用既有 `membership_*` 表。
**Rationale**: 订阅权益与代币发放需要可消费实例与账本，独立实体可支持幂等与审计。
**Alternatives considered**: 直接把权益/代币写入订单 metadata（缺乏可查询与扣减能力）。

## Decision 2: 幂等键优先级
**Decision**: 幂等键优先使用 `transaction_id`，回退 `out_trade_no`，再回退 `order_id`。
**Rationale**: 兼容回调与订单侧唯一性，避免重复发放。
**Alternatives considered**: 仅使用 `order_id` 或 `out_trade_no`（在多交易或重试场景下易冲突）。

## Decision 3: 权益叠加与有效期
**Decision**: 默认叠加策略为 `stack`，无限权益使用 `quantity = -1`。
**Rationale**: 符合订阅额度可累加的直觉，且与数量型/代币型统一。
**Alternatives considered**: `replace` 或 `max` 作为默认（不符合多数订阅预期）。

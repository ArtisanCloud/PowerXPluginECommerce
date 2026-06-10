# Phase 0 Research — 完整优惠券机制（结论汇总）

## Decision 1: 优惠计算权威时机
- **Decision**: 在订单提交时进行服务端最终复算；展示层仅作预估。
- **Rationale**: 避免前后端金额漂移，且便于将券占用与订单写入放入同一事务边界。
- **Alternatives considered**:
  - 查价阶段直接算券后价：展示一致性更高，但与用户资产上下文耦合过重。
  - 双引擎并行：体验更好，但一致性治理复杂度明显增加。

## Decision 2: 券状态机
- **Decision**: `available -> reserved -> redeemed`，失败路径 `reserved -> available`，退款可选 `redeemed -> refunded`。
- **Rationale**: 与订单支付状态天然对齐，便于运营理解与对账。
- **Alternatives considered**:
  - 提交即核销：支付失败回滚成本高。
  - 支付前不预占：并发下容易重复消费。

## Decision 3: 叠加策略与计算顺序
- **Decision**: 由模板声明可叠加/互斥/优先级，计算顺序“先商品级券，再订单级券，同层按优先级”。
- **Rationale**: 兼顾运营灵活性与解释性，支持部分退款分摊。
- **Alternatives considered**:
  - 全局最大优惠优先：可解释性较差，调试成本高。
  - 单券模型：实现简单但业务价值受限。

## Decision 4: 金额分摊与退款返券
- **Decision**: 订单级优惠按订单行金额占比分摊；退款返券按模板策略（默认不返券）。
- **Rationale**: 保证售后与财务对账口径统一，降低后续补偿歧义。
- **Alternatives considered**:
  - 不分摊：部分退款无法准确核算。
  - 全额退款才返券：规则简单但不满足运营灵活诉求。

## Decision 5: 有效期与释放边界
- **Decision**: 秒级有效期，`submit_at <= expire_at` 视为可用；预占释放与订单超时关闭策略保持一致。
- **Rationale**: 边界规则明确、用户体验稳定、减少定时任务与订单状态冲突。
- **Alternatives considered**:
  - 不含边界：临界秒误伤概率更高。
  - 固定释放窗口：与订单关闭时间不一致会引发状态漂移。

## Decision 6: 幂等与一致性控制
- **Decision**: 关键动作（reserve/redeem/release/refund）引入幂等键与唯一约束，状态更新采用条件写入。
- **Rationale**: 防止重复回调、乱序事件造成重复扣减。
- **Alternatives considered**:
  - 仅依赖业务代码判断：并发场景下不可靠。
  - 全局分布式锁：实现复杂且成本高于必要水平。

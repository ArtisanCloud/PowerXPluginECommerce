# Phase 0 Research — 自营下单（Mini-app + Admin）MVP（结论汇总）

> 本文件用于将“影响实现与验收的关键决策”显式化，并给出替代方案对比。结论来源：`specs/007-order-mini-app/spec.md` 的 Clarifications 与 `docs/plan/marketing/order/checkout_mvp.md` 的 MVP 范围声明。

## Decision 1：路由前缀与鉴权边界

- **Decision**: 小程序订单接口放在 `/api/v1/mini-app/orders`（customer token 保护）；后台订单接口放在 `/api/v1/admin/orders`（JWT + RBAC）。
- **Rationale**: 与仓库现有 miniapp/admin 路由体系一致，且符合宪章的“管理端点与反代合同”约束。
- **Alternatives considered**:
  - 统一放在 `/v1/orders`：需要引入公共业务合同与现有 `/api/v1/**` 体系的兼容层，MVP 成本更高。

## Decision 2：创建订单幂等（防重试/重复点击）

- **Decision**: 创建订单必须支持幂等键（适用于小程序与后台）。同一幂等键在有效期内重复提交，返回同一订单结果且不得重复锁定库存。
- **Rationale**: 直接降低重复订单与库存锁定风险，减少运营与对账成本；也能让验收更稳定。
- **Alternatives considered**:
  - 不做幂等：实现最省，但重复订单/锁库存会成为高频线上问题。
  - 仅后台幂等：小程序仍有重试/弱网重复提交风险。

## Decision 3：多 SKU 下单的并发语义（整单原子）

- **Decision**: 多 SKU 下单采用整单原子：任一明细项不可售或库存不足则整单失败，不创建订单、不锁任何库存。
- **Rationale**: 语义清晰、易解释，且最容易保证“无超卖 + 无残余锁定”。
- **Alternatives considered**:
  - 允许部分成功：会引入“部分下单/拆单”的复杂语义与对账问题。

## Decision 4：价格与金额快照的时点

- **Decision**: 以“提交时服务端计算的当前价”为准直接创建订单，并保存价格/金额快照；MVP 不做“价格变化二次确认”门槛。
- **Rationale**: 规则简单、最可落地；价格快照保证后续查询与追溯一致。
- **Alternatives considered**:
  - 价格变化拒绝并要求用户重确认：交互与重试成本增加。
  - 价格变化进入二次确认态：引入额外状态与流程，超出 MVP。

## Decision 5：库存锁定策略（最小强一致）

- **Decision**: 创建订单在同一租户事务内，对每个 SKU 的 default 仓库存行加锁并更新 `locked_qty += qty`；取消订单执行 `locked_qty -= qty`（下限为 0）。
- **Rationale**: 强一致、实现成本低，能优先保证正确性；后续可升级为 reservation 表与异步补偿。
- **Alternatives considered**:
  - 引入 `inventory_reservations`：可追溯更好，但需要额外数据模型与补偿机制，非 MVP 必需。

## Decision 6：取消语义（权限与可取消状态）

- **Decision**: 仅后台（运营/客服）可取消订单，且仅 `pending_payment` 状态允许取消；`paid` 在 MVP 中不可取消；`draft` 可按草稿作废/删除处理但不视为取消。
- **Rationale**: 避免引入退款/回滚等复杂联动，保持 MVP 边界。
- **Alternatives considered**:
  - 客户可取消：需要小程序 UI/权限与更多风控策略。
  - `paid` 也可取消：会牵涉支付与退款语义，超出 MVP。

## Decision 7：人可读订单号

- **Decision**: 生成并对外展示人可读订单号（租户内唯一），用于后台检索与客服沟通。
- **Rationale**: 只用 UUID 不利于人工沟通与检索；租户内唯一通常已足够。
- **Alternatives considered**:
  - 仅内部 ID：对运营支持不友好。
  - 全局唯一：约束更强但业务价值有限，且实现复杂度更高。

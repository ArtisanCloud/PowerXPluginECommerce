# Phase 0 Research — 库存（Stock）MVP（结论汇总）

> 本文件用于将“影响实现与验收的关键决策”显式化，并给出替代方案对比。结论来源：`specs/006-sku-inventory-stock/spec.md` 的 Clarifications 与 `docs/plan/inventory/stock.md` 的库存状态语义。

## Decision 1：库存维度与默认仓

- **Decision**: MVP 只维护 `warehouse_id=default` 的库存（可扩展到多仓）。
- **Rationale**: 先让“可售库存→能上架”闭环跑通，避免被仓库主数据、库区库位等复杂度阻塞。
- **Alternatives considered**:
  - 直接做多仓：需要仓库主数据、权限、分配策略与更多 UI 入口，超出 MVP。

## Decision 2：可售库存口径

- **Decision**: “可售（salable）”口径固定为 `available_qty > 0`（仅评估 default 仓）。
- **Rationale**: 一期尚未实现订单锁定/安全库存，使用最小口径即可支撑上架把关；后续可升级为 `available-locked-safety`。
- **Alternatives considered**:
  - `available-locked`：需要锁定来源与一致性策略。
  - `available-locked-safety`：需要引入 safety_stock 维护与告警策略。

## Decision 3：库存写入语义（delta 调整）

- **Decision**: 写入定义为“增量调整（delta）”，且 `delta` 仅支持整数（按件）。
- **Rationale**: 审计更直观（记录变更量），也更贴近仓库操作；整数件数适配绝大多数 SKU。
- **Alternatives considered**:
  - 覆盖写入（set to value）：易产生“来源不明”的跳变，且并发下更容易覆盖别人调整结果。
  - 支持小数：涉及计量单位/舍入与更多校验，非 MVP 必需。

## Decision 4：并发与一致性策略（最小保证）

- **Decision**: 调整过程必须保证最终 `available_qty` 不为负；并发冲突应以“拒绝 + 可解释错误”或“串行化更新”确保结果可解释。
- **Rationale**: 防止出现不可履约的负库存与对账争议；MVP 先保证正确性与可追溯。
- **Alternatives considered**:
  - 允许负库存：会导致上架/履约/售后逻辑不可解释。

## Decision 5：把关触发点与失败策略

- **Decision**: 在 **SPU 发布** 与 **渠道上架/同步发布** 两处都进行库存校验；校验失败采取“仅阻止”策略（不自动下架/不自动改状态）。
- **Rationale**: 双触发点可防止绕过；仅阻止策略实现成本低且风险可控，后续再引入自动下架任务治理。
- **Alternatives considered**:
  - 仅在其中一处校验：存在绕过路径。
  - 失败自动下架：需要任务中心、状态回滚与更多沟通成本，放到下一阶段。


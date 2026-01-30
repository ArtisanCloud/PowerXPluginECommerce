# Feature Specification: 订阅会籍权益代币发放

**Feature Branch**: `010-membership-entitlements`  
**Created**: 2026-01-30  
**Status**: Draft  
**Input**: User description: "根据 docs/plan/customer/membership.md 与 docs/plan/customer/tokens.md 补充内容，编写订阅支付后会籍/权益/代币发放的规格说明文档"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 支付成功自动授予会籍与权益 (Priority: P1)

作为订阅商品的客户，在支付成功后系统应自动授予会籍与对应权益（包含组合型权益），并避免重复发放。

**Why this priority**: 支付完成即兑现权益是订阅体验的核心，且必须保证幂等避免重复发放。

**Independent Test**: 通过一次成功支付回调即可验证会籍与权益是否生成且不重复。

**Acceptance Scenarios**:

1. **Given** 订阅计划已绑定会员等级与权益，**When** 支付回调成功，**Then** 生成/更新会籍并发放权益。
2. **Given** 同一笔交易触发多次回调，**When** 系统处理回调，**Then** 只发放一次权益且不会重复记账。

---

### User Story 2 - 订阅计划绑定权益与代币 (Priority: P2)

作为运营人员，我需要在订阅计划上配置会籍、权益包以及代币发放规则，确保不同订阅档位映射不同服务组合。

**Why this priority**: 若无绑定规则，支付成功也无法知道该发放哪些权益与代币。

**Independent Test**: 仅通过订阅计划配置页面即可验证绑定字段是否可保存并被读取。

**Acceptance Scenarios**:

1. **Given** 订阅计划可编辑，**When** 保存会籍/权益/代币字段，**Then** 绑定关系可被后续发放逻辑读取。

---

### User Story 3 - 客户查看已获得权益与代币余额 (Priority: P3)

作为订阅客户，我希望在订单或会员视角中看到已获得权益与代币余额，以确认支付结果。

**Why this priority**: 有助于客户确认订阅已生效并可使用服务。

**Independent Test**: 通过查询客户视角的权益/代币数据即可完成验证。

**Acceptance Scenarios**:

1. **Given** 订阅已支付成功，**When** 客户访问权益/余额视图，**Then** 可看到对应权益与代币数量。

---

### Edge Cases

- 支付成功但订阅计划未配置权益或会籍时，系统应仅更新订单状态并记录异常原因。
- 权益有效期到达后继续访问服务时，系统应提示权益已过期且不再可用。
- 当权益数量不足（如 0）时，系统应阻止服务使用并提示余额不足。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须支持订阅计划绑定会籍等级、权益包与代币发放规则。
- **FR-002**: 系统必须在订阅支付成功后生成或更新会籍（membership assignment）。
- **FR-002a**: 会籍、权益与代币的生效时间以支付成功时间为准。
- **FR-003**: 系统必须按权益定义生成可消费权益，支持单体权益与组合型权益（bundle），默认叠加策略为 `stack`。
- **FR-004**: 系统必须支持权益的有效期与数量配置，并以 `quantity = -1` 表示无限量。
- **FR-005**: 系统必须支持订阅支付成功时发放代币余额。
- **FR-005a**: 代币发放频率为每计费周期一次（含首笔支付）。
- **FR-006**: 系统必须保证对同一交易回调的幂等处理，避免重复发放权益与代币。
- **FR-006a**: 幂等键优先使用 `transaction_id`，缺失则回退 `out_trade_no`，再回退 `order_id`。
- **FR-007**: 客户必须能够查询其当前会籍、权益与代币余额。
- **FR-008**: 系统必须记录权益与代币的发放来源（订阅计划/交易/订单）。

### Key Entities *(include if feature involves data)*

- **SubscriptionPlanBinding**: 订阅计划与会籍/权益/代币的绑定信息（metadata 字段）。
- **MembershipTier**: 会员等级定义（等级名称、规则、状态）。
- **MembershipAssignment**: 客户与会员等级的关联，含生效期与状态。
- **Benefit**: 权益定义，支持 `single` 与 `bundle` 类型。
- **Entitlement**: 可消费权益实例（服务项、数量、有效期、来源）。
- **TokenAccount**: 客户代币账户与余额。
- **TokenTransaction**: 代币发放/扣减记录。

## Assumptions & Dependencies

- 订阅计划已配置会籍/权益/代币绑定字段（metadata）。
- 支付成功回调可稳定送达，且可识别同一交易的唯一标识。
- 客户身份在订阅订单中可唯一定位。

## Clarifications

### Session 2026-01-30

- Q: 无限权益的数量如何表示？ → A: `quantity = -1` 表示无限。
- Q: 权益默认叠加策略是什么？ → A: 默认 `stack`（数量累加）。
- Q: 订阅支付后的生效时间如何定义？ → A: 支付成功立即生效。
- Q: 代币发放频率如何定义？ → A: 每个计费周期发放一次（含首笔）。
- Q: 幂等键应该如何选择？ → A: 优先 `transaction_id`，回退 `out_trade_no`，再回退 `order_id`。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 订阅支付成功后，权益与代币发放在 60 秒内可被客户查询到。
- **SC-002**: 重复回调不会导致重复发放，重复率为 0。
- **SC-003**: 至少 95% 的订阅订单在首次支付后能正确关联会籍与权益。
- **SC-004**: 客户查询权益/余额接口平均响应时间 ≤ 1 秒（50 条以内）。

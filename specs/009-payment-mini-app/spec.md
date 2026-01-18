# Feature Specification: 小程序支付

**Feature Branch**: `009-payment-mini-app`  
**Created**: 2026-01-15  
**Status**: Draft  
**Input**: User description: "docs/plan/marketing/payments/miniapp"

目标：完成小程序支付闭环（发起支付 → 回调确认 → 结果页展示/重试），并确保支付状态与订单状态一致可追溯；支付使用微信支付（PowerWechat 封装）。

## Clarifications

### Session 2026-01-15

- Q: 支付回调与前端确认的优先级？ → A: 回调优先，前端短时轮询兜底
- Q: 支付失败后的重试策略？ → A: 允许 1 次立即重试 + 订单列表继续支付
- Q: 支付风险拦截默认处理方式？ → A: 记录风险事件，提示联系支持

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 小程序支付与结果反馈 (Priority: P1)

作为小程序用户，我需要在下单后完成支付并获得明确结果反馈，失败时能快速重试或查看订单状态。

**Why this priority**: 直接影响支付转化率与用户体验。

**Independent Test**: 调用小程序支付创建接口，拉起微信支付完成支付，回调落库后结果页展示成功；失败场景可重试并回到订单列表。

**Acceptance Scenarios**:

1. **Given** 用户发起支付，**When** 支付成功，**Then** 结果页展示成功状态并可进入订单详情。
2. **Given** 用户支付取消或失败，**When** 返回结果页，**Then** 能看到失败原因并提供重试入口。
3. **Given** 回调到达延迟，**When** 结果页打开，**Then** 显示“支付处理中”并允许轮询确认。

### Edge Cases

- 支付渠道不可用或凭证失效时，提示渠道不可用并记录告警，禁止继续发起支付。
- 支付回执延迟导致结果页状态不一致时，提示“支付处理中”，引导用户刷新订单并触发轮询确认。
- 交易金额不一致或疑似欺诈时，阻断交易并记录风控事件，提示联系支持。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 小程序端必须支持支付发起、结果展示与失败重试，回调优先且前端短时轮询兜底。
- **FR-002**: 系统必须提供微信支付回调处理接口，支持幂等校验与状态确认。
- **FR-003**: 系统必须支持支付失败后 1 次立即重试，并提供订单列表入口继续支付。
- **FR-004**: 支付状态必须与订单状态一致可追溯（支付成功后更新订单为 `paid`）。
- **FR-005**: 所有支付相关数据必须包含 `tenant_uuid` 并遵循多租户 RLS 约束。

### Key Entities *(include if feature involves data)*

- **支付单/交易**: 交易编号、订单关联、金额、渠道、状态、时间、手续费。
- **支付回调**: 回调 payload、幂等键、签名验真结果与状态确认。
- **风险事件**: 异常类型、触发规则、处理动作、审计记录。

**Assumptions & Dependencies**: 使用 PowerWechat 作为微信支付 SDK；订单与支付数据一致；小程序具备调用支付 API 权限。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 支付拉起成功率 ≥ 98%。
- **SC-002**: 支付结果确认时间 p95 ≤ 1s（含回调与轮询兜底）。
- **SC-003**: 支付失败可在 3 分钟内完成重试并成功支付（在可支付条件下）。

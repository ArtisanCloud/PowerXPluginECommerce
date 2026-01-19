# Feature Specification: 小程序支付

**Feature Branch**: `009-payment-mini-app`  
**Created**: 2026-01-19  
**Status**: Draft  
**Input**: User description: "docs/plan/marketing/payments/miniapp.md"

目标：完成小程序支付闭环（下单确认 → 拉起支付 → 结果确认 → 结果页展示/重试），并确保支付状态与订单状态一致可追溯。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 小程序内完成支付并得到结果 (Priority: P1)

作为小程序用户，我需要在下单后完成支付并获得明确结果反馈，失败或取消时能快速重试或查看订单状态。

**Why this priority**: 直接影响支付转化率与用户体验。

**Independent Test**: 用户从下单确认页发起支付并完成支付，结果页展示成功；失败或取消时展示原因并提供一次重试与返回订单入口。

**Acceptance Scenarios**:

1. **Given** 用户发起支付，**When** 支付成功，**Then** 结果页展示成功状态并可进入订单详情。
2. **Given** 用户支付取消或失败，**When** 返回结果页，**Then** 能看到原因并提供一次重试入口。
3. **Given** 回调确认延迟，**When** 结果页打开，**Then** 展示“支付处理中”并允许短时重试确认。

---

### User Story 2 - 订单内继续支付与状态追踪 (Priority: P2)

作为小程序用户，我需要在订单列表或订单详情中看到支付状态，并在待支付时继续支付。

**Why this priority**: 支付失败或中断后能快速恢复，减少流失。

**Independent Test**: 在订单列表/详情中看到支付状态，点击“继续支付”可进入支付流程并完成结果确认。

**Acceptance Scenarios**:

1. **Given** 订单为待支付状态，**When** 用户点击继续支付，**Then** 可以重新发起支付并完成结果确认。
2. **Given** 订单支付成功，**When** 查看订单详情，**Then** 展示支付成功状态与时间。

### Edge Cases

- 支付渠道不可用时，提示“当前暂不可支付”并引导返回订单。
- 回调确认延迟导致结果页暂不一致时，提示“支付处理中”并允许短时确认。
- 金额不一致或疑似风险时，阻断交易并提示联系支持。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须支持小程序内发起支付并展示支付结果（成功/失败/取消/处理中）。
- **FR-002**: 系统必须以服务端确认结果为准，并提供短时结果确认兜底机制。
- **FR-003**: 系统必须允许支付失败或取消后 1 次立即重试，并在订单列表/详情提供继续支付入口。
- **FR-004**: 支付状态变更必须驱动订单状态同步更新且可追溯。
- **FR-005**: 重复结果通知不得导致支付或订单状态回退。

### Key Entities *(include if feature involves data)*

- **支付单/交易**: 交易编号、订单关联、金额、渠道、状态、时间。
- **支付结果**: 结果状态、结果来源、确认时间、失败原因。
- **订单状态**: 订单支付状态、支付时间、结果追溯信息。
- **风险事件**: 异常类型、处理动作、记录信息。

**Assumptions & Dependencies**: 小程序具备支付能力与调用权限；订单与支付状态需要保持一致；客服入口可用。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 支付发起成功率 ≥ 98%。
- **SC-002**: 95% 用户在 10 秒内获得最终支付结果（成功/失败/取消/处理中后的最终确认）。
- **SC-003**: 在可支付条件下，支付失败后的即时重试成功率 ≥ 30%。

# Tasks: 小程序支付

**Input**: Design documents from `/specs/009-payment-mini-app/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 对齐小程序支付 PRD 与规格说明（`docs/plan/marketing/payments/miniapp.md`、`specs/009-payment-mini-app/spec.md`）

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T002 统一支付状态机与回调幂等策略在 `specs/009-payment-mini-app/research.md`
- [x] T003 统一对外 API 合同与字段命名在 `specs/009-payment-mini-app/contracts/openapi.yaml`
- [x] T004 设计支付相关领域模型草案在 `specs/009-payment-mini-app/data-model.md`

---

## Phase 3: User Story 1 - 小程序支付与结果反馈 (Priority: P1)

**Goal**: 完成支付发起、结果确认、失败/取消重试与结果页展示。

**Independent Test**: 小程序可发起支付、显示成功/失败/取消/处理中结果，失败可立即重试并返回订单继续支付。

### Implementation for User Story 1

- [ ] T005 [US1] 新增支付单创建接口在 `backend/internal/transport/http/agent/payments/transactions_handler.go`
- [ ] T006 [US1] 新增支付回调处理在 `backend/internal/transport/http/agent/payments/providers_callback_handler.go`
- [ ] T007 [US1] 新增支付服务（含 PowerWechat 调用）在 `backend/internal/services/agent/payments/transactions_service.go`
- [ ] T008 [US1] 在小程序增加支付请求封装在 `mini-app/src/services/miniapp-payment.ts`
- [ ] T009 [US1] 在订单列表接入支付入口在 `mini-app/src/pages/order/list.vue`
- [ ] T010 [US1] 在订单详情接入支付入口在 `mini-app/src/pages/order/detail.vue`
- [ ] T011 [US1] 在结果页补充支付状态与重试入口在 `mini-app/src/pages/order/success.vue`
- [ ] T012 [US1] 增加支付中状态展示与短时确认逻辑在 `mini-app/src/pages/order/`

---

## Phase 4: Polish & Cross-Cutting Concerns

- [ ] T013 支付回调幂等与状态确认记录（`backend/internal/services/agent/payments/`）
- [ ] T014 运行 quickstart 检查项与自检记录在 `specs/009-payment-mini-app/quickstart.md`

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
- [x] T004-1 补充第三方账号绑定模型在 `specs/009-payment-mini-app/data-model.md`

---

## Phase 3: User Story 1 - 小程序支付与结果反馈 (Priority: P1)

**Goal**: 完成支付发起、结果确认、失败/取消重试与结果页展示。

**Independent Test**: 小程序可发起支付、显示成功/失败/取消/处理中结果，失败可立即重试并返回订单继续支付。

### Implementation for User Story 1

- [x] T005 [US1] 新增支付单创建接口在 `backend/internal/transport/http/miniapp/payments/transactions_handler.go`
- [x] T006 [US1] 新增支付回调处理在 `backend/internal/transport/http/miniapp/payments/providers_callback_handler.go`（支持 providerType/mchId/appId 路由）
- [x] T007 [US1] 新增支付服务（含 PowerWechat 调用）在 `backend/internal/services/agent/payments/transactions_service.go`
- [x] T008 [US1] 在小程序增加支付请求封装在 `mini-app/src/services/miniapp-payment.ts`
- [x] T009 [US1] 在订单列表接入支付入口在 `mini-app/src/pages/order/list.vue`
- [x] T010 [US1] 在订单详情接入支付入口在 `mini-app/src/pages/order/detail.vue`
- [x] T011 [US1] 在结果页补充支付状态与重试入口在 `mini-app/src/pages/order/success.vue`
- [x] T012 [US1] 增加支付中状态展示与短时确认逻辑在 `mini-app/src/pages/order/`

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T013 支付回调幂等与状态确认记录（`backend/internal/services/agent/payments/`）
- [x] T014 运行 quickstart 检查项与自检记录在 `specs/009-payment-mini-app/quickstart.md`

---

## Phase 4.1: MiniApp Auth (Wechat)

- [x] T014-1 [US1] 后端：小程序授权登录（code 换 openid/unionid）接口实现（`backend/internal/transport/http/miniapp/auth/`）
- [x] T014-2 [US1] 后端：第三方身份绑定表（模型/迁移/索引）（`backend/internal/entity/models/customer/`）
- [x] T014-3 [US1] 后端：绑定逻辑接入登录流程（`backend/internal/services/customer/auth/`）
- [x] T014-4 [US1] 文档：接口契约补齐（`specs/009-payment-mini-app/contracts/openapi.yaml`）
- [x] T014-5 [US1] 前端：UniApp 登录对接（`mini-app/src/`）

---

## Phase 5: Payment Provider Configuration (Admin)

- [x] T015 [US3] 支付渠道凭证加密存储/解密读取（`backend/internal/services/admin/payments/`）
- [x] T016 [US3] 支付渠道配置管理接口（`backend/internal/transport/http/admin/payments/`）
- [x] T017 [US3] 支付渠道管理页面（仅微信可配置，其它展示禁用）（`web-admin/`）

# Tasks: 订阅会籍权益代币发放

## Phase 1: Setup

- [x] T001 确认 specs 与计划文档路径正确（`specs/010-membership-entitlements/spec.md`, `specs/010-membership-entitlements/plan.md`）

## Phase 2: Foundational

- [x] T002 建立权益/代币最小实体清单与字段映射（参考 `specs/010-membership-entitlements/data-model.md`）
- [x] T003 定义幂等键优先级与统一写入位置（`specs/010-membership-entitlements/spec.md` FR-006a）

## Phase 3: User Story 1 (P1) — 支付成功自动授予会籍与权益

**Goal**: 订阅支付成功后生成会籍与权益，支持 bundle 与幂等。

**Independent Test**: 单笔支付成功回调触发后，会籍/权益生成且不会重复。

- [ ] T004 [US1] 新增权益实例与代币账本模型（`backend/internal/entity/models`）
- [ ] T005 [US1] 新增权益实例与代币账本仓储（`backend/internal/entity/repository`）
- [ ] T006 [US1] 在支付回调服务中接入会籍/权益/代币发放编排（`backend/internal/services/agent/payments` 或对应支付服务）
- [ ] T007 [US1] 幂等处理：按 `transaction_id` → `out_trade_no` → `order_id` 去重发放（支付回调服务）
- [ ] T008 [US1] 处理缺失绑定字段的异常并记录（支付回调服务日志/审计）

## Phase 4: User Story 2 (P2) — 订阅计划绑定权益与代币

**Goal**: 订阅计划可配置会籍/权益/代币绑定字段。

**Independent Test**: 在订阅计划配置页保存绑定字段并可读取。

- [ ] T009 [US2] 订阅计划 metadata 支持 membershipTierId/benefitIds/tokenCode/tokenAmount/tokenExpireDays/tokenRollover（后端 DTO/service）
- [ ] T010 [US2] web-admin 订阅计划配置页补充绑定字段 UI（`web-admin/app/components/product/SubscriptionPlanPanel.vue`）
- [ ] T011 [US2] 确保订阅计划列表/详情可返回绑定字段（mini-app 或 admin API）

## Phase 5: User Story 3 (P3) — 客户查看权益与代币余额

**Goal**: 客户可查询当前权益与代币余额。

**Independent Test**: 客户接口可返回权益与代币余额。

- [ ] T012 [US3] 新增客户权益查询接口（`backend/internal/transport/http/miniapp` 或 `admin`）
- [ ] T013 [US3] 新增代币余额查询接口（`backend/internal/transport/http/miniapp` 或 `admin`）
- [ ] T014 [US3] mini-app 订单/会员视图展示权益与代币余额（`mini-app/src/pages/order` 或 `profile`）

## Phase 6: Polish & Cross-Cutting

- [ ] T015 补充 quickstart 验证步骤与预期结果（`specs/010-membership-entitlements/quickstart.md`）
- [ ] T016 记录文档变更索引（`docs/plan/customer/readme.md`）

## Dependencies

- US1 依赖 Foundational 完成
- US2/US3 可并行推进，但 US1 的模型/仓储优先

## Parallel Execution Examples

- [P] T004 / T005 可并行（模型/仓储）
- [P] T009 / T010 可并行（后端绑定字段 / 前端 UI）
- [P] T012 / T013 可并行（权益/代币查询接口）

## Implementation Strategy

- MVP 先完成 US1（支付成功发放会籍/权益/代币）
- 再完成 US2（配置绑定）与 US3（查询/展示）

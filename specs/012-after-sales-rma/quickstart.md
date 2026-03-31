# Quickstart: 售后 RMA 与退货门户

## 0. Phase 1 初始化结果
- 已建立售后域目录骨架：`backend/internal/{entity/models,entity/repository,services,transport/http}/{admin,miniapp}/after_sales/`
- 已建立前端骨架入口：`web-admin/app/pages/customer/returns-portal.vue`
- 已建立售后 API composable：`web-admin/app/composables/api/useAfterSales.ts`

## 1. 前置条件
- 当前分支：`012-after-sales-rma`
- 本地可启动 backend 与 web-admin
- 已完成数据库迁移：`make migrate`
- 可访问 API 前缀：`/api/v1/admin/**` 与 `/api/v1/mini-app/**`

## 2. 启动服务
1. 启动后端：`make run`（或 `make dev`）
2. 启动管理端：`cd web-admin && npm run dev`

## 3. 客户侧流程验证（US1）
1. 准备一笔已支付订单且包含可售后明细。
2. 调用 `POST /api/v1/mini-app/after-sales` 提交售后申请（仅退款/退货退款/换货）。
3. 调用 `GET /api/v1/mini-app/after-sales` 查询列表，确认新申请可见。
4. 调用 `GET /api/v1/mini-app/after-sales/{id}` 查询详情，确认状态与时间线正确。

预期结果：
- 申请提交成功并生成唯一申请编号。
- 列表与详情可查询，状态为初始受理态。

## 4. 管理侧流程验证（US2）
1. 调用 `GET /api/v1/admin/after-sales/cases` 查询待处理售后单。
2. 调用 `POST /api/v1/admin/after-sales/cases/{id}/accept` 受理。
3. 调用 `POST /api/v1/admin/after-sales/cases/{id}/approve` 或 `/reject` 完成审核。
4. 对通过单调用 `POST /api/v1/admin/after-sales/cases/{id}/complete` 完结。

预期结果：
- 状态流转符合状态机约束。
- 每次操作可在时间线中看到操作者与动作记录。

## 5. 最小联动验证（US3）
1. 对退款类售后审核通过后，确认订单侧出现售后处理中标记且防止重复退款申请。
2. 对退货类售后调用 `POST /api/v1/admin/after-sales/cases/{id}/reverse-logistics` 绑定逆向物流。
3. 在售后详情中确认逆向物流关联结果可见。

预期结果：
- 订单/售后展示状态一致。
- 同一明细进行中售后不允许重复提交。

## 6. 质量门禁
1. 后端：`make lint && make test`
2. 前端：`make test-admin-ci`
3. 可选构建验证：`make build-admin`

## 7. Phase 6 执行记录（2026-03-31）

### 7.1 后端回归（T042）
- 执行命令：`make test`
- 结果：通过（exit code 0）
- 说明：覆盖 `backend` 全量 `go test ./...`，包含售后模块与新增联动逻辑编译/测试通过。

### 7.2 前端回归（T043）
- 执行命令：`make test-admin-ci`
- 结果：通过（exit code 0）
- 明细：
  - unit：4 files / 11 tests passed
  - component：4 files / 9 tests passed

### 7.3 端到端冒烟（T044，US1 → US2 → US3）
- 执行命令：
  - `go test ./internal/services/admin/after_sales -run TestAfterSalesUS123Smoke -v -count=1`
- 执行顺序：
  1. US1：客户创建售后申请（miniapp `Create`）
  2. US2：运营受理→审核中→审核通过（admin `Transition` + `Approve`）
  3. US3：逆向物流关联 + 订单事件联动校验（`Link` + `after_sales.approved` 事件）
- 结果：通过（exit code 0）

### 7.4 SC 指标样本（T045）
- 样本来源：`TestAfterSalesUS123Smoke` 的单次执行日志
- 指标记录：
  - US1 处理时长样本：`355.584µs`
  - US2 处理时长样本：`676.25µs`
  - US3 处理时长样本：`433.208µs`
  - 非法流转拦截：状态机冲突返回 `AFTER_SALES_INVALID_STATE_TRANSITION`（服务映射已接入）
  - 5 秒可见性：订单侧联动事件 `after_sales.*` 与售后列表回写在同事务链路完成，样本远低于 5 秒

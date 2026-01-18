# 小程序支付 PRD

> 覆盖端：小程序前端支付流程（下单 -> 拉起支付 -> 回调/轮询 -> 结果页）以及与后台支付单/订单的联动。面向用户支付体验、稳定性与风控提示。

## 1. 背景与目标
- 当前支付 PRD 聚焦后台管理与对账，需要补充小程序侧的支付流程、异常处理与体验设计。

**目标**
1. 支付流程清晰：创建支付单、拉起支付、回调/轮询、结果页一致。
2. 异常可控：失败/超时/取消有明确引导与重试策略。
3. 与订单/支付单闭环联动：支付状态与订单状态一致可追踪。

## 2. 角色
| 角色 | 诉求 |
| --- | --- |
| 终端用户 | 支付快速、失败可重试、结果明确 |
| 客服 | 可定位支付问题，协助用户恢复 |
| 运营 | 支付转化率、失败原因可观测 |

## 3. 端内页面与信息架构
1. **下单确认页**（`mini-app/src/pages/order/confirm.vue`）
   - 展示订单摘要、优惠、应付金额；当前未展示支付方式。
   - CTA：提交订单（创建订单后进入结果页）。
2. **支付中状态页/弹窗**
   - 目标：拉起支付后展示状态，提供“重新拉起支付/返回订单”。
   - 现状：暂无独立支付中页。
3. **支付结果页**（`mini-app/src/pages/order/success.vue`）
   - 现状：展示“订单已创建/待支付”，引导继续逛逛或查看订单。
   - 目标：成功/失败/取消分态展示，提供重试与客服入口。
4. **订单列表/详情页**（`mini-app/src/pages/order/list.vue`、`mini-app/src/pages/order/detail.vue`）
   - 展示支付状态、支付时间、支付方式；支持继续支付。
   - 现状：列表“去支付”按钮提示“支付待接入”。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 创建支付单 | 下单后生成支付单并返回支付参数 |
| 拉起支付 | 调用小程序支付 SDK（如微信支付） |
| 结果确认 | 回调优先，超时轮询兜底 |
| 失败重试 | 失败/取消后可重新发起支付 |
| 结果展示 | 成功/失败页，统一状态解释 |
| 订单联动 | 支付状态同步更新订单状态 |
| 风控提示 | 风控拦截或异常提示给用户 |

## 5. 流程
**现状流程**
1. **下单**：小程序提交订单 → 后端创建订单 → 跳转 `order/success`（状态为待支付）。
2. **支付入口**：订单列表/详情提供“去支付”入口，但尚未接入支付。

**目标流程**
1. **下单**：小程序提交订单 → 后端创建订单与支付单。
2. **拉起支付**：返回支付参数 → 调起支付 SDK。
3. **结果确认**：
   - 成功：SDK 回调 → 服务端确认 → 前端展示成功。
   - 失败/取消：SDK 结果 → 可重试；若无结果则轮询。
4. **兜底轮询**：前端轮询支付单状态（短周期 + 失败降级）。
5. **订单联动**：支付成功 → 订单状态更新 → 进入订单详情。

## 6. 接口与数据
**已对齐小程序现有接口**（`mini-app/src/services/miniapp-order.ts`）
- `POST /orders`（创建订单；当前通过 `Idempotency-Key` 做幂等）
- `GET /orders`（订单列表）
- `GET /orders/{id}`（订单详情）

**待接入支付接口**
- `POST /v1/agent/payments/transactions`（创建支付单）
- `GET /v1/agent/payments/transactions/{id}`（查询支付状态）
- `POST /v1/agent/payments/providers/{id}/callback`（渠道回调）

数据：支付单状态（待支付/支付中/成功/失败/退款）、订单状态、失败原因码。

### 6.1 支付与订单状态一致性策略
- **权威源**：支付单状态以服务端回调确认为准，订单状态由支付状态驱动更新。
- **更新时序**：回调处理成功后，同事务或幂等链路内更新支付单状态与订单状态（`paid`）。
- **幂等处理**：回调以支付单号/商户订单号为幂等键，重复回调不得回滚状态。
- **前端展示**：支付结果页/订单详情以服务端状态为准，短时轮询只用于缩短展示延迟。

**微信支付接入约定（PowerWechat 封装，对齐库内接口）**
- 服务端支付下单（小程序 JSAPI）：
  - PowerWechat 调用：`payment/order.Client.JSAPITransaction`
    - 请求结构：`request.RequestJSAPIPrepay`
      - `description`（商品描述）
      - `out_trade_no`（商户订单号）
      - `amount.total` / `amount.currency`
      - `payer.openid`（小程序用户 openid）
      - `notify_url`（不传则由 `config.notify_url` 注入）
      - 可选：`attach`、`detail.goods_detail`、`scene_info.payer_client_ip`
    - 返回结构：`response.ResponseUnitfy`（`prepay_id`）
  - 建议对外 API（业务层）：`POST /v1/agent/payments/transactions`
    - 入参建议：`orderId`、`orderNo`、`amountMinor`、`currency`、`channel`、`payMethod=wechat_jsapi`、`openid`、`client=miniapp`、`idempotencyKey`
- 服务端生成 JSAPI 拉起参数：
  - PowerWechat 调用：`payment/jssdk.Client.BridgeConfig(prepayID, false)`
  - 返回：`appId`、`timeStamp`、`nonceStr`、`package`（`prepay_id=...`）、`signType`、`paySign`
- 前端拉起支付（小程序）：
  - `uni.requestPayment({ provider: "wxpay", ...BridgeConfig })`
  - 成功回调仅代表客户端完成，最终以服务端回调为准
- 服务端回调（微信订单支付回调接口）：
  - 建议路径：`POST /v1/agent/payments/providers/{id}/callback`（provider=wechat）
  - PowerWechat 处理：`payment/notify.NewPaidNotify(app, r).Handle(...)`
    - `request.RequestNotify`：解密后可得到 `models.Transaction`
    - 核心字段：`out_trade_no`、`transaction_id`、`trade_state`、`success_time`、`amount.total`、`payer.openid`
  - 回调处理要点：验签/解密 → 幂等更新支付单 → 更新订单状态 → 返回 `{"code":"SUCCESS"}` 或 `{"code":"FAIL"}`（PowerWechat 已内置）

**支付状态映射（建议）**
- `pending_payment`：支付单已创建，待拉起支付/待支付
- `paying`：已拉起支付，等待回调确认（对应微信 `USERPAYING`）
- `paid`：支付成功（对应微信 `SUCCESS`），订单进入待发货/履约
- `failed`：支付失败
- `canceled`：用户取消
- `timeout`：支付超时
- `refunded`：已退款（全额/部分）

**微信支付回调报文（解密后）字段清单**
- `out_trade_no`：商户订单号
- `transaction_id`：微信支付订单号
- `trade_state` / `trade_state_desc`：交易状态/描述（如 `SUCCESS`、`NOTPAY`、`CLOSED`、`USERPAYING`、`PAYERROR`）
- `trade_type`：交易类型（JSAPI）
- `success_time`：支付成功时间（rfc3339）
- `amount.total`：订单总金额（分）
- `amount.currency`：币种（如 `CNY`）
- `payer.openid`：用户 openid
- `attach`：下单透传字段（可用于业务回填）

**解密后示例（节选）**
```json
{
  "appid": "wx1234567890",
  "mchid": "1900000109",
  "out_trade_no": "ORDER202501010001",
  "transaction_id": "4200001234202501012345678901",
  "trade_type": "JSAPI",
  "trade_state": "SUCCESS",
  "trade_state_desc": "支付成功",
  "success_time": "2025-01-01T10:29:42+08:00",
  "payer": {
    "openid": "oAuaP0TRUMwP169nQfg7XCEAw3HQ"
  },
  "amount": {
    "total": 129900,
    "currency": "CNY",
    "payer_total": 129900,
    "payer_currency": "CNY"
  },
  "attach": "orderId=...&channel=miniapp"
}
```

**失败原因码字典（建议）**
- `USER_CANCEL`：用户取消
- `PAY_TIMEOUT`：支付超时
- `SIGN_VERIFY_FAILED`：回调签名失败
- `ORDER_NOT_FOUND`：订单不存在
- `AMOUNT_MISMATCH`：金额不一致
- `CHANNEL_ERROR`：渠道异常/系统错误
- `RISK_BLOCKED`：风控拦截

## 7. 异常与风控
- 支付失败：显示原因 + 可重试入口（可限制次数）。
- 支付超时：提示稍后查询订单，提供刷新/轮询。
- 风控拦截：提示需联系客服或更换支付方式。

## 8. KPI
| 指标 | 目标 |
| --- | --- |
| 支付成功率 | ≥ 98% |
| 支付转化率 | ≥ 90% |
| 支付失败重试成功率 | ≥ 30% |

## 9. Backlog
- 多支付方式（余额/分期/礼品卡）。
- 失败原因聚类与智能引导。
- 支付风险提示与验证增强。

## 10. 交付任务拆解（前端/后端）
**前端（小程序）**
1. 支付按钮联动：订单列表/详情的“去支付”触发支付流程。
2. 支付拉起：调用 `POST /v1/agent/payments/transactions` 获取微信参数，执行 `uni.requestPayment`。
3. 支付结果处理：成功回调后跳转结果页，失败/取消显示原因并可重试。
4. 轮询兜底：支付发起后短周期轮询 `GET /v1/agent/payments/transactions/{id}` 直至确认。
5. UI 对齐：补充支付中态与失败态展示。

**后端（支付服务 + PowerWechat）**
1. 支付下单：封装 PowerWechat 下单（JSAPI）并返回 `paySign` 等参数。
2. 回调处理：校验签名、幂等更新支付单状态、写审计日志。
3. 状态查询：`GET /v1/agent/payments/transactions/{id}` 返回最新状态。
4. 订单联动：支付成功后更新订单状态并触发履约流程。
5. 风控与限流：对异常频繁发起/失败场景做策略控制。

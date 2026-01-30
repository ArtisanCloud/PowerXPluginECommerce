# 小程序支付 PRD

> 覆盖端：小程序前端支付流程（下单 → 拉起支付 → 回调/轮询 → 结果页）以及与后台支付单/订单的联动。面向用户支付体验、稳定性与风控提示。

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

## 3. 范围
**包含**
- 小程序支付闭环：下单确认 → 拉起支付 → 结果确认 → 结果页展示。
- 支付失败/取消/超时的提示与重试入口。
- 支付单与订单状态联动与可追溯。
- 订单列表/详情中的支付状态与继续支付入口。

**不包含**
- 多支付方式（余额/分期/礼品卡）。
- 退款与对账流程。
- 后台管理端的支付配置与报表能力。

## 4. 目标流程
**主流程**
1. 用户提交订单，系统创建订单与支付单。
2. 拉起支付（小程序内支付能力）。
3. 支付完成后优先使用服务端回调确认结果。
4. 前端展示结果页（成功/失败/取消/处理中）。
5. 支付成功后，订单状态同步更新并进入订单详情。

**异常/兜底流程**
- 回调延迟或无结果时，前端短周期轮询支付单状态作为兜底。
- 失败/取消支持 1 次立即重试，并在订单列表/详情提供“继续支付”。
- 风控拦截提示用户联系客服或更换支付方式。

## 5. 端内页面与体验
1. **下单确认页**（`mini-app/src/pages/order/confirm.vue`）
   - 展示订单摘要、优惠、应付金额。
   - CTA：提交订单并发起支付。
2. **支付中状态页/弹窗**
   - 拉起支付后展示状态，提供“重新拉起支付/返回订单”。
3. **支付结果页**（`mini-app/src/pages/order/success.vue`）
   - 成功/失败/取消/处理中分态展示。
   - 提供重试入口与客服入口。
4. **订单列表/详情页**（`mini-app/src/pages/order/list.vue`、`mini-app/src/pages/order/detail.vue`）
   - 展示支付状态、支付时间、支付方式；支持继续支付。

## 6. 功能与业务规则
| 功能 | 规则描述 |
| --- | --- |
| 创建支付单 | 下单后生成支付单并返回支付参数 |
| 拉起支付 | 使用小程序支付能力发起支付 |
| 结果确认 | 回调优先；前端短时轮询兜底（避免长时间阻塞） |
| 失败重试 | 允许 1 次立即重试；订单列表/详情继续支付 |
| 结果展示 | 成功/失败/取消/处理中统一状态解释 |
| 订单联动 | 支付状态同步更新订单状态并可追溯 |
| 风控提示 | 风控拦截或异常提示给用户 |

**接口约定（小程序）**
- 发起支付必须显式携带 `providerType/mchId/appId` 以定位渠道配置。
- 回调地址建议使用：`/api/v1/mini-app/payments/providers/{type}/{mchId}/{appId}/callback`。

## 7. 小程序登录与 OpenID 获取
**目标**：使用 PowerWeChat 完成小程序授权登录与 OpenID 绑定，为支付提供用户标识。

**核心流程**
1. 小程序端调用 `wx.login()` 获取 `code`。
2. 前端将 `code` 传给后端登录接口。
3. 后端使用 PowerWeChat `MiniProgramApp.Auth.Session(ctx, code)` 换取 `openid/unionid` 并签发 token。

泳道图（登录与资料获取）：
[MiniApp Wechat Login Flow](https://www.figma.com/online-whiteboard/create-diagram/331d6c07-7ca9-44b8-b542-ab2763148855?utm_source=other&utm_content=edit_in_figjam&oai_id=&request_id=26b651e4-df17-44cd-aa28-959711b8fca7)

**后端接口**
- `POST /api/v1/mini-app/auth/wechat/login`
  - 入参：`providerId`（必填）、`code`（必填），可选 `nickname/avatarUrl`
  - 出参：`token/expiresAt/customerId/tenantUuid/customerName/openid/unionid`
- `POST /api/v1/mini-app/auth/wechat/phone`
  - 入参：`code`（必填）
- `POST /api/v1/mini-app/auth/wechat/decrypt`
  - 入参：`encryptedData/sessionKey/iv`
- `POST /api/v1/mini-app/auth/wechat/check-encrypted`
  - 入参：`hash` 或 `encryptedData`
- `POST /api/v1/mini-app/auth/wechat/paid-unionid`
  - 入参：`openid`（必填）+ `transactionId/mchId/outTradeNo` 任选

**配置项**
```yaml
wechat_miniapp:
  app_id: "wx..."
  app_secret: "..."
```
或环境变量：
`POWERX_WECHAT_MINIAPP_APP_ID` / `POWERX_WECHAT_MINIAPP_APP_SECRET`

## 8. 状态与一致性
- **权威源**：支付单状态以服务端回调确认为准。
- **更新时序**：回调确认后更新支付单状态，并驱动订单状态更新。
- **幂等处理**：同一支付单重复回调不得回滚状态。
- **前端展示**：短时轮询仅用于缩短展示延迟，最终状态以服务端为准。

**支付状态（建议）**
- `pending_payment`：待支付/未拉起
- `paying`：已拉起支付，等待确认
- `paid`：支付成功
- `failed`：支付失败
- `canceled`：用户取消
- `timeout`：支付超时
- `refunded`：已退款

## 9. KPI
| 指标 | 目标 |
| --- | --- |
| 支付成功率 | ≥ 98% |
| 支付转化率 | ≥ 90% |
| 支付失败重试成功率 | ≥ 30% |

## 10. Backlog
- 多支付方式（余额/分期/礼品卡）。
- 失败原因聚类与智能引导。
- 支付风险提示与验证增强。

## 11. 交付任务拆解（前端/后端）
**前端（小程序）**
1. 支付入口联动：订单列表/详情触发支付流程。
2. 支付拉起：创建支付单并拉起支付。
3. 支付结果处理：成功/失败/取消/处理中展示与重试入口。
4. 轮询兜底：短周期轮询支付单状态直至确认。
5. UI 对齐：补充支付中态与失败态展示。

**后端（支付服务）**
1. 支付下单：生成支付单并返回支付参数。
2. 回调处理：回调验真与幂等更新支付单状态。
3. 状态查询：返回最新支付状态。
4. 订单联动：支付成功后更新订单状态并记录审计。
5. 风控与限流：对异常频繁发起/失败场景做策略控制。

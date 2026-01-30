# Quickstart: 小程序支付

## Checks

- [ ] T001 已配置微信支付凭证（PowerWechat）
- [ ] T001-1 已配置微信小程序 AppID/AppSecret（用于 code 换取 openid）
- [ ] T002 支付创建接口可返回唤起参数
- [ ] T003 回调验签成功并更新支付单状态
- [ ] T004 支付状态查询接口返回最新状态
- [ ] T005 结果页展示支付状态并支持重试/处理中轮询

## Runbook

1) 小程序登录：`POST /api/v1/mini-app/auth/wechat/login`（需传 providerId）
2) 发起支付：`POST /api/v1/mini-app/payments/transactions`
3) 回调（推荐）：`POST /api/v1/mini-app/payments/providers/{type}/{mchId}/{appId}/callback`
4) 回调（兼容）：`POST /api/v1/mini-app/payments/providers/id/{id}/callback`
5) 查询状态：`GET /api/v1/mini-app/payments/transactions/{id}`

发起支付请求需显式指定支付渠道定位信息：
- `providerType`（如 `wechat`）
- `mchId`
- `appId`
6) 前端轮询确认：订单/支付状态一致

小程序登录配置：
```yaml
wechat_miniapp:
  app_id: "wx..."
  app_secret: "..."
```
或环境变量：
`POWERX_WECHAT_MINIAPP_APP_ID` / `POWERX_WECHAT_MINIAPP_APP_SECRET`

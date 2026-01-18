# Quickstart: 小程序支付

## Checks

- [ ] T001 已配置微信支付凭证（PowerWechat）
- [ ] T002 支付创建接口可返回唤起参数
- [ ] T003 回调验签成功并更新支付单状态
- [ ] T004 结果页展示支付状态并支持重试

## Runbook

1) 发起支付：`POST /v1/agent/payments/transactions`
2) 回调：`POST /v1/agent/payments/providers/{id}/callback`
3) 前端轮询确认：订单/支付状态一致

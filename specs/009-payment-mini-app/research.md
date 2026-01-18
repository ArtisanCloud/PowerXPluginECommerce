# Research: 小程序支付

## Decisions

- 支付回调优先，前端短时轮询兜底
- 回调幂等：同一订单/支付单重复回调必须可重放
- 失败重试：允许 1 次立即重试

## Notes

- 微信支付使用 PowerWechat 封装
- 回调需验签并记录 request_id/tenant_uuid

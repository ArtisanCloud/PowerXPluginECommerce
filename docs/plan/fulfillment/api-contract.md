# 履约与物流 API Contract（示例版）

## 1. 承运商
### GET /api/v1/admin/logistics/carriers
Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "c1",
      "name": "顺丰",
      "type": "sf",
      "status": "active"
    }
  ]
}
```

### POST /api/v1/admin/logistics/carriers
Request:
```json
{
  "name": "顺丰",
  "type": "sf",
  "status": "active",
  "credentials": {"app_id": "xxx", "key": "yyy"}
}
```

## 2. 运费模板
### POST /api/v1/admin/logistics/templates
Request:
```json
{
  "name": "华东默认模板",
  "currency": "CNY",
  "zones": [
    {"region": "华东", "first_weight": 1, "first_fee": 8, "add_weight": 1, "add_fee": 2}
  ]
}
```

## 3. 运单
### POST /api/v1/admin/logistics/waybills
Request:
```json
{
  "order_id": "o1",
  "carrier_id": "c1",
  "service_code": "SF_EXPRESS"
}
```
Response:
```json
{
  "success": true,
  "data": {
    "id": "w1",
    "waybill_no": "SF123456",
    "status": "created",
    "label_url": "https://..."
  }
}
```

## 4. 轨迹
### POST /api/v1/admin/logistics/waybills/{id}/track
Response:
```json
{
  "success": true,
  "data": [
    {"status": "in_transit", "description": "已发出", "occurred_at": "2026-02-06T10:00:00Z"}
  ]
}
```


## 5. 错误码规范（示例）
- `LOGISTICS_CARRIER_NOT_FOUND`
- `LOGISTICS_TEMPLATE_NOT_FOUND`
- `LOGISTICS_WAYBILL_CREATE_FAILED`
- `LOGISTICS_TRACKING_UNAVAILABLE`
- `FULFILLMENT_TASK_INVALID_STATUS`
- `REVERSE_WAYBILL_NOT_FOUND`

## 6. 响应包规范（示例）
```json
{
  "success": false,
  "error": {
    "code": "LOGISTICS_WAYBILL_CREATE_FAILED",
    "message": "承运商创建运单失败",
    "details": {}
  }
}
```


## 7. 权限矩阵（简版）
- `logistics.carriers.read/manage`
- `logistics.templates.read/manage`
- `logistics.waybills.read/manage`
- `fulfillment.tasks.read/manage`
- `reverse.logistics.read/manage`


## 8. 请求参数校验规则（示例）
- carrier
  - `name` 必填，长度 2-64
  - `type` 必填，enum
- template
  - `currency` 必填，ISO 4217
  - `zones` 至少 1 条
- waybill
  - `order_id` 必填
  - `carrier_id` 必填
  - `service_code` 必填

## 9. Webhook 回调契约（轨迹）
### POST /api/v1/admin/logistics/webhook
Request:
```json
{
  "provider": "sf",
  "event_id": "evt_123",
  "waybill_no": "SF123456",
  "status": "in_transit",
  "description": "已发出",
  "occurred_at": "2026-02-06T10:00:00Z"
}
```
Response:
```json
{
  "success": true
}
```


## 10. 幂等与签名校验（建议）
- 幂等键：`event_id` + `waybill_no`
- 重复回调：仅更新最后状态，不重复写入运单/轨迹
- 签名校验：
  - Header: `X-Provider-Signature`
  - Body: 原始 JSON 字符串
  - 算法：HMAC-SHA256

## 11. 失败重试策略
- 网络错误：重试 3 次（指数退避 1s/3s/9s）
- 业务错误：记录异常并报警，不自动重试


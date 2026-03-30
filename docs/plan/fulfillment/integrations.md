# 第三方承运商适配（集成层）

> 目标：在不改业务流程的前提下，通过适配器接入顺丰/京东物流/菜鸟/中通等。

## 1. 适配器接口（必须实现）
- `CreateShipment(order, address, items) -> { waybill_no, label_url }`
- `GetTracking(waybill_no) -> events[]`
- `CancelShipment(waybill_no)`
- `GetLabel(waybill_no) -> label_url`

## 2. 认证与配置字段
- `app_id` / `mch_id` / `api_key` / `api_secret`
- `sandbox`（是否沙箱）
- `callback_url`（轨迹回调/状态回调）
- `service_code_map`（服务编码映射表）

## 3. 回调与幂等
- 回调必须带 `provider_event_id`（幂等键）
- 重复回调仅更新最新状态，不生成新运单

## 4. 错误与重试
- 网络错误：重试 3 次（指数退避）
- 业务错误：记录异常并通知客服

## 5. 适配器列表（规划）
- 顺丰（SF）
- 京东物流（JD Logistics）
- 菜鸟（Cainiao）
- DHL / FedEx（国际件）


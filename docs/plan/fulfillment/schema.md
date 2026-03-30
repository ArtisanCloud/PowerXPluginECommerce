# 履约与物流数据结构（字段级）

> 说明：字段以最小可用为主，后续可扩展。所有表默认包含 `tenant_uuid`、`created_at`、`updated_at`。

## 1. 承运商
### 1.1 logistics_carriers
- `id` (uuid)
- `name` (string)
- `type` (string) 例如：`self`/`sf`/`jd`/`cainiao`
- `status` (string) `active`/`disabled`
- `contact_name` (string)
- `contact_phone` (string)
- `fee_rate` (numeric)
- `sla` (jsonb) 例如：`{"avg_days":2,"on_time":0.95}`
- `credentials` (jsonb) 认证信息

### 1.2 logistics_carrier_services
- `id` (uuid)
- `carrier_id` (uuid)
- `code` (string) 服务编码
- `name` (string)
- `service_type` (string) `express`/`heavy`/`cold`
- `status` (string) `active`/`disabled`
- `config` (jsonb)

## 2. 运费模板
### 2.1 logistics_rate_templates
- `id` (uuid)
- `name` (string)
- `currency` (string)
- `status` (string) `draft`/`published`
- `channels` (jsonb) 适用渠道列表
- `version` (int)

### 2.2 logistics_rate_zones
- `id` (uuid)
- `template_id` (uuid)
- `region` (string)
- `first_weight` (numeric)
- `first_fee` (numeric)
- `add_weight` (numeric)
- `add_fee` (numeric)
- `free_threshold` (numeric)

## 3. 运单与轨迹
### 3.1 logistics_waybills
- `id` (uuid)
- `order_id` (uuid)
- `carrier_id` (uuid)
- `service_code` (string)
- `waybill_no` (string)
- `status` (string) `created`/`picked`/`in_transit`/`delivered`/`exception`
- `fee_amount` (numeric)
- `label_url` (string)

### 3.2 logistics_tracking_events
- `id` (uuid)
- `waybill_id` (uuid)
- `status` (string)
- `description` (string)
- `occurred_at` (timestamp)

### 3.3 logistics_exceptions
- `id` (uuid)
- `waybill_id` (uuid)
- `type` (string) `delay`/`lost`/`damaged`
- `status` (string) `open`/`resolved`
- `reason` (string)

## 4. 履约任务
### 4.1 fulfillment_tasks
- `id` (uuid)
- `order_id` (uuid)
- `warehouse_id` (uuid)
- `status` (string) `pending`/`picking`/`packed`/`completed`/`exception`
- `assigned_to` (string)

### 4.2 fulfillment_task_logs
- `id` (uuid)
- `task_id` (uuid)
- `action` (string)
- `operator` (string)
- `created_at` (timestamp)

## 5. 逆向物流
### 5.1 reverse_waybills
- `id` (uuid)
- `order_id` (uuid)
- `waybill_no` (string)
- `status` (string) `created`/`in_transit`/`received`/`closed`
- `reason` (string)


## 6. 约束与索引建议
- `logistics_carriers (tenant_uuid, name)` 唯一
- `logistics_carrier_services (tenant_uuid, carrier_id, code)` 唯一
- `logistics_rate_templates (tenant_uuid, name, version)` 唯一
- `logistics_waybills (tenant_uuid, waybill_no)` 唯一
- `logistics_waybills (tenant_uuid, order_id)` 索引
- `logistics_tracking_events (tenant_uuid, waybill_id, occurred_at)` 索引
- `fulfillment_tasks (tenant_uuid, order_id)` 索引
- `reverse_waybills (tenant_uuid, order_id)` 索引

## 7. 枚举值
- carrier.status: `active` | `disabled`
- service.status: `active` | `disabled`
- template.status: `draft` | `published`
- waybill.status: `created` | `picked` | `in_transit` | `delivered` | `exception` | `cancelled`
- task.status: `pending` | `picking` | `packed` | `completed` | `exception`
- reverse.status: `created` | `in_transit` | `received` | `closed`


## 8. 字段约束（建议）
- `logistics_carriers.name` 非空
- `logistics_carriers.type` 非空（enum）
- `logistics_carrier_services.code` 非空
- `logistics_rate_templates.currency` 非空（ISO 4217）
- `logistics_waybills.order_id` 非空
- `logistics_waybills.waybill_no` 非空
- `logistics_tracking_events.occurred_at` 非空
- `fulfillment_tasks.order_id` 非空

## 9. 审计事件（建议）
- carrier.created / carrier.updated
- template.published / template.updated
- waybill.created / waybill.cancelled / waybill.exception
- task.completed / task.exception
- reverse.created / reverse.received


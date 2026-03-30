# Data Model: 履约与物流全模块闭环

## 1. Carrier
- Purpose: 承运商基础信息与可用性配置。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - name (string, 2-64)
  - type (enum: self/sf/jd/cainiao/dhl/other)
  - status (enum: active/disabled)
  - contact_name (string)
  - contact_phone (string)
  - credentials (json)
  - created_at / updated_at (timestamp)
- Rules:
  - (tenant_uuid, name) 唯一。
  - disabled 状态不得用于新建运单。

## 2. CarrierService
- Purpose: 承运商服务维度（快递/重货等）。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - carrier_id (uuid)
  - code (string)
  - name (string)
  - service_type (enum: express/heavy/cold/local)
  - status (enum: active/disabled)
  - config (json)
- Rules:
  - (tenant_uuid, carrier_id, code) 唯一。

## 3. RateTemplate
- Purpose: 运费模板与区域规则。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - name (string)
  - currency (string, ISO 4217)
  - status (enum: draft/published)
  - channels (json)
  - version (int)
- Rules:
  - (tenant_uuid, name, version) 唯一。
  - draft 可编辑，published 只读。

## 4. RateZone
- Purpose: 模板分区计费条款。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - template_id (uuid)
  - region (string)
  - first_weight (decimal)
  - first_fee (decimal)
  - add_weight (decimal)
  - add_fee (decimal)
  - free_threshold (decimal)

## 5. Waybill
- Purpose: 正向发货凭证。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - order_id (uuid)
  - carrier_id (uuid)
  - service_code (string)
  - waybill_no (string)
  - status (enum: created/picked/in_transit/delivered/exception/cancelled)
  - fee_amount (decimal)
  - label_url (string)
- Rules:
  - (tenant_uuid, waybill_no) 唯一。
  - 创建入口主来源：订单触发；人工补单仅异常场景。

## 6. TrackingEvent
- Purpose: 运单轨迹节点。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - waybill_id (uuid)
  - waybill_no (string)
  - event_id (string)
  - status (string)
  - description (string)
  - occurred_at (timestamp)
  - source (enum: provider/manual)
- Rules:
  - 幂等联合键：event_id + waybill_no。
  - 状态源优先级：provider > manual。

## 7. FulfillmentTask
- Purpose: 仓内履约执行任务。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - order_id (uuid)
  - warehouse_id (uuid)
  - status (enum: pending/picking/packed/completed/exception)
  - assigned_to (string)
  - created_at / updated_at (timestamp)
- Rules:
  - 订单进入可履约状态后自动创建。

## 8. FulfillmentException
- Purpose: 履约异常及升级链路。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - task_id (uuid)
  - waybill_id (uuid, nullable)
  - type (enum: delay/lost/damaged/stockout/other)
  - status (enum: open/in_progress/resolved/escalated)
  - reason (string)
  - first_action_at (timestamp, nullable)
  - escalated_at (timestamp, nullable)
- Rules:
  - open 且超过 24h 未 first_action_at 时自动升级。

## 9. ReverseWaybill
- Purpose: 退货/换货/补发逆向物流凭证。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - order_id (uuid)
  - after_sale_id (uuid)
  - waybill_no (string)
  - status (enum: created/in_transit/received/closed)
  - inspection_result (enum: resellable/damaged/rejected, nullable)
  - disposition (enum: restock/scrap/compensate, nullable)
- Rules:
  - 仅在售后审批通过后允许创建。

## Relationships
- Carrier 1:N CarrierService
- RateTemplate 1:N RateZone
- Order 1:N Waybill
- Waybill 1:N TrackingEvent
- Order 1:N FulfillmentTask
- FulfillmentTask 1:N FulfillmentException
- Order 1:N ReverseWaybill

## State Transitions
- Waybill: created → picked → in_transit → delivered | exception | cancelled
- FulfillmentTask: pending → picking → packed → completed | exception
- ReverseWaybill: created → in_transit → received → closed
